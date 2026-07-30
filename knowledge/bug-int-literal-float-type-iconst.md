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

## ✅ SECOND ROOT CAUSE, FIXED 2026-07-31 — the coercion existed only for f32, at two sites

Everything below was the *other half* of the same symptom with a different cause. Typecheck never
hinted the float type at these positions, so the literal's node type stayed INTEGER and there was
no float constant to materialize (measured: `%2: t3 = iconst %3` feeding an f64 `setfld`); and for
an int **VARIABLE** a hint could never have helped anyway, because an ident's storage is already
built at its own width. Measured before/after (bit set = row CORRECT):

| position | probe | before | after |
|---|---|---|---|
| `H(a: 3)` into an **f64 struct field** | `bin/probe4/i1.ax` | 0 | **42** |
| `argf(3)` / `takes_f64(9)` **f64 fn param** | `bin/probe4/h4.ax`, `bin/probe6/q1.ax` | 0 / 1 | **42** |
| `mm.m(3)` **f64 method param** | `bin/probe4/g6.ax` | 250 | **255** |
| `let c: f64 = 3 + 1` (int **expression**) | `bin/probe4/g5.ax` | 187 | **255** |
| int **VARIABLE** at let / assign / arg | `bin/probe4/g7.ax` | 100 | **107** |

Root cause, again the repo's signature shape: **a rule existing in a narrower form than the thing
it must cover.** `air_builder.coerce_float_arg_ft` only converted float→float (it returned
unchanged the moment the argument was not already a float), and `typecheck.ax`'s expected-type hint
was `TYPE_F32` only — its own comment admitted the narrowing: *"Scoped to F32 only, so int/i64
scalar fields keep their prior UNKNOWN inference and the self-host fixpoint is unaffected."* So
every **f32** twin passed and every **f64** one read garbage.

**Fix — one rule, every value site** (`bootstrap/stage1/air_builder.ax`):
`coerce_int_to_float(target_type, src_node, src_reg)`, deliberately built as the twin of
`coerce_struct_to_interface` and called at exactly the same list of sites: let, assign, array
element (literal + assign), struct field (init + assign), return, call argument (free / method /
dynamic), global init. Self-guarding, so it is a no-op unless the target is f32/f64 AND the source
node's type is a *definite* integer — in particular a literal typecheck already adopted (the fix
above) arrives float-typed and is untouched. `coerce_field_to_interface` was renamed
`coerce_field_value` because a helper named after one rule is precisely how the previous coercion
drifted into f32-only. **Nothing in typecheck changed**, so the feared hint-widening fixpoint
exposure never materialized.

**§9 invariant added**: `verify_air_no_int_into_float` — an INT-classed value must not reach a
float consumer (float-typed OP_COPY, OP_F{ADD,SUB,MUL,DIV}) without an OP_ITOF. Tiny one-sided
abstract domain; `type_id == 0` never counts as INT (so a no-init `mut x: f64` is not a false
alarm). Argument and struct-field sites are **not** covered — AIR carries neither the param type on
OP_CALL nor the field type on OP_SET_FIELD — and the oracle guards those instead.

⚠️ **Bug found while writing that verifier, worth more than the verifier**: AIR's `src1`/`src2` are
OVERLOADED — an IMMEDIATE for OP_ICONST, the two 32-bit halves of the IEEE pattern for OP_FCONST, a
BLOCK index for OP_JUMP/OP_BRANCH. Sizing a per-vreg map by `max(dest, src1, src2)` therefore made
one `3.0` literal (src2 = 1074266112) request a gigabyte, and the compiler *appeared to hang* (no
error, no output file, rc=124). It did not reproduce on the compiler's own source — which is
integer-only — only on a program with float literals. **Bound anything per-vreg by `dest` only.**

Gate: **A == B == C**, `DA5C96AC76D268B2626BC9538764051D7096540D432E59EDAE109377E5BCB72F`. A==B is
the criterion *and* the proof of self-inertness: A is built by the old logic from the new source and
B by the new logic from the same source, so their equality says the new coercion emits nothing extra
for the compiler's own code.

Still open, deliberately untouched (separate backlog rows): **f64 → f32 narrowing** is still
silently accepted and wrong (`let s: f32 = d` → `bin/probe6/g_f64tof32.ax` still exits 1 — per §4 it
should be a REJECT, not a conversion), and the reverse direction `let a: i64 = 3.0`
(`bin/probe4/g13.ax`, exit 3 = raw IEEE bits copied into an integer).

## Oracle traps that bit this investigation (both are encoded in the oracle's header)

- **Never route the value under test through a parameter of the same float type** to check it — the
  arg site coerces it and hides the defect.
- **A row can pass in a multi-row program purely because an earlier float compare left the right
  value in an XMM register.** `return 3` from an `-> f64` fn did exactly that. Keep rows
  independent and re-verify any suspicious pass in a minimal file (`bin/probe4/g8.ax`).

Related: [[lesson-exit-code-8bit-masking]], [[bug-generic-explicit-typearg-float-literal]],
[[bug-float-arg-reg-unprotected]].
