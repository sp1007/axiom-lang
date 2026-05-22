package cgen

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// DeferStack tracks deferred expressions per scope.
// Each scope has its own defer slice; defers are executed LIFO at scope exit.
type DeferStack struct {
	scopes [][]uint32 // stack of scopes, each with AST node indices of deferred exprs
}

// NewDeferStack creates an empty DeferStack.
func NewDeferStack() *DeferStack {
	return &DeferStack{}
}

// PushScope enters a new defer scope.
func (d *DeferStack) PushScope() {
	d.scopes = append(d.scopes, nil)
}

// PopScope exits the current scope, returning deferred node indices in LIFO order.
func (d *DeferStack) PopScope() []uint32 {
	if len(d.scopes) == 0 {
		return nil
	}
	top := d.scopes[len(d.scopes)-1]
	d.scopes = d.scopes[:len(d.scopes)-1]
	// Reverse for LIFO order
	for i, j := 0, len(top)-1; i < j; i, j = i+1, j-1 {
		top[i], top[j] = top[j], top[i]
	}
	return top
}

// Push adds a deferred expression AST node to the current scope.
func (d *DeferStack) Push(exprNodeIdx uint32) {
	if len(d.scopes) == 0 {
		d.PushScope()
	}
	d.scopes[len(d.scopes)-1] = append(d.scopes[len(d.scopes)-1], exprNodeIdx)
}

// CurrentDefers returns the deferred expressions in the current scope (LIFO order).
// Does not pop the scope.
func (d *DeferStack) CurrentDefers() []uint32 {
	if len(d.scopes) == 0 {
		return nil
	}
	top := d.scopes[len(d.scopes)-1]
	result := make([]uint32, len(top))
	copy(result, top)
	// Reverse for LIFO
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// StmtGen generates C statements from the typed AST.
// It maintains a defer stack and delegates expression generation to ExprGen.
type StmtGen struct {
	W       *IndentWriter
	ExprGen *ExprGen
	Defers  *DeferStack
	Table   *types.TypeTable
	Intern  *ast.InternPool
	Symbols *sema.SymbolTable
	Tree    *ast.AstTree
	Queue   *TypeDeclQueue
}

// NewStmtGen creates a new statement generator.
func NewStmtGen(
	w *IndentWriter,
	table *types.TypeTable,
	intern *ast.InternPool,
	symbols *sema.SymbolTable,
	tree *ast.AstTree,
	queue *TypeDeclQueue,
) *StmtGen {
	sg := &StmtGen{
		W:       w,
		Defers:  NewDeferStack(),
		Table:   table,
		Intern:  intern,
		Symbols: symbols,
		Tree:    tree,
		Queue:   queue,
	}
	sg.ExprGen = NewExprGen(table, intern, symbols, tree, queue)
	return sg
}

// EmitStmt emits C code for a single AST statement node.
func (g *StmtGen) EmitStmt(nodeIdx uint32) {
	node := g.Tree.Node(nodeIdx)

	switch node.Kind {
	case ast.NodeVarDecl:
		g.emitVarDecl(nodeIdx, node)
	case ast.NodeAssignStmt:
		g.emitAssign(nodeIdx, node)
	case ast.NodeReturnStmt:
		g.emitReturn(nodeIdx, node)
	case ast.NodeIfStmt:
		g.emitIf(nodeIdx, node)
	case ast.NodeWhileStmt:
		g.emitWhile(nodeIdx, node)
	case ast.NodeForStmt:
		g.emitFor(nodeIdx, node)
	case ast.NodeDeferStmt:
		g.emitDefer(nodeIdx, node)
	case ast.NodeUnsafeBlock:
		g.emitUnsafe(nodeIdx, node)
	case ast.NodeArenaBlock:
		g.emitArena(nodeIdx, node)
	case ast.NodeBlock:
		g.emitBlock(nodeIdx, node)
	case ast.NodeDestroyStmt:
		g.emitDestroy(nodeIdx, node)
	case ast.NodeAliasStmt:
		g.emitAlias(nodeIdx, node)
	case ast.NodeMatchStmt:
		g.emitMatch(nodeIdx, node)
	default:
		// Expression statement (call, etc.)
		expr := g.ExprGen.Emit(nodeIdx)
		g.W.Linef("%s;", expr)
	}
}

// EmitBlock emits all statements in a block node.
func (g *StmtGen) EmitBlock(nodeIdx uint32) {
	node := g.Tree.Node(nodeIdx)
	child := node.FirstChild
	for child != ast.NullIdx {
		g.EmitStmt(child)
		child = g.Tree.Node(child).NextSibling
	}
}

// EmitFuncBody emits the body of a function, with defer scope management.
func (g *StmtGen) EmitFuncBody(bodyNodeIdx uint32) {
	g.Defers.PushScope()
	g.EmitBlock(bodyNodeIdx)
	// Emit remaining defers at function end (implicit return)
	deferred := g.Defers.PopScope()
	for _, d := range deferred {
		expr := g.ExprGen.Emit(d)
		g.W.Linef("%s;", expr)
	}
}

// emitVarDecl generates: type name = initializer;
func (g *StmtGen) emitVarDecl(idx uint32, node *ast.AstNode) {
	name := string(g.Tree.TokenText(node.TokenIdx))
	symIdx := node.Payload
	var typeID types.TypeID
	useSym := false
	if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
		sym := g.Symbols.SymbolAt(symIdx)
		if sym.Kind == sema.SymVar || sym.Kind == sema.SymParam || sym.Kind == sema.SymConst {
			useSym = true
			typeID = types.TypeID(sym.TypeID)
			name = resolveName(sym.NameID, g.Intern)
		}
	}
	if !useSym {
		typeID = types.TypeID(node.Payload)
	}
	// Find initExpr by scanning children and skipping NodeTypeExpr and NodeGenericType
	var initExprIdx uint32 = ast.NullIdx
	child := node.FirstChild
	for child != ast.NullIdx {
		childKind := g.Tree.Node(child).Kind
		if childKind != ast.NodeTypeExpr && childKind != ast.NodeGenericType {
			initExprIdx = child
			break
		}
		child = g.Tree.Node(child).NextSibling
	}

	fEntry := g.Table.Entry(typeID)
	if fEntry.Kind == types.KindArray {
		elemID := g.Table.ArrayElem(typeID)
		elemC := CTypeName(elemID, g.Table, g.Intern, g.Queue)
		length := g.Table.ArrayLength(typeID)
		if initExprIdx != ast.NullIdx {
			initExpr := g.ExprGen.Emit(initExprIdx)
			g.W.Linef("%s %s[%d] = %s;", elemC, name, length, initExpr)
		} else {
			g.W.Linef("%s %s[%d] = {0};", elemC, name, length)
		}
	} else {
		ctype := CTypeName(typeID, g.Table, g.Intern, g.Queue)
		if initExprIdx != ast.NullIdx {
			initExpr := g.ExprGen.Emit(initExprIdx)
			g.W.Linef("%s %s = %s;", ctype, name, initExpr)
		} else {
			g.W.Linef("%s %s = {0};", ctype, name)
		}
	}
}

