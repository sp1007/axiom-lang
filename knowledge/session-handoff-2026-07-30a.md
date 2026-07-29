---
name: session-handoff-2026-07-30a
description: "HANDOFF 2026-07-30a — HEAD 7509184 (code cuối ở bcf7746), driver A==B+B==C 42F49C73, 575/575 (+lượt -O0 575/575), KHÔNG còn bug OPEN. Đã SỬA 3 silent bug: (1) literal nguyên không vừa i32 vẫn bị type i32 (-O0 sai kết quả); (2) aggregate local không initializer bind vào NULL (segfault mọi opt level); (3) HOLE#7 hàm user trùng tên stdlib + param interface không hề được gọi. CẢ 4 SHAPE nay trong mốc M6-codegen (fib 1,05x so floor NGHIÊM NHẤT sau khi fold copy tham số, −13,9%). Ship: peephole 1d+1e, căn lề hàm 16-byte, LEA/MOV fold sang thanh ghi vật lý. Floor cả 4 shape ĐÃ soi (fib: full variant study; 3 shape kia: alignment + biến thể cấu trúc) và ĐỨNG VỮNG. Còn 1 câu hỏi cho USER: đo mốc bằng perf counter thay wall-clock?"
metadata:
  type: project
---

# HANDOFF 2026-07-30a — **ĐỌC ĐẦU TIÊN**

> ⚠️⚠️ **CÁCH ĐỌC FILE NÀY**: phần thân được viết **THEO THỨ TỰ THỜI GIAN**, nên nhiều kết luận
> ở GIỮA file **ĐÃ BỊ THAY THẾ** bởi kết luận sau đó (đặc biệt: "mốc chưa đạt / fib 1,18x ❌"
> và "căn lề đo ra 0, đã revert" — **CẢ HAI ĐỀU ĐÃ LỖI THỜI**). **Khối TÓM TẮT ngay dưới đây
> là trạng thái ĐÚNG.** Chỉ đọc phần thân để lấy *bằng chứng và bài học*, đừng lấy kết luận.

## ✅ TÓM TẮT CHỐT — trạng thái ĐÚNG tính đến `5967dc8`
- HEAD **`5967dc8`** đã push; driver `bin/axc_native.exe` = **`6C9165C8`** (`B==C`; A!=B bình thường vì codegen đổi).
  Gate: **575/575**,
  **lượt `-O0` toàn suite 575/575**, **KHÔNG còn bug OPEN**, ELF 12/12, ctgc 16/16, exe_size 4/4,
  lib_collision 6/6, so_export ✓.
