---
name: perf-div-pow2-strength-reduction
description: "SHIPPED 38a9d81: div/mod by constant power-of-two now emits shifts, not IDIV. Collatz 10.5x->2.89x vs clang. The #1 arithmetic codegen gap, closed. B==C 838F42AE, 342/342. Records the 3 self-host hazards that had to be solved."
metadata: 
  node_type: memory
  type: project
  originSessionId: da03b89d-0ea2-4e5a-91f3-4ec2ce91f9c5
---

# Div/mod-by-power-of-two strength reduction — SHIPPED `38a9d81` (2026-07-17)

`x / 2^k` and `x % 2^k` on 8-byte ints used to lower to a full `cqo; idiv` (~20-90
cycles); now emit shifts (~1 cycle). Found via a FAIR benchmark (collatz — clang cannot
constant-fold it): AXIOM was **10.5x** slower than clang -O2 purely from the two
per-iteration idivs. Strength reduction closes it to **2.89x** (3.6x faster). Benefits
ALL integer div/mod by a literal power of two — the dominant arithmetic gap (fib's ~2.4x
is separate: call-count / recursion->loop).

## Implementation (x86_selector.ax OP_IDIV/OP_IMOD)
- `const_divisor_pow2(fn,vreg)` -> k iff the divisor is a **single-def** ICONST == 2^k.
- Sequences computed in **RAX/RDX**, mirroring the IDIV path. Unsigned div: `SHR k`.
  Signed div (round-to-zero): bias = `(x<0)?(2^k-1):0` via `sar 63; shr 64-k`, then
  `(x+bias) sar k`. Mod: `x - (x/2^k)<<k` — stays in RAX+RDX, no mask register.
- `MACH_DIVCLOBBER` (opcode **45**): zero-byte marker emitted before the sequence;
  emitters render a nop, regalloc treats it like IDIV so vregs spanning it avoid RAX/RDX.
- Gate `sel_type_is_8byte_int` (i64/u64 only) sidesteps sub-64-bit dirty-upper hazards.

## THREE self-host hazards that had to be solved (each caused a distinct failure)
1. **Non-SSA multi-def divisor** (wrong value: t_numtheory 95 vs 127). AIR is non-SSA
   (see [[bug-cse-redef-operand-miscompile]]); a first-match def scan strength-reduced a
   later-reassigned runtime divisor by a stale power of two. FIX = require the divisor
   vreg to have exactly ONE definition, else fall back to idiv.
2. **Fresh vregs broke self-host** (B segfaulted in codegen). The first version computed
   into `next_vreg` temps; in the compiler's large SPILLING functions this interacted
   badly (B!=C, B crashed). FIX = compute in RAX/RDX physical regs (like idiv), no vreg
   temps. In spill-all all vregs are in memory so rax/rdx are free; in graph-coloring the
   DIVCLOBBER marker forbids rax/rdx for spanning vregs.
3. **Opcode collision** (silent f32 miscompile). MACH_DIVCLOBBER was first assigned 43,
   which was already MACH_CVTSS2SD (f32->f64 widen) — the emitter's `NOP or DIVCLOBBER`
   branch then turned every widen into a nop. **B==C still PASSED** (the compiler doesn't
   widen f32->f64 itself) but t_f32/t_f32argcoerce failed. Moved to opcode 45.
   LESSON (reinforces is_param): **B==C is NOT sufficient for a backend change — run the
   FULL regression.** Only the f32 oracles caught this.

## Bench recipe
`scratch/collatz.ax` / `scratch/c1.ax` (100k = fast), best-of-3 vs `clang -O2`
(`/c/msys64/ucrt64/bin/clang`). Oracle `t_divpow2` (=16) banked in regression_repros.sh
covers signed-negative, unsigned, k=1..6, and the non-pow2 idiv fallback.
Remaining perf gap to clang on collatz (2.89x) is call overhead + loop quality (clang
inlines collatz_len); see [[m6-perf-gate-fib-benchmark]].
