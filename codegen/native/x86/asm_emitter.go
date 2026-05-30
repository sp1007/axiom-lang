package x86

import (
	"fmt"
	"strings"
)

// --------------------------------------------------------------------------
// Assembly Text Emitter for NASM & FASM
//
// Converts x86-64 MachInsts (with register allocation and frame completed)
// into standard Intel-syntax assembly code compatible with NASM and FASM.
// --------------------------------------------------------------------------

// AsmEmitter converts MachInsts into standard Intel-syntax assembly text.
type AsmEmitter struct {
	allocs map[uint32]RegAllocation
	Format string // "nasm" or "fasm"
}

// NewAsmEmitter creates a new assembly text emitter.
func NewAsmEmitter(allocs map[uint32]RegAllocation, format string) *AsmEmitter {
	return &AsmEmitter{
		allocs: allocs,
		Format: format,
	}
}

// FormatOperand formats a MachOperand to Intel syntax.
func (ae *AsmEmitter) FormatOperand(op MachOperand) string {
	switch op.Kind {
	case OpndPhys:
		return op.Phys.String()
	case OpndVReg:
		if alloc, ok := ae.allocs[op.VReg]; ok {
			return alloc.Phys.String()
		}
		return "none"
	case OpndImm:
		return fmt.Sprintf("%d", op.Imm)
	case OpndLabel:
		if ae.Format == "winasm" {
			return fmt.Sprintf("L_b_%d", op.Label)
		}
		return fmt.Sprintf(".L_b_%d", op.Label)
	default:
		return "none"
	}
}

