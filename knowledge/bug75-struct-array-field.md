---
name: bug75-struct-array-field
description: "BUG#75 FIXED (ef90bd5): reading `b.data[i]` where a struct has a fixed-array field `data: [i32; 4]` segfaulted, and `b.data[i] = v` silently wrote to a throwaway copy — ONLY when the array field is EXACTLY 16 bytes (12- and 20-byte array fields worked). Twin of BUG#64: OP_GET_FIELD caught the 16-byte field in the str-style {ptr,len} inline-copy branch before the by-address branch, returning the array's bytes as an inline value instead of the field address. Fix scoped to arrays (kind 3): field_is_array helper excludes them from the size==16 GET_FIELD inline path AND from regalloc_is_16byte's inline home; 16-byte structs untouched (tstruct_abi). Fixpoint A==B, regression 103/103."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#75 — FIXED 2026-07-06, ef90bd5.** Tìm ra khi proactive-probing struct chứa
FIXED-ARRAY field (`struct Buf: data: [i32; 4]`).

## Triệu chứng
- `b.data[i]` (đọc) → **SEGFAULT**.
- `b.data[i] = v` (ghi) → không crash nhưng ghi vào BẢN SAO tạm (đọc lại thấy giá trị cũ).
- **CHỈ khi array field ĐÚNG 16 byte** (`[i32;4]`). `[i32;3]` (12B) và `[i32;5]` (20B)
  chạy đúng cả đọc lẫn ghi. Plain array local (`mut a := [...]`) luôn đúng. Struct
  scalar field luôn đúng.

## Root cause — sinh đôi với BUG#64

`x86_selector.ax::OP_GET_FIELD` dispatch theo THỨ TỰ: `size>16` LEA → **`size==16`
str-inline copy** → `field_is_aggregate` LEA. Field `[i32;4]` = đúng 16B → rơi vào nhánh
`size==16` (dành cho `str` {ptr,len} 16B inline) TRƯỚC nhánh by-address → copy 16 byte của
array vào slot như 1 giá trị inline. Rồi `b.data[i]` deref 2 i32 gói lại (vd 10|20) làm con
trỏ → fault. Y hệt BUG#64 (đã fix cho OP_INDEX) nhưng ở OP_GET_FIELD — chưa từng test vì
array-field-16B là tổ hợp mới.

Ghi vào copy: `regalloc_is_16byte` else-branch cũng home load array-field-16B thành inline-16
→ OP_STORE dùng `LEA &slot` (ghi vào copy) còn OP_INDEX `LOAD [slot]` (lấy địa chỉ) → bất
đối xứng: sau khi GET_FIELD trả địa chỉ thì ĐỌC đúng nhưng GHI vẫn vào copy.

## Fix — scoped CHỈ cho array (kind 3)

Thêm helper `field_is_array` (mirror `field_is_aggregate` nhưng chỉ kind==3). Dùng ở 2 chỗ:
1. `OP_GET_FIELD`: `elif size==16 and not field_is_array(...)` → array-field-16B rơi xuống
   nhánh `field_is_aggregate` (LEA trả ĐỊA CHỈ field).
2. `regalloc_is_16byte` else-branch: `not is_byaddr_array (tk==3)` → array-field-16B home
   8-byte (giữ địa chỉ), ĐỌC (OP_INDEX) và GHI (OP_STORE) đều MOV địa chỉ + index.

**Vì sao scoped array-only, KHÔNG đụng struct 16B:** array LUÔN by-address, KHÔNG dùng
register-pair by-value ABI (RFC 0001) mà struct 16B dùng. Comment trong regalloc cảnh báo
widening chung phá `tstruct_abi` (test struct 16B register-pair). Giữ struct nguyên → tstruct_abi
(A=7 B=12 C=15 D=6 E=99) vẫn xanh. `OP_GET_FIELD.type_id = node_types[field-expr]` = kiểu array
(kind 3, size 16) nên else-branch nhận diện được.

Fixpoint A==B, regression 103/103, oracle `bin/t_structarrfield.ax` (đọc + ghi-lại 16B array
field, exit 119). Liên quan: [[bug64-vec-big-aggregate-element]] (sinh đôi, cùng size==16-vs-
aggregate ordering), BUG#56/60 (cùng lãnh thổ 16-byte aggregate vs str/tagged-ptr repr).

## Lead CHƯA điều tra (cùng đợt probe, session này)
- **Nested generic struct** `Pair(first:.., second: <Pair value>)` rồi đọc inner qua generic
  fn → SEGFAULT (probe p1_nested). Có thể cùng họ BUG#74 hoặc aggregate-field. CHƯA đào.
- `type Color = Red|Green|Blue` + `match` trong hàm → parse "expected expression nud"
  offset lớn (probe p5_match). NGHI là syntax test của tôi sai (enum hoạt động ở t_enum.ax),
  chưa xác nhận.
