---
name: bug-freefn-stdlib-collision-noarg
description: "FIXED cluster (4 holes): user free fn colliding with a bundled stdlib fn mis-dispatched. (1) zero-arg 730ee9e; (2) bare-mangle f7bc186 (both mangle ax_<name> -> link-merge); (3) bare-generic first param 89cd5c8; (4) NUMERIC-WIDENING df269d5 — i32 literal arg vs i64-param user fn skipped -> fell through to generic Vec/HashMap head -> monomorph on garbage -> SIGSEGV; fix widens narrow-int arg0->i64 in resolve_free_call_overload. Latest: A==B 5F5704B8, 331/331."
metadata:
  node_type: memory
  type: project
  originSessionId: 7fa6650d-d306-4f61-9562-d7dcda064554
---

✅ **FIXED** `730ee9e` (frontend-only typecheck.ax, A==B `982630AE`, 141/141) — found by feature-combo probing 2026-07-10 (Bug B; Bug A of that round shipped as `c103b9f`).

## Fix shipped
`resolve_free_call_overload` zero-arg branch: when the call has 0 args and the bound head takes >0 params, walk the overload chain for a 0-param overload (the user's own `f()`) and switch to it. Param count from `sym_decl_param_count` (counts NODE_PARAM_DECL on the decl node in the symbol's OWN tree — tree-safe, independent of signature-inference order). Conservative: only switches on arity MISMATCH so correctly-bound zero-arg calls are unchanged. Oracle `bin/t_freefncollision.ax` (user `run()` shadowing stdlib `run(self)` → 15). Confirmed: `run()` 0→15, `close()` SIGSEGV→7.

