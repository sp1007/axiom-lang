package x86

// --------------------------------------------------------------------------
// p11-t10: x86-64 Machine Code Emitter
//
// Translates MachInst (with physical registers) into actual x86-64 machine
// code bytes. Handles label resolution, branch fixups, and records
// relocations for external symbols.
// --------------------------------------------------------------------------

// RelocKind identifies the type of relocation.
type RelocKind uint8

const (
	RelocPC32   RelocKind = iota // 32-bit PC-relative
	RelocAbs64                   // 64-bit absolute
	RelocPLT32                   // PLT entry (for calls to shared libs)
)

// Relocation records a reference that needs to be fixed up by the linker.
type Relocation struct {
	Offset   int       // byte offset in the .text section
	Kind     RelocKind
	SymName  uint32    // interned symbol name
	Addend   int64     // addend for the relocation
}

// Fixup records an internal branch fixup (label → offset).
type Fixup struct {
	Offset   int    // byte offset of the rel32 field to fix
	LabelID  uint32 // target block label
	InstSize int    // size of the instruction containing the fixup
}

// Emitter converts MachInsts into machine code bytes.
type Emitter struct {
	Code     []byte            // emitted machine code
	Relocs   []Relocation      // external relocations
	Labels   map[uint32]int    // label ID → byte offset
	Fixups   []Fixup           // internal branch fixups
	allocs   map[uint32]RegAllocation
}

// NewEmitter creates a new machine code emitter.
func NewEmitter(allocs map[uint32]RegAllocation) *Emitter {
	return &Emitter{
		Labels: make(map[uint32]int),
		allocs: allocs,
	}
}

// EmitFunction emits machine code for a complete function.
func (e *Emitter) EmitFunction(insts []MachInst, frame *StackFrame) {
	// Emit prologue
	for _, inst := range EmitPrologue(frame) {
		e.emitMachInst(inst)
	}

	// Emit body
	for _, inst := range insts {
		if inst.Op == MachRet {
			// Emit epilogue before returning
			for _, epInst := range EmitEpilogue(frame) {
				if epInst.Op != MachRet {
					e.emitMachInst(epInst)
				}
			}
			e.emit(EncodeRet())
		} else {
			e.emitMachInst(inst)
		}
	}

	// Resolve internal fixups
	e.resolveFixups()
}

