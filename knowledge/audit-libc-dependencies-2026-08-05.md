---
name: audit-libc-dependencies
description: Full libc-dependency audit of std/ + the compiler + ax_runtime.dll, with the measured PE import tables and a sequenced removal plan
metadata:
  type: project
---

# Audit: libc dependencies (2026-08-05)

Chỉ thị user: *"rà soát các thư viện thuộc std xem có cái nào phụ thuộc lib c không, có thì xử lý đi,
axiom cần độc lập với libc"*.

## ⭐ SỰ THẬT ĐO ĐƯỢC — bảng import PE, không phải suy đoán

⛔ **`strings | grep <tên hàm>` KHÔNG dùng được để trả lời câu hỏi này.** Tôi đã thử và nó báo
`printf`/`puts`/`malloc` "có mặt" — sai, chúng là **hằng chuỗi** trong whitelist ở `cgen.ax:140`.
Phải **parse bảng import PE thật**. Script dựng sẵn: xem lịch sử phiên này (parse `IMAGE_DIRECTORY_
ENTRY_IMPORT`, RVA→offset qua bảng section).

`bin/axc_native.exe` (2.308.096 byte, mốc `2234451`) import ĐÚNG:
- `kernel32.dll` → GetStdHandle, WriteFile, VirtualAlloc, ExitProcess, GetLastError, VirtualFree,
  DeleteFileA, GetCommandLineA
- `ax_runtime.dll` → ax_get_global_state_internal, ax_actor_step, ax_actor_is_running,
  ax_actor_has_messages, ax_str_len, ax_print_str, ax_str_eq, ax_i64_to_str, ax_panic, ax_str_slice
- **`ucrtbase.dll` (= libc) → 16 ký hiệu**: memset, memcpy, strlen, fopen, fclose, fread, fwrite,
  fputs, atof, exit, fflush, fseek, ftell, rewind, clock, system

Chương trình tầm thường (`bin/probe9/pn1.exe`) chỉ còn **3**: memset, memcpy, strlen.
Khớp `docs/linux-target.md:18` (ELF `DT_NEEDED libc.so.6` chỉ vì memset/memcpy).

## ⭐⭐ HAI PHÁT HIỆN LẬT NGƯỢC GIẢ THIẾT

### 1. "Cross-module extern-C resolver issue" ở `std/io.ax:4-5` là **COMMENT LỖI THỜI, KHÔNG PHẢI BUG**
Comment (từ `ce7c4e2`, 13/06) đổ cho resolver. Sai. Resolver + mangler **đã đúng**:
- `main_air.ax:1875` export extern **bất kể `pub`** — có từ `a61b19e` (30/05), tức **hai tuần TRƯỚC**
  khi comment kia được viết.
- `x86_regs.ax:254-256` `if is_extern: return fn_name` — lời gọi extern xuyên module phát tên C trần.
- `linker.ax:710` đã whitelist sẵn `CreateFileA`/`ReadFile`/`CloseHandle`.
- Tiền lệ CHẠY ĐƯỢC: `std/net.ax` gọi `std.os.linux_sys.syscall(...)` (`pub extern "C"` ở
  `linux_sys.ax:13`) từ 10 vị trí.

Cái thật sự hỏng nằm ở **đường BUNDLING trong driver**, không phải resolver:
`main_air.ax:1061 concatenate_stdlib` nối **danh sách 8 file CỨNG** (`main_air.ax:401-426`) và
**`std/os/win32.ax` KHÔNG có trong đó**; `main_air.ax:225-254 strip_imports` xoá mọi dòng
`import std`; `:264-283 strip_package_prefixes` viết lại tiền tố (mục `std.os.win32.` **đã có sẵn ở
`:277`** — dấu hiệu đây đúng là ý đồ thiết kế ban đầu). ⇒ `CreateFileA` mất khai báo.
Bằng chứng phụ: **`std/os/win32.ax` là FILE CHẾT** — `CreateFileA:20` không được gọi từ đâu cả; mọi
module std khác cần Win32 đều **khai báo lại extern tại chỗ** (`std/runtime.ax:8-10`,
`std/mem/alloc.ax:94-95`, `std/os.ax:124-129`, …), và **trùng khai báo extern giữa các file bundled
đang hoạt động bình thường** (`GetStdHandle` khai ở cả `runtime.ax:8` lẫn `mem/alloc.ax:94`).
⇒ **`std/io.ax` viết lại được NGAY, không cần đụng compiler.** Đây là kết quả rẻ nhất có thể.

