package cgen_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/types"
)

// --- FFITypeName Tests ---

func TestFFITypeName_Primitives(t *testing.T) {
	table, intern, queue := helper()

	cases := []struct {
		id   types.TypeID
		want string
	}{
		{types.TypeI8, "signed char"},
		{types.TypeI16, "short"},
		{types.TypeI32, "int"},
		{types.TypeI64, "long long"},
		{types.TypeU8, "unsigned char"},
		{types.TypeU16, "unsigned short"},
		{types.TypeU32, "unsigned int"},
		{types.TypeU64, "unsigned long long"},
		{types.TypeF32, "float"},
		{types.TypeF64, "double"},
		{types.TypeBool, "int"},
		{types.TypeString, "const char*"},
		{types.TypeChar8, "char"},
		{types.TypeVoid, "void"},
		{types.TypeISize, "long long"},
		{types.TypeUSize, "unsigned long long"},
		{types.TypeUnknown, "void"},
	}

	for _, tc := range cases {
		got := cgen.FFITypeName(tc.id, table, intern, queue)
		if got != tc.want {
			t.Errorf("FFITypeName(%v) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestFFITypeName_PointerToU8(t *testing.T) {
	table, intern, queue := helper()
	ptrU8 := table.RegisterPointer(types.TypeU8)
	got := cgen.FFITypeName(ptrU8, table, intern, queue)
	if got != "void*" {
		t.Errorf("FFITypeName(*u8) = %q, want \"void*\"", got)
	}
}

func TestFFITypeName_PointerToVoid(t *testing.T) {
	table, intern, queue := helper()
	ptrVoid := table.RegisterPointer(types.TypeVoid)
	got := cgen.FFITypeName(ptrVoid, table, intern, queue)
	if got != "void*" {
		t.Errorf("FFITypeName(*void) = %q, want \"void*\"", got)
	}
}

func TestFFITypeName_PointerToI32(t *testing.T) {
	table, intern, queue := helper()
	ptrI32 := table.RegisterPointer(types.TypeI32)
	got := cgen.FFITypeName(ptrI32, table, intern, queue)
	if got != "int*" {
		t.Errorf("FFITypeName(*i32) = %q, want \"int*\"", got)
	}
}

func TestFFITypeName_StructUsesAxName(t *testing.T) {
	table, intern, queue := helper()
	fooName := intern.InternString("Foo")
	fooID := table.RegisterStruct(fooName, nil, nil)
	got := cgen.FFITypeName(fooID, table, intern, queue)
	if got != "struct ax_Foo" {
		t.Errorf("FFITypeName(struct Foo) = %q, want \"struct ax_Foo\"", got)
	}
}

func TestFFITypeName_DiffersFromCTypeName(t *testing.T) {
	table, intern, queue := helper()

	// FFITypeName should return raw C types, not ax_ aliases
	ffiI32 := cgen.FFITypeName(types.TypeI32, table, intern, queue)
	cI32 := cgen.CTypeName(types.TypeI32, table, intern, queue)
	if ffiI32 == cI32 {
		t.Errorf("FFITypeName and CTypeName should differ for i32: both returned %q", ffiI32)
	}
	if ffiI32 != "int" {
		t.Errorf("FFITypeName(i32) = %q, want \"int\"", ffiI32)
	}
	if cI32 != "ax_i32" {
		t.Errorf("CTypeName(i32) = %q, want \"ax_i32\"", cI32)
	}
}

// --- FFIDecl Tests ---

func TestFFIDecl_Malloc(t *testing.T) {
	table, intern, queue := helper()

	// extern "C" fn malloc(size: u64) -> *u8
	decl := cgen.FFIFuncDecl{
		Name: "malloc",
		Params: []cgen.FFIParam{
			{Name: "size", TypeID: types.TypeU64},
		},
		ReturnType: table.RegisterPointer(types.TypeU8),
		IsVariadic: false,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "void* malloc(unsigned long long size);" {
		t.Errorf("FFIDecl(malloc) = %q", got)
	}
}

func TestFFIDecl_Printf(t *testing.T) {
	table, intern, queue := helper()

	// extern "C" fn printf(fmt: string, ...) -> i32
	decl := cgen.FFIFuncDecl{
		Name: "printf",
		Params: []cgen.FFIParam{
			{Name: "fmt", TypeID: types.TypeString},
		},
		ReturnType: types.TypeI32,
		IsVariadic: true,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "int printf(const char* fmt, ...);" {
		t.Errorf("FFIDecl(printf) = %q", got)
	}
}

func TestFFIDecl_Free(t *testing.T) {
	table, intern, queue := helper()

	// extern "C" fn free(ptr: *u8) -> void
	ptrU8 := table.RegisterPointer(types.TypeU8)
	decl := cgen.FFIFuncDecl{
		Name: "free",
		Params: []cgen.FFIParam{
			{Name: "ptr", TypeID: ptrU8},
		},
		ReturnType: types.TypeVoid,
		IsVariadic: false,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "void free(void* ptr);" {
		t.Errorf("FFIDecl(free) = %q", got)
	}
}

func TestFFIDecl_NoParams(t *testing.T) {
	table, intern, queue := helper()

	// extern "C" fn getpid() -> i32
	decl := cgen.FFIFuncDecl{
		Name:       "getpid",
		Params:     nil,
		ReturnType: types.TypeI32,
		IsVariadic: false,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	// Must emit (void) not ()
	if got != "int getpid(void);" {
		t.Errorf("FFIDecl(getpid) = %q, want \"int getpid(void);\"", got)
	}
}

func TestFFIDecl_UnnamedParams(t *testing.T) {
	table, intern, queue := helper()

	// extern "C" fn foo(i32, i32) -> void
	decl := cgen.FFIFuncDecl{
		Name: "foo",
		Params: []cgen.FFIParam{
			{Name: "", TypeID: types.TypeI32},
			{Name: "", TypeID: types.TypeI32},
		},
		ReturnType: types.TypeVoid,
		IsVariadic: false,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "void foo(int, int);" {
		t.Errorf("FFIDecl(unnamed params) = %q, want \"void foo(int, int);\"", got)
	}
}

func TestFFIDecl_MathFunctions(t *testing.T) {
	table, intern, queue := helper()

	// extern "C" fn sin(x: f64) -> f64
	decl := cgen.FFIFuncDecl{
		Name: "sin",
		Params: []cgen.FFIParam{
			{Name: "x", TypeID: types.TypeF64},
		},
		ReturnType: types.TypeF64,
		IsVariadic: false,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "double sin(double x);" {
		t.Errorf("FFIDecl(sin) = %q", got)
	}
}

func TestFFIDecl_StructParam(t *testing.T) {
	table, intern, queue := helper()

	barName := intern.InternString("Bar")
	barID := table.RegisterStruct(barName, nil, nil)

	// extern "C" fn process(data: Bar) -> i32
	decl := cgen.FFIFuncDecl{
		Name: "process",
		Params: []cgen.FFIParam{
			{Name: "data", TypeID: barID},
		},
		ReturnType: types.TypeI32,
		IsVariadic: false,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "int process(struct ax_Bar data);" {
		t.Errorf("FFIDecl(struct param) = %q", got)
	}
}

// --- StructAttrAnnotation Tests ---

func TestStructAttrAnnotation_Packed(t *testing.T) {
	got := cgen.StructAttrAnnotation(true, 0)
	if got != "__attribute__((packed))" {
		t.Errorf("packed = %q, want \"__attribute__((packed))\"", got)
	}
}

func TestStructAttrAnnotation_Align(t *testing.T) {
	got := cgen.StructAttrAnnotation(false, 32)
	if got != "__attribute__((aligned(32)))" {
		t.Errorf("align(32) = %q, want \"__attribute__((aligned(32)))\"", got)
	}
}

func TestStructAttrAnnotation_PackedAndAlign(t *testing.T) {
	got := cgen.StructAttrAnnotation(true, 16)
	if got != "__attribute__((packed, aligned(16)))" {
		t.Errorf("packed+align(16) = %q, want \"__attribute__((packed, aligned(16)))\"", got)
	}
}

func TestStructAttrAnnotation_None(t *testing.T) {
	got := cgen.StructAttrAnnotation(false, 0)
	if got != "" {
		t.Errorf("no attrs = %q, want \"\"", got)
	}
}

// --- StructAttrAnnotationMSVC Tests ---

func TestStructAttrAnnotationMSVC_Packed(t *testing.T) {
	pre, post := cgen.StructAttrAnnotationMSVC(true, 0)
	if !strings.Contains(pre, "#pragma pack(push, 1)") {
		t.Errorf("MSVC packed pre = %q, want containing #pragma pack(push, 1)", pre)
	}
	if !strings.Contains(post, "#pragma pack(pop)") {
		t.Errorf("MSVC packed post = %q, want containing #pragma pack(pop)", post)
	}
}

func TestStructAttrDeclspec_Align(t *testing.T) {
	got := cgen.StructAttrDeclspec(32)
	if got != "__declspec(align(32))" {
		t.Errorf("declspec align = %q, want \"__declspec(align(32))\"", got)
	}
}

func TestStructAttrDeclspec_NoAlign(t *testing.T) {
	got := cgen.StructAttrDeclspec(0)
	if got != "" {
		t.Errorf("declspec no align = %q, want \"\"", got)
	}
}

// --- EmitStructWithAttrs Tests ---

func TestEmitStructWithAttrs_NoAttrs(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitStructWithAttrs(w, "ax_Foo", false, 0, func() {
		w.Line("int x;")
	})
	out := buf.String()
	if !strings.Contains(out, "struct ax_Foo {") {
		t.Errorf("no attrs: missing struct declaration in:\n%s", out)
	}
	if !strings.Contains(out, "int x;") {
		t.Errorf("no attrs: missing field in:\n%s", out)
	}
	if strings.Contains(out, "#ifdef") {
		t.Errorf("no attrs: should not contain #ifdef:\n%s", out)
	}
}

func TestEmitStructWithAttrs_Packed(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitStructWithAttrs(w, "ax_PacketHeader", true, 0, func() {
		w.Line("unsigned int magic;")
	})
	out := buf.String()
	if !strings.Contains(out, "#ifdef _MSC_VER") {
		t.Errorf("packed: missing MSVC guard in:\n%s", out)
	}
	if !strings.Contains(out, "#pragma pack(push, 1)") {
		t.Errorf("packed: missing pragma pack in:\n%s", out)
	}
	if !strings.Contains(out, "__attribute__((packed))") {
		t.Errorf("packed: missing GCC attribute in:\n%s", out)
	}
}

func TestEmitStructWithAttrs_Align(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitStructWithAttrs(w, "ax_SimdVec", false, 32, func() {
		w.Line("float data[8];")
	})
	out := buf.String()
	if !strings.Contains(out, "__attribute__((aligned(32)))") {
		t.Errorf("align: missing GCC attribute in:\n%s", out)
	}
	if !strings.Contains(out, "__declspec(align(32))") {
		t.Errorf("align: missing MSVC declspec in:\n%s", out)
	}
}

func TestEmitStructWithAttrs_PackedAndAlign(t *testing.T) {
	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	cgen.EmitStructWithAttrs(w, "ax_Combined", true, 16, func() {
		w.Line("int x;")
	})
	out := buf.String()
	if !strings.Contains(out, "__attribute__((packed, aligned(16)))") {
		t.Errorf("packed+align: missing combined GCC attribute in:\n%s", out)
	}
	if !strings.Contains(out, "#pragma pack(push, 1)") {
		t.Errorf("packed+align: missing pragma pack in:\n%s", out)
	}
	if !strings.Contains(out, "__declspec(align(16))") {
		t.Errorf("packed+align: missing MSVC declspec in:\n%s", out)
	}
}

// --- FFIEmitter Tests ---

func TestFFIEmitter_EmitTo(t *testing.T) {
	table, intern, queue := helper()
	emitter := cgen.NewFFIEmitter(table, intern, queue)

	emitter.AddDecl(cgen.FFIFuncDecl{
		Name: "sin",
		Params: []cgen.FFIParam{
			{Name: "x", TypeID: types.TypeF64},
		},
		ReturnType: types.TypeF64,
	})
	emitter.AddDecl(cgen.FFIFuncDecl{
		Name: "cos",
		Params: []cgen.FFIParam{
			{Name: "x", TypeID: types.TypeF64},
		},
		ReturnType: types.TypeF64,
	})

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	emitter.EmitTo(w)
	out := buf.String()

	if !strings.Contains(out, "/* FFI extern declarations */") {
		t.Errorf("missing FFI header comment in:\n%s", out)
	}
	if !strings.Contains(out, "double sin(double x);") {
		t.Errorf("missing sin declaration in:\n%s", out)
	}
	if !strings.Contains(out, "double cos(double x);") {
		t.Errorf("missing cos declaration in:\n%s", out)
	}
}

func TestFFIEmitter_EmitToEmpty(t *testing.T) {
	table, intern, queue := helper()
	emitter := cgen.NewFFIEmitter(table, intern, queue)

	var buf bytes.Buffer
	w := cgen.NewIndentWriter(&buf)
	emitter.EmitTo(w)
	out := buf.String()

	if out != "" {
		t.Errorf("empty emitter should produce no output, got:\n%s", out)
	}
}

func TestFFIEmitter_Emit(t *testing.T) {
	table, intern, queue := helper()
	emitter := cgen.NewFFIEmitter(table, intern, queue)

	emitter.AddDecl(cgen.FFIFuncDecl{
		Name:       "abort",
		ReturnType: types.TypeVoid,
	})

	results := emitter.Emit()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "void abort(void);" {
		t.Errorf("Emit()[0] = %q, want \"void abort(void);\"", results[0])
	}
}

// --- Determinism ---

func TestFFITypeName_Deterministic(t *testing.T) {
	table, intern, queue := helper()

	r1 := cgen.FFITypeName(types.TypeI32, table, intern, queue)
	r2 := cgen.FFITypeName(types.TypeI32, table, intern, queue)
	if r1 != r2 {
		t.Errorf("non-deterministic: %q vs %q", r1, r2)
	}
}

func TestFFIDecl_Deterministic(t *testing.T) {
	table, intern, queue := helper()

	decl := &cgen.FFIFuncDecl{
		Name: "test",
		Params: []cgen.FFIParam{
			{Name: "x", TypeID: types.TypeI32},
		},
		ReturnType: types.TypeVoid,
		IsVariadic: true,
	}

	r1 := cgen.FFIDecl(decl, table, intern, queue)
	r2 := cgen.FFIDecl(decl, table, intern, queue)
	if r1 != r2 {
		t.Errorf("non-deterministic: %q vs %q", r1, r2)
	}
}

// --- Pointer-to-Struct FFI ---

func TestFFITypeName_PointerToStruct(t *testing.T) {
	table, intern, queue := helper()
	fooName := intern.InternString("Ctx")
	fooID := table.RegisterStruct(fooName, nil, nil)
	ptrFoo := table.RegisterPointer(fooID)
	got := cgen.FFITypeName(ptrFoo, table, intern, queue)
	if got != "struct ax_Ctx*" {
		t.Errorf("FFITypeName(*Ctx) = %q, want \"struct ax_Ctx*\"", got)
	}
}

// --- Multiple params with pointer types ---

func TestFFIDecl_MemcpyLike(t *testing.T) {
	table, intern, queue := helper()
	ptrVoid := table.RegisterPointer(types.TypeVoid)

	// memcpy(dest: *void, src: *void, n: u64) -> *void
	decl := cgen.FFIFuncDecl{
		Name: "memcpy",
		Params: []cgen.FFIParam{
			{Name: "dest", TypeID: ptrVoid},
			{Name: "src", TypeID: ptrVoid},
			{Name: "n", TypeID: types.TypeU64},
		},
		ReturnType: ptrVoid,
	}

	got := cgen.FFIDecl(&decl, table, intern, queue)
	if got != "void* memcpy(void* dest, void* src, unsigned long long n);" {
		t.Errorf("FFIDecl(memcpy) = %q", got)
	}
}
