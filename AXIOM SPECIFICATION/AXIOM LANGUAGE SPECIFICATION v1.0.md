# AXIOM LANGUAGE SPECIFICATION v1.0
*Bản đặc tả kỹ thuật kiến trúc, trình biên dịch và môi trường thực thi*

---

## 1. DESIGN PHILOSOPHY [Stable]
*   **Rationale:** C++ mang theo gánh nặng lịch sử, Rust có "borrow checker" làm chậm tốc độ học, Python quá chậm. Axiom ra đời để giải quyết bài toán: Hiệu năng C, An toàn bộ nhớ của Rust, Cú pháp của Python, và Khả năng đọc hiểu nội tại cho Trí tuệ Nhân tạo (AI-Native).
*   **Architecture:** 
    *   **Data-oriented:** Tối ưu hóa CPU Cache mặc định.
    *   **AI-Native Observability:** Không chỉ sinh mã máy, trình biên dịch xuất ra Đồ thị tri thức (Knowledge Graph).
    *   **Zero-Friction Safety:** An toàn bộ nhớ qua Tham chiếu thế hệ (Generational References) thay vì Trình kiểm tra mượn (Borrow Checker).
*   **Trade-offs:** Ưu tiên tính tất định (Determinism) và tốc độ biên dịch hơn là tối ưu hóa mã máy cực hạn (cấp độ vi mô) ở các bản build debug.

---

## 2. SYNTAX & GRAMMAR [Stable]
*   **Rationale:** Giảm thiểu "nhiễu" thị giác, tối ưu hóa số lượng token để các mô hình LLM có thể đọc lượng code lớn hơn.
*   **Architecture:** Ngữ pháp phi ngữ cảnh (Context-Free LL(1)), dựa trên thụt lề (Indentation-based) với khoảng trắng bắt buộc (4 spaces).
*   **Examples:**
    ```axiom
    import std.fs

    pub fn read_config(path: string) -> Result[string, Error]:
        mut file := await fs.open(path)
        defer file.close()
        return file.read_all()
    ```
*   **Implementation Notes:** Lexer không cấp phát bộ nhớ (Zero-copy), tự động sinh ra các token `INDENT` và `DEDENT`. Parser dùng đệ quy xuống (Recursive Descent) kết hợp Pratt Parser cho biểu thức.

---

## 3. TYPE SYSTEM [Stable]
*   **Rationale:** Cần một hệ thống kiểu nghiêm ngặt tĩnh nhưng không cần viết quá nhiều (Type Inference).
*   **Architecture:** Nominal typing. Hỗ trợ Monomorphized Generics (nhân bản tĩnh), Structural Interfaces (Duck-typing), và Sum Types.
*   **Examples:**
    *   `!T`: Kiểu sở hữu độc quyền (Owned).
    *   `Isolated[T]`: Đồ thị bộ nhớ không có tham chiếu ngoại lai.
*   **Trade-offs:** Monomorphization có thể làm phình to kích thước file nhị phân (Code bloat) so với Type Erasure.
*   **Implementation Notes:** Sử dụng thuật toán Lazy Field Analysis (Phân tích lười biếng kiểu Zig) - chỉ phân giải kiểu cho những hàm thực sự được gọi.

---

## 4. MEMORY MODEL [Stable]
*   **Rationale:** Quản lý bộ nhớ thủ công gây lỗi, Garbage Collection (GC) gây giật lag (latency).
*   **Architecture:** Lấy cảm hứng từ Vale.
    *   **Single Ownership:** Mọi phân vùng Heap chỉ có 1 chủ sở hữu. Trình biên dịch tự chèn `free()` khi thoát scope (CTGC).
    *   **Generational References:** Mỗi phân vùng Heap chứa một ID thế hệ 64-bit. Con trỏ lưu cả địa chỉ và ID. Giải tham chiếu sai ID sẽ văng `Panic` an toàn thay vì lỗi `Use-After-Free`.
*   **Trade-offs:** Tốn thêm 8 byte header cho mỗi object trên Heap và 1 lệnh so sánh khi giải tham chiếu. 
*   **Implementation Notes:** Cung cấp block `unsafe {}` hoặc `in [Arena]` để tắt kiểm tra ID trong vòng lặp game/toán học để đạt tốc độ C nguyên thủy.

---

## 5. CONCURRENCY MODEL [Stable]
*   **Rationale:** Xử lý hàng triệu kết nối mạng mà không bị bế tắc (deadlock) do khóa bộ nhớ (Mutex).
*   **Architecture:** Actor Model (như Erlang) kết hợp CSP. Mỗi Actor sở hữu một vùng nhớ Heap hoàn toàn độc lập (Isolated Heaps).
*   **Examples:**
    ```axiom
    let actor_id = spawn worker_process(isolate(my_data))
    ```
