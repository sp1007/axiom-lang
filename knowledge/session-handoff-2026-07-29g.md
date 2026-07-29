---
name: session-handoff-2026-07-29g
description: "HANDOFF 2026-07-29g — HEAD 6141531 sạch+pushed, driver seed==A==B 2077495B, 561/561. Coalescing ĐÃ SHIP: xorshift −30,2%, bằng asm floor. Việc kế tiếp: đo lại toàn bộ perf_suite để xem mốc M6-codegen (≤15% so ASM floor) đã ĐẠT chưa."
metadata:
  type: project
---

# HANDOFF 2026-07-29g — **ĐỌC ĐẦU TIÊN**

## Trạng thái
- HEAD `6141531`, đã **push lên `origin/main`**, cây sạch (chỉ `.claude/settings.json`
  untracked — **của user, đừng đụng**).
- Daily driver `bin/axc_native.exe` = **`2077495B`**, và nó **tự tái tạo**: fast_fixpoint cho
  **seed == A == B**, cộng thêm **B == C** đã chạy riêng trước đó.
- Gate đầy đủ XANH: regression **561/561** (+3 dòng `t_coalescedest`), ELF 12/12, ctgc 16/16,
  exe_size 4/4, lib_collision 6/6, so_export ✓.

## Đã ship phiên này (một commit)
`6141531` **coalesce_dest_copy** — fold cặp copy của dạng 2-toán-hạng phá huỷ.
Chi tiết đầy đủ + bài học ở [[m6-perf-baseline]] mục 2026-07-29g. Tóm tắt:
- xorshift **313,7 → 218,8 ms (−30,2%)**, asm floor 217,4 ⇒ **1,006x**. Đo ghép cặp xen kẽ,
  **2 vòng độc lập** cho cùng kết quả. Trần dự báo 89,8 ms, đo được 94,9 ms — **dự báo ĐÚNG**.
- fib −2,6%, arrwalk −1,8%, callloop +0,5%.
- Trả lời câu hỏi bỏ ngỏ của handoff 07-29f: **giả thuyết 1 đúng** — `vD` loop-carried nên
  interval gộp-def phủ cả vòng lặp, temp nằm hẳn bên trong ⇒ interfere THẬT ⇒ bias trong
  allocator **về nguyên tắc** không với tới. Không phải hạn chế first-move-wins.

## VIỆC TIẾP THEO (đã chọn, chưa bắt đầu)
**Đo lại toàn bộ `scripts/perf_suite.ps1` để định vị mốc M6-codegen.**
Mốc (quyết định D1 của user): **≤15% so với ASM floor** viết tay cùng hình dạng.
- xorshift nay **1,006x** ⇒ ĐẠT.
- fib trước phiên này là 1,24x floor; coalescing chỉ mua −2,6% ⇒ **có thể vẫn ~1,2x**, cần đo.
- arrwalk/callloop chưa rõ sau thay đổi này.
⇒ Chạy suite, xác định shape nào CÒN ngoài 15%, rồi định giá khiếm khuyết của riêng shape đó
bằng biến thể NASM (`scripts/perf_asm_variants.ps1`) **TRƯỚC KHI viết code** — quy trình này đã
loại được 3/6 ứng viên vô giá trị và vừa dự báo đúng lần này.
⚠️ **Một lần chạy `perf_suite`/`perf_fib` KHÔNG đáng tin** (phương sai 8–10%/lần chạy trên máy
này). Mọi con số phải đo GHÉP CẶP xen kẽ, các vòng KHÔNG chồng lấp, và lặp lại ít nhất 2 vòng.

## Backlog còn lại (sau việc trên)
- RFC 0035: method/global/ctor vẫn dùng scheme cũ (`axS_`/`axG_`/`axC_` chưa làm) ⇒ fn-vs-struct
  vẫn dựa reject ở typecheck. P3 (E0501 → error) vẫn bị chặn bởi shim runtime trùng lặp hợp lệ.
- `mod_name` rỗng ở `register_module_from_lib` (binding `mod.NAME` là đồ chết) — vô hại, nhưng
  là bẫy cho người đọc.
- Module path nhiều đoạn không resolve ở call site (`bin.libcol.liba.helper()`) — có sẵn từ trước.
- M6-opt (accumulator/tail-rec) là milestone RIÊNG với M6-codegen — ROI cao hơn allocator và
  KHÔNG đụng code self-host-critical.
- Nếu hết việc: `axiom-bug-probe` để nạp lại backlog.

## ⛔ Cảnh báo còn hiệu lực
- **ĐỪNG cài loop rotation / bottom-test** — đã đo **+0,1%**.
- **ĐỪNG định giá coalescing/copy trên fib** — fib latency-bound, copy vô hình ở đó.
- Backend/ABI/linker ⇒ **B==C bắt buộc** trước commit (A!=B là bình thường khi codegen đổi).
