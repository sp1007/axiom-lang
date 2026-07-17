---
name: bug64-vec-big-aggregate-element
description: "BUG#64 FIXED (b08867b): Vec element aggregate >8B (regalloc_is_16byte OP_INDEX). +diag fix (d0c12ea): reject 'call trên field không phải hàm' (v.len()) — TypeChecker.diags_count mới, driver HALT trước codegen. BUG#65 FIXED (RFC 0013, 4fb74f4): generic-lồng-generic (Vec[Vec[T]]/MV[MV[i32]]) mangled '__' delimiter ambiguous → StructInfo giờ mang generic_args riêng (mono.instantiate_function ghi trực tiếp), get_generic_args ưu tiên đọc đó thay vì parse tên. Xem [[rfc0013-nested-generic-args]]."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#64 FIXED 2026-07-05 (b08867b, backend-only, fixpoint n3==n4).**

Phần tử của Vec (và mọi collection) là aggregate **LỚN HƠN 8 byte** từng bị trả về RÁC/segfault từ `.get()` → `Some(elem)` → `.unwrap()`:
- `Vec[P8]` (struct 8 byte) → ĐÚNG kể cả trước fix (exit 42).
- `Vec[Big]` (struct 16 byte a,b,c,d i32) → TRƯỚC: rác (91/119/235/14/105 tùy build — non-deterministic-looking nhưng thực ra deterministic theo từng biến thể fix); SAU: đúng (42).
- `Vec[Vec[i32]]` (Vec 24 byte data+len+cap) → TRƯỚC: **SEGFAULT**; SAU (phần tử-size): FIXED nhưng lộ **BUG#65 riêng** (xem dưới) khi làm `v0.get(0)` (chained generic method).

**Root cause thật (tìm bằng gdb, KHÔNG đoán):** `regalloc_is_16byte_cached` (x86_selector.ax) KHÔNG có case riêng cho `OP_INDEX` → rơi vào catch-all `else`, mà BUG#56 fix ở đó chỉ loại trừ kind 6/11/12 (sum/option/result) khỏi phân loại "16-byte-inline", KHÔNG loại trừ struct/array thường (kind 1/3/8) tình cờ đúng 16 byte (Big = 4×i32). Struct đó được giữ BY-ADDRESS như mọi aggregate khác trong backend này, nên `Vec[Big].get(i)->Big` (OP_INDEX trả về ĐỊA CHỈ phần tử 8 byte) bị phân loại nhầm "16-byte-inline" → `OP_RETURN`'s is_16byte branch tách địa chỉ 8-byte đó thành 2 lần load-vào-register (RDX=[addr+8], RAX=[addr]); nếu temp-address-vreg tình cờ được cấp phát TRÙNG RDX, load đầu tiên tự ghi đè lên chính nó trước khi load thứ hai đọc — rác hoặc null-deref. Xác nhận qua gdb: `mov (%rdx),%rax` với rdx=0 crash; và riêng case Vec[Vec] ban đầu debug thấy `rip` nhảy tới đúng giá trị 42 vừa push (call qua register rác).

**Fix (SCOPED hẹp, KHÔNG mở rộng catch-all chung):** thêm case `elif op == OP_INDEX` riêng trong `regalloc_is_16byte_cached`, áp `type_is_aggregate` exclusion (giống OP_DEREF/OP_COPY/OP_CALL đã fix BUG#56/#60). **Lần thử đầu mở rộng catch-all CHUNG bị regression `tstruct_abi`** (struct NHÚNG BY-VALUE trong struct khác, vd `Outer.inner: P` với P 16-byte, dựa vào phân loại HẸP cũ cho register-pair ABI convention tại call-boundary) — đây là bài học: catch-all dùng chung cho OP_GET_FIELD/nhiều case khác, KHÔNG được đụng; chỉ sửa case cụ thể gây lỗi.

**Fixpoint lưu ý:** cần THÊM 1 hop mới hội tụ — n1(driver cũ)≠n2(build đầu với codegen mới, EXPECTED vì thay đổi đụng chính source compiler), n2≠n3 (VẪN khác, bất ngờ), nhưng **n3==n4** xác nhận hội tụ thật. Gate: build tay 4 hop khi đổi core-codegen ảnh hưởng pervasive tới chính compiler.

Test vec_big_element_BUG64.ax un-SKIP (println+.expected). Regression 93/93 + tstruct_abi xác nhận KHÔNG regress.

---

**BUG#65 OPEN — phát hiện khi verify BUG#64 (Vec[Vec] full case).**

`v0.get(0)` với `v0 = outer.get(0)` (kết quả CHAINED của một generic call khác) → lower thành `getfld`+indirect-call (KHÔNG resolve thành call trực tiếp tới `get[i32]` mono instance) → segfault. Cùng họ triệu chứng BUG#61/#63 (method-dispatch fallback) nhưng **KHÁC vị trí**: đây là lời gọi TRỰC TIẾP trong `main` (không phải trong thân hàm generic clone), trên một LOCAL VAR (`v0`) được gán từ kết quả của MỘT generic call khác (`outer.get(0)`, outer: `Vec[Vec[i32]]`/`MV[MV[i32]]`).

Giả thuyết: type-inference cho `v0`'s type (phải là `Vec[i32]`/`MV[i32]`, tức generic-inst của outer's element type) không được thread đúng qua kết quả trả về của method generic `get[T]` khi T chính nó là một generic-inst (`T=MV[i32]`) — nên khi typecheck cố resolve `.get` trên `v0`, method-resolution không nhận diện `v0`'s type khớp template `get[T]` (có thể do tên mangled `MV__i32` không match logic `resolve_method_sym`/`match_method_name` như mong đợi, hoặc return-type recovery của `outer.get(0)` (theo kiểu BUG#61's `method_ret_type`) không xử lý trường hợp return type CHÍNH NÓ là generic-inst).

