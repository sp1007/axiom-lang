---
name: task-cross-library-name-collision
description: "TASK (do later, user-requested 2026-07-24): giải quyết TRIỆT ĐỂ họ name-collision khi link nhiều thư viện (std + non-std). Hiện chỉ có mangling fn-vs-fn cross-module (BUG#50, flag 2048 -> ax_<name>__m<idx>); còn hở fn-vs-struct và các họ khác. Cần một scheme định danh symbol nhất quán + kiểm tra collision ở link."
metadata:
  node_type: memory
  type: project
---

**TASK (user 2026-07-24, thực hiện SAU):** "giải quyết triệt để vấn đề hàm trùng tên giữa các thư viện."

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
