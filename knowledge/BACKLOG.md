# BACKLOG — orientation file (READ THIS, not MEMORY.md)

⚠️ **Đây là FILE CON TRỎ, không phải kho sự thật.** Mọi chi tiết/bằng chứng nằm trong
`knowledge/<topic>.md` và `knowledge/MEMORY.md`. Nếu hai file nói khác nhau, **`MEMORY.md` đúng** —
và hãy sửa file này ngay, đừng để trôi lệch (hai bản sao của một sự thật chính là lớp defect đã
sinh ra bug interface-return).

**Vì sao file này tồn tại:** `knowledge/MEMORY.md` = **175 KB ≈ 87k token**, vượt read cap; chỉ một
trang bị cắt của nó đã tốn ~25k token mỗi phiên **trước khi làm bất cứ việc gì**. Định hướng bằng
file này (~2k token) + handoff mới nhất; vào `MEMORY.md` **chỉ bằng `Grep`** theo tên topic.

---

## Trạng thái cây (cập nhật 2026-09-05b — TẠM DỪNG, cây KHÔNG sạch)
- ⛔ **ĐỌC `knowledge/session-handoff-2026-09-05b.md` TRƯỚC.** Cây còn **thay đổi B3b CHƯA COMMIT và
  CHƯA QUA GATE**; bản lưu ở nhánh **`wip/b3b-current-tree` (`f71f40c`, local, UNVERIFIED)**.
- 🆕 **B3b — ĐÃ TÌM RA NGUYÊN NHÂN GỐC, chưa kiểm chứng:** `ax_driver_load_module` ghi đè
  `symtable.current_tree` và không khôi phục; nạp module là **TÁI NHẬP** nên khi quay ra, thanh ghi
  còn trỏ vào module A ⇒ mọi symbol B định nghĩa **sau dòng `import`** bị đóng dấu thuộc cây của A.
  Giải thích cả "kích hoạt chỉ theo SỐ NODE của A" lẫn `undefined name 'aparam'`.
- Driver `bin/axc_native.exe` **nguyên vẹn**: A==B `29DB6D68…`, **2.330.112 byte** (đã đo lại sau khi
  giết agent). Baseline **714/714** là số của phiên trước — **phiên 09-05b không chạy trọn suite**.

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

## ✅ probe4 (2026-07-30) — **CẢ BỐN ĐÃ ĐÓNG 2026-07-31** → `knowledge/probe4-closed-bugs-2026-07-31.md`
#1 `OP_ICONST` mang type_id float (`6febd02`, sửa ở **tầng lowering** chứ không phải hint typecheck) ·
#2 dispatch bỏ coercion đối số (RFC 0029 §9, gate thực tế **A==B==C**) · #3 `let a: i64 = 3.0` nay
`error[E3031]` · #4 **GIẢI THỂ** — ⛔ **đừng đi săn "bug thanh ghi float"**, nguyên nhân thật là khớp
**substring chặn bởi `_`** khi phân giải tên method (`a281992`); thanh ghi cũ chỉ là **hệ quả**.
⚠️ Bẫy driver ghi ở đó vẫn sống: **thiếu subcommand `build` thì driver exit 0 mà KHÔNG sinh file.**

## 🐞🐞 probe8 (2026-07-31, sau `a281992`) — **3 LỖ SỐNG** → `knowledge/probe8-scan-2026-07-31-live-holes.md`
⭐ **Bài học bao trùm:** `a281992` cho tên CHÍNH XÁC **thắng điểm** cửa sổ lỏng, nhưng **không chặn cửa sổ lỏng
LÀM CÂU TRẢ LỜI khi không có tên chính xác nào** — **sửa một nửa của một luật**. Repro ở `bin/probe8/`.
✅ ĐÃ SỬA: rank-1 loose (`a538983`) · overload trên receiver dạng gọi **TỰ DO** (`52D1ABD4`, RFC 0035 §2bis).
🔴 **CÒN MỞ (mỗi cái một item):**
- gọi bằng **cú pháp method** bind theo RECEIVER, **bỏ qua arity** ⇒ [[bug-method-call-overload-ignores-arity]].
  ⚠️ **Phân kỳ theo mức tối ưu**: `g10_seq` = **2 ở -O0/default, 42 ở -O1** — **-O1 chỉ CHE bằng inlining**.
