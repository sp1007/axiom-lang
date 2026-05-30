# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-6)

Tài liệu này ghi lại chi tiết các hướng đi chiến lược tiếp theo để đưa ngôn ngữ lập trình AXIOM đạt đến cấp độ hoàn thiện hoàn mỹ của một hệ thống biên dịch freestanding tự trị và an toàn bộ nhớ tĩnh.

---

## 🧭 Tổng Quan Các Hướng Đi Chiến Lược (Strategic Directions)

### 🚀 Hướng 1: Native Self-Hosting Integration (Tích hợp luồng Tự trị Mã máy Gốc)
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào trình biên dịch GCC làm bộ sinh mã máy trung gian C (C-transpiler) trong các phiên bản tự trị. Tích hợp trực tiếp luồng sinh mã máy gốc (x86_64 ELF/COFF) và tự liên kết làm luồng mặc định cho Stage 1 và Stage 2.
* **Nhiệm vụ cụ thể**:
  - Tích hợp các hàm `compile_native_binary` (`x86_coff.ax`) và `axiom_linker_link` (`linker.ax`) làm đường ống biên dịch mặc định trong trình điều khiển tự trị của AXIOM (`main_air.ax`) khi nhận tùy chọn `build`.
  - Loại bỏ hoàn toàn việc gọi `system("gcc ...")` ngoại vi để biên dịch các file `.obj` tạm thời, chuyển sang cơ chế tự liên kết nhị phân thông qua thư viện linker gốc viết bằng AXIOM (`linker.ax`).
  - Đảm bảo các tệp thực thi nhị phân do trình biên dịch tự trị tự tạo ra có thể khởi chạy và hoạt động trơn tru (ví dụ: `hello_native_s2.exe`).

### 🚀 Hướng 2: OS Call Autonomy & Freestanding Standard Library (Tách rời I/O và OS Khỏi C-FFI)
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào UCRT (`ucrtbase.dll`) trên Windows và `libc.so.6` trên Linux cho các tác vụ I/O cơ bản. Giúp AXIOM sinh ra các tệp thực thi freestanding cực nhẹ thông qua Direct OS API hoặc Syscalls.
* **Nhiệm vụ cụ thể**:
  - **Trên Windows**: Viết lại `std/io.ax` để gọi trực tiếp các API Win32 trong `kernel32.dll` (`CreateFileA`, `ReadFile`, `WriteFile`, `CloseHandle`) thay thế cho `fopen`/`fread`/`fwrite`/`fclose`.
  - **Trên Linux**: Sử dụng trực tiếp chỉ thị `syscall` gốc trong `std/io.ax` (Syscall 0 cho `sys_read`, Syscall 1 cho `sys_write`, Syscall 60 cho `sys_exit`).
  - **Trên Windows CLI**: Viết lại `std/os.ax` để truy xuất biến môi trường và đối số dòng lệnh trực tiếp bằng cách đọc cấu trúc **Process Environment Block (PEB)** hoặc thông qua `GetCommandLineW` để đơn giản hóa xử lý, tránh phụ thuộc vào C-runtime FFI.

### 🚀 Hướng 3: Memory Safety Pipeline & Advanced Collections (An toàn Bộ nhớ Tĩnh CTGC)
* **Mục tiêu**: Tích hợp toàn diện các giai đoạn phân tích đồ thị sở hữu bộ nhớ để tự động tiêm giải phóng bộ nhớ tĩnh tại thời điểm biên dịch (Compile-Time Garbage Collection - CTGC), loại bỏ triệt để nguy cơ rò rỉ bộ nhớ hoặc lỗi truy cập con trỏ hoang mà không cần GC runtime.
* **Nhiệm vụ cụ thể**:
  - Tích hợp và kích hoạt toàn diện các module `ownership.ax` (kiểm soát đơn sở hữu), `escape.ax` (phân tích thoát biến), `ctgc.ax` (tự động tiêm destructors `=destroy`), và `alias_reuse.ax` (tối ưu hóa tái sử dụng ô nhớ liền kề) trong luồng biên dịch chính.
  - Triển khai các cấu trúc dữ liệu Generic cốt lõi (`Vector[T]` và `HashMap[K, V]` sử dụng giải thuật Robin Hood Hashing) viết 100% bằng AXIOM thuần, hỗ trợ quản lý bộ nhớ tự động CTGC mà không rò rỉ.
  - Sử dụng AddressSanitizer/Valgrind để stress-test trình biên dịch tự trị, đảm bảo rò rỉ bộ nhớ bằng 0 (Zero Memory Leaks).

---

## 🛠️ Kế Hoạch Thực Hiện: Ưu Tiên Hướng 1 (Native Self-Hosting Integration)

Chúng ta sẽ thực hiện Hướng 1 trước để đóng gói luồng biên dịch mã nhị phân máy gốc hoàn toàn tự trị mà không cần GCC.

### 1. Phân Tích & Xác Định Vấn Đề Hiện Tại
Khi chúng ta chạy Stage 2 tự trị để sinh nhị phân gốc:
```powershell
.\bin\axc_stage2_selfhosted.exe build .\tests\sema\valid_hello.ax -o .\bin\hello_native_s2.exe
```
Quá trình biên dịch mã máy gốc và tự liên kết (Self-Linking) thông qua `linker.ax` chạy thành công rực rỡ và tạo ra `hello_native_s2.exe`. Tuy nhiên, khi thực thi:
- Chương trình bị thoát ngay lập tức với mã lỗi `1` và không in ra màn hình.
- Nguyên nhân: `ax_runtime.dll` chứa các hàm runtime như `ax_println_str`. Trình liên kết tự trị `linker.ax` phân giải các ký hiệu ngoại vi và liên kết động chúng sang `ax_runtime.dll`. Khi thực thi, nếu các tên ký hiệu (symbol names) bị lệch hoặc không khớp định dạng mong đợi giữa runtime DLL và code gốc, Windows Loader sẽ thoát chương trình ngay lập tức với mã lỗi `1`.

### 2. Các Bước Triển Khai Kỹ Thuật
1. **Kiểm tra và Chuẩn hóa Phân giải Ký hiệu Runtime DLL**:
   - Rà soát hàm `get_dll_for_symbol` và `is_valid_runtime_dll_symbol` trong `bootstrap/stage1/linker.ax`.
   - Đảm bảo các ký hiệu runtime nội bộ (như `ax_println_str`, `ax_println_i64`, `ax_free`, `ax_alloc`) được ánh xạ chính xác sang `"ax_runtime.dll"` hoặc `"none"` một cách tất định.
2. **Sửa đổi luồng điều khiển mặc định trong `main_air.ax`**:
   - Bật tùy chọn `self_link` làm luồng mặc định khi biên dịch nhị phân gốc để đảm bảo AXIOM tự liên kết bằng `linker.ax` thay vì gọi GCC hệ thống khi chạy lệnh `build`.
3. **Kiểm tra và Khắc phục sự tương thích của Hello World nhị phân gốc**:
   - Đảm bảo chương trình `hello_native_s2.exe` thực thi hoàn hảo và in ra kết quả `Hello, world!` chính xác 30 lần mà không lỗi.
