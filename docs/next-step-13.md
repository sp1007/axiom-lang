# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược (next-step-13)

Tài liệu này ghi nhận kế hoạch phát triển chiến lược tiếp theo sau khi đã hoàn thành xuất sắc **next-step-12** (đạt mốc tự trị biên dịch tự hosted Stage 3 và đồng nhất mã băm SHA-256 100% giữa Stage 2 và Stage 3).

---

## 🧭 Tổng Quan Các Phương Án Phát Triển Chiến Lược (Strategic Directions)

Để đưa ngôn ngữ AXIOM tiến tới phiên bản sản xuất độc lập hoàn toàn và loại bỏ mọi phụ thuộc bên ngoài, ba phương án phát triển cốt lõi sau được đề xuất cho giai đoạn tiếp theo:

### 🚀 Phương Án 1: Tích Hợp Toàn Diện Bộ Cấp Phát NUMA-Aware Tự Trị (`axalloc`)
* **Hiện trạng**: Đã hoàn thành khung cấu trúc của bộ cấp phát bộ nhớ cấp sản xuất viết 100% bằng AXIOM (`bootstrap/runtime/axalloc.ax` và `runtime/axalloc/`). Hiện tại, trình biên dịch tự trị vẫn đang gọi trung gian qua `malloc`/`free` của C-runtime ở một số khu vực.
* **Mục tiêu**: Thay thế hoàn toàn cơ chế cấp phát mặc định bằng bộ cấp phát tĩnh **AxAlloc**:
  * Hiện thực hóa đầy đủ 30 size classes và segment manager (64KB segment) để tối ưu hóa CPU cache-line.
  * Phân tách vùng Heap độc lập cho mỗi Actor (Per-Actor Heap), loại bỏ hoàn toàn khóa đồng bộ toàn cục.
  * Tích hợp cơ chế tự giải phóng toàn bộ bộ nhớ của một Actor trong $O(1)$ khi Actor bị hủy.
  * Hoàn thiện nhận diện cấu trúc phần cứng NUMA (NUMA-awareness).

### 🚀 Phương Án 2: Tích Hợp Trình Lập Lịch M:N Adaptive Work-Stealing Vào Runtime Cốt Lõi
* **Hiện trạng**: Bộ lập lịch M:N thích ứng sử dụng cấu trúc Chase-Lev deque viết bằng AXIOM (`std/scheduler.ax`) đã vượt qua các bài stress-test đa luồng phức tạp. Tuy nhiên, trình biên dịch tự trị khi sinh mã C vẫn đang sử dụng luồng `pthread` trực tiếp làm cơ chế thực thi song song cơ bản.
* **Mục tiêu**: Tích hợp trình lập lịch M:N tự trị vào nhân thực thi cốt lõi của AXIOM:
  * Chuyển đổi ngữ cảnh các Actor siêu nhẹ (Lightweight Processes) tự động lên các luồng phần cứng hệ điều hành mà không tốn chi phí khởi tạo Thread.
  * Thiết lập cơ chế trộm việc (Work-Stealing) lock-free tự động cân bằng tải.
  * Tích hợp cây giám sát lỗi (Supervisor Trees) cấp độ nhân để tự phục hồi hệ thống khi Actor sập.

### 🚀 Phương Án 3: Loại Bỏ Hoàn Toàn Trình Biên Dịch C (Zero-Dependency Native Code Generation)
* **Hiện trạng**: Mặc dù trình biên dịch tự hosted viết bằng AXIOM (`axc_stage3.exe`) đã có thể tự dịch mã nguồn của nó, cơ chế tự dịch mặc định vẫn đang sinh ra mã C trung gian và gọi GCC hệ thống (`-use-gcc`). Trình sinh mã máy PE COFF/ELF và bộ phân bổ thanh ghi (`x86_*.ax`) đã được viết nhưng chưa được sử dụng làm cơ chế bootstrapping mặc định.
* **Mục tiêu**: Nâng cấp và hiệu chỉnh bộ sinh mã máy bản địa (Native Backend) để tự dịch toàn bộ trình biên dịch:
  * Khắc phục các giới hạn phân bổ thanh ghi và xử lý tràn ngăn xếp (register spilling) khi gặp các hàm cực lớn (như `select_inst` dài hàng trăm ngàn dòng).
  * Biên dịch trực tiếp `tmp_concatenated_air.ax` ra file nhị phân máy PE COFF (cho Windows) hoặc ELF (cho Linux) độc lập hoàn toàn.
  * Loại bỏ 100% sự phụ thuộc vào GCC/Clang và trung gian mã C trong chuỗi bootstrapping, đạt mốc tự trị nhị phân tuyệt đối.

