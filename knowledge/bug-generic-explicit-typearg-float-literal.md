---
name: bug-generic-explicit-typearg-float-literal
description: "OPEN 2026-07-30 (probe-found) — a FLOAT LITERAL passed to a generic FREE function with an EXPLICIT type argument (`idf[f32](2.5)`) is never coerced to the instantiated width: it stays f64 against an f32 param and the callee reads garbage (returns 0). The INFERRED path coerces correctly, and INT literals coerce correctly on the same explicit path — so it is one missing float clause on a sibling branch."
metadata:
  node_type: memory
  type: project
---

# OPEN — float literal + explicit generic type arg is not coerced

Probe-found 2026-07-30, batch 2. **Accept-then-miscompile** (BUG#53 class): no diagnostic, silent
wrong value.

## Minimal repro

    fn idf[T](a: T) -> T:
        return a
    fn main() -> i64:
        let y = idf[f32](2.5)
        return (y * 100.0) as i64      // returns 0, expected 250

`bin/probe2/t3.ax`. Two-parameter form `bin/probe2/r1.ax` (`pick[f32](2.5, 1.0)`) behaves the same.

Mechanism: a float literal defaults to **f64**, the instantiated parameter is **f32**, and no
coercion is inserted on this path — so an 8-byte double bit pattern is passed where a 4-byte single
is expected and the callee reads garbage.

## The boundary, measured — this is the useful part

**FAILS** (both `bin/probe2/`):

| probe | shape |
|---|---|
| `t3` | `idf[f32](2.5)` — single param, explicit type arg |
| `t4` | `idf[f32](2.0 + 0.5)` — same, folded expression |
| `r1` | `pick[f32](2.5, 1.0)` — two params |
| `r3` | `let y: f32 = pick[f32](2.5, 1.0)` — an explicit BINDING annotation does NOT rescue it |

**WORKS** — each one narrows the diagnosis:

| probe | shape | why it matters |
|---|---|---|
| `t2` | `pick(2.5, q)`, T INFERRED f32 from an f32 var | ⭐ **the inferred path DOES coerce the float literal** |
| `r2` | `pick[f32](p, q)` with f32 VARIABLES | not literal-specific to the callee — it is the LITERAL that is uncoerced |
| `r4` | `pick(p, q)` inferred from f32 vars | — |
| `r5` | `pick[f64](2.5, 1.0)` | f64 needs no coercion (it is the literal's default) |
| `t1` | `pick[u8](300, 1)` → **44** = 300 & 255 | ⚠️ proves NOTHING — see the retraction below |
| `s1`,`s2` | `pick[i32](250,1)`, `pick[u8](250,1)` | — |
| `s3` | generic struct ctor `W[f32](v: 2.5)` | generic CTOR path is fine |
| `s4` | generic METHOD `w.setv(2.5)` with T from the receiver | generic METHOD path is fine |
| `s5` | `Vec[f32].push(2.5)` | builtin generic container path is fine |
| `s6` | `Some(2.5)` into `Option[f32]` | variant ctor path is fine |

⇒ **The defect is exactly one branch**: generic FREE function × EXPLICIT type argument × FLOAT
literal argument. Everything adjacent — inferred type args, generic methods, generic struct ctors,
variant ctors, Vec/Option builtins, and integer literals on the very same explicit path — is
correct.

## ⚠️ RETRACTION — `t1` does not prove int literals are coerced
I first read `pick[u8](300, 1) -> 44` as proof that the explicit-type-arg path coerces INT literals
and therefore that only a float clause was missing. **That inference is invalid.** 44 is `300 & 255`,
which is what you get EITHER from a real coercion OR from the callee simply reading the low 8 bits of
a register holding 300. For integers those two are indistinguishable by construction — reading the
low N bits IS truncation — so **no integer test can discriminate here.** My original "integers merely
mask the defect" hypothesis, which I dismissed on the strength of `t1`, is back on the table and is
now the more likely one.

This matters because it changes the fix: if nothing is coerced on the generic path, the fix is to
thread an expected type there at all, not to add one clause beside an existing one.

## Located, 2026-07-30 — the site and why the shape is what it is
`typecheck.ax:5257` builds `fp_data`, the parameter list used to thread an expected type down to each
argument. It is gated:

    if not did_ctor_args and not is_generic_call and ...nodes.data[callee].kind == NODE_IDENT and ...

and at `:5278` it already contains **exactly the needed clause, for non-generic calls only**:

    elif cand == TYPE_F32:
        // An f32 scalar param: pass f32 as expected so a float literal arg
        // (`takef(40.0)`) is typed f32 and materialized as single instead of a DOUBLE
        // -> read back wrong at the f32 param slot -> 0

⇒ **A generic call gets NO expected-type threading whatsoever**, so the float literal keeps its f64
default. The non-generic F32 clause is the reference implementation and its comment describes this
exact failure mode ("materialized as a DOUBLE → read back wrong → 0"). Note that clause cites
[[bug-f32-boundary-abi-open]], so the non-generic half of this was already found and fixed once; the
generic half was left.

The `not is_generic_call` gate is deliberate — its comment says it exists "so method/generic arg
positions are never misaligned", because for a generic callee the declared params are type
PARAMETERS, not concrete types, and `fp_data` would thread a generic placeholder. So the fix needs
the MONOMORPHISED param type, which is why it is not a one-word edit.

## Why this is a known SHAPE, and where to look
This is the third instance of one recurring defect shape in this codebase: **a coercion rule present
on one branch and missing on its sibling.**

- [[bug-f32-compare-float-literal]] — the arithmetic branch coerced float literals to the operand
  width; the COMPARISON branch coerced only INT literals. Identical omission.
- [[bug-negative-literal-compare-o0]] — int-literal typing correct in one position, wrong in the
  sibling.

Here the sibling branches are **generic vs non-generic calls**: the non-generic branch threads an
f32 expected type (`:5278`) and the generic branch threads nothing.

**Fix direction**: for a generic call with a resolvable instantiation, thread the MONOMORPHISED
parameter type as the argument's expected type — at minimum for the float cases, which is the scope
the non-generic clause already limits itself to. The INFERRED path works today (`t2`), so whatever it
does with the literal is the behavioural reference to match.

⚠️ **Do NOT hotfix blind.** `knowledge/backlog-open-items.md` records a previous attempt at a
generic-call-arg coercion post-pass that was **REVERTED** — it did not fix the target (tuple element
widths) and actively caused an O0/O1 divergence, because monomorphisation can consume the argument
node before a re-inference post-pass runs. That warning was about re-inferring at the wrong time, not
about adding a coercion clause where one already exists for ints, but generics are self-host
critical: verify by probe that the clause FIRES (per the peephole-1d lesson — prove it acts, do not
merely prove it is safe) and gate with `A==B` + full regression.

## Why the suite misses it
No test calls a generic free function with an explicit narrow-float type argument and a literal.
`t_genfnexpltypearg` covers the explicit-type-arg form but with non-float types.

Oracle to add when fixed: the full FAILS table plus the neighbouring WORKS rows, since a fix that
coerced too eagerly would break the inferred path or the ctor paths.
