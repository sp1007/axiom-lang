---
name: feedback-autonomous
description: "Làm việc hoàn toàn tự chủ — tự chọn hướng tối ưu và thực hiện, không hỏi lại"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

User KHÔNG thường xuyên quan sát → **không hỏi lại, không chờ xác nhận, không "ngủ
đông"**. Tự động xác định phương hướng tối ưu (giá trị/rủi ro) và THỰC HIỆN luôn.

**Why:** User đã nói rõ nhiều lần ("đừng chờ hỏi tôi để ngủ đông", "tự động xác định
phương hướng tối ưu và thực hiện, đừng hỏi lại tôi, tôi không thường xuyên quan sát").

**How to apply:**
- Chọn task tối ưu và làm ngay; chỉ dừng hỏi khi thực sự bế tắc cần quyết định của
  user (hiếm).
- Khi build degraded / chờ lâu: chuyển sang việc build-independent có giá trị (RFC,
  spec, docs, cleanup, phân tích) thay vì đứng chờ.
- Giữ standing rules: fix→commit→push thẳng main ([[feedback-auto-commit]]); backend/
  optimizer đổi → regression + fixpoint TRƯỚC khi commit; library-only → không cần
  fixpoint. Report trung thực (test fail thì nói fail).
- Kỷ luật verify (bài học [[next-step-16-fnptr-shipped]]): đúng cú pháp, isolated build,
  đừng tự báo động sai.
