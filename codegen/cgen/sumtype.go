package cgen

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// EmitSumTypeDecl generates the full C declaration for a sum type:
// - A tag enum with variant names
// - A tagged union struct with a discriminant and data union
//
// Example output for `type Shape = Circle(f64) | Rect(f64, f64) | Empty`:
//
//	enum ax_Shape_tag { ax_Shape_Circle = 0, ax_Shape_Rect = 1, ax_Shape_Empty = 2 };
//	struct ax_Shape { enum ax_Shape_tag tag; union { ax_f64 Circle; ax_f64 Rect[2]; } data; };
func EmitSumTypeDecl(
	typeID types.TypeID,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	entry := table.Entry(typeID)
	info := table.SumInfo(typeID)
	baseName := resolveName(entry.NameID, intern)

	var b strings.Builder

	// 1. Tag enum
	fmt.Fprintf(&b, "enum ax_%s_tag {\n", baseName)
	for _, v := range info.Variants {
		vname := resolveName(v.NameID, intern)
		fmt.Fprintf(&b, "    ax_%s_%s = %d,\n", baseName, vname, v.Tag)
	}
	b.WriteString("};\n\n")

	// 2. Tagged union struct
	cname := "ax_" + baseName
	fmt.Fprintf(&b, "struct %s {\n", cname)
	fmt.Fprintf(&b, "    enum ax_%s_tag tag;\n", baseName)

	// Check if any variant has a payload
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
				payloadC := CTypeName(v.PayloadType, table, intern, queue)
				fmt.Fprintf(&b, "        %s %s;\n", payloadC, vname)
			}
		}
		b.WriteString("    } data;\n")
	}
	b.WriteString("};\n")

	return b.String()
}

// EmitSumTypeConstructor generates a constructor function for a sum type variant.
//
// Example: ax_Option_i32_some(ax_i32 value) → returns an Option with tag=Some
func EmitSumTypeConstructor(
	typeID types.TypeID,
	variant types.VariantInfo,
	table *types.TypeTable,
	intern *ast.InternPool,
	queue *TypeDeclQueue,
) string {
	entry := table.Entry(typeID)
	baseName := resolveName(entry.NameID, intern)
	vname := resolveName(variant.NameID, intern)
	structName := "struct ax_" + baseName

	var b strings.Builder

	if variant.PayloadType == types.TypeUnknown {
		// No payload: constructor takes no args
		fmt.Fprintf(&b, "static inline %s ax_%s_%s(void) {\n",
			structName, baseName, strings.ToLower(vname))
		fmt.Fprintf(&b, "    %s _result;\n", structName)
		fmt.Fprintf(&b, "    _result.tag = ax_%s_%s;\n", baseName, vname)
		fmt.Fprintf(&b, "    return _result;\n")
		b.WriteString("}\n")
	} else {
		payloadC := CTypeName(variant.PayloadType, table, intern, queue)
		fmt.Fprintf(&b, "static inline %s ax_%s_%s(%s value) {\n",
			structName, baseName, strings.ToLower(vname), payloadC)
		fmt.Fprintf(&b, "    %s _result;\n", structName)
		fmt.Fprintf(&b, "    _result.tag = ax_%s_%s;\n", baseName, vname)
		fmt.Fprintf(&b, "    _result.data.%s = value;\n", vname)
		fmt.Fprintf(&b, "    return _result;\n")
		b.WriteString("}\n")
	}

	return b.String()
}
