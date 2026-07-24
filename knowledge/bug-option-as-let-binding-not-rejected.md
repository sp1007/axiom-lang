---
name: bug-option-as-let-binding-not-rejected
description: "FIXED 2026-07-24g — probe-found silent miscompile: `let x: T = <Option/Result/sum value>` with a plain-typed annotation and NO .unwrap() was accepted-then-miscompiled (the 8-byte tagged box stored raw into the scalar → garbage: `let x:i64 = o` returned 8; `let x:i64 = v.get(0)` returned garbage). Now REJECTED. Let-binding sibling of the call-arg reject (bug-option-as-call-arg-not-rejected)."
metadata:
  node_type: memory
  type: project
---

## ✅ FIXED 2026-07-24g (autopilot probe) — driver `C53A2DD1`, 540/540, A==B
**Probe-found** during a feature-cross batch (nested `Vec[Vec[i64]]`). The nested-Vec probe
mis-typed `let inner: Vec[i64] = outer.get(0)` — `Vec.get` returns `Option[T]`, so the RHS was
`Option[Vec[i64]]` — and the compiler **silently accepted** the Option→plain binding and
miscompiled it. Minimized to the scalar case:
```
let o: Option[i64] = Some(55)
let x: i64 = o          // returned 8 (the box read raw), should REJECT
```
and the call-result case `let x: i64 = v.get(0)` (returned garbage). This is the **let-binding
sibling** of [[bug-option-as-call-arg-not-rejected]] (`cf42579b`, call-args) — same root class (an
Option/Result value used where its inner `T` is expected; AXIOM has NO auto-unwrap), a DIFFERENT
site the call-arg fix never touched.

## Root & fix
`typecheck.ax` NODE_VAR_DECL/NODE_CONST_DECL handler (~L3454): the let-binding site already had the
str-vs-numeric-literal reject (BUG#53 m5) but **no Option/Result/sum→scalar reject**. Added a mirror
of the call-arg reject right after it (~L3498): when the ANNOTATED `exp_type` kind is
`PRIMITIVE or STRUCT` and the RHS's TRUE type kind is `OPTION/RESULT/SUM`, reject
"cannot bind an Option/Result/sum value to a `let`/`mut` of a plain type; unwrap it first".
- RHS TRUE type: for an IDENT init read the **symbol's** `type_id` (an annotated `Option[i64]` local
  resolves to a **SUM-kind** type here — the same trap as the call-arg fix; an OPTION-kind check on
  the already-computed `inferred` alone would miss it).
- For a non-ident init (e.g. a call `v.get(0)`) **REUSE the already-computed `inferred`** (L3477).
  ⚠️ First attempt RE-INFERRED with `infer_node(init_expr, TYPE_UNKNOWN)` and **regressed t_u64cmp
  7→1**: infer_node has SIDE EFFECTS — for `let big: u64 = 18000000000000000000` the second call
  re-defaulted the int literal to i64 and overwrote node_types, flipping `big > small` to a SIGNED
  compare. `inferred` (derived once with the real `exp_type` hint) is side-effect-free AND still
  carries the Option/Result type for a call result — the hint cannot coerce `Option→scalar` (that IS
  the miscompile), so `v.get(0)` still infers as `Option[i64]` there and L1 still rejects. Confirmed:
  reusing `inferred` catches the call-result case (L1) with no re-inference.
- Gate `PRIMITIVE or STRUCT` annotated type: an `Option`-annotated local has an OPTION-kind exp_type
  and is skipped, so a correct `let o: Option[i64] = Some(5)` never trips. Cannot over-reject —
  binding an aggregate box where a scalar/struct is expected is always a type error.

## Verification
Fast fixpoint **A==B `C53A2DD1`** (compiler self-builds byte-identical; its own source never binds an
un-unwrapped Option to a plain local → INERT frontend reject). Regression **540/540** (+`t_optletreject`
reject, +`t_optletok`=37 positive companion). Oracles: `bin/t_optletreject.ax` (Option→i64 let, reject),
`bin/t_optletok.ax` (Option-annotated local + unwrapped scalar bind + nested `Vec[Vec[i64]]`
retrieve+unwrap → 37, guards against over-reject). Positives verified: `Some(5)`+unwrap=42, `.get(0).unwrap()`=77,
nested `Vec[Vec[i64]]` written correctly = 20.

## Lessons
- ⭐ **Nested `Vec[Vec[i64]]` WORKS** when written correctly (`outer.get(0).unwrap().get(1).unwrap()`);
  the "nested Vec segfault" the probe first showed was ONLY the missing-unwrap miscompile, not a
  container bug. `Vec.get` returns `Option[T]` — a probe that omits `.unwrap()` is ill-formed.
- Same family as the call-arg and field-access ([[bug-option-as-call-arg-not-rejected]],
  `t_forgotunwrap`) rejects: an Option/Result value used where its inner `T` is expected must be a
  diagnostic, not a silent box-reinterpret. Remaining un-covered siblings of this class (if any):
  assignment `x = o` to a plain-typed mut (untested here), and SUM param/binding receiving a
  DIFFERENT sum (needs a type_id compare) — niche, safe, not yet probed.

Related: [[bug-option-as-call-arg-not-rejected]], [[bug-option-arith-miscompile-open]].
