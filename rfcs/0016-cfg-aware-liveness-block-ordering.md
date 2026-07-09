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
- **P2' (revised) — CFG-aware liveness is REQUIRED, not optional.** Replace the linear-index interval computation with proper per-block live-in/live-out dataflow (backward fixpoint over the CFG: `live_in[b] = use[b] ∪ (live_out[b] − def[b])`, `live_out[b] = ∪ live_in[succ]`), then build intervals over a block-linearized order. This makes correctness independent of block emission order and handles loops, diamonds, and mid-expression blocks uniformly. Gate: **A!=B → B==C mandatory** + full regression (MUST include `t_math`, the loop-carried-float canary) + control-flow oracles.
- **P3 — enable short-circuit `and`/`or`:** re-apply the lowering from [[bug86-short-circuit-open]] on top of P2'. Gate: `scB` compiles trivial + B==C + regression + `sc1/sc2/sc3=42` (crash-based) + `scv=3` (truth table). Ship BUG#86.
  - **Proven prerequisite met:** the short-circuit *lowering* is correct (isolation test), and the P2 experiment additionally proved a self-hosting short-circuit compiler is *achievable* once liveness is sound — so P3 is purely gated on P2'.

## 6. Alternatives

- **A. Keep eager `and`/`or` forever, document the spec deviation.** ✗ Violates spec (`docs/tasks/p09-t06`, p08-t04) and the natural `p != null and p[0]` idiom faults. Only acceptable as status quo, not a resolution.
- **B. Lower short-circuit without new blocks** (e.g. `cmov`-based select of both eagerly-evaluated sides). ✗ Defeats the purpose — the RHS still executes (faults/side-effects). Not short-circuit.
- **C. Special-case: only short-circuit when the RHS is provably side-effect-free and fault-free.** ✗ Loses the main benefit (avoiding the faulting-RHS case), and "fault-free" is undecidable in general (a deref can fault).

## 7. Success criteria

- P1: terminator invariant holds/enforced; A==B; regression green.
- P2: RPO serialization; B==C fixpoint; regression green; control-flow-heavy oracles pass; no perf regression in self-build time beyond noise.
- P3: short-circuit shipped; `scB` self-hosts; crash-based + truth-table oracles pass; the guard idiom `if p != null and p[0] > 0` no longer faults.

## 8. Status

DRAFT — root cause diagnosed, reproduced, and **P2 (RPO) experimentally falsified as a sufficient fix** (2026-07-09). Key empirical results:
- Short-circuit lowering is correct in isolation.
- RPO block serialization makes a short-circuit compiler **self-host** (B==C) — proving the goal is reachable — but **regresses `t_math` 127→124** (loop-carried float), proving `compute_liveness` must become genuinely CFG-aware (P2'), not merely fed a better block order.
- Reverted cleanly; tree at BUG#88 fixpoint `0D672CC8`.

Next actionable step: **P2' (CFG-aware live-in/live-out dataflow)**, the real fix, with `t_math` as the mandatory loop-carried-value canary in the gate. P1 (terminator normalization) remains a low-risk independent precursor but investigation shows the invariant already holds (`ensure_return` + per-construct jumps). This RFC exists so the foundational change is designed deliberately rather than patched ad-hoc (which broke self-host twice: the naive short-circuit vreg, and RPO-only).

🤖 Generated with [Claude Code](https://claude.com/claude-code)
