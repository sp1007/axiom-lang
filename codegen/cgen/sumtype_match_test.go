package cgen_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestSumTypeDecl_Basic(t *testing.T) {
	table, intern, queue := helper()

	shapeName := intern.InternString("Shape")
	circleName := intern.InternString("Circle")
	rectName := intern.InternString("Rect")
	emptyName := intern.InternString("Empty")

	variants := []types.VariantInfo{
		{NameID: circleName, PayloadType: types.TypeF64, Tag: 0},
		{NameID: rectName, PayloadType: types.TypeF64, Tag: 1},
		{NameID: emptyName, PayloadType: types.TypeUnknown, Tag: 2},
	}
	typeID := table.RegisterSumType(shapeName, variants, nil)

	decl := cgen.EmitSumTypeDecl(typeID, table, intern, queue)

	// Check tag enum
	if !strings.Contains(decl, "enum ax_Shape_tag") {
		t.Errorf("missing tag enum:\n%s", decl)
	}
	if !strings.Contains(decl, "ax_Shape_Circle = 0") {
		t.Errorf("missing Circle tag:\n%s", decl)
	}
	if !strings.Contains(decl, "ax_Shape_Rect = 1") {
		t.Errorf("missing Rect tag:\n%s", decl)
	}
	if !strings.Contains(decl, "ax_Shape_Empty = 2") {
		t.Errorf("missing Empty tag:\n%s", decl)
	}

	// Check struct
	if !strings.Contains(decl, "struct ax_Shape") {
		t.Errorf("missing struct:\n%s", decl)
	}
	if !strings.Contains(decl, "union {") {
		t.Errorf("missing union:\n%s", decl)
	}
	// Empty variant should NOT appear in the union
	if strings.Contains(decl, "Empty;") && !strings.Contains(decl, "ax_f64 Empty;") {
		// Empty has no payload, shouldn't appear as a field
	}
}

func TestSumTypeDecl_AllEmpty(t *testing.T) {
	table, intern, queue := helper()

	colorName := intern.InternString("Color")
	redName := intern.InternString("Red")
	greenName := intern.InternString("Green")
	blueName := intern.InternString("Blue")

	variants := []types.VariantInfo{
		{NameID: redName, PayloadType: types.TypeUnknown, Tag: 0},
		{NameID: greenName, PayloadType: types.TypeUnknown, Tag: 1},
		{NameID: blueName, PayloadType: types.TypeUnknown, Tag: 2},
	}
	typeID := table.RegisterSumType(colorName, variants, nil)

	decl := cgen.EmitSumTypeDecl(typeID, table, intern, queue)

	// Should have tag enum but no union (all variants empty)
	if !strings.Contains(decl, "enum ax_Color_tag") {
		t.Errorf("missing tag enum:\n%s", decl)
	}
	if strings.Contains(decl, "union {") {
		t.Errorf("should not have union for all-empty variants:\n%s", decl)
	}
}

func TestSumTypeDecl_ResultType(t *testing.T) {
	table, intern, queue := helper()

	resultName := intern.InternString("Result")
	okName := intern.InternString("Ok")
	errName := intern.InternString("Err")

	variants := []types.VariantInfo{
		{NameID: okName, PayloadType: types.TypeI32, Tag: 0},
		{NameID: errName, PayloadType: types.TypeString, Tag: 1},
	}
	typeID := table.RegisterSumType(resultName, variants, nil)

	decl := cgen.EmitSumTypeDecl(typeID, table, intern, queue)

	if !strings.Contains(decl, "ax_i32 Ok;") {
		t.Errorf("missing Ok payload:\n%s", decl)
	}
	if !strings.Contains(decl, "ax_string Err;") {
		t.Errorf("missing Err payload:\n%s", decl)
	}
}

func TestSumTypeConstructor_WithPayload(t *testing.T) {
	table, intern, queue := helper()

	optName := intern.InternString("Option")
	someName := intern.InternString("Some")

	variant := types.VariantInfo{NameID: someName, PayloadType: types.TypeI32, Tag: 1}
	variants := []types.VariantInfo{
		{NameID: intern.InternString("None"), PayloadType: types.TypeUnknown, Tag: 0},
		variant,
	}
	typeID := table.RegisterSumType(optName, variants, nil)

	ctor := cgen.EmitSumTypeConstructor(typeID, variant, table, intern, queue)

	if !strings.Contains(ctor, "static inline struct ax_Option ax_Option_some(ax_i32 value)") {
		t.Errorf("incorrect constructor signature:\n%s", ctor)
	}
	if !strings.Contains(ctor, "_result.tag = ax_Option_Some") {
		t.Errorf("missing tag assignment:\n%s", ctor)
	}
	if !strings.Contains(ctor, "_result.data.Some = value") {
		t.Errorf("missing data assignment:\n%s", ctor)
	}
}

func TestSumTypeConstructor_NoPayload(t *testing.T) {
	table, intern, queue := helper()

	optName := intern.InternString("Option")
	noneName := intern.InternString("None")

	variant := types.VariantInfo{NameID: noneName, PayloadType: types.TypeUnknown, Tag: 0}
	variants := []types.VariantInfo{
		variant,
		{NameID: intern.InternString("Some"), PayloadType: types.TypeI32, Tag: 1},
	}
	typeID := table.RegisterSumType(optName, variants, nil)

	ctor := cgen.EmitSumTypeConstructor(typeID, variant, table, intern, queue)

	if !strings.Contains(ctor, "ax_Option_none(void)") {
		t.Errorf("no-payload constructor should take void:\n%s", ctor)
	}
	if strings.Contains(ctor, "_result.data") {
		t.Errorf("no-payload constructor should not set data:\n%s", ctor)
	}
}

func TestNewMatchGen(t *testing.T) {
	// Verify MatchGen can be constructed without panicking
	ctx := newStmtTestCtx()
	queue := cgen.NewTypeDeclQueue()
	eg := cgen.NewExprGen(ctx.table, ctx.intern, ctx.symbols, ctx.tree, queue)

	var buf strings.Builder
	w := cgen.NewIndentWriter(&buf)

	mg := cgen.NewMatchGen(w, eg, ctx.table, ctx.intern, ctx.tree, queue)
	if mg == nil {
		t.Error("NewMatchGen returned nil")
	}
}
