---
name: feedback-autonomy
description: "Với lựa chọn về định hướng/kiến trúc, tự chọn phương án tối ưu và hoàn thành mục tiêu — KHÔNG hỏi lại user"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

Khi gặp lựa chọn về ĐỊNH HƯỚNG (thứ tự task, cách tiếp cận, trade-off kiến trúc), tự chọn phương án **tối ưu/hiệu quả nhất** và thực hiện tới khi hoàn thành — KHÔNG dừng để hỏi user xác nhận "có nên làm không".

**Why:** User nói rõ (2026-06-26): "lần sau nếu có lựa chọn về định hướng như vậy, hãy chọn phương án tối ưu, hiệu quả nhất, đừng hỏi lại tôi. cứ tự động hoàn thành mục tiêu."

**How to apply:** Chỉ dùng AskUserQuestion khi câu trả lời thực sự đổi ngữ NGHĨA ngôn ngữ (semantics chưa chốt trong spec) hoặc ảnh hưởng ABI/grammar mà spec im lặng. Còn lại: chọn, ghi rationale vào RFC/commit, làm. Kết hợp [[feedback-auto-commit]] (tự commit+push main mỗi bug xong) và [[feedback_compact]].

**Tái khẳng định 2026-06-28:** "tiếp tục, lần sau bắt đầu ngay theo hướng có lợi nhất, không cần hỏi tôi" — KHÔNG kết thúc turn bằng câu hỏi "bạn muốn tôi làm X không"; cứ bắt đầu hướng có lợi nhất ngay. VÀ: trong mọi tác vụ chạy lâu (verify/build), **bắt buộc Monitor persistent phát heartbeat mỗi 10 phút** (600s) để user biết agent còn hoạt động (liveness check `ps -W | grep axc_stage`, KHÔNG dùng pgrep — rc=127 trên box này).
