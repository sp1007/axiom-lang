# Lộ Trình Phát Triển Ngôn Ngữ AXIOM — Kế Hoạch Tiếp Theo (next-step-4)

Tài liệu này ghi lại các hướng đi chiến lược tiếp theo để đạt được sự tự trị hoàn toàn của hệ sinh thái ngôn ngữ lập trình AXIOM. Các hạng mục này sẽ được thực hiện tuần tự để nâng cấp trình biên dịch tự dịch Stage 1 và thư viện chuẩn (`std/`).

---

## 🧭 Các Lựa Chọn Chiến Lược (Roadmap Choices)

### 🚀 Hướng 1: Biên dịch trực tiếp ra Mã máy & Liên kết Tự trị (Bỏ qua GCC Transpiler)
* **Mục tiêu**: Loại bỏ hoàn toàn sự phụ thuộc vào GCC làm bộ dịch trung gian C (transpiler to C). Cho phép Stage 1 tự dịch trực tiếp mã nguồn AXIOM thành tập tin đối tượng máy (`.obj` / `.o`) và liên kết trực tiếp thành file nhị phân thực thi độc lập.
* **Nhiệm vụ cụ thể**:
  - Tích hợp bộ chọn lệnh mã máy x86 (`x86_selector.ax`), bộ phân bổ thanh ghi (`x86_regalloc.ax`), và bộ sinh mã máy (`x86_asm_emitter.ax`) vào luồng điều khiển chính của Stage 1 (`main_air.ax`).
  - Sử dụng trình liên kết tự trị sẵn có (`linker.ax`) để ghi các bảng relocations, import table `.idata` (Windows PE) và Procedure PLT/GOT (Linux ELF64) trực tiếp từ Stage 1.
  - Đảm bảo file thực thi tạo ra có thể chạy độc lập không phụ thuộc vào trình biên dịch GCC bên ngoài.

### 🚀 Hướng 2: Tự trị hóa Thư viện chuẩn I/O và Tương tác Hệ điều hành (`std/io.ax` & `std/os.ax`)
* **Mục tiêu**: Cắt đứt hoàn toàn FFI C đối với các tác vụ nhập/xuất tiêu chuẩn, đọc/ghi tập tin và tương tác hệ thống.
* **Nhiệm vụ cụ thể**:
  - Viết lại `std/io.ax` để thực hiện file read/write trực tiếp thông qua Syscall 0 (`sys_read`) và Syscall 1 (`sys_write`) trên Linux, và API nhân `ReadFile` / `WriteFile` trên Windows.
  - Viết lại `std/os.ax` để truy xuất biến môi trường, tham số dòng lệnh trực tiếp từ cấu trúc PEB (Process Environment Block) trên Windows và ELF Stack arguments trên Linux.
  - Cắt đứt hoàn toàn sự phụ thuộc vào UCRT (`ucrtbase.dll`) trên Windows và `libc.so.6` trên Linux cho các tác vụ I/O của thư viện chuẩn.

### 🚀 Hướng 3: Hoàn thiện Dynamic Collections chuẩn công nghiệp (`std/collections.ax`)
* **Mục tiêu**: Xây dựng cấu trúc dữ liệu Generic `Vector[T]` và `HashMap[K, V]` viết 100% bằng AXIOM thuần, hỗ trợ giải phóng tự động thông qua đường ống CTGC (Compile-Time Garbage Collection).
* **Nhiệm vụ cụ thể**:
  - Triển khai `HashMap` sử dụng giải thuật Robin Hood hashing và open-addressing tối ưu cache line CPU.
  - Tích hợp bộ tiêm destructor `=destroy` tự động để quản lý dọn dẹp tài nguyên động của mảng co giãn và bảng băm một cách an toàn, tuyệt đối không rò rỉ bộ nhớ.
  - Tối ưu hóa hiệu năng và benchmark độ trễ so với các thư viện chuẩn của các ngôn ngữ hiện hành.

---

## 🔄 Tiến Trình Thực Hiện Tuần Tự (Execution Pipeline)

Chúng tôi sẽ tiến hành thực hiện tuần tự theo thứ tự ưu tiên:
1. **Pha 1**: Tích hợp bộ sinh mã máy trực tiếp và trình liên kết nhị phân độc lập vào Stage 1, bypass GCC.
2. **Pha 2**: Tách rời I/O và tương tác hệ điều hành khỏi Glibc/UCRT, thay thế bằng direct OS Syscalls trong `std/io.ax` và `std/os.ax`.
3. **Pha 3**: Hoàn thiện bộ sưu tập Generic dynamic collections hiệu năng cao `Vector[T]` và `HashMap[K, V]` tích hợp CTGC.
