package sema

import (
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// OwnershipChecker enforces AXIOM's single-ownership rules at compile time.
// Each value has exactly one owner. Moving a value invalidates the source.
// Borrowed values cannot be stored or moved. Mutable access is controlled.
type OwnershipChecker struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable

	// graph is the ConnectionGraph being populated during analysis.
	graph *ConnectionGraph
	// moved tracks which symbols have been moved (invalidated).
	moved map[uint32]bool
	// scopeDepth tracks the current nesting depth for lifetime assignment.
	scopeDepth uint32
	// currentFunc tracks the current function symbol for return analysis.
	currentFunc uint32

	errors []diagnostics.Diagnostic

	// FunctionGraphs maps function symID to its computed ConnectionGraph
	FunctionGraphs map[uint32]*ConnectionGraph
	// FunctionMoved maps function symID to its computed moved set
	FunctionMoved map[uint32]map[uint32]bool
}

// NewOwnershipChecker creates a new OwnershipChecker.
func NewOwnershipChecker(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable) *OwnershipChecker {
	return &OwnershipChecker{
		ast:            tree,
		intern:         intern,
		symtable:       st,
		types:          tt,
		graph:          NewConnectionGraph(),
		moved:          make(map[uint32]bool),
		FunctionGraphs: make(map[uint32]*ConnectionGraph),
		FunctionMoved:  make(map[uint32]map[uint32]bool),
	}
}

// Graph returns the populated ConnectionGraph.
func (oc *OwnershipChecker) Graph() *ConnectionGraph {
	return oc.graph
}

// Moved returns the set of moved symbol IDs.
func (oc *OwnershipChecker) Moved() map[uint32]bool {
	return oc.moved
}

func (oc *OwnershipChecker) errorf(nodeIdx uint32, code int, format string, args ...any) {
	oc.errors = append(oc.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      diagnostics.Pos{},
	})
}

// Check runs ownership analysis on the entire AST.
// Returns diagnostics for ownership violations.
func (oc *OwnershipChecker) Check() []diagnostics.Diagnostic {
	if oc.ast == nil || oc.ast.NodeCount() == 0 {
		return oc.errors
	}
	oc.checkNode(0)
	return oc.errors
}

func (oc *OwnershipChecker) checkNode(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]

	switch node.Kind {
	case ast.NodeFuncDecl:
		oc.checkFuncDecl(nodeIdx)
	case ast.NodeVarDecl:
		oc.checkVarDecl(nodeIdx)
	case ast.NodeAssignStmt:
		oc.checkAssign(nodeIdx)
	case ast.NodeReturnStmt:
		oc.checkReturn(nodeIdx)
	case ast.NodeCallExpr:
		oc.checkCall(nodeIdx)
	case ast.NodeIdent:
		oc.checkIdentUse(nodeIdx)
	case ast.NodeBlock:
		oc.scopeDepth++
		oc.checkChildren(nodeIdx)
		oc.scopeDepth--
		return // already handled children
	default:
		oc.checkChildren(nodeIdx)
		return
	}
}

func (oc *OwnershipChecker) checkChildren(nodeIdx uint32) {
	child := oc.ast.Nodes[nodeIdx].FirstChild
	for child != 0 {
		oc.checkNode(child)
		child = oc.ast.Nodes[child].NextSibling
	}
}

func (oc *OwnershipChecker) checkFuncDecl(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]
	prevFunc := oc.currentFunc
	oc.currentFunc = node.Payload

	// Reset graph and moved set for this function
	prevGraph := oc.graph
	prevMoved := oc.moved
	oc.graph = NewConnectionGraph()
	oc.moved = make(map[uint32]bool)

	// Add a special RETURN_SLOT node for escape analysis
	oc.graph.AddValueNode(0, 0, 0) // Node 0 = return slot

	// Register parameters as value nodes
	child := node.FirstChild
	for child != 0 {
		childNode := &oc.ast.Nodes[child]
		if childNode.Kind == ast.NodeParamDecl {
			symID := childNode.Payload
			if symID != 0 {
				oc.graph.AddValueNode(symID, 0, oc.scopeDepth)
			}
		}
		child = childNode.NextSibling
	}

	oc.checkChildren(nodeIdx)

	// Save computed graph and moved set before restoring previous ones
	funcSym := node.Payload
	if funcSym != 0 {
		oc.FunctionGraphs[funcSym] = oc.graph
		// Deep copy the moved map to ensure it isn't mutated later
		movedCopy := make(map[uint32]bool)
		for k, v := range oc.moved {
			movedCopy[k] = v
		}
		oc.FunctionMoved[funcSym] = movedCopy
	}

	oc.currentFunc = prevFunc
	oc.graph = prevGraph
	oc.moved = prevMoved
}

func (oc *OwnershipChecker) checkVarDecl(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]
	symID := node.Payload

	if symID != 0 {
		// Add a value node for this variable
		oc.graph.AddValueNode(symID, 0, oc.scopeDepth)
	}

	// Check initializer — resolve children for type exprs and init expr
	child := node.FirstChild
	for child != 0 {
		childNode := &oc.ast.Nodes[child]
		if childNode.Kind == ast.NodeIdent {
			// Initializer uses a value — this is a move if the source type is non-Copy
			srcSym := childNode.Payload
			if srcSym != 0 && oc.isMoveContext(child) {
				if oc.moved[srcSym] {
					name := oc.symName(srcSym)
					oc.errorf(child, 4001, "use of moved value '%s'", name)
				}
				// Add FlowsTo edge: source flows to new variable
				if srcNode, ok := oc.graph.NodeOfSym(srcSym); ok {
					if dstNode, ok2 := oc.graph.NodeOfSym(symID); ok2 {
						oc.graph.AddEdge(srcNode, dstNode, EdgeFlowsTo)
					}
				}
				oc.moved[srcSym] = true
			}
		}
		oc.checkNode(child)
		child = childNode.NextSibling
	}
}

