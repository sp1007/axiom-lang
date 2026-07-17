---
name: bug77-16byte-struct-byaddress-unified
description: "BUG#77 FIXED (ee02b72): the ROOT unification of the 16-byte-struct representation split behind BUG#75/76. A 16-byte struct from OP_GET_FIELD was held INLINE while the same struct from OP_ALLOC/param was a by-address POINTER; crossing the two faulted (nested `c.b.a.v`, array field, struct-field return). Fix: regalloc_is_16byte's catch-all + OP_GET_FIELD size==16 path now use `not type_is_aggregate` (like the param path and every other op case already did) → 16-byte aggregates are by-address EVERYWHERE; only `str` (primitive) stays inline. Supersedes BUG#75's array-only field_is_array guard (removed). First fix this session where A!=B (compiler self-codegen changed) → verified via B==C. Regression 105/105, tstruct_abi intact."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#77 — FIXED 2026-07-06, ee02b72. ROOT cause chung của cả cụm BUG#75/76/77.**

## Nguyên lý (quan trọng, nhớ cho mọi bug 16-byte tương lai)
`str` (primitive 16B, {ptr,len}) là thứ DUY NHẤT held INLINE-16 (giá trị nằm thẳng trong
slot/RAX:RDX). MỌI **aggregate** 16B (struct/array/tuple/generic-inst/sum/option/result) held
**BY ADDRESS** (reg giữ con trỏ 8B tới data). `regalloc_is_16byte` phải trả `true` CHỈ khi
`size==16 and not type_is_aggregate` — đây là quy tắc thống nhất mà **param path + COPY/CALL/
DEREF cases ĐÃ dùng sẵn**; chỉ còn catch-all else-branch (phân loại OP_GET_FIELD dest) là
lạc điệu, vẫn home aggregate 16B inline.

## Triệu chứng (facet thứ 3, mới lộ)
```
struct A: v: i64
struct B: a: A; w: i64        // 16 byte
struct C: b: B; z: i64
fn main(): ... c.b.a.v         // SEGFAULT
```
`c.b` (field struct 16B) load INLINE → `.a` làm pointer+disp LEA trên data inline → địa chỉ
rác → fault. `let bb = c.b; bb.a.v` crash y hệt. 2-level (`b.a.v` trên local literal) OK vì
literal → OP_ALLOC → con trỏ; chỉ khi struct 16B đến từ GET_FIELD mới inline → lệch repr.

## Fix
1. `regalloc_is_16byte` else-branch (catch-all, phân loại GET_FIELD dest): `not is_tagged_ptr
   and not is_byaddr_array` → **`not type_is_aggregate(table, type_id)`** (thống nhất với mọi
   case khác).
2. `OP_GET_FIELD` size==16 branch: `not field_is_array` → **`not field_is_aggregate`** (struct
   16B field cũng trả ĐỊA CHỈ, không copy inline).
3. Xóa `field_is_array` (BUG#75 tạo, giờ thừa — quy tắc chung bao trùm).

## tstruct_abi — nỗi lo cũ ĐÃ LỖI THỜI
Comment BUG#56/64 cũ bảo "giữ narrow kẻo vỡ tstruct_abi register-pair ABI". Thực tế: param
path ĐÃ treat struct 16B by-address, nên làm GET_FIELD nhất quán chính là điều register-pair
path muốn. tstruct_abi (A=7 B=12 C=15 D=6 E=99) XANH. Nỗi lo cũ không còn đúng.

## Bài học fixpoint (QUAN TRỌNG — khác BUG#74/75/76)
Đây là fix ĐẦU TIÊN session này đổi **codegen của CHÍNH compiler** (compiler source có field-
access struct 16B) → **A != B là ĐÚNG** (bootstrap transition, không phải lỗi). `fast_fixpoint
.ps1` báo "FAILURE A!=B" nhưng đó là script chỉ check A==B (đủ cho fix không đụng self-codegen
như BUG#74/75/76). Với fix đụng self-codegen: **phải build C từ B tay và check B==C** (gate
thật). B==C OK + regression 105/105 bằng B = an toàn. (BUG#74/75/76 A==B vì chúng chỉ sửa
codegen cho CODE DÙNG feature, không phải compiler tự-compile.)

Liên quan: [[bug75-struct-array-field]] (array facet, giờ là case riêng của quy tắc chung),
[[bug76-return-16byte-struct-field]] (return facet), [[bug64-vec-big-aggregate-element]] +
BUG#56/60 (cùng cụm size==16-vs-aggregate). Oracle `bin/t_nested3field.ax`.
