---
name: bug-none-first-generic-inference
description: "FIXED: generic call with a bare `None` arg + a concrete sibling arg mis-monomorphized (None poisoned T via first-wins) → silent segfault. Residual OPEN sub-bug: None-value lowering when T comes only from inference/expected-type."
metadata:
  node_type: memory
  type: project
---

# `None`-first generic type-arg inference — silent segfault (FIXED 2026-07-18, probe-found)

## Symptom (found by bug-probe)
`pick(None, Some(42))` on a generic `fn pick[T](a: Option[T], b: Option[T]) -> Option[T]`
**segfaulted (139)** deterministically at BOTH -O0 and -O1 — accept-then-miscompile.
Minimal repro (`vD`): body just `return b`. Isolation:
- `pick(Some(42), None)` (None NOT first) → OK.
- `pick[i64](None, Some(42))` (explicit type args) → OK.
- `let n: Option[i64] = None; pick(n, Some(42))` (annotated None) → OK.
So the trigger = **bare unannotated `None` as an argument whose param's T must instead be
inferred from a LATER concrete arg.**

## Root cause
A bare `None` (no call args) is typed as the **Option TEMPLATE** (T = the template's own
generic param) — see `try_instantiate_variant_call` (typecheck.ax:366, returns bare
`sum_type_id` when `first_call_arg == 0`). In `infer_generic_type_args` (typecheck.ax ~630)
the arg→param match was **first-wins**: `if inferred[g] == TYPE_UNKNOWN: inferred[g] = arg_type`.
So arg0 `None` bound `T := <generic template param>`; arg1 `Some(42)` (=Option[i64]) then hit
the guard (`inferred[T]` no longer UNKNOWN) and **could not correct it**. `T` stayed generic
→ `has_generic_arg` path baked a wrong/garbage instance → segfault.

## Fix (`FIXED`, A==B fixpoint `9A2788B6`, regression 366/366 incl. `t_noneinfer`)
typecheck.ax `infer_generic_type_args`: let a **concrete** arg override a previously-inferred
**generic** binding (concrete never overridden by a later generic; UNKNOWN still taken by anything):
```
if inferred.data[g_idx] == TYPE_UNKNOWN:
    inferred.data[g_idx] = arg_type
elif self.is_generic(inferred.data[g_idx]) and not self.is_generic(arg_type):
    inferred.data[g_idx] = arg_type
```
Frontend-only (inference → mono selection), A==B held. Oracle `bin/t_noneinfer.ax` (exit 42).
Fixes the realistic `first_some`/`or_else`-shaped pattern (`Some` + `None` args). Family of
[[bug91-generic-sum-ctor-inference-open]] / [[bug79-vec-option-none-mono]] but distinct (arg-order,
concrete-overrides-generic).

## Residual OPEN sub-bug (banked, NOT fixed by the above) — None-value lowering
The inference fix picks the right INSTANCE, but the `None` VALUE at a generic call site is still
lowered from the template repr, so two harder cases remain:
- **`vG`**: `pick(None, Some(42))` returning `a` (the None) → now compiles (T=i64) but returns
  **231** (want 999 via the None arm). Crash → silent-wrong (the None arg round-trip is mis-lowered).
- **`n2`**: `let r: Option[i64] = id(None)` on `fn id[T](a: Option[T]) -> Option[T]` → **segfault**.
  Here T can only come from the expected RETURN type (no concrete arg to override the template),
  which the generic-call inference doesn't flow in.
Owning stage: air_builder lowering of a `None` argument in a generic call (value repr must follow
the monomorphized element type), + expected-return-type flow into generic-call inference. Deeper;
needs a dedicated session. Repro programs in `/tmp/probe/{vG,n2}.ax` (regenerate from this note).

## LESSON
When shrinking a probe, the isolation matrix (None-first vs None-second, annotated vs bare,
explicit vs inferred type args) pinned the exact trigger fast. A concrete-vs-generic override in
type-arg inference is the general shape — a template-typed arg must never lock a generic param.
