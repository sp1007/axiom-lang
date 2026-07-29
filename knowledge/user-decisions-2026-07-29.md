---
name: user-decisions-2026-07-29
description: "Bốn quyết định user chốt 2026-07-29 (mốc M6, M4 suite, cross-library collision, phạm vi tự chủ autopilot) — ĐỪNG HỎI LẠI, thi hành trực tiếp."
metadata:
  type: feedback
---

User chốt 2026-07-29 khi được hỏi các mục "cần user quyết" trong backlog, kèm chỉ thị rõ:
**"ghi nhớ, lần sau không hỏi lại, sau đó tự động xử lý"** ⇒ đây là quyết định THƯỜNG TRỰC.
Không hỏi lại bất kỳ mục nào dưới đây ở các phiên sau.

## D1 — Mốc M6: TÁCH làm hai milestone
- **M6-codegen** = khoảng cách so với **NASM viết tay cùng hình dạng** (ASM floor), mục tiêu **≤15%**.
  Hiện fib ở **1.18x floor** ⇒ còn ~90–100 ms, toàn bộ là **register coalescing** + copy-propagation.
- **M6-opt** = milestone RIÊNG cho biến đổi thuật toán (accumulator / tail-recursion). ROI cao hơn
  allocator và KHÔNG đụng code self-host-critical.
- **Why:** đã CHỨNG MINH bằng đo (không phải suy luận) rằng mốc cũ "≤5% clang" bất khả bằng backend:
  clang biến lời gọi đệ quy thứ 2 của fib thành VÒNG LẶP, nên backend hoàn hảo vẫn = 1.55x clang.
  Trộn 2 loại gap vào một con số làm mốc không đo được. Xem [[m6-perf-baseline]].
- **How to apply:** báo cáo perf từ nay dùng cột **ASM floor**, không dùng riêng tỉ lệ clang. Mốc
  "≤5% clang" trong `docs/tasks/milestones.md` đã bị THAY THẾ — đừng khôi phục.

## D2 — M4 "100 compliance tests": VIẾT LẠI suite theo grammar thật
- Chuyển `tests/axiom_compliance_suite.ax` (681 dòng) sang grammar AXIOM chính thức; **CẮT** các test
  cần `std.gpu` / `std.quantum` / `std.net` / `std.compiler.ai` (module chưa tồn tại).
- **Why:** dialect `=>` / `.length()` / `impl Trait for Type` đã bị TỪ CHỐI ở quyết định A4
  (2026-07-22, [[group-a-design-decisions-2026-07-22]]) vì mâu thuẫn với design — nên suite hiện
  không parse được và M4 là gate bất khả. Viết lại biến M4 thành gate ĐO ĐƯỢC.
- **How to apply:** làm dần, mỗi lô test = 1 tick của autopilot. KHÔNG thêm cú pháp dialect để suite
  chạy được — đó chính là thứ A4 đã bác.

## D3 — Cross-library name collision: DUYỆT TRƯỚC cure thật
- Được phép viết RFC + implement **symbol module-namespaced nhất quán** (không dựa `sym_idx`) +
  **linker BÁO LỖI khi trùng** thay vì first-wins âm thầm. Đóng dứt điểm cả họ 5 lỗ.
- **Why:** mitigate hiện tại (mangling fn-vs-fn flag 2048 + reject fn-vs-struct `C432EA9E`) chỉ che
  các ca ĐÃ BIẾT; scheme symbol vẫn mong manh và linker vẫn im lặng khi trùng.
- **How to apply:** ABI/linker change ⇒ **B==C bắt buộc TRƯỚC commit** (§24). Xem
  [[task-cross-library-name-collision]].

## D4 — Phạm vi tự chủ autopilot: TOÀN QUYỀN (chọn cả 4)
1. **Tự commit sau khi gate GREEN** — chạy fixpoint (A==B frontend / **B==C backend-linker**) +
   regression đầy đủ; GREEN thì commit luôn, RED thì tự sửa. Không hỏi.
2. **Tự viết RFC khi §13 yêu cầu** — soạn trong `rfcs/`, tự duyệt theo spec, rồi implement. KHÔNG
   chờ user duyệt RFC.
3. **Tự chạy mọi lệnh build/test/git routine** — xác nhận lại directive §24, không dừng xin phép.
4. **Được phép ôm việc rủi ro cao kéo dài nhiều phiên** (allocator maturity, namespacing ABI), miễn
   mỗi bước **isolated + measured + reversible**.
- **Why:** user muốn vòng lặp tự chủ thật sự, không bị chặn bởi câu hỏi xác nhận.
- **How to apply:** tự chủ đổi *tốc độ*, KHÔNG đổi *luật*. §3 (absolute rules), §9 (IR verification),
  §13 (RFC policy) và fixpoint gate vẫn ràng buộc nguyên vẹn. Vẫn báo cáo trung thực khi RED.

## Thứ tự thi hành đã chốt (bắt đầu ngay phiên 2026-07-29)
1. 🔴 Bug allocator XMM (hai vreg float sống đồng thời dùng chung register) — [[m6-perf-baseline]] "OPEN LEAD".
2. Re-land copy-propagation (đã build + đo −4.4% fib, đang bị #1 chặn).
3. Register coalescing George–Appel iterated có precolored → đóng M6-codegen.
4. RFC namespacing + linker collision diagnostic (D3).
5. Viết lại M4 suite (D2), làm dần.
