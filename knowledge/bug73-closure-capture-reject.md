---
name: bug73-closure-capture-reject
description: "BUG#73 FIXED (56ec6cc): closures capturing an outer-scope variable (`|x| base + x` where `base` is from the enclosing function) silently computed WRONG results (not a crash) — RFC 0008 P1 shipped zero-capture lambda-lifting only, P2 (real capture) is still Draft/unimplemented; a captured identifier resolved post-lift via an unscoped global symbol scan to an unrelated variable's stale register slot. Fixed with a free-variable scan in lift_closures rejecting any body identifier that isn't the closure's own param or a known top-level global. Regression 100/100."
metadata:
  node_type: memory
  type: project
  originSessionId: ea9a6a4a-c908-4da1-80b2-8b5c3cfad2da
---

**BUG#73 FIXED — 56ec6cc, 2026-07-06.** Fourth dormant/half-shipped feature found via
proactive probing this session, after [[bug70-array-literal-shipped]],
[[bug71-interface-dynamic-dispatch]], [[bug72-range-index-reject]] — but this one is
notably different: **a silent WRONG ANSWER, not a segfault.** The other three all crashed
(making them easy to notice); this one just quietly computed the wrong number, which is a
scarier class of bug (no signal at all that something is wrong).

## Triệu chứng

```
fn test() -> i32:
    let base = 100 as i32
    let add = |x: i32| -> i32 base + x
    return add(7 as i32)     // trả về 14 (7+7), ĐÚNG phải là 107 (100+7)
```
Không segfault, không lỗi biên dịch — chỉ SAI KẾT QUẢ âm thầm. Test đổi hằng số để phân
biệt "đọc x hai lần" (14=7+7) khỏi "đọc rác tình cờ đúng x" xác nhận: closure tính `x + x`,
hoàn toàn bỏ qua giá trị `base` thật.

## Root cause

**RFC 0008 (`rfcs/0008-closures-with-capture.md`) Status line ghi RÕ**: "P1 Implemented
(zero-capture lambda-lift); **P2 capture = Draft**" — capture THẬT SỰ CHƯA TỪNG ĐƯỢC LÀM,
chỉ mới thiết kế trên giấy. `lift_one_closure` (parser.ax) — pass "lambda-lifting" chạy
SAU parse, TRƯỚC resolution — chỉ RE-PARENT thân lambda thành 1 hàm top-level MỚI, KHÔNG hề
sinh env struct, KHÔNG rewrite tham chiếu biến ngoài. `t_lambda.ax` (test RFC 0008 P1 sẵn
có) tự ghi chú "RFC 0008 zero-capture lambdas" — TẤT CẢ 5 case của nó chỉ dùng tham số CỦA
CHÍNH lambda, không case nào tham chiếu biến ngoài — nên gap này chưa từng lộ ra.

Khi identifier "base" (không phải param của lambda) được resolve TRONG hàm lifted mới:
`SymbolTable::resolve_global` (resolver.ax:616) — **quét TUYẾN TÍNH KHÔNG PHÂN BIỆT SCOPE**
qua TOÀN BỘ `self.symbols` — tìm thấy symbol "base" (declare ở hàm `test`, KHÁC hàm lifted)
và trả về ĐÚNG symbol_idx đó (không sai theo nghĩa "tìm nhầm tên"), NHƯNG symbol này gắn với
STACK FRAME/REGISTER của hàm `test` (đã kết thúc, không còn liên quan) — codegen của hàm
LIFTED MỚI đọc reg/slot số đó theo cách đánh số CỤC BỘ của CHÍNH NÓ, tình cờ trùng với slot
của "x" (tham số duy nhất của lambda) → đọc nhầm x thay vì base.

## Fix — free-variable scan TRƯỚC KHI lift, reject nếu có capture

Thêm `scan_closure_free_vars` + `check_closure_no_capture` (parser.ax, trước
`lift_closures`): với MỖI closure, TRƯỚC khi lift, duyệt thân lambda tìm mọi `NODE_IDENT`;
nếu identifier KHÔNG phải (a) 1 trong các param CỦA CHÍNH closure, VÀ KHÔNG phải (b) 1 tên
GLOBAL đã biết (quét trực tiếp children của program root cho
FUNC_DECL/STRUCT_DECL/INTERFACE_DECL/CONST_DECL/TYPE_ALIAS_DECL TRƯỚC khi duyệt closure nào)
→ đây là CAPTURE → lỗi rõ ràng + đếm. `lift_closures` đổi signature từ `void` sang trả về
`i64` (tổng lỗi); `main_air.ax` check giá trị này NGAY SAU lời gọi, HALT giống hệt pattern
`parser.diags_count` (BUG#53).

**Không cần symbol table** (chạy TRƯỚC resolution) — chỉ cần quét AST thô, tận dụng
`NODE_TYPE_EXPR` (không phải `NODE_IDENT`) tách biệt namespace type khỏi namespace value nên
không cần lo tên kiểu (`i32`, ...) bị nhầm là capture.

**Verify**: case segfault-kiểu-sai-kết-quả gốc → giờ lỗi biên dịch rõ ràng. `t_lambda.ax`
(5 case zero-capture, RFC 0008 P1) vẫn PASS nguyên (exit 5, không false-positive). Closure
gọi HÀM GLOBAL + đọc CONST global (hợp lệ, không phải capture) vẫn chạy đúng (test riêng,
không phải capture — 3*2+100=106). Compiler tự-host (~855 hàm, nhiều closures nội bộ nếu
có) build sạch — xác nhận KHÔNG có capture-closure nào tồn tại trong chính compiler/stdlib
hiện tại (an toàn, fixpoint không bị ảnh hưởng). Fixpoint A==B, regression 100/100 (test mới
`bin/t_closurecap.ax` case hợp lệ + `tests/sema/err_closure_capture.ax` case reject).

**Ngụ ý**: closures trong AXIOM hiện tại CHỈ dùng được ở dạng zero-capture (pure function
literal, `|params| -> Ret expr` chỉ tham chiếu params của chính nó + global). Muốn capture
biến ngoài (RFC 0008 P2: env struct, capture-by-value, heap alloc) — CHƯA implement, cần
follow-up riêng theo đúng thiết kế đã có trong RFC 0008 §2.2-2.3.

Liên quan: [[bug70-array-literal-shipped]], [[bug71-interface-dynamic-dispatch]],
[[bug72-range-index-reject]] (cùng session, cùng pattern "tính năng half-shipped lộ qua
proactive probing"); khác các bug#70-72 ở chỗ đây là SAI KẾT QUẢ ÂM THẦM chứ không phải
crash — nhắc nhở: luôn thử NHIỀU GIÁ TRỊ khác nhau khi verify một fix/probe (đổi hằng số để
phân biệt "tình cờ đúng" khỏi "thực sự đúng"), đừng chỉ tin 1 test case dễ trùng hợp.
