// Package air — opcode definitions for the AXIOM Intermediate Representation.
//
// Opcodes are organized into classes using hex prefixes:
//
//	0x00xx — Special (NOP)
//	0x01xx — Memory (alloc, load, store, GEP, etc.)
//	0x02xx — ALU (arithmetic, logic, comparison, conversion)
//	0x03xx — Control flow (jump, branch, call, return, phi)
//	0x04xx — SIMD (vector operations)
//	0x05xx — Comptime (compile-time evaluation)
package air

// ---------------------------------------------------------------------------
// Special
// ---------------------------------------------------------------------------

const OpNop Opcode = 0x0000

// ---------------------------------------------------------------------------
// Memory class (0x01xx)
// ---------------------------------------------------------------------------

const (
	OpAlloc      Opcode = 0x0101 // Dest = alloc(TypeID)
	OpFree       Opcode = 0x0102 // dealloc(Src1)
	OpLoad       Opcode = 0x0103 // Dest = *Src1
	OpStore      Opcode = 0x0104 // *Dest = Src1
	OpGEP        Opcode = 0x0105 // GetElementPointer: Dest = &Src1[Src2]
	OpCopy       Opcode = 0x0106 // Dest = copy(Src1)
	OpMove       Opcode = 0x0107 // Dest = move(Src1), source poisoned
	OpMakeRef    Opcode = 0x0108 // Dest = &Src1
	OpDeref      Opcode = 0x0109 // Dest = *Src1 (with ownership check)
	OpArenaAlloc Opcode = 0x010A // Dest = arena_alloc(Src1=arena, TypeID)
	OpDestroy    Opcode = 0x010B // destroy owned value Src1
	OpAliasReuse Opcode = 0x010C // reuse alias Src1 as Dest
	OpGetField   Opcode = 0x010D // Dest = Src1.field[Src2]
	OpSetField   Opcode = 0x010E // Src1.field[Src2] = Dest
	OpIndex      Opcode = 0x010F // Dest = Src1[Src2]
	OpSlice      Opcode = 0x0110 // Dest = Src1[Src2:Dest] (uses ExtraIdx for end)
)

// ---------------------------------------------------------------------------
// ALU class (0x02xx)
// ---------------------------------------------------------------------------

const (
	OpIConst Opcode = 0x0201 // load integer constant
	OpFConst Opcode = 0x0202 // load float constant
	OpIAdd   Opcode = 0x0203 // Dest = Src1 + Src2  (integer)
	OpISub   Opcode = 0x0204 // Dest = Src1 - Src2  (integer)
	OpIMul   Opcode = 0x0205 // Dest = Src1 * Src2  (integer)
	OpIDiv   Opcode = 0x0206 // Dest = Src1 / Src2  (integer)
	OpIMod   Opcode = 0x0207 // Dest = Src1 % Src2  (integer)
	OpFAdd   Opcode = 0x0208 // Dest = Src1 + Src2  (float)
	OpFSub   Opcode = 0x0209 // Dest = Src1 - Src2  (float)
	OpFMul   Opcode = 0x020A // Dest = Src1 * Src2  (float)
	OpFDiv   Opcode = 0x020B // Dest = Src1 / Src2  (float)
	OpEq     Opcode = 0x020C // Dest = Src1 == Src2
	OpNe     Opcode = 0x020D // Dest = Src1 != Src2
	OpLt     Opcode = 0x020E // Dest = Src1 < Src2
	OpLe     Opcode = 0x020F // Dest = Src1 <= Src2
	OpGt     Opcode = 0x0210 // Dest = Src1 > Src2
	OpGe     Opcode = 0x0211 // Dest = Src1 >= Src2
	OpAnd    Opcode = 0x0212 // Dest = Src1 & Src2  (bitwise)
	OpOr     Opcode = 0x0213 // Dest = Src1 | Src2  (bitwise)
	OpXor    Opcode = 0x0214 // Dest = Src1 ^ Src2
	OpShl    Opcode = 0x0215 // Dest = Src1 << Src2
	OpShr    Opcode = 0x0216 // Dest = Src1 >> Src2
	OpNot    Opcode = 0x0217 // Dest = ~Src1        (bitwise not)
	OpNeg    Opcode = 0x021D // Dest = -Src1
	OpIToF   Opcode = 0x021E // Dest = int-to-float(Src1)
	OpFToI   Opcode = 0x021F // Dest = float-to-int(Src1)
	OpZExt   Opcode = 0x0220 // Dest = zero-extend(Src1)
	OpSExt   Opcode = 0x0221 // Dest = sign-extend(Src1)
	OpTrunc  Opcode = 0x0222 // Dest = truncate(Src1)
	OpCast   Opcode = 0x0223 // Dest = cast(Src1) to TypeID
)

