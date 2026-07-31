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
| float → int (vd f32→i32) | **CẤM ngầm → LỖI**; phải `as` (truncate toward zero) — ✅ cưỡng chế ở **§6.2** (`E3031`) |
| f32 → f64 (widen) | **ngầm OK** (cvtss2sd) — ✅ cưỡng chế ở **§7.3** (convert thật ở MỌI value site) |
| f64 → f32 (narrow, lossy) | **CẤM ngầm → LỖI**; phải `as` (cvtsd2ss) — ✅ cưỡng chế ở **§6.3** (`E3032`) |
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
| **đối số của PHƯƠNG THỨC** (2026-07-31) | `s.setv(300)` với `setv(self, x: u8)` | wrap 44 |
| **đối số ở dạng gọi TĨNH** (2026-07-31) | `S.setv(&s, 300)` | wrap 44 |
| **đối số qua DYNAMIC DISPATCH** (2026-07-31) | `let i: Sink = s` … `i.take(300)` | wrap 44 |

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
3b. ~~**ĐỐI SỐ của lời gọi PHƯƠNG THỨC**~~ → **ĐÃ PHỦ 2026-07-31** (ba hàng cuối bảng trên).
   Đây là **một lỗ hổng CHUNG của cả ba luật**, không phải thiếu sót riêng của luật nào: phân
   giải method đi qua khối quét `mfi.params` (`typecheck.ax`, khối kiểm arity của nhánh callee
   `NODE_FIELD_EXPR`), còn dispatch động đọc từ `interface_method_sig` — cả hai đều **không**
   dùng `fp_data`. Vá bằng **MỘT** hook (`check_method_args_annotated`) gọi đúng
   `check_annotated_target`, nên cả ba luật bật cùng lúc ở vị trí này và không thể phủ lệch
   nhau. Vẫn chưa phủ: **gán vào PHẦN TỬ mảng** (`a[0] = 300`) — cố ý, vì kiểu phần tử có thể
   do suy diễn (`mut a = [1,2,3]` là `[i32;3]`), không phải kiểu người dùng viết.
4. **Vị trí KHÔNG chú thích**: literal quá lớn cho i32 vẫn suy ra i64 **theo độ lớn**
   ([[bug-negative-literal-compare-o0]]) — ca đó không có chú thích nào để tôn trọng.
   Cũng vì vậy `let arr = [1, 5000000000]` (kiểu phần tử suy từ phần tử ĐẦU) vẫn giữ hành
   vi cũ (phần tử 2 bị cắt) — một khiếm khuyết RIÊNG, không do quy tắc này.

