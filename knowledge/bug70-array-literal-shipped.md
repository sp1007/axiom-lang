---
name: bug70-array-literal-shipped
description: "BUG#70 FIXED + fixed-size array literals SHIPPED (3e94c4e, 2026-07-06): `[e0, e1, ...]` had zero parser support (NODE_ARRAY_LIT existed with backend lowering already written but unreachable); added parser NUD + typecheck inference, which exposed a real dormant backend bug (lower_array_lit's element OP_STORE hardcoded type_id=0 -> wrong/doubled element stride on write vs read) — fixed to match the element type, mirroring index-assignment lowering. Fixpoint + regression 99/99 green."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**Fixed-size array literals `[e0, e1, ...]` SHIPPED + BUG#70 FIXED — 3e94c4e, 2026-07-06.**
Found via proactive probing after [[bug69-ctgc-ownership-escape-noop]] (needed a fresh,
lower-risk area to probe after that discovery turned out to need a dedicated effort rather
than a quick fix).

## Phát hiện: tính năng có sẵn backend nhưng KHÔNG BAO GIỜ reachable

`bootstrap/stage1/ast.ax` có `NODE_ARRAY_LIT` (kind 43), và `air_builder.ax` đã có
`lower_array_lit` viết sẵn (OP_ALLOC + loop OP_STORE từng phần tử) — nhưng **parser
KHÔNG BAO GIỜ tạo ra node này**: không có case nào cho `[` ở vị trí NUD (primary
expression) trong `parse_primary`/nud dispatch (`parser.ax`, hàm kết thúc bằng
`self.errorf(tok, "expected expression nud")`). `[1, 2, 3]` luôn là parse error. Kiểu mảng
CỐ ĐỊNH `[T; N]` (annotation, vd tham số hàm `a: [i32; 5]`) đã hoạt động đầy đủ (typecheck
có `register_array`, air_builder có `builder_type_size_and_align` case TYPE_KIND_ARRAY) —
chỉ riêng LITERAL EXPRESSION là thiếu. Cùng loại "half-shipped" như `NODE_STRUCT_LIT`
(cũng tồn tại + có lowering `lower_struct_lit`, nhưng construction thật `Type(field: v)` đi
qua đường NODE_CALL_EXPR hoàn toàn khác — NODE_STRUCT_LIT vẫn dead code, không đụng tới).

## Fix 1 — parser (an toàn, không ambiguous)

Thêm case `tok.kind == TK_L_BRACKET` vào primary-expression nud (trước dòng
`self.errorf(...)`): parse danh sách expr phân cách dấu phẩy tới `]`, tạo `NODE_ARRAY_LIT`.
AN TOÀN vì `[` ở vị trí NUD (bắt đầu 1 expression) trước đây LUÔN là lỗi — không có cú
pháp hợp lệ nào bị ảnh hưởng; index hậu tố `a[i]` là LED (áp lên 1 primary đã parse xong),
khác vị trí hoàn toàn.

## Fix 2 — typecheck (case mới, mirror NODE_INT_LIT default-i32 pattern)

Thêm `elif kind == NODE_ARRAY_LIT:` trong `infer_node` (gần NODE_CHAR_LIT): element type
lấy từ phần tử ĐẦU (hoặc từ `expected` hint nếu là TYPE_KIND_ARRAY — vd var-decl có
annotation `[i32;5]`), các phần tử sau infer với hint đó (coercion literal); `count` = số
con; gọi `self.types.register_array(elem_type, count)`. Fallback rỗng → i32 (giống
NODE_INT_LIT).

## BUG#70 — phát hiện NGAY khi test lần đầu (backend bug ẩn từ trước, chưa từng lộ)

`arr[0..4]` cho `[1,2,3,4,5]` trả về `[1,0,2,0,3]` thay vì `[1,2,3,4,5]` — pattern lộ rõ
element bị GHI với stride ĐÔI (8 byte) trong khi ĐỌC với stride ĐÚNG (4 byte, i32).
Root cause: `lower_array_lit`'s vòng lặp emit `OP_STORE` với **`type_id: 0 as u32` cứng**
cho mỗi phần tử — trong khi lowering `a[i] = v` (index-assignment thật, air_builder.ax
~2728-2733) tính `type_id = entry.extra` (kiểu phần tử của mảng/slice/pointer, đọc từ
`TypeEntry.extra`) để codegen scale đúng offset = index × elem_size. Store với type_id=0
khiến backend dùng elem_size mặc định SAI (khác cỡ so với đọc) → ghi lệch vị trí.

**Fix**: tính `elem_type_id` từ `TypeEntry` của mảng (case ARRAY/SLICE/POINTER đọc
`entry.extra`) NGAY ĐẦU `lower_array_lit`, dùng làm `type_id` cho mỗi `OP_STORE` — mirror
chính xác logic index-assignment đã có, không phát minh cơ chế mới.

**Verify**: `arr[0]==1`, `arr[4]==5`, `sum_arr(arr)` (truyền mảng theo giá trị vào hàm,
đọc qua vòng lặp `while`) = 15 — cả đọc trực tiếp lẫn truyền tham số đều đúng. Fixpoint
A==B (double-hop thủ công xác nhận độc lập với `fast_fixpoint.ps1`), daily-driver
`bin/axc_native.exe` rebuilt, regression 99/99 (test mới `bin/t_arrlit.ax`).

**Bài học methodology, nhất quán với các bug trước**: một tính năng "trông như đã có" ở
backend (lowering function tồn tại, được gọi tên rõ ràng) không có nghĩa nó ĐÃ TỪNG CHẠY —
đây là bug ẩn thứ 2 trong session này (sau [[bug69-ctgc-ownership-escape-noop]]) chỉ lộ ra
NGAY KHI một tính năng dormant lần đầu được làm reachable. Luôn viết oracle test ngay khi
mở khóa 1 tính năng "sẵn có nhưng chưa test", đừng giả định code cũ đã đúng chỉ vì nó tồn
tại lâu.
