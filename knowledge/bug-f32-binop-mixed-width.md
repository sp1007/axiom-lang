---
name: bug-f32-binop-mixed-width
description: "✅ FIXED (frontend, A==B): a binary float op with a mixed-width operand miscompiled to 0. A float literal defaults to f64, so `y * 2.0` with `y: f32` built f32*f64 operands, the op was typed F64, and codegen emitted a 64-bit mulsd over y's 32-bit bits -> garbage (0). Fix = float-literal bidirectional width inference in typecheck (mirror the int-literal rule). Found by probing while shipping the HOLE#6 f32 arg-coercion."
metadata:
  node_type: memory
  type: project
  originSessionId: 83ebf198-e937-49ec-a738-064db47952bb
---

# ✅ FIXED — mixed-width float binary op miscompiled to 0

Found 2026-07-16 by probing during the HOLE#6 float-arg-coercion work
([[bug-freefn-stdlib-collision-noarg]]).

## Symptom
Any binary float op where one operand is a float LITERAL (which defaults to f64) and
the other is a concrete f32 miscompiled to **0**:
```
let y: f32 = 1.5
let z: f32 = y * 2.0     # 0, want 3
let w: f32 = y + 2.0     # 0
let q: f32 = 1.5 * 2.0   # 0 (both literals, f32-expected)  — want 3
```
Controls that WORKED: `f32 * f32-var` (3), `f64 * f64-lit` (3), explicit `y * (b as f32)`
(3). So only the IMPLICIT f32/f64-literal mix was wrong.

## Root cause
`infer_node` binary-arithmetic branch (typecheck.ax ~L3108). The bidirectional literal
inference there only handled `NODE_INT_LIT` — a float literal (NODE_FLOAT_LIT=37) was
left at its default f64. With `t1=f32, t2=f64`, `result_type` became F64 (`if t1==F64 or
t2==F64`), but `y` is a 32-bit value, so the selector emitted a 64-bit mulsd reading y's
32-bit bits as a double → garbage.

## Fix (`ec8a0d0`, frontend-only, A==B `28B81830`, 337/337)
Added the float-literal analog right after the int-literal block: coerce a float-literal
operand to the OTHER operand's concrete float width (`t2 = infer_node(rhs, t1)` etc.), or
— when BOTH operands are float literals — to an f32 EXPECTED context. Same-width and
pure-f64 code is unaffected (the default f64 already matched). Oracle `bin/t_f32binop.ax`
(f32*f64lit + f64lit*f32 + both-lit-f32 = 42, O0==O1).

## Residual (not chased, rarer)
General mixed CONCRETE-width vars (`let a:f32; let b:f64; a*b` with no literal, no cast)
would still type F64 over mixed operands. Needs a real cvt on one operand at the op site
(codegen), not just literal retyping. Uncommon (users usually match widths or cast);
explicit `a * (b as f32)` already works. Same family as [[bug-f32-boundary-abi-open]] and
the HOLE#6 arg-coercion — the general "insert cvt at a mixed-width float boundary" theme.
