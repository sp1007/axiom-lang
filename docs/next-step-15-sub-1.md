# next-step-15 / sub-1 — ADT codegen native (BUG#32)

Theo dõi việc implement codegen cho user-defined sum type / `enum` (RFC 0004).
Bối cảnh: `enum` sugar (RFC 0003) đã xong nhưng vô dụng runtime vì backend native
KHÔNG lower được match + variant constructor (BUG#32, knowledge/bugs.md).

## Phát hiện (2026-06-26)
- `match` có node (NODE_MATCH_STMT=21, NODE_MATCH_ARM=22), parser+resolver+typecheck
  XỬ LÝ đầy đủ, nhưng `air_builder.lower_stmt` (1686) THIẾU nhánh → match sinh 0 AIR.
- Variant constructor: `air_builder` (1033) CHỈ hard-code Ok/Err/Some/None. User variant
  (Red/Circle) → rơi vào call path → `ax_Circle` unresolved.
- **Compiler tự host KHÔNG dùng match thật** (các "match" trong source chỉ là substring
  tên hàm/comment) và KHÔNG dùng user sum type → thêm codegen này KHÔNG đụng fixpoint.
- Result/Option dùng METHOD (is_ok/unwrap, `ax_sum_layout_is_pointer`), KHÔNG dùng match.
  → ADT codegen mới chỉ phục vụ TYPE_KIND_SUM (user). match trên Result/Option = follow-up.
- Type table ĐÃ lưu đủ: `VariantInfo{name_id, payload_type, tag}` (tag tuần tự 0,1,2…),
  `register_sum_type` (typetable.ax:288), điền tại typecheck.ax:940 (pre_infer_sum).
  variant symbol.type_id = sum type id. **HẠN CHẾ:** chỉ lưu 1 payload_type/variant
  (đọc first_child) → multi-field `Rect(i64,i64)` lossy ở type table.

## Thiết kế (RFC 0004) — representation tag-in-box
- Giá trị TYPE_KIND_SUM = con trỏ 8-byte tới heap box (đồng nhất aggregate-by-pointer, BUG#30).
- Box layout: `word[0]=tag (i64)`, `word[1..]=payload slots` (8B scalar/ptr, 16B str).
- Box size = 8 + payload_area; payload_area = max over variants của payload size.
- **v1 scope (tối thiểu, đúng, test được):** no-payload + single-payload variant.
  Multi-field (`Rect(i64,i64)`) → diagnostic "chưa hỗ trợ", follow-up (cần mở rộng
  VariantInfo lưu list payload). match trên Result/Option = follow-up.

## Lowering
1. **Constructor** (lower_expr NODE_IDENT no-payload / lower_call payload):
   nhận diện callee là variant của TYPE_KIND_SUM → tag; OP_ALLOC box; OP_STORE tag@0;
   nếu có payload: lower arg, OP_STORE @8; trả box ptr.
2. **match** (lower_stmt NODE_MATCH_STMT): lower scrutinee→box ptr; OP_LOAD tag@0;
   mỗi arm: so tag==variant_tag → block arm; bind payload (OP_LOAD @8 → local); lower body;
   jump end. Wildcard/binding = default. Merge block cuối.

## Checklist
- [x] Helper find_variant_info (air_builder): tra tag+payload_type theo variant name_id.
- [x] typetable: sum entry size=16, align=8 (box [tag@0, payload@8]).
- [x] Constructor lowering no-payload (lower_ident SYM_VARIANT kind SUM) + single-payload
      (lower_call SYM_VARIANT kind SUM, SAU path Ok/Err/Some/None để giữ pointer-layout).
- [x] match lowering (lower_match): OP_GET_FIELD tag@0; mỗi arm OP_EQ+OP_BRANCH;
      payload bind = OP_GET_FIELD field1; wildcard/non-variant-binding = default catch-all;
      bare-ident-là-variant = tag match no-bind. Chỉ kind-6 SUM (is_sum guard).
- [x] Test pass: t_enum_np=6, t_enum=42, t_adt2=104 (sumtype+wildcard+payload),
      t_adt3=19 (catch-all), t_builtin_opt=15 (Option/Result built-in KHÔNG hỏng).
- [~] regression_repros.sh (đang chạy, +5 repro ADT).
- [ ] RFC 0004 (rfcs/0004-adt-codegen.md).
- [ ] verify_bug29_selfhost.sh: fixpoint stage3==stage4 GIỮ (~2.5h).
- [ ] Commit+push; cập nhật knowledge/bugs.md (đóng BUG#32 phần v1).

## Quyết định kỹ thuật quan trọng
- **Representation:** sum value = 8-byte pointer tới heap box [tag i64 @0, payload 8B @8],
  size=16. Dùng thẳng sum type id cho OP_ALLOC/SET_FIELD/GET_FIELD: get_register_type
  trả sum id, field_offset(sum,idx) rơi vào fallback `idx*8` → field0@0, field1@8 ĐÚNG.
  field_size(sum)=8. type_is_aggregate(kind6)=true → value vẫn 8-byte by-address pointer.
- **Built-in vs user:** Option/Result CŨNG là `type Option=Some(T)|None` (Some/Ok là
  SYM_VARIANT kind SUM). Phân biệt CHỈ bằng TÊN (Ok/Err/Some/None → pointer-layout path,
  giữ nguyên). User variant path đặt SAU path tên + guard kind==SUM. ĐỪNG reorder lên
  trước (sẽ chặn nhầm Some/Ok → is_some/unwrap đọc sai layout → bug 3 thay vì 15).
- **Tên trùng:** user KHÔNG được đặt variant tên Ok/Err/Some/None (sẽ bị built-in nuốt).

## Hạn chế v1 (follow-up)
- Multi-field variant `Rect(i64,i64)`: type table chỉ lưu 1 payload_type; box chỉ 1 slot.
- str/16-byte payload: field_size(sum)=8 → mất 8 byte. (scalar/pointer 8B OK.)
- Generic user sum instance (kind 8 GENERIC_INST) + match trên Option/Result (kind 11/8):
  is_sum guard bỏ qua → match KHÔNG lower (dùng is_ok/unwrap như cũ). KHÔNG mis-compile.

## Tiến độ
- 2026-06-26: implement xong v1, 5/5 test ADT + built-in pass. Đang chạy regression gate,
  rồi RFC 0004 + fixpoint verify.