// ---------------------------------------------------------------------------
// Control class (0x03xx)
// ---------------------------------------------------------------------------

const (
	OpJump      Opcode = 0x0301 // jump to block Src1
	OpBranch    Opcode = 0x0302 // if Src1 jump Src2 else Dest
	OpCall      Opcode = 0x0303 // Dest = call Src1(args via Extras)
	OpReturn    Opcode = 0x0304 // return Src1
	OpPhi       Opcode = 0x0305 // SSA phi node
	OpLoopBegin Opcode = 0x0306 // marks loop header
	OpLoopEnd   Opcode = 0x0307 // marks loop exit
	OpSpawn     Opcode = 0x0308 // spawn actor with fn Src1
	OpSend      Opcode = 0x0309 // send Src1 to channel Src2
	OpRecv      Opcode = 0x030A // Dest = recv from channel Src1
	OpAwait     Opcode = 0x030B // Dest = await Src1
)

// ---------------------------------------------------------------------------
// SIMD class (0x04xx)
// ---------------------------------------------------------------------------

const (
	OpSIMDLoad  Opcode = 0x0401 // SIMD vector load
	OpSIMDStore Opcode = 0x0402 // SIMD vector store
	OpSIMDAdd   Opcode = 0x0403 // SIMD element-wise add
	OpSIMDMul   Opcode = 0x0404 // SIMD element-wise multiply
	OpSIMDFMA   Opcode = 0x0405 // SIMD fused multiply-add
)

// ---------------------------------------------------------------------------
// Comptime class (0x05xx)
// ---------------------------------------------------------------------------

const (
	OpComptime Opcode = 0x0501 // compile-time evaluation
)

// ---------------------------------------------------------------------------
// Backward compatibility aliases — old iota-based names map to new scheme.
// These allow incremental migration of downstream code.
// ---------------------------------------------------------------------------

const (
	OpcodeNop        = OpNop
	OpcodeConst      = OpIConst
	OpcodeAdd        = OpIAdd
	OpcodeSub        = OpISub
	OpcodeMul        = OpIMul
	OpcodeDiv        = OpIDiv
	OpcodeMod        = OpIMod
	OpcodeEq         = OpEq
	OpcodeNe         = OpNe
	OpcodeLt         = OpLt
	OpcodeLe         = OpLe
	OpcodeGt         = OpGt
	OpcodeGe         = OpGe
	OpcodeAnd        = OpAnd
	OpcodeOr         = OpOr
	OpcodeXor        = OpXor
	OpcodeShl        = OpShl
	OpcodeShr        = OpShr
	OpcodeNeg        = OpNeg
	OpcodeNot        = OpNot
	OpcodeLoad       = OpLoad
	OpcodeStore      = OpStore
	OpcodeAlloc      = OpAlloc
	OpcodeDealloc    = OpFree
	OpcodeCall       = OpCall
	OpcodeReturn     = OpReturn
	OpcodeJump       = OpJump
	OpcodeBranch     = OpBranch
	OpcodePhi        = OpPhi
	OpcodeGetField   = OpGetField
	OpcodeSetField   = OpSetField
	OpcodeIndex      = OpIndex
	OpcodeSlice      = OpSlice
	OpcodeCast       = OpCast
	OpcodeSpawn      = OpSpawn
	OpcodeAwait      = OpAwait
	OpcodeDestroyVal = OpDestroy
)

// OpcodeCount is kept for backward compatibility but is now meaningless
// with class-prefixed opcodes. Use Opcode.Class() instead.
const OpcodeCount Opcode = 0xFFFF

// ---------------------------------------------------------------------------
// Mnemonic table
// ---------------------------------------------------------------------------

