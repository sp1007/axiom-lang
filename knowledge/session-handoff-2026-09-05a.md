---
name: session-handoff-2026-09-05a
description: Resume point — B3/B6 fixed, gated and pushed (360b3a2); B3b/stage-2 scoping investigation dispatched
metadata:
  type: project
---

# Handoff 2026-09-05a

## ✅ TRẠNG THÁI: SẠCH, ĐÃ COMMIT + PUSH, GREEN
HEAD = **`360b3a2`** `fix(typecheck): a lazily loaded module may call the bundled stdlib again`.
Đã push lên `main`. Không có việc dở dang trong cây.

| | |
|---|---|
| Driver `bin/axc_native.exe` | A==B **`F7146F3DBCDD3280E838B06538E946B33D2F3D19F32A18909BBAD81E22BC77E3`**, **2.329.600 byte** (đã `ls -l`) |
| Baseline | **711/711** ở **CẢ** default **LẪN `-O0`** — **tự chạy lại, không lấy số của agent** |
| Mốc **B==C** gần nhất | vẫn `c3eae77` / `52D1ABD4…` ⇒ **thay đổi backend kế tiếp phải dựng lại B==C từ driver MỚI này** |

## Đã ship phiên này
**B3 + B6 = MỘT bug** — `pre_infer_func_signature` (`typecheck.ax:3770`) đánh chỉ số `decl_node` của
unit KHÁC vào `self.tree`; ba call site (`:1626`, `:5898`, `:7081`) trong khi **bản sao thứ tư `:1985`
đã được canh sẵn**. Chi tiết đầy đủ: **`knowledge/bug-cross-tree-decl-node-segv.md`**.
`import std.{sync,thread,collections,string}` + module user dùng `@compiler_intrinsic`: **139 → 0**.

## 🔴 VIỆC TIẾP THEO — B3b (stage 2 của lộ trình stdlib)
`import std.{iter,process,cli}` **vẫn 139**. Nguồn module nạp trễ **không được tiền xử lý** ⇒ nạp
`std/collections.ax` **lần thứ hai**. **Đã chứng minh độc lập với bản vá B3** bằng biến thể chỉ-P0.
Repro đã gửi ngân hàng (untracked): `nestbundledmod.ax` + `bin/probe_nestbundled.ax`.
⚠️ `std/iter.ax` còn ~23 lỗi parse của riêng nó ⇒ **không dùng nó làm acceptance test**.

**Một `axiom-investigator` đang chạy nền** để scope stage 2 (agent `ae16a6660309714ee`). Nhiệm vụ:
xác nhận/bác bỏ cơ chế, liệt kê 8 tên bundled từ bảng thật, chỉ ra chỗ sửa đúng lớp, trả lời câu hỏi
**định danh** (D1 quyết định 3: khớp theo **identity**, KHÔNG theo **cách viết**), quan hệ với
`strip_package_prefixes` (stage 3), có cần RFC không, oracle + hiệu chuẩn, và rủi ro với baseline 711.
Nếu phiên mới không thấy báo cáo của nó ⇒ **chạy lại investigator với cùng đề bài**.

## Hàng đợi sau đó (không đổi)
- **Stage 3**: **XOÁ `strip_package_prefixes`** ⇒ hết **B1** (nó **làm hỏng HẰNG CHUỖI**) + mở khoá CỘT chẩn đoán.
- **libc 6-9 (đều B==C)**: `system`, `memcpy`/`memset` (`x86_regs.ax:233`), `strlen`
  (`x86_selector.ax:1202`), và giết `ax_runtime.dll` (gate ELF-only ở `linker.ax:3162-3175`).
- **P4** (RFC 0038): `println` do user định nghĩa bị selector cướp vì khớp **chuỗi tên** ⇒ B==C.
- ⛔ **`atof` cấm đụng nếu chưa có RFC** (`std/string.ax:818` không có số mũ, không làm tròn đúng).
- **Di trú oracle sang §7.1** (stdout thay exit code) — làm **dần**.

