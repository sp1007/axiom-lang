---
name: bug-narrowing-cast-compare-width
description: "✅ FIXED `aede254`: an overflowing narrowing cast like `300 as u8` left DIRTY UPPER BITS (scalar OP_CAST emitted a bare MOV, skipping emit_wrap_to_width unlike arithmetic), so a 64-bit `==` read 300 not 44. Fix = call emit_wrap_to_width after the scalar cast MOV, restoring the documented invariant. B==C 07D2D13F, 341/341, oracle t_castwidth."
metadata:
  node_type: memory
  type: project
  originSessionId: 3228306b-52d7-4378-bb1c-a0b6cef57eba
---

# ✅ BUG FIXED `aede254`: narrowing-cast value not masked to width

**FIX:** [x86_selector.ax](../../../../d--projects-compiler-Axiom) scalar `OP_CAST` path now calls
`emit_wrap_to_width(sel, inst.dest, inst.type_id, out_insts)` after the cast MOV (line ~1160),
matching what every arithmetic op already did. Repro below now returns 3 (and oracle t_castwidth,
covering u8/u16/i8, returns 15). B==C 07D2D13F, regression 341/341, ELF 10/10. Root was a single
missing normalize call — the invariant in emit_wrap_to_width's comment ("every producer... load,
cast, arithmetic... normalizes") was stated but the cast producer didn't honor it.

## (historical) BUG: narrowing-cast value not masked to width at comparison

Found while designing M6 immediate-compare folding [[m6-perf-gate-fib-benchmark]].

## Repro
```axiom
fn main() -> i64:
    let a: u8 = 300 as u8      // 300 & 0xFF == 44
    let b: u8 = 44 as u8
    mut r := 0 as i64
    if a == b:            r = r + 1   // 44==44 SHOULD be true
    if b == (300 as u8):  r = r + 2   // 44==(300 as u8=44) SHOULD be true
    return r                          // CORRECT=3; actual=0
```
`axc_native` (D6AF9DC7) returns **0** — both compares FALSE.

## Diagnosis
- `let y: u8 = 300 as u8; return y as i64` → **44** (correct): the `as i64` widening masks/zero-extends.
- But `a == b` and `b == (300 as u8)` → false: the u8 value carries **dirty upper bits** (300 = 0x12C)
  and the integer comparison runs at 64-bit **without masking to the u8 width**, so 0x12C != 0x2C.
- So narrowing casts (`300 as u8`) do NOT actually reduce the stored/register value to the target
  width; only a subsequent widening cast cleans it. Comparisons (and likely other width-sensitive
  ops) see the un-masked value.

## Scope / priority
Narrow: only triggers for casts of an OUT-OF-RANGE constant/value to a smaller type used directly in
a compare (rare in real code — most u8/u32 values are already in-range with clean bits, which is why
340/340 regression passes). LOW priority. **But same width-handling class that broke immediate-fold**:
the fix (mask/zero-or-sign-extend a narrowing cast's result to its width, or make compares width-aware
via `emit_wrap_to_width`) would ALSO unblock safe `CMP reg,imm32` folding for sub-64-bit operands.

## Fix sketch (needs care, backend)
Either (a) emit a width-mask (`movzx`/`and`) after a narrowing OP_CAST so the value is clean in the
register, or (b) make `select_comparison` (and imm folding) width-aware — compare at the operand's
actual byte width (REX/operand-size prefix + `movzx` the reg first). Verify against the whole
u8/u16/u32 compare surface; fixpoint B==C mandatory (backend). Oracle: the repro above → 3.
