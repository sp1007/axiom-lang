---
name: m4-compliance-suite-spec-vs-impl-gap
description: "M4 milestone assessment (2026-07-18): tests/axiom_compliance_suite.ax (100 tests, 10 groups) + its std/testing.ax dependency are written in the ASPIRATIONAL SPEC DIALECT (brace blocks, !not, default args, trait bounds, format(), async/gpu/quantum) which the implemented self-hosting compiler does NOT parse. M4 '100 tests pass' is blocked on a large spec-vs-impl gap, not a bounded bug. Needs a user decision on approach."
metadata:
  node_type: memory
  type: project
---

# M4 compliance suite — spec-vs-implementation dialect gap (needs a decision)

## ✅ RESOLVED 2026-07-18 — approach (b)/(c) shipped: `bin/t_compliance.ax` (60/60, gated)
User (2026-07-18) chose **rewrite-in-real-grammar** and granted standing autonomy
([[autopilot-direction-2026-07-18]]). M4 is now **DEFINED** as: the compliance suite runs on
the ACTUAL shipped grammar. Deliverable = **`bin/t_compliance.ax`** — 60 tests across groups
1–6 (primitives, control flow, functions/lambdas/tuples, structs & methods, generics &
collections, sum types & error handling), each asserting a known-correct value; **exit code ==
number of passing tests = 60**. Builds clean on daily driver + runs 60 on the native path.
**Gated forever** as `t_compliance|exit|60` in `scripts/regression_repros.sh`. **Part 2 added**
(`bin/t_compliance2.ax`, `t_compliance2|exit|28`): 28 more real-grammar tests over the stdlib
surface — Vec HOFs (map/fold/filter/for-in), HashMap (insert/get.unwrap/len/for-in keys), arrays
(literal+index+for-in), strings (concat/split/trim/contains/to_upper/starts_with/index_of/for-in
chars), numeric formatting (to_str/to_hex UFCS), and control-flow variations (nested for, while
break/continue, cast chain, tuple .0/.1, nested if-expr). Built clean + exit 28 first try. Combined
M4 compliance surface = **88 real-grammar tests** across the language + stdlib.
- The aspirational spec-dialect original is preserved at `tests/axiom_compliance_suite_aspirational.ax`;
  `tests/axiom_compliance_suite.ax` is now a pointer doc to the real suite.
- **Dialect gaps SIDESTEPPED (not blockers):** local `const` → module-level `const X: T = v`;
  `match`-as-expression → RFC 0023 if-expression value form (`let r = if c: a else: b`); `=>`
  arms → `Pattern:` block arms; `interface`/`impl` → struct-embedded + duck-typed methods;
  closure-capture-of-locals → zero-capture lambdas only. The one real bug from measurement
  (`type ID = i32|string` match segfault) was already fixed ([[bug-sum-of-primitives-match-segfault]]).
- **Groups 7–10 (async/comptime `#run`/gpu/quantum/AI) remain out-of-scope** — milestone-scale
  subsystems, tracked separately; M4 no longer waits on them.
- Frontend/test-only change (no compiler/stdlib source touched) → no fixpoint required; gated by
  full regression staying GREEN.

Original assessment (kept for history) follows.


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

### Running M4-core baseline (measured on axc_native `1C2E3D6A`, impl grammar rewrite)
⚠️ CRITICAL: the impl lambda syntax is **Rust-style pipes with an EXPRESSION body**:
`|a: i64, b: i64| -> i64 a + b` (see bin/t_lambda.ax), NOT the spec's `fn(a)->T: return ...`.
An interim note wrongly flagged lambdas as broken (I used spec syntax); CORRECTED below.
- **Group 1 (001-010): 9/10.** GAP: **LOCAL `const NAME = value`** (const inside a fn body →
  parse error). ✅ TOP-LEVEL `const X: T = v` WORKS (bin/t_closurecap.ax, exit verified). So the
  gap is narrow: local-scope const. Rest of group 1 supported.
- **Group 2 (011-020): 10/10.** All control flow works. `=>` match-arm is surface dialect
  (impl uses `pattern:` block); the suite has a malformed `let arr =` (missing `[1,2,3]`).
