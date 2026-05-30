package x86

import (
	"github.com/axiom-lang/axiom/compiler/types"
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
	MachFAdd
	MachFSub
	MachFMul
	MachFDiv
	MachFCmp
	MachItof
	MachFtoi
	MachMovDQ
	MachMovQD
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

type instructionSelector struct {
	fn                *air.AirFunc
	abi               ABI
	table             *types.TypeTable
	paramIdxProcessed int
	maxVReg           uint32
}

func (sel *instructionSelector) nextVReg() uint32 {
	sel.maxVReg++
	return sel.maxVReg
}

func (sel *instructionSelector) is16Byte(reg uint32) bool {
	if reg == 0 {
		return false
	}
	t := sel.getRegisterType(reg)
	if t == 0 {
		if reg >= 1 && reg <= uint32(len(sel.fn.Params)) {
			t = uint16(sel.fn.Params[reg-1])
		}
	}
	if t != 0 && sel.table != nil {
		size, _ := sel.typeSizeAndAlign(t)
		return size == 16
	}
	return false
}

func (sel *instructionSelector) emitBlockCopyExt(dstReg uint32, dstDisp uint32, srcReg uint32, srcDisp uint32, size uint32) []MachInst {
	var insts []MachInst
	offset := uint32(0)
	remaining := size
	for remaining >= 8 {
		tmpVal := sel.nextVReg()
		insts = append(insts, MachInst{
			Op:   MachLoad,
			Dst:  VReg(tmpVal),
			Src1: VReg(srcReg),
			Src2: Imm(int64(srcDisp + offset)),
		})
		insts = append(insts, MachInst{
			Op:   MachStore,
			Dst:  VReg(dstReg),
			Src1: VReg(tmpVal),
			Src2: Imm(int64(dstDisp + offset)),
		})
		offset += 8
		remaining -= 8
	}
	if remaining >= 4 {
		tmpVal := sel.nextVReg()
		insts = append(insts, MachInst{
			Op:   MachLoad,
			Dst:  VReg(tmpVal),
			Src1: VReg(srcReg),
			Src2: Imm(int64(srcDisp + offset)),
		})
		insts = append(insts, MachInst{
			Op:   MachStore,
			Dst:  VReg(dstReg),
			Src1: VReg(tmpVal),
			Src2: Imm(int64(dstDisp + offset)),
		})
		offset += 4
		remaining -= 4
	}
	if remaining >= 2 {
		tmpVal := sel.nextVReg()
		insts = append(insts, MachInst{
			Op:   MachLoad,
			Dst:  VReg(tmpVal),
			Src1: VReg(srcReg),
			Src2: Imm(int64(srcDisp + offset)),
		})
		insts = append(insts, MachInst{
			Op:   MachStore,
			Dst:  VReg(dstReg),
			Src1: VReg(tmpVal),
			Src2: Imm(int64(dstDisp + offset)),
		})
		offset += 2
		remaining -= 2
	}
	if remaining >= 1 {
		tmpVal := sel.nextVReg()
		insts = append(insts, MachInst{
			Op:   MachLoad,
			Dst:  VReg(tmpVal),
			Src1: VReg(srcReg),
			Src2: Imm(int64(srcDisp + offset)),
		})
		insts = append(insts, MachInst{
			Op:   MachStore,
			Dst:  VReg(dstReg),
			Src1: VReg(tmpVal),
			Src2: Imm(int64(dstDisp + offset)),
		})
	}
	return insts
}

func (sel *instructionSelector) emitBlockCopy(dstReg uint32, dstDisp uint32, srcReg uint32, size uint32) []MachInst {
	finalSrcReg := srcReg
	var insts []MachInst
	if sel.is16Byte(srcReg) {
		finalSrcReg = sel.nextVReg()
		insts = append(insts, MachInst{
			Op:   MachLea,
			Dst:  VReg(finalSrcReg),
			Src1: VReg(srcReg),
		})
	}
	insts = append(insts, sel.emitBlockCopyExt(dstReg, dstDisp, finalSrcReg, 0, size)...)
	return insts
}

// Select translates an AirFunc into a list of MachInst per basic block.
// Virtual registers map 1:1 to AIR registers.
func Select(fn *air.AirFunc, abi ABI, table *types.TypeTable) []MachInst {
	maxVReg := uint32(len(fn.Params))
	for i := range fn.Insts {
		if fn.Insts[i].Dest > maxVReg {
			maxVReg = fn.Insts[i].Dest
		}
	}

	sel := &instructionSelector{
		fn:      fn,
		abi:     abi,
		table:   table,
		maxVReg: maxVReg,
	}
	var result []MachInst

	// If blocks have instruction indices, emit per-block
	if len(fn.Blocks) > 0 && hasBlockInstrs(fn) {
		for bi := range fn.Blocks {
			blk := &fn.Blocks[bi]
			result = append(result, MachInst{Op: MachLabel, Dst: LabelOp(blk.ID)})
			for _, idx := range blk.Instrs {
				if int(idx) < len(fn.Insts) {
					result = append(result, sel.selectInst(&fn.Insts[idx])...)
				}
			}
		}
	} else {
		// Flat instruction list
		for i := range fn.Insts {
			result = append(result, sel.selectInst(&fn.Insts[i])...)
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
func (sel *instructionSelector) selectInst(inst *air.AirInst) []MachInst {
	switch inst.Opcode {
	case air.OpNop:
		return nil

	case air.OpIConst:
		tID := inst.TypeID
		isStr := (tID == 12)
		if tID != 0 && sel.table != nil {
			size, _ := sel.typeSizeAndAlign(tID)
			if size == 16 {
				isStr = true
			}
		}
		if isStr {
			tmpStructAddr := sel.nextVReg()
			tmpDestAddr := sel.nextVReg()
			var insts []MachInst
			insts = append(insts, MachInst{
				Op:   MachMovImm,
				Dst:  VReg(tmpStructAddr),
				Src1: MachOperand{Kind: OpndImm, VReg: 2, Imm: int64(inst.Src1)},
			})
			insts = append(insts, MachInst{
				Op:   MachLea,
				Dst:  VReg(tmpDestAddr),
				Src1: VReg(inst.Dest),
			})
			insts = append(insts, sel.emitBlockCopy(tmpDestAddr, 0, tmpStructAddr, 16)...)
			return insts
		}
		if inst.Src1 == 0 {
			// XOR-zeroing idiom
			return []MachInst{{Op: MachXorZero, Dst: VReg(inst.Dest)}}
		}
		return []MachInst{{Op: MachMovImm, Dst: VReg(inst.Dest), Src1: Imm(int64(int32(inst.Src1)))}}

	case air.OpFConst:
		// Convert integer constant Src1 to double in Dest
		tmpGPR := inst.Dest + 50000
		return []MachInst{
			{Op: MachMovImm, Dst: VReg(tmpGPR), Src1: Imm(int64(inst.Src1))},
			{Op: MachItof, Dst: VReg(inst.Dest), Src1: VReg(tmpGPR)},
		}

	case air.OpCopy, air.OpMove:
		// Check if this is the initial parameter copy instruction
		if sel.paramIdxProcessed < len(sel.fn.Params) && inst.Src1 == uint32(sel.paramIdxProcessed+1) {
			paramIdx := sel.paramIdxProcessed
			sel.paramIdxProcessed++

			var phys PhysReg = RegNone
			if types.TypeID(sel.fn.Params[paramIdx]).IsFloat() {
				if paramIdx < len(sel.abi.FloatArgRegs()) {
					phys = sel.abi.FloatArgRegs()[paramIdx]
				}
			} else {
				if paramIdx < len(sel.abi.IntArgRegs()) {
					phys = sel.abi.IntArgRegs()[paramIdx]
				}
			}

			if phys != RegNone {
				destType := sel.getRegisterType(inst.Dest)
				destSize := uint32(8)
				if destType != 0 && sel.table != nil {
					destSize, _ = sel.typeSizeAndAlign(destType)
				}
				if destSize == 16 {
					tmpVal1 := sel.nextVReg()
					tmpVal2 := sel.nextVReg()
					tmpDestAddr := sel.nextVReg()
					return []MachInst{
						{Op: MachLoad, Dst: VReg(tmpVal1), Src1: Phys(phys)},
						{Op: MachLoad, Dst: VReg(tmpVal2), Src1: Phys(phys), Src2: Imm(8)},
						{Op: MachLea, Dst: VReg(tmpDestAddr), Src1: VReg(inst.Dest)},
						{Op: MachStore, Dst: VReg(tmpDestAddr), Src1: VReg(tmpVal1)},
						{Op: MachStore, Dst: VReg(tmpDestAddr), Src1: VReg(tmpVal2), Src2: Imm(8)},
					}
				} else {
					return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(phys)}}
				}
			} else {
				var offset int64
				if sel.abi.Name() == "win64" {
					offset = 48 + int64(paramIdx-4)*8
				} else {
					offset = 16 + int64(paramIdx-6)*8
				}
				return []MachInst{{Op: MachLoad, Dst: VReg(inst.Dest), Src1: Phys(RBP), Src2: Imm(offset)}}
			}
		}
		if sel.is16Byte(inst.Dest) {
			tmpDestAddr := sel.nextVReg()
			tmpSrcAddr := sel.nextVReg()
			var insts []MachInst
			insts = append(insts, MachInst{Op: MachLea, Dst: VReg(tmpDestAddr), Src1: VReg(inst.Dest)})
			insts = append(insts, MachInst{Op: MachLea, Dst: VReg(tmpSrcAddr), Src1: VReg(inst.Src1)})
			insts = append(insts, sel.emitBlockCopy(tmpDestAddr, 0, tmpSrcAddr, 16)...)
			return insts
		}
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

	case air.OpFAdd:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachFAdd, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpFSub:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachFSub, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpFMul:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachFMul, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpFDiv:
		return []MachInst{
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
			{Op: MachFDiv, Dst: VReg(inst.Dest), Src1: VReg(inst.Src2)},
		}

	case air.OpIToF:
		return []MachInst{{Op: MachItof, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpFToI:
		return []MachInst{{Op: MachFtoi, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpNeg:
		t := sel.getRegisterType(inst.Src1)
		if types.TypeID(t).IsFloat() {
			tmpSign := inst.Dest + 60000
			return []MachInst{
				{Op: MachMovImm, Dst: VReg(tmpSign), Src1: Imm(int64(-1 << 63))},
				{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)},
				{Op: MachXor, Dst: VReg(inst.Dest), Src1: VReg(tmpSign)},
			}
		}
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
		return sel.selectComparison(inst, CCE, CCE)
	case air.OpNe:
		return sel.selectComparison(inst, CCNE, CCNE)
	case air.OpLt:
		return sel.selectComparison(inst, CCL, CCB)
	case air.OpLe:
		return sel.selectComparison(inst, CCLE, CCBE)
	case air.OpGt:
		return sel.selectComparison(inst, CCG, CCA)
	case air.OpGe:
		return sel.selectComparison(inst, CCGE, CCAE)

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
			var retReg PhysReg = RAX
			if types.TypeID(sel.fn.RetType).IsFloat() {
				retReg = XMM0
			}
			return []MachInst{
				{Op: MachMov, Dst: Phys(retReg), Src1: VReg(inst.Src1)},
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
		var insts []MachInst

		// If there is an argument list index in Src2
		argStart := inst.Src2
		argCount := uint32(0)
		if argStart < uint32(len(sel.fn.Extras)) {
			argCount = sel.fn.Extras[argStart]
		}

		// Win64 requires 32-byte shadow space before CALL
		if sel.abi.Name() == "win64" {
			insts = append(insts, MachInst{
				Op:   MachSub,
				Dst:  Phys(RSP),
				Src1: Imm(32),
			})
		}

		// Copy arguments to physical registers based on ABI (float -> XMM, int -> GPR)
		for i := uint32(0); i < argCount; i++ {
			argReg := sel.fn.Extras[argStart+1+i]
			argType := sel.getRegisterType(argReg)
			var phys PhysReg = RegNone
			if types.TypeID(argType).IsFloat() {
				if i < uint32(len(sel.abi.FloatArgRegs())) {
					phys = sel.abi.FloatArgRegs()[i]
				}
			} else {
				if i < uint32(len(sel.abi.IntArgRegs())) {
					phys = sel.abi.IntArgRegs()[i]
				}
			}
			if phys != RegNone {
				insts = append(insts, MachInst{
					Op:   MachMov,
					Dst:  Phys(phys),
					Src1: VReg(argReg),
				})
			}
		}

		// Emit the call itself
		insts = append(insts, MachInst{
			Op:   MachCall,
			Src1: Imm(int64(inst.Src1)),
		})

		// Restore shadow space for Win64
		if sel.abi.Name() == "win64" {
			insts = append(insts, MachInst{
				Op:   MachAdd,
				Dst:  Phys(RSP),
				Src1: Imm(32),
			})
		}

		// Copy return value if needed
		if inst.Dest != 0 {
			var retReg PhysReg = sel.abi.ReturnReg()
			if types.TypeID(inst.TypeID).IsFloat() {
				retReg = XMM0
			}
			insts = append(insts, MachInst{
				Op:   MachMov,
				Dst:  VReg(inst.Dest),
				Src1: Phys(retReg),
			})
		}

		return insts

	case air.OpAlloc:
		size, _ := sel.typeSizeAndAlign(inst.TypeID)
		arg0 := sel.abi.IntArgRegs()[0]
		return []MachInst{
			{Op: MachMovImm, Dst: Phys(arg0), Src1: Imm(int64(size))},
			{Op: MachCall, Src1: Imm(-1)}, // placeholder for malloc
			{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(RAX)},
		}

	case air.OpFree, air.OpDestroy:
		arg0 := sel.abi.IntArgRegs()[0]
		return []MachInst{
			{Op: MachMov, Dst: Phys(arg0), Src1: VReg(inst.Src1)},
			{Op: MachCall, Src1: Imm(-2)}, // placeholder for free
		}

	case air.OpLoad:
		return []MachInst{{Op: MachLoad, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}

	case air.OpStore:
		return []MachInst{{Op: MachStore, Dst: VReg(inst.Src2), Src1: VReg(inst.Src1)}}

	case air.OpGetField:
		structType := sel.getRegisterType(inst.Src1)
		disp := sel.fieldOffset(structType, inst.Src2)
		return []MachInst{
			{
				Op:   MachLoad,
				Dst:  VReg(inst.Dest),
				Src1: VReg(inst.Src1),
				Src2: Imm(int64(disp)),
			},
		}

	case air.OpSetField:
		structType := sel.getRegisterType(inst.Src1)
		disp := sel.fieldOffset(structType, inst.Src2)
		return []MachInst{
			{
				Op:   MachStore,
				Dst:  VReg(inst.Src1),
				Src1: VReg(inst.Dest),
				Src2: Imm(int64(disp)),
			},
		}

	case air.OpCast:
		tID := inst.TypeID
		isStr := (tID == 12)
		if tID != 0 && sel.table != nil {
			size, _ := sel.typeSizeAndAlign(tID)
			if size == 16 {
				isStr = true
			}
		}
		if isStr {
			tmpLen := sel.nextVReg()
			arg0 := sel.abi.IntArgRegs()[0]
			var insts []MachInst
			if sel.abi.Name() == "win64" {
				insts = append(insts, MachInst{
					Op:   MachSub,
					Dst:  Phys(RSP),
					Src1: Imm(32),
				})
			}
			insts = append(insts, MachInst{
				Op:   MachMov,
				Dst:  Phys(arg0),
				Src1: VReg(inst.Src1),
			})
			insts = append(insts, MachInst{
				Op:   MachCall,
				Src1: Imm(-21),
			})
			if sel.abi.Name() == "win64" {
				insts = append(insts, MachInst{
					Op:   MachAdd,
					Dst:  Phys(RSP),
					Src1: Imm(32),
				})
			}
			insts = append(insts, MachInst{
				Op:   MachMov,
				Dst:  VReg(tmpLen),
				Src1: Phys(RAX),
			})
			tmpDestAddr := sel.nextVReg()
			insts = append(insts, MachInst{
				Op:   MachLea,
				Dst:  VReg(tmpDestAddr),
				Src1: VReg(inst.Dest),
			})
			insts = append(insts, MachInst{
				Op:   MachStore,
				Dst:  VReg(tmpDestAddr),
				Src1: VReg(inst.Src1),
			})
			insts = append(insts, MachInst{
				Op:   MachStore,
				Dst:  VReg(tmpDestAddr),
				Src1: VReg(tmpLen),
				Src2: Imm(8),
			})
			return insts
		} else {
			srcType := sel.getRegisterType(inst.Src1)
			is16Byte := (srcType == 12)
			if srcType != 0 && sel.table != nil {
				size, _ := sel.typeSizeAndAlign(srcType)
				if size == 16 {
					is16Byte = true
				}
			}
			if is16Byte {
				if sel.paramIdxProcessed < len(sel.fn.Params) && inst.Src1 == uint32(sel.paramIdxProcessed+1) {
					paramIdx := sel.paramIdxProcessed
					sel.paramIdxProcessed++
					phys := RegNone
					if paramIdx < len(sel.abi.IntArgRegs()) {
						phys = sel.abi.IntArgRegs()[paramIdx]
					}
					destType := sel.getRegisterType(inst.Dest)
					destSize := uint32(8)
					if destType != 0 && sel.table != nil {
						destSize, _ = sel.typeSizeAndAlign(destType)
					}
					if phys != RegNone {
						if destSize == 16 {
							tmpVal1 := sel.nextVReg()
							tmpVal2 := sel.nextVReg()
							tmpDestAddr := sel.nextVReg()
							return []MachInst{
								{Op: MachLoad, Dst: VReg(tmpVal1), Src1: Phys(phys)},
								{Op: MachLoad, Dst: VReg(tmpVal2), Src1: Phys(phys), Src2: Imm(8)},
								{Op: MachLea, Dst: VReg(tmpDestAddr), Src1: VReg(inst.Dest)},
								{Op: MachStore, Dst: VReg(tmpDestAddr), Src1: VReg(tmpVal1)},
								{Op: MachStore, Dst: VReg(tmpDestAddr), Src1: VReg(tmpVal2), Src2: Imm(8)},
							}
						} else {
							return []MachInst{
								{Op: MachLoad, Dst: VReg(inst.Dest), Src1: Phys(phys)},
							}
						}
					} else {
						var offset int64
						if sel.abi.Name() == "win64" {
							offset = 48 + int64(paramIdx-4)*8
						} else {
							offset = 16 + int64(paramIdx-6)*8
						}
						tmpAddr := sel.nextVReg()
						var insts []MachInst
						insts = append(insts, MachInst{
							Op:   MachLoad,
							Dst:  VReg(tmpAddr),
							Src1: Phys(RBP),
							Src2: Imm(offset),
						})
						if destSize == 16 {
							tmpVal1 := sel.nextVReg()
							tmpVal2 := sel.nextVReg()
							tmpDestAddr := sel.nextVReg()
							insts = append(insts, MachInst{Op: MachLoad, Dst: VReg(tmpVal1), Src1: VReg(tmpAddr)})
							insts = append(insts, MachInst{Op: MachLoad, Dst: VReg(tmpVal2), Src1: VReg(tmpAddr), Src2: Imm(8)})
							insts = append(insts, MachInst{Op: MachLea, Dst: VReg(tmpDestAddr), Src1: VReg(inst.Dest)})
							insts = append(insts, MachInst{Op: MachStore, Dst: VReg(tmpDestAddr), Src1: VReg(tmpVal1)})
							insts = append(insts, MachInst{Op: MachStore, Dst: VReg(tmpDestAddr), Src1: VReg(tmpVal2), Src2: Imm(8)})
						} else {
							insts = append(insts, MachInst{Op: MachLoad, Dst: VReg(inst.Dest), Src1: VReg(tmpAddr)})
						}
						return insts
					}
				} else {
					destType := sel.getRegisterType(inst.Dest)
					destSize := uint32(8)
					if destType != 0 && sel.table != nil {
						destSize, _ = sel.typeSizeAndAlign(destType)
					}
					if destSize == 16 {
						return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}
					} else {
						return []MachInst{{Op: MachLoad, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}
					}
				}
			} else {
				if sel.paramIdxProcessed < len(sel.fn.Params) && inst.Src1 == uint32(sel.paramIdxProcessed+1) {
					paramIdx := sel.paramIdxProcessed
					sel.paramIdxProcessed++
					phys := RegNone
					if paramIdx < len(sel.abi.IntArgRegs()) {
						phys = sel.abi.IntArgRegs()[paramIdx]
					}
					if phys != RegNone {
						return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: Phys(phys)}}
					} else {
						var offset int64
						if sel.abi.Name() == "win64" {
							offset = 48 + int64(paramIdx-4)*8
						} else {
							offset = 16 + int64(paramIdx-6)*8
						}
						return []MachInst{{Op: MachLoad, Dst: VReg(inst.Dest), Src1: Phys(RBP), Src2: Imm(offset)}}
					}
				}
				return []MachInst{{Op: MachMov, Dst: VReg(inst.Dest), Src1: VReg(inst.Src1)}}
			}
		}

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

func (sel *instructionSelector) getRegisterType(reg uint32) uint16 {
	if reg == 0 {
		return 0
	}
	for i := range sel.fn.Insts {
		inst := &sel.fn.Insts[i]
		if inst.Dest == reg {
			if inst.TypeID == 0 {
				println("[DEBUG-TYPE] getRegisterType found defining inst for reg", reg, "but TypeID is 0! Inst opcode:", inst.Opcode.Mnemonic(), "in function NameID=", sel.fn.Name)
			}
			return inst.TypeID
		}
	}
	println("[DEBUG-TYPE] getRegisterType COULD NOT FIND defining inst for reg", reg, "in function NameID=", sel.fn.Name, "Total instructions:", len(sel.fn.Insts))
	for i := range sel.fn.Insts {
		inst := &sel.fn.Insts[i]
		println("  inst", i, "opcode:", inst.Opcode.Mnemonic(), "dest:", inst.Dest, "src1:", inst.Src1, "src2:", inst.Src2, "typeID:", inst.TypeID)
	}
	return 0
}

func (sel *instructionSelector) typeSizeAndAlign(typeID uint16) (uint32, uint32) {
	if typeID == 0 {
		return 8, 8 // default pointer size
	}
	if sel.table == nil {
		return 8, 8 // fallback
	}
	entry := sel.table.Entry(types.TypeID(typeID))
	if entry.Size != 0 {
		return entry.Size, entry.Align
	}

	switch entry.Kind {
	case types.KindPrimitive:
		return entry.Size, entry.Align
	case types.KindArray:
		elemID := sel.table.ArrayElem(types.TypeID(typeID))
		length := sel.table.ArrayLength(types.TypeID(typeID))
		elemSize, elemAlign := sel.typeSizeAndAlign(uint16(elemID))
		if elemAlign == 0 {
			elemAlign = 8
		}
		size := elemSize * length
		entry.Size = size
		entry.Align = elemAlign
		return size, elemAlign
	case types.KindStruct:
		// Compute struct layout dynamically
		info := sel.table.StructInfo(types.TypeID(typeID))
		offset := uint32(0)
		maxAlign := uint32(1)
		for i := range info.Fields {
			f := &info.Fields[i]
			fSize, fAlign := sel.typeSizeAndAlign(uint16(f.TypeID))
			if fAlign == 0 {
				fAlign = 8
			}
			// Align offset
			offset = (offset + fAlign - 1) & ^(fAlign - 1)
			f.Offset = offset
			offset += fSize
			if fAlign > maxAlign {
				maxAlign = fAlign
			}
		}
		// Align total size
		size := (offset + maxAlign - 1) & ^(maxAlign - 1)
		if size == 0 {
			size = 8 // non-empty struct minimum size
		}
		// Cache size/align in the entry
		entry.Size = size
		entry.Align = maxAlign
		return size, maxAlign
	default:
		return 8, 8 // fallback pointer size
	}
}

func (sel *instructionSelector) fieldOffset(structTypeID uint16, fieldIdx uint32) uint32 {
	sel.typeSizeAndAlign(structTypeID) // Ensure layout is computed
	if sel.table == nil {
		return fieldIdx * 8 // fallback
	}
	info := sel.table.StructInfo(types.TypeID(structTypeID))
	if int(fieldIdx) < len(info.Fields) {
		return info.Fields[fieldIdx].Offset
	}
	return fieldIdx * 8 // fallback
}

func (sel *instructionSelector) selectComparison(inst *air.AirInst, intCC, floatCC CondCode) []MachInst {
	t := sel.getRegisterType(inst.Src1)
	if types.TypeID(t).IsFloat() {
		return []MachInst{
			{Op: MachFCmp, Dst: VReg(inst.Src1), Src1: VReg(inst.Src2)},
			{Op: MachSetCC, CC: floatCC, Dst: VReg(inst.Dest)},
			{Op: MachMovzxB, Dst: VReg(inst.Dest), Src1: VReg(inst.Dest)},
		}
	}
	return []MachInst{
		{Op: MachCmp, Dst: VReg(inst.Src1), Src1: VReg(inst.Src2)},
		{Op: MachSetCC, CC: intCC, Dst: VReg(inst.Dest)},
		{Op: MachMovzxB, Dst: VReg(inst.Dest), Src1: VReg(inst.Dest)},
	}
}

