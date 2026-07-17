---
name: bug56-nested-sum-payload
description: "BUG#56 FIXED (10a0161) — native backend miscompiled extraction of a SUM-typed variant payload; root cause in regalloc_is_16byte (sum box=16B but reg holds 8B pointer). Fixpoint held with a NARROW fix."
metadata: 
  node_type: memory
  type: project
  originSessionId: 05d3f904-e67c-4f1a-9bfa-33caeb26ab45
---

**BUG#56 FIXED (commit 10a0161, pushed main 2026-07-04).** A variant whose PAYLOAD
is itself a sum — `type Outer = Wrap(Inner) | Empty`, `type Inner = A | B` —
miscompiled on the **native backend** when the payload was extracted in a match arm
(`Wrap(i)`) and then used (nested `match i` / passing `i` to a fn). Native emitted
empty/garbage; the C backend `bin/axc.exe` was always correct. BUG#51 family.

**Root cause (found via objdump of `axiom_temp.obj`):**
```
mov 0x8(%rsi),%r11    ; i = [outer+8] = inner box pointer   (correct)
mov %r11,-0x28(%rbp)  ; spill i
lea -0x28(%rbp),%r10  ; base = &slot                        (WRONG — should MOV slot)
mov (%r10),%rbx       ; tag = [&slot] = box pointer, not [box+0]
```
`x86_selector.regalloc_is_16byte_cached` default case (covers an OP_GET_FIELD-defined
reg) returned true for ANY size-16 type. A sum/option/result box is size 16
([tag,payload]) but its **register holds the 8-byte BOX POINTER** (pointer-repr).
Misclassified 16-byte ⇒ once spilled and reused as a GET_FIELD base, the spill
reload (x86_regalloc L908 `is_16` branch) emitted LEA(&slot) — the "value lives
inline at the slot" path — instead of MOV(slot→reg).

**Fix:** exclude ONLY the tagged pointer-repr kinds (sum 6 / option 11 / result 12)
from the size-16 classification in that default case — mirroring the guard already
in the OP_COPY/OP_CALL/param cases (cf. BUG#30 comment). **CRITICAL:** a broader
`not type_is_aggregate` guard (also excluding struct/array/tuple) **BROKE the
self-host build** — genA segfaulted compiling itself; the **fixpoint gate caught
it**. Narrowing to kinds 6/11/12 restored bit-identical fixpoint. Lesson: size-16
struct/array/tuple keep their prior 16-byte classification; don't touch them.

**Verified:** fixpoint bit-identical (genA==genB, both build cleanly); build-mode
compile outcomes AND runtime output on every `.expected` test identical to pre-fix
HEAD control (zero regressions). Test: tests/generics/nested_sum_payload.ax
(+.expected, prints nested-B/inner-B; pre-fix printed nothing).

Found while fixing [[bug54-qualified-variant]] (BUG#54/55) — probing enum/sum
patterns. See [[bug51-hunt-progress]] for the native-vs-cgen representation family.
