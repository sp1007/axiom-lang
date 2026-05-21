package x86

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p11-t03: x86-64 Instruction Selector
//
// Translates AIR instructions into MachInst (machine-level instructions
// with virtual registers). This is the first step of lowering from the
// platform-independent AIR to platform-specific x86-64 code.
// --------------------------------------------------------------------------

// MachOpKind identifies the type of machine instruction.
type MachOpKind uint16

const (
	MachNop MachOpKind = iota
	MachMov
	MachMovImm
	MachAdd
	MachSub
	MachImul
	MachIdiv
	MachCqo
	MachNeg
	MachNot
	MachAnd
	MachOr
	MachXor
	MachShl
	MachSar
	MachCmp
	MachTest
	MachSetCC
	MachMovzxB
	MachJmp
	MachJcc
	MachCall
	MachCallIndirect
	MachRet
	MachPush
	MachPop
	MachLea
	MachLoad
	MachStore
	MachXorZero
	MachLabel
)

// OperandKind describes whether an operand is a register, immediate, or memory.
type OperandKind uint8

const (
	OpndNone OperandKind = iota
	OpndVReg             // virtual register
	OpndPhys             // physical register (pre-allocated)
	OpndImm              // immediate value
	OpndLabel            // block label
	OpndMem              // memory [base + disp]
)

// MachOperand is a single operand for a MachInst.
type MachOperand struct {
	Kind  OperandKind
	VReg  uint32   // virtual register index (for OpndVReg)
	Phys  PhysReg  // physical register (for OpndPhys)
	Imm   int64    // immediate value (for OpndImm)
	Label uint32   // block label (for OpndLabel)
}

// MachInst is a single machine-level instruction with virtual registers.
type MachInst struct {
	Op   MachOpKind
	CC   CondCode    // condition code (for Jcc, SetCC)
	Dst  MachOperand // destination
	Src1 MachOperand // first source
	Src2 MachOperand // second source (if any)
}

// VReg creates a virtual register operand.
func VReg(id uint32) MachOperand {
	return MachOperand{Kind: OpndVReg, VReg: id}
}

// Phys creates a physical register operand.
func Phys(r PhysReg) MachOperand {
	return MachOperand{Kind: OpndPhys, Phys: r}
}

// Imm creates an immediate operand.
func Imm(v int64) MachOperand {
	return MachOperand{Kind: OpndImm, Imm: v}
}

// Label creates a block label operand.
func LabelOp(id uint32) MachOperand {
	return MachOperand{Kind: OpndLabel, Label: id}
}

// Select translates an AirFunc into a list of MachInst per basic block.
// Virtual registers map 1:1 to AIR registers.
func Select(fn *air.AirFunc) []MachInst {
	var result []MachInst

	// If blocks have instruction indices, emit per-block
	if len(fn.Blocks) > 0 && hasBlockInstrs(fn) {
		for bi := range fn.Blocks {
			blk := &fn.Blocks[bi]
			result = append(result, MachInst{Op: MachLabel, Dst: LabelOp(blk.ID)})
			for _, idx := range blk.Instrs {
				if int(idx) < len(fn.Insts) {
					result = append(result, selectInst(&fn.Insts[idx])...)
				}
			}
		}
	} else {
		// Flat instruction list
		for i := range fn.Insts {
			result = append(result, selectInst(&fn.Insts[i])...)
		}
	}

	return result
}

func hasBlockInstrs(fn *air.AirFunc) bool {
	for bi := range fn.Blocks {
		if len(fn.Blocks[bi].Instrs) > 0 {
			return true
		}
	}
	return false
}

