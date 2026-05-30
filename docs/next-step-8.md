# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-8)

Tài liệu này ghi nhận chi tiết các hướng đi chiến lược tiếp theo để phát triển ngôn ngữ lập trình AXIOM, kế thừa và mở rộng từ các thành tựu rực rỡ của `next-step-7`.

---

## 🧭 Tổng Quan Các Nhiệm Vụ Chiến Lược (Strategic Tasks)

### 🚀 Nhiệm Vụ 1: Optimization Pipeline — SSA Constant Folding & Dead Code Elimination (DCE)
* **Mục tiêu**: Tối ưu hóa hiệu năng mã máy thông qua việc tối giản hóa đồ thị SSA IR ở Phase 6.
* **Chi tiết kỹ thuật**:
  - Triển khai thuật toán gập hằng số (Constant Folding) cho các kiểu dữ liệu cơ bản trực tiếp trong bộ tối ưu hóa SSA (`ssa_opt.ax` và `ir/opt/`).
  - Triển khai giải thuật loại bỏ mã chết (Dead Code Elimination - DCE) thông qua việc phân tích tầm vực sử dụng biến (Def-Use Chains), loại bỏ các khối cơ bản (Basic Blocks) không thể chạm tới hoặc các đăng ký SSA không được tiêu thụ.

### 🚀 Nhiệm Vụ 2: Package Manager — Multi-File Compilation & Native Module Resolution
* **Mục tiêu**: Nâng cấp trình biên dịch từ dạng biên dịch đơn tệp (Single-file) lên Phase 7 hỗ trợ nhiều tệp nguồn và tự động phân giải import.
* **Chi tiết kỹ thuật**:
  - Cải tiến Driver trình biên dịch (`compiler/driver/` và `driver.ax`) để hỗ trợ danh sách nhiều tệp đầu vào, tự động phân giải cấu trúc cây thư mục.
  - Thiết lập cơ chế tự phân giải module thông minh: Khi gặp câu lệnh `import std.os`, trình biên dịch sẽ tự động tìm kiếm và biên dịch tích hợp tệp `std/os.ax` mà không cần người dùng phải ghép tệp thủ công thông qua tập lệnh PowerShell.

### 🚀 Nhiệm Vụ 3: Language Tooling — Canonical Code Formatter (`axfmt`)
* **Mục tiêu**: Chuẩn hóa phong cách lập trình của hệ sinh thái AXIOM ở Phase 9.
* **Chi tiết kỹ thuật**:
  - Phát triển công cụ định dạng mã nguồn tự động `axfmt` dựa trên cây cú pháp AST của Parser, hỗ trợ căn lề chuẩn, loại bỏ khoảng trắng thừa, định dạng khối mã và biểu thức nhất quán.

---

## 🛠️ Kế Hoạch Thực Hiện: Ưu Tiên Nhiệm Vụ 2 (Multi-File Module Resolution)

Nhiệm vụ cấp bách nhất hiện nay là **Nhiệm vụ 2**: Mở rộng khả năng xử lý của Trình biên dịch để hỗ trợ biên dịch đa tệp và tự động phân giải import. Điều này sẽ giúp loại bỏ hoàn toàn các tập lệnh nối chuỗi trung gian, cho phép biên dịch trực tiếp các dự án Axiom lớn và phức tạp.
