---
name: perf-immediate-operand-folding-inert
description: "In-flight `cmp/add/sub reg,imm` immediate-operand folding (const_imm32) in x86_selector was INERT and got REVERTED; complete root-cause + the sound fix a future M6-perf session must implement."
metadata:
  node_type: memory
  type: project
---

# Immediate-operand folding (const_imm32) — diagnosed INERT, REVERTED 2026-07-18

A prior session left UNCOMMITTED in-flight work in `bootstrap/stage1/x86_selector.ax`:
an `cmp/add/sub reg, imm` immediate-operand folding pass (`const_imm32` helper +
folds in `select_comparison`/OP_IADD/OP_ISUB + `remove_dead_imm_loads` post-pass),
targeting the **M6 fib-class perf gate** (fib 2.44x vs clang — see [[m6-perf-gate-fib-benchmark]]).
I gated + investigated it; it was **completely inert** (zero measurable effect) AND its
obvious completion is **unsound**, so I **reverted to clean `f885c75` / daily driver `97A0703F`**
and banked this diagnosis. The revert is NOT a loss — the diagnosis below makes a future
sound implementation cheap.

## What was verified
- **B==C held** (A==B==C byte-identical when the folding compiler builds the folding source) —
  SAFE, deterministic. Full regression **365/365** (incl. a new `t_immfold` oracle). So the change
  never miscompiled — it just did **nothing**.
- **Byte-identical output** on fib at -O0/-O1 AND on the compiler self-build. The folding never
  changed a single emitted byte on real programs.

## Root cause (fully nailed down, evidence-backed)
The fold path IS live and the emitter DOES honor immediate operands (proof: forcing
`const_imm32` to `return 5` made fib **segfault 139** instead of 231 — the emitter emitted the
folded imm form). The emitter already encodes the compact form: `x86_emitter.ax` MACH_ADD/SUB/CMP
branch on `inst.src1.kind == OPND_IMM` → `x86_encode_{add,sub,cmp}_ri`. So plumbing is complete.

The bug: **`const_imm32` stops its def-chain walk at `OP_CAST`**, but small integer literals reach
an arithmetic/compare op **through a widening `OP_CAST`** (the literal is emitted as a default-width
`iconst` then cast to the op's type, e.g. `i32 1 → cast i64 → isub`). `dump-air` COLLAPSES this cast
in its printed view (shows `isub %1, %4` with `%4 = iconst` directly), which is why it looked like a
direct iconst. The proven-working sibling `const_divisor_pow2` (div/pow2 strength-reduction, shipped
`38a9d81`) **does** follow `OP_CAST` (line ~980) — that's the ONLY behavioral difference, and it's
exactly why div-strength-reduction fires but immediate-folding didn't.

- Adding `OP_CAST` to `const_imm32`'s chain-follow → folding **fires** (fib bytes change, code
  shrinks, exit still 231 correct). CONFIRMED root cause.

## Why the obvious fix is UNSOUND (do not just add OP_CAST)
Blindly following `OP_CAST` reads the **pre-cast** constant, which is WRONG for **narrowing** casts:
`b == (300 as u8)` must fold `44` (=300 as u8), not `300`. Confirmed: with naive cast-follow,
`t_castwidth` returns **13 (want 15)** — a silent miscompile (the exact bug class this project treats
as cardinal). The RHS `300 as u8` folded as `cmp b, 300` instead of `cmp b, 44`.

## The sound fix a future M6-perf session must implement
Follow `OP_CAST` only when it is **value-preserving** for the fold. Candidate minimal-sound rule:
follow the cast iff the cast **target type is an 8-byte integer** (i64/u64) — then a constant already
gated to `[-2^31, 2^31)` widens without changing value, and the op runs at 8-byte width so the
sign-extended imm32 matches. This captures the fib/i64 hot case and rejects all narrowing (u8/u16/i8/
i16) and the conservative i32-target case. Subtleties to handle carefully (each is a potential silent
miscompile): signed↔unsigned target (i32→u64 of a negative value), sub-8-byte op width, and threading
the type table (`const_imm32` currently takes only `fn_ptr`; `sel_type_is_8byte_int`/`sel_type_is_unsigned`
need `sel` or a type-id-only width helper). REQUIRES new oracles: (a) fold-through-widening-cast correct,
(b) fold-through-narrowing-cast NOT applied / still correct. Gate = **B==C + full regression + -O2 acceptance**
(backend change). See [[m6-perf-gate-fib-benchmark]] and [[perf-div-pow2-strength-reduction]] (the sibling
that got cast-following right, but note ITS cast-follow is also latently unsound for narrowing — e.g.
`x / (256 as u8)`=div-by-0 would be mis-strength-reduced; worth hardening together).

## Artifacts kept
- `bin/t_immfold.ax` + regression row `t_immfold|exit|42` — a general constant-arithmetic /
  constant-folding correctness oracle (independent of the reverted feature; kept as added coverage).

## LESSON
Measure a perf "optimization" against a CLEAN baseline before believing it. My first fib compare was
folding-build vs folding-build (both had the inert change) → falsely read "inert = no effect at all";
the real inertness was cast-guarded. Always diff against the committed clean compiler, not another
build of the same in-flight source.