// selectInst translates a single AIR instruction into one or more MachInsts.
func selectInst(inst *air.AirInst) []MachInst {
	switch inst.Opcode {
	case air.OpNop:
		return nil

	case air.OpIConst:
		if inst.Src1 == 0 {
			// XOR-zeroing idiom
			return []MachInst{{Op: MachXorZero, Dst: VReg(inst.Dest)}}
		}
		return []MachInst{{Op: MachMovImm, Dst: VReg(inst.Dest), Src1: Imm(int64(int32(inst.Src1)))}}

	case air.OpFConst:
		return []MachInst{{Op: MachMovImm, Dst: VReg(inst.Dest), Src1: Imm(int64(inst.Src1))}}

	case air.OpCopy, air.OpMove:
		return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpIAdd:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachAdd, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpISub:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachSub, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpIMul:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachImul, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpIDiv:
		// IDIV: RDX:RAX / src → RAX (quotient), RDX (remainder)
		return []MachInst{
			{Op: MachMov, Dst: Phys(RAX), Src1: VReg(inst.Src1)},
			{Op: MachCqo},
			{Op: MachIdiv, Src1: VReg(inst.Src2)},
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(RAX)},
		}

	case air.OpIMod:
		// IDIV: remainder in RDX
		return []MachInst{
			{Op: MachMov, Dst: Phys(RAX), Src1: VReg(inst.Src1)},
			{Op: MachCqo},
			{Op: MachIdiv, Src1: VReg(inst.Src2)},
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(RDX)},
		}

	case air.OpNeg:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachNeg, Dst: VReg(inst.Dest)},
		}

	case air.OpNot:
		// Logical not: CMP src, 0; SETE dst; MOVZX dst, dst
		return []MachInst{
			{Op: MachCmp, Dst: VReg(inst.Src1), Src1: Imm(0)},
			{Op: MachSetCC, CC: CCE, Dst: VReg(inst.Dest)},
			{Op: MachMovzxB, Dst: VReg(inst.Dest), Src1: VReg(inst.Dest)},
		}

	case air.OpEq:
		return selectCmp(inst, CCE)
	case air.OpNe:
		return selectCmp(inst, CCNE)
	case air.OpLt:
		return selectCmp(inst, CCL)
	case air.OpLe:
		return selectCmp(inst, CCLE)
	case air.OpGt:
		return selectCmp(inst, CCG)
	case air.OpGe:
		return selectCmp(inst, CCGE)

	case air.OpAnd:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachAnd, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpOr:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachOr, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpXor:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachXor, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpShl:
		// SHL requires shift amount in CL
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachMov, Dst: Phys(RCX), Src1: VReg(inst.Src2)},
			{Op: MachShl, Dst: VReg(inst.Dest)},
		}

	case air.OpShr:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachMov, Dst: Phys(RCX), Src1: VReg(inst.Src2)},
			{Op: MachSar, Dst: VReg(inst.Dest)},
		}

	case air.OpReturn:
		if inst.Src1 != 0 {
			return []MachInst{
				{Op: MachMov, Dst: Phys(RAX), Src1: VReg(inst.Src1)},
				{Op: MachRet},
			}
		}
		return []MachInst{{Op: MachRet}}

	case air.OpJump:
		return []MachInst{{Op: MachJmp, Dst: LabelOp(inst.Src1)}}

	case air.OpBranch:
		return []MachInst{
			{Op: MachTest, Dst: VReg(inst.Src1), Src1: VReg(inst.Src1)},
			{Op: MachJcc, CC: CCNE, Dst: LabelOp(inst.Src2)},
			{Op: MachJmp, Dst: LabelOp(inst.Dest)},
		}

	case air.OpCall:
		mi := MachInst{Op: MachCall, Src1: Imm(int64(inst.Src1))}
		if inst.Dest != 0 {
			return []MachInst{
				mi,
				{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(RAX)},
			}
		}
		return []MachInst{mi}

	case air.OpAlloc:
		// External call to malloc
		return []MachInst{
			{Op: MachCall, Src1: Imm(-1)}, // placeholder for malloc
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(RAX)},
		}

	case air.OpFree:
		return []MachInst{
			{Op: MachMov, Dst: Phys(RDI), Src1: VReg(inst.Src1)},
			{Op: MachCall, Src1: Imm(-2)}, // placeholder for free
		}

	case air.OpLoad:
		return []MachInst{{Op: MachLoad, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpStore:
		return []MachInst{{Op: MachStore, Dst: VReg(inst.Src2), Src1: VReg(inst.Src1)}}

	default:
		return []MachInst{{Op: MachNop}}
	}
}

// selectCmp generates CMP + SETcc + MOVZX for comparison operators.
func selectCmp(inst *air.AirInst, cc CondCode) []MachInst {
	return []MachInst{
		{Op: MachCmp, Dst: VReg(inst.Src1), Src1: VReg(inst.Src2)},
		{Op: MachSetCC, CC: cc, Dst: VReg(inst.Dest)},
		{Op: MachMovzxB, Dst: VReg(inst.Dest), Src1: VReg(inst.Dest)},
	}
}
