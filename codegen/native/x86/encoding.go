package x86

// --------------------------------------------------------------------------
// p11-t02: x86-64 Instruction Encoding Tables
//
// Provides functions to encode individual x86-64 instructions into
// machine code bytes. Each function returns a byte slice containing
// the complete encoded instruction including prefixes, opcode, ModRM,
// SIB, and immediate fields.
// --------------------------------------------------------------------------

// EncodeRet encodes a RET instruction (0xC3).
func EncodeRet() []byte {
	return []byte{0xC3}
}

// EncodeNop encodes a 1-byte NOP instruction (0x90).
func EncodeNop() []byte {
	return []byte{0x90}
}

// EncodeInt3 encodes an INT3 breakpoint (0xCC).
func EncodeInt3() []byte {
	return []byte{0xCC}
}

// EncodePush encodes PUSH reg (64-bit).
func EncodePush(reg PhysReg) []byte {
	if reg.NeedsREX() {
		return []byte{REXBase | REXB, 0x50 + reg.RegField()}
	}
	return []byte{0x50 + reg.RegField()}
}

// EncodePop encodes POP reg (64-bit).
func EncodePop(reg PhysReg) []byte {
	if reg.NeedsREX() {
		return []byte{REXBase | REXB, 0x58 + reg.RegField()}
	}
	return []byte{0x58 + reg.RegField()}
}

// EncodeMovRR encodes MOV dst, src (64-bit register-to-register).
// REX.W + 0x89 /r (MOV r/m64, r64)
func EncodeMovRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
		return []byte{rex, 0x89, modrm}
	}
	return []byte{EncodeREX(true, false, false, false), 0x89, modrm}
}

// EncodeMovRI encodes MOV reg, imm32 (sign-extended to 64-bit).
// REX.W + 0xC7 /0 id (MOV r/m64, imm32)
func EncodeMovRI(dst PhysReg, imm32 int32) []byte {
	rex := EncodeREX(true, false, false, dst.NeedsREX())
	modrm := (ModRegDirect << 6) | (0 << 3) | dst.RegField()
	return []byte{
		rex, 0xC7, modrm,
		byte(imm32), byte(imm32 >> 8), byte(imm32 >> 16), byte(imm32 >> 24),
	}
}

// EncodeMovRI64 encodes MOV reg, imm64 (full 64-bit immediate).
// REX.W + 0xB8+rd io (MOV r64, imm64)
func EncodeMovRI64(dst PhysReg, imm64 int64) []byte {
	rex := EncodeREX(true, false, false, dst.NeedsREX())
	buf := []byte{rex, 0xB8 + dst.RegField()}
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(imm64>>(i*8)))
	}
	return buf
}

// EncodeAddRR encodes ADD dst, src (64-bit).
// REX.W + 0x01 /r
func EncodeAddRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x01, modrm}
}

// EncodeAddRI encodes ADD reg, imm32 (64-bit).
// REX.W + 0x81 /0 id
func EncodeAddRI(dst PhysReg, imm32 int32) []byte {
	rex := EncodeREX(true, false, false, dst.NeedsREX())
	modrm := (ModRegDirect << 6) | (0 << 3) | dst.RegField()
	return []byte{
		rex, 0x81, modrm,
		byte(imm32), byte(imm32 >> 8), byte(imm32 >> 16), byte(imm32 >> 24),
	}
}

// EncodeSubRR encodes SUB dst, src (64-bit).
// REX.W + 0x29 /r
func EncodeSubRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x29, modrm}
}

// EncodeSubRI encodes SUB reg, imm32 (64-bit).
// REX.W + 0x81 /5 id
func EncodeSubRI(dst PhysReg, imm32 int32) []byte {
	rex := EncodeREX(true, false, false, dst.NeedsREX())
	modrm := (ModRegDirect << 6) | (5 << 3) | dst.RegField()
	return []byte{
		rex, 0x81, modrm,
		byte(imm32), byte(imm32 >> 8), byte(imm32 >> 16), byte(imm32 >> 24),
	}
}

// EncodeImulRR encodes IMUL dst, src (64-bit signed multiply).
// REX.W + 0x0F 0xAF /r
func EncodeImulRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(dst, src)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x0F, 0xAF, modrm}
}

// EncodeCqo encodes CQO (sign-extend RAX into RDX:RAX).
// REX.W + 0x99
func EncodeCqo() []byte {
	return []byte{EncodeREX(true, false, false, false), 0x99}
}

// EncodeIdivR encodes IDIV reg (signed division: RDX:RAX / reg → RAX, RDX).
// REX.W + 0xF7 /7
func EncodeIdivR(divisor PhysReg) []byte {
	rex := EncodeREX(true, false, false, divisor.NeedsREX())
	modrm := (ModRegDirect << 6) | (7 << 3) | divisor.RegField()
	return []byte{rex, 0xF7, modrm}
}

