---
name: bug-option-array-element-read-open
description: "FIXED bf1361c: reading a builtin Option/Result element out of a FIXED array now works (construct/index/match/for-in/is_some/unwrap). Scoped typecheck fix; global retype was unviable (self-host oscillation). Vec[Option] still out of scope."
metadata: 
  node_type: memory
  type: project
  originSessionId: 3a1742c4-c4b5-4670-948f-940237abef60
---

## ✅ FIXED `bf1361c` (2026-07-12, frontend A==B `8AF4B46C`, regression 194/194)
Fixed arrays of Option/Result now fully work: construct, `arr[i]`, `match arr[i]`, `for o in arr` (incl. None), `.is_some()`/`.unwrap()`, Result Ok/Err. Oracles `t_optarray(40)` `t_resultarray(40)` `t_optarrayiter(123)` `t_optarraymethod(40)`.

**The fix (4 parts, all `typecheck.ax`, SCOPED so self-host untouched — global retype oscillated, see history):**
1. `NODE_ARRAY_TYPE` (~4258): resolve an `Option[T]`/`Result[T,E]` ELEMENT (base-name test on the inner NODE_GENERIC_TYPE) to `register_option`/`register_result` (kind 11/12, size 16, pointer-repr) so the store stride == the OP_INDEX load. Scoped to array-element position ONLY (compiler source has no `[Option;N]` → fixpoint safe).
2. `match_is_confidently_partial` (~1574) + `NODE_MATCH_ARM` (~2418): OPTION/RESULT have NO sumtypes entry — `.extra` is the inner/ok TYPE id, NOT a sumtypes index. Feeding it to `sumtypes.data[]` read a garbage `variants` ptr → **crashed the COMPILER** (surfaced once a scrutinee actually typed OPTION). Fix: return false in the partiality check; in the arm resolve the pattern payload DIRECTLY from the entry (Some/Ok→`.extra`, Err→`.name_id`, None→none, bare binding→whole value).
3. method reject-gates — call-level (~3352 `is_opt_res_m` block) AND `NODE_FIELD_EXPR` member (~3991): the backend-intercepted `is_some/unwrap/is_none/is_ok/is_err/unwrap_err` have no method symbol → an OPTION/RESULT-typed receiver false-rejected (`no method 'is_some' on type ''`). Accept these methods on OPTION/RESULT receivers (air_builder lowers them by name).

**BÀI HỌC:** Option=kind-11 / Result=kind-12 have `.extra`=inner type (NOT a sumtypes index) — any `sumtypes.data[extra]` on them reads garbage. Typecheck was built assuming it NEVER emits kind-11/12; emitting them (even scoped) surfaces every place that assumed a real SUM (crash sites + reject-gates). The intercepted Option/Result methods live ONLY in air_builder by name — every typecheck method-existence gate must whitelist them for OPTION/RESULT receivers.

**`Vec[Option[T]]` — VERIFIED WORKING** (2026-07-12, follow-up): `v[i]`, nested `v.get(i)` (Option[Option]), and `for o in v` all correct (→40). The earlier "Vec[Option] segfault/link" was a MALFORMED PROBE (`Vec()` + `-self-link`); the right form is `Vec[T].new()` built `-O1` WITHOUT `-self-link` (Vec needs `import std.collections` dynamic-link, like the passing `t_vecindex`/`t_vecset`). No bug here. Repros `scratch/probe/vopt.ax`/`vget.ax`/`vfor.ax` + `p3*.ax`.

---
## (history) The confirmed bug + 2 rejected approaches
Reading a **builtin `Option`/`Result`** element out of a fixed array `[Option[T];N]` (or `Vec[Option[T]]`) **silent-SEGFAULTed** at run (exit 139). Construction was fine; crash on ELEMENT READ. User sum (plain OR generic) as element already worked — only builtin Option/Result failed. Root: element type fell back to i32 while the ctor lowered an 8-byte pointer-repr box → OP_STORE/OP_INDEX truncated it → match/method derefs garbage.

**Rejected approach 1 — retype Option/Result globally in typecheck** (annotation NODE_GENERIC_TYPE + ctor NODE_CALL_EXPR): DESTABILIZES SELF-HOST — the compiler's own source uses Option/Result pervasively; kind-11/12 shifts every EQ/COPY/MAKE_REF path (RFC 0012) → self-compilation enters a **period-3 fixpoint cycle** A→B→C→A. Annotation-only edit alone also oscillates. This is why air_builder does box-typing LOCALLY "not via typecheck."

**Rejected approach 2 — scoped array hook ALONE** reached A==B but then crashed the compiler in match inference (the sumtypes[extra] bug, part 2 above) + false-rejected methods (part 3). The SHIPPED fix = scoped hook + all the downstream OPTION/RESULT guards. Related: [[bug78-array-of-option-none]] (construction side), [[bug90-option-method-segfault-open]], [[bug-option-arith-miscompile-open]].
