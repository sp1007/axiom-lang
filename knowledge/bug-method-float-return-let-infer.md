---
name: bug-method-float-return-let-infer
description: "BOTH FIXED 2026-07-30 (probe-found, two distinct bugs) — (A) `let a = <in-struct method>()` returning f32/f64 infers the wrong type and yields 0, fixed in eb88586 (A==B 0148CBB3, 587/587); (B) interface dynamic dispatch resolved NO return type so any non-i64 return through an interface was wrong, fixed in 376af08 (A==B 4F359C9B, 584/584)."
metadata:
  node_type: memory
  type: project
---

# TWO silent miscompiles around method RETURN TYPE resolution — BOTH NOW FIXED

## ✅ Status
- **(B) interface dynamic dispatch** — FIXED `376af08`, `A==B 4F359C9B`, 584/584. Deleted the
  duplicated AST walk and delegated to `interface_method_sig`, which already had the correct
  predicate. Oracle `bin/t_ifaceretnoni64.ax` (calibrated: 13 pre-fix, 42 after).
- **(A) float-returning in-struct method** — FIXED `eb88586`, `A==B 0148CBB3`, 587/587. Root: the
  method-return recovery (`typecheck.ax` ~L4979) restored the result type only for STRUCTURED
  returns, on the premise that "an untyped copy of a scalar already works via the register return".
  **False for floats** — a float return arrives in **XMM0, not RAX**, so an untyped copy defaulting
  to an integer reads the wrong register. Fix = add `TYPE_F32`/`TYPE_F64` to that condition.
  **Floats were the THIRD class to fall out of that same premise, after struct and str/bytes.**
  Oracle `bin/t_methfloatret.ax` (calibrated: 1 pre-fix, 42 after).

⚠️ **Oracle-writing trap, caught by calibration.** The first `t_methfloatret` checked the f64 case
via `near(a, 9.0)` and **PASSED on the broken compiler** — passing the binding to an f64-typed
parameter **coerces it and masks the defect entirely**. Only a DIRECT comparison against float
literals observes the binding's own type. Calibration exposed it: the broken compiler returned 2,
sailing through the f64 case and failing at the f32 one.

The investigation record below is retained because the control matrix is the diagnosis.

---


Found by probing 2026-07-30 while checking whether the `interface_method_ret_type` follow-up noted
in [[bug-iface-conformance-signature-not-checked]] was really harmless. **It is not.** And the probe
turned up a second, independent bug next to it.

Both are **accept-then-miscompile** (BUG#53 class): no diagnostic, wrong answer.

## BUG A — `let` binding a FLOAT-returning IN-STRUCT method infers the wrong type

    struct Sq:
        s: f64
        fn area(self: Sq) -> f64:
            return self.s * self.s
    fn main() -> i64:
        let q = Sq(s: 3.0)
        let a = q.area()          // <-- a is NOT f64
        if a > 8.99 and a < 9.01:
            return 42
        return (a * 10.0) as i64  // returns 0

**Returns 0, expected 42.** Repro `bin/probe/zf_meth.ax`. f32 form returns 7: `bin/probe/zf_f32.ax`.

The trigger is a THREE-WAY combination, and each control isolates one leg:

| probe | method declared | result usage | return | result |
|---|---|---|---|---|
| `zf_free` | free fn | `let` | f64 | 42 ✅ |
| `zf_ufcs` | free fn with `self: Sq`, called UFCS `q.area()` | `let` | f64 | 42 ✅ |
| `zf_nolet` | **in-struct body** | used INLINE in the compare | f64 | 42 ✅ |
| `zf_mstr` | in-struct body | `let` | **str** | 6 ✅ |
| `zf_mi64` | in-struct body | `let` | **i64** | 36 ✅ |
| **`zf_meth`** | **in-struct body** | **`let`** | **f64** | **0 ❌** |
| **`zf_f32`** | **in-struct body** | **`let`** | **f32** | **7 ❌** |

⭐ **It is INFERENCE, proven, not codegen**: `let a: f64 = q.area()` (explicit annotation) returns 42
correctly — `bin/probe/zf_annot.ax`. So the value is produced fine; only the binding's inferred type
is wrong. Almost certainly defaulting to an integer type, which reads the float return from the
wrong register/width.

⚠️ Note how narrow the escape is: **free functions are fine and UFCS is fine**, so the entire
existing float test suite (`t_interpolation`, `t_colorhsl`, `t_quatrot` — all free functions) misses
this. And in-struct methods returning `i64`/`str` are fine, so the interface/method oracles miss it
too. **No test in the suite binds a float-returning in-struct method to a `let`.**

## BUG B — interface dynamic dispatch resolves NO return type

    interface Shape:
        fn area(self: Self) -> f64
    struct Sq:
        s: f64
        fn area(self: Sq) -> f64:
            return self.s * self.s
    fn useit(sh: Shape) -> i64:
        let a = sh.area()
        if a > 8.99 and a < 9.01:
            return 42
        return (a * 10.0) as i64

**Returns 80, expected 42.** Repro `bin/probe/zi_f64.ax`. The `str` form returns **115** instead of
6 (`bin/probe/zi_str.ax`). `i64` through an interface is correct (36, `bin/probe/zi_i64.ax`) — which
is why every existing interface oracle passes: **they all return `i64`.**

**Root cause is exact and already written down.** `interface_method_ret_type`
(`typecheck.ax:1664`), the sole resolver of a dynamic-dispatch call's result type
(called at `typecheck.ax:4332`), matches the interface declaration with:

    if dn.kind == NODE_INTERFACE_DECL and dn.payload == iname:      // iname = type entry name_id

but **the resolver rewrites `NODE_INTERFACE_DECL.payload` from the interned name to the SYMBOL
INDEX** (the Phase-0 registration walk at `typecheck.ax:2544` reads `node.payload` as `sym_idx`).
So the predicate essentially never matches and the function always returns 0.

⭐⭐ **This exact defect was identified and dodged during `f614c11`** (see
[[bug-iface-conformance-signature-not-checked]]), which fixed the sibling function by matching
`symbols[dn.payload].type_id == iface_type` instead. That session recorded
`interface_method_ret_type` as "worth fixing in a follow-up, but out of scope since its caller
already copes." **The caller does NOT cope** — the comment at `typecheck.ax:4330` claims downstream
recovery handles a UNKNOWN result, and it evidently recovers to the WRONG type (i64) rather than the
declared one. That assumption was never tested against a non-i64 return.

**Fix**: mirror the validated predicate — `symbols[dn.payload].type_id == iface_type`. Frontend-only
⇒ `A==B` gate. ⚠️ Watch for the `Self`/generic return positions the sibling function had to skip,
and for the cross-tree hazard (`symbol_trees[iface_sym_idx]` may differ from `self.tree`) noted at
`typecheck.ax:1733`.

## Meta-lesson
**"Its caller copes with the degenerate value" is a claim about the caller, not a proof.** Two
sessions carried this function as harmless because its one caller tolerated a 0 — nobody probed what
the caller then DID with the 0. Costs nothing to check: one program per return type.

Related: [[bug-f32-compare-float-literal]] and [[bug-negative-literal-compare-o0]] are the same
family — a type/coercion rule present on one path and missing on its sibling path.
