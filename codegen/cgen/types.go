package cgen

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// TypeDeclQueue tracks which types require C struct/enum declarations,
// and preserves dependency order so forward declarations appear before uses.
//
// Thread-safety: not safe for concurrent use.
type TypeDeclQueue struct {
	seen    map[types.TypeID]bool
	ordered []types.TypeID
}

// NewTypeDeclQueue creates an empty TypeDeclQueue.
func NewTypeDeclQueue() *TypeDeclQueue {
	return &TypeDeclQueue{
		seen: make(map[types.TypeID]bool),
	}
}

// Enqueue adds a type ID to the queue if not already seen.
func (q *TypeDeclQueue) Enqueue(id types.TypeID) {
	if !q.seen[id] {
		q.seen[id] = true
		q.ordered = append(q.ordered, id)
	}
}

// Drain returns all enqueued type IDs in dependency order and clears the queue.
func (q *TypeDeclQueue) Drain() []types.TypeID {
	out := q.ordered
	q.ordered = nil
	return out
}

// Len returns the number of enqueued types.
func (q *TypeDeclQueue) Len() int {
	return len(q.ordered)
}

// CTypeName returns the C11 type name for the given AXIOM TypeID.
// For compound types (structs, slices, sum types, generic instantiations),
// it enqueues entries in the TypeDeclQueue so that forward declarations
// are emitted before first use.
//
// The intern pool is needed to resolve interned names (NameIDs) back to strings.
// If intern is nil, numeric placeholders are used for names.
//
// Determinism guarantee: given the same TypeTable and TypeID,
// CTypeName always returns the identical string.
func CTypeName(id types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	entry := table.Entry(id)

	switch entry.Kind {
	case types.KindPrimitive:
		return primitiveCName(id)

	case types.KindPointer:
		inner := CTypeName(types.TypeID(entry.Extra), table, intern, queue)
		return inner + "*"

	case types.KindSlice:
		inner := CTypeName(types.TypeID(entry.Extra), table, intern, queue)
		name := "ax_slice_" + sanitizeName(inner)
		queue.Enqueue(id)
		return name

	case types.KindStruct:
		queue.Enqueue(id)
		return "struct ax_" + resolveName(entry.NameID, intern)

	case types.KindSum:
		name := buildMangledName(entry.NameID, table.SumInfo(id).Variants, table, intern, queue)
		queue.Enqueue(id)
		return name

	case types.KindGenericInst:
		typeArgs := table.GenericInstArgs(id)
		parts := make([]string, len(typeArgs))
		for i, arg := range typeArgs {
			parts[i] = sanitizeName(CTypeName(arg, table, intern, queue))
		}
		name := "struct ax_" + resolveName(entry.NameID, intern) + "_" + strings.Join(parts, "_")
		queue.Enqueue(id)
		return name

	case types.KindFunction:
		fi := table.FuncInfo(id)
		ret := CTypeName(fi.Return, table, intern, queue)
		params := make([]string, len(fi.Params))
		for i, p := range fi.Params {
			params[i] = CTypeName(p, table, intern, queue)
		}
		if len(params) == 0 {
			return fmt.Sprintf("%s (*)(void)", ret)
		}
		return fmt.Sprintf("%s (*)(%s)", ret, strings.Join(params, ", "))

	case types.KindRef:
		// Heap references are represented as AxRef in the C runtime.
		return "AxRef"

	case types.KindInterface:
		// Interfaces are represented as fat pointers (ptr + vtable).
		queue.Enqueue(id)
		return "struct ax_iface_" + resolveName(entry.NameID, intern)

	case types.KindArray:
		// Fixed-size arrays: Extra stores element TypeID.
		// Size is stored in the entry itself (total bytes).
		inner := CTypeName(types.TypeID(entry.Extra), table, intern, queue)
		return inner // arrays are passed by pointer in C; declaration handles the size

	case types.KindGeneric:
		// Unresolved generic parameter — should never appear in code generation.
		// This indicates a bug in the monomorphization pass.
		panic(fmt.Sprintf("CTypeName: unresolved generic type parameter (TypeID %d)", id))

	default:
		panic(fmt.Sprintf("CTypeName: unknown type kind %d for TypeID %d", entry.Kind, id))
	}
}

// CTypeDecl returns the full C struct/enum/typedef declaration for a compound type.
// Returns "" for primitive types that need no declaration.
// This is used during the declaration emission phase to output struct definitions.
func CTypeDecl(id types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	entry := table.Entry(id)

	switch entry.Kind {
	case types.KindStruct:
		return structDecl(id, table, intern, queue)

	case types.KindSlice:
		return sliceDecl(id, table, intern, queue)

	case types.KindSum:
		return sumTypeDecl(id, table, intern, queue)

	case types.KindGenericInst:
		// Generic instantiations may be structs or sum types.
		// For now, emit a forward declaration comment.
		cname := CTypeName(id, table, intern, queue)
		return fmt.Sprintf("/* forward: %s (generic inst) */", cname)

	default:
		return "" // primitives, pointers, functions don't need declarations
	}
}

