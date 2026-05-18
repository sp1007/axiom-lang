# AXIOM LANGUAGE SPECIFICATION v1.0

## 1. LANGUAGE OVERVIEW

### 1.1. Mục tiêu (Goals)
**Axiom** là ngôn ngữ lập trình hệ thống biên dịch trước (AOT-compiled), định kiểu tĩnh, được thiết kế cho kỷ nguyên Trí tuệ Nhân tạo. Mục tiêu cốt lõi là cung cấp hiệu năng ngang C++, an toàn bộ nhớ tuyệt đối mà không cần Trình mượn (Borrow Checker) hay Bộ thu gom rác (Garbage Collector), và khả năng đọc hiểu ngữ nghĩa hoàn hảo cho các mô hình AI.

### 1.2. Triết lý thiết kế (Design Philosophy)
*   **Data-Oriented by Default:** Dữ liệu và logic bị tách rời hoàn toàn. Trình biên dịch có quyền tự do tối ưu hóa cấu trúc bộ nhớ (AoS sang SoA) mà không làm vỡ logic.
*   **AI-Native Observability:** Mã nguồn không chỉ sinh ra mã máy mà còn xuất ra Đồ thị tri thức thực thi (Executable Knowledge Graph - `.axmeta`) cho phép AI suy luận luồng dữ liệu (Dataflow) tĩnh.
*   **Zero-Friction Safety:** Đạt được an toàn bộ nhớ bằng "Generational References" (Tham chiếu thế hệ) và "Single Ownership".

### 1.3. Nguyên tắc cốt lõi (Core Principles)
*   **Deterministic Behavior (Tính tất định):** Phân tích cú pháp 100% Context-Free. Không có luồng điều khiển ẩn (No hidden control flow).
*   **Low Context-Sensitivity:** Cú pháp minh bạch, một từ khóa chỉ có một ý nghĩa duy nhất.
*   **Zero-Cost Abstractions:** Các trừu tượng hóa bậc cao (Generics, Interfaces, Async) không phát sinh chi phí lúc thực thi (runtime overhead).

### 1.4. Non-goals (Không thuộc phạm vi)
*   Không hỗ trợ Kế thừa Hướng đối tượng (No OOP inheritance).
*   Không có Garbage Collection chạy ngầm (No tracing GC).
*   Không xây dựng một ABI hoàn toàn mới (Sử dụng chuẩn C-ABI để tương thích ngược).

### 1.5. Thuật ngữ (Terminology)
*   **AIR (Axiom Intermediate Representation):** Biểu diễn trung gian đa tầng dạng mảng phẳng.
*   **Connection Graph:** Đồ thị theo dõi vòng đời đối tượng tĩnh, dùng cho Escape Analysis.
*   **Actor:** Một tiến trình siêu nhẹ có bộ nhớ Heap cô lập hoàn toàn.

---

## 2. LEXICAL STRUCTURE

### 2.1. Encoding & Unicode
*   Mã nguồn phải được mã hóa chuẩn **UTF-8**.
*   Hỗ trợ ký tự Unicode trong chuỗi (String) và nhận dạng trực tiếp các kiểu `char8`, `char16`, `char32`.

### 2.2. Whitespace Rules (Quy tắc khoảng trắng)
*   Axiom là ngôn ngữ **Indentation-based** (dựa trên thụt lề).
*   **Chỉ cho phép dùng Space (khoảng trắng)** để thụt lề. Bắt buộc **4 spaces** cho mỗi cấp độ khối lệnh (Block). Việc sử dụng ký tự Tab (`\t`) sẽ gây ra lỗi `SyntaxError` ngay lập tức.

### 2.3. Identifiers (Định danh)
*   Regex: `[a-zA-Z_][a-zA-Z0-9_]*`
*   Định danh phân biệt chữ hoa, chữ thường. Khuyến khích `snake_case` cho hàm/biến, và `PascalCase` cho Struct/Interface.

### 2.4. Keywords (Từ khóa)
Các từ khóa sau được bảo lưu tĩnh (không thể dùng làm tên biến):
`fn`, `mut`, `struct`, `interface`, `impl`, `match`, `spawn`, `async`, `await`, `comptime`, `import`, `pub`, `lock`, `isolate`, `return`, `if`, `else`, `for`, `in`, `SOA`.

### 2.5. Literals (Trực quan)
*   **Integer:** `42`, `0x2A` (Hex), `0b101010` (Bin), `1_000_000` (Digit separators).
*   **Float:** `3.14`, `1e-9`.
*   **String:** `"Hello"`, Triple-quoted `"""Block string"""` cho văn bản nhiều dòng.

### 2.6. Comments (Chú thích)
*   Ghi chú đơn dòng: `// comment`
*   Ghi chú văn bản tài liệu (Docstrings): `/// documentation` (Trình biên dịch parse trực tiếp thành AST Node để xuất tài liệu/Meta cho AI).

---

## 3. SYNTAX GRAMMAR (EBNF)

Axiom sử dụng 문 pháp Context-Free, LL(1) thân thiện với việc parse bằng Recursive Descent.

```ebnf
Program         ::= { ImportDecl | Statement }
ImportDecl      ::= "import" Identifier ["as" Identifier] NEWLINE

(* --- Declarations --- *)
Decl            ::= FnDecl | StructDecl | InterfaceDecl | ImplDecl | ConstDecl
ConstDecl       ::= ["pub"] ["comptime"] "const" Identifier "=" Expr NEWLINE

(* --- Functions --- *)
FnDecl          ::= ["pub"] ["async"] "fn" Identifier [GenericParams] "(" [ParamList] ")" ["->" Type] ":" Block
ParamList       ::= Param { "," Param }
Param           ::= ["mut" | "isolate"] Identifier ":" Type
GenericParams   ::= "[" GenericParam { "," GenericParam } "]"
GenericParam    ::= Identifier [ ":" Type ]

(* --- Structures & Interfaces --- *)
StructDecl      ::= ["pub"] ["@SOA"] "struct" Identifier [GenericParams] ":" Block
InterfaceDecl   ::= ["pub"] "interface" Identifier [GenericParams] ":" Block
ImplDecl        ::= "impl" Identifier "for" Type ":" Block

(* --- Blocks & Statements --- *)
Block           ::= NEWLINE INDENT { Statement } DEDENT
Statement       ::= Decl | VarDecl | Assignment | ExprStmt | IfStmt | ForStmt | MatchStmt | SpawnStmt | ReturnStmt

VarDecl         ::= ["mut"] Identifier [":" Type] ":=" Expr NEWLINE
Assignment      ::= Expr "=" Expr NEWLINE
SpawnStmt       ::= "spawn" Expr NEWLINE
ReturnStmt      ::= "return" [Expr] NEWLINE

IfStmt          ::= "if" Expr ":" Block {"elif" Expr ":" Block} ["else" ":" Block]
ForStmt         ::= "for" Identifier "in" Expr ":" Block

(* --- Pattern Matching --- *)
MatchStmt       ::= "match" Expr ":" NEWLINE INDENT { MatchBranch } DEDENT
MatchBranch     ::= Pattern "=>" ( Expr NEWLINE | Block )
Pattern         ::= "." Identifier ["(" Identifier ")"] | "_"

(* --- Expressions --- *)
Expr            ::= OrExpr
OrExpr          ::= AndExpr { "or" AndExpr }
AndExpr         ::= EqExpr { "and" EqExpr }
EqExpr          ::= CompExpr { ("==" | "!=") CompExpr }
CompExpr        ::= AddExpr { ("<" | ">" | "<=" | ">=") AddExpr }
AddExpr         ::= MulExpr { ("+" | "-") MulExpr }
MulExpr         ::= UnaryExpr { ("*" | "/" | "%") UnaryExpr }
UnaryExpr       ::= ("-" | "not" | "await" | "#run") UnaryExpr | PrimaryExpr
PrimaryExpr     ::= Literal | Identifier | CallExpr | MemberAccess | "(" Expr ")"
CallExpr        ::= PrimaryExpr "(" [ArgList] ")"
MemberAccess    ::= PrimaryExpr "." Identifier
```

