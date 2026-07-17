---
name: bug72-range-index-reject
description: "BUG#72 FIXED (6601643): `arr[x..y]` (range as array index / slicing syntax) segfaulted — `..` has NO dedicated Range/Slice type, it parses into a plain NODE_BINARY_EXPR (parser.ax treats it exactly like +/-), and NODE_FOR_STMT is the ONLY place giving it real meaning (loop bounds). Used as an index it silently typechecked as ordinary arithmetic and the garbage value became a raw element offset. Fixed with a diagnostic reject at NODE_INDEX_EXPR (same class as BUG#53/64/68/70/71); slicing-from-range is a real unimplemented feature, not something fixed here."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#72 FIXED — 6601643, 2026-07-06.** Third dormant/half-shipped feature found via
proactive probing this session (after [[bug70-array-literal-shipped]] and
[[bug71-interface-dynamic-dispatch]]), continuing directly from testing arrays (this is
literally the next thing anyone would try after array literals: slicing a fixed array).

## Triệu chứng

```
let arr: [i32; 5] = [1 as i32, 2 as i32, 3 as i32, 4 as i32, 5 as i32]
let sl = arr[1 as i64..4 as i64]   // <-- SEGFAULT
```

## Root cause — AXIOM KHÔNG CÓ kiểu Range/Slice-từ-range thật

`..` (`TK_DOT_DOT`) KHÔNG có node/type riêng — `parser.ax::parse_led` xử lý nó y hệt
`+`/`-`/so sánh: tạo `NODE_BINARY_EXPR` chung (binding power 35, "range" chỉ là tên biến,
không phải semantics đặc biệt). **`NODE_FOR_STMT` là nơi DUY NHẤT** gán ý nghĩa thật cho nó
(typecheck.ax:1622-1626: `range_expr = node.first_child` — for-loop biết đọc bound
start/end từ CẤU TRÚC con của binary-expr, KHÔNG phải nhờ 1 kiểu Range riêng).

Ở NODE_INDEX_EXPR (`arr[idx]`, typecheck.ax nhánh `else:` ~dòng 2739, xử lý indexing THẬT
— khác nhánh phía trên xử lý `Vec[T]` generic type-args), code gọi
`self.infer_node(idx, TYPE_UNKNOWN)` KHÔNG kiểm tra `idx` có phải scalar hợp lệ hay không —
`1..4` (NODE_BINARY_EXPR, flags=0 vì `..` không rơi vào nhóm so sánh/and-or được gán flags)
đi qua ĐÚNG cùng code path như `1+4`, nhận 1 kiểu số nguyên bất kỳ — rồi giá trị số (vô
nghĩa) đó được dùng làm OFFSET PHẦN TỬ THẬT trong codegen → đọc/ghi bộ nhớ sai vị trí →
segfault. KHÔNG có diagnostic nào trước fix.

## Fix — diagnostic reject tại điểm dùng nguy hiểm (index), KHÔNG cấm `..` toàn cục

Theo đúng pattern đã dùng suốt session (BUG#53/64/68/70/71): KHÔNG cố cấm `..` ở MỌI nơi
bất hợp lệ (cần parent-tracking, phạm vi lớn hơn nhiều) — chỉ chặn tại điểm CHẮC CHẮN nguy
hiểm đã chứng minh (index expression). Tại `typecheck.ax`'s NODE_INDEX_EXPR real-indexing
nhánh: nếu `idx`'s node là `NODE_BINARY_EXPR` VÀ text toán tử == `".."` → lỗi rõ ràng
("range-slicing ('a[x..y]') is not yet supported; index must be a single value") +
`diags_count++`, TRƯỚC KHI gọi `infer_node(idx, ...)`.

**Verify**: case segfault gốc → giờ lỗi biên dịch sạch, không segfault. `for i in 0..30`
(dùng CHÍNH `..` hợp lệ, qua NODE_FOR_STMT — KHÔNG đi qua nhánh INDEX_EXPR) vẫn hoạt động
bình thường (test `tests/sema/valid_hello.ax` build sạch). Array literal (`bin/t_arrlit.ax`,
BUG#70) không bị ảnh hưởng. Fixpoint A==B, regression 99/99, negative test mới
`tests/sema/err_range_index.ax`.

**Ngụ ý**: slicing-từ-range (`arr[a..b]` trả về 1 slice con) là TÍNH NĂNG THẬT chưa tồn tại
— cần thiết kế Range/Slice type riêng (RFC), KHÔNG phải phạm vi fix này. Nếu sau này muốn
làm slicing thật: (1) cho `..` một representation riêng (không dùng chung NODE_BINARY_EXPR
với arithmetic), (2) NODE_INDEX_EXPR cần nhánh RIÊNG nhận diện range-index → tạo slice
value (data_ptr + len) thay vì scalar offset.

Liên quan: [[bug70-array-literal-shipped]] (cùng buổi probing arrays),
[[bug71-interface-dynamic-dispatch]] (cùng pattern "tính năng half-shipped lộ qua proactive
probing", cùng session).
