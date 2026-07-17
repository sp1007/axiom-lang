---
name: bug-accept-then-miscompile-cluster-0711
description: "FIXED cluster 2026-07-11: 4 accept-then-miscompile diagnostics (undefined method, non-exhaustive no-payload enum, free-fn arity, nonexistent field). Fixpoint 75998F4C, 154/154."
metadata: 
  node_type: memory
  type: project
  originSessionId: b9556362-e358-4881-8a42-df3ba54b80bb
---

✅ **Accept-then-miscompile diagnostic cluster (BUG#53 convention) — 2026-07-11.**
daily-driver = A==B **`75998F4C`**, `origin/main`=**`fbc7272`**, **154/154**. All
frontend-only (typecheck.ax), all surfaced by two bug-probe rounds. Each was a
silent accept → air_builder garbage lowering → wrong value/SIGSEGV, now a clean
compile-time REJECT.

1. **Undefined method** `1fa2d63` — `p.ghost()` (name ≠ field/method/UFCS-fn) →
   indirect-call through garbage. [[bug-undefined-method-call-segfault]]. Oracle `t_nomethod`.
2. **Non-exhaustive no-payload enum match** `0ccdb97` — `match c` on `Red|Green|Blue`
   omitting Blue was treated as TOTAL (bare-variant `Red` binding-pat misread as
   catch-all) → missing-return diagnostic suppressed → garbage fall-through. Fix:
   `match_is_confidently_partial` variant-aware (new `binding_pat_is_variant`,
   mirrors air_builder find_variant_info-by-name). Oracles `t_nonexhenum`/`t_exhaustenum`.
3. **Free-fn arity** `41ab338` — `add3(5)` for `fn add3(a,b,c)` returned garbage;
   `add2(5,6,7,8)` dropped extras. Fix: compare arg count to `fi.params.len` in the
   concrete-FUNC branch. Oracles `t_arityfew`/`t_aritymany`/`t_arityok`.
4. **Nonexistent field value-load** `fbc7272` — `c.zzz` on `Cell{x}` read `c.x`
   (offset 0). Fix: reject in value-load NODE_FIELD_EXPR. Oracles `t_badfield`/`t_goodfield`.
5. **Vec direct subscript** — `v[i]` on `Vec[T]` read the `{data,len,cap}` header
   (`c[1]`=len; wide elems segfault) instead of `vec.data[i]`. First REJECT `d94aa8e`
   (typecheck NODE_INDEX_EXPR only typed pointer/slice/array/str → Vec fell through
   UNKNOWN → OP_INDEX on Vec address), then **IMPLEMENTED as RFC 0021 `afe0dbc`**:
   typecheck types index to elem T (`get_generic_args[0]`), air_builder
   `lower_index_expr` loads `data` field (OP_GET_FIELD) before OP_INDEX (mirrors
   for-in-vec ~3688). Unchecked O(1) like ptr/array `[]`; `.get(i)`=bounds-safe.
   Verified Vec[i64]+Vec[str](16B)+O0/O1. **Lvalue `v[i]=x` cũng DONE `a56881e`**
   (lower_assign load `data` field trước OP_STORE, mirror read side; write-side là
   silent miscompile mới do read-side typing mở ra — store hit header). Oracles
   `t_vecindex`(7)/`t_vecset`(42). `rfcs/0021-vec-index-operator.md`.

⚠️ **KEY LESSONS (repeated self-host breakage during this cluster):**
- Typecheck rejects that depend on the RECEIVER TYPE false-positive: cross-module
  types (`in_file: File`) / fields resolve LATE during inference (read as placeholder
  "R"). → For #1 and #4 use a **RECEIVER-AGNOSTIC** name scan (reject only when the
  name matches NO function / NO struct-field ANYWHERE). Robust to resolution timing.
- Arity (#3): scope to NODE_IDENT free-fn callees only (method/UFCS/generic arg+receiver
  shape is unreliable → `'map' expects 1 found 0`). EXCLUDE the internal variadic
  intrinsic `compiler_intrinsic` (`@compiler_intrinsic("op")[T](...)`, declared 1 param).
- Every typecheck reject: run the A==B gate BEFORE trusting it; a false-positive on
  the compiler's OWN source makes stage A error-out and A!=B (with a STALE axc_fpB.exe
  giving a misleading B hash == the pre-change fixpoint). Read `fpB.log` (UTF-16, strip
  nulls) for the `error:` lines to find the offending call.

6. **Scalar member access** `05fff84` — `n.foo` on a primitive (i32/bool/f64…) →
   SIGSEGV (garbage field load). Fix: reject in a new TYPE_KIND_PRIMITIVE branch of
   value-load NODE_FIELD_EXPR. ⚠️ receiver-agnostic scan MUST include struct fields
   (not just fns): an operator-overload result `a+(3 as i64)` reads as its RHS scalar
   i64 before the Num return type settles, so `d.v` (v∈Num) would false-reject —
   caught by regression `t_opmix` first try. Oracle `t_scalarmember`.

**Residual / NOT yet caught (safe follow-ups, low priority):**
- ✅ **Forgot-unwrap `.x` DONE** `0036495` (2026-07-11, A==B **`19E26F1A`**, 164/164):
  `o.v` on `Option[Cell]` AND chained `v.get(i).v` silently read box/tag=garbage 8 → now
  REJECT. Fix = **POST-TYPECHECK pass (Phase 5 of run_type_checker)** — runs after inference
  so node_types fully populated + generic returns recovered → catches the CHAINED case the
  reverted inference-time attempt couldn't. Walk all NODE_FIELD_EXPR value-loads (exclude
  call-callees via a callee bitmap); if receiver type is Option/Result → reject. Detect
  Option/Result by BOTH raw TYPE_KIND_OPTION/RESULT AND monomorphized sum name containing
  `Option__`/`Result__` (mono name = `_AX_std_Option__Cell`, so `starts_with("Option")`
  FAILS — must use `contains "Option__"`). Oracles `t_forgotunwrap`(reject)/`t_unwrapok`(42).
  ⚠️ **LESSON: debug-print in `run_type_checker` post-inference is INERT** (like ssa_opt) —
  `ax_puts_local`/`ax_printf_local("[DBG..")` produce NO output there; only `"error:"`-prefixed
  ax_printf_local flushes (driver flushes on the diag path). Diagnose by behavior-bisection
  (temp `diags_count++` + `"error: PROBE.."` lines), not plain printf.
- ✅ **Method-call arity mismatch DONE** `de1ff47` (2026-07-11, A==B **`256741CD`**, 162/162):
  `p.getx(99)`/`p.add(3)` silently accepted → now REJECT. Non-generic methods aren't bound
  by resolve_method_sym (generic-only, BUG#45), so mirror `method_ret_type`'s name+receiver
  scan: param[0]=receiver → INSTANCE call expects (decl params - 1), STATIC `Type.m(recv,...)`
  expects (decl params). Overload-safe (reject only if name+receiver methods exist but none
  match arity); gated non-generic + STRUCT/SUM receiver. Oracles `t_methodmany`/`t_methodfew`/`t_methodok`(40).
- Const div-by-0 / const out-of-range shift — loud runtime trap, optional compile reject.

Related: [[bug93-qualified-str-call-segfault]], [[missing-return-diagnostic-shipped]].