---

## 4. TYPE SYSTEM

Hệ thống kiểu của Axiom đánh giá theo danh nghĩa (Nominal Typing) nhưng hỗ trợ Structural Pattern Matching thông qua Sum types.

### 4.1. Primitive Types (Kiểu nguyên thủy)
*   **Integer:** `i8`, `i16`, `i32`, `i64`, `isize` / `u8`, `u16`, `u32`, `u64`, `usize`
*   **Floating Point:** `f32`, `f64`
*   **Boolean:** `bool`
*   **String:** `string` (UTF-8 by default)

### 4.2. Composite Types (Kiểu phức hợp)
*   **Structs:** Dữ liệu định hướng bản ghi (record). Struct mặc định nằm trên Stack trừ khi bị giới hạn vòng đời (escape analysis). Có thể được gán decorator `@SOA` để compiler tự động biên dịch thành Structure of Arrays trong bộ nhớ.
*   **Sum Types (Union):** Được định nghĩa ngầm định qua các biến thể enum và khớp vét cạn (Exhaustive Pattern Matching) qua `match`.

### 4.3. Traits/Interfaces & Generics
*   **Interfaces:** Định nghĩa hợp đồng (contract) các phương thức.
*   **Generics:** Không dùng Type Erasure. Trình biên dịch sử dụng cơ chế Monomorphization (Tạo bản sao tĩnh) để sinh mã máy tối ưu hóa (zero-cost) cho từng kiểu dữ liệu truyền vào.
*   **Compile-time Verification:** Mọi Generic bounding (`T: Drawable`) đều được phân tích và báo lỗi ngay tại AST, không chờ lúc instantiation như C++ Templates.

### 4.4. Hệ thống hiệu ứng (Effect System)
Hệ thống type theo dõi hiệu ứng phụ (Side-effects):
*   Hàm không đánh dấu được mặc định là Pure (Thuần túy).
*   Trình biên dịch có chỉ thị `#run` chỉ cho phép thực thi ở compile-time đối với các hàm không chứa hiệu ứng phụ về I/O hay Mutability ngoại vi.

### 4.5. Memory-Aware Typing & Capabilities
*   **Ownership Qualifier `!T`:** Xác định biến giữ quyền sở hữu tuyệt đối (Single Owner) của phân vùng bộ nhớ.
*   **`Isolated[T]`:** Kiểu năng lực (Capability Type) chứng minh đồ thị bộ nhớ của `T` không có bất kỳ liên kết tham chiếu nào từ bên ngoài. Bắt buộc dùng để truyền dữ liệu qua luồng bằng `spawn`.
*   **`Locker[T]`:** Cấu trúc ép buộc đồng bộ dữ liệu. Không thể truy cập `T` nếu không đặt trong khối `lock`.

---

## 5. MEMORY MODEL

Axiom thay thế hoàn toàn Borrow Checker và GC bằng mô hình **Hybrid Single-Ownership & Generational References**.

### 5.1. Ownership Rules (Quy tắc sở hữu)
1. Mỗi giá trị Heap có đúng một Chủ sở hữu (Single Owner).
2. Khi Chủ sở hữu ra khỏi Tầm vực (Scope), trình biên dịch chèn mã gọi hàm `=destroy` ngầm định (Implicit Hook) và tái sử dụng bộ nhớ (Compile-time GC - CTGC).
3. Dữ liệu có thể được di chuyển (Moved) sang chủ sở hữu khác. Sau khi di chuyển, biến gốc lập tức mất hiệu lực tĩnh tại compile-time (Tương tự Rust).

### 5.2. Generational References (Tham chiếu thế hệ)
Để cho phép Aliasing (nhiều biến cùng trỏ vào một dữ liệu) linh hoạt:
*   Mọi vùng nhớ Heap cấp phát động đều đi kèm một metadata siêu nhỏ chứa **Generation ID (64-bit)**.
*   Con trỏ tham chiếu lưu địa chỉ bộ nhớ VÀ Generation ID lúc mượn.
*   Mỗi khi giải tham chiếu (dereference), CPU so sánh hai ID. Nếu Chủ sở hữu đã xóa dữ liệu (ID vùng nhớ đã tăng lên), chương trình sẽ `panic` an toàn thay vì truy cập bộ nhớ rác (Use-after-free).

### 5.3. Region Memory (Bộ nhớ vùng / Arena)
Đối với các tác vụ có vòng đời giống hệt nhau (VD: HTTP Request, Game Tick):
*   Sử dụng cú pháp `in [Arena]`. Trình biên dịch tắt kiểm tra Generational ID bên trong Arena, đạt được hiệu suất ngang ngửa raw pointers của C/C++. Bộ nhớ được giải phóng với chi phí $O(1)$.

### 5.4. Concurrency Memory Safety (An toàn bộ nhớ đồng thời)
Sử dụng mô hình Actor cô lập bộ nhớ (Isolated Heaps) như Erlang:
*   Các Thread/Actor **KHÔNG** chia sẻ bộ nhớ Heap gốc.
*   Giao tiếp qua Channel (CSP) hoặc Message Passing.
*   Muốn gửi dữ liệu lớn giữa các Actor, phải chuyển đổi quyền sở hữu thông qua `Isolated[T]` để đạt zero-copy move semantics. Trình biên dịch kiểm tra tính hợp lệ qua Connection Graph.

