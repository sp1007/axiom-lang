---
name: bug74-generic-struct-inferred-ctor-args
description: "BUG#74 FIXED (8ce470f, 2026-07-06): a user-defined multi-param generic struct (`Pair[A,B]`) constructed WITHOUT explicit type args (`Pair(first:.., second:..)`, field-value inference) silently computed WRONG field values when passed through a generic function — every field past the first read 0/garbage. Two linked defects: (1) typecheck's inferred-ctor path never recorded generic_args (set_struct_generic_args was mono.ax-only) → callee defaulted type params to i32; (2) air_builder built a NODE_IDENT struct ctor with the TEMPLATE type (generic 8-byte slots, second@8) not the monomorphized packed type (second@4). Fixed by finish_generic_instantiation + try_instantiate_struct_ctor (typecheck) and preferring node_types[idx] over sym.type_id (air_builder). Fixpoint A==B, regression 101/101."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#74 — FIXED 2026-07-06, commit 8ce470f** (found 2026-07-06 via proactive
probing after BUG#70-73). A self-written 2-param generic struct
(`struct Pair[A,B]`) constructed via field-value inference miscomputed fields.

## Triệu chứng
```
struct Pair[A, B]:
    first: A
    second: B
fn get_second[A, B](p: Pair[A, B]) -> B:
    return p.second
fn main() -> i32:
    let p = Pair(first: 3 as i32, second: 7 as i64)   // KHÔNG Pair[i32,i64] tường minh
    return get_second(p) as i32   // trả 0, đáng lẽ 7 — SAI ÂM THẦM (không crash)
```

Ma trận thu hẹp (quan trọng cho chẩn đoán): `get_first` đúng; direct-read trong
main đúng; **chỉ field-thứ-hai-trở-đi qua HÀM GENERIC sai**. Sau fix phần 1, mọi
combo đúng TRỪ `Pair[i32,i32]` (struct đúng 8 byte) → lộ ra defect thứ 2.

## Root cause — HAI defect liên kết

**Defect 1 (typecheck — generic_args không được ghi):** đường construct suy-luận
KHÔNG BAO GIỜ gọi `set_struct_generic_args`; nó chỉ đạt tới qua đường type-args
tường minh (`mono.ax:422`). Nên `get_generic_args(Pair-instance)` rỗng →
`infer_generic_type_args` bind 0 param → `get_second[A,B]` default A=B=i32 (fallback
BUG#66) → monomorph hóa callee SAI layout.

**Defect 2 (air_builder — dùng TEMPLATE thay vì mono type):** kể cả sau khi (1) fix
để typecheck ghi type mono lên call-node, `lower_call_expr` build struct-ctor
NODE_IDENT bằng `sym.type_id` (TEMPLATE, field-slot generic 8-byte → second@8) chứ
KHÔNG dùng `node_types[idx]` (mono packed → second@4). Giá trị dựng ở second@8 còn
callee đọc packed second@4. **Che giấu ở MỌI combo mà second rơi vào offset 8 ở cả
hai layout** (i32/i64, i64/i32, i64/i64) — chỉ `Pair[i32,i32]` (packed second@4)
lệch → trả 0. Đường tường minh `Pair[i32,i32]` (callee NODE_GENERIC_TYPE, dòng
1489-1496) VỐN dùng `node_types` nên luôn đúng → chính là manh mối phân biệt.

## Fix (8ce470f)

- **typecheck.ax**: tách lõi instantiate của đường tường minh (`Name[T,U]`) thành
  `finish_generic_instantiation(sym_idx, args)` (nhận sở hữu `args`); thêm
  `try_instantiate_struct_ctor(callee, struct_type_id)` — suy type-args từ giá trị
  field-init (positional, khớp cách backend lower named-arg thành positional, xem
  BUG#21), rồi chạy CÙNG monomorphization → construct và param hàm generic nhận nó
  chia sẻ MỘT packed layout. Gọi ở nhánh `TYPE_KIND_STRUCT` của NODE_CALL_EXPR.
  Non-generic struct / không bind được gì → trả template nguyên (no-op).
- **air_builder.ax**: nhánh NODE_IDENT struct-ctor giờ ưu tiên `node_types[idx]`
  (type mono) hơn `sym.type_id` (template), guard `ct != sym.type_id` +
  kind∈{STRUCT,GENERIC_INST} nên non-generic struct KHÔNG đổi.

## Bài học phương pháp luận
- **Ma trận combo type là chìa khóa**: fix defect-1 xong tưởng hết, nhưng test
  i32/i32 vs i64/i64 vs i64/i32 lộ ra rằng "hết bug ở combo A" ≠ "hết bug"; combo
  8-byte-đúng-mức che defect ABI thứ 2. Luôn thử combo có kích thước field KHÁC nhau.
- So sánh đường **explicit vs inferred** cho CÙNG type args (cả hai qua
  finish_generic_instantiation) khoanh vùng defect-2 vào air_builder chính xác.
- `ax_printf_local` nuốt prefix `[D`/`[T`/... — dùng `XTRACE` (xem
  [[fast-fixpoint-workflow]]).

## Phạm vi
KHÔNG ảnh hưởng self-host/stdlib (HashMap/Vec/Option/Result đều `.new()` type-args
tường minh — luôn đúng). Chỉ sửa/mở khóa code người dùng tự viết generic struct đa
tham số construct suy-luận. Fixpoint A==B, regression 101/101, oracle
`bin/t_genericstruct2.ax` (phủ ca i32/i32 packed-8-byte + i64 mixed).

Liên quan: [[bug70-array-literal-shipped]], [[bug71-interface-dynamic-dispatch]],
[[bug72-range-index-reject]], [[bug73-closure-capture-reject]] (chuỗi
proactive-probing); [[bug66-hashmap-i64-value-corruption]] (cùng fallback
unbound→i32); RFC 0013 / [[bug64-vec-big-aggregate-element]] (cơ chế generic_args).
