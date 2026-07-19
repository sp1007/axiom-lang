---
name: probe-batch-2026-07-19b-shift-find
description: "Probe session 2026-07-19b: FOUND + fixed a real load-bearing codegen bug (variable-shift-in-loop, see [[bug-variable-shift-in-loop]]) then swept ~17 more loop-context crosses CLEAN across 4 batches. Documents which crosses are covered so future autopilot sessions don't re-probe them. Yield after the find: 0/17 → plateau reconfirmed."
metadata:
  type: project
---

# Probe session 2026-07-19b — one real find, then clean plateau

## The find (fixed, shipped)
`p4_shift` cross (arithmetic + shift in a loop) returned an **O-level-consistent** wrong
answer (42 vs hand-oracle 99) — invisible to O0-vs-O1 comparison, caught only by a
hand-oracle + decomposition. Root-caused to `const_shift_amount` missing a non-SSA
def-count guard → **variable-shift-in-loop miscompile**, FIXED `5a22500` (B==C af3ba8e3,
439/439). Full detail [[bug-variable-shift-in-loop]]. Reusable technique: **a consistent
wrong answer across O-levels needs a HAND-oracle** — O0-vs-O1 alone is blind to it.

## Crosses swept CLEAN after the find (do NOT re-probe these — all O0==O1==O2, hand-oracle matched)
All in loop contexts (the fruitful neighborhood), driver af3ba8e3:
- **Adjacent to the shift bug:** div by loop-variant divisor (`120/k`), mod by loop-variant
  (`100%k`), doubly-loop-variant shift (`acc << k`, both operands change → 138 ✓, the fix
  handles it), AND/OR/XOR with loop-variant operand, div/mod-by-const-pow2 with loop-variant
  value (strength-reduction path). → `const_divisor_pow2` guard confirmed sound at runtime.
- **Fresh neighborhood (memory-store / index-address path):** in-place bubble sort (nested
  loops, `a[j]`/`a[j+1]` read+write aliasing), two-pointer array reverse, sum-payload match
  in a loop (`Circle(k)`), Vec-of-struct field accumulation (`for r in v: s += r.w`), nested
  Option unwrap in a loop with a None branch.
- **Cast / signed / global path:** narrowing cast wrap in loop (`x as u8` crossing 256),
  signed division round-toward-zero over a negative→positive range, signed modulo sign,
  global-variable mutation across calls in a loop (`counter` RFC 0017), i32/i64 mixed-width
  accumulator.
- **First batch (the one that found it):** HashMap[i64,Vec[i64]] (container-valued map),
  mixed-width sum payload (`I(i32)|F(f64)|S(i64)`), HOF chain `map→filter→fold`,
  mutual-recursion struct return, nested `Option[Result[i64,i64]]` double-match,
  array-of-struct index+field mutation.

## Session 2026-07-19c continuation — 2nd find (defer) + more clean batches
- **2nd real find:** `defer` in control flow silently miscompiled (static registration) →
  REJECTED `dcac520`. See [[bug-defer-in-control-flow]]. Found via F3 (defer in loop ran once)
  + G1 (defer under `if false:` ran anyway).
- **Clean batches after the defer find (don't re-probe):** short-circuit reachability
  (`false and side()` / `true or side()` do NOT run RHS; while-cond short-circuit), nested
  break/continue (exit/skip only inner loop), integer boundary (near i64::MAX/MIN, O0==O1
  const-fold vs runtime), match wildcard `_` + nested match + match-on-sum-inside-match,
  Option `.unwrap_or`. All O0==O1==O2, hand-oracle matched.
- **Self-as-param CONFIRMED WORKING** (SP6/SP8/SP9 →42): not a gap; see [[backlog-open-items]].
  Grammar notes banked: `defer Expr` only (assignment isn't an expr); untyped `self` needs a
  struct body (RFC 0002), top-level free-fn method needs `self: Type`.

## Conclusion
1 real bug per ~18 programs at this plateau — probing still pays (this find was load-bearing),
but the loop-codegen surface is now well-covered. Future probing should target GENUINELY new
axes (string ops in loops, deeply-nested generics 3+ levels, defer×loop×early-return ordering,
closures over loop state) rather than re-crossing arithmetic/index/container/option-in-loop.
See prior clean batches [[probe-batch-1012-clean-2026-07-18]], [[probe-batch-clean-2026-07-18-b]].
