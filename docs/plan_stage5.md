# Kế hoạch Thực hiện Stage 5 — Tích hợp Kiểm soát Sở hữu, CTGC & Trình Sinh Mã C Tự Trị (100% Self-Hosted CodeGen)

Tài liệu này chi tiết hóa kế hoạch kỹ thuật nhằm tích hợp bộ kiểm soát sở hữu (Memory Safety Pipeline) và hoàn thiện bộ sinh mã tự trị từ SSA AIR sang C11 trực tiếp trong trình biên dịch tự trị AXIOM Stage 1.

---

## 1. MỤC TIÊU & TẦM NHÌN KIẾN TRÚC

Hiện tại, `axc_stage1.exe` đã phân tích cú pháp + ngữ nghĩa và sinh ra **SSA AIR (Axiom Intermediate Representation)** chính xác, nhưng chưa thể tự giải phóng bộ nhớ tĩnh (CTGC) hoặc tự dịch mã nguồn sang file thực thi nhị phân thông qua FFI/C compiler.

Hoàn thành Stage 5 sẽ giúp:
1. **An toàn bộ nhớ tĩnh hoàn toàn (Static Memory Safety)**: Kiểm soát Single Ownership, Escape Analysis, và tự động chèn giải phóng bộ nhớ tĩnh tại thời điểm biên dịch (CTGC) + tối ưu hóa tái sử dụng bộ nhớ (Alias Reuse).
2. **Loại bỏ hoàn toàn trình biên dịch Go**: `axc_stage1.exe` sẽ tự sinh mã C trực tiếp từ SSA AIR, gọi C compiler (`gcc`/`clang`/`cl`) hệ thống để sinh mã máy, tự biên dịch chính nó (Self-Hosting thành công hoàn toàn).

---

## 2. PHÂN TÍCH THIẾT KẾ CÁC SUBSYSTEM KỸ THUẬT

### 2.1. Đồ thị Kết nối & Ownership Checker (`connection_graph.ax`, `ownership.ax`)
* **Kiến trúc**:
  * Mô hình hóa mối quan hệ sở hữu dưới dạng Đồ thị Kết nối (Connection Graph) với các loại cạnh: `EDGE_OWNS`, `EDGE_BORROWS`, `EDGE_FLOWS_TO`, `EDGE_ESCAPES_TO`, `EDGE_REUSED_BY`.
  * Flat Data-Oriented: một `U32Vec` (được lập chỉ mục trực tiếp bởi `SymID` tuần tự) trỏ tới `NodeID` của đồ thị trong thời gian $O(1)$.
* **Quy tắc sở hữu (Ownership Rules)**:
  * Tránh dùng lại biến đã bị chuyển vùng sở hữu (`use of moved value` -> `E4001`).
  * Kiểm tra tính khả biến (mutability) của biến khi gán giá trị (`cannot assign to immutable variable` -> `E4002`).

### 2.2. Phân tích Thải hồi & CTGC (`escape.ax`, `ctgc.ax`, `alias_reuse.ax`)
* **Escape Analysis**: Duyệt qua đồ thị kết nối bằng thuật toán DFS khử đệ quy/vòng lặp để phát hiện các biến thoát (`EscapesToHeap`). Nếu biến không thoát và kích thước nhỏ hơn ngưỡng (ví dụ: 1024 bytes), nó sẽ được cấp phát trên Stack thay vì Heap.
* **CTGC Injection**: Tự động chèn các câu lệnh `NodeDestroyStmt` tại cuối các khối block hoặc trước các lệnh `return` (theo thứ tự LIFO - vào sau ra trước) cho các biến heap chưa bị di chuyển sở hữu.
* **Alias Reuse**: Tìm kiếm cặp lệnh giải phóng và cấp phát tiếp theo của cùng một kiểu dữ liệu để chuyển đổi chúng thành `NodeAliasStmt`, tái sử dụng con trỏ vùng nhớ nhằm loại bỏ hoàn toàn chi phí hệ điều hành cấp phát lại.

### 2.3. Sinh Mã C Từ SSA AIR (`cgen.ax`)
* **Sinh mã từ AIR**:
  * SSA AIR đã làm phẳng hoàn toàn các vòng lặp (`while`, `for`), các biểu thức nhánh phức tạp (`if`/`elif`/`else`, `match` expressions), các phép tính lồng nhau và các lệnh quản lý tài nguyên.
  * Bộ sinh mã C từ SSA AIR (`AirCGen`) chỉ tốn ~350 dòng mã nguồn Axiom.
