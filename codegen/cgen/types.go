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
	seen        map[types.TypeID]bool
	ordered     []types.TypeID
	emittedTags map[string]bool
}

// NewTypeDeclQueue creates an empty TypeDeclQueue.
func NewTypeDeclQueue() *TypeDeclQueue {
	return &TypeDeclQueue{
		seen:        make(map[types.TypeID]bool),
		emittedTags: make(map[string]bool),
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
			typeNameStr := CTypeName(arg, table, intern, queue)
			parts[i] = sanitizeName(typeNameStr)
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
		inner := CTypeName(table.ArrayElem(id), table, intern, queue)
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
		// Find the underlying template structure or sum type
		var templateID types.TypeID
		var templateEntry *types.TypeEntry
		for idx := 0; idx < table.Count(); idx++ {
			e := table.Entry(types.TypeID(idx))
			if (e.Kind == types.KindStruct || e.Kind == types.KindSum) && e.NameID == entry.NameID {
				templateID = types.TypeID(idx)
				templateEntry = e
				break
			}
		}

		if templateEntry == nil {
			panic(fmt.Sprintf("CTypeDecl: generic base template not found for %s", resolveName(entry.NameID, intern)))
		}

		if templateEntry.Kind == types.KindStruct {
			return genericInstStructDecl(id, templateID, table, intern, queue)
		} else if templateEntry.Kind == types.KindSum {
			if queue.emittedTags == nil {
				queue.emittedTags = make(map[string]bool)
			}
			return genericInstSumTypeDecl(id, templateID, table, intern, queue, queue.emittedTags)
		}
		return ""

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
		fEntry := table.Entry(f.TypeID)
		fname := resolveName(f.NameID, intern)
		if fEntry.Kind == types.KindArray {
			elemID := table.ArrayElem(f.TypeID)
			elemC := CTypeName(elemID, table, intern, queue)
			length := table.ArrayLength(f.TypeID)
			fmt.Fprintf(&b, "    %s %s[%d];\n", elemC, fname, length)
		} else {
			ftype := CTypeName(f.TypeID, table, intern, queue)
			fmt.Fprintf(&b, "    %s %s;\n", ftype, fname)
		}
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

	// Emit constructor functions for each variant (non-generic)
	for _, v := range info.Variants {
		vname := resolveName(v.NameID, intern)
		vnameLower := strings.ToLower(vname)
		structName := "struct " + cname
		if v.PayloadType == types.TypeUnknown {
			fmt.Fprintf(&b, "\nstatic inline %s ax_%s_%s(void) {\n",
				structName, baseName, vnameLower)
			fmt.Fprintf(&b, "    %s _result;\n", structName)
			fmt.Fprintf(&b, "    _result.tag = ax_%s_%s;\n", baseName, vname)
			fmt.Fprintf(&b, "    return _result;\n")
			b.WriteString("}\n")
		} else {
			payloadC := CTypeName(v.PayloadType, table, intern, queue)
			fmt.Fprintf(&b, "\nstatic inline %s ax_%s_%s(%s value) {\n",
				structName, baseName, vnameLower, payloadC)
			fmt.Fprintf(&b, "    %s _result;\n", structName)
			fmt.Fprintf(&b, "    _result.tag = ax_%s_%s;\n", baseName, vname)
			fmt.Fprintf(&b, "    _result.data.%s = value;\n", vname)
			fmt.Fprintf(&b, "    return _result;\n")
			b.WriteString("}\n")
		}
	}

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

func genericInstStructDecl(id types.TypeID, templateID types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	cname := CTypeName(id, table, intern, queue)
	mangledName := strings.TrimPrefix(cname, "struct ")

	info := table.StructInfo(templateID)
	params := info.GenericParams
	args := table.GenericInstArgs(id)

	var b strings.Builder
	fmt.Fprintf(&b, "struct %s {\n", mangledName)
	for _, f := range info.Fields {
		fname := resolveName(f.NameID, intern)
		specType := table.SubstituteGenericType(f.TypeID, params, args)
		fEntry := table.Entry(specType)
		if fEntry.Kind == types.KindArray {
			elemID := table.ArrayElem(specType)
			elemC := CTypeName(elemID, table, intern, queue)
			length := table.ArrayLength(specType)
			fmt.Fprintf(&b, "    %s %s[%d];\n", elemC, fname, length)
		} else {
			ftype := CTypeName(specType, table, intern, queue)
			fmt.Fprintf(&b, "    %s %s;\n", ftype, fname)
		}
	}
	b.WriteString("};\n")
	return b.String()
}

func genericInstSumTypeDecl(id types.TypeID, templateID types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue, emittedTags map[string]bool) string {
	templateEntry := table.Entry(templateID)
	info := table.SumInfo(templateID)
	baseName := resolveName(templateEntry.NameID, intern)
	cname := CTypeName(id, table, intern, queue)
	mangledName := strings.TrimPrefix(cname, "struct ")

	params := info.GenericParams
	args := table.GenericInstArgs(id)

	var b strings.Builder

	// 1. Emit the tag enum once per template name
	if !emittedTags[baseName] {
		emittedTags[baseName] = true
		fmt.Fprintf(&b, "enum ax_%s_tag {\n", baseName)
		for _, v := range info.Variants {
			vname := resolveName(v.NameID, intern)
			fmt.Fprintf(&b, "    ax_%s_%s = %d,\n", baseName, vname, v.Tag)
		}
		b.WriteString("};\n\n")
	}

	// 2. Emit the concrete tagged union struct
	fmt.Fprintf(&b, "struct %s {\n", mangledName)
	fmt.Fprintf(&b, "    enum ax_%s_tag tag;\n", baseName)

	hasPayload := false
	for _, v := range info.Variants {
		if v.PayloadType != types.TypeUnknown {
			hasPayload = true
			break
		}
	}

	if hasPayload {
		b.WriteString("    union {\n")
		for _, v := range info.Variants {
			if v.PayloadType != types.TypeUnknown {
				vname := resolveName(v.NameID, intern)
				specType := table.SubstituteGenericType(v.PayloadType, params, args)
				payloadC := CTypeName(specType, table, intern, queue)
				fmt.Fprintf(&b, "        %s %s;\n", payloadC, vname)
			}
		}
		b.WriteString("    } data;\n")
	}
	b.WriteString("};\n")

	// Emit constructor functions for each variant (generic instantiation)
	for _, v := range info.Variants {
		vname := resolveName(v.NameID, intern)
		vnameLower := strings.ToLower(vname)
		structName := "struct " + mangledName
		if v.PayloadType == types.TypeUnknown {
			fmt.Fprintf(&b, "\nstatic inline %s ax_%s_%s(void) {\n",
				structName, mangledName, vnameLower)
			fmt.Fprintf(&b, "    %s _result;\n", structName)
			fmt.Fprintf(&b, "    _result.tag = ax_%s_%s;\n", baseName, vname)
			fmt.Fprintf(&b, "    return _result;\n")
			b.WriteString("}\n")
		} else {
			specType := table.SubstituteGenericType(v.PayloadType, params, args)
			payloadC := CTypeName(specType, table, intern, queue)
			fmt.Fprintf(&b, "\nstatic inline %s ax_%s_%s(%s value) {\n",
				structName, mangledName, vnameLower, payloadC)
			fmt.Fprintf(&b, "    %s _result;\n", structName)
			fmt.Fprintf(&b, "    _result.tag = ax_%s_%s;\n", baseName, vname)
			fmt.Fprintf(&b, "    _result.data.%s = value;\n", vname)
			fmt.Fprintf(&b, "    return _result;\n")
			b.WriteString("}\n")
		}
	}

	return b.String()
}