// EmitInst converts a single MachInst to standard Intel-syntax string.
func (ae *AsmEmitter) EmitInst(inst MachInst) string {
	switch inst.Op {
	case MachNop:
		return "    nop"
	case MachLabel:
		if ae.Format == "winasm" {
			return fmt.Sprintf("L_b_%d:", inst.Dst.Label)
		}
		return fmt.Sprintf(".L_b_%d:", inst.Dst.Label)
	case MachRet:
		return "    ret"
	case MachMov:
		dstReg := ae.resolveReg(inst.Dst)
		srcReg := ae.resolveReg(inst.Src1)
		if dstReg.IsXMM() && srcReg.IsXMM() {
			return fmt.Sprintf("    movsd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
		} else if dstReg.IsXMM() || srcReg.IsXMM() {
			return fmt.Sprintf("    movq %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
		}
		return fmt.Sprintf("    mov %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachMovImm:
		return fmt.Sprintf("    mov %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachXorZero:
		dst := ae.FormatOperand(inst.Dst)
		return fmt.Sprintf("    xor %s, %s", dst, dst)
	case MachAdd:
		return fmt.Sprintf("    add %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachSub:
		return fmt.Sprintf("    sub %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachImul:
		return fmt.Sprintf("    imul %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachIdiv:
		return fmt.Sprintf("    idiv %s", ae.FormatOperand(inst.Src1))
	case MachCqo:
		return "    cqo"
	case MachNeg:
		return fmt.Sprintf("    neg %s", ae.FormatOperand(inst.Dst))
	case MachNot:
		return fmt.Sprintf("    not %s", ae.FormatOperand(inst.Dst))
	case MachAnd:
		return fmt.Sprintf("    and %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachOr:
		return fmt.Sprintf("    or %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachXor:
		return fmt.Sprintf("    xor %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachShl:
		return fmt.Sprintf("    shl %s, cl", ae.FormatOperand(inst.Dst))
	case MachSar:
		return fmt.Sprintf("    sar %s, cl", ae.FormatOperand(inst.Dst))
	case MachCmp:
		return fmt.Sprintf("    cmp %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachTest:
		return fmt.Sprintf("    test %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachSetCC:
		dstReg := ae.resolveReg(inst.Dst)
		byteReg := toByteReg(dstReg)
		return fmt.Sprintf("    set%s %s", inst.CC.String(), byteReg)
	case MachMovzxB:
		srcReg := ae.resolveReg(inst.Src1)
		byteReg := toByteReg(srcReg)
		return fmt.Sprintf("    movzx %s, %s", ae.FormatOperand(inst.Dst), byteReg)
	case MachPush:
		return fmt.Sprintf("    push %s", ae.FormatOperand(inst.Src1))
	case MachPop:
		return fmt.Sprintf("    pop %s", ae.FormatOperand(inst.Dst))
	case MachJmp:
		return fmt.Sprintf("    jmp %s", ae.FormatOperand(inst.Dst))
	case MachJcc:
		return fmt.Sprintf("    j%s %s", inst.CC.String(), ae.FormatOperand(inst.Dst))
	case MachLoad:
		dstReg := ae.resolveReg(inst.Dst)
		base := ae.FormatOperand(inst.Src1)
		disp := int64(0)
		if inst.Src2.Kind == OpndImm {
			disp = inst.Src2.Imm
		}
		addrStr := ""
		if disp == 0 {
			addrStr = fmt.Sprintf("[%s]", base)
		} else if disp > 0 {
			addrStr = fmt.Sprintf("[%s + %d]", base, disp)
		} else {
			addrStr = fmt.Sprintf("[%s - %d]", base, -disp)
		}
		if dstReg.IsXMM() {
			if ae.Format == "winasm" {
				return fmt.Sprintf("    movsd %s, qword ptr %s", ae.FormatOperand(inst.Dst), addrStr)
			}
			return fmt.Sprintf("    movsd %s, %s", ae.FormatOperand(inst.Dst), addrStr)
		}
		if ae.Format == "winasm" {
			return fmt.Sprintf("    mov %s, qword ptr %s", ae.FormatOperand(inst.Dst), addrStr)
		}
		return fmt.Sprintf("    mov %s, %s", ae.FormatOperand(inst.Dst), addrStr)
	case MachStore:
		srcReg := ae.resolveReg(inst.Src1)
		base := ae.FormatOperand(inst.Dst)
		disp := int64(0)
		if inst.Src2.Kind == OpndImm {
			disp = inst.Src2.Imm
		}
		addrStr := ""
		if disp == 0 {
			addrStr = fmt.Sprintf("[%s]", base)
		} else if disp > 0 {
			addrStr = fmt.Sprintf("[%s + %d]", base, disp)
		} else {
			addrStr = fmt.Sprintf("[%s - %d]", base, -disp)
		}
		if srcReg.IsXMM() {
			if ae.Format == "winasm" {
				return fmt.Sprintf("    movsd qword ptr %s, %s", addrStr, ae.FormatOperand(inst.Src1))
			}
			return fmt.Sprintf("    movsd %s, %s", addrStr, ae.FormatOperand(inst.Src1))
		}
		if ae.Format == "winasm" {
			return fmt.Sprintf("    mov qword ptr %s, %s", addrStr, ae.FormatOperand(inst.Src1))
		}
		return fmt.Sprintf("    mov %s, %s", addrStr, ae.FormatOperand(inst.Src1))
	case MachFAdd:
		return fmt.Sprintf("    addsd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachFSub:
		return fmt.Sprintf("    subsd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachFMul:
		return fmt.Sprintf("    mulsd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachFDiv:
		return fmt.Sprintf("    divsd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachFCmp:
		return fmt.Sprintf("    comisd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachItof:
		return fmt.Sprintf("    cvtsi2sd %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachFtoi:
		return fmt.Sprintf("    cvttsd2si %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachMovDQ:
		return fmt.Sprintf("    movq %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	case MachMovQD:
		return fmt.Sprintf("    movq %s, %s", ae.FormatOperand(inst.Dst), ae.FormatOperand(inst.Src1))
	default:
		return "    ; unknown instruction"
	}
}

// resolveReg maps a MachOperand to a physical register.
func (ae *AsmEmitter) resolveReg(op MachOperand) PhysReg {
	switch op.Kind {
	case OpndPhys:
		return op.Phys
	case OpndVReg:
		if alloc, ok := ae.allocs[op.VReg]; ok {
			return alloc.Phys
		}
		return RegNone
	default:
		return RegNone
	}
}

// toByteReg maps a 64-bit physical register to its 8-bit equivalent name.
func toByteReg(reg PhysReg) string {
	hw := reg.HWReg()
	if reg.IsGPR() {
		names := [16]string{
			"al", "cl", "dl", "bl", "spl", "bpl", "sil", "dil",
			"r8b", "r9b", "r10b", "r11b", "r12b", "r13b", "r14b", "r15b",
		}
		if hw < 16 {
			return names[hw]
		}
	}
	return "al"
}

// EmitFunction generates assembly text for a complete function.
func (ae *AsmEmitter) EmitFunction(fnName string, insts []MachInst, frame *StackFrame, symNameResolver func(uint32) string) string {
	var sb strings.Builder

	// Write function label
	if ae.Format == "winasm" {
		fmt.Fprintf(&sb, "%s PROC\n", fnName)
	} else {
		fmt.Fprintf(&sb, "%s:\n", fnName)
	}

	// Emit prologue
	for _, inst := range EmitPrologue(frame) {
		sb.WriteString(ae.FormatInst(inst, fnName, symNameResolver))
		sb.WriteString("\n")
	}

	// Emit body
	for _, inst := range insts {
		if inst.Op == MachRet {
			// Emit epilogue before returning
			for _, epInst := range EmitEpilogue(frame) {
				if epInst.Op != MachRet {
					sb.WriteString(ae.FormatInst(epInst, fnName, symNameResolver))
					sb.WriteString("\n")
				}
			}
			sb.WriteString("    ret\n")
		} else {
			sb.WriteString(ae.FormatInst(inst, fnName, symNameResolver))
			sb.WriteString("\n")
		}
	}

	if ae.Format == "winasm" {
		fmt.Fprintf(&sb, "%s ENDP\n", fnName)
	}

	return sb.String()
}

// FormatInst formats an instruction, resolving relocations / symbols.
func (ae *AsmEmitter) FormatInst(inst MachInst, fnName string, symNameResolver func(uint32) string) string {
	if inst.Op == MachCall {
		if inst.Src1.Imm == 0 {
			return fmt.Sprintf("    call %s", fnName)
		} else {
			symName := symNameResolver(uint32(inst.Src1.Imm))
			return fmt.Sprintf("    call %s", symName)
		}
	}
	return ae.EmitInst(inst)
}