*   **Trade-offs:** Không thể chia sẻ một biến đếm (counter) toàn cục một cách dễ dàng, phải dùng Message Passing.
*   **Implementation Notes:** Zero-copy message passing đạt được bằng cách chuyển giao pointer của kiểu `Isolated[T]` giữa 2 Actor Heaps.

---

## 6. COMPILER ARCHITECTURE [Stable]
*   **Rationale:** Phải biên dịch được 1 triệu dòng code/giây để hỗ trợ phản hồi tức thì (hot-reloading).
*   **Architecture:**
    *   Luồng: `Lexer` $\to$ `Flat AST` $\to$ `Semantic Graph` $\to$ `AIR` $\to$ `C-Backend/Native`.
    *   Cấu trúc dữ liệu: Flat Array. Các AST Node được tham chiếu bằng Index (`u32`), không dùng con trỏ (Pointer-free).
*   **Trade-offs:** Viết mã thao tác trên Flat AST khó hơn dùng Tree/Pointer thông thường.
*   **Implementation Notes:** Trình biên dịch MVP sử dụng C-Backend, dịch thẳng AST sang ngôn ngữ C (tuân thủ C11/C23), sau đó gọi GCC/Clang dưới nền để sinh mã máy.

---

## 7. INTERMEDIATE REPRESENTATION (AIR) [Stable]
*   **Rationale:** Giữ lại ý định ngữ nghĩa (Semantic intent) thay vì làm phẳng quá mức như LLVM IR.
*   **Architecture:** Dạng SSA (Static Single Assignment) 3 địa chỉ, đóng gói tĩnh 16 bytes/lệnh. Hỗ trợ khối `loop_region` bậc cao.
*   **Examples:** `%r1: i32 = add %r2, 1`.
*   **Trade-offs:** Ít chi tiết phần cứng hơn LLVM IR, nhưng dễ dàng cho AI đọc hiểu và suy luận tĩnh.
*   **Implementation Notes:** Luồng dữ liệu phải duy trì trạng thái của *Connection Graph* để phục vụ cho Compile-Time GC (Tái sử dụng bộ nhớ in-place).

---

## 8. RUNTIME ARCHITECTURE [Experimental]
*   **Rationale:** Cần một vi nhân (microkernel) siêu nhẹ để điều phối Actors.
*   **Architecture:**
    *   **M:N Scheduler:** Ánh xạ M Actors vào N OS-Threads (như Goroutines).
    *   **AxAlloc:** Bộ cấp phát bộ nhớ NUMA-aware, lấy cảm hứng từ `mimalloc`, quản lý bộ nhớ độc lập cho từng Actor. Khi Actor crash, toàn bộ Segment được trả về OS trong $O(1)$.
*   **Trade-offs:** Có Runtime nghĩa là không thể dễ dàng nhúng Axiom vào các hệ thống Bare-metal không có OS (trừ khi dùng cờ `--runtime=none`).
*   **Implementation Notes:** Ngân sách thực thi (Reduction budget) được áp dụng để preempt (chiếm quyền) các Actor chiếm CPU quá lâu.

---

## 9. EXECUTABLE FORMAT [Stable]
*   **Rationale:** Phải tương thích với hệ điều hành hiện tại nhưng mở rộng khả năng suy luận cho AI.
*   **Architecture:** Xuất định dạng chuẩn (ELF/PE/Mach-O).
    *   Chèn thêm Section `.axmeta`: Chứa toàn bộ Đồ thị ngữ nghĩa (Semantic Graph) dạng JSON nén.
*   **Trade-offs:** Tăng nhẹ dung lượng file nhị phân (có thể bị strip nếu không cần thiết).
*   **Implementation Notes:** `.axmeta` được Linker nhúng trực tiếp sau khi hoàn tất Codegen.

---

## 10. SECURITY MODEL [Experimental]
*   **Rationale:** Mã độc chuỗi cung ứng (Supply chain attacks) đang là vấn nạn.
*   **Architecture:** Capability-based Security.
    *   Các Module phải khai báo quyền tường minh: `import std.fs { read }`.
*   **Trade-offs:** Gây thêm rườm rà khi sử dụng thư viện bên thứ 3.
*   **Implementation Notes:** Compiler chặn tĩnh bất kỳ hàm nào cố gắng gọi syscall ngoài scope quyền hạn đã xin.

