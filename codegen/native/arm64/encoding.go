package arm64

// --------------------------------------------------------------------------
// p13-t01: ARM64 Instruction Encoding
//
// Encodes ARM64 instructions into 32-bit fixed-width words.
// ARM64 uses a fixed 32-bit instruction size with structured encoding.
// --------------------------------------------------------------------------

import "encoding/binary"

// CondCode represents ARM64 condition codes.
type CondCode uint8

const (
	CondEQ CondCode = 0b0000 // Equal (Z=1)
	CondNE CondCode = 0b0001 // Not equal (Z=0)
	CondCS CondCode = 0b0010 // Carry set / unsigned >=
	CondCC CondCode = 0b0011 // Carry clear / unsigned <
	CondMI CondCode = 0b0100 // Minus / negative
	CondPL CondCode = 0b0101 // Plus / positive
	CondVS CondCode = 0b0110 // Overflow
	CondVC CondCode = 0b0111 // No overflow
	CondHI CondCode = 0b1000 // Unsigned >
	CondLS CondCode = 0b1001 // Unsigned <=
	CondGE CondCode = 0b1010 // Signed >=
	CondLT CondCode = 0b1011 // Signed <
	CondGT CondCode = 0b1100 // Signed >
	CondLE CondCode = 0b1101 // Signed <=
	CondAL CondCode = 0b1110 // Always
)

// String returns the condition code name.
func (cc CondCode) String() string {
	names := [16]string{
		"eq", "ne", "cs", "cc", "mi", "pl", "vs", "vc",
		"hi", "ls", "ge", "lt", "gt", "le", "al", "nv",
	}
	if cc < 16 {
		return names[cc]
	}
	return "??"
}

// ---------- Data Processing (register) ----------