func (oc *OwnershipChecker) checkAssign(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]

	// AssignStmt children: [lhs, rhs]
	lhsIdx := node.FirstChild
	if lhsIdx == 0 {
		return
	}
	rhsIdx := oc.ast.Nodes[lhsIdx].NextSibling

	// Check mutability of LHS
	lhsNode := &oc.ast.Nodes[lhsIdx]
	if lhsNode.Kind == ast.NodeIdent {
		symIdx := lhsNode.Payload
		if symIdx != 0 && int(symIdx) < len(oc.symtable.Symbols) {
			sym := oc.symtable.SymbolAt(symIdx)
			// Check both symbol-level mut flag AND AST-level FlagIsMut.
			// The name resolver may not propagate FlagIsMut to SymFlagMut,
			// so we also check the declaring VarDecl's AST flags.
			isMut := sym.Flags&SymFlagMut != 0
			if !isMut {
				// Search for the VarDecl node that declared this symbol
				// and check its AST flags
				isMut = oc.isDeclaredMut(symIdx)
			}
			if sym.Kind == SymVar && !isMut {
				name := oc.symName(symIdx)
				oc.errorf(nodeIdx, 4002, "cannot assign to immutable variable '%s'", name)
			}
		}
	}

	// Check RHS for moved values
	if rhsIdx != 0 {
		oc.checkNode(rhsIdx)
	}
}

func (oc *OwnershipChecker) checkReturn(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]

	// If there's a return value, add EscapesTo edge to return slot
	child := node.FirstChild
	if child != 0 {
		childNode := &oc.ast.Nodes[child]
		if childNode.Kind == ast.NodeIdent {
			symID := childNode.Payload
			if symID != 0 {
				if int(symID) < len(oc.symtable.Symbols) {
					sym := oc.symtable.SymbolAt(symID)
					typeID := types.TypeID(sym.TypeID)
					if !typeID.IsPrimitive() {
						if srcNode, ok := oc.graph.NodeOfSym(symID); ok {
							// Return slot is node 0 (added in checkFuncDecl)
							oc.graph.AddEdge(srcNode, 0, EdgeEscapesTo)
						}
					}
				}
			}
		}
		oc.checkNode(child)
	}
}

func (oc *OwnershipChecker) checkCall(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]
	// CallExpr children: [callee, arg1, arg2, ...]
	calleeIdx := node.FirstChild
	if calleeIdx == 0 {
		return
	}

	// Process arguments
	argIdx := oc.ast.Nodes[calleeIdx].NextSibling
	for argIdx != 0 {
		argNode := &oc.ast.Nodes[argIdx]
		if argNode.Kind == ast.NodeIdent {
			symID := argNode.Payload
			if symID != 0 && oc.moved[symID] {
				name := oc.symName(symID)
				oc.errorf(argIdx, 4001, "use of moved value '%s'", name)
			}
			// Note: in full implementation, check if param is !T (sink) or lent T
			// For now, function calls don't consume by default
		}
		oc.checkNode(argIdx)
		argIdx = oc.ast.Nodes[argIdx].NextSibling
	}

	oc.checkNode(calleeIdx)
}

func (oc *OwnershipChecker) checkIdentUse(nodeIdx uint32) {
	node := &oc.ast.Nodes[nodeIdx]
	symID := node.Payload
	if symID != 0 && oc.moved[symID] {
		name := oc.symName(symID)
		oc.errorf(nodeIdx, 4001, "use of moved value '%s'", name)
	}
}

// isMoveContext returns true if the given node is in a position that constitutes a move.
// For MVP, all value bindings are moves (non-Copy types).
func (oc *OwnershipChecker) isMoveContext(nodeIdx uint32) bool {
	// For MVP, treat all value assignments as moves.
	// A full implementation would check if the type is Copy (primitives are Copy).
	node := &oc.ast.Nodes[nodeIdx]
	if node.Kind == ast.NodeIdent {
		symID := node.Payload
		if symID != 0 && int(symID) < len(oc.symtable.Symbols) {
			sym := oc.symtable.SymbolAt(symID)
			// Primitive types, built-in types, and constants are not moved
			if sym.Kind == SymBuiltinType || sym.Kind == SymConst || sym.Kind == SymFunc {
				return false
			}
			// Variables with struct types are moved; primitives are copied
			// For MVP, we only move non-primitive variables
			if sym.Kind == SymVar || sym.Kind == SymParam {
				return true
			}
		}
	}
	return false
}

// symName returns the human-readable name of a symbol.
func (oc *OwnershipChecker) symName(symIdx uint32) string {
	if symIdx == 0 || int(symIdx) >= len(oc.symtable.Symbols) {
		return "<unknown>"
	}
	sym := oc.symtable.SymbolAt(symIdx)
	name := oc.intern.Get(sym.NameID)
	if len(name) > 0 {
		return string(name)
	}
	return fmt.Sprintf("sym%d", symIdx)
}

// isDeclaredMut checks whether the VarDecl node for the given symbol
// has FlagIsMut set in the AST. This is needed because the name resolver
// may not propagate mut flags to the symbol table.
func (oc *OwnershipChecker) isDeclaredMut(symIdx uint32) bool {
	for i := range oc.ast.Nodes {
		node := &oc.ast.Nodes[i]
		if node.Kind == ast.NodeVarDecl && node.Payload == symIdx {
			return node.Flags&ast.FlagIsMut != 0
		}
	}
	return false
}
