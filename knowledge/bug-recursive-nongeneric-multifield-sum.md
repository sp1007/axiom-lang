---
name: bug-recursive-nongeneric-multifield-sum
description: "OPEN bug (probe-found 2026-07-24e): a NON-generic self-recursive multi-field sum variant `type Tree = Leaf(i64) | Node(Tree, Tree)` miscompiles — a match-bound payload (`Node(l,r)` → l) is typed as the synth STRUCT, not Tree, so sum(l) reads garbage (returns 0 not 7) and `match l` is rejected 'matching on struct not supported'. Distinct from the FIXED generic bug92 (Tree[T]/generic_inst path)."
metadata:
  node_type: memory
  type: project
---

## ✅ FIXED 2026-07-24e (`03F808DE`, 531/531, B==C) — forward-declare gated on a bare self-ref
Shipped: `reserve_sum_type`/`set_sum_variants` (typetable.ax) + a pre-scan in `pre_infer_type_alias`
that reserves the sum's type_id and sets it on the symbol BEFORE resolving variant fields — but ONLY
when a variant field is a **bare self-reference** (`NODE_TYPE_EXPR` whose `payload == sym_idx`, the
sum's own symbol). Generic templates and mono clones self-reference as `Tree[..]` (`NODE_GENERIC_TYPE`),
so the kind check excludes them → t_gentree/t_gentreestr stay 15/26; non-recursive sums take the exact
original path. Result: s2=7, s3=3 (accept), q1=28, oracle `bin/t_rectreesum.ax` (=28) in the suite.
Gate: A==B==C `03F808DE`, full regression **531/531**. The critical insight (2nd attempt) was that the
first, un-gated version reserved for ALL non-generic sums and broke the **mono clone** `Tree[i64]`
(generic params stripped, so a NODE_GENERIC_PARAMS check missed it) — the bare-self-ref field-kind gate
is what cleanly separates genuine recursion from clones.

## (history) OPEN — self-recursive non-generic multi-field sum variant mistypes match-bound payload
**Found by proactive probing on the mature-plateau compiler (driver `AFA6529F`), 2026-07-24e.**
The classic linked-tree / AST shape miscompiles:
```
type Tree = Leaf(i64) | Node(Tree, Tree)     // Node's fields ARE the enclosing sum
fn sum(t: Tree) -> i64:
    match t:
        Leaf(v): return v
        Node(l, r): return sum(l) + sum(r)
fn main() -> i64:
    return sum(Node(Leaf(3), Leaf(4)))       // WANT 7, GET 0
```

## Precisely characterized (probe matrix)
| case | shape | result |
|---|---|---|
| s1 | `Leaf(5)` match | ✅ 5 |
| s6 | `Pt(i64,i64)` multi-field SCALAR (RFC 0019 baseline) | ✅ 34 |
| s4 | `P(Color,Color)` multi-field SUM-typed fields, **non-recursive** + nested match | ✅ 2 |
| s5 | `Cons(i64)` single scalar field | ✅ 7 |
| **s2** | `Node(Tree,Tree)` **self-recursive** multi-field, `sum(l)+sum(r)` | ❌ **0** (want 7) |
| **s3** | same, then `match l` (l bound from `Node(l,r)`) | ❌ **REJECT** "matching on a struct/array/tuple/... value is not supported" |

⇒ The distinguishing factor is **SELF-RECURSION**: the variant's field type is the ENCLOSING sum
(`Tree`) that is still being defined when `Node` is desugared. Non-recursive sum-typed multi-field
payloads (s4) work, so the multi-field + sum-payload machinery is fine; only the self-reference
breaks. s3's reject message is the smoking gun: the payload binder `l` is typed as the synth STRUCT
(the RFC 0019 `__mfv_` struct for `Node`), NOT as `Tree` the sum → `match l` sees a struct → rejects;
and in s2 `sum(l)` passes a mistyped/garbage value → 0.

## ⚠️ ATTEMPTED 2026-07-24e — fix WORKS for the target but broke a mono clone; REVERTED to green
Implemented the forward-declaration exactly as below (reserve_sum_type + set_sum_variants in
typetable.ax; reserve+set-sym-type-id before the variant loop in pre_infer_type_alias). Result:
**s2=7, s3=3 (accept), q1=28 — the bug IS fixed**, non-recursive sums (s4/s5/s6) unchanged, and
**B==C `095CFB1E`** held. BUT full regression = **529/530, `t_gentree` SEGFAULTS** (the GENERIC
recursive `Tree[i64]`). Reverted (driver stays `AFA6529F`, A==B green).
- First refinement TRIED: skip forward-decl when the alias decl has a `NODE_GENERIC_PARAMS` child
  (generic template). **Did NOT fix t_gentree** — the breakage is the **MONO CLONE** `Tree[i64]`,
  not the template. Mono strips generic params from clones (`remove_generic_params_child`,
  mono.ax:533) and re-runs `pre_infer_type_alias(cloned_root)` (:3189), so the clone has NO
  `NODE_GENERIC_PARAMS` → my check treated it as non-generic → reserved a concrete sum id + set the
  clone's sym.type_id early → the clone's self-ref field `Tree[i64]` (a NODE_GENERIC_TYPE) got
  mis-collapsed → segfault. (Curiously `t_gentreestr` = `Tree[str]` still PASSED — the >8B-payload
  path differs; only the i64 clone broke.)
- **PRECISE REFINED FIX for next time:** apply forward-declaration ONLY when a variant field is a
  **BARE self-reference** — a field type node that is NOT `NODE_GENERIC_TYPE` and names the enclosing
  sum. Genuine non-generic recursive sums use a bare `Tree` field; BOTH generic templates AND mono
  clones use `Tree[T]`/`Tree[i64]` = `NODE_GENERIC_TYPE` fields, so keying on "has a bare-name self-ref
  field" cleanly includes s2 and excludes t_gentree. Pre-scan the variant fields (the `ptn` sibling
  chain at typecheck.ax:2824-2833, BEFORE resolving) for a node with `kind != NODE_GENERIC_TYPE` whose
  identifier == `sym.name_id` (compare via the field node's resolved payload symbol or its token text
  against `self.intern.get(sym.name_id)`); reserve only if found. Gate: B==C + FULL 530 regression
  (t_gentree/t_gentreestr MUST stay 15/26) + new `t_rectreesum`.