* **Quy trình hoạt động**:
  1. Khai báo tất cả các thanh ghi SSA (`ax_i32 r_1;` v.v.) ở đầu hàm C để tránh xung đột khai báo chéo goto.
  2. Phát nhãn cho các Basic Block (`block_n: ;`).
  3. Dịch các chỉ thị SSA sang câu lệnh C11 tương thích.
  4. Trình điều khiển biên dịch sẽ nối file sinh ra với hệ thống runtime (`runtime/*.c`) và gọi compiler hệ thống thông qua FFI `system()`.

---

## 3. CHI TIẾT CÁC PHÂN ĐOẠN TASK THỰC THI

### Task 1: Thiết lập Trình Điều khiển Thống nhất (`main_driver.ax`)
* **Đầu vào**: Tệp mã nguồn AXIOM.
* **Đầu ra**: Tùy chọn xuất ra SSA AIR (cho kiểm thử so khớp tương thích) hoặc tự sinh mã C và gọi GCC/Clang/MSVC hệ thống để biên dịch.
* **Hỗ trợ giao diện dòng lệnh**:
  * Dạng kiểm thử (cho `triple_build.ps1`): `axc_stage1.exe <file.ax>` hoặc `axc_stage1.exe dump-air <file.ax>` -> Xuất ra SSA AIR.
  * Dạng sinh mã C: `axc_stage1.exe emit-c <file.ax> -o <out.c>` -> Xuất ra tệp C.
  * Dạng biên dịch trực tiếp: `axc_stage1.exe build <file.ax> -o <out.exe>` -> Sinh mã C tạm, gọi trình biên dịch C hệ thống để sinh nhị phân.

### Task 2: Chèn các giai đoạn xử lý Ownership & Memory Safety vào pipeline
* Tích hợp tuần tự:
  1. `OwnershipChecker`: Kiểm tra và báo lỗi Single Ownership (`E4001`) và Mutability (`E4002`).
  2. `EscapeAnalyser`: Phân tích sự thoát của các biến để đặt cờ `FLAG_ESCAPES_TO_HEAP`.
  3. `CtgcInjector`: Tiêm các lệnh destroy (`NODE_DESTROY_STMT`) tự động giải phóng tài nguyên.
  4. `AliasReuseOptimizer`: Chuyển đổi các cặp destroy + alloc kề nhau của cùng một loại sang `OP_ALIAS_REUSE`.
* Chú ý: Cần tắt các bước chèn này đối với lệnh `dump-air` để khớp hoàn hảo 100% từng byte với Stage 0 (vốn không chạy các bước này trong lệnh `dump-air`).

### Task 3: Cải tiến Script Triple-Build (`scripts/triple_build.ps1`)
* Cập nhật danh sách tệp biên dịch của Stage 1 để bao gồm:
  - `bootstrap/stage1/connection_graph.ax`
  - `bootstrap/stage1/ownership.ax`
  - `bootstrap/stage1/escape.ax`
  - `bootstrap/stage1/ctgc.ax`
  - `bootstrap/stage1/alias_reuse.ax`
  - `bootstrap/stage1/cgen.ax`
  - `bootstrap/stage1/main_driver.ax` (thay thế `main_air.ax` làm entry point)

### Task 4: Kiểm chứng & So khớp Độc lập các test cases trong `tests/`
* Chạy `triple_build.ps1` và xác thực 100% các ca thử nghiệm đều vượt qua, đảm bảo sự toàn vẹn của mã và tính tất định (byte-for-byte deterministic).

---

## 4. KẾ HOẠCH KIỂM THỬ & XÁC MINH (VERIFICATION PLAN)

### Kiểm thử Tự động (Automated Tests)
* Chạy `scripts/triple_build.ps1` để xác nhận toàn bộ quá trình tự dịch hoạt động đúng đắn.
* Đảm bảo tính tất định: So khớp SHA256/MD5 của các trình biên dịch được dịch chéo qua các thế hệ.

### Kiểm thử Thủ công (Manual Verification)
* Sử dụng Stage 1 mới biên dịch `tests/valid_hello.ax` hoặc `tests/valid_fibonacci.ax` ra tệp `.exe` thực thi trực tiếp trên Windows và chạy thử nghiệm.
