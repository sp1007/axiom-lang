---
name: bug-recursive-nongeneric-multifield-sum
description: "OPEN bug (probe-found 2026-07-24e): a NON-generic self-recursive multi-field sum variant `type Tree = Leaf(i64) | Node(Tree, Tree)` miscompiles — a match-bound payload (`Node(l,r)` → l) is typed as the synth STRUCT, not Tree, so sum(l) reads garbage (returns 0 not 7) and `match l` is rejected 'matching on struct not supported'. Distinct from the FIXED generic bug92 (Tree[T]/generic_inst path)."
metadata:
  node_type: memory
  type: project
---

## OPEN — self-recursive non-generic multi-field sum variant mistypes match-bound payload
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

## Distinct from the FIXED bug92
[[bug92-generic-recursive-multifield-open]] (`Tree[T]=Node(T,Tree[T],Tree[T])`, `938c48b`) was the
GENERIC recursive case, fixed via `field_is_pointer_sum` recognising a **generic_inst-of-SUM** (kind
8). This bug is the NON-generic case: `Node(Tree,Tree)` where `Tree` is a plain SUM (not generic_inst),
so it goes through a different resolution path that doesn't resolve the forward/self reference.

## Fix direction (for a dedicated gated session — NOT yet attempted)
The multi-field-variant desugar (RFC 0019, synth `__mfv_` struct) resolves each field's type when
`Node` is processed, but `Tree` is not yet fully registered at that point (self-reference during its
own definition) → the field resolves to the synth struct / an unresolved id rather than the SUM type
`Tree`. Then the match binder inherits that wrong type. Likely fix: make the payload field-type
resolution recognise a reference to the enclosing sum being defined (like `field_is_pointer_sum` does
for generic_inst, but for a plain forward/self SUM reference), OR defer field-type resolution until
after the whole sum is registered, then re-point `Node`'s field types to `Tree`. Verify the match
binder (`lower_match` / the typecheck of match-arm payload bindings) then types `l`/`r` as `Tree`.
Gate: A==B (frontend/typecheck; inert on self-host — the compiler uses no self-recursive multi-field
sum) + regression + a new oracle `t_rectreesum` (=7) + the `match l` accept. See probe files
`/tmp/probe2/s2.ax`,`s3.ax` (banked into bin/ when fixing).
