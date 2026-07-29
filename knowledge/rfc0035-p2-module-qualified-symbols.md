---
name: rfc0035-p2-module-qualified-symbols
description: "RFC 0035 P2 SHIPPED 2026-07-29 (A==B DC4FD242, 557/557 + lib_collision 5/5): symbol của thư viện nay mang tên module (ax_libpa_helper) suy từ TÊN ĐƠN VỊ BIÊN DỊCH, không phải sym_idx — repro 2 thư viện cùng tên hàm trả về ĐÚNG 30."
metadata:
  type: project
---

# RFC 0035 P2 — module-qualified symbols cho đường `.lib` (SHIPPED 2026-07-29)

Gate: `A==B DC4FD242`, regression **557/557**, ELF 12/12, ctgc 16/16, exe_size 4/4,
so_export ✓, **`scripts/lib_collision_check.sh` 5/5 (MỚI)**.

## Bằng chứng thật là `30`, KHÔNG phải gate xanh

Đây là điểm quan trọng nhất. Self-host build không dùng thư viện ⇒ thay đổi **trơ** ⇒
`A==B` là điều đương nhiên. Chính RFC này đã một lần đạt `A==B B0AEA1C0` + 557/557 trên một
increment **rỗng về cấu trúc** ([[session-handoff]] / RFC §7bis). Nên tiêu chí nghiệm thu là
repro §2: 2 thư viện cùng khai `pub fn helper` (10 và 20), app gọi cả hai → **exit 30**.
Trước P2 nó thậm chí không link được (`unresolved external symbol 'ax_helper__m1755'`).

## Mô hình đúng: qualifier lấy từ ĐƠN VỊ BIÊN DỊCH

`module_qualifier` (main_air.ax) là **hàm thuần** gộp hai cách viết cùng một tên module mà hai
phía đang cầm — thư viện cầm `libpa.ax`/`libpa.lib`, app cầm tên import `libpa` — bằng cách bỏ
đuôi `.ax`/`.lib` rồi map `/`, `\`, `.` → `_`. Không hash, không thứ tự, không index bảng ⇒ hai
lần biên dịch RIÊNG BIỆT tính ra CÙNG một chuỗi mà không chia sẻ gì. Đó đúng là tính chất
`sym_idx` không thể có (RFC §2).

Hai site, một tên, **cả hai return TRƯỚC nhánh trang trí method `ax_<Struct>_<fn>`** để định
nghĩa và tham chiếu không thể lệch nhau khi param đầu là struct:
- **Phía thư viện**: `SymbolTable.unit_qualifier` (field mới), driver chỉ set khi `--staticlib`,
  `x86_resolve_sym_name` đọc cho hàm `pub` của chính unit. Giới hạn ở `--staticlib` vì đó là
  output DUY NHẤT được link vào một compilation KHÁC ⇒ mọi build thường byte-identical.
- **Phía app**: cờ mới `SYM_FLAG_MODQUAL` (8192) trên symbol body-less mà
  `register_module_from_lib` đăng ký từ iface `.lib`; `name_id` của nó ĐƯỢC ĐẶT THÀNH tên đã
  qualify. Tên LOCAL vẫn điều khiển mọi lookup (binding `mod.fn` ở global scope + `ModuleExport`)
  nên resolution mức nguồn không đổi. name_id khác nhau cũng khiến **flag 2048 không còn nổ** —
  chính nó sinh ra lời gọi `ax_helper__m1755` mà không ai định nghĩa.

## 4 sự thật ĐO ĐƯỢC, đừng suy luận lại

1. **`args[0]` KHÔNG phải file nguồn** — nó là argv[0] (đường dẫn exe của chính compiler;
   `ensure_import_libs` truyền nó làm `self_exe`). Dùng nhầm ra qualifier
   `D:_projects_compiler_Axiom_axA_exe`. Biến đúng là `filename = args[file_arg_idx]`.
2. **`mod_name` tới `register_module_from_lib` là RỖNG** (đã probe). Vì thế qualifier phía app
   suy từ `lib_path` (vừa `file_exists()` xong nên chắc chắn còn nguyên). ⇒ binding `mod.NAME`
   mà `register_lib_symbol` ghi từ `mod_name` là **đồ chết**; resolution thực đi qua
   `ModuleExport`. Bẫy cho người đọc code sau.
3. **`std.string.replace` TRẢ VỀ CHÍNH INPUT khi không có match** (std/string.ax:760). Chuỗi
   `replace` + `@free` từng cái = use-after-free đúng trên những tên KHÔNG cần thay — tức phần
   lớn — và nó **segfault resolver ngay lần chạy đầu**. Đã thay bằng MỘT vòng map ký tự.
4. **Module path nhiều đoạn không resolve ở call site**: `import bin.libcol.liba` biên dịch
   thư viện đúng nhưng `bin.libcol.liba.helper()` báo `undefined name 'bin'`. **CÓ SẴN TỪ TRƯỚC**
   (compiler đã ship fail y hệt), không phải do P2. Vì vậy oracle dùng tên module một đoạn.

## Cảnh báo §4 về predicate so tên hoá ra KHÔNG cần làm gì thêm

DFE root set tính MỌI tên nó so sánh **qua chính `x86_resolve_sym_name`**, nên nó tự động đi
theo scheme mới; `#[export]` vốn đã khớp cả intern-id LẪN tên. Runtime/ABI được loại trừ bằng
cách **ĐỌC `is_valid_runtime_dll_symbol` từ linker.ax** (không chép lại — tiền lệ
`dfe_is_abi_name`), thử cả cách viết nguồn lẫn tên phát mặc định.

