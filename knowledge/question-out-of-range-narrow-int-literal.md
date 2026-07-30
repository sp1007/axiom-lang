---
name: question-out-of-range-narrow-int-literal
description: "DECIDED 2026-07-30 (user, option 2 = REJECT) — an integer literal that does not fit its ANNOTATED narrow type is now a compile error, error[E3030]. Was: silently accepted at full width (`let x: u8 = 300` held 300, no diagnostic). Implemented in typecheck.ax::check_int_lit_range at 8 annotated positions (method arguments still uncovered); RFC 0006 §6.1. Unannotated magnitude-based inference is deliberately untouched."
metadata:
  node_type: memory
  type: project
---

# DECIDED — out-of-range integer literal for an annotated narrow type ⇒ REJECT (E3030)

**Status: DECIDED by the user on 2026-07-30 — option (2) REJECT. Implemented the same day.**
Filed originally as a QUESTION (the behaviour was measured, the correct answer was a spec
decision nobody had taken). The alternatives below are kept for the historical record; do not
re-open them without a new user decision.

## The decision

An **integer literal** written where the programmer **explicitly annotated** a narrow integer
type, whose value is not representable in that type, is a **compile error**:

```
error[E3030]: integer literal `300` is out of range for type `u8` (valid range 0 to 255)
   |
   |     let x: u8 = 300
   |                 ^^^ value does not fit in `u8`
   |
   = note: the type `u8` is written explicitly at this `let` binding, so the literal is not widened to fit
   = help: write `300 as u8` if truncating to `u8` is intended
```

Rationale as given: it is the only option with **no silent outcome**. The value is known at
compile time; the user wrote the type themselves, so silently widening the binding betrays what
they wrote and silently wrapping loses data. `300 as u8` remains the spelling for "I intend to
truncate". Matches the project's BUG#53 convention that accept-then-miscompile is the worst
outcome.

Spec: **RFC 0006 §6.1** (a requirement, not a proposal — it is the first enforced piece of the
constant-overflow checker §6 already mandated). Implementation:
`bootstrap/stage1/typecheck.ax::check_int_lit_range` + 7 call sites (8 positions).

## What was measured before the fix (2026-07-30)

    fn main() -> i64:
        let x: u8 = 300
        let y = x as i64
        if y == 300:
            return 42          // <-- this fired

`bin/probe2/w4.ax` returned **42**: the value was **300**, NOT narrowed to `300 & 255 = 44`, and
**no diagnostic** was emitted. Same for the generic form `pick[u8](300, 1)` (`bin/probe2/w3.ax`),
so it was never generic-specific.

⚠️ Read through an exit code this looked like 44 and therefore like correct wrapping. It was not
— see [[lesson-exit-code-8bit-masking]]. Every one of the nine reject oracles exits **44** (or 255) on the
pre-fix compiler for exactly this reason. Compare in-program.

## Positions covered (8) — and the OLD behaviour each had, measured

| Position | Old behaviour |
|---|---|
| `let`/`const` with an annotation (incl. a global) | kept **300** |
| assignment to an **annotated** binding (`mut x: u8 = 1; x = 300`) | kept 300 |
| assignment to a struct **field** (`s.v = 300`) | **wrapped to 44** (byte store) |
| call argument against a declared narrow param | kept 300 |
| call argument in the **explicit type-arg** form (`pick[u8](300,1)`) | kept 300 |
| struct-field initializer (`P(a: 300)`) | wrapped to 44 |
| element of an array literal under an **annotated** array type | wrapped to 44 |
| `return` against a declared narrow return type | kept 300 |

⭐ Two different silent failure modes, one verdict. Positions that keep a value out of its own
type's range and positions that quietly truncate are both "silent"; the decision rejects both.

Ranges enforced: i8 `-128..127`, i16 `-32768..32767`, i32 `-2147483648..2147483647`,
u8 `0..255`, u16 `0..65535`, u32 `0..4294967295`; every `u*` rejects every negative literal.
Literals are quoted AS WRITTEN, so `0xFFFFFFFF` at `i32` is reported as `0xFFFFFFFF`.

## Deliberately NOT covered (report honestly, do not describe as complete)

1. ⚠️ **METHOD arguments** — `s.setv(300)` with `setv(self, x: u8)` is still ACCEPTED and
   silently wraps to 44. Measured, not assumed. Different mechanism: method calls resolve
   params through the `mfi.params` symbol scan (`typecheck.ax` ~L4640), not the `fp_data` path
   the free-function/explicit-type-arg checks use. That scan is the hook if this is extended.
2. **Assignment to an array ELEMENT** (`a[0] = 300`) — deliberate: an array's element type may
   be inferred (`mut a = [1,2,3]` is `[i32;3]`), which is not a type the user wrote.
3. **i64/isize** — `parse_comptime_int` wraps at 64 bits, so a literal beyond i64 cannot be told
   apart from a legal one.
4. **u64/usize** — only a syntactically negative literal is rejected. `let big: u64 =
   18000000000000000000` parses to a negative i64 and MUST stay legal (t_u64cmp pins it).
5. **Folded constant expressions** (`let b: u8 = 255 + 1`, mandated by RFC 0006 §6) — needs
   constant folding in typecheck; still open.
6. **Unannotated positions** — see the precedent below.
7. **Runtime narrowing** (`let x: u8 = some_i64`) — a separate, still-undecided policy question.

## ⭐ The precedent that constrained the design

[[bug-negative-literal-compare-o0]] deliberately made an integer literal too large for i32 infer
**i64 by MAGNITUDE**. That case has **no annotation to respect**; this one does. So the reject
fires *only* where an explicit type exists, and magnitude inference at unannotated positions is
untouched — verified by a control row (`t_intrangeok` rows 14–16), not assumed.

A related consequence, measured and unchanged before/after: `let arr = [1, 5000000000]` still
truncates element 1, because an unannotated array literal takes its element type from the FIRST
element. That is a separate latent defect, not part of this rule.

## The alternatives that were rejected (historical record)

1. **Wrap to width** (`300 -> 44`). Consistent with `emit_wrap_to_width` (RFC 0006) and with C.
   Rejected: silent data loss at a place where the compiler knows the answer.
2. **Reject** — chosen.
3. **Widen the binding's type**. What [[bug-negative-literal-compare-o0]] chose for the DIFFERENT,
   unannotated case. Rejected here: the user wrote `u8`, so widening contradicts them.

## Verification (2026-07-30)

- Breakage audit: **0** hits across the compiler's own 2 MB source (twice, through both fixpoint
  hops) and **833** other `.ax` files (`bin/*.ax`, `tests/`, `examples/`, `std/`, `stdlib/`) —
  the only E3030 hits are the 9 intentional reject oracles.
- Calibration: all 9 reject oracles **built** (exit 44 / 255) on **two** independent pre-change
  compilers (`bin/axc_pre1f.exe`, `bin/axc_baseline_keep.exe`) and are **rejected** after;
  `t_intrangeok` returns 42 before and after, at -O0/-O1/-O2.
- Boundary table verified row by row (21 rows): `u8` accepts 0/255 and rejects -1/-128/256/300;
  `i8` accepts -128/127 and rejects -129/128; i16/u16/i32/u32 edges likewise; `u64`/`usize`
  reject only `-1`; `let big: u64 = 18000000000000000000` and `i64` max still accepted.
- Gate: fast fixpoint **A == B** = `7829550985DFDE0208648A73ADA8D0E64553A69B472404E41DA289E39C13B1B9`;
  regression **607/607** at default and at `-O0` (597 baseline + 10 new rows).