// structDecl emits a C struct definition for a struct type.
func structDecl(id types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	entry := table.Entry(id)
	info := table.StructInfo(id)
	name := "ax_" + resolveName(entry.NameID, intern)

	var b strings.Builder
	fmt.Fprintf(&b, "struct %s {\n", name)
	for _, f := range info.Fields {
		ftype := CTypeName(f.TypeID, table, intern, queue)
		fname := resolveName(f.NameID, intern)
		fmt.Fprintf(&b, "    %s %s;\n", ftype, fname)
	}
	b.WriteString("};\n")
	return b.String()
}

// sliceDecl emits a C struct definition for a slice type.
func sliceDecl(id types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	elemID := table.SliceElem(id)
	elemC := CTypeName(elemID, table, intern, queue)
	sliceName := "ax_slice_" + sanitizeName(elemC)

	var b strings.Builder
	fmt.Fprintf(&b, "typedef struct {\n")
	fmt.Fprintf(&b, "    %s* ptr;\n", elemC)
	fmt.Fprintf(&b, "    ax_u64 len;\n")
	fmt.Fprintf(&b, "    ax_u64 cap;\n")
	fmt.Fprintf(&b, "} %s;\n", sliceName)
	return b.String()
}

// sumTypeDecl emits a C tagged union for a sum type.
func sumTypeDecl(id types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	entry := table.Entry(id)
	info := table.SumInfo(id)
	baseName := resolveName(entry.NameID, intern)

	var b strings.Builder

	// Emit tag enum
	fmt.Fprintf(&b, "enum ax_%s_tag {\n", baseName)
	for _, v := range info.Variants {
		vname := resolveName(v.NameID, intern)
		fmt.Fprintf(&b, "    ax_%s_%s = %d,\n", baseName, vname, v.Tag)
	}
	b.WriteString("};\n\n")

	// Emit struct with tag + data union
	cname := "ax_" + baseName
	fmt.Fprintf(&b, "struct %s {\n", cname)
	fmt.Fprintf(&b, "    enum ax_%s_tag tag;\n", baseName)
	b.WriteString("    union {\n")
	for _, v := range info.Variants {
		vname := resolveName(v.NameID, intern)
		if v.PayloadType != types.TypeUnknown {
			payloadC := CTypeName(v.PayloadType, table, intern, queue)
			fmt.Fprintf(&b, "        %s %s;\n", payloadC, vname)
		}
	}
	b.WriteString("    } data;\n")
	b.WriteString("};\n")

	return b.String()
}

// primitiveCName maps primitive TypeIDs to their ax_ C type names.
func primitiveCName(id types.TypeID) string {
	switch id {
	case types.TypeI8:
		return "ax_i8"
	case types.TypeI16:
		return "ax_i16"
	case types.TypeI32:
		return "ax_i32"
	case types.TypeI64:
		return "ax_i64"
	case types.TypeU8:
		return "ax_u8"
	case types.TypeU16:
		return "ax_u16"
	case types.TypeU32:
		return "ax_u32"
	case types.TypeU64:
		return "ax_u64"
	case types.TypeF32:
		return "ax_f32"
	case types.TypeF64:
		return "ax_f64"
	case types.TypeBool:
		return "ax_bool"
	case types.TypeString:
		return "ax_string"
	case types.TypeChar8:
		return "ax_char"
	case types.TypeVoid:
		return "void"
	case types.TypeISize:
		return "ax_isize"
	case types.TypeUSize:
		return "ax_usize"
	case types.TypeUnknown:
		return "void" // unknown/never maps to void in C
	default:
		return fmt.Sprintf("ax_type_%d", id)
	}
}

// resolveName converts an interned NameID back to a string.
// Falls back to a numeric placeholder if the intern pool is nil or NameID is 0.
func resolveName(nameID uint32, intern *ast.InternPool) string {
	if nameID == 0 || intern == nil {
		return fmt.Sprintf("_anon_%d", nameID)
	}
	return intern.Get(nameID)
}

// buildMangledName creates a unique C name for a sum type by including variant info.
func buildMangledName(nameID uint32, variants []types.VariantInfo, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	name := resolveName(nameID, intern)
	return "struct ax_" + name
}

// sanitizeName replaces characters invalid in C identifiers with underscores.
// Valid characters: [a-zA-Z0-9_]
func sanitizeName(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