---

## 11. AI-NATIVE ARCHITECTURE [Stable]
*   **Rationale:** Ngôn ngữ lập trình đầu tiên có giao tiếp 2 chiều với Trí tuệ Nhân tạo.
*   **Architecture:**
    *   AI có thể đọc file `.axmeta` để hiểu ngữ nghĩa cấp cao.
    *   Lập trình viên dùng annotation `@[ai::suggest_layout]` để kích hoạt LLM trong quá trình build, đề xuất đổi từ cấu trúc AoS sang SoA.
*   **Trade-offs:** Phụ thuộc vào công cụ LSP và AI Engine cục bộ.
*   **Implementation Notes:** Cung cấp API nội bộ `std.compiler.ai` để thao tác AST bằng LLM.

---

## 12. QUANTUM EXTENSION [Future Features]
*   **Rationale:** Chuẩn bị cho kỷ nguyên phần cứng Lượng tử lai (Hybrid QPU-CPU).
*   **Architecture:** Mở rộng `OIR-Q` với thanh ghi lượng tử `%q`. Kiểu dữ liệu `qbit` bị ép buộc Move-semantics (định lý No-cloning).
*   **Trade-offs:** Chỉ nằm trên giấy tờ thiết kế, loại bỏ hoàn toàn khỏi MVP.

---

## 13. PACKAGE ECOSYSTEM [Stable]
*   **Rationale:** Quản lý gói phi tập trung, chống sập server trung tâm.
*   **Architecture:** Dùng file `axiom.toml`. Kéo source trực tiếp từ Git repo.
*   **Implementation Notes:** File `axiom.lock` tự động sinh mã băm SHA-256 để bảo vệ tính toàn vẹn (Integrity Verification).

---

## 14. STANDARD LIBRARY [Stable]
*   **Rationale:** Triết lý "Batteries included" (đầy đủ đồ nghề).
*   **Architecture:** 
    *   `std.collections` (memory-aware).
    *   `std.concurrency` (Actors, Channels).
    *   `std.gpu` [Future] (Zero-copy tới VRAM).
*   **Implementation Notes:** Toàn bộ Stdlib được viết 100% bằng Axiom, đóng vai trò làm Test-case cho Trình biên dịch.

---

## 15. TOOLCHAIN [Stable]
*   **Rationale:** Tránh phân mảnh công cụ (như Make, CMake, Ninja của C++).
*   **Architecture:** Monolithic Binary mang tên `axc`. Bao gồm compiler, formatter, profiler, package manager và LSP.
*   **Implementation Notes:** Cờ CLI thiết kế tối giản: `axc build`, `axc fmt`, `axc lsp`. Trình Formatter là Zero-configuration (Không cho phép cấu hình style).

---

## 16. SELF-HOSTING ROADMAP [Stable]
*   **Giai đoạn 1 (Tháng 1-6):** Prototype Compiler bằng Go. Lexer/Parser ra Flat AST. Dịch ra mã C.
*   **Giai đoạn 2 (Tháng 7-12):** Bootstrap. Viết mã nguồn Compiler bằng Axiom. Dùng Prototype biên dịch ra file C.
*   **Giai đoạn 3 (Tháng 13-18):** 100% Self-hosted. Compiler mới tự biên dịch chính nó. Đạt tốc độ > 500k loc/s.

---

## 17. MVP IMPLEMENTATION ROADMAP [Stable]
*Kế hoạch triển khai cho 1 kỹ sư và AI trợ lực.*
*   **Tuần 1-4:** EBNF Grammar $\to$ Zero-copy Lexer $\to$ Recursive Descent Parser $\to$ Flat AST (Bằng Go).
*   **Tuần 5-8:** Semantic Graph $\to$ Type Checker (Types cơ bản) $\to$ C-Transpiler Backend. *Mục tiêu: Dịch thành công Hello World ra C.*
*   **Tuần 9-12:** Ownership Semantics $\to$ Tự chèn `free()` (CTGC) $\to$ Tiêm Generational ID $\to$ Fix bugs. Phóng bản `v0.0.1`.

---
**TỔNG KẾT:** 
Bản thiết kế này đã khóa mục tiêu kiến trúc. Bằng cách sử dụng **Flat AST**, **C-Backend cho MVP**, và **M:N Actor Scheduler**, một kỹ sư hoàn toàn có thể hiện thực hóa Axiom bằng Go trong vòng 1 năm đầu tiên mà vẫn duy trì nguyên vẹn tầm nhìn về một ngôn ngữ "AI-Native" và an toàn bộ nhớ thế hệ mới.