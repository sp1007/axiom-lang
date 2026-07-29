---
name: task-cross-library-name-collision
description: "TASK (do later, user-requested 2026-07-24): giải quyết TRIỆT ĐỂ họ name-collision khi link nhiều thư viện (std + non-std). Hiện chỉ có mangling fn-vs-fn cross-module (BUG#50, flag 2048 -> ax_<name>__m<idx>); còn hở fn-vs-struct và các họ khác. Cần một scheme định danh symbol nhất quán + kiểm tra collision ở link."
metadata:
  node_type: memory
  type: project
---

**TASK (user 2026-07-24, thực hiện SAU):** "giải quyết triệt để vấn đề hàm trùng tên giữa các thư viện."

## 🟢 2026-07-29e — P1 ĐÃ SHIP (`b5c1c83`, B==C `4955808B`, 554/554) + **RFC 0035** đã viết
Xem `rfcs/0035-symbol-namespacing-and-link-collisions.md`. Diagnostic `E0501` ở linker
(`report_duplicate_definitions`) báo khi một emit-name bị định nghĩa >1 lần — trước đây mọi lookup
là **linear FIRST-MATCH** nên bản định nghĩa thứ hai bị bỏ ÂM THẦM.

⭐⭐ **Viết diagnostic xong là nó bắt NGAY một defect SỐNG trên compiler đã ship** — repro:
`libpa.ax`/`libpb.ax` cùng khai `pub fn helper`, app `--auto-lib` gọi cả hai:
- `warning[E0501] ... ax_helper` (định nghĩa 2 lần)
- `error: linker: unresolved external symbol 'ax_helper__m1755'`

**Root (đây mới là điều quan trọng nhất, và nó bác bỏ hướng vá cũ):** mitigation cũ
`SYM_FLAG_MODDUP`(2048) + `ax_<name>__m<sym_idx>` là **state THEO TỪNG LẦN BIÊN DỊCH**, nên khi
thư viện được biên dịch RIÊNG (RFC 0011 `--auto-lib`) thì:
1. **Callee KHÔNG mangle** — mỗi lib compile một mình, không có gì trùng tên trong lần compile đó
   ⇒ flag 2048 không bật ⇒ cả hai phát `ax_helper` ⇒ trùng, first-wins.
2. **Caller CÓ mangle, ra một tên không ai định nghĩa** — app thấy 2 import trùng tên ⇒ gọi
   `ax_helper__m1755`, mà `sym_idx` là chỉ số trong bảng symbol của *bên import*.
⇒ **Caller và callee bất đồng về TÊN của cùng một hàm.** Một scheme mangling suy ra từ state
per-compilation **không thể** là hợp đồng link-time. `sym_idx` KHÔNG ổn định giữa các lần build —
đúng như note này dự đoán, nay đã CHỨNG MINH bằng repro.

⚠️ **Giữ mức WARNING, không phải error — đây là phát hiện chứ không phải rụt rè:** cùng lần chạy
thấy `ax_Ok`/`ax_Err`/`ax_Some`/`ax_None`/`ax_sum_layout_is_pointer`/`ax_block_size` cũng bị trùng
vì **mỗi lib nhúng bản sao runtime shim riêng** ⇒ trùng định nghĩa là **BÌNH THƯỜNG** trên đường
multi-lib hiện nay; đổi thành error sẽ phá đường đó. Phân biệt "shim trùng vô hại" với "hai hàm
khác nhau" chính là thứ P2 làm cho quyết định được (P3 mới nâng lên error).

Gate cost: gated on multi-object link (1 object thì scan không thể nổ). Đo trung thực: gated vs
ungated self-build đều 8,3–8,5s, chồng lấp ⇒ **không đo được khác biệt**; con số 7,3→8,1s trước đó
là NHIỄU của máy. Zero false positive trên 554 chương trình + self-link.

## Trạng thái hiện tại (cơ chế đã có, và các lỗ hổng)
- **fn-vs-fn cross-module ĐÃ có mangling**: `x86_resolve_sym_name` (x86_regs.ax:245-271) — nếu symbol có `flags & 2048` (cross-module same-name) thì emit `ax_<name>__m<sym_idx>` để hai module có `close`/`open`… không merge first-import-wins ở link. Đây là bản vá BUG#50 (`82d0565`, `b966f73`).
- **CÒN HỞ — họ collision (memory: [[bug-user-fn-stdlib-struct-name-collision]])**: đã ghi nhận **5 hole**; 4 hole đầu là fn-vs-fn, hole thứ 5 là **fn-vs-STRUCT name** (user `fn worker` trùng `struct worker` bundled trong scheduler.ax → `call worker` link sang symbol stdlib khác → user fn không chạy → sai/crash). Workaround hiện tại: đừng đặt tên trùng bundled stdlib. Repro `bin/known_fail_worker_name_collision.ax`.
- **Vấn đề gốc**: định danh symbol emit ra KHÔNG mang namespace module một cách nhất quán — mangling chỉ bật theo flag 2048 (heuristic "same-name cross-module"), không phải scheme toàn cục. Struct-method/ctor/global cũng phát symbol theo tên, có thể đụng.

## Hướng cần thiết kế (khi thực hiện)
1. **Scheme mangling nhất quán mọi symbol** (fn, method, ctor, global) mang định danh module/namespace ổn định — KHÔNG dựa `sym_idx` (không ổn định giữa build) mà dùng module-path hash hoặc intern-id ổn định. ⚠️ Phải giữ **deterministic** (§3) và không phá **B==C fixpoint** (backend change).
2. **Collision check ở LINKER**: khi hai symbol định nghĩa cùng tên emit-name, phải là DIAGNOSTIC (E-code) chứ không phải first-wins âm thầm. Đây là "diagnostics là product feature" (§8).
3. **Bao trùm CẢ fn-vs-struct/ctor**: symbol namespace của type-constructor và free-fn phải phân tách (prefix khác nhau) để `fn worker` và `struct worker` không đụng.
4. **RFC**: đây là ABI/linker change → cần RFC (§13) + B==C trước commit.

## Liên quan
- [[bug-user-fn-stdlib-struct-name-collision]] — bug OPEN cụ thể (hole thứ 5), có repro.
- DFE root kind 4 vừa ship (2026-07-24) cũng là họ "predicate so tên qua ranh giới tầng → mangling là một phần phép so" — cùng bài học.
- Câu hỏi kèm theo của user (library bloat) → trả lời: xem [[rfc0031-dead-function-elimination]] + phần dưới.

## Ghi chú trả lời user: "thư viện non-std có bị attach TOÀN BỘ mã vào exe không?"
- **Trước DFE default-on**: CÓ — native path bundle nguyên stdlib + toàn bộ hàm của module import vào exe (~75KB overhead cố định; `return 42` = 77.824B).
- **Sau DFE default-on (ship 2026-07-24)**: KHÔNG với hàm CHẾT — DFE prune ở codegen mọi hàm không reachable từ root set (entry + `#[export]` + ABI-shadow + extern-C-shadow + spawn). Thư viện non-std import-as-source (vào `mod.funcs`) chỉ còn lại hàm chương trình THỰC SỰ gọi (transitive). `return 42`: 77.824 → 22.016B (−71,7%).
- **Ngoại lệ**: `--staticlib`/`--shared` KHÔNG prune (thư viện phải giữ nguyên public surface). Và `.lib` prebuilt link qua `-l` thì DFE ở codegen KHÔNG thấy nội bộ `.lib` (linker xử lý archive member riêng — thường chỉ kéo member được tham chiếu).