## Distinct from the FIXED bug92
[[bug92-generic-recursive-multifield-open]] (`Tree[T]=Node(T,Tree[T],Tree[T])`, `938c48b`) was the
GENERIC recursive case, fixed via `field_is_pointer_sum` recognising a **generic_inst-of-SUM** (kind
8). This bug is the NON-generic case: `Node(Tree,Tree)` where `Tree` is a plain SUM (not generic_inst),
so it goes through a different resolution path that doesn't resolve the forward/self reference.

## Fix direction — EXACT location + recipe (investigated 2026-07-24e; NOT yet implemented)
ROOT confirmed by reading `pre_infer_type_alias` (typecheck.ax):
- **:2827** `let ft = self.infer_node(ptn, TYPE_UNKNOWN)` resolves each `__mfv_` field type. For
  `Node(Tree,Tree)`, `ptn` is `Tree`. `infer_node(Tree)` resolves the NAME `Tree` via its symbol's
  `type_id` — but that is set only at **:2860** (`symbols[sym_idx].type_id = type_id`), AFTER
  `register_sum_type` at **:2859**, i.e. AFTER this field loop. So the self-ref `Tree` resolves to
  TYPE_UNKNOWN → synth field `_f0/_f1` gets UNKNOWN → the `__mfv_Tree_Node` struct has garbage field
  types → match binder `l` (from `Node(l,r)`) inherits UNKNOWN/struct → s2 wrong, s3 reject.

**FIX = forward-declare the sum's type_id BEFORE the variant/field loop (classic self-ref pattern):**
1. Split `register_sum_type` (typetable.ax:299) into **reserve** (push an empty `SumInfo` + a
   `TYPE_KIND_SUM` entry, return the id) and **populate** (set `sumtypes[extra].variants = variants`).
   Or add `reserve_sum_type(name_id)->u32` + `set_sum_variants(type_id, variants)`.
2. In `pre_infer_type_alias`, BEFORE building `variants` / the variant loop: `let type_id =
   self.types.reserve_sum_type(sym.name_id)` and `self.symtable.symbols.data[sym_idx].type_id =
   type_id` immediately. Now `infer_node(Tree)` at :2827 resolves the self-ref to `type_id`.
3. After the loop, replace `register_sum_type(...)` at :2859 with `self.types.set_sum_variants(type_id,
   variants)`; keep :2860/:2861 (already sets sym + node type, now redundant for sym but harmless).
⚠️ **BLAST RADIUS: every sum type flows through this path** (530 tests use sums). The reserve/populate
split must be behaviour-identical for NON-recursive sums (the common case). Gate HARD: A==B (frontend
typecheck; inert on self-host — the compiler declares no self-recursive multi-field sum, so freeable
here is nil and the reserved-id path only changes self-ref resolution) + FULL regression 530/530 +
new oracles `t_rectreesum`(=7, s2) and the `match l` accept (s3). Watch for: dedup/ordering
differences in type_id assignment shifting hashes (should stay A==B since sum ids are positional and
the reserve just moves the push earlier by a few lines — verify the entry ORDER vs structs registered
inside the loop, e.g. the `__mfv_` synth struct at :2841 is registered DURING the loop, so reserving
the sum FIRST puts the sum's id BEFORE its `__mfv_` structs' ids — a numbering change that is
self-consistent but verify it doesn't perturb any id-order-dependent code). Repro `/tmp/probe2/s2.ax`,
`s3.ax` → bank as `bin/t_rectreesum.ax` when fixing.
