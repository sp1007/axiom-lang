# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-10)

Tài liệu này ghi nhận chi tiết các hướng đi chiến lược tiếp theo để phát triển và hoàn thiện ngôn ngữ lập trình AXIOM, kế thừa và mở rộng từ các thành tựu của `next-step-9`.

---

## 🧭 Tổng Quan Các Nhiệm Vụ Chiến Lược (Strategic Tasks)

### 🚀 Nhiệm Vụ 1: Native Backend Completion — Floating Point & Float Comparison Support
* **Mục tiêu**: Hoàn thiện pha phát sinh mã máy trực tiếp cho kiến trúc x86-64 (Phase 6) để hỗ trợ đầy đủ các phép toán dấu phẩy động (Floating Point), loại bỏ hạn chế chỉ chạy được số thực trên C-Backend.
* **Chi tiết kỹ thuật**:
  * Triển khai bộ chọn chỉ thị (Instruction Selector) cho các phép toán số thực (`f32`, `f64`) sử dụng tập lệnh SSE2/AVX (ví dụ: `ADDSS`, `SUBSS`, `MULSS`, `DIVSS` cho đơn chính xác; `ADDSD`, `SUBSD`, `MULSD`, `DIVSD` cho song chính xác).
  * Hỗ trợ các chỉ thị so sánh số thực (`COMISS`, `COMISD`) và ánh xạ sang các cờ điều kiện.
  * Hỗ trợ thanh ghi số thực (thanh ghi `XMM0` đến `XMM15`) trong bộ phân bổ thanh ghi (Register Allocator).

### 🚀 Nhiệm Vụ 2: Direct Native Self-Hosting Bootstrap — Direct ELF64/COFF Emission
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào trình biên dịch C (GCC/MSVC) ở Stage 2/3 bằng cách cho phép trình biên dịch AXIOM tự trị dịch chính nó trực tiếp ra mã máy nhị phân.
* **Chi tiết kỹ thuật**:
  * Cải tiến driver của trình biên dịch tự trị (`main_air.ax`) để khi chạy lệnh `build`, hệ thống sẽ sử dụng trực tiếp Native Backend sinh mã PE COFF/ELF64 và tự liên kết bằng `linker.ax` thay vì transpile qua C.

### 🚀 Nhiệm Vụ 3: Freestanding Standard Library — Full OS Autonomy
* **Mục tiêu**: Hoàn thiện các module còn lại của thư viện chuẩn để đảm bảo tính độc lập hoàn toàn (Zero libc).
* **Chi tiết kỹ thuật**:
  * Hoàn thiện `std/net.ax` sử dụng các API Socket hệ điều hành trực tiếp thông qua FFI thô (Winsock trên Windows, syscall trên Linux) để hỗ trợ truyền thông mạng.
  * Hoàn thiện `std/process.ax` cho việc quản lý tiến trình con độc lập.

---

## 🛠️ Kế Hoạch Thực Hiện: Ưu Tiên Nhiệm Vụ 1 (Native Floating Point Support)

Chúng ta sẽ tiến hành Nhiệm vụ 1 trước: Triển khai phát sinh mã máy cho các phép toán và so sánh số thực (`f32`, `f64`) trực tiếp trên x86-64 Native Backend của AXIOM.

### 📋 Checklist các bước cần thực hiện:
- [x] **Bước 1**: Cập nhật `selector.go` để bổ sung logic `air.OpCall` hỗ trợ truyền đối số số thực vào thanh ghi XMM và nhận giá trị trả về số thực từ `XMM0`.
- [x] **Bước 2**: Cập nhật `selector.go` để định nghĩa helper `selectComparison` hỗ trợ so sánh số thực thông qua `MachFCmp` và ánh xạ chính xác các cờ điều kiện.
- [x] **Bước 3**: Cập nhật `emitter.go` để phát sinh mã byte nhị phân SSE2 tương ứng cho các chỉ thị thực (`MachFAdd`, `MachFSub`, `MachFMul`, `MachFDiv`, `MachFCmp`, `MachItof`, `MachFtoi`, `MachMovDQ`, `MachMovQD`).
- [x] **Bước 4**: Cập nhật `asm_emitter.go` để hỗ trợ xuất text assembly chính xác cho các chỉ thị SSE2.
- [x] **Bước 5**: Cập nhật `backend.go` để chạy hai lượt phân bổ thanh ghi độc lập: GPR cho các thanh ghi số nguyên/con trỏ và XMM cho thanh ghi số thực.
- [x] **Bước 6**: Cập nhật kiểm thử `tests/codegen/native_diff_test.go` kích hoạt `TestDiffFloat` và chạy kiểm thử chéo để đảm bảo tính đúng đắn tuyệt đối.
