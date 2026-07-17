---
name: rfc0011-p4-separate-compilation
description: "RFC 0011 P4 import-driven auto-libraries (separate compilation): __axiom_iface member design + progress. inc1–inc3c SHIPPED 2026-07-04 — `--auto-lib` find-or-build+cache; import x auto-resolves x.lib end-to-end (opt-in)."
metadata:
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

**RFC 0011 P4 — import-driven automatic libraries (Go/Rust/Zig model).** DECISION chốt
2026-07-04 (commit 9345bfd, RFC §3bis): `import x` sẽ tự tìm `x.lib` (tươi→dùng,
thiếu/stale→biên dịch `x.ax`→`x.lib` tối ưu, đệ quy). `-l` giữ làm escape-hatch cho lib
không-có-source. Cùng một tính năng với "hết recompile stdlib" (stdlib = client đầu).
Cần SEPARATE COMPILATION qua **self-describing `.lib`** mang interface member.

**Cơ chế `.lib` (COFF `!<arch>`):** member code `axiom.o` + member metadata
`__axiom_iface/` (đặt CUỐI để symbol-index offsets không đổi; tên prefix `__` để consumer
`-l` bỏ qua — không phải code). Format text: `AXIFACE1\n` + mỗi `pub fn` một dòng
`F <name> <nparams> <p0> … -> <ret>`, type = token ổn-định (primitive TypeID cố định;
ptr/slice = "ptr"/"slice"; struct/generic = "?" placeholder). Consumer suy ra symbol
linker = `ax_<name>` (mangling đồng nhất).

**Đã SHIP:**
- **inc1** (ec56533): plumbing member `__axiom_iface` + verb `axc iface <x.lib>` in ra;
  consumer loop bỏ qua member prefix `__`; `axiom_read_lib_iface`. Payload v1 = danh sách
  symbol. Round-trip OK, `-l` link vẫn chạy (app_static exit 84).
- **inc2** (ec41356): `build_lib_iface_text(symtable, pool, types)` trong linker.ax duyệt
  symbol table, xuất chữ ký hàm thật. `axiom_write_static_lib` thêm param `iface_text`
  (rỗng→fallback danh-sách-symbol). Dump: `F ax_add 2 i32 i32 -> i32`. Gặp+né BUG#53
  (single-line-if miscompile, xem [[bug53-single-line-if-miscompile]]).
- **inc3a** (e556224): interface READER — `parse_lib_iface(lib_path) -> IfaceFuncVec`
  (đảo của writer), `iface_token_to_type`, `read_lib_iface_bytes`. `axc iface` in raw +
  parsed round-trip (khớp = reader OK).
- **inc3b** (57c733c): **XONG end-to-end** — `import x` tự tìm `x.lib`. main_air
  `register_module_from_lib`: mỗi pub fn → SYM_FUNC body-less (decl_node 0, chỉ
  SYM_FLAG_PUB ⇒ cgen mangle `ax_<name>` khớp lib), bind qualified `x.fn` vào global
  scope, push export; status=LOADED + ast_tree=null ⇒ build_module KHÔNG emit body
  (codegen-skip free), linker cấp body. `ax_driver_load_module` check `<mod>.lib` TRƯỚC
  source path (gated ⇒ không có .lib = hành vi cũ byte-identical). Linker inputs threaded
  từ lazy.modules. **Test:** `import imp_mymath` + imp_mymath.lib → imp_app.exe exit 42
  KHÔNG `-l`; xóa lib → source fallback vẫn 42 (tests/ffi/imp_*).