// emitAssign generates: lhs = rhs;
func (g *StmtGen) emitAssign(idx uint32, node *ast.AstNode) {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		g.W.Line("/* invalid assign: missing children */")
		return
	}

	// Emit bounds checks for LHS first
	g.emitLHSBoundsChecks(children[0])

	// Emit LHS in unsafe (no bounds check) mode so it is a valid C lvalue
	wasUnsafe := g.ExprGen.Unsafe
	g.ExprGen.Unsafe = true
	lhs := g.ExprGen.Emit(children[0])
	g.ExprGen.Unsafe = wasUnsafe

	rhs := g.ExprGen.Emit(children[1])

	// Check for compound assignment operators via ExtraIdx
	op := assignOp(node.ExtraIdx)
	g.W.Linef("%s %s %s;", lhs, op, rhs)
}

func (g *StmtGen) emitLHSBoundsChecks(idx uint32) {
	node := g.Tree.Node(idx)
	if node.Kind == ast.NodeIndexExpr {
		children := g.Tree.Children(idx)
		if len(children) >= 2 {
			// First recursively emit bounds checks for children (nested indexes)
			g.emitLHSBoundsChecks(children[0])
			g.emitLHSBoundsChecks(children[1])

			// Emit bounds check for this index expression itself
			arr := g.ExprGen.Emit(children[0])
			index := g.ExprGen.Emit(children[1])

			colType := g.ExprGen.NodeType(children[0])
			if colType != types.TypeUnknown {
				entry := g.Table.Entry(colType)
				if entry.Kind == types.KindArray {
					length := g.Table.ArrayLength(colType)
					g.W.Linef("ax_bounds_check((ax_u64)(%s), (ax_u64)(%d));", index, length)
					return
				}
				if entry.Kind == types.KindPointer {
					// Pointers don't have bounds checks
					return
				}
			}
			g.W.Linef("ax_bounds_check((ax_u64)(%s), (%s).len);", index, arr)
		}
	} else {
		// Recursively walk other children to find any index expressions (e.g. self.field.arr[i])
		child := node.FirstChild
		for child != ast.NullIdx {
			g.emitLHSBoundsChecks(child)
			child = g.Tree.Node(child).NextSibling
		}
	}
}