### 2. Không thể "độc lập libc" bằng cách sửa `std/` — vì **`ax_runtime.dll` là artifact C link UCRT**
`runtime/ax_print.c` CHÍNH LÀ implementation của `ax_println_str` & co (`:82-128` nhánh Win32,
`:212-243` POSIX). `bin/ax_runtime.dll` (155.944 byte, 31/05) import `api-ms-win-crt-*` :
strtod, strtoll, calloc/free/malloc/realloc, memcmp/memcpy/memset, strchr/strlen/strncmp, isspace,
fflush, abort, _exit, `__stdio_common_vfprintf`, `_initterm`, … `README.md:65-67` xác nhận đây là
phụ thuộc triển khai bắt buộc trên Windows.
⭐ **NHƯNG runtime libc-free cho Windows đã viết ~80% và đang BỊ TẮT:** `bootstrap/runtime/panic.ax`
hiện thực `ax_print_str:120`, `ax_println_str:123`, `ax_print_i64:127`, `ax_println_i64:134`,
`ax_print_bool:138`, `ax_println_bool:144`, `ax_print_f64:151/169`, `ax_str_eq:175`, `ax_panic:39`,
… **có đủ nhánh Windows** (`:41-51` GetStdHandle/WriteFile/ExitProcess). Nó chỉ được bundle khi
`target_linux` (`main_air.ax:1039-1052`) và chỉ được bind khi `format == "elf64"`
(`linker.ax:3162-3175`, comment nói rõ "trên COFF các tên này là import ax_runtime.dll thật").
Twin còn thiếu: `ax_str_len`, `ax_i64_to_str`, `ax_str_slice`, `ax_actor_{step,is_running,has_messages}`
— ba cái đầu đã có thân ở `std/runtime.ax:173,186`.

## Cơ chế phân loại DLL (gốc của mọi thứ)
`linker.ax:696-731 get_dll_for_symbol` = whitelist cứng + **fall-through libc ở `:731`
`return "ucrtbase.dll"`**. Tức là **quên whitelist một tên = nó âm thầm thành import libc**.
Đã tìm ra hai nhóm **bị phân loại nhầm** (bug tiềm ẩn, chưa ai chạm):
- `std/process.ax:9,11` `CreateProcessA`, `GetExitCodeProcess` → phải là kernel32.
- `std/net.ax:10-19` `WSAStartup, socket, bind, listen, accept, connect, send, recv, closesocket,
  ioctlsocket` → phải là **ws2_32.dll** (linker chưa có bucket này).
⇒ chương trình dùng `std.net`/`std.process` trên Windows sẽ import từ DLL **không hề export** chúng.

## Builtin ≠ libc-free (điểm dễ sai nhất)
`resolver.ax:480-483` đăng ký `alloc/free/memcpy/memset` là builtin, nhưng tên phát ra khác nhau:
| builtin | tên phát | giải quyết bởi |
|---|---|---|
| `@alloc` | `ax_alloc` (`x86_regs.ax:242-246`) | **trong chương trình** (`std/runtime.ax:123`) — không libc |
| `@free` | `ax_free` (`:236-240`) | **trong chương trình** (`std/runtime.ax:138`) — không libc |
| `@memcpy` | `memcpy` **trần** (`x86_regs.ax:233`) | **LIBC THẬT** |
| `@memset` | `memset` **trần** (`:233`) | **LIBC THẬT** |
Loại thứ ba, **backend tự chế, không có khai báo nguồn nào**: `strlen` — `x86_selector.ax:1202` phát
`MACH_CALL` callee ma thuật `-21`, `x86_coff.ax:361-362` ánh xạ `-21` → `"strlen"` cho MỌI cast
`ptr[u8] → str`. (`abort` = `-8` ở `x86_coff.ax:357-358`, hiện chưa emitter nào phát — latent.)

## Kế hoạch tuần tự (mỗi bước gate riêng, nhiều commit nhỏ)
⚠️ **Ràng buộc self-host**: 8 file này được nối vào MỌI build native **kể cả compiler**
(`main_air.ax:401-426`): `result.ax, mem/alloc.ax, scheduler.ax, runtime.ax, os.ax, string.ax,
io.ax, collections.ax` ⇒ đụng vào là **đổi binary compiler**, phải lập lại A==B.
`net/time/process/thread/timer/reactor` **không** ảnh hưởng image compiler.