- **inc3c** (d3d2c7c): **XONG staleness+auto-build** sau cờ opt-in `--auto-lib`. Pre-pass
  `ensure_import_libs` (main_air, sau lift_closures, trước resolve): mỗi non-std import có
  x.ax → nếu x.lib thiếu HOẶC manifest-hash != djb2(x.ax) → `system("<self>" build x.ax -o
  x.lib --staticlib --no-stdlib …)` (self=args[0]; **cmd /c bọc NGOÀI 1 cặp `"`** để tránh
  bug Windows nuốt quote đầu). Unchanged→cache hit; đổi→rebuild. **TẤT CẢ đường import→.lib
  (register-from-lib branch, linker threading, pre-pass) gated trên `lazy.auto_lib`/cờ.**
  **BÀI HỌC:** un-gated auto-CREATE làm t_modcollide FAIL — biến app-module (trùng tên,
  BUG#50) thành lib mất unique-mangling. ⇒ phải opt-in vì AXIOM chưa phân biệt app-module
  vs library. Default = source path (byte-identical). Test: `--auto-lib` build lần đầu tạo
  lib (exit 42), unchanged=cache hit, sửa x*3→x*4 rebuild=56; không cờ→bỏ qua lib on-disk.

- **inc3d** (19de137): **`library <dotted.name>`** — từ khóa contextual đầu file đánh dấu
  module là library biên-dịch-riêng. Module có marker → auto-lib MẶC ĐỊNH (không cần cờ);
  không marker → source path (an toàn cho app-module trùng tên/dùng stdlib). parser nuốt
  `library` (không tạo node); `source_is_library(ax_path)` quét dòng đầu có nghĩa. Các gate
  (loader lib-branch, pre-pass, linker threading) fire khi marker HOẶC --auto-lib. Test:
  module marked, import KHÔNG cờ → tự build lib → exit 52 (13*4). ⟹ **giải quyết** "phân
  biệt app-module vs library" mà user đề xuất (giống `package` của Go).
- **inc3e** (b2234d5): **`--staticlib --shared` cùng lúc → xuất CẢ `<base>.lib` VÀ
  `<base>.dll`** từ 1 compile (base = -o bỏ đuôi .lib/.dll). Consumer chọn: static-link
  .lib (`-l`) hoặc bind DLL theo tên (`extern "C" from "x.dll"`) — AXIOM không cần import-lib
  riêng. DLL export theo `#[export]`; .lib phơi mọi symbol. `strip_lib_dll_ext`; nhánh
  staticlib ghi .lib rồi rơi xuống linker ra .dll. Test axmath → axmath.lib + axmath.dll.

- **inc3f** (04ca70d): build DLL (--shared) → tự sinh **`<base>_ffi.ax`** = `pub extern "C"
  from "<dll>" fn name(p0: T,…) -> T` cho mỗi hàm #[export]. Consumer `import <base>_ffi`
  MỘT file thay vì gõ extern tay từng hàm; gọi `wrapper.fn(...)` → linker tự thêm dll vào
  .idata. `build_dll_wrapper_text` (linker.ax). Test: axmath.dll→axmath_ffi.ax; consumer
  import → exit 84, objdump thấy axmath.dll trong .idata. ⟹ trả lời đúng câu hỏi user.

- **inc3g** (8fea53a): **library-marked DLL export mọi pub** (không chỉ #[export]). main_air:
  `export_all_pub = source_is_library(filename)` → nếu true, thread MỌI pub SYM_FUNC (tid!=0)
  vào `linker.export_names` thay vì export_syms. `build_dll_wrapper_text` thêm cờ `export_all`
  (true → bind mọi pub trong `<base>_ffi.ax`). Unmarked --shared giữ #[export] (fixpoint
  byte-identical). Test tests/ffi/libapi.ax (marked, KHÔNG #[export], 2 pub fn) + libapi_use →
  exit 34; cả lib_add/lib_sub export + bind qua wrapper. ⟹ .dll khớp API .lib.

**CÒN LẠI:**
- **(nâng cao) import x động thẳng DLL không cần wrapper:** register-from-iface nhưng đánh
  dấu symbol là dll-import (dll_bind) thay vì link .lib — `import x` chọn static/dynamic.
- **inc4a (const) SHIPPED (27a65ed, sha=2E254BA3):** `pub const` int qua interface. Producer
  `build_lib_iface_text(+tree)` emit `C <name> <type> <value>` cho mỗi pub SYM_CONST có
  initializer NODE_INT_LIT kiểu primitive-int (đọc value từ token literal). Reader mới
  `IfaceConst`/`parse_lib_iface_consts`. Consumer `register_module_from_lib` đăng ký mỗi const
  = SYM_CONST body-less (decl_node=0), **stash value vào field `next_overload`** (const không
  dùng field này — overload chain chỉ SYM_FUNC). Lowerer `lower_field_expr`: `mod.CONST` với
  decl_node=0 → emit OP_ICONST từ next_overload (path const-nguồn walk-AST giữ nguyên; chỉ
  module-member path cần vá vì lib const luôn qualified). Giới hạn: int ≥0 <2^32 (float/bool/
  struct = lát sau). Test constlib.ax MAX_ITEMS=42 → constuse 42-8=34. ⟹ **khai báo
  non-function ĐẦU TIÊN round-trip qua __axiom_iface**.
- **inc4b (struct) SHIPPED (28a7875, sha=6BE28772):** `pub struct` field-primitive qua interface.
  Producer emit `S <name> <nfields> <fname> <ftype> ...` (chỉ struct field toàn primitive), TRƯỚC
  F/C lines. `iface_type_token(+pool)` trả TÊN struct cho TYPE_KIND_STRUCT ⇒ chữ ký hàm round-trip
  (`F origin 0 -> Point`, `F add_pt 2 Point Point -> i32`). Consumer: **rewrite register_module_from_lib
  parse raw 2-pass** — pass1 đăng ký struct (register_struct trên tt consumer, layout TÍNH LẠI từ
  cùng fields ⇒ size/align/offset khớp), build map tên→type_id + SYM_STRUCT bind `mod.Name`; pass2
  đăng ký hàm, resolve param/ret struct-typed qua map. Helper mới register_lib_symbol + iface_resolve_tok.
  **Struct-by-value ABI (return VÀ param) chạy qua đơn vị biên dịch riêng** — test geolib/geouse
  add_pt(origin(),origin())=14. KEY: offset/size không serialize, tính lại ⇒ khớp tự động. Lưu ý:
  `iface` verb round-trip printer vẫn in "?" cho struct (standalone parse_lib_iface không có map) —
  chỉ cosmetic, path đăng ký thật (register_module_from_lib) struct-aware.
- **inc4c type-alias BỊ CHẶN Ở PARSER (khảo sát 2026-07-04):** `type X = <cái gì>` LUÔN
  parse thành **sum-type** (parse_type_variant), KHÔNG phải alias. `type Word = i32` → sum
  một biến-thể-nullary tên "i32" (dùng làm param → tag-box, KHÁC i32); `type CPtr = ptr[u8]`
  → **LỖI PARSE** (`[` sau `ptr` không nuốt được). std/ffi.ax (`pub type CPtr=ptr[u8]`) KHÔNG
  nằm trong concatenate_stdlib (chỉ result/alloc/scheduler/runtime/os/string/io) ⇒ chưa từng
  compile. ⟹ inc4c cần TRƯỚC: parser hỗ trợ alias thật `type X = <type-expr>` (đụng ngữ nghĩa
  sum-vs-alias, có thể RFC). Alias trong stdlib thực chỉ: CPtr/CStr (ptr) + Option/Result (sum
  generic = inc5). TẠM HOÃN inc4c.
- **BUG#54 (phát hiện+FIX khi khảo sát trên):** `Color.Blue` qualified-variant segfault → fix
  1bb4359. Xem [[bug54-qualified-variant]].
- **inc4 CÒN:** struct field ptr/nested-struct (token mất pointee/struct → field-access typing sai;
  cần topo-order cho nested). type-alias (cần parser work, xem trên). Gỡ giới hạn --no-stdlib cho lib.
  Const: float/bool/negative/64-bit + const=biểu thức (const-fold producer, hiện chỉ INT_LIT).
  **LƯU Ý const:** value stash ở `next_overload:u32` ⇒ i64/negative bị chặn bởi width field +
  immediate OP_ICONST (src1:u32); mở rộng cần side-table hoặc đổi storage.
- **inc5:** generic (template trong interface → monomorphize tại nơi dùng).
- **name collision:** lib-fn trùng tên local-fn chưa xử lý (đường BUG#50 flag-2048).
- Đích cuối: stdlib thành `std.lib` cached → bỏ whole-program `concatenate_stdlib`
  (main_air:343) → hết recompile stdlib mỗi build. Xem [[rfc0011-static-libs]].

**Gate:** đây là backend/linker → BẮT BUỘC self-host fixpoint (n1==n2) + regression
93/93 TRƯỚC commit. inc1 sha=522995B3, inc2=4121B425, inc3a=80C0A12F, inc3b=2E87B4A3,
inc3g=E151098E.
