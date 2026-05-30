package sema

import (
	"fmt"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// ResolvedCall contains the result of overload resolution.
type ResolvedCall struct {
	SymbolID uint32
	Score    int
	IsBuiltin bool
	BuiltinOp string // E.g. "add_i32"
}

// OverloadResolver handles resolving function and method calls.
type OverloadResolver struct {
	st *SymbolTable
	tt *types.TypeTable
}

// NewOverloadResolver creates a new overload resolver.
func NewOverloadResolver(st *SymbolTable, tt *types.TypeTable) *OverloadResolver {
	return &OverloadResolver{
		st: st,
		tt: tt,
	}
}

// Resolve searches for the best matching overload for the given function name.
func (or *OverloadResolver) Resolve(nameID uint32, argTypes []types.TypeID) (ResolvedCall, *diagnostics.Diagnostic) {
	candidates := or.collectCandidates(nameID)
	
	if len(candidates) == 0 {
		return ResolvedCall{}, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     3030,
			Message:  "no matching overload found",
		}
	}

	bestScore := -1
	var bestCandidates []uint32

	for _, candID := range candidates {
		sym := or.st.SymbolAt(candID)
		if sym.Kind != SymFunc {
			continue // only functions can be overloaded
		}
		
		score := or.scoreCandidate(sym, argTypes)
		if score < 0 {
			continue // -1 means no match
		}
		
		if score > bestScore {
			bestScore = score
			bestCandidates = []uint32{candID}
		} else if score == bestScore {
			bestCandidates = append(bestCandidates, candID)
		}
	}

	if len(bestCandidates) == 0 {
		return ResolvedCall{}, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     3030,
			Message:  "no matching overload found for the given argument types",
		}
	}
	
	if len(bestCandidates) > 1 {
		return ResolvedCall{}, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     3031,
			Message:  fmt.Sprintf("ambiguous call: %d overloads have the same score", len(bestCandidates)),
		}
	}

	return ResolvedCall{
		SymbolID: bestCandidates[0],
		Score:    bestScore,
	}, nil
}

// collectCandidates searches all active scopes from innermost to outermost.
func (or *OverloadResolver) collectCandidates(nameID uint32) []uint32 {
	var candidates []uint32
	// Search from top of stack (innermost) to bottom (global)
	for i := len(or.st.stack) - 1; i >= 0; i-- {
		scopeIdx := or.st.stack[i]
		if symIdx, found := or.st.Scopes[scopeIdx].get(nameID); found {
			currIdx := symIdx
			for {
				candidates = append(candidates, currIdx)
				sym := or.st.SymbolAt(currIdx)
				if sym.NextOverload == 0 {
					break
				}
				currIdx = sym.NextOverload
			}
		}
	}
	return candidates
}

// scoreCandidate computes a score for a single candidate based on argument types.
// Returns -1 if the candidate is not a match.
func (or *OverloadResolver) scoreCandidate(sym *Symbol, argTypes []types.TypeID) int {
	funcTypeID := types.TypeID(sym.TypeID)
	if funcTypeID == 0 || funcTypeID == types.TypeUnknown {
		return -1
	}

	entry := or.tt.Entry(funcTypeID)
	if entry.Kind != types.KindFunction {
		return -1
	}
	
	fInfo := or.tt.FuncInfo(funcTypeID)
	if len(fInfo.Params) != len(argTypes) {
		return -1 // Arg count mismatch
	}

	totalScore := 0
	for i, paramType := range fInfo.Params {
		argType := argTypes[i]
		
		if paramType == argType {
			totalScore += 4
		} else if or.tt.IsAssignableTo(argType, paramType) {
			totalScore += 3
		} else if or.isCoercible(argType, paramType) {
			totalScore += 2
		} else {
			// No match
			return -1
		}
	}
	
	return totalScore
}

// isCoercible returns true if 'from' type can be implicitly coerced to 'to' type.
func (or *OverloadResolver) isCoercible(from, to types.TypeID) bool {
	// Integers coerce to larger integers
	if from.IsInteger() && to.IsInteger() && from <= to {
		return true
	}
	// Integer to float
	if from.IsInteger() && to.IsFloat() {
		return true
	}
	// Float to larger float
	if from.IsFloat() && to.IsFloat() && from <= to {
		return true
	}
	return false
}

// ResolveBuiltinOp resolves built-in operator pseudo-overloads (e.g. '+').
func (or *OverloadResolver) ResolveBuiltinOp(op string, left, right types.TypeID) types.TypeID {
	if left == types.TypeUnknown || right == types.TypeUnknown {
		return types.TypeUnknown
	}
	
	switch op {
	case "+":
		if left == types.TypeString && right == types.TypeString {
			return types.TypeString
		}
		fallthrough
	case "-", "*", "/", "%":
		// Numeric rules: wider of the two.
		// For simplicity, we assume they are either same or coercible.
		if or.isCoercible(left, right) {
			return right
		} else if or.isCoercible(right, left) {
			return left
		} else if left.IsFloat() || right.IsFloat() { // float widening
			if left == types.TypeF64 || right == types.TypeF64 {
				return types.TypeF64
			}
			return types.TypeF32
		} else if left.IsInteger() && right.IsInteger() {
			if left > right {
				return left
			}
			return right
		}
	case "==", "!=", "<", ">", "<=", ">=":
		// Assuming comparable types
		return types.TypeBool
	case "and", "or":
		if left == types.TypeBool && right == types.TypeBool {
			return types.TypeBool
		}
	case "&", "|", "^", "<<", ">>":
		if left.IsInteger() && right.IsInteger() {
			return left
		}
	}
	return types.TypeUnknown
}
