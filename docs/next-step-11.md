# AXIOM Language Project — Lộ Trình Phát Triển Chiến Lược Và Tự Biên Dịch Tự Trị (next-step-11)

Tài liệu này ghi nhận kết quả hoàn thành giai đoạn sửa lỗi thư viện tiến trình freestanding và đặt ra lộ trình kiểm chứng biên dịch tự trị Native Backend cùng phát triển thư viện mạng độc lập.

---

## 🧭 Danh Sách Nhiệm Vụ (Task Checklist)

### Giai Đoạn 1: Sửa Lỗi Thư Viện Tiến Trình Freestanding (Đã hoàn thành 100%)
- [x] **Nhiệm vụ 1.1: Khắc phục lỗi Struct Order & Incomplete Type trong C-codegen**
  * **Giải pháp**: Trì hoãn việc sinh mã định nghĩa struct trong `processTree` (file `codegen/cgen/decls.go`). Đưa TypeID của struct vào `e.queue` để hàm `e.drainTypeDecls()` tự động phân tích và sinh mã struct/generic theo đúng thứ tự phụ thuộc.
- [x] **Nhiệm vụ 1.2: Khắc phục lỗi Conflicting Types cho các hàm Win32 Extern**
  * **Giải pháp**: Thêm các hàm extern như `CreateProcessA`, `WaitForSingleObject` vào danh sách `stdLibFuncs` trong `codegen/cgen/decls.go` để trình biên dịch bỏ qua việc sinh prototype C đè lên.
- [x] **Nhiệm vụ 1.3: Chạy Kiểm Thử Tích Hợp và Xác Minh**
  * **Hành động**: Chạy lệnh `go test -v -run TestAxiomProcess ./tests/codegen` trên Windows. Kết quả đã biên dịch thành công và kiểm thử pass hoàn toàn.

### Giai Đoạn 2: Kiểm Chứng Trình Biên Dịch Tự Trị Natively và Freestanding (Đang thực hiện)
- [ ] **Nhiệm vụ 2.1: Viết test case biên dịch tự trị thông qua Native Backend trực tiếp**
  * **Mục tiêu**: Chạy thử nghiệm trình biên dịch tự trị viết bằng Axiom (`main_air.ax` hay `bin/axc_stage2_selfhosted.exe`) với cờ tự liên kết (`self_link = true`) để sinh mã máy PE COFF trực tiếp cho một chương trình đơn giản, sau đó liên kết nhị phân độc lập không qua GCC/MSVC.
- [ ] **Nhiệm vụ 2.2: Kiểm tra tính đúng đắn của file thực thi tự trị sinh ra**
  * **Hành động**: Thực thi nhị phân sinh ra từ Native Self-Linker để đảm bảo nó hoạt động không có lỗi bộ nhớ.

### Giai Đoạn 3: Thiết Kế Thư Viện Mạng Freestanding `std/net.ax` (Đang thực hiện)
- [ ] **Nhiệm vụ 3.1: Định nghĩa các cấu trúc Socket FFI thuần túy**
  * **Chi tiết**: Khai báo cấu trúc FFI của Winsock trên Windows (`SOCKET`, `sockaddr_in`) và syscall Socket trên Linux.
- [ ] **Nhiệm vụ 3.2: Triển khai các API cơ bản (socket, bind, listen, accept, connect, send, recv)**
  * **Mục tiêu**: Hỗ trợ giao tiếp TCP/IP cơ bản freestanding hoàn toàn không phụ thuộc C standard library.
