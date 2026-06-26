# RFC 0004 — Native codegen for user-defined sum types (ADT)

- **Status:** Implemented v1 (2026-06-26)
- **Author:** self-host team
- **Tracking:** next-step-15 T2 / sub-1; đóng BUG#32 (phần v1)
- **Liên quan:** RFC 0001 (aggregate ABI), RFC 0003 (enum sugar), knowledge/bugs.md BUG#32

## 1. Motivation

`type X = A | B(T)` và `enum` (RFC 0003) parse + typecheck đúng nhưng backend native
KHÔNG sinh code (BUG#32): `match` không được lower, và constructor variant do người
dùng định nghĩa không được lower (chỉ Ok/Err/Some/None hard-code). Hệ quả: enum/sum
type vô dụng lúc runtime. RFC này thêm codegen để construct + `match` chạy thật.

## 2. Representation

Một giá trị sum (TYPE_KIND_SUM, kind 6) là **con trỏ 8-byte by-address** tới heap box:

```
box: [ tag: i64 @ offset 0 ][ payload: 8 bytes @ offset 8 ]   // size 16, align 8
```

- `tag` = số thứ tự variant (0,1,2…) do typecheck gán (VariantInfo.tag).
- `payload` = 1 slot 8-byte (scalar / pointer / struct-pointer by-address).
- Đồng nhất quy ước aggregate-by-pointer (BUG#30): `type_is_aggregate(kind6)=true`
  nên VALUE là con trỏ 8-byte, 16 chỉ là kích thước cấp phát box.

Tái dùng tối đa máy sẵn có: vì box reg mang type = sum id, `field_offset(sum, idx)`
rơi vào nhánh fallback `idx*8` → field 0 @0, field 1 @8 (đúng), `field_size(sum)=8`.
Do đó constructor/match dùng thẳng OP_ALLOC/OP_SET_FIELD/OP_GET_FIELD với sum type id,
KHÔNG cần backing struct, KHÔNG sửa selector.

## 3. Lowering (air_builder.ax)

- **Constructor no-payload** (`Green`): `lower_ident`, nhánh `SYM_VARIANT` + kind==SUM →
  `lower_variant_construct(sym, 0)`.
- **Constructor có payload** (`Circle(r)`): `lower_call_expr`, SAU path tên Ok/Err/Some/None,
  nhánh `SYM_VARIANT` + kind==SUM → `lower_variant_construct(sym, payload_arg)`.
- `lower_variant_construct`: OP_ALLOC(sum_id) → box; OP_ICONST tag; OP_SET_FIELD field0=tag;
  nếu có payload: OP_SET_FIELD field1=payload. Trả box.
- **match** (`lower_match`, lower_stmt nhánh NODE_MATCH_STMT): OP_GET_FIELD field0 = tag (1 lần);
  mỗi arm: OP_ICONST tag + OP_EQ + OP_BRANCH(then/else); then-block bind payload
  (OP_GET_FIELD field1 → local_map_put) rồi lower body, jump merge; else-block tiếp arm sau.
  - `NODE_VARIANT_PAT` → tag + bind payload.
  - bare `NODE_BINDING_PAT` mà tên LÀ variant → tag, no-bind (vì grammar coi bare-ident là
    binding; phải resolve theo tên với variant của sum type lúc codegen).
  - bare `NODE_BINDING_PAT` tên KHÔNG phải variant → catch-all, bind toàn scrutinee.
  - `NODE_WILDCARD_PAT` `_` → catch-all.

## 4. Built-in Option/Result vs user sum

Option/Result cũng khai báo `type Option[T]=Some(T)|None` (std/result.ax) nên Some/Ok/
Err/None là SYM_VARIANT kind SUM — **chỉ phân biệt với user sum bằng TÊN**. Path tên
Ok/Err/Some/None (pointer-layout: None=null, Err=ptr|1, payload boxed @0) chạy TRƯỚC và
return; user-variant path đặt SAU + guard kind==SUM. Vì vậy:
- Option/Result giữ nguyên pointer-layout (is_ok/unwrap/is_some… không đổi).
- User KHÔNG được đặt tên variant Ok/Err/Some/None (sẽ bị built-in nuốt).
- `match` trên Option/Result (kind 11/GENERIC_INST 8) bị `is_sum` guard bỏ qua → KHÔNG
  lower (dùng method như cũ); KHÔNG mis-compile.

## 5. Tại sao an toàn cho self-host

Compiler tự host KHÔNG dùng `match` thật và KHÔNG dùng user sum (chỉ Result/Option qua
method + if/elif). Constructor path mới chỉ kích hoạt cho SYM_VARIANT kind==SUM không
trùng tên built-in; Result/Option đi path tên cũ. Nên stage1→stage2 không đổi hành vi →
fixpoint phải giữ. Đã verify: t_builtin_opt=15 (Option/Result OK), regression 19/19.

## 6. Hạn chế (follow-up)
- Multi-field variant (`Rect(i64,i64)`): VariantInfo lưu 1 payload_type, box 1 slot → chưa hỗ trợ.
- Payload str / >8 byte: field_size(sum)=8 → mất byte. (scalar/pointer 8B OK.)
- Generic user sum instance (kind 8) + match trên Option/Result: chưa lower.

## 7. Test plan / DoD
- [x] t_enum_np=6, t_enum=42, t_adt2=104, t_adt3=19, t_builtin_opt=15.
- [x] regression_repros.sh 19/19 PASS.
- [ ] verify_bug29_selfhost.sh: fixpoint stage3==stage4 GIỮ.
