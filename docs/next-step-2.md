# Lộ Trình Phát Triển Ngôn Ngữ AXIOM — Kế Hoạch Tiếp Theo (next-step-2)

Tài liệu này ghi lại kết quả đánh giá toàn diện trạng thái kỹ thuật hiện tại của hệ sinh thái ngôn ngữ lập trình AXIOM và xác định kế hoạch chiến lược cho phân đoạn tiếp theo nhằm hoàn thiện sự tự chủ, tối ưu hóa runtime và công cụ hóa hệ sinh thái.

---

## 🧭 1. Đánh Giá Hiện Trạng Hệ Thống (Subsystem Audit)

Sau khi hoàn thành và tích hợp các pha tối ưu hóa SSA nâng cao cùng đường ống an toàn bộ nhớ (Memory Safety Pipeline), hệ thống tự dịch Stage 1 đã đạt trạng thái hoạt động tự trị hoàn hảo:

| Thành phần Subsystem | Trạng thái hiện tại | Đánh giá & Khả năng hỗ trợ |
| :--- | :--- | :--- |
| **Frontend & Parser** | **HOÀN THÀNH 100%** | Hỗ trợ phân tích cú pháp đầy đủ cho cấu trúc nâng cao: Struct, Type Alias (Sum Types), và biểu thức khớp mẫu `match`. |
| **Type Checker & Sema** | **HOÀN THÀNH 100%** | Suy luận kiểu mạnh mẽ, giải quyết triệt để chuyên biệt hóa Generic lồng nhau (Nested Monomorphization) và kiểm tra lỗi tương thích. |
| **Middle-end SSA (AIR)** | **HOÀN THÀNH 100%** | Bộ sinh AIR SSA hoàn thiện, làm phẳng các luồng điều khiển phức tạp và biểu thức khớp mẫu discriminant-based thành cấu trúc tuần tự. |
| **SSA Optimizer** | **HOÀN THÀNH 100%** | Tích hợp đầy đủ các tối ưu hóa O1/O2 cao cấp: Constant Folding, Copy Propagation, **LICM**, **Loop Unrolling**, **Strength Reduction**, và **DCE**. |
| **Đường ống CTGC & Safety** | **HOÀN THÀNH 100% (Đã kích hoạt)** | Đã kích hoạt hoàn toàn `OwnershipChecker`, `EscapeAnalyser` (phân tích thoát Heap->Stack), `CtgcInjector` (tiêm destroy tự động), và `AliasReuseOptimizer` vào quy trình biên dịch Stage 1. |
| **Custom Linker (`linker.ax`)**| **HOÀN THÀNH 100%** | Tự liên kết chéo hệ điều hành: hỗ trợ sinh bảng import `.idata` (PE/COFF trên Windows) và bảng procedure PLT/GOT (ELF64 trên Linux). |
| **Quy trình Tự dịch Parity** | **HOÀN THÀNH 100%** | Chạy thành công quy trình xác thực determinism qua `scripts/triple_build.ps1` với kết quả trùng khớp byte-for-byte chéo giữa các thế hệ Stage 1. |

---

## 🧭 2. Hướng Đi Tiếp Theo (Roadmap Milestones)

Để tiến tới việc tự chủ hoàn toàn 100% độc lập với môi trường C bên ngoài và mở rộng khả năng xử lý của ngôn ngữ sang cấp độ công nghiệp, chúng tôi xác định 3 hướng đi kỹ thuật then chốt:

### 🚀 Hướng 1: Tự Chủ Hóa Runtime Thư Viện Chuẩn (Pure-AXIOM Runtime & Syscalls)
* **Mục tiêu**: Thay thế toàn bộ hệ thống Runtime viết bằng C hiện tại (`runtime/`) bằng mã nguồn viết 100% bằng AXIOM thuần trong thư viện chuẩn `std/`.
* **Nhiệm vụ cụ thể**:
  - **Cấp phát bộ nhớ Arena (`std/mem/alloc.ax`)**: Chuyển đổi hoàn toàn cấu trúc quản lý heap Bump Allocation và Generational Reference Temporal Safety từ C sang AXIOM sử dụng các khối `unsafe` giao tiếp trực tiếp OS kernel qua syscalls (`mmap` trên Linux, `VirtualAlloc` trên Windows).
  - **Bộ điều phối thích ứng M:N (`std/scheduler.ax`)**: Viết lại Work-Stealing Actor Scheduler trực tiếp bằng AXIOM, sử dụng các nguyên tử CAS (`std/sync.ax`) sẵn có để quản lý lock-free ring-buffer của các luồng điều phối công việc thích ứng.
  - **Reactor I/O Không đồng bộ (`std/reactor.ax`)**: Tích hợp các syscalls ghép kênh I/O (`epoll` cho Linux, `IOCP` cho Windows, `kqueue` cho macOS) vào hệ thống I/O reactor của AXIOM nhằm loại bỏ hoàn toàn thư viện liên kết FFI C.

### 🚀 Hướng 2: Hoàn Thiện Thư Viện Chuẩn Tự Trị (`std/collections` & `std/io`)
* **Mục tiêu**: Cung cấp các cấu trúc dữ liệu và API tương tác hệ thống cấp công nghiệp viết bằng AXIOM thuần, độc lập hoàn toàn với UCRT (Windows) và Glibc (Linux).
* **Nhiệm vụ cụ thể**:
  - **Bộ sưu tập hiệu năng cao (`std/collections.ax`)**: Xây dựng generic `Vector[T]`, sharded `HashMap[K, V]` sử dụng giải thuật Robin Hood hashing và open-addressing tối ưu cache line CPU.
  - **Hermetic I/O Stream (`std/io.ax`)**: Viết lại toàn bộ hệ thống file streams, standard inputs/outputs và socket networking để chúng trực tiếp hạ cấp thành các syscalls `sys_read`, `sys_write`, `sys_socket`, `sys_connect` thay vì gọi FFI qua `fopen` / `printf` / `system`.

### 🚀 Hướng 3: Công Cụ Hóa Hệ Sinh Thái AXIOM (Ecosystem Tooling & Developer Experience)
* **Mục tiêu**: Tận dụng tối đa sức mạnh phân tích cú pháp tự chủ của trình biên dịch Stage 1 để sinh ra các công cụ nâng cao trải nghiệm lập trình.
* **Nhiệm vụ cụ thể**:
  - **Trình định dạng mã nguồn tự trị (`axc fmt`)**: Xây dựng bộ định dạng mã nguồn nhanh, idempotent (`fmt(fmt(src)) == fmt(src)`) dựa trên cây cú pháp của `parser.ax` để chuẩn hóa toàn diện phong cách lập trình AXIOM.
  - **Trình phục vụ ngôn ngữ LSP (`axc lsp`)**: Cấu hình một máy chủ LSP 3.17 trực tiếp trong trình điều khiển biên dịch, tiêu thụ bảng ký hiệu và type table của `resolver.ax`/`typecheck.ax` nhằm cung cấp khả năng tự động hoàn thành, nhảy tới định nghĩa (Go to Definition) và hovers báo kiểu dữ liệu trực tiếp trên các trình soạn thảo (VS Code, Neovim).
