# BACKLOG — orientation file (READ THIS, not MEMORY.md)

⚠️ **Đây là FILE CON TRỎ, không phải kho sự thật.** Mọi chi tiết/bằng chứng nằm trong
`knowledge/<topic>.md` và `knowledge/MEMORY.md`. Nếu hai file nói khác nhau, **`MEMORY.md` đúng** —
và hãy sửa file này ngay, đừng để trôi lệch (hai bản sao của một sự thật chính là lớp defect đã
sinh ra bug interface-return).

**Vì sao file này tồn tại:** `knowledge/MEMORY.md` = **175 KB ≈ 87k token**, vượt read cap; chỉ một
trang bị cắt của nó đã tốn ~25k token mỗi phiên **trước khi làm bất cứ việc gì**. Định hướng bằng
file này (~2k token) + handoff mới nhất; vào `MEMORY.md` **chỉ bằng `Grep`** theo tên topic.

---

## Trạng thái cây (cập nhật 2026-07-31)
- **HEAD** = `f6ac69e` — E3030/E3031/E3032 phủ cả **đối số method** bằng MỘT hook. Driver
  `bin/axc_native.exe` = **A==B `9D8C7D68`** (mốc B==C gần nhất: `D3EABC61` ở `b8ac125`).
- **BASELINE = 649/649**, đo ở **cả default lẫn `-O0`**. Dưới 649 là RED.
  (611 trong bản cũ của file này đã lỗi thời: +32 hàng ở `b8ac125`, +7 ở `f6ac69e`.)
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

## ✅ Bề mặt đã QUÉT SẠCH (probe4, bankable làm oracle — đừng quét lại)
Option/Result payload × type class (`a1`–`a4`), generics × type class kể cả type arg tường minh
(`b1`,`b2`), str/bytes qua interface (`b3`), aggregate × float + global float (`c1`,`c2`), ABI float
qua biên 6 đối số + xen kẽ int/float (`c3`), cast/độ chính xác f32 kể cả chia single-rounded
(`c4`,`d1`), ADT user payload non-i64 (`e1`), tuple có float/str/struct (`e2`), stdlib container ở
element non-i64 (`e3`), interface breadth (`e4` — chỉ hàng f32 fail = bug #2).

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
