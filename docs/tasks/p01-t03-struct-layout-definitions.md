# p01-t03: Struct Layout Definitions (FROZEN)

## Purpose
Define the core data structure layouts in Go that ALL other compiler modules depend on. These structures — `Token`, `AstNode`, and `AirInst` — are FROZEN: their field layout, size, and byte offsets must never change without a formal RFC and full compiler migration. Getting these right at the start prevents cascading breakage across every compiler phase. The sizes (8, 24, 16 bytes respectively) are cache-line friendly and designed for bulk array processing.

## Context
These three structures form the backbone of the entire compiler pipeline:
- `Token` (8 bytes): output of the lexer, input to the parser. Zero-copy design — stores offset+length into the original source buffer, no string allocations.
- `AstNode` (24 bytes): the flat-array AST. All nodes live in a single `[]AstNode` slice; tree structure is encoded via index fields (FirstChild, NextSibling), not pointers. This enables cache-friendly traversal and O(1) serialization.
- `AirInst` (16 bytes): AXIOM Intermediate Representation instruction in SSA form. Dest, Src1, Src2 are value IDs (indices into a value table), not pointers.

The `unsafe.Sizeof` tests written here become permanent regression guards — if anyone accidentally adds a field, the test fails immediately.

Spec references: `05. IR thật sự.md` (AirInst), `03. Thiết kế parser thực tế.md` (AstNode), `01.minimal core.md` (Token).

## Inputs
- `go.mod` (established in p01-t01)
- `compiler/lexer/` package stub (p01-t01)
- `compiler/ast/` package stub (p01-t01)
- `ir/air/` package stub (p01-t01)
- AXIOM LANGUAGE SPECIFICATION v1.0.md

## Outputs
- `compiler/lexer/token.go` — `Token` struct (8 bytes) + `NodeKind` placeholder
- `compiler/ast/node.go` — `AstNode` struct (24 bytes), `NodeKind` enum, `Flags` bit constants
- `ir/air/inst.go` — `AirInst` struct (16 bytes), `Opcode` enum stub
- `compiler/lexer/token_test.go` — size assertion tests
- `compiler/ast/node_test.go` — size assertion tests
- `ir/air/inst_test.go` — size assertion tests

## Dependencies
- p01-t01: repository-bootstrap — packages must exist with correct module path

## Subsystems Affected
- `compiler/lexer/`: Token defined here; lexer implementation (p02-t02) builds on this
- `compiler/ast/`: AstNode defined here; AST builder (p03-t01) extends this
- `compiler/parser/`: Parser produces AstNodes; depends on NodeKind enum
- `ir/air/`: AirInst defined here; AIR builder (p03-t01 via ir/builder) uses this
- All other compiler phases: reference these types by import

## Detailed Requirements

1. **Token struct** in `compiler/lexer/token.go` — must be exactly 8 bytes:
   ```go
   package lexer

   import "unsafe"

   // Token represents a single lexical token.
   // Layout is FROZEN at 8 bytes. Do not add fields without an RFC.
   // Zero-copy design: the token does not own its text.
   // Use src[tok.Offset : tok.Offset+uint32(tok.Len)] to recover the text.
   type Token struct {
       Kind   TokenKind // 1 byte (uint8)
       _      uint8     // 1 byte padding (reserved, must remain zero)
       Len    uint16    // 2 bytes: length of token text in source bytes
       Offset uint32    // 4 bytes: byte offset of token start in source
   }

   // TokenKind is the discriminant of a Token.
   // Must fit in uint8 (max 255 values). See token_kind.go for the full enum.
   type TokenKind uint8

   var _ = [1]struct{}{}[unsafe.Sizeof(Token{})-8] // compile-time size assert
   ```
   Note: The padding byte `_` is reserved for future use (e.g., a flags nibble). It must always read as zero. The compile-time size assertion using array trick ensures the struct never grows silently.

