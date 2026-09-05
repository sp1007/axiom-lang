---
name: bug-println-arity-and-type-defects
description: The print/println defect family P1-P6 — P1/P2/P3 closed by RFC 0038, P4 (user-defined println hijacked by the selector) still OPEN and needs B==C, P5 was a console codepage issue not a bug
metadata:
  type: project
---

# `print` / `println` — họ defect P1–P6 (user báo 2026-08-05)

> Tách khỏi `BACKLOG.md` ngày 2026-09-05: `BACKLOG.md` là **file con trỏ**, không phải kho sự thật
> (CLAUDE.md §24). Nội dung dưới đây là **nguyên văn**, không lược bớt — mỗi mục dạy một bài học
> chẩn đoán riêng. Trạng thái tóm tắt: **P1/P2/P3 ĐÃ SỬA (RFC 0038)**, **P4 CÒN MỞ (chạm backend ⇒
> B==C)**, **P5 không phải bug**, **P6 ĐÃ SỬA 2026-08-06**.

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

Liên quan: [[bug-cross-tree-decl-node-segv]] · [[lesson-exit-code-8bit-masking]]
