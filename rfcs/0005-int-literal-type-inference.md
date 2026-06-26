# RFC 0005 — Integer literal type inference in binary expressions

- **Status:** Implemented v1 (2026-06-26)
- **Author:** self-host team
- **Tracking:** next-step-15 (ergonomics)
- **Liên quan:** typecheck.ax NODE_BINARY_EXPR / NODE_INT_LIT

## 1. Motivation

Lập trình viên phải gõ `as` cho integer literal khắp nơi vì literal mặc định `i32`
trong khi biến thường là `u32`/`u64`/`i64`:
```axiom
let a = b + 1 as u32      // bực mình: phải cast literal
if x < 10 as u32:         // cả trong so sánh
total = total + 1 as u32
```
Root cause: trong `infer_node` NODE_BINARY_EXPR, cả hai vế được infer với
`TYPE_UNKNOWN`; literal không có ngữ cảnh → mặc định `i32` (NODE_INT_LIT). Khi vế kia
là `u32`, hai toán hạng lệch kiểu → buộc cast literal.

## 2. Design

**Bidirectional literal type inference:** trong biểu thức nhị phân (số học VÀ so sánh),
nếu đúng MỘT vế là **bare integer literal** (NODE_INT_LIT) còn vế kia là **kiểu số cụ
thể** (`tc_is_numeric`), literal tự nhận kiểu của vế kia. Sau đó không cần `as`:
```axiom
let a = b + 1       // 1 : u32
if x < 10:          // 10 : u32
total = total + 1   // 1 : u32
let d = 100 - b     // 100 : u32 (literal ở vế trái cũng được)
```

Cơ chế: literal đã honor `expected` sẵn (NODE_INT_LIT: nếu expected là kiểu số →
result = expected). Nên chỉ cần **infer lại** node literal với kiểu cụ thể của vế kia
(`infer_node` ghi lại node_type ở cuối). Áp dụng cho nhánh `op==1` (so sánh, kết quả
bool) và nhánh số học.

## 3. Tại sao an toàn cho self-host (fixpoint)

Thay đổi có **guard `t1 != t2`**: chỉ kích hoạt khi literal (mặc định i32) LỆCH kiểu
với vế kia. Code compiler hiện tại viết `as` tường minh → literal nằm trong
`NODE_CAST_EXPR` (KHÔNG phải NODE_INT_LIT trực tiếp) nên `lhs_lit/rhs_lit` = false →
nhánh mới KHÔNG chạy → node_types không đổi → codegen y hệt → fixpoint giữ. Tính năng
chỉ tác động code KHÔNG cast (vốn trước đây phải cast hoặc lỗi). Thuần ergonomics.

## 4. Drawbacks / giới hạn (follow-up)

- Chỉ literal trực tiếp (NODE_INT_LIT). Literal âm `-1` (NODE_UNARY_EXPR(NEG,INT_LIT))
  hoặc literal trong ngoặc CHƯA được coerce — vẫn cần `as` (hiếm). Follow-up: nhận diện
  unary-neg-of-literal.
- Chỉ trong biểu thức nhị phân. Đối số hàm / return / index dùng cơ chế `expected` sẵn có
  (đã hoạt động cho literal trực tiếp).
- KHÔNG phải implicit int coercion giữa hai BIẾN khác kiểu (vd `u32_var + i64_var` vẫn
  theo promotion cũ); chỉ literal mới được suy kiểu. (Cố ý: tránh bug width/sign ngầm.)

## 5. Test plan / DoD

- [x] `bin/t_litinfer.ax`: `b+5`, `100-b`, `x<20`, `c==15`, `d>80`, `total+1` không cast → 111.
- [x] regression_repros.sh PASS (literal đã-cast cũ không đổi).
- [ ] verify_bug29_selfhost.sh: fixpoint stage3==stage4 GIỮ.
