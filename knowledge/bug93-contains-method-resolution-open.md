---
name: bug93-contains-method-resolution-open
description: "BUG#93 CLOSED. (a) FIXED: HashSet.contains inline-match. (b) FIXED 9f89b64: m.contains(k) on HashMap silently miscompiled to getfld+call*garbage SIGSEGV — reject gate was receiver-agnostic; made self-type-aware (base-name). + added HashMap.contains_key."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

**BUG#93 — `contains` method resolution mess.** Found while re-probing BUG#90 [[bug90-option-method-segfault-open]] (the `hmc` "unwrap" crash was actually `contains`).

### (a) ✅ FIXED (pending commit) — `HashSet.contains` returned true for ANY key
`std/collections.ax` `contains[T](self: HashSet[T], ...)` used INLINE match arms:
```
match self.map.get(item):
    Some(_): true      # inline arm -> mis-lowers to ALWAYS first arm
    None: false
```
Inline match arms [[inline-match-arm-unsupported]] mis-lower (always take arm 0), so `contains` always returned true (`s.contains(7)` on an absent 7 → true). Fix: rewrite to BLOCK-form arms (`Some(x):` newline `return true` / `None:` newline `return false`). Verified: `s.contains(present)`=true, `s.contains(absent)`=false.

### (b) ✅ FIXED `9f89b64` (A==B `B7531B55`, 185/185) — cross-type method call now REJECTS instead of segfaulting
**Real mechanism (memory below was half-wrong):** `m.contains(k)` did NOT mis-resolve to HashSet.contains — resolution *correctly* returned 0 (self-type guard via `is_method_compatible` base-name works). The bug: the accept-then-miscompile **reject gate** (`typecheck.ax` ~3224) name-existence scan was **receiver-AGNOSTIC** — ANY fn named `contains` (HashSet's) → name_exists=true → NO reject → callee (a NODE_FIELD_EXPR) lowered as getfld + `call *garbage` → SIGSEGV. **Fix:** for a reliably-named concrete aggregate receiver (struct/sum/generic-inst, real name, NOT Option/Result), scan self-type-aware via new `has_method_for_base(mname, rec_base)` (base-name match over ALREADY-INFERRED symbols only — never pre_infer, cross-module safe). **2 exclusions or self-host false-rejects:** (1) Option/Result receivers, (2) intercepted method NAMES `unwrap/unwrap_err/is_some/is_none/is_ok/is_err` (appear on payload-typed receivers e.g. `tok.unwrap()` where tok=Token — trap that broke first fixpoint attempt). Also added `HashMap.contains_key(k)->bool` (Robin-Hood probe, distinct name from HashSet.contains → no dispatch ambiguity). So `m.contains(k)` REJECTs, `m.contains_key(k)` works. Oracles `t_hashcontainskey(7)` + `t_crosstypereject(reject)`. LESSON: reject-gate name scans must be self-type-aware for named aggregates, agnostic for primitives + Option/Result-intercepts.

### (b-orig) OLD DIAGNOSIS (superseded) — "mis-resolves to HashSet.contains → SEGFAULT"
`m.contains(k)` where `m: HashMap[K,V]` — HashMap defines get/insert/remove/len but NO `contains`/`contains_key`. The call resolves to `HashSet.contains` (self: HashSet[T]) despite the receiver being a HashMap → wrong struct layout → SIGSEGV (deterministic). ROOT: **method resolution is self-type-BLIND** for same-named generic methods — it matches by NAME and does not reject when the receiver type ≠ the method's `self` type. Confirmed: adding a `HashMap.contains` made `HashSet.contains` dispatch to the HashMap one instead (broke HashSet) — so it's not just "missing method", the resolver genuinely can't disambiguate two generic `contains`.

**Fix directions (pick in a dedicated iteration):**
1. Make method resolution REJECT a call whose receiver base type ≠ the candidate's `self` base type (convention: clean "no method `contains` on HashMap[K,V]" diagnostic, class BUG#53) — this alone stops the segfault. Look at `method_ret_type` (typecheck.ax:958, checks `fp == unwrapped_rec` — but the ACTUAL dispatch in air_builder may not) and air_builder method-call resolution (`match_mangled_method_raw_bytes`, ~L1494-1530).
2. THEN optionally add real `HashMap.contains_key`/`contains` (get-probe returning bool) — only safe once resolution disambiguates by self type.

**Workaround today:** use `m.get(k).is_some()` instead of `m.contains(k)` on a HashMap (works — Option methods fixed).

Backend/typecheck; if air_builder dispatch changes → B==C. Oracle: `m.contains()` on HashMap rejected/works; HashSet.contains present/absent correct.