✅✅ **RESIDUAL FIXED 2026-07-10 (g)** `f7bc186` (mangling; A!=B, **B==C `353AA12B`**, **143/143**, origin/main=`f7bc186`). Two coordinated changes: (1) `resolve_free_call_overload` now disambiguates by **param-count + arg0-type** (walk chain, prefer arity match; preserves BUG#80 equal-arity-by-arg0) so the call binds to the user's symbol; (2) `run_type_checker` **Phase 3.5** tags every bare-mangling fn sharing its name with an EARLIER bare-mangling fn with `SYM_FLAG_MODDUP` (new `free_fn_bare_mangles` mirrors x86_regs.ax) → backend emits unique `ax_<name>__m<idx>`; the FIRST (stdlib) keeps the bare name so existing callers are unchanged. Zero self-collisions in the compiler → only manifests for user programs. Oracle `t_freefncollision_arity` (`len(str,i32)`→105; the i32 SIGSEGV repro→107). **Whole free-fn/stdlib collision cluster now CLOSED** (zero-arg `730ee9e` + bare-mangle residual `f7bc186`). Diagnosis below kept for reference.

✅✅✅ **THIRD HOLE FIXED 2026-07-16 `bced0d2`** (A==B `A7E36C48`, 324/324) — the arity-disambig from `f7bc186` STILL failed when the user overload's FIRST PARAM is a BARE generic type-param. Repro: bundled `Vec.zip[T,U](self,other)` (2-arg) shadows user `fn zip[A,B,C](a,b,f)` (3-arg); bare call `zip(20,22,f)` SEGFAULTED. Root: `is_method_compatible` (typecheck.ax ~939) computes `kinds_match` only for same-kind OR both-structured types → a first param of `TYPE_KIND_GENERIC` (bare `A`) vs a concrete i64 arg returns FALSE. So `resolve_free_call_overload` (which needs arity-match AND arg0-compat, ~888) skipped the same-arity user `zip` and fell through to the wrong head. Fix (FINAL `89cd5c8`): in `resolve_free_call_overload`'s ARITY-matched branch (~888), accept an overload whose first param is TYPE_KIND_GENERIC (once its param count already == call arity). Diagnosed READ-ONLY by tracing kinds_match. **UNBLOCKED Vec.zip** (shipped, t_veczip=62). ⚠️ FIRST attempt `bced0d2` put the rule GLOBALLY in is_method_compatible (`if kind==GENERIC: return true`) — too broad: is_method_compatible is also used by resolve_method_overload WITHOUT arity gate, so it OVER-MATCHED user function overloading (`fn pick(x:i64)` + `fn pick[T](x,y)`: `pick(30)` picked the 2-arg pick[T]→11 not 31). Caught by POST-CHANGE PROBING (validation discipline). Narrowed to the arity-gated branch (`89cd5c8`, oracle t_useroverload=42). Also re-greened `t_genfnopt`. Residual may remain for a STRUCTURED (non-bare-generic) first-param mismatch (original Vec.first[T](self:Vec[T]) case) — not re-hit.

✅✅✅✅ **FOURTH HOLE FIXED 2026-07-16 `df269d5`** (frontend-only, A==B `5F5704B8`, 331/331) — NUMERIC-WIDENING residual. Repro: `fn find(a: i64, b: i64) -> i64` (or ANY user free fn named after a bundled GENERIC Vec/HashMap/HashSet method: find/get/push/map/filter/contains/index_of/count/reverse) + bare call `find(40, 2)` → **SIGSEGV**. Root: integer literals default to **i32 (type 3)**, but the user overload has **i64 (4)** params; `is_method_compatible` (non-generic path, ~1003) does an EXACT `unwrapped_param==unwrapped_rec` i.e. `4==3` → FALSE. So `resolve_free_call_overload`'s arity-matched branch (~906) skipped the user `find`, fell through to `resolve_method_overload` → returned the bundled **generic** stdlib head (`find[T](self: Vec[T],...)`); air_builder took the generic-call path and MONOMORPHIZED `find[i32]` on the i64 arg treated as a Vec receiver → crash. Diagnosed by a throwaway debug compiler (dbgsym print in x86_coff.ax + dbgov in resolve_free_call_overload): showed exactly ONE `ax_find` (user, unique) BUT a spurious `ax__AX_std_Vec__i32__AX_std_get__i32__o2Vec` instance materialized for a program that never uses Vec → proof the CALL mis-dispatched (not a link collision). Fix (`df269d5`): in the arity-matched branch, widen a narrow-int arg0 (i8/i16/i32/u8/u16/u32 = 1,2,3,5,6,7) to a CONCRETE i64 (4) first param → `ci_arg0_widen`, added to the accept condition. Mirrors air_builder UFCS widening (~1726). Only concrete-i64 widens; generic `Vec[T]` receiver (GENERIC_INST, not 4) untouched → real `v.get(i)`/`v.find(pred)` still dispatch (verified =32/=42). Oracle `bin/t_freefnfind.ax` (user find+get + real Vec.get =42). **This is the residual foreseen at the end of the 3rd-hole note** (STRUCTURED/concrete first-param mismatch) — now closed for the integer-widening case. ⚠️ Any FLOAT arg0 (f32→f64) is deliberately NOT widened (needs cvtss2sd; mirror UFCS exclusion) — untested, potential 5th hole if a user fn shadows a float-first-param generic method.

✅ **HOLE#4 GENERALIZED `ece1cbf`** (frontend, A==B `0B9FD818`, 331/331) — df269d5 only widened an **i64 (type 4)** first param; a user fn with **u64/isize/usize** param (e.g. `fn count(a: u64, b: u64)` called `count(40,2)`) still SIGSEGV'd (probe p_u). Generalized to: integer arg0 compatible with ANY concrete integer first param (`tc_is_numeric(fp0) and fp0 not in {F32,F64}` AND same for arg0). Generic `Vec[T]` receiver isn't a numeric primitive → never over-picked. Oracle t_freefnfind extended with u64 `count`.

✅✅✅✅✅✅ **HOLE#6 FIXED 2026-07-16 (A==B `507EE552`, 336/336)** — FLOAT f32↔f64 arg
widening, the residual flagged at the end of HOLE#4/#5. Repro: `fn map(x: f64) -> i64`
(shadowing the bundled generic `Vec.map`) called with an f32 arg → the exact gate skipped
the user overload (f64 != f32), fell through to the generic Vec.map head → monomorphized on
the float treated as a Vec receiver → **SIGSEGV**. ALSO a plain-value bug independent of the
collision: a unique `fn f(x: f64)` called with an f32 variable read **0** (no cvt emitted).
Two-part fix: (A) `free_call_gate` (typecheck.ax) now gates a concrete-float first param
against a float arg (`ci_arg0_float`), so BOTH the user (concrete float) and the generic head
are gated → the HOLE#5 tie-break binds the user fn. (B) air_builder `coerce_float_arg` emits
a real OP_CAST (→ cvtss2sd/cvtsd2ss) at the call site when an arg's static float width ≠ the
param's, keyed off the resolved callee FUNC symbol's param types (param index == temp_count,
which counts the method receiver). Dormant for self-compile (compiler passes no f32→f64 args)
→ A==B directly. ⚠️ REGRESSION caught in post-change probing: adding the float overload to the
gate created a TIE that head-first resolved WRONG (`pick(f32)`+`pick(f64)` called with an f64
literal picked the head f32). Fixed by GRADING the tie-break score: exact type match = +2,
coercion-compatible = +1 (standard "exact preferred over promotion") — also improved
int-width overloading (`conv(i32)`+`conv(i64)` with an i64 arg now picks i64 exactly). Oracle
`bin/t_f32argcoerce.ax` (collision map(f64)/f32-arg + real Vec.map dispatch = 42). ⚠️ Kept the
oracle to a SINGLE width-cvt call: two BACK-TO-BACK width-converting float calls hit a
separate PRE-EXISTING xmm-regalloc bug (reproduces with explicit `as` casts) — logged as
[[bug-consecutive-float-cvt-call-regalloc]]. **The free-fn/stdlib collision cluster is now
CLOSED across all 6 holes** (zero-arg / bare-mangle / bare-generic-first-param / integer-
widening / same-arg0-tie / float-widening).

✅✅✅✅✅ **HOLE#5 FIXED `018d49c`** (frontend, A==B `489CDDAE`, 332/332) — ambiguous same-arity, SAME-arg0-type collision. Repro (probe p_s): `fn contains(s: str, n: i64)` + `contains("hi", 2)` → SIGSEGV (bound to NON-generic bundled `std.string.contains(str,str)` — both arity 2, arg0=str; call passed i64 `2` where str `sub` expected → read as ptr → crash). arg0-type CANNOT disambiguate (identical). Fix = **two-pass full-signature resolution** in `resolve_free_call_overload`: Pass 1 counts arity+arg0-gated candidates (gate extracted to new `free_call_gate` so both passes accept the SAME set). 1 candidate (or arg0 UNKNOWN) → bind head-first, exactly as before (zero behaviour change for the common case). ≥2 → infer ALL arg types, score each candidate's full param list via new `param_arg_match` (exact | integer↔integer | float↔float; un-inferrable arg = neutral), bind highest; score-tie keeps head-first (best_score starts −1). A bare-generic `T` param scores 0 vs a concrete arg → a concrete overload out-scores a generic one for a concrete call (intended). Only the genuinely-ambiguous path pays the extra inference. Verified: p_s=42, real `str.contains`/`Vec.get`/`Vec.find` still dispatch, user overloading `pick(i64)`+`pick[T](x,y)` `pick(30)`=31 unchanged. Oracle t_freefncontains. **WHOLE free-fn/stdlib collision cluster now CLOSED across all 5 holes** (zero-arg / bare-mangle / bare-generic-first-param / integer-widening / same-arg0-tie). ⚠️ Residual (untested, low-priority): FLOAT f32→f64 arg widening still not done — a user `fn m(x: f64)` shadowing a generic method called with an EXPLICIT f32 arg could mis-dispatch (float literals default to f64 = exact, so normal code is unaffected).

---
**ROOT-CAUSE (g): NOT a resolution bug, it was a MANGLING COLLISION.**
Repro (constructed, not user-reported): `fn len(a: str, b: i32) -> i32: return b+100` then `len("hi",5)` returns **2** (= stdlib `len("hi")`), and `fn len(a: i32, b: i32)` + `len(3,4)` **SIGSEGVs** — both silently mis-dispatch to the bundled stdlib `std.string.len(s: str)`.

Why the earlier zero-arg / BUG#80 fixes WORKED but this doesn't: dispatch has TWO stages and only the first was the problem before.
1. **Resolution (typecheck):** I extended `resolve_free_call_overload` to disambiguate by arity+arg0-type and it CORRECTLY picks the user's symbol (verified by trace: chain walk returns the user `len` sym 311). So resolution is fixable and works.
2. **Mangling (backend `x86_resolve_sym_name`, x86_regs.ax:236-302):** a fn whose first param (unwrapped) is a **struct/sum** mangles to `ax_<Struct>_<name>` (unique per receiver); a fn with **scalar/str/no** first param mangles to bare **`ax_<name>`**. User `len(str,i32)` AND stdlib `len(str)` BOTH bare-mangle to `ax_len` → link-time ODR collision → the call binds to whichever `ax_len` was emitted first (stdlib). **Picking the right symbol at stage 1 does NOT help — codegen still emits `call ax_len`.** So my arity fix was REVERTED (ineffective end-to-end; tree restored to `fa9e76a`/`BC276EF6`).
   - BUG#80 `get(Box)` worked because user→`ax_Box_get` vs stdlib→`ax_Vec_get` (distinct). Zero-arg `run()`/`close()` worked because user→`ax_run`/`ax_close` vs stdlib→`ax_scheduler_run`/`ax_File_close` (distinct). The collision ONLY bites when the user fn AND a bundled fn BOTH bare-mangle (both scalar/str/no first param): `len`, `concat`, `is_valid_utf8`, etc.
   - `SYM_FLAG_MODDUP` (=2048, resolver.ax:551-553) already produces a unique `ax_<name>__m<idx>` — BUT it's only set when the overload head lives in a DIFFERENT tree. The driver CONCATENATES stdlib+user into ONE tree, so same-tree collisions never get MODDUP. At `define()` (resolve phase) first-param types aren't resolved yet, so "does it bare-mangle?" can't be decided there.

**Proper fix (BACKEND, B==C required, own session):** make the user's colliding fn uniquely mangled end-to-end. Two options: (a) at typecheck (types resolved) detect a chain with ≥2 bare-mangling members and set MODDUP on the non-stdlib one so both def+calls emit `ax_<name>__m<idx>` (calls already go through `x86_resolve_sym_name(sym_idx)` so they'd agree) — needs the arity-resolution fix too; or (b) reject at typecheck with a clean diagnostic (convention BUG#53/#80/#88) — simplest/safest but UX friction on common names (`len`). Bare-mangle predicate = mirror x86_regs.ax:264-302 (no params OR first param unwrapped not struct/sum, not extern/MODDUP/special-name). All REPORTED symptoms were the zero-arg case (already FIXED `730ee9e`); this is a constructed extension.

---
(original diagnosis below, kept for reference)

## Symptom
A user module-level `fn` whose name collides with a bundled stdlib `pub fn` silently binds to the STDLIB overload when arities differ:
- `fn run() -> i32: ...15...` then `run()` → **0** (binds to `std/scheduler.ax:332 run(mut self)->i32`; reads garbage `self`).
- `fn close() ...` → **139 (SIGSEGV)** (binds to `std/io.ax` `close(self)` → syscall on garbage).
- `fn clear() ...` → **1** (binds to `std/collections.ax` clear).
- Non-colliding names (`start`, `step`, `run2`) work. This is the true cause behind probe p15 (match/continue/return were all fine in isolation).

## Root cause
`typecheck.ax:709-722` `resolve_free_call_overload`. When the call has NO arguments (`first_arg_node == 0`, ~L714) it returns `sym_idx` unchanged — so it cannot re-pick when the name resolved to a bundled stdlib overload. BUG#80 [[bug80-free-call-overload-collision]] fixed the SAME class by disambiguating on `arg[0]` type, which structurally cannot handle a zero-arg call or an arity mismatch. Silent accept-then-miscompile (sometimes crash), no diagnostic — same convention family as BUG#53/#68/#88.

## Proposed fix (frontend, not yet done)
Add **arity-based** disambiguation to `resolve_free_call_overload`: thread the call's argument COUNT in; when arg0-type resolution is unavailable/ambiguous, prefer the overload whose parameter count matches the call (user `run()`=0 params vs stdlib `run(self)`=1). Among equal-arity candidates, prefer the one declared in the user's MAIN module over a bundled one. If still not unique → REJECT with a diagnostic (BUG#53/#80 convention). Frontend-only → A==B + async fixpoint. Oracle `t_freefn_collision_noarg` (`fn run()->i32{..15}` → 15; + a reject-mode variant if policy is reject-on-ambiguous).

## CLEAN areas from that probe round (don't re-probe)
Nested aggregates crossing fn boundaries + mutation (reference semantics by design), globals RFC 0017 P1 (loop bound / while-RMW; aggregate globals cleanly rejected = P2 limit), generics×option/result×aggregates (Option[Struct].unwrap().field, HashMap[K,Struct].get().field, Vec[Vec[T]]), multi-field variant×match×Vec, non-generic recursive linked list, nested for/while break/continue, int width/cast/unsigned-compare. All correct. (Exit codes are 8-bit: 1000&0xFF=232 — use `$LASTEXITCODE`/masked oracle.) **Toolchain gotcha: run the compiler from repo root — it resolves bundled stdlib relative to CWD, else emits a broken binary.**
