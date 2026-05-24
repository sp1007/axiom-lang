# Lộ Trình Phát Triển Ngôn Ngữ AXIOM — Kế Hoạch Tiếp Theo (next-step-3)

Tài liệu này ghi lại các hướng đi chiến lược tiếp theo để đạt được sự tự chủ hoàn toàn của hệ sinh thái ngôn ngữ lập trình AXIOM. Các hạng mục này sẽ được thực hiện tuần tự để nâng cấp trình biên dịch tự dịch Stage 1 và thư viện chuẩn (`std/`).

---

## 🧭 Các Lựa Chọn Chiến Lược (Roadmap Choices)

### 🚀 Lựa Chọn 1: Công Cụ Hóa Hệ Sinh Thái — Trình Định Dạng Tự Trị (`axc fmt` viết bằng AXIOM)
* **Mục tiêu**: Xây dựng bộ định dạng mã nguồn nhanh, idempotent (`fmt(fmt(src)) == fmt(src)`) hoàn toàn độc lập được viết 100% bằng AXIOM thuần ( Stage 1).
* **Nhiệm vụ cụ thể**:
  - Chuyển đổi giải thuật định dạng AXIOM từ Go (`tools/fmt`) sang AXIOM thuần trong [bootstrap/stage1/fmt.ax](file:///d:/projects/compiler/Axiom/bootstrap/stage1/fmt.ax).
  - Sử dụng Token Stream và cây AST phân giải trực tiếp từ `lexer.ax` và `parser.ax` để tái cấu trúc văn bản nguồn canonically.
  - Tích hợp lệnh `fmt` vào trình điều khiển tự trị của Stage 1 (`main_air.ax`).

### 🚀 Lựa Chọn 2: Tự Chủ Hóa Runtime — Bộ Cấp Phát Bộ Nhớ Arena Syscall-based (`std/mem/alloc.ax`)
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào FFI C (`malloc`, `free`, `realloc`) cho hệ thống Heap Allocator của AXIOM bằng cách giao tiếp trực tiếp với OS Kernel.
* **Nhiệm vụ cụ thể**:
  - Triển khai chỉ thị gọi hệ thống trực tiếp (Direct OS Syscalls) trong AXIOM: sử dụng các lệnh Assembly `syscall` qua bộ chọn lệnh x86 (`x86_selector.ax`).
  - Viết lại hàm xin cấp phát phân mảnh và thu hồi bộ nhớ lớn trực tiếp qua `mmap` / `munmap` (Linux/macOS) và `VirtualAlloc` / `VirtualFree` (Windows).
  - Tối ưu hóa bộ quản lý bộ nhớ Generational Temporal Safety để chống rò rỉ và phát hiện Use-After-Free độc lập với UCRT/Glibc.

### 🚀 Lựa Chọn 3: Bộ Điều Phối M:N Adaptative & Reactor Không Đồng Bộ (`std/scheduler.ax` & `std/reactor.ax`)
* **Mục tiêu**: Xây dựng lõi runtime xử lý concurrency hiệu năng cao dạng Lock-Free Work-Stealing viết 100% bằng AXIOM.
* **Nhiệm vụ cụ thể**:
  - Tối ưu hóa Adaptive Scheduler sử dụng các CAS Atomics (`std/sync.ax`) sẵn có của AXIOM.
  - Tích hợp ghép kênh I/O không đồng bộ trực tiếp qua Syscall của OS: Epoll trên Linux, IOCP trên Windows, và Kqueue trên macOS.
  - Đảm bảo FFI C hoàn toàn không tồn tại trong hệ thống Async Engine của AXIOM.

---

## 🔄 Tiến Trình Thực Hiện Tuần Tự (Execution Pipeline)

Chúng tôi sẽ tiến hành thực hiện tuần tự theo thứ tự ưu tiên:
1. **Pha 1**: Triển khai trình định dạng tự trị `fmt.ax` viết bằng AXIOM thuần và tích hợp vào Stage 1.
2. **Pha 2**: Tách rời malloc/free khỏi UCRT/Glibc, thay thế bằng direct OS Kernel Syscalls trong `alloc.ax`.
3. **Pha 3**: Lock-free Scheduler & Asynchronous Reactor FFI-free.
