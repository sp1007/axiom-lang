---
name: bug80-free-call-overload-collision
description: "BUG#80 FIXED — user free-function whose name collides with a stdlib fn (get/push/len...) no longer silently miscompiles; overloaded free-calls now disambiguate by argument type"
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#80 FIXED** ✅ 6b310f4 (2026-07-08, frontend/typecheck-only, fixpoint A==B `E8FF2AE5…`, regression 108/108).

Tìm ra bằng **proactive probing** (đúng cadence BUG#70-79). Probe p2 (struct chứa Option field truyền by-value) trả 0; thu hẹp dần phát hiện trigger THẬT không phải Option/16B-struct mà là **mọi probe fail đều đặt tên hàm helper là `get`** (p2c=`get_tag`, p2d=`get_val`, p2m=`sumP` → PASS; đổi `get`→`combine` → PASS ngay). `get` trùng stdlib.

## Triệu chứng
User định nghĩa free function tên trùng stdlib (`get`: Vec.get/HashMap.get; cũng `push`,`len`,...). Gọi `get(a)` ở free-position → trả **0 âm thầm, KHÔNG lỗi biên dịch** (silent accept-then-miscompile, cùng lớp BUG#73). stdlib LUÔN được link nên footgun rộng: bất kỳ hàm user trùng tên stdlib đều hỏng.

## Root cause
Resolver bind callee ident của call về **overload-chain HEAD** (`resolver.ax` NODE_IDENT ~L997 `symtable.resolve(name)` trả symbol ĐẦU TIÊN cùng tên; `define()` chain overload qua `next_overload` nhưng KHÔNG `scope_put` lại → scope vẫn trỏ head). **KHÔNG có disambiguation theo arg cho free-call** — chỉ method-call `x.f()` chọn overload (qua `resolve_method_overload` keyed on receiver). Head của `get` = stdlib `get[T]` (generic) → typecheck NODE_CALL_EXPR path `is_generic_call` (L1935+) instantiate `get[T=UserType]` trên sai chữ ký → 0.

## Fix (frontend, typecheck.ax)
`resolve_free_call_overload(sym_idx, first_arg_node)` (cạnh `resolve_method_overload`): nếu head có `next_overload!=0`, infer kiểu arg[0], gọi `resolve_method_overload(head, arg0_type)` (nó check first-param vs "receiver"=arg0) + `is_method_compatible` xác nhận → chọn overload khớp. Chèn ở đầu nhánh `callee_node.kind==NODE_IDENT` của NODE_CALL_EXPR, ghi `tree.nodes[callee].payload` (air_builder đọc payload để lower) + refresh `callee_node` (đổi sang `mut`; infer arg có thể realloc nodes — BUG#51). **Bảo thủ**: `resolve_method_overload` test HEAD trước → call đang bind đúng KHÔNG đổi; chỉ switch khi head **incompatible** arg0 và chain-member khác khớp = thuần repair. Không đụng backend/ABI/resolver.

## Bài học / methodology
- **Đổi HẰNG SỐ + đổi TÊN khi thu hẹp**: bug "16B struct by-value" hóa ra là "tên hàm `get`". Nếu chỉ tin 1 repro đã kết luận nhầm ABI. p2c/p2d/p2m PASS (không tên `get`) là manh mối chốt.
- Cũng cải thiện free-function overloading nói chung (chọn theo arg[0], giống method chọn theo receiver) — không chỉ ca collision stdlib.
- Giới hạn: disambiguate chỉ theo arg[0] (giống `resolve_method_overload` chỉ theo receiver). Overload khác nhau ở arg thứ 2+ nhưng arg[0] trùng → chưa phân biệt (đủ cho ca stdlib-collision).
- Đóng cùng họ silent-miscompile: [[bug73-closure-capture-reject]], [[bug68-struct-eq-no-overload]].
