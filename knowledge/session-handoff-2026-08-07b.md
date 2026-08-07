---
name: session-handoff-2026-08-07b
description: Resume point after diagnostics stages 0-2b and libc step 3; step 4 is next and its binary-read trap is already analysed here
metadata:
  type: project
---

# Handoff 2026-08-07b

## Trạng thái — cây SẠCH, mọi thứ đã commit, GREEN

| | |
|---|---|
| HEAD | `cf0c37a` (chore: gitignore) — công việc thực chất ở `a522762` |
| Driver `bin/axc_native.exe` | A==B `1E371716677E987A89E3630457A7C557A93E95807CE7CB90C5A230CFBE3FCC92`, **2.321.920 byte** |
| Baseline | **699/699** (cả default lẫn `-O0`) |
| Mốc B==C gần nhất | `c3eae77` / `52D1ABD4…` — **backend kế tiếp phải dựng lại B==C từ driver MỚI** |
| libc — compiler | `ucrtbase.dll` **13**: atof fclose fopen fputs fread fseek ftell fwrite memcpy memset rewind strlen system |
| libc — **chương trình user** | **3**: memcpy memset strlen ⭐ (sàn) |

## Đã ship phiên này (tất cả gate A==B, tự kiểm chứng lại chứ không tin báo cáo agent)
| Commit | Việc |
|---|---|
| `ca7a98d` | **RFC 0038** — `print`/`println` variadic (mọi đối số sau đối số đầu bị nuốt im lặng). Desugar ở **air_builder**, không ở selector ⇒ cả 3 backend sửa miễn phí. Kèm `println()` trần hết segfault + `error[E3033]`. |
| `50d92ce` | Audit libc → [audit-libc-dependencies](audit-libc-dependencies-2026-08-05.md) |
| `3f3ea74` | libc bước 0-2: bỏ `clock`, `exit`, `fflush` (16 → 13) |
| `74eab1d` | **P6** — `payload` đọc như symbol chỉ khi cờ 2048 nói vậy. **User chạm được**: 7,7% tên method thường gặp (`count`, `alloc`, …) không biên dịch được. |
| `1a12f40` | **RFC 0039** — `let c: T = (a: 1, b: 2)` suy diễn từ annotation, `error[E3034]` khi không có ngữ cảnh |
| `0515e30` | Chẩn đoán stage 0-1: cascade 52 dòng → 3; gỡ dump token, `nud`, `total_len` |
| `9f9b265` | Chẩn đoán stage 2a: `strip_imports` **bôi trắng** thay vì xóa ⇒ **giữ CẢ offset LẪN số dòng** |
| `5916c1d` | Chẩn đoán stage 2b: **mọi** chẩn đoán có `--> file.ax:LINE`, từ **MỘT** renderer (trước có BA, đã trôi lệch) |
| `a522762` | **libc bước 3**: `std/io.ax` sang handle native. Chương trình user: ucrtbase 8 → **3** |

## 🔜 VIỆC TIẾP THEO — libc bước 4+5 (đã phân tích sẵn, agent bị cắt trước khi chạy)
**Mục tiêu đo được:** `fopen, fclose, fread, fwrite, fputs, fseek, ftell, rewind` **biến mất** khỏi
import của `bin/axc_*.exe` ⇒ còn `atof, memcpy, memset, strlen, system` (5).

Các file còn khai báo extern stdio **riêng** (đó là lý do compiler vẫn 13 dù bước 3 đã xong):
- `linker.ax:6-11` — dùng ở `:1499-1509`, `:1608-1614`, `:1627-1640`, `:1655-1662`, …
- `x86_coff.ax:71-73` (dùng `:270-278`), `x86_elf64.ax:6-8` (dùng ~`:353-361`)
- `cgen.ax:6-8`, `wasm.ax:6-8`, `x86_asm_emitter.ax:560-561`, `print_helpers.ax:8` (bước 5, text)

### ⛔ BẪY LỚN NHẤT — đã phân tích, đừng phát hiện lại bằng cách hỏng link
**`std.io.read_all` dựa trên chuỗi NUL-terminated** (`@memset` + `buf as str`) ⇒ **dừng ở byte 0 đầu
tiên**. **`linker.ax` đọc file OBJECT NHỊ PHÂN, đầy byte NUL.** Dùng `read_all` ở đó sẽ **cắt cụt mọi
`.obj`** và cho ra bản link hỏng — nhiều khả năng hỏng **âm thầm**. Đúng lớp accept-then-miscompile.
✅ **Tin tốt đã xác minh:** bộ `fseek(f,0,2)` + `ftell` + `rewind` trong `linker.ax` **chỉ dùng để lấy
KÍCH THƯỚC file** rồi đọc trọn ⇒ **KHÔNG có truy cập ngẫu nhiên** ⇒ **không cần thêm seek/tell vào
`std.io`**. Hai lựa chọn:
- **(a) khuyến nghị** — thêm `read_all_bytes` trả buffer byte có độ dài, lặp `read` vào buffer lớn dần
  cho tới khi trả 0. Không cần lời gọi OS mới, không cần whitelist thêm.
