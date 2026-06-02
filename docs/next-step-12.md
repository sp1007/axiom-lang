# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-12)

Tài liệu này ghi nhận kế hoạch phát triển chiến lược tiếp theo để xây dựng hệ sinh thái AXIOM v0.0.2 tự trị hoàn toàn, kế thừa và mở rộng từ các thành tựu của `next-step-11`.

---

## 🧭 Tổng Quan Các Nhiệm Vụ Chiến Lược (Strategic Tasks)

Các phương án phát triển sẽ được thực hiện tuần tự:

### 🚀 Phương Án 1: Thiết Kế & Hiện Thực Bộ Quản Lý Gói AXIOM Package Manager (`tools/pkg`)
* **Hiện trạng**: Tệp [tools/pkg/pkg.go](file:///d:/projects/compiler/Axiom/tools/pkg/pkg.go) mới chỉ là một stub trống rỗng. Trình biên dịch chưa hỗ trợ khai báo, tải và liên kết các thư viện ngoài.
* **Mục tiêu**: Xây dựng hệ thống quản lý gói cơ bản:
  * Định nghĩa cấu trúc tệp cấu hình dự án (`axiom.toml` hoặc `axiom.json`).
  * Triển khai cơ chế phân giải gói phụ thuộc (Dependency Resolution) và tải gói qua Git hoặc bộ nhớ đệm cục bộ (`axc get` hoặc `axc install`).
  * Tích hợp mã băm SHA-256 (`axiom.lock`) để kiểm tra tính toàn vẹn.
* **Trạng thái**: [x] Đã hoàn thành 100%.

### 🚀 Phương Án 2: Xây Dựng Trình Lập Lịch Đa Luồng M:N Adaptive Work-Stealing (`runtime/actors`)
* **Hiện trạng**: Triển khai hoàn chỉnh bộ lập lịch M:N thích ứng Chase-Lev deque viết bằng AXIOM (`std/scheduler.ax`) cùng các Unit Tests (`std/scheduler_test.ax`) và C-bridge (`scheduler_test_debug.c`).
* **Mục tiêu**: Triển khai bộ lập lịch M:N thích ứng hoàn chỉnh viết bằng AXIOM:
  * Điều phối và chuyển đổi ngữ cảnh các Actor siêu nhẹ (Lightweight Processes) lên các luồng hệ điều hành thực tế.
  * Xây dựng hàng đợi công việc không khóa (Lock-free Work Queues) phục vụ cơ chế trộm việc (Work-Stealing).
  * Triển khai cây giám sát lỗi (Supervisor Trees) phục vụ cơ chế tự phục hồi ("Let it crash").
* **Trạng thái**: [x] Đã hoàn thành 100%.

### 🚀 Phương Án 3: Phát Triển Bộ Cấp Phát NUMA-Aware Cấp Sản Xuất (`axalloc`)
* **Hiện trạng**: Đã hoàn thành cấu trúc bộ cấp phát bộ nhớ cấp sản xuất (`bootstrap/runtime/axalloc.ax` và `runtime/axalloc/`) viết 100% bằng AXIOM, hỗ trợ per-actor independent heaps và `ax_numa_alloc` cache-line aligned.
* **Mục tiêu**: Xây dựng bộ cấp phát bộ nhớ cấp sản xuất viết 100% bằng AXIOM:
  * Phân mảnh bộ nhớ theo các lớp kích thước (8, 16, 32, 64, 128, ... bytes) để tối ưu hóa CPU cache-line.
  * Phân vùng Heap độc lập cho mỗi Actor (Per-Actor Heap), loại bỏ khóa đồng bộ toàn cục.
  * Hỗ trợ giải phóng toàn bộ bộ nhớ của một Actor trong độ phức tạp thời gian $O(1)$.
  * Tích hợp khả năng nhận biết cấu trúc phần cứng NUMA (NUMA-awareness).
* **Trạng thái**: [/] Đã hoàn thành cấu trúc.

#### 🚀 Phương Án 4: Độc Lập Chuỗi Công Cụ (Porting `tools/fmt` & `tools/lsp` sang AXIOM)
* **Hiện trạng**: Trình định dạng [tools/fmt/fmt.go](file:///d:/projects/compiler/Axiom/tools/fmt/fmt.go) và Language Server [tools/lsp/lsp.go](file:///d:/projects/compiler/Axiom/tools/lsp/lsp.go) hiện tại đang được viết bằng ngôn ngữ Go.
* **Mục tiêu**: Chuyển đổi toàn bộ chuỗi công cụ phụ trợ sang AXIOM:
  * Viết lại trình định dạng mã nguồn `axc fmt` bằng ngôn ngữ AXIOM.
  * Viết lại Language Server `axc lsp` sử dụng cây Flat AST và siêu dữ liệu `.axmeta` để tích hợp trực tiếp trên IDE.
  * Đưa dự án đạt trạng thái 100% tự trị hoàn toàn, không còn mã nguồn Go hay C trong chuỗi công cụ cốt lõi.
* **Trạng thái**: [/] Đã hoàn thành cấu trúc.

---

## 🛠️ Kế Hoạch Thực Hiện Hiện Tại: Phương Án 2 (M:N Adaptive Work-Stealing Scheduler)

Chúng ta đã hoàn thành xuất sắc Phương án 1 và Phương án 2.

### 📋 Checklist các bước cần thực hiện của Phương Án 2:
- [x] **Bước 1**: Phác thảo và xây dựng cấu trúc trạng thái của Actor (`Actor` struct) lưu trữ id, state, mailbox, per-actor heap, và reduction budget.
- [x] **Bước 2**: Triển khai cơ chế tráo đổi ngữ cảnh các coroutine (Context Switching) an toàn không khóa.
- [x] **Bước 3**: Triển khai hàng đợi công việc thích ứng (Adaptive Work-Stealing Queue) hỗ trợ cơ chế trộm công việc giữa các luồng phần cứng.
- [x] **Bước 4**: Triển khai Mailbox và cơ chế gửi nhận thông điệp an toàn bất đồng bộ (Channel & Message Passing).
- [x] **Bước 5**: Tích hợp mô hình Cây giám sát lỗi ("Let it crash" Supervisor Tree) tự động hồi phục các Actor sập.
- [x] **Bước 6**: Viết bộ kiểm thử stress-test chịu tải cao và kiểm tra độ ổn định đa luồng.
