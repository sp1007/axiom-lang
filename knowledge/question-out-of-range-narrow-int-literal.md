---
name: question-out-of-range-narrow-int-literal
description: "OPEN QUESTION (not a claimed bug) 2026-07-30 — an integer literal that does NOT fit its annotated narrow type keeps its full value: `let x: u8 = 300` yields 300, not 44, and no diagnostic is emitted. Whether the correct behaviour is wrap, reject, or something else is a spec decision that has not been made."
metadata:
  node_type: memory
  type: project
---

# OPEN QUESTION — out-of-range integer literal for a narrow type

**This is deliberately filed as a QUESTION, not a bug.** The behaviour is measured; what it *should*
be is a spec/design decision that has not been taken, and asserting one in a test would pin an
unverified expectation.

## Measured (2026-07-30)

    fn main() -> i64:
        let x: u8 = 300
        let y = x as i64
        if y == 300:
            return 42          // <-- this fires
        if y == 44:
            return 43
        return 99

`bin/probe2/w4.ax` returns **42**: the value is **300**. It is NOT narrowed to `300 & 255 = 44`, and
**no diagnostic is emitted**. Same for the generic form `pick[u8](300, 1)` (`bin/probe2/w3.ax`), so
it is not generic-specific.

⚠️ Read via an exit code this looks like 44 and therefore like correct wrapping. It is not — see
[[lesson-exit-code-8bit-masking]]. Compare in-program.

## The three defensible answers, and why the choice is not obvious
1. **Wrap to width** (`300 -> 44`). Consistent with `emit_wrap_to_width` (RFC 0006), which the
   backend already applies to narrow arithmetic. Silent data loss, but predictable and matches C.
2. **Reject** with a diagnostic ("literal 300 does not fit u8"). Consistent with this project's
   strong BUG#53 convention — accept-then-miscompile is treated as the worst outcome, and a literal
   that cannot be represented is knowable at compile time. Rust does this.
3. **Widen the binding's type** — what [[bug-negative-literal-compare-o0]] chose for a DIFFERENT
   case: an integer literal too large for i32 now infers i64 *by magnitude*. But that case had no
   annotation to respect; here the user WROTE `u8`, so silently widening would contradict them.

⭐ Note that (3) is already precedent in this codebase for unannotated literals, which is exactly why
this needs a decision rather than an assumption: the existing rule "type by magnitude" cannot simply
be extended to an annotated position.

**Recommendation if asked: (2) reject.** It is the only option with no silent outcome, it matches the
project's own BUG#53 convention, and the value is fully known at compile time. Wrapping can be
offered later behind an explicit cast (`300 as u8`), which is already the spelling for "I mean this
truncation."

## Scope check before acting
Current behaviour stores an out-of-range value in a narrow-typed binding, so `u8` does not hold a
`u8`. Anything downstream that assumes the range invariant (array indexing, `& 0xFF` elision, sum
tags) could be affected. Verify how far it reaches before choosing — and note that a REJECT is a
frontend change (`A==B` gate) while a WRAP touches codegen.
