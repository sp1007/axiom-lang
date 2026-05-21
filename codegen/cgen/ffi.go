package cgen

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// FFITypeName returns the raw C type name for a type used in FFI context.
// Unlike CTypeName, FFI types use standard C names (int, char*, etc.)
// rather than ax_ aliases, to ensure correct ABI at the FFI boundary.
//
// Determinism guarantee: given the same TypeTable and TypeID,
// FFITypeName always returns the identical string.
func FFITypeName(id types.TypeID, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	// Handle primitives directly — no need to look up entry for well-known IDs.
	switch id {
	case types.TypeI8:
		return "signed char"
	case types.TypeI16:
		return "short"
	case types.TypeI32:
		return "int"
	case types.TypeI64:
		return "long long"
	case types.TypeU8:
		return "unsigned char"
	case types.TypeU16:
		return "unsigned short"
	case types.TypeU32:
		return "unsigned int"
	case types.TypeU64:
		return "unsigned long long"
	case types.TypeF32:
		return "float"
	case types.TypeF64:
		return "double"
	case types.TypeBool:
		return "int" // C99 _Bool promoted to int at FFI boundary
	case types.TypeString:
		return "const char*" // null-terminated C string, not ax_string
	case types.TypeChar8:
		return "char"
	case types.TypeVoid, types.TypeUnknown:
		return "void"
	case types.TypeISize:
		return "long long" // platform-dependent; 64-bit default
	case types.TypeUSize:
		return "unsigned long long"
	}

	// Non-primitive types require entry lookup.
	entry := table.Entry(id)

	switch entry.Kind {
	case types.KindPointer:
		inner := FFITypeName(types.TypeID(entry.Extra), table, intern, queue)
		if inner == "void" || inner == "unsigned char" {
			return "void*"
		}
		return inner + "*"

	case types.KindStruct:
		// Struct types in FFI use the same ax_ struct name (shared layout).
		return CTypeName(id, table, intern, queue)

	default:
		// For any other compound type (slice, sum, generic inst, etc.),
		// fall back to the internal CTypeName.
		return CTypeName(id, table, intern, queue)
	}
}

// FFIParam holds a single parameter in an extern "C" function declaration.
type FFIParam struct {
	Name   string       // parameter name (may be empty for unnamed params)
	TypeID types.TypeID // AXIOM type of the parameter
}

// FFIFuncDecl holds all information needed to emit an extern "C" function prototype.
type FFIFuncDecl struct {
	Name       string     // C function name (no mangling)
	Params     []FFIParam // parameters
	ReturnType types.TypeID
	IsVariadic bool // true if the function is variadic (...)
}

// FFIDecl generates a C function prototype string for an extern "C" function declaration.
// The returned string includes the trailing semicolon.
//
// Examples:
//
//	int printf(const char* fmt, ...);
//	void* malloc(unsigned long long size);
//	void free(void* ptr);
func FFIDecl(decl *FFIFuncDecl, table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) string {
	ret := FFITypeName(decl.ReturnType, table, intern, queue)

	params := make([]string, 0, len(decl.Params)+1)
	for _, p := range decl.Params {
		ctype := FFITypeName(p.TypeID, table, intern, queue)
		if p.Name != "" {
			params = append(params, fmt.Sprintf("%s %s", ctype, p.Name))
		} else {
			params = append(params, ctype)
		}
	}
	if decl.IsVariadic {
		params = append(params, "...")
	}

	// C requires (void) for zero-parameter functions, not ()
	if len(params) == 0 {
		params = []string{"void"}
	}

	return fmt.Sprintf("%s %s(%s);", ret, decl.Name, strings.Join(params, ", "))
}

// StructAttrAnnotation returns the C compiler attribute annotation string
// for a struct with @packed and/or @align(N) attributes.
//
// If both packed and align are specified, the attributes are combined.
// Returns "" if no layout attributes are specified.
//
// The annotation is the string to insert between "struct" and the struct name,
// e.g., "__attribute__((packed, aligned(32)))".
func StructAttrAnnotation(packed bool, alignN int) string {
	if !packed && alignN <= 0 {
		return ""
	}

	var attrs []string
	if packed {
		attrs = append(attrs, "packed")
	}
	if alignN > 0 {
		attrs = append(attrs, fmt.Sprintf("aligned(%d)", alignN))
	}

	return fmt.Sprintf("__attribute__((%s))", strings.Join(attrs, ", "))
}