### 5.5. Fragmentation Handling (Xử lý phân mảnh)
*   **CTGC In-place Reuse:** Trình biên dịch Axiom theo dõi AST. Nếu đối tượng $A$ chết ở dòng 10, và đối tượng $B$ cùng kiểu được tạo ở dòng 12, trình biên dịch sẽ sinh mã tái sử dụng thẳng địa chỉ của $A$ cho $B$ mà không cần gọi hệ thống (malloc/free).
*   **Size-Classed Pools:** Vùng Heap của mỗi Actor được cấu trúc thành các mảng block có độ dài cố định. Việc giải phóng chỉ đơn giản là đẩy index vào Free-list, triệt tiêu phân mảnh ngoại.

## 6. EXECUTION MODEL (MÔ HÌNH THỰC THI)

### 6.1. Runtime Behavior (Hành vi Runtime)
Axiom hỗ trợ hai chế độ thực thi:
*   **Bare-metal (Zero-runtime):** Trình biên dịch không nhúng bất kỳ runtime nào (cờ `--runtime=none`). Các tính năng như Actor, M:N Scheduling bị vô hiệu hóa. Phân bổ bộ nhớ hoàn toàn do lập trình viên kiểm soát tĩnh thông qua `alloc`/`free` hoặc `Arena`. Phù hợp cho lập trình nhúng, WebAssembly, hoặc viết OS.
*   **Managed Runtime (Mặc định):** Kích hoạt hệ sinh thái Actor và Async. Trình biên dịch nhúng một Runtime siêu nhẹ (vài KB) khởi chạy M:N Scheduler.

### 6.2. Stack Model (Mô hình ngăn xếp)
*   **Stack Allocation by Default:** Bất kỳ đối tượng nào không "thoát" (escape) khỏi đồ thị kết nối cục bộ của hàm đều được tự động cấp phát trên Stack thông qua phân tích tĩnh (Escape Analysis).
*   Ngăn xếp có thể mở rộng linh hoạt (Segmented/Growable Stacks) khi chạy trong chế độ Actor, khởi điểm chỉ từ vài trăm bytes.

### 6.3. Heap Model (Mô hình Heap)
*   **Isolated Heaps:** Không có "Global Heap" (Heap toàn cục) chia sẻ giữa các luồng. Mỗi Actor sở hữu một vùng Heap hoàn toàn độc lập.
*   Khi một Actor kết thúc hoặc bị lỗi (crash), toàn bộ Heap cục bộ của nó được giải phóng trả về cho OS ngay lập tức ($O(1)$) mà không cần quét rác.

### 6.4. Scheduling (Lập lịch)
*   Sử dụng trình lập lịch **M:N Adaptive Work-Stealing**. Ánh xạ $M$ tác vụ (Actors/Coroutines) vào $N$ luồng hệ điều hành.
*   Lập lịch ưu tiên (Preemptive scheduling) dựa trên số chu kỳ tính toán và các điểm chặn (I/O block points) để đảm bảo không một vòng lặp vô hạn nào có thể treo hệ thống.

### 6.5. Async & Actor Execution (Thực thi Bất đồng bộ và Actor)
*   **Zero-cost State Machines:** Các hàm `async` không phân bổ đối tượng `Promise` hay `Future` trên Heap. Chúng được trình biên dịch làm phẳng thành các Máy trạng thái (State Machines) hữu hạn, lưu trữ trực tiếp trạng thái trên Stack.
*   **Structured Concurrency:** Khối `awaitAll` đảm bảo mọi tác vụ con (spawned tasks) phải hoàn tất trước khi luồng cha tiếp tục, không bao giờ sinh ra tác vụ mồ côi (orphan tasks).

---

## 7. CONCURRENCY MODEL (MÔ HÌNH ĐỒNG THỜI)

### 7.1. Task Model & Actors
*   Mô hình đồng thời cốt lõi của Axiom là sự kết hợp giữa **Actor Model** và **CSP (Communicating Sequential Processes)**. 
*   Các Actor không chia sẻ trạng thái. Giao tiếp độc quyền thông qua việc gửi thông điệp (Message Passing).

### 7.2. Channels & Message Passing
*   **Channels:** Các kênh định kiểu tĩnh. Hỗ trợ truyền dữ liệu đồng bộ (synchronous) hoặc bất đồng bộ có bộ đệm (buffered).
*   **Zero-copy Send:** Bằng cách sử dụng kiểu năng lực `Isolated[T]` (đảm bảo dữ liệu không có tham chiếu từ bên ngoài), việc gửi một cấu trúc dữ liệu khổng lồ qua Channel chỉ tốn chi phí sao chép một con trỏ (1 word), thay vì sao chép toàn bộ dữ liệu (deep copy).

### 7.3. Lock-free Guarantees (Đảm bảo phi khóa)
*   Axiom mặc định **Data-race Free** ở cấp độ biên dịch. Trình biên dịch từ chối mã cố tình thay đổi dữ liệu từ 2 luồng khác nhau mà không thông qua kênh truyền.
*   Không có hiện tượng `Deadlock` gây ra bởi khóa bộ nhớ vì Runtime không sử dụng cơ chế `Mutex` toàn cục.

### 7.4. Deterministic Concurrency (Đồng thời tất định)
*   **Data-flow Variables:** Việc phân chia các tác vụ song song trên một mảng lớn sử dụng toán tử ghi trực tiếp vào một luồng dữ liệu (Data-flow variables) chưa được đánh giá. Thao tác đọc trên biến này bị tạm dừng (block) cho đến khi tác vụ ghi hoàn thành, đảm bảo kết quả mỗi lần chạy đều giống hệt nhau (100% deterministic).

### 7.5. Synchronization Primitives (Nguyên thủy đồng bộ)
*   Chỉ khi thật sự cần thiết, lập trình viên sử dụng kiểu `Locker[T]`. Để truy cập dữ liệu bên trong `Locker`, hệ thống kiểu ép buộc việc sử dụng khối lệnh tĩnh `lock var_name: ...`. Khi thoát khỏi khối lệnh, lock tự động được nhả (Higher RAII).

---

## 8. MODULE & PACKAGE SYSTEM (HỆ THỐNG MODULE & GÓI)

### 8.1. Module Layout & Imports
*   Một tệp (file) `.ax` tương đương với một Module. Axiom không có tệp tiêu đề (Header files) như C/C++.
*   Cú pháp: `import math.linalg as la`. Trình biên dịch phân giải (resolve) theo đường dẫn thư mục tuyệt đối hoặc tương đối.

### 8.2. Dependency Graph (Đồ thị phụ thuộc)
*   Hệ thống từ chối tĩnh (statically rejects) các Module phụ thuộc vòng tròn (Circular Dependencies) ở cấp độ kiến trúc tệp. Việc này ép buộc thiết kế phần mềm sạch (clean architecture) và tối ưu hóa tốc độ biên dịch đa luồng.

