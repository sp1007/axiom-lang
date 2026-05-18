# p09-t04: AIR Verifier

## Purpose
Implement an AIR verifier that checks SSA invariants, CFG consistency, type correctness, and ownership rules at the IR level. The verifier runs after every AIR transformation and serves as the correctness oracle for all passes.

## Context
"Trust but verify" — every AIR pass must produce valid AIR. The verifier catches bugs in AIR builders and optimization passes early, before they manifest as mysterious codegen failures. It is the AIR equivalent of LLVM's `verifyModule()`. Running after every pass (not just at the end) makes bugs easier to locate.

## Inputs
- `AirFunc` or `AirModule` to verify
- `TypeTable` — for type consistency checks

## Outputs
- `[]Diagnostic` — empty if valid, diagnostic list if invalid
- `VerificationResult{Valid bool, Errors []Diagnostic}`

## Dependencies
- p09-t02: air-basic-blocks — CFG structure to verify
- p09-t01: air-instruction-set — opcode semantics

## Subsystems Affected
- All AIR passes: verifier runs after each pass
- CI: verification runs on all AIR output in tests
- Debugging: errors pinpoint the exact invalid instruction

## Detailed Requirements

1. `Verifier` struct: `module *AirModule, tt *TypeTable`
2. `Verify(func *AirFunc) []Diagnostic` — returns all errors found (not fail-fast)
3. Checks to perform:
   **SSA invariants:**
   - Each virtual register (Dest) defined exactly once across all instructions
   - Every use (Src1, Src2) of a register preceded by a definition (in dominator order)
   - Phi nodes only at start of blocks (before any non-phi instruction)
   - Phi operands: one per predecessor block
   **CFG invariants:**
   - Every block has exactly one terminator as last instruction
   - Terminator targets reference valid block IDs
   - Entry block has no predecessors
   - All blocks reachable from entry (no orphan blocks)
   - No critical edges (if needed by optimizer)
   **Type invariants:**
   - Dest TypeID matches the operation's result type (e.g., OpIAdd result must be integer)
   - Src operand types compatible with opcode
   - OpCall: arg types match callee's param types
   - OpReturn: value type matches function return type
   **Ownership invariants:**
   - OpMove: source register not used after the OpMove
   - OpDestroy: target is a heap-allocated value (OwnerInfo in metadata)
4. Error format: `"block_3:inst_12: %r5 defined twice"`.
5. Performance: O(V+E) for SSA checks, O(V) for CFG checks.

## Implementation Steps

1. Create `ir/air/verifier.go`.
2. Implement `checkSSA(func *AirFunc)`: build def-set, check each Src has a prior def.
3. Implement `checkCFG(func *AirFunc)`: verify terminators, edges, reachability (BFS from entry).
4. Implement `checkTypes(func *AirFunc)`: for each instruction, verify TypeID rules.
5. Implement `checkOwnership(func *AirFunc)`: verify OpMove/OpDestroy invariants.
6. Export `RunVerifier(module *AirModule) []Diagnostic` — convenience wrapper.
7. Write extensive unit tests.

## Test Plan

- `TestVerifyValidFunc`: valid simple function → 0 errors
- `TestVerifyDoubleDefine`: register defined twice → error
- `TestVerifyUseBeforeDef`: use of register before its definition → error
- `TestVerifyMissingTerminator`: block with no terminator → error
- `TestVerifyDeadBlock`: unreachable block → error (warning level)
- `TestVerifyTypeMismatch`: OpIAdd with string operands → type error
- `TestVerifyCallArgTypes`: call with wrong arg types → error

## Validation Checklist

- [ ] SSA uniqueness enforced
- [ ] Use-before-def detected
- [ ] Missing terminator detected
- [ ] Type consistency checked
- [ ] Verifier itself never panics on invalid input
- [ ] O(V+E) performance

## Acceptance Criteria

- Verifier finds all injected errors in malformed test AIR modules
- Valid AIR from AIR builder produces 0 verifier errors
- Verifier runs in < 10ms for 1000-instruction function

## Definition of Done

- [ ] `ir/air/verifier.go` implemented
- [ ] Unit tests pass
- [ ] Integrated into AIR builder — runs after each function is built
- [ ] Integrated into optimization pipeline — runs after each pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Verifier is too strict and rejects valid AIR | Start with must-have invariants; add more as passes mature |
| Verifier performance degrades for large functions | Profile; implement lazy checks for O(1) common cases |

## Future Follow-up Tasks

- p09-t11: axc-dump-air runs verifier and reports failures
- p10-t01: opt-pipeline-manager runs verifier between passes