// EncodeAddReg: ADD Xd, Xn, Xm (64-bit)
// [sf=1] [0 0 0 1 0 1 1] [shift=00] [0] [Rm] [imm6=000000] [Rn] [Rd]
func EncodeAddReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0x8B000000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeSubReg: SUB Xd, Xn, Xm
func EncodeSubReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0xCB000000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeMulReg: MUL Xd, Xn, Xm (alias for MADD Xd, Xn, Xm, XZR)
func EncodeMulReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0x9B007C00) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeSdivReg: SDIV Xd, Xn, Xm
func EncodeSdivReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0x9AC00C00) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeMsubReg: MSUB Xd, Xn, Xm, Xa (for modulo: Xa - Xn*Xm)
func EncodeMsubReg(rd, rn, rm, ra PhysReg) []byte {
	inst := uint32(0x9B008000) | uint32(rm.HWReg())<<16 | uint32(ra.HWReg())<<10 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeAndReg: AND Xd, Xn, Xm
func EncodeAndReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0x8A000000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeOrrReg: ORR Xd, Xn, Xm
func EncodeOrrReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0xAA000000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeEorReg: EOR Xd, Xn, Xm
func EncodeEorReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0xCA000000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeLslReg: LSL Xd, Xn, Xm (alias for LSLV)
func EncodeLslReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0x9AC02000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeAsrReg: ASR Xd, Xn, Xm (alias for ASRV)
func EncodeAsrReg(rd, rn, rm PhysReg) []byte {
	inst := uint32(0x9AC02800) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeNeg: NEG Xd, Xm (alias for SUB Xd, XZR, Xm)
func EncodeNeg(rd, rm PhysReg) []byte {
	return EncodeSubReg(rd, PhysReg(31), rm) // XZR = register 31
}

// ---------- Data Processing (immediate) ----------

// EncodeAddImm: ADD Xd, Xn, #imm12
func EncodeAddImm(rd, rn PhysReg, imm12 uint16) []byte {
	inst := uint32(0x91000000) | uint32(imm12&0xFFF)<<10 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeSubImm: SUB Xd, Xn, #imm12
func EncodeSubImm(rd, rn PhysReg, imm12 uint16) []byte {
	inst := uint32(0xD1000000) | uint32(imm12&0xFFF)<<10 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeMovz: MOVZ Xd, #imm16, LSL #shift
// shift must be 0, 16, 32, or 48.
func EncodeMovz(rd PhysReg, imm16 uint16, shift uint8) []byte {
	hw := uint32(shift / 16)
	inst := uint32(0xD2800000) | hw<<21 | uint32(imm16)<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeMovk: MOVK Xd, #imm16, LSL #shift
func EncodeMovk(rd PhysReg, imm16 uint16, shift uint8) []byte {
	hw := uint32(shift / 16)
	inst := uint32(0xF2800000) | hw<<21 | uint32(imm16)<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeMovReg: MOV Xd, Xn (alias for ORR Xd, XZR, Xn)
func EncodeMovReg(rd, rn PhysReg) []byte {
	return EncodeOrrReg(rd, PhysReg(31), rn) // XZR = register 31
}

// ---------- Comparison ----------

// EncodeCmpReg: CMP Xn, Xm (alias for SUBS XZR, Xn, Xm)
func EncodeCmpReg(rn, rm PhysReg) []byte {
	inst := uint32(0xEB000000) | uint32(rm.HWReg())<<16 | uint32(rn.HWReg())<<5 | uint32(31) // Rd=XZR
	return encode32(inst)
}

// EncodeCmpImm: CMP Xn, #imm12 (alias for SUBS XZR, Xn, #imm12)
func EncodeCmpImm(rn PhysReg, imm12 uint16) []byte {
	inst := uint32(0xF1000000) | uint32(imm12&0xFFF)<<10 | uint32(rn.HWReg())<<5 | uint32(31) // Rd=XZR
	return encode32(inst)
}

// EncodeCset: CSET Xd, cond (alias for CSINC Xd, XZR, XZR, !cond)
func EncodeCset(rd PhysReg, cc CondCode) []byte {
	invCond := cc ^ 1 // invert condition
	inst := uint32(0x9A9F07E0) | uint32(invCond)<<12 | uint32(rd.HWReg())
	return encode32(inst)
}

// ---------- Memory ----------

// EncodeLdrImm: LDR Xd, [Xn, #offset] (64-bit, unsigned offset)
func EncodeLdrImm(rd, rn PhysReg, offset int16) []byte {
	imm12 := uint32(uint16(offset/8)) & 0xFFF
	inst := uint32(0xF9400000) | imm12<<10 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeStrImm: STR Xd, [Xn, #offset] (64-bit, unsigned offset)
func EncodeStrImm(rd, rn PhysReg, offset int16) []byte {
	imm12 := uint32(uint16(offset/8)) & 0xFFF
	inst := uint32(0xF9000000) | imm12<<10 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeLdrPreIdx: LDR Xd, [Xn, #simm9]! (pre-indexed)
func EncodeLdrPreIdx(rd, rn PhysReg, simm9 int16) []byte {
	imm9 := uint32(uint16(simm9) & 0x1FF)
	inst := uint32(0xF8400C00) | imm9<<12 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeStrPreIdx: STR Xd, [Xn, #simm9]! (pre-indexed)
func EncodeStrPreIdx(rd, rn PhysReg, simm9 int16) []byte {
	imm9 := uint32(uint16(simm9) & 0x1FF)
	inst := uint32(0xF8000C00) | imm9<<12 | uint32(rn.HWReg())<<5 | uint32(rd.HWReg())
	return encode32(inst)
}

// EncodeStp: STP Xt1, Xt2, [Xn, #offset]! (pre-indexed, 64-bit pair)
func EncodeStp(rt1, rt2, rn PhysReg, offset int16) []byte {
	imm7 := uint32(uint16(offset/8) & 0x7F)
	inst := uint32(0xA9800000) | imm7<<15 | uint32(rt2.HWReg())<<10 | uint32(rn.HWReg())<<5 | uint32(rt1.HWReg())
	return encode32(inst)
}

// EncodeLdp: LDP Xt1, Xt2, [Xn], #offset (post-indexed, 64-bit pair)
func EncodeLdp(rt1, rt2, rn PhysReg, offset int16) []byte {
	imm7 := uint32(uint16(offset/8) & 0x7F)
	inst := uint32(0xA8C00000) | imm7<<15 | uint32(rt2.HWReg())<<10 | uint32(rn.HWReg())<<5 | uint32(rt1.HWReg())
	return encode32(inst)
}

// ---------- Branch ----------

// EncodeB: B label (unconditional branch, 26-bit signed offset)
func EncodeB(offset int32) []byte {
	imm26 := uint32(offset/4) & 0x03FFFFFF
	inst := uint32(0x14000000) | imm26
	return encode32(inst)
}

// EncodeBCond: B.cond label (conditional branch, 19-bit signed offset)
func EncodeBCond(cc CondCode, offset int32) []byte {
	imm19 := uint32(offset/4) & 0x7FFFF
	inst := uint32(0x54000000) | imm19<<5 | uint32(cc)
	return encode32(inst)
}

// EncodeBl: BL label (branch with link, 26-bit signed offset)
func EncodeBl(offset int32) []byte {
	imm26 := uint32(offset/4) & 0x03FFFFFF
	inst := uint32(0x94000000) | imm26
	return encode32(inst)
}

// EncodeBlr: BLR Xn (branch with link to register)
func EncodeBlr(rn PhysReg) []byte {
	inst := uint32(0xD63F0000) | uint32(rn.HWReg())<<5
	return encode32(inst)
}

// EncodeRet: RET {Xn} (default X30/LR)
func EncodeRet() []byte {
	inst := uint32(0xD65F03C0) // RET X30
	return encode32(inst)
}

// EncodeNop: NOP
func EncodeNop() []byte {
	return encode32(0xD503201F)
}

// EncodeCbz: CBZ Xn, label (compare and branch if zero)
func EncodeCbz(rn PhysReg, offset int32) []byte {
	imm19 := uint32(offset/4) & 0x7FFFF
	inst := uint32(0xB4000000) | imm19<<5 | uint32(rn.HWReg())
	return encode32(inst)
}

// EncodeCbnz: CBNZ Xn, label (compare and branch if not zero)
func EncodeCbnz(rn PhysReg, offset int32) []byte {
	imm19 := uint32(offset/4) & 0x7FFFF
	inst := uint32(0xB5000000) | imm19<<5 | uint32(rn.HWReg())
	return encode32(inst)
}

// ---------- Helpers ----------

func encode32(inst uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, inst)
	return buf
}