**Repro tối giản (self-contained, KHÔNG cần std.collections):**
```
struct MV[T]:
    data: ptr[T]
    len: i64
fn new_mv[T]() -> MV[T]: ...
fn push[T](mut self: MV[T], item: T): ...
fn get[T](self: MV[T], i: i64) -> T:
    return self.data[i]
fn main() -> i32:
    mut outer = new_mv[MV[i32]]()
    mut a = new_mv[i32]()
    a.push(42)
    outer.push(a)
    let v0 = outer.get(0)
    let x = v0.get(0)   // <-- CRASH: getfld+call fallback thay vì resolve get[i32]
    return x
```
Lưu tại `scratch/mvmv.ax` (session này, KHÔNG commit — dùng để tái hiện bug lần sau). AIR dump xác nhận: `main`'s `%12=getfld %10; %13=call %12,%11` (v0.get(0)) — pattern giống hệt BUG#61 khi chưa fix.

**ROOT CAUSE TÌM RA (traced, KHÔNG phải đoán) — LÀ VẤN ĐỀ MANGLING/ARCHITECTURE, KHÔNG PHẢI ONE-LINE BUG:**

`v0 = outer.get(0)` bind SAI type (AIR cho thấy `%10: t3 = copy %9` — t3=i32, ĐÚNG PHẢI LÀ `MV[i32]`). Truy ngược:
1. `get[T]`'s `T` được suy ra từ `infer_generic_type_args` khớp `self: MV[T]` với `outer`'s type = `MV[MV[i32]]`. Logic gọi `get_generic_args(outer_type)` để lấy type-args của outer.
2. `get_generic_args`: nếu `entry.kind == TYPE_KIND_GENERIC_INST` thì dùng vector `generic_insts` CHÍNH XÁC. NHƯNG mono hóa STRUCT (`mono.instantiate_function`) đăng ký kiểu qua `typetable.register_struct(mangled_id, ...)` → kết quả có `entry.kind == TYPE_KIND_STRUCT` (kind=1), KHÔNG PHẢI GENERIC_INST! Nên `get_generic_args` rơi vào fallback **`get_generic_args_from_mangled`** — parse type-args bằng cách TÁCH CHUỖI tên mangled trên mọi dấu `"__"`.
3. **Tên mangled của `MV[MV[i32]]`** (qua `mono.mangle_name`+`get_type_name_recursive`): recursive nên = `"_AX_std_MV" + "__" + get_type_name_recursive(MV[i32]_type_id)`. Vì `MV[i32]` ĐÃ được mono hóa trước đó (struct instance, `entry.name_id` = chuỗi interned `"MV__i32"`), `get_type_name_recursive` trả THẲNG chuỗi đó (không escape) → mangled name outer = `"_AX_std_MV__MV__i32"`.
4. `get_generic_args_from_mangled` tách trên MỌI `"__"` — với `"_AX_std_MV__MV__i32"`, sau prefix nó thấy 2 segment `"MV"` và `"i32"` → hiểu nhầm outer có **2 type-arg đơn** (`MV`, `i32`) thay vì **1 type-arg lồng nhau** (`MV[i32]`). `parse_type_from_name("MV")` (tên template trần, không tham số) rất có thể trả về sai/0/rơi vào fallback — khiến `T` được suy luận SAI (rơi về i32 qua fallback default `TYPE_UNKNOWN → TYPE_I32` ở nơi khác trong `infer_node`).