## ⚠️ BẪY HẠ TẦNG MỚI PHÁT HIỆN PHIÊN NÀY — ĐỌC TRƯỚC KHI CHẠY SUITE
**`knowledge/lesson-taskstop-leaves-suite-running.md`**. Tóm tắt:
1. **`TaskStop` KHÔNG giết script suite.** Nó chạy tiếp và ghi đè `$REGTMP/reg_<name>.exe` mà lượt sau
   đang đọc ⇒ lượt `-O0` báo `FAIL t_hoftup (got 42 want 44)` **không tái hiện được**. Bằng chứng:
   exe suite đã chạy còn trên đĩa, **25600 byte = kích thước bản `-O1`**, trong khi lượt `-O0` chỉ sinh
   **26112**. **Không được gọi là flake** (§24 mục 4) — phải quy về nguyên nhân có tên.
2. ⇒ **Luôn đặt `REGTMP` riêng cho mỗi lượt.** Lệnh đúng:
   ```sh
   rm -rf /tmp/regO0_iso && mkdir -p /tmp/regO0_iso
   AXC=bin/axc_fpA.exe AXEXTRA=-O0 REGTMP=/tmp/regO0_iso bash scripts/regression_repros.sh
   ```
3. **`regression_repros.sh` KHÔNG đọc `AXFLAGS`** — biến đúng là **`AXEXTRA`** (`:23`). Chạy sai biến
   ⇒ suite chạy **y hệt mức mặc định** rồi báo GREEN cho một mức **chưa từng được kiểm**.
   Dòng build chính `:1057` là `-O1 $AXEXTRA` ⇒ `AXEXTRA=-O0` thành `-O1 -O0` (cờ sau thắng: 26112 byte).
4. **Forensics đúng chỗ:** exe mà suite đã chạy **vẫn còn trên đĩa** — `ls -l` + `cmp` với bản dựng
   sạch trả lời ngay "suite có chạy đúng file nó vừa dựng không", rẻ hơn mọi suy đoán.

## Bài học phương pháp phiên này
- ⭐ **Hai "bug" mô tả bằng triệu chứng bề mặt có thể là MỘT nguyên nhân gốc** (B3 & B6). Ma trận đo
  callee đã cho ra điều kiện kích hoạt thật; hai "manh mối" `raw32` và `hash_key` **đều là hệ quả**.
- ⭐⭐ **Biến thể IM LẶNG nguy hiểm hơn biến thể crash**: thí nghiệm nhồi lên 44.011 node ⇒ **exit 0**,
  chỉ biên mảng đổi. Một OOB read "may mắn trúng vùng hợp lệ" thì đóng dấu chữ ký sai, không ai thấy.
- **Ghi chép cũ trôi lệch**: handoff trước liệt kê `std.thread`/`std.sync` là B3/B6 crash — lúc kiểm
  lại thì `std.sync` đã là **reject sạch**, và `std.string`/`std.collections` **không crash trên đường
  `build`** (chúng được bundle). **Luôn tự chạy lại repro trước khi dựa vào nó.**
- ⚠️ **Đừng để hai bản sao của một sự thật**: lần đầu tôi chép chi tiết B3 vào **cả** `BACKLOG.md` lẫn
  `MEMORY.md` — đúng lớp defect mà CLAUDE.md §24 cảnh báo. Đã tách ra file topic riêng; `BACKLOG.md`
  chỉ còn **con trỏ**.
- **Bản vá frontend KHÔNG nhất thiết byte-identical với seed** (tôi đã đoán sai và đã đính chính):
  code mới vẫn nằm trong self-image dù tự-biên-dịch không đi qua đường đó. **A==B mới là tiêu chí.**

## File chưa track (cố ý để lại, không phải rác)
`nestbundledmod.ax`, `bin/probe_nestbundled.ax` (repro B3b) · `bin/probe_b3.ax`, `bin/probe_b6.ax`,
`bin/probe_b6_min.ax`, `std/testmod.ax` (probe cũ) · `.claude/settings.json` (**file của user — KHÔNG đụng**).
