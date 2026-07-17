---
name: bug88-str-concat-operator
description: "BUG#88 FIXED — `str + str` silently miscompiled (OP_IADD on 16-byte str values → garbage/crash) instead of concatenating; now lowered to std.string.concat. std/os.ax `s + \".\" + ext` intends concat."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#88 FIXED** ✅ 9b45355 (2026-07-09, air_builder, **A!=B/B==C**, regression 114/114). ⚠️ **commit local, push BỊ CHẶN auth env** (token hết hạn giữa phiên — chờ push lại). Tìm bằng proactive probing (batch #9).

## Triệu chứng
`let c = a + b` (a,b: str) → `std.string.len(c)` = 127/crash (đáng lẽ concat). `std.string.concat(a,b)` tường minh chạy đúng (6). str `+` SAI ÂM THẦM.

## Root cause
`str` là PRIMITIVE (TYPE_STRING=12), KHÔNG phải STRUCT/SUM/GENERIC_INST → RFC 0007 operator-overload path (air_builder L716-732) BỎ QUA → `+` rơi xuống eager OP_IADD trên 2 giá trị str 16-byte (inline) → cộng con trỏ → str rác → len rác/crash. Không diagnostic.

## Fix (implement, KHÔNG reject)
Ngôn ngữ ĐỊNH `+` concat string (std/os.ax dùng `s + "." + ext`; compiler dùng `"ax_"+name`/`"."+field_name` mangling). Reject sẽ vỡ os.ax. → `lower_binary_expr`: nếu op `+` và cả 2 operand TYPE_STRING → `resolve_op_method("concat", TYPE_STRING, TYPE_STRING)` + `lower_op_overload` (tái dùng call-emission path). ~12 dòng.

## Gate A!=B/B==C
Compiler tự dùng str+ (linker.ax:965 `"ax_"+name`, intern.ax:107/typecheck.ax:725 `"."+field_name` — MANGLING) → đổi self-codegen → A!=B transition (như [[bug85-scalar-field-address]]); gate=B==C tay (`0D672CC8…`) + regression 114/114 (name-mangling tests t_method/t_modcollide PASS → mangling B đúng) + oracle. Các site này self-host được TRƯỚC fix (có thể dead/hiếm-hit path); sau fix đúng+hội tụ.

## Ghi chú
- Oracle bin/t_strconcat.ax (15 = len6+len7+char2). Chained `c + "!"` + mixed var/literal OK.
- Chỉ `+` string; các op str khác (`-`,`*`...) vẫn OP_IADD rác — nhưng vô nghĩa cho str, user khó gặp; `==` str đã handle riêng (st2 OK). Cùng họ silent-op-miscompile [[bug68-struct-eq-no-overload]] (nhưng #68 REJECT vì không có intent, #88 IMPLEMENT vì có intent+std dùng).
