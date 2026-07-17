---
name: feedback-fixpoint-async-rule
description: Rule user-approved 2026-07-03 — fixpoint chạy ASYNC (sau commit) cho thay đổi frontend-thuần; chỉ bắt buộc fixpoint-trước-commit với codegen/ABI/linker
metadata: 
  node_type: memory
  type: feedback
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

User duyệt (2026-07-03): **"fixpoint-async cho thay đổi frontend-thuần"** để tránh chu kỳ chờ nhiều giờ mỗi commit (fixpoint = 3 lần self-compile 790 hàm, có thể 6-8h khi regalloc còn O(n²)).

**Why:** fixpoint KHÔNG validate tính đúng của feature frontend mới (compiler source không dùng feature đó → dead-code khi self-compile); nó chỉ chứng minh non-regression + determinism. Phần non-regression đã được full regression (scripts/regression_repros.sh, 92 test) cover. Chờ 6-8h cho mỗi commit frontend là phí.

**How to apply:**
- **Frontend-thuần** (parser/resolver/typecheck/AIR-lowering mới, KHÔNG đụng ssa_opt/x86_*/linker/cgen): regression GREEN + oracle test đủ để **commit + push ngay**; fixpoint chạy async sau đó. Nếu NO FIXPOINT → revert commit.
- **Backend/ABI/linker/optimizer** (ssa_opt.ax, x86_selector/regalloc/emitter/coff/elf, linker.ax, cgen.ax): fixpoint **BẮT BUỘC trước commit** như cũ ([[feedback-auto-commit]]).
- Ranh giới không chắc → coi là backend (an toàn trước).

Áp dụng lần đầu: closures P1 commit trong khi fixpoint stage2 đang chạy (regression 92/92 GREEN, t_lambda oracle 5/5).