- **Group 3 (021-030): 9/10 (CORRECTED from a wrong ~6/10).** SUPPORTED: fn call, `mut` param
  by-value, recursion, higher-order fn-ptr param (BUG#49), `let op = named_fn`, tuple return
  `-> (i32,i32)` + destructuring `let (x,y)=f()`, **lambda literals `|x|-> T expr`** (impl
  syntax — [[next-step-16-fnptr-shipped]]/2cc67ed CONFIRMED correct), **zero-capture closures**
  (over globals/consts). GAP: **closure capturing a LOCAL var** (test_030 `outer`) → clean
  REJECT "only zero-capture closures are currently supported (RFC 0008 P2 not yet implemented)".
- **Group 4 (031-040): 8/10.** SUPPORTED: struct instantiate `Point(x:..,y:..)`, mut field
  assign, value copy/move `let b = a`, **borrow `&a` + `view.x`**, UFCS method `p.area()`
  (top-level fn), **`in [Arena]:` region block** (works!), scoped-lifetime/CTGC scopes.
  GAPS (both aspirational ownership annotations, syntax not parsed): **`@SOA` struct attribute**
  (033, "expected top level declaration") and **`!Point` sink parameter** (036, "expected type
  expression"). test_037 uses a NESTED fn decl (rewritable to top-level; UFCS itself works).

**Emerging pattern (updated):** the core LANGUAGE is MORE complete than first thought —
**~36/40 of groups 1-4 work**. Real gaps are all cleanly diagnosed (no silent miscompiles):
`local const`, `closure-capture-of-locals` (RFC 0008 P2), `@SOA`, `!Point` sink. The rest are
pure surface-syntax dialect (`fn(...)` lambdas → impl `|..|`, `=>` match arms → `pattern:`,
brace blocks → indentation) that an approach-(b) suite rewrite sidesteps. Groups 5-6
(interfaces/generics, error/sum — both heavily exercised & fixed already, likely high pass)
still to measure. **Bounded tasks surfaced, smallest first:** (1) **local-`const` support**
(parser: allow `const` in a block; top-level machinery exists — likely small, but a syntax
addition ⇒ light RFC per §13); (2) **RFC 0008 P2 local capture** (larger; RFC exists);
(3) `@SOA` / `!` sink (aspirational ownership — defer, design-level).

- **Group 5 (041-050): interfaces/generics — PARTIAL.** SUPPORTED (from prior work + spot
  checks): generic structs `Box[T]`, multi-param `Pair[K,V]`, nested generics `Box[Box[i32]]`,
  generic inference `Box(value:88)`, monomorphization. GAP: **`interface`/`impl ... for`
  declaration syntax does NOT parse** ("unexpected token") — zero interface examples exist in
  bin/*.ax; interface dispatch + trait-bound `[T: Drawable]` are unimplemented (matches backlog
  "interface-vtable dispatch open (niche)"). Tests 041/044/045 need it. Structural-duck-typing
  method dispatch `p.method()` via a top-level fn DOES work (UFCS).
- **Group 6 (051-060): error/sum — mostly SUPPORTED with 1 BUG.** SUPPORTED: Result Ok/Err,
  Option Some/None, match on them, error propagation via match, custom error struct in Err
  (all heavily exercised, [[probe-batch-clean-2026-07-18-b]]). GAPS: **match-as-expression**
  `let x = match ...:` (059) → parse error (not supported; RFC 0023 did if-expr, not match-expr);
  **`{.raises: [E].}` effect annotation** (060) → parse error (aspirational effects). 🐞 **REAL
  BUG:** `type ID = i32 | string` bare-primitive union + `i32(v)` variant match → accept-then-
  SEGFAULT (057/058) — see [[bug-sum-of-primitives-match-segfault]] (bounded reject fix
  available).

### M4-core (groups 1-6) SUMMARY — measurement COMPLETE
Core language is largely implemented. **Bounded/small tasks** (fast gate): local `const`;
reject bare-primitive-union match [[bug-sum-of-primitives-match-segfault]]. **Larger/RFC:**
closure-capture-of-locals (RFC 0008 P2); interface/impl + trait bounds; match-as-expression.
**Aspirational/defer (design-level):** `@SOA`, `!` sink, `{.raises.}` effects, and groups 7-10
(async/comptime/gpu/quantum/AI). Everything is CLEANLY DIAGNOSED except the one sum-primitive
segfault. Approach-(b) suite rewrite would score ~45-50/60 on core today; the rest map to the
bounded/RFC tasks above. **This method (rewrite a compliance group in impl grammar, run, triage
gaps) is the productive replacement for exhausted random probing — it already found a real bug.**
