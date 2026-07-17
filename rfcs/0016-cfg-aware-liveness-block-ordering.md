# RFC 0016 — CFG-aware liveness / RPO block ordering (unblocks short-circuit `and`/`or`, BUG#86)

Status: **DRAFT** (2026-07-09). Blocks: BUG#86 (short-circuit evaluation), and any future lowering that creates control-flow blocks *in the middle of an expression*.

---

## 1. Problem — the register allocator's liveness is not control-flow-aware

`x86_regalloc.compute_liveness` (`bootstrap/stage1/x86_regalloc.ax:55`) builds each vreg's live interval as a single contiguous range **`[first_def_index, last_use_index]` measured in linear instruction-array position**. It is NOT CFG-aware:

- A vreg's interval is created at its *first* definition; a *second* definition of the same vreg is ignored (`:78` only creates when the slot is null).
- Uses extend `end` to the current index (`:102`, `:114`).
- The only concession to control flow is a back-edge pass (`:164-182`): a `JMP`/`JCC` whose target index is *earlier* than itself extends the intervals of vregs live across the target — i.e. loop-carried values.

This is correct **only if the linear instruction-array order is a topological order of the CFG** (every block appears after its non-loop predecessors). Blocks are serialized in **creation order** (`x86_selector.ax:2211-2235` iterates `f.blocks.data` by id; `air.new_block` assigns `id = blocks.len`). Control transfers are all **explicit** (`OP_BRANCH` → `JCC target` + `JMP else`; `OP_JUMP` → `JMP`), so physical block order does not affect *correctness of control flow* — only the *linear indices the liveness sees*.

### Why the existing lowerings survive this
`lower_if` creates `merge_block` **first** (id M), then `lower_if_chain` creates `then`(M+1)/`else`(M+2). So the serialized order is `[merge, then, else]` — already **out of control-flow order** (merge executes last but is serialized first). This works only by luck of two properties:
1. The `merge` join block is **empty** (just a label + whatever *following* code lands there).
2. Values that cross the diamond are **mut-local vregs** (`OP_COPY` into an `existing_reg` from `local_map`) whose first def is *before* the `if` and whose real use is *after* the whole `if` — so the interval `[before_if, after_if]` conservatively covers every reassignment regardless of block order.

## 2. Why short-circuit `and`/`or` breaks it (BUG#86)

Short-circuit lowering (see [[bug86-short-circuit-open]]) must create a `rhs_block` + `merge_block` **in the middle of expression evaluation**, and its result `sc_res` is **consumed AT the merge join** (not after the enclosing structure). When the `and`/`or` sits in an `if` condition, `lower_if_chain` has *already* created `then`/`else`, so the emitted stream is `... P(cond), merge_if, then, else, sc_rhs, sc_merge` while control flow is `P → sc_rhs → sc_merge → then/else → merge_if`. The linear liveness now computes intervals against an order that is *not* a CFG topological order, and the `sc_merge`-consumed value has no post-structure use to paper over it. Result: wrong register/slot assignment → clobbering → **the self-hosted compiler miscompiles itself and hangs on a trivial input**.

**Proven facts** (isolation test, 2026-07-09): the short-circuit *lowering is correct* — a short-circuit compiler built by an eager compiler (`fpA`) compiles small programs perfectly (`false and boom(0)` does not crash; full truth table passes). Only the *self-compiled* compiler (`scB`) breaks — i.e. the failure is the liveness/ordering fragility surfacing in large, high-block-count functions, not the lowering.

## 3. Why the "obvious" quick fixes are NOT the fix

- **Diamond-merged result vreg** (`sc_res` written in two blocks): tried; broke self-host. Structurally identical to how mut-locals cross diamonds, but without the post-structure use.
- **Explicit stack slot via `OP_ALLOC`**: INVALID — `OP_ALLOC` emits a runtime allocator **call** (`x86_selector.ax:1648`, `MACH_CALL imm=-1`), i.e. a heap allocation. Using it per `and`/`or` would malloc on every evaluation and leak inside loops.
- **Reorder just the short-circuit blocks by id**: block ids are sequential; you cannot make a mid-expression block sort before blocks created earlier without a global ordering pass.

The fragility is a **foundational property of the backend**, not specific to short-circuit. Any future lowering that creates blocks mid-expression (ternary `if`-expr as a value, `?`-propagation, inline pattern guards) will hit the same wall.

## 4. Decision points (resolve before implementing)

D1. **Fix location** — (A) serialize blocks in RPO before selection so linear index == CFG topo order; OR (B) make `compute_liveness` walk the CFG (per-block live-in/live-out via dataflow) instead of linear indices. (A) is smaller but requires the terminator invariant (D2); (B) is more robust but rewrites the liveness core.

