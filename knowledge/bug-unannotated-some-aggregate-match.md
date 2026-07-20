---
name: bug-unannotated-some-aggregate-match
description: "FIXED 56cab6e + a5c410f (A==B==C 93D1B4B1, 465/465). Unannotated `let x = Some(<aggregate>)`/user-generic-sum `Wrap(agg)` + match read only field 0 — GENERIC_INST scrutinee bound the pattern var to the TEMPLATE generic param T. Fix: typecheck subst + air_builder deref-type on BOTH tagged (Option/Result) and linear (user sum) paths. Residual width-mismatch follow-up open."
metadata:
  type: project
---

# ✅ FIXED `56cab6e` — unannotated `Some/Ok(aggregate)` match-payload read only field 0

**Symptom (silent miscompile, found via autopilot feature-combo probing 2026-07-20):**
`let x = Some((3, 4))` (no annotation) then `match x: Some(p): ... p.0 ... p.1 ...` read
**every field as field 0**. Shape varied by payload size:
- 16B payload (tuple `(i64,i64)` / struct 2×i64): `p.1 == p.0` (e.g. `p.0+p.1*10` = 33 not 43).
- ≤8B multi-field payload (tuple `(i32,i32)` / struct 2×i32): `p.1 == 0` (33 → 3 mid-fix).

`.unwrap()` and **annotated** lets (`let x: Option[(i64,i64)] = …`) were CORRECT — they carry
the concrete payload type. Result **from a fn return** was correct (return type = the hint).
Only the unannotated-let + match combination broke. Scalar payloads (`Some(42)`) were fine
(BUG#91 `57b2218`) — no field access to mis-offset.

**Root cause:** an unannotated `Some(agg)` types as a `TYPE_KIND_GENERIC_INST`
(`Option[(i64,i64)]`), NOT a monomorphized SUM. In `typecheck.ax` match-arm typing
(~L3084 GENERIC_INST branch), `find_variant_info`/`sum_info` resolve to the **TEMPLATE**
`Option` sum by base name, whose `Some` variant payload is the generic param **`T`**
(`TYPE_KIND_GENERIC`, kind 7). The pattern var `p` was thus typed as a single-slot generic →
`.N`/`.field` all default to offset 0. air_builder's tagged-match payload deref
(air_builder.ax ~L3242) has the SAME issue: it deref'd with `pl_type = T` (generic scalar 8B),
so a ≤8B aggregate loaded into a register where field1 is unaddressable (→0), while a 16B
aggregate happened to work by-address once the bound-var offsets were fixed.

**Fix (two halves):**
1. `typecheck.ax` new helper **`subst_variant_payload_type(scrutinee_type, payload_type)`**:
   when the scrutinee is a GENERIC_INST and the template variant payload is a bare generic
   param, substitute the concrete type-arg the generic-inst carries — matched by the param's
   NAME against the alias's ordered generic params (same mechanism try_instantiate_variant_call
   uses to BIND them). Wired in right after `payload_type = v_info.payload_type`. Inert on the
   concrete-SUM path (payload already concrete → helper returns input unchanged).
2. `air_builder.ax` tagged-match deref (`lower_match_tagged`): when `find_variant_info` yields
   0 OR a bare generic param (kind 7), prefer the bound var's now-concrete `type_id` for the
   deref, so a ≤8B aggregate payload is loaded by-address (fields addressable), not into a reg.
3. `air_builder.ax` LINEAR-match single-binding OP_GET_FIELD (`lower_match`, `a5c410f`): the
   SAME fix — a USER generic sum (`type Box[T] = Wrap(T) | Empty`) matched unannotated goes
   through the linear path, not the Option/Result tagged path, and had the identical ≤8B gap
   (`Wrap((3,4))` → 3). 16B user-sum payloads already worked via the typecheck half.

**Gate (backend change → B==C):** A==B==C **`05E08D93`**, regression **465/465**, oracle
**t_optmatchagg (57)** O0+O1 (covers 16B/8B tuple + 16B/8B struct + Result Ok + 3-field tuple,
each needing field1 ≠ field0). Daily driver `axc_native` = `05E08D93`.

