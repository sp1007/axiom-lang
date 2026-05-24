# Lộ Trình Phát Triển Ngôn Ngữ AXIOM — Kế Hoạch Tiếp Theo (next-step-1)

Tài liệu này ghi lại các hướng đi kỹ thuật tiếp theo để phát triển và hoàn thiện hệ sinh thái ngôn ngữ lập trình AXIOM, tập trung vào việc tự chủ hóa toàn diện trình biên dịch, trình liên kết (linker) và thư viện chuẩn (stdlib).

---

## 🧭 Các Hướng Đi Chính

### 🚀 Hướng 1: Hỗ Trợ Tham Chiếu Thế Hệ Dưới Dạng Mã Máy (x86_64 Native Generational Reference Support)
* **Mục tiêu**: Đảm bảo an toàn bộ nhớ theo thời gian (temporal safety) trực tiếp ở tầng mã máy x86_64 nhằm phát hiện lỗi Use-After-Free một cách tất định (deterministic) mà không cần Garbage Collector (GC).
* **Nhiệm vụ cụ thể**:
  - Định nghĩa khuôn mẫu sinh mã (machine code emitter patterns) cho `OP_MAKE_REF` và `OP_DEREF` trong bộ chọn lệnh mã máy `x86_selector.ax`.
  - Cấu hình hỗ trợ dịch chuyển offset `-8` để truy cập trực tiếp trường kiểm tra `gen_id` của cấu trúc `AxHeader` nằm ngay trước vùng nhớ được cấp phát trong `x86_encoding.ax`.
  - Viết và chạy các bài kiểm thử cưỡng bức (torture tests) để đảm bảo khi truy cập vào một tham chiếu đã bị giải phóng, chương trình lập tức kích hoạt lỗi generation mismatch panic một cách an toàn.

### 🚀 Hướng 2: Tự Chủ Hóa Hệ Thống Liên Kết & Tương Tác OS Kernel (Stage A: Linker Autonomy & Syscalls)
* **Nhiệm vụ cụ thể**:
  - **Tối ưu hóa ELF64 Linker trên Linux**: Hỗ trợ sinh Procedure Linkage Table (PLT) và Global Offset Table (GOT) cho các thư viện `.so` trên môi trường Linux để linker tự chủ hoàn toàn xuyên nền tảng (cross-platform).
  - **Sinh mã máy `syscall` trực tiếp**: Cấu hình bộ chọn lệnh `x86_selector.ax` hạ cấp các cuộc gọi hệ thống trực tiếp thông qua lệnh CPU `syscall` thay vì gọi FFI qua C Standard Library (`libc`), giúp loại bỏ hoàn toàn sự phụ thuộc vào hệ thống DLL/SO ngoại vi.

### 🚀 Hướng 3: Tái Cấu Trúc Thư Viện Chuẩn Tự Chủ 100% (Stage B: 100% Autonomous Standard Library - `std/`)
* **Nhiệm vụ cụ thể**:
  - Viết lại `std/io.ax`, `std/os.ax` và `std/process.ax` để chúng sử dụng trực tiếp các hàm `syscall` hệ điều hành mã máy và bộ quản lý bộ nhớ per-actor heap phân mảnh mới của AXIOM (`std/mem/alloc.ax`).
  - Cắt đứt hoàn toàn sự phụ thuộc vào UCRT (`ucrtbase.dll`) trên Windows và `libc.so.6` trên Linux cho các tác vụ I/O của thư viện chuẩn.

### 🚀 Hướng 4: Tối Ưu Hóa SSA Pipeline Nâng Cao (Stage C: Optimization Passes)
* **Nhiệm vụ cụ thể**:
  - Tiếp tục phát triển các pass tối ưu hóa SSA nâng cao như di chuyển mã bất biến trong vòng lặp (LICM - Loop-Invariant Code Motion), trải vòng lặp (Loop Unrolling) và tối giản hóa độ phức tạp của các phép toán (Strength Reduction).