- `check_iface_conformance` không phủ **gán** và **payload Option** ⇒ SEGV ⇒ [[bug-iface-conformance-unchecked-sites]].
- method generic mono hoá trả `-> T` với T ↦ f64 ra **0.0** ⇒ [[bug-mono-generic-ret-typaram-f64]].
- **chưa có file riêng:** kiểu **TRẢ VỀ** của operator method không được kiểm (`eq -> f64` vẫn làm `==` ra true);
  `h1_rank2_eq`/`h2_axstd_prefix` (**rank 2**, tên user bắt chước mangle — sửa theo **D1-3**: khoá vào **danh tính**
  instantiation, ⛔ **KHÔNG** hạn chế theo prefix `_AX_std_`); `resolver.ax:653/662` chọn giữa nhiều `M.foo`
  là **tuỳ tiện theo thứ tự slot hash** — ghi lại để không ai tưởng đã kiểm.

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

## 🐞 `print`/`println` — họ defect P1–P6 → **`knowledge/bug-println-arity-and-type-defects.md`**
**P1/P2/P3 ĐÃ SỬA** (RFC 0038, A==B `99F795C2…`, baseline 677 lúc đó). **P5 không phải bug** (codepage
console, không phải lexer/string — đừng đi săn). **P6 ĐÃ SỬA** 2026-08-06 (A==B `A58F762A…`).
🔴 **CÒN MỞ — P4:** hàm `println` do **user định nghĩa** bị selector cướp vì khớp theo **chuỗi tên**
(`x86_selector.ax:1730`→`:1737`) chứ không theo **danh tính symbol** — cùng lớp defect với RFC 0037.
Sửa đúng = đưa việc chọn symbol runtime ra khỏi selector ⇒ **chạm backend ⇒ bắt buộc B==C**.
⚠️ Item phụ chưa xử lý: `cgen.ax:1527-1541` có bảng phân loại kiểu **HẸP HƠN** selector x86.

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

## 🧭 Stdlib reachability — option C, 4 stage → `knowledge/stdlib-reachability-option-c-stages.md`
**Stage 1 ✅ XONG 2026-08-07** (A==B `9C6726C1…`): `strip_imports` bôi trắng **chỉ khi khớp CHÍNH XÁC**
một mục của bảng DUY NHẤT `preprocessed_module_name()` (11 mục) ⇒ hết "hai danh sách"; B2 rơi ra miễn phí.
⛔⛔ **THỨ TỰ STAGE GỐC LÀ SAI** (đo 2026-09-05) — xem `bug-b3b-cross-module-index-and-loader-defects.md`
§5–§7: **stage 4 là TIỀN ĐỀ của stage 2** (thiếu danh tính theo module ⇒ stage 2 buộc khớp theo chính tả
⇒ **vi phạm D1-3**, vỡ trên overload `map`/`unwrap`); **stage 3 PHỤ THUỘC stage 2**, không rẻ hơn.
**Stage 2 BẮT BUỘC RFC.** Thứ tự đúng: B3b-3 ✅ → B3b-2 ✅ → B3b (chỉ số xuyên cây) →
`strip_package_prefixes` biết-literal (đóng B1, mở khoá CỘT, **không cần RFC**) → *rồi mới* stage 2.
⛔ **KHÔNG làm option B** (nhồi thêm module vào bundle) — đã đo và bác bỏ; lý do + số đo DFE ở file trên.

## 🐞 BUG PHỤ tìm được cùng lúc (mỗi cái filable riêng)
- **B2 ✅ ĐÃ SỬA (stage 1)** `import stdthing` bị xoá oan — `match_prefix(s,i,"import std")` **không
  có biên phân cách**. Rơi ra miễn phí khi đổi sang so khớp **chính xác** với bảng module.
  Oracle `bin/t_stdprefixmod.ax` (+ fixture `std_util.ax` ở gốc repo): reject trước, 42 sau.
- **B3 + B6 ✅ ĐÃ SỬA 2026-09-05 — HAI CÁI LÀ MỘT BUG.** `pre_infer_func_signature` đánh chỉ số
  `decl_node` của unit KHÁC vào `self.tree` ⇒ compiler SEGV 139. Chi tiết, ma trận kích hoạt, thí
  nghiệm nhồi và cách sửa: **`knowledge/bug-cross-tree-decl-node-segv.md`**.
