---
name: bug67-option-struct-payload-unwrap
description: "BUG#67 HOÀN TOÀN FIXED (651d13c + c2a2a15 + 3d3c615): Option[T]/Result[T,E] với T là STRUCT/aggregate bất kỳ (Some(x)/Ok(x) top-level VÀ gọi từ bên trong hàm generic khác như HashMap[K,Vec[V]].get) giờ .unwrap() đúng, hết segfault/field-corruption. 3 lớp fix: (1) try_instantiate_variant_call cho Some(x) constructor call, (2) SumInfo.generic_args cho sum-type instantiation qua type-annotation, (3) fix bug scope thật sự (cloned_kind đọc ngoài scope → luôn garbage → (2) chưa bao giờ chạy) mới thật sự kích hoạt (2). Regression 97/97."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#67 OPEN — phát hiện qua proactive-probing sau khi ship BUG#65/66 (RFC 0013, 4fb74f4), 2026-07-06. Root cause đã truy CHÍNH XÁC bằng objdump ground-truth + 1 debug print xác nhận trực tiếp, nhưng CHƯA FIX (thuộc loại kiến trúc, cần cân nhắc kỹ trước khi patch).**

## Triệu chứng (CONFIRMED, nhiều biến thể)

```
struct Big: a: i32, b: i32, c: i32, d: i32
let big = Big(a: 7, b: 12, c: 15, d: 6)
let opt = Some(big)
opt.unwrap().a + ... .b + ... .c + ... .d   // SEGFAULT
```
Repro tối giản nhất, **KHÔNG cần std.collections/Vec** — bất kỳ struct thường nào cũng lộ bug. Lưu ở `scratch/bug67_opt_struct.ax` (session-local, không commit).

Cũng lộ qua `std.collections.Vec[i64]` làm payload (`Some(v)` với `v: Vec[i64]`, 24 byte) — segfault y hệt. Và `HashMap[i64, Vec[i64]]` (nested collection) cũng segfault vì bên trong dùng đúng cơ chế này.

**Quan trọng: KHÔNG liên quan RFC 0013/BUG#65** — test dùng struct THƯỜNG (`Big`, không generic-lồng-generic gì cả) vẫn segfault y hệt, nên đây là bug HOÀN TOÀN KHÁC, chỉ tình cờ phát hiện ngay sau khi ship RFC 0013 nhờ tiếp tục proactive-probe cùng họ generic/collections.

## Vì sao cụm test hiện có (95/95 xanh) không bắt được bug này