2. **AstNode struct** in `compiler/ast/node.go` — must be exactly 24 bytes:
   ```go
   package ast

   import "unsafe"

   // AstNode is a node in the flat-array AST.
   // Layout is FROZEN at 24 bytes. Do not add fields without an RFC.
   //
   // Tree structure is encoded via index fields, not pointers.
   // All nodes of a compilation unit live in a single []AstNode slice.
   // Index 0 is always the root Program node.
   //
   // Field layout (24 bytes total):
   //   Kind        NodeKind  1B   @ offset 0
   //   _pad        uint8     1B   @ offset 1  (reserved)
   //   Flags       uint16    2B   @ offset 2
   //   TokenIdx    uint32    4B   @ offset 4
   //   FirstChild  uint32    4B   @ offset 8
   //   NextSibling uint32    4B   @ offset 12
   //   Payload     uint32    4B   @ offset 16
   //   ExtraIdx    uint32    4B   @ offset 20
   type AstNode struct {
       Kind        NodeKind // discriminant
       _           uint8    // padding, reserved
       Flags       uint16   // bit flags, see Flags* constants below
       TokenIdx    uint32   // index into the token slice for this node's primary token
       FirstChild  uint32   // index of first child node (0 = no children)
       NextSibling uint32   // index of next sibling node (0 = last sibling)
       Payload     uint32   // multipurpose: SymbolIdx, TypeID, or literal value depending on Kind
       ExtraIdx    uint32   // index into AstTree.Extras for overflow data
   }

   var _ = [1]struct{}{}[unsafe.Sizeof(AstNode{})-24] // compile-time size assert
   ```

3. **NodeKind enum** — all node kinds the parser can emit:
   ```go
   // NodeKind is the type discriminant for AstNode.
   type NodeKind uint8

   const (
       NodeInvalid NodeKind = iota // 0: sentinel / error node

       // Top-level declarations
       NodeProgram       // root node
       NodeFuncDecl      // fn foo(...)
       NodeStructDecl    // struct Foo:
       NodeInterfaceDecl // interface Bar:
       NodeImportDecl    // import std.fs
       NodeConstDecl     // const X: T = expr
       NodeTypeAliasDecl // type Result = Ok(i32) | Err(string)

       // Sub-declarations
       NodeParamDecl  // function parameter
       NodeFieldDecl  // struct field
       NodeMethodSig  // interface method signature
       NodeVariantDecl // sum type variant

       // Statements
       NodeBlock      // indented block
       NodeVarDecl    // let x: T = expr
       NodeAssignStmt // x = expr, x += expr
       NodeReturnStmt // return expr
       NodeIfStmt     // if/elif/else chain
       NodeElifClause // elif branch
       NodeElseClause // else branch
       NodeForStmt    // for x in expr:
       NodeWhileStmt  // while cond:
       NodeMatchStmt  // match expr:
       NodeMatchArm   // pattern: body
       NodeDeferStmt  // defer expr
       NodeUnsafeBlock // unsafe:
       NodeArenaBlock // in [arena]:

       // Expressions
       NodeBinaryExpr  // lhs op rhs
       NodeUnaryExpr   // op expr
       NodeCallExpr    // fn(args)
       NodeIndexExpr   // expr[idx]
       NodeFieldExpr   // expr.field
       NodeCastExpr    // expr as Type
       NodeDerefExpr   // expr.*
       NodeSpawnExpr   // spawn expr
       NodeAwaitExpr   // await expr
       NodeClosureExpr // |params| body

       // Literals and atoms
       NodeIntLit    // integer literal
       NodeFloatLit  // float literal
       NodeStringLit // string literal
       NodeCharLit   // character literal
       NodeBoolLit   // true / false
       NodeNilLit    // nil
       NodeIdent     // identifier reference
       NodeArrayLit  // [expr, ...]
       NodeStructLit // TypeName{field: expr, ...}
       NodeNamedArg  // field: expr in call/struct

       // Type expressions
       NodeTypeExpr       // generic type node wrapping a type expression
       NodePtrType        // *T or *mut T
       NodeSliceType      // [T]
       NodeArrayType      // [T; N]
       NodeFuncType       // fn(A, B) -> C
       NodeGenericType    // Foo[T]
       NodeIsolatedType   // Isolated[T]
       NodeFutureType     // Future[T]
       NodeSumType        // A | B

       // Patterns
       NodeWildcardPat // _
       NodeLiteralPat  // literal in match arm
       NodeBindingPat  // name binding in match arm
       NodeVariantPat  // Variant(inner)
       NodeTuplePat    // (a, b)

       // Generics
       NodeGenericParams // [T: Interface]
       NodeGenericParam  // single T: Interface

       // Effects
       NodeEffectAnnotation // {.raises: [T].}

       // Error recovery
       NodeError // parse error node

       NodeKindCount // sentinel — total count
   )
   ```

