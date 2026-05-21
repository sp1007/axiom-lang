package x86

// --------------------------------------------------------------------------
// p11-t17: ModRM/SIB Encoding Library
//
// Implements the x86-64 ModRM and SIB byte encoding, including REX prefix
// generation. These are the fundamental building blocks for all instruction
// encoding in the x86 backend.
//
// ModRM byte layout: [mod:2][reg:3][rm:3]
// SIB byte layout:   [scale:2][index:3][base:3]
// REX byte layout:   [0100][W][R][X][B]
// --------------------------------------------------------------------------

// ModRM mode values
const (
	ModIndirect   uint8 = 0b00 // [reg] — register indirect
	ModDisp8      uint8 = 0b01 // [reg + disp8]
	ModDisp32     uint8 = 0b10 // [reg + disp32]
	ModRegDirect  uint8 = 0b11 // reg — register direct
)

// REX prefix base and bit positions
const (
	REXBase uint8 = 0x40 // 0100xxxx
	REXW    uint8 = 0x08 // W bit: 64-bit operand size
	REXR    uint8 = 0x04 // R bit: extends ModRM.reg
	REXX    uint8 = 0x02 // X bit: extends SIB.index
	REXB    uint8 = 0x01 // B bit: extends ModRM.rm or SIB.base
)

// EncodeREX builds a REX prefix byte.
// w=true for 64-bit operand size, r/x/b extend register fields.
func EncodeREX(w, r, x, b bool) byte {
	rex := REXBase
	if w {
		rex |= REXW
	}
	if r {
		rex |= REXR
	}
	if x {
		rex |= REXX
	}
	if b {
		rex |= REXB
	}
	return rex
}

// NeedsREXW returns true if a REX.W prefix is needed for 64-bit operations.
func NeedsREXW() bool {
	return true // All 64-bit operations need REX.W
}

// EncodeModRM_RR encodes a ModRM byte for register-to-register (mod=11).
// reg is the /r field, rm is the R/M field.
func EncodeModRM_RR(reg, rm PhysReg) (modrm byte, rex byte, needREX bool) {
	modrm = (ModRegDirect << 6) | (reg.RegField() << 3) | rm.RegField()
	needREX = reg.NeedsREX() || rm.NeedsREX()
	if needREX {
		rex = REXBase
		if reg.NeedsREX() {
			rex |= REXR
		}
		if rm.NeedsREX() {
			rex |= REXB
		}
	}
	return
}

// EncodeModRM_RM encodes a ModRM byte for register + memory [base+disp].
// reg is the /r field, base is the R/M base register, disp is displacement.
//
// Special cases handled:
// - RSP/R12 as base → emit SIB byte (RSP encoding conflicts with SIB marker)
// - RBP/R13 as base with disp=0 → use disp8=0 (RBP encoding conflicts with RIP-relative)
func EncodeModRM_RM(reg, base PhysReg, disp int32) (buf []byte) {
	needsSIB := base.RegField() == RSP.RegField() // RSP/R12
	needsDisp8ForRBP := base.RegField() == RBP.RegField() && disp == 0

	// Determine mod field
	var mod uint8
	switch {
	case needsDisp8ForRBP:
		mod = ModDisp8 // force disp8=0 for RBP/R13
	case disp == 0:
		mod = ModIndirect
	case disp >= -128 && disp <= 127:
		mod = ModDisp8
	default:
		mod = ModDisp32
	}

	// REX prefix
	rex := byte(0)
	if reg.NeedsREX() || base.NeedsREX() {
		rex = REXBase
		if reg.NeedsREX() {
			rex |= REXR
		}
		if base.NeedsREX() {
			rex |= REXB
		}
	}

	if rex != 0 {
		buf = append(buf, rex)
	}

	if needsSIB {
		// ModRM with SIB: rm=100 (SIB follows)
		modrm := (mod << 6) | (reg.RegField() << 3) | 0x04
		sib := (0 << 6) | (0x04 << 3) | base.RegField() // scale=1, index=RSP(none), base=base
		buf = append(buf, modrm, sib)
	} else {
		modrm := (mod << 6) | (reg.RegField() << 3) | base.RegField()
		buf = append(buf, modrm)
	}

	// Displacement
	switch mod {
	case ModDisp8:
		buf = append(buf, byte(int8(disp)))
	case ModDisp32:
		buf = append(buf, byte(disp), byte(disp>>8), byte(disp>>16), byte(disp>>24))
	}

	return buf
}

// EncodeModRM_RIP encodes a ModRM byte for RIP-relative addressing.
// mod=00, rm=101 indicates RIP-relative with 32-bit displacement.
func EncodeModRM_RIP(reg PhysReg, disp32 int32) (buf []byte) {
	rex := byte(0)
	if reg.NeedsREX() {
		rex = REXBase | REXR
	}
	if rex != 0 {
		buf = append(buf, rex)
	}

	modrm := (ModIndirect << 6) | (reg.RegField() << 3) | 0x05
	buf = append(buf, modrm)
	buf = append(buf, byte(disp32), byte(disp32>>8), byte(disp32>>16), byte(disp32>>24))
	return buf
}

// EncodeModRM_SIB encodes a ModRM+SIB for [base + index*scale + disp].
func EncodeModRM_SIB(reg, base, index PhysReg, scale uint8, disp int32) (buf []byte) {
	// Scale encoding: 1→0, 2→1, 4→2, 8→3
	var scaleBits uint8
	switch scale {
	case 1:
		scaleBits = 0
	case 2:
		scaleBits = 1
	case 4:
		scaleBits = 2
	case 8:
		scaleBits = 3
	default:
		scaleBits = 0
	}

	// Determine mod
	var mod uint8
	needsDisp8ForRBP := base.RegField() == RBP.RegField() && disp == 0
	switch {
	case needsDisp8ForRBP:
		mod = ModDisp8
	case disp == 0:
		mod = ModIndirect
	case disp >= -128 && disp <= 127:
		mod = ModDisp8
	default:
		mod = ModDisp32
	}

	// REX
	rex := byte(0)
	if reg.NeedsREX() || base.NeedsREX() || index.NeedsREX() {
		rex = REXBase
		if reg.NeedsREX() {
			rex |= REXR
		}
		if index.NeedsREX() {
			rex |= REXX
		}
		if base.NeedsREX() {
			rex |= REXB
		}
	}

	if rex != 0 {
		buf = append(buf, rex)
	}

	modrm := (mod << 6) | (reg.RegField() << 3) | 0x04 // rm=100 → SIB follows
	sib := (scaleBits << 6) | (index.RegField() << 3) | base.RegField()
	buf = append(buf, modrm, sib)

	switch mod {
	case ModDisp8:
		buf = append(buf, byte(int8(disp)))
	case ModDisp32:
		buf = append(buf, byte(disp), byte(disp>>8), byte(disp>>16), byte(disp>>24))
	}

	return buf
}
