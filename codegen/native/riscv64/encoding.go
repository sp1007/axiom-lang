package riscv64

import "encoding/binary"

// --------------------------------------------------------------------------
// p13-t05: RISC-V 64-bit Instruction Encoding
//
// RV64I base instruction encodings. RISC-V uses 6 instruction formats:
// R-type, I-type, S-type, B-type, U-type, J-type. All 32 bits wide.
// --------------------------------------------------------------------------

// R-type: [funct7:7][rs2:5][rs1:5][funct3:3][rd:5][opcode:7]
func encodeR(opcode, funct3, funct7 uint32, rd, rs1, rs2 PhysReg) []byte {
	inst := opcode | uint32(rd.HWReg())<<7 | funct3<<12 |
		uint32(rs1.HWReg())<<15 | uint32(rs2.HWReg())<<20 | funct7<<25
	return encode32(inst)
}

// I-type: [imm:12][rs1:5][funct3:3][rd:5][opcode:7]
func encodeI(opcode, funct3 uint32, rd, rs1 PhysReg, imm12 int16) []byte {
	inst := opcode | uint32(rd.HWReg())<<7 | funct3<<12 |
		uint32(rs1.HWReg())<<15 | (uint32(imm12)&0xFFF)<<20
	return encode32(inst)
}

// S-type: [imm:12 split][rs2:5][rs1:5][funct3:3][imm:5][opcode:7]
func encodeS(opcode, funct3 uint32, rs1, rs2 PhysReg, imm12 int16) []byte {
	imm := uint32(imm12) & 0xFFF
	inst := opcode | (imm&0x1F)<<7 | funct3<<12 |
		uint32(rs1.HWReg())<<15 | uint32(rs2.HWReg())<<20 | (imm>>5)<<25
	return encode32(inst)
}

// B-type: conditional branch [imm:13 split]
func encodeB(funct3 uint32, rs1, rs2 PhysReg, offset int32) []byte {
	imm := uint32(offset) & 0x1FFE // bits [12:1]
	inst := uint32(0x63) |        // opcode = BRANCH
		((imm>>11)&1)<<7 |
		((imm>>1)&0xF)<<8 |
		funct3<<12 |
		uint32(rs1.HWReg())<<15 |
		uint32(rs2.HWReg())<<20 |
		((imm>>5)&0x3F)<<25 |
		((imm>>12)&1)<<31
	return encode32(inst)
}

// U-type: [imm:20][rd:5][opcode:7]
func encodeU(opcode uint32, rd PhysReg, imm20 uint32) []byte {
	inst := opcode | uint32(rd.HWReg())<<7 | (imm20 << 12)
	return encode32(inst)
}

// J-type: [imm:21 split][rd:5][opcode:7]
func encodeJ(opcode uint32, rd PhysReg, offset int32) []byte {
	imm := uint32(offset) & 0x1FFFFF
	inst := opcode | uint32(rd.HWReg())<<7 |
		((imm>>12)&0xFF)<<12 |
		((imm>>11)&1)<<20 |
		((imm>>1)&0x3FF)<<21 |
		((imm>>20)&1)<<31
	return encode32(inst)
}

// ---------- RV64I Instructions ----------

// EncodeAdd: ADD rd, rs1, rs2
func EncodeAdd(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 0, 0, rd, rs1, rs2) }

// EncodeSub: SUB rd, rs1, rs2
func EncodeSub(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 0, 0x20, rd, rs1, rs2) }

// EncodeAnd: AND rd, rs1, rs2
func EncodeAnd(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 7, 0, rd, rs1, rs2) }

// EncodeOr: OR rd, rs1, rs2
func EncodeOr(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 6, 0, rd, rs1, rs2) }

// EncodeXor: XOR rd, rs1, rs2
func EncodeXor(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 4, 0, rd, rs1, rs2) }

// EncodeSll: SLL rd, rs1, rs2 (shift left logical)
func EncodeSll(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 1, 0, rd, rs1, rs2) }

// EncodeSra: SRA rd, rs1, rs2 (shift right arithmetic)
func EncodeSra(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 5, 0x20, rd, rs1, rs2) }

// EncodeSlt: SLT rd, rs1, rs2 (set less than)
func EncodeSlt(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 2, 0, rd, rs1, rs2) }

// RV64M multiply/divide extension
func EncodeMul(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 0, 1, rd, rs1, rs2) }
func EncodeDiv(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 4, 1, rd, rs1, rs2) }
func EncodeRem(rd, rs1, rs2 PhysReg) []byte { return encodeR(0x33, 6, 1, rd, rs1, rs2) }

// I-type instructions
func EncodeAddi(rd, rs1 PhysReg, imm int16) []byte { return encodeI(0x13, 0, rd, rs1, imm) }
func EncodeAndi(rd, rs1 PhysReg, imm int16) []byte { return encodeI(0x13, 7, rd, rs1, imm) }
func EncodeOri(rd, rs1 PhysReg, imm int16) []byte  { return encodeI(0x13, 6, rd, rs1, imm) }
func EncodeXori(rd, rs1 PhysReg, imm int16) []byte  { return encodeI(0x13, 4, rd, rs1, imm) }

// Load/Store (64-bit)
func EncodeLd(rd, rs1 PhysReg, offset int16) []byte { return encodeI(0x03, 3, rd, rs1, offset) }
func EncodeSd(rs2, rs1 PhysReg, offset int16) []byte { return encodeS(0x23, 3, rs1, rs2, offset) }

// Branches
func EncodeBeq(rs1, rs2 PhysReg, offset int32) []byte { return encodeB(0, rs1, rs2, offset) }
func EncodeBne(rs1, rs2 PhysReg, offset int32) []byte { return encodeB(1, rs1, rs2, offset) }
func EncodeBlt(rs1, rs2 PhysReg, offset int32) []byte { return encodeB(4, rs1, rs2, offset) }
func EncodeBge(rs1, rs2 PhysReg, offset int32) []byte { return encodeB(5, rs1, rs2, offset) }

// Jumps
func EncodeJal(rd PhysReg, offset int32) []byte   { return encodeJ(0x6F, rd, offset) }
func EncodeJalr(rd, rs1 PhysReg, offset int16) []byte { return encodeI(0x67, 0, rd, rs1, offset) }

// LUI: LUI rd, imm20
func EncodeLui(rd PhysReg, imm20 uint32) []byte { return encodeU(0x37, rd, imm20) }

// AUIPC: AUIPC rd, imm20
func EncodeAuipc(rd PhysReg, imm20 uint32) []byte { return encodeU(0x17, rd, imm20) }

// Pseudo instructions
func EncodeNop() []byte    { return EncodeAddi(Zero, Zero, 0) }
func EncodeRet() []byte    { return EncodeJalr(Zero, RA, 0) }
func EncodeLi(rd PhysReg, imm int32) []byte {
	if imm >= -2048 && imm <= 2047 {
		return EncodeAddi(rd, Zero, int16(imm))
	}
	// LUI + ADDI for larger immediates
	upper := uint32(imm+0x800) >> 12
	lower := int16(imm - int32(upper<<12))
	result := EncodeLui(rd, upper)
	if lower != 0 {
		result = append(result, EncodeAddi(rd, rd, lower)...)
	}
	return result
}

func EncodeCall(offset int32) []byte { return EncodeJal(RA, offset) }

func encode32(inst uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, inst)
	return buf
}