### 8.3. Package Structure (Cấu trúc gói)
*   Mọi Package Axiom đều được quản lý bởi file `axiom.toml` chứa Metadata, Tác giả, và Danh sách phụ thuộc.
*   Không có kho lưu trữ trung tâm duy nhất (như NPM hay Crates.io). Có thể kéo gói trực tiếp từ Git repo hoặc đường dẫn cục bộ.

### 8.4. Versioning & Cryptographic Verification
*   **SemVer (Semantic Versioning):** Bắt buộc sử dụng chuẩn `MAJOR.MINOR.PATCH` cho toàn bộ thư viện.
*   **Lockfile (`axiom.lock`):** Tự động sinh ra khi biên dịch lần đầu. Chứa chữ ký băm mật mã (Cryptographic Hashes - SHA-256) của toàn bộ cây phụ thuộc, đảm bảo tính toàn vẹn (Integrity) và bảo vệ khỏi các cuộc tấn công chuỗi cung ứng (Supply chain attacks).

---

## 9. COMPILER ARCHITECTURE (KIẾN TRÚC TRÌNH BIÊN DỊCH)

Kiến trúc Axiom Compiler (AxiomC) được thiết kế từ đầu (Self-contained) theo hướng Data-Oriented để đạt tốc độ biên dịch hàng triệu dòng code/giây.

### 9.1. Lexer & Parser
*   **Lexer:** Zero-copy, không cấp phát chuỗi mới trên Heap. Chỉ xuất ra một mảng phẳng (Flat Array) chứa cặp `[TokenID, Offset_in_Source]`.
*   **Parser:** Parallel & Deterministic. Vì cấu trúc Module không phụ thuộc vòng và không có `#include`, nhiều tệp có thể được phân tích cú pháp song song ra Flat AST hoàn toàn phi khóa (Lock-free).

### 9.2. AST & Semantic Graph
*   **Flat Array AST:** AST không dùng con trỏ (pointers), mà là các mảng (arrays) liền kề được liên kết bằng các chỉ số Index. Cực kỳ thân thiện với bộ nhớ đệm CPU (CPU Cache).
*   **Lazy Field Analysis:** Bắt chước Zig, trình phân tích ngữ nghĩa (Semantic Checker) kiểm tra lười biếng. Nó chỉ phân giải kiểu (Type resolution) đối với những hàm/trường cấu trúc thực sự được gọi ở runtime.
*   **Connection Graph:** Chạy trên bề mặt của AST để theo dõi và xây dựng bản đồ luồng di chuyển quyền sở hữu (Ownership Flow) nhằm phục vụ Escape Analysis và chèn các hook `=destroy` cho việc thu gom vùng nhớ tự động (CTGC).

### 9.3. Optimizer & Register Allocation
*   **Debug Mode:** Không có tối ưu hóa, sử dụng thuật toán cấp phát thanh ghi quét tuyến tính (Linear Scan Register Allocation) siêu tốc.
*   **Release Mode:** Áp dụng các Pass tối ưu như Constant Folding, Dead Code Elimination, Inlining.

### 9.4. Code Generation & Linker
*   **C-Backend (MVP Target):** Trình biên dịch dịch thẳng AST/IR ra ngôn ngữ C chuẩn (`C23` / `C11`) tương tự thiết kế của ngôn ngữ Nim và V. Sau đó, gọi `GCC` hoặc `Clang` nội bộ để sinh nhị phân. Cách này lợi dụng hệ sinh thái tối ưu hóa khổng lồ của C, tương thích chéo mọi hệ điều hành ngay lập tức.
*   **Native x86-64/ARM Backend (Future Target):** Khả năng sinh thẳng mã máy vào bộ nhớ đệm mà không cần Linker trung gian.

---

## 10. INTERMEDIATE REPRESENTATION (AIR - Axiom IR)

### 10.1. IR Structure & Format
*   Sử dụng mô hình **SSA (Static Single Assignment)**, thiết kế tuyến tính (Flat list of instructions) chia theo Basic Blocks.
*   Các thanh ghi ảo mang kiểu dữ liệu tĩnh (Strongly Typed IR), khác với hợp ngữ (assembly) cấp thấp. Định dạng: `%r1: i32 = add %r2, %r3`.

### 10.2. Optimization Passes
Các pha tối ưu hóa trên AIR bao gồm:
*   **Monomorphization Pass:** Xóa bỏ Generics, sinh mã cụ thể cho từng kiểu dữ liệu truyền vào (Zero-cost abstraction).
*   **CTGC Pass (Compile-time GC):** Tận dụng Đồ thị kết nối, AIR chèn các lệnh ghi đè vùng nhớ (In-place Memory Reuse) nếu một vòng đời đối tượng vừa kết thúc và đối tượng cùng kiểu mới chuẩn bị cấp phát.

### 10.3. AI Semantic Metadata (`.axmeta`)
Sức mạnh lớn nhất của Axiom IR nằm ở khả năng xuất (export) metadata cho các công cụ Trí tuệ Nhân tạo.
*   Khi compile với cờ `--emit-meta`, trình biên dịch đóng gói AST đã phân giải ngữ nghĩa và Đồ thị kết nối (Connection Graph) thành một tệp `.axmeta` dạng JSON.
*   **Khả năng Suy luận của AI:** Thay vì một con bot Copilot phải "đoán" logic từ văn bản nguồn, nó sẽ đọc trực tiếp `.axmeta`. File này chỉ cho AI biết đích xác: Biến này trỏ đi đâu, vùng nhớ nằm ở Heap hay Stack, và hàm này có Effect ngoại vi (I/O, database) nào không. Quá trình sinh mã (Code Generation) và tái cấu trúc (Refactoring) của AI vì thế trở nên chuẩn xác toán học.

### 10.4. Portability (Tính di động)
Nhờ sử dụng C-Backend làm mục tiêu trung gian trong giai đoạn thiết kế MVP, IR của Axiom được đảm bảo tính di động tối thượng. Mọi thiết bị, từ hệ thống nhúng (Embedded microcontrollers) tới máy chủ Linux hay Windows, miễn là có trình biên dịch C (chuẩn C99/C11 trở lên), đều có thể chạy Axiom với hiệu năng gốc (native performance).

## 11. EXECUTABLE FORMAT (ĐỊNH DẠNG THỰC THI)

Để đảm bảo tính tương thích với các hệ điều hành hiện tại nhưng vẫn mang trong mình sức mạnh của hệ sinh thái Axiom, trình biên dịch xuất ra các định dạng chuẩn (ELF cho Linux, PE cho Windows, Mach-O cho macOS) nhưng mở rộng chúng thông qua các phân vùng (sections) tùy chỉnh.

