# BACKLOG — orientation file (READ THIS, not MEMORY.md)

⚠️ **Đây là FILE CON TRỎ, không phải kho sự thật.** Mọi chi tiết/bằng chứng nằm trong
`knowledge/<topic>.md` và `knowledge/MEMORY.md`. Nếu hai file nói khác nhau, **`MEMORY.md` đúng** —
và hãy sửa file này ngay, đừng để trôi lệch (hai bản sao của một sự thật chính là lớp defect đã
sinh ra bug interface-return).

**Vì sao file này tồn tại:** `knowledge/MEMORY.md` = **175 KB ≈ 87k token**, vượt read cap; chỉ một
trang bị cắt của nó đã tốn ~25k token mỗi phiên **trước khi làm bất cứ việc gì**. Định hướng bằng
file này (~2k token) + handoff mới nhất; vào `MEMORY.md` **chỉ bằng `Grep`** theo tên topic.

---

## Trạng thái cây (cập nhật 2026-08-07 — RFC 0039)
- 🆕 **RFC 0039 (struct literal suy diễn từ annotation) — A==B
  `84A13E958B59D2A1022C860C8E4637E81716BA65A66BFDDFD096034E4DB3FF68`** (2.313.728 byte).
  `let c: TmpStruct = (a: 64, b: 64)` nay hợp lệ; `error[E3034]` khi không có ngữ cảnh suy diễn.
  Tự kiểm chứng **685/685** ở cả hai mức + control tuple/biểu thức ngoặc.
- Mốc trước: sửa P6 `A58F762A…` (2.307.584 byte), tự kiểm chứng 682/682;
  dọn libc `0585124E…`; RFC 0038 (`ca7a98d`) `99F795C2…` 2.308.096 byte.
  ⚠️ Mốc **B==C** gần nhất vẫn là `c3eae77` /
  `52D1ABD4AE9E6EF11216AD3B8318D1592C1C03F383D49F5464B6ABF0A6C9478B` (2.297.856 byte) — cả RFC 0038
  lẫn dọn-libc đều frontend/std nên chỉ cần A==B; **thay đổi backend kế tiếp phải dựng lại B==C từ
  driver MỚI này**.
- **libc: `ucrtbase.dll` 16 → 13 ký hiệu** (mất `clock`, `exit`, `fflush`).
  `QueryPerformanceCounter`/`Frequency` thêm vào **kernel32** (không phải libc).
  ⚠️ Đo bằng **parse bảng import PE**, KHÔNG bằng `strings` — xem
  [audit-libc-dependencies](audit-libc-dependencies-2026-08-05.md).
- **BASELINE = 685/685**, đo ở **cả default lẫn `-O0`**. Dưới 685 là RED. (672 → 677 RFC 0038, → 679 dọn libc 0-2, → 682 sửa P6, → 685 RFC 0039.)
  (611 → 649 → 662 → 672: +32 hàng ở `b8ac125`, +7 ở `f6ac69e`, +13 ở `a538983`, +10 khi đóng
  hole C — gồm một khối `-no-dfe` riêng, vì DFE che đúng cái defect đó.)
- `bin/axc_pre1f.exe` = compiler tham chiếu tiền-1f, giữ để định giá ghép cặp.
  `bin/probe5/axc_new.exe` (30/07 21:59) = mốc **trước** `6febd02`, hữu ích để quy trách nhiệm.
- Handoff mới nhất: [session-handoff-2026-07-30d](session-handoff-2026-07-30d.md) — ⚠️ phần
  "REMAINING QUEUE" của nó **đã lỗi thời**: bug #2, bug #3 và task 0 đều đã ship sau đó.

