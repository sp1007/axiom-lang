---
name: m4-compliance-suite-spec-vs-impl-gap
description: "M4 milestone assessment (2026-07-18): tests/axiom_compliance_suite.ax (100 tests, 10 groups) + its std/testing.ax dependency are written in the ASPIRATIONAL SPEC DIALECT (brace blocks, !not, default args, trait bounds, format(), async/gpu/quantum) which the implemented self-hosting compiler does NOT parse. M4 '100 tests pass' is blocked on a large spec-vs-impl gap, not a bounded bug. Needs a user decision on approach."
metadata:
  node_type: memory
  type: project
---

# M4 compliance suite — spec-vs-implementation dialect gap (needs a decision)

**Context:** M4 ("MVC v0.1.0", milestones.md:12) gate = "`axc build` works, **100 compliance
tests pass**". M6 (:14) reuses the same 100 tests via the native/ELF backend. So the
compliance suite is the gate for the next two externally-demonstrable milestones.

**The artifact:** `tests/axiom_compliance_suite.ax` — 681 lines, exactly **100 test fns**
(`test_001`..`test_100`) in **10 groups**:
1. Basic syntax & primitives (001-010)   6. Error handling & sum types (051-060)
2. Control flow (011-020)                 7. Concurrency, actors & async (061-070)
3. Functions & closures (021-030)         8. Compile-time execution `#run` (071-080)
4. Structs & ownership/CTGC (031-040)     9. Standard library (081-090)
5. Interfaces & generics (041-050)       10. AI hooks / quantum / GPU / native interop (091-100)

**BLOCKER (found 2026-07-18): the suite is written in the aspirational SPEC dialect, which
the implemented self-hosting compiler does NOT parse.** It fails to compile against daily
driver `axc_native 1C2E3D6A` (parse errors in the suite body). Both the suite and its
dependency `std/testing.ax` use constructs the bootstrap language lacks:
- **brace blocks** `fn assert(...) { ... }` (impl uses `:` + indentation)
- **C-style `!condition`** (impl uses `not`)
- **default parameter values** `message: str = "assertion failed"` (impl: none)
- **trait bounds** `[T: Eq + Display]` (impl: generics without bound syntax)
- **`format("... {} ...", x)`** interpolation, `panic()` (aspirational stdlib)
- top-level imports of **aspirational modules** `std.net/gpu/quantum/compiler.ai/concurrency`
  (don't exist / don't parse — see [[next-step-16-fnptr-shipped]], [[backlog-open-items]])
- even group 1 uses `.length()` (impl: `str.len` is a FIELD, not `.length()` method),
  `const NAME = ...`, `0xFF`/`0b1010` literals — each needs verification against the impl.

**Implication:** M4 as literally specified is NOT a bounded bug — it is a large program of
work bridging the spec/impl gap. Groups 7-10 (async, comptime `#run`, GPU/quantum/AI) depend
on whole unimplemented subsystems; groups 1-6 are core language but still use spec-dialect
surface syntax the suite assumes.

## Decision needed (user) — three approaches
- **(a) Implement the spec dialect** incrementally (brace blocks, `!`, default args, trait
  bounds, format-strings, then comptime/async/...). Faithful to spec but multi-milestone;
  brace-vs-indentation is an RFC-level grammar decision (the impl deliberately chose
  indentation). Groups 7-10 still need major subsystems.
- **(b) Rewrite the suite in the implemented grammar** (indentation, `not`, `.len`, no
  default args) for the subset of tests whose FEATURES exist — realistically groups 1-6
  (~60 tests), deferring 7-10 until their subsystems land. Mechanical, measurable, and turns
  each remaining failure into a bounded feature-gap task (a far richer backlog than the now-
  exhausted random probing — [[probe-batch-clean-2026-07-18-b]]).
- **(c) Redefine M4** around the implemented language (a new core-compliance suite authored
  in real grammar). Cleanest measurement of what actually ships.

## Recommendation
**(b)-scoped**: author `tests/compliance_core.ax` — groups 1-6 rewritten in the implemented
grammar with a plain `main()` harness (no std.testing braces; use direct `if x: return N`
exit-code asserts like the existing `bin/t_*.ax` oracles). Compile+run against `axc_native`,
record each feature that fails as a bounded spec-driven task, fix incrementally through the
fast gate ([[infra-defender-build-throttle]]). This gives a real, growing M4-core score and
a productive bounded-bug pipeline. Defer groups 7-10 (blocked on async/comptime/gpu/quantum).
**Autonomous-safe** except approach choice (a/b/c) and the brace-grammar RFC — those need the
user. Until a choice is made, the compiler stays stable; this is the derived next direction.

## Baseline measurement started 2026-07-18 (approach-b spike, group 1)
Rewrote group 1 (tests 001-010) in the implemented grammar (indentation, `not`, `.len`
field, direct `if COND:`\n`    ...` blocks, oracle-style pass-count return) and ran on
`axc_native 1C2E3D6A`: **9/10 core features already work.** Confirmed SUPPORTED: i32/f64/
bool decls, `mut`+reassign, `and`/`not`, string literal + `.len`, **block string `"""..."""`
(RFC 0024)**, `char8 'A'`, type inference `let x = 1000`, hex `0xFF` + bin `0b1010`.
**ONLY GAP: `const NAME = value`** → parse error "expected expression nud" (test_008). `const`
is a new keyword/syntax the impl lacks (impl has `let` immutable + `mut`); adding it is a
SYNTAX change ⇒ RFC per CLAUDE.md §13 (not an autonomous bugfix). Substituting `let` for
`const` makes group 1 pass 10/10.
Also re-confirmed impl-grammar constraints when rewriting the suite: **inline `if x: stmt`
is REJECTED** ("inline ':' suites are not supported" — must be an indented block on the next
line); `.length()` doesn't exist (`.len` is a field). 
**Immediate next autonomous step:** measure groups 2-6 the same way (each ~10 tests, quick
rewrite+run) to produce the full M4-core (001-060) baseline + gap list. Each gap is then a
bounded task (RFC if syntax, else a plain fix). Groups 7-10 stay deferred (async/comptime/
gpu/quantum subsystems). First known gap to schedule: **RFC + impl for `const` declarations.**