| # | Việc | Gate | Rủi ro |
|---|---|---|---|
| 0 | Sửa comment lỗi thời `std/io.ax:4-5`; xoá extern chết `fseek/ftell/rewind/fputs` (`:14-16,18`), `puts`/`remove` (`main_air.ax:40,43`), `GetStdHandle`/`WriteFile` chết (`print_helpers.ax:6-7`) | A==B | ~0 |
| 1 | Xoá mọi `fflush(null)` + 5 extern `fflush` — console đã đi qua `WriteFile`, chúng là no-op | A==B | thấp |
| 2 | `exit` → `ExitProcess`/`syscall(60)`; `clock` → `QueryPerformanceCounter` | A==B | thấp |
| 3 | **Viết lại `std/io.ax`** sang handle native (CreateFileA/ReadFile/WriteFile/CloseHandle ‖ syscall 2/0/1/3), tách nền bằng `@compiler_intrinsic("is_windows")` (idiom sẵn có `std/runtime.ax:17,57,79,111`) | A==B | **trung-cao** — `main_air.ax:163 read_file_content` đọc TOÀN BỘ input compiler qua đây |
| 3b | *(biến thể kiến trúc của 3)* thêm `std/os/win32.ax` vào `concatenate_stdlib`, cho extern `pub`, `std/io.ax` gọi `std.os.win32.*` | A==B | trung |
| 4 | ✅ **XONG 2026-08-07** — `linker.ax` (7 chỗ, không phải 4: còn `:1824`, `:1956`, `:3247` và **`main_air.ax:701,725`**), `x86_coff.ax`, `x86_elf64.ax` sang `std.io`. Thêm `read_all_bytes` (out-param độ dài, KHÔNG dùng `read_all` vì nó cắt ở byte NUL) + `write_bytes` + token `stream_*`. 13 → **8** | A==B | trung |
| 5 | ✅ **XONG 2026-08-07** — `cgen.ax`, `wasm.ax`, `x86_asm_emitter.ax`, `print_helpers.ax`, `main_air.ax:80`. 8 → **5** (`atof memcpy memset strlen system`) | A==B | thấp |
| 6 | Sửa whitelist: thêm `CreateProcessA`/`GetExitCodeProcess` vào kernel32, thêm bucket `ws2_32.dll`; rồi `system` → `CreateProcessA` | **B==C** | trung |
| 7 | **`memcpy`/`memset`** → `rep movsb`/`rep stosb` inline hoặc twin AXIOM (`x86_regs.ax:233`). **Đây là ký hiệu libc CUỐI CÙNG trên Linux** | **B==C** | **cao** — mọi phép copy aggregate của compiler |
| 8 | **`strlen`** → thay callee ma thuật `-21` bằng vòng quét inline/twin | **B==C** | cao |
| 9 | **Giết `ax_runtime.dll` trên Windows**: mở gate `linker.ax:3167` cho COFF, bundle `panic.ax` trên Windows (`main_air.ax:1039`), viết 6 twin còn thiếu | **B==C** | **cao nhất, giá trị cao nhất** |
| 10 | `atof` → strtod thuần AXIOM **làm tròn đúng** | A==B + **RFC** | cao (trôi số âm thầm) |

## ⛔ `atof` — KHÔNG được thay bằng cái đang có
`std/string.ax:818 str_parse_f64_impl` thuần AXIOM nhưng **không hỗ trợ số mũ** (`1e10` bị từ chối)
và cộng dồn từng chữ số bằng `*10.0`/`*0.1` ⇒ **không làm tròn đúng** (khác `strtod`). Thay vào sẽ
**âm thầm đổi giá trị MỌI float literal mà compiler biên dịch, kể cả hằng trong chính image của nó**.
⇒ giữ `atof` cho tới khi có parser decimal→binary làm tròn đúng. Cỡ RFC.

## Phán quyết trung thực
- **11/16** ký hiệu libc rời được `axc_native.exe` bằng thay đổi frontend/std, gate **A==B**:
  `fopen, fclose, fread, fwrite, fputs, fseek, ftell, rewind, fflush, exit, clock`.
- **4** cái nữa cần backend, gate **B==C**: `memcpy, memset, strlen, system`. Sau đó **binary Linux
  không còn `DT_NEEDED` nào**, Windows không còn import `ucrtbase`.
- **`ax_runtime.dll`** là phần dư mà sửa `std/` không bao giờ chạm tới được — bước 9.
- **`atof`** cần RFC.

## Sạch sẵn — không có extern nào
`std/math.ax` tự hiện thực `sqrt:75, pow:147, sin:193, cos:213, fmod:48` thuần AXIOM ⇒ **không phụ
thuộc libm**. Cùng nhóm sạch: `collections, json, sort, iter, fmt, cli, log, crypto, sync`.
`bootstrap/runtime/` (`syscall.ax:5-11`, `panic.ax:6-11`, `axalloc.ax:88-90`) chỉ dùng `syscall` +
kernel32 ⇒ **zero libc**.

Liên quan: [[audit-method-trap]] · BACKLOG mục "GỠ PHỤ THUỘC libc".