// emitReturn generates deferred expressions then the return statement.
func (g *StmtGen) emitReturn(idx uint32, node *ast.AstNode) {
	// Emit current scope defers in LIFO order
	deferred := g.Defers.CurrentDefers()
	for _, d := range deferred {
		expr := g.ExprGen.Emit(d)
		g.W.Linef("%s;", expr)
	}

	if node.FirstChild != ast.NullIdx {
		retExpr := g.ExprGen.Emit(node.FirstChild)
		g.W.Linef("return %s;", retExpr)
	} else {
		g.W.Line("return;")
	}
}

// emitIf generates if/elif/else chains.
func (g *StmtGen) emitIf(idx uint32, node *ast.AstNode) {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		g.W.Line("/* invalid if: missing children */")
		return
	}

	cond := g.ExprGen.Emit(children[0])
	g.W.Linef("if (%s) {", cond)
	g.W.Indent()
	g.EmitBlock(children[1])
	g.W.Dedent()

	// Process elif/else siblings
	for i := 2; i < len(children); i++ {
		childNode := g.Tree.Node(children[i])
		switch childNode.Kind {
		case ast.NodeElifClause:
			elifChildren := g.Tree.Children(children[i])
			if len(elifChildren) >= 2 {
				elifCond := g.ExprGen.Emit(elifChildren[0])
				g.W.Linef("} else if (%s) {", elifCond)
				g.W.Indent()
				g.EmitBlock(elifChildren[1])
				g.W.Dedent()
			}
		case ast.NodeElseClause:
			g.W.Line("} else {")
			g.W.Indent()
			g.EmitBlock(children[i])
			g.W.Dedent()
		}
	}
	g.W.Line("}")
}

// emitWhile generates while loops.
func (g *StmtGen) emitWhile(idx uint32, node *ast.AstNode) {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		g.W.Line("/* invalid while: missing children */")
		return
	}

	cond := g.ExprGen.Emit(children[0])
	g.W.Linef("while (%s) {", cond)
	g.W.Indent()
	g.Defers.PushScope()
	g.EmitBlock(children[1])
	deferred := g.Defers.PopScope()
	for _, d := range deferred {
		expr := g.ExprGen.Emit(d)
		g.W.Linef("%s;", expr)
	}
	g.W.Dedent()
	g.W.Line("}")
}

// emitFor generates for-in loops over slices or ranges.
func (g *StmtGen) emitFor(idx uint32, node *ast.AstNode) {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		g.W.Line("/* invalid for: missing children */")
		return
	}

	symIdx := node.Payload
	varName := ""
	var elemTypeID types.TypeID = types.TypeI32

	if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
		sym := g.Symbols.SymbolAt(symIdx)
		varName = resolveName(sym.NameID, g.Intern)
		if sym.TypeID != 0 {
			elemTypeID = types.TypeID(sym.TypeID)
		}
	} else {
		// Fallback if resolver didn't run
		varName = resolveName(symIdx, g.Intern)
		if varName == "" {
			varName = "i" // default fallback
		}
	}

	elemType := CTypeName(elemTypeID, g.Table, g.Intern, g.Queue)

	// Check if iterator is a range expression: A..B
	iterNode := g.Tree.Node(children[0])
	isRange := false
	var startExpr, endExpr string
	if iterNode.Kind == ast.NodeBinaryExpr {
		opText := string(g.Tree.TokenText(iterNode.TokenIdx))
		if opText == ".." {
			isRange = true
			iterChildren := g.Tree.Children(children[0])
			if len(iterChildren) >= 2 {
				startExpr = g.ExprGen.Emit(iterChildren[0])
				endExpr = g.ExprGen.Emit(iterChildren[1])
			} else {
				startExpr = "0"
				endExpr = "0"
			}
		}
	}

	if isRange {
		g.W.Linef("for (%s %s = %s; %s < %s; %s++) {", elemType, varName, startExpr, varName, endExpr, varName)
		g.W.Indent()
		g.Defers.PushScope()
		if len(children) > 1 {
			g.EmitBlock(children[1])
		}
		deferred := g.Defers.PopScope()
		for _, d := range deferred {
			expr := g.ExprGen.Emit(d)
			g.W.Linef("%s;", expr)
		}
		g.W.Dedent()
		g.W.Line("}")
	} else {
		iterExpr := g.ExprGen.Emit(children[0])
		idxVar := fmt.Sprintf("_ax_i_%s", varName)

		g.W.Linef("for (ax_u64 %s = 0; %s < (%s).len; %s++) {", idxVar, idxVar, iterExpr, idxVar)
		g.W.Indent()
		g.W.Linef("%s %s = (%s).ptr[%s];", elemType, varName, iterExpr, idxVar)
		g.Defers.PushScope()
		if len(children) > 1 {
			g.EmitBlock(children[1])
		}
		deferred := g.Defers.PopScope()
		for _, d := range deferred {
			expr := g.ExprGen.Emit(d)
			g.W.Linef("%s;", expr)
		}
		g.W.Dedent()
		g.W.Line("}")
	}
}

