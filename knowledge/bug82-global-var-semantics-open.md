---
name: bug82-global-var-semantics-open
description: "BUG#82 FIXED (RFC 0017 P1 `9e00177`). Scalar non-const init SHIPPED `33d4ead`. AGGREGATE globals (struct/array/tuple) SHIPPED `6264ff6` (A==B==C `9ce4ff45`, 219/219) via runtime-init+block-copy. STILL OPEN: pointer-repr (Option/Result/sum) globals, .bss, ELF, uninit-decl."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

## ✅ FIXED 2026-07-09 (RFC 0017 P1, commit `9e00177`, pushed `origin/main`=`73be5db`)
ROOT CAUSE was WORSE than first thought: globals had **NO storage at all** — reads returned stale registers, writes were no-ops. The earlier "plain set/get works" was a FALSE positive (coincidental RAX leftover), proven via `dump-air`.

**Fix (RFC 0017 P1, [[backlog-open-items]]):** real static storage in a writable COFF `.data` section (section 3), addressed via new **OP_GLOBAL_ADDR (0x0113)** = `lea rip + RELOC_PC32` (selector vreg=4 marker + emitter), placed by the linker at the end of the writable `.idata` region. Reads = OP_GLOBAL_ADDR+OP_LOAD, writes = OP_GLOBAL_ADDR+OP_STORE. Constant scalar initializers (int/bool/char/float, unary +/-, **casts** `0 as u64`) folded into the data image. Typecheck **Phase 4 gate** rejects non-scalar / non-constant-init globals (BUG#53 convention).
- **Backend change → A!=B, gated by hand-built B==C = `6676E76D9253FE69`** (new daily-driver hash). Regression **120/120** (added `t_globals`=113). Oracles t1..t6 (init=7, RMW=3, set/get=42).
- **Fixpoint-safe** because zero-globals output is byte-identical; the compiler's only global (`g_documents`, lsp.ax) is never touched during `build`.
- **P2 non-const scalar init SHIPPED `33d4ead` (2026-07-11, A==B, 166/166):** `mut g := compute()` / `mut v := f(K)` now accepted. Non-const init → storage zeroed + init AST node stashed (`AirGlobal.init_node`); `lower_func` emits a **prologue in `main`** that evals each and stores (OP_GLOBAL_ADDR+OP_STORE) in source-declaration order before user code. Chose main-prologue over synthetic `__axiom_global_init` (no new call-site/linker wiring). A==B (compiler has no non-const globals). Oracle `t_globalinit`(210). Forward-ref reads pre-init .data value (declaration order = defined semantics).
- **✅ P2 AGGREGATE globals SHIPPED `6264ff6` (2026-07-12, A==B==C `9ce4ff45`, 219/219)** — struct/array/tuple module-level globals. Approach = **runtime-init + block-copy** (NOT the reverted const-blob approach below): (1) typecheck accept STRUCT/ARRAY/TUPLE; (2) `collect_global` slot=full size + always runtime-init (init_node, never fold); (3) `lower_global_read` returns `OP_GLOBAL_ADDR` **typed as the aggregate** (no LOAD) → get_register_type reports struct type → OP_GET_FIELD/INDEX/SET_FIELD/STORE compose base+offset like a local aggregate. `lower_global_write` UNCHANGED (selector OP_STORE block-copies when type_id is aggregate). **Why this succeeded where (k) failed:** (k) tried const-folded blob + returned addr typed as pointer(4) → OP_INDEX mis-based at sym+(size-esz); typing the OP_GLOBAL_ADDR result as the aggregate + runtime-init sidesteps both. Oracles t_globstruct(30)/globarr(80)/globnested(40)/globarriter(119)/globbig(65)/globpass(42). **Fixpoint-safe:** compiler has no aggregate globals → new path never exercised self-compiling.
- **✅ P2 POINTER-REPR globals SHIPPED `3a44577` (2026-07-12, A==B==C `dc6a18a5`, 236/236)** — module-level `Option[T]`/`Result[T,E]`/user-sum globals. Their VALUE is an 8-byte tagged box pointer (entry.size 16 = box, NOT slot), so slot=8B + store/load the pointer directly (NOT block-copy). Runtime-init at main-prologue heap-allocates the box + stores its pointer. **3 pieces:** (1) `typecheck check_module_global` accept sum/Option/Result kinds; `Option[T]`/`Result[T,E]` **annotations** are left UNRESOLVED (type_id 0) by the general inference path — retyping them globally shifts RFC 0012 str-box EQ/COPY/MAKE_REF and breaks self-compilation — so resolve via register_option/result **scoped to module-global decls only** (compiler has no such globals → fixpoint-safe), pin symbol type_id. (2) `collect_global` pointer-repr → `size=8` (override entry.size 16) + runtime-init. (3) `x86_selector` OP_LOAD forces 8B load for pointer-repr dest; OP_STORE plain path (src2==0) adds `store_is_ptr_sum` 8B pointer store (mirror BUG#78 array/field). Oracles t_globopt(42)/t_globresult(42)/t_globsum(42), all -O0==-O1. **BẪY probe:** build test programs từ **repo ROOT WITHOUT `-self-link`** (`-self-link` is compiler-self-build only → segfault; imports resolve theo CWD → build from root); run needs ax_runtime.dll in cwd.
- **✅ P2 STR/BYTES (16B-inline) globals SHIPPED `288c86a` (2026-07-12, A==B==C `81522e76`, 238/238)** — `mut g: str = "..."`. str/bytes are 16-byte INLINE primitives ({ptr,len}); slot=16B, reads/writes move BOTH 8B halves inline (NOT scalar-8, NOT box-ptr, NOT by-address block-copy). 3 pieces: typecheck accept 16B PRIMITIVE; `collect_global` `is_inline16` → 16B slot + runtime-init; `x86_selector` OP_LOAD 16B-non-aggregate dest → two-8B-halves inline load into dest home (single 16B MACH_LOAD into GP reg INVALID — mirror OP_GET_FIELD str size==16). OP_STORE unchanged (emit_block_copy LEAs 16B inline src). Oracles t_globstr/t_globstridx(42). **RFC 0017 storage COMPLETE: every value category (scalar const/non-const, aggregate, pointer-repr, 16B-inline str/bytes).**
- **P2 STILL deferred (low value):** uninitialized decl (`mut g: S` no-init = parse error, decl needs init), `.bss`, ELF `.data`, dedicated `.data` PE section. See `rfcs/0017-global-variable-storage.md`.

### ⚠️ P2 array-globals ATTEMPT 2026-07-10 (k) — REVERTED (tree clean at `62c0619`), bug NOT solved
Attempted const-init fixed-array globals (`let tbl:[i32;3]=[10,20,30]`). Wired end-to-end: `AirGlobal` + `init_bytes:ptr[u8]`+`align` (air.ax); typecheck gate allows array-of-scalar + const array-lit (`ax_global_init_is_const` accepts NODE_ARRAY_LIT); `collect_global` folds the literal into a byte blob (elem i at i*esz); `lower_global_read` returns the OP_GLOBAL_ADDR **address** (no LOAD) for aggregates; x86_coff emits the blob aligned by `g.align`. Scalar path stays byte-identical (init_bytes=null), struct globals still rejected.
**BLOCKER (unsolved):** array reads are wrong for N>1. `[i32;N]`: `tbl[0]` reads the element at offset `(N-1)*esz` (the LAST element), `tbl[1..]` read past-end zeros; **N==1 works** (tbl[0] correct). `[i64;3]` reads garbage (188, not a clean offset) → smells like the larger blob overruns/misplaces in the linker's writable-`.idata` global region. The blob bytes + COFF section-3 symbol (offset 0) + `OP_GLOBAL_ADDR` (type-agnostic reloc) all look correct on paper; `OP_INDEX` base handling and `linker.ax:2346` global-region append also look right. **Needs actual disassembly / memory dump** of the emitted `.data` VA + the `lea rip+reloc` displacement to find where the base lands at `sym+ (size-esz)`. Not a quick fix — do NOT retry without a disasm harness. Scalar globals + all committed work unaffected.

---
### (Lịch sử — mô tả lúc còn OPEN)
**BUG#82 — phát hiện 2026-07-08** qua proactive probing (batch #3, sau khi ship BUG#80/#81).

## Triệu chứng (module-level global `let`/`mut` var, KHÔNG phải `const`)
- **Initializer non-zero KHÔNG chạy**: `mut g := 7` rồi `return g` → **0** (không phải 7); `let g = 9` → 0. Global mặc định 0 (BSS-like), initializer expression không bao giờ được emit/chạy trước main.
- **Cross-function read-modify-write SAI**: `mut counter := 0`; `fn bump()->i32: counter=counter+1; return counter`; gọi `bump()` 2 lần rồi cộng → **4** (đáng lẽ 1+2=3).
- **Plain set/get u64 cross-function ĐÚNG**: `mut g:u64=0`; `setit(x): g=x`; `getit(): return g`; `setit(42); getit()` → **42** ✓.

## Vì sao compiler tự-host vẫn chạy
Compiler + std DÙNG `mut` globals: `lsp.ax:15 mut g_documents: u64 = 0`, `std/scheduler.ax:444 mut g_supervisor_table: u64 = 0`, `std/scheduler_test.ax:98`. TẤT CẢ init `= 0` (né lỗi initializer) + pattern set/get đơn giản (né lỗi RMW). Nên self-build không lộ bug. `pub const` (dùng khắp air.ax) hoạt động vì inline compile-time tại call site (SYM_CONST path trong air_builder), KHÔNG phải storage runtime.

## Vì sao CHƯA fix
1. **Không reject được** như [[bug81-multifield-variant-payload]]: compiler tự phụ thuộc `mut` globals → reject vỡ self-build.
2. **Fix đúng = feature runtime sâu**: cần (a) static storage cho global (có phần rồi — set/get chạy), (b) **init sequence** chạy initializers trước main (chưa có; .init_array/startup), (c) sửa RMW lowering cross-function. Đụng chính đường globals mà compiler đang dùng → **rủi ro fixpoint cao**. Xứng đáng RFC riêng về global-variable semantics + đo lại pattern nào compiler thực sự dựa vào.
3. Không phải blocker: workaround = `const` cho hằng compile-time; tránh global mutable (đúng tinh thần CLAUDE.md §11 "avoid global mutable state").

## Next khi làm
- Viết RFC global-var: storage model + init ordering + RMW correctness.
- Oracle cần: init non-zero, cross-fn RMW (counter), nhiều global, generic/aggregate global.
- Gate: fixpoint BẮT BUỘC (đụng lowering globals mà compiler dùng) + regression + build lsp.ax/scheduler.ax vẫn đúng.
- Liên quan silent-miscompile family [[bug80-free-call-overload-collision]], [[bug81-multifield-variant-payload]].

## Follow-up nhỏ khác cùng phiên (q5): multi-arg free-function overload
Fix BUG#80 (`resolve_free_call_overload`) disambiguate free-call overload chỉ theo **arg[0]** (mirror `resolve_method_overload` chỉ theo receiver). Overload khác nhau ở arg thứ 2+ nhưng arg[0] trùng (vd `combine(i32,i32)` vs `combine(i32,i64)`) → vẫn chọn head, SAI. Pre-existing limitation, không phải regression. Fix đầy đủ = so khớp TOÀN BỘ arg-list khi chọn overload.
