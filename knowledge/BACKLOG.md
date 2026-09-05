# BACKLOG — orientation file (READ THIS, not MEMORY.md)

⚠️ **Đây là FILE CON TRỎ, không phải kho sự thật.** Mọi chi tiết/bằng chứng nằm trong
`knowledge/<topic>.md` và `knowledge/MEMORY.md`. Nếu hai file nói khác nhau, **`MEMORY.md` đúng** —
và hãy sửa file này ngay, đừng để trôi lệch (hai bản sao của một sự thật chính là lớp defect đã
sinh ra bug interface-return).

**Vì sao file này tồn tại:** `knowledge/MEMORY.md` = **175 KB ≈ 87k token**, vượt read cap; chỉ một
trang bị cắt của nó đã tốn ~25k token mỗi phiên **trước khi làm bất cứ việc gì**. Định hướng bằng
file này (~2k token) + handoff mới nhất; vào `MEMORY.md` **chỉ bằng `Grep`** theo tên topic.

---

## Trạng thái cây (cập nhật 2026-09-05 — sửa B3/B6)
- 🆕 **B3+B6 (một bug: `decl_node` xuyên cây trong `pre_infer_func_signature`) — A==B
  `F7146F3DBCDD3280E838B06538E946B33D2F3D19F32A18909BBAD81E22BC77E3`** (2.329.600 byte).
  Module nạp trễ gọi được stdlib đã bundle; `import std.{sync,thread,collections,string}` hết SEGV.
  Regression **711/711 ở CẢ default lẫn `-O0`** ⇒ **BASELINE MỚI = 711**.
  ⚠️ Frontend-only ⇒ chỉ cần A==B; **mốc B==C gần nhất vẫn là `c3eae77`/`52D1ABD4…`**.
  ⚠️ Bản vá **không** byte-identical với seed (code mới nằm trong self-image dù quá trình tự biên
  dịch không hề đi qua đường nạp trễ) — **A==B mới là tiêu chí**, đừng kỳ vọng byte-identical.
- Mốc trước: **Stage 1 stdlib-reachability + B2/B4/B5 — A==B
  `9C6726C11F366ACA5BA3970F72D0C0502C7495506F82AEBBCEF33A6C14C326E2`** (2.329.600 byte).
  `import std.math` + `std.math.sqrt(16.0)` chạy end-to-end. Regression **709/709 ở CẢ
  default lẫn `-O0`** ⇒ **BASELINE MỚI = 709**. Chi tiết ở §"HƯỚNG ĐÃ ĐỊNH GIÁ" bên dưới.
- Mốc trước: **RFC 0039 (struct literal suy diễn từ annotation) — A==B
  `84A13E958B59D2A1022C860C8E4637E81716BA65A66BFDDFD096034E4DB3FF68`** (2.313.728 byte).
  `let c: TmpStruct = (a: 64, b: 64)` nay hợp lệ; `error[E3034]` khi không có ngữ cảnh suy diễn.
  Tự kiểm chứng **685/685** ở cả hai mức + control tuple/biểu thức ngoặc.
- Mốc trước: sửa P6 `A58F762A…` (2.307.584 byte), tự kiểm chứng 682/682;
  dọn libc `0585124E…`; RFC 0038 (`ca7a98d`) `99F795C2…` 2.308.096 byte.
  ⚠️ Mốc **B==C** gần nhất vẫn là `c3eae77` /
  `52D1ABD4AE9E6EF11216AD3B8318D1592C1C03F383D49F5464B6ABF0A6C9478B` (2.297.856 byte) — cả RFC 0038
  lẫn dọn-libc đều frontend/std nên chỉ cần A==B; **thay đổi backend kế tiếp phải dựng lại B==C từ
  driver MỚI này**.
- **libc: compiler `ucrtbase.dll` 16 → 5** (`atof memcpy memset strlen system`); chương trình user = **3** (sàn).
  `QueryPerformanceCounter`/`Frequency` thêm vào **kernel32** (không phải libc).
  ⚠️ Đo bằng **parse bảng import PE**, KHÔNG bằng `strings` — xem
  [audit-libc-dependencies](audit-libc-dependencies-2026-08-05.md).
- **BASELINE = 709/709**, đo ở **cả default lẫn `-O0`**. Dưới 709 là RED. (672 → 677 RFC 0038, → 679 libc 0-2, → 682 P6, → 685 RFC 0039, → 689 chẩn đoán 0-1, → 697 chẩn đoán 2b, → 699 libc 3, → 700 srcmap module, → 703 libc 4+5, → 709 stage 1 stdlib reachability.)
  (611 → 649 → 662 → 672: +32 hàng ở `b8ac125`, +7 ở `f6ac69e`, +13 ở `a538983`, +10 khi đóng
  hole C — gồm một khối `-no-dfe` riêng, vì DFE che đúng cái defect đó.)
- `bin/axc_pre1f.exe` = compiler tham chiếu tiền-1f, giữ để định giá ghép cặp.
  `bin/probe5/axc_new.exe` (30/07 21:59) = mốc **trước** `6febd02`, hữu ích để quy trách nhiệm.
- Handoff mới nhất: [session-handoff-2026-07-30d](session-handoff-2026-07-30d.md) — ⚠️ phần
  "REMAINING QUEUE" của nó **đã lỗi thời**: bug #2, bug #3 và task 0 đều đã ship sau đó.

## ⛔ CẦN USER QUYẾT (không tự quyết — hạng D1)
*(Trống. Cả 7 mục treo đã được user quyết 2026-08-07 — xem ngay dưới. Đừng hỏi lại.)*

## ✅ ĐÃ QUYẾT (D1) — không hỏi lại

### Chốt 2026-08-07 (7 quyết định, user chọn trực tiếp)
1. **Gate M6 ⇒ SO PHÂN PHỐI: median trên N layout + LUÔN in spread cạnh con số.** Giữ fib trong
   gate. Ngưỡng phán quyết đọc trên median, không trên một lượt đo đơn lẻ. Hạ tầng đã có
   (`scripts/perf_layout_dist.ps1`, `scripts/perf_m6_gate.ps1`); giá phải trả là N lần build mỗi
   lần đo — chấp nhận. ⛔ **KHÔNG** ghim một layout tham chiếu (sẽ che chính hiệu ứng bimodal vừa
   đo được), **KHÔNG** loại fib (mất phủ sóng đệ quy/tail-call, đúng chỗ 1f trả lãi −14%).
   Thay thế mục "fib không phán quyết được" ở handoff 07-30c.
2. **`atof` ⇒ VIẾT PARSER FLOAT ĐÚNG CHUẨN BẰNG AXIOM.** RFC + hiện thực strtod đúng IEEE-754
   (số mũ, round-to-nearest-even, subnormal), kèm **test đối chiếu bit-exact với `atof`** trên
   corpus literal trước khi chuyển. Đây là con đường tới "AXIOM độc lập libc" thật.
   ⛔ **KHÔNG** dùng `std/string.ax:818` (thiếu số mũ, làm tròn sai ⇒ đổi ngữ nghĩa im lặng mọi
   float literal — §3 cấm). ⛔ **KHÔNG** giữ `atof` vĩnh viễn (một import kéo cả UCRT vào mọi binary).
3. **Rank 2/3 (RFC 0037) ⇒ KHOÁ VÀO DANH TÍNH, KHÔNG VÀO CÁCH VIẾT.** Monomorphizer đánh dấu
   tường minh symbol nào **THỰC SỰ là một instantiation** (cờ/bảng); rank 2/3 chỉ chấp nhận symbol
   mang dấu đó. Đóng `h1_rank2_eq` + `h2_axstd_prefix` **và** xoá cả lớp defect "khớp theo chính
   tả" (cùng lớp với P4). Chạm monomorphizer ⇒ đo A==B cẩn thận.
   ⛔ **KHÔNG** hạn chế rank 2 theo prefix `_AX_std_` (đã đo: làm hỏng method instantiated của
   struct generic — mangle đặt prefix TRƯỚC dấu chấm). ⛔ **KHÔNG** cấm user đặt tên chứa `__`/`_AX_`
   (lấy không gian tên của user để bù khuyết tật nội bộ = đổi chỗ nợ, không trả nợ).