**ĐÂY LÀ LỖ HỔNG THIẾT KẾ (mangling scheme), KHÔNG PHẢI BUG cục bộ:** chuỗi mangled tên kiểu generic-lồng-generic dùng CHUNG delimiter `"__"` đệ quy → không thể phân biệt "N type-arg đơn" với "1 type-arg lồng nhau" khi round-trip qua string parsing, vì không có dấu ngoặc/độ dài tiền tố nào đánh dấu ranh giới. Theo CLAUDE.md §13 (RFC Policy: "ABI changes" cần RFC) và §20 (uncertainty → không đoán ngầm semantics) — đây LÀ loại thay đổi cần RFC, không phải patch nhanh. Sửa đúng cách cần MỘT trong:
(a) Khi mono hóa STRUCT generic, cũng lưu `generic_insts` (args vector CHÍNH XÁC) như cách GENERIC_INST làm — để `get_generic_args` KHÔNG BAO GIỜ cần fallback string-parse cho type đã mono hóa qua compiler (chỉ cần fallback cho type export/import qua .lib interface, nơi string LÀ nguồn thật duy nhất).
(b) Đổi format mangling để tự-phân-định (length-prefix mỗi arg, hoặc ký hiệu ngoặc `<...>` thay vì `__` phẳng) — rủi ro cao hơn vì đụng khắp linkage/symbol-matching/DLL export.

**Khuyến nghị hướng (a)** — an toàn hơn format-wise (không đổi chuỗi mangled hiện có, giữ tương thích .lib interface), nhưng **CHI PHÍ THẬT KHÔNG NHỎ**: đã kiểm tra `TypeTable` struct (`typetable.ax:241-254`) — constructor `new_type_table` dùng **`@alloc(120) as ptr[TypeTable]`, kích thước HARDCODE** (không phải `size_of`/sizeof tự động). Thêm field mới (vd `struct_generic_args: U32VecVec`, 24 byte) đòi hỏi sửa magic number 120→144 và rà soát MỌI chỗ khác có giả định layout/size TypeTable cứng (rủi ro giống các "stage0 ceiling"/hardcoded-struct-size bug đã gặp trước đây — xem [[next-step-15-selfhost-status]]). KHÔNG PHẢI "chỉ thêm field" đơn giản như tưởng ban đầu — củng cố thêm lý do đây là việc cần làm CẨN THẬN có RFC, không tự ý patch giữa phiên.

---

**Diag fix phụ SHIPPED 2026-07-05 (d0c12ea)** — thay vì rush BUG#65, pivot sang việc AN TOÀN + giá trị cao hơn: thêm diagnostic reject "gọi field không phải hàm" (`v.len()` khi Vec chỉ có FIELD `len: i64`, không có method). Thêm `TypeChecker.diags_count: i64` (struct này dùng `size_of[TypeChecker]()` — KHÔNG hardcode size như TypeTable, nên an toàn thêm field). Check đặt trong `infer_node`'s NODE_CALL_EXPR khi callee là NODE_FIELD_EXPR: nếu callee_type resolve được (không UNKNOWN) nhưng kind KHÔNG thuộc {FUNC, STRUCT, SUM, GENERIC_INST, OPTION, RESULT} → emit lỗi + tăng diags_count. SCOPE hẹp đúng chỗ: qualified constructor call (`Mod.Struct(...)`, `Enum.Variant(x)`) CŨNG là FIELD_EXPR callee nhưng resolve về STRUCT/SUM nên KHÔNG bị chặn nhầm (đã verify qua qualified_variant tests). Driver (main_air.ax) HALT trước codegen khi `checker.diags_count>0`, giống HỆT pattern BUG#53 (parser.diags_count). Verify: compiler tự build sạch (không false-positive trên CHÍNH source ~1.35MB), regression 93/93, toàn bộ generics/match/ABI test pass, fixpoint n1==n2 (hội tụ ngay, không đụng codegen pervasive). Test mới: tests/sema/err_call_field_not_function.ax (+.diag, quy ước `err_*.ax` = expect non-zero exit).

