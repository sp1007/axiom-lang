---
name: session-handoff-2026-08-07c
description: Resume point — stdlib reachability stage 1 gated but UNCOMMITTED; B3/B6 compiler crashes are the next task and are now user-reachable
metadata:
  type: project
---

# Handoff 2026-08-07c — ĐỌC MỤC "⚠️ VIỆC DỞ DANG" TRƯỚC

## ✅ TRẠNG THÁI: SẠCH, ĐÃ COMMIT HẾT, GREEN — không có việc dở dang
Stage 1 stdlib-reachability **đã gate, đã promote, đã commit** (`7faee73`).
`git status` sạch (chỉ còn `.claude/settings.json` untracked — file của user, KHÔNG đụng vào).

⚠️ **Bài học lặp lại lần thứ 3 trong ngày — ghi lại vì nó suýt lọt:** agent báo **709/709**,
lần **tự chạy** của tôi ra **708/1** (`diagloc-module` FAIL). Agent đã sửa comment trong
`bin/t_diagloc_modlib.ax` ⇒ **đẩy dòng lỗi từ 8 sang 12**, nhưng khối assert vẫn ghi 8, và agent
khẳng định khối đó "vẫn pass" mà **không hề kiểm**. Chẩn đoán bản thân nó ĐÚNG; chỉ số dòng trong
assert là cũ. Đã sửa assert về **dòng 14** và sửa câu "line 8" tự mâu thuẫn trong fixture.
⇒ **LUÔN tự chạy lại suite. Đừng chuyển tiếp con số của agent.**

## Trạng thái đã commit (HEAD `7faee73`)
| | |
|---|---|
| Driver `bin/axc_native.exe` | A==B `9C6726C11F366ACA5BA3970F72D0C0502C7495506F82AEBBCEF33A6C14C326E2`, **2.329.600 byte** |
| Baseline | **709/709** (cả default lẫn `-O0`, **tự chạy lại**) |
| Mốc **B==C** gần nhất | `c3eae77` / `52D1ABD4…` — **backend kế tiếp phải dựng lại B==C từ driver MỚI** |
| libc — compiler | `ucrtbase` **5**: atof memcpy memset strlen system |
| libc — chương trình user | **3**: memcpy memset strlen (sàn) |

## Đã ship hôm nay (mỗi cái gate A==B, tự kiểm chứng lại chứ không tin báo cáo agent)
`ca7a98d` RFC 0038 print/println variadic · `50d92ce` audit libc · `3f3ea74` libc 0-2 (16→13) ·
`74eab1d` **P6** (payload đọc như symbol chỉ khi cờ 2048; **user chạm được**, 7,7% tên method thường
gặp fail) · `1a12f40` RFC 0039 struct literal suy diễn · `0515e30` chẩn đoán 0-1 (cascade 52→3) ·
`9f9b265` `strip_imports` bôi trắng (giữ CẢ offset LẪN dòng) · `5916c1d` chẩn đoán 2b (mọi lỗi có
`--> file:line`, từ **MỘT** renderer thay vì ba) · `a522762` libc 3 (`std/io.ax` native) ·
`78463c3` srcmap module import · `3f54ed5` libc 4+5 (**13→5**) · `f188d7c` + `4c92d57` phát hiện
std/ + B1-B6.

## 🔴 VIỆC TIẾP THEO — B3 và B6, và chúng ĐÃ TRỞ NÊN NGHIÊM TRỌNG HƠN
Stage 1 mở đường loader cho `std.X` ⇒ **các module này nay TỚI ĐƯỢC loader và LÀM SEGV COMPILER**:
`import std.{sync,thread,process,iter,cli}` ⇒ **SEGV 139, không chẩn đoán**.
Trước stage 1 chúng bị **nuốt im lặng** (đo: `import std.thread` = ACCEPT trên driver cũ).
⚠️ **`iter` và `cli` là MỚI trong danh sách này** — bản định giá xếp `iter` vào nhóm "không parse
nổi", thực tế nó **crash** vì import `std.collections` (= B3).
- **B3**: module user `import` một module **ĐÃ BUNDLED** (`std.string`/`std.collections`) ⇒ SEGV 139
  khi typecheck module được nạp. Deterministic 4/4, cả -O0 lẫn -O1. Không crash dưới `--no-stdlib`
  ⇒ do **trùng lặp bundled-vs-loaded**. **Chưa localize được dòng** (đã loại trừ trùng tên hàm và
  trùng tên struct).
