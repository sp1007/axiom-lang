---
name: session-handoff-2026-09-05b
description: Resume point — B3b root cause FOUND (loader clobbers symtable.current_tree) but UNVERIFIED on branch wip/b3b-current-tree; BACKLOG.md restructured -39%; reject-guard calibrated and staged
metadata:
  type: project
---

# Handoff 2026-09-05b (phiên tạm dừng theo yêu cầu user)

## ⛔ ĐỌC TRƯỚC: cây làm việc KHÔNG sạch, và phần chưa commit là CHƯA ĐƯỢC KIỂM
HEAD = **`49f4dbe`** trên `main`. Cây còn **thay đổi chưa commit của B3b** + fixture untracked.
**Không có gì trong đó đã qua gate.**

| | |
|---|---|
| Driver `bin/axc_native.exe` | **NGUYÊN VẸN**, đã `ls -l` + `sha256sum` SAU khi giết agent: A==B `29DB6D68A6BDD7EE07542577C0344970F99AFAD1B0B640A7811211707C4D37DB`, **2.330.112 byte** |
| Baseline | **714/714** (số của phiên TRƯỚC; phiên này **không chạy trọn suite** — đừng trích dẫn như đã đo hôm nay) |
| Mốc **B==C** gần nhất | vẫn `c3eae77` / `52D1ABD4…` |

## 🔴 VIỆC TIẾP THEO — B3b: đã có CHẨN ĐOÁN THẬT, chỉ còn KIỂM CHỨNG
Bản vá nằm ở nhánh **`wip/b3b-current-tree` (`f71f40c`)** VÀ trong cây làm việc. Hai bản giống nhau;
nhánh là bản lưu chống mất, cây là bản để chạy tiếp.

### Nguyên nhân gốc (implementer tìm ra — hợp lý về kiến trúc, **CHƯA XÁC MINH**)
`ax_driver_load_module` (`main_air.ax`) **ghi đè `symtable.current_tree`** — thanh ghi mà `define`
đóng dấu vào `symbol_trees` (`resolver.ax:562/606/621`) — **và không bao giờ trả lại**. Nạp module là
**TÁI NHẬP (re-entrant)**: nó chạy **giữa chừng** quá trình resolve của kẻ đi import
(`root → load(B) → B resolve → load(A) → A resolve`). Khi quay ra, thanh ghi vẫn trỏ vào **A**, nên
**mọi symbol mà B — rồi cả unit gốc — định nghĩa SAU dòng `import` đều bị đóng dấu thuộc cây của A.**
`sym_decl_tree` sau đó áp **trung thực** một `decl_node` cục bộ của B lên cây của A.

⭐ Giải thích được **cả hai** sự thật đã đo: (1) yếu tố kích hoạt **chỉ là SỐ NODE của A**;
(2) `undefined name 'aparam'` — tham số **chỉ tồn tại trong A** — bị báo trong lúc kiểm B.
⭐ Hai caller (`resolver.ax:1142/:1203`) **đã** save+restore *mảnh trạng thái resolve kia* mà họ biết
(ngăn xếp scope); thanh ghi này **chỉ đơn giản bị bỏ sót** — đúng lớp "một luật, hai bản sao".

**Bản vá:** lưu `current_tree` lúc vào, khôi phục ở **cả bốn** đường `return`.
**Kèm theo:** guard `ax_panic` cho hai chỗ đọc `decl_node` của const **không kiểm biên**
(`air_builder.ax:769/:2096`), và đặt tên **B3b-4** cho phần còn lại: sửa đúng là **phân giải cây SỞ
HỮU** theo cách `:2704` đã làm, cần một handle cây gốc trên `AirModuleBuilder`.
**Guard KHÔNG phải bản sửa đó.**

### Cách chạy tiếp (frontend-only ⇒ tiêu chí là **A==B**, KHÔNG cần B==C)
1. Dựng lại driver, **`ls -l` NGAY sau khi dựng** — bản bị cắt cụt 25–30 KB đọc y hệt một regression
   thảm hoạ; bản đúng ~2,33 MB. ⛔ **Không dựng compiler song song với suite.**
2. `scripts/fast_fixpoint.ps1` ⇒ **A==B**. ⚠️ **Đừng kỳ vọng byte-identical với seed.**
3. Suite ở **CẢ** default **LẪN `-O0`**, mỗi lượt **`REGTMP` RIÊNG**; biến đúng là **`AXEXTRA`**
   (`AXFLAGS` bị bỏ qua ⇒ suite báo GREEN cho một mức chưa từng được kiểm):
   ```sh
   rm -rf /tmp/regO0_iso && mkdir -p /tmp/regO0_iso
   AXC=bin/axc_fpA.exe AXEXTRA=-O0 REGTMP=/tmp/regO0_iso bash scripts/regression_repros.sh
   ```
4. Oracle mới `t_b3bcrossmod` ⇒ baseline **714 → 715**. Hiệu chuẩn trước-khi-sửa đã ghi trong hàng suite.
5. Repro đối chiếu: `./bin/axc_native.exe build bin/probe_b3b_dotted.ax -o /tmp/pb3b.exe`
   ⇒ **trước khi sửa**: exit 1, `undefined name 'aparam'` ×2 (tự chạy lại 2026-09-05, vẫn đúng).

## ✅ ĐÃ SHIP PHIÊN NÀY (5 commit trên `main`)
| commit | nội dung |
|---|---|
| `e2b1ad6` | `BACKLOG.md` 59 KB → 41 KB: tách 3 mục sự thật ra file topic, để lại con trỏ |
| `ccf5402` | tách nốt probe8 ⇒ **35,7 KB (−39%)**; con trỏ nay **gọi tên từng lỗ còn sống** |
| `4699856` | **đo trước** hiệu chuẩn cho reject-guard + **đính chính đề xuất của chính lesson đó** |
| `7638bb8` | chẩn đoán "not bundled on this build" là **SAI** với 3 module đã bundled |
| `49f4dbe` | handoff này + con trỏ `BACKLOG.md` |