- 🐞→✅ **ĐÃ SỬA (HOLE#7, `bcf7746`)** — **KHÔNG còn bug OPEN nào**:
  [[bug-user-fn-builtin-name-iface-param]] — hàm user trùng tên hàm stdlib bundled
  (`get`/`len`/`find`/`push`) **VÀ** có tham số kiểu **INTERFACE** ⇒ **overload resolution chọn
  bản STDLIB**, hàm user KHÔNG hề được gọi. `fn get(sh: Shape)` → **0** (đúng 36);
  `fn len(sh: Shape)` → **8** (= đọc length từ `len(s: str)`). Đổi tên → đúng ngay.
  ⭐ Arity giải thích trọn bảng: `set` an toàn vì bản stdlib cần **3** tham số.
  ⭐ Chỉ INTERFACE param mới thua — `fn get(p: Sq)`/`(s: str)`/`(x: i64)` đều ĐÚNG ⇒ giá trị
  interface khớp **lỏng** với param generic (`T`/`Vec[T]`/`MutexGuard[T]`) theo cách struct cụ thể
  không khớp — cơ chế THẬT (đã probe): `arg0_type` suy từ **biểu thức đối số** nên là struct
  CỤ THỂ, còn param là interface ⇒ `is_method_compatible` không biết struct implement interface ⇒
  **`n_gated==0`, không ứng viên nào được nhận** ⇒ rơi xuống `return sym_idx` = HEAD **sai arity**.
  ⭐ **FIX** (`typecheck.ax:1241`): thêm MỘT nhánh accept — `param[0]` là `TYPE_KIND_INTERFACE` và
  `arg0` là struct implement nó (`interface_missing_method==0`). Chỉ NỚI tập accept ⇒ không thể
  làm hẹp ứng viên đang đúng. Cả 5 tên nay trả **36** (kể cả `push` trước đó BUILD FAIL).
  Oracle `t_ifacefnbuiltinname`(36) **đã đăng ký** ở `-O1` và khối `-O0`.
  ⚖️ Lưới `n_gated==0` **cố ý KHÔNG làm** — sau fix không còn input nào chạm ⇒ guard không thể nổ.
- 🐞→✅ **SỬA THÊM một CRASH im lặng** (probe-found cuối phiên):
  [[bug-uninitialized-local-array-segfault]] — aggregate LOCAL khai báo **không initializer**
  (`mut arr: [i64;2]`, `mut p: P`) bị bind vào **OP_ICONST 0 = con trỏ NULL** ⇒ store phần tử
  đầu tiên **SEGFAULT ở MỌI opt level, không một chẩn đoán nào**. Fix 1 site
  (`air_builder.lower_var_decl`): aggregate phát **`OP_ALLOC`** thay vì `OP_ICONST`; scalar giữ
  nguyên đường cũ. **Phủ struct + array + tuple cùng lúc** (sửa riêng array = partial fix).
  ⭐ Tôi từng đóng khung nó là "câu hỏi thiết kế cần user quyết" — **SAI**: local không-init
  **đã chạy đúng với scalar/str** nên hướng sửa đã được ngôn ngữ quyết sẵn.
- 🐞→✅ **ĐÃ SỬA một SILENT MISCOMPILE** tìm được cuối phiên bằng probe:
  [[bug-negative-literal-compare-o0]] — literal nguyên không vừa i32 vẫn bị **TYPE là i32**
  ⇒ `emit_wrap_to_width` cắt mọi bit trên 32 ⇒ `if c == -3000000000` **rẽ SAI nhánh** và
  `to_str(i64::MIN)` in ra `"0"`. **Chỉ ở `-O0`** (const-fold SSA ở -O1+ che mất).
  Fix theo **ĐỘ LỚN** ở `typecheck.ax` ~L5595 (literal vừa i32 giữ NGUYÊN kiểu ⇒ không xáo trộn
  overload / `t_u64cmp`; chỉ đổi những literal mà kiểu VỐN ĐÃ SAI).
- ⭐⭐ **HẠ TẦNG MỚI: khối `-O0`-only trong `regression_repros.sh`** (`t_negbiglitcmp`+`t_tostr`).
  **`t_tostr` đã bắt được bug này từ NGÀY NÓ ĐƯỢC VIẾT** — chỉ là suite build ở `-O1` trở lên nên
  **chưa bao giờ chạy `-O0`**. Muốn quét rộng: `AXEXTRA=-O0 scripts/regression_repros.sh`
  (xanh tính đến hôm nay). ⇒ **Bài học: một lớp defect có thể vô hình chỉ vì gate không bao giờ
  chạy ở cấu hình đó — không phải vì thiếu test.**
- **CẢ 4 SHAPE ĐỀU TRONG MỐC M6-codegen (≤15%)**, đo tiền định sau khi đã căn lề hàm:

  | shape | vs floor | ghi chú |
  |---|---|---|
  | fib | **1,05x** | so floor **NGHIÊM NHẤT** (V1); so V3 = 1,03x; so V0 = 0,98x |
  | xorshift | **1,00x** | |
  | arrwalk | **1,08x** | |
  | callloop | **1,08–1,12x** | |

- ✅ **FLOOR CỦA 3 SHAPE CÒN LẠI ĐÃ ĐƯỢC SOI BẰNG BIẾN THỂ CẤU TRÚC — chúng ĐỨNG VỮNG**
  (`scripts/floor_struct_variants.ps1`, ghép cặp 3 vòng xen kẽ). Thử dạng KHÁC VỀ CẤU TRÚC, đúng
  loại thí nghiệm đã lộ khiếm khuyết V0 của fib:
  | shape | floor | biến thể | delta |
  |---|---|---|---|
  | xorshift | 215,5 | 214,2 (unroll ×2) | −0,61% |
  | arrwalk | 341,4 | 337,3 (unroll ×2) | −1,20% |
  | callloop | 60,9 | 61,4 (loop control 2 lệnh thay vì 3) | **+0,78% CHẬM HƠN** |
  ⇒ Không biến thể nào nhanh hơn đáng kể; cải thiện lớn nhất (arrwalk −1,2%) chỉ đẩy tỷ lệ AXIOM
  từ 1,08x lên ~1,09x — **vô nghĩa với gate 15%**. ⭐ callloop: bỏ MỘT lệnh khỏi loop control lại
  **CHẬM HƠN** — thêm một lần "số lệnh ≠ chi phí".
  ⚠️ **Giới hạn trung thực**: mỗi shape chỉ thử MỘT biến thể cấu trúc (unroll / dạng loop control),
  không vét cạn. Nhưng cộng với thí nghiệm căn lề (đã bác) và lập luận cấu trúc (cả ba thân vòng
  đã là chuỗi phụ thuộc tối thiểu), **kết luận mốc KHÔNG đổi**.
- **Đã ship 5 thay đổi codegen**: peephole **1d** `fold_alu_immediate`, **1e**
  `strength_reduce_imul`, **căn lề 16-byte cho function entry**, **LEA/MOV fold sang thanh ghi
  VẬT LÝ**, và **fold so-sánh-với-0** (`XOR_ZERO v ; CMP n,v` → `CMP n,imm(0)`).
  fib cải thiện ~17% toàn phiên; callloop 1,28→1,08x; arrwalk 1,14→1,08x.
- 🆕 **SHAPE BENCHMARK MỚI: `bin/t_tailrecloop.ax`** (tail-recursive, 20M call) + floor NASM trong
  `scripts/price_tailrec.ps1`. **Suite trước đây KHÔNG có shape tail-recursive nào** ⇒ cả một lớp
  khe hở không đo được. Nó lộ ra NGAY:
  - ⭐ **AXIOM ĐÃ tự chuyển self-tail-recursion thành LOOP** (`sumto` phát ra KHÔNG có `call`) ⇒
    **RFC 0036 bị RÚT** vì đề xuất thứ đã có (xem `rfcs/0036-*.md`; `t_selfrec`/`t_selfrec2` đã
    đặt tên đúng việc này — **phải đối soát với thứ ĐÃ SHIP trước khi viết RFC**).
  - ⭐ **Fold so-sánh-với-0** (−5,85% đo ghép cặp 4 cặp; NASM dự báo 8,83% ⇒ **dự báo lạc quan
    ~1/3**). 4 shape cũ PHẲNG vì không shape nào có so-sánh-với-0 nóng.
  - 📍 **KHE HỞ CÒN LẠI ĐÃ ĐO: AXIOM 21,4 ms vs loop floor 16,2 ms = 1,32x** trên shape này.
    Nguyên nhân đọc từ disassembly: **shuffle tham số** (`mov %rdx,%rax ; mov %rbx,%rcx` mỗi vòng)
    + prologue mà bản viết tay không cần. ⇒ **Mục tiêu codegen kế tiếp, đã có shape để đo.**
- **Việc kế tiếp đề xuất** (backlog **KHÔNG còn bug OPEN**; mục (a) cũ — biến thể cấu trúc cho 3 floor — **ĐÃ LÀM XONG**, xem mục FLOOR ở trên): (b) ⛔ **M6-opt: RFC 0036 ĐÃ BỊ RÚT** — transform tôi đề xuất (**self-tail-recursion → loop**) **ĐÃ ĐƯỢC CÀI SẴN**: `sumto` phát ra KHÔNG có `call` nào, chỉ `jmp` về entry (`t_selfrec`/`t_selfrec2` trong suite đã đặt tên đúng việc này — tôi lẽ ra phải đọc TRƯỚC khi viết RFC). ⭐ **Kết quả ĐO được vẫn có giá trị**: trên shape tail-recursive MỚI (20M call, ghép cặp 3 vòng) — AXIOM **22,6 ms** vs **loop floor 16,2 ms** = **1,40x**; khe hở KHÔNG phải thiếu opt pass mà là **shuffle tham số** (`mov rdx,rax; mov rbx,rcx`) + prologue ⇒ thuộc **M6-codegen**, cùng họ các copy fold đã ship phiên này, và là khe hở MỚI trên shape suite chưa từng có. ⇒ Việc kế tiếp cho M6-opt: (i) đóng khe shuffle tham số cho hàm tail-recursive (codegen, rẻ, nay ĐO ĐƯỢC), hoặc (ii) xét lại accumulator transform cho đệ quy KHÔNG-tail như `fib` (các phản đối ở §2 RFC 0036 vẫn đứng); **CỐ Ý TỪ CHỐI** accumulator transform (đòi reassociate recurrence, KHÔNG hợp lệ trên float, và chỉ có `fib` được lợi ⇒ bẫy "tối ưu cho benchmark"). ⚠️ RFC nói rõ: **bản thân nó có thể KHÔNG cải thiện shape nào hiện có** vì `fib` không phải tail-recursive ⇒ **việc ĐẦU TIÊN là thêm một shape benchmark TAIL-RECURSIVE + NASM floor rồi ĐỊNH GIÁ, TRƯỚC khi viết pass**; nếu định giá <~5% thì ĐÓNG RFC không cài; (c) `axiom-bug-probe` — **rất đáng làm**:
  phiên này probe ra **3 silent bug** (1 sai kết quả, 1 crash, 1 hàm không hề được gọi) — **cả ba** đều ở vùng **chưa có test** chứ không phải regression, và **cả ba đã được sửa trong phiên**.
- ❓ **Chờ USER**: có nên đo mốc bằng **perf counter** thay vì tỷ lệ wall-clock? Kết quả
  "xoá 1 lệnh = −14%" cho thấy wall-clock ở mức này đang đo front-end nhiều ngang codegen.
- 🧹 **VÙNG ĐÃ QUÉT SẠCH cuối phiên — ĐỪNG probe lại** (batch 3–4, 0 phát hiện):
  (a) **literal lớn trong ngữ cảnh GENERIC**: `ident[T](-3000000000)`, `Box[i64](v: ...)`,
  `Vec[i64].push(...)`, array literal `[-3000000000, 5]` — đúng ở cả -O0/-O1.
  (b) **aggregate không-initializer, các biến thể**: tuple, array-of-struct, struct lồng,
  `Option` — đều đúng sau fix. (`Option` đúng **do cách dùng** — `o = Some(..)` thay cả binding,
  không ghi xuyên qua null — chứ KHÔNG phải do fix; nếu sau này sum có mutation payload tại chỗ
  thì phải xử lý như aggregate.)
  (c) **closure**: capture **global const** (`const N: i64 = -3000000000` trong closure) → **42
  ĐÚNG** cả -O0/-O1. Capture **LOCAL** → **chẩn đoán ĐÚNG ĐẮN** và abort trước codegen:
  *"closure captures 'n' from an enclosing scope; only zero-capture closures are currently
  supported (RFC 0008 P2 not yet implemented)"* ⇒ **KHÔNG phải bug** — là feature chưa cài, có
  lỗi rõ ràng. (Tôi từng tưởng BUILDFAIL này là defect; nó là compiler ĐÚNG.)
  (d) **for-loop trên `Vec`** → 12 đúng; **for trên `str` UTF-8** (`"héllo"`) → 5 đúng, -O0==-O1.
  (e) **interface dispatch** trực tiếp (`let sh: Shape = sq; sh.area()`) → 36 đúng.
  (f) **HashMap với KHOÁ âm lớn** (`m.insert(-3000000000, 7)` rồi `.get(...).unwrap()`) → **42
  ĐÚNG** ⇒ fix literal phủ cả khoá HashMap. (g) **Result** `Ok`/`Err` + match → 42 ĐÚNG.
  (h) `defer` trong **vòng lặp** → **chẩn đoán ĐÚNG ĐẮN, có nêu LÝ DO**: *"`defer` inside a loop or
  conditional is not supported (static registration would run it when the branch is not taken, or
  only once in a loop); move it to the function body top level"* ⇒ KHÔNG phải bug.
  (i) **TỔ HỢP 3 FEATURE** (batch 5) — đều ĐÚNG 42 ở -O0/-O1: generic fn × for-loop × `Vec[Struct]`;
  `defer` × `Result` × `match` (defer ở top-level hàm, có early-return trong nhánh Err);
  `HashMap[i64, Struct]` × giá trị struct × `Option` match.
  ⇒ **Tổng kết probe phiên này: 5 batch. Batch 1–2 ra 3 bug THẬT (đã sửa hết); batch 3–5 SẠCH.**
  ⇒ Bề mặt dễ tiếp cận nay đã sạch; probe tiếp nên nhắm vùng **thật sự chưa chạm**: `spawn`/thread
  + message passing, generic × interface (vtable trên type generic), bignum/float biên,
  `--staticlib`/`--shared` đường multi-lib, và **tổ hợp 3 feature** thay vì 2.

- 🔧 **HARNESS (2026-07-30)**: phiên này chạy **BA** monitor heartbeat cùng lúc (`betmnqwcz`,
  `ba4hfsc7e`, `bnsu3yw53`, cùng một command) ⇒ tick về mỗi ~40 giây thay vì 5 phút. Đã `TaskStop`
  hai cái, giữ một. ⚠️ **Phase 0 của skill nói "nếu CHƯA có monitor thì arm ĐÚNG MỘT" — không phủ
  trường hợp đã có BA.** Sửa lại thành: **đếm** monitor đang chạy; nếu 0 thì arm 1; nếu >1 thì
  `TaskStop` phần thừa. Tick gấp ba vừa tốn việc vừa làm loãng tín hiệu của chính vòng lặp.

### ⭐⭐⭐ BÀI HỌC PHƯƠNG PHÁP CỦA PHIÊN (giá trị lâu dài hơn cả bản vá)
0. **Gate không chạy ở cấu hình nào thì mù ở cấu hình đó.** Suite build `-O1`+ ⇒ cả một lớp
   defect chỉ-ở-`-O0` vô hình, dù **đã có test bắt được nó từ lâu** (`t_tostr`). Không phải
   thiếu test — là thiếu **cấu hình chạy**.
1. **Fix về TÍNH TIỀN ĐỊNH phải đo trên HIỆU GIỮA HAI BẢN BUILD, không đo một bản build.**
   Đo sai kiểu này suýt chôn vĩnh viễn thay đổi căn lề (xem mục căn lề).
2. **Peephole phải được chứng minh là CÓ NỔ, không chỉ AN TOÀN** — bản đầu của 1d khớp 0/4 hằng
   mà vẫn qua sạch mọi gate.
3. **Gate xanh có thể chứng minh SỐ KHÔNG** (3 ca trong phiên: 1d rỗng, thứ tự 1e/1c, và
   RFC 0035 cursor trước đó).
4. **Định giá codegen trên MỘT shape liên tục cho ra "số 0 tự tin"** — immediate-folding bị chấm
   nhiễu trên fib rồi đáng 14% trên callloop.
5. **Disassembly là bằng chứng về BINARY ĐÃ XONG, không phải về mảng mà peephole pre-alloc khớp.**

---

## ⚠️ CẬP NHẬT CUỐI PHIÊN (`60a975a`) — đã ship **peephole 1e**, và mốc VẪN CHƯA ĐẠT, nhưng shape trượt ĐÃ ĐỔI
> ⛔ **KẾT LUẬN TRONG MỤC NÀY ĐÃ BỊ THAY THẾ** — fib sau đó được sửa và nay ĐẠT. Giữ lại vì
> phần **quy trách nhiệm alignment** vẫn đúng và hữu ích.
`IMUL vD,imm(2)` → `ADD vD,vD`; `IMUL vD,imm(2^k)` → `SHL vD,imm(k)`. Driver mới **`A2AD800D`**
(seed==A==B==C), 564/564. Đo ghép cặp 2 vòng:

| shape | trước 1e | sau 1e | |
|---|---|---|---|
| callloop | 1,21 / 1,23 | **1,07 / 1,10** | ✅ nay TRONG mốc |
| arrwalk | 1,14 / 1,15 | **1,08 / 1,09** | ✅ (`imul $8` của index mảng) |
| xorshift | 1,00 / 1,01 | 1,01 / 1,01 | phẳng |
| fib | 1,13 / 1,13 | **1,19 / 1,20** | ❌ **rơi RA NGOÀI mốc** |

⇒ **M6-codegen vẫn CHƯA ĐẠT, nhưng shape chặn nay là `fib` chứ không phải `callloop`.**

⭐ **fib regress là ALIGNMENT, đã QUY TRÁCH NHIỆM chứ không đoán**: fib không có phép nhân nào,
nên 1e không thể đổi instruction selection của nó. Build cả 2 chiều rồi diff: thay đổi lệnh DUY
NHẤT trong binary fib nằm ở **runtime helper bundled** (`imul $0x1000`→`shl $0xc` cho page size,
`imul $0x8`→`shl $0x3` cho mảng con trỏ); **thân hàm fib đệ quy KHÔNG ĐỔI**. `.text` cùng kích
thước, nhưng `shl` mã hoá ngắn hơn `imul` imm32 ⇒ mọi thứ sau đó dịch **16 byte** ⇒ fib rơi vào
alignment khác. **Cùng hiện tượng đã bank 2026-07-29f cho xorshift** (+4,6% từ một thay đổi ÍT
lệnh hơn, chuỗi phụ thuộc không đổi). ⛔ **ĐỪNG "sửa" bằng cách nhét lại lệnh chết.**
⇒ Hệ quả: **con số 1,13x của fib trước đây một phần là MAY MẮN.**

## ✅⭐⭐⭐ CĂN LỀ 16-BYTE CHO FUNCTION ENTRY — **ĐÃ RE-LAND** (`7ac52f5`, `B==C 522BEA6B`)
### ⚠️ ĐÍNH CHÍNH: mục này TRƯỚC ĐÓ ghi "đo ra 0, đã revert, ĐỪNG THỬ LẠI" — **KẾT LUẬN ĐÓ SAI**
Bản thân thay đổi không đổi một dòng; **phép ĐO của tôi mới là thứ sai.**
- **Tôi đã đo**: aligned-có-1e vs **UN**aligned-có-1e ⇒ mọi shape phẳng ⇒ "lợi ích 0" ⇒ revert.
- **Vì sao SAI**: mục đích của thay đổi **không phải làm một bản build nhanh hơn**, mà làm
  **HAI bản build SO SÁNH ĐƯỢC với nhau**. Tác dụng của nó nằm ở **HIỆU giữa hai bản build**,
  nên phép đo một-bản-build **về cấu trúc không thể thấy nó**. Đánh giá một fix về TÍNH TIỀN ĐỊNH
  bằng câu hỏi "có con số nào đẹp hơn không" là **sai phạm trù**.
- **Thí nghiệm ĐÚNG** (fib, best-of-9, 4 cặp xen kẽ mỗi bên):
  | | 1e làm fib chậm đi bao nhiêu |
  |---|---|
  | **KHÔNG** căn lề | **+5,1%** (+3,4…+7,3; **cùng dấu cả 4 cặp**) |
  | **CÓ** căn lề | **−0,2%** (−0,84…+0,51; **vắt qua 0**) |
  ⇒ Độ nhạy vị trí là THẬT, và căn lề **xoá hẳn kênh đó**.
- **Chi phí trung thực** (không phải win miễn phí): fib nay đọc **~605 ms TIỀN ĐỊNH**, so với
  580 ở vị trí unaligned MAY và 609 ở vị trí unaligned RỦI ⇒ nó rơi vào GIỮA dải thay vì bốc
  thăm trong dải đó; tốn ≤15 NOP/hàm (~8 KB).
- ⭐⭐ **HỆ QUẢ CHO MỐC**: con số **1,13x của fib từng ĐẠT gate 15% chính là lần bốc thăm trúng**.
  Số TIỀN ĐỊNH là **~1,18x và TRƯỢT**. Kết quả gate xấu đi vì **phép đo trở nên trung thực** —
  đó là hướng ĐÚNG: một con số chỉ đúng 1 trong 3 lần chạy không phải là điểm đậu.
- ⭐⭐⭐ **BÀI HỌC ĐỂ ĐỜI**: khi thay đổi nhằm mục tiêu *tính tiền định / khả năng so sánh*,
  **phải đo trên HIỆU giữa hai bản build**, không đo trên một bản build. (Chi tiết cài đặt cũ:)
- Cài đặt: pad `0x90` tới bội số 16 trong `all_code` **TRƯỚC** khi chốt `current_offset`
  (`x86_coff.ax` ~L830). An toàn theo cấu trúc: mọi symbol value + reloc per-function đều
  rebase theo `info.offset` nên tự tính cả padding. Đã xác nhận: mọi function entry kết thúc
  bằng `0`, 524 NOP đệm, `B==C 522BEA6B`, **564/564**, ELF/ctgc/exe_size/lib_collision/so_export
  đều xanh. ⇒ **Thay đổi ĐÚNG, không phải bug.**
### 📊 BASELINE TIỀN ĐỊNH (driver `522BEA6B`) — ⛔ **ĐÃ LỖI THỜI, xem khối TÓM TẮT đầu file**
> Bảng dưới đây là baseline TẠI THỜI ĐIỂM ĐÓ (fib 1,18x ❌). Sau đó `1faefc9` + `1a61166` đưa
> fib xuống **1,05x so floor nghiêm nhất** ⇒ **cả 4 shape đều ĐẠT**. Giữ lại để đối chiếu lịch sử.
| shape | vòng 1 | vòng 2 | ≤15%? |
|---|---|---|---|
| fib | 1,18x | 1,18x | ❌ |
| xorshift | 1,00x | 1,00x | ✅ |
| arrwalk | 1,10x | 1,08x | ✅ |
| callloop | 1,08x | 1,09x | ✅ |

⇒ **M6-codegen CHƯA ĐẠT, chỉ còn `fib` chặn.** **DÙNG BẢNG NÀY LÀM BASELINE**, đừng so với
số unaligned cũ. ⭐ Chú ý fib nay khớp **1,18/1,18** — chặt hơn MỌI cặp đo trước đó trong phiên
(vốn trải 1,05…1,20): đó chính là căn lề trả cổ tức bằng **khả năng tái lập**.
⚠️ Suy luận "căn lề entry không lấy lại 1,13x ⇒ thứ nhạy cảm không phải vị trí đầu hàm" **cũng
SAI** và đã bị bác bỏ bởi thí nghiệm hiệu-hai-bản-build ở trên: vị trí đầu hàm **ĐÚNG LÀ** thứ
nhạy cảm; 1,13x đơn giản là không có thật để mà "lấy lại".

⭐⭐ **Thứ tự 1e SAU 1c là load-bearing, và GATE KHÔNG THỂ THẤY ĐIỀU ĐÓ.** 1c khớp theo
`counts[vT]==3`; viết `ADD vT,vT` thêm mention thứ TƯ ⇒ vô hiệu hoá 1c. Trên `x = x * 2` trong
vòng lặp: sau-1c → `add %rax,%rax`; trước-1c → `mov %rax,%rdx ; add %rdx,%rdx ; mov %rdx,%rax`.
**Nhưng compile chính nguồn 2 MB của compiler theo CẢ HAI thứ tự cho ra BYTE-IDENTICAL**
(`413C6A12`), vì trong đó không có chỗ nào viết `x = x * 2^k` dạng compound assignment ⇒
**A==B/B==C vẫn XANH dù thứ tự SAI.** Lại một ca "gate xanh chứng minh số không".

⭐ Flags KHÔNG cần phân tích (giả định ban đầu của tôi là SAI): `IMUL` để SF/ZF/AF/PF
**undefined** theo ISA ⇒ không consumer đúng đắn nào đọc cờ xuyên qua nó.

## 🏁⭐⭐⭐ `1a61166` — FOLD COPY NẠP THAM SỐ: fib **−13,9%**, và MỌI floor đều nói fib ĐẠT
Mở rộng shape B cho producer `MOV vT, <phys>` (copy tham số vào). fib mở đầu bằng
`mov %rcx,%rax ; mov %rax,%rbx` → nay chỉ `mov %rcx,%rbx`. **Chỉ nhận nguồn VẬT LÝ** (thanh ghi
ABI không do allocator vreg định giá ⇒ về cấu trúc không thể tái hiện thất bại của copy-prop cũ).
Driver **`0E24570B`** (B==C), 564/564, phụ trợ xanh hết.

**ĐO GHÉP CẶP 4 cặp xen kẽ** (vì "xoá 1 lệnh được 14%" tự nó KHÔNG đáng tin):
`OFF 587,0 ms → ON 505,2 ms = **−13,93%**` (−12,95…−15,26; cùng dấu cả 4 cặp).
⇒ Một lệnh không thể tốn 14% thời gian chạy ⇒ **hiệu ứng NGƯỠNG ở front-end** (thân hàm ngắn
hơn rơi sang phía tốt của ranh giới fetch/uop-cache). **KHÔNG tuyên bố cơ chế** — cần perf counter.

### ⚠️ ĐÍNH CHÍNH commit `7d5218c`
Tôi đã bank 13,6% chưa giải thích được của W-series là **"không phải codegen"**, vì bản chép
NGUYÊN VĂN sang NASM chạy nhanh hơn binary AXIOM 13,6% với cùng hệt từng lệnh. **Suy luận đó
QUÁ MẠNH.** Hiệu ứng bị kích hoạt bởi **kích thước + nội dung CHÍNH XÁC của hàm phát ra**, nên
nó **CÓ thể với tới từ codegen** — thay đổi một lệnh vừa đẩy fib qua ngưỡng. Khe hở là THẬT;
kết luận của tôi về "cái gì có thể xử lý nó" mới là sai.

### 📐 CÂU HỎI VỀ FLOOR NAY THÀNH VÔ NGHĨA (đo back-to-back)
| | ms | tỷ lệ fib |
|---|---|---|
| AXIOM fib | 510,5 | — |
| V3 (đúng hình dạng AXIOM) | 495,8 | **1,030x** |
| V1 (viết tay nhanh nhất) | 486,4 | **1,050x** |

⇒ **fib ĐẠT gate 15% với MỌI định nghĩa floor, kể cả nghiêm nhất.** Cộng xorshift 1,00x,
arrwalk 1,08x, callloop 1,08–1,12x ⇒ **cả 4 shape đều TRONG mốc M6-codegen.**

⚠️ **NHƯNG CHƯA tuyên bố mốc HOÀN TẤT**: chỉ floor của fib được soi bằng nghiên cứu biến thể.
Ba floor còn lại là **một** bản viết tay duy nhất, **chưa hề kiểm tra xem có phải bản nhanh nhất
của thuật toán đó không** — đúng khiếm khuyết đã phát hiện ở V0 của fib. **Đã kiểm một phần** (`scripts/floor_check_variants.ps1`): giả thuyết khả dĩ nhất — **căn lề THÂN VÒNG LẶP** (entry đã căn, nhưng đích nhảy lùi thì không) — **BỊ BÁC BỎ cho cả ba**: xorshift −0,09%, arrwalk +0,85%, callloop +0,30%, đều vắt qua 0. Cộng lập luận cấu trúc: thân vòng của cả ba **đã là chuỗi phụ thuộc tối thiểu** (xorshift = 3 bước shift-xor phụ thuộc; arrwalk = pointer-chase thuần ở độ trễ load; callloop = lea→add→and 3 chu kỳ). ⚠️ **Nhưng đây là MỘT giả thuyết, không phải nghiên cứu biến thể đầy đủ như fib** — bằng chứng YẾU HƠN. Muốn chốt mốc thì thử thêm dạng khác về CẤU TRÚC (như V1 của fib khác V0 ở chỗ có rbp).


## Trạng thái (CUỐI PHIÊN — con số CHÍNH XÁC ở đây, các mục bên trên là lịch sử theo thứ tự thời gian)
- HEAD **`1a61166`**, đã **push lên `origin/main`**, cây sạch (chỉ `.claude/settings.json`
  untracked — **của user, đừng đụng**).
- Daily driver `bin/axc_native.exe` = **`0E24570B`** (B==C; A!=B là bình thường vì codegen đổi).
- Gate đầy đủ XANH: regression **564/564**, ELF 12/12, ctgc 16/16, exe_size 4/4,
  lib_collision 6/6, so_export ✓.
- ⚠️ **Baseline TIỀN ĐỊNH mới nhất** (so floor V0 hiện tại): fib **1,14/1,15**, xorshift
  **1,00/1,01**, arrwalk **1,09/1,10**, callloop **1,08/1,09**. ⛔ **CHƯA tuyên bố mốc ĐẠT** —
  xem mục floor bị đặc tả sai + 13,6% không phải codegen.

## Đã ship phiên này (6 commit code + memory)
1. `ae8516a` — **NASM floor cho arrwalk + callloop** (chỉ `scripts/perf_suite.ps1`).
2. `1172648` — **peephole 1d `fold_alu_immediate`** + 4 encoder mới (`and_ri` /4, `or_ri` /1,
   `xor_ri` /6, `imul_ri` 0x69) + oracle `t_aluimmfold`(42).
3. `60a975a` — **peephole 1e `strength_reduce_imul`** (`IMUL imm(2^k)` → ADD/SHL).
4. `7ac52f5` — **căn lề 16-byte cho function entry** (re-land sau khi revert nhầm).
5. `1faefc9` — **LEA copy-fold nhắm được thanh ghi VẬT LÝ** (bỏ 2 copy/call).
6. `cbdb873` + `7d5218c` — định giá lại bằng NASM (`scripts/price_fib_variants.ps1`,
   `scripts/price_fib_wseries.ps1`).

## ❓ HAI CÂU HỎI ĐANG CHỜ USER (đã nêu, chưa có trả lời — **đừng tự quyết**)
1. **Floor của fib nên là bản nào?** V0 (hiện tại, tâng bốc ta) / V3 (đúng hình dạng AXIOM) /
   **V1 (nhanh nhất cùng thuật toán — tôi ĐỀ XUẤT)**. Đây là đổi cách đo một mốc do USER đặt (D1).
2. **Có nên đo mốc bằng PERF COUNTER thay vì tỷ lệ wall-clock?** Với số hạng môi trường 13,6%,
   gate 15% gần như **không thể bác bỏ được** bằng wall-clock.

## ⭐ KẾT QUẢ LỚN NHẤT: mốc M6-codegen nay ĐO ĐƯỢC ĐẦY ĐỦ — và **CHƯA ĐẠT**
Trước phiên này chỉ fib/xorshift có floor nên **không thể kết luận**. Nay cả 4 shape đều có
floor NASM cùng hình dạng. Đo ghép cặp, 2 vòng không chồng lấp, **sau** khi ship 1d:

| shape | vs asm floor (2 vòng) | ≤15%? |
|---|---|---|
| fib | 1,13x / 1,13x | ✅ |
| xorshift | 1,00x / 1,01x | ✅ |
| arrwalk | 1,15x / 1,14x | ✅ (sát mép) |
| **callloop** | **1,21x / 1,23x** | ❌ **MISS** |

⇒ **callloop là shape DUY NHẤT còn chặn mốc M6-codegen.** (Trước 1d: 1,28x/1,29x.)

⚠️ Máy chạy chậm hơn ~5% ở vòng đo thứ hai — **mọi cột floor cố định đều tăng** (fib floor
519→548, xorshift 216→222, arrwalk 338→362). Vì vậy **chỉ tin TỶ LỆ, đừng tin ms tuyệt đối**
giữa các phiên. callloop giảm 79,2→77,4 ms tuyệt đối TRONG KHI máy chậm đi ⇒ win là thật.

## ✅ CẬP NHẬT `1faefc9` — LEA copy-fold nay nhắm được THANH GHI VẬT LÝ
Mở rộng shape B của `coalesce_dest_copy` cho `b1.dst.kind == OPND_PHYS`: `LEA vT,[..] ; MOV rcx,vT`
→ `LEA rcx,[..]`. Trước đây chỉ nhận VREG nên **mọi call site** đều giữ lại copy nạp tham số.
fib bỏ được **2 copy mỗi lời gọi đệ quy**. Driver **`B289E6C4`** (B==C), 564/564, phụ trợ xanh hết.
- Đo 2 vòng: fib tuyệt đối **611,6/607,9 → 602,5/598,4 ms (−1,5%)**; tỷ lệ 1,18 → 1,14/1,15.
- ⚠️ **ĐỌC TỶ LỆ CẨN THẬN**: floor của fib tự nó đọc 518,5/515,5 trước và 530,3/521,5 sau ⇒
  **~2 trong 4 điểm cải thiện là do floor trôi**. Con số trung thực của thay đổi này là **−1,5%**.
- ⛔ **KHÔNG tuyên bố M6-codegen ĐẠT**: cả 4 shape nay ≤15% so floor **HIỆN TẠI (V0)**, nhưng V0
  đã được chứng minh là đặc tả SAI theo hướng có lợi cho ta — so V3 fib ~1,22x, so V1 ~1,24x,
  **đều TRƯỢT**; cộng thêm 13,6% khe hở không phải codegen. Tuyên bố đạt trên V0 = chọn thước
  đo yếu nhất trong ba.
- Còn lại trong fib: chuỗi `mov rcx→rax→rbx` lúc vào và `mov rsi→rax` lúc ra.

## 🎯 VIỆC KẾ TIẾP ĐÃ CHỌN — **fib là shape DUY NHẤT còn chặn mốc (1,18x / gate 1,15x)**, và khoảng cách chỉ **~2,6%**
Đã so từng lệnh giữa fib của AXIOM và NASM floor (`bin/bench/fib_ax.exe` vs `fib_hand.exe`):

| | AXIOM | floor |
|---|---|---|
| **nhánh base** (≈NỬA trong ~331 triệu lời gọi) | **15 lệnh** — prologue ĐẦY ĐỦ rồi mới test | **4 lệnh**: `cmp; jl; mov rax,rcx; ret` |
| prologue | 5 (`push rbp; mov rsp,rbp; push rbx; push rsi; sub $0x20`) | 3, và **chỉ trên nhánh đệ quy** |
| nạp tham số | `mov %rcx,%rax; mov %rax,%rbx` + `mov %rax,%rcx` mỗi call | `mov rbx,rcx` |

⇒ Floor **test base case TRƯỚC prologue** = **shrink-wrapping**. AXIOM trả trọn 5 lệnh prologue
+ 5 lệnh epilogue cho **mọi** lời gọi base — tức khoảng **một nửa** số lời gọi.

### ✅ ĐÃ ĐỊNH GIÁ LẠI XONG (4 biến thể NASM, **có `align 16`**, xen kẽ 4 vòng, lặp 2 lần)
| biến thể | ms | vs V0 |
|---|---|---|
| V0 shrink-wrap, KHÔNG rbp — **floor hiện tại** | 513,6 | — |
| V1 shrink-wrap, **CÓ rbp** | **483,3** | **−5,9%** |
| V2 prologue-trước, KHÔNG rbp | 515,4 | +0,4% |
| V3 prologue-trước, CÓ rbp — **ĐÚNG hình dạng AXIOM** | 492,6 | −4,1% |

**1. Shrink-wrapping chỉ đáng ~0,4–0,7% ⇒ phán quyết 2026-07-29 ("+0,5 ms, vô giá trị") LÀ ĐÚNG.**
Nghi ngờ của tôi ở mục trên **SAI**; ghi lại để không ai đào lại lần thứ ba.

**2. CÓ frame pointer NHANH HƠN 4–6% trên shape này** (V1<V0, V3<V2; nhất quán 2 lần chạy, đã
khống chế alignment; **frame size hai bên BẰNG NHAU = 64 byte** nên không phải do khác cỡ frame).
AXIOM **đã có sẵn** rbp frame ⇒ ⛔ **"bỏ frame pointer" KHÔNG chỉ vô giá trị mà CÓ HẠI — ĐỪNG LÀM.**
(Không suy đoán cơ chế: đây là con số đo được, chưa giải thích.)

**3. ⭐⭐⭐ FLOOR CỦA fib ĐANG BỊ ĐẶC TẢ SAI, VÀ NÓ TÂNG BỐC AXIOM.** V0 vừa **không** phải hình
dạng AXIOM phát (AXIOM có rbp, không shrink-wrap = **V3**), vừa **không** phải bản nhanh nhất
(**V1**). Hệ quả về con số:
- so với floor hiện tại V0: **608/513 = 1,18x** (đang báo cáo)
- so với ĐÚNG hình dạng AXIOM V3: **608/493 = 1,23x**
- so với bản viết tay NHANH NHẤT cùng thuật toán V1: **608/483 = 1,26x**
⇒ **Khoảng cách thật của fib LỚN HƠN 1,18x**, và câu "chỉ còn ~2,6%" ở trên **là SAI, đã rút lại**.
⚠️ **CẦN USER QUYẾT** (D1 là quyết định của user nên tôi KHÔNG tự đổi thước đo): nên định nghĩa
floor là **V1** (bản viết tay nhanh nhất cùng thuật toán — hợp với mục đích "đo chất lượng
codegen") hay **V3** (cùng hình dạng)? Cả hai đều làm mốc khó hơn con số đang báo cáo.
⇒ **Và không ứng viên nào trong hai cái trên đóng được gate**: chỗ tốn kém của fib nằm ở nơi khác.

### 🔬⭐⭐⭐ W-SERIES: **13,6% khe hở của fib TỒN TẠI VỚI CÙNG HỆT TỪNG LỆNH** — không quy được cho codegen
Chép **NGUYÊN VĂN từng lệnh** của `ax_fib` (driver `522BEA6B`, -O3) sang NASM (`V4`), so với
`V3` (cùng frame shape nhưng thân gọn) và với chính binary AXIOM:

| | best ms |
|---|---|
| V3 (frame shape, thân gọn) | 495,9 |
| **V4 — chép nguyên văn `ax_fib`** | **538,3** |
| **AXIOM (binary thật)** | **611,3** |

**1. Bốn copy reg-reg thừa đáng 42,4 ms = +8,6%** (V4−V3). ⇒ **ĐÍNH CHÍNH một phần** ghi chú cũ
"copy vô hình ở vòng LATENCY-BOUND (fib)": chúng **KHÔNG** miễn phí. (Không mâu thuẫn với thí
nghiệm 07-29e — lần đó bỏ copy bằng **bias allocator**, tức đổi luôn cách gán thanh ghi; đây là
so hai chuỗi lệnh cố định.) Đây là phần **CÓ THỂ HÀNH ĐỘNG**: `mov rcx→rax→rbx` lúc vào,
`mov rax→rcx` trước mỗi call, `mov rsi→rax` lúc ra.

**2. ⛔ CÒN 73 ms = 13,6% KHÔNG GIẢI THÍCH ĐƯỢC, với chuỗi lệnh GIỐNG HỆT.** V4 và `ax_fib` có
cùng từng lệnh, cùng `align 16`, cùng frame 64 byte — mà chênh 13,6%.
- **Đã loại trừ startup**: chương trình AXIOM rỗng chạy **10,9 ms**, không phải 73 ms.
- Nghi can còn lại (chưa kiểm chứng được bằng đo hộp đen): bố cục cache/ITLB của binary lớn hơn,
  aliasing của branch predictor, khác biệt do self-link. **Bước tiếp theo cần PERF COUNTER phần
  cứng**, không phải đọc thêm disassembly.

⭐⭐ **HỆ QUẢ QUAN TRỌNG NHẤT**: một phần lớn "khe hở codegen" của fib **KHÔNG nằm ở lệnh mà
compiler chọn**. ⇒ **Đuổi theo tối ưu mức-lệnh cho fib có thể là đuổi theo cái không tồn tại.**
Kết hợp với mục floor bị đặc tả sai ở trên: **con số fib (1,18x hay 1,23x hay 1,26x) đang trộn
ít nhất BA thứ** — chất lượng codegen thật, định nghĩa floor, và một hiệu ứng môi trường 13,6%.
**Đừng lên kế hoạch lớn cho fib trước khi tách được ba thứ này.**

### (bối cảnh cũ, giữ lại) hai ứng viên này từng bị định giá và loại
⚠️ **CẢ HAI ứng viên này ĐÃ TỪNG BỊ ĐỊNH GIÁ VÀ LOẠI** (2026-07-29, bằng `perf_asm_variants.ps1`):
shrink-wrapping **+0,5 ms**, bỏ rbp frame **−17 ms** — đều bị coi là "vô giá trị/nhiễu".
**CẦN ĐỊNH GIÁ LẠI, vì hai lý do CỤ THỂ chứ không phải vì hoài nghi chung:**
1. Lần định giá đó chạy **TRƯỚC khi có căn lề hàm** ⇒ mỗi biến thể NASM rơi vào một vị trí
   ngẫu nhiên trong dải ±5% (chính là hiệu ứng đã chứng minh ở mục căn lề bên trên). Một tín
   hiệu 3% **không thể phân biệt được với nhiễu** dưới điều kiện đó.
2. Tiền lệ **immediate-folding**: cũng bị NASM chấm "nhiễu" (−0,5/−13 ms) ngày 07-24e, rồi hoá
   ra đáng **14%** trên callloop khi đo đúng shape. Đây là lần thứ HAI một "số 0 tự tin".
   Và **−17 ms trên 541 ms = 3,1%** — **ĐỦ ĐỂ ĐÓNG GATE 2,6%** nếu nó là thật.
**Cách làm đúng**: dựng biến thể NASM (có/không shrink-wrap; có/không rbp) và đo **GHÉP CẶP XEN
KẼ, nhiều cặp, so HIỆU** — đúng phương pháp đã dùng để cứu mục căn lề. **ĐỪNG** tin một lần chạy.

## (đã xong) callloop — disassembly cũ, giữ để tham chiếu
Thân vòng lặp hiện tại (10 lệnh, `-O3`, driver `E72FB62E`):
```
mov  $0x7,%rdx          <- hằng BẤT BIẾN trong vòng, nạp lại mỗi vòng
mov  %rax,%rbx          <- copy acc (destructive form)
mov  %rcx,%rsi          <- copy i
imul $0x2,%rsi,%rsi     <- nên là add/lea
add  %rsi,%rbx
imul $0x3,%rdx,%rdx     <- 7*3 là hằng HOÀN TOÀN, vẫn tính lại mỗi vòng
add  %rdx,%rbx
and  $0xfffff,%rbx
mov  %rbx,%rax          <- copy ra
lea  0x1(%rcx),%rcx
```
Floor tương ứng chỉ 5 lệnh: `lea rax,[rax+rcx*2] ; add rax,21 ; and rax,1048575 ; inc rcx ; cmp/jb`.

Ba mục còn lại, **mỗi mục đo RIÊNG (đừng batch)**:
1. **Constant folding `7*3` → `21`** sau inline. `work(acc,i,7)` được inline nên compiler ĐÃ có
   thông tin; hai lệnh `mov $7` + `imul $3` là thuần lãng phí. ⚠️ Đây là **AIR-level const
   propagation**, ranh giới M6-codegen / M6-opt cần cân nhắc — nhưng floor đã tính nó vào.
2. `IMUL vD, imm(2^k)` → `SHL vD,k`; `imm(2)` → `ADD vD,vD`. Rẻ, cùng thành ngữ peephole.
3. Hai copy quanh dạng destructive mà `coalesce_dest_copy` chưa với tới (nó cần 3 lệnh LIỀN KỀ).
4. (arrwalk, riêng) địa chỉ có scale `(%rbx,%rax,8)` + hoist địa chỉ bảng bất biến — `imul` nằm
   TRONG chuỗi phụ thuộc pointer-chasing nên đắt hơn số lệnh gợi ý.

## ⭐⭐⭐ BÀI HỌC PHƯƠNG PHÁP PHIÊN NÀY (quan trọng hơn cả bản vá)
**Bản ĐẦU của peephole 1d KHÔNG LÀM GÌ CẢ — và qua sạch mọi gate trong khi không làm gì.**
Viết với cửa sổ 2 lệnh liền kề (`MOV_IMM vC,k ; ALU vD,vC`), nó khớp **0/4** hằng của callloop.
Lý do: selection **materialise hằng TRƯỚC** khi phát copy `MOV vT,lhs` của dạng destructive, nên
MOV_IMM và bên tiêu thụ cách nhau HAI lệnh. Shape thật là:
`MOV_IMM vC,k ; MOV vT,vD ; ALU vT,vC`.

⛔ **Disassembly KHÔNG THỂ phát hiện điều này** — ở đó cặp lệnh trông LIỀN KỀ
(`mov $0x2,%rbx; imul %rbx,%rdx`), vì các tầng sau không giữ nguyên thứ tự này. **Disassembly là
bằng chứng về BINARY ĐÃ XONG, không phải về mảng mà một peephole PRE-ALLOCATION khớp trên.**
Cách bắt: in thẳng instruction stream mà pass nhìn thấy — 1 lần chạy là ra.
⇒ **Luật mới: một peephole phải được chứng minh là CÓ NỔ, không chỉ là AN TOÀN.**
Cùng họ thất bại với "module cursor" RFC 0035 (`A==B` xanh chứng minh SỐ KHÔNG).

⭐ **Và lần thứ HAI cùng một sai lầm định giá**: chính ứng viên này đã bị NASM định giá
2026-07-24e ra −0,5/−13 ms (= nhiễu) rồi **loại bỏ** — vì định giá trên **fib**, vốn
latency-bound, nơi lệnh nạp hằng nằm NGOÀI chuỗi phụ thuộc nên vô hình. Y hệt chuyện register
coalescing: đo 0 trên fib, rồi −30% trên xorshift. **Định giá một thay đổi codegen trên MỘT
shape sẽ liên tục cho ra số 0 tự tin cho những thứ đáng vài phần trăm ở chỗ khác.**

## ⭐ Oracle `t_aluimmfold` — ĐÃ HIỆU CHUẨN bằng 3 lần phá có chủ ý
- `and_ri` viết nhầm digit /1 (thành `or`) → **crash** (nó phá luôn `and rsp,-16` căn stack).
- Bỏ guard imm32 → **exit 8**, đúng như thiết kế (0xFFFFFFFF sign-extend thành −1 ⇒ giá trị
  không đổi).
- Bỏ REX.R của `imul_ri` → **PASS 42** lần đầu! ⇒ **lộ lỗ hổng THẬT**: không check nào chạm
  r8–r15, mà `imul_ri` là encoder DUY NHẤT trong 4 cái mới cần REX.R (dst nằm ở CẢ reg lẫn rm).
  Đã thêm khối 12 biến sống đồng thời để ép allocator dùng thanh ghi cao; nay phá là crash.
  **Bài học: encoder mới phải có test chạm r8–r15, nếu không nửa trường REX không có coverage.**

## 🔎 ĐÃ ĐIỀU TRA (read-only, cuối phiên) — **MACH_LEA KHÔNG có index/scale**, và đó là nút thắt chung
Đã đọc thẳng nguồn, không suy đoán:
- `x86_encode_lea(dst, base, disp)` (`x86_encoding.ax:322`) chỉ nhận **base + displacement**.
- `x86_encode_modrm_rm` (`x86_modrm.ax:102`) chỉ phát **SIB** cho đúng ca bắt buộc
  (`base & 7 == 4`, tức RSP/R12), và khi đó **hard-code index = 0x04 = "không có index",
  scale = 0** (`x86_modrm.ax:134`). **REX.X không bao giờ được set** ở đường này.
⇒ `lea (%rdx,%rdx,2)` hay `mov (%rbx,%rax,8),%rsi` **hiện KHÔNG diễn đạt được**.

**Đây là nút thắt CHUNG của 3 mục backlog còn lại**, nên làm nó TRƯỚC sẽ mở khoá cả ba:
1. `imul $2/$3` → LEA scale (callloop). LEA scale phủ k ∈ {2,3,4,5,8,9}: `x*3` =
   `lea (%r,%r,2)`. Latency 1 thay vì 3, mà `imul` đang nằm TRONG chuỗi tích luỹ.
   ⚠️ **ĐÍNH CHÍNH ghi chú trước đó trong chính file này**: tôi đã viết rằng `IMUL → SHL/ADD`
   "phải chứng minh không JCC/SETCC nào đọc cờ ở giữa". **Rào cản đó KHÔNG tồn tại.** `IMUL`
   vốn đã phá cờ (ISA để SF/ZF/AF/PF **undefined**), nên **không có consumer ĐÚNG ĐẮN nào có
   thể đọc cờ xuyên qua nó** ⇒ thay `IMUL` bằng `ADD`/`SHL` không thể phá code đúng, xét về cờ.
   Đã kiểm chứng thêm: compiler **không** phát nhánh đọc overflow sau phép toán — wrap theo bề
   rộng làm bằng mask/sign-extend (`emit_wrap_to_width` → `emit_load_extend`, `x86_selector.ax:887`),
   không đọc OF; `CC_O` có khai báo nhưng không có site nào phát `JO` sau IMUL.
   ⇒ **`imul $2` → `ADD vD,vD` là một thay đổi ~5 dòng, an toàn, KHÔNG cần chờ SIB.** Vẫn nên
   ưu tiên LEA cho bức tranh chung (nó phủ cả `$3`, và không ghi cờ nên không ràng buộc thứ tự),
   nhưng nếu muốn một win nhỏ, độc lập, đo được trước khi làm SIB thì đây là mục rẻ nhất.
2. Địa chỉ có scale cho index mảng (arrwalk) — mục ĐẮT NHẤT của arrwalk vì `imul` nằm trong
   chuỗi pointer-chasing.
3. `base + idx*k` tổng quát.

**Việc cần làm (ước lượng, chưa code):** thêm cách mang index-vreg + scale trong `MachInst`
(dùng `src2` cho index vreg — regalloc đã đếm `src2` là một operand nên liveness tự đúng; scale
nhét vào `padding` hoặc `imm`), một encoder SIB thật (set cả **REX.X** cho index r8–r15 — xem
bài học REX.R của `imul_ri` ở trên: nửa trường REX không có test thì không có coverage), và mở
rộng `format_operand` cho đường asm-text. Backend ⇒ **B==C bắt buộc**.

## Backlog còn lại (sau callloop)
- RFC 0035: method/global/ctor vẫn scheme cũ (`axS_`/`axG_`/`axC_`); P3 (E0501 → error) vẫn bị
  chặn bởi shim runtime trùng lặp hợp lệ.
- `mod_name` rỗng ở `register_module_from_lib` (binding `mod.NAME` là đồ chết) — vô hại.
- Module path nhiều đoạn không resolve ở call site (`bin.libcol.liba.helper()`) — có sẵn từ trước.
- M6-opt (accumulator/tail-rec) là milestone RIÊNG với M6-codegen.
- Nếu hết việc: `axiom-bug-probe`.

## ⛔ Cảnh báo còn hiệu lực
- **ĐỪNG cài loop rotation / bottom-test** — đã đo **+0,1%**.
- **ĐỪNG định giá copy/coalescing/hằng số trên fib** — fib latency-bound.
- Backend/ABI/linker ⇒ **B==C bắt buộc** trước commit (A!=B là bình thường khi codegen đổi).
- Một lần chạy `perf_suite` KHÔNG đáng tin (phương sai 8–10%): đo ghép cặp xen kẽ, ≥2 vòng, và
  so TỶ LỆ chứ không so ms giữa các phiên.
