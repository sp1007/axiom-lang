---
name: bug83-underscore-digit-separator
description: "BUG#83 FIXED — numeric digit separators '_' (1_000_000, 0x1234_0011) truncated the literal at the first '_' in BOTH integer parsers; now skipped correctly. std.time.Duration was silently wrong."
metadata:
  node_type: memory
  type: project
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**BUG#83 FIXED** ✅ 1fb6397 (2026-07-09, fixpoint A==B `2D0B219E…`, regression 110/110). Tìm ra bằng **proactive probing** (batch #5).

## Triệu chứng
Numeric literal có dấu `_` phân nhóm chữ số (`1_000_000`, `0x1234_0011`) bị **cắt cụt tại `_` đầu tiên**, SAI ÂM THẦM (không lỗi). `0x1234_0011 & 0xFF` → 0x34=52 (đáng lẽ 0x11=17), chứng minh literal parse thành `0x1234`. Lexer ĐÃ chấp nhận `_` trong token số (scan_dec/hex_digits, b==95) nên tạo token "1_000" đầy đủ → accept-then-miscompile.

## Root cause
HAI đường parse integer-literal, cả hai `break` vòng lặp khi gặp `_` (digit=-1):
- `air_builder.ax:parse_int_from_str` (~L127) — dùng lúc lowering emit hằng.
- `typecheck.ax:parse_comptime_int` (~L1040) — dùng lúc const-fold comptime.

## Fix
Cả hai: trước tính digit, `if c == '_' as u8: i = i + 1; continue` (skip separator, tiếp tục quét). `continue` được AXIOM hỗ trợ.

## Fixpoint an toàn
Bootstrap/stage1 KHÔNG có numeric `_` literal nào — mọi `_` là trong string/comment/identifier (`"x86_64"`, `// R_X86_64_64`, `EM_X86_64`) → parse_int không chạm → self-codegen không đổi → A==B như kỳ vọng (không phải backend transition). Numeric `_` literal THẬT duy nhất = **std/time.ax** (`1_000_000_000`, `1_000_000`, `1_000` cho Duration) — trước đây parse thành `1` → `std.time.Duration` SAI hoàn toàn; nay đúng. std/time.ax không thuộc compiler self-build nên không rủi ro fixpoint.

## Ghi chú
- Oracle: bin/t_underscore.ax (exit 27 = 17 hex-mask + 10 dec-scale). Regression row `t_underscore|exit|27`.
- Cùng họ silent accept-then-miscompile: [[bug80-free-call-overload-collision]], [[bug81-multifield-variant-payload]]. Bài học: khi lexer chấp nhận 1 lexeme mà parser/lowering không xử đủ → miscompile thầm; luôn kiểm cả 2 đường parse (air_builder lowering + typecheck comptime).