### ⭐ Phát hiện đáng giá nhất phiên này (không phải bản vá)
1. **107 hàng `reject` mù với việc compiler SẬP.** Đã **đo trước** để gỡ thế chặn: chạy cả 107 qua một
   **bản sao riêng** của driver ⇒ **107/107 exit=1, không sinh exe, ≥2 dòng chẩn đoán**
   ⇒ **guard là THUẦN BỔ SUNG, không làm đỏ hàng nào** ⇒ không còn bị chặn bởi "phải điều tra từng cái".
   ⚠️ **Đính chính cho chính lesson đó:** tiêu chí "có in chẩn đoán" **KHÔNG** bắt được ca gốc —
   `import std.process` in nhiều `undefined name 'raw32'` **rồi mới** sập (139). **Chỉ tiêu chí
   mã-thoát-là-tín-hiệu-crash mới bắt.** Chi tiết + fixture hiệu chuẩn:
   `knowledge/lesson-reject-comparator-blind-to-crashes.md`.
   ⛔ **Chưa sửa `scripts/regression_repros.sh`** vì lúc đó suite đang chạy (sửa giữa chừng giết suite).
2. **"Hai danh sách" CHƯA hết.** Stage 1 hợp nhất `strip_imports` với bảng 11 mục, nhưng
   `strip_package_prefixes` (`main_air.ax:411-419`) **vẫn giữ danh sách riêng chỉ 8 tiền tố** —
   thiếu đúng **`std.result.`, `std.runtime.`, `std.collections.`**. Cặp control chứng minh:
   `ax_assert(...)` trần **chạy**, `std.runtime.ax_assert(...)` **bị từ chối** kèm câu *"not bundled on
   this build"* — mà bare name chính là bằng chứng nó **ĐÃ** bundled. §8: chẩn đoán phải actionable.
3. ⭐ **B1 tự hiện ra ngay trong probe của tôi**: tôi viết literal `"stripped prefix std.string.len: "`,
   chương trình in ra **`stripped prefix len: `**. Bộ viết lại văn bản sửa **hằng chuỗi** — §3 cấm.
4. ⭐ **Meta-pattern lại xuất hiện, lần 6–7**: `strip_package_prefixes` **tự né B1 cho chính nó**
   (comment BUG#26 ở `:401-409` + chuỗi `concat` để nguồn của nó không chứa literal tiền tố), và
   `preprocessed_module_name` mục 8–9 viết `concat("std.os", ".win32")` **vì literal sẽ tự viết lại
   chính nó**. ⇒ **Đụng phải bug → mô tả đúng cơ chế → né tại chỗ, không sửa.** Xem
   `knowledge/lesson-comment-protects-one-line-only.md`.

## ⚠️ BẪY ĐÃ TÁI XÁC NHẬN BẰNG ĐO ĐẠC PHIÊN NÀY
**`TaskStop` KHÔNG giết script suite** — xác nhận lại: sau khi `TaskStop` agent, `ps` vẫn thấy
**`bin/axc_fpA` đang chạy**; phải `kill -9` thủ công. ⇒ **Luôn `ps` sau khi giết agent**, rồi
**`ls -l` + `sha256sum` driver** trước khi tin bất cứ số đo nào. (Đã làm: driver còn nguyên.)

## 🧾 Việc dọn dẹp còn treo cho phiên sau
- Nhánh **`wip/b3b-current-tree` (`f71f40c`)** là **local, chưa push**. Nó **UNVERIFIED** — đừng
  merge; dùng làm lưới an toàn nếu cây làm việc bị mất.
- ⭐ **Bài học tự rút:** commit `7638bb8` (bản đầu `23ef7cf`) **vô tình nuốt 3 file fixture** vì implementer đã
  `git add` sẵn chúng và `git commit` commit **TOÀN BỘ index**. Đã sửa (soft-reset + `restore --staged`).
  ⇒ **Sau khi một agent chạy xong, luôn `git status` TRƯỚC khi commit bất cứ thứ gì "chỉ là docs".**
- File untracked cố ý giữ: `bin/probe_b3b_dotted.ax`, `bin/probe_b3*.ax`, `bin/probe_b6*.ax`,
  `bin/probe_nestbundled.ax`, `nestbundledmod.ax`, `b3bmod_a.ax`, `b3bmod_b.ax`, `std/testmod.ax`.
  `.claude/settings.json` = **file của user, KHÔNG đụng**.

## Hàng đợi sau B3b (thứ tự đã định giá, không đổi)
1. **B3b** (đang dở, trên) → 2. **reject-guard** (đã hiệu chuẩn, rẻ, thuần bổ sung) →
3. **`strip_package_prefixes` biết-literal** ⇒ đóng **B1** + mở khoá **CỘT** trong chẩn đoán;
   nay **rẻ hơn** vì đã biết chính xác 3 tiền tố còn thiếu (mục 2 ở trên). ⚠️ Phải **đo** A==B: chính
   `cgen.ax:773,777` và `resolver.ax:807-818` so với literal `"std.string.len"` **đang bị viết lại**.
4. **Rồi mới** định giá lại stage 2 — **BẮT BUỘC RFC**, và **stage 4 là TIỀN ĐỀ**, không phải phần sau.

Liên quan: [[bug-b3b-cross-module-index-and-loader-defects]] · [[bug-cross-tree-decl-node-segv]] ·
[[lesson-reject-comparator-blind-to-crashes]] · [[lesson-taskstop-leaves-suite-running]]
