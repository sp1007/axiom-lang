# p09-t02: AIR Basic Blocks & CFG

## Purpose
Define the basic block and control flow graph (CFG) data structures for AIR, providing the organizational container for AIR instructions and enabling all CFG-based analyses (dominator trees, loop detection, liveness analysis).

## Context
A basic block is a straight-line sequence of instructions with exactly one entry (the block label) and one exit (a terminator instruction: jump, branch, or return). The CFG is the directed graph of basic blocks connected by control flow edges. All optimization passes and register allocation operate on the CFG.

## Inputs
- `ir/air/opcodes.go` — instruction set from p09-t01
- `compiler/types/types.go` — TypeIDs for function signatures

## Outputs
- `ir/air/cfg.go` — BasicBlock, AirFunc, AirModule types
- API for building and traversing the CFG

## Dependencies
- p09-t01: air-instruction-set — AirInst type used

## Subsystems Affected
- AIR builder (p09-t06 through p09-t10): builds CFG nodes
- AIR verifier (p09-t04): validates CFG structure
- Optimization passes: traverse and modify CFG
- Register allocator (p11-t04, p11-t05): uses CFG for liveness

## Detailed Requirements

1. `BasicBlock` struct:
   ```go
   type BasicBlock struct {
       ID        uint32
       Instrs    []uint32    // indices into AirFunc.Insts slice
       Succs     []uint32    // successor block IDs
       Preds     []uint32    // predecessor block IDs
       LoopDepth uint8       // 0 = not in loop, 1 = outermost loop, etc.
       IsEntry   bool
       IsExit    bool
   }
   ```
2. `AirFunc` struct:
   ```go
   type AirFunc struct {
       SymID    uint32
       Name     uint32        // interned name
       Params   []uint32      // TypeIDs of parameters
       RetType  uint32        // return TypeID
       Blocks   []BasicBlock
       Insts    []AirInst     // flat instruction array
       Extras   []uint32      // extra operands for variadic instructions
       IsAsync  bool
       IsExtern bool
   }
   ```
3. `AirModule` struct: `Funcs []AirFunc, TypeTable *TypeTable, StringPool *InternPool`
4. CFG invariants:
   - Every BasicBlock has exactly one terminator as its last instruction (OpJump, OpBranch, or OpReturn)
   - Entry block has no predecessors
   - Exit block(s) have OpReturn terminators
   - Phi nodes only appear at the start of a block (before any non-phi)
5. `AirFuncBuilder` helper: stateful builder pattern for constructing AirFunc:
   - `NewBlock() uint32` — create new empty block, return ID
   - `SwitchTo(blockID uint32)` — set current emission block
   - `Emit(inst AirInst) uint32` — add inst to current block, return inst index
   - `FreshReg() uint32` — allocate new virtual register (SSA: each written once)
   - `Terminate(jumpTarget or branchInst)` — add terminator to current block

## Implementation Steps

1. Create `ir/air/cfg.go` with BasicBlock, AirFunc, AirModule.
2. Create `ir/air/builder.go` with AirFuncBuilder.
3. Implement `NewBlock()`, `SwitchTo()`, `Emit()`, `FreshReg()`.
4. Implement `AddEdge(from, to uint32)` — adds to Succs/Preds.
5. Implement `ComputeDominators() []uint32` — dominator tree using Cooper et al. iterative algorithm.
6. Implement `PostOrder() []uint32` — post-order traversal of CFG.
7. Implement `ReversePostOrder() []uint32` — used by liveness analysis.
8. Write unit tests for all builder operations.

## Test Plan

- `TestBasicBlockCreate`: create block, emit instructions, verify Instrs
- `TestCFGEdges`: add edge, verify Succs/Preds updated
- `TestFreshReg`: verify each call returns unique ID
- `TestTerminator`: verify block with no terminator is caught by verifier
- `TestDominators`: compute dominators for a simple if/else CFG
- `TestPostOrder`: verify post-order traversal order

## Validation Checklist

- [ ] BasicBlock enforces single terminator
- [ ] Succs and Preds are consistent (if A→B in Succs, B has A in Preds)
- [ ] FreshReg returns monotonically increasing IDs
- [ ] Dominator computation handles loops correctly
- [ ] AirModule owns all AirFuncs (one module per compilation unit)

## Acceptance Criteria

- A simple function (hello world) produces 1 entry block, 1 exit block, 1 return inst
- if/else produces 3 blocks (header, then, else) with correct edges

## Definition of Done

- [ ] `ir/air/cfg.go` implemented
- [ ] `ir/air/builder.go` implemented
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| CFG edge consistency bugs (Succs without Preds) | Always use `AddEdge()` helper, never modify Succs/Preds directly |
| Virtual register count overflow (> uint32) | Panic with "too many virtual registers" at 4B limit |

## Future Follow-up Tasks

- p09-t04: air-verifier validates CFG invariants
- p11-t04: liveness-analysis uses ReversePostOrder and CFG edges
