---
name: bug-consecutive-float-cvt-call-regalloc
description: "CLOSED (verified 2026-07-16 pm, daily driver ec8a0d0): two back-to-back calls each needing an f32<->f64 width conversion at the call boundary — the reported 'second call reads first call arg' no longer reproduces (explicit-as, identity id64/id32, and implicit-coercion forms all correct at O0 and O1). Likely fixed as a side effect of HOLE#6 05ef55f + float-lit-width ec8a0d0. (was: OPEN backend regalloc miscompile.)"
metadata:
  node_type: memory
  type: project
  originSessionId: 83ebf198-e937-49ec-a738-064db47952bb
---

# ✅ CLOSED (verified 2026-07-16 pm) — no longer reproduces on daily driver ec8a0d0

Re-ran the exact repros below on `bin/axc_native.exe` @ec8a0d0 (post HOLE#6 `05ef55f`
+ float-lit-width `ec8a0d0`): explicit-`as` two-call form → 10 (want 10), identity
id64/id32 variant → 37 (want 37), implicit-coercion form → 10 — all CORRECT at O0 and O1.
The regalloc miscompile is gone; likely incidental fix via the float-literal width-inference
work. Kept oracle candidates; no dedicated backend session needed. Original OPEN report below.

# (historical) OPEN — consecutive float width-cvt calls miscompile (xmm regalloc)

Found 2026-07-16 while shipping the HOLE#6 float f32↔f64 implicit-argument coercion
([[bug-freefn-stdlib-collision-noarg]]). This is a **separate, PRE-EXISTING backend
regalloc bug**, NOT introduced by the coercion fix.

## Symptom
Two consecutive function calls, each of which converts a float argument between f32 and
f64 at the call boundary (cvtss2sd / cvtsd2ss), miscompile: the SECOND call reads the
FIRST call's argument value (or garbage) instead of its own.

```
fn take64(x: f64) -> i64: return (x + 0.5) as i64
fn take32(x: f32) -> i64: return (x + 0.5) as i64
fn main() -> i64:
    let a: f32 = 3.0
    let b: f64 = 7.0
    let w = take64(a)   # widen f32->f64 at the call
    let n = take32(b)   # narrow f64->f32 — reads 3.0, returns 3 not 7
    return w + n        # 6, want 10
```

## Key evidence it is PRE-EXISTING (not the coercion fix)
- Reproduces on the OLD compiler (`bin/axc_native.exe` at 6eb63db) using EXPLICIT casts:
  `take64(a as f64)` then `take32(b as f32)` → 6, not 10. The implicit-coercion fix reuses
  the exact same OP_CAST → cvt path, so it merely makes the bug reachable without an
  explicit `as`.
- **Two same-width f64 calls (no cvt) are CORRECT** (`f(a); g(b)` = 10) → the trigger is
  specifically the width-conversion at the call, not two float calls per se.
- **Order-sensitive:** narrow-first-then-widen (`take32(b)` then `take64(a)`) is CORRECT
  (10); only widen-first-then-narrow fails. Suggests the first call's cvt result / arg reg
  isn't preserved across the call and collides with the second.
- **A filler statement between the calls MASKS it:** inserting `let filler = w + 1` between
  the two calls → CORRECT (37 in the id64/id32 variant). Classic register-pressure /
  spill-reload dependency ⇒ the bug is in xmm caller-saved handling around CVT + CALL.

## Where to look
- `x86_regalloc.ax` (MACH_CVTSS2SD/CVTSD2SS handling, ~L87) + the float-arg / call-clobber
  path. Likely the cvt destination xmm (or the pre-cvt source) is treated as call-preserved
  when it must be caller-saved across the intervening CALL, or the two calls' arg xmm regs
  alias without a spill.
- x86_selector OP_CAST float-pair lowering (~L1050) emits the cvt into `inst.dest`; check
  whether that vreg gets a stable home across a following call in regalloc.

## Priority / scope
Niche (needs two back-to-back width-converting float calls; float literals default to f64
so most code never hits it), but a genuine silent miscompile. Backend regalloc → **B==C
fixpoint required** before commit. Deferred from the HOLE#6 session (kept the t_f32argcoerce
oracle to a SINGLE width-cvt call to avoid this bug). Do in a dedicated backend session with
a MachInst/asm dump of the failing two-call sequence. Repro pattern above; also
`fn id64(f64)->f64` / `fn id32(f32)->f32` identity variant returns 1 (want 37).