D2. **Terminator invariant** — RPO reordering is only sound if **every basic block ends with an explicit terminator** (`JMP`/`JCC`/`RET`). Today the selector relies on physical fallthrough for any block that lacks one (e.g. a final block ending in the last statement with an implicit return, or straight-line code split across blocks). Reordering such a block changes what it falls through to → miscompile. Need a normalization pass that appends an explicit `OP_JUMP next` to every non-terminated block **before** reordering, OR prove the invariant already holds.

D3. **Back-edge detection** — with RPO, the current index-based back-edge pass (`target_idx < be_i`) still identifies loop back-edges correctly (RPO guarantees forward edges go low→high, back-edges high→low). Verify no reducibility assumptions are violated by the blocks short-circuit introduces.

D4. **Determinism** — RPO must be computed deterministically (fixed DFS successor order) so builds stay reproducible (fixpoint requirement).

## 5. Recommended phased plan

- **P1 — terminator normalization (independent, low-risk):** add a pass (or assert) ensuring every AIR block ends with a terminator; append `OP_JUMP fallthrough_successor` where missing. Gate: A==B (should be a no-op if the invariant already holds) + regression. This is shippable on its own and de-risks everything after. (Note: investigation shows `ensure_return` + per-construct jumps already make the invariant hold in practice.)
- **~~P2 — RPO serialization~~ — ATTEMPTED 2026-07-09, INSUFFICIENT (reverted).** Implemented reverse-postorder block emission (DFS from block 0 over succ edges, with a safety net appending unreachable blocks). Result: **the RPO compiler self-hosts — `scB` compiles trivial, B==C fixpoint holds** — AND ~30 spot-checked tests pass, BUT **`t_math` regressed 127→124**: a loop-carried float value (Newton/series iterations) is miscompiled. RPO produces *correct* loop ordering (postorder+reverse naturally serializes `cond,body,exit`), so the failure is **not** the ordering — it is `compute_liveness` itself being wrong for loop-carried values under the new indices. **Conclusion: reordering alone cannot fix a liveness algorithm that is not CFG-aware; it just moves which CFG shapes break.** Reverted; tree back to the BUG#88 fixpoint (`0D672CC8`).
- **P2' (revised) — CFG-aware liveness is REQUIRED, not optional. ✅ SHIPPED 2026-07-09 (`x86_regalloc.compute_liveness`).** Replaced the old linear back-edge hack with proper per-block live-in/live-out dataflow (backward fixpoint over the CFG: `live_in[b] = use[b] ∪ (live_out[b] − def[b])`, `live_out[b] = ∪ live_in[succ]`). The base `[first_def,last_use]` linear intervals are kept; the dataflow only *extends* each vreg's contiguous interval to the linear hull of every block where it is live (via `live_in`→cover `block_lo`, `live_out`→cover `block_hi`). Because extension only grows intervals, it can never shrink a real live range — worst case is conservative (extra spills), never a miscompile. Implementation details:
    - New helper `inst_defs_dst(op)`: the def-set is built ONLY from opcodes that actually overwrite dst. CRITICAL: `CMP`/`TEST`/`STORE`/`FCMP` read dst but must NOT be defs (marking them defs would kill the value and shrink liveness). When unsure → not a def (conservative).
    - Blocks: split at index 0, every `MACH_LABEL`, and the instruction after any `JMP`/`JCC`/`RET`. Successors: `JMP`→target, `JCC`→target+fallthrough, `RET`→none, else fallthrough. Deterministic (fixed block/word iteration order) so builds stay reproducible.
    - Bitset dataflow (bit index = vreg id, `u64` words); monotone fixpoint converges.
    - **Gate result (all green):** A!=B → **B==C fixpoint** (new hash `BCEFC38F…`); **full regression 114/114**; **`t_math`=127** (the loop-carried-float canary that RPO regressed to 124 — now correct); +9 control-flow/feature oracles pass. This confirms the RFC thesis: it was `compute_liveness` (not block order) that had to become CFG-aware.
