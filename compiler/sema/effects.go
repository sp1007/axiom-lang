package sema

import (
	"fmt"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// EffectSet represents the side effects a function can have.
type EffectSet struct {
	Raises  []types.TypeID
	IsPure  bool
	IsAsync bool
}

// HasRaise returns true if the EffectSet contains the given exception type.
func (es *EffectSet) HasRaise(typ types.TypeID) bool {
	for _, r := range es.Raises {
		if r == typ {
			return true
		}
	}
	return false
}

// EffectChecker verifies effect propagation rules (pure, async, raises).
type EffectChecker struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable
	infer    *InferenceEngine

	FuncEffects map[uint32]EffectSet
	errors      []diagnostics.Diagnostic
}

// NewEffectChecker creates a new EffectChecker.
func NewEffectChecker(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable, ie *InferenceEngine) *EffectChecker {
	return &EffectChecker{
		ast:         tree,
		intern:      intern,
		symtable:    st,
		types:       tt,
		infer:       ie,
		FuncEffects: make(map[uint32]EffectSet),
	}
}

func (ec *EffectChecker) errorf(nodeIdx uint32, code int, format string, args ...any) {
	ec.errors = append(ec.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      diagnostics.Pos{},
	})
}

// Check walks the AST to verify effects.
func (ec *EffectChecker) Check() []diagnostics.Diagnostic {
	if ec.ast == nil || ec.ast.NodeCount() == 0 {
		return ec.errors
	}
	// First pass: populate declared FuncEffects
	ec.populateEffects(0)

	// Second pass: verify body propagates effects correctly
	ec.verifyEffects(0, 0)
	
	return ec.errors
}

// populateEffects collects declared effects from FuncDecl nodes.
func (ec *EffectChecker) populateEffects(nodeIdx uint32) {
	if nodeIdx == 0 && ec.ast.Nodes[0].Kind != ast.NodeProgram {
		return
	}

	node := &ec.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeFuncDecl {
		symIdx := node.Payload
		if symIdx != 0 {
			if _, exists := ec.FuncEffects[symIdx]; !exists {
				isAsync := node.Flags&uint16(ast.FlagIsAsync) != 0
				ec.FuncEffects[symIdx] = EffectSet{
					IsAsync: isAsync,
				}
			}
		}
	}

	child := node.FirstChild
	for child != 0 {
		ec.populateEffects(child)
		child = ec.ast.Nodes[child].NextSibling
	}
}

// verifyEffects walks function bodies and checks calls against the enclosing function's effects.
func (ec *EffectChecker) verifyEffects(nodeIdx uint32, currentFuncSym uint32) {
	if nodeIdx == 0 && ec.ast.Nodes[0].Kind != ast.NodeProgram {
		return
	}

	node := &ec.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeFuncDecl {
		currentFuncSym = node.Payload
	}

	if node.Kind == ast.NodeCallExpr && currentFuncSym != 0 {
		callee := node.FirstChild
		if callee != 0 {
			calleeNode := &ec.ast.Nodes[callee]
			if calleeNode.Kind == ast.NodeIdent {
				calleeSymIdx := calleeNode.Payload
				if calleeSymIdx != 0 {
					calleeEffects := ec.FuncEffects[calleeSymIdx]
					callerEffects := ec.FuncEffects[currentFuncSym]

					// Check pure
					if callerEffects.IsPure && !calleeEffects.IsPure {
						ec.errorf(nodeIdx, 3040, "pure function cannot call impure function")
					}

					// Check raises
					// In a full implementation, we'd check if we are inside a try/match block.
					// For MVP, we just check if the caller declares it.
					for _, r := range calleeEffects.Raises {
						if !callerEffects.HasRaise(r) {
							// Check if there is an enclosing try/match handling it (mocked as false here)
							handled := false
							// Walk up AST to find try/match (omitted for MVP unless we add parent pointers or track it during descent)
							if !handled {
								ec.errorf(nodeIdx, 3041, "unhandled effect: function raises %d but it is not handled or declared", r)
							}
						}
					}
				}
			}
		}
	}

	if node.Kind == ast.NodeAwaitExpr {
		if currentFuncSym != 0 {
			callerEffects := ec.FuncEffects[currentFuncSym]
			if !callerEffects.IsAsync {
				ec.errorf(nodeIdx, 3042, "cannot await in non-async function")
			}
		} else {
			// Await outside function (e.g., top-level) - usually allowed if script, but let's say no
			ec.errorf(nodeIdx, 3042, "cannot await outside of async function")
		}
	}

	child := node.FirstChild
	for child != 0 {
		ec.verifyEffects(child, currentFuncSym)
		child = ec.ast.Nodes[child].NextSibling
	}
}