**Mã lỗi:** `E3030` (dải type errors E3000–E3999, docs/CONTRIBUTING.md). Chẩn đoán in
giá trị, kiểu, dải hợp lệ, đoạn mã nguồn + caret, và chỉ ra cách sửa (`300 as u8`). KHÔNG
in `file.ax:line:col`: `tree.src` của frontend là compilation unit đã NỐI (driver ghép
std/*.ax trước file người dùng) nên số dòng tuyệt đối sẽ chỉ sai file — snippet + caret là
phần chính xác được của format §8 hiện nay.

**Cài đặt & test:** `typecheck.ax::check_int_lit_range`, gọi qua entry point chung
`check_annotated_target` (**8 call site** sau 2026-07-31 — 7 vị trí trực tiếp + 1 nằm trong
`check_method_args_annotated`, bản thân helper này được gọi từ 2 chỗ: khối quét method cụ thể
và site dispatch động; dùng chung với §6.2/§6.3);
`bin/t_intrange{let,assign,fieldasg,arg,gen,field,arr,ret,neg}.ax` = reject rows,
`bin/t_intrangeok.ax` = mọi biên hợp lệ + tiền lệ suy-theo-độ-lớn phải vẫn chạy (42).
Gate: A==B `78295509`, regression **607/607** ở default và `-O0`. Audit breakage: 0 hit trên
source của chính compiler (qua cả hai hop) và trên 833 file `.ax` khác trong repo.
Ba hàng METHOD/STATIC/DISPATCH thêm 2026-07-31: `bin/t_metharg{range,static,iface}.ax` = reject
rows (cả ba BUILD và trả **44** trên compiler trước hook), `bin/t_methargok.ax` = overreach
guard (42, trước và sau, ở `-O0` và `-O1`). Gate: A==B `9D8C7D68`, regression **649/649** ở
default và `-O0`.

### 6.2 REQUIREMENT — float → int NGẦM ⇒ LỖI (thi hành §4, user chốt 2026-06-26) ✅ IMPLEMENTED

Đây là **ảnh gương của §6.1** và là nửa còn lại của CÙNG chính sách conversion ở §4
(int→float ngầm OK / float→int **CẤM ngầm, phải `as`**). Quyết định đã chốt từ 2026-06-26
(`knowledge/bugs.md:1015`, BUG#35): "`let a: i32 = 3.0` → LỖI". Phần này chỉ cưỡng chế.

**Quy tắc (bắt buộc):** một **giá trị float** viết tại vị trí mà lập trình viên đã
**CHÚ THÍCH TƯỜNG MINH** một kiểu **nguyên** (i8/i16/i32/i64/isize/u8/u16/u32/u64/usize)
⇒ **`error[E3031]`** + `diags_count++` (driver dừng trước codegen). Muốn cắt phần lẻ thì
viết `3.0 as i64` (vẫn hợp lệ, không đổi). **Chiều ngược lại KHÔNG đổi:** int → float vẫn
ngầm hợp lệ (§4), `let x: f64 = 3` phải tiếp tục biên dịch (76de988, `t_intlitfloatctx`).

Lý do REJECT: giống §6.1, đây là lựa chọn DUY NHẤT không IM LẶNG. Trước thay đổi này
`let a: i64 = 3.0` **biên dịch được** và `a` giữ **nguyên mẫu bit IEEE** của 3.0
(`0x4008000000000000` = 4611686018427387904) — nên `a > 4000000000` là ĐÚNG. Một `i64`
thật ra là một `f64`, không một chẩn đoán nào (đo bằng `bin/probe4/g13.ax`, exit 3). Đúng
dạng accept-then-miscompile mà quy ước BUG#53 coi là kết cục tệ nhất.

**Nguồn float được nhận diện (chỉ những dạng CÚ PHÁP chắc chắn):** float literal (kể cả
khi bọc trong unary `-`), và **identifier có khai báo `let`/`const` ghi rõ `f32`/`f64`**.
KHÔNG suy từ `infer_node`: tại mọi call site biểu thức đã được infer VỚI kiểu nguyên đích
làm hint, nên kiểu suy ra báo lại chính ĐÍCH chứ không phải nguồn — cùng lý do các reject
str/numeric kế bên trong file này đều gate theo node kind. `3.0 as i64` là node kind khác
nên không bao giờ chạm tới check này.

**Vị trí áp dụng** (dùng CHUNG một call-site list với §6.1 —
`typecheck.ax::check_annotated_target` — nên hai luật không thể trôi lệch nhau):

| Vị trí | Ví dụ bị từ chối | Hành vi CŨ (đã đo) |
|---|---|---|
| `let`/`const` có chú thích | `let a: i64 = 3.0` | giữ **bit IEEE** (a > 4e9) |
| `let`/`const` **global** | `let G: i64 = 3.0` | như trên |
| literal float ÂM | `let a: i64 = -2.5` | như trên |
| **IDENT khai báo f32/f64** | `let f: f64 = 3.0` … `let a: i64 = f` | như trên |
| gán vào binding CÓ chú thích | `mut x: i32 = 1; x = 2.5` | nhận |
| gán vào **field struct** | `s.v = 2.5` | nhận |
| đối số của tham số nguyên | `fn takes(v: i32)` … `takes(4.5)` | nhận |
| đối số dạng **type-arg tường minh** | `pick[i64](3.0, 1)` | nhận |
| khởi tạo field struct | `P(a: 3.5)` | nhận |
| phần tử mảng có kiểu phần tử chú thích | `let a: [i32;3] = [1, 2.5, 3]` | nhận |
| `return` của hàm khai báo kiểu nguyên | `fn g() -> i32: return 3.0` | nhận |
| **đối số của PHƯƠNG THỨC** (2026-07-31) | `s.setv(3.0)` với `setv(self, x: i64)` | nhận, **rác** ⬅ xem dưới |
| **đối số gọi TĨNH / DISPATCH** (2026-07-31) | `S.setv(&s, 3.0)` · `i.take(3.0)` | như trên |

**Ngoài phạm vi (đã ĐO từng ca 2026-07-31, ghi rõ để không bị hiểu là thiếu sót):**
1. **Biểu thức** `let a: i64 = f + 1.0` — vẫn nhận. Cần kiểu suy diễn đáng tin ở vị trí có
   hint; xem lý do "không dùng infer_node" ở trên.
2. **Kết quả CALL** `let a: i64 = getf()` (getf → f64) — vẫn nhận.
3. **IDENT là PARAM** `fn h(x: f64): let a: i64 = x` — vẫn nhận. Kiểu ghi của param có thể
   là tên generic, và `type_id` của symbol bị ghi đè theo từng bản monomorphise (`fn id[T]
   (v: T)` khởi tạo ở f64 rồi i64 sẽ khiến bản i64 bị TỪ CHỐI OAN) ⇒ cố ý không đọc.
4. **Đọc field** `let a: i64 = s.f` (S.f: f64) — vẫn nhận.
5. **Gán vào binding SUY DIỄN** `mut x := 0; x = 3.0` — vẫn nhận (dùng chung annotation
   gate với §6.1) và **gán vào PHẦN TỬ mảng** `a[0] = 3.0` — cùng ranh giới §6.1 vẽ.
6. ~~**Đối số của PHƯƠNG THỨC** `s.setv(3.0)`~~ → **ĐÃ PHỦ 2026-07-31**, cùng một hook với
   §6.1 mục 3b và §6.3 mục 1. ⚠️ Đo lại kỹ hơn khi đóng: hành vi CŨ **không** phải là giữ
   mẫu bit IEEE như ca `let` — literal float đi vào **XMM**, còn tham số `i64` được đọc từ
   **thanh ghi nguyên chưa khởi tạo**, nên field nhận một số nguyên dương RÁC (không phải 3,
   không phải 0, không > 4e9; `bin/probe8/m2.ax` exit **16**, `m2b.ax` phân loại). Tệ hơn ca
   `let`: giá trị còn không tất định theo ngữ cảnh gọi.
7. ~~**f64 → f32 (narrowing)** — hàng KHÁC của bảng §4, chưa cưỡng chế~~ → **ĐÃ LÀM: §6.3
   (`E3032`, 2026-07-31)**, dùng chung `check_annotated_target` nên phủ đúng 10 vị trí như
   E3030/E3031. Chiều widen f32 → f64 (hợp lệ) được sửa ở **§7.3**.
8. ⚠️ Phát hiện phụ (KHÔNG do luật này, đã hỏng từ trước — `bin/probe6/q1.ax`): chiều
   int→float NGẦM còn thiếu ở hai vị trí — đối số `takes_f64(9)` và khởi tạo field
   `Mixed(f: 5)` không chèn cvt, f64 đọc ra rác. Là khiếm khuyết RIÊNG, cần fix riêng.

**Mã lỗi:** `E3031` (kế tiếp E3030 trong dải type errors E3000–E3999). Chẩn đoán in giá
trị, kiểu NGUỒN, kiểu ĐÍCH, snippet + caret, và cách sửa (`3.0 as i64`). Dùng lại đúng bộ
render của E3030 (`print_annotated_expr_snippet`) nên hai chẩn đoán không thể lệch hình
dạng. KHÔNG in `file.ax:line:col`, cùng lý do §6.1. Chuỗi in ra **thuần ASCII** (console
Windows làm hỏng UTF-8 — `§`/em-dash ra mojibake; ký tự non-ASCII chỉ nằm trong COMMENT).

**Cài đặt & test:** `typecheck.ax::{check_float_to_int, ident_declared_float,
is_int_family_type, float_type_display}` + entry point chung `check_annotated_target`
(**8 call site** sau 2026-07-31 — xem §6.1; dùng chung với `check_int_lit_range`);
`bin/t_f2i{let,var,assign,fieldasg,arg,gen,field,arr,ret,neg}.ax` = reject rows (cả 10 đều
BUILD trên compiler trước fix), `bin/t_f2iok.ax` = `3.0 as i64` + int→float widening +
float→float + mọi vị trí nguyên phải vẫn chạy (42), pass cả TRƯỚC và SAU.
Gate: A==B `407E0805`, regression **630/630** ở default và `-O0`. Audit breakage: 0 hit
trên source của chính compiler (qua cả hai hop) và trên 190 file `.ax` có float trong repo
(chỉ 10 oracle reject cố ý).

### 6.3 REQUIREMENT — f64 → f32 NGẦM ⇒ LỖI (thi hành §4, hàng 3) ✅ IMPLEMENTED

Hàng thứ BA của **cùng** bảng §4, cưỡng chế từ **cùng** danh sách vị trí với §6.1/§6.2.
§4 dòng 42: "f64 → f32 (narrow, lossy) | **CẤM ngầm → LỖI**; phải `as` (cvtsd2ss)"; §4 quy
tắc cài: "target f32 & expr f64 → **LỖI**. target f64 & expr f32 → chèn cvt widen." Chiều
widen là HỢP LỆ và được sửa riêng ở §7.3 — hai nửa của một hàng, làm cùng lượt.

**Quy tắc:** một **giá trị f64** viết tại vị trí mà lập trình viên đã **CHÚ THÍCH TƯỜNG
MINH** `f32` ⇒ **`error[E3032]`** + `diags_count++`. Muốn chấp nhận mất chính xác thì viết
`d as f32` (không đổi).

Lý do REJECT: giống §6.1/§6.2 — lựa chọn DUY NHẤT không IM LẶNG. Trước thay đổi này
`let d: f64 = 3.5; let s: f32 = d` **biên dịch được** và `s` **không phải 3.5**: mẫu bit f64
bị `OP_COPY` vào ô f32 nên nửa thấp của một double được đọc như một single
(đo: `bin/probe6/g_f64tof32.ax` exit 1). Accept-then-miscompile, lớp BUG#53.

**Nguồn f64 được nhận diện (chỉ dạng CÚ PHÁP chắc chắn):** **identifier có khai báo
`let`/`const` ghi rõ `f64`** (kể cả khi bọc trong unary `-`).
⚠️ **Float LITERAL KHÔNG bị bắt, có chủ ý**: RFC 0005 (mở rộng cho float literal ở §2) nói
literal ADOPT kiểu ngữ cảnh, nên `let a: f32 = 1.5` là literal f32 thật, không phải f64 bị
thu hẹp. Đọc từ **văn bản chú thích đã viết** (`ident_declared_float`), KHÔNG từ `type_id`
của symbol — `type_id` bị ghi đè theo từng bản monomorphise nên `idf[f32](13.5)` sẽ bị TỪ
CHỐI OAN (đúng cái bẫy §6.2 mục 3 đã ghi; `t_f64f32ok` ghim cả hai bản f64/f32).

**Vị trí áp dụng** (10 vị trí, dùng CHUNG `check_annotated_target` với §6.1/§6.2). Cột cuối
là giá trị ĐO ĐƯỢC trên compiler trước fix cho `let d: f64 = 3.5` … `return X as i64`
(đúng phải là **3**):

| Vị trí | Ví dụ bị từ chối | Hành vi CŨ (đã đo) |
|---|---|---|
| `let`/`const` có chú thích | `let s: f32 = d` | biên dịch, **0** |
| `let` **global** | `let S: f32 = D` | biên dịch, **0** |
| gán vào binding CÓ chú thích | `mut x: f32 = 1.0; x = d` | biên dịch, **0** |
| gán vào **field struct** | `s.v = d` | biên dịch, **0** |
| đối số của tham số `f32` | `fn takes(v: f32)` … `takes(d)` | biên dịch, **3** ⬅ xem dưới |
| đối số dạng **type-arg tường minh** | `pick[f32](d, 1.0)` | biên dịch, **0** |
| khởi tạo field struct | `P(a: d)` | biên dịch, **0** |
| phần tử mảng có kiểu phần tử chú thích | `let a: [f32;3] = [1.0, d, 3.0]` | biên dịch, **0** |
| `return` của hàm khai báo `-> f32` | `fn g() -> f32: return d` | biên dịch, **0** |
| IDENT f64 bị NEGATE | `let s: f32 = -d` | biên dịch, **0** |
| **đối số của PHƯƠNG THỨC** (2026-07-31) | `s.setv(d)` với `setv(self, x: f32)` | biên dịch, **3** |
| **đối số gọi TĨNH / DISPATCH** (2026-07-31) | `S.setv(&s, d)` · `i.take(d)` | biên dịch, **3** |

⬅ Hàng ĐỐI SỐ trả **3** (đúng) trong khi chín hàng kia trả **0** — đó chính là bằng chứng
đo được rằng *đối số là vị trí DUY NHẤT* có chuyển đổi bề rộng float, vì
`coerce_float_arg_ft` giữ một BẢN SAO RIÊNG của luật. Đúng hình dạng "một luật tồn tại ở
dạng hẹp hơn thứ nó phải phủ" mà §7.2 đã mô tả; §7.3 gộp bản sao đó vào luật chung.

**Ngoài phạm vi (đã ĐO từng ca 2026-07-31 trên compiler SAU fix, ghi rõ để không bị hiểu là
thiếu sót). Lưu ý quan trọng: nhờ §7.3 mọi ca dưới đây nay cho GIÁ TRỊ ĐÚNG (3) thay vì rác
(0) — chúng là "nhận mà không chẩn đoán", KHÔNG còn là miscompile:**
1. ~~**Đối số của PHƯƠNG THỨC** `s.setv(d)`~~ → **ĐÃ PHỦ 2026-07-31** bằng **một** hook chung
   cho cả ba luật (`check_method_args_annotated` → `check_annotated_target`), đúng như mục
   này dự đoán. Ở luật NÀY đó là ca "nhận mà không chẩn đoán" (giá trị đã đúng, 3, nhờ §7.3),
   trong khi §6.1/§6.2 ở cùng vị trí là miscompile thật — một lý do nữa để không tách ba hook.
2. **Gán vào PHẦN TỬ mảng** `a[0] = d` — vẫn nhận (0 → 3). Cùng ranh giới §6.1 vẽ.
3. **Biểu thức** `let s: f32 = d + 1.0` — vẫn nhận (0 → 4). Cần kiểu suy diễn đáng tin ở vị
   trí có hint; xem lý do "không dùng infer_node" ở §6.2.
4. **Kết quả CALL** `let s: f32 = getd()` — vẫn nhận (0 → 3).
5. **IDENT là PARAM** `fn h(x: f64): let s: f32 = x` — vẫn nhận (0 → 3), CỐ Ý: kiểu ghi của
   param có thể là tên generic (§6.2 mục 3).
6. **Đọc field** `let s: f32 = o.d` — vẫn nhận (0 → 3).
7. **Gán vào binding SUY DIỄN** `mut x := 1.5 as f32; x = d` — vẫn nhận (0 → 3).

**Mã lỗi:** `E3032` (kế tiếp E3031 trong dải type errors E3000–E3999). In tên nguồn, kiểu
nguồn `f64`, kiểu đích `f32`, snippet + caret, và cách sửa (`d as f32`). Dùng lại đúng bộ
render của E3030/E3031 (`print_annotated_expr_snippet`) nên ba chẩn đoán không thể lệch hình
dạng. KHÔNG in `file.ax:line:col`, cùng lý do §6.1. Chuỗi in ra **thuần ASCII**.

Hình dạng in ra (đo, không phải phác thảo):

```
error[E3032]: cannot implicitly convert `d` (type `f64`) to the narrower float type `f32`
   |
   |     let s: f32 = d
   |                  ^ f64 value here; `f32` has less precision and a smaller range
   |
   = note: the type `f32` is written explicitly at this `let` binding, and AXIOM never narrows f64 to f32 implicitly (RFC 0006 section 4) -- the loss of precision must be visible in the source
   = help: write `d as f32` if the loss of precision is intended
```

**Cài đặt & test:** `typecheck.ax::check_f64_to_f32` + entry point chung
`check_annotated_target` (**8 call site** sau 2026-07-31 — xem §6.1; dùng chung với
`check_int_lit_range` và
`check_float_to_int`); `bin/t_f64f32{let,global,assign,fieldasg,arg,gen,field,arr,ret,neg}.ax`
= reject rows (cả 10 đều BUILD trên compiler trước fix), `bin/t_f64f32ok.ax` = overreach guard
(42 cả TRƯỚC và SAU, ở -O0 và -O1). Audit breakage: quét 120 file `.ax` có `f32` trong
`bin/ tests/ examples/ std/ stdlib/ bootstrap/ docs/ rfcs/` bằng compiler sau fix — chỉ 12
hit, đúng 10 oracle reject + 2 probe cố ý (`probe6/g_f64tof32`, `probe7/w2`); 0 hit trên
source của chính compiler (hop B build sạch).

### 6.4 ĐỐI SỐ của PHƯƠNG THỨC — vị trí chung của cả ba luật (2026-07-31) ✅ IMPLEMENTED

Không phải luật thứ tư: là **VỊ TRÍ thứ mười một** của ba luật §6.1/§6.2/§6.3, và là lỗ hổng
duy nhất mà cả ba cùng có. Ghi riêng vì **cơ chế phân giải** ở đây khác hẳn mọi vị trí khác.

**Vì sao nó lọt:** bảy call site cũ của `check_annotated_target` đều lấy kiểu đích từ đường
`fp_data` (kiểu khai báo của một hàm tự do). Lời gọi phương thức **không đi đường đó**:

- `s.m(a)` / `S.m(&s, a)` — kiểu tham số đến từ **khối quét symbol theo tên + kiểu receiver**
  trong nhánh callee `NODE_FIELD_EXPR` của `infer_node` (khối kiểm arity, đọc `mfi.params`);
- `i.m(a)` (dispatch động) — kiểu tham số đến từ **`interface_method_sig`**, đọc thẳng AST
  của khai báo interface.

⇒ **MỘT** hook (`check_method_args_annotated`) gọi đúng `check_annotated_target`, dùng ở CẢ HAI
chỗ. Không thêm luật, không nhân bản luật — đây chính là lớp defect (hai bản sao của một luật
trôi lệch) đã sinh ra tám bug trong repo này, gần nhất là bản sao riêng của luật độ-rộng-float
trong `coerce_float_arg_ft` (§6.3, `b8ac125`).

**Cái bẫy monomorphise — đã ĐO, không đoán.** §6.2 mục 3 / §6.3 ghi rõ: `type_id` của một
symbol bị **ghi đè theo từng bản monomorphise**, nên đọc nó ở phía NGUỒN sẽ từ chối oan bản
i64 của `fn id[T](v: T)`. Ở đây kiểu đến từ một symbol scan, nên phải trả lời cho phía ĐÍCH.
Kết luận: **là kiểu KHAI BÁO**, vì ba lý do độc lập, mỗi lý do đủ một mình:

1. khối quét **bỏ qua** symbol có `SYM_FLAG_GENERIC`;
2. một bản monomorphise là **symbol RIÊNG** với tên **đã mangle** (`f__i64`), mà
   `name_matches_method` chỉ khớp tên chính xác hoặc dạng `Type.name` ⇒ không bao giờ khớp;
3. `pre_infer_func_signature` đăng ký func type **một lần cho mỗi symbol** (guard
   `has_existing`), và `register_function` luôn **push entry mới** ⇒ entry không bị ghi đè
   theo từng lần khởi tạo.

Chứng cứ đo được (`bin/t_methargok.ax`, hàng 40–44 và 30–31, chạy 42 ở `-O0` lẫn `-O1`, trước
và sau): **một** method generic khởi tạo ở **i64, f64, u8** rồi **quay lại i64** — không hàng
nào bị từ chối oan; và **hai struct có method TRÙNG TÊN, kiểu tham số KHÁC nhau**
(`Wide.setv(x: i64)` nhận 300, `Narrow.setv(x: u8)` từ chối 300) ⇒ kiểu đọc ra đúng là của
khai báo thuộc receiver đó. Với nhiều hơn một method khớp (overload cùng receiver + arity),
hook **không kích hoạt** — thà bỏ sót còn hơn từ chối oan.

**Đã phủ:** instance `s.m(a)` · static `S.m(&s, a)` · dispatch `i.m(a)` · method khai báo
**inline** trong thân struct · receiver là **enum/sum** (`bin/probe8/{c5,c6,c7,c8}.ax`).

**Chưa phủ (đo, ghi rõ):** receiver là builtin/generic-inst (`Vec`/`Option`/`HashMap` — khối
quét chỉ nhận receiver STRUCT/SUM của người dùng, cố ý để không đụng method bị backend chặn);
lời gọi method **generic** (đi nhánh `is_generic_call`, tham số là tên generic nên cả ba luật
đằng nào cũng trơ); và mọi vị trí §6.2/§6.3 vốn đã ngoài phạm vi (biểu thức, kết quả call,
đọc field, ident là param, `a[0] = …`).

**Test:** `bin/t_metharg{range,f2i,f64f32,static,iface}.ax` = reject rows — **cả năm BUILD**
trên compiler trước hook và trả giá trị sai âm thầm (44 / rác dương / 3 / 44 / 44);
`bin/t_methargok.ax` = overreach guard (42 trước và sau, `-O0` và `-O1`).
Gate: A==B `9D8C7D68`, regression **649/649** ở default và `-O0`. Audit breakage: quét **1193**
file `.ax` trong `bin/ tests/ examples/ std/ stdlib/ bootstrap/ docs/ rfcs/` bằng cả hai
compiler và **diff** danh sách bị từ chối — mới thêm **đúng 16 file**, toàn bộ là oracle/probe
của chính thay đổi này cộng hai probe có sẵn được viết ra để GHI LẠI lỗ hổng
(`bin/probe6/g_method.ax`, `bin/probe7/gap_method.ax`); 0 hit trong `std/ stdlib/ tests/
examples/ bootstrap/`. Riêng 31 chương trình có `interface` được chạy đối chứng cũ/mới ở `-O1`
(vì site dispatch được đổi sang gọi `interface_method_sig` một lần cho cả ret lẫn params):
mọi exit code trùng khớp.

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

### 7.2 int → float NGẦM: chèn ở TẦNG LOWERING, một quy tắc cho MỌI value site (2026-07-31) ✅ IMPLEMENTED

§4 nói int→float ngầm OK và "phải là lệnh convert thật". §4 còn đề xuất chèn ở **typecheck**.
Thực tế cài **khác** — và đây là một sai lệch có chủ ý, ghi lại theo CLAUDE.md §20:

- Typecheck chỉ có thể *adopt* kiểu cho **literal** (RFC 0005). Với một **BIẾN** nguyên
  (`let a: f64 = n`) hay một **BIỂU THỨC** nguyên (`let c: f64 = 3 + 1`), storage đã được
  dựng ở bề rộng của chính nó; một "expected type hint" không dựng lại nó. Nên hint không
  bao giờ đủ.
- Ngược lại, air_builder ĐÃ sở hữu chuyển đổi int↔float cho `as` (`lower_cast_expr`) và cho
  toán hạng của phép toán float (`lower_binary_expr`). Đặt quy tắc ngầm ở cùng tầng giữ
  đúng một chỗ phát OP_ITOF, thay vì hai cơ chế song song sẽ trôi lệch.

Cài đặt: `air_builder.coerce_to_float_target(target_type, src_node, src_reg)` (2026-07-31 đổi
tên từ `coerce_int_to_float` khi §7.3 gộp thêm nhánh float→float — tên cũ đã thành sai) — anh
em song sinh của `coerce_struct_to_interface`, tự-bảo-vệ (no-op trừ khi target là f32/f64 **và**
kiểu node nguồn là kiểu nguyên XÁC ĐỊNH), và được gọi ở **MỌI** value site:
let / assign / phần tử mảng (literal + gán) / field struct (khởi tạo + gán) / return /
đối số lời gọi (free fn, method, dynamic dispatch) / khởi tạo global.

Vì sao liệt kê đủ: lỗ hổng vừa vá chính là *một quy tắc tồn tại ở dạng hẹp hơn thứ nó phải
phủ* — coercion cũ chỉ chạy cho **f32** ở **hai** trong số các site đó (comment tự khai:
"Scoped to F32 only … the self-host fixpoint is unaffected"), nên f32 đúng còn f64 đọc ra rác:
`H(a: 3)`, `takes_f64(9)`, `mm.m(3)`, `let c: f64 = 3 + 1`, và toàn bộ nhánh biến nguyên
(`let a: f64 = n`, `b = n`, `argf(n)`). Đây là accept-then-miscompile, lớp BUG#53.

Ghi chú dấu: OP_ITOF = cvtsi2sd (có dấu), nên u64/usize > 2^63 chuyển ra số âm. Đó **đã là**
hành vi của `x as f64` (cùng một phép chọn opcode); làm nhánh ngầm khớp nhánh tường minh là
bước tối thiểu đúng — chuyển đổi unsigned chính xác là **một** fix cho cả hai site, không phải
lý do để một site tiếp tục miscompile im lặng.

**Bất biến §9 kèm theo** (`air_builder.verify_air_no_int_into_float`): một giá trị lớp NGUYÊN
không được tới một consumer lớp FLOAT mà thiếu OP_ITOF. Miền trừu tượng cực nhỏ và một chiều
(INT/FLOAT/UNKNOWN; `type_id == 0` không bao giờ thành INT), consumer bị bắt = OP_COPY kiểu
float và OP_FADD/FSUB/FMUL/FDIV. **Giới hạn đã biết, nói thẳng để không ai đọc sự im lặng
thành bằng chứng**: site ĐỐI SỐ và site FIELD *không* được phủ, vì AIR không mang kiểu tham số
trên OP_CALL, cũng không mang kiểu field trên OP_SET_FIELD (type_id ở đó là kiểu STRUCT); hai
site này do oracle canh. Mở rộng AIR để mang các kiểu đó là thay đổi IR ⇒ cần RFC riêng.

Calibrate (không phải giả định): stub `coerce_to_float_target` thành no-op thì check báo đúng
`main / inst #5 … type_id=10 src1=5` cho `let a: f64 = n`, `#22` cho `b = n`, `#50` cho
`let c: f64 = 3 + 1`, và **không** sinh file output. Với fix, check im lặng trên cả 1053 hàm
của chính compiler. Oracle `bin/t_intlitfloatctx.ax`: **4 → 42** ở -O0/-O1/-O2.

### 7.3 f32 → f64 NGẦM: cvt THẬT ở MỌI value site, không chỉ ở đối số (2026-07-31) ✅ IMPLEMENTED

Nửa HỢP LỆ của hàng §4 thứ ba (nửa kia là §6.3, reject chiều narrow). §4 dòng 47 nói
"target f64 & expr f32 → **chèn cvt widen**", dòng 48 nói mọi cvt `f64↔f32` phải là **lệnh
convert thật, KHÔNG phải OP_COPY reinterpret**. Cả hai đều bị vi phạm ở mọi vị trí trừ đối số.

Đo được (không phải suy luận): `let f: f32 = 1.5; let x: f64 = f` sinh AIR

```
%2: t9  = copy %1
%3: t10 = copy %2      <- bit của một single đọc như nửa thấp của một double
```

nên `x` là denormal chứ không phải 1.5 (`bin/probe7/w1.ax` exit **1**). Sau fix AIR có
`%4 = cast %2` và máy sinh đúng **một** lệnh `cvtss2sd %xmm8,%xmm9` (objdump, `0x140006579`),
`w1` exit **42**.

Nguyên nhân gốc **giống hệt** §7.2: luật tồn tại ở dạng hẹp hơn thứ nó phải phủ.
`coerce_float_arg_ft` giữ BẢN SAO RIÊNG của chuyển đổi bề rộng float (một `OP_CAST` phát tại
chỗ), nên **đối số** là vị trí DUY NHẤT chuyển đúng, còn let/assign/field/element/return/global
thì reinterpret. Bảng §6.3 đo được đúng vân tay này: hàng đối số trả **3**, chín hàng kia trả
**0**.

Cài đặt (tối thiểu, đúng hình dạng chống-trôi của §7.2): nhánh float→float chuyển vào **cùng**
hàm `coerce_to_float_target` đã được gọi ở MỌI value site, và `coerce_float_arg_ft` nay chỉ còn
việc TÌM kiểu tham số rồi ủy quyền. Một cài đặt, một danh sách vị trí.

```
int  -> f32/f64  (§4 hàng 1)  -> OP_ITOF  (cvtsi2sd / cvtsi2ss)
f32  -> f64      (§4 hàng 3)  -> OP_CAST  (cvtss2sd, chọn ở x86_selector theo kiểu vreg nguồn)
```

Hệ quả PHỤ đáng ghi (đo trên compiler sau fix): những vị trí mà §6.3 **không** reject (đối số
phương thức, `a[0] = d`, biểu thức, kết quả call, param ident, đọc field, binding suy diễn) nay
phát `cvtsd2ss` thật nên cho **giá trị đúng (3)** thay vì rác (**0**). Chúng chuyển từ
*accept-then-miscompile* sang *accept-without-diagnostic* — vẫn là lỗ hổng chẩn đoán, nhưng
không còn là lỗi sinh mã. Nói thẳng ra để không ai đọc "3" thành "đã phủ".

**Chưa có bất biến §9 cho bề rộng float.** `verify_air_no_int_into_float` chỉ phân lớp
INT/FLOAT/UNKNOWN, không phân biệt f32 với f64, nên một `OP_COPY` kiểu f64 đọc vreg f32 vẫn
KHÔNG bị bắt — chính là bug này. Mở rộng miền trừu tượng sang bề rộng là việc riêng (backlog);
hiện tại `bin/t_f32widen.ax` là thứ canh, và nó FAIL (exit 1) trên compiler trước fix — đó là
calibration, không phải giả định.

Gate: thay đổi chạm air_builder ⇒ **B == C**. Đo được A == B == C =
`D3EABC61F38B7A0BEA6E58598660D0475D7CA77ACCC832EC0FDF987940085CD4` (A == B vì source của chính
compiler không có luồng f32↔f64 ngầm nào ⇒ thay đổi trơ với self-codegen). Regression
**642/642** ở default và `-O0` (630 baseline + 12 hàng mới).

## 8. Test plan

`tests/arith/` (oracle Python, bit-exact): mode `int` / `float` / `mixed` (cast) /
`imix` (promote ngầm). ≥100k biểu thức/seed. Thêm `.diag` test cho các ca PHẢI lỗi
(float→int, f64→f32 thiếu `as`). Mỗi phần implement: chạy harness + giữ self-host fixpoint
(compiler integer-only nên thay đổi không được đổi codegen path đang dùng — verify).

## 9. Thứ tự implement (đề xuất)
1. Float codegen (OP_F* theo kiểu) + int→float/f64→f32 cvt thật — làm `float`+`mixed`+`imix` pass.
2. Unsigned div/mod (DIV) + shift signed-aware — làm `int` pass hết.
3. Conversion policy diagnostics (float→int, f64→f32 → lỗi) + float-literal-adopt (mở rộng RFC 0005).
   — **ĐÃ XONG cả hai chẩn đoán: §6.2 (E3031, float→int) và §6.3 (E3032, f64→f32), cùng nửa
   hợp lệ §7.3 (f32→f64 cvt thật)**. Còn lại của mục này: bộ `.diag` test có hệ thống trong
   `tests/arith/` (§8), và các vị trí ngoài phạm vi liệt kê ở §6.3.
4. int→int narrowing policy (quyết định cuối). — **ca literal trần đã chốt & implement:
   §6.1 (REJECT, E3030, 2026-07-30)**; còn lại: biểu thức hằng gấp, và narrowing từ một
   GIÁ TRỊ runtime (`let x: u8 = some_i64`) vẫn chưa quyết.
