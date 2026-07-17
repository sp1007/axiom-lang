---
name: bug54-qualified-variant
description: "BUG#54 FIXED — qualified nullary-variant construction `EnumType.Variant` (e.g. Color.Blue) segfaulted; only bare `Blue` worked. Fix in typecheck+air_builder, fixpoint OK."
metadata: 
  node_type: memory
  type: project
  originSessionId: 05d3f904-e67c-4f1a-9bfa-33caeb26ab45
---

**BUG#54 FIXED (commit 1bb4359, pushed main 2026-07-04).** `Color.Blue` — a
qualified enum-variant name in value position — **segfaulted at runtime**; only
the bare `Blue` form worked. Discovered while probing type-alias support for
RFC 0011 P4 inc4c (see [[rfc0011-p4-separate-compilation]]).

**Root cause:** typecheck NODE_FIELD_EXPR resolved the base `Color` to the sum
type (TYPE_KIND_SUM) but, finding no *struct field* named `Blue`, fell through to
method resolution and left the node untyped. The lowerer then took the generic
`lower_field_expr` path → `lower_expr(Color)` on a bare type name → garbage
receiver reg → OP_GET_FIELD → segfault.

**Fix (scoped to no-payload variants — the confirmed crash):**
- `typecheck.ax` NODE_FIELD_EXPR: when base is a user sum + field name matches a
  nullary variant (`vinfo.payload_type == TYPE_UNKNOWN`), search symtable for the
  SYM_VARIANT with matching name_id AND type_id, bind node.payload=variant_sym +
  flag 2048, result_type=sum type. Mirrors bare-variant resolution.
- `air_builder.ax` lower_field_expr flag-2048 path: added `if sym.kind ==
  SYM_VARIANT` (sum-typed) → `lower_variant_construct(sym_idx, 0)`.
- Payload constructors `Enum.Ctor(x)` are a call → untouched, never fires.

**Gate passed (backend/lowering change ⇒ fixpoint required, see
[[feedback-fixpoint-async-rule]]):** self-host fixpoint bit-identical (genA==genB);
compile exit-codes across the whole test tree IDENTICAL to a no-fix git-stash
control (zero new compile regressions); pre-fix `axc_control` segfaults on the new
test, fixed prints blue/red/green. Test: tests/generics/qualified_variant.ax
(+.expected). Daily-driver binaries (axc_native + stage2/3_native) promoted to the
fixed compiler.

**BUG#55 FIXED (commit a84e936, sibling of #54).** `Shape.Circle(4)` — qualified
constructor for a PAYLOAD variant — segfaulted too; only bare `Circle(4)` worked.
Bare path keyed on `callee_node.kind == NODE_IDENT`; a qualified callee is a
NODE_FIELD_EXPR → fell to generic field-load on the type name → crash. Fix: (1)
typecheck qualified-variant recognition now matches payload variants too (dropped
nullary guard); (2) lower_field_expr variant branch stays nullary-only (guarded via
find_variant_info — payload variant only completes when called); (3) lower_call_expr
new branch: flag-2048 field-expr callee bound to a sum variant → lower_variant_construct
with call arg as payload. Test: tests/generics/qualified_variant_payload.ax. Same
gate (fixpoint bit-identical + zero-regression build-mode diff vs HEAD control).

**Method:** the harness `-Mode run -Stage 2` shows ~19 pre-existing FAILs even at
baseline (CRLF/env noise + a STALE bak binary that predated BUG#53's parse-error
halt). Do NOT trust run-mode counts as the gate; the real signals are (a) fixpoint
SHA, (b) build-mode exit-code diff vs a stashed control. Build the control by
`git stash push <files>` → regen_concat → build → `git stash pop`.
