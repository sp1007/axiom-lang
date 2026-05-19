// Package air defines the AXIOM Intermediate Representation instruction set.
// AIR is a typed, block-structured SSA-form IR used as the primary
// optimization and lowering representation in the AXIOM compiler pipeline.
package air

import "unsafe"

// AirInst is one instruction in the AXIOM Intermediate Representation.
// Layout is FROZEN at 16 bytes. Do not add fields without an RFC.
//
// AIR is in SSA form. Dest, Src1, Src2 are value IDs (indices into
// the function's value table, not memory addresses).
// TypeID indexes into the global TypeTable.
//
// Field layout (16 bytes):
//
//	Opcode  uint16  2B  @ offset 0
//	TypeID  uint16  2B  @ offset 2
//	Dest    uint32  4B  @ offset 4
//	Src1    uint32  4B  @ offset 8
//	Src2    uint32  4B  @ offset 12
//
// FROZEN: do not modify without RFC
type AirInst struct {
	Opcode Opcode // instruction opcode
	TypeID uint16 // result type (index into TypeTable)
	Dest   uint32 // destination value ID (SSA def)
	Src1   uint32 // first source value ID
	Src2   uint32 // second source value ID (or auxiliary data)
}

// Compile-time size assertions — ensures AirInst is exactly 16 bytes.
var _ = [1]struct{}{}[16-unsafe.Sizeof(AirInst{})]
var _ = [1]struct{}{}[unsafe.Sizeof(AirInst{})-16]

// Opcode is the instruction discriminant.
type Opcode uint16

const (
	OpcodeNop        Opcode = iota // no operation
	OpcodeConst                    // load constant: Dest = immediate value stored in Src1
	OpcodeAdd                      // Dest = Src1 + Src2
	OpcodeSub                      // Dest = Src1 - Src2
	OpcodeMul                      // Dest = Src1 * Src2
	OpcodeDiv                      // Dest = Src1 / Src2
	OpcodeMod                      // Dest = Src1 % Src2
	OpcodeEq                       // Dest = Src1 == Src2
	OpcodeNe                       // Dest = Src1 != Src2
	OpcodeLt                       // Dest = Src1 < Src2
	OpcodeLe                       // Dest = Src1 <= Src2
	OpcodeGt                       // Dest = Src1 > Src2
	OpcodeGe                       // Dest = Src1 >= Src2
	OpcodeAnd                      // Dest = Src1 & Src2 (bitwise)
	OpcodeOr                       // Dest = Src1 | Src2 (bitwise)
	OpcodeXor                      // Dest = Src1 ^ Src2
	OpcodeShl                      // Dest = Src1 << Src2
	OpcodeShr                      // Dest = Src1 >> Src2
	OpcodeNeg                      // Dest = -Src1
	OpcodeNot                      // Dest = ~Src1 (bitwise not)
	OpcodeLoad                     // Dest = *Src1
	OpcodeStore                    // *Dest = Src1
	OpcodeAlloc                    // Dest = alloc(TypeID)
	OpcodeDealloc                  // dealloc(Src1)
	OpcodeCall                     // Dest = call Src1(args via Extras)
	OpcodeReturn                   // return Src1
	OpcodeJump                     // jump to block Src1
	OpcodeBranch                   // if Src1 jump Src2 else Dest
	OpcodePhi                      // SSA phi node
	OpcodeGetField                 // Dest = Src1.field[Src2]
	OpcodeSetField                 // Src1.field[Src2] = Dest
	OpcodeIndex                    // Dest = Src1[Src2]
	OpcodeSlice                    // Dest = Src1[Src2:Dest] (uses ExtraIdx for end)
	OpcodeCast                     // Dest = cast(Src1) to TypeID
	OpcodeSpawn                    // spawn actor with fn Src1
	OpcodeAwait                    // Dest = await Src1
	OpcodeDestroyVal               // destroy owned value Src1

	OpcodeCount // sentinel
)