- (b) thêm `size()` (GetFileSizeEx/fstat) rồi giữ hình dạng đọc-một-lần.
`read_all` (bản `str`) **phải giữ nguyên hành vi** — nơi khác đang phụ thuộc.

### Bẫy khác
1. **`std.io.write_all` TỰ ĐÓNG file.** Code cũ `fopen; fwrite; fclose` tường minh ⇒ coi chừng đóng 2 lần.
2. **`print_helpers.ax` nay chứa renderer chẩn đoán** (`print_diag_location`). `fputs` của nó dùng cho
   `print_to_file`/`ax_fprintf_local`. **Kiểm chiều phụ thuộc trước** — `std/io.ax` bundle từ đĩa,
   `print_helpers.ax` là stage1. Nếu vướng, **làm bước 4 thôi**, báo bước 5 chưa xong kèm lý do.
3. `x86_coff.ax`/`x86_elf64.ax` ghi **object nhị phân** ⇒ kiểm `write_all` có phải dạng `str` không.
4. ⚠️ Ký hiệu **không** có trong whitelist kernel32 của `linker.ax:708-718` sẽ **âm thầm rơi về
   `ucrtbase.dll`** (default `:731`) ⇒ import một tên ucrtbase không export ⇒ chương trình không chạy.
5. **Oracle mạnh nhất chính là fixpoint**: compiler đọc object của chính nó qua `linker.ax` và ghi qua
   `x86_coff.ax` ⇒ **A==B byte-identical là bằng chứng trực tiếp** đường nhị phân đúng. Thêm một hàng
   round-trip nhị phân **có byte NUL nhúng**.

## Hàng đợi sau đó
- **libc 6-8 (B==C)**: `system` (cần whitelist `CreateProcessA`/`GetExitCodeProcess`), `memcpy`/`memset`
  (`x86_regs.ax:233` phát tên trần), `strlen` (`x86_selector.ax:1202` callee ma thuật `-21`).
- **libc 9 (B==C, giá trị cao nhất)**: giết `ax_runtime.dll` trên Windows — runtime libc-free **đã viết
  ~80%** ở `bootstrap/runtime/panic.ax`, chỉ bị gate ELF-only ở `linker.ax:3162-3175`.
- ⛔ **`atof` cấm đụng nếu chưa có RFC** — `std/string.ax:818` không có số mũ, **không làm tròn đúng**.
- **Chẩn đoán còn lại** — xem BACKLOG §3b: không có CỘT (cố ý), node mono hóa không định vị được (cần
  RFC), cây module import `srcmap == null` (**rẻ**), parser trỏ vào dòng SAU.
- **P4** (RFC 0038): `println` do user định nghĩa bị selector cướp vì khớp theo **chuỗi tên** ⇒ B==C.
- **Di trú oracle sang §7.1** (stdout thay exit code) — làm **dần**.

## ⚠️ LINUX CHƯA TỪNG CHẠY
Bước 3 viết nhánh Linux (syscall 2/0/1/3) nhưng **host là Windows, chưa hề thực thi**. Chưa kiểm:
syscall trả đúng như giả định, cờ open `577`/`420`, `fd < 0` là phép thử đúng, đường fd-0.
Đã kiểm: `--target linux` dựng được ELF 37 KB và ELF linker vá import Windows chết về 0 (mẫu có sẵn).

## Bài học đo lường (đắt — đừng học lại)
1. ⛔ **`strings | grep <tên libc>` KHÔNG trả lời được "có phụ thuộc libc không"** — báo
   `printf`/`puts`/`malloc` "có", nhưng đó là **hằng chuỗi** trong whitelist `cgen.ax:140`.
   **Phải parse bảng import PE.**
2. ⭐ **Compiler bị CẮT CỤT (25 KB) xuất hiện KHÔNG cần tải đồng thời** — log thành công đầy đủ, exit 0.
   ⇒ **`ls -l` sau MỌI lần build** (~2,32 MB). Compiler cụt đọc y hệt "regression thảm hoạ".
3. **Một test fail không bao giờ là "flake"** — `t_ifaceconsumer` fail ⇒ bisect ra P6 trong ~6 lệnh.
4. **Comment trong repo trôi lệch và đã làm lạc hướng cả một cuộc điều tra.** Thêm ca mới: **repro
   `bin/probe11/s1.ax` của CHÍNH TÔI đã lỗi thời sau 2 giờ** vì RFC 0039 hợp thức hoá cú pháp đó —
   investigator bắt được. ⇒ **kiểm lại repro trước khi dựa vào nó**, kể cả repro mình vừa viết.
5. **Vứt lượt chạy BỊ GIẾT, đừng union với lượt sạch.**
6. `grep -c` **exit 1 khi đếm 0** ⇒ `cmd | grep -c x && next` nuốt mất `next`.
7. **Session limit cắt agent giữa chừng** đã xảy ra 3 lần hôm nay. ⇒ commit mọi thứ GREEN ngay khi
   xong, và giữ handoff này luôn cập nhật — lần bị cắt gần nhất **không mất gì**.