**Phát hiện phụ (KHÔNG fix, ngoài phạm vi):** `tests/sema/err_call_non_func.ax` (gọi biến int thường `x()`, callee là NODE_IDENT chứ không phải NODE_FIELD_EXPR) hiện VẪN compile thành công (exit 0) dù tên test ngụ ý phải lỗi — gap CŨ, KHÔNG liên quan tới fix này (fix chỉ scope NODE_FIELD_EXPR). Để dành cho lần khác nếu cần.

Repro tối giản (KHÔNG commit, giữ ở `scratch/mvmv.ax` session-local): xem code trong description phía trên. AIR xác nhận `%12=getfld %10; %13=call %12,%11` (v0.get(0)) = pattern method-dispatch-fallback giống BUG#61 CŨ.

Liên quan: [[bug60-61-option-followups]] (cụm tagged-type đã đóng), method-dispatch-fallback family. Ưu tiên: THẤP-TRUNG (Vec[Vec[T]]/generic-lồng-generic collection ít phổ biến hơn Vec[BigStruct] đã fix ở BUG#64), nhưng là landmine kiến trúc cần RFC trước khi đụng.

---

**BUG#65 FIXED — RFC 0013, `4fb74f4`, 2026-07-06.**

Hướng (a) đã ship, nhưng **KHÔNG đụng `TypeTable`** như lo ngại ban đầu (dòng 65 ở trên): thay vì thêm field vào `TypeTable` (hardcode `@alloc(120)`, rủi ro cao), field mới `generic_args: U32Vec` được thêm vào **`StructInfo`** (phần tử của `TypeTable.structs`) — struct này KHÔNG hardcode size, `StructInfoVec.push` grow qua `@compiler_intrinsic("size_of")[StructInfo]()` tự động, nên thêm field AN TOÀN, không cần sửa magic number nào. Chi phí thật hoá ra thấp hơn nhiều so với đánh giá ban đầu.

Cơ chế: `mono.ax::instantiate_function` đã CÓ SẴN `args: U32Vec` (type-arg cụ thể) làm tham số ngay tại chỗ gọi `register_struct` (mono.ax:421) — gọi thêm `typetable.set_struct_generic_args(struct_type_id, args)` (typetable.ax, hàm mới) ngay sau đó, DEEP-COPY `args` (không chỉ alias con trỏ) vì caller (`typecheck.ax`'s `inferred` vector) tự `@free` nó sau khi `instantiate_function` return — quên deep-copy sẽ tạo dangling pointer. `typecheck.ax::get_generic_args` thêm 1 nhánh MỚI: nếu `entry.kind==TYPE_KIND_STRUCT` và `structs.data[entry.extra].generic_args.len>0` → trả thẳng, chỉ fallback `get_generic_args_from_mangled` khi rỗng (struct đến từ .lib/.dll interface import, RFC 0011 P5, nơi string vẫn là nguồn thật duy nhất — CÒN MỞ, ngoài phạm vi RFC 0013).

**Cạm bẫy suýt gây "regression" giả (đã debug, không phải bug thật):** `pre_infer_struct`'s nhánh `else` (typecheck.ax ~1118, re-đăng-ký fields cho struct đã có type_id) GHI ĐÈ TOÀN BỘ `StructInfo` bằng struct-literal MỚI chỉ có `fields` — nếu không cẩn thận sẽ XÓA MẤT `generic_args` vừa set. Fix: đọc `prev_generic_args` từ entry cũ trước khi ghi đè, carry-forward vào struct-literal mới.

**Fixpoint xác nhận riêng lẻ KHÔNG đủ — bài học methodology quan trọng:** lần chạy full regression đầu tiên "FAIL t_hashi64" tưởng là regression thật do fix BUG#65, nhưng hoá ra do `bin/axc_native.exe` (daily-driver, KHÔNG track git) bị STALE — build từ TRƯỚC khi BUG#66 (`73f396d`) được commit, nên thiếu fix đó. `scripts/fast_fixpoint.ps1` chỉ tạo `axc_fpA.exe`/`axc_fpB.exe` riêng, KHÔNG tự cập nhật `axc_native.exe`. **Quy tắc mới: sau MỖI commit fix (đặc biệt đụng typecheck/mono/codegen), phải tự rebuild `bin/axc_native.exe` (self-compile 2-hop, xác nhận hash A==B) TRƯỚC KHI dùng nó làm `AXC=` cho regression suite** — nếu không, false-positive/false-negative do binary cũ sẽ đánh lừa.

Test: `bin/t_nestedgen.ax` (row `t_nestedgen|exit|42`). Regression 95/95. RFC: [[rfc0013-generic-instantiation-type-args]] (file `rfcs/0013-generic-instantiation-type-args.md`).
