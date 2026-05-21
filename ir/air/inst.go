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
// Opcodes use class-prefixed hex values. See opcodes.go for definitions.
type Opcode uint16