// EncodeNegR encodes NEG reg (64-bit two's complement negation).
// REX.W + 0xF7 /3
func EncodeNegR(reg PhysReg) []byte {
	rex := EncodeREX(true, false, false, reg.NeedsREX())
	modrm := (ModRegDirect << 6) | (3 << 3) | reg.RegField()
	return []byte{rex, 0xF7, modrm}
}

// EncodeNotR encodes NOT reg (64-bit bitwise NOT).
// REX.W + 0xF7 /2
func EncodeNotR(reg PhysReg) []byte {
	rex := EncodeREX(true, false, false, reg.NeedsREX())
	modrm := (ModRegDirect << 6) | (2 << 3) | reg.RegField()
	return []byte{rex, 0xF7, modrm}
}

// EncodeCmpRR encodes CMP dst, src (64-bit comparison).
// REX.W + 0x39 /r
func EncodeCmpRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x39, modrm}
}

// EncodeCmpRI encodes CMP reg, imm32 (64-bit).
// REX.W + 0x81 /7 id
func EncodeCmpRI(reg PhysReg, imm32 int32) []byte {
	rex := EncodeREX(true, false, false, reg.NeedsREX())
	modrm := (ModRegDirect << 6) | (7 << 3) | reg.RegField()
	return []byte{
		rex, 0x81, modrm,
		byte(imm32), byte(imm32 >> 8), byte(imm32 >> 16), byte(imm32 >> 24),
	}
}

// EncodeTestRR encodes TEST dst, src (64-bit AND test, sets flags).
// REX.W + 0x85 /r
func EncodeTestRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x85, modrm}
}

// EncodeAndRR encodes AND dst, src (64-bit).
// REX.W + 0x21 /r
func EncodeAndRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x21, modrm}
}

// EncodeOrRR encodes OR dst, src (64-bit).
// REX.W + 0x09 /r
func EncodeOrRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x09, modrm}
}

// EncodeXorRR encodes XOR dst, src (64-bit).
// REX.W + 0x31 /r
func EncodeXorRR(dst, src PhysReg) []byte {
	modrm, rex, needREX := EncodeModRM_RR(src, dst)
	if needREX {
		rex |= REXW
	} else {
		rex = EncodeREX(true, false, false, false)
	}
	return []byte{rex, 0x31, modrm}
}

// EncodeShlRCL encodes SHL reg, CL (shift left by CL amount).
// REX.W + 0xD3 /4
func EncodeShlRCL(reg PhysReg) []byte {
	rex := EncodeREX(true, false, false, reg.NeedsREX())
	modrm := (ModRegDirect << 6) | (4 << 3) | reg.RegField()
	return []byte{rex, 0xD3, modrm}
}

// EncodeShrRCL encodes SAR reg, CL (arithmetic shift right by CL amount).
// REX.W + 0xD3 /7
func EncodeShrRCL(reg PhysReg) []byte {
	rex := EncodeREX(true, false, false, reg.NeedsREX())
	modrm := (ModRegDirect << 6) | (7 << 3) | reg.RegField()
	return []byte{rex, 0xD3, modrm}
}

// EncodeXorZero encodes XOR reg, reg (zero a 32-bit register, implicitly zero-extends).
// 0x31 /r (no REX.W — 32-bit clears upper 32)
func EncodeXorZero(reg PhysReg) []byte {
	modrm := (ModRegDirect << 6) | (reg.RegField() << 3) | reg.RegField()
	if reg.NeedsREX() {
		rex := REXBase | REXR | REXB
		return []byte{rex, 0x31, modrm}
	}
	return []byte{0x31, modrm}
}

// EncodeSetCC encodes SETcc r/m8 (set byte on condition).
// 0x0F 0x9X /0 where X is the condition code.
func EncodeSetCC(cc CondCode, dst PhysReg) []byte {
	var buf []byte
	if dst.NeedsREX() || dst.HWReg() >= 4 {
		// Need REX to access SPL/BPL/SIL/DIL or R8B+
		buf = append(buf, REXBase)
		if dst.NeedsREX() {
			buf[0] |= REXB
		}
	}
	modrm := (ModRegDirect << 6) | (0 << 3) | dst.RegField()
	buf = append(buf, 0x0F, 0x90+uint8(cc), modrm)
	return buf
}

// EncodeMovzxBR encodes MOVZX r64, r/m8 (zero-extend byte to 64-bit).
// REX.W + 0x0F 0xB6 /r
func EncodeMovzxBR(dst, src PhysReg) []byte {
	rex := EncodeREX(true, dst.NeedsREX(), false, src.NeedsREX())
	modrm := (ModRegDirect << 6) | (dst.RegField() << 3) | src.RegField()
	return []byte{rex, 0x0F, 0xB6, modrm}
}

// EncodeJmpRel32 encodes JMP rel32 (near jump, 5 bytes).
// 0xE9 cd
func EncodeJmpRel32(rel32 int32) []byte {
	return []byte{
		0xE9,
		byte(rel32), byte(rel32 >> 8), byte(rel32 >> 16), byte(rel32 >> 24),
	}
}

