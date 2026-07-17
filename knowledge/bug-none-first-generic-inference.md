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

## ✅ UPDATE 2026-07-18 (later) — un-inferable-param sub-case now REJECTS cleanly
`s3` and `n2` (a type param NO argument can determine, in a CONCRETE calling context) now emit
`error: cannot infer generic type parameter for this call in a non-generic context; add explicit
type arguments (f[T](...)) or annotate the argument` and halt (BUG#53) instead of segfaulting.
Implemented exactly per the pointer below: added `current_fn_is_generic` to `TypeChecker`
(set from `FLAG_IS_GENERIC` at NODE_FUNC_DECL), and at the `has_generic_arg` defer site reject
when `not self.current_fn_is_generic`. Generic BODIES still defer (self-host untouched: self-build
OK, A==B `32F4DF0E`, regression 367/367). Oracle `bin/t_uninferreject.ax` (reject mode). The defer
site is correct-by-construction: the compiler's own concrete code never leaves a param generic, so
the reject never fires on self-build. **vG (below) is NOT covered — it is inferable (T=i64) so it
is not rejected. See the correction below — vG turned out NOT to be a bug.**

## ✅✅ CLUSTER FULLY RESOLVED 2026-07-18 — all three sub-cases closed
- **Realistic Some+None inference** (vD/p7) — FIXED `4aa4868` (concrete overrides generic binding).
- **Un-inferable param in concrete context** (s3: `combine[A,B](None, Some(100))`; n2:
  `let r: Option[i64] = id(None)`) — both genuine segfaults (139), now REJECT cleanly `2438ba0`
  (`current_fn_is_generic` discriminator). n2's T-from-expected-return remains un-inferable at the
  CALL, so reject is the correct answer (user annotates / gives explicit type args).
- **`vG`** (`pick(None, Some(42))` returning the None arg) — **NOT A BUG.** It always returned the
  correct value; I misread a **truncated exit code**: the None arm returned `999`, and `999 & 0xFF
  = 231`, which is what bash showed. Re-checked with a sub-256 value (None arm `200` → exit `200`)
  and with the identical `w2` program (None arm `99` → `99`). The None VALUE round-trip through a
  generic call is CORRECT; there is no None-value-lowering bug.

⚠️ **LESSON (probe hygiene):** bash exit codes are **8-bit** (`value & 0xFF`). An oracle value ≥256
(like `999`) will alias (`999→231`) and look like a miscompile. ALWAYS keep probe oracle values in
`[0,255]`, or read the full value via PowerShell `$LASTEXITCODE`. I banked a phantom "None-value-
lowering residual" for ~an hour on this misread before w2/w3 (same shape, sub-256 arm) exposed it.

### Implementation pointer (worked out 2026-07-18, NOT yet built — turnkey for the session)
The defer decision is `if has_generic_arg:` at typecheck.ax ~3488 — it FREEs and returns (defers,
leaving the call's result_type generic). For a call inside a CONCRETE function this defer is wrong
(no later mono pass instantiates it → segfault). The missing discriminator = **is the ENCLOSING
function generic?** Not currently tracked. Add it:
1. New field `current_fn_is_generic: bool` on `struct TypeChecker` (~L84, beside `current_return`);
   init false in the ctor (~L106).
2. At `NODE_FUNC_DECL` (L2590–2613): save/restore like `current_return`; set
   `self.current_fn_is_generic = (self.tree.nodes.data[node_idx].flags & FLAG_IS_GENERIC) != 0`.
3. At the defer point: `if has_generic_arg and not self.current_fn_is_generic:` → the param is
   genuinely un-inferable in a concrete context. Choose per sub-case:
   - **s3** (leftover generic param is UNUSED for values, only matched as None): default it to a
     concrete type (mirror the existing `has_unresolved -> TYPE_I32` at L3453) so it monomorphizes
     correctly. LOW risk for the unused case.
   - **n2** (param drives the RETURN type via the caller's expected type): needs expected-return
     (`self.current_return` / the `let` annotation) flowed INTO inference before defaulting, else the
     i32 default mismatches the Option[i64] slot. Do NOT blind-default this one.
   - Fallback when neither resolves: clean REJECT "cannot infer type parameter <name>" (BUG#53), not
     a segfault.
GATE HAZARD: the compiler's OWN concrete functions must not legitimately rely on `has_generic_arg`
deferral — verify via A==B fixpoint + full regression (a wrong discriminator either breaks self-host
build or wrongly rejects). That uncertainty is exactly why this is a dedicated session, not an
autopilot-tick change. Repro programs in `/tmp/probe/{vG,n2}.ax` + `/tmp/pb4/s3.ax` (regenerate
from this note). NOTE: the realistic `Some`+`None` case (a concrete sibling determines T) is FIXED;
all residuals are the "no concrete arg determines the param" family.

## LESSON
When shrinking a probe, the isolation matrix (None-first vs None-second, annotated vs bare,
explicit vs inferred type args) pinned the exact trigger fast. A concrete-vs-generic override in
type-arg inference is the general shape — a template-typed arg must never lock a generic param.
