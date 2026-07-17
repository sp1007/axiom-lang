---
name: bug-o2-large-code-miscompile
description: "FIXED 2026-07-17 by disabling loop_unroll_func. Bisected: a compiler built at -O2 SEGFAULTED on every input; the culprit is loop_unroll (full-unroll of constant-trip-count<=4 loops), NOT strength_reduction (proven safe) and NOT LICM (already off). loop_unroll has correctness bugs on non-SSA AIR beyond a stale-instrs_start fix; disabled pending an RFC-scale rewrite. -O2-built compiler now passes the full regression."
metadata:
  node_type: memory
  type: project
  originSessionId: 002335d7-bcf4-411a-bdcf-96e89f5848ff
---

> **UPDATE `f885c75` (2026-07-17 pm):** `loop_unroll_func` is NO LONGER disabled — it was
> rewritten soundly and RE-ENABLED. Root cause was the per-iteration rename-map RESET (an
> accumulator read the original value each iteration → miscompile); fixed with a persistent
> `cur[]` value map + copy-back + iv-final, plus conservative gates (unroll only unique-pre-header/
> 2-pred-header/single-block pure-register bodies). The -O2-built compiler passes 364/364 with it
> on. See [[rfc0025-licm-shipped]]. The bisection notes below stand as the diagnosis history.

# ✅ FIXED 2026-07-17 — culprit = loop_unroll_func (now soundly RE-ENABLED, see update above)

**Bisected** (build compiler at -O2, toggle passes): with both strength_reduction and
loop_unroll off → -O2 compiler WORKS; strength_reduction only → WORKS; loop_unroll only →
SEGFAULTS. So **`loop_unroll_func` is the sole culprit**; strength_reduction is safe and
stays enabled. Fix = **disable the `loop_unroll_func` call** in `SsaOptimizer.run` (like
licm_func), with a note. -O2-built compiler then passes the full regression suite.

## Why loop_unroll miscompiles (partial root-cause)
It full-unrolls a constant-trip-count (≤4) loop by cloning the body into the pre-header
with fresh-register renaming (`next_reg++`) + a loop-carried copy-back, on NON-SSA AIR.
A first fix — reading the body block's `instrs_start` FRESH each access instead of via a
captured `body_blk` (stale across the `insert_inst_at` calls that shift block positions,
same class as the licm/insert_inst_at bug) — did NOT resolve the -O2 segfault, so there is
at least one more bug (suspected: the loop-carried copy-back condition
`def_block[dest] != loop_body_block_id` skips accumulators reassigned in the body → post-loop
reads a stale reg; and/or the unreachable original header/body blocks left after the
pre-header is rewired to jump straight to the exit). A correct re-enable needs a careful
rewrite (SSA or verified renaming/copy-back) = RFC-scale. Reverted the exploratory
fresh-read edit; shipped the clean disable.

## Original report (still valid context)

Found 2026-07-17 while attempting RFC 0025 (re-enable LICM). Distinct from the
already-fixed -O2/-O3 **compile-time** crash [[bug-o2-o3-loop-compile-crash]] — that
one made the compiler SEGV *while building* a loop; this one is a **run-time
miscompile of the produced binary** when the input program is large.

## Symptom
Build the compiler itself at -O2:
`bin/axc_native.exe build bootstrap/stage1/tmp_concatenated_air.ax -o axc_o2.exe -self-link -O2`
→ builds successfully (exit 0, ~2.38 MB exe), but the **-O2-built compiler SEGFAULTS on
every input** (`axc_o2.exe build any.ax ...` → SIGSEGV, exit 139). The -O1-built daily
driver compiles the same inputs fine (346/346). So -O2 codegen is wrong for something in
the compiler's own ~800-function codebase.

## Scope / isolation
- Reproduces with **LICM disabled** (the shipped state) — so it is NOT the LICM work; it
  is strength_reduction / loop_unroll / regalloc-under-pressure on **mid-size** functions.
  (Functions > OPT_LARGE_FUNC_THRESHOLD=2000 insts skip structural passes, so the culprit
  is a < 2000-inst function that uses a loop/pattern the -O2 passes mishandle at scale.)
- Small/medium programs compiled at -O2/-O3 are CORRECT (17-program differential battery
  O0==O1==O2==O3, + regression guards t_licm/t_loopcall/t_loopstruct@-O2/-O3). So the bug
  needs the register pressure / function-count / pattern mix of a large program to surface.
- Likely related to the fragile linear-interval register liveness [[rfc0016-p2prime-cfg-liveness]]
  (RFC 0016) — loop-crossing values in higher-pressure functions.

## Priority / impact
LATENT: the daily driver, self-host, and all benchmarks use **-O1**, which is solid.
Nobody builds real programs at -O2 yet. But it means -O2/-O3 are NOT trustworthy for large
programs, and it **blocks the acceptance test for RFC 0025** (can't validate LICM by
building the compiler at -O2 when -O2 is already broken on large code).

## Repro / bisect recipe
1. `& scripts/build_native.ps1` (daily driver, -O1).
2. `bin/axc_native.exe build bootstrap/stage1/tmp_concatenated_air.ax -o /tmp/axc_o2.exe -self-link -O2`
3. `/tmp/axc_o2.exe build bin/t_licm.ax -o /tmp/x.exe -O1; /tmp/x.exe; echo $?` → 139 (segfault).
4. Bisect: in `SsaOptimizer.run`, disable strength_reduction_func / loop_unroll_func one at
   a time (regen+build), rebuild axc_o2, retest. Whichever, when disabled, makes the
   -O2-built compiler WORK is the culprit pass. Then dump-air the miscompiled function.