// EncodeJccRel32 encodes Jcc rel32 (conditional jump near, 6 bytes).
// 0x0F 0x8X cd
func EncodeJccRel32(cc CondCode, rel32 int32) []byte {
	return []byte{
		0x0F, 0x80 + uint8(cc),
		byte(rel32), byte(rel32 >> 8), byte(rel32 >> 16), byte(rel32 >> 24),
	}
}

// EncodeCallRel32 encodes CALL rel32 (near call, 5 bytes).
// 0xE8 cd
func EncodeCallRel32(rel32 int32) []byte {
	return []byte{
		0xE8,
		byte(rel32), byte(rel32 >> 8), byte(rel32 >> 16), byte(rel32 >> 24),
	}
}

// EncodeCallR encodes CALL reg (indirect call through register).
// 0xFF /2
func EncodeCallR(reg PhysReg) []byte {
	var buf []byte
	if reg.NeedsREX() {
		buf = append(buf, REXBase|REXB)
	}
	modrm := (ModRegDirect << 6) | (2 << 3) | reg.RegField()
	buf = append(buf, 0xFF, modrm)
	return buf
}

// EncodeLea encodes LEA dst, [base + disp32] (64-bit).
// REX.W + 0x8D /r
func EncodeLea(dst, base PhysReg, disp int32) []byte {
	rex := EncodeREX(true, dst.NeedsREX(), false, base.NeedsREX())
	buf := []byte{rex, 0x8D}
	modrmBuf := EncodeModRM_RM(dst, base, disp)
	// EncodeModRM_RM may prepend its own REX — strip it if present
	if len(modrmBuf) > 0 && (modrmBuf[0]&0xF0) == 0x40 {
		// Merge REX bits
		buf[0] |= modrmBuf[0] & 0x0F
		modrmBuf = modrmBuf[1:]
	}
	buf = append(buf, modrmBuf...)
	return buf
}

// EncodeMovLoad encodes MOV dst, [base + disp32] (64-bit load).
// REX.W + 0x8B /r
func EncodeMovLoad(dst, base PhysReg, disp int32) []byte {
	rex := EncodeREX(true, dst.NeedsREX(), false, base.NeedsREX())
	buf := []byte{rex, 0x8B}
	modrmBuf := EncodeModRM_RM(dst, base, disp)
	if len(modrmBuf) > 0 && (modrmBuf[0]&0xF0) == 0x40 {
		buf[0] |= modrmBuf[0] & 0x0F
		modrmBuf = modrmBuf[1:]
	}
	buf = append(buf, modrmBuf...)
	return buf
}

// EncodeMovStore encodes MOV [base + disp32], src (64-bit store).
// REX.W + 0x89 /r
func EncodeMovStore(base PhysReg, disp int32, src PhysReg) []byte {
	rex := EncodeREX(true, src.NeedsREX(), false, base.NeedsREX())
	buf := []byte{rex, 0x89}
	modrmBuf := EncodeModRM_RM(src, base, disp)
	if len(modrmBuf) > 0 && (modrmBuf[0]&0xF0) == 0x40 {
		buf[0] |= modrmBuf[0] & 0x0F
		modrmBuf = modrmBuf[1:]
	}
	buf = append(buf, modrmBuf...)
	return buf
}

// CondCode represents x86 condition codes for Jcc/SETcc/CMOVcc.
type CondCode uint8

const (
	CCO   CondCode = 0x00 // Overflow
	CCNo  CondCode = 0x01 // No overflow
	CCB   CondCode = 0x02 // Below (CF=1) — unsigned <
	CCAE  CondCode = 0x03 // Above or equal (CF=0) — unsigned >=
	CCE   CondCode = 0x04 // Equal (ZF=1)
	CCNE  CondCode = 0x05 // Not equal (ZF=0)
	CCBE  CondCode = 0x06 // Below or equal (CF=1 or ZF=1) — unsigned <=
	CCA   CondCode = 0x07 // Above (CF=0 and ZF=0) — unsigned >
	CCS   CondCode = 0x08 // Sign (SF=1)
	CCNS  CondCode = 0x09 // No sign (SF=0)
	CCPE  CondCode = 0x0A // Parity even
	CCPO  CondCode = 0x0B // Parity odd
	CCL   CondCode = 0x0C // Less (SF≠OF) — signed <
	CCGE  CondCode = 0x0D // Greater or equal (SF=OF) — signed >=
	CCLE  CondCode = 0x0E // Less or equal (ZF=1 or SF≠OF) — signed <=
	CCG   CondCode = 0x0F // Greater (ZF=0 and SF=OF) — signed >
)

// String returns the condition code mnemonic.
func (cc CondCode) String() string {
	names := [16]string{
		"o", "no", "b", "ae", "e", "ne", "be", "a",
		"s", "ns", "pe", "po", "l", "ge", "le", "g",
	}
	if cc < 16 {
		return names[cc]
	}
	return "??"
}
