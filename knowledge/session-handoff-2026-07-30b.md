---
name: session-handoff-2026-07-30b
description: "HANDOFF 2026-07-30b — peephole 1f (non-adjacent copy-chain collapse) is BUILT, PRICED at -14.2% on the tail-recursive shape, and RED: it miscompiles FLOAT PARAMETER prologues. Change is in the working tree UNCOMMITTED. Driver restored to a clean HEAD fixpoint 6C9165C8."
metadata:
  node_type: memory
  type: project
---

# HANDOFF 2026-07-30b — 1f built + priced + RED on float params. DO NOT COMMIT THE SELECTOR AS-IS.

## State of the tree
- **HEAD = `fa9798a`** (unchanged; nothing from this session is committed except the oracle below).
- **Daily driver `bin/axc_native.exe` = `6C9165C8`** — a CLEAN HEAD fixpoint (R2==R3, rebuilt twice
  from HEAD source). Trust it.
- **BASELINE = 578 / 578, 0 failed** — verified on this driver AFTER the commit, not assumed. Was
  575; `t_copychain` adds 3 rows (main list at -O1, plus the O2 and O3 sweep). A drop below 578 is
  RED.
  ⚠️ Mid-session the driver was briefly the 1f build (`1FFD5C2C`), which miscompiles float code. It
  was replaced. If any binary in `bin/` looks suspect, rebuild it — it may have been produced by
  that driver.
- **`bootstrap/stage1/x86_selector.ax` is MODIFIED and UNCOMMITTED (+165 lines)** = peephole 1f.
  It is **RED**. Do not commit it without fixing the float-param defect below.
- `CLAUDE.md` is also modified in the tree. **That edit is NOT mine** (it was clean at session
  start; it refines the context-hygiene rule). Leave it for the user.

## What shipped (committed, GREEN)
`bin/t_copychain.ax` (exit 42) + wiring in `scripts/regression_repros.sh` (main list + O2/O3 sweep).
A VALUE-checking oracle for copy-chain folds: `sumto`(55), `twin`(10), `walk`(7). Passes O0–O3 on
the clean driver, so it is a valid pin on HEAD behaviour today and the acceptance test for 1f.

⭐ **It is CALIBRATED, and the calibration produced the session's most useful fact.** Flipping either
vD-interference check to `ok = true` makes it return **1 at -O1** (sumto gives 45, not 55). Meanwhile
**`bin/t_tailrecloop.ax` still returns its expected 128 with the guard broken** — the benchmark is
value-BLIND to this exact miscompile. The bench program could never have gated this work.

## Peephole 1f — what it is, and it DOES fire
`collapse_copy_chain` in `x86_selector.ax`, wired BEFORE 1c (`coalesce_dest_copy`):

    DEF vT,...  (MOV|LEA) ; <k insts touching neither vT nor vD> ; MOV vD,vT   ->   DEF vD,...

Shape B of 1c with adjacency lifted; `gap >= 2` keeps the two passes disjoint so deltas stay
attributable. Guards: `counts[vT]==2` (whole-function, the single criterion), span scan rejecting
any surviving mention of vD, bail on control flow / calls / div-family, `CHAIN_WINDOW = 8`.

⭐ **Proven to fire on the intended shape, not merely "somewhere"** (the trap that made the first
version of 1d match 0 constants while passing every gate). Instrumented count on
`bin/t_tailrecloop.ax`: 84 folds, and the two in `sumto` are exactly the predicted links —
`i=14 j=17 vt=6 vd=10 op=LEA` and `i=18 j=20 vt=11 vd=2 op=MOV`.

⭐⭐ **The ordering hazard is respected in practice.** The `v10 -> v1` link is ABSENT from the fold
log: `ADD v7,v1` reads the OLD v1, and the span scan refuses it. That refusal is what keeps the
answer correct, and it was verified by observation, not by reading the code.

## Measured (paired, alternating, per-pair sign — not best-of-all)
- **`t_tailrecloop`: -14.0% / -14.3% / -14.2%** across three tight pairs (pair 0 discarded as a
  cold-start outlier at 31.3 ms). Same exit 128 both builds. **The win is REAL.**
  Predicted was -24% for reaching the 5-instruction floor; we land at ~6 instructions because the
  LEA is not SUNK below the ADD (that needs reordering, not a peephole). 22.1 -> 19.0 ms against a
  16.2 ms floor, i.e. **1.36x -> 1.17x**.