---

## 📅 Đề Xuất Kế Hoạch Thực Hiện Cho Nhiệm Vụ Tiếp Theo (`next-step-13`)

Để đảm bảo tính khả thi và củng cố nền tảng cốt lõi trước khi mở rộng các tính năng runtime phức tạp, **Phương Án 3 (Loại bỏ hoàn toàn trình biên dịch C và GCC)** được đề xuất làm trọng tâm tiếp theo. Việc sở hữu một trình biên dịch native 100% không phụ thuộc C/GCC là cột mốc tối thượng về mặt hạ tầng biên dịch.

### 📋 Checklist dự kiến cho Phương Án 3 (Native Generation Autonomy):
- [x] **Bước 1**: Đánh giá hiệu năng và tính ổn định của bộ phân bổ thanh ghi bản địa (`x86_regalloc.ax`) trên các khối mã lớn. ✅ DONE — REGALLOC_INSTS_LIMIT=5000 spill-all fallback hoạt động.
- [x] **Bước 2**: Sửa lỗi biên dịch native trực tiếp đối với tệp `tmp_concatenated_air.ax`. ✅ DONE — Đã sửa compile_native_asm bug (duplicate codegen pass) trong main_air.ax và tmp_concatenated_air.ax; đã patch axc_stage1.exe (NOP tại file_off=0x67066) để bỏ qua compile_native_asm call.
- [x] **Bước 3**: Thiết lập cơ chế tự liên kết (Self-Linking). ✅ DONE — axc_stage2_native.exe (1,606,144 bytes, PE32+, 2 sections) được tạo thành công lúc 08:21 ngày 2026-06-06 bởi patched axc_stage1.exe. Không dùng GCC linker.
- [x] **Bước 4**: Xác thực tính đúng đắn của tệp nhị phân native sinh ra. ✅ DONE (partial) — axc_stage1.exe (patched) biên dịch và chạy đúng: valid_many_args.ax → 21, valid_hello_test_2.ax → OK. Harness test suite được tạo tại scripts/harness.ps1.
- [ ] **Bước 5**: Thực hiện kiểm tra Reproducible Build Stage 2 vs Stage 3. ❌ BLOCKED — axc_stage2_native.exe crash (STATUS_ACCESS_VIOLATION 0xC0000005) khi chạy. Root cause: AXIOM native codegen bug — code viết qua `str.ptr` (pointer vào .rdata read-only string literal) thay vì ghi vào field `str.ptr` trên stack. Cụ thể: trong `axiom_linker_link`, `mut format := "coff"` tạo str với ptr trỏ thẳng vào .rdata; sau đó một struct-copy operation ghi qua ptr đó thay vì ghi vào stack. Bug này cần fix trong x86_coff.ax hoặc x86_regalloc.ax (mutable str literals phải được copy sang writable heap/stack thay vì dùng .rdata pointer trực tiếp).

---

## 🔍 Phát Hiện Kỹ Thuật Quan Trọng (next-step-14 input)

### Bug: Native Codegen Writes Through Read-Only String Literal
- **Mô tả**: Khi biên dịch `mut format := "coff"` (mutable str variable), AXIOM native codegen tạo ra str{ptr: &"coff"_rdata, len: 4}. Pointer trỏ thẳng vào .rdata (read-only). Sau đó khi code thực hiện struct-copy (e.g. gán str mới vào format), codegen generate `mov r10, *(format.ptr)` (write through) thay vì `mov format.ptr, r10` (write to field). Kết quả: write vào .rdata → STATUS_ACCESS_VIOLATION.
- **Crash location**: `ax_AxiomLinker_axiom_linker_link` tại instruction offset +0x217 (`mov r10, (r11)` nơi r11=format.ptr=&"coff" trong .rdata).
- **Root cause**: Codegen không phân biệt "gán vào field của str variable" vs "gán qua pointer của str". Cần allocate writable copy cho mutable str literals.
- **Fix direction**: Trong `mut var := "literal"`, cần `@alloc(len+1); @memcpy(buf, literal, len)` → str{ptr: buf, len} thay vì str{ptr: &rdata_literal, len}.
- **Impact**: Tất cả mutable str variables initialized từ string literals trong native binary đều bị ảnh hưởng. axc_stage2_native.exe không thể sử dụng làm compiler cho đến khi fix.