- **B3b 🔴 CÒN MỞ — và cơ chế ghi ở lộ trình đã bị BÁC BỎ (đo 2026-09-05).** Nạp trùng module là
  hiện tượng **đi kèm, không phải nguyên nhân**: bản `--no-stdlib` (không hề trùng) vẫn hỏng y hệt,
  còn trùng `std.result` thì **build sạch**. Nguyên nhân thật là **chỉ số node xuyên cây** (đúng họ
  B3/B6, đường mono/typecheck mà bản vá B3 chưa phủ) và **hai module user thuần tái hiện được** ⇒
  **không phải vấn đề stdlib/bundling**. ⛔ **Stage 2 như đề xuất VI PHẠM D1-3** (buộc phải khớp theo
  chính tả; vỡ trên overload `map`/`unwrap`) — **stage 4 là TIỀN ĐỀ của stage 2**, lộ trình ghi ngược.
  Kèm hai defect loader mới: **B3b-2** UAF ở `main_air.ax:1957`, **B3b-3** cổng B5 kiểm quá muộn.
  Toàn bộ đo đạc, thứ tự khuyến nghị và phán quyết RFC: **`knowledge/bug-b3b-cross-module-index-and-loader-defects.md`**.
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

## 🐞 CLI — CỜ KHÔNG NHẬN DẠNG BỊ **NUỐT IM LẶNG** (đo 2026-09-05, chưa sửa)
Chuỗi `if/elif` phân tích tuỳ chọn ở `main_air.ax:1058-1169` kết thúc ở `-ctgc-free-report`
**KHÔNG có `else` cuối** ⇒ mọi tham số không khớp nhánh nào **rơi tọt qua**, vòng lặp đi tiếp.
⇒ `axc foo.ax --no-stdib` (gõ sai) biên dịch **CÓ stdlib**, báo thành công, và **không hề nói** rằng
cờ vừa nhận là vô nghĩa. Tương tự với `-O2`, `--target`, `--shared` gõ sai.
⭐ Đây là **lớp accept-then-miscompile (BUG#53) áp vào CLI**: user nhận một bản build **không phải**
bản họ yêu cầu, mà không có tín hiệu nào.
⚠️ **Hệ quả cho việc kiểm thử**: mọi cờ mới "chạy được" ngay cả trước khi hiện thực ⇒ oracle phải
assert **HIỆU QUẢ**, không bao giờ assert việc cờ **được chấp nhận** (đã ghi vào RFC 0040 §7.1).
Chưa sửa: bề mặt breakage riêng (mọi script đang truyền cờ mà compiler này không biết sẽ bắt đầu
fail — đúng mục đích, nhưng phải audit riêng).

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
⚠️ **ĐÍNH CHÍNH câu "chẩn đoán tốt" ngay trên (đo lại 2026-09-05, có cặp control):** chẩn đoán đó
**KHÔNG im lặng — nhưng nó NÓI SAI NGUYÊN NHÂN** cho đúng ba module *đã* bundled:
```
ax_assert(true, "x")            ⇒ build exit 0, chạy, in 42     // bare  ⇒ CHỨNG MINH nó ĐÃ bundled
std.runtime.ax_assert(true,"x") ⇒ "unresolved call ... on undefined namespace 'std'
                                   -- module not imported or not bundled on this build"
```
Câu *"not bundled on this build"* **đúng** cho `std.math` (thật sự không bundled), nhưng **SAI** cho
`std.result` / `std.runtime` / `std.collections` — chúng **có** trong bundle (`preprocessed_module_name`
0/3/7) và bare name chạy được. §8 đòi chẩn đoán **actionable**; câu này đẩy user đi bundle một thứ
đã bundled rồi, thay vì nói sự thật: **tiền tố không nằm trong danh sách rewrite**.

📏 **Con số chính xác của "hai danh sách" (đo 2026-09-05):** stage 1 đã hợp nhất `strip_imports` với
bảng `preprocessed_module_name` (**11** mục) — nhưng **`strip_package_prefixes` (`main_air.ax:411-419`)
vẫn giữ danh sách ĐỘC LẬP riêng của nó, chỉ **8** tiền tố**: `mem.alloc, scheduler, os.win32,
os.linux_sys, os, string, io, net`. **Thiếu đúng 3: `std.result.`, `std.runtime.`, `std.collections.`**
⇒ đó là toàn bộ nguyên nhân của hàng `std.collections` trong ma trận trên. **"Hai danh sách" CHƯA hết**,
mới hết ở một trong hai chỗ.

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
