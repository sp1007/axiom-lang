---
name: bug76-return-16byte-struct-field
description: "BUG#76 FIXED (cdc084c): `fn f(o: Outer) -> Inner: return o.inner` segfaulted in the CALLER when Inner is a >=16-byte struct FIELD returned directly. Root cause: callee/caller disagreed on the 16-byte-struct return ABI — caller classifies a 16-byte aggregate return as an 8-byte POINTER (RFC 0001), OP_ALLOC-returns already match, but a GET_FIELD-of-16B-struct-field result is classified 16-byte-INLINE so OP_RETURN sent field halves in RAX:RDX while caller read RAX as a pointer. Fix: OP_RETURN keys the convention off the RETURN TYPE (LEA pointer for a 16-byte aggregate return), not the value's incidental classification. tstruct_abi intact; fixpoint A==B, regression 104/104."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#76 — phát hiện + FIXED 2026-07-06, cdc084c** (proactive-probing, cùng đợt với
[[bug75-struct-array-field]]). Ban đầu tưởng nested-generic-struct nhưng thu hẹp cho thấy
KHÔNG liên quan generics — là ABI trả struct 16B.

## FIX (cdc084c)
`OP_RETURN` giờ quyết định convention 16-byte theo **KIỂU TRẢ VỀ** (khớp caller), KHÔNG theo
`regalloc_is_16byte(src1)` (phân loại tình cờ của giá trị). Trả về aggregate 16B → LEA con
trỏ vào RAX (giống OP_ALLOC-return sẵn có); chỉ `str` (primitive inline-16) mới trả RAX:RDX.
Caller luôn coi aggregate-16B-return là con trỏ (RFC 0001, `not type_is_aggregate` ở CALL
case của regalloc_is_16byte) → giờ callee khớp. tstruct_abi nguyên (mkP trả OP_ALLOC=con trỏ,
không đụng nhánh mới). Fixpoint A==B, regression 104/104, oracle `bin/t_retstructfield.ax`.
Cũng fix luôn ca nested-generic (generic fn trả field generic-struct 16B) đã khơi ra bug này.

## Chi tiết root cause (giữ lại)

## Triệu chứng (đã thu hẹp chính xác)
```
struct Inner:
    a: i32
    b: i64          // Inner = 16 byte
struct Outer:
    x: i32
    inner: Inner
fn get_inner(o: Outer) -> Inner:
    return o.inner          // <-- SEGFAULT: trả TRỰC TIẾP field struct 16B
fn main() -> i32:
    let o = Outer(x: 1 as i32, inner: Inner(a: 3 as i32, b: 7 as i64))
    let gi = get_inner(o)
    return gi.b as i32
```

Ma trận thu hẹp:
- `Inner` = 8 byte (1 field): trả field trực tiếp → **OK**.
- Trả 1 struct 16B LOCAL (`fn make() -> Inner: return Inner(...)`) → **OK**.
- Bind field vào local rồi dùng (`let gi = o.inner; return gi.b`) → **OK** (workaround).
- **CHỈ**: trả TRỰC TIẾP field struct **>=16 byte** (`return o.inner`, GET_FIELD → thẳng
  OP_RETURN) → **SEGFAULT**.

## Vùng root cause (chưa pin-point)
`x86_selector.ax`: OP_GET_FIELD field struct 16B đi nhánh str-inline (size==16, KHÔNG phải
array nên [[bug75-struct-array-field]] fix không đụng — cố ý giữ struct trên nhánh register-
pair ABI cho tstruct_abi). Kết quả GET_FIELD (16B inline value) đưa THẲNG vào OP_RETURN
(`is_16byte` → LEA &src1 → load RAX/RDX). Lý thuyết cả hai đều coi 16-byte-inline nên "khớp",
nhưng thực tế crash — nghi regalloc/spill của temp vreg hoặc tương tác GET_FIELD-str-path ↔
OP_RETURN khi field ở offset != 0. Cần objdump-level tracing như BUG#60/64.

## Vì sao CHƯA fix (quyết định có chủ đích)
- Cùng lãnh thổ 16-byte-aggregate ABI cực nhạy (BUG#56/60/64/75; tstruct_abi dễ vỡ) —
  fix vội rủi ro cao, cần objdump + nhiều vòng.
- Workaround tầm thường (bind local trước). KHÔNG ảnh hưởng self-host (grep: compiler
  không có pattern `return <struct>.field` cho aggregate 16B field trực tiếp — dùng
  Option/Result/Vec, đều pointer-repr).
- Session đã ship 3 fix gated (BUG#74, RFC 0015 P1, BUG#75). Dừng đúng chỗ, để lại lead
  chính xác cho session focus riêng (giống cách [[bug74-generic-struct-inferred-ctor-args]]
  từng được document trước khi fix ở session sau).

**Điểm khởi đầu cho session fix:** build `/tmp/probe/n2.ax` (hoặc tái tạo ở trên), objdump
hàm `get_inner`, xem OP_RETURN 16-byte path (`x86_selector.ax:1366-1381`) + GET_FIELD str-
inline path (dòng ~1791) tương tác thế nào; so với n2_ret16 (local 16B, hoạt động) để thấy
khác biệt mã sinh. Prefix debug an toàn = `XTRACE` (KHÔNG `[D`, xem [[fast-fixpoint-workflow]]).
