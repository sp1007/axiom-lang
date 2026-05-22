package cgen

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// MatchGen generates C code for AXIOM match expressions/statements.
// It transforms match arms into a C switch statement on the discriminant tag.
type MatchGen struct {
	W       *IndentWriter
	ExprGen *ExprGen
	Table   *types.TypeTable
	Intern  *ast.InternPool
	Tree    *ast.AstTree
	Queue   *TypeDeclQueue
}

// NewMatchGen creates a new MatchGen.
func NewMatchGen(
	w *IndentWriter,
	exprGen *ExprGen,
	table *types.TypeTable,
	intern *ast.InternPool,
	tree *ast.AstTree,
	queue *TypeDeclQueue,
) *MatchGen {
	return &MatchGen{
		W:       w,
		ExprGen: exprGen,
		Table:   table,
		Intern:  intern,
		Tree:    tree,
		Queue:   queue,
	}
}

// EmitMatchStmt emits a match statement as a C switch.
// The match expression must be a sum type; the arms switch on .tag.
//
// matchNodeIdx is the AST node index of the NodeMatchStmt.
func (mg *MatchGen) EmitMatchStmt(matchNodeIdx uint32) {
	mg.EmitMatchStmtWithReturning(matchNodeIdx, false)
}

func (mg *MatchGen) EmitMatchStmtWithReturning(matchNodeIdx uint32, returning bool) {
	node := mg.Tree.Node(matchNodeIdx)
	children := mg.Tree.Children(matchNodeIdx)

	if len(children) < 2 {
		mg.W.Line("/* invalid match: missing children */")
		return
	}

	// First child: the discriminant expression
	discrimExpr := mg.ExprGen.Emit(children[0])

	// Get the type of the discriminant to resolve variant names
	discrimTypeID := mg.ExprGen.NodeType(children[0])
	_ = node // silence unused

	mg.W.Linef("switch ((%s).tag) {", discrimExpr)

	// Process match arms (children 1..n)
	for i := 1; i < len(children); i++ {
		armNode := mg.Tree.Node(children[i])
		if armNode.Kind != ast.NodeMatchArm {
			continue
		}

		armChildren := mg.Tree.Children(children[i])
		if len(armChildren) < 2 {
			continue
		}

		patternNode := mg.Tree.Node(armChildren[0])
		mg.emitMatchArmWithReturning(discrimExpr, discrimTypeID, patternNode, armChildren, returning)
	}

	// Add exhaustiveness guard
	mg.W.Line("    default: {")
	mg.W.Line("        /* unreachable: exhaustiveness checked by type checker */")
	mg.W.Line("        __builtin_unreachable();")
	mg.W.Line("    }")
	mg.W.Line("}")
}

// emitMatchArm emits a single match arm as a case clause.
func (mg *MatchGen) emitMatchArm(
	discrimExpr string,
	discrimTypeID types.TypeID,
	patternNode *ast.AstNode,
	armChildren []uint32,
) {
	mg.emitMatchArmWithReturning(discrimExpr, discrimTypeID, patternNode, armChildren, false)
}

