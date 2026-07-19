---
name: session-handoff-2026-07-09
description: "HANDOFF 2026-07-09 → next session. State: tree clean at BUG#88 fixpoint 0D672CC8; 3 local commits UNPUSHED (git credential expired). Shipped BUG#83-88 + for-collection reject. BUG#86 short-circuit blocked on RFC 0016 P2' (CFG-aware liveness)."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

# 🟢 STATE 2026-07-19b (autopilot) — READ FIRST
HEAD=**`5a22500`** (pushed to origin/main), daily driver `bin/axc_native.exe` = **`af3ba8e3`**
(shift-in-loop fix on top of CTGC + container free-glue), **439/439**, ctgc_free_check **12/12**.
Tree CLEAN (except pre-existing untracked `scratch/self_linked_concatenated.ax` artifact +
user's `.claude/settings.json`). ⚠️ Rebuild the daily driver first thing next session:
`& scripts/build_native.ps1` (bin/*.exe gitignored — regenerated; expect fixpoint `af3ba8e3`).
⭐ **This session (2026-07-19b): FIXED a load-bearing variable-shift-in-loop codegen miscompile
`5a22500`** — `x << k`/`x >> k` with a loop-induction-variable count computed `x << 0` (count
frozen at loop-entry value); root = `const_shift_amount` missing the non-SSA def-count guard its
twin `const_divisor_pow2` already had. Backend change, B==C=`af3ba8e3`, A!=B (compiler's own
source hit it). Surfaced by autopilot bug-probing (hand-oracle caught an O-level-consistent wrong
answer, invisible to O0-vs-O1). Oracle `t_shiftloop`(63). Full detail [[bug-variable-shift-in-loop]].
Backlog unchanged below (all dedicated-session scale). Probe batch also cleared 7 novel crosses
CLEAN (HashMap-valued-Vec, mixed-width-sum-payload, HOF map→filter→fold chain, mutual-recursion
struct return, nested Option[Result], array-of-struct index+field mutation) — bank if useful.
Session cleared the two remaining OPEN bugs + the inflight CTGC P3, then COMPLETED the CTGC free story:
0. ⭐⭐ **RFC 0027 CONTAINER FREE-GLUE SHIPPED `10eceb6`** — closes the P3-activation container leak.
   `air_builder.ax::emit_container_buffer_frees` frees a non-escaping `Vec`/`HashMap`/`HashSet` local's
   owned buffers (data/keys/values/hashes/occupied) before the header OP_DESTROY, under opt-in
   `-ctgc-free`. Byte-identical self-host (A==B `9A178747`), 436/436, sweep clean (0 crashes, only
   intended t_drop diff), 8 aliasing/escape probes correct (escaping/aliased containers spared → no
   UAF). [[ctgc-p3-scoping-2026-07-18]] RFC 0027 path D (field-ownership annotations, general/user
   containers) = future.
1. ⭐ **3+ hash-container teardown SIGSEGV — FIXED `cebea3f`** (self-link early-exit mitigation:
   skip the teardown free-chain on a successful `-self-link` build; OS reclaims memory). **0 crashes
   in 160 stress runs** (baseline ~23%), fixpoint `84D204E8`, 435/435. Root cause not fully pinned
   (needs symbolized ASAN) but RULED OUT the two leading hypotheses via a size-classed **AX_CANARY**
   allocator built into `std/mem/alloc.ax` (committed **disabled**, `AX_CANARY=false`): NOT a
   contiguous ≤16B adjacent overflow (+16 slack only halved it; block-end footer sweep clean) and NOT
   free-list-link corruption (per-pop guard + free-list integrity sweep never tripped). Remaining
   suspect = an OOB/indexed write clobbering a POINTER field in a compiler struct. Flip AX_CANARY=true
   to resume the hunt. [[bug-3hashmap-mono-teardown-crash]] Discovery: the running Windows compiler's
   `ax_alloc` is bundled from `std/mem/alloc.ax` (NOT the C `ax_runtime.dll` — that only provides
   global-state/panic/string helpers).
2. ⭐⭐⭐⭐ **CTGC P3 general-free ACTIVATED & SHIPPED `18db268`** — `lower_destroy` else-branch frees
   non-escaping owned NON-drop locals with a plain `OP_DESTROY`. Behind the **opt-in `-ctgc-free`**
   flag → default builds byte-identical. The 2026-07-18 "container free-glue crash" blocker was the
   teardown flake above (now fixed) — DEBUNKED (t_hashi64 `-ctgc-free` 0/30 crashes). Gate GREEN:
   A==B `40BC8158` (inert self-host, freeable=0), 435/435, ctgc_free_check 10/10, broad 455-program
   `-ctgc-free` sweep (`scratch/ctgc_sweep.sh`): 0 flag-crashes, 1 intended diff (t_drop 0→42).
   (The container inner-heap-leak caveat noted here was RESOLVED by item 0 above.)
   [[ctgc-p3-scoping-2026-07-18]]
3. **m2b arity-match shadow — DEFERRED (documented `b97b1b3`)**: dropping the arity-spare RE-CONFIRMED
   to break self-build (compiler A rejects its own `let len = std.string.len(s)` sites → fixpoint RED;
   an A-only build misleadingly succeeds — ALWAYS gate reject-tightening with fast_fixpoint). Real fix
   = resolver (a qualified module path must not resolve to a same-named local); HIGH risk/LOW reward.
   [[bug-malformed-input-robustness-cluster]]
**Gate cmd** unchanged (see 2026-07-18 STATE below). **Backlog now (all dedicated-session scale):**
RFC 0027 **path D** (field-ownership annotations → general free for USER containers, `rfcs/0027`),
**m2b resolver fix** (qualified path must not resolve to a same-named local), true **root-cause** of
the 3-hashmap corruption (flip `AX_CANARY=true` in std/mem/alloc.ax + WSL/valgrind), and the standing
large items: **async/spawn-await**, **macOS/Mach-O**, **M6 perf**. Nothing tick-sized remains
(probing this session found 0 real bugs — mature plateau). Pick one per user priority.

---

# HANDOFF phiên 2026-07-09 (đọc đầu tiên khi vào phiên mới)

## ✅ CẬP NHẬT 2026-07-18 (autopilot) — M4 COMPLIANCE DỨT ĐIỂM + hướng do user chốt
User chốt 3 hướng chiến lược ([[autopilot-direction-2026-07-18]]) rồi trao quyền tự chủ:
(1) **M4 = rewrite suite bằng grammar thật**; (2) milestone kế = **M6 perf**; (3) **RFC 0015 P3
CTGC-free = attempt có gate chặt** (session riêng, B==C + revert-on-red).
- **M4 SHIPPED**: `bin/t_compliance.ax` — 60 test grammar THẬT (groups 1–6: primitives / control-flow
  / functions+lambdas+tuples / structs+methods / generics+collections / sum-types+errors), mỗi test
  assert giá trị đúng, **exit code == số test pass = 60**. Build sạch trên daily driver, chạy 60.
  Gated `t_compliance|exit|60`. **Regression 401/401 (was 400) GREEN.** Test-only (KHÔNG đụng
  compiler/stdlib source) → ko cần fixpoint. Aspirational suite giữ ở
  `tests/axiom_compliance_suite_aspirational.ax`; `tests/axiom_compliance_suite.ax` = pointer doc.
  Dialect gaps SIDESTEPPED: local const→module-const, match-expr→if-expr, `=>`→`Pattern:`,
  interface/impl→struct/duck methods, capture→zero-capture lambda. Chi tiết [[m4-compliance-suite-spec-vs-impl-gap]].
- **M6 GROUNDWORK SHIPPED** (`8d05f96`): benchmark harness `scripts/bench_perf.sh` +
  `benchmarks/{fib,collatz}.{ax,c}` + **RFC 0026**. Baseline daily driver: fib **2.72x**, collatz
  **3.19x** vs clang -O2. Inliner FULLY SPECed + staged (`1095b4d`, [[m6-perf-gate-fib-benchmark]]
  "INLINER IMPLEMENTATION SPEC"). ⚠️ Safe pure-arith inliner = perf-NEUTRAL (fib recursive/collatz
  multi-block both skipped) → next dedicated B==C session = **accumulator-recursion→loop (fib)** or
  **control-flow inliner (collatz)**. Insertion `ssa_opt.ax:1798`.
- **COMPLIANCE PART 2** (`d77813b`) + **cross-opt guard** (`94e1379`): `bin/t_compliance2.ax` = 28
  more real-grammar tests; both compliance suites now gated at -O2/-O3 too. Combined M4 surface = **88 tests**.
- ✅ **RFC 0026 P1 INLINER SHIPPED** (`d64a68d`): pure-scalar single-block inliner in `ssa_opt.ax`
  (`inline_func`, level≥1). **B==C bit-identical `aae2ea1f`, regression 407/407**, daily driver
  promoted (fixpoint now `aae2ea1f`). Win: `benchmarks/hotloop` **185→72ms = 2.57x**; fib/collatz
  unchanged (skipped). 2 bring-up bugs fixed (AirInst aliasing via `mut nn:=cin`; params.data=TYPE
  ids not vregs). Oracle `t_inline`. Details [[m6-perf-gate-fib-benchmark]]. `origin/main`=**`d64a68d`**.
- ✅ **RFC 0026 P1.5 GETTER INLINING SHIPPED** (`67a0bbe`): scalar-field getters (`fn get_x(p)=return p.x`)
  now inline. `SsaOptimizer.run`/`inline_func` take the `TypeTable` (threaded from main_air.ax:1085);
  `OP_GET_FIELD` whitelisted but gated to a scalar field of a param (field_size≤8, not aggregate/
  pointer-sum). **B==C `f286cac9`, 412/412**, daily driver promoted. Win: `benchmarks/getter`
  **185→64ms = 2.89x**. Also covers computed multi-field accessors (`area=w*h`). **7 inliner oracles**
  banked (`t_inline`..`t_inline7`: pure-arith, edge cases, generics/loop/branch crosses, scalar/
  multi-field getters, negative safety-gate for 16B/pointer-sum fields, mixed scalar/aggregate struct).
  Multiple probe batches CLEAN (no inliner miscompile).
- ✅ **OP_INDEX SHIPPED** (`90ca7e4`): scalar array-element getter inlining (`at(a,i)=a[i]`), gated to
  resolved scalar element (type_id!=0, not aggregate, size≤8). **B==C `343fa03b`, 418/418**. Oracle
  `t_inline8`. **GETTER FAMILY COMPLETE** (pure-arith + field + multi-field + array-element; 8 oracles
  `t_inline`..`t_inline8`). Full suite **green**. `origin/main`=**`90ca7e4`**.
- **NEXT (dedicated-session backlog)**: (1) RFC 0026 P2 — control-flow inliner (catches collatz) +
  accumulator-recursion→loop (fib), higher bench leverage; also multi-field/aggregate getters +
  `OP_INDEX` (array-element getters) as further P1.x; (2) RFC 0015 P3 CTGC-free (high-risk,
  attempt-with-tight-gate per user). All need full B==C + -O2 regression + revert-on-red.

## ✅ CẬP NHẬT 2026-07-12 — RFC 0017 GLOBAL STORAGE HOÀN CHỈNH + probing
`origin/main`=`ceceb7e`, 243/243, daily driver `bin/axc_native.exe` **A==B==C `81522e76`** (đổi ở `288c86a`=backend; sau đó chỉ test-only).
**RFC 0017 storage COMPLETE mọi value category** (scalar const/non-const, aggregate struct/array, pointer-repr Option/Result/sum, 16B-inline str/bytes). Commit chain:
- `288c86a` = **STR/BYTES (16B-inline) globals**: `mut g: str="…"` → 16B `.data` slot, runtime-init, đọc/ghi CẢ HAI nửa 8B inline. 3 mảnh: typecheck accept 16B PRIMITIVE; `collect_global` `is_inline16`→16B slot+runtime-init; `x86_selector` OP_LOAD 16B-non-agg dest→two-8B-halves inline load vào dest home (single 16B MACH_LOAD vào GP reg INVALID — mirror OP_GET_FIELD str). OP_STORE ko đổi (emit_block_copy LEA 16B inline src). Oracle t_globstr/t_globstridx(42).
- `3a44577` = **POINTER-REPR globals** (Option/Result/user-sum). Value=8B tagged box ptr → slot=8B store/load pointer TRỰC TIẾP (KHÔNG block-copy). 3 mảnh: (1) `typecheck check_module_global` accept sum/opt/res; Option/Result **annotation** UNRESOLVED bởi general inference → resolve register_option/result **scoped module-global-decl only** (compiler ko có global loại này → fixpoint-safe), pin symbol type_id; (2) `collect_global` ptr-repr `size=8` override + runtime-init; (3) `x86_selector` OP_LOAD force 8B + OP_STORE plain-path (src2==0) `store_is_ptr_sum` 8B ptr store (mirror BUG#78). Oracle t_globopt/globresult/globsum.
- `6264ff6` = **aggregate globals** (struct/array). Tái dùng by-address machinery: typecheck accept STRUCT/ARRAY/TUPLE; `collect_global` slot=full size+runtime-init block-copy; `lower_global_read` aggregate trả OP_GLOBAL_ADDR (địa chỉ, KHÔNG load) → GET_FIELD/INDEX compose như local. `lower_global_write` ko đổi.
- `f5ef298` = parser no-init var-decl fix (`check_raw` thay `check(TK_EQ)`→peek nuốt newline); no-init global → zeroed `.data`.
- `5d49d9d`/`1a66360` = bank 5/6 aggregate-global write+interaction oracle (O0==O1).
- `ceceb7e` = **bank 5 DEEP-CROSS oracle (test-only, binary ko đổi)**: 19 probe/4 batch (globals × ctrl-flow/generics/methods/str-16B/ptr-repr/init-order) = 0 miscompile. Bank: t_globarrstr([str;3]), t_globoptstr(Option[str] payload=str), t_globarrstructstr([{i64,str};2]), t_globarropt([Option;3] None-store), t_globinitorder(cross-global init top-to-bottom).

**2 finding KHÔNG-bug (defer):** (a) **tuple-literal expr `(a,b)` KHÔNG parse** — parser chỉ có `NODE_TUPLE_PAT` (pattern), KHÔNG tuple-EXPR node → tuple globals unconstructible (typecheck-accept-list liệt kê TUPLE nhưng KHÔNG có oracle nào). Grammar gap → cần RFC §13. (b) aggregate `return g` = REFERENCE alias (RFC 0001 §5), KHÔNG copy — đúng design.
**BẪY probe:** build từ repo ROOT WITHOUT `-self-link` (import resolve theo CWD; `-self-link`=compiler-self-build → segfault giả); self-build có thể OOM → `rm -f bin/axc_fp*.exe` trước mỗi hop. RFC 0015 P2 escape OPEN (blocked, reverted `cff2552`).

## ✅ CẬP NHẬT 2026-07-09 (phiên sau) — BUG#86 ĐÓNG
- **Tất cả ĐÃ PUSH.** `origin/main` = `755d7b8`.
- **RFC 0016 P2' (CFG-aware liveness) SHIP** = `e3f9539`. Chi tiết [[rfc0016-p2prime-cfg-liveness]].
- **RFC 0016 P3 (short-circuit and/or) SHIP** = `755d7b8` → **đóng BUG#86** [[bug86-short-circuit-open]]. 2 bug: (1) lowering diamond, (2) `lower_while` CFG-edge (bug làm B hang: -O0 pass/-O1 hang → DCE `remove_unreachable_blocks` NOP `ret` exit).
- **Daily driver `bin/axc_native.exe` hash `c777ef7b`** (thay `BCEFC38F`←`0D672CC8`).

## Trạng thái tree (SẠCH)
- HEAD local = `755d7b8` = `origin/main`. Tree clean (chỉ `scratch/self_linked_concatenated.ax` + `.claude/` pre-existing, KHÔNG đụng).
- **`bin/axc_native.exe` fixpoint `c777ef7b`.** Regression baseline **114 tests + 5 sc oracle** (sc1/sc2/scv/scw/scstress). t_math=127.
- Gate nhanh: `scripts/fast_fixpoint.ps1` (A==B, hash = `c777ef7b`). Regression: `AXC=bin/axc_native.exe bash scripts/regression_repros.sh`; spot-check nhanh `bin/axc_native.exe build bin/<t>.ax -o /tmp/x -O1; /tmp/x; echo $?`.

## Đã SHIP (đã gate)
BUG#83/84/85/87/88 · **RFC 0016 P2' (CFG liveness) + P3 (BUG#86 short-circuit)**. Tất cả pushed.

## OPEN — ưu tiên cao
- **BUG#82 globals** ([[bug82-global-var-semantics-open]]): module-level `let`/`mut` init non-zero không chạy + cross-fn RMW sai. Cần RFC riêng.
- `for x in <collection>` iteration thật (iterator protocol) — hiện REJECT sạch; feature tương lai.

## Tài liệu chính
- `rfcs/0016-cfg-aware-liveness-block-ordering.md` — thiết kế fix nền cho BUG#86 (P1 terminator-norm, P2 RPO=insufficient, **P2' CFG-aware liveness=fix thật**, P3 short-circuit). ĐỌC KỸ trước khi làm BUG#86.
- Design: [[axiom-struct-reference-semantics]] — struct = REFERENCE semantics (RFC 0001 §5), `mut cpy:=src` alias KHÔNG copy; ĐỪNG "fix" (suýt phá self-host).

## Bài học phiên này
- Proactive probing (batch feature-combo, so exit code) tìm được BUG#83-88; yield về 0 sau ~18 batch (compiler rất chắc).
- Exit code bash 8-bit: giá trị >255 bị truncate (vd 921→153); dùng PowerShell `$LASTEXITCODE` cho full 32-bit.
- CHECK SPEC trước khi "fix" hành vi lạ (struct alias là design, không phải bug).
- Backend đổi self-codegen → A!=B là transition ĐÚNG, gate = B==C tay (build C từ B). Frontend-thuần → A==B.