4. **Debug output ⇒ CỜ `--verbose` THẬT, MẶC ĐỊNH IM LẶNG.** Cần RFC (đổi bề mặt CLI). Mọi output
   debug đi qua **một** cổng có cấp độ; mặc định compiler chỉ in chẩn đoán. Xoá luôn whitelist
   ~19 chuỗi `[D...` của `is_verbose_debug` (`print_helpers.ax:108-182`).
   ⛔ **KHÔNG** dùng env var `AXIOM_DEBUG` (kém khám phá được + trạng thái ngầm, §19).
5. **`str.len` ⇒ GIỮ NGHĨA SỐ BYTE, O(1)**, ghi RÕ vào spec; **bổ sung hàm đếm ký tự riêng**
   (`chars()` / `char_count()`). Tiền lệ Rust/Go; không đổi call site nào trong self-image.
   ⛔ **KHÔNG** đổi `len` thành số ký tự (đổi ngữ nghĩa hàm stdlib compiler đang dùng + `len` thành O(n)).
6. **P5 console ⇒ `SetConsoleOutputCP(65001)` lúc khởi động runtime, CHỈ KHI stdout LÀ CONSOLE.**
   Byte ra pipe/file **không đổi** ⇒ giữ tính tất định và không ảnh hưởng harness regression.
   Thêm 1 import **kernel32** (không phải libc). ⛔ **KHÔNG** đặt vô điều kiện (runtime magic, §11).
7. **Provenance cho node mono hoá ⇒ LÀM RFC NGAY.** Mở rộng `Token` (hiện 8 byte, dư 1 byte
   `padding`) hoặc bảng phụ offset→origin, để node clone giữ được nguồn gốc và lỗi trong code
   generic in đúng chỗ. RFC phải kèm **số đo** chi phí bộ nhớ + tốc độ biên dịch (Token là cấu
   trúc nóng nhất của lexer). Thay thế hiện trạng `ast.ax:265` (`offset = len(old_src)` ⇒ rơi
   ngoài mọi region ⇒ suy biến thành "không có vị trí").

- **`let x: u8 = 300` ⇒ REJECT `error[E3030]`** (user chốt 2026-07-30, option 2). Literal nguyên
  ngoài dải của kiểu **được chú thích tường minh** là lỗi biên dịch tại 8 vị trí (let/assign vào
  binding có chú thích/gán field struct/đối số/đối số dạng `f[u8](..)`/field init/phần tử mảng
  có kiểu chú thích/`return`). ✅ **ĐỐI SỐ của PHƯƠNG THỨC ĐÃ PHỦ 2026-07-31** (`s.setv(300)`,
  dạng gọi tĩnh `S.setv(&s,300)`, và dispatch động `i.take(300)`) — lỗ hổng CHUNG của cả BA
  luật chú thích (E3030 int-range, E3031 float→int, E3032 f64→f32) đã đóng bằng **MỘT** hook
  `check_method_args_annotated` → `check_annotated_target` (RFC 0006 **§6.4**), gọi từ khối quét
  `mfi.params` và từ site dispatch. Kiểu tham số đọc ra là **KIỂU KHAI BÁO** (đo bằng method
  generic khởi tạo ở 3 kiểu + 2 struct trùng tên method khác kiểu tham số). `300 as u8`
  vẫn là cách nói "cố ý cắt bit". Suy-theo-độ-lớn ở vị trí KHÔNG chú thích giữ nguyên (tiền lệ
  [[bug-negative-literal-compare-o0]]). Spec: **RFC 0006 §6.1**; chi tiết + phần CHƯA phủ
  (i64/u64, biểu thức hằng gấp, narrowing từ giá trị runtime):
  [question-out-of-range-narrow-int-literal](question-out-of-range-narrow-int-literal.md).
  Commit `abfe985` (E3030 gốc) + `f6ac69e` (phủ đối số method cho cả ba luật).

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

## 🐞🐞 BUG MỞ — quét sau `a281992` (probe8, 2026-07-31). **KHÔNG phải quét sạch: 3 lỗ SỐNG.**
Tất cả đều đo được, có control đối chứng, ở `bin/probe8/` (runner `run8.sh`, ma trận `matrix8.sh`).
Bài học bao trùm: `a281992` làm tên CHÍNH XÁC **thắng điểm** cửa sổ lỏng, nhưng **không chặn cửa sổ
lỏng LÀM CÂU TRẢ LỜI khi không có tên chính xác nào**. Sửa một nửa của một luật.

**A+B+B′ — rank-1 "loose": ✅ ĐÃ SỬA `a538983`** (A==B `B10DABE6`, regression 662/662, breakage
audit 1215 file: 7 reject mới đúng như dự định, 2 accept mới, **0 collateral**). Đã xoá luôn **bốn**
bản sao chết của cùng luật. Ngoài ra **hoàn thiện nhánh SỐ HỌC của RFC 0007 §3.1** (§2.2 chỉ hoãn
tới khi có error infra — nay đã có): `a + b` trên struct không có `add` trước đây **SEGFAULT (139)**,
nay là lỗi biên dịch. Tự kiểm chứng độc lập trên compiler CŨ: accept + exit 139.
⚠️ **`h1_rank2_eq` KHÔNG được sửa và không thể sửa bằng việc bỏ rank 1** — nó là **rank 2**
(`Num.eq__fast` khớp `eq` + dấu phân cách type-arg `__`). Cùng hình dạng với `h2_axstd_prefix`:
**tên do user đặt bắt chước cách mangle của monomorphizer**. Xem "Over-reach còn lại" trong RFC 0037
— sửa đúng là khoá rank 2/3 vào việc symbol **THỰC SỰ là một instantiation**, không phải vào cách
viết; hạn chế rank 2 theo prefix `_AX_std_` **sẽ làm hỏng** method instantiated của struct generic
(mangle đặt prefix TRƯỚC dấu chấm). Đó là **quyết định thiết kế**, không phải siết cơ học.
Mô tả gốc — chọn method theo **substring chặn `_`**:
- `==`/`<`/`+` gọi hàm user chưa từng đặt tên: chỉ có `deep_eq` ⇒ `a == b` trên hai struct KHÁC nhau
  ra **true** (`a1_op_deepeq` = 7). Chẩn đoán RFC 0007 §2.2 **đã tồn tại** trong compiler và bị
  rank 1 **bịt miệng**. Cùng cơ chế: `total_lt`→`<`, `checked_add`→`+`, `eq__fast` (rank 2).
- **Drop glue TỰ CHẾ lời gọi**: kiểu chỉ khai báo `pre_drop(self)` bị gọi làm drop glue ở MỌI lối ra
  scope (`a2` 5 lời gọi; `d4` 3 lời gọi vào method trả `ptr[i64]`) ⇒ **UAF đang chờ xảy ra**.
  Quy trách nhiệm chính xác: 42 dưới `-no-ctgc-free`, 7 mặc định ⇒ `resolve_drop_method`+`lower_destroy`.
  RFC 0014 định nghĩa hook là `Type.drop(self)`; `pre_drop` **không phải** `drop`.
- **`ownership.ax:138,162` CHƯA TỪNG được chuyển** sang rank mới ⇒ `type_has_drop` true cho
  `pre_drop` ⇒ `let b = a` bị **E4003 oan** (`a3`; control `a3c` dùng `cleanup` = 42). **Đúng hình
  dạng "hai bản sao một luật, một bản không bao giờ được sửa".**