### 11.1. Binary Layout (Bố cục Nhị phân)
*   **`.text` / `.data` / `.rodata`**: Chứa mã máy gốc (native machine code) và dữ liệu tĩnh như C/C++.
*   **`.axmeta` (Axiom Metadata)**: Phân vùng đặc biệt chứa siêu dữ liệu (metadata) của chương trình dưới định dạng nhị phân tinh gọn (BSON/MessagePack). 
*   **`.axsig` (Axiom Signatures)**: Phân vùng chứa chữ ký mật mã (cryptographic signatures) phục vụ cho quá trình xác thực tính toàn vẹn (Integrity Verification).

### 11.2. Metadata & Runtime Metadata (Siêu dữ liệu)
*   Khác với C++ chỉ giữ lại thông tin RTTI (Run-Time Type Information) hạn chế, Axiom lưu trữ toàn bộ cấu trúc kiểu (Type definitions), Traits, và Memory Regions vào `.axmeta`.
*   Tại thời gian chạy (runtime), API `std.reflect` cho phép truy cập các siêu dữ liệu này với chi phí $O(1)$ để phục vụ cho Serialization (như JSON) hoặc Dependency Injection.

### 11.3. Encrypted Sections (Phân vùng Mã hóa - Chế độ Hardened)
*   Khi biên dịch với cờ `--hardened`, phân vùng `.text` sẽ được mã hóa tĩnh.
*   Trình biên dịch tự động nhúng một Micro-Loader (Trình nạp siêu nhỏ) vào tệp thực thi. Khi hệ điều hành khởi chạy chương trình, Micro-Loader này sẽ giải mã Just-In-Time (JIT) đoạn mã `.text` trực tiếp trên RAM và xóa khóa giải mã ngay sau đó để ngăn chặn hành vi dump bộ nhớ.

### 11.4. Symbol Model (Mô hình Ký hiệu)
*   Axiom sử dụng **Mangled Symbols** (Ký hiệu được mã hóa tên) chứa đầy đủ thông tin về Không gian tên (Namespace), Kiểu dữ liệu (Types), và Hiệu ứng (Effects). Ví dụ: `_AX_std_net_fetch_!IO_@async`.
*   Hỗ trợ `extern "C"` để vô hiệu hóa mangling, đảm bảo tương thích ABI hai chiều (bidirectional interoperability) tuyệt đối với C/C++,.

### 11.5. Integrity Verification (Xác thực Toàn vẹn)
*   Phân vùng `.axsig` chứa một Cây Merkle (Merkle Tree) băm toàn bộ các block của file nhị phân. Khi ứng dụng khởi động, Axiom Runtime có thể tùy chọn quét cây Merkle này để đảm bảo file chưa bị chỉnh sửa (tampered) bởi mã độc.

---

## 12. SECURITY MODEL (MÔ HÌNH BẢO MẬT)

Axiom thiết lập một ranh giới bảo mật nghiêm ngặt từ lúc biên dịch cho đến lúc thực thi.

### 12.1. Sandboxing & Permissions (Cơ chế Hộp cát và Quyền hạn)
*   Axiom áp dụng **Capability-based Security** (Bảo mật dựa trên năng lực) ở cấp độ Module.
*   Một module không thể tự ý gọi I/O hoặc truy cập mạng. Nó phải yêu cầu quyền tĩnh: `import std.fs { read, write }`. Trình biên dịch theo dõi các quyền này trên toàn bộ Đồ thị Cú pháp (AST) và có thể chặn quá trình build nếu phát hiện module bên thứ 3 (third-party) cố tình thực hiện các tác vụ không được khai báo.

### 12.2. Secure Modules & Signed Packages (Gói bảo mật và Ký điện tử)
*   Hệ thống Package Manager của Axiom bắt buộc mọi thư viện (packages) phải có file `axiom.lock` chứa mã băm SHA-256.
*   Trình biên dịch sẽ từ chối tải bất kỳ module nào có mã băm không khớp với chữ ký đã được tác giả ký điện tử, loại bỏ hoàn toàn các cuộc tấn công chuỗi cung ứng (Supply Chain Attacks).

### 12.3. Anti-Tamper & Anti-Debug (Chống can thiệp và Gỡ lỗi)
*(Tính năng dành cho cấu hình Release-Hardened)*
*   Trình biên dịch tự động chèn các lệnh kiểm tra thời gian thực thi (Timing Checks, ví dụ dùng `RDTSC`) vào giữa các khối Basic Blocks. Nếu phát hiện thời gian thực thi bị trễ bất thường (dấu hiệu của việc sử dụng Debugger như GDB), tiến trình sẽ tự động sập (panic) một cách an toàn.
*   Tránh sử dụng `libc` để thực hiện các cuộc gọi hệ thống nhạy cảm, thay vào đó sử dụng **Direct Syscalls** để qua mặt các kỹ thuật móc hàm (API Hooking) của công cụ bẻ khóa.

---

## 13. AI-NATIVE FEATURES (TÍNH NĂNG TRÍ TUỆ NHÂN TẠO NỘI TẠI)

Axiom là ngôn ngữ đầu tiên trên thế giới coi các mô hình Trí tuệ Nhân tạo (LLMs, AI Agents) không chỉ là công cụ sinh mã, mà là "công dân hạng nhất" tham gia vào quá trình đọc hiểu và tối ưu hóa codebase.

### 13.1. Token-Efficient Syntax (Cú pháp tối ưu Token cho AI)
*   Sử dụng cú pháp dựa trên thụt lề (Indentation-based), loại bỏ ngoặc nhọn `{}` và dấu chấm phẩy `;` thừa thãi.
*   Giảm thiểu lượng token LLM cần phân tích (context window) lên tới 30% so với C++ hoặc Rust, giúp AI suy luận nhanh hơn và ít bị ảo giác (hallucination) hơn.

### 13.2. Semantic AST & Executable Knowledge Graph (Đồ thị Tri thức Thực thi)
*   Thông qua lệnh `axc export-meta`, trình biên dịch xuất ra một tệp JSON thể hiện toàn bộ **Đồ thị Tri thức (Knowledge Graph)** của chương trình.
*   AI có thể truy vấn đồ thị này (Ví dụ: *"Biến X này có thoát khỏi hàm không?"*, *"Hiệu ứng phụ của hàm Y là gì?"*) để hiểu hoàn hảo ngữ nghĩa (semantic) của mã nguồn thay vì chỉ đọc text tuyến tính. Tính tất định (deterministic parsing) của Axiom đảm bảo AI không bao giờ hiểu sai ý định của lập trình viên.

