---
name: bug81-multifield-variant-payload
description: "BUG#81 FIXED — enum/sum variant with >1 payload field (Two(i32,i32) / Node(Tree,Tree)) now rejected with a diagnostic instead of silently miscompiling (2nd+ field dropped → wrong answer / segfault)"
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#81 FIXED** ✅ df2f867 (2026-07-08, frontend/typecheck-only, fixpoint A==B `671B5BF5…`, regression 109/109).

Tìm ra bằng **proactive probing** (batch #2, cùng cadence BUG#70-80). Probe recursive enum `Node(Tree,Tree)` segfault; thu hẹp → non-recursive `Two(i32,i32)` cũng sai: field[0]=10 đúng, field[1] đọc **0** (đáng lẽ 7).

## Triệu chứng
Enum/sum variant khai báo **>1 payload field**: `Two(i32, i32)`, `Node(Tree, Tree)`. Field payload thứ 2+ bị **bỏ hoàn toàn**: scalar → đọc 0 (SAI ÂM THẦM, không lỗi), pointer (recursive tree) → deref rác → **segfault 139**. Parser parse đủ N type-children (`parse_type_variant` loop append), nhưng phần sau chỉ xử field đầu.

## Root cause
Design variant là **single-payload**: `VariantInfo` chỉ mang MỘT `payload_type`. Tại `pre_infer` sum registration ([typecheck.ax] ~L1368, vòng NODE_VARIANT_DECL): `payload_type = infer_node(v_node_struct.first_child)` — chỉ đọc type-child ĐẦU; `.next_sibling` (field 2+) bị lờ. `lower_variant_construct(vcsym, payload_arg)` nhận DUY NHẤT `callee_node.next_sibling` (arg[0]) → construct chỉ lưu field đầu; match extract 1 field.

## Fix (diagnostic reject, convention BUG#53/71/72)
Tại chính vòng registration: nếu `first_child.next_sibling != 0` (variant có ≥2 payload type) → `ax_printf_local("error: variant '%s' has multiple payload fields... wrap them in a struct")` + `diags_count += 1` (driver HALT trước codegen — BUG#53). Đóng cả ca wrong-answer lẫn segfault. **Workaround hợp lệ = single struct payload** (`Two(Pair)` với `struct Pair`, path BUG#58). Không compiler/stdlib nào khai multi-field variant → self-build không ảnh hưởng (grep xác nhận trước khi build).

## Ghi chú
- Multi-field (tuple) variant payload = **feature CHƯA implement** (cần RFC + typecheck tuple-payload + air_builder construct/match N-offset + sizing). Future work nếu muốn hỗ trợ thật; hiện dùng struct payload.
- Cùng họ silent-miscompile: [[bug80-free-call-overload-collision]], [[bug73-closure-capture-reject]], [[bug68-struct-eq-no-overload]]. Test: bin/t_variantstruct.ax (oracle 17), tests/sema/err_multifield_variant.ax (reject, .diag rỗng theo convention).
- Bài học probing: recursive segfault → thu hẹp về non-recursive + tách field[0]/field[1] riêng lẻ (exit 8-bit truncation gây khó đọc; trả từng field một).
