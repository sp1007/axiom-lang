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
	g.EmitStmtWithReturning(nodeIdx, false)
}

// EmitStmtWithReturning emits C code for a single AST statement node, propagating the returning state.
func (g *StmtGen) EmitStmtWithReturning(nodeIdx uint32, returning bool) {
	node := g.Tree.Node(nodeIdx)

	switch node.Kind {
	case ast.NodeVarDecl:
		g.emitVarDecl(nodeIdx, node)
	case ast.NodeAssignStmt:
		g.emitAssign(nodeIdx, node)
	case ast.NodeReturnStmt:
		g.emitReturn(nodeIdx, node)
	case ast.NodeIfStmt:
		g.emitIfWithReturning(nodeIdx, node, returning)
	case ast.NodeWhileStmt:
		g.emitWhile(nodeIdx, node)
	case ast.NodeForStmt:
		g.emitFor(nodeIdx, node)
	case ast.NodeDeferStmt:
		g.emitDefer(nodeIdx, node)
	case ast.NodeUnsafeBlock:
		g.emitUnsafeWithReturning(nodeIdx, node, returning)
	case ast.NodeArenaBlock:
		g.emitArenaWithReturning(nodeIdx, node, returning)
	case ast.NodeBlock:
		g.emitBlockWithReturning(nodeIdx, node, returning)
	case ast.NodeDestroyStmt:
		g.emitDestroy(nodeIdx, node)
	case ast.NodeAliasStmt:
		g.emitAlias(nodeIdx, node)
	case ast.NodeMatchStmt:
		g.emitMatchWithReturning(nodeIdx, node, returning)
	default:
		// Expression statement (call, etc.)
		expr := g.ExprGen.Emit(nodeIdx)
		if returning && g.ExprGen.ReturnType != types.TypeVoid && g.ExprGen.ReturnType != types.TypeUnknown && !g.IsVoidExpr(nodeIdx) {
			deferred := g.Defers.CurrentDefers()
			for _, d := range deferred {
				defExpr := g.ExprGen.Emit(d)
				g.W.Linef("%s;", defExpr)
			}
			g.W.Linef("return %s;", expr)
		} else {
			g.W.Linef("%s;", expr)
		}
	}
}

// IsVoidExpr returns true if the AST node is a void expression or a void builtin call.
func (g *StmtGen) IsVoidExpr(nodeIdx uint32) bool {
	if nodeIdx == ast.NullIdx {
		return true
	}
	node := g.Tree.Node(nodeIdx)
	if node.Kind == ast.NodeCallExpr {
		children := g.Tree.Children(nodeIdx)
		if len(children) >= 1 {
			callee := children[0]
			calleeNode := g.Tree.Node(callee)
			if calleeNode.Kind == ast.NodeIdent {
				name := string(g.Tree.TokenText(calleeNode.TokenIdx))
				switch name {
				case "panic", "assert", "assert_eq", "print", "println", "eprint", "eprintln", "exit", "abort":
					return true
				}
			}
		}
	}
	tID := g.ExprGen.NodeType(nodeIdx)
	return tID == types.TypeVoid
}

// EmitBlock emits all statements in a block node.
func (g *StmtGen) EmitBlock(nodeIdx uint32) {
	g.EmitBlockWithReturning(nodeIdx, false)
}

// EmitBlockWithReturning emits all statements in a block node, propagating the returning state to the last statement.
func (g *StmtGen) EmitBlockWithReturning(nodeIdx uint32, returning bool) {
	node := g.Tree.Node(nodeIdx)
	child := node.FirstChild
	for child != ast.NullIdx {
		next := g.Tree.Node(child).NextSibling
		isLast := next == ast.NullIdx
		g.EmitStmtWithReturning(child, returning && isLast)
		child = next
	}
}

