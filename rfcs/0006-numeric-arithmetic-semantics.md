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

## 7. Float codegen (BUG#33)

`lower_binary_expr` phải phát OP_FADD/FSUB/FMUL/FDIV khi kiểu kết quả là f32/f64 (thay vì
luôn OP_IADD…). OP_F* selector đã có (x86_selector.ax:899). Float const = fconst (đã có).

## 8. Test plan

`tests/arith/` (oracle Python, bit-exact): mode `int` / `float` / `mixed` (cast) /
`imix` (promote ngầm). ≥100k biểu thức/seed. Thêm `.diag` test cho các ca PHẢI lỗi
(float→int, f64→f32 thiếu `as`). Mỗi phần implement: chạy harness + giữ self-host fixpoint
(compiler integer-only nên thay đổi không được đổi codegen path đang dùng — verify).

## 9. Thứ tự implement (đề xuất)
1. Float codegen (OP_F* theo kiểu) + int→float/f64→f32 cvt thật — làm `float`+`mixed`+`imix` pass.
2. Unsigned div/mod (DIV) + shift signed-aware — làm `int` pass hết.
3. Conversion policy diagnostics (float→int, f64→f32 → lỗi) + float-literal-adopt (mở rộng RFC 0005).
4. int→int narrowing policy (quyết định cuối).