4. **Flags bit constants** for `AstNode.Flags`:
   ```go
   const (
       FlagIsPub    uint16 = 1 << 0  // declaration is pub
       FlagIsMut    uint16 = 1 << 1  // variable is mut / pointer is *mut
       FlagIsAsync  uint16 = 1 << 2  // function is async
       FlagIsExtern uint16 = 1 << 3  // function is extern
       FlagIsSink   uint16 = 1 << 4  // parameter is !T (sink/consumed)
       FlagIsLent   uint16 = 1 << 5  // parameter is lent (borrowed)
       FlagIsPacked uint16 = 1 << 6  // struct is packed
       FlagEscapesToHeap uint16 = 1 << 7 // escape analysis: value escapes to heap
       FlagUsesArena     uint16 = 1 << 8 // allocation uses arena allocator
       FlagIsGeneric     uint16 = 1 << 9 // declaration has generic parameters
       FlagIsMoved       uint16 = 1 << 10 // value has been moved (ownership tracking)
   )
   ```

5. **AirInst struct** in `ir/air/inst.go` — must be exactly 16 bytes:
   ```go
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
   //   Opcode  uint16  2B  @ offset 0
   //   TypeID  uint16  2B  @ offset 2
   //   Dest    uint32  4B  @ offset 4
   //   Src1    uint32  4B  @ offset 8
   //   Src2    uint32  4B  @ offset 12
   type AirInst struct {
       Opcode Opcode // instruction opcode
       TypeID uint16 // result type (index into TypeTable)
       Dest   uint32 // destination value ID (SSA def)
       Src1   uint32 // first source value ID
       Src2   uint32 // second source value ID (or auxiliary data)
   }

   var _ = [1]struct{}{}[unsafe.Sizeof(AirInst{})-16] // compile-time size assert

   // Opcode is the instruction discriminant.
   type Opcode uint16
   ```

