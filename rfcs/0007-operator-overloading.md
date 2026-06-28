# RFC 0007 — Operator overloading (convention-based)

- **Status:** Implemented v1 (2026-06-28)
- **Author:** self-host team
- **Tracking:** next-step-15 (ergonomics) / BUG#39
- **Liên quan:** air_builder.ax lower_binary_expr / lower_op_overload / resolve_op_method

## 0. Ghi chú implement (khác đề xuất §3)

Resolution + dispatch làm **HOÀN TOÀN trong air_builder.lower_binary_expr** (KHÔNG đụng
typecheck/parser/grammar như §3 dự kiến). Lý do: `resolve_method_sym` của typecheck chỉ
tìm được generic method ở fallback (guard `SYM_FLAG_GENERIC`), KHÔNG resolve được struct
method non-generic. air_builder đã có cơ chế đúng (`match_mangled_method_raw_bytes` + so
first-param-type) dùng cho `obj.method()`. Nên: air_builder kiểm tra `node_types[lhs]` là
user type → `op_to_method_name` → `resolve_op_method` (reuse match loop) → `lower_op_overload`
emit OP_CALL `[self, rhs]`. Result type KHÔNG cần typecheck đổi: arithmetic = lhs type
(numeric-promotion fallback đã trả `t1`), comparison = bool. `!=` = `eq` + OP_NOT.
Verified: bin/t_opover.ax exit 44.
- **Mở khóa:** BigInt/BigNum (số lớn cho tính toán khoa học) — `a + b` trên kiểu người dùng

## 1. Motivation

