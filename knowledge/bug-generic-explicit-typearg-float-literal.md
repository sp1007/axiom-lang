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
| `t1` | `pick[u8](300, 1)` → **44** = 300 & 255 | ⭐⭐ **INT literals ARE coerced on the SAME explicit-type-arg path** |
| `s1`,`s2` | `pick[i32](250,1)`, `pick[u8](250,1)` | — |
| `s3` | generic struct ctor `W[f32](v: 2.5)` | generic CTOR path is fine |
| `s4` | generic METHOD `w.setv(2.5)` with T from the receiver | generic METHOD path is fine |
| `s5` | `Vec[f32].push(2.5)` | builtin generic container path is fine |
| `s6` | `Some(2.5)` into `Option[f32]` | variant ctor path is fine |

⇒ **The defect is exactly one branch**: generic FREE function × EXPLICIT type argument × FLOAT
literal argument. Everything adjacent — inferred type args, generic methods, generic struct ctors,
variant ctors, Vec/Option builtins, and integer literals on the very same explicit path — is
correct.

## Why this is a known SHAPE, and where to look
This is the third instance of one recurring defect shape in this codebase: **a coercion rule present
on one branch and missing on its sibling.**

- [[bug-f32-compare-float-literal]] — the arithmetic branch coerced float literals to the operand
  width; the COMPARISON branch coerced only INT literals. Identical omission.
- [[bug-negative-literal-compare-o0]] — int-literal typing correct in one position, wrong in the
  sibling.

Here, the explicit-type-arg generic call coerces INT literals (t1 proves it) but not FLOAT literals.
So the fix is very likely **one missing float clause beside an existing int clause** at the
generic-call argument coercion site, mirroring what `f614c11`-era fixes did for the comparison
branch. Find where the explicit-type-arg path coerces int literals to the instantiated param type
and add the float case; the inferred path (which works, see t2) is the reference implementation.

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
