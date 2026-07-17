---
name: feedback-compact
description: "User wants Claude to auto-compact context when it's getting full"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

Tự động /compact context khi gần đầy (khi conversation context gần đạt giới hạn).

**Why:** User explicitly requested this behavior — tránh mất context giữa chừng.

**How to apply:** Khi conversation đang dài và gần limit, chủ động compact trước khi bị truncate. Không cần hỏi user.
