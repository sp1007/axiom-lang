---
name: rfc0025-licm-shipped
description: RFC 0025 SHIPPED bd6616e — sound model-A LICM re-enabled at -O2/-O3; RFC 0016 liveness blocker-1 was a misdiagnosis (liveness is CFG-aware/conservative).
metadata: 
  node_type: memory
  type: project
  originSessionId: c3655292-a9ba-4316-abf6-b7660cb40ed5
---

**RFC 0025 Loop-Invariant Code Motion RE-ENABLED at -O2/-O3 — SHIPPED `bd6616e` (2026-07-17).**
`licm_func` in `bootstrap/stage1/ssa_opt.ax` is active again in `SsaOptimizer.run` (the
-O2/-O3 structural-pass block). Model-A design (sound on NON-SSA AIR without building SSA):

1. **Whitelist `is_hoistable_op`**: only pure, fault-free ALU ops (IADD/ISUB/IMUL,
   FADD/FSUB/FMUL, compares EQ..GE, bitwise AND/OR/XOR/SHL/SHR, NEG/NOT). EXCLUDES constants
   (ICONST/FCONST — no compute benefit + needless loop-crossing register pressure), ALL memory
   ops (LOAD/GET_FIELD/INDEX — may fault), IDIV/IMOD/FDIV (trap), casts (src2 may be a width),
   side-effecting ops.
2. **Single-def gate** (the non-SSA soundness key): hoist only if dest AND every register
   operand are defined EXACTLY ONCE in the whole function, operands defined OUTSIDE the loop.
   `def_count[reg]==1` makes `def_block[reg]` unambiguous. Induction/accumulator vars are
   multiply-defined → never hoisted. This closes the hole that made the OLD naive pass segfault
   at RUNTIME (it hoisted one def of a multi-def reg → loop read stale reg).
3. **Unique dominating pre-header**: new `compute_loop_preheaders(f, preheader[])` maps each
   loop-body block to the pre-header of its INNERMOST loop, requiring the header have EXACTLY
   ONE predecessor with lower loop_depth (which then dominates it). Loops without one are
   skipped. Enables hoisting from BODY blocks, not just the header → actually useful.

**Blocker-1 ("RFC 0016 backend liveness — loop-crossing value miscompile") was a MISDIAGNOSIS.**
The shipped P2' `x86_regalloc.compute_liveness` IS CFG-aware: base `[first_def,last_use]` linear
interval + a CFG live-in/live-out dataflow HULL that only ever *grows* intervals. So it
over-approximates loop-crossing values (extra spills at worst), never under-approximates → it
CANNOT cause the clobber-miscompile the blocker described. Proven empirically by
`bin/t_loopcross.ax` (runtime bound live across a loop, used in the condition, 7 competing body
temps → identical correct result O0..O3). No liveness change was needed. Constants stay excluded
from hoisting NOT for liveness safety but because immediates are rematerialized for free.
Blocker-2 (no large-scale acceptance test) had already been resolved by `570d5cb` (disable
`loop_unroll_func`, the separate -O2-large-code miscompile [[bug-o2-large-code-miscompile]]).

**Acceptance gate (RFC 0025 §4 — MANDATORY, not the fixpoint).** Self-build fixpoint is BLIND
to LICM because self-build uses -O1 (LICM never runs during it). The real gate:
- fast fixpoint **A==B** (`A4B63601…`) — -O1 self-build stays deterministic.
- daily-driver full regression **355/355**.
- **the -O2-BUILT compiler passes the full regression 355/355, 0 failures** ← the definitive test.
- **B==C**: the -O2-built compiler builds source → byte-identical to the daily driver.
- `dump-air` confirms LICM active: `x*y` hoisted to pre-header, `+7` const stays in loop.

Oracles banked in `scripts/regression_repros.sh` (O0..O3): `t_licmhoist` (must-hoist positive,
exit 235), `t_loopcross` (liveness loop-crossing, exit 10). Both under `bin/`.

**How to re-run the acceptance test after any optimizer change:**
`bin/axc_native.exe build bootstrap/stage1/tmp_concatenated_air.ax -o bin/axc_O2built.exe -self-link -O2`
then `AXC=bin/axc_O2built.exe bash scripts/regression_repros.sh` must be 355/355.

**FOLLOW-UP `f885c75` — LICM chaining + sound loop unroll (daily driver `97A0703F`, 364/364).**
- **LICM invariant-chain hoisting:** `copy_prop_func` now runs immediately BEFORE `licm_func` in
  the -O2/-O3 structural block. A lowered chain `t1=a*b; t2=t1+c` puts a copy between the links;
  collapsing it first lets LICM's within-pass chaining hoist BOTH links (dump-air verified).
  Oracle `t_licmchain`. Residual (documented, not a bug): a link whose OPERAND vreg is REUSED in
  the loop (non-SSA, def_count>1) is soundly declined → needs reaching-defs (model-B).
- **Loop unroll re-enabled (`loop_unroll_func`, const-trip ≤4):** fixed the loop-carried threading
  (persistent `cur[]` value map instead of per-iteration reset — the old reset made an accumulator
  read the ORIGINAL value every iteration → -O2-built compiler segfault) + copy-back each
  body-defined reg's final value + set iv to post-loop value. Plus CONSERVATIVE gates (forced by
  the acceptance test: threading-fixed still segfaulted on the compiler's own loop shapes): unroll
  ONLY a unique-pre-header, 2-pred-header, single-block body whose sole successor is the header and
  that is pure straight-line register arithmetic (no branch/call/memory/concurrency op). Oracle
  `t_licmunroll` + edge probes (multi-accum, trip-1, iv-after-loop, step-2). Closes
  [[bug-o2-large-code-miscompile]] for loop_unroll. `loop_unroll` is NO LONGER disabled.

**Post-ship probe hardening (`df512d3`):** 10 LICM edge-case programs (nested/triple-nested
loops, break/continue, calls-in-loop, float & bitwise invariants, invariant chains, a
multi-def-operand candidate that must NOT hoist, a zero-trip loop for speculation safety, and an
invariant loop bound) all correct O0..O3. Found ONE correctness-neutral effectiveness gap:
invariants behind an OP_COPY (LICM runs before copy-prop) only hoist their first link. Tried
whitelisting OP_COPY — safe but gave no extra motion, reverted. Documented in `is_hoistable_op`.
Future fix if a loop-heavy workload needs it: run copy-prop before LICM, or iterate LICM in the
value-pass loop (both = pipeline changes needing a full gate).

Related: [[bug-o2-o3-loop-compile-crash]] (insert_inst_at physical-order + ICONST src2 guard —
kept), [[bug-o2-large-code-miscompile]] (loop_unroll still DISABLED), [[rfc0016-p2prime-cfg-liveness]].
Still DISABLED: `loop_unroll_func` (RFC-scale rewrite needed).
