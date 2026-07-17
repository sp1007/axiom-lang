---
name: bug69-ctgc-ownership-escape-noop
description: "BUG#69 ✅ ĐÓNG TRỌN (P1+P2+P3): 3 pass (Ownership/Escape/Ctgc) từng no-op do guard node_idx==0. P1 mutability-only SHIPPED 5c50011. P2 (Escape ACTIVE) SHIPPED f06d939 — phá án 'non-determinism period-2' của 2026-07-12 là SAI, thực chất CRASH (wild-free run() do save/restore ConnectionGraph value-semantics trên aggregate-alias); fix→A==B, mark SYM_FLAG_ESCAPES=4096. P3 (CTGC free) SHIPPED 038c2ea 2026-07-16 — opt-in -ctgc-free (default OFF→self-host A==B): ctor-detection + scalar-field flow refinement (node_types) + fall-through-only OP_DESTROY. Validate: 13-prog corpus + compiler tự-build với -ctgc-free chạy đúng (6 self-free). Còn niche: return-path leak, Option-ctor ko free. Mở khóa RFC 0014."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#69 — phát hiện 2026-07-06, khi điều tra tại sao RFC 0014 (`drop(self)` hook cho
CTGC, xem `rfcs/0014-drop-glue.md`) không hoạt động dù implement đúng thiết kế.**

## Triệu chứng ban đầu

`bin/t_drop.ax` (struct có `fn drop(self)`, dùng làm local, không copy/alias) — `drop`
KHÔNG BAO GIỜ được gọi. Debug-print (`ax_printf_local` tạm, cách an toàn đã dùng nhiều lần
trước — xem [[bug67-option-struct-payload-unwrap]]) đặt ngay đầu `CtgcInjector::
traverse_and_inject` cho thấy: **hàm này không bao giờ xử lý bất kỳ node nào, kể cả một
`NODE_BLOCK` đơn giản nhất trong chương trình đơn giản nhất.**

## Root cause

