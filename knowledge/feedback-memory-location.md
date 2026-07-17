---
name: feedback-memory-location
description: "User directive (2026-07-17): store all project memory/knowledge in ./knowledge inside the repo, not the global .claude memory dir."
metadata:
  type: feedback
---

Lưu TẤT CẢ project memory/knowledge vào **`d:\projects\compiler\Axiom\knowledge\`** (relative `./knowledge`), KHÔNG dùng global `.claude/projects/.../memory/` nữa.

**Why:** user muốn knowledge version-controlled cùng code trong repo (2026-07-17: "về sau, memory lưu vào ./knowledge" + "các thứ cần ghi nhớ cũng lưu vào knowledge").

**How to apply:** ghi mọi memory file MỚI vào `knowledge/` với cùng frontmatter format; index chính = `knowledge/MEMORY.md` (thêm 1 dòng con trỏ mỗi memory). Global `.claude/.../memory/MEMORY.md` giờ chỉ là stub chuyển hướng → đọc `knowledge/MEMORY.md` ĐẦU TIÊN mỗi session. Các topic file cũ đã copy sang `knowledge/`.