// emitMachInst emits a single machine instruction.
func (e *Emitter) emitMachInst(inst MachInst) {
	switch inst.Op {
	case MachNop:
		e.emit(EncodeNop())

	case MachLabel:
		e.Labels[inst.Dst.Label] = len(e.Code)

	case MachRet:
		e.emit(EncodeRet())

	case MachMov:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		if dst != RegNone && src != RegNone {
			e.emit(EncodeMovRR(dst, src))
		}

	case MachMovImm:
		dst := e.resolveReg(inst.Dst)
		if dst != RegNone {
			imm := inst.Src1.Imm
			if imm >= -2147483648 && imm <= 2147483647 {
				e.emit(EncodeMovRI(dst, int32(imm)))
			} else {
				e.emit(EncodeMovRI64(dst, imm))
			}
		}

	case MachXorZero:
		dst := e.resolveReg(inst.Dst)
		if dst != RegNone {
			e.emit(EncodeXorZero(dst))
		}

	case MachAdd:
		dst := e.resolveReg(inst.Dst)
		if inst.Src1.Kind == OpndImm {
			e.emit(EncodeAddRI(dst, int32(inst.Src1.Imm)))
		} else {
			src := e.resolveReg(inst.Src1)
			e.emit(EncodeAddRR(dst, src))
		}

	case MachSub:
		dst := e.resolveReg(inst.Dst)
		if inst.Src1.Kind == OpndImm {
			e.emit(EncodeSubRI(dst, int32(inst.Src1.Imm)))
		} else {
			src := e.resolveReg(inst.Src1)
			e.emit(EncodeSubRR(dst, src))
		}

	case MachImul:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeImulRR(dst, src))

	case MachIdiv:
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeIdivR(src))

	case MachCqo:
		e.emit(EncodeCqo())

	case MachNeg:
		dst := e.resolveReg(inst.Dst)
		e.emit(EncodeNegR(dst))

	case MachNot:
		dst := e.resolveReg(inst.Dst)
		e.emit(EncodeNotR(dst))

	case MachAnd:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeAndRR(dst, src))

	case MachOr:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeOrRR(dst, src))

	case MachXor:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeXorRR(dst, src))

	case MachShl:
		dst := e.resolveReg(inst.Dst)
		e.emit(EncodeShlRCL(dst))

	case MachSar:
		dst := e.resolveReg(inst.Dst)
		e.emit(EncodeShrRCL(dst))

	case MachCmp:
		dst := e.resolveReg(inst.Dst)
		if inst.Src1.Kind == OpndImm {
			e.emit(EncodeCmpRI(dst, int32(inst.Src1.Imm)))
		} else {
			src := e.resolveReg(inst.Src1)
			e.emit(EncodeCmpRR(dst, src))
		}

	case MachTest:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeTestRR(dst, src))

	case MachSetCC:
		dst := e.resolveReg(inst.Dst)
		e.emit(EncodeSetCC(inst.CC, dst))

	case MachMovzxB:
		dst := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		e.emit(EncodeMovzxBR(dst, src))

	case MachPush:
		src := e.resolveReg(inst.Src1)
		e.emit(EncodePush(src))

	case MachPop:
		dst := e.resolveReg(inst.Dst)
		e.emit(EncodePop(dst))

	case MachJmp:
		if inst.Dst.Kind == OpndLabel {
			e.Fixups = append(e.Fixups, Fixup{
				Offset:   len(e.Code) + 1,
				LabelID:  inst.Dst.Label,
				InstSize: 5,
			})
			e.emit(EncodeJmpRel32(0)) // placeholder
		}

	case MachJcc:
		if inst.Dst.Kind == OpndLabel {
			e.Fixups = append(e.Fixups, Fixup{
				Offset:   len(e.Code) + 2,
				LabelID:  inst.Dst.Label,
				InstSize: 6,
			})
			e.emit(EncodeJccRel32(inst.CC, 0)) // placeholder
		}

	case MachCall:
		disp := int32(0)
		if inst.Src1.Imm == 0 {
			// Recursive call to current function (starts at 0)
			disp = -int32(len(e.Code) + 5)
		} else {
			// Record relocation for the call target
			e.Fixups = append(e.Fixups, Fixup{
				Offset:   len(e.Code) + 1,
				LabelID:  uint32(inst.Src1.Imm),
				InstSize: 5,
			})
		}
		e.emit(EncodeCallRel32(disp))

	case MachLoad:
		dst := e.resolveReg(inst.Dst)
		base := e.resolveReg(inst.Src1)
		disp := int32(0)
		if inst.Src2.Kind == OpndImm {
			disp = int32(inst.Src2.Imm)
		}
		e.emit(EncodeMovLoad(dst, base, disp))

	case MachStore:
		base := e.resolveReg(inst.Dst)
		src := e.resolveReg(inst.Src1)
		disp := int32(0)
		if inst.Src2.Kind == OpndImm {
			disp = int32(inst.Src2.Imm)
		}
		e.emit(EncodeMovStore(base, disp, src))
	}
}

// resolveReg maps a MachOperand to a physical register.
func (e *Emitter) resolveReg(op MachOperand) PhysReg {
	switch op.Kind {
	case OpndPhys:
		return op.Phys
	case OpndVReg:
		if alloc, ok := e.allocs[op.VReg]; ok {
			return alloc.Phys
		}
		return RegNone
	default:
		return RegNone
	}
}

// emit appends bytes to the code buffer.
func (e *Emitter) emit(code []byte) {
	e.Code = append(e.Code, code...)
}

// resolveFixups patches branch targets with the correct offsets.
func (e *Emitter) resolveFixups() {
	for _, fix := range e.Fixups {
		target, ok := e.Labels[fix.LabelID]
		if !ok {
			// external symbol — needs linker relocation
			e.Relocs = append(e.Relocs, Relocation{
				Offset:   fix.Offset,
				Kind:     RelocPC32,
				SymName:  fix.LabelID,
				Addend:   -4,
			})
			continue
		}

		// Compute PC-relative offset
		// offset = target - (fixup_position + 4)
		// The fixup position points to the rel32 field, instruction ends at fixup+4
		rel := int32(target - (fix.Offset + 4))

		// Patch the rel32 field
		e.Code[fix.Offset+0] = byte(rel)
		e.Code[fix.Offset+1] = byte(rel >> 8)
		e.Code[fix.Offset+2] = byte(rel >> 16)
		e.Code[fix.Offset+3] = byte(rel >> 24)
	}
}

// CodeSize returns the current size of emitted code.
func (e *Emitter) CodeSize() int {
	return len(e.Code)
}