### 13.3. Self-documenting Metadata (Siêu dữ liệu tự tài liệu hóa)
*   Mọi comment bắt đầu bằng `///` (Docstrings) được Trình biên dịch gắn trực tiếp vào Node tương ứng trên AST. AI có thể sử dụng dữ liệu này để tự động sinh tài liệu chuẩn xác hoặc giải thích mã cho kỹ sư con người.

### 13.4. AI Optimization Hooks (Điểm móc tối ưu hóa cho AI)
*   Lập trình viên có thể sử dụng Annotation `#[ai_suggest]`. Trình biên dịch Axiom (hoặc LSP) sẽ gửi đoạn mã đó cho AI cục bộ, AI sẽ đọc Đồ thị bộ nhớ và đề xuất biến đổi cấu trúc (ví dụ: chuyển từ Array of Structures - AoS sang Structure of Arrays - SoA,) để tối ưu hóa CPU Cache mà không làm vỡ logic.

---

## 14. QUANTUM EXTENSION (MỞ RỘNG LƯỢNG TỬ - DỰ PHÓNG TƯƠNG LAI)

Để đảm bảo tính tiến hóa dài hạn (long-term evolution), Axiom thiết kế sẵn không gian ngữ nghĩa cho các hệ thống máy tính lai Cổ điển - Lượng tử (Classical-Quantum Hybrid).

### 14.1. Qubit Abstraction (Trừu tượng hóa Qubit)
*   Kiểu dữ liệu nguyên thủy `qbit`. Dựa trên định lý "Không sao chép" (No-cloning theorem) của vật lý lượng tử, `qbit` là kiểu dữ liệu bị khóa hoàn toàn vào cơ chế Move-semantics (Chỉ di chuyển, cấm Copy) tương tự các tài nguyên tuyến tính (Linear Types) được quản lý bởi Single Ownership.

### 14.2. Quantum IR (OIR-Q)
*   Biểu diễn trung gian của Axiom (AIR) có một tập lệnh mở rộng `OIR-Q`. Khác với các thanh ghi tĩnh `%r`, `OIR-Q` sử dụng thanh ghi lượng tử `%q`. Trình biên dịch sẽ tách (split) các đoạn mã chứa `%q` để biên dịch riêng ra OpenQASM hoặc các định dạng lượng tử phần cứng.

### 14.3. Probabilistic Semantics (Ngữ nghĩa Xác suất)
*   Các kết quả từ hàm đo lường lượng tử (`measure`) không trả về kiểu `bool` hay `int` ngay lập tức, mà trả về kiểu `Superposition[T]` (Chồng chập).
*   Lập trình viên bị ép buộc phải sử dụng Pattern Matching (`match`) để "làm sụp đổ" (collapse) trạng thái này về giá trị tĩnh trước khi giao tiếp với logic Cổ điển.

### 14.4. Hybrid Execution (Thực thi Lai)
*   Môi trường thực thi (Runtime) coi Bộ xử lý lượng tử (QPU) như một Actor độc lập. Các luồng CPU sẽ giao tiếp với QPU Actor thông qua Message Passing bất đồng bộ, loại bỏ tình trạng thắt cổ chai (bottleneck) giữa hai hệ thống.

---

## 15. STANDARD LIBRARY (THƯ VIỆN CHUẨN - `std`)

Thư viện chuẩn của Axiom được thiết kế với triết lý "Pin đi kèm" (Batteries-included) nhưng hoàn toàn tuân thủ Zero-cost Abstractions và Data-Oriented Design.

### 15.1. Collections (`std.collections`)
*   Bao gồm `List[T]`, `Map[K, V]`, `Set[T]`.
*   Mọi collection đều nhận thức được bộ nhớ (memory-aware) thông qua Generational References. Các phần tử được xếp liên kề (contiguous) trên bộ nhớ mặc định để tối ưu Cache. Hỗ trợ thuộc tính `@SOA` để tự động xoay trục bộ nhớ.

### 15.2. Concurrency (`std.concurrency`)
*   Sử dụng mô hình Actor siêu nhẹ mượn ý tưởng từ Erlang/Elixir, kết hợp với Cơ chế Đồng thời có Cấu trúc (Structured Concurrency).
*   Các hàm cốt lõi: `spawn`, `awaitAll`, và cơ chế giao tiếp an toàn lock-free qua `Channel[T]`.

### 15.3. Filesystem & Networking (`std.fs`, `std.net`)
*   Thiết kế hoàn toàn Bất đồng bộ (Asynchronous by default). Mọi lệnh gọi hệ thống như `fs.read()` hoặc `net.tcp_listen()` đều tự động nhường luồng (yield) trả lại cho Scheduler thông qua State Machines ẩn mà không block luồng hệ điều hành.

### 15.4. Crypto (`std.crypto`)
*   Tích hợp sẵn các thuật toán mã hóa hiện đại (AES-GCM, ChaCha20, SHA-3) sử dụng các nhân (kernels) tăng tốc bằng SIMD/AVX-512 tự viết bằng Axiom mà không phụ thuộc vào OpenSSL.

### 15.5. GPU Primitives (`std.gpu`)
*   Lấy cảm hứng từ Mojo,, thư viện chuẩn cho phép định nghĩa các Kernel chạy trực tiếp trên GPU. Dữ liệu mảng `Array[T]` trên CPU có thể được gửi thẳng qua vùng nhớ GPU thông qua Zero-copy memory truyền con trỏ cô lập (`Isolated[T]`).

### 15.6. AI Primitives (`std.ai`)
*   Cung cấp kiểu `Tensor[T]` là công dân hạng nhất (first-class citizen).
*   Tích hợp các hàm tối ưu hóa cho Tính toán Ma trận (Matrix Multiplication), Hỗ trợ Auto-differentiation (Tự động đạo hàm) ở cấp độ cú pháp thông qua Metaprogramming, lúc biên dịch để phục vụ cho kỷ nguyên xây dựng Mô hình Máy học hiệu năng cao trực tiếp bằng Axiom.

## 16. TOOLCHAIN (CHUỖI CÔNG CỤ)

Chuỗi công cụ của Axiom được thiết kế nguyên khối (monolithic binary) mang tên `axc`, nhằm tránh sự phân mảnh của hệ sinh thái (như C++ với Make, CMake, Ninja, vcpkg). 

### 16.1. Rationale (Cơ sở lý luận)
Một ngôn ngữ lập trình hiện đại cần một bộ công cụ chuẩn hóa ngay từ Ngày 1 để loại bỏ các cuộc tranh luận không cần thiết (bike-shedding) về cấu trúc dự án, định dạng mã, và quản lý gói. Đặc biệt, việc tích hợp Language Server trực tiếp vào trình biên dịch đảm bảo độ chính xác 100% của Đồ thị Ngữ nghĩa (Semantic Graph) khi xuất ra cho AI.

