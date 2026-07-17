---
name: bug58-sum-struct-payload
description: "BUG#58 FIXED (6e27fb1) — sum variant with a STRUCT payload read the wrong field in a match arm (p.y read x). Pass-ordering bug in run_type_checker: sum pre-infer ran before struct type-ids registered → payload UNKNOWN."
metadata: 
  node_type: memory
  type: project
  originSessionId: 05d3f904-e67c-4f1a-9bfa-33caeb26ab45
---

**BUG#58 FIXED (commit 6e27fb1, pushed main 2026-07-04).** A sum variant carrying a
STRUCT payload — `type Shape = Dot(Point) | Nothing` — bound in a match arm
(`Dot(p)`): reading a NON-first field (`p.y`) resolved the wrong field index. `p.y`
read `x` (`p.x + p.y` = 20 not 30; `p.y` alone = 10; the printing form even
SEGFAULTED on the pre-fix native compiler). Passing the binding to a fn typed
`Point` masked it (the parameter re-typed it) — only a direct `p.field` in the arm
was wrong.

**Root cause:** pass ordering in `typecheck.ax run_type_checker`. Phase 0
(`pre_infer_type_alias` for sums — records each variant's payload type via
`infer_node(payload-type-expr)`) ran BEFORE the phase that pre-registers struct
type-ids. A struct-typed variant payload was inferred while the struct had no
type-id ⇒ payload_type = UNKNOWN. The match binding (typecheck L1355 sets
`sym.type_id = payload_type`) was then UNKNOWN, so `p.field` (NODE_FIELD_EXPR)
couldn't find the field and `extra_idx` defaulted to 0.

**Fix:** register struct type-ids (empty, forward-ref) BEFORE the sum-type
pre-inference pass. Struct FIELDS are still filled in the later struct pass; the
sum variant payload only needs the struct's type-id to exist early. Reorder only —
Phase 0.0 (struct ids) → Phase 0.1 (sums) → Phase 1 (struct fields) → sigs → infer.

**Diagnosis method:** `dump-air` showed `%11 = getfld %10` with field index 0 for
`p.y`; isolation showed the payload VALUE was correct (passing `p` to a Point-typed
fn worked) — only the arm's field-index resolution was wrong ⇒ a typecheck (not
codegen) bug. Same gate as always (fixpoint + zero-regression control diff).
Test: tests/generics/sum_struct_payload.ax (10/20/30).

One of a batch of enum/sum bugs found 2026-07-04 — see [[bug54-qualified-variant]],
[[bug56-nested-sum-payload]]. Still OPEN: [[bug57-match-option-native]] (match on
builtin Option/Result emits no code).
