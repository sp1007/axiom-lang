---
name: axiom-struct-reference-semantics
description: "DESIGN (not a bug): AXIOM structs/aggregates use REFERENCE semantics — `mut cpy := src`, `b = a`, passing struct to fn all ALIAS the same storage, no implicit copy. Decided in RFC 0001 §5. Do NOT 'fix' this."
metadata:
  node_type: memory
  type: reference
  originSessionId: 044ec622-2518-45eb-9368-07febdfca8f1
---

**AXIOM struct/aggregate = REFERENCE semantics (INTENTIONAL, RFC 0001 §5 lines 129-137).** KHÔNG phải bug.

## Behavior (trông như bug nhưng ĐÚNG)
- `mut cpy := src` (src struct) → `cpy` ALIAS `src`, KHÔNG copy. Mutate `cpy.field` đổi luôn `src.field`.
- `b = a` (assignment struct) → alias, không copy.
- `fn f(b: Box)` truyền struct → callee nhận con trỏ tới object gốc; mutate trong callee ẢNH HƯỞNG caller.
- Array element `mut t := arr[i]` (element struct) → alias arr[i].
- **Scalar** (i32/u64/bool/f64...) → copy bình thường (value). Chỉ AGGREGATE mới reference.

## Vì sao (RFC 0001 §5)
"Quyết định: struct dùng REFERENCE semantics nhất quán (không memcpy bản sao). Truyền con trỏ 8-byte tới object gốc là ĐÚNG và nhất quán. §4.2 (memcpy bản sao) bị BÁC vì sẽ tạo bất nhất với assignment-alias." Aggregate = by-address ở MỌI chỗ (BUG#30/Family C); binding/param/return đều truyền con trỏ 8B, không copy nội dung. Compiler tự dựa vào điều này (dùng `&expr` khi muốn alias tường minh nhưng binding cũng là reference).

## HỆ QUẢ khi probing
- Probe kiểu "mut copy := struct; mutate copy; kỳ vọng gốc không đổi" sẽ THẤY gốc đổi — **KHÔNG BÁO BUG**, đó là design. Đã suýt "fix" (2026-07-09) → check spec kịp, tránh phá self-host.
- Value-type thật (copy-on-assign) = feature tương lai `Copy[T]`/explicit clone, cần RFC RIÊNG, KHÔNG sửa ngầm ABI.
- Liên quan RFC 0001 (16-byte aggregate by-value ABI), [[bug77-16byte-struct-byaddress-unified]] (aggregate by-address), [[bug51-hunt-progress]] (RFC 0010 aggregate-alias landmine — cùng gốc reference-semantics).
