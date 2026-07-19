---
name: backlog-open-items
description: "Reconciled list of genuinely-open AXIOM backlog items (from next-step-15 TASK QUEUE + RFC follow-ups) — cross-checked against what already shipped, so autopilot doesn't redo completed work."
metadata: 
  node_type: memory
  type: project
  originSessionId: f482e19c-5639-4d54-a2ca-f24d4c2c01aa
---

Reconciled 2026-07-09 from `docs/next-step-15.md` + `docs/next-step-15-sub-1.md` (dated 2026-06-26, user flagged as "dang dở") against MEMORY.md. **Most of those queues already SHIPPED** (RFC 0002–0007, ADT v1 codegen). Do NOT redo shipped work — verify against git/memory before starting.

## Genuinely OPEN (feature backlog)
- **Negative match-arm literals** (`match n: -5: ...`) — DON'T PARSE (2026-07-19c): `parse_pattern`
  (parser.ax:876) only accepts TK_INT_LIT/FLOAT/STRING/CHAR/TRUE/FALSE/NIL for a NODE_LITERAL_PAT;
  a leading `-` (TK_MINUS) → "expected pattern" (both native AND C paths reject, so it's a parser
  gap, not a codegen bug — and RFC 0028's jump-table bsearch never sees them, so it's safe there).
  Fix needs: (1) parse `-` + INT/FLOAT literal in parse_pattern → NODE_LITERAL_PAT carrying the
  SIGN; (2) negate the value at BOTH native lowering sites (air_builder.ax lower_match ~L3190
  `parse_int_from_str(ltext)` AND try_lower_match_bsearch's value collection); (3) maybe cgen.
  BLOCKER on a clean impl: **all 16 u16 flag bits are used** (ast.ax:101-116) so there is NO free
  flag bit for "negative" — either contextually reuse a decl-only bit (e.g. FLAG_IS_MUT) on a
  LITERAL_PAT, or carry the value another way. Modest value; bounded but multi-site.
- ~~**BUG#82 globals**~~ — ✅ RFC 0017 P1 (scalar storage) + P2 non-const scalar init + P2 aggregate (struct/array/tuple) `6264ff6` + **P2 pointer-repr (Option/Result/sum) globals SHIPPED `3a44577`** (A==B==C `dc6a18a5`, 236/236; 8B box-ptr slot + store_is_ptr_sum, scoped annotation resolution). See [[bug82-global-var-semantics-open]]. **RFC 0017 storage surface COMPLETE for all value categories.** Còn defer (low value): uninit-decl (`mut g: S` no-init = parse error), .bss, ELF `.data`.
- **`for x in <collection>`** iteration — **fixed arrays ✓ (P1 `96dd586`), Vec[T] ✓, string ✓, HashMap ✓ (`4b1a8f4`)**. `for k in m` iterates occupied KEYS: air_builder lower_for walks 0..cap, body guarded by `occupied[i]`, binds `keys[i]`; A==B (compiler ko iterate hashmap, shared for-path byte-identical). Oracle `t_hashiter(15)`. **DONE** — no remaining for-in gaps. (HashMap VALUES/entries iteration = future if needed.)
- ~~**Multi-field variant** `Rect(i64,i64)`~~ — ✅ **SHIPPED** RFC 0019 `59bc731` [[rfc0019-multifield-variant-shipped]] (desugar-to-synth-struct). Closes BUG#81. Robust: scalar/mixed/aggregate fields, 2-3 fields, arrays+for. DONE.
- ~~**str / >8B variant payload**~~ — ✅ **SHIPPED** `53e62ed` (extends RFC 0019: wrap single >8B primitive in flagged synth struct → payload is an 8B ptr). [[rfc0019-multifield-variant-shipped]]. DONE.
- **`Self` type** — ✅ **Self-as-param CONFIRMED WORKING 2026-07-19c** (probe SP6/SP8/SP9 all →42): `other: Self` resolves to the receiver type in in-struct methods AND free-function methods AND combined with `-> Self` return + method chaining (`a.merge(b).merge(c)`). `-> Self` return + chaining already worked (`575a69a` [[self-type-return-fixed]]). **Only INTERFACE VTABLE DYNAMIC DISPATCH remains open** (calling a method through an interface-TYPED value / `Box[Interface]` — BUG#71, rejected pending a vtable; this is what std/log's `Box[LogSink]` needs, and the real remaining chunk of the "Self+vtable" greenlit item). NOTE: untyped `self` (`fn m(self)`) requires the method be INSIDE a struct body (RFC 0002 `current_struct`); a top-level free-fn method must write `self: Type` — by design, not a bug.
- **Rewrite aspirational std modules** (iter/json/log/net/fmt/…) from out-of-grammar dialect to real grammar — see `std/MODULE_STATUS.md`; needs enum/Self. Backlog, not blocking.
  - **`to_str` number formatting** — ✅ **i64 DONE** `0be41f8` (`t_tostr(88)`), ✅ **f64 DONE** `b81293a` (overload by arg type, mirror `ax_sprintf_local`, 6-digit frac; `t_tostrf(91)`), ✅ **hex DONE** `afbf5e3` (`to_hex(i64)` lowercase; `t_tohex(123)`), ✅ **general radix DONE** `d92a6e4` (`to_radix(i64,base)` base 2..36, subsumes bin/oct/hex; `t_toradix(99)`). ✅ **numeric UFCS DONE** `796a7a8` (`n.to_str()`/`n.to_hex()`/`n.to_radix(b)`/`f.to_str()` — widened UFCS dispatch to ALL primitives; `t_numufcs(11)`). ✅ **numeric UFCS widening DONE** `a02db69` (narrow-int receiver i8/i16/i32/u8/u16/u32 → i64 free fn, gates-only no cast; `t_numufcswide(7)`; also closed misdiagnosed "generic-inst field mono Layer 2"). ✅ **u64 decimal DONE** `732963c` (`to_str(u64)` unsigned overload; `t_tostru64(19)`). **STILL OPEN:** uppercase hex (cosmetic, low value). Bigger fmt-module rewrite (json/log/iter still out-of-grammar) remains backlog. NOTE: free-fn overload by arg type CONFIRMED working (to_str i64+f64 coexist).
- **INTRINSIC-NOT-LOWERED bug class (native backend):** many `@compiler_intrinsic("str_*")` are ONLY mapped by the C backend (cgen.ax → `ax_str_*`); the native/self-link path leaves them unlowered → garbage RAX (0/false/segfault) SILENTLY. Audited std/string.ax 2026-07-11: split/char_count/is_valid_utf8 FIXED (reimpl pure-AXIOM). Remaining `@compiler_intrinsic` bodies: std/sync.ax atomics+channels (concurrency, no native threading runtime — bigger). **When adding/using any stdlib fn, verify it works on the NATIVE path, not just C.**
  - **RE-AUDIT 2026-07-18 (native intrinsic reachability, read-only):** enumerated all `@compiler_intrinsic` in std/*. Native path (air_builder.ax ~L1571) compile-time-folds `is_windows/is_linux/is_macos/os_name/arch_name/path_separator` + `size_of` (builtin). REACHABLE formatting verified CLEAN on native driver `633913E9`: `to_hex(255)`="ff", `to_str(-42)`="-42", `to_str(3.5)`="3.5" (correct lens 2/3/3, no garbage). The UNLOWERED intrinsics are all in NON-REACHABLE code: atomic_load/store/cas (std/sync.ax, blocked on threading runtime — big), and simd_*/q*(quantum)/sha*/rng_* (std/gpu,quantum,crypto — aspirational, out-of-grammar, don't parse). **CONCLUSION: no tick-sized native-intrinsic miscompile remains** — remaining risk is gated behind blocked (threading) or non-parsing (aspirational) modules. Don't re-audit without new evidence.
- ~~**Tuple-literal expressions** (`(a, b)`)~~ — ✅ **RFC 0022 P1 SHIPPED** `af9b15a` (A==B `C8408F62`, 244/244). `(e0,e1,...)` = anonymous struct fields `_f0.._fN` (reuse RFC 0019 synth-struct); `.N` field access; NODE_TUPLE_EXPR(69); parser NUD-comma + `.`-INT LED; typecheck register_struct; air lower_tuple_lit. Oracle t_tuple(66). **P2 SHIPPED** `7be0682` (A==B `D1D77721`, 245/245): tuple TYPE annotations `(T0,T1)` (NODE_TUPLE_TYPE) → annotated let + params + `-> (T..)` returns + tuple GLOBALS. 2 cơ chế bắt buộc: (1) **canonicalization** (register_tuple_type reuse `__tup` struct trùng element types → literal==annotation cùng type_id; guard prefix `__tup` tránh alias RFC 0019 `__mfv_`); (2) **element coercion by expected tuple type** (nếu ko, `(10,20)`=i32{8B} vào slot i64{16B}→field 2 rác). Tuple ARG lệch width tuple PARAM (unannotated `let t=(10,15)`→{i32,i32}) = REJECT (BUG#53). Oracle t_tuple2(84). **P3 SHIPPED** `bb2681f` (A==B `5E09C11B`, 246/246): tuple destructuring `let (a,b)=EXPR`/`mut (…)` — desugar PARSER-only vào temp (eval 1 lần) + 1 binding/element qua field access, splice children vào enclosing stmt list (NODE_BLOCK là thứ duy nhất parse_stmt trả ở stmt-position → unambiguous). `_`=skip. 1 backend-touch: `lower_ident` tin payload đã-resolve khi symbol là temp `__td`-prefix (nó re-derive tên từ TOKEN để guard stale payload → synth ident token placeholder bị reject → getfld base=0 → segfault; AIR dump chỉ ra). Oracle t_tuple3(50). **P4 SHIPPED** `d273ce1` (A==B `4B3AE685`, 247/247): chained `t.N.M` — `.` LED thấy FLOAT selector (`0.1`) → split thành nested `._f0._f1` (2-level; sâu hơn cần intermediate bind). ACCEPT-only, float literal/method ko ảnh hưởng. Oracle t_tuple4(10). **RFC 0022 TUPLE FEATURE-COMPLETE** (literal, `.N`/`.N.M`, TYPE annotation, params/returns/globals, destructuring). **P5 DEFER (niche):** match/param destructuring, nested patterns (`let ((a,b),c)=…`). rfcs/0022.
- **Jump-table dispatch** optimization (RFC-gated, semantics-preserving) — self-host uses many `if/elif op == OP_*` chains; user flagged interest 2026-06-26. next-step-15 item 7 (truncated).
- ~~**Vec/collection HOF methods**~~ — ✅ **SHIPPED**: map/filter/fold (`b956c01`), any/all/find (`d57de80`), count/position/take_while/skip_while/reverse (`f7ed1e6`). Surface non-capturing HOÀN CHỈNH (276/276). CÒN THIẾU (blocked): for_each/reduce (cần capturing lambda — BUG#73 zero-capture reject), enumerate/partition/zip (tuple×generic element bug § dưới). Ghi chú lịch sử: — NOW UNBLOCKED by `2cc67ed` (inline lambda × generic monomorphization works, incl. U≠T & multi-arg — see [[bug-lambda-generic-fntype-infer-open]]). Currently CLEAN-REJECT ("no method 'map' on type Vec"). `std/vec.ax` is the MATH lib (Vec2/3/4); dynamic `Vec[T]` (push/get/index) is compiler-builtin/runtime — need to find where Vec[T] methods live and add generic HOFs returning a new Vec[U]. `Option.map`/`filter` already work (std/result.ax). Moderate feature; good demonstrator of the lambda fix. Watch: element aggregate/16B, U≠T Vec[U] alloc, and the deferred tuple-element-in-generic coercion (§ below) if elements are tuples.

## ✅ str/bytes globals SHIPPED (2026-07-12 `288c86a`)
- ~~**`str` module-level globals REJECTED**~~ — ✅ SHIPPED `288c86a` (A==B==C `81522e76`, 238/238). `mut g: str = "..."` now stores real 16-byte inline `{ptr,len}` in `.data`, runtime-init. 3 pieces: typecheck accept 16B PRIMITIVE; `collect_global` 16B slot + runtime-init (never fold, `is_inline16` flag); `x86_selector` OP_LOAD 16B-non-aggregate dest → two-8B-halves inline load into dest home (single 16B MACH_LOAD into GP reg invalid; mirror OP_GET_FIELD str branch). OP_STORE unchanged (emit_block_copy LEAs the 16B inline source). Oracles t_globstr(42, cross-fn reassign+.len)/t_globstridx(42, byte-index). **RFC 0017 global storage now covers EVERY value category: scalar const/non-const, aggregate (struct/array/tuple), pointer-repr (Option/Result/sum), 16B-inline (str/bytes).**

## ✅ CLOSED 2026-07-18 — tuple × generic element coercion (re-verified all pass)
- **✅ VARIANT-CTOR half FIXED `1aff7ca`** (2026-07-14): `Some((20,22))`/`Ok(..)`/match-bound/
  direct-unwrap/Result/3-tuple all correct now (thread `expected` to try_instantiate_variant_call
  + sticky NODE_TUPLE_EXPR coercion). Oracle t_tupctor(60). See [[bug-tuple-generic-payload-unwrap-open]].
- **✅ Vec-push half CLOSED 2026-07-18** — re-probed the full matrix on daily driver `F81E2A77`, ALL
  correct at O0+O1: `Vec[(i64,i64)]` 2-push index-read (37), single-push (37), `Vec[(i64,str)]`
  mixed index+`.len` (47), 3-tuple (37), `for t in v` iteration (37), `HashMap[str,(i64,i64)]` value
  (42). The miscompile was fixed by `aa419a2` (tuple-generic-element mono, `is_generic` recursive over
  `__tup` field — see MEMORY.md [[bug-vec-generic-tuple-element-mono-open]]) which postdated this
  entry; the backlog note was stale. Banked oracle **t_vectupmix** (Vec[(i64,str)] mixed, exit 47).
- **(historical) `Vec[(i64,i64)].push((10,20))` miscompiled** — tuple literal in a GENERIC call arg (`ax_push`) không coerce về monomorphized element type `(i64,i64)` → literal = `{i32,i32}` 8B đẩy vào Vec element stride 16B → đọc sai (probe b3: 32 thay vì 37, O0==O1). Cùng CLASS với struct-field bug đã fix `696d1b2` nhưng ở path GENERIC-arg (P2 fn-arg coercion CỐ TÌNH loại `is_generic_call`). Fix = extend generic-arg/mono coercion cho `__tup` element type. Non-generic struct-field + fn-arg + annotated let/return/global ĐÃ đúng. Same fix pattern likely applies to HashMap/array-of-Vec tuple elements. **ĐÃ THỬ (2026-07-13, REVERTED):** post-pass ở generic-call arg inference (typecheck ~L3110, sau `infer_generic_type_args`) re-infer NODE_TUPLE_EXPR arg với `inferred[gk]` khi param là bare gen-param `T`→`__tup`. KHÔNG hiệu quả (v1 vẫn 32) VÀ có hại (v2 O0=32/O1=127 divergence) → mono instantiation có thể chạy TRƯỚC/độc lập với re-inference nên arg node_type mới ko khớp instance đã mono; hoặc receiver element type ko phải canonical `__tup` như kỳ vọng. **Cần iteration chuyên sâu map mono flow (khi nào push instantiated, element type lưu ở đâu) TRƯỚC khi sửa — ĐỪNG hotfix.** Rủi ro cao (generics = self-compile critical). **INVESTIGATION 2026-07-13 (AIR dump xác nhận):** `dump-air` cho `Vec[(i64,i64)].push((10,20))` cho `%7: t3 = iconst 10` → **t3=i32**, tuple build {i32,i32} 8B (int-literal element default i32) trong khi Vec element = `__tup{i64,i64}` 16B → push store 8B vào slot 16B → single-push CRASH (127), two-push đọc rác (32). Re-infer post-pass thất bại VÌ: mono instantiate_function nhân bản body push với T=__tup RIÊNG, nhưng arg node (caller tree) re-infer SAU/độc lập ko propagate — nghi ordering: arg đã bị mono/overload path tiêu thụ trước post-pass. **Next-step cụ thể:** (a) trace-print xác nhận post-pass CÓ fire + inferred[gk] có phải __tup{i64,i64} 16B; (b) tìm điểm mono đọc arg node_type; (c) coerce arg TRƯỚC mono, hoặc reject sạch (BUG#53) nếu ko coerce được. Cần rebuild+trace (multi-build).

## ✅ FIXED `4c66c86` — generic free-fn explicit type-args f[T]() inside generic body
- **FIXED 2026-07-13** (A==B `44D8B1A7…`, 302/302, oracle t_genfnexpltypearg). Root =
  NODE_INDEX_EXPR inference (typecheck.ax ~4402): `make[T]` with `has_generic_arg` was
  registered as a generic-inst named after the FUNCTION (`make`) — correct for generic TYPES,
  wrong for generic FUNCTIONS. Fix: when the indexed symbol is `SYM_FUNC`, override result_type
  with the fn's RETURN type (pre-infer signature if needed). Concrete-arg `make[i64]()` already
  worked (mono path); `Vec[T].new()` workaround still works. Frontend ACCEPT-only → A==B.
  Original report below (kept for context):
- Inside a generic fn body, `mut result := new_vec[U]()` (calling a generic free-fn that
  returns `Vec[U]`, U = the outer fn's type param) mis-types `result` as the callee name
  (`new_vec`) instead of `Vec[U]` → `error: no method 'push' on type 'new_vec'`. Surfaced
  while adding `Vec.map`/`filter`. **Workaround used: direct struct ctor `Vec[U](data:.., len:.., cap:..)`
  types correctly and works.** So the gap is narrow: it's the RETURN-type resolution of a
  generic free-fn call whose result is a generic-inst parameterized by an enclosing generic
  param, at TEMPLATE-typecheck time (before mono). Likely infer_node for the `:=` RHS call
  falls back to the callee symbol name when the generic return can't be resolved pre-mono.
  Low priority (clean workaround exists); fix = resolve generic-inst return types symbolically
  (keep `Vec[U]` with U as a type-param placeholder) during template typecheck.
  **REFINED 2026-07-13g (repro `/tmp/t_genret.ax` vs `/tmp/t_gr2.ax`):** the INFERRED form
  `mut r := make_vec(seed)` (U inferred from an arg) WORKS end-to-end (BUG#61 recovery block
  typecheck.ax:3424 fires). The break is ONLY the EXPLICIT-type-args form `make_vec[U]()` —
  which is forced when the fn has no value arg to infer U from (zero-arg generic ctor). There,
  `make_vec[U]` is inferred as a GENERIC_INST *type* whose base name is the fn (`make_vec`), so
  `result_type` is already non-UNKNOWN → recovery skipped → `.push` rejects "no method push on
  type 'make_vec'". Root = explicit `f[U](...)` call syntax mis-inferred as type-instantiation
  of the fn name. Fix = when callee ident resolves to a FUNC symbol, treat `[..]` as fn type-args
  (bind + use the fn's return type), not a generic-inst of the name. DEEP (self-host-critical
  generic inference); workaround `Vec[U](...)` direct ctor stands. Defer to dedicated session.

## ✅ CLOSED 2026-07-16 (pm) — free-fn-vs-method collision verified fixed by HOLE#5/6 cluster
- **VERIFIED 2026-07-16 pm** on daily driver @ec8a0d0: user `fn find(x:i64)` (1-arg) → 105 ✓;
  2-arg `find(a,b)` → 42 ✓; user `find` free-fn + REAL `Vec.find(pred)` method BOTH used in
  same program → 110 ✓; batch of 8 flagged method names (count/position/get/any/all/take/first/
  last) → 8/8 clean. The RECONFIRM below predates the HOLE#5/6 two-pass full-signature tie-break
  in `resolve_free_call_overload`; that work closed the method-collision case too. NOT open anymore.
- **(historical) RECONFIRM 2026-07-16 (probe):** user `fn find(x: i64) -> i64` (1-arg, KỂ CẢ trả i64 thường,
  ko chỉ Option) → **SEGFAULT 139**; đổi tên `find`→`myget` → chạy đúng. `find` = Vec HOF method
  (shipped `d57de80`) → free-fn-vs-METHOD collision hole VẪN OPEN. Post-P2/P2.1 escape-activation
  probe (~17 combo aggregate×option×generic×tuple×ctrl-flow×str-field×alias×early-return) SẠCH khi
  tên unique → escape activation KHÔNG gây regression. BẪY probe (lại): mất ~10 probe đuổi "Option-
  from-fn miscompile" ảo trước khi nhận ra `find` collision — **LUÔN đổi tên fn khi shrink repro.**
- **CORRECTION 2026-07-13j:** this is worse than previously logged as "safe reject". Adding
  `Vec.first[T](self)` to std.collections made the oracle t_genfnopt (user `fn first[T](a,b)`)
  **SEGFAULT** (not a clean reject) — the 2-arg user call mis-resolved to the 1-arg stdlib method.
  So a stdlib method whose name matches a user free fn is an accept-then-**miscompile**, not just a
  reject. Practical mitigation applied: DON'T add common-named methods (`first`/`last`/…) to bundled
  stdlib (see std/collections.ax comment); only added collision-safe `is_empty`/`extend` (`5bd9936`).
  Real fix = harden free-fn-vs-method overload resolution by arg0-type/arity (BUG#80 class, deeper).
  **INVESTIGATED 2026-07-13j (read-only):** `resolve_free_call_overload` (typecheck.ax:848-853)
  ALREADY walks the overload chain preferring an overload with matching arity AND arg0-compatible —
  yet `Vec.first` still broke `first(V2,V2)` (2-arg user fn). So the hole is SUBTLER than a missing
  arity check (that exists); likely the method symbol `Vec.first` is/ isn't in the plain-`first`
  overload chain as expected, or method-vs-free registration diverges, or the segfault is downstream
  in mono/codegen after resolution. NOT bounded — self-host-critical resolver; dedicated session.
- **User free-fn shadowing a BUNDLED stdlib fn** (e.g. `fn split(n: i64)…`) with different arity/arg-types → call resolves to the stdlib one → "expects N argument(s)" REJECT thay vì dùng user fn. Native path bundles toàn bộ stdlib (KHÔNG cần import) nên tên như `split`/`get` va chạm. SAFE (reject, KHÔNG miscompile) — BUG#80 class (đã fix case miscompile-to-0, còn case arity-reject). Fix lý tưởng: user fn arg[0] khác stdlib → shadow. Thấp ưu tiên. **BÀI HỌC probe: LUÔN đổi tên fn khi shrink repro** (dính bẫy này 2026-07-13 với tuple probe w5).

## STRATEGIC RE-PLAN 2026-07-13 (milestone survey — ĐỪNG survey lại mỗi tick)
Compiler ở PLATEAU trưởng thành: self-host native x86-64 OK (qua M1–M5, phần lớn M6). Safe tactical RFC backlog CẠN (tuples/globals/strings đều probe-clean). `docs/tasks/milestones.md` = roadmap dài hạn (Month 30 cho M11). Gate gần nhất chưa đạt:
- **M4 "100 compliance tests"** — `tests/axiom_compliance_suite.ax` (681 dòng) KHÔNG parse: dùng grammar ASPIRATIONAL (import `std.gpu`/`std.quantum`/`std.compiler.ai`/`std.net`; block string `"""…"""`; `.length()` method; `assert`). Pass toàn bộ = milestone-scale, KHÔNG tick-sized.
- **M6** ELF export (RFC 0009 P3) + perf gate (Fib(40) ≤5% clang) + **M6b** Mach-O — đều lớn.
**Candidate tick-sized work (an toàn, ACCEPT-only, tiến tới compliance):** block strings `"""…"""` (lexer, cần chọn raw-vs-dedent); `.length()` alias cho `.len` (design: spec dùng `.len`); `assert(cond)` builtin/std. **Rủi ro/sâu:** Vec-tuple mono (cần deep-debug session). **KL:** loop nên (1) làm feature nhỏ tiến tới compliance HOẶC (2) probe combo mới HOẶC (3) user chỉ định milestone lớn. Big progress cần user hướng hoặc dedicated session.

## ✅ Block strings `"""..."""` — SHIPPED `72e4858` (RFC 0024, A==B `21ED17B1`, 307/307)
Multi-line `"""..."""`: spans newlines, closes at next unescaped `"""`, interior single/double
quotes literal, escapes (`\n \t \" \\`) processed, content verbatim (no dedent). 3 sites exactly
as scoped below: lexer.ax scan_string (block-scan), print_helpers.ax unescape_string_literal
(strip 3), typecheck.ax strip_quotes (strip 3). Oracles t_blockstr/ml/esc/byte/quote. LESSON:
the test-driven gate (build+O0==O1+regression) CATCHES a missed-site miscompile, so "multi-site"
is NOT a reason to defer a NON-self-host-critical frontend feature — over-conservatism was wrong.
DEFERRED sub-features: dedent policy, raw-string variant (RFC 0024 §4). Original scope note:
Compliance-suite (M4) uses `"""multi-line"""`. Implementation is MULTI-SITE (accept-then-miscompile
risk if any site missed → do it with full context, gate carefully):
1. **lexer.ax `scan_string`** (~L285, dispatch L531): on `"`, peek for `""` → block mode: consume
   until closing `"""`, ALLOW newlines (normal strings terminate at `\n` L299). Token text includes
   the `"""…"""`.
2. **`unescape_string_literal`** (print_helpers.ax:24) — called by x86_asm_emitter (L175,732),
   x86_emitter (L232), x86_coff (L431): strips surrounding quotes + processes escapes for the NATIVE
   byte emission. Must strip 3 delimiters each side for a `"""` literal (currently 1).
3. **`strip_quotes`** (typecheck.ax:13) — comptime value; same 3-vs-1 fix.
Design decisions (§20, pick SAFEST minimal): content = VERBATIM bytes between delims with the SAME
escape processing as normal strings (no dedent — dedent = separate design). Leading-newline: keep
verbatim (don't auto-strip) for predictability. ACCEPT-only new syntax → A==B gate. Needs an RFC
(syntax change, §13). Est. bounded but touches ~5 files; verify NATIVE path bytes (not just C).

## M4 compliance suite — gap analysis 2026-07-13j (do NOT re-derive; mostly design-questionable)
`tests/axiom_compliance_suite.ax` (681 lines) written in an ASPIRATIONAL dialect diverging from
AXIOM's real grammar. What now parses: basic decls (let/mut/i32/f64/bool), `assert()` (shipped),
**block strings `"""..."""`** (shipped `72e4858`, test_006). Remaining BLOCKERS (each a decision,
not just "implement"):
- **`=>` match-arm syntax** (`1 => assert(false)`) — AXIOM uses `Pattern:` (parse_match_arm). Adding
  `=>` = REDUNDANT syntax, contradicts design. DON'T add without explicit spec decision.
- **`.length()`** — AXIOM uses `.len` (spec-authoritative). `.length()` = dialect drift. A `.len`
  alias is trivial (method resolution) but redundant; defer as a design call.
- **`impl Trait for Type`** trait-impl syntax (line 274) — AXIOM uses structural/duck method match
  (line 304 comment says Square works WITHOUT `impl`). Trait-impl = bigger design/RFC.
- **Unimplemented stdlib imports**: std.compiler.ai / gpu / quantum / net / concurrency / fs /
  testing.assert — aspirational modules, milestone-scale.
KL: making the suite pass is milestone-scale AND partly contradicts AXIOM design. Block strings was
the only clearly-good tick-sized piece. Rest = user-directed design decisions or big features.

## RFC follow-ups still open
- **RFC 0009 P3** — ELF export (`.edata` for ELF); P1/P2 (COFF import/export) shipped.
- **RFC 0014 drop-glue** — BLOCKED on [[bug69-ctgc-ownership-escape-noop]] (needs escape/ctgc real analysis).
- **RFC 0015 P2 SHIPPED `f06d939`** (2026-07-16) — EscapeAnalyser ACTIVE (crash fixed: was a wild-free segfault mis-logged as "non-determinism"; see [[bug69-ctgc-ownership-escape-noop]]), marks escaping locals w/ `SYM_FLAG_ESCAPES=4096`, inert (A==B `184E35B4`, 327/327). **P3 (CTGC free) STILL OPEN, high-risk** — needs borrow/alias tracking (INDEX/FIELD-init borrow-edge, never-free) + module whitelist + fix ctgc.ax guard/flag, else UAF self-host (frees borrow-locals aliasing AST vectors). Unblocks RFC 0014 drop-glue.

## Note
next-step-15 items RFC 0005 (int-literal inference), RFC 0006 (numeric/arith, BUG#33/34), ADT v1, untyped-self (RFC 0002), enum (RFC 0003) are DONE. The `docs/next-step-1..17` files are historical logs; the LIVE state is MEMORY.md + the newest session-handoff.
