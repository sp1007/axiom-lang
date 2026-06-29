# tests/mathlib — careful accuracy verification for the math library

Math results must be verified carefully (independent oracle, many points, tight
tolerance), not just a couple of exit-code sanity checks.

## Tools
- **`gen_verify.py`** — oracle generator. Computes reference values with Python's
  `math` across a grid per function and emits a bundled AXIOM validator
  (std.math + unrolled relative-error checks; exit = number of functions that
  exceed their tolerance). Use for a broad sweep:
  `python tests/mathlib/gen_verify.py > sweep.ax` then `cat std/math.ax sweep.ax`
  and build. NOTE: keep the generated `main` modest — a very large single
  function is slow to compile and can hit the regalloc spill-all path.
- **`accuracy_probe.ax`** + **`run.sh`** — fast curated gate. Reference bands are
  oracle-computed; checks are in *argument position* (no many-live-float locals,
  see BUG#48). `bash tests/mathlib/run.sh` → asserts exit 127.

## Verified (2026-06-29)
std.math `sqrt/sin/exp/erf/asin/cbrt/tgamma` (and by extension the elementary
family) are accurate to ~1e-5 at the sampled points. The pure-AXIOM
transcendentals and the erf/Lanczos-gamma approximations are sound.

## Known limitation (NOT a math error) — BUG#48
Holding **≥8 float call-results in locals simultaneously** corrupts a spilled
one (register-pressure float-spill bug in x86_regalloc; same family as BUG#46).
The math formulas are correct — only this codegen edge case is affected. Verify
in argument position; avoid ≥8 simultaneously-live float locals until BUG#48 is
fixed (a backend keystone needing a fixpoint verify). See knowledge/bugs.md.
