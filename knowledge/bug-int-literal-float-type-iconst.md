---
name: bug-int-literal-float-type-iconst
description: "FIXED 2026-07-30 — an INTEGER literal that typecheck adopted into a float type (let a: f64 = 3, RFC 0005) was still emitted as OP_ICONST carrying type_id 9/10, so the constant landed in a GPR-classed vreg while the float ops read the aliasing XMM: exactly 0.0, or a stale float (so -O0 and -O1 disagreed). Root cause: lower_int_lit was a drifted copy of lower_float_lit. Fix = one shared emit site + a new AIR verifier rule."
metadata:
  node_type: memory
  type: project
---

# `let a: f64 = 3` evaluated to 0.0 — int literal adopted into a float type was emitted as OP_ICONST

Status: **✅ FIXED 2026-07-30**, B==C `824807E2F2F17F346C1F4038D01CC89BFB0BD0163CDC8006F491AC4F4B8F8A48`,
regression **597/597** at default AND at `-O0`. Oracle `bin/t_intlitfloatctx.ax` (before: exit **1**;
after: **42** at -O0/-O1/-O2).

## The defect

RFC 0005 lets typecheck **adopt** an integer literal into the expected float type, so
`let a: f64 = 3` types the `3` as f64 and *nothing downstream believes a conversion is owed*. But
`air_builder.lower_int_lit` read that adopted `type_id` (9=f32, 10=f64) and still emitted
**`OP_ICONST`** with it. Measured AIR (`dump-air bin/probe4/g12.ax`):

    %1: t10 = iconst %3      ; let a: f64 = 3   <- INTEGER constant carrying f64
    %2      = cast %1        ;   f64->f64, a no-op
    %6: t10 = iconst %3      ; c + 3
    %7: t10 = fadd %5, %6    ;   FADD reading an integer materialized in a GPR

The value is materialized with `mov imm` into a GPR-classed vreg whose hw index aliases the XMM the
float op reads (the `R10 ≡ XMM10` family) ⇒ **exactly 0.0**, or a **stale** float left in that XMM —
which is why the symptom moved between opt levels (`bin/probe4/h2.ax` passed -O0 and failed -O1).
Accept-then-miscompile, BUG#53 class, and broad: annotated `let`, assignment to a `mut` float,
array element + element assign, `return` from a `-> f64` fn, `3 as f64` (the *documented*
"workaround" — equally broken, since the cast is a no-op over the same constant), `c + 3`,
`3 / 2.0`, `c > 2`, global initializer.

## Root cause: two copies of one mechanism, one never extended

`lower_float_lit` did the right thing — IEEE-754 bit pattern split across src1(low32)/src2(high32),
f32 narrowed first — and even documented RFC 0006 on the line. Its sibling `lower_int_lit` was a
copy that was **never extended** when float constants got real support. Same defect class as the
interface-return miscompile and the f32-only hint below: **a duplicated mechanism drifts, and the
copy that is exercised less keeps the old behaviour.**

## Fix (`bootstrap/stage1/air_builder.ax`)

1. New **single emit site** `emit_float_const(reg, type_id, value: f64)`; `lower_float_lit` and
   `lower_int_lit` both call it. `lower_int_lit` routes to it when the adopted `type_id` is 9/10
   (`val as f64`; explicit int→float casts were verified working first, incl. exact bit patterns).
2. **§9 verifier rule `verify_air_const_types`**, run on every function as it leaves the AIR builder
   (`lower_func`): **OP_ICONST must not carry a float type_id**, and OP_FCONST must not carry a
   definite integer one (type_id 0 = unset/void is not flagged). Fatal (`ax_panic`) — the
   alternative is a plausible-looking binary that computes the wrong number.
   **Calibrated**, not assumed: disabling the fix makes the build print
   `internal error: OP_ICONST carries a float type_id ... main / inst #0 OP_ICONST dest=1 type_id=10`
   and exit 1 with **no** output file. This invariant was true-but-unchecked for the whole life of
   the bug; *a fact in a comment protects only its own line, a check protects every line.*

Self-host inert (the compiler is integer-only): **A == B == C**, all three the same hash.

## ⚠️ STILL OPEN — a DIFFERENT root cause at the same-looking positions

Four positions still yield the wrong value, and they are **not** literal-materialization:
typecheck never hints the float type there, so the literal's node type stays INTEGER and there is no
float constant to materialize (measured: `%2: t3 = iconst %3` feeding an f64 `setfld`).

| position | probe | f32 twin |
|---|---|---|
| `H(a: 3)` into an **f64 struct field** | `bin/probe4/i1.ax` | `i2.ax` **passes** |
| `argf(3)` into an **f64 fn param** | `bin/probe4/h4.ax` | `g6.ax` bit 2 passes |
| `mm.m(3)` into an **f64 method param** | `bin/probe4/g6.ax` bit 1 | — |
| `let c: f64 = 3 + 1` (int **expression**) | `bin/probe4/g5.ax` bit 64 | — |

Every **f32** twin passes, because `typecheck.ax:5208` (and its relatives) hint `TYPE_F32` only —
comment: *"Scoped to F32 only, so int/i64 scalar fields keep their prior UNKNOWN inference and the
self-host fixpoint is unaffected."* Same drifted-copy class again. Fixing it = propagate the hint /
insert a coercion at those POSITIONS (a typecheck change with real fixpoint exposure), which is the
open follow-up (b) at [[bugs]] BUG#33. Row numbers **4, 8, 9, 11 are reserved** for them in
`bin/t_intlitfloatctx.ax` — do not renumber.

Also still open and deliberately untouched: int **VARIABLE** → float at let/assign/arg
(`bin/probe4/g7.ax`, exit 100 = all three rows wrong), and the reverse direction
`let a: i64 = 3.0` (`bin/probe4/g13.ax`, exit 3 = the raw IEEE bits are copied into an integer).

## Oracle traps that bit this investigation (both are encoded in the oracle's header)

- **Never route the value under test through a parameter of the same float type** to check it — the
  arg site coerces it and hides the defect.
- **A row can pass in a multi-row program purely because an earlier float compare left the right
  value in an XMM register.** `return 3` from an `-> f64` fn did exactly that. Keep rows
  independent and re-verify any suspicious pass in a minimal file (`bin/probe4/g8.ax`).

Related: [[lesson-exit-code-8bit-masking]], [[bug-generic-explicit-typearg-float-literal]],
[[bug-float-arg-reg-unprotected]].
