---
name: bug-o2-o3-loop-compile-crash
description: "FIXED 2026-07-17: compiler SEGFAULT while compiling any `while` loop at -O2/-O3 is gone. Two real crash bugs in the shared insert_inst_at/licm code fixed (physical-order instrs_start adjust; ICONST/FCONST immediate treated as register in def_block index). The unsound licm_func pass itself DISABLED (non-SSA hazard) — strength_reduction+loop_unroll remain and are verified sound. -O2/-O3 now compile AND run loops correctly."
metadata:
  node_type: memory
  type: project
  originSessionId: da03b89d-0ea2-4e5a-91f3-4ec2ce91f9c5
---

> **UPDATE 2026-07-17 (pm/later):** `licm_func` is NO LONGER disabled — it was
> rewritten as sound model-A and RE-ENABLED at -O2/-O3 in `bd6616e`. See
> [[rfc0025-licm-shipped]]. The two crash fixes described below still stand
> (shared by strength_reduction). `loop_unroll_func` remains disabled.

# ✅ FIXED 2026-07-17 (pm) — -O2/-O3 loop compile crash resolved

The compiler no longer SEGFAULTs building `while` loops at -O2/-O3, and the emitted
program runs correctly (differential O1==O2==O3 on 5 loop programs + regression
guard `t_licm@-O2/-O3` = 44). Daily driver still -O1 (unchanged output — this is
a -O2/-O3-only change, self-host A==B trivially held).

## What was actually wrong (three layers, peeled by tracing)
The original diagnosis (dangling `inst` ptr + stale `blk` copy) was only the
surface. Real chain, found by instrumenting `licm_func`/`insert_inst_at` with
`ax_printf_local` traces and rebuilding:

1. **`insert_inst_at` (ssa_opt.ax) adjusted `instrs_start` by BLOCK INDEX, not
   physical layout.** `block_instrs` is a flat array ordered by `instrs_start`
   (physical), which is NOT block-index order — an optimizer can make a loop's
   pre-header have a HIGHER block index than the loop body (`pre_header_id > bi`).
   The old `while adjust_bi = block_id+1..len` bumped the wrong blocks, leaving a
   physically-later but lower-index block's window misaligned → duplicate/garbage
   inst indices → OOB → compiler SEGV. FIX = bump `instrs_start` of every OTHER
   block whose `instrs_start >= insert_pos` (physical), regardless of index.
   This helper is shared by strength_reduction/loop_unroll, so the fix is general.
2. **licm src2 invariance test indexed `def_block` with an immediate.** For
   `OP_ICONST`/`OP_FCONST` the 64-bit constant is packed into src1+src2 (seen in
   trace: `op=513(ICONST) s1=0xFFFFFFFE s2=0xFFFFFFFF`). The check only excluded
   `OP_BRANCH`, so `def_block[0xFFFFFFFF]` was a wild OOB read → SEGV. FIX = also
   exclude ICONST/FCONST/JUMP for src2 (mirror `max_reg_id`'s register-operand
   rules). src1 already excluded them.
3. **licm_func is fundamentally UNSOUND on non-SSA AIR** — after 1+2 the compiler
   stopped crashing but the *program* segfaulted at runtime. AIR is non-SSA
   (a register is reassigned in a loop, e.g. `i=i+1`); LICM decides invariance
   from operand def-sites and hoists ONE def while NOP-ing the in-loop copy, valid
   only under single-assignment. On multiply-defined regs it erases/moves the
   wrong def → loop reads stale reg → runtime crash. So **`licm_func` is DISABLED**
   in `SsaOptimizer.run` with a full note at its definition. Re-enable path =
   **RFC 0025** (rfcs/0025-licm-re-enablement.md): conservative single-def gate +
   true dominating pre-header + speculation safety, mandatory differential
   O1==O2==O3 verification. Low urgency. [[bug-cse-redef-operand-miscompile]]

## Final state
- `insert_inst_at`: physical-position `instrs_start` adjust (correctness fix, kept).
- `licm_func`: ICONST/FCONST/JUMP src2 guard added; pass CALL disabled (unsound).
- `strength_reduction_func` + `loop_unroll_func`: ENABLED, verified sound
  (5 loop programs O1==O2==O3; they benefit from the insert_inst_at fix too).
- Oracles: `bin/t_licm.ax` (row `t_licm|exit|44` at -O1 + dedicated -O2/-O3 guard
  block in regression_repros.sh), `bin/t_floatcvt.ax` (closes the #2 float-cvt item
  [[bug-consecutive-float-cvt-call-regalloc]]).

## Bisect recipe (for future LICM re-enable work)
Comment the 3 structural calls (ssa_opt.ax `SsaOptimizer.run`) one at a time,
`& scripts/build_native.ps1`, then compile a loop at **-O2** and RUN it (both the
compiler crash AND the runtime miscompile must be checked — B==C is not enough, a
program-run oracle is mandatory for these passes).