// StructAttrAnnotationMSVC returns the MSVC-compatible attribute string
// for a struct with @packed and/or @align(N) attributes.
//
// Returns a pair of (pre, post) strings where:
//   - pre is placed before the struct definition (e.g., "#pragma pack(push, 1)")
//   - post is placed after the struct definition (e.g., "#pragma pack(pop)")
//
// If no attributes are needed, both strings are empty.
func StructAttrAnnotationMSVC(packed bool, alignN int) (pre, post string) {
	if !packed && alignN <= 0 {
		return "", ""
	}

	var preParts, postParts []string

	if packed {
		preParts = append(preParts, "#pragma pack(push, 1)")
		postParts = append(postParts, "#pragma pack(pop)")
	}

	// MSVC align is handled via __declspec(align(N)) on the struct itself
	// so it's not in the pragma pre/post — it goes inline.
	// For simplicity, we put it in pre as a comment marker.

	return strings.Join(preParts, "\n"), strings.Join(postParts, "\n")
}

// StructAttrDeclspec returns the MSVC __declspec(align(N)) string if needed.
// Returns "" if no alignment is specified.
func StructAttrDeclspec(alignN int) string {
	if alignN <= 0 {
		return ""
	}
	return fmt.Sprintf("__declspec(align(%d))", alignN)
}

// EmitStructWithAttrs emits a struct declaration with platform-specific
// attribute annotations for @packed and @align.
// It wraps the declaration in #ifdef _MSC_VER guards to support both
// GCC/Clang and MSVC compilers.
func EmitStructWithAttrs(w *IndentWriter, name string, packed bool, alignN int, emitFields func()) {
	hasAttrs := packed || alignN > 0

	if !hasAttrs {
		w.Linef("struct %s {", name)
		w.Indent()
		emitFields()
		w.Dedent()
		w.Line("};")
		return
	}

	// Emit with compiler-specific attributes
	gccAttr := StructAttrAnnotation(packed, alignN)
	msvcDeclspec := StructAttrDeclspec(alignN)
	msvcPre, msvcPost := StructAttrAnnotationMSVC(packed, alignN)

	w.Line("#ifdef _MSC_VER")
	if msvcPre != "" {
		w.Line(msvcPre)
	}
	if msvcDeclspec != "" {
		w.Linef("%s struct %s {", msvcDeclspec, name)
	} else {
		w.Linef("struct %s {", name)
	}
	w.Indent()
	emitFields()
	w.Dedent()
	w.Line("};")
	if msvcPost != "" {
		w.Line(msvcPost)
	}
	w.Line("#else")
	w.Linef("struct %s %s {", gccAttr, name)
	w.Indent()
	emitFields()
	w.Dedent()
	w.Line("};")
	w.Line("#endif")
}

// FFIEmitter accumulates extern "C" declarations and emits them as C prototypes.
// It is used by the DeclEmitter to collect all FFI declarations during module processing.
type FFIEmitter struct {
	table  *types.TypeTable
	intern *ast.InternPool
	queue  *TypeDeclQueue
	decls  []FFIFuncDecl
}

// NewFFIEmitter creates a new FFIEmitter.
func NewFFIEmitter(table *types.TypeTable, intern *ast.InternPool, queue *TypeDeclQueue) *FFIEmitter {
	return &FFIEmitter{
		table:  table,
		intern: intern,
		queue:  queue,
	}
}

// AddDecl adds an extern "C" function declaration to be emitted.
func (e *FFIEmitter) AddDecl(decl FFIFuncDecl) {
	e.decls = append(e.decls, decl)
}

// Emit returns all accumulated FFI declarations as C prototype strings.
func (e *FFIEmitter) Emit() []string {
	results := make([]string, len(e.decls))
	for i, d := range e.decls {
		results[i] = FFIDecl(&d, e.table, e.intern, e.queue)
	}
	return results
}

// EmitTo writes all FFI declarations to the given IndentWriter.
func (e *FFIEmitter) EmitTo(w *IndentWriter) {
	if len(e.decls) == 0 {
		return
	}
	w.Line("/* FFI extern declarations */")
	for _, d := range e.decls {
		w.Line(FFIDecl(&d, e.table, e.intern, e.queue))
	}
	w.BlankLine()
}
