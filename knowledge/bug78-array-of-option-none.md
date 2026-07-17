---
name: bug78-array-of-option-none
description: "BUG#78 ARRAY-LITERAL PART FIXED (091908b): `[Some(3), None]` segfaulted at construction — array OP_STORE block-copied 16 bytes from the pointer-repr Option value; None (null) read from address 0. Fixed with type_is_pointer_repr (kind 6/11/12, OR kind-8 GENERIC_INST whose base name is Option/Result — Some/None infer to GENERIC_INST 'Option' not kind-11) → store/load 8-byte pointer in OP_STORE(array)+OP_INDEX, mirroring field_is_pointer_sum (BUG#59). A==B fixpoint, regression 106/106. REMAINING: Vec[Option[T]] with a None element still faults via a different storage path (v.push(None)) — was already broken (not a regression), needs a separate fix."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#78 — ARRAY-LITERAL PART FIXED 2026-07-06, 091908b. Vec[Option] part CÒN MỞ.**

## FIX (091908b) — array literal of Option
Thêm `type_is_pointer_repr(table, pool, type_id)` = kind 6/11/12 **HOẶC** kind-8 GENERIC_INST
mà `extract_qualified_base_name(name)` ∈ {"Option","Result"}. (Some/None infer thành
GENERIC_INST tên "Option" qua BUG#67, KHÔNG phải kind-11 — đây là điểm mấu chốt; kind 8 cũng
là generic struct by-value thật như Pair nên PHẢI check tên). Dùng ở OP_STORE (nhánh array
src2!=0) + OP_INDEX → store/load con trỏ 8B thay vì block-copy/address. A==B fixpoint (không
đụng self-codegen), regression 106/106, oracle `bin/t_arroption.ax`. all-Some/all-None/mixed
đều chạy.

## CÒN MỞ — Vec[Option[T]] với None (lead cho session sau)
`Vec[Option[i32]].new(); v.push(None); v.get(i).unwrap()` VẪN segfault — **đã hỏng TỪ TRƯỚC**
fix này (test trên committed compiler cũng crash), KHÔNG phải regression. Vec dùng đường lưu
KHÁC (push/get method + buffer heap, nested Option[Option[i32]] từ get) không đi qua OP_STORE
array-path đã sửa. Cần điều tra Vec's element storage riêng (có thể cùng type_is_pointer_repr
nhưng ở lowering khác) + xử nested Option[Option]. Workaround: tránh None trong Vec[Option].

## Root cause gốc (giữ lại)

## Triệu chứng
```
let arr = [Some(3 as i32), None]     // hoặc [None, None], hoặc [Some(3), None, Some(7)]
return 7 as i32                       // CHỈ construct thôi -> SEGFAULT runtime
```
- `[Some(3), Some(7)]` (toàn Some) → OK.
- Có BẤT KỲ `None` element → segfault NGAY khi construct (trước mọi access). Index phần tử
  Some khác trong cùng mảng cũng crash (None làm hỏng cả construction).
- `None` đứng riêng (`let o: Option[i32] = None; match o`) → OK. Chỉ None-trong-array-literal.
- PRE-EXISTING, KHÔNG phải regression BUG#77 (type_is_aggregate + OP_STORE không bị BUG#77 đụng;
  105/105 regression + mọi test Option xanh với compiler BUG#77).

## Root cause
`air_builder::lower_array_lit` (BUG#70) store mỗi element bằng `OP_STORE type_id=elem_type`.
Element = Option → `type_is_aggregate`=true + `size`=16 → OP_STORE **block-copy 16 byte TỪ giá
trị Option** như thể nó là địa chỉ. Nhưng Option là POINTER-REPR (giá trị LÀ con trỏ box 8B):
- Some(3) → con trỏ box non-null → block-copy đọc 16B từ box → sai repr nhưng không crash.
- None → con trỏ NULL (0) → block-copy đọc 16B từ **địa chỉ 0** → **SEGFAULT**.
Đây là bản-array của BUG#59 (`field_is_pointer_sum`: field sum/option/result lưu con trỏ 8B,
không block-copy).

## Fix attempt (ĐÃ REVERT) & vì sao khó
Thêm `store_is_ptr_sum`/`idx_is_ptr_sum` (kind 6/11/12) vào OP_STORE + OP_INDEX để store/load
con trỏ 8B thay vì block-copy/address. **KHÔNG hiệu quả**: `Some(x)`/`None` tạo Option dạng
**GENERIC_INST (kind 8)** tên "Option" (qua `try_instantiate_variant_call`, BUG#67), KHÔNG phải
kind-11 OPTION → guard 6/11/12 trượt. Không thể thêm kind 8 vào guard vì **generic struct
by-value thật (Pair[i32,i64]) CŨNG kind 8** — sẽ phá chúng. Cần phát hiện "GENERIC_INST mà THỰC
CHẤT là Option/Result" bằng tên (fragile) — đúng lãnh thổ rối repr Option (BUG#57/60/67) từng
mất nhiều session. Revert sạch (git checkout, khớp trạng thái BUG#77 committed).

## Hướng fix tương lai (session focus riêng)
1. Xác định cơ chế chuẩn phân biệt "pointer-repr type" (Option/Result/sum ở MỌI dạng: kind
   11/12/6 VÀ GENERIC_INST kind 8 tên Option/Result). Có thể cần helper
   `type_is_pointer_repr(table, type_id)` dùng chung — kiểm kind 6/11/12 HOẶC (kind 8 AND
   base-name ∈ {Option, Result}). Grep xem air_builder đã có cơ chế tương tự chưa (BUG#57
   register_option/`match_arms_tagged_kind` phân loại theo TÊN arm — tham khảo).
2. Dùng helper đó ở OP_STORE (src2!=0 array path) + OP_INDEX để store/load con trỏ 8B.
3. Gate: fixpoint (có thể A!=B nếu đụng self-codegen → check B==C tay) + regression + oracle
   `[Some,None,Some]` array + Vec[Option[T]] (đảm bảo không phá).

## Phạm vi
Hẹp: chỉ ARRAY LITERAL / fixed-array của Option/Result (mảng có None/Err). Vec[Option[T]]
dùng .push/.get (heap, có thể cùng OP_STORE/OP_INDEX — cần kiểm). KHÔNG ảnh hưởng self-host
(compiler không dùng array-literal-of-Option). Workaround: dùng Vec thay array literal, hoặc
tránh None trong array literal.

Liên quan: [[bug57-match-option-native]] + BUG#67 (repr Option GENERIC_INST-vs-kind-11 — gốc
của khó khăn), BUG#59 (field_is_pointer_sum — bản-field của cùng vấn đề), [[bug70-array-literal-shipped]]
(array literal lowering nơi lỗi nằm).