- **B6**: `@compiler_intrinsic("is_windows")` trong module nạp lazy ⇒ SEGV 139. Repro tối giản 5 dòng.
⇒ **KHÔNG được "sửa" bằng cách nhét lại các module đó vào bảng blank** — làm vậy là khôi phục đúng
cái defect "nuốt import im lặng" vừa mới sửa.
0 file trong corpus bị ảnh hưởng (breakage audit: 0 collateral), nhưng user gõ `import std.thread`
nay nhận **crash không chẩn đoán** thay vì im lặng.

## Hàng đợi sau B3/B6
- **Stage 2** stdlib: đăng ký 8 tên bundled thành `SYM_MODULE` cờ *bundled*;
  `lazy_resolver_resolve_field` tra scope toàn cục thay vì nạp file ⇒ `std.collections.new_vec` chạy,
  `std.string.len` bind **cùng symbol** với `len` trần ⇒ delta codegen = 0.
- **Stage 3** stdlib: **XOÁ `strip_package_prefixes`** ⇒ hết **B1** (nó đang **làm hỏng HẰNG CHUỖI**:
  `println("literal: std.string.len")` in ra `literal: len` — đã tự kiểm chứng), và **mở khoá CỘT**
  trong chẩn đoán (§3b).
- **libc 6-9 (đều B==C)**: `system` (+whitelist `CreateProcessA`/`GetExitCodeProcess`),
  `memcpy`/`memset` (`x86_regs.ax:233`), `strlen` (`x86_selector.ax:1202` callee `-21`), và
  **giết `ax_runtime.dll`** (runtime libc-free đã viết ~80% ở `bootstrap/runtime/panic.ax`, đang bị
  gate ELF-only ở `linker.ax:3162-3175`).
- ⛔ **`atof` cấm đụng nếu chưa có RFC** — `std/string.ax:818` không có số mũ, **không làm tròn đúng**
  ⇒ sẽ đổi âm thầm mọi float literal compiler sinh ra.
- **P4** (RFC 0038): `println` do user định nghĩa bị selector cướp vì khớp **chuỗi tên** ⇒ B==C.
- **Di trú oracle sang §7.1** (stdout thay exit code) — làm **dần**, 709 hàng đa số vẫn `exit|`.

## ⚠️ LINUX CHƯA TỪNG CHẠY
Toàn bộ nhánh Linux của `std/io.ax` (syscall 2/0/1/3, `read_all_bytes`, `stream_*`) **chưa hề thực
thi** — host là Windows. Đã kiểm: `--target linux` dựng được ELF và ELF linker vá import Windows
chết về 0. Chưa kiểm: syscall trả đúng như giả định, cờ open `577`/`420`, `fd < 0`, đường fd-0.

## 📏 SỨC KHOẺ std/ (đo, không đoán)
`math` ✅ `sort` ✅ dùng được end-to-end sau stage 1. `sync/thread/process/iter/cli` ⇒ **crash** (B3/B6).
`net, json, fmt, time, crypto, log` ⇒ **không parse nổi** (7-16 lỗi mỗi file) = **aspirational**.

## Bài học đo lường (đắt — đừng học lại)
1. ⛔ **`strings | grep <tên libc>` KHÔNG trả lời được "có phụ thuộc libc không"** — báo
   `printf`/`puts`/`malloc` "có", nhưng đó là **hằng chuỗi** trong whitelist `cgen.ax:140`.
   **Phải parse bảng import PE.**
2. ⭐ **Compiler bị CẮT CỤT (25 KB) xuất hiện KHÔNG cần tải đồng thời** — log thành công đầy đủ,
   exit 0. ⇒ **`ls -l` sau MỌI lần build** (~2,32 MB). Compiler cụt đọc y hệt "regression thảm hoạ".
3. **Một test fail không bao giờ là "flake"** — `t_ifaceconsumer` fail ⇒ bisect ra P6 trong ~6 lệnh.
4. **Repro của CHÍNH MÌNH cũng lỗi thời**: `bin/probe11/s1.ax` chết sau 2 giờ vì RFC 0039 hợp thức
   hoá cú pháp đó. **Kiểm lại repro trước khi dựa vào nó.**
5. **Vứt lượt chạy BỊ GIẾT, đừng union với lượt sạch.**
6. `grep -c` **exit 1 khi đếm 0** ⇒ `cmd | grep -c x && next` nuốt mất `next`.
7. **Chạy suite ở background mà pipe qua `tail`** ⇒ file output rỗng tới tận khi xong; muốn theo dõi
   thì bỏ `tail`.
8. **Comment trong repo trôi lệch và đã làm lạc hướng cả một cuộc điều tra** (`std/io.ax:4-5`).
   Luôn `git log` file mình sắp tin.
