---
name: bug-generic-fn-float-return-default-int
description: "OPEN silent-miscompile: a generic fn whose return type is a bare type param T, called WITHOUT an expected-type hint, resolves its result type to generic-T which defaults to INT — correct for int instances, garbage for float (and likely struct) returns. Workaround: annotate the result."
metadata:
  node_type: memory
  type: project
---

# Generic fn returning bare `T` → result type defaults to INT (float returns miscompile) — OPEN

## Symptom (probe batch 6, 2026-07-18)
```
fn id[T](a: T) -> T:
    return a
fn main() -> i64:
    let r = id(3.5)          // r SHOULD be f64 3.5
    return (r * 12.0) as i64 // want 42; GOT 0 (-O0) / 96 (-O1) — divergent garbage
```
Deterministic, O0≠O1. Repro `/tmp/pb6/y2.ax` (or x2). Common pattern: generic `id`/`max`/`min`/
`clamp` on `f64`.

## Isolation matrix (all `return (r*12.0) as i64`, want 42)
- x1 non-generic `f64` max → 42 ✓ (float ABI fine non-generically)
- x4 / y4 generic **int** `id(42)` → 42 ✓ (generic works for int)
- x3 generic float COMPARE returning bool → 42 ✓
- **y2/x2/z1 generic float return (no annotation) → 0/96 ✗**
- **z2 `let r: f64 = id(3.5)` (result ANNOTATED) → 42 ✓  ← the tell**
- z3 int-then-float instantiation → 0/96 ✗ (int instance doesn't poison)

## Root cause (AIR-confirmed)
`dump-air` of y2:
```
fn @4 (main):
  %2 = call                 ; id(3.5) result — UNTYPED
  %5: t10 = itof %3         ; ← treats the f64 return as an INTEGER, spurious int→float
  %6: t10 = fmul %5, %4
fn @51(t10) -> t10:  %1: t10 = copy %1; ret %1   ; instance is CORRECT f64->f64
```
The INSTANCE is correctly monomorphized to f64. The CALLER mis-types `id(3.5)` as int, so
`r * 12.0` inserts a bogus `itof`, reading the float return (XMM/return slot) as an integer bit
pattern → garbage. Because z2 (explicit `f64` expected) works, the generic-call result-type
resolution only lands on f64 when an expected type flows in; with `expected == UNKNOWN` the return
type stays the bare generic param `T` and is then DEFAULTED to an integer type (i64/i32) — right for
int instances, wrong for float.

## Fix location (typecheck.ax generic-call block)
`result_type = fi.ret` at ~L3629, from `callee_type = infer_node(callee, TYPE_UNKNOWN)` (L3624),
where `callee.payload` was set to the monomorphized instance (L3614). The instance's SIGNATURE type
(`symbols[inst_sym_idx].type_id`'s `fi.ret`) is evidently still the generic param (not concretized
to f64) even though the instance BODY is f64 — so either:
- (a) concretize the instance signature return in `pre_infer_func_signature(cloned_root)` (L3611)
  so `fi.ret` reads f64 directly (preferred — fixes it at the source), OR
- (b) at L3629, if `is_generic(result_type)`, substitute it via `inferred[gk]` for the matching
  `gen_params` entry — BUT `inferred`/`gen_params` are FREED at L3617-3622 before L3629, so this
  needs the free moved after the substitution.
Confirm with a trace of `symbols[inst_sym_idx].type_id`'s ret vs `inferred[T]` (needs one
rebuild). FRONTEND fix → A==B gate. Add an oracle once fixed (e.g. `t_genfloatret` id/max on f64
→ 42); a WORKAROUND oracle (annotated `let r: f64 =`) already passes today.

## Distinct from t_f32generic
`t_f32generic` covers an f32 LITERAL ARG re-typing in `Vec[f32].push` (arg side). This is the
RETURN side of a generic fn — uncovered. Not a duplicate.

## Workaround (works today)
Annotate the call result's type: `let r: f64 = id(3.5)`. Or use a non-generic float fn.

## LESSON
Generic type-param inference resolving a return to "generic default = int" is invisible for int
instantiations (the default happens to be right) and only surfaces on float/other-repr instances —
exactly why it survived to now. Probe generic functions with NON-integer type params (float,
struct, str) to flush this class.
