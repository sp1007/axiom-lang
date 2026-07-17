---
name: bug-lambda-param-field-capture
description: "FIXED b37c14a: lambda body accessing a field/method of its own param (p.y, s.len, n.to_str()) was wrongly rejected as 'closure captures'' — FIELD_EXPR field-name child is an empty-payload placeholder ident."
metadata: 
  node_type: memory
  type: project
  originSessionId: 1e902df0-dc42-4a78-ba53-850e359b6f93
---

**FIXED `b37c14a`** (A==B `BD81C572…`, 278/278). Found via HOF×struct probe batch (p1/p3).

**Symptom:** any `param.field` / `param.method()` inside a lambda body →
`error: closure captures '' from an enclosing scope` (empty name). Blocked a huge
common class: no lambda passed to `Vec.map/filter/fold/count/...` (or any HOF)
could read a field or call a method of its element.

**Root:** `scan_closure_free_vars` (parser.ax) recursed into EVERY NODE_IDENT and
flagged any name not in params/globals. The field-name child of a NODE_FIELD_EXPR
is a **structural placeholder**: the field name lives in the FIELD_EXPR's own
payload ([parser.ax:461](../../../../../d--projects-compiler-Axiom/bootstrap/stage1/parser.ax)),
and the child ident ([parser.ax:463](../../../../../d--projects-compiler-Axiom/bootstrap/stage1/parser.ax))
is created with payload 0 (never set). `pool.get(0)` = `""` → flagged as capturing `''`.

**Fix:** capture scan ignores payload==0 idents (`if node.kind == NODE_IDENT and node.payload != 0`).
Payload-0 idents are never variable references — only placeholders (also covers
tuple `._f0._f1` chain idents). ACCEPT-only → A==B held.

**Lesson:** same false-positive family as `7e5965a` (variant ctors). The zero-capture
scan (BUG#73) over-rejects whenever the AST has synthetic/placeholder idents; when
adding lambda features, probe field/method access on the param. Related [[bug73-closure-capture-reject]].
Oracles t_lambdafield(24)/t_lambdamethod(42).
