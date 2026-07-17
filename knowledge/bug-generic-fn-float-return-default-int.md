---
name: bug-generic-fn-float-return-default-int
description: "FIXED: a generic fn whose return type is a bare type param T, called without a result annotation, mis-typed a FLOAT return as int (caller read the generic TEMPLATE signature, not the concrete instance). Fix: prefer the monomorphized instance's signature (callee.payload, flag 2048) over infer_node's template re-resolution."
metadata:
  node_type: memory
  type: project
---

# Generic fn returning bare `T` → result type defaulted to INT (float returns miscompiled) — ✅ FIXED

## ✅ FIXED 2026-07-18 (A==B `6F1A1ED5`, regression 368/368, oracle t_genfloatret)
**Real root cause (trace-nailed, NOT the earlier NODE_TYPE_EXPR guesses):** the monomorphized
instance's SIGNATURE is CORRECT — `pre_infer_func_signature` computes ret f64(10) for the f64
instance, i32(3) for the int instance, and stores it on `symbols[callee.payload].type_id`. But at
the call site (typecheck.ax ~L3627) `callee_type = infer_node(callee)` **re-resolves the ident to
the generic TEMPLATE** (func type with generic/`ret=0` return, id 63 in the trace) instead of the
instance (id 457, ret=10/3). So `result_type = fi.ret` became 0 → defaulted to int downstream →
correct for int returns (default matched), a spurious `itof` corrupting f64/other returns.
Decisive trace: `XCALL callee=_AX_std_id__f64__o1T ctype=63 ctRET=0 ptid=457 ptRET=10` (float) vs
`ptRET=3` (int) — the payload (instance) type had the right return; infer_node's `ctype` did not.

**Fix:** at L3627, when the call was monomorphized (`callee.flags & 2048`, set right after the
instance is bound), override `callee_type` with the instance's own signature
(`symbols[callee.payload].type_id`) when it is a FUNC type. Frontend-only, A==B held, self-build OK,
no regressions (int generics unaffected — they only worked by the int-default coincidence). Oracle
`bin/t_genfloatret.ax` (maxof + id on f64 → 42). This also fixes generic STRUCT/other-repr returns
that would have hit the same default.

## (historical below — symptom, isolation, and the WRONG earlier hypotheses/dead-ends)

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

## ❌ DEAD ENDS (2026-07-18, attempt 2 — do NOT repeat these)
Two fix attempts on the "NODE_TYPE_EXPR substituted resolution" hypothesis below BOTH failed:
1. Guarding the first_child short-circuit (`if child != 0 and not is_substituted`) at typecheck
   L4138: changed the float result (0/96 → 127/96/224) but did NOT fix it; int stayed correct.
2. A trace `if is_substituted:` in the NODE_TYPE_EXPR branch (L4124) printed **NOTHING** when
   compiling y2 — i.e. **the substituted-NODE_TYPE_EXPR branch is NEVER HIT for the instance's
   return type.** So the return-type resolution does NOT go through that branch at all; the L4140
   narrowing below is WRONG.
CONCLUSION: the instance's SIGNATURE return type (which the caller reads as i64, per the L3629
trace `ret=4` for both int and float) is set on a path I have NOT located. Next attempt MUST first
find it methodically, not guess: (a) trace `pre_infer_func_signature` (typecheck L2370) — print the
KIND of the return annotation child at L2416 and the `ret_type` it computes for the mono instance
(gate the print to the mangled instance name to cut stdlib noise); (b) if the instance takes the
EXISTING-instance branch (typecheck L3579) its type_id was set on an earlier call — check whether
mono creates the instance signature with a concrete or generic/defaulted return; (c) also check
`mono.ax` `substitute_type_params` actually visits the return annotation node (it may skip the
`-> T` node, leaving ret generic → defaulted to int downstream). Confirm the ACTUAL kind/flags of
the return node in the cloned instance before touching any resolver branch.

## ⚠️ (SUPERSEDED, kept for history) Root cause NARROWED to one branch (trace-confirmed 2026-07-18)
Added a temp `XTRACE` at typecheck.ax L3629 and diffed y2 (float) vs y4 (int) traces: the user
`id` call resolves **`ret=4` (i64) for BOTH** — so it is the monomorphized INSTANCE's SIGNATURE
return type that defaults to i64 (not the caller, not the arg). Chain, confirmed piece by piece:
1. `mono.ax substitute_type_params` DOES correctly map the return node: type_id 10 → `intern("f64")`,
   sets `payload=intern("f64")` + `FLAG_IS_SUBSTITUTED` (the `elif type_id==10: "f64"` case exists).
   So the substitution side is FINE.
2. `typecheck.ax pre_infer_func_signature` L2416 computes `ret_type = infer_node(return_node)`.
   The return node is a **NODE_TYPE_EXPR** (not NODE_IDENT), handled at L4124–4210.
3. The NODE_IDENT substituted branch (L4079–4114) maps `"f64"→10` correctly, but the **NODE_TYPE_EXPR**
   branch resolves differently: **L4140–4144 first infers `first_child` and, if non-UNKNOWN, sets
   `result_type = inner` BEFORE the `"f64"→10` text mapping at L4146+ can run.** Prime suspect: the
   substituted return TYPE_EXPR still has a child (the old `T` ident) that infers to a generic/int,
   short-circuiting the correct name resolution — for i64 the wrong path still lands on 4, so int is
   invisibly "correct"; for f64 it lands on 4 (i64) = the bug.
NEXT (turnkey): trace `pre_infer_func_signature` for the instance (gate the print to the mangled
name) to confirm whether L4142 `inner` or the L4146 text path wins for the substituted `-> f64` node;
then make the NODE_TYPE_EXPR substituted-primitive path resolve the name FIRST (mirror NODE_IDENT
L4079–4114) or skip the first_child short-circuit when `is_substituted`. Frontend → A==B. Verify
no regression on generic INT returns (they currently work by coincidence of the same default).

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
