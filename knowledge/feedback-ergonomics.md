---
name: feedback-ergonomics
description: "User coi trọng ergonomics/DX của ngôn ngữ AXIOM — ghét boilerplate thừa (vd phải gõ `as u32` khắp nơi)"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

User (tác giả AXIOM) coi trọng trải nghiệm lập trình viên: phàn nàn việc phải gõ ép kiểu `as u32` quá nhiều (kể cả trong biểu thức so sánh) là "rất bực mình". Dẫn tới RFC 0005 (int literal type inference trong binary expr).

**Why:** AXIOM nhắm production-grade + tự host; ngôn ngữ phải dễ chịu khi viết, không bắt người dùng lặp annotation/cast máy móc.

**How to apply:** Khi thiết kế/review feature ngôn ngữ, chủ động giảm boilerplate ở những chỗ AN TOÀN — ưu tiên type inference, literal suy kiểu theo ngữ cảnh, sugar — miễn KHÔNG hi sinh tính tường minh ở chỗ nguy hiểm (vd implicit coercion giữa hai biến khác kiểu thì KHÔNG, dễ bug width/sign). Mỗi thay đổi type-system phải có RFC + giữ fixpoint (guard để code đã-cast không đổi codegen). Liên quan [[next-step-15-selfhost-status]].
