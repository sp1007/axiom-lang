# BACKLOG — orientation file (READ THIS, not MEMORY.md)

⚠️ **Đây là FILE CON TRỎ, không phải kho sự thật.** Mọi chi tiết/bằng chứng nằm trong
`knowledge/<topic>.md` và `knowledge/MEMORY.md`. Nếu hai file nói khác nhau, **`MEMORY.md` đúng** —
và hãy sửa file này ngay, đừng để trôi lệch (hai bản sao của một sự thật chính là lớp defect đã
sinh ra bug interface-return).

**Vì sao file này tồn tại:** `knowledge/MEMORY.md` = **175 KB ≈ 87k token**, vượt read cap; chỉ một
trang bị cắt của nó đã tốn ~25k token mỗi phiên **trước khi làm bất cứ việc gì**. Định hướng bằng
file này (~2k token) + handoff mới nhất; vào `MEMORY.md` **chỉ bằng `Grep`** theo tên topic.

---

## Trạng thái cây (cập nhật 2026-07-30)
- **HEAD `e3222f0`** + fix int-literal→float + reject E3030, main. Driver `bin/axc_native.exe` = **A==B `78295509`**.
- **BASELINE = 607/607** (597 + 10 hàng E3030: 9 reject `t_intrange*` + `t_intrangeok`). Dưới 607 là RED.
  Driver sau khi ship E3030 = **A==B `78295509`**.
- `bin/axc_pre1f.exe` = compiler tham chiếu tiền-1f, giữ để định giá ghép cặp.
- Handoff mới nhất: [session-handoff-2026-07-30c](session-handoff-2026-07-30c.md).

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
  có kiểu chú thích/`return`). **CHƯA phủ: đối số của PHƯƠNG THỨC** `s.setv(300)` (vẫn wrap 44). `300 as u8`
  vẫn là cách nói "cố ý cắt bit". Suy-theo-độ-lớn ở vị trí KHÔNG chú thích giữ nguyên (tiền lệ
  [[bug-negative-literal-compare-o0]]). Spec: **RFC 0006 §6.1**; chi tiết + phần CHƯA phủ
  (i64/u64, biểu thức hằng gấp, narrowing từ giá trị runtime):
  [question-out-of-range-narrow-int-literal](question-out-of-range-narrow-int-literal.md). Commit `abfe985`.

## 🐞 BUG MỞ — probe4 tìm ra 2026-07-30, ĐÃ TỰ XÁC MINH LẠI (không chỉ theo báo cáo agent)
Cả ba đều **accept-then-miscompile** (lớp BUG#53) và đều có ma trận đối chứng trong
`bin/probe4/`. Repro chạy bằng `sh bin/probe4/run.sh <file.ax>`.
⚠️ Driver cần subcommand `build`: `bin/axc_native.exe build f.ax -o out.exe -O0`. **Thiếu `build`
thì nó exit 0 mà KHÔNG sinh file** — đừng đọc đó là build thành công (tôi đã mắc đúng bẫy này).

1. ⛔⛔⛔ **`OP_ICONST` mang `type_id` FLOAT** — `let a: f64 = 3` ra **0.0**. Xác minh: `probe4/h1.ax`
   -O0 exit **1** (đúng phải 42). Root: `air_builder.ax:572-589` `lower_int_lit` phát `OP_ICONST` với
   type_id 9/10, còn sibling `lower_float_lit` (`:591`, emit `:601-614`) làm đúng — **bản sao int
   chưa từng được mở rộng**. Breadth CỰC RỘNG: `let`/assign/param/field-init/array-elem/`return`/
   `3 as f64`/`c + 3`/`3 / 2.0`/`c > 2`. -O0 và -O1 **phân kỳ** (họ aliasing `R10 ≡ XMM10`) ⇒ oracle
   phải chạy CẢ HAI. ⚠️ `knowledge/bugs.md:1015-1019` khẳng định *"`let x: f64 = 3` → OK"* — **câu đó
   SAI và chưa từng được kiểm**. **ĐANG SỬA** (sub-agent, kèm invariant §9: `OP_ICONST` không được
   mang type_id float).
2. **Dynamic dispatch bỏ HẾT coercion cho ĐỐI SỐ** — `i.c32(1.5)` callee nhận đúng **0.0**. Xác minh:
   `probe4/f1.ax` exit **61** (đúng phải 42). Root: `typecheck.ax:4336` chỉ phân giải **kiểu KẾT QUẢ**
   từ contract; `interface_method_sig` (`:1742`) chỉ expose `out_nparams`+`out_ret`, **không có
   accessor cho kiểu THAM SỐ** ⇒ args suy với `expected = TYPE_UNKNOWN`. **Cùng họ hệt hai bug đã sửa
   hôm nay** (`376af08` kiểu trả về, `0bf34ee` type arg tường minh) — lần này là THAM SỐ. Frontend ⇒
   **A==B**. Dấu hiệu "đúng 0.0 chứ không phải rác": low 32 bit của double 1,5 là 0, `movss` đọc đúng
   chỗ đó. Int narrowing "đúng" chỉ do **tình cờ** (sub-register GPR trùng nhau, không cần lệnh nào).
3. **`let a: i64 = 3.0` được NHẬN, cho ra bit pattern IEEE thô** — `probe4/g13.ax` exit **3** (i64 giữ
   `0x4008000000000000`). **Spec ĐÃ chốt hướng này**: `knowledge/bugs.md:1015` — *"float → int: CẤM
   ngầm, phải `as`"* ⇒ **REJECT + chẩn đoán**, KHÔNG cần user quyết (khác ca `u8 = 300` ở §CẦN USER
   QUYẾT, ca đó chưa có phán quyết). Frontend ⇒ A==B.

## ✅ Bề mặt đã QUÉT SẠCH (probe4, bankable làm oracle — đừng quét lại)
Option/Result payload × type class (`a1`–`a4`), generics × type class kể cả type arg tường minh
(`b1`,`b2`), str/bytes qua interface (`b3`), aggregate × float + global float (`c1`,`c2`), ABI float
qua biên 6 đối số + xen kẽ int/float (`c3`), cast/độ chính xác f32 kể cả chia single-rounded
(`c4`,`d1`), ADT user payload non-i64 (`e1`), tuple có float/str/struct (`e2`), stdlib container ở
element non-i64 (`e3`), interface breadth (`e4` — chỉ hàng f32 fail = bug #2).

## 🔜 TASK MỞ (tự làm được, theo thứ tự giá trị)
0. **int → f64 tại các VỊ TRÍ mà typecheck không lan hint** (field f64, param f64, param method f64,
   biểu thức `let c: f64 = 3 + 1`) → [bug-int-literal-float-type-iconst](bug-int-literal-float-type-iconst.md)
   mục "STILL OPEN". Bản **f32** của cả bốn đều ĐÚNG vì `typecheck.ax:5208` chỉ hint `TYPE_F32` ⇒
   lại là "hai bản sao trôi lệch". Sửa ở typecheck ⇒ **có thể đổi codegen của chính compiler**
   (fixpoint không miễn nhiễm) ⇒ chạy full gate. Hàng 4/8/9/11 đã đặt chỗ sẵn trong
   `bin/t_intlitfloatctx.ax`. Probe: `bin/probe4/{i1,i2,h4,g5,g6,g7,g13}.ax`.
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
commit**. Regression: `scripts/regression_repros.sh` (≥593), có lượt `-O0`.
