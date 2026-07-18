---
name: bug-sum-of-primitives-match-segfault
description: "✅ FIXED (2026-07-18, A==B fixpoint EC01FECA, 398/398): `type ID = i32 | string` (a union whose variants are BARE PRIMITIVE types) now REJECTED at DECLARATION — was accept-then-SEGFAULT (139) on the payload-extracting match `i32(v)`. Reject via is_primitive_type_name() in pre_infer_type_alias. Oracle t_sumprimmatch (reject)."
metadata:
  node_type: memory
  type: project
---

# ✅ FIXED — Sum-of-primitives union rejected at declaration

**FIXED 2026-07-18** (frontend-only, A==B fixpoint `EC01FECA`, regression 398/398).
Attempt-2 succeeded via the **declaration-level** reject the refined direction (below)
predicted. In `pre_infer_type_alias` (typecheck.ax ~2238), inside the per-variant loop
(gated `not sum_is_builtin`, so Option/Result untouched), a new helper
`is_primitive_type_name(vnm)` detects a "variant" whose NAME is a bare builtin/primitive
type (`i32`/`string`/`bool`/…) and emits `error: sum type variant '%s' is a bare primitive
type; a union of primitive types is not supported — use a named constructor …`, bumping
`diags_count` (driver HALTs before codegen). This cleanly distinguishes the bug from valid
payload-less constructors (`Nil`/`Empty`) and from Option/Result entirely, sidestepping the
match-site `payload_type == TYPE_UNKNOWN` imprecision that sank attempt-1. Self-build-safe:
grep confirmed ZERO sum/enum pipe declarations in bootstrap/stage1 + std. Oracle
`t_sumprimmatch` (reject mode). Full union-of-primitives SUPPORT (tag + inline value, no box)
remains a larger RFC-gated feature if ever wanted.

# (historical) Sum-of-primitives union + variant-pattern match → accept-then-SEGFAULT

**Found 2026-07-18** by the M4 compliance-suite measurement (group 6, tests 057/058 —
"union ngầm định"). See [[m4-compliance-suite-spec-vs-impl-gap]].

**Minimal repro (segfaults 139 at O0 AND O1, deterministic):**
```
type ID = i32 | string
fn main() -> i32:
    let my_id: ID = 100
    match my_id:
        i32(v):
            return v
        string(s):
            return 0
```
**Isolation:** `type ID = i32 | string` decl alone → builds+runs fine (exit 1).
`... let my_id: ID = 100` (assign, no match) → builds+runs fine. **Only the MATCH with
type-named variant patterns `i32(v)`/`string(s)` crashes.** So the defect is in
matching/payload-extracting a sum whose "variants" are bare primitive TYPES: the pattern
`i32(v)` treats the primitive variant as if it has an extractable payload box; there is no
box → deref of garbage → SIGSEGV. Same accept-then-crash family as m4 (match-non-sum,
FIXED 7c4f058) and the malformed-input cluster [[bug-malformed-input-robustness-cluster]],
but here the scrutinee IS a sum type — just with primitive (payload-less) variants.

**Open question — is `type T = prim | prim` intended AXIOM?** Named-variant sums
(`type T = A(i64) | B(str)`) are the norm and work. Bare-primitive unions are an aspirational
spec feature (compliance suite uses them); no compiler-source or oracle uses them (verified:
zero `type .. = prim | prim` in bin/*.ax). Two fix directions:
- **(a) REJECT (bounded, safe first cut):** at typecheck, reject a variant pattern `T(x)` that
  binds a payload when the matched variant is a bare primitive type with no payload — OR reject
  the `type = prim | prim` declaration itself. Mirrors m4's NODE_MATCH_ARM reject
  (typecheck.ax ~L2773). Gate: A==B + full regression (compiler source has none → self-build
  safe). Oracle `t_sumprimmatch` (reject mode).
- **(b) SUPPORT (larger):** properly represent primitive-variant unions (tag + inline value,
  no box) and lower the match to read the value directly. Only if bare-primitive unions are a
  committed language feature (needs a spec/user decision).

**Priority:** medium — it's a real silent SIGSEGV, but on an aspirational/rare construct. The
(a) reject is the bounded, self-host-safe fix; gate is fast now
([[infra-defender-build-throttle]]). Schedule as an M4-core bounded task.

## ❌ Attempt 1 (2026-07-18, REVERTED) — match-level "payload-less variant" reject is UNSOUND
Added to the NODE_MATCH_ARM SUM path (typecheck.ax ~2962): reject when a variant pattern
`V(x)` binds a payload (`first_child` is NODE_BINDING_PAT) but the found variant's
`payload_type == TYPE_UNKNOWN`. It correctly rejected the bug AND passed **A==B fixpoint**
(3F7E05A8) — but the **full regression caught an OVER-REJECTION**: `t_treeoptchild`
(`match root.left { Some(p): ... }` where `root.left: Option[ptr[Node]]`) was wrongly
rejected. ROOT: builtin **Option/Result — and generic sums — reach this path with a variant
`Some`/`Ok` whose `payload_type` resolves to `TYPE_UNKNOWN` at the match check** (the generic
payload isn't concretized there), indistinguishable from a genuinely payload-less variant.
Gating to `s_entry.kind == TYPE_KIND_SUM` (excluding GENERIC_INST) did **not** help — Option
is a builtin generic sum that still presents as SUM here with an unresolved `Some` payload.
**LESSON (same imprecision wall as m2):** `payload_type == TYPE_UNKNOWN` at the match site is
NOT a reliable "genuinely payload-less" signal. Reverted cleanly (driver back to 1C2E3D6A);
A==B ≠ correctness — the full regression (t_treeoptchild) is what caught it.
**Refined fix direction:** do the reject at the **DECLARATION** of `type T = <primitive-type> |
...` — detect that a "variant" node is a BARE TYPE expression (primitive/type name) rather than
a named constructor (with/without payload). That distinguishes the bug (`i32`/`string` as
variants) from valid payload-less constructors (`Nil`, `Empty`) and from generic Option/Result
entirely. Requires reading the parser's sum-variant node structure (does a bare-type variant
differ structurally from a named `Nil`?). NOT the match site. Still low priority (rare construct).