// emitDefer pushes the deferred expression onto the defer stack.
func (g *StmtGen) emitDefer(idx uint32, node *ast.AstNode) {
	if node.FirstChild != ast.NullIdx {
		g.Defers.Push(node.FirstChild)
	}
}

// emitUnsafe generates an unsafe block.
func (g *StmtGen) emitUnsafe(idx uint32, node *ast.AstNode) {
	g.W.Line("{ /* unsafe */")
	g.W.Indent()
	g.EmitBlock(idx)
	g.W.Dedent()
	g.W.Line("}")
}

// emitArena generates an arena-scoped block.
func (g *StmtGen) emitArena(idx uint32, node *ast.AstNode) {
	g.W.Line("{ /* arena */")
	g.W.Indent()
	g.W.Line("ax_arena_scope _ax_arena = ax_arena_begin();")
	g.EmitBlock(idx)
	g.W.Line("ax_arena_end(&_ax_arena);")
	g.W.Dedent()
	g.W.Line("}")
}

// emitBlock generates a C block with braces.
func (g *StmtGen) emitBlock(idx uint32, node *ast.AstNode) {
	g.W.Line("{")
	g.W.Indent()
	g.EmitBlock(idx)
	g.W.Dedent()
	g.W.Line("}")
}

// emitDestroy generates CTGC-injected destroy (free) calls.
func (g *StmtGen) emitDestroy(idx uint32, node *ast.AstNode) {
	var typeID types.TypeID
	var name string

	if node.FirstChild != ast.NullIdx {
		expr := g.ExprGen.Emit(node.FirstChild)
		typeID = g.ExprGen.NodeType(node.FirstChild)
		name = expr
	} else {
		symID := node.Payload
		if symID != 0 && g.Symbols != nil && int(symID) < len(g.Symbols.Symbols) {
			sym := g.Symbols.SymbolAt(symID)
			typeID = types.TypeID(sym.TypeID)
			name = resolveName(sym.NameID, g.Intern)
		}
	}

	if typeID != types.TypeUnknown && g.Table != nil && int(typeID) < g.Table.Count() {
		entry := g.Table.Entry(typeID)
		if entry.Kind == types.KindPointer || entry.Kind == types.KindRef {
			g.W.Linef("ax_free(%s);", name)
		} else {
			g.W.Linef("/* skip destroy for value type %s */", name)
		}
	} else {
		g.W.Linef("ax_free(%s);", name)
	}
}

// emitAlias generates CTGC alias reuse.
func (g *StmtGen) emitAlias(idx uint32, node *ast.AstNode) {
	children := g.Tree.Children(idx)
	if len(children) >= 2 {
		dest := g.ExprGen.Emit(children[0])
		src := g.ExprGen.Emit(children[1])
		g.W.Linef("/* alias reuse */ %s = %s;", dest, src)
	}
}

// emitMatch generates a match/switch statement.
func (g *StmtGen) emitMatch(idx uint32, node *ast.AstNode) {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		g.W.Line("/* invalid match: missing children */")
		return
	}

	matchExpr := g.ExprGen.Emit(children[0])
	g.W.Linef("switch ((%s).tag) {", matchExpr)
	g.W.Indent()

	// Process match arms
	for i := 1; i < len(children); i++ {
		armNode := g.Tree.Node(children[i])
		if armNode.Kind == ast.NodeMatchArm {
			armChildren := g.Tree.Children(children[i])
			if len(armChildren) >= 2 {
				pattern := g.ExprGen.Emit(armChildren[0])
				g.W.Linef("case %s: {", pattern)
				g.W.Indent()
				g.EmitBlock(armChildren[1])
				g.W.Line("break;")
				g.W.Dedent()
				g.W.Line("}")
			}
		}
	}

	g.W.Dedent()
	g.W.Line("}")
}

// assignOp returns the C assignment operator string.
// ExtraIdx encodes the operator variant (0 = plain "=", etc.)
func assignOp(extraIdx uint32) string {
	switch extraIdx {
	case 0:
		return "="
	case 1:
		return "+="
	case 2:
		return "-="
	case 3:
		return "*="
	case 4:
		return "/="
	case 5:
		return "%="
	default:
		return "="
	}
}
