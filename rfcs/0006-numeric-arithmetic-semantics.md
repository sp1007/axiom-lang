# RFC 0006 — Numeric & arithmetic semantics (DRAFT)

- **Status:** Draft (spec chốt qua thảo luận 2026-06-26; chưa implement)
- **Author:** self-host team
- **Tracking:** next-step-15 task 4; gom BUG#33, BUG#34, BUG#35
- **Liên quan:** RFC 0005 (int literal inference — mở rộng cho float literal), tests/arith/

## 1. Motivation

Phép toán là nền tảng → phải đúng & nhất quán. Hiện native backend có nhiều lỗ hổng
(đã đo bằng dump-air): float arithmetic không phát (luôn integer opcode), unsigned
div dùng IDIV signed, int↔float convert bằng copy-reinterpret (sai giá trị), float→int
ngầm không báo lỗi. RFC này định nghĩa ngữ nghĩa số học và conversion, kèm bộ test
differential `tests/arith/` (≥100k biểu thức, oracle fixed-width).

## 2. Ngữ nghĩa toán hạng & literal

- **Toán hạng quyết định phép toán.** Kiểu khai báo (let/return target) KHÔNG lan
  xuống toán hạng. VD `let b: f64 = a/3` (a:i32) → `a/3` là **int division** = 3, rồi
  int→float ở chỗ gán → **b = 23.0** (KHÔNG phải float division).
- **Literal adopt kiểu (RFC 0005, mở rộng cả float literal):** một literal (int hoặc
  float) nhận kiểu của **toán-hạng-anh-em CỤ THỂ** nếu có; nếu không (cả hai literal,
  hoặc không có sibling cụ thể) thì nhận **kiểu ngữ cảnh** (expected của let/return).
  - `a/3` (a:i32 biến): `3` adopt i32 (sibling cụ thể) → int division.
  - `x < 10` (x:u32): `10` adopt u32 → so sánh u32.
  - `10/3.0` (cả hai literal): adopt ngữ cảnh; `let a: f32 = 10/3.0` → cả 2 → f32 →
    f32 division = 3.333 (f32), không narrowing.
  - `let a: f32 = 3.0` → 3.0 adopt f32 (không còn luôn-f64).

## 3. Mixed int/float trong một phép toán

Một toán hạng float → cả phép là float; toán hạng int được **promote ngầm int→float**
(chèn cvtsi2ss/sd). VD `a/3.0` (a:i32, 3.0 f64) → a promote f64, float division.

## 4. Conversion policy (BẤT ĐỐI XỨNG)

| Từ → Đến | Chính sách |
|---|---|
| int → float (vd i32→f64) | **ngầm OK**, chèn cvt thật (cvtsi2sd) |
| float → int (vd f32→i32) | **CẤM ngầm → LỖI**; phải `as` (truncate toward zero) |
| f32 → f64 (widen) | **ngầm OK** (cvtss2sd) |
| f64 → f32 (narrow, lossy) | **CẤM ngầm → LỖI**; phải `as` (cvtsd2ss) |
| int → int khác kiểu | [TBD] widening ngầm OK; narrowing/đổi-dấu: cảnh báo hay cấm? |

Quy tắc cài (typecheck, ở let/assign/arg/return): so *category* target vs expr:
- target int & expr float → LỖI. target float & expr int → chèn cvt int→float.
- target f32 & expr f64 → LỖI. target f64 & expr f32 → chèn cvt widen.
Mọi cvt int↔float / f64↔f32 phải là **lệnh convert thật**, KHÔNG phải OP_COPY reinterpret
(bug hiện tại: `let a:f32 = 3` ra copy → a = denormal rác).

VD chốt:
- `let a:u64 = (b+2.0)*3` (b:f64) → **LỖI** (float→int). `((b+2.0) as u32)*3` → OK.
- `let a:f32=10/3; let b:i64=a+10` → **LỖI** tại b (float→int). `(a+10) as i64` → b=13.
  (10/3 = int div = 3 → a = 3.0 (int→float cvt); a+10 = 13.0; as i64 = 13.)
- `let a:f32 = 10/3.0; let b:i32 = a as i32` → a=3.333(f32), b = **3**.

## 5. Signed / unsigned