## ⛔ CẦN USER QUYẾT (không tự quyết — hạng D1)
1. **fib không phán quyết được bằng gate M6 như đang viết.** Layout spread của fib 17,2% (AXIOM) /
   13,7% (floor) > toàn bộ biên gate 15%; mẫu số của tỉ số **tự nó bimodal** theo parity của dịch
   16 byte. Tăng n không cứu. ⇒ **phát biểu lại gate** (so phân phối, hoặc ghim MỘT layout tham
   chiếu cho cả hai phía) **hoặc loại fib khỏi gate**. Chi tiết: handoff 07-30c.
   *(mục 2 — `let x: u8 = 300` — ĐÃ ĐƯỢC QUYẾT 2026-07-30, xem ngay dưới)*

## ✅ ĐÃ QUYẾT (D1) — không hỏi lại
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
3b. 🆕 **NHIỄU CHẨN ĐOÁN — vi phạm §8 "diagnostics là TÍNH NĂNG SẢN PHẨM"** (đo 2026-08-06).
   - `main_air.ax:489` in **`total_len=<số>` VÔ ĐIỀU KIỆN, KHÔNG có tiền tố `[Debug]`**, trên **mọi
     lần biên dịch thành công**. Đây là dòng nhiễu duy nhất không gắn nhãn ⇒ xoá (hoặc gộp vào dòng
     `[Debug]` ở `:488`, vốn không có `\n` nên hai lệnh này định gộp thành MỘT dòng — hiện đang tách).
   - Cả khối `[Debug] Reading *.ax...` / `[Debug] Stage N...` cũng **vô điều kiện** ⇒ nên đưa sau cờ
     verbose. Task lớn hơn, tách riêng.
   - 🔴 **LỖI PARSE là ca TỆ NHẤT — đo 2026-08-06, repro `bin/probe11/s1.ax`** (9 dòng nguồn,
     dùng cú pháp struct literal sai `let c: TmpStruct = (a: 64, b:64)`):
     ```
       tokens[32928]: kind=67 offset=142042      <- dump token THÔ, kind là SỐ
     error: unexpected token at offset 142042    <- offset BYTE vào buffer ĐÃ NỐI
     error: expected newline at offset 142042
     error: expected expression nud at offset 142042
     ... (9 lỗi dây chuyền từ MỘT sai sót)
     ```
     Ba khuyết tật chồng nhau: (1) **offset 142042** là vị trí byte trong **nguồn đã nối stdlib**
     (~142 KB) — file user chỉ 9 dòng, nên con số này **vô nghĩa với user**; không file/dòng/cột,
     không trích đoạn nguồn; (2) **9 lỗi dây chuyền** từ một sai sót, không có recovery;
     (3) **dump token thô** (`kind=67`, kind là số nguyên nội bộ) in ra trên đường lỗi NGƯỜI DÙNG.
     `nud` cũng là thuật ngữ Pratt-parser nội bộ, không phải ngôn ngữ người dùng.
     ⇒ Ưu tiên cao hơn hai mục dưới: đây là ấn tượng đầu tiên của mọi user gõ sai cú pháp.
   - **THIẾU vị trí nguồn ở MỌI chẩn đoán** (không riêng E3033): §8 quy định
     `--> file.ax:12:8` + **số dòng trong gutter**. Hiện E3030 và E3033 đều in gutter RỖNG (`   |`)
     và không có dòng `-->`. Đã kiểm chứng cả hai ⇒ **thiếu sót toàn cục của hạ tầng diagnostic**,
     không phải sót của một mã lỗi. Sửa ở nơi render diagnostic dùng chung.
   Cả ba đều frontend ⇒ **A==B**.
4. 🆕 **QUY TẮC TEST MỚI (user, 2026-08-05) — xem CLAUDE.md §7.1.** Không phán quyết bằng exit code
   nữa; oracle phải `println("<chuỗi UTF-8>", <giá trị>)` và so **stdout tường minh**.
   ⚠️ **BỊ CHẶN BỞI P1**: `println(str, val)` hiện nuốt đối số thứ hai ⇒ **phải sửa P1 trước**, rồi
   mới di trú oracle. 672 hàng baseline hiện đa số là `exit|` — di trú dần, không big-bang.

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