6. **Opcode stub enum** (to be expanded in ir phase):
   ```go
   const (
       OpcodeNop    Opcode = iota // no operation
       OpcodeConst               // load constant: Dest = immediate value stored in Src1
       OpcodeAdd                 // Dest = Src1 + Src2
       OpcodeSub                 // Dest = Src1 - Src2
       OpcodeMul                 // Dest = Src1 * Src2
       OpcodeDiv                 // Dest = Src1 / Src2
       OpcodeMod                 // Dest = Src1 % Src2
       OpcodeEq                  // Dest = Src1 == Src2
       OpcodeNe                  // Dest = Src1 != Src2
       OpcodeLt                  // Dest = Src1 < Src2
       OpcodeLe                  // Dest = Src1 <= Src2
       OpcodeGt                  // Dest = Src1 > Src2
       OpcodeGe                  // Dest = Src1 >= Src2
       OpcodeAnd                 // Dest = Src1 & Src2 (bitwise)
       OpcodeOr                  // Dest = Src1 | Src2 (bitwise)
       OpcodeXor                 // Dest = Src1 ^ Src2
       OpcodeShl                 // Dest = Src1 << Src2
       OpcodeShr                 // Dest = Src1 >> Src2
       OpcodeNeg                 // Dest = -Src1
       OpcodeNot                 // Dest = ~Src1 (bitwise not)
       OpcodeLoad                // Dest = *Src1
       OpcodeStore               // *Dest = Src1
       OpcodeAlloc               // Dest = alloc(TypeID)
       OpcodeDealloc             // dealloc(Src1)
       OpcodeCall                // Dest = call Src1(args via Extras)
       OpcodeReturn              // return Src1
       OpcodeJump                // jump to block Src1
       OpcodeBranch              // if Src1 jump Src2 else Dest
       OpcodePhi                 // SSA phi node
       OpcodeGetField            // Dest = Src1.field[Src2]
       OpcodeSetField            // Src1.field[Src2] = Dest
       OpcodeIndex               // Dest = Src1[Src2]
       OpcodeSlice               // Dest = Src1[Src2:Dest] (uses ExtraIdx for end)
       OpcodeCast                // Dest = cast(Src1) to TypeID
       OpcodeSpawn               // spawn actor with fn Src1
       OpcodeAwait               // Dest = await Src1
       OpcodeDestroyVal          // destroy owned value Src1

       OpcodeCount // sentinel
   )
   ```