- **fib +14.5% / +14.5% / +28.1%** and **callloop +8.6% / +6.0% / +3.0%** — consistent sign, REAL
  for these binaries. xorshift and arrwalk flat (cross zero => noise).
- ⭐ **The fib/callloop cost is LAYOUT, and this is proven rather than assumed.** All four bench
  shapes report exactly **80 folds** — identical across four unrelated programs, so every fold is in
  bundled code and **zero** are in the benchmarks' own functions. Confirmed by diffing objdump
  output: fib's instruction stream is **IDENTICAL** between builds (total .text 6514 -> 6478, i.e.
  the 36 removed instructions are all upstream). fib and callloop call no stdlib in their hot loops,
  so unchanged code cannot have got slower — only its address did. Same class as the peephole-1e
  alignment story, second instance.

## ⛔ WHY IT IS RED — and the exact next step
Full regression: **575 passed, 3 FAILED** — `t_interpolation` (127 -> 79), `t_colorhsl` (127 -> 120),
`t_quatrot` (8 -> 3). Identical failures at O0/O1/O2, so not an SSA/opt-level interaction.
Confirmed caused by 1f: a reference compiler built from HEAD source returns 8/127/127.

`t_interpolation`'s 79 = bits 1+2+4+8+64, so exactly `ip_catmull_rom` (16) and
`ip_bezier3`/`ip_bilerp` (32) failed — the functions with the most simultaneously-live f64 temps.

**Localized by instrumentation (`CHAINFN len=83 maxv=57`, the catmull_rom function):**

      i=0 j=4 vt=45 vd=1        i=1 j=5 vt=46 vd=2
      i=2 j=6 vt=47 vd=3        i=3 j=7 vt=48 vd=4

`vd = 1,2,3,4` are **PARAMETER vregs**, and this is the **parameter prologue**: `MOV vTemp,<xmm arg
reg>` … `MOV vParam,vTemp`, NON-adjacent because the four parameters interleave. The 5th param is
untouched (Windows x64 passes only 4 in registers).

⇒ **1f reaches into the ABI parameter prologue, and on FLOAT params that miscompiles.** Note that
1c's shape B already folds this same sequence in its ADJACENT form and is fine, which is why the
defect is specific to the widened window.

**Two candidate mechanisms, NOT yet distinguished — settle this BEFORE writing a guard:**
1. **Register class.** `is_float_vreg` (`x86_regalloc.ax:1398`) classifies vregs 1..params.len from
   `fn_ptr.params`, but any OTHER vreg by scanning AIR for the inst whose `dest == vreg`. The
   deleted temps (v45..v48) are selector-invented and have no AIR def, so `is_float_vreg` returns
   FALSE for them — they classify as GPR. Understand what the prologue relies on there before
   changing it.
2. **Float register pressure.** The failing functions are precisely the float-pressure-heavy ones,
   and 1f moves a destination's def up to 8 instructions earlier. This is the shape that killed
   forward copy-propagation (-6.5% fib) and, in 2026-07-29d, "uncovered an allocator bug". It may be
   EXPOSING a latent float-spill defect rather than being wrong itself.

⚠️ Do NOT just exclude parameter vregs as fold destinations to make the suite green: **sumto's
winning `vt=11 vd=2` fold has a PARAMETER destination too** (sumto's params are v1,v2), so that
guard deletes half the measured win. The discriminator is float-ness, not param-ness — and a guard
whose mechanism is unexplained is exactly what `f3e3915` warns against.

Reproduce in one build with:

    fn ip_catmull_rom(p0: f64, p1: f64, p2: f64, p3: f64, t: f64) -> f64  (body from bin/t_interpolation.ax)

called as `ip_catmull_rom(0,0,1,1,0.5)`, expect 0.5. REF gives 42 / 1f gives -62 on the small
harness. Recreate it as `bin/zz_fmin3.ax` (deleted as scratch).

## Method notes worth keeping
- **A reference compiler is cheap and decisive**: `git show HEAD:bootstrap/stage1/x86_selector.ax`
  over the file, regen, build, swap back. Behaviour follows SOURCE, not which binary built it — one
  build turns "is this mine?" into a fact.
- ⚠️ But its BYTES come from whatever built it. After a buggy driver, restore by rebuilding HEAD
  twice and checking the fixpoint (R2==R3), which is what produced `6C9165C8`.
- The bench binaries have **no per-function symbols** (`objdump` shows only `.text`), so per-function
  attribution needs normalized whole-stream diffs plus instruction counts, or in-compiler prints.
  There is no `-S`/asm-dump driver flag.