func (mg *MatchGen) emitMatchArmWithReturning(
	discrimExpr string,
	discrimTypeID types.TypeID,
	patternNode *ast.AstNode,
	armChildren []uint32,
	returning bool,
) {
	entry := mg.Table.Entry(discrimTypeID)
	baseName := resolveName(entry.NameID, mg.Intern)

	switch patternNode.Kind {
	case ast.NodeVariantPat:
		// Variant pattern: match a specific variant
		variantName := string(mg.Tree.TokenText(patternNode.TokenIdx))
		mg.W.Linef("    case ax_%s_%s: {", baseName, variantName)

		// Emit binding: extract payload into a local variable
		if patternNode.FirstChild != ast.NullIdx {
			bindingNode := mg.Tree.Node(patternNode.FirstChild)
			if bindingNode.Kind == ast.NodeBindingPat {
				bindName := string(mg.Tree.TokenText(bindingNode.TokenIdx))
				// Find the variant's payload type
				info := mg.Table.SumInfo(discrimTypeID)
				for _, v := range info.Variants {
					vname := resolveName(v.NameID, mg.Intern)
					if vname == variantName && v.PayloadType != types.TypeUnknown {
						payloadC := CTypeName(v.PayloadType, mg.Table, mg.Intern, mg.Queue)
						mg.W.Linef("        %s %s = (%s).data.%s;",
							payloadC, bindName, discrimExpr, variantName)
						mg.W.Linef("        (void)%s;", bindName)
						break
					}
				}
			}
		}

		// Emit arm body (second child onwards)
		mg.emitArmBodyWithReturning(armChildren, returning)

		mg.W.Line("        break;")
		mg.W.Line("    }")

	case ast.NodeBindingPat:
		// Binding pattern can be a variant with no payload (e.g. None)
		symIdx := patternNode.Payload
		isVariant := false
		var variantName string
		if symIdx != 0 && mg.ExprGen.Symbols != nil && int(symIdx) < len(mg.ExprGen.Symbols.Symbols) {
			sym := mg.ExprGen.Symbols.SymbolAt(symIdx)
			if sym.Kind == sema.SymVariant {
				isVariant = true
				variantName = resolveName(sym.NameID, mg.Intern)
			}
		}

		if isVariant {
			mg.W.Linef("    case ax_%s_%s: {", baseName, variantName)
			mg.emitArmBodyWithReturning(armChildren, returning)
			mg.W.Line("        break;")
			mg.W.Line("    }")
		} else {
			// Otherwise it's a wildcard / fallback variable binding
			mg.W.Line("    default: {")
			bindName := string(mg.Tree.TokenText(patternNode.TokenIdx))
			if bindName != "_" && bindName != "" {
				discrimC := CTypeName(discrimTypeID, mg.Table, mg.Intern, mg.Queue)
				mg.W.Linef("        %s %s = %s;", discrimC, bindName, discrimExpr)
				mg.W.Linef("        (void)%s;", bindName)
			}
			mg.emitArmBodyWithReturning(armChildren, returning)
			mg.W.Line("        break;")
			mg.W.Line("    }")
		}

	case ast.NodeWildcardPat:
		// Wildcard: default case
		mg.W.Line("    default: {")
		mg.emitArmBodyWithReturning(armChildren, returning)
		mg.W.Line("        break;")
		mg.W.Line("    }")

	case ast.NodeLiteralPat:
		// Literal pattern: case <literal>:
		litValue := mg.ExprGen.Emit(armChildren[0])
		mg.W.Linef("    case %s: {", litValue)
		mg.emitArmBodyWithReturning(armChildren, returning)
		mg.W.Line("        break;")
		mg.W.Line("    }")
	}
}

func (mg *MatchGen) emitArmBody(armChildren []uint32) {
	mg.emitArmBodyWithReturning(armChildren, false)
}

func (mg *MatchGen) emitArmBodyWithReturning(armChildren []uint32, returning bool) {
	if len(armChildren) >= 2 {
		mg.W.Indent()
		mg.W.Indent()
		sg := &StmtGen{
			W:       mg.W,
			ExprGen: mg.ExprGen,
			Defers:  NewDeferStack(),
			Table:   mg.Table,
			Intern:  mg.Intern,
			Symbols: mg.ExprGen.Symbols,
			Tree:    mg.Tree,
			Queue:   mg.Queue,
		}

		bodyNodeIdx := armChildren[1]
		bodyNode := mg.Tree.Node(bodyNodeIdx)
		if bodyNode.Kind == ast.NodeBlock {
			sg.EmitBlockWithReturning(bodyNodeIdx, returning)
		} else {
			// It is a single expression, not a NodeBlock.
			// Let's emit it directly!
			expr := sg.ExprGen.Emit(bodyNodeIdx)
			if returning && sg.ExprGen.ReturnType != types.TypeVoid && sg.ExprGen.ReturnType != types.TypeUnknown && !sg.IsVoidExpr(bodyNodeIdx) {
				mg.W.Linef("return %s;", expr)
			} else {
				mg.W.Linef("%s;", expr)
			}
		}
		mg.W.Dedent()
		mg.W.Dedent()
	}
}

// EmitMatchExpr emits a match used as an expression.
// It creates a temporary variable, assigns the result of each arm to it,
// and returns the variable name.
func (mg *MatchGen) EmitMatchExpr(
	matchNodeIdx uint32,
	resultTypeID types.TypeID,
	tempVarName string,
) string {
	resultC := CTypeName(resultTypeID, mg.Table, mg.Intern, mg.Queue)
	mg.W.Linef("%s %s;", resultC, tempVarName)

	node := mg.Tree.Node(matchNodeIdx)
	_ = node
	children := mg.Tree.Children(matchNodeIdx)
	if len(children) < 2 {
		return tempVarName
	}

	discrimExpr := mg.ExprGen.Emit(children[0])
	mg.W.Linef("switch ((%s).tag) {", discrimExpr)

	for i := 1; i < len(children); i++ {
		armNode := mg.Tree.Node(children[i])
		if armNode.Kind != ast.NodeMatchArm {
			continue
		}
		armChildren := mg.Tree.Children(children[i])
		if len(armChildren) >= 2 {
			// Pattern
			patText := mg.ExprGen.Emit(armChildren[0])
			mg.W.Linef("    case %s:", patText)
			// Body — last expression is the result value
			bodyExpr := mg.ExprGen.Emit(armChildren[1])
			mg.W.Linef("        %s = %s;", tempVarName, bodyExpr)
			mg.W.Line("        break;")
		}
	}

	mg.W.Line("    default: __builtin_unreachable();")
	mg.W.Line("}")

	return tempVarName
}