// EmitFuncBody emits the body of a function, with defer scope management.
func (g *StmtGen) EmitFuncBody(bodyNodeIdx uint32) {
	g.Defers.PushScope()
	hasReturnVal := g.ExprGen.ReturnType != types.TypeVoid && g.ExprGen.ReturnType != types.TypeUnknown
	g.EmitBlockWithReturning(bodyNodeIdx, hasReturnVal)
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
			oldExpected := g.ExprGen.ExpectedType
			g.ExprGen.ExpectedType = typeID
			initExpr := g.ExprGen.Emit(initExprIdx)
			g.ExprGen.ExpectedType = oldExpected
			g.W.Linef("%s %s[%d] = %s;", elemC, name, length, initExpr)
		} else {
			g.W.Linef("%s %s[%d] = {0};", elemC, name, length)
		}
	} else {
		ctype := CTypeName(typeID, g.Table, g.Intern, g.Queue)
		if initExprIdx != ast.NullIdx {
			oldExpected := g.ExprGen.ExpectedType
			g.ExprGen.ExpectedType = typeID
			initExpr := g.ExprGen.Emit(initExprIdx)
			g.ExprGen.ExpectedType = oldExpected
			g.W.Linef("%s %s = %s;", ctype, name, initExpr)
		} else {
			g.W.Linef("%s %s = {0};", ctype, name)
		}
		if name == "dummy" {
			g.W.Linef("(void)%s;", name)
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

	lhsType := g.ExprGen.NodeType(children[0])
	oldExpected := g.ExprGen.ExpectedType
	g.ExprGen.ExpectedType = lhsType
	rhs := g.ExprGen.Emit(children[1])
	g.ExprGen.ExpectedType = oldExpected

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
		oldExpected := g.ExprGen.ExpectedType
		g.ExprGen.ExpectedType = g.ExprGen.ReturnType
		retExpr := g.ExprGen.Emit(node.FirstChild)
		g.ExprGen.ExpectedType = oldExpected
		g.W.Linef("return %s;", retExpr)
	} else {
		g.W.Line("return;")
	}
}

// emitIf generates if/elif/else chains.
func (g *StmtGen) emitIf(idx uint32, node *ast.AstNode) {
	g.emitIfWithReturning(idx, node, false)
}

func (g *StmtGen) emitIfWithReturning(idx uint32, node *ast.AstNode, returning bool) {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		g.W.Line("/* invalid if: missing children */")
		return
	}

	cond := g.ExprGen.Emit(children[0])
	g.W.Linef("if (%s) {", cond)
	g.W.Indent()
	g.EmitBlockWithReturning(children[1], returning)
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
				g.EmitBlockWithReturning(elifChildren[1], returning)
				g.W.Dedent()
			}
		case ast.NodeElseClause:
			g.W.Line("} else {")
			g.W.Indent()
			g.EmitBlockWithReturning(children[i], returning)
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
	g.emitUnsafeWithReturning(idx, node, false)
}

func (g *StmtGen) emitUnsafeWithReturning(idx uint32, node *ast.AstNode, returning bool) {
	g.W.Line("{ /* unsafe */")
	g.W.Indent()
	g.EmitBlockWithReturning(idx, returning)
	g.W.Dedent()
	g.W.Line("}")
}

// emitArena generates an arena-scoped block.
func (g *StmtGen) emitArena(idx uint32, node *ast.AstNode) {
	g.emitArenaWithReturning(idx, node, false)
}

func (g *StmtGen) emitArenaWithReturning(idx uint32, node *ast.AstNode, returning bool) {
	g.W.Line("{ /* arena */")
	g.W.Indent()
	g.W.Line("ax_arena_scope _ax_arena = ax_arena_begin();")
	g.EmitBlockWithReturning(idx, returning)
	g.W.Line("ax_arena_end(&_ax_arena);")
	g.W.Dedent()
	g.W.Line("}")
}

// emitBlock generates a C block with braces.
func (g *StmtGen) emitBlock(idx uint32, node *ast.AstNode) {
	g.emitBlockWithReturning(idx, node, false)
}

func (g *StmtGen) emitBlockWithReturning(idx uint32, node *ast.AstNode, returning bool) {
	g.W.Line("{")
	g.W.Indent()
	g.EmitBlockWithReturning(idx, returning)
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
	g.emitMatchWithReturning(idx, node, false)
}

func (g *StmtGen) emitMatchWithReturning(idx uint32, node *ast.AstNode, returning bool) {
	mg := NewMatchGen(g.W, g.ExprGen, g.Table, g.Intern, g.Tree, g.Queue)
	mg.EmitMatchStmtWithReturning(idx, returning)
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
