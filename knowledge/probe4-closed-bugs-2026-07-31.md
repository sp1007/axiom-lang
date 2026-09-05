---
name: probe4-closed-bugs-2026-07-31
description: The four probe4 bugs found 2026-07-30 — all CLOSED by 2026-07-31 (#1/#2/#3 fixed, #4 dissolved as a wrong problem statement); kept because each teaches a distinct diagnostic lesson
metadata:
  type: project
---

# probe4 — bốn bug tìm 2026-07-30, **CẢ BỐN ĐÃ ĐÓNG 2026-07-31**

> Tách khỏi `BACKLOG.md` ngày 2026-09-05: `BACKLOG.md` là **file con trỏ** (CLAUDE.md §24), không giữ
> sự thật. Không mục nào ở đây còn MỞ — giữ lại vì mỗi mục dạy một bài học chẩn đoán khác nhau, đặc
> biệt **#4 là ví dụ mẫu của một phát biểu vấn đề SAI HOÀN TOÀN** ("bug thanh ghi float") che mất
> nguyên nhân thật (khớp substring khi phân giải tên method).

## 🐞 BUG — probe4 tìm ra 2026-07-30, ĐÃ TỰ XÁC MINH LẠI (không chỉ theo báo cáo agent)
**CẢ BỐN ĐỀU ĐÃ ĐÓNG 2026-07-31** — #1/#2/#3 được sửa, #4 **giải thể** (phát biểu sai, xem dưới).
Giữ nguyên bản ghi vì mỗi mục đều dạy một bài học chẩn đoán khác nhau, và ba mục đầu là
**accept-then-miscompile** (lớp BUG#53) đều có ma trận đối chứng trong
`bin/probe4/`. Repro chạy bằng `sh bin/probe4/run.sh <file.ax>`.
⚠️ Driver cần subcommand `build`: `bin/axc_native.exe build f.ax -o out.exe -O0`. **Thiếu `build`
thì nó exit 0 mà KHÔNG sinh file** — đừng đọc đó là build thành công (tôi đã mắc đúng bẫy này).

1. ✅ **ĐÃ SỬA 2026-07-31 (`6febd02`)** — luật nằm ở **tầng lowering** (`coerce_int_to_float`, nay là
   `coerce_to_float_target`), gọi ở MỌI value site, **không** phải mở rộng hint ở typecheck: hint chỉ
   retype được LITERAL, còn int-**biến** → f64 thì storage đã dựng theo bề rộng của chính nó. Làm cả
   hai nơi sẽ tái dựng đúng hình dạng "hai bản sao một cơ chế" đã đẻ ra cả họ bug này. Chẩn đoán ban
   đầu (mở rộng `typecheck.ax:5208`) **SAI**, và implementer đã nói ra thay vì làm theo. Mô tả gốc:
   **`OP_ICONST` mang `type_id` FLOAT** — `let a: f64 = 3` ra **0.0**. Xác minh: `probe4/h1.ax`
   -O0 exit **1** (đúng phải 42). Root: `air_builder.ax:572-589` `lower_int_lit` phát `OP_ICONST` với
   type_id 9/10, còn sibling `lower_float_lit` (`:591`, emit `:601-614`) làm đúng — **bản sao int
   chưa từng được mở rộng**. Breadth CỰC RỘNG: `let`/assign/param/field-init/array-elem/`return`/
   `3 as f64`/`c + 3`/`3 / 2.0`/`c > 2`. -O0 và -O1 **phân kỳ** (họ aliasing `R10 ≡ XMM10`) ⇒ oracle
   phải chạy CẢ HAI. ⚠️ `knowledge/bugs.md:1015-1019` khẳng định *"`let x: f64 = 3` → OK"* — **câu đó
   SAI và chưa từng được kiểm** — nay đã đúng. Oracle: `bin/t_intlitfloatctx.ax`, `bin/probe7/{p1,p2}.ax`.
2. ✅ **ĐÃ SỬA 2026-07-31** — dynamic dispatch bỏ hết coercion cho đối số (`i.c32(1.5)` → 0.0):
   [bug-iface-dispatch-arg-coercion](bug-iface-dispatch-arg-coercion.md), RFC 0029 §9, oracle
   `bin/t_ifacefloatarg.ax` (hiệu chuẩn **30** trước → **42** sau). ⚠ Chẩn đoán “sửa ở typecheck, gate
   A==B” **SAI PHA**: coercion đối số của method call nằm ở `air_builder.coerce_float_arg` (đọc kiểu
   tham số từ **SYMBOL callee**), dispatch không có symbol ⇒ phải publish signature của interface cho
   backend. Gate thực tế **A==B==C**. ⚠ `probe4/f1.ax` nay ra **110** chứ không phải 42 — vì chạm một
   defect **KHÁC, có từ trước**: f32 **RETURN** qua dispatch đọc thanh ghi cũ khi lời gọi dispatch trước
   đó cũng trả f32 và có đối số (`bin/probe5/r6e.ax` = **101** trên CẢ compiler cũ lẫn mới).
3. ✅ **ĐÃ SỬA 2026-07-31 (`e6c507c`, mở rộng sang đối số method ở `f6ac69e`)** — nay là
   `error[E3031]`, dùng chung `check_annotated_target` với E3030/E3032. ⚠️ Đính chính bản ghi cũ: ở
   **đối số method**, ca float→int KHÔNG phải hình dạng "bit IEEE thô" như `let`: literal f64 đi vào
   XMM còn tham số i64 đọc từ **thanh ghi nguyên CHƯA KHỞI TẠO** ⇒ giá trị không cả deterministic
   theo call context (RFC 0006 §6.2). Mô tả gốc: **`let a: i64 = 3.0` được NHẬN, cho ra bit pattern
   IEEE thô** — `probe4/g13.ax` exit **3** (i64 giữ
   `0x4008000000000000`). **Spec ĐÃ chốt hướng này**: `knowledge/bugs.md:1015` — *"float → int: CẤM
   ngầm, phải `as`"* ⇒ **REJECT + chẩn đoán**, KHÔNG cần user quyết (khác ca `u8 = 300` ở §CẦN USER
   QUYẾT, ca đó chưa có phán quyết). Frontend ⇒ A==B.
4. ❌ **GIẢI THỂ (DISSOLVED) — cách phát biểu ban đầu SAI HOÀN TOÀN. Đừng đi săn bug thanh ghi float
   không tồn tại.** Sự thật (`3741afc` chép lại, fix ở **`a281992`**): phân giải tên method khớp một
   tên trần với **SUBSTRING chặn bởi `_`** của tên method khác ⇒ **chọn NHẦM HÀM**. `p32_r32` kết
   thúc bằng `_r32`, nên **cả hai** slot vtable nhận địa chỉ `p32_r32`; `i.r32()` gọi một hàm 2 tham
   số qua call site 1 đối số và trả về một tham số đọc từ thứ mà lời gọi trước để lại trong XMM1.
   **Thanh ghi cũ là HỆ QUẢ, không phải nguyên nhân.** Chốt bằng disassembly: `mov $0x2a` không hề
   xuất hiện trong image ⇒ `S.r32` chưa bao giờ được sinh. Root cause có trước công việc dispatch
   **hai tháng** (`ec5667d`) — đó chính là lý do repro fail y hệt nhau ở CẢ hai phía của `5359a39`.
   Bug này **type-agnostic** (i64/f32/f64/str/bytes), **không riêng dispatch** (gọi tĩnh cũng sai),
   với dạng nặng hơn là **nhầm KIỂU** (-O0 và -O1 bất đồng) và **nhầm ARITY**.
   ✅ Tự đo lại 2026-07-31: `r6e` = **42**, `probe4/f1.ax` = **42** (trước 110), cả `-O0` lẫn default.
   Mô tả gốc (SAI, giữ làm bản ghi): **f32 RETURN qua dynamic
   dispatch đọc thanh ghi CŨ** khi lời gọi dispatch **ngay trước** cũng trả **f32** *và có đối số*:
   lời gọi thứ hai trả về giá trị của lời gọi thứ nhất. Repro `bin/probe5/r6e.ax` = **101** (đúng phải 42)
   trên **CẢ** `axc_native` cũ lẫn compiler mới ⇒ đã quy trách nhiệm rõ, không đổ cho fix #2. Thu hẹp sẵn:
   hai lời gọi f32 **không đối số** liên tiếp thì ĐÚNG (`r6f.ax` = 42); riêng **vị trí slot** không phải nguyên
   nhân (`r6c.ax` = 42, slot 5). Đây là lý do `probe4/f1.ax` dừng ở **110** thay vì 42.