- div/mod: signed type → IDIV (CQO); unsigned type → DIV (xor rdx). Hiện LUÔN IDIV
  signed (BUG#34.1) → sai cho u32/u64 lớn. `map_binary_op` + selector phải chọn theo
  signedness của kiểu toán hạng.
- shift phải: signed `>>` = SAR (arithmetic), unsigned `>>` = SHR (logical).
- compare unsigned dùng JB/JA…, signed dùng JL/JG…

## 6. Overflow (user chốt 2026-06-26: wrap runtime + lỗi nếu tràn hằng)

- **Runtime:** tràn = **wrap** fixed-width (two's complement). Phép toán GIỮ kiểu toán
  hạng (i16+i16 → i16), backend phải **mask về bề rộng** sau mỗi op (hiện KHÔNG mask →
  tính trong 64-bit, BUG#36). KHÔNG promotion kiểu C (i16+i16 vẫn i16, không lên i32).
- **Compile-time:** nếu tràn THẤY ĐƯỢC lúc biên dịch (toán hạng toàn HẰNG) → **LỖI**
  (constant-overflow checker). VD `let b: u8 = 255 + 1` → lỗi; `i16: 1234*56` → lỗi.
- Hệ quả test: ca tràn-hằng = EXPECT-ERROR (tests/arith/diag/e08-e10), KHÔNG phải value
  test. Runtime-wrap value test cần toán hạng non-const (opaque) — TODO harness.

### 6.1 REQUIREMENT — literal nguyên ngoài dải của kiểu ĐƯỢC CHÚ THÍCH ⇒ LỖI (user chốt D1, 2026-07-30) ✅ IMPLEMENTED

Đây là phần ĐẦU TIÊN của "constant-overflow checker" ở §6 được cưỡng chế, và là quyết
định cuối cho §9 mục 4 **ở ca literal trần**.

**Quy tắc (bắt buộc):** một **literal nguyên** viết tại vị trí mà lập trình viên đã
**CHÚ THÍCH TƯỜNG MINH** một kiểu nguyên hẹp, mà giá trị KHÔNG biểu diễn được trong kiểu
đó ⇒ **`error[E3030]`** + `diags_count++` (driver dừng trước codegen). KHÔNG mở rộng kiểu
ngầm, KHÔNG wrap ngầm. Muốn cắt bit thì viết `300 as u8` (vẫn hợp lệ, không đổi).

Lý do chọn REJECT (thay vì wrap hoặc widen): đây là lựa chọn DUY NHẤT không có kết cục
IM LẶNG. Giá trị đã biết lúc biên dịch; người dùng TỰ VIẾT kiểu đó, nên widen là phản lại
điều họ viết, còn wrap là mất dữ liệu không báo. Trước thay đổi này `let x: u8 = 300` giữ
nguyên **300** trong một ô u8 — `u8` không chứa u8 — đúng dạng accept-then-miscompile mà
quy ước BUG#53 của dự án coi là kết cục tệ nhất.

**Vị trí áp dụng** (mọi vị trí có kiểu do người dùng viết):

| Vị trí | Ví dụ bị từ chối | Hành vi CŨ (đã đo) |
|---|---|---|
| `let`/`const` có chú thích (kể cả global) | `let x: u8 = 300` | giữ **300** |
| gán vào binding CÓ chú thích | `mut x: u8 = 1; x = 300` | giữ 300 |
| gán vào **field struct** | `struct S: v: u8` … `s.v = 300` | **wrap 44** (store 1 byte) |
| đối số của tham số hẹp | `fn f(v: u8)` … `f(300)` | giữ 300 |
| đối số ở dạng **type-arg tường minh** | `fn pick[T](a:T,b:T)` … `pick[u8](300,1)` | giữ 300 |
| khởi tạo field struct | `struct P: a: u8` … `P(a: 300)` | wrap 44 |
| phần tử mảng có kiểu phần tử chú thích | `let a: [u8;3] = [1,2,300]` | wrap 44 |
| `return` của hàm khai báo kiểu hẹp | `fn g() -> u8: return 300` | giữ 300 |

**Dải hợp lệ** (bất đối xứng của min có dấu là bắt buộc đúng): i8 `-128..127`, i16
`-32768..32767`, i32 `-2147483648..2147483647`, u8 `0..255`, u16 `0..65535`, u32
`0..4294967295`. Mọi `u*` từ chối MỌI literal âm.

**Ngoài phạm vi (có chủ ý, ghi rõ để không bị hiểu là thiếu sót):**
1. **i64/isize**: `parse_comptime_int` wrap ở 64 bit nên literal vượt i64 không phân biệt
   được với literal hợp lệ ⇒ không kiểm.
2. **u64/usize**: chỉ từ chối literal ÂM cú pháp. Literal không dấu > i64-max (vd
   `let big: u64 = 18000000000000000000`) parse ra i64 âm và PHẢI tiếp tục hợp lệ
   (t_u64cmp ghim điều này).
3. **Biểu thức hằng gấp** (`let b: u8 = 255 + 1` ở §6) — vẫn TODO; cần constant folding ở
   typecheck, không thuộc thay đổi này.
3b. ⚠️ **ĐỐI SỐ của lời gọi PHƯƠNG THỨC** (`s.setv(300)` với `setv(self, x: u8)`) — **CHƯA
   phủ**, đã đo: vẫn nhận và **wrap âm thầm về 44**. Cơ chế khác (phân giải method đi qua
   khối quét `mfi.params` ở `typecheck.ax` ~L4640, không dùng `fp_data`); đó là chỗ móc nếu
   mở rộng. Cũng chưa phủ: **gán vào PHẦN TỬ mảng** (`a[0] = 300`) — cố ý, vì kiểu phần tử có
   thể do suy diễn (`mut a = [1,2,3]` là `[i32;3]`), không phải kiểu người dùng viết.
4. **Vị trí KHÔNG chú thích**: literal quá lớn cho i32 vẫn suy ra i64 **theo độ lớn**
   ([[bug-negative-literal-compare-o0]]) — ca đó không có chú thích nào để tôn trọng.
   Cũng vì vậy `let arr = [1, 5000000000]` (kiểu phần tử suy từ phần tử ĐẦU) vẫn giữ hành
   vi cũ (phần tử 2 bị cắt) — một khiếm khuyết RIÊNG, không do quy tắc này.

**Mã lỗi:** `E3030` (dải type errors E3000–E3999, docs/CONTRIBUTING.md). Chẩn đoán in
giá trị, kiểu, dải hợp lệ, đoạn mã nguồn + caret, và chỉ ra cách sửa (`300 as u8`). KHÔNG
in `file.ax:line:col`: `tree.src` của frontend là compilation unit đã NỐI (driver ghép
std/*.ax trước file người dùng) nên số dòng tuyệt đối sẽ chỉ sai file — snippet + caret là
phần chính xác được của format §8 hiện nay.

**Cài đặt & test:** `typecheck.ax::check_int_lit_range` (+ 7 call site);
`bin/t_intrange{let,assign,fieldasg,arg,gen,field,arr,ret,neg}.ax` = reject rows,
`bin/t_intrangeok.ax` = mọi biên hợp lệ + tiền lệ suy-theo-độ-lớn phải vẫn chạy (42).
Gate: A==B `78295509`, regression **607/607** ở default và `-O0`. Audit breakage: 0 hit trên
source của chính compiler (qua cả hai hop) và trên 833 file `.ax` khác trong repo.

## 7. Float codegen (BUG#33)

`lower_binary_expr` phải phát OP_FADD/FSUB/FMUL/FDIV khi kiểu kết quả là f32/f64 (thay vì
luôn OP_IADD…). OP_F* selector đã có (x86_selector.ax:899). Float const = fconst (đã có).

### 7.1 Bất biến AIR: opcode của HẰNG phải khớp LỚP KIỂU (thêm 2026-07-30)

"Float const = fconst" ở trên là ĐIỀU KIỆN, không phải mô tả: **một hằng mang type_id f32/f64
BẮT BUỘC là OP_FCONST**, vì OP_ICONST vật chất hoá bằng `mov imm` vào vreg lớp GPR còn lệnh
float đọc XMM (hw index alias, `R10 ≡ XMM10`) ⇒ giá trị đọc lại là **0.0 hoặc giá trị cũ còn
sót**. Đây chính là lỗ hổng đã khiến `let a: f64 = 3` cho 0.0: RFC 0005 adopt f64 cho literal
nguyên (đúng), nhưng `lower_int_lit` vẫn phát OP_ICONST với type_id đó — nghĩa là bất biến này
ĐÚNG-MÀ-KHÔNG-ĐƯỢC-KIỂM suốt thời gian tồn tại của bug.

Nay được **cưỡng chế** (CLAUDE.md §9): `air_builder.verify_air_const_types` chạy trên mọi hàm
khi rời AIR builder, panic nếu OP_ICONST mang type_id 9/10 (hoặc OP_FCONST mang một kiểu
nguyên xác định; type_id 0 = unset không bị bắt). Mọi chỗ phát hằng float phải đi qua **một**
điểm duy nhất `emit_float_const`. Chi tiết + calibration:
[[bug-int-literal-float-type-iconst]].

## 8. Test plan

`tests/arith/` (oracle Python, bit-exact): mode `int` / `float` / `mixed` (cast) /
`imix` (promote ngầm). ≥100k biểu thức/seed. Thêm `.diag` test cho các ca PHẢI lỗi
(float→int, f64→f32 thiếu `as`). Mỗi phần implement: chạy harness + giữ self-host fixpoint
(compiler integer-only nên thay đổi không được đổi codegen path đang dùng — verify).

## 9. Thứ tự implement (đề xuất)
1. Float codegen (OP_F* theo kiểu) + int→float/f64→f32 cvt thật — làm `float`+`mixed`+`imix` pass.
2. Unsigned div/mod (DIV) + shift signed-aware — làm `int` pass hết.
3. Conversion policy diagnostics (float→int, f64→f32 → lỗi) + float-literal-adopt (mở rộng RFC 0005).
4. int→int narrowing policy (quyết định cuối). — **ca literal trần đã chốt & implement:
   §6.1 (REJECT, E3030, 2026-07-30)**; còn lại: biểu thức hằng gấp, và narrowing từ một
   GIÁ TRỊ runtime (`let x: u8 = some_i64`) vẫn chưa quyết.
