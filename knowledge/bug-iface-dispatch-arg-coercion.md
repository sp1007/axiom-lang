---
name: bug-iface-dispatch-arg-coercion
description: "BUG 2026-07-31 (FIXED) — dynamic dispatch through an interface receiver dropped ALL argument coercion: `i.c32(1.5)` delivered exactly 0.0. Static calls were correct because air_builder.coerce_float_arg reads the param types off the resolved callee SYMBOL, and dispatch has no callee symbol. Fixed by publishing the interface's declared method signatures on its type entry (from the single interface_method_sig walk) and calling the SAME coercion rule from the dispatch site. RFC 0029 §9."
metadata:
  node_type: memory
  type: bug
---

# Dynamic dispatch passed every argument unconverted (`i.c32(1.5)` → 0.0)

**Status:** FIXED 2026-07-31. Oracle `bin/t_ifacefloatarg.ax` (30 → 42). RFC 0029 §9.
Gate: **A == B == C** (all three hashes `5b0eb92c…`), regression **611/611** at default and `-O0`.

## Symptom

`i.c32(1.5)` on an interface-typed receiver delivered **exactly 0.0** to the callee. Not garbage —
`0.0`, which is the tell: the low 32 bits of the double `1.5` (`0x3FF8000000000000`) are zero and
the callee's `movss` reads precisely those. Accept-then-miscompile, BUG#53 class.

Control matrix (measured, `bin/probe4/`):

| form | before |
|---|---|
| static `s.c32(1.5)` | correct |
| static `s.c64(v)`, `v: f32` (needs cvtss2sd) | correct |
| dynamic `i.c32(v)`, `v: f32` (no conversion needed) | correct |
| dynamic `i.c32(d as f32)` (explicit cast) | correct |
| dynamic `i.c64(1.5)` (f64 param, float literal) | correct |
| dynamic `i.i32lit(5)` / `i.u8lit(200)` | correct **by accident** (GPR sub-register overlap needs no instruction) |
| **dynamic `i.c32(1.5)`** | **0.0** |
| **dynamic `i.c32(2.0+2.0)`** (folded f64 expr) | **wrong** |
| **dynamic `i.c64(v)`**, `v: f32` → f64 param | **wrong** |
| **dynamic, f32 param in slot 1 / 2 / after i64 / after f64** | **wrong in every position** |

⇒ *any* dynamic-dispatch argument requiring a real conversion instruction was passed unconverted.

## Root cause — and why the first diagnosis pointed at the wrong phase

The task was briefed as a typecheck bug: thread the declared parameter types in as `expected` at
the dispatch site. **Measurement says argument coercion for method calls does not happen in
typecheck at all.**

- `typecheck.ax`'s expected-type threading block is gated to `NODE_IDENT` / `NODE_INDEX_EXPR`
  callees. A method call's callee is a `NODE_FIELD_EXPR`, so **static** method arguments are
  already inferred with `expected = TYPE_UNKNOWN` — yet they are correct.
- They are correct because `air_builder.coerce_float_arg` emits `OP_CAST` after reading the
  parameter types off the **resolved callee symbol**.
- `coerce_float_arg` returns at its first line when `fn_sym == 0`. Dynamic dispatch has no callee
  symbol *by construction*. That is the whole bug: not a missing accessor — a missing **signature**
  at the one site that already does this job.
- The accessor-only fix could not have worked anyway: `i.c64(v)` with `v: f32` needs an emitted
  `cvtss2sd`, and threading an expected type does not rebuild an ident's value (the codebase says
  so itself, at the Option/Result width reject: *"Threading an expected type does not REBUILD an
  ident's value, so a coercion is impossible here"*).

## Fix (RFC 0029 §9)

1. `typecheck.interface_method_sig` — the single walk that already resolves an interface method's
   arity and return type — now also yields the declared **parameter** types (optional out-param,
   `self` at index 0). Same walk, same predicate.
2. **Phase 2.5** of `run_type_checker` publishes them as `TYPE_KIND_FUNC` type ids on the
   interface's type entry (`TypeTable.iface_method_sigs`), slot-aligned with the method-name list
   registered in Phase 0.0 and pushed in lockstep with it so the two cannot desynchronize. It
   enumerates interface SYMBOLS, not `NODE_INTERFACE_DECL` nodes, and delegates every extraction to
   `interface_method_sig`.
3. `air_builder.coerce_float_arg` split into `coerce_float_arg(fn_sym, …)` →
   `coerce_float_arg_ft(fn_type, …)`; the dispatch branch calls the latter with the published FUNC
   type. **One rule, two entry points** — a static and a dynamic call cannot diverge on what a
   float argument means.

`NODE_INTERFACE_DECL` remains the sole authority; the type-table entry is derived, by one function,
from one walk. No new AIR opcode (`OP_CAST` is what the static path already emits), no change to the
box layout, vtable slots, ABI, linker or syntax.

## ⭐ The lesson that outlives the bug

**Before threading a value into a phase, check which phase already solves the same problem for the
working case.** The CORRECT column of the control matrix was the whole answer: static calls work,
and they work *without any typecheck involvement*. Implementing the fix where it was first assumed
to belong would have created a **second** argument-coercion mechanism, in a different phase, for a
problem the backend already solves — the exact "two copies of one mechanism drift apart" pattern
that produced `376af08` (duplicated interface walk), `76de988` (`lower_int_lit` vs
`lower_float_lit`), `0bf34ee` (explicit type args) and this bug. The fix that removes a divergence
must not introduce one.

Corroborating datum in the same matrix: the ONE static row that fails, `s.f64int(3)` (int literal →
f64 param), fails precisely because `coerce_float_arg`'s predicate covers f32↔f64 only. Both the
successes and the single failure of the static column are explained by the backend, never by
typecheck.

## Calibration

- Oracle `bin/t_ifacefloatarg.ax`: **30** before (first WRONG row; all 12 static controls and all 5
  already-correct dynamic rows pass), **42** after, at `-O0`/`-O1`/`-O2`.
- Bitmask variant `bin/probe5/orcmask.ax` proves each row independently rather than stopping at the
  first: **100** before = all seven coercion rows wrong (none passing by accident), **227** after =
  all seven correct.
- probe4 rows: `f2` 12→42, `f3` 14→255, `f6` 105→107, `e4` 8→42.

## Left open on purpose (parity with static calls, not new debt)

- **`s.f64int(3)` — int literal at an `f64` parameter — is broken for STATIC method calls too**
  (`bin/probe4/f4.ax` returns 253 before AND after). Same family as the open task
  [[bug-int-literal-float-type-iconst]] "STILL OPEN" (int→f64 where typecheck spreads no hint).
- **f32 RETURN through dispatch reads a stale register** when the preceding dispatch call also
  returned f32 *and took arguments*: `bin/probe5/r6e.ax` exits **101** on the OLD compiler and the
  NEW one alike (the second call's result is the first call's value). Pre-existing and untouched —
  verified by running both compilers on the identical program. It is why `bin/probe4/f1.ax` moved
  61 → 110 rather than to 42: rows R1–R5 now pass and execution reaches this *different* defect.
  Two consecutive no-arg f32 dispatch calls are fine (`r6f.ax` = 42), and slot index alone does not
  trigger it (`r6c.ax` = 42).
- Interface-typed parameters of an interface method are not boxed at the dispatch site (the
  `coerce_interface_arg` analogue).
- **E3030 does not fire at dispatch arguments** — no expected type is threaded there. Identical to
  static method calls, whose range check is gated to `NODE_IDENT` free-function callees.
