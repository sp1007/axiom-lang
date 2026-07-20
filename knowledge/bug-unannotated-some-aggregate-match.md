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

## ⚠️ RESIDUAL follow-up (separate, pre-existing, NOT a regression) — OPEN
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
