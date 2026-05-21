package opt

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t02: Constant Folding Pass
//
// Evaluates constant expressions at compile time and replaces them with
// their computed values. This reduces runtime computation and enables
// further optimizations (e.g., DCE of now-unused operands).
// --------------------------------------------------------------------------

// ConstantFoldingPass implements OptPass for constant folding.
type ConstantFoldingPass struct{}

func (p *ConstantFoldingPass) Name() string { return "constant-fold" }

// Run scans all instructions in all functions, folding constant expressions.
// Returns true if any instruction was modified.
func (p *ConstantFoldingPass) Run(mod *air.AirModule) bool {
	changed := false
	for fi := range mod.Funcs {
		if foldFunc(&mod.Funcs[fi]) {
			changed = true
		}
	}
	return changed
}

// constVal represents a known constant value for an SSA register.
type constVal struct {
	known bool
	val   uint32 // raw value (integer or float bits)
}

// foldFunc performs constant folding on a single function.
func foldFunc(fn *air.AirFunc) bool {
	// Build a map of register → known constant value
	vals := make(map[uint32]constVal, len(fn.Insts))
	changed := false

	for i := range fn.Insts {
		inst := &fn.Insts[i]

		// Record constants
		if inst.Opcode == air.OpIConst || inst.Opcode == air.OpFConst {
			if inst.Dest != 0 {
				vals[inst.Dest] = constVal{known: true, val: inst.Src1}
			}
			continue
		}

		// Skip non-ALU instructions
		if !inst.Opcode.IsBinaryALU() && !isUnaryFoldable(inst.Opcode) {
			continue
		}

		// Try to fold binary operations
		if inst.Opcode.IsBinaryALU() {
			src1, ok1 := vals[inst.Src1]
			src2, ok2 := vals[inst.Src2]
			if ok1 && src1.known && ok2 && src2.known {
				result, ok := evalBinary(inst.Opcode, src1.val, src2.val)
				if ok {
					// Replace with constant
					inst.Opcode = air.OpIConst
					inst.Src1 = result
					inst.Src2 = 0
					vals[inst.Dest] = constVal{known: true, val: result}
					changed = true
					continue
				}
			}
		}

		// Try to fold unary operations
		if isUnaryFoldable(inst.Opcode) {
			src1, ok1 := vals[inst.Src1]
			if ok1 && src1.known {
				result, ok := evalUnary(inst.Opcode, src1.val)
				if ok {
					inst.Opcode = air.OpIConst
					inst.Src1 = result
					inst.Src2 = 0
					vals[inst.Dest] = constVal{known: true, val: result}
					changed = true
					continue
				}
			}
		}

		// Result is not a known constant
		if inst.Dest != 0 {
			vals[inst.Dest] = constVal{known: false}
		}
	}

	return changed
}

// evalBinary evaluates a binary operation on two constant values.
// Returns (result, true) if the operation can be folded, (0, false) otherwise.
func evalBinary(op air.Opcode, a, b uint32) (uint32, bool) {
	ai := int32(a)
	bi := int32(b)

	switch op {
	case air.OpIAdd:
		return uint32(ai + bi), true
	case air.OpISub:
		return uint32(ai - bi), true
	case air.OpIMul:
		return uint32(ai * bi), true
	case air.OpIDiv:
		if bi == 0 {
			return 0, false // division by zero — do not fold
		}
		return uint32(ai / bi), true
	case air.OpIMod:
		if bi == 0 {
			return 0, false
		}
		return uint32(ai % bi), true
	case air.OpEq:
		return boolToU32(ai == bi), true
	case air.OpNe:
		return boolToU32(ai != bi), true
	case air.OpLt:
		return boolToU32(ai < bi), true
	case air.OpLe:
		return boolToU32(ai <= bi), true
	case air.OpGt:
		return boolToU32(ai > bi), true
	case air.OpGe:
		return boolToU32(ai >= bi), true
	case air.OpAnd:
		return a & b, true
	case air.OpOr:
		return a | b, true
	case air.OpXor:
		return a ^ b, true
	case air.OpShl:
		if b >= 32 {
			return 0, true
		}
		return a << b, true
	case air.OpShr:
		if b >= 32 {
			return 0, true
		}
		return uint32(ai >> b), true
	default:
		return 0, false
	}
}

// evalUnary evaluates a unary operation on a constant value.
func evalUnary(op air.Opcode, a uint32) (uint32, bool) {
	switch op {
	case air.OpNeg:
		return uint32(-int32(a)), true
	case air.OpNot:
		if a == 0 {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// isUnaryFoldable returns true if the opcode is a unary operation
// that can be constant-folded.
func isUnaryFoldable(op air.Opcode) bool {
	return op == air.OpNeg || op == air.OpNot
}

// boolToU32 converts a boolean to 0/1 uint32.
func boolToU32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