User muốn dùng số siêu lớn/chính xác (BigInt, BigDecimal) với **cú pháp giống số
builtin**: `let c = a + b`, `if a < b:`, `a * b`. Hiện AXIOM KHÔNG có operator
overloading (BUG#39): `a + b` luôn map cứng sang OP_IADD/OP_FADD theo kiểu builtin;
toán hạng là struct → sai/không hợp lệ. Đây là rào cản DUY NHẤT khiến bignum bất khả
thi (đã có struct, method, generics; chỉ thiếu operator dispatch).

## 2. Design — convention-based, KHÔNG đổi grammar

Operator trên **kiểu người dùng** (struct/ADT, không phải builtin numeric) **desugar
sang method gọi theo TÊN QUY ƯỚC**. KHÔNG thêm cú pháp khai báo operator mới
(không `operator +`, không `fn +`): tái dùng method thường (đã parse + resolve +
codegen ổn định), nên rủi ro tối thiểu.

### 2.1 Bảng ánh xạ operator → tên method

| Operator | Method | Trả về |
|----------|--------|--------|
| `a + b`  | `a.add(b)` | T |
| `a - b`  | `a.sub(b)` | T |
| `a * b`  | `a.mul(b)` | T |
| `a / b`  | `a.div(b)` | T |
| `a % b`  | `a.rem(b)` | T |
| `a == b` | `a.eq(b)`  | bool |
| `a != b` | `a.eq(b)` rồi phủ định | bool |
| `a < b`  | `a.lt(b)`  | bool |
| `a <= b` | `a.le(b)`  | bool |
| `a > b`  | `a.gt(b)`  | bool |
| `a >= b` | `a.ge(b)`  | bool |
| `-a`     | `a.neg()`  | T (unary, follow-up) |

(Tên quy ước theo Rust core::ops, đã quen thuộc. `!=`/`>=`/`<=`/`>` có thể derive từ
`eq`/`lt` nhưng v1 yêu cầu method tường minh cho rõ ràng + đơn giản codegen.)

### 2.2 Điều kiện kích hoạt

Chỉ khi **kiểu của toán hạng TRÁI (lhs) là kiểu người dùng** (TYPE_KIND_STRUCT /
SUM / GENERIC_INST — KHÔNG phải i*/u*/f*/bool/char/str builtin). Builtin numeric giữ
NGUYÊN codegen native hiện tại (OP_IADD/OP_FADD/mask...) → **byte-identical → fixpoint
giữ**. Nếu lhs là struct mà KHÔNG có method tương ứng → lỗi typecheck (BUG#35 infra,
tạm thời: bỏ qua/giữ hành vi cũ cho tới khi có error infra).

## 3. Implementation

1. **typecheck.ax** (NODE_BINARY_EXPR, cả nhánh so sánh op==1 và số học):
   sau khi có `t1` (lhs type), nếu `t1` là kiểu người dùng (không `tc_is_numeric`,
   không str/bool): tra method theo tên quy ước trong method-set của type `t1`,
   lưu `sym_id` method vào node (extra_idx), set `result_type` = ret của method (so
   sánh → bool). Tái dùng đúng đường resolve method của NODE_FIELD_EXPR call.
2. **air_builder.ax** (lower_binary_expr): nếu node có `op_method_sym` (extra_idx):
   lower lhs + rhs, emit **call 2-tham-số** (`self`=lhs, arg0=rhs) tới method sym
   (tái dùng path emit OP_CALL của lower_call_expr method branch). KHÔNG vào path
   numeric. `!=` → gọi `eq` rồi OP_NOT kết quả bool.
3. KHÔNG đụng parser/resolver/grammar.

## 4. Tại sao an toàn cho self-host (fixpoint)

- Chỉ kích hoạt khi lhs là **kiểu người dùng**. Compiler self-host KHÔNG dùng
  `struct <op> struct` ở đâu → toàn bộ thay đổi LATENT → stage* byte-identical → fixpoint
  giữ. Verify đầy đủ + repro runtime (BigInt mini) để chứng minh chạy.

## 4b. Giới hạn MVP v1 (QUAN TRỌNG cho biểu thức trộn kiểu)

Khi bignum/user-type trộn với primitive trong cùng biểu thức, v1 có 3 giới hạn:

1. **Dispatch CHỈ theo toán hạng TRÁI.** `big + 5` → `big.add(...)` (OK), nhưng
   `5 + big` → lhs là primitive → đi đường numeric → `OP_IADD(int, ptr-bignum)` = RÁC.
   → Hiện phải đặt user-type bên TRÁI. (Follow-up: dispatch đối xứng — nếu lhs
   primitive mà rhs là user-type, thử `rhs.<op>` hoặc reflected method.)
2. **Resolve theo CẢ HAI operand type (cập nhật 2026-06-28).** `resolve_op_method`
   giờ khớp param0(self)==kiểu lhs VÀ param1==kiểu rhs (khi rhs type biết). →
   (a) hỗ trợ operator method mixed-type đơn `add(self, o: i64)` cho `Num + i64`
   (verified t_opmix.ax exit 13); (b) KHÔNG còn gọi nhầm `add(self,o:BigInt)` với
   rhs khác kiểu (trả no-match thay vì miscompile thầm lặng). **NHƯNG** muốn có
   ĐỒNG THỜI `add(self,o:BigInt)` VÀ `add(self,o:i64)` (method overloading cùng
   tên) thì CHƯA được: 2 method cùng tên collision về 1 symbol (mangling theo tên,
   không theo chữ ký) → segfault. → Hiện: MỘT operator method mỗi tên/type (có thể
   mixed-type). Follow-up: method overloading cần signature-based mangling
   (resolver+symtab+codegen) — RFC riêng, lớn.
3. **KHÔNG ép kiểu ngầm primitive → user-type.** Mọi primitive phải convert tường
   minh sang bignum trước khi vào biểu thức. (Đồng nhất với chính sách RFC 0006:
   không lan kiểu thầm lặng.)

Primitive trộn primitive (vd `(a >> 8) as u8`) KHÔNG bị các giới hạn này — chỉ cần
`as` tường minh khi thu hẹp/đổi dấu (RFC 0006).

## 5. Out of scope (follow-up)

- **Const generics `bignum[256]`** (kích thước tham số hóa) — feature RIÊNG, lớn hơn
  (generic param hiện chỉ là KIỂU, chưa nhận giá trị hằng). Cần RFC khác.
- Auto-derive `!=` từ `eq`, `>=`/`>`/`<=` từ `lt` (v1 yêu cầu tường minh).
- Operator cho assignment compound `+=` trên struct (desugar `a = a.add(b)`).
- Index `a[i]` overloading (`index`/`index_set`).

## 6. Migration / compatibility

Không phá vỡ gì: builtin operator không đổi. Thuần mở rộng cho kiểu người dùng.
