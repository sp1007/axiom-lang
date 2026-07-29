---
name: bug-f32-compare-float-literal
description: "FIXED 2026-07-24g (998F7199, 546/546, A==B): comparing an f32 value against a float LITERAL (`c == 4.0` with c:f32) was silently miscompiled — the literal defaulted to f64, producing a mixed-precision compare that returned the wrong result. The comparison branch of NODE_BINARY_EXPR lacked the float-literal width coercion the arithmetic branch already had. Probe-found."
metadata:
  node_type: memory
  type: project
---

## ✅ FIXED 2026-07-24g — driver `998F7199`, 546/546, A==B
**Probe-found** (float/string batch). `let c: f32 = 4.0; if c == 4.0: ...` took the FALSE branch —
`4.0f32 == 4.0` returned false. Isolated cleanly:
| case | result |
|---|---|
| f32 sum `== ` f32 VARIABLE | ✅ 42 (works) |
| f32 `==` float LITERAL (`c == 4.0`) | ❌ wrong (took else) |
| f32 sum `as i64` | ✅ 4 (the value is correct) |
| f64 `==` float literal | ✅ 42 (works) |
| f32 PARAM `==` literal | ❌ wrong |

So the f32 arithmetic and f32==f32 compare are fine; only **f32 compared to a float LITERAL** breaks,
in any operand order and for params or locals.

## Root & fix
`typecheck.ax` `infer_node` NODE_BINARY_EXPR: a float literal defaults to **f64**. The **arithmetic**
branch (op != 1) already coerced a float-literal operand to the other operand's concrete float width
(RFC 0006, ~L4124), but the **comparison** branch (op == 1, ~L4047) only did this for **INT** literals —
the float-literal coercion was missing. So `c(f32) == 4.0(f64)` stayed mixed-precision and codegen
compared the f32 bits against f64 bits → wrong. Fix: mirror the arithmetic branch's float-literal
coercion into the comparison branch (coerce the literal operand to the other operand's f32/f64 width via
`infer_node(lit, other_t)`), both operand orders.

Gate: fast fixpoint **A==B `998F7199`** (frontend, self-host inert — the compiler compares no f32 to a
float literal), regression **546/546** (+`t_f32cmplit`: let-bound f32, f32 param, both operand orders,
and a `>` comparison, O0==O1=42). Oracle `bin/t_f32cmplit.ax`.

## ✅ EXTENDED (same session) — NEGATED float literals — driver `C29FF51D`, 546/546, A==B
Immediately probing the just-shipped fix (per the lesson below) found the same gap for a NEGATED
literal: `c == -2.5` (comparison) AND `y * -2.0` (arithmetic) still failed, because `-2.5` parses as
`NODE_UNARY_EXPR(neg, FLOAT_LIT)`, not a bare `NODE_FLOAT_LIT`, so the literal-gate missed it — in BOTH
branches. Fix: a helper `node_is_float_litish` that also matches a unary-neg wrapping a FLOAT_LIT
(re-inferring a unary-neg node with an expected type propagates it to the inner literal, since
NODE_UNARY_EXPR infers its operand with `expected`). Used in both binop branches. A==B `C29FF51D`,
regression 546/546, oracle `t_f32cmplit` extended with negated-literal cases (`0.0 - 2.5 == -2.5`,
`3.0 * -2.0 == -6.0`). (Negated INT literals in comparisons may have an analogous latent gap but were
NOT confirmed broken and int mixed-width behavior differs — left unprobed.)

## Lesson
A coercion rule added to ONE binop branch (arithmetic) but not its sibling (comparison) is a classic
partial fix — the same mixed-precision hazard exists at every operand site. When adding an
operand-coercion rule, apply it to ALL binop branches AND to every literal SHAPE (bare literal +
unary-negated literal). This bug took TWO passes: first the missing comparison branch, then the missing
negated-literal shape — each found by immediately probing the fix's neighborhood. Sibling of the
int-literal width coercion that exists in both branches.
