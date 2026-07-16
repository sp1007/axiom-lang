# RFC 0014 — `drop(self)` hook cho CTGC (drop glue tối giản)

- **Status:** ✅ **Implemented** (2026-07-16) — drop-glue live behind `-ctgc-free`
  (opt-in; self-host builds declare no `drop`, so they are byte-identical / A==B).
  `resolve_drop_method` (air_builder.ax) + `lower_destroy` call `Type.drop(self)` before
  freeing the block, for every non-escaping owned local whose type declares `drop`.
  Oracle `bin/t_drop.ax` (drop fires 42× with the flag, 0 without); validated on a drop
  corpus (single-local fires once, escaping/aliased instances never dropped -> no
  double-drop) + `scripts/ctgc_free_check.sh`. Two bugs were fixed en route: (1) BLOCKER
  from 2026-07-06 (§8) — CTGC/Escape were no-ops — resolved by RFC 0015 P2/P3; (2) a
  latent `lower_destroy` reg-lookup bug (used `sym.name_id`, but the local_map is keyed
  by sym_idx) that had silently discarded every CTGC free. **Scope note:** the FREE is
  emitted only for `drop`-typed locals (the memory-reclaim half of drop-glue); RFC 0015's
  GENERAL free of all non-escaping owned locals stays deferred — it is not yet sound on
  the alias-heavy self-host compiler (RFC 0010 §9; a `-ctgc-free` self-build UAF-crashed
  once free was made real). Drop types are user-declared and simple, so freeing exactly
  them is sound (self-build with `-ctgc-free` reproduces the fixpoint, since the compiler
  has no drop types).
- **Author:** self-host team
- **Tracking:** giải quyết [[bignum-ctgc-conflict]] (rò rỉ heap thật trong `std.bignum`)
- **Liên quan:** ctgc.ax (`CtgcInjector`), air_builder.ax (`lower_destroy`, `resolve_op_method`
  — RFC 0007 operator overloading), ownership.ax
- **Blocks:** không gì trong pipeline hiện tại phụ thuộc RFC này; đây là tính năng CỘNG THÊM
  (opt-in), không đổi hành vi hiện có cho struct không định nghĩa `drop`.

---

## 1. Motivation

CTGC (`ctgc.ax::traverse_and_inject`) chèn `NODE_DESTROY_STMT` cuối scope cho biến
struct/generic-inst/sum/option/result không escape. `lower_destroy` (air_builder.ax:3083)
lower thẳng thành `OP_DESTROY`, và **mọi backend** (cgen.ax:1090, x86_selector.ax:1634,
wasm.ax:458) đều hiện thực giống nhau: `ax_free(reg)` — free **ĐÚNG MỘT** khối, là khối
chứa chính instance đó. Cơ chế này KHÔNG có khái niệm "field bên trong cũng sở hữu heap".