- Bằng chứng an toàn: commit `a281992` ghi rằng bản compiler **bỏ hẳn rank 1** qua **619/619 và tự
  dựng lại byte-identical** (stdlib đều gọi bằng tên đầy đủ = rank 3). Còn `air_builder.ax:1571`
  `match_base_names` và `:1755` `match_mangled_method_name` = **bản sao CHẾT, 0 caller** ⇒ xóa.

**C — overload cùng tên trên receiver struct đè nhau ở symbol: ✅ ĐÃ SỬA 2026-07-31** cho dạng gọi
**TỰ DO** (A==B==C `52D1ABD4`, RFC 0035 §2bis, oracle `bin/t_structoverload{,dfe}.ax`).
⚠️ **Nửa còn lại VẪN MỞ và là bug KHÁC:** gọi bằng **cú pháp method** (`s.f(41)`, method inline,
`S.f(&s,41)`) bind theo RECEIVER, **bỏ qua arity** ⇒ [[bug-method-call-overload-ignores-arity]]
(`resolve_method_overload` không có arity, `resolve_free_call_overload` thì có — lại là **một luật,
hai bản sao**). Quy trách nhiệm chắc: cùng khai báo, **một** compiler, tự do ra 42 / method ra 2.
Mô tả gốc: `typecheck.ax:1233-1246` `free_fn_bare_mangles` trả false khi param 0 là struct/sum, tin rằng
`x86_regs.ax:338` mangle ra `ax_<Struct>_<fn>` "duy nhất theo receiver" — nhưng đó là duy nhất theo
**(receiver, tên)**, KHÔNG theo chữ ký. Vòng uniquing Phase-3.5 (`typecheck.ax:3227-3240`) không gắn
`MODDUP` cho overload thứ hai ⇒ **mọi lời gọi bind vào body khai báo TRƯỚC**.
⚠️ **Phân kỳ theo mức tối ưu**: `g10_seq` = **2 ở -O0/default, 42 ở -O1** (deterministic 5/5, dựng
lại 2 lần, binary byte-identical; default ≡ -O0 vì `optimize` mặc định false ở `main_air.ax:854/943`).
**-O1 chỉ CHE bằng inlining.** `g11_revboth` (đảo thứ tự khai báo) = 82. `g5_only2` = 22 nhưng **11
dưới `-no-dfe`** ⇒ các ca "đúng" chỉ là **ảo giác do DFE** xóa hàm bị che. Receiver i64/str (`g12`,
`g13`) = 42 vì chúng CÓ bare-mangle nên Phase 3.5 unique được.

**Phát hiện KỀ BÊN (không phải luật tên) — mỗi cái là một item riêng:**
- `check_iface_conformance` chỉ được gọi ở **3 site** (`typecheck.ax:764` field init, `:4193` let
  init, `:4208` return) + check param inline `:5439`. **Gán (assignment)** và **payload Option**
  KHÔNG được phủ ⇒ `f1_iface_optpayload` **SEGFAULT (139)** ở cả ba mức, `f5_iface_assign` = 7.
  Control `f1d`/`f5d` (struct không có method na ná) fail y hệt ⇒ gốc là **thiếu check**, không phải
  luật tên. → [[bug-iface-conformance-unchecked-sites]].
- **Kiểu TRẢ VỀ của operator method không được kiểm**: `eq(self,o) -> f64` đặt tên CHÍNH XÁC vẫn làm
  `a == b` ra true (đọc thanh ghi nguyên trong khi callee trả XMM0) — `d1c_eq_f64_exact`.
- **Method generic mono hoá trả `-> T` với T ↦ f64 ra 0.0**: `i8_genf64ret` = 40 ở cả ba mức.
  Controls thu hẹp rất gọn: `i6` đọc **field** f64 generic = 41 ✅, `i7` method f64 không generic
  = 41 ✅, `i9` struct generic có `-> f64` **cụ thể** = 42 ✅, `i5` `Box[i32].get` = 42 ✅ ⇒ khuyết tật
  đúng ở **`-> T` mono hoá sang f64**, nghi lớp thanh ghi trả về (**chưa xác minh** — xem objdump
  trước khi tin). → [[bug-mono-generic-ret-typaram-f64]].
- `h2_axstd_prefix` = 7 (mà `h3` đảo thứ tự = 42): phép **strip `_AX_std_`** biến một
  `_AX_std_eq` do user định nghĩa thành **tie rank-3** cướp `==` khỏi `eq` thật — over-reach do
  **chính `a281992`** đẻ ra, nằm ở luật phá hoà, không phải ở cửa sổ lỏng.
- `resolver.ax:653/662`: biên `.` đúng, nhưng **việc CHỌN** giữa nhiều `M.foo` là tuỳ tiện (theo thứ
  tự slot hash) — chưa bị chạm tới, ghi lại để không ai tưởng đã kiểm.