7. **Size assertion tests** — these are regression guards, not unit tests:

   `compiler/lexer/token_test.go`:
   ```go
   package lexer

   import (
       "testing"
       "unsafe"
   )

   func TestTokenSize(t *testing.T) {
       const want = 8
       got := unsafe.Sizeof(Token{})
       if got != want {
           t.Fatalf("Token size = %d bytes, want %d bytes. "+
               "Token layout is FROZEN. Do not add fields without an RFC.", got, want)
       }
   }

   func TestTokenFieldOffsets(t *testing.T) {
       var tok Token
       base := uintptr(unsafe.Pointer(&tok))
       kindOff := uintptr(unsafe.Pointer(&tok.Kind)) - base
       lenOff  := uintptr(unsafe.Pointer(&tok.Len)) - base
       offOff  := uintptr(unsafe.Pointer(&tok.Offset)) - base
       if kindOff != 0 { t.Errorf("Token.Kind offset = %d, want 0", kindOff) }
       if lenOff  != 2 { t.Errorf("Token.Len offset = %d, want 2", lenOff) }
       if offOff  != 4 { t.Errorf("Token.Offset offset = %d, want 4", offOff) }
   }
   ```

   `compiler/ast/node_test.go`:
   ```go
   package ast

   import (
       "testing"
       "unsafe"
   )

   func TestAstNodeSize(t *testing.T) {
       const want = 24
       got := unsafe.Sizeof(AstNode{})
       if got != want {
           t.Fatalf("AstNode size = %d bytes, want %d bytes. "+
               "AstNode layout is FROZEN. Do not add fields without an RFC.", got, want)
       }
   }

   func TestAstNodeFieldOffsets(t *testing.T) {
       var n AstNode
       base := uintptr(unsafe.Pointer(&n))
       check := func(name string, got, want uintptr) {
           t.Helper()
           if got != want {
               t.Errorf("AstNode.%s offset = %d, want %d", name, got, want)
           }
       }
       check("Kind",        uintptr(unsafe.Pointer(&n.Kind))-base,        0)
       check("Flags",       uintptr(unsafe.Pointer(&n.Flags))-base,       2)
       check("TokenIdx",    uintptr(unsafe.Pointer(&n.TokenIdx))-base,    4)
       check("FirstChild",  uintptr(unsafe.Pointer(&n.FirstChild))-base,  8)
       check("NextSibling", uintptr(unsafe.Pointer(&n.NextSibling))-base, 12)
       check("Payload",     uintptr(unsafe.Pointer(&n.Payload))-base,     16)
       check("ExtraIdx",    uintptr(unsafe.Pointer(&n.ExtraIdx))-base,    20)
   }

   func TestNodeKindCount(t *testing.T) {
       // Ensure NodeKindCount fits in uint8 (TokenKind is uint8)
       if NodeKindCount > 255 {
           t.Fatalf("NodeKindCount = %d exceeds uint8 max (255)", NodeKindCount)
       }
   }

   func TestFlagConstants(t *testing.T) {
       // All flag constants must be distinct powers of 2
       flags := []uint16{
           FlagIsPub, FlagIsMut, FlagIsAsync, FlagIsExtern,
           FlagIsSink, FlagIsLent, FlagIsPacked, FlagEscapesToHeap,
           FlagUsesArena, FlagIsGeneric, FlagIsMoved,
       }
       seen := map[uint16]bool{}
       for _, f := range flags {
           if f == 0 { t.Error("flag must not be zero") }
           if f&(f-1) != 0 { t.Errorf("flag 0x%04x is not a power of 2", f) }
           if seen[f] { t.Errorf("duplicate flag value 0x%04x", f) }
           seen[f] = true
       }
   }
   ```

   `ir/air/inst_test.go`:
   ```go
   package air

   import (
       "testing"
       "unsafe"
   )

   func TestAirInstSize(t *testing.T) {
       const want = 16
       got := unsafe.Sizeof(AirInst{})
       if got != want {
           t.Fatalf("AirInst size = %d bytes, want %d bytes. "+
               "AirInst layout is FROZEN. Do not add fields without an RFC.", got, want)
       }
   }

   func TestAirInstFieldOffsets(t *testing.T) {
       var inst AirInst
       base := uintptr(unsafe.Pointer(&inst))
       check := func(name string, got, want uintptr) {
           t.Helper()
           if got != want {
               t.Errorf("AirInst.%s offset = %d, want %d", name, got, want)
           }
       }
       check("Opcode", uintptr(unsafe.Pointer(&inst.Opcode))-base, 0)
       check("TypeID", uintptr(unsafe.Pointer(&inst.TypeID))-base, 2)
       check("Dest",   uintptr(unsafe.Pointer(&inst.Dest))-base,   4)
       check("Src1",   uintptr(unsafe.Pointer(&inst.Src1))-base,   8)
       check("Src2",   uintptr(unsafe.Pointer(&inst.Src2))-base,   12)
   }

   func TestOpcodeCount(t *testing.T) {
       if OpcodeCount > 65535 {
           t.Fatalf("OpcodeCount = %d exceeds uint16 max", OpcodeCount)
       }
   }
   ```

## Implementation Steps

1. Create `compiler/lexer/token.go` with the `Token` struct, `TokenKind` type, and compile-time size assertion. Do NOT define `TokenKind` constants here — that is p02-t01's job. Only declare the type.

2. Create `compiler/ast/node.go` with:
   - `AstNode` struct with all 7 fields
   - Compile-time size assertion
   - `NodeKind` enum with all constants listed in Requirement 3
   - `Flags` bit constants listed in Requirement 4

3. Create `ir/air/inst.go` with:
   - `AirInst` struct with all 5 fields
   - Compile-time size assertion
   - `Opcode` type declaration
   - `Opcode` constants listed in Requirement 6

4. Create `compiler/lexer/token_test.go` with `TestTokenSize` and `TestTokenFieldOffsets`.

5. Create `compiler/ast/node_test.go` with `TestAstNodeSize`, `TestAstNodeFieldOffsets`, `TestNodeKindCount`, `TestFlagConstants`.

6. Create `ir/air/inst_test.go` with `TestAirInstSize`, `TestAirInstFieldOffsets`, `TestOpcodeCount`.

7. Run `go test ./compiler/lexer/ ./compiler/ast/ ./ir/air/` — all size and offset tests must pass.

8. Run `go build ./...` — no compilation errors anywhere.

