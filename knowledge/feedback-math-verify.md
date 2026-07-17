---
name: feedback-math-verify
description: "Hàm toán phải kiểm chứng kết quả CẨN THẬN — đối chiếu oracle nhiều điểm, sai số chặt, báo max-error; không chỉ vài check exit-code"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 9720703b-c141-45a7-9477-748fe53e183d
---

User (2026-06-29): "những hàm toán học thì phải kiểm chứng kết quả cẩn thận".

**Why:** Vài check sanity exit-code với tolerance lỏng + giá trị tham chiếu tự-tính-tay KHÔNG đủ tin cậy (đặc biệt hàm float/xấp xỉ: trig/exp/ln/erf/gamma/cdf). Có thể "pass" mà vẫn sai số lớn.

**How to apply:** Với mọi hàm toán (nhất là float):
- Dùng ORACLE độc lập (Python math/mpmath) sinh giá trị tham chiếu chuẩn.
- Kiểm trên LƯỚI nhiều điểm phủ miền xác định (gồm biên/góc đặc biệt), không chỉ 1-2 điểm.
- Sai số tương đối CHẶT (vd 1e-6 cho elementary; nêu rõ tolerance từng hàm theo bản chất xấp xỉ — erf~1e-7, Lanczos gamma~1e-10).
- BÁO CÁO max-error thực tế, không chỉ pass/fail; nếu hàm xấp xỉ kém → cải thiện thuật toán.
- Hàm nguyên (combinatorics/numtheory) so khớp tuyệt đối.
- Hạ tầng: tests/mathlib/ (oracle.py sinh + bundle std/math.ax + validator). Liên hệ [[next-step-15-selfhost-status]] (đã có tests/arith/ cho matrix số học).
- Pattern bundle-verify: `cat std/math.ax + main` rồi build axc_stage1 (compiler không import std.*); array literal `[T;N]` HỎNG → unroll check hoặc @alloc.
