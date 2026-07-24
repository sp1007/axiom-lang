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
  diagnostic, not a silent box-reinterpret.

## ✅ ASSIGNMENT sibling ALSO FIXED (same session) — driver `C3A821C0`, 542/542, A==B
Immediately probed the assignment site (the "untested" sibling this note first flagged) and it had the
**identical silent miscompile**: `x = o` (`o:Option`, `x:i64`) → 8; `x = v.get(0)` → 127; `s.a = o`
(struct field) → 8 — all accept-then-store-box-raw. Fixed with the same mirror at the NODE_ASSIGN_STMT
handler (typecheck.ax ~L6329): gate on the LVALUE kind PRIMITIVE/STRUCT + RHS true-type OPTION/RESULT/
SUM; RHS true type from symbol (IDENT rhs) else **reuse the existing `rhs_inferred`** (the handler
already infers the RHS with the lvalue hint — capture that return value, do NOT re-infer, same
t_u64cmp side-effect trap). Oracles `t_optasgreject` (reject) + `t_optasgok` (=42: unwrapped assign +
Option-lvalue reassign + sum-lvalue reassign, guards over-reject). A==B `C3A821C0`, regression 542/542.
- Still un-covered (niche, safe): a SUM param/lvalue receiving a DIFFERENT sum/Option (kind SUM vs SUM
  needs a type_id compare to avoid rejecting the same-sum case) — left as the call-arg note also left it.

## Soundness re-confirmed 2026-07-24g (post-ship probing)
Verified NO over-reject on valid generic code: generic free-fn calls returning a plain type bound to a
plain local (`let s: str = identity[str]("hi")`, `let n: i64 = identity[i64](42)`, inferred form) all
compile+run; `b.get()` method calls returning str don't reject. Decisive: the **A==B self-build is
byte-identical**, and the 2 MB compiler source is saturated with generic-call-bound-to-local patterns —
an over-reject there would have failed self-compilation. The ONE edge that trips the reject is a
**malformed** call `get[str](b)` passing a `Box[str]` VALUE to a `ptr[Box[T]]` param (unresolved generic
inference yields a bogus SUM-kind → the reject fires with a misleading "Option/Result/sum" message). It
is on ALREADY-invalid code (the idiomatic `b.get()` form doesn't reject), so not a regression of valid
code — at most a diagnostic-quality nit. Encountered while re-confirming the separate >8B generic-method
return bug ([[bug-generic-struct-inline-method]], still reproduces on driver `C3A821C0`: R1 inline
method→str = 127, R3m free-fn method→str = 74 — backend 16B-return-with-receiver codegen, B==C, fresh
session per that note).

Related: [[bug-option-as-call-arg-not-rejected]], [[bug-option-arith-miscompile-open]], [[bug-generic-struct-inline-method]].
