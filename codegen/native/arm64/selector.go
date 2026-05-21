package arm64

import (
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p13-t01: ARM64 Instruction Selector
//
// Translates AIR instructions into ARM64 MachInst with virtual registers.
// ARM64 is a three-operand RISC architecture (unlike x86's two-operand).
// --------------------------------------------------------------------------

// MachOpKind identifies the type of ARM64 machine instruction.
type MachOpKind uint16

const (
	MachNop MachOpKind = iota
	MachMov
	MachMovz
	MachMovk
	MachAdd
	MachSub
	MachMul
	MachSdiv
	MachMsub     // Rd = Ra - Rn*Rm (for modulo)
	MachNeg
	MachAnd
	MachOrr
	MachEor
	MachLsl
	MachAsr
	MachCmp
	MachCmpImm
	MachCset
	MachLdr
	MachStr
	MachStp
	MachLdp
	MachB        // unconditional branch
	MachBCond    // conditional branch
	MachBl       // branch with link (call)
	MachBlr      // call via register
	MachRet
	MachCbz
	MachCbnz
	MachLabel
)

// OperandKind describes whether an operand is a register, immediate, or label.
type OperandKind uint8

const (
	OpndNone OperandKind = iota
	OpndVReg
	OpndPhys
	OpndImm
	OpndLabel
)

// MachOperand is a single operand.
type MachOperand struct {
	Kind  OperandKind
	VReg  uint32
	Phys  PhysReg
	Imm   int64
	Label uint32
}

// MachInst is a single ARM64 machine instruction with virtual registers.
type MachInst struct {
	Op   MachOpKind
	CC   CondCode
	Dst  MachOperand
	Src1 MachOperand
	Src2 MachOperand
	Src3 MachOperand // for 4-operand instructions (MSUB)
}

// Operand constructors
func VReg(id uint32) MachOperand    { return MachOperand{Kind: OpndVReg, VReg: id} }
func Phys(r PhysReg) MachOperand    { return MachOperand{Kind: OpndPhys, Phys: r} }
func Imm(v int64) MachOperand       { return MachOperand{Kind: OpndImm, Imm: v} }
func LabelOp(id uint32) MachOperand { return MachOperand{Kind: OpndLabel, Label: id} }

// Select translates an AirFunc into ARM64 MachInsts.
func Select(fn *air.AirFunc) []MachInst {
	var result []MachInst

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

func selectInst(inst *air.AirInst) []MachInst {
	switch inst.Opcode {
	case air.OpNop:
		return nil

	case air.OpIConst:
		if inst.Src1 == 0 {
			// MOV Xd, XZR
			return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(PhysReg(31))}}
		}
		// MOVZ with up to 3 MOVK for large constants
		return materializeImm(inst.Dest, int64(int32(inst.Src1)))

	case air.OpFConst:
		return materializeImm(inst.Dest, int64(inst.Src1))

	case air.OpCopy, air.OpMove:
		return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpIAdd:
		return []MachInst{{Op: MachAdd, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}

	case air.OpISub:
		return []MachInst{{Op: MachSub, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}

	case air.OpIMul:
		return []MachInst{{Op: MachMul, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}

	case air.OpIDiv:
		return []MachInst{{Op: MachSdiv, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}

	case air.OpIMod:
		// ARM64 has no MOD instruction. Use SDIV + MSUB:
		// t = SDIV src1, src2
		// dest = MSUB t, src2, src1  (src1 - t*src2)
		tmpVReg := inst.Dest + 0x8000 // temporary VReg
		return []MachInst{
			{Op: MachSdiv, Dst: VReg(tmpVReg), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)},
			{Op: MachMsub, Dst: VReg(inst.Dest), Src1: VReg(tmpVReg), Src2: VReg(inst.Src2), Src3: VReg(inst.Src1)},
		}

	case air.OpNeg:
		return []MachInst{{Op: MachNeg, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpNot:
		// CMP src, #0; CSET dst, EQ
		return []MachInst{
			{Op: MachCmpImm, Dst: VReg(inst.Src1), Src1: Imm(0)},
			{Op: MachCset, CC: CondEQ, Dst: VReg(inst.Dest)},
		}

	case air.OpEq:
		return selectCmp(inst, CondEQ)
	case air.OpNe:
		return selectCmp(inst, CondNE)
	case air.OpLt:
		return selectCmp(inst, CondLT)
	case air.OpLe:
		return selectCmp(inst, CondLE)
	case air.OpGt:
		return selectCmp(inst, CondGT)
	case air.OpGe:
		return selectCmp(inst, CondGE)

	case air.OpAnd:
		return []MachInst{{Op: MachAnd, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}
	case air.OpOr:
		return []MachInst{{Op: MachOrr, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}
	case air.OpXor:
		return []MachInst{{Op: MachEor, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}

	case air.OpShl:
		return []MachInst{{Op: MachLsl, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}
	case air.OpShr:
		return []MachInst{{Op: MachAsr, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1), Src2: VReg(inst.Src2)}}

	case air.OpReturn:
		if inst.Src1 != 0 {
			return []MachInst{
				{Op: MachMov, Dst: Phys(X0), Src1: VReg(inst.Src1)},
				{Op: MachRet},
			}
		}
		return []MachInst{{Op: MachRet}}

	case air.OpJump:
		return []MachInst{{Op: MachB, Dst: LabelOp(inst.Src1)}}

	case air.OpBranch:
		return []MachInst{
			{Op: MachCbnz, Dst: VReg(inst.Src1), Src1: LabelOp(inst.Src2)},
			{Op: MachB, Dst: LabelOp(inst.Dest)},
		}

	case air.OpCall:
		mi := MachInst{Op: MachBl, Src1: Imm(int64(inst.Src1))}
		if inst.Dest != 0 {
			return []MachInst{
				mi,
				{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(X0)},
			}
		}
		return []MachInst{mi}

	case air.OpLoad:
		return []MachInst{{Op: MachLdr, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpStore:
		return []MachInst{{Op: MachStr, Dst: VReg(inst.Src1), Src1: VReg(inst.Src2)}}

	default:
		return []MachInst{{Op: MachNop}}
	}
}

func selectCmp(inst *air.AirInst, cc CondCode) []MachInst {
	return []MachInst{
		{Op: MachCmp, Dst: VReg(inst.Src1), Src1: VReg(inst.Src2)},
		{Op: MachCset, CC: cc, Dst: VReg(inst.Dest)},
	}
}

// materializeImm generates MOVZ + MOVK sequence for a 64-bit constant.
func materializeImm(dest uint32, val int64) []MachInst {
	uval := uint64(val)
	result := []MachInst{{
		Op:   MachMovz,
		Dst:  VReg(dest),
		Src1: Imm(int64(uval & 0xFFFF)),
		Src2: Imm(0), // shift = 0
	}}

	if (uval>>16)&0xFFFF != 0 {
		result = append(result, MachInst{
			Op:   MachMovk,
			Dst:  VReg(dest),
			Src1: Imm(int64((uval >> 16) & 0xFFFF)),
			Src2: Imm(16),
		})
	}
	if (uval>>32)&0xFFFF != 0 {
		result = append(result, MachInst{
			Op:   MachMovk,
			Dst:  VReg(dest),
			Src1: Imm(int64((uval >> 32) & 0xFFFF)),
			Src2: Imm(32),
		})
	}
	if (uval>>48)&0xFFFF != 0 {
		result = append(result, MachInst{
			Op:   MachMovk,
			Dst:  VReg(dest),
			Src1: Imm(int64((uval >> 48) & 0xFFFF)),
			Src2: Imm(48),
		})
	}

	return result
}
