---
name: bug-variable-shift-in-loop
description: "FIXED 2026-07-19 5a22500 (B==C af3ba8e3): a variable-amount shift `x << k` / `x >> k` whose COUNT is a loop induction variable miscompiled — const_shift_amount (x86_selector) chased the count vreg to the init ICONST and emitted a CONSTANT `SHL/SHR dst,imm8` with k's loop-ENTRY value, so every `1 << k` in a loop computed `1 << 0`. Root = missing non-SSA def-count guard that the sibling const_divisor_pow2 already had. LOAD-BEARING (compiler's own source hit it → A!=B). Fix = accept the immediate form only for a single-def count vreg, else fall back to the correct CL-form. Surfaced by autopilot bug-probing."
metadata:
  type: project
---

# Variable-shift-in-loop miscompile — FIXED 2026-07-19 (`5a22500`)

## Symptom
`1 << k` (or `256 >> k`) where `k` is a loop induction variable computed the WRONG
value at every iteration — the shift count was frozen at k's value on loop ENTRY.
`s += (1 << k)` for k=0..5 gave 6 (six 1s) instead of 63 (`1<<0`=1 each). O-level
INDEPENDENT (O0==O1==O2), so it was a base-codegen bug, not an optimizer bug — the
kind O0-vs-O1 probing cannot catch (needs a hand-oracle).

## Discriminators (bisect, driver 9A178747)
- `1 << 2` / `let k=3; 1<<k` (no loop) → correct. Single-shot variable shift is fine.
- `while k<6: s += (1<<k)` → 6 (bug). `1 << k` == `1 << 0` each iter (count read as 0).
- start `mut k := 2` → froze at `1<<2`=4 each iter → the stale value is k's LOOP-ENTRY value.
- constant count in loop (`1 << 2`) → correct. loop-INVARIANT var count (`let m=3; 1<<m`)
  → correct. **`1 * k` (MUL by loop var) → correct** — the loop var itself is read fine;
  the defect is SPECIFIC to the shift operator's count-operand classification.

## Root cause
`x86_selector.ax::const_shift_amount` chases a shift's count vreg back through
casts/copies to an OP_ICONST to select the fast `SHL/SHR dst, imm8` immediate form
(BUG#24: the CL-form clobbers RCX, which the regalloc doesn't model, so the immediate
form is preferred where possible). AIR is **non-SSA**: a vreg can have multiple defs.
A loop counter `mut k := 0; ... k = k + 1` has TWO defs (init ICONST + the OP_IADD
increment). The first-match instruction scan found the init ICONST first and returned
its value → the shift was emitted as a CONSTANT shift by the loop-entry value.

The sibling `const_divisor_pow2` (same file, added for div/mod strength-reduction)
ALREADY had the exact guard for this hazard — *"AIR is non-SSA: a first-match scan of a
re-defined vreg would pick a stale ICONST"* — it counts defs and bails if `!= 1`.
`const_shift_amount` was simply missing that guard.

## Fix (`5a22500`)
Add the identical def-count guard to `const_shift_amount`: before accepting the
immediate form, count definitions of the current chase vreg; if `def_count != 1`,
return -1 → the caller falls back to the correct variable CL-form
(`MOV RCX,count; SHL/SHR dst,%cl`), which genuine variable shifts already use
successfully. Shared by OP_SHL and OP_SHR (both call const_shift_amount).

## Why load-bearing (A!=B)
The compiler's own source contains a loop-variant shift the OLD (buggy) compiler
miscompiled, so the fixed source built by the old compiler (A) differs from the
fixed compiler building itself (B); **B==C=`af3ba8e3`** confirms the fixpoint. This is
exactly why a backend change requires the B==C gate, not just A==B.

## Gate
Backend change → B==C mandatory BEFORE commit. B==C=`af3ba8e3`, full regression
**439/439** (+`bin/t_shiftloop.ax` oracle registered @O1 and in the O2/O3 block, exit
63), repro validated O0/O1/O2/O3. Daily driver promoted `9A178747`→**`af3ba8e3`**.

## Lessons
- A consistent (O-level-independent) wrong answer is invisible to O0-vs-O1 probing;
  a HAND-oracle caught it (p4_shift expected 99, got 42 → decomposed to `1<<k`=1).
- When one helper is hardened against a non-SSA multi-def hazard, GREP for siblings
  doing the same first-match vreg chase — `const_shift_amount` and `const_divisor_pow2`
  are twins; only one had the guard. Related: [[bug-cse-redef-operand-miscompile]]
  (same non-SSA "stop at redef" class), [[perf-div-pow2-strength-reduction]].
