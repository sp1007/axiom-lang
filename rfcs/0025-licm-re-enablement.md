# RFC 0025 — Loop-Invariant Code Motion re-enablement (sound LICM on non-SSA AIR)

Status: **DRAFT** (2026-07-17). Supersedes the ad-hoc `licm_func` that shipped disabled in commit `3703d52`. Blocks: nothing (LICM is currently OFF and the pipeline is sound without it); enables a future perf lever for loop-heavy code.

Related: [[bug-o2-o3-loop-compile-crash]], [[bug-cse-redef-operand-miscompile]] (non-SSA operand hazard), RFC 0016 (CFG-aware liveness — a prerequisite building block).

---

## 1. Background — why LICM is disabled today

`ssa_opt.licm_func` (`bootstrap/stage1/ssa_opt.ax:599`) is present but its **call is commented out** in `SsaOptimizer.run` (the -O2/-O3 structural-pass block). It was disabled in `3703d52` after being found to both crash the compiler and, once the crashes were fixed, silently miscompile.

Three distinct defects were peeled apart (full trace log in the bug memory):

1. **`insert_inst_at` physical-vs-index adjustment** (FIXED in `3703d52`, kept). The helper adjusted block `instrs_start` by *block index* while `block_instrs` is laid out in *physical* order; a pre-header with a higher block index than its loop body corrupted a block's instruction window → OOB → compiler SEGV. Now adjusts by `instrs_start >= insert_pos`. Shared by strength_reduction/loop_unroll, so this fix stands regardless of LICM.

2. **ICONST/FCONST immediate indexed as a register** (FIXED in `3703d52`, kept). LICM's src2 invariance test indexed `def_block` with the packed 64-bit immediate of an `OP_ICONST`/`OP_FCONST` (e.g. `0xFFFFFFFF`) → wild OOB read → SEGV. Now excludes ICONST/FCONST/JUMP for src2, mirroring `max_reg_id`'s register-operand rules.

3. **Fundamental unsoundness on non-SSA AIR** (the reason LICM stays OFF). AIR is **not** SSA: a register is reassigned inside a loop (an induction variable `i = i + 1`, an accumulator `acc = acc + i`). `licm_func` decides a value is loop-invariant purely from where its *operands* are defined (`def_block[src].loop_depth < cur`), then hoists *one* definition into the pre-header and NOP-s the in-loop copy. That rewrite is valid **only under single-assignment**. On a multiply-defined register it moves/erases the wrong definition, so the loop body reads a stale/undefined register → the compiled program crashes at run time. Verified: a plain `while i<1000: acc=acc+i; i=i+1` sum loop segfaults at run time with LICM enabled, even though it compiles cleanly.

The same non-SSA operand hazard bit CSE before (`bug-cse-redef-operand-miscompile`, fixed by scan-down-stop-at-redef). LICM is a stronger form of the same problem: it does not just *reuse* a value, it *relocates a definition across a region where the register may be redefined*.

## 2. Goal / non-goal

- **Goal:** re-enable LICM as a *sound* -O2/-O3 pass that never miscompiles and never crashes the compiler, hoisting genuinely loop-invariant, side-effect-free computations into a true pre-header.
- **Non-goal:** a maximal LICM. A conservative pass that hoists only provably-safe instructions (and bails on anything it cannot prove) is acceptable and preferred (§10 correctness > cleverness). Loads/aliasing-sensitive motion are explicitly out of scope for v1.

## 3. Decision points (resolve before implementing)

**D1. Soundness model — how do we license moving a definition?** Two viable paths:

- **(A) Single-def gate (minimal).** Hoist a candidate instruction only if its destination register is defined **exactly once** in the whole function AND every operand's definition is outside the loop (loop_depth < cur) AND the pre-header dominates the loop header. Compute a per-register def-count in one linear pass (reuse `max_reg_id`'s scan shape). Reject (leave in place) any candidate whose dest is multiply-defined. This is small, obviously correct, and captures the common case (a truly loop-invariant temp computed once). It will NOT hoist an induction-derived value, which is fine — those are not invariant anyway.

- **(B) Real SSA construction (maximal, RFC-scale).** Build SSA (φ-nodes) for the function, run textbook LICM on it, then destruct SSA. Far more powerful and reusable (helps CSE, copy-prop, future GVN), but a large new subsystem with its own verification burden. Likely its own RFC.

  **Recommendation: ship (A) first.** It is a self-contained, verifiable win; (B) is a separate epic that (A) does not preclude.

**D2. True pre-header / dominance.** The current "any predecessor with lower loop_depth" heuristic is unsound — such a predecessor need not dominate the loop (could be a loop exit reached on another path, or a block that only sometimes precedes the header). v1 must either (i) require a *unique* predecessor with lower loop_depth that also dominates the header, or (ii) **materialize** a dedicated pre-header block (single edge into the loop header) when one does not already exist. Materializing is the standard fix and interacts with RFC 0016's terminator/ordering invariants — coordinate. If a unique dominating pre-header cannot be established, **bail on that loop** (hoist nothing).

**D3. Speculation safety.** Only hoist instructions with **no side effects and no trap potential**. `has_side_effect` already excludes STORE/CALL/SYSCALL/etc. Additionally exclude anything that can fault when executed unconditionally: division/modulo (div-by-zero if the loop never ran), and **loads** (address may only be valid inside the loop). v1: hoist only pure arithmetic/bitwise/compare/const on register or immediate operands. No loads, no div/mod, in v1.

**D4. Insertion point.** Hoist to **before the pre-header's terminator** (already implemented — appending after the `JMP`/`JCC` into the header would be dead code). Keep the terminator-scan guard.

**D5. Determinism.** Iterate blocks/instructions in a fixed order; do not depend on hash iteration. Required for the fixpoint/reproducible-build guarantee.

## 4. Verification plan (mandatory before re-enable)

Backend/optimizer change → the fixpoint gate alone is **insufficient** (a mis-hoist need not change the compiler's own -O1 self-build, which does not run LICM). Required:

1. **Differential O1==O2==O3** on a loop battery (≥15 programs: nested, break/continue, early-return, call-in-loop, struct-field-mutation-in-loop, pow2 div/mod, data-dependent trip counts). The programs banked as `t_licm`, `t_loopcall`, `t_loopstruct` oracles (+ the -O2/-O3 guard block in `regression_repros.sh`) are the seed set — extend it.
2. **A "must-hoist" positive test.** Construct a loop with a genuinely invariant sub-expression and assert (via AIR dump or a perf/iteration-count proxy) it is actually hoisted — otherwise v1 could silently become a no-op and "pass" trivially.
3. Full regression GREEN + self-build fixpoint (`build_native.ps1`), on Windows and Linux (`elf_linux_check.sh`).

## 5. Drawbacks / alternatives

- **Do nothing (keep LICM off).** Zero risk. LICM's benefit on the current benchmarks is limited (`m6-perf-gate-fib-benchmark`): collatz's `while x!=1` has a data-dependent trip count (not the target), fib has no loop; the larger perf levers are inlining and recursion→loop. So this RFC is **low urgency** — schedule it behind higher-value perf work unless a loop-heavy workload appears.
- **Jump straight to SSA (D1-B).** More power but couples this to a large new subsystem; not justified until multiple passes want SSA.

## 6. Migration / compatibility

None user-visible. Pure -O2/-O3 codegen quality change; -O0/-O1 (daily driver, self-host) unaffected. The two crash fixes and the disabled-pass note already shipped; this RFC only governs turning the call back on under model (A).