## ✅ Bề mặt đã QUÉT SẠCH (probe4, bankable làm oracle — đừng quét lại)
Option/Result payload × type class (`a1`–`a4`), generics × type class kể cả type arg tường minh
(`b1`,`b2`), str/bytes qua interface (`b3`), aggregate × float + global float (`c1`,`c2`), ABI float
qua biên 6 đối số + xen kẽ int/float (`c3`), cast/độ chính xác f32 kể cả chia single-rounded
(`c4`,`d1`), ADT user payload non-i64 (`e1`), tuple có float/str/struct (`e2`), stdlib container ở
element non-i64 (`e3`), interface breadth (`e4` — chỉ hàng f32 fail = bug #2).

## 🧪 BÀI HỌC ĐO — breakage audit: **đừng union một lượt chạy BỊ GIẾT với lượt chạy sạch**
Audit 1215 file báo **10 file "mới bị reject"** ngoài dự kiến (kể cả `tests/lexer/hello.ax`), mâu
thuẫn với 662/662. Kiểm lại từng file: **cả 10 đều biên dịch bình thường**. Nguyên nhân **có tên**:
cả 10 đều nằm ở chỉ số ≥ 822 — đúng dải mà một chunk **bị cap 10 phút giết**; lượt chạy bị giết ghi
"không có exe" thành "reject", và tôi đã **union** phần dở đó với lượt chạy lại sạch.
**Union chỉ có thể BÁO THỪA reject, không bao giờ giấu mất một reject thật** ⇒ kết luận audit vẫn
đứng vững. Lần sau: chunk < 10 phút, và **vứt** output của lượt bị giết thay vì gộp.

## 🐞🐞🐞 `print`/`println` (user báo 2026-08-05) — **P1/P2/P3 ĐÃ SỬA, P4 CÒN MỞ, P5 không phải bug**
✅ **P1+P2+P3 đóng bằng RFC 0038** (`air_builder.ax:2500-2595` desugar + `typecheck.ax:2599-2692`
`error[E3033]`). Gate **A==B = `99F795C212B3CFE2BF28DCB3CEF06CDA0CAF133F09EBC7D12F5400B13B3FF783`**,
regression **677/677 ở CẢ default lẫn `-O0`** (672 cũ + 5 hàng mới) ⇒ **BASELINE MỚI = 677**.
Driver `bin/axc_native.exe` = **2.308.096 byte** (mốc A==B này).
Tự kiểm chứng độc lập (không chỉ theo báo cáo agent): `Kết quả tính toán thử nghiệm: 15` exit 15;
byte `4b e1 ba bf …` = UTF-8 đúng; `println()` trần hết segfault; `error[E3033]: cannot print a value
of type \`Point\``.
⚠️ Hai điều implementer làm KHÁC brief, đều có lý do đo được: (a) dòng UTF-8 gộp vào **một** dòng
output vì harness so bằng command substitution (chỉ strip newline CUỐI) ⇒ hai dòng thì `want` không
khớp được; (b) **E3033 bỏ qua `TYPE_KIND_GENERIC`** — thân template chưa instantiate (`fn dump[T](v: T)`)
sẽ bị false-positive, bản mono hoá mang kiểu cụ thể và VẪN bị kiểm.
⚠️ Ghi nhận chưa xử lý: `cgen.ax:1527-1541` cùng hình dạng 1-đối-số, được desugar AIR sửa cho **miễn
phí**, nhưng bảng phân loại kiểu của nó **HẸP HƠN** selector x86 (`i8/i16/u8/u16/isize/usize` không
ánh xạ sang `_i64`). Có từ trước, backend-local ⇒ item riêng.

Bản ghi gốc (giữ nguyên vì mỗi mục dạy một bài học chẩn đoán):
Repro ở `bin/probe9/` (chạy **từ REPO ROOT**: `./bin/axc_native.exe build bin/probe9/pn1.ax -o
bin/probe9/pn1.exe -O0` — chạy từ `/tmp` sẽ fail "cannot open source file" vì driver phân giải
đường dẫn stdlib theo cwd).

**P1 — MỌI đối số sau đối số ĐẦU bị NUỐT IM LẶNG.** `println("val: ", s)` in `val: ` rồi hết.
Lớp accept-then-miscompile (họ BUG#53). Gốc: `x86_selector.ax:1737-1776` chỉ đọc
`extras.data[arg_start+1]` (đối số ĐẦU), đổi target thành hàm runtime **1 tham số**
(`ax_println_{str,i64,f64,bool}`, sym_imm −10..−17), nhưng **không sửa `arg_count`** ⇒ đối số 2..N
vẫn được nạp vào RDX/R8/… nơi không ai đọc. **AIR vẫn giữ đủ đối số** (dump `pn1.ax`: extras[0]=2,
extras[1]=%3, extras[2]=%2) ⇒ mất mát xảy ra **CHỈ ở selector**. `resolver.ax:484-485` đăng ký
`print`/`println` là `SYM_BUILTIN_TYPE` **không có kiểm arity ở bất cứ đâu**; `typecheck.ax` không
hề chứa chuỗi `"print"`/`"println"`.
⇒ **Hướng đã chốt: desugar ở `air_builder`, KHÔNG ở selector** (§4 CLAUDE.md — selector không được
tự chế N lời gọi từ 1). Điểm chèn **`lower_call_expr`, `air_builder.ax:1881`, ngay sau `:2498`
trước `:2500`**; tiền lệ cùng hình dạng: desugar builtin `assert` ở `:2481-2498`.
⚠️ **Lower HẾT đối số TRƯỚC rồi mới phát N lời gọi** — viết thẳng `print(a); print(b)` sẽ đổi thứ tự
đan xen tác dụng phụ (mất tính chất "đánh giá hết trước khi in" của printf). `temp_count < 2` ⇒ AIR
**byte-identical** với hôm nay.
⚠️ Ba bẫy: (1) `print_sym` phải **tra cứu** (`name_id == intern("print") and kind ==
SYM_BUILTIN_TYPE`), **không** suy ra bằng `println_sym - 1`; (2) guard theo `SYM_BUILTIN_TYPE` chứ
không theo tên, để hàm `println` do user định nghĩa không bị desugar; (3) chuỗi rỗng tổng hợp phải
intern **kèm dấu nháy** `"\"\""` — `lower_string_lit` (`:669-682`) intern **nguyên văn token có
`"`), backend strip bằng `unescape_string_literal` (`print_helpers.ax:24-37`). (Đường `assert` ở
`:2494` intern **không** nháy — đừng chép mù, nó sống sót chỉ vì `unescape` no-op khi ký tự đầu
không phải `"`.)
✅ **Blast radius = 0**: quét đủ 1202 file `*.ax` — **không có** call site ≥2 đối số nào trong
`bootstrap/stage1/`, `std/`, `stdlib/`, `tests/`. Compiler chỉ dùng `print_helpers.ax:75,80`
(**1 đối số, `str`**) và **không bao giờ** dùng `println()` trần ⇒ self-image không đổi ⇒ **gate
A==B**. Chỉ `examples/bigfloat128.ax:124` (chính file user sửa) là call site thật.
📄 **Cần RFC** (§13 — đổi hợp đồng bề mặt ngôn ngữ của builtin; chưa RFC nào phủ `print`): arity ≥ 0;
đánh giá trái→phải **trước mọi output**; chỉ lời gọi CUỐI của `println` phát newline; `println()` ≡
`println("")`; **không** chèn dấu phân cách. Ghi chú: `std/fmt.ax:47,52` **đã khai báo**
`pub fn print(args: ...)` variadic ⇒ desugar là *thực thi* chữ ký stdlib, không phải nới luật.

**P2 — `println()` KHÔNG đối số ⇒ SEGFAULT 139** (`bin/probe9/pn3.ax`). `x86_selector.ax:1770-1776`
chọn `ax_println_str` mà không có đối số ⇒ runtime deref rác trong RCX. (Bản ghi cũ của tôi nói
"`println()` đã chạy được" — **SAI**, đã đo lại.)

**P3 — lỗ KIỂU, độc lập arity, CÓ TỪ TRƯỚC.** `x86_selector.ax:1742` phân loại 1..8/15/16=int,
11=bool, 9/10=float, **còn lại ⇒ `str`** (deref như con trỏ chuỗi):
`char8` (id 13) ⇒ **SEGFAULT**; struct / `ptr[T]` / `void` ⇒ **dòng trắng, exit 0** (accept-then-
miscompile); `u32`/rune ⇒ in số chứ không in ký tự. Desugar **nhân bội phơi nhiễm** (đối số 2 hôm
nay bị nuốt vô hại, sau desugar thành lời gọi thật) ⇒ nên ship **kèm** chẩn đoán frontend
`error[Exxxx]: cannot print a value of type T`, cho phép đúng `{i8..u64, isize, usize, bool, f32,
f64, str, bytes}`. Frontend ⇒ không đổi gate.

**P4 — hàm `println` DO USER ĐỊNH NGHĨA bị cướp** (`bin/probe9/pn8.ax`: khai báo
`fn println(a: str, b: i64) -> i32`, gọi ra in `x` và trả **1** thay vì 7). Selector khớp theo
**chuỗi TÊN** đã phân giải (`x86_selector.ax:1730`→`:1737`), không theo **danh tính symbol** — đúng
lớp defect "khớp theo cách VIẾT, không theo danh tính" đã ghi cho RFC 0037. Sửa đúng = đưa việc chọn
symbol runtime ra khỏi selector ⇒ **chạm backend ⇒ B==C**. **Tách item riêng, không gộp vào P1.**

**P6 ✅ ĐÃ SỬA 2026-08-06** (A==B `A58F762AACDB79C1164481FFED09AD2DD4B0A357B662A4C08420C06BABB5FE97`,
regression **682/682 ở CẢ default lẫn `-O0`**). Gốc **KHÔNG** nằm ở dispatch interface như nghi ban
đầu: `typecheck.ax` (khối phục hồi BUG#61) đọc `NODE_FIELD_EXPR.payload` **như CHỈ SỐ SYMBOL** mà
không kiểm cờ phân biệt **2048**. Với method call thường, payload là **ID INTERN CỦA TÊN**, nên
`intern("count")` tra vào bảng symbol và **kiểu trả về của một symbol stdlib xa lạ** (Option/Result/
Vec) bị đóng dấu lên lời gọi trả `i64` ⇒ `c + 30` bị từ chối. Sửa bằng **một** accessor
`callee_symbol` (bản cài đặt DUY NHẤT của luật cờ-2048; site `:5470` cũng đi qua nó) **+ kiểm
`kind == SYM_FUNC`**.
⭐ Kiểm SYM_FUNC **không phải thừa**: nếu thiếu, callee `NODE_IDENT` là **VARIANT** (`Ok`/`Err`/
`Some`) rơi vào `pre_infer_func_signature(cdecl)` trên node khai báo VARIANT, đóng dấu chữ ký FUNC
**0 tham số** lên symbol variant ⇒ **từ chối oan** `'Ok' expects 0 argument(s), found 1`. Đo bằng
bản dựng có instrument: `std/result.ax` dính 3 lần (symkind=5, nodekind=42) và bị **reject với 7
lỗi**; `scratch/stage2_preprocessed.ax` **134 lỗi**. Cả hai nay biên dịch được.
Breakage audit **1649 file**, before/after: **0 reject mới**, 6 accept mới (3 file kể trên +
`std/collections_test.ax` + 2 oracle mới).
Bốn extern chết `fseek`/`ftell`/`rewind`/`fputs` ở `std/io.ax` **ĐÃ XOÁ**; `bin/t_ifaceconsumer.ax`
vẫn exit 46. Oracle: `bin/t_p6methodname.ax` (+ control `bin/t_p6methodnamectl.ax`),
`bin/t_p6count.ax`.
⚠️ Câu **"THÊM một extern thì vô hại; chỉ XOÁ mới kích hoạt"** trong bản ghi gốc dưới đây là **SAI**
(ảo giác lấy mẫu trên một ánh xạ tuỳ tiện) — thêm extern cũng làm hỏng ở Δ=+2,+5,+6.

Bản ghi gốc (giữ lại vì dấu vân tay chẩn đoán vẫn đáng học):
**XOÁ một khai báo `extern` CHẾT làm HỎNG biên dịch một chương trình
KHÔNG LIÊN QUAN.** Phát hiện 2026-08-06 khi dọn libc bước 0; **deterministic, đã bisect xong.**
Repro (compiler nào cũng dính, kể cả binary đã commit — vì `concatenate_stdlib` đọc `std/*.ax`
**từ đĩa lúc biên dịch**, nên đổi `std/io.ax` là đủ, không cần dựng lại compiler):
1. xoá **bất kỳ MỘT** dòng trong bốn dòng `extern "C" fn {fseek,ftell,rewind,fputs}` ở `std/io.ax`;
2. `./bin/axc_native.exe build bin/t_ifaceconsumer.ax -o x.exe -O1`
⇒ `error: operator '+' is not defined for Option/Result operands (missing .unwrap()?)` +
`error: type 'Option' does not implement 'add' ...`.
Biểu thức bị tố là `a.run(8) + b.run(20)` — **hai toán hạng đều là `i64`** trả về từ dispatch
interface, **cả chương trình không có Option/Vec nào**.
⭐ **Dấu vân tay quyết định:** *tên kiểu báo lỗi thay đổi theo SỐ LƯỢNG khai báo bị xoá* — xoá 1 ⇒
`Option`, xoá cả 4 ⇒ `Vec`. Đó là chữ ký của việc **đọc kiểu qua một CHỈ SỐ trôi theo bảng symbol**,
không phải kiểu gắn với biểu thức. (Câu tiếp theo của bản ghi gốc — *"THÊM một extern thì vô hại;
chỉ XOÁ mới kích hoạt"* — đã bị **BÁC BỎ**, xem phần đã-sửa ở trên.) Không
phải "nhiễu do đổi số dòng" (xoá một dòng COMMENT cũng vô hại — đã đo).
Không phải phụ thuộc sử dụng: **không file bundled nào dùng** fseek/ftell/rewind/fputs (đã grep đủ 8
file trong `main_air.ax:401-426`).
Hiện đang **fail-safe** (lỗi biên dịch, không phải miscompile) — nhưng cùng LỚP với các defect
"đọc qua chỉ số/tên không ổn định" đã đẻ ra miscompile ở repo này. Nghi vấn *(SAI — không phải
dispatch, xem phần đã-sửa)*: phân giải **kiểu TRẢ VỀ của lời gọi dispatch** trong typecheck.
✅ Bốn extern chết ở `std/io.ax` **đã được xoá** cùng bản sửa, đúng như dự kiến làm regression test.

**P5 — KHÔNG phải bug compiler: UTF-8 in ra sai là do CODEPAGE console.** Đã đo: exe phát **đúng
byte UTF-8** (`4b e1 ba bf 74 …` = "Kết quả…"), và render đúng trong cả Git Bash lẫn PowerShell khi
`chcp` = 65001. Runtime ghi byte thô bằng `WriteFile` (`bootstrap/runtime/syscall.ax:9-10`,
`panic.ax:42-50`) ⇒ console codepage ≠ 65001 sẽ mojibake. Cải tiến sản phẩm khả dĩ:
`SetConsoleOutputCP(65001)` lúc khởi động runtime **chỉ khi stdout là console** (không đổi byte ra
pipe/file ⇒ giữ tính tất định). **Đừng đi săn bug lexer/string — không có.**

## 🔴🔴 B1 — `strip_package_prefixes` LÀM HỎNG **HẰNG CHUỖI** (đo 2026-08-07, TỰ KIỂM CHỨNG)
```axiom
println("literal: std.string.len")   // in ra:  literal: len      ⛔
println("also: std.io.open")         // in ra:  also: open        ⛔
```
`main_air.ax` `strip_package_prefixes` là **replace văn bản trên TOÀN BỘ nguồn**, không phân biệt
code với **chuỗi**. Mọi chương trình user chứa `std.string.` / `std.io.` / `std.os.` … trong một
literal đều bị **sửa âm thầm, không cảnh báo**. Đây là **thay đổi ngữ nghĩa im lặng** — §3 cấm.
Nạn nhân sống trong chính compiler: `cgen.ax:773,777` so `fn_name == "std.string.len"`,
`resolver.ax:807-818` `intern_string("std.string.len")` — trong image self-host chúng thành `"len"`.
(Hiện *chưa* vỡ vì các nhánh đó liệt kê cả `"len"`, nhưng đó là **may**, không phải thiết kế.)

## ⭐ RESOLVER ĐÃ ĐỦ SỨC — `strip_package_prefixes` là thứ THỪA, không phải thứ cần thiết
Bằng chứng quyết định: `--no-stdlib` **bỏ qua cả `strip_imports` lẫn `strip_package_prefixes`**
(`main_air.ax:1096,1128`), và dưới nó:
`import std.math` + `std.math.sqrt(16.0)` ⇒ **build sạch, chạy ra `4.000000`** (cả -O0 lẫn -O1);
control âm: `std.math.no_such_thing(...)` ⇒ lỗi đúng, exit 1.
⇒ Chuỗi resolver + lazy loader **đã hoàn chỉnh và đang chạy**: `resolver.ax:785-800` →
`:824-888` → `main_air.ax:1817 ax_driver_load_module` (`replace(mod_name,".","/")` ở `:1823`).
❌ **Bác bỏ tiền lệ trong audit libc**: `std/net.ax` gọi `std.os.linux_sys.syscall(...)` **KHÔNG**
phải bằng chứng resolver — tiền tố đó nằm trong danh sách rewrite, và **`std/net.ax` không parse
nổi** (10 lỗi parse) nên chưa từng được biên dịch.
⚠️ `std.string.len` "chạy được" **KHÔNG PHẢI** do resolver — nó bị **xoá 11 byte khỏi text** nên
lexer chỉ thấy `len(...)`. Chứng minh: `std.string.no_such_fn` báo lỗi **tên trần**, và literal bị
hỏng (B1 ở trên).

## 🧭 HƯỚNG ĐÃ ĐỊNH GIÁ — option C, làm 4 stage, mỗi stage A==B
**Stage 1 ✅ XONG 2026-08-07** (A==B `9C6726C1…`, 709/709 ở cả hai mức, breakage audit 870 file).
`strip_imports` nay bôi trắng **chỉ khi tên module đúng bằng** một mục trong **MỘT bảng duy nhất**
`preprocessed_module_name()` (11 mục: 8 bundled + `std.os.win32`/`std.os.linux_sys`/`std.net` —
ba tên chỉ bị `strip_package_prefixes` viết lại, không bundled, nhưng import của chúng vẫn phải bôi
trắng vì call site đã mất tiền tố). `concatenate_stdlib` đọc đường dẫn qua `bundled_module_path(i)`
suy ra **từ chính bảng đó** ⇒ hết "hai danh sách".
⇒ `import std.math` + `std.math.sqrt(16.0)` **chạy end-to-end** (`bin/t_stdmath.ax` = 12.000000).
✅ **B2 rơi ra miễn phí** (so khớp CHÍNH XÁC ⇒ có biên phân cách theo cấu trúc); **B4** và **B5**
đã sửa cùng stage (xem §BUG PHỤ).
⚠️ **Hệ quả đã đo:** `import std.{sync,thread,process,iter,cli}` nay **tới được loader ⇒ compiler
SEGV** (B3/B6) thay vì bị nuốt im lặng như trước. Không file nào trong corpus dính (audit 0
collateral), nhưng đây là **bug tiếp theo phải sửa** — crash không chẩn đoán là tệ hơn cả im lặng.
**Stage 2:** đăng ký 8 tên bundled thành `SYM_MODULE` có cờ *bundled*; `lazy_resolver_resolve_field`
tra cứu **scope toàn cục của unit hiện tại** thay vì nạp file ⇒ `std.collections.new_vec` chạy, và
`std.string.len` bind **cùng một symbol** với `len` trần ⇒ **delta codegen = 0**.
**Stage 3:** **XOÁ HẲN `strip_package_prefixes`** ⇒ hết B1, và **mở khoá CỘT trong chẩn đoán** (§3b).
**Stage 4 (tuỳ chọn):** module scoping thật cho symbol spliced, dùng `Symbol.decl_node` → offset →
`srcmap_find` (hạ tầng đã có từ stage 2b chẩn đoán).
⛔ **KHÔNG làm option B** (nhồi thêm module vào bundle): chỉ 2/12 module thêm được an toàn
(`math`, `sort`), trả **+0,15 s mỗi lần biên dịch VĨNH VIỄN** cho mọi user kể cả người không dùng,
và **không sửa gì về cấu trúc**.
📏 Đo được (bác bỏ lo ngại của tôi): DFE **bật mặc định** (`main_air.ax:928`) ⇒ bundle thêm module
không dùng cho ra **binary BYTE-IDENTICAL**. Giá thật là **thời gian biên dịch**, không phải kích thước.

## 🐞 BUG PHỤ tìm được cùng lúc (mỗi cái filable riêng)
- **B2 ✅ ĐÃ SỬA (stage 1)** `import stdthing` bị xoá oan — `match_prefix(s,i,"import std")` **không
  có biên phân cách**. Rơi ra miễn phí khi đổi sang so khớp **chính xác** với bảng module.
  Oracle `bin/t_stdprefixmod.ax` (+ fixture `std_util.ax` ở gốc repo): reject trước, 42 sau.
- **B3 + B6 ✅ ĐÃ SỬA 2026-09-05 — HAI CÁI LÀ MỘT BUG.** `pre_infer_func_signature` đánh chỉ số
  `decl_node` của unit KHÁC vào `self.tree` ⇒ compiler SEGV 139. Chi tiết, ma trận kích hoạt, thí
  nghiệm nhồi và cách sửa: **`knowledge/bug-cross-tree-decl-node-segv.md`**.
- **B3b 🔴 CÒN MỞ (defect KHÁC)** — `import std.{iter,process,cli}` vẫn 139: nguồn module nạp trễ
  không được tiền xử lý ⇒ nạp `std/collections.ax` **lần thứ hai**. Đã chứng minh độc lập với bản vá
  B3 ⇒ thuộc **stage 2** (đổi thiết kế). Repro: `nestbundledmod.ax` + `bin/probe_nestbundled.ax`.
- **B4 ✅ ĐÃ SỬA (stage 1)** — `mod.no_such_member()` trên module **đã nạp thành công**: nhận, sinh
  exe, SEGV lúc chạy, KHÔNG chẩn đoán. Nay `error: module 'X' has no member 'Y'` + `--> file:line`
  + caret. ⭐ Kiểm BUG#93 cũ (`typecheck.ax`) **không thấy** hình dạng này: nó hỏi "gốc chuỗi có
  **chưa** bind không", mà ở đây gốc **đã** bind — vào chính module. Luật mới đọc **receiver**:
  FIELD_EXPR có cờ 2048, hoặc IDENT có payload là symbol thật (so `name_id`). Hai luật loại trừ
  nhau (`recv_mod_sym == 0`) nên không bao giờ in hai lỗi cho một chỗ.
  Oracle `bin/t_modnomember.ax` (dạng IDENT, 139 trước) + `bin/t_stdmathnomember.ax` (dạng
  FIELD_EXPR); đối chứng dương `bin/t_modcollide.ax` = 101.
- **B5 ✅ ĐÃ SỬA (stage 1)** — module nạp lazy có lỗi parse/typecheck KHÔNG dừng pipeline.
  Nay `LazyResolver.load_errors` cộng dồn `parser_ptr.diags_count` + `mod_checker.diags_count`,
  driver chặn ngay trước cổng `checker.diags_count`. Bonus: srcmap của module được gắn **TRƯỚC**
  `parse_program` ⇒ lỗi parse trong module in `--> std/foo.ax:4` thay vì "byte offset N".
  Oracle `bin/t_modparseerr.ax` (in 42 trước, reject sau).

## 📋 SỨC KHOẺ 12 module (đo, không đoán)
Đo lại **sau stage 1** (bằng `import std.X` thật, không phải suy đoán):
| module | trạng thái sau stage 1 |
|---|---|
| `math`, `sort` | ✅ **chạy end-to-end** (`sqrt(16)+pow(2,3)` = 12.000000, cả -O0 lẫn -O1) |
| `sync`, `thread` | ✅ **biên dịch sạch** sau khi sửa B3/B6 (2026-09-05) |
| `process`, `iter`, `cli` | 🔴 vẫn **SEGV** — nhưng vì **B3b** (nạp trùng module đã bundle), không phải B3 |
| `json, fmt, time, crypto, log` | ❌ reject sạch (lỗi parse của module, nay CHÍ TỬ nhờ B5) |
| `net` | vẫn nằm trong bảng bôi trắng ⇒ không đổi (không có gì để gọi) |
⇒ **Chỉ `math` + `sort` dùng được**, đúng như định giá trước khi làm — stage 1 gỡ **rào cấu trúc**,
không phải là lời hứa 12 module chạy được.
✅ Xác nhận `std/math.ax` là code THẬT: 0 import, 0 extern, 45+ `pub fn`;
`sqrt(16)+pow(2,10)+sin(.5)+cos(.5)+ln(3)` = **1030.455620** đúng.

## 🔴 PHÁT HIỆN 2026-08-07 — PHẦN LỚN `std/` KHÔNG DÙNG ĐƯỢC trên đường build native
✅ **Nửa `strip_imports` của vấn đề này ĐÃ ĐÓNG ở stage 1** (một bảng module duy nhất). Phần còn
lại — `strip_package_prefixes` viết lại văn bản, và B1 làm hỏng hằng chuỗi — là stage 2/3.
Bản ghi gốc giữ nguyên vì ma trận đo được vẫn là bằng chứng:
Đo trực tiếp (không suy đoán). Có **HAI danh sách CỨNG phải khớp nhau, và chúng KHÔNG khớp**:
- `concatenate_stdlib` (`main_air.ax`) nối **8 file**: result, mem/alloc, scheduler, runtime, os,
  string, io, collections.
- `strip_package_prefixes` viết lại **8 tiền tố**: `std.mem.alloc.`, `std.scheduler.`,
  `std.os.win32.`, `std.os.linux_sys.`, `std.os.`, `std.string.`, `std.io.`, `std.net.`
- `strip_imports` **xoá mọi dòng `import std`** ⇒ module không nằm trong danh sách bundle thì
  **không bao giờ được nạp**.

Ma trận đo được:
| module | trong bundle? | tiền tố được rewrite? | gọi `std.X.f()` | gọi trần `f()` |
|---|---|---|---|---|
| `std.string` | ✅ | ✅ | ✅ **chạy** (`std.string.len("chào")` = 5) | ✅ |
| `std.collections` | ✅ | ❌ | ❌ `undefined name 'std'` | ✅ **chạy** |
| `std.math` (và sort/json/fmt/time/process/thread/crypto/log/sync/iter) | ❌ | ❌ | ❌ **KHÔNG DÙNG ĐƯỢC** | ❌ |
| `std.net` | ❌ | ✅ | ❌ (rewrite rồi nhưng chẳng có gì để gọi) | ❌ |

⇒ **`std/math.ax` hiện thực `sqrt/pow/sin/cos` thuần AXIOM (không cần libm) nhưng user KHÔNG GỌI
ĐƯỢC.** Cùng cảnh: `sort, json, fmt, time, process, thread, crypto, log, sync, iter`.
✅ Chẩn đoán **tốt, không im lặng**: `error: unresolved call to 'sqrt' on undefined namespace 'std'
-- module not imported or not bundled on this build`.
⚠️ **ĐÍNH CHÍNH audit libc:** mục "`std/process.ax` + `std/net.ax` bị phân loại nhầm DLL ⇒ import
từ `ucrtbase` không export" là **KHÔNG TỚI ĐƯỢC** — hai module đó **không dùng được** trên đường
native ngay từ đầu. Vẫn nên sửa whitelist, nhưng **không phải bug user gặp**; hạ ưu tiên.
⇒ Đây đúng lớp defect **"hai bản sao của một sự thật"** đắt nhất repo này. Sửa đúng = **một nguồn
sự thật** (danh sách module + tiền tố sinh ra từ cùng một chỗ), hoặc bỏ hẳn
`strip_package_prefixes` bằng resolver qualified-name (xem §3b — cũng xoá luôn bài toán CỘT của
chẩn đoán). **Cần định giá trước khi làm.**
📌 Ghi chú phụ: `std.string.len("chào")` = **5** ⇒ trả **SỐ BYTE**, không phải số ký tự. Hợp lý cho
`str.len`, nhưng cần ghi rõ trong spec vì AXIOM tuyên bố "UTF-8 mặc định".

## 🔜 TASK MỞ (tự làm được, theo thứ tự giá trị)
0. ✅ **XONG (`6febd02`)** — int → f64 ở các vị trí typecheck không lan hint. Xem bug #1 ở trên.
   ⚠️ Còn **nợ thật sự** từ `b8ac125`: bảy vị trí f32→f64 (đối số method, gán phần tử mảng, biểu
   thức, kết quả lời gọi, param, đọc field, binding suy diễn) nay trả **3 thay vì 0** — đó **KHÔNG
   phải là đã phủ**, chỉ là chuyển từ "nhận-rồi-miscompile" sang "nhận-mà-không-có-chẩn-đoán".
   Và `verify_air_no_int_into_float` phân loại INT/FLOAT/UNKNOWN **không có khái niệm BỀ RỘNG**, nên
   một `OP_COPY` type f64 đọc vreg f32 vẫn qua được §9. Mở rộng miền trừu tượng đó = RFC 0006 §7.3.
1. **Probing tiếp các bề mặt chưa quét** — đã trả lãi **3 miscompile trong phiên 07-30c, rồi 3 nữa
   trong phiên 07-30d**. Dùng skill `axiom-bug-probe`. Bề mặt sạch bank ở `bin/t_methretbreadth.ax`
   + danh sách probe4 ngay trên.
2. **Nợ kỹ thuật đo:** mọi tuyên bố perf phải là **median trên nhiều layout + spread bên cạnh**
   (`scripts/perf_layout_dist.ps1`, `scripts/perf_m6_gate.ps1`). **KHÔNG dùng `perf_suite.ps1`** để
   phán quyết gate nữa.
3. 🆕 **GỠ PHỤ THUỘC libc — AXIOM phải ĐỘC LẬP với libc** (user, 2026-08-05). **ĐÃ AUDIT XONG.**
   📄 Chi tiết + kế hoạch 11 bước: [audit-libc-dependencies](audit-libc-dependencies-2026-08-05.md).
   Số đo (bảng import PE thật, **không** phải `strings`): `axc_native.exe` import **16 ký hiệu
   `ucrtbase.dll`**; chương trình tầm thường chỉ **3** (memset/memcpy/strlen).
   ⭐ **Hai giả thiết bị LẬT:** (a) "blocker resolver extern-C" ở comment `std/io.ax:4-5` là **comment
   LỖI THỜI, không phải bug** — resolver/mangler đã đúng từ `a61b19e`, cái hỏng là **danh sách bundle
   CỨNG** `main_air.ax:401-426` thiếu `std/os/win32.ax` ⇒ **`std/io.ax` viết lại được NGAY**;
   (b) sửa `std/` **không đủ** — `bin/ax_runtime.dll` là artifact **C link UCRT** (`runtime/ax_print.c`),
   nhưng bản runtime libc-free cho Windows **đã viết ~80% trong `bootstrap/runtime/panic.ax` và đang
   bị gate ELF-only** ở `linker.ax:3162-3175`.
   ⚠️ `linker.ax:731` **fall-through về `ucrtbase.dll`** ⇒ quên whitelist một tên = âm thầm thành
   import libc. Đã lộ 2 nhóm phân loại nhầm: `std/process.ax:9,11` (phải kernel32) và
   `std/net.ax:10-19` (phải **ws2_32.dll**, linker chưa có bucket).
   ⛔ **`atof` KHÔNG được thay** bằng `std/string.ax:818` (không có số mũ, không làm tròn đúng) —
   sẽ đổi âm thầm mọi float literal compiler sinh ra. Cần RFC.
   Phán quyết: **11/16 gỡ được bằng frontend/std (A==B)**, 4 cái cần **B==C**, `ax_runtime.dll` là
   phần dư riêng, `atof` cần RFC.
3b. ✅ **CHẨN ĐOÁN §8 — STAGE 0/1/2a/2b ĐÃ XONG** (2026-08-07, `0515e30`, `9f9b265`, `5916c1d`).
   Mọi chẩn đoán nay có `--> file.ax:LINE` + gutter có số dòng + caret, phát từ **MỘT**
   renderer `print_diag_location` (`print_helpers.ax`). Trước đó có **BA** renderer đã trôi lệch
   (chỉ typecheck in trích đoạn; parser và ownership chỉ in offset). Cascade: 52 dòng → 3.
   `total_len` + dump token + thuật ngữ `nud`/`INDENT` đã gỡ.
   ⭐ **`strip_imports` nay BÔI TRẮNG dòng thành dấu cách thay vì XÓA** ⇒ **bảo toàn CẢ offset
   byte LẪN số dòng** (mạnh hơn thiết kế ban đầu chỉ nhắm giữ số dòng). Đo: file 3 import,
   lỗi dòng 7 ⇒ offset 141596 − region start 141506 = byte 90 = **đúng dòng 7, cột 3**.
   ⚠️ **CÒN LẠI (stage 3+), đừng tưởng đã xong:**
   - **KHÔNG in CỘT** — cố ý. `strip_package_prefixes` viết lại dòng tại chỗ (xóa `std.string.`
     = 11 byte) nên cột sẽ **sai lặng lẽ**. `file.ax:12` đúng hơn `file.ax:12:8` sai.
     ⭐ Investigator gợi ý hướng TỐT HƠN stage 3: **khừ hẳn `strip_package_prefixes`** bằng
     resolver qualified-name (audit libc nói resolver đã đúng từ `a61b19e`) ⇒ xóa luôn bài toán
     cột thay vì lách nó. **Cần định giá trước khi làm stage 3.**
   - **Node mono hóa KHÔNG định vị được**: `ast.ax:265` cho clone `offset = len(old_src)` ⇒ rơi
     ngoài mọi region (đo: 147025 vs buồn‑đệm hết ở 141593). Nay suy biến thành "không có vị trí"
     thay vì đoán bừa. Sửa thật cần **provenance trên `Token`** (hiện 8 byte, chỉ dư 1 byte
     `padding`) ⇒ **cần RFC**.
   - **Cây module IMPORT có `srcmap == null`** ⇒ chẩn đoán trong đó vẫn in ghi chú byte offset.
     Đường dẫn module **đã biết** ở `lazy_resolver_preload_module` ⇒ map 1 entry là đủ. **RẺ.**
   - **Parser báo token NƠI MONG ĐỢI biểu thức, không phải toán tử thiếu toán hạng**: `let a = 1 +`
     trỏ vào **dòng SAU** (`return`). Đúng về kỹ thuật, khó đọc. Có từ trước, nay **lộ ra**
     vì số dòng đã thật.
   - Khối `[Debug] Reading *.ax` / `[codegen]` vẫn **vô điều kiện**. Gốc sâu hơn:
     `is_verbose_debug` (`print_helpers.ax:108-182`) lọc theo **CÁCH VIẾT** (whitelist ~19 chuỗi
     `[D...`), không theo cờ verbose — **cùng lớp defect "khớp theo chính tả" với RFC 0037 và P4**.
     Thay bằng cờ `--verbose` thật = đổi bề mặt CLI ⇒ **cần RFC**.
4. 🆕 **QUY TẮC TEST MỚI (user, 2026-08-05) — xem CLAUDE.md §7.1.** Không phán quyết bằng exit code
   nữa; oracle phải `println("<chuỗi UTF-8>", <giá trị>)` và so **stdout tường minh**.
   ✅ **HẾT BỊ CHẶN (xác nhận 2026-09-05)** — P1 đã đóng bằng RFC 0038 (xem §`print`/`println` ở trên),
   nên `println("...", val)` **in đúng giá trị**. Bằng chứng end-to-end: hai oracle mới
   `t_b3lazyintrinsic` / `t_b3stdsync` chạy qua suite ở **cả** default lẫn `-O0` và in
   `... : 42`. ⇒ **Di trú oracle sang §7.1 nay THỰC HIỆN ĐƯỢC**; 711 hàng baseline vẫn đa số là
   `exit|` — di trú **dần**, không big-bang.
   ⚠️ Dòng "bị chặn bởi P1" cũ đã đứng sai suốt từ khi RFC 0038 ship — đúng cái bẫy CLAUDE.md §24
   mô tả: **một TODO ghi giữa phiên thường bị một commit SAU đó đóng, mà không ai quay lại gạch đi.**

## ✅ ĐÃ XONG gần đây (giao thức đo M6)
- **arrwalk có bản đọc phân phối** — đã xong ở commit `3ef26f0` (cùng ngày, template `hot()` trong
  `perf_m6_gate.ps1` đã mở rộng cho global array), rồi được thêm vào `$Shapes` mặc định 2026-07-30
  (phiên sau). Đọc lại xác nhận: **1.087–1.092x, PASS** (2 lần đo độc lập), xorshift control không
  đổi (0.995x) ⇒ không hồi quy. Dòng "còn nợ" trong handoff 07-30c bị để sót — đã sửa. Bài học: khi
  một TODO được giải quyết bởi commit SAU trong CÙNG phiên, sửa luôn file chứa TODO trong commit đó.

## ⚠️ LATENT — cố ý KHÔNG sửa
- **XMM0–XMM3 không được bảo vệ ở allocator** → [bug-float-arg-reg-unprotected](bug-float-arg-reg-unprotected.md).
  Hiện KHÔNG tới được (đúng nhờ **thứ tự** `emit_param_prologue`, không nhờ may). §10 CLAUDE.md:
  lợi ích không đo được thì không mua độ phức tạp ở thành phần self-host-critical nhất.
  Nếu sửa: **backend ⇒ B==C bắt buộc**. Repro `bin/t_floatparamchain.ax`.

## ⛔ ĐÃ BÁC BỎ — đừng xây lại
- **Loop-header alignment**: dịch +16 byte **bảo toàn** địa chỉ mod 16 ⇒ pass căn-lề-16 không thể
  giải thích delta; và đo được **layout NHANH NHẤT là cái LỆCH LỀ NHẤT**. Refuted trước khi viết code.
- **Loop rotation / bottom-test**: đo +0,1%.
- **Copy-propagation xuôi**: kéo dài live range ⇒ spill (fib −6,5%). Bản đúng là fold NGƯỢC
  (`coalesce_dest_copy`, đã ship).
- **"1f làm callloop chậm 7%"**: bóng ma layout. Re-price trên 8 layout: median 143,5 vs 143,5 =
  **đúng bằng 0**. 1f là **lợi ích thuần** (tailrec −14%).

## 📏 BÀI HỌC PHẢI ĐỌC TRƯỚC KHI VIẾT ORACLE / ĐO PERF
- [lesson-exit-code-8bit-masking](lesson-exit-code-8bit-masking.md) — exit code bị mask 8 bit;
  `return 300` → 44. **Mọi giá trị kỳ vọng < 256**; tính+so sánh TRONG chương trình rồi `return 42`.
  Đã có **guard abort cả suite** trong `regression_repros.sh`.
- **Ghép cặp KHÔNG khử được thiên lệch layout** — 4/4 cặp cùng dấu đã báo "+7%" cho hiệu ứng thật = 0.
- **Harness perf phải in SÀN STARTUP cạnh con số** (sàn thật ≈ 10,3 ms).
- **Peephole phải được chứng minh là CÓ NỔ, không chỉ AN TOÀN** (bản đầu của 1d khớp 0/4 hằng mà vẫn
  qua sạch mọi gate).
- **Floor chỉ là floor nếu chương trình khớp CHÍNH XÁC** (xorshift bị paraphrase → đo sai 1,50x).

## 🔁 Cách chạy gate
`Skill(axiom-fixpoint-gate)`. Frontend-only ⇒ **A==B**; backend/ABI/linker ⇒ **B==C BẮT BUỘC trước
commit**. Regression: `scripts/regression_repros.sh` (**≥649**), có lượt `-O0`.

## 🧭 QUY TRÌNH — đọc `git log` TRƯỚC KHI dispatch (học lại 2026-07-31)
Skill autopilot đã có luật "cross-check mục backlog với `git log` trước khi dispatch", **và tôi vẫn
bỏ qua**: dispatch một investigator đi phán quyết bug #4 theo giả thuyết "thanh ghi float cũ", trong
khi `git log --oneline -14` cho thấy ngay hai commit `3741afc`/`a281992` đã **bác bỏ và sửa** nó từ
02:04 cùng ngày. Mất một lượt dispatch. **Một `git log` rẻ hơn mọi dispatch** — và ở repo này, file
backlog trôi lệch nhanh hơn ta tưởng vì bug được sửa trong CÙNG phiên mà mục backlog không ai gạch.
