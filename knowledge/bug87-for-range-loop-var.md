---
name: bug87-for-range-loop-var
description: "BUG#87 FIXED — for-range loop variable (`for i in a..b`) was unreadable in the body (read 0/garbage): lower_for keyed it in local_map by sym.name_id but reads use sym_idx. Loop count was correct, only the variable value broke."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#87 FIXED** ✅ 694da03 (2026-07-09, frontend-only, fixpoint A==B, regression 113/113). Tìm bằng proactive probing (batch #8).

## Triệu chứng
`for i in 0..5: sum = sum + i` → sum = **0** (đáng lẽ 10). Loop CHẠY đúng số lần (`for i in 0..7: c+=1` → 7 ✓) nhưng ĐỌC biến `i` trong body ra 0/rác (`last = i` để last nguyên -1). Chỉ hỏng khi body THAM CHIẾU biến lặp.

## Root cause
`lower_for` (air_builder) `local_map_put(name_id, iter_reg)` với `name_id = sym.name_id` (tên interned). Nhưng MỌI đọc identifier (`lower_ident` L691) tra `local_map_get(sym_idx)` với sym_idx = node.payload (chỉ số SYMBOL) — y hệt cách `let`/`mut` var-decl lưu (`name_id = node.payload`, L2693). → loop var lưu 1 key, đọc 1 key khác → miss → 0. Count đúng vì lower_for tự get/put nội bộ đều dùng name_id (nhất quán với nhau, lệch với read path).

## Fix
`lower_for`: `name_id = sym_idx` (dùng chỉ số symbol làm key, khớp read path). Giữ type_id lookup từ sym. 1 dòng.

## Ghi chú
- Frontend-only, compiler KHÔNG dùng `for` (chỉ `while`) → A==B, không rủi ro self-host. Oracle bin/t_forrange.ax (28 = sum10+prod6+last12).
- **Follow-up SHIPPED** ✅ b504e70: `for x in <collection>` (non-range aggregate: Vec/HashMap/array/struct) giờ **REJECT sạch** ở typecheck (NODE_FOR_STMT: nếu không phải range `..` và iteree kind STRUCT/SUM/GENERIC_INST/ARRAY → diagnostic + diags_count++). Trước đây hang (~vô hạn vì so counter < địa-chỉ-Vec). `for i in n` (int) + `for i in a..b` (range) VẪN chạy. A==B (compiler không dùng for). Test tests/sema/err_for_collection.ax. Iteration thật (iterator protocol std.iter) = future. Cùng convention [[bug72-range-index-reject]].
- Vec API: `get[T](self,i)->Option[T]` (KHÔNG phải T) — probe phải `.get(i)` rồi unwrap. `for x in v` chưa chạy.
