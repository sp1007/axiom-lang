---
name: bug68-struct-eq-no-overload
description: "BUG#68 FIXED (91bae28): so sánh struct VALUE bằng == khi struct KHÔNG có eq overload (RFC 0007) từng âm thầm miscompile thành SO SÁNH ĐỊA CHỈ (structs giữ by-address) thay vì field-by-field — 2 giá trị struct khác instance nhưng field giống hệt bị coi là 'khác nhau', không có diagnostic nào. Hệ quả cụ thể: HashMap[StructKey, V] không bao giờ tìm thấy key vừa insert. Fix: diagnostic reject (không phải auto-derive equality) tại typecheck, giống hệt pattern BUG#53/#64."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#68 FIXED — 91bae28, 2026-07-06. Phát hiện qua proactive-probing (thử HashMap[StructKey,i64] và struct==struct trực tiếp), NGAY SAU khi đóng hoàn toàn BUG#67.**

## Triệu chứng (CONFIRMED, tách biệt hoàn toàn khỏi cụm generics/Option)

```
struct Point: x: i32, y: i32
let a = Point(x: 1, y: 2)
let b = Point(x: 1, y: 2)
a == b   // → FALSE (SAI! phải là true, field giống hệt)
```
Không cần generic/collection gì cả — struct THƯỜNG, so sánh TRỰC TIẾP. Và hệ quả thực tế: `HashMap[Point, i64]` không BAO GIỜ tìm thấy key vừa `insert()` (vì `self.keys[idx] == key` bên trong `get()` luôn false).

## Root cause (đọc code, không cần objdump lần này — rõ ràng ngay khi đọc air_builder.ax)

RFC 0007 (operator overloading, air_builder.ax `lower_binary_expr` dòng ~707-728): khi lhs của `a <op> b` là STRUCT/SUM/GENERIC_INST, code tìm method overload tương ứng (`==`→`eq`, `<`→`lt`, ...) qua `resolve_op_method`. **NẾU KHÔNG TÌM THẤY** (không có overload), code **RƠI XUỐNG dòng 730 TRỞ ĐI** — đường xử lý SCALAR MẶC ĐỊNH (`OP_EQ` với 2 toán hạng là raw register value). Với struct (aggregate, giữ BY-ADDRESS trong backend này), `lhs_reg`/`rhs_reg` MỖI CÁI giữ ĐỊA CHỈ của instance riêng — `OP_EQ` so sánh 2 ĐỊA CHỈ, không phải field. 2 instance khác nhau (dù field giống hệt) → địa chỉ khác → luôn "not equal". **KHÔNG CÓ DIAGNOSTIC NÀO** — accept-then-miscompile, đúng lớp bug đã fix nhiều lần trước (BUG#53 dòng if 1-line, BUG#64 field-not-function).

`typecheck.ax` (dòng ~1565+, `NODE_BINARY_EXPR`, nhánh `op==1` xử lý so sánh) hoàn toàn KHÔNG kiểm tra việc này — chỉ gán `result_type = TYPE_BOOL` bất kể operand có method overload hay không. Không có bất kỳ validation nào ở typecheck level cho RFC 0007 — TOÀN BỘ cơ chế sống ở air_builder (codegen), typecheck "mù" về nó.

## Fix (91bae28) — diagnostic reject, KHÔNG PHẢI auto-derive equality

Theo đúng tinh thần tối giản CLAUDE.md: KHÔNG implement 1 tính năng mới ("auto-derive structural equality" cho struct không có overload — sẽ là RFC riêng, phạm vi lớn hơn nhiều: memcmp toàn bộ hay field-by-field, xử lý padding, xử lý field kiểu con trỏ bên trong, v.v.). Thay vào đó: THÊM DIAGNOSTIC reject sớm, giống hệt pattern BUG#53/#64 (driver HALT khi `diags_count>0` trước codegen).

`typecheck.ax`'s `NODE_BINARY_EXPR` (`op==1`, so sánh): sau khi suy luận `t1`/`t2`, nếu `t1` (lhs) là STRUCT/SUM/GENERIC_INST → lấy operator token text → map sang method name (==/!= → eq, < → lt, <= → le, > → gt, >= → ge) → kiểm tra method THẬT SỰ tồn tại → nếu KHÔNG → emit lỗi + `diags_count++`.

**Cạm bẫy suýt gây 2 false-positive regression (t_opover, t_u128)** — dùng `self.resolve_method_sym(mname_id, t1)` (cơ chế method-resolution SẴN CÓ của typecheck) để kiểm tra tồn tại — NHƯNG method khai báo THEO KIỂU INLINE trong struct (`struct Num: fn eq(self, o): ...`, phong cách t_opover.ax/t_u128.ax) được đăng ký với tên MANGLED CÓ DẤU CHẤM (vd `"Num.eq"`), trong khi `resolve_method_sym` chỉ resolve theo TÊN THUẦN qua scope (`self.symtable.resolve("eq")`) — KHÔNG tìm thấy `"Num.eq"`. `resolve_op_method` (air_builder.ax, cơ chế THẬT được dùng ở codegen) lại dùng SUBSTRING-MATCH (`match_mangled_method_raw_bytes`) — tìm thấy đúng. **Bài học: diagnostic PHẢI dùng CHÍNH XÁC cùng cơ chế resolution với quyết định lowering thật, không được dùng 1 cơ chế "trông tương tự" khác — nếu không, diagnostic và hành vi thật sẽ LỆCH NHAU** (ở đây: false-positive reject 1 case ĐÁNG LẼ hoạt động đúng). Fix: viết `diag_resolve_op_method` (typecheck.ax) MIRROR CHÍNH XÁC `resolve_op_method` (cùng logic unwrap ptr/ref, cùng `match_mangled_method_raw_bytes`, cùng cách match rhs type) thay vì dùng `resolve_method_sym`.

**Verify**: struct KHÔNG có eq overload → error rõ ràng ("type 'Point' does not implement 'eq' for this comparison..."), HALT trước codegen — hết miscompile im lặng. Struct CÓ eq overload (t_opover.ax, t_u128.ax, `HashMap[Point,i64]` với `fn eq` tự viết) → hoạt động ĐÚNG END-TO-END (đã test cả HashMap[StructKey,V] thật). Fixpoint A==B, regression 98/98 (test mới `bin/t_structeq.ax` + `tests/sema/err_struct_eq_no_overload.ax`).

**Ngụ ý cho code hiện có**: nếu có bất kỳ chỗ nào trong compiler/stdlib TỰ MÌNH so sánh struct bằng `==` mà KHÔNG có eq overload, session này compiler VẪN self-compile sạch (không lỗi) → xác nhận KHÔNG có usage nào như vậy tồn tại trong codebase hiện tại (an toàn, không phải sửa thêm ở đâu).

Liên quan: [[bug64-vec-big-aggregate-element]] (diag fix d0c12ea, cùng pattern "reject thay vì miscompile"), RFC 0007 (operator overloading, chưa có memory riêng — xem `rfcs/0007-operator-overloading.md`).
