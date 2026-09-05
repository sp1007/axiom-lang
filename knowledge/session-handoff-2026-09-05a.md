---
name: session-handoff-2026-09-05a
description: Resume point — B3/B6 + B3b-3 shipped (cc4a2cf, baseline 713); B3b mechanism refuted and the stage order re-priced; B3b-2 in flight
metadata:
  type: project
---

# Handoff 2026-09-05a

## ✅ TRẠNG THÁI: SẠCH, ĐÃ COMMIT + PUSH, GREEN
HEAD = **`cc4a2cf`** `fix(loader): a module with parse errors stops the build instead of crashing it`.
Đã push lên `main`. **Một implementer đang chạy nền cho B3b-2** (UAF loader) — chưa commit gì.

| | |
|---|---|
| Driver `bin/axc_native.exe` | A==B **`E44344648837F413CB631C3CC1C3D8137C0F8D9B8B6EE35614DBF180C7284744`**, **2.329.600 byte** (đã `ls -l`) |
| Baseline | **713/713** ở **CẢ** default **LẪN `-O0`** — **tự chạy lại, không lấy số của agent** |
| Mốc **B==C** gần nhất | vẫn `c3eae77` / `52D1ABD4…` ⇒ **thay đổi backend kế tiếp phải dựng lại B==C từ driver MỚI này** |

## Đã ship phiên này
**B3 + B6 = MỘT bug** — `pre_infer_func_signature` (`typecheck.ax:3770`) đánh chỉ số `decl_node` của
unit KHÁC vào `self.tree`; ba call site (`:1626`, `:5898`, `:7081`) trong khi **bản sao thứ tư `:1985`
đã được canh sẵn**. Chi tiết đầy đủ: **`knowledge/bug-cross-tree-decl-node-segv.md`**.
`import std.{sync,thread,collections,string}` + module user dùng `@compiler_intrinsic`: **139 → 0**.

## 📦 ĐÃ SHIP THÊM SAU B3/B6 (cùng phiên)
`cc4a2cf` **B3b-3** — loader dừng khi module có lỗi parse (§9): `std.iter`/`std.cli` **139 → 1**,
`std.process` vẫn 139 (đúng: nó chết ở B3b thật). Baseline **711 → 713**.
`4cc1617` bỏ chặn quy tắc oracle §7.1 (P1 đã đóng từ RFC 0038 — dòng "bị chặn" đã cũ).
`9646de4` **RFC 0040** `--verbose` (proposed). `5a2c385` + `9beeda6` + `500d550` + `5b659f0` tài liệu.

## ⚠️ HAI PHÁT HIỆN VỀ HẠ TẦNG KIỂM THỬ (quan trọng hơn bản vá)
1. ⛔ **`cmp=reject` KHÔNG phân biệt "từ chối sạch" với "compiler SẬP"** — nó chỉ xét "không có exe",
   mà SIGSEGV cũng không sinh exe. **107 hàng `reject`** đang mù với việc compiler sập trên đúng
   chương trình chúng canh. Tôi đã **đưa gợi ý comparator SAI** trong đề bài B3b-3; chỉ **quy tắc hiệu
   chuẩn bắt buộc** mới bắt được. Đề xuất guard: `knowledge/lesson-reject-comparator-blind-to-crashes.md`
   — **chưa cài**, cần hiệu chuẩn riêng và có thể làm đỏ vài hàng trong 107 (đó là tính năng).
2. ⭐ **Meta-defect lặp 4 lần trong một ngày**: luật được **ghi đúng** rồi vẫn bị vi phạm, vì prose gắn
   với **vị trí** còn luật sống ở **khái niệm**. Xem `knowledge/lesson-comment-protects-one-line-only.md`.
   Áp luật đó vào chính nó đã **sửa được chẩn đoán**: có **HAI** chỗ free alias, không phải một.

## 🔴 VIỆC TIẾP THEO — theo THỨ TỰ ĐÃ ĐỊNH GIÁ LẠI (không phải thứ tự cũ của lộ trình)
⛔⛔ **Cơ chế cũ ghi ở đây ("nạp `std/collections.ax` lần thứ hai") ĐÃ BỊ BÁC BỎ 2026-09-05.**
Ba phép đo: `--no-stdlib` (không hề trùng lặp) **vẫn hỏng**; trùng `std.result` **build sạch**; và
**hai module user thuần** tái hiện y hệt. Nguyên nhân thật là **chỉ số node xuyên cây** (họ B3/B6,
đường mono/`clone_subtree_from`), yếu tố kích hoạt là **SỐ NODE của module được import**.
⇒ **B3b KHÔNG phải vấn đề stdlib/bundling.** Toàn bộ bằng chứng + phán quyết RFC:
**`knowledge/bug-b3b-cross-module-index-and-loader-defects.md`**.

**Thứ tự đúng (rủi ro thấp trước):**
1. ✅ **B3b-3** — xong (`cc4a2cf`).
2. 🔄 **B3b-2** — UAF loader, **đang chạy nền**. **HAI** chỗ free alias: `main_air.ax:1951` (tệ hơn —
   free rồi truyền `mod_name` vào `register_module_from_lib`) **và** `:1957`, cộng chỗ thứ ba mới
   thêm ở `cc4a2cf`. Làm **route (a)** (sửa call site); ⛔ **KHÔNG** đụng `replace` (route (b) chạm
   24 call site, phải tách commit + đo A==B riêng).
3. **B3b thật** — chỉ số xuyên cây: đưa qua `typecheck.ax:3793 sym_decl_tree` + kiểm biên §9.
   Site cần audit: `typecheck.ax:5416-5427` (**không kiểm biên**), `:1206-1216`, `mono.ax:418-437`,
   `mono.ax:473-529`, `air_builder.ax:769/:2096` (không kiểm) vs `:2705/:4005` (có kiểm).
   Repro: `b3bpkg/moda.ax` + `b3bpkg/modb.ax` + `bin/probe_b3b_dotted.ax` (untracked, đã bank).
4. **`strip_package_prefixes` biết-literal** ⇒ đóng **B1** + mở khoá **CỘT** trong chẩn đoán.
   **Rẻ hơn stage 3 rất nhiều**, một hàm, không RFC.
5. **Rồi mới** định giá lại stage 2 — ⛔ **stage 2 như đề xuất VI PHẠM D1-3** (buộc khớp theo chính
   tả; vỡ trên overload `map`/`unwrap`), và **stage 4 là TIỀN ĐỀ của nó**, không phải phần đi sau.
   **BẮT BUỘC RFC.**

⚠️ `std/iter.ax` còn ~23 lỗi parse của riêng nó ⇒ **không dùng nó làm acceptance test**.

## Hàng đợi sau đó
- ⛔ **Stage 3 (xoá `strip_package_prefixes`) KHÔNG rẻ — nó PHỤ THUỘC stage 2** (lời gọi
  `std.string.replace` của chính compiler hôm nay chỉ chạy nhờ viết lại văn bản). Dùng mục 4 ở trên thay thế.
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