`bootstrap/stage1/ctgc.ax` (dòng ~31), `escape.ax` (dòng ~34), `ownership.ax` (dòng ~33) —
**CẢ BA** pass bắt đầu hàm traverse chính bằng:
```
if node_idx == 0xffffffff as u32 or node_idx == 0 as u32:
    return
```
Nhưng `ast.ax::new_ast_tree` (dòng 147-150, comment tường minh *"Append root program node
at index 0"*) xác nhận: **node index 0 CHÍNH LÀ AST root thật (`NODE_PROGRAM`)**, không
phải sentinel rỗng. Đồng thời `NULL_IDX: u32 = 0` (ast.ax:102) CŨNG dùng giá trị 0 làm
sentinel "không có child/sibling" cho `first_child`/`next_sibling` — một sự trùng hợp giá
trị (0 vừa là root thật, vừa là NULL) mà 3 pass này hiểu SAI thành "0 luôn là NULL".

Cả ba pass gọi entry point bằng chính node 0: `ctgc_inj.run()` → `self.
traverse_and_inject(0 as u32, 0 as u32)`; `ownership_chk.check()` → `self.check_node(0 as
u32)`; `escape_an.run()` → `self.traverse_nodes(0 as u32)`. Guard `node_idx==0` bắt NGAY
LẦN GỌI ĐẦU TIÊN (root thật) → return tức khắc → **không bao giờ xử lý bất kỳ node nào,
suốt vòng đời của cả ba pass.**

## Hệ quả — nghiêm trọng hơn RFC 0014 (hay bất kỳ bug nào trong list BUG#5x-6x) rất nhiều

Output driver `[Debug] Running Ownership Checker...` / `Running Escape Analyser...` /
`Running CTGC Injector...` **ĐÁNH LỪA** — in ra như đang chạy thật nhưng không kiểm
tra/không đánh dấu/không chèn gì cả. Xác nhận độc lập bằng thực nghiệm: `let b = a; let c
= a` (copy 2 lần cùng struct) compile sạch KHÔNG lỗi, dù `check_move` (nếu chạy) lẽ ra phải
đánh dấu `a` là MOVED sau lần dùng đầu và báo `E4001 use of moved value` ở lần thứ hai.

1. **KHÔNG có compile-time GC nào đang chạy, cho BẤT KỲ chương trình AXIOM nào, từ trước
   tới giờ.** Mọi struct/sum/option/result/generic-inst local đều RÒ RỈ — không chỉ
   [[bignum-ctgc-conflict]]'s `BigUint.limbs`, mà CẢ khối chứa struct/sum/option/result
   TỰ THÂN cũng không bao giờ được free (CTGC đáng lẽ tự free đúng MỘT khối này qua
   `OP_DESTROY`→`ax_free`, nhưng chưa từng chạy). Toàn bộ self-hosting compiler + 98/98
   regression xanh **CHƯA TỪNG chạy dưới CTGC thật** — mọi thành công hiện tại hoàn toàn
   độc lập với CTGC (nó không tham gia gì cả).
2. **KHÔNG có ownership/move-checking nào đang chạy** — single-owner chỉ là ý định trên
   spec/CLAUDE.md, chưa từng được enforce bởi compiler.
3. **KHÔNG có escape analysis nào đang chạy** — `FLAG_ESCAPES_TO_HEAP` (field tồn tại, có
   code ĐỌC nó ở air_builder.ax/ctgc.ax) không bao giờ được SET bởi pass thật → luôn đọc
   thấy 0/false ở mọi nơi dùng nó.

## Đã thử nghiệm (KHÔNG commit) — vì sao chưa fix ngay

Sửa guard cả 3 file (bỏ `or node_idx == 0 as u32`, giữ `0xffffffff`) — về lý thuyết AN
TOÀN vì mọi lời gọi đệ quy nội bộ đã tự lọc `!= 0` ở call site (vòng lặp `while child_idx
!= 0: recurse(...)`) TRƯỚC KHI gọi, nên guard `==0` bên trong callee chỉ thực sự cần cho
đúng MỘT lần gọi gốc (từ `run()`/`check()`). Build thử (không commit):

- **Fixpoint A==B vẫn đạt** (compiler tự-build được, tự-host không vỡ ở mức fixpoint).
- **Regression: 0/98 — MỌI build đều lỗi ngay ở OwnershipChecker**, kể cả
  `bin/t_drop.ax` tối giản nhất (không import gì đặc biệt) — vì runtime/stdlib lõi LUÔN
  được bundle bất kể import gì (`result.ax, alloc.ax, scheduler.ax, runtime.ax, os.ax,
  string.ax, io.ax, collections.ax`) và bản thân các file lõi này đã vi phạm move-rule của
  OwnershipChecker (`error[E4002] cannot assign to immutable variable`, `error[E4001] use
  of moved value`) ở hàng chục vị trí. Toàn bộ codebase (stdlib lõi + self-hosting
  compiler, ~nhiều tháng phát triển) được viết và tiến hóa **giả định ba pass này không
  bao giờ chạy** — bật lên đòi hỏi rà + sửa (hoặc thiết kế lại luật checker cho đúng ngữ
  nghĩa dự định, đặc biệt `check_move` hiện đánh dấu MỌI struct dùng làm var-decl-
  init/assign-rhs/call-arg là "moved" vô điều kiện — không phân biệt copy hợp lệ, trong khi
  RFC 0010 (`rfcs/0010-aggregate-value-semantics.md`) đã chỉ ra struct-copy hiện là NGỮ
  NGHĨA THỰC của codebase, không phải lỗi cần cấm) trên diện RẤT RỘNG — không phải việc một
  session có thể làm xong.
- Guard fix + toàn bộ code `resolve_drop_method`/`lower_destroy` wiring (RFC 0014) đã
  **REVERT HOÀN TOÀN** (khôi phục từ backup) — không thể verify đúng vì 0 lần thực thi thật
  (chỉ chứng minh được "unreachable/an toàn", không chứng minh được "correct").

## Đo quy mô thật (2026-07-06, thử nghiệm KHÔNG commit, đã revert) — nhỏ hơn lo ngại ban đầu

Áp lại guard-fix tạm, build `bin/tsp.ax` (chỉ bundle runtime lõi tối thiểu): **5 lỗi**
(2× E4002, 3× E4001). Build CHÍNH `bootstrap/stage1/tmp_concatenated_air.ax` (TOÀN BỘ
self-hosting compiler + stdlib, ~1.37MB) qua chính nó: **260 lỗi tổng** — **258/260 là
E4001 "use of moved value"**, chỉ **2/260 là E4002**. Soi 1 sample E4001 (offset 93153,
`bootstrap/stage1/parser.ax` vùng `parse_...`): `node` (một AstNode/struct) được dùng lại
ở `self.append_child(node, body)` sau khi đã "moved" bởi một lần dùng trước đó trong cùng
hàm — đây là PATTERN HOÀN TOÀN BÌNH THƯỜNG (dùng lại một biến struct nhiều lần, đọc-only,
không di chuyển ownership thật) mà `check_move`'s luật "mọi struct dùng 1 lần là moved vô
điều kiện" hiểu sai thành vi phạm.

⟹ **Kết luận quan trọng: đây KHÔNG phải 260 bug riêng lẻ cần sửa tay** — tuyệt đại đa số
(258/260) đều cùng MỘT nguyên nhân (luật `check_move` quá thô, coi MỌI lần dùng struct thứ
2 trở đi là lỗi, không phân biệt reuse-đọc hợp lệ với move-ra thật). Sửa ĐÚNG MỘT chỗ (luật
move trong `ownership.ax::check_move` — cần thiết kế lại thế nào là "move thật" cho ngôn
ngữ này, có thể: chỉ coi là move khi RÕ RÀNG chuyển ra khỏi scope hiện tại — return/gán cho
biến có lifetime dài hơn — không phải MỌI cách dùng) sẽ giải quyết phần lớn 258 ca, không
phải sửa từng call site. 2 ca E4002 còn lại (offset 54360/54403, gần
`bootstrap/stage1/resolver.ax` vùng hash-table `mask`/`idx`) cần soi riêng — có thể là bug
thật (thiếu `mut`) hoặc checker gán nhầm offset, chưa xác nhận được trong lần đo này.

**Điều này làm cho "CTGC activation" bớt đáng sợ hơn ước tính ban đầu** ("diện rất rộng,
không phải việc một session") — vẫn KHÔNG phải việc nhỏ (cần thiết kế lại ngữ nghĩa move,
rồi RE-KIỂM TRA toàn bộ 260 vị trí sau khi đổi luật để chắc không còn false-negative/false-
positive, rồi mới tính tới audit CTGC free thật + escape-analysis độ chính xác ở mục dưới)
— nhưng có NỀN TẢNG RÕ RÀNG hơn để bắt đầu, thay vì một biển lỗi không định hình.

## Việc cần làm trước khi bật lại (KHÔNG phải việc nhỏ — cần RFC/kế hoạch riêng)

1. Sửa guard 3 file (1 dòng/file) — bản thân fix rất nhỏ, chỉ là ĐIỀU KIỆN CẦN.
2. **OwnershipChecker cần thiết kế lại `check_move`** trước khi bật — luật hiện tại
   ("mọi struct dùng lại sau khi từng là var-decl-init/assign-rhs/call-arg là lỗi") sẽ biến
   HẦU HẾT code hợp lệ hiện có thành lỗi biên dịch giả nếu bật nguyên trạng.
3. **CTGC cần audit an toàn free** trước khi bật — một khi chạy thật sẽ free hàng nghìn
   struct instance CHƯA TỪNG bị free trước đây trong compiler tự-host; RFC 0010 đã cảnh báo
   compiler tự-host DỰA VÀO alias-qua-buffer (đọc sau khi mutate cùng buffer, không
   realloc) — free thật lần đầu có thể biến các alias này thành UAF thật (hiện "an toàn"
   chỉ vì free chưa từng chạy, không phải vì logic đúng).
4. EscapeAnalyser's heuristic (vd `NODE_CALL_EXPR` luôn coi MỌI argument là "escape") chưa
   được audit độ chính xác — có thể quá bảo thủ hoặc sai ở nhiều trường hợp thực tế.

**Đề xuất hướng đi:** mở RFC/công việc riêng ("CTGC activation") theo dõi việc bật lại 3
pass này một cách có kiểm soát (có thể: bật từng pass riêng lẻ, hoặc bật dần theo
module/whitelist, đo blast radius thật trước khi bật toàn cục) — tách biệt hoàn toàn khỏi
[[rfc0014-drop-glue-blocked]] (`drop(self)` hook), vốn CHỈ nên tiếp tục triển khai THẬT sau
khi CTGC activation xong và ổn định trên toàn bộ regression.

## Cập nhật 2026-07-06 — RFC 0015 ĐÃ VIẾT (a48d07a)

`rfcs/0015-ctgc-activation.md` (Draft) đóng khung vấn đề + quyết định + kế hoạch phân pha.
**Chốt quan trọng (§3-§4):** `check_move` không "sai chi tiết" mà SAI TIỀN ĐỀ — nó giả định
move-only aggregate, TRÁI với alias/copy semantics của AXIOM (RFC 0010, mà compiler tự-host
DỰA VÀO). ⟹ đây là quyết định NGÔN NGỮ (CLAUDE.md §13), không phải guard-fix. Quyết định
§4.1 khuyến nghị: **AXIOM KHÔNG áp dụng move semantics** → OwnershipChecker's move-checking
phải BỎ (không chỉ retune); chỉ mutability-check (E4002) may có giá trị (cần audit 2 ca).
Kế hoạch: P0 doc (xong) → P1 mutability-only checker → P2 escape analysis đúng đắn → P3 CTGC
free whitelist. **Đã thêm comment `// no-op — RFC 0015` tại 3 guard** (ownership/escape/ctgc)
để không đánh lừa người đọc — comment-only, output self-compile BIT-IDENTICAL (cùng SHA
E2A2B33...), fixpoint A==B.

## Cập nhật 2026-07-06 — RFC 0015 P1 SHIPPED (5c50011) 🎉

**Lần ĐẦU TIÊN 1 trong 3 pass chạy thật.** Quyết định §4.1 CHỐT: **AXIOM KHÔNG move
semantics** (aggregate = copy/alias, RFC 0010). Nên:
- `check_move` GỌT thành no-op (đánh dấu mọi aggregate-use là moved → ~258 E4001 giả);
  nhánh E4001 chết. CHỈ **E4002 (assign vào `let` non-mut)** được enforce.
- Fix guard ownership.ax: `if node_idx == 0xffffffff: return` (bỏ `or ==0`) → pass CHẠY.
- **CHỈ ownership.ax bật; escape.ax + ctgc.ax VẪN no-op** (P2/P3, nguy hiểm — CTGC free
  gây UAF theo RFC 0010 §9).

**Bug thật lộ ra khi bật:** 2 E4002 khi self-build đều là FALSE POSITIVE — gán vào **`mut`
PARAMETER** (`fn substr(s, mut start, mut end): start = 0`). Root cause: parser set
`FLAG_IS_MUT` trên param (parse_param) NHƯNG resolver `NODE_PARAM_DECL` BỎ nó → `mut` param
không bao giờ mang `SYM_FLAG_MUT`. Fix resolver.ax (mirror NODE_VAR_DECL). **`SYM_FLAG_MUT`
CHỈ ownership.ax đọc** (grep xác nhận) → ZERO codegen effect; compiler đổi binary chỉ vì
source đổi, vẫn fixpoint A==B. Kết quả: 0 false-positive toàn stdlib+compiler, regression
102/102. Tests: `bin/t_mutparam.ax` (accept), `tests/sema/err_immutable_assign.ax` (reject).
**AXIOM giờ enforce `let` immutability** — gán vào non-`mut` binding = lỗi biên dịch.

Còn lại: P2 (escape analysis sound) + P3 (CTGC free whitelist) → mở khóa
[[rfc0014-drop-glue-blocked]]. Vẫn nhiều session, KHÔNG rush (RFC 0010 §9 UAF risk).

## Cập nhật 2026-07-12 — RFC 0015 P2 THỬ & REVERT (cff2552, doc-only)

Thử activate EscapeAnalyser (P2). **REVERT về baseline A==B `8AF4B46C`.** 2 blocker (chi tiết ở `rfcs/0015` §5 P2 + §8):
1. **Flag collision (ĐÃ GIẢI):** `FLAG_ESCAPES_TO_HEAP = 128` **TRÙNG** `SYM_FLAG_MOVED = 128`. escape.ax gốc `sym.flags |= FLAG_ESCAPES_TO_HEAP` → set SYM_FLAG_MOVED → `ownership.ax:79` đọc → **E4001 flood** (258 lỗi trở lại). Và `decl_node.flags |= 128` → air_builder heap-box (lower_var_decl ~3143 → lower_ownership_aware ~2419). ⟹ CẢ HAI write gốc đổi behavior. Giải: **`SYM_FLAG_ESCAPES = 4096` mới** (trống trong sym-flag space, chưa ai đọc) cho P3 đọc thay vì bit 128.
2. **BLOCKER MỞ (thật):** chỉ cần **BẬT pass chạy** (sửa guard `node_idx == 0xffffffff` only), **KỂ CẢ khi tắt hết flag-write** (pass KHÔNG mutate shared state, chỉ build/free ConnectionGraph mỗi hàm) → **A != B**: dao động **period-2 `8AF4B46C ⇄ A70D6243`**. seed(pass no-op) và A(pass chạy) compile CÙNG source ra binary KHÁC. CGNode=20/CGEdge=12 sizes ĐÚNG (không phải size bug). Pass ghi thứ codegen KHÔNG đọc → nghi ngờ **latent non-determinism codegen phụ thuộc allocation-order/địa-chỉ**, bị nhiễu bởi alloc/free traffic của pass (hoặc corruption trong connection_graph never-run ở scale lớn).

**NEXT SESSION (de-risked):** isolate nguyên nhân A!=B TRƯỚC KHI wiring SYM_FLAG_ESCAPES. Cách: (a) chạy pass trên 1 hàm nhỏ + dump, xác nhận deterministic; (b) bisect — thử pass CHỈ alloc/free 1 graph không traverse gì (isolate alloc-traffic vs traversal); (c) audit codegen có path nào phụ thuộc pointer-value/allocation-order. Chỉ khi pass chạy A==B mới set flag + validate soundness. **ĐỪNG autopilot thường** — cần điều tra determinism sâu.

## Cập nhật 2026-07-16 — RFC 0015 P2 SHIPPED (`f06d939`) 🎉 + phá án "non-determinism"

**"Latent codegen non-determinism / period-2 oscillation" của P2 (2026-07-12) là CHẨN
ĐOÁN SAI — thực chất escape pass CRASH (SIGSEGV, exit 139).** Bằng chứng: (a) fpB.log
dừng ngay tại "Running Escape Analyser..."; (b) `axc_fpB.exe` mtime KHÔNG đổi qua các
hop (hop-2 `A build -o axc_fpB` crash → ko ghi → file CŨ sót lại → hash "nhảy" giữa 2 build
cuối = **stale-artifact phantom**, ko phải 2-cycle thật); (c) chạy trực tiếp trên toy
2-hàm cũng segfault. **BÀI HỌC: luôn check mtime/exit-code của artifact trước khi kết luận
non-determinism — đừng tin hash từ fixpoint script khi một hop có thể đã crash.**

**Root cause crash:** trong `escape.ax::traverse_nodes`, pattern per-function
`let old_cg = self.curr_cg; self.curr_cg = new_connection_graph(); …; self.curr_cg = old_cg`
viết theo VALUE-COPY semantics, nhưng AXIOM aggregate = ALIAS ([[axiom-struct-reference-semantics]],
RFC 0010). `old_cg` ALIAS field `curr_cg` (ko snapshot) → sau hàm cuối free graph xong,
`curr_cg = old_cg` (self-alias no-op) → `curr_cg` giữ con trỏ buffer DANGLING → `run()`
gọi `free_connection_graph(self.curr_cg)` cuối cùng iterate `adj_out`/`adj_in` dangling →
`@free` rác → WILD FREE → SIGSEGV. **Fix:** hàm ko lồng nhau → BỎ save/restore; mỗi
NODE_FUNC_DECL tạo graph mới, free, rồi RESET `curr_cg = new_connection_graph()` (empty hợp lệ)
→ ko còn dangling. Pass chạy trên TOÀN BỘ self-host compiler, **A==B `F100027D`** → chứng minh
KHÔNG hề có non-determinism.

**P2 landed:** escaping local giờ mark bằng `SYM_FLAG_ESCAPES = 4096` (SYM-flag space,
tránh collision 128=SYM_FLAG_MOVED → E4001 flood). CHƯA ai đọc flag → fixpoint-neutral
(**A==B `184E35B4`, 327/327**). Oracle `bin/t_escape.ax` (exit 41). Frontend-neutral, commit
theo fixpoint-async.

**P3 (CTGC free) VẪN MỞ — high-risk, cần session riêng.** Blocker cụ thể (xác nhận 2026-07-16):
kích hoạt CTGC free vô điều kiện sẽ UAF-crash self-host — `ctgc.ax` free MỌI block-local
struct/sum/opt/result/gen-inst có bit escape=0, NHƯNG self-host đầy **borrow-local** kiểu
`let node = tree.nodes.data[i]` (alias VÀO vector sống lâu hơn, RFC 0010 §9). EscapeAnalyser
hiện KO mark local borrow-từ-INDEX/FIELD → CTGC sẽ free slot TRONG `tree.nodes.data` → corrupt
AST → UAF thảm họa. P3 cần: (a) borrow-edge trong ConnectionGraph cho INDEX_EXPR/FIELD_EXPR init
(mark borrow → never-free); (b) whitelist module/hàm; (c) fix guard no-op của `ctgc.ax` +
đổi read `FLAG_ESCAPES_TO_HEAP(128)` → `SYM_FLAG_ESCAPES(4096)`. Mở khóa [[rfc0014-drop-glue-blocked]].

## Cập nhật 2026-07-16 (tiếp) — P2.1 soundness + P3 machinery READY (`dc7cd0c`)

**P2.1 ownership-origin soundness SHIPPED `dc7cd0c`** (A==B `1A220D0B`, 327/327): trong
`escape.ax::analyze_stmt` NODE_VAR_DECL, local chỉ OWN memory tươi khi init =
`NODE_STRUCT_LIT`/`NODE_ARRAY_LIT`; init khác (ident=alias, INDEX/FIELD=borrow, call=maybe-
borrow) → `add_edge(local → escape_node)` → `escapes()=true` → P3 KO free. Kết hợp flow
return/call-arg/field-store sẵn có → cover luôn `let y=x` aliasing. Verified qua dump per-fn
toàn compiler: freeable set nhỏ/conservative (fn 19-local → 1 freeable, 22-local → 0).

**P3 machinery XÁC NHẬN ĐỦ (read-only 2026-07-16):** `OP_DESTROY` lowering ĐÃ implement mọi
backend — `air_builder.ax:4101 lower_destroy` (CTGC-injected NODE_DESTROY_STMT payload=sym →
`local_map_get(sym.name_id)` → emit OP_DESTROY(reg)), `x86_selector.ax:1719` → `ax_free`,
cgen/wasm cũng có. Chưa bao giờ chạy (ctgc no-op). ⟹ **substrate P3 HOÀN CHỈNH** (analysis
P2.1 + free machinery). CÒN LẠI để bật P3 (rủi ro cao, session riêng): (a) whitelist HOẶC
flag-gate `-ctgc-free` (mặc định off → self-host A==B) để bound blast radius; (b) fix guard
no-op `ctgc.ax` (node_idx==0 → 0xffffffff); (c) đổi read `FLAG_ESCAPES_TO_HEAP(128)` →
`SYM_FLAG_ESCAPES(4096)`; (d) validate no-UAF (allocator deterministic-fail trên full
regression). LƯU Ý UAF: `lower_destroy` free GIÁ TRỊ reg = con trỏ; borrow-local = con trỏ
NỘI BỘ vào vector → ax_free = corrupt allocator → chính xác lý do P2.1 tồn tại.

## Cập nhật 2026-07-16 (tiếp) — 🎉 P3 SHIPPED `038c2ea` — CTGC FREE HOẠT ĐỘNG

Xây TRỌN GÓI 1 unit (Stage 1→3, mỗi stage gate riêng), opt-in `-ctgc-free` (default OFF
→ self-host **A==B `AD550CE4`**, 329/329):
- **Stage 1 (escape.ax):** (a) ctor-detection — `NODE_CALL_EXPR` có callee ident resolve
  `SYM_STRUCT` = OWNING (vì `Box(f:v)` parse thành CALL_EXPR chứ ko phải NODE_STRUCT_LIT,
  node đó gần như unused); (b) scalar-field flow refinement — thread `checker.node_types`
  vào `new_escape_analyser` (thêm 2 param), trong analyze_expr special-case FIELD/INDEX:
  nếu `node_types[idx]`=`TYPE_KIND_PRIMITIVE` VÀ base là plain IDENT → value copy, KO flow
  base tới dest. **Conservative mọi trục** (unresolved=0/aggregate/complex-base → propagate)
  → chỉ free ÍT hơn, KO BAO GIỜ UAF. A==B `BEFA3A23` (inert). Dump: freeable=190/122 fn.
- **Stage 2 (ctgc.ax + main_air.ax):** ctgc gate `if not free_enabled: return`; fix guard
  (`node_idx==0` → chỉ `0xffffffff`); đọc `SYM_FLAG_ESCAPES` (KO phải FLAG_ESCAPES_TO_HEAP
  128=SYM_FLAG_MOVED); **BỎ before-return injection** (free trước `return tmp.v` = UAF) →
  chỉ inject OP_DESTROY tại block FALL-THROUGH (return-path leak = an toàn). `-ctgc-free`
  flag → `ctgc_inj.free_enabled`. A==B `AD550CE4` flag-off.
- **Stage 3 validate:** flag-on==flag-off 13-prog corpus (loop/alias/escape-return/escape-call/
  nested-block/aggregate-field-read/Option/Result/method/array); destroys fire (đếm >0); **TEST
  QUYẾT ĐỊNH: compiler tự-build với `-ctgc-free` (6 self-free trong chương trình alias-heavy
  nhất) → binary chạy ĐÚNG** (compile c1→5/c8→42/c12→42). Gate lặp lại `scripts/ctgc_free_check.sh`
  → `CTGC_FREE_OK`. Oracle `bin/t_ctgcfree.ax`(42).

**Machinery `OP_DESTROY`→`ax_free`** (air_builder lower_destroy + x86_selector) ĐÃ có sẵn từ trước.
**CÒN niche:** return-path frees leak (cần return-value-temp transform), Option/Result-ctor
locals ko free (chỉ SYM_STRUCT ctor = owning). **Mở khóa [[rfc0014-drop-glue-blocked]].**

Liên quan: [[bignum-ctgc-conflict]] (động lực ban đầu phát hiện ra bug này), RFC 0010
(`rfcs/0010-aggregate-value-semantics.md`, cảnh báo về alias-qua-buffer mà compiler
tự-host dựa vào — trực tiếp liên quan tới rủi ro mục 3 ở trên), full chi tiết kỹ thuật +
code thử nghiệm nằm trong `rfcs/0014-drop-glue.md` mục 8.
