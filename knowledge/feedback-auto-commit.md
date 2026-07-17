---
name: feedback-auto-commit
description: Tự động git commit + push mỗi khi giải quyết xong một bug/vấn đề
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

Mỗi khi fix xong một bug/vấn đề (không phải mỗi thay đổi nhỏ), TỰ commit và push lên git — không cần hỏi.

**Why:** User muốn lịch sử bug-fix được lưu ngay, tránh mất việc khi context dài/compact.
**How to apply:** Sau khi xác nhận một fix hoạt động (verified), commit các file SOURCE liên quan (vd bootstrap/stage1/*.ax, knowledge/bugs.md) với message mô tả bug + root cause, rồi `git push`. User làm việc trực tiếp trên branch `main` (xem lịch sử commit) → commit thẳng main, KHÔNG cần tạo branch. KHÔNG commit file build tạm (axiom_temp.obj, *.exe, *.log). Liên kết [[next-step-15-selfhost-status]].