### 16.2. Examples (Ví dụ)
*   **Compiler CLI:** `axc build main.ax -O3 --target=linux-x64`
*   **Package Manager:** `axc get github.com/user/lib@1.2.0` (Tự động cập nhật `axiom.toml` và `axiom.lock`).
*   **Formatter:** `axc fmt .` (Định dạng toàn bộ dự án).
*   **Profiler:** `axc profile main.ax` (Chạy và xuất ra file `flamegraph.svg`).
*   **Language Server:** `axc lsp --emit-meta` (Khởi chạy giao thức LSP cho VSCode/Cursor, đồng thời cung cấp payload JSON `.axmeta` cho AI Agent).

### 16.3. Constraints (Ràng buộc kỹ thuật)
*   **Trình định dạng (Formatter):** Zero-configuration (Không cho phép cấu hình). `axc fmt` sử dụng một bộ quy tắc AST duy nhất (ví dụ: thụt lề bắt buộc 4 khoảng trắng, giới hạn 100 ký tự/dòng).
*   **Trình gỡ lỗi (Debugger):** Ở giai đoạn MVP (sử dụng C-Backend), `axc` phải tự động chèn `#line` directives vào mã C sinh ra để ánh xạ ngược (map back) breakpoint về tệp `.ax` gốc, đảm bảo tương thích với GDB/LLDB.

### 16.4. Trade-offs (Đánh đổi)
*   **Kích thước file nhị phân toolchain:** Việc gộp Compiler, LSP, Profiler và Package Manager vào một tệp thực thi duy nhất làm tăng kích thước `axc` (~50MB). Tuy nhiên, đổi lại là trải nghiệm cài đặt "chỉ 1 click" (zero-dependency installation) cho người dùng cuối.

---

## 17. SELF-HOSTING ROADMAP (LỘ TRÌNH TỰ BIÊN DỊCH)

Quá trình tự biên dịch (Self-hosting) là minh chứng kỹ thuật cho độ chín muồi của ngôn ngữ. Lộ trình của Axiom chia làm 3 giai đoạn (Stages) rõ ràng.

### 17.1. Rationale (Cơ sở lý luận)
Việc viết trình biên dịch Axiom bằng chính Axiom ngay lập tức là bất khả thi. Mượn chiến lược từ Go và V, chúng ta cần một Trình biên dịch mồi (Bootstrap Compiler) bằng ngôn ngữ hiện có (như Go/C), sau đó dịch dần sang Axiom để tận dụng chính các tính năng an toàn bộ nhớ của nó.

### 17.2. Examples (Các giai đoạn)
*   **Stage 0 (Bootstrap Compiler):** Viết bằng Go (do AI sinh mã Go rất tốt). Nó chỉ chứa Lexer, Parser và C-Backend cơ bản. Chỉ biên dịch được tập con (subset) của Axiom.
*   **Stage 1 (Self-hosted via C):** Trình biên dịch Axiom được viết hoàn toàn bằng Axiom (file `.ax`). Dùng Stage 0 để dịch mã nguồn này ra file `axc.c`. Dùng GCC/Clang biên dịch `axc.c` thành `axc` nhị phân.
*   **Stage 2 (Native Self-hosted):** Trình biên dịch `axc` (từ Stage 1) tự biên dịch chính mã nguồn `.ax` của nó ra mã máy x86-64/ARM thuần túy (thông qua Axiom IR), không cần GCC/Clang nữa. Đạt tốc độ build > 1 triệu dòng/giây.

### 17.3. Constraints (Ràng buộc kỹ thuật)
*   Mã nguồn của Trình biên dịch (Stage 1) không được sử dụng các tính năng thử nghiệm (experimental) của ngôn ngữ để đảm bảo quá trình bootstrap có thể lặp lại (reproducible builds) trên mọi hệ điều hành.

### 17.4. Trade-offs (Đánh đổi)
*   Duy trì kiến trúc C-Backend ở Stage 1 tốn thêm nỗ lực bảo trì bộ AST-to-C, nhưng bù lại mang đến khả năng chạy Axiom trên mọi kiến trúc phần cứng (kể cả chip nhúng) mà không cần viết bộ cấp phát thanh ghi (Register Allocator) riêng cho từng chip.

---

## 18. MVP SPEC (ĐẶC TẢ SẢN PHẨM KHẢ THI TỐI THIỂU)

Để dự án được hoàn thành bởi 1 kỹ sư và sự trợ giúp của AI trong 12-15 tháng, thiết kế cần sự tàn nhẫn trong việc cắt giảm phạm vi (Scope reduction).

### 18.1. Rationale (Cơ sở lý luận)
Nguyên tắc Pareto (80/20): 20% tính năng ngôn ngữ giải quyết 80% nhu cầu thực tế. Cần phân định rõ ranh giới để ngăn chặn "Feature Creep" (lan man tính năng).

### 18.2. Examples (Phân vùng tính năng)
*   **BẮT BUỘC IMPLEMENT ĐẦU TIÊN (Must-have):**
    *   Cú pháp thụt lề (Indentation grammar), Flat AST.
    *   Memory Model: Single Ownership (`mut`, `!T`) và Generational References.
    *   C-Backend Codegen.
    *   Siêu lập trình nội tuyến cơ bản (`#run` cho constant folding).
*   **TRÌ HOÃN (Postponed to v2.0):**
    *   Native x86-64 backend (chưa cần, dùng GCC ẩn dưới nền).
    *   Hệ thống Actor phân tán qua mạng LAN (chỉ làm Actor đa luồng cục bộ trước).
    *   Quantum Subsystem (QPU IR).
*   **THỬ NGHIỆM (Experimental trong MVP):**
    *   Thu gom bộ nhớ lúc biên dịch (Compile-time GC - CTGC). Nếu CTGC phân tích đồ thị quá phức tạp ở MVP, compiler được phép chèn `free()` vào cuối scope (như Destructor bình thường).

### 18.3. Constraints (Ràng buộc kỹ thuật)
*   Chỉ tập trung hỗ trợ nền tảng POSIX (Linux/macOS) ở MVP để giảm thiểu mã tương thích OS. Windows sẽ hỗ trợ thông qua WSL/MinGW ban đầu.

### 18.4. Trade-offs (Đánh đổi)
*   **Thời gian chạy (Runtime performance) vs. Thời gian phát triển (Dev time):** Việc dùng C-Backend có thể sinh ra mã C không đẹp (unidiomatic C), nhưng nó đảm bảo MVP ra mắt đúng hạn. Tối ưu mã máy cấp thấp (Low-level optimization) được nhường lại cho GCC `-O3`.

---

## 19. FORMAL SEMANTICS (NGỮ NGHĨA HÌNH THỨC)

Để trình biên dịch và AI có cùng một cách hiểu tuyệt đối (deterministic) về Axiom, ngôn ngữ được định nghĩa bằng các tiên đề toán học (Axiomatic semantics).