9. Verify that the compile-time size assertions (`var _ = [1]struct{}{}[unsafe.Sizeof(...)-N]`) produce a helpful compiler error message if the size is wrong (test this by temporarily adding a field and observing the error).

## Test Plan

- **TestTokenSize**: asserts `unsafe.Sizeof(Token{}) == 8`. Fails if layout grows.
- **TestTokenFieldOffsets**: asserts each field is at the expected byte offset.
- **TestAstNodeSize**: asserts `unsafe.Sizeof(AstNode{}) == 24`. Critical regression guard.
- **TestAstNodeFieldOffsets**: asserts all 7 fields at exact offsets 0,2,4,8,12,16,20.
- **TestNodeKindCount**: asserts the enum has fewer than 256 values (fits in uint8).
- **TestFlagConstants**: asserts all flags are distinct powers of 2.
- **TestAirInstSize**: asserts `unsafe.Sizeof(AirInst{}) == 16`.
- **TestAirInstFieldOffsets**: asserts all 5 fields at offsets 0,2,4,8,12.
- **TestOpcodeCount**: asserts fewer than 65536 opcodes.

All tests must pass on amd64 (primary), arm64, and 386 (no pointer-sized fields, so sizes should be architecture-independent).

## Validation Checklist
- [ ] `unsafe.Sizeof(Token{}) == 8` verified by test
- [ ] `unsafe.Sizeof(AstNode{}) == 24` verified by test
- [ ] `unsafe.Sizeof(AirInst{}) == 16` verified by test
- [ ] All field offsets match specification (Token: 0,2,4; AstNode: 0,2,4,8,12,16,20; AirInst: 0,2,4,8,12)
- [ ] NodeKind enum has all 60+ node types listed
- [ ] Flags constants are all distinct powers of 2
- [ ] Opcode enum has all instruction types listed
- [ ] Compile-time size assertions present in all three files
- [ ] `go test ./...` passes
- [ ] `go build ./...` passes
- [ ] No pointer-sized fields in any of the three structs (sizes would be arch-dependent)

## Acceptance Criteria
- All 9 test functions pass on Linux amd64, macOS arm64, and Windows amd64
- If a developer adds any field to Token, AstNode, or AirInst, a compile-time error is produced immediately
- `NodeKindCount` is correctly updated when new node kinds are added
- `OpcodeCount` is correctly updated when new opcodes are added
- All Go files pass `gofmt` and `golangci-lint`

## Definition of Done
- [ ] All 9 tests pass across 3 platforms
- [ ] Compile-time assertions verified to catch size changes
- [ ] Reviewed and approved — these layouts are now FROZEN
- [ ] Committed to repository with tag comment `// FROZEN: do not modify without RFC`
- [ ] Referenced in `docs/CONTRIBUTING.md` (p01-t05) as FROZEN structures

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Go compiler adds padding unexpectedly | Use explicit `_` padding fields where needed; verify with offset tests |
| AstNode needs more than 255 node kinds | uint8 allows 256 kinds; currently ~65 defined; document limit; future: promote Kind to uint16 via RFC |
| AirInst needs more than 65535 opcodes | uint16 allows 65536; current needs ~40; document limit |
| Token's 1-byte reserved field is misused | Add a comment `// must remain zero`; linter check for non-zero writes |
| 32-bit Offset limits source files to 4GB | Acceptable limit; document in code; future: extend via ExtraIdx if needed |
| Architecture-specific padding on 32-bit platforms | No pointer fields in any struct; sizes are deterministic on all platforms |

## Future Follow-up Tasks
- p02-t01: Define all `TokenKind` constants using the `TokenKind` type declared here
- p02-t02: Implement lexer that produces `[]Token` using this Token struct
- p03-t01: Implement `AstTree` and builder using `AstNode` from this task
- ir phase: Expand `Opcode` enum and implement AIR builder using `AirInst` from here
- RFC required before any field change to Token, AstNode, or AirInst