## Oracle mới `scripts/lib_collision_check.sh` — ĐÃ HIỆU CHUẨN

5 kiểm tra: exit 30; mỗi lib phát `ax_<mod>_helper`; shim ABI (`ax_sum_layout_is_pointer`) GIỮ
tên cố định; iface vẫn ghi tên LOCAL (`F helper 0 -> i32`). Trên compiler CŨ nó **fail 3/5 đúng
triệu chứng §2**, còn 2 guard chống over-qualify thì pass ở CẢ HAI bản — nên nó thật sự phân
biệt được, không phải test luôn xanh.

## Flag day ĐÃ ĐÓNG (cùng phiên): `AX_LIB_SCHEME_ID`

Manifest staleness vốn chỉ hash NGUỒN thư viện ⇒ `.lib` do compiler scheme CŨ tạo bị coi là
"fresh" và link thẳng vào tên không ai phát. Nay hằng `AX_LIB_SCHEME_ID` (main_air.ax, hiện
= 2) được trộn vào giá trị ghi/so trong manifest qua **một hàm duy nhất** `lib_manifest_hash`
(mọi producer/consumer phải đi qua nó, nếu không sẽ thấy lib stale vĩnh viễn). **Bump khi nào**:
tên public phát ra của build `--staticlib` đổi, hoặc format `__axiom_iface` đổi.

⭐ **Hiệu chuẩn phải dùng HAI binary** — không thể khẳng định từ trong một lần chạy: cùng một
nguồn, compiler TRƯỚC guard ghi `d93a771928dbb2fc`, compiler SAU ghi `fb955e482b400e93`; bản
mới thấy stale → **rebuild** (`[import] building library from source`) → chạy đúng. Chiều
ngược lại cũng kiểm: lib đang up-to-date rebuild **0 lần** (cache không hỏng). Row
`stale_manifest_rebuild` trong oracle CHỈ chứng minh *cơ chế* (manifest lạ ⇒ rebuild) — nó sẽ
pass cả khi không có guard, và comment trong script nói đúng như vậy thay vì nhận vơ.

## Còn mở sau P2

Method/global/ctor vẫn dùng scheme cũ (`axS_`/`axG_`/`axC_` chưa làm) ⇒ fn-vs-struct vẫn dựa
vào reject ở typecheck ([[bug-user-fn-stdlib-struct-name-collision]]). P3 (E0501 thành error)
vẫn bị chặn: mỗi lib nhúng bản sao runtime shim riêng và những tên đó CỐ Ý không qualify.
Xem [[task-cross-library-name-collision]].