### 19.1. Rationale (Cơ sở lý luận)
Sự mơ hồ (Ambiguity) trong C++ (như Undefined Behavior khi tràn bộ đệm) là gốc rễ của lỗ hổng bảo mật. Axiom từ chối UB (Undefined Behavior). Mọi trạng thái phải được định nghĩa bằng ngữ nghĩa hoạt động (Operational Semantics).

### 19.2. Examples (Các mô hình ngữ nghĩa)

*   **Execution Semantics (Ngữ nghĩa thực thi):**
    *   Việc gán biến tuân theo Semantics Dòng dữ liệu (Dataflow).
    *   Ký hiệu: `Eval(expr, Env) => (Value, Env')`.
    *   Nếu `expr` là hàm `#run`, `Eval` được thực thi ngay trong chu kỳ `CompileEnv` thay vì `RuntimeEnv`.
*   **Memory Semantics (Ngữ nghĩa Bộ nhớ):**
    *   Một con trỏ (Pointer) $P$ trong Axiom là một tuple: $P = (Addr, GenID)$.
    *   Một vùng Heap $H$ chứa Metadata: $H[Addr].GenID$.
    *   **Toán tử Dereference (`*P`):**
        ```text
        if P.GenID == H[P.Addr].GenID:
            return Memory[P.Addr]
        else:
            trigger SafePanic("Use-After-Free detected")
        ```
*   **Concurrency Semantics (Ngữ nghĩa Đồng thời):**
    *   Sử dụng đại số Actor (Actor Algebra). Một Actor $\alpha$ là tập hợp $(State, Mailbox, Behavior)$.
    *   Gửi thông điệp $M$: Đồ thị sở hữu (Ownership graph) của $M$ bị xóa khỏi $\alpha_1$ và chuyển sang $Mailbox$ của $\alpha_2$. Quá trình này là Atomic (nguyên tử) ở mức con trỏ.
*   **Error Semantics (Ngữ nghĩa Lỗi):**
    *   Không có `try/catch` ẩn. Lỗi là một giá trị (Errors as values) được ánh xạ qua Sum Types (`Result<T, E>`).

### 19.3. Constraints (Ràng buộc kỹ thuật)
*   Trình biên dịch phải báo lỗi tĩnh (Compile-time Error) nếu phát hiện luồng logic có thể dẫn đến việc đọc một biến chưa được khởi tạo (Uninitialized memory is forbidden).

### 19.4. Trade-offs (Đánh đổi)
*   **Tính tất định vs Hiệu năng cực hạn:** Việc ép buộc kiểm tra `GenID` trong Memory Semantics loại bỏ hoàn toàn rủi ro bảo mật (Memory Safety), nhưng tiêu tốn thêm 1 chu kỳ CPU cho mỗi lần giải tham chiếu (Dereference). Trong các vòng lặp game/toán học, lập trình viên phải đổi sang `Arena` (Region memory) để tắt kiểm tra này, đánh đổi sự an toàn toàn cục lấy hiệu năng $O(1)$.

---

## 20. APPENDIX (PHỤ LỤC)

Phần này cung cấp các ví dụ thực tiễn, mã giả và các ghi chú triển khai dành riêng cho đội ngũ phát triển Core Compiler.

### 20.1. Rationale (Cơ sở lý luận)
Các đặc tả trừu tượng cần được minh họa bằng các đoạn mã cụ thể để kiểm chứng cú pháp và đảm bảo tính khả thi khi phân tích cây cú pháp (AST Parsing).

### 20.2. Examples (Ví dụ & Mã giả)

**A. Ví dụ Cú pháp (Syntax Example): Thuật toán đếm từ song song**
```axiom
import std.fs
import std.concurrency as sync

pub async fn count_words(filepath: string) -> int:
    // 'mut' báo hiệu quyền sở hữu và khả năng đột biến
    mut file := await fs.open(filepath)
    defer file.close() // Higher RAII: Chắc chắn gọi khi thoát scope

    mut count := 0
    for line in file.read_lines():
        count += line.split(" ").length()
        
    return count

fn main():
    // Khởi tạo Actor con thông qua 'spawn' và chờ kết quả
    let result_a = spawn count_words("data1.txt")
    let result_b = spawn count_words("data2.txt")
    
    // Đảm bảo không có tác vụ mồ côi (Structured Concurrency)
    await_all:
        echo "Total: " + (await result_a + await result_b)
```

**B. Mã giả triển khai Generational Reference (Compiler Implementation Note)**
Khi biên dịch đoạn mã Axiom sử dụng con trỏ an toàn, C-Backend sẽ sinh ra cấu trúc C như sau:
```c
// Mã C được sinh ra bởi Axiom Compiler
typedef struct {
    void* addr;
    uint64_t gen_id;
} AxRef;

typedef struct {
    uint64_t current_gen;
    // ... data ...
} AxHeapMeta;

// Hàm giải tham chiếu ngầm định
static inline void* ax_deref(AxRef ref) {
    AxHeapMeta* meta = (AxHeapMeta*)((char*)ref.addr - sizeof(AxHeapMeta));
    if (meta->current_gen != ref.gen_id) {
        ax_panic("Use after free!");
    }
    return ref.addr;
}
```

### 20.3. Constraints (Ràng buộc kỹ thuật)
*   **Generics Instantiation:** Để tránh hiện tượng "Code Bloat" (phình to mã nhị phân) do Monomorphization (Tạo bản sao hàm Generic cho từng kiểu), linker nội bộ của AxiomC phải có Pass gộp các hàm Generic có mã máy sinh ra giống hệt nhau (ví dụ: `List[PointerA]` và `List[PointerB]` dùng chung mã).

### 20.4. Trade-offs & Future Evolution (Đánh đổi và Tiến hóa tương lai)
*   **Trade-off hiện tại:** Axiom MVP bỏ qua tính năng đa kế thừa (Multiple Inheritance) và các khái niệm OOP truyền thống. Điều này khiến việc chuyển đổi trực tiếp mã kiến trúc GUI cũ từ C++ sang Axiom gặp khó khăn, yêu cầu phải thiết kế lại theo mô hình Component (ECS - Entity Component System).
*   **Tiến hóa tương lai (v3.0+):** Khi LLM đủ thông minh, Trình biên dịch Axiom sẽ mở API `std.compiler.mutate`. Một mô hình AI cục bộ có thể lắng nghe các cảnh báo hiệu năng (Profiler feedback), tự động thay đổi cấu hình `Region/Arena` trong mã gốc, và build lại file thực thi tối ưu nhất theo thời gian thực. Hướng tới một hệ điều hành nơi phần mềm liên tục "tự chữa lành" và "tự tiến hóa" dựa trên phần cứng (Self-evolving software).