- **P3 — enable short-circuit `and`/`or`. ✅ SHIPPED 2026-07-09 (closes BUG#86).** Added `lower_short_circuit` in `air_builder.ax` (diamond: seed `sc_res` = lhs, branch, rhs block overwrites `sc_res`, merge) + a token-based interception in `lower_binary_expr` (`and`/`&&`/`or`/`||` only — NOT the bitwise `&`/`|`). Correct now that liveness is CFG-aware (P2').
  - **Second bug found & fixed during P3 (`lower_while` CFG edges):** the short-circuit condition of a `while` creates a diamond, so the `OP_BRANCH` lands in the diamond's *merge* block, not `cond_block`. `lower_while` was attributing the body/exit CFG edges to `cond_block` (stale), leaving the merge block with no recorded successor to the exit → `remove_unreachable_blocks()` (DCE) NOP'd the exit's `ret` → infinite loop **at -O1 only** (-O0 has no SSA opt, so it worked, which localized the bug). Fixed by capturing `branch_block = current_block()` after lowering the condition — exactly what `lower_if_chain` already did (that asymmetry is why `if …and…` worked but `while …and…` hung). `lower_for`/`**`-loop conditions are compiler-generated comparisons (never diamonds), so they were left as-is.
  - **Gate (all green):** isolation oracles `sc1=sc2=42` (faulting-RHS skipped), `scv=3` (truth table), `scw=5` (while+and), `scstress=32` (nested chains, while-cond, mixed precedence, as-value); B compiles trivial; **B==C fixpoint** (`c777ef7b…`); full regression 114/114; t_math=127.

## 6. Alternatives

- **A. Keep eager `and`/`or` forever, document the spec deviation.** ✗ Violates spec (`docs/tasks/p09-t06`, p08-t04) and the natural `p != null and p[0]` idiom faults. Only acceptable as status quo, not a resolution.
- **B. Lower short-circuit without new blocks** (e.g. `cmov`-based select of both eagerly-evaluated sides). ✗ Defeats the purpose — the RHS still executes (faults/side-effects). Not short-circuit.
- **C. Special-case: only short-circuit when the RHS is provably side-effect-free and fault-free.** ✗ Loses the main benefit (avoiding the faulting-RHS case), and "fault-free" is undecidable in general (a deref can fault).

## 7. Success criteria

- P1: terminator invariant holds/enforced; A==B; regression green.
- P2: RPO serialization; B==C fixpoint; regression green; control-flow-heavy oracles pass; no perf regression in self-build time beyond noise.
- P3: short-circuit shipped; `scB` self-hosts; crash-based + truth-table oracles pass; the guard idiom `if p != null and p[0] > 0` no longer faults.

## 7b. Re-confirmation 2026-07-17 — liveness IS sound for loop-crossing values (RFC 0025 blocker-1 disproved)

RFC 0025 (LICM re-enablement) originally recorded a "blocker 1" claiming the
register allocator's liveness "is a single linear `[def,use]` interval, not
CFG-aware (RFC 0016), and mishandles loop-crossing values" — i.e. that hoisting a
value so it becomes live across a loop and used in the loop condition would
miscompile. **That claim was investigated at the code level and empirically, and is
FALSE for the shipped P2' liveness.** Two facts:

1. **Code:** `compute_liveness` after P2' keeps the base `[first_def,last_use]`
   linear interval but *extends* it via the CFG live-in/live-out dataflow hull
   (`x86_regalloc.ax` §"CFG-aware liveness extension"). Extension only ever *grows*
   an interval, so a loop-carried value's interval is grown to cover the whole loop
   region. Liveness therefore only ever **over**-approximates — it can add spurious
   interference (extra spills), never under-approximate, so it cannot cause the
   clobber-miscompile the blocker described.
2. **Empirical (`bin/t_loopcross.ax`):** a runtime-computed bound defined before
   a loop, live across it, used in the loop condition, alongside seven independent
   loop-body temps pressuring the allocator — compiles to the identical correct
   result at -O0/-O1/-O2/-O3.

The real reason naive LICM crashed was the **non-SSA multiply-defined hoist**
(RFC 0025 §1 item 3), not liveness. RFC 0025 model-A's single-def gate addresses
that directly, and sound LICM is now shipped at -O2/-O3 (see RFC 0025 §0). No change
to the liveness core was needed.

## 8. Status

**RESOLVED — P2' + P3 SHIPPED (2026-07-09). BUG#86 CLOSED.** Short-circuit `and`/`or` now evaluate lazily per spec; the CFG-aware liveness foundation (P2') plus the `lower_while` CFG-edge fix (found during P3) make the diamond self-host. Final fixpoint `c777ef7b`.

History / empirical results:
- Short-circuit lowering is correct in isolation.
- P2 RPO block serialization self-hosts (B==C) but **regresses `t_math` 127→124** (loop-carried float) → proved reordering alone is insufficient; the liveness algorithm itself had to become CFG-aware. Reverted.
- **P2' (CFG dataflow) — SHIPPED, all green:** A!=B → B==C fixpoint (`BCEFC38F…`); regression 114/114; `t_math`=127 (canary correct); control-flow oracles pass. Kept the base linear intervals and only *grow* them via dataflow hull, so the change is safe-by-construction.

Next actionable step: **P3** — re-apply the short-circuit `and`/`or` lowering from [[bug86-short-circuit-open]] on top of P2'. Expected gate: `scB` compiles trivial + B==C + regression + `sc1/sc2/sc3=42` + `scv=3` + the guard idiom `if p != null and p[0] > 0` no longer faults. P1 (terminator normalization) remains a no-op precursor (invariant already holds).

🤖 Generated with [Claude Code](https://claude.com/claude-code)
