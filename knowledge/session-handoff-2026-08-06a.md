---
name: session-handoff-2026-08-06a
description: Resume point after RFC 0038 (variadic print/println) and libc cleanup steps 0-2; P6 latent type-index defect is the open thread
metadata:
  type: project
---

# Handoff 2026-08-06a

## Trạng thái: GREEN, cây sạch, đã commit hết

| | |
|---|---|
| HEAD | `3f3ea74` refactor(libc): drop clock, exit and fflush |
| Driver `bin/axc_native.exe` | **A==B `0585124E9B3B81B4878E609A6279AD4DAF7B18852524A4D660F968057F6FFF08`**, 2.307.584 byte |
| Baseline regression | **679/679** ở **cả default lẫn `-O0`** |
| Mốc B==C gần nhất | `c3eae77` / `52D1ABD4…` (2.297.856 byte) — **thay đổi backend kế tiếp phải dựng lại B==C từ driver MỚI** |
| libc | `ucrtbase.dll` **16 → 13** ký hiệu |

## Đã ship phiên này
1. **`ca7a98d` — RFC 0038: `print`/`println` variadic.** Mọi đối số sau đối số đầu **bị nuốt im
   lặng** (`println("val: ", s)` chỉ in `val: `). Sửa bằng **desugar ở `air_builder`, KHÔNG ở
   selector** — selector không được tự chế N lời gọi từ 1 (§4), và làm ở AIR thì cả ba backend
   (x86_selector/cgen/wasm) được sửa miễn phí. Kèm: `println()` trần hết SEGFAULT, và
   `error[E3033]` cho kiểu không in được (char8 từng segfault, struct/ptr in dòng trắng exit 0).
2. **`50d92ce` — audit libc đầy đủ** → [audit-libc-dependencies](audit-libc-dependencies-2026-08-05.md).
3. **`3f3ea74` — dọn libc bước 0-2**: bỏ `clock`, `exit`, `fflush`.

## 🔴 VIỆC TIẾP THEO — P6 (đang có investigator chạy nền lúc viết file này)
**Xoá MỘT khai báo `extern` CHẾT ở `std/io.ax` làm HỎNG biên dịch một chương trình không liên quan.**
Repro (không cần dựng lại compiler — `concatenate_stdlib` đọc `std/*.ax` **từ đĩa lúc biên dịch**):
xoá bất kỳ một trong `fseek/ftell/rewind/fputs` ở `std/io.ax`, rồi
`./bin/axc_native.exe build bin/t_ifaceconsumer.ax -o x.exe -O1`
⇒ `operator '+' is not defined for Option/Result operands` trỏ vào `a.run(8) + b.run(20)`, **hai
toán hạng đều i64**, chương trình **không có Option nào**.
⭐ Vân tay: **tên kiểu báo lỗi đổi theo SỐ khai báo bị xoá** (1 ⇒ `Option`, 4 ⇒ `Vec`) ⇒ kiểu được
đọc qua **chỉ số trôi theo bảng symbol**. **Thêm** extern thì vô hại, **chỉ xoá** mới kích hoạt; xoá
dòng comment cũng vô hại. Không file bundled nào dùng bốn symbol đó.
Chi tiết đầy đủ: `knowledge/BACKLOG.md` mục **P6**.
⛔ Vì P6, bốn extern chết **cố ý giữ lại** trong `std/io.ax` (có comment cảnh báo tại chỗ). **Xoá
được bốn dòng đó = regression test của P6.**
Câu hỏi mở quan trọng nhất giao cho investigator: **user có tự chạm được P6 không** (chỉ bằng hình
dạng chương trình của họ, không sửa `std/`)? Nếu có ⇒ P6 khẩn cấp, không chỉ là vật cản dọn dẹp.

## Hàng đợi sau P6
- **libc bước 3** — viết lại `std/io.ax` sang handle native (CreateFileA/ReadFile/WriteFile/
  CloseHandle ‖ syscall 2/0/1/3). Bỏ thêm **4** ký hiệu libc. ⚠️ rủi ro trung-cao:
  `main_air.ax:163 read_file_content` đọc TOÀN BỘ input compiler qua đây.
- **libc bước 4-5** — chuyển `linker.ax`, `x86_coff.ax`, `x86_elf64.ax`, `cgen/wasm/asm_emitter`
  sang `std.io`; bỏ nốt stdio. Tất cả **A==B**.
- **libc bước 6-9** — `system`, `memcpy`/`memset`, `strlen`, và giết `ax_runtime.dll` trên Windows.
  Tất cả **B==C**. Bước 9 giá trị cao nhất: runtime libc-free **đã viết ~80%** ở
  `bootstrap/runtime/panic.ax`, chỉ đang bị gate ELF-only ở `linker.ax:3162-3175`.
- ⛔ **`atof` cấm đụng** nếu chưa có RFC — `std/string.ax:818` không có số mũ và **không làm tròn
  đúng**, thay vào sẽ đổi âm thầm mọi float literal compiler sinh ra.
- **P4** (RFC 0038): hàm `println` do user định nghĩa bị selector cướp vì khớp theo **chuỗi tên**
  chứ không theo **danh tính symbol** (`x86_selector.ax:1730`→`:1737`). Sửa đúng = đưa việc chọn
  symbol runtime ra khỏi selector ⇒ **B==C**.
- **Di trú oracle sang CLAUDE.md §7.1** (stdout thay vì exit code). Nay đã khả thi vì `println`
  variadic đã chạy. **Di trú dần**, không big-bang 679 hàng.

## Bài học đo lường phiên này (đắt, đừng học lại)
1. ⛔ **`strings | grep <tên hàm libc>` KHÔNG trả lời được "có phụ thuộc libc không".** Nó báo
   `printf`/`puts`/`malloc` "có" — sai, đó là **hằng chuỗi** trong whitelist `cgen.ax:140`.
   **Phải parse bảng import PE.** (Script parse có trong lịch sử phiên này.)
2. **Comment trong repo này trôi lệch và đã làm lạc hướng cả một cuộc điều tra.** Comment
   `std/io.ax:4-5` đổ lỗi cho "resolver extern-C xuyên module" — sai hoàn toàn; thủ phạm là **danh
   sách bundle CỨNG** `main_air.ax:401-426`. Cái đúng ở `main_air.ax:1875` đã có từ **hai tuần
   TRƯỚC** khi comment kia được viết. ⇒ Luôn `git log` cái file mình sắp tin.
3. **Một test fail không bao giờ là "flake".** `t_ifaceconsumer` fail ⇒ bisect ra P6 trong ~6 lệnh.
   Nếu đã gán nhãn flake thì mất luôn một defect thật.
4. **`&&` với `grep -c` là bẫy**: `grep -c` **exit 1 khi đếm được 0** ⇒ vế sau `&&` không chạy và
   trông như lỗi. Đã suýt đọc nhầm một lần.
