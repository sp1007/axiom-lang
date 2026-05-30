# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-7)

Tài liệu này ghi nhận chi tiết các hướng đi chiến lược tiếp theo để phát triển ngôn ngữ lập trình AXIOM, kế thừa và mở rộng từ các thành tựu rực rỡ của `next-step-6`.

---

## 🧭 Tổng Quan Các Nhiệm Vụ Chiến Lược (Strategic Tasks)

### 🚀 Nhiệm Vụ 1: OS Call Autonomy — Freestanding CLI Args & Windows Unicode Support
* **Mục tiêu**: Chuẩn hóa việc phân giải đối số dòng lệnh trực tiếp từ Hệ điều hành (OS) mà không phụ thuộc vào `argc`/`argv` truyền thống của C runtime, hỗ trợ đầy đủ ký tự Unicode thông qua `GetCommandLineW` trên Windows.
* **Chi tiết kỹ thuật**:
  - Tích hợp giải thuật phân giải Unicode từ `GetCommandLineW` (được thử nghiệm thành công trong `scratch/test_args.ax`) trực tiếp vào tệp thư viện chuẩn `std/os.ax` dưới dạng freestanding.
  - Sử dụng bộ giải mã chuỗi rộng UTF-16 thành UTF-8 bytes thuần AXIOM, đưa `std/os.ax` đạt độ độc lập 100% khỏi C runtime.

### 🚀 Nhiệm Vụ 2: OS Call Autonomy — Freestanding I/O Standard Library
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào UCRT (`ucrtbase.dll` trên Windows) và `libc.so.6` trên Linux cho các tác vụ I/O tệp cơ bản.
* **Chi tiết kỹ thuật**:
  - **Trên Windows**: Viết lại tệp thư viện chuẩn `std/io.ax` để gọi trực tiếp các API Win32 thô từ `kernel32.dll` (`CreateFileA`, `ReadFile`, `WriteFile`, `CloseHandle`) thay thế cho `fopen`/`fread`/`fwrite`/`fclose`.
  - **Trên Linux**: Sử dụng trực tiếp chỉ thị `syscall` gốc trong `std/io.ax` (Syscall 0 cho `sys_read`, Syscall 1 cho `sys_write`, Syscall 60 cho `sys_exit`, Syscall 3 cho `sys_close`).

### 🚀 Nhiệm Vụ 3: Custom Linker — Dynamic Extern Symbol Resolution
* **Mục tiêu**: Loại bỏ hoàn toàn địa chỉ hardcoded (`0x500000`) cho các ký hiệu ngoại vi (extern symbols) trong trình liên kết tự gốc `linker.ax`.
* **Chi tiết kỹ thuật**:
  - Triển khai bảng băm định vị địa chỉ thực tế từ PE/COFF IAT (Import Address Table) và ELF PLT/GOT để phân giải động địa chỉ biểu tượng import tại thời điểm liên kết.

### 🚀 Nhiệm Vụ 4: Memory Safety — Pure Axiom Generic Collections (CTGC)
* **Mục tiêu**: Triển khai các cấu trúc dữ liệu cơ bản dạng Generic viết 100% bằng AXIOM thuần, tận dụng tối đa cơ chế tự giải phóng tĩnh CTGC.
* **Chi tiết kỹ thuật**:
  - Phát triển cấu trúc `Vector[T]` và `HashMap[K, V]` (Robin Hood Hashing) có khả năng tự dọn dẹp và giải phóng thông qua destructor `=destroy` tĩnh.
  - Loại bỏ hoàn toàn rò rỉ bộ nhớ khi chạy các chương trình phức tạp tự trị.

---

## 🛠️ Kế Hoạch Thực Hiện: Ưu Tiên Nhiệm Vụ 1 (OS Call Autonomy - CLI Args)

Chúng ta sẽ tiến hành Nhiệm vụ 1 trước: Tích hợp trình phân giải đối số dòng lệnh Unicode freestanding trực tiếp vào `std/os.ax`.