var mnemonicTable = map[Opcode]string{
	OpNop:        "nop",
	OpAlloc:      "alloc",
	OpFree:       "free",
	OpLoad:       "load",
	OpStore:      "store",
	OpGEP:        "gep",
	OpCopy:       "copy",
	OpMove:       "move",
	OpMakeRef:    "mkref",
	OpDeref:      "deref",
	OpArenaAlloc: "aalloc",
	OpDestroy:    "destroy",
	OpAliasReuse: "areuse",
	OpGetField:   "getfld",
	OpSetField:   "setfld",
	OpIndex:      "index",
	OpSlice:      "slice",

	OpIConst: "iconst",
	OpFConst: "fconst",
	OpIAdd:   "iadd",
	OpISub:   "isub",
	OpIMul:   "imul",
	OpIDiv:   "idiv",
	OpIMod:   "imod",
	OpFAdd:   "fadd",
	OpFSub:   "fsub",
	OpFMul:   "fmul",
	OpFDiv:   "fdiv",
	OpEq:     "eq",
	OpNe:     "ne",
	OpLt:     "lt",
	OpLe:     "le",
	OpGt:     "gt",
	OpGe:     "ge",
	OpAnd:    "and",
	OpOr:     "or",
	OpXor:    "xor",
	OpShl:    "shl",
	OpShr:    "shr",
	OpNot:    "not",
	OpNeg:    "neg",
	OpIToF:   "itof",
	OpFToI:   "ftoi",
	OpZExt:   "zext",
	OpSExt:   "sext",
	OpTrunc:  "trunc",
	OpCast:   "cast",

	OpJump:      "jump",
	OpBranch:    "branch",
	OpCall:      "call",
	OpReturn:    "ret",
	OpPhi:       "phi",
	OpLoopBegin: "loopbeg",
	OpLoopEnd:   "loopend",
	OpSpawn:     "spawn",
	OpSend:      "send",
	OpRecv:      "recv",
	OpAwait:     "await",

	OpSIMDLoad:  "vload",
	OpSIMDStore: "vstore",
	OpSIMDAdd:   "vadd",
	OpSIMDMul:   "vmul",
	OpSIMDFMA:   "vfma",

	OpComptime: "comptime",
}

// ---------------------------------------------------------------------------
// Opcode methods
// ---------------------------------------------------------------------------

// Mnemonic returns a short lowercase name for this opcode (e.g. "iadd").
func (op Opcode) Mnemonic() string {
	if s, ok := mnemonicTable[op]; ok {
		return s
	}
	return "???"
}

// Class returns the high byte of the opcode, identifying its class.
// e.g. 0x01 for memory, 0x02 for ALU, 0x03 for control, etc.
func (op Opcode) Class() uint16 {
	return uint16(op) >> 8
}

// IsMemory returns true if the opcode belongs to the memory class (0x01xx).
func (op Opcode) IsMemory() bool {
	return op.Class() == 0x01
}

// IsALU returns true if the opcode belongs to the ALU class (0x02xx).
func (op Opcode) IsALU() bool {
	return op.Class() == 0x02
}

// IsControl returns true if the opcode belongs to the control-flow class (0x03xx).
func (op Opcode) IsControl() bool {
	return op.Class() == 0x03
}

// IsSIMD returns true if the opcode belongs to the SIMD class (0x04xx).
func (op Opcode) IsSIMD() bool {
	return op.Class() == 0x04
}

// IsTerminator returns true if this instruction terminates a basic block.
// Only OpJump, OpBranch, and OpReturn are terminators.
func (op Opcode) IsTerminator() bool {
	return op == OpJump || op == OpBranch || op == OpReturn
}

// IsBinaryALU returns true if this is a binary ALU operation (two source operands).
// This excludes unary ops (neg, not) and conversions (itof, ftoi, zext, sext, trunc).
func (op Opcode) IsBinaryALU() bool {
	switch op {
	case OpIAdd, OpISub, OpIMul, OpIDiv, OpIMod,
		OpFAdd, OpFSub, OpFMul, OpFDiv,
		OpEq, OpNe, OpLt, OpLe, OpGt, OpGe,
		OpAnd, OpOr, OpXor, OpShl, OpShr:
		return true
	default:
		return false
	}
}

// String returns the mnemonic for use in fmt.Println etc.
func (op Opcode) String() string {
	return op.Mnemonic()
}
