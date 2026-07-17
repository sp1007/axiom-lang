---
name: bug85-scalar-field-address
description: "BUG#85 FIXED — &s.f on a SCALAR struct field yielded the address of a temp holding the loaded value (writes lost); new OP_FIELD_ADDR LEAs the real field address. Twin of BUG#84 (&a[i]). Aggregate &s.aggfield was always fine."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#85 FIXED** ✅ 80f09e6 (2026-07-09, backend x86_selector + new opcode). **Twin của [[bug84-scalar-array-elem-address]]** — tìm ngay sau khi fix #84 bằng cách probe sibling `&s.field`.

## Triệu chứng
`let px = &s.x` với field `x: i32` (SCALAR): ghi qua `set(px, 40)` KHÔNG bền (s.x giữ nguyên) → `s.x+s.y` = 8 thay 45. Giống hệt #84 nhưng cho struct field.

## Root cause
`lower_unary_expr` cho `&`: `lower_expr(child)` với child=NODE_FIELD_EXPR scalar → OP_GET_FIELD **LOAD** value; OP_MAKE_REF LEA slot temp → con trỏ trỏ temp, không phải field trong struct. **Aggregate `&s.aggfield` LUÔN đúng**: OP_GET_FIELD trả field ADDRESS (LEA) cho field_is_aggregate; compiler dùng `&self.blocks.data[i]` (field aggregate) an toàn.

## Fix (mirror #84)
- **`OP_FIELD_ADDR: u16 = 0x0112`** (air.ax) + mnemonic "fldaddr". x86: LEA dest = src1 + `field_offset(get_register_type(src1), src2)`, src2=OPND_IMM (spill rewrite hiểu pointer+disp). KHÔNG load.
- `lower_unary_expr`: emit CHỈ khi `op=="&"` VÀ child.kind==NODE_FIELD_EXPR VÀ có first_child VÀ `(flags & 2048)==0` (không phải qualified-variant/const field-expr) VÀ `not type_is_aggregate(node_types[child])`. src1=obj_reg, src2=`child_node.extra_idx` (field index). Aggregate giữ đường OP_GET_FIELD+MAKE_REF cũ.

## Gate: A!=B, B==C (KHÁC #84!)
Compiler tự dùng `&s.scalar_field` → đổi self-codegen → **A!=B là bootstrap transition ĐÚNG** (như [[bug77-16byte-struct-byaddress-unified]]); gate thật = build C từ B tay + **B==C** (`8d203c3d…`) + regression 112/112 + oracle. (BUG#84 A==B vì compiler KHÔNG dùng scalar `&a[i]`; #85 A!=B vì CÓ dùng scalar `&s.f` — self-host cũ vẫn chạy nghĩa là usage cũ read-only/benign, B mới đúng+hội tụ+regression xanh.)

## Ghi chú
- Oracle bin/t_fldelemaddr.ax (exit 150): write-through 2 field (`&s.x`, `&s.z`) + read-through `&s.y`. Regression row `t_fldelemaddr|exit|150`.
- cgen/wasm chưa handle OP_FIELD_ADDR nhưng native self-build không sao (native path). Cùng họ address/aggregate-repr [[bug84-scalar-array-elem-address]], [[bug64-vec-big-aggregate-element]].