## ⚠️ RESIDUAL follow-up — struct-field context FIXED `6132b15`; Vec-push + param contexts still OPEN
**Struct-field context CLOSED (`6132b15`, A==B `8C63A414`, 467/467, oracle t_ctorfieldopt=43):**
`Box(v: Some((3,4)))` with field `v: Option[(i64,i64)]` now coerces the tuple to the field's
element widths. Fix = extend the struct-ctor arg coercion (typecheck.ax ~L4446) to thread the
declared field type as expected for OPTION/RESULT/GENERIC_INST/SUM fields (it already did so
for ARRAY/F32/__tup), so try_instantiate_variant_call propagates the concrete payload to the
literal. The two contexts below remain, each harder for a distinct reason:
- **Vec.push(Some((a,b)))** into `Vec[Option[(i64,i64)]]` (probe_gen3=20) — the expected
  ELEMENT type must thread through the GENERIC METHOD ARG (`ax_push`) to the inline ctor. This
  is the mono/generic-call-arg coercion path that was ATTEMPTED+REVERTED before (arg re-infer
  runs after/independently of mono instantiation). Still needs the deep mono-flow session.
- **passing `let x = Some((3,4))` to a param `o: Option[(i64,i64)]`** (pp=3) — DIFFERENT: `x` is
  a VARIABLE already built as Option[{i32,i32}] (8B). Threading the param's expected type to the
  arg does NOT rebuild an ident's value, so coercion cannot help. Correct fix = REJECT the
  Option[(i32,i32)]→Option[(i64,i64)] width mismatch (BUG#53 convention), a conservative
  diagnostic rather than a silent miscompile. A separate, careful reject (watch self-host).

## ⚠️ RESIDUAL follow-up (separate, pre-existing, NOT a regression) — OPEN (original note)
An unannotated `Some((int, int))` builds an `{i32,i32}` **8-byte** payload (tuple int-literal
elements default to i32). When that VALUE later crosses into a **16-byte-expecting** context it
is not width-coerced and field1 reads past-end (0):
- `Vec[Option[(i64,i64)]]` push `Some((20,22))` then read `v[0]` and match → field1 = 0
  (probe `scratch/probe_gen3.ax` returns 20 not 42).
- passing unannotated `let x = Some((3,4))` to a param `o: Option[(i64,i64)]` → returns 3 not 43
  (`scratch/narrow/pp.ax`).
- `Box(v: Some((3,4)))` where field `v: Option[(i64,i64)]`, then `match b.v` → 3 not 43
  (`scratch/p2/q3.ax`). Expected type IS threaded to the field-init inference
  (typecheck.ax:4937 passes `einfo.fields.data[j].type_id`), and the coerced TYPE reaches the
  match — but the tuple VALUE is still built `{i32,i32}` (air-build tuple-literal element
  coercion does not reach the nested `Some(tuple)`). Confirmed 3 contexts (Vec element / fn
  param / struct field) all fail identically; `as i64` elements or a width-matching field type
  fix all three (q3b/q3c). NOTE the confirmed-clean nearby combos (do NOT re-probe): Result
  `Err((i64,i64))` match, nested `Some(Some(tuple))`, `[Some(tuple),None]` array element,
  aggregate bound + mutated in a match arm — all correct post-56cab6e.
Same class as the known "tuple ARG width ≠ tuple PARAM width" reject/coerce gap
([[backlog-open-items]] RFC 0022 note) but routed THROUGH `Some()`, so neither the coercion
nor the BUG#53 reject fires. Fix direction: coerce the ctor's aggregate-literal element widths
to the expected container/param element type (thread the expected type through the Vec-push /
call-arg mono path), or reject on width mismatch. Deferred — needs the mono/coercion flow, not a
match tweak. Related: [[bug-vec-generic-tuple-element-mono-open]], [[bug-tuple-generic-payload-unwrap-open]].

## Clean probe coverage 2026-07-20 (do NOT re-probe these axes — all correct O0+O1 on driver 93D1B4B1)
Three probe batches (13 programs) beyond the aggregate-match fix were all correct:
- **integer/cast/if-expr**: generic `max` chained, u8 wraparound-in-loop (300→44), signed
  div/mod negatives, `if..elif..else` value-expr, variable-shift bit-OR loop.
- **Vec HOF chains**: map→fold, filter→fold, all-predicate, find→Option-match, map.filter.fold.
- **HOF × aggregates / strings**: `map(|x|->(i64,i64))` producing tuple elements (banked as
  oracle **t_hoftup**=44), `Vec[(i64,i64)].filter` by tuple field, string slice `s[0..5]` +
  byte-index loop, `map(|x|->Option[i64])` + Some-match.
Corroborates the mature plateau: the ONE real find this session was the generic-inst
match-payload bug (fixed both tagged+linear paths); everything else in aggregate/HOF/generic/
integer land is sound. Remaining known gap = the width-coerce residual above.

## 🔬 Vec-push context — DEEP DIAGNOSIS 2026-07-20 (dedicated session, ATTEMPTED→REVERTED, root pinpointed)
Attempted the Vec-push coercion (`Vec[Option[(i64,i64)]].push(Some((20,22)))` → gen3=20). Traced
it to the exact blocker; the fix touches self-host-critical mono code so it was REVERTED (change
was inert: A==B `DE4B6230`, but gen3 stayed 20 — ineffective, not shipped). Findings:
- **Fix lever CONFIRMED**: `push(Some((20 as i64, 22 as i64)))` (explicit i64) → 42 ✓ (`g_i64`).
  So making the inner tuple {i64,i64} fully fixes it — it IS a tuple-element-width coercion.
- **Where to coerce**: the existing phase-2 re-infer loop (typecheck.ax ~L3905-3959, inside the
  `is_generic_call` branch) already re-infers `NODE_TUPLE_EXPR`→`__tup` and `f32` args after
  `inferred[]` is resolved. Added an `elif arg_n.kind==NODE_CALL_EXPR and et is Option/Result/
  GENERIC_INST/SUM: self.infer_node(ap, et)` to re-infer the variant-ctor arg with the resolved
  element type. It FIRES (traced) but does NOT coerce.
- **ROOT BLOCKER (traced)**: `inferred[T]` for the push = the receiver's element type = a
  MONOMORPHIZED SUM (kind 6, e.g. type 463), NOT a GENERIC_INST. `get_generic_args(463)` returns
  **EMPTY** (`ega_len=0`), so `try_instantiate_variant_call`'s hint extraction (`exp_args =
  get_generic_args(expected)`) gets nothing → the inner tuple keeps its {i32,i32} default. The
  mono element SUM lacks `generic_args` because `set_sum_generic_args` is only called for
  `fresh_type_alias_inst` (typecheck.ax:2850); the Vec-element sum is monomorphized via a
  DIFFERENT path that skips it.
- **Two fix options for the focused session** (both touch shared/mono code — gate hard, revert-on-red,
  the prior 2026-07-13 attempt caused O0/O1 DIVERGENCE, so O0==O1 cross-check is mandatory):
  (A) populate `generic_args` on the Vec-element mono sum at its creation path (find where the
  Vec-arg Option[(i64,i64)] sum 463 is registered and add `set_sum_generic_args`), so the existing
  coercion works; OR (B) a fallback in `try_instantiate_variant_call`: when `expected` is a concrete
  SUM with empty `generic_args`, extract the matching variant's `payload_type` from the sum's
  variants directly as the `exp_hint`. (A) is more principled (helps any get_generic_args caller);
  (B) is more localized. Watch: on re-visit the Some call may already be flag-2048 resolved and skip
  try_instantiate_variant_call — verify the coercion path still runs (option-B may need the first
  visit to carry the hint, i.e. derive T from the receiver BEFORE the arg_types loop at L3823).
- **NOT conflated**: `g_annot` (push an ANNOTATED Option variable `let e: Option[(i64,i64)]=…; v.push(e)`)
  → 20 too — a SEPARATE issue (pushing a 16B-payload Option *variable*, distinct from the inline-ctor
  width path). Investigate independently.

### Option A ATTEMPTED 2026-07-20 — INSUFFICIENT (reverted)
Tried option A: in `finish_generic_instantiation` (typecheck.ax ~L2849), set the result SUM's
`generic_args` whenever it's a concrete SUM with empty generic_args (not only `fresh_type_alias_inst`),
idempotent + empty-guarded. Re-added the variant-ctor re-infer branch to consume it. **A==B `3AEF2FB2`
(inert), but gen3 STILL 20.** Trace confirmed `get_generic_args(et=463)` remained `ega_len=0` — so the
Vec-element mono SUM (463) is NOT created via `finish_generic_instantiation`; it is instantiated by a
DIFFERENT path (likely during Vec-annotation type resolution / the initial `pre_infer_type_alias` pass
at L2220, which calls `register_sum_type` at L2502 and never sets generic_args). Reverted (revert-on-red).
**Remaining approaches for the focused session (both real changes to shared generic code, mandatory
O0==O1 gate):**
  (A') find the ACTUAL creation path of the Vec-element sum (trace `register_sum_type` callers for a
       concrete Option/Result arg) and call `set_sum_generic_args` there; OR
  (B)  in `try_instantiate_variant_call` (typecheck.ax ~L439), when `get_generic_args(expected)` is empty
       AND `expected` is a concrete SUM, derive `exp_args` from the sum's VARIANT payload_types (for
       Option: `exp_args[0]` = the Some variant's payload_type = the concrete tuple). Needs the
       variant→param ordinal mapping (Result has T,E). This is creation-path-independent, so likely the
       more robust route. NOTE the re-infer branch DID run (rr changed 463→468), so on re-visit the Some
       call is NOT flag-2048-skipped — the only missing piece is the exp_hint, so (B) should suffice.
The struct-field context (`6132b15`) stays fixed; only Vec-push + param remain.

### ✅ Vec-push context FIXED `f301ae9` (option B) — the flagged-hard case is SOLVED
Option B worked (after option A / plain re-infer were reverted). Two parts, frontend-only
(A==B `61D345BD`, 468/468, oracle t_vecoptpush=42):
1. The existing generic-method-arg phase-2 re-infer loop (typecheck.ax, `is_generic_call` branch,
   already handles `__tup`/`f32` args) now also re-infers a variant-CTOR arg (`NODE_CALL_EXPR`)
   with the receiver-resolved element type once `inferred[]` is final.
2. `try_instantiate_variant_call` (typecheck.ax ~L439): when `get_generic_args(expected)` is empty
   AND `expected` is a concrete monomorphized SUM, recover the coercion hint DIRECTLY from the
   matching variant's concrete `payload_type` (single-payload variants only; flagged RFC 0019
   multi-field synth structs excluded; `__tup` coercion self-guards → inert otherwise).
This is the creation-path-independent route — it sidesteps the elusive mono-sum-args gap entirely.
Fixes `Vec[Option[(i64,i64)]].push(Some((20,22)))` (gen3: 20→42), Option multi-push, and Result
**Err**-payload tuple. **NARROW residual still open**: Result with the tuple in the FIRST type
param (`Vec[Result[(i64,i64),i64]].push(Ok((20,22)))`, vp_res) stays 20 — for the Ok payload param
(gi=0) `exp_args[0]` comes back non-UNKNOWN-but-wrong (bypassing the fallback), whereas Err (gi=1)
and Option both work. Likely a get_generic_args ORDERING mismatch for a 2-param Result mono; a
small follow-up (trace exp_args order vs gp_name_ids for Result). The **param-passing context**
(pp: pass an already-built Option[{i32,i32}] value to a 16B param) remains a REJECT follow-up.
