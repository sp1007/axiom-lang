---
name: bug-defer-in-control-flow
description: "FIXED (reject) 2026-07-19c dcac520 (A==B 1a934e34): `defer Expr` is registered STATICALLY at AIR lowering (air_builder self.defers, textual order) and flushed at function exit, so a `defer` inside a non-taken `if` branch ran anyway, and a `defer` in a loop ran ONCE regardless of iterations. Both = accept-then-miscompile (BUG#53 class). Fix = typecheck check_defer_placement rejects `defer` nested in if/while/for/match; unconditional top-level defer (the only sound case) still works. defer unused in compiler/std -> self-build-safe. Found via autopilot probing (F3/G1)."
metadata:
  type: project
---

# `defer` in control flow — silent miscompile, now REJECTED (`dcac520`)

## Symptom (found via probing)
- `defer bump()` inside a `while i<3:` loop ran the deferred call **once** (c=1), not 3× (F3).
- `defer bump()` inside `if cond:` with **cond=false** ran the deferred call **anyway**
  (c=1, should be 0) (G1) — the sharper witness.
- Straight-line top-level defer is fine: runs at function exit AFTER the return value is
  captured, LIFO order (F1=2, F2=21).
- `defer x = x + 100` is a parse error — the grammar is `defer Expr` and an assignment is a
  statement, not an expression (do not mistake this for a bug).

## Root cause
`air_builder::lower_defer` records the deferred expression's node index into `self.defers`
(a `U32Vec`) at **lowering time** — statically, once per textual `defer`, in source order —
and `flush_defers` runs them (LIFO) at the single function-exit point. So registration is
compile-time/static, NOT runtime/per-execution. For unconditional straight-line code
static==dynamic. Under control flow it diverges: a defer under a non-taken branch is still in
`self.defers` (runs anyway); a defer in a loop is registered once (runs once, not per-iter).

Go/dynamic semantics would register per *execution* of the `defer` statement (loop → N runs,
non-taken branch → 0 runs). Implementing that needs a runtime defer stack (push captured
action on execution, pop-and-run at exit) + argument-capture semantics = an RFC-scale change.

## Fix (minimal, BUG#53 convention: reject > silent miscompile)
`typecheck.ax::check_defer_placement(node, in_cf)` — a one-shot recursive AST walk called per
top-level function-body statement (in_cf=false). It sets in_cf=true when descending into an
`if`/`while`/`for`/`match` node and, on reaching a `NODE_DEFER_STMT` with in_cf=true, emits a
diagnostic + `diags_count++` (driver halts before codegen). Recursion **stops at a nested
`NODE_FUNC_DECL`** so each function's defers are judged in its own scope. The only sound case
— an unconditional top-level defer — is unaffected.

Self-build-safe: `defer` has ZERO uses in the compiler/std sources (grep), so nothing is
newly rejected → A==B `1a934e34`.

## Gate
Frontend reject → A==B `1a934e34`, full regression **441/441** (+`t_deferctrl` reject =
defer-in-if; +`t_defertop` exit 2 = top-level defer + LIFO + capture-then-run still correct).

## Follow-up (deferred, needs RFC)
Proper **dynamic defer** (per-execution registration via a runtime defer stack, Go-style)
would lift the restriction. Until then defer is top-level-only. Related: BUG#53
accept-then-miscompile cluster [[bug-malformed-input-robustness-cluster]].