`match opt: Some(x) => ... | None => ...` (NATIVE match, cơ chế BUG#57/58/59) hoạt động ĐÚNG — vì `match` được air_builder xử lý HOÀN TOÀN RIÊNG (dispatch trực tiếp trên tag/box pointer ở codegen), KHÔNG đi qua std library's `.unwrap()`/`.is_some()` (những hàm generic THẬT trong `std/result.ax`, gọi qua cơ chế generic-function-call bình thường của typecheck). Toàn bộ test suite hiện có dùng `match` để kiểm Option/Result, KHÔNG dùng `.unwrap()` ở TOP-LEVEL (ngoài 1 hàm generic khác) trên payload STRUCT — nên gap này chưa từng bị lộ.

## Root cause THẬT SỰ (đã CONFIRM bằng debug print `[BUG67-DBG]`, không phải đoán)

1. `pub type Option[T] = Some(T) | None` (std/result.ax) là 1 SUM TYPE generic. `pre_infer_type_alias` (typecheck.ax ~1014-1068) đăng ký sum type NÀY MỘT LẦN DUY NHẤT qua `register_sum_type`, với payload_type của variant "Some" = **T (generic placeholder type_id)**, và gán `v_sym.type_id = type_id` (type_id của SUM TEMPLATE) cho CHÍNH symbol biến thể "Some".
2. Khi typecheck 1 lời gọi **KHÔNG ĐỦ ĐIỀU KIỆN** (`Some(big)`, constructor KHÔNG qualified, tức không phải `Option.Some(big)`) — NODE_CALL_EXPR's handler (typecheck.ax dòng ~1688-1701): `callee_type = infer_node(callee)` trả về `sym.type_id` của "Some" = TYPE TEMPLATE (kind=6 SUM, generic, KHÔNG PHẢI instantiation cụ thể). Dòng 1698-1699:
   ```
   elif entry.kind == TYPE_KIND_SUM:
       result_type = callee_type       // <-- BUG: gán THẲNG type TEMPLATE, KHÔNG instantiate với T=Big
   ```
   Không có bước "instantiate sum type với concrete type-arg" nào tương tự như generic STRUCT (`mono.instantiate_function`, RFC 0013 vừa fix) hay generic HÀM (cùng file, generic-call machinery dòng 1478+). `opt`'s type_id **VẪN LÀ TEMPLATE**, không mang thông tin "T=Big" ở đâu cả.
3. Khi gọi `opt.unwrap()` (một hàm GENERIC THẬT `unwrap[T](self: Option[T]) -> T` trong std/result.ax) — is_generic_call machinery (dòng 1478+) chạy `infer_generic_type_args("Option[T]" AST, arg_types[0]=opt's type_id, ...)`, mà `infer_generic_type_args`'s "user-defined generic template" branch gọi `get_generic_args(opt's type_id)`. Vì `opt`'s type_id là TEMPLATE thô (không GENERIC_INST, không STRUCT có `generic_args` — xem RFC 0013), `get_generic_args` rơi vào `get_generic_args_from_mangled` — parse tên KHÔNG CÓ type-arg gì để parse (tên template thuần "Option", không "Option__Big") → trả về VECTOR RỖNG.
4. `inferred[T-index]` VẪN LÀ `TYPE_UNKNOWN` sau vòng lặp → dòng 1605-1607 (default fallback, giống hệt cơ chế BUG#66!) đặt `inferred[T] = TYPE_I32`. `unwrap[T]` bị monomorphize với **T=i32**, KHÔNG PHẢI T=Big.
5. Xác nhận bằng debug print trực tiếp (`ax_printf_local`, đã revert sạch sau khi confirm): với test `bug67_opt_struct.ax`, dòng debug in ra `tmpl=unwrap rec_type=51 inferred0=0` (inferred0=0=TYPE_UNKNOWN, TRƯỚC khi default hoá thành i32) — xác nhận CHÍNH XÁC cơ chế trên, không phải suy đoán.
6. Hệ quả runtime (xem objdump `bin/probe_opt_struct.exe`, đã xoá sau khi trace xong): field access `got.a/.b/.c/.d` (compiler tưởng `got: i32`, KHÔNG BIẾT layout của `Big`) đều compile thành `QWORD PTR [rax]` LẶP LẠI CÙNG OFFSET 0 cho MỌI field (không tăng offset field nào cả) — vì compiler hoàn toàn mất field-layout info của `got` (tưởng nó là scalar). Dẫn tới đọc rác/sai địa chỉ → segfault.

## Vì sao SCALAR T (i32/i64/f64/str) "work" — hiện tượng che dấu

Với T scalar, `unwrap()` bị monomorphize sai (T=i32 default) NHƯNG vẫn "âm thầm hoạt động" trong nhiều trường hợp vì: (a) hầu hết lời gọi `.unwrap()` trong CHÍNH compiler/stdlib xảy ra BÊN TRONG một hàm generic khác (T flow tự nhiên từ enclosing generic scope, không cần qua `get_generic_args` — khác hẳn kịch bản top-level `main()` của bug này), và (b) khi THẬT SỰ ở top-level với T scalar nhỏ (i32/i64 cùng kích thước sổ registry-based), sai lệch có thể không đủ để gây segfault ngay (dù giá trị vẫn có thể sai/truncate — CHƯA kiểm chứng riêng biệt, nghi ngờ nhưng chưa test). Đây LÀ LÝ DO test suite 95/95 hiện có không bắt được — không phải vì bug hẹp, mà vì cách test hiện tại (`match`, không `.unwrap()` top-level trên struct) tình cờ né được.

## ĐÃ FIX (651d13c, 2026-07-06) — case chính

`typecheck.ax` thêm hàm `try_instantiate_variant_call(callee, sum_type_id)`, gọi ngay tại nhánh `TYPE_KIND_SUM` của `NODE_CALL_EXPR` (thay vì `result_type = callee_type` thẳng). Cơ chế:
1. Callee phải resolve về 1 `SYM_VARIANT` (vd "Some"), và sum type đó phải GENERIC (`FLAG_IS_GENERIC` trên alias's decl_node — lấy alias bằng `self.symtable.resolve(entry.name_id)`, entry.name_id chính là tên alias vì `register_sum_type` đăng ký bằng `sym.name_id` của alias).
2. Lấy danh sách gen_params CỦA ALIAS theo TÊN (đọc `NODE_GENERIC_PARAMS` con của alias's decl_node) — KHÔNG cần type_id, chỉ cần chuỗi tên ("T", "E").
3. Duyệt SONG SONG các payload type-expr của variant CỤ THỂ đang gọi (vd "Some(T)" → 1 node "T") với các argument THẬT của lời gọi (vd `big`), khớp TÊN payload-expr với TÊN gen_param để biết VỊ TRÍ cần bind, rồi gán `args[vị_trí] = infer_node(argument_thật)` — **chỉ nếu** kết quả KHÔNG PHẢI generic placeholder (`not self.is_generic(concrete)`, guard thêm sau lần test đầu phát hiện thiếu — xem residual gap).
4. Nếu bind được ÍT NHẤT 1 param: `self.types.register_generic_inst(entry.name_id, args)` thay cho type template thô. Param không bind được (vd E khi gọi `Ok(x)`) giữ TYPE_UNKNOWN — CHƯA giải quyết (không phải regression, là giới hạn suy luận cục bộ vốn có).

**KHÔNG đụng AIR_BUILDER's `register_option`/`register_result`** (kind-11/12, RFC 0012) — đúng như lo ngại (c) ở trên đã dự đoán, chỉ sửa phần TYPECHECK gán type_id, để `get_generic_args`'s nhánh ĐẦU (TYPE_KIND_GENERIC_INST, đã có sẵn) tự động phục hồi đúng T sau này.

**Verify**: `bug67_opt_struct.ax` (Some(Big 16-byte struct).unwrap()) exit 40 đúng (7+12+15+6), không còn segfault. `Option[Vec[i64]]` (`Some(v).unwrap().len`) exit 42 đúng. Fixpoint A==B, regression 96/96 (test mới `bin/t_optstruct.ax`).

## RESIDUAL GAP — ✅ FIXED (3d3c615, cùng phiên tiếp tục)

`HashMap[K, Vec[V]]` (2 tầng generic khác nhau: HashMap chứa Vec làm value): `m.get(1).unwrap().len` từng trả sai (test `bin/t_optnested.ax`).

**Quá trình điều tra (đều trong CÙNG 1 phiên, liên tục không dừng):**
1. Loại trừ `try_instantiate_variant_call` (fix 651d13c): debug print xác nhận nó BIND ĐÚNG `V=Vec[i64]` ngay cả trong ngữ cảnh nested (`concrete=392 kind=1 name=_AX_std_Vec__i64`) — không phải nguyên nhân.
2. Thử hướng "hardening" `SumInfo.generic_args` (c2a2a15) — verify an toàn nhưng KHÔNG fix được case nested. Debug tiếp `get_generic_args`'s nhánh SUM mới: `sum_args.len=0` — nghĩa là dữ liệu KHÔNG BAO GIỜ ĐƯỢC GHI, dù code "trông như" phải chạy.
3. Debug ngay tại **call site** của `set_sum_generic_args` (typecheck.ax, nhánh `NODE_GENERIC_TYPE` xử lý `Option[V]` làm type-ANNOTATION, vd return-type `-> Option[V]`): phát hiện `cloned_kind` đọc ra **GIÁ TRỊ RÁC KHỔNG LỒ** (địa chỉ bộ nhớ, không phải enum nhỏ như `NODE_TYPE_ALIAS_DECL=7`)! **ROOT CAUSE THẬT: bug trong CHÍNH fix c2a2a15 của tôi** — `cloned_kind` được khai báo bằng `let` BÊN TRONG nhánh "fresh instantiation" (`else` của `existing_sym_idx != 0`), nhưng bị tham chiếu ở NGOÀI phạm vi đó. Khi gặp CACHE HIT (`existing_sym_idx != 0`, tức instance đã tồn tại từ trước — TRƯỜNG HỢP PHỔ BIẾN vì Option[Vec[i64]] thường được request nhiều lần) → nhánh fresh KHÔNG chạy → `cloned_kind` giữ nguyên RÁC từ stack → check `if cloned_kind == NODE_TYPE_ALIAS_DECL` LUÔN LUÔN false → **`set_sum_generic_args` KHÔNG BAO GIỜ được gọi, dù chỉ 1 lần, cho BẤT KỲ sum type nào** (Option/Result đều bị). Toàn bộ commit c2a2a15 là "INERT" (vô hiệu) từ đầu.

**Fix (3d3c615)**: thêm `mut fresh_type_alias_inst := false` khai báo CÙNG SCOPE với `inst_sym_idx` (trước if/else), set `= true` CHỈ bên trong nhánh fresh-instantiation khi `cloned_kind == NODE_TYPE_ALIAS_DECL` (đúng chỗ, đúng scope), rồi check `if fresh_type_alias_inst:` tại call site thay vì check `cloned_kind` trực tiếp.

**Verify**: `HashMap[i64,Vec[i64]].get(1).unwrap()` giờ trả đúng Vec (len=3, elements 10/20/30 đúng thứ tự) — exit 42 đúng thay vì exit=2/8 rác. Fixpoint A==B, regression 97/97 (test mới `bin/t_optnested.ax`).

**Bài học quan trọng nhất của cả BUG#67**: khi thêm 1 guard dựa trên biến `let`/`mut` được khai báo TRONG 1 nhánh if/else lồng sâu, PHẢI kiểm tra kỹ biến đó CÓ THỰC SỰ được gán trên MỌI đường đi tới điểm sử dụng hay không — AXIOM's scoping cho phép tham chiếu biến "hoisted" ngoài block khai báo mà KHÔNG báo lỗi compile-time, nên một guard tưởng như đúng có thể ÂM THẦM vô hiệu hoàn toàn (không crash, không lỗi, chỉ đơn giản KHÔNG BAO GIỜ chạy nhánh mong muốn) — nguy hiểm hơn cả 1 lỗi rõ ràng vì tự-test "verify an toàn" (fixpoint+regression xanh) vẫn PASS bình thường (vì code inert không đổi behavior nào cả, không phải vì nó đúng). Chỉ phát hiện được nhờ debug print TRỰC TIẾP tại call site, không chỉ tin vào "code trông hợp lý".

## Bài học methodology (áp dụng lại từ phiên trước, được xác nhận lại lần nữa)

- `bin/axc_native.exe` (daily-driver, KHÔNG track git) PHẢI được rebuild (self-compile 2-hop, xác nhận hash A==B) NGAY SAU MỖI commit fix trước khi dùng làm `AXC=` cho regression suite — quên bước này gây ra 1 "regression giả" (`t_hashi64` FAIL) trong phiên fix BUG#65, chỉ vì binary cũ thiếu fix BUG#66 đã commit trước đó. Đã fix bằng cách rebuild lại và xác nhận đúng.
- Debug print SCALAR-ONLY (`ax_printf_local`, KHÔNG copy struct) vẫn là cách AN TOÀN nhất để xác nhận giả thuyết trước khi patch — dùng lại thành công lần nữa ở bug này (in `tmpl`, `rec_type`, `inferred0` — xác nhận `inferred0=0` TRƯỚC default hoá).
- objdump trên binary build SẴN (không cần sửa compiler) vẫn là cách tốt nhất để lấy ground-truth khi nghi ngờ field-offset/scale — lần này phát hiện field access LẶP LẠI OFFSET 0 cho MỌI field, dấu hiệu rõ ràng của "mất field-layout info hoàn toàn" (khác với BUG#66's "offset đúng nhưng scale sai").

Liên quan: [[bug64-vec-big-aggregate-element]] (RFC 0013, generic STRUCT instantiation — pattern tương tự cần áp dụng cho SUM type ở bug này), [[bug66-hashmap-i64-value-corruption]] (cùng cơ chế fallback "TYPE_UNKNOWN → default i32" khi generic-param binding thất bại), [[bug57-match-option-native]] (lý do TẠI SAO test hiện có không bắt được — match là đường code HOÀN TOÀN riêng).
