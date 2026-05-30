# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-9)

Tài liệu này ghi nhận chi tiết các hướng đi chiến lược tiếp theo để phát triển ngôn ngữ lập trình AXIOM, kế thừa và mở rộng từ các thành tựu của `next-step-8`.

---

## 🧭 Tổng Quan Các Nhiệm Vụ Chiến Lược (Strategic Tasks)

### 🚀 Nhiệm Vụ 1: Optimization Pipeline — SSA Constant Folding & Dead Code Elimination (DCE)
* **Mục tiêu**: Tối ưu hóa hiệu năng mã máy thông qua việc tối giản hóa đồ thị SSA IR ở Phase 6.
* **Chi tiết kỹ thuật**:
  - **Constant Folding (Gập hằng số)**: Nhận diện và thực hiện trước các phép toán tĩnh trên các kiểu số nguyên và logic (ví dụ: `2 + 3` thành `5`, `true and false` thành `false`) trực tiếp trên cấu trúc SSA IR.
  - **Dead Code Elimination (Loại bỏ mã chết)**: Phân tích đồ thị sử dụng biến (Def-Use Chains). Loại bỏ các chỉ thị SSA không bao giờ được sử dụng hoặc các khối cơ bản (Basic Blocks) không có đường truyền tới (unreachable blocks), thu gọn dung lượng nhị phân và cải thiện tốc độ CPU.

### 🚀 Nhiệm Vụ 2: Package Manager — Multi-File Compilation & Native Module Resolution
* **Mục tiêu**: Nâng cấp trình biên dịch từ dạng biên dịch đơn tệp (Single-file) lên Phase 7 hỗ trợ nhiều tệp nguồn và tự động phân giải import.
* **Chi tiết kỹ thuật**:
  - Cải tiến Driver trình biên dịch tự trị (`main_air.ax`) để hỗ trợ nhận danh sách nhiều tệp đầu vào, tự động phân giải cấu trúc cây thư mục.
  - Thiết lập cơ chế nạp module lười biếng (Lazy Module Loading) trực tiếp trong phiên bản tự trị: khi gặp câu lệnh `import std.string` hay `import std.collections`, trình biên dịch tự động tìm kiếm, phân tích và liên kết tệp `.ax` tương ứng mà không cần kịch bản PowerShell ghép nối tệp thủ công.

### 🚀 Nhiệm Vụ 3: OS Call Autonomy — Freestanding I/O Standard Library (Zero libc)
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào các thư viện liên kết động C Runtime (`ucrtbase.dll` trên Windows và `libc.so` trên Linux) cho các tác vụ I/O cơ bản.
* **Chi tiết kỹ thuật**:
  - **Trên Windows**: Viết lại module `std/io.ax` để gọi trực tiếp các API Win32 thô từ `kernel32.dll` (`CreateFileA`, `ReadFile`, `WriteFile`, `CloseHandle`) thay thế cho `fopen`/`fread`/`fwrite`/`fclose`.
  - **Trên Linux**: Sử dụng trực tiếp chỉ thị `syscall` gốc trong `std/io.ax` (Syscall 0 cho `sys_read`, Syscall 1 cho `sys_write`, Syscall 3 cho `sys_close`, Syscall 60 cho `sys_exit`).

---

## 🛠️ Kế Hoạch Thực Hiện: Ưu Tiên Nhiệm Vụ 1 (SSA Optimization Pipeline)

Chúng ta sẽ tiến hành Nhiệm vụ 1 trước: Triển khai Constant Folding và Dead Code Elimination (DCE) trực tiếp trên trình tối ưu hóa SSA của AXIOM.
