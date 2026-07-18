---
name: bug-sum-of-primitives-match-segfault
description: "OPEN (found 2026-07-18 via M4 compliance group 6, test_057/058): `type ID = i32 | string` (a sum/union of BARE PRIMITIVE types) builds, but matching it with type-named variant patterns `i32(v)` / `string(s)` accept-then-SEGFAULTs (139) at BOTH O0 and O1. Decl-only and let-assign build & run fine; only the payload-extracting match crashes. BUG#53 class: must reject or work, not silent crash."
metadata:
  node_type: memory
  type: project
---

# Sum-of-primitives union + variant-pattern match → accept-then-SEGFAULT (OPEN)

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
