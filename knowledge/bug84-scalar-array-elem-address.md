---
name: bug84-scalar-array-elem-address
description: "BUG#84 FIXED — &a[i] on a SCALAR array/pointer element yielded the address of a temp holding the loaded value (writes lost, reads past index 0 → 0); new OP_INDEX_ADDR computes the real element address. Aggregate &agg[i] was always fine."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#84 FIXED** ✅ c7072f0 (2026-07-09, backend x86_selector + new opcode, fixpoint A==B `63002833…`, regression 111/111). Tìm bằng **proactive probing** (batch #5, follow-up nghi ngờ t3b từ session trước).

## Triệu chứng
`let p = &a[0]` với `a: [i32;4]` (element SCALAR): `p[0]` đúng nhưng `p[1..]` = **0** (không phải garbage → zeroed stack), và ghi qua con trỏ (`store_at(p,1,20)`) KHÔNG bền (a[1] giữ nguyên). `&a[1]` rồi `p[0]` = a[1] đúng (con trỏ trỏ đúng temp giữ 1 giá trị). → `&a[k]` trả **địa chỉ của TEMP giữ VALUE a[k]**, không phải lvalue trong mảng.

## Root cause
`lower_unary_expr` cho `&`: `operand_reg = lower_expr(child)` → với child là NODE_INDEX_EXPR scalar, OP_INDEX **LOAD** giá trị a[i] vào reg tạm; rồi OP_MAKE_REF (x86 `LEA [slot]`) lấy địa chỉ SLOT của reg tạm đó. p[0] "đúng" chỉ vì đọc lại temp vừa load. **Aggregate `&agg[i]` LUÔN đúng**: OP_INDEX trả element ADDRESS cho aggregate (x86 line ~1784 MOV tmp_addr), MAKE_REF cho aggregate = MOV pass-through → địa chỉ. Compiler tự dùng `&self.blocks.data[i]`, `&table.entries.data[id]` (TẤT CẢ aggregate struct-element) nên self-host không lộ bug.

## Fix (scoped, đúng convention không đổi self-codegen)
- **`OP_INDEX_ADDR: u16 = 0x0111`** (air.ax) + mnemonic "idxaddr". Tính base + i*elem_size, trả ADDRESS, KHÔNG load. elem-size lấy từ pointee của src1 (`get_register_type(src1)` → POINTER/REF/SLICE/ARRAY → `.extra`), KHÔNG từ `inst.type_id` (= ptr[elem] result, size 8). Có nhánh str-16byte (`regalloc_is_16byte` → LEA+LOAD ptr field) như OP_INDEX để `&s[i]` (byte của str) an toàn.
- `lower_unary_expr`: chỉ khi `op=="&"` VÀ child.kind==NODE_INDEX_EXPR VÀ `not type_is_aggregate(node_types[child])` → emit OP_INDEX_ADDR (type_id = node_types[unary] = ptr[elem]) rồi return, BỎ QUA MAKE_REF. Aggregate giữ nguyên đường cũ → **compiler self-codegen byte-identical → A==B đủ** (khác BUG#77 cần B==C vì đó đổi self-codegen).
- x86_selector: case OP_INDEX_ADDR (mirror OP_INDEX base+offset, MOV địa chỉ vào dest thay vì LOAD).

## Ghi chú
- ssa_opt không whitelist opcode (fold chỉ ICONST) → op mới pass-through, dest được dùng → sống qua DCE. cgen/wasm KHÔNG handle OP_INDEX_ADDR nhưng compiler không emit nó cho chính mình → self-build/fixpoint qua native an toàn; chỉ chương trình dùng scalar `&a[i]` biên qua backend gcc/wasm mới cần (future).
- Oracle bin/t_arrelemaddr.ax (exit 66): write-through `store_at(p,1,20)` + read-through `sum4(p)` + `&a[2]`. Regression row `t_arrelemaddr|exit|66`.
- Cùng họ address/aggregate-repr: [[bug64-vec-big-aggregate-element]], [[bug77-16byte-struct-byaddress-unified]]. Cùng cadence probing [[bug83-underscore-digit-separator]], [[bug80-free-call-overload-collision]].