`std.bignum`'s `BigUint = {limbs: ptr[u64], n: i64}` là ví dụ thật: CTGC free đúng khối
16-byte header của `BigUint`, nhưng **buffer `limbs` (cấp bởi `@alloc` riêng) không bao giờ
được free** — không có `bu_free`/`destroy` nào trong `std/bignum.ax` (đã grep, rỗng). Mọi
phép toán trả struct-mới (`bu_shl`, `bf_mul`, ...) rò rỉ buffer trung gian. Đây là leak
THẬT (không phải use-after-free hay sai kết quả), nhưng vi phạm §11 CLAUDE.md ("Allocator
phải support stress testing... deterministic behavior") — allocator torture test cho
bignum sẽ OOM dần dần, không bao giờ ổn định.

## 2. Ngữ nghĩa hiện tại vs đề xuất

**Hiện tại:** không có cách nào để một struct chạy code tùy ý khi bị CTGC destroy. Field
kiểu `ptr[T]` được coi như dữ liệu bất kỳ (giống `i64`) — CTGC không phân biệt "con trỏ sở
hữu" (owning) với "con trỏ view" (borrow, vd `let p = &arr[i]`).

**Đề xuất:** một struct có thể định nghĩa method `fn drop(self):` (self by-value, giống
các operator-overload method RFC 0007 — không tham số khác, không giá trị trả về). Nếu có,
`lower_destroy` gọi `drop(self)` **TRƯỚC KHI** emit `OP_DESTROY` free khối instance. Struct
KHÔNG định nghĩa `drop` giữ nguyên hành vi hiện tại (chỉ free khối riêng, byte-identical).

```axiom
struct BigUint:
    limbs: ptr[u64]
    n: i64

    fn drop(self):
        @free(self.limbs as ptr[u8])
```

Không có "owned field" annotation, không auto field-walking, không topo-order. Người dùng
viết CHÍNH XÁC những gì struct sở hữu — giống Rust's `Drop`/C++ destructor, nhưng tối giản
hơn (không trait, không gọi đệ quy tự động cho field kiểu struct khác — nếu cần, người dùng
tự gọi `self.child.drop()` trong thân `drop`).

## 3. Vì sao KHÔNG chọn "owned field annotation + auto free"

Phương án thay thế (đánh dấu field `ptr[T]` là `owned`, compiler tự sinh glue free đệ quy
mọi field owned, kể cả struct lồng nhau) được cân nhắc và BÁC BỎ cho iteration đầu:

1. **Topo-order phức tạp ngay từ đầu**: struct lồng struct-có-owned-field cần graph phân
   tích thứ tự free — đúng loại rủi ro CLAUDE.md §19 cảnh báo ("avoid overly generic designs
   early"). RFC 0011 P4 đã gặp đúng vấn đề "topo-order cho nested struct" ở một ngữ cảnh
   khác (interface serialization) và vẫn đang hoãn.
2. **Không rõ ngữ nghĩa cho con trỏ KHÔNG sở hữu**: rất nhiều `ptr[T]` trong codebase hiện
   tại là view/borrow (`let p = &arr[i]`, tham số `self: ptr[T]` của method). Suy luận
   "owned vs borrow" từ kiểu khai báo đơn thuần là SAI hầu hết các trường hợp (RFC 0010's
   audit đã cho thấy phân biệt alias-vs-copy tinh vi hơn tưởng, dẫn tới regression tự-host
   khi đoán sai). Một annotation tường minh (`owned ptr[T]`) giải quyết được, nhưng đó là
   thay đổi cú pháp — cần RFC cú pháp riêng, phạm vi lớn hơn.
3. **`drop(self)` tường minh không cần đoán gì cả** — không owned/borrow inference, không
   cú pháp mới, tái dùng NGUYÊN VẸN cơ chế method-resolution đã có (RFC 0007). Đủ để giải
   quyết bignum ngay, và là nền tảng an toàn nếu sau này muốn thêm auto-glue (method `drop`
   tự sinh có thể gọi glue tự động cho field owned — bổ sung sau, không chặn bởi RFC này).

## 4. Design — implementation tối thiểu

**Method resolution:** tái dùng CHÍNH XÁC cơ chế `resolve_op_method` (air_builder.ax:830):
substring-match trên tên mangled (`match_mangled_method_raw_bytes`) để tìm `<Type>.drop`
kể cả khi khai báo INLINE trong struct (như `Num.eq`/`Num.lt` ở t_opover.ax) — vì method
inline mang tên dotted mangled, KHÔNG resolve được qua `resolve_method_sym` (bài học
BUG#68, xem [[bug68-struct-eq-no-overload]]). Viết `resolve_drop_method(rec_type) -> u32`
(0 nếu không có), signature khác `resolve_op_method` chỉ ở chỗ không cần `rhs_type` (drop
không có tham số thứ hai) — check param count == 1 (chỉ `self`) thay vì 2.

**Lowering** (`lower_destroy`, air_builder.ax:3083): trước khi emit `OP_DESTROY` (cả hai
nhánh — `node.first_child != 0` và nhánh `sym_id`/`reg`), nếu `resolve_drop_method(type of
reg) != 0`: emit `OP_CALL` tới method đó với `self=reg` làm argument, TRƯỚC OP_DESTROY hiện
tại. Method `drop` chạy XONG rồi mới free khối instance — thứ tự đúng (drop có thể đọc field
trước khi khối bị free).

**Diagnostics (không phải phần lõi, nhưng rẻ và nhất quán với BUG#53/#64/#68):** không cần
— khác BUG#68 (nơi thiếu overload dẫn tới SO SÁNH SAI ngầm), ở đây thiếu `drop` chỉ có nghĩa
"không có gì để free thêm", hành vi hiện tại (đúng, không đổi) — không có "miscompile ngầm"
để chặn.

**Không đụng:** CTGC's escape-analysis (`FLAG_ESCAPES_TO_HEAP` check) giữ nguyên — struct
escape thì không destroy (đúng, ownership chuyển đi, callee sẽ destroy) — `drop` cũng không
được gọi trong trường hợp đó, đúng ngữ nghĩa (escape = ownership moved).

## 5. Rủi ro & câu hỏi mở

1. **Move vào struct khác / return / truyền tham số**: nếu instance có `drop` bị MOVE (trả
   về, gán cho biến khác còn sống) thay vì thực sự chết ở cuối scope, CTGC's escape-check
   (mục "If it is non-escaping heap variable") đã lo việc này — không được thêm vào
   `active_vars` nếu escape → không destroy → không double-drop. Cần XÁC NHẬN bằng test
   trước khi ship (không giả định — audit rule CLAUDE.md §20).
2. **Copy-by-value rồi cả hai bị destroy → double-free**: đây CHÍNH LÀ vấn đề RFC 0010 đang
   điều tra dở (aggregate `let x = y` là alias hay copy). Nếu hai biến struct trỏ CÙNG
   `limbs` (như BigUint hiện tại, `let b = a` chia sẻ con trỏ) đều có `drop` → double-free
   XUẤT HIỆN NGAY LẦN ĐẦU DÙNG. **RFC này KHÔNG tự giải quyết được** vấn đề aliasing của RFC
   0010 — nó chỉ cấp CƠ CHẾ free; đúng đắn triệt để đòi hỏi RFC 0010 xong trước (hoặc: quy
   ước "struct có `drop` thì cấm copy-by-value ngầm", enforce ở ownership.ax, phạm vi hẹp
   hơn nhiều so với RFC 0010 tổng quát — xem mục 6 P1).
3. **Có thể mất drop nếu nhánh code trả về sớm nằm ngoài traverse hiện tại?** CTGC's
   `NODE_RETURN_STMT` branch đã destroy mọi active_vars trước return — cơ chế sẵn có, không
   cần thay đổi, chỉ cần `lower_destroy` gọi đúng `drop` tại mọi điểm `OP_DESTROY` đã có sẵn
   (không thêm điểm chèn mới).

## 6. Kế hoạch thực hiện (tối thiểu, đúng scope RFC)

- **P0 (RFC này):** chốt thiết kế — `drop(self)` convention, resolve qua cơ chế RFC 0007,
  gọi trước `OP_DESTROY` tại `lower_destroy`.
- **P1 (bắt buộc trước khi bật cho bignum):** enforce hẹp — struct có `drop` method bị CẤM
  copy-by-value ngầm (`let b = a` khi `a`'s type có `drop`) ở `ownership.ax`/typecheck, lỗi
  compile rõ ràng thay vì double-free im lặng (giống pattern reject BUG#53/64/68). Phạm vi
  hẹp hơn RFC 0010 nhiều — chỉ áp dụng cho struct CÓ `drop`, không phải MỌI aggregate.
- **P2:** implement `resolve_drop_method` + lowering; test oracle (`bin/t_drop.ax`: struct
  có `drop` in ra giá trị/side-effect quan sát được qua `ax_printf_local`, đếm số lần gọi
  qua exit code) chạy cả 2 backend (cgen + x86 native).
- **P3:** áp dụng cho `std.bignum` (`BigUint.drop` free `limbs`) — đóng
  [[bignum-ctgc-conflict]] thật sự; allocator torture test cho bignum không còn leak.
- **P4 (tương lai, KHÔNG trong scope RFC này):** nếu cần drop tự động đệ quy cho field
  struct lồng nhau (thay vì gọi tay), cân nhắc mở rộng riêng — không block bởi RFC này.

## 7. Alternatives

- **Owned-field annotation + auto recursive free**: xem mục 3 — bác bỏ cho v1, quá rộng.
- **Không làm gì, chấp nhận leak trong bignum**: bignum là pure library (không ảnh hưởng
  self-host, xem [[next-step-15-selfhost-status]]), nhưng vi phạm §11/§21 CLAUDE.md
  (allocator phải survive stress test) nếu bignum được dùng nặng trong chương trình dài hạn.
  Bác bỏ vì để lại nợ kỹ thuật vô thời hạn.
- **`bu_free` tường minh, không cần cơ chế ngôn ngữ mới**: nhanh nhất, nhưng đặt gánh nặng
  lên MỌI call site (dễ quên gọi — đúng lớp lỗi CLAUDE.md muốn tránh, "hidden runtime
  behavior" kiểu ngược — ở đây là "hidden LEAK behavior" vì thiếu free tường minh mà không
  ai nhắc). `drop(self)` gắn liền lifetime, không thể quên (CTGC tự gọi ở mọi exit point).

## 8. ⛔ PHÁT HIỆN CHẶN ĐƯỜNG: OwnershipChecker/EscapeAnalyser/CtgcInjector là NO-OP (2026-07-06)

Khi thử implement P2 (mục 6) và build oracle `bin/t_drop.ax`, `drop` KHÔNG BAO GIỜ được
gọi — kể cả cho local đơn giản nhất, không copy, không alias. Điều tra bằng debug-print
(`ax_printf_local` tạm trong `ctgc.ax`/`air_builder.ax`, cách an toàn đã dùng nhiều lần
trước) cho thấy `CtgcInjector::traverse_and_inject` **KHÔNG BAO GIỜ VIẾNG THĂM BẤT KỲ NODE
NÀO** — kể cả một `NODE_BLOCK` đơn giản nhất.

**Root cause:** `ctgc.ax` (dòng ~31), `escape.ax` (dòng ~34), `ownership.ax` (dòng ~33) —
CẢ BA đều bắt đầu bằng:
```
if node_idx == 0xffffffff as u32 or node_idx == 0 as u32:
    return
```
`ast.ax::new_ast_tree` (dòng 147-150, comment tường minh: *"Append root program node at
index 0"*) xác nhận: **node index 0 CHÍNH LÀ AST ROOT thật (`NODE_PROGRAM`)**, không phải
sentinel "không có node". Đồng thời `NULL_IDX: u32 = 0` (ast.ax:102) — 0 CŨNG được dùng
làm sentinel "không có child/sibling" cho `first_child`/`next_sibling`. Ba pass trên gọi
entry point bằng `self.traverse_and_inject(0 as u32, 0 as u32)` / `self.check_node(0 as
u32)` / `self.traverse_nodes(0 as u32)` — **guard `node_idx == 0` bắt luôn LẦN GỌI ĐẦU
TIÊN (root thật), return ngay lập tức, không bao giờ xử lý bất kỳ node nào.**

⟹ **OwnershipChecker, EscapeAnalyser, VÀ CtgcInjector — CẢ BA PASS — là NO-OP HOÀN TOÀN**,
rất có thể từ khi được viết. "[Debug] Running Ownership Checker..." / "Running Escape
Analyser..." / "Running CTGC Injector..." trong output driver **ĐÁNH LỪA** — in ra như thể
đang chạy thật nhưng KHÔNG kiểm tra/không đánh dấu/không chèn gì cả. Xác nhận độc lập:
`let b = a; let c = a` (copy 2 lần một struct) compile sạch KHÔNG lỗi (đáng lẽ
OwnershipChecker phải báo `E4001 use of moved value` ở lần dùng thứ 2 theo `check_move`).

**Đây là bug NGHIÊM TRỌNG HƠN toàn bộ RFC 0014** — nó có nghĩa:
1. **KHÔNG CÓ compile-time GC nào đang chạy, cho BẤT KỲ chương trình AXIOM nào, từ trước
   tới giờ.** Mọi struct/sum/option/result/generic-inst local đều RÒ RỈ (không chỉ
   `BigUint.limbs` — bản thân khối chứa struct cũng không bao giờ được free). Toàn bộ
   self-hosting compiler + mọi test 98/98 xanh **CHƯA TỪNG chạy dưới CTGC thật** — tất cả
   thành công hiện tại độc lập với CTGC hoàn toàn.
2. **KHÔNG CÓ ownership/move-checking nào đang chạy** — `mut`/single-owner chỉ là DỰ ĐỊNH
   trên giấy (spec), không được enforce.
3. **KHÔNG CÓ escape analysis nào đang chạy** — `FLAG_ESCAPES_TO_HEAP` không bao giờ được
   set bởi pass thật (dù field này tồn tại và được ĐỌC ở nơi khác — luôn đọc thấy 0/false).

**Đã THỬ NGHIỆM (không commit) sửa guard** (bỏ `or node_idx == 0 as u32`, giữ lại
`0xffffffff` — an toàn về lý thuyết vì mọi lời gọi đệ quy đã tự lọc `!= 0` ở call site
trước khi gọi, guard `==0` bên trong chỉ cần cho lần gọi gốc): fixpoint A==B **vẫn đạt**
(compiler tự-build được), nhưng **regression 0/98 — MỌI build đều lỗi ngay ở
OwnershipChecker** (`error[E4002] cannot assign to immutable variable`, `error[E4001] use
of moved value`) **kể cả `bin/t_drop.ax` tối giản** (vi phạm nằm trong RUNTIME/STDLIB LÕI
luôn được bundle bất kể import gì — `result.ax, alloc.ax, scheduler.ax, runtime.ax, os.ax,
string.ax, io.ax, collections.ax`). Toàn bộ codebase (stdlib lõi + self-hosting compiler)
được viết và tiến hóa ~nhiều tháng **giả định ba pass này không bao giờ chạy** — bật lên
cần rà + sửa (hoặc nới lỏng luật checker cho đúng ngữ nghĩa dự định) trên diện RẤT RỘNG,
không phải một session.

**Quyết định:** REVERT hoàn toàn thử nghiệm (`ownership.ax`/`escape.ax`/`ctgc.ax` về
nguyên trạng qua backup), REVERT luôn code `resolve_drop_method`/lowering trong
`air_builder.ax` (không thể verify đúng — 0 lần thực thi, chỉ chứng minh được
"unreachable" chứ không chứng minh được "correct"). Giữ lại: RFC này (thiết kế `drop(self)`
vẫn đúng, sẵn sàng dùng khi CTGC thật sự chạy), `bin/t_drop.ax` (oracle tài liệu, giống
tinh thần `bin/t_aggcopy.ax` của RFC 0010 — CHƯA đăng ký regression vì hành vi hiện tại
không khớp oracle: mọi giá trị exit=0 do không destroy nào chạy).

**Việc CẦN LÀM TRƯỚC KHI bật lại (không phải phạm vi RFC này — cần RFC/kế hoạch riêng,
lớn hơn nhiều):**
1. Sửa guard 3 file (1 dòng mỗi file) — bản thân fix rất nhỏ, nhưng đây chỉ là ĐIỀU KIỆN
   CẦN.
2. **OwnershipChecker cần rà lại NGỮ NGHĨA trước** — `check_move` hiện đánh dấu MỌI struct
   dùng làm var-decl-init/assign-rhs/call-arg là "moved" vô điều kiện, không phân biệt
   copy hợp lệ (RFC 0010 đang cho thấy struct-copy hiện là NGỮ NGHĨA THỰC, không phải lỗi)
   — bật nguyên trạng sẽ biến MỌI reuse hợp lệ thành lỗi biên dịch giả (false positive
   diện rộng, không phải phát hiện lỗi thật). Cần thiết kế lại quy tắc move trước.
3. **CTGC cần audit an toàn free** — một khi thật sự chạy, sẽ free hàng nghìn struct instance
   chưa từng bị free trước đây; RFC 0010 đã cảnh báo alias-qua-buffer (đọc sau khi mutate
   cùng buffer) mà compiler tự-host DỰA VÀO — free thật có thể biến các alias này thành
   UAF thật (hiện đang "an toàn" chỉ vì free chưa từng chạy).
4. Escape analysis (`escape.ax`) cũng chưa audit độ chính xác (heuristic đơn giản, vd
   NODE_CALL_EXPR luôn coi mọi arg là "escape" — có thể quá bảo thủ hoặc sai ở nhiều case).

⟹ Đề xuất: mở **RFC/bug riêng ("CTGC activation")** theo dõi việc bật lại 3 pass này, tách
biệt khỏi RFC 0014 — RFC 0014 CHỈ nên tiếp tục (P2 thật) SAU KHI CTGC activation xong và
CTGC chạy thật, ổn định trên toàn bộ regression.

## 9. Success criteria

- `resolve_drop_method` dùng CHUNG cơ chế với `resolve_op_method` (không lệch, học từ
  BUG#68) — cùng `match_mangled_method_raw_bytes`.
- Test oracle `t_drop.ax`: struct có `drop` → method chạy ĐÚNG MỘT LẦN mỗi lần biến chết
  (scope-exit VÀ mọi return path), KHÔNG chạy khi biến escape (move/return giá trị).
- P1 enforce: copy-by-value một struct-có-drop bị compile error rõ ràng (không double-free
  im lặng) — test âm `tests/sema/err_drop_copy.ax`.
- `BigUint` gắn `drop` → chương trình lặp bignum-heavy nhiều vòng không tăng RSS (đo bằng
  công cụ hệ điều hành hoặc allocator instrumentation hiện có, xem §11 CLAUDE.md).
- Fixpoint A==B + full regression GREEN trên native driver trước khi commit (backend-level
  change).
