---
name: bug-option-arith-miscompile-open
description: "FIXED 313cb51: arithmetic on un-unwrapped Option/Result (v.get(0)+v.get(1)) silently added box pointers -> garbage. Root: Option/Result desugar to SUM (kind 6) named Option/Result, missed by kind-11/8/struct checks. Fix = base-name test tc_is_opt_res in the arith reject."
metadata:
  node_type: memory
  type: project
  originSessionId: 549fa823-f129-4f77-b9fb-7bcc94df6352
---

# ✅ FIXED `313cb51` (A==B `597F7099`, 190/190) — arithmetic on Option/Result now REJECTS

Found by feature-combo probe 2026-07-12 (batch after BUG#93b fix `9f89b64`).

## THE FIX (after two wrong looks — see history below)
**Root cause:** `Option[T]`/`Result[T,E]` desugar to a **SUM (type kind 6)** whose
mangled base name is "Option"/"Result" (e.g. `_AX_std_Option__i64`). At the
NODE_BINARY_EXPR arith branch (typecheck.ax ~2648) the Option operand `infer_node`s
to **type-id kind 6** — NOT kind OPTION(11), GENERIC_INST(8), or STRUCT — which is
exactly why the first attempt (checking those kinds) never fired. (Confirmed via
ZIDENT instrument: `let a: Option[i64]` → tid=415 kind=6 name=`_AX_std_Option__i64`.)
**Fix:** new `tc_is_opt_res(t)` = base-name "Option"/"Result" over kinds SUM/OPTION/
RESULT/GENERIC_INST/STRUCT; the arith `else`-branch rejects if either operand is
opt/res. Base-name (not kind/type-id) is what makes it self-host safe: the
`hash_key[K]()%cap` generic-call mis-inference is a GENERIC_INST named "hash_key"
(base ≠ Option/Result) → not flagged → A==B holds. Oracle `t_optarithreject` (reject);
`t_optunwrapadd(112)` proves unwrapped arithmetic still works.

## ORIGINAL (superseded — kept for the debugging lesson)

## Repro (BUG#53 class silent-accept-then-miscompile)
```
import std.collections
fn main() -> i64:
    mut a = new_vec[i64]()
    a.push(5 as i64)
    a.push(7 as i64)
    let x = a.get(0)      // Option[i64], NOT unwrapped
    let y = a.get(1)
    return x + y          // adds two Option BOX POINTERS -> garbage (ran 112), no reject
```
`Vec.get -> Option[T]`. `x + y` should REJECT (no arithmetic on Option) but builds
and returns garbage. Also reproduces with a dedicated `let a: Option[i64] = Some(5)` +
`a + b`. A user forgetting `.unwrap()` gets silent garbage. Low severity (obviously-
wrong code), but it is a real silent miscompile.

## Attempted fix (REVERTED — did not work)
Added a reject in `typecheck.ax` at the NODE_BINARY_EXPR arithmetic `else`-branch
(~2647, the `op != comparison/and-or` case): if either operand type is Option/Result,
emit a diagnostic. **It did not fire.** Instrumenting (XARITH print of t1/t2 kind+name
right where `mut t1 := infer_node(lhs, TYPE_UNKNOWN)` runs) showed the operands there
NEVER present as Option — for the typed `Option+Option`, the `Vec.get` version, AND a
plain `i64+i64`, the operand types read the SAME (type-ids 0/4/8, kind 0, empty name).
So `return a + b` on Options is **not** reaching that arith branch with Option-typed
operands — the operand types are already resolved to something primitive/unknown there
(likely the return-statement's expected-type inference path handles `a+b` before/around
this branch, or the operand node_types are read elsewhere). The comparison-operator
reject right above (~2619, struct `==` without overload) DOES work via the same
infer_node, so it is specific to how Option operands flow here.

Two traps hit while attempting (documented so they aren't re-hit):
1. Restricting to kinds OPTION/RESULT (11/12) misses it — `Vec.get(0)` result is a
   GENERIC_INST (kind 8) named "Option", NOT kind 11. (A base-name "Option"/"Result"
   check is needed, which also correctly EXCLUDES the `hash_key[K](k) % cap` generic-
   call mis-inference — that generic-inst is named after the FN "hash_key", so a
   struct/generic-inst arith reject keyed on base-name Option/Result is self-host safe.)
2. A broad struct/sum/generic-inst arith reject FALSE-rejects the self-host: `hash_key
   [K](key) % (cap as u64)` — the explicit-generic-call `f[T](..)` result is mis-inferred
   as GENERIC_INST named "hash_key" (kind 8) instead of its u64 return type (a separate
   latent inference quirk — the value is really u64 and lowers fine).

## SHARPENED ROOT CAUSE (2026-07-12, second look — the earlier "operands don't reach the branch" was WRONG)
Decisive ZARITH-count experiment (print at typecheck.ax:2648 arith `else`-branch,
count per program): `fn main: return 0` → 706 prints (all stdlib); `return a+b` on
i64 → 707; `return a+b` on **Option[i64]** → 707. So main's `a+b` **DOES reach the
arith branch** (707−706=1) in BOTH — and for the Option program the operand infers to
**type-id 4 (i64), kind 0, empty name**, NOT Option. **So the Option WRAPPER TYPE is
lost at the expression site: an Option-typed local (annotated `let a: Option[i64]` OR
inferred `let x = v.get(0)`) reads as its PAYLOAD type (i64) in `a+b`.** The register
holds the Option BOX POINTER, added as i64 → garbage (112). This is NOT a missing
reject — a reject keyed on operand type can't fire because the type already says i64.
It is a **type-propagation bug**: `infer_node(ident)` (typecheck.ax:3404 returns
`symbols[sym].type_id`) yields i64 for an Option local at the arith operand site,
while `match x`/`x.unwrap()` still work (they use the initializer/value, not the
symbol type) — so symbols[x].type_id (or how the arith path reads it) drops the
Option wrapper. Fixing = making the Option local's type stick through to expression
operands, a SELF-HOST-SENSITIVE type-system change (Option locals are everywhere in
the compiler), NOT a minimal reject. Deferred — needs a dedicated careful session +
A==B gating; low severity (obviously-wrong code) so low urgency.

## Fix direction (dedicated session)
Trace WHY `symbols[x].type_id` / the arith operand type for an Option local is i64 not
Option[i64] — start at the `let`-binding type stamp (does `let x = v.get(0)` record
Option or unwrap to payload?) and `infer_node` NODE_IDENT (typecheck.ax:3397+). If the
symbol genuinely holds Option, find where the arith path re-reads it as payload. Only
then can a base-name Option/Result reject (or correct typing) be added. Guard any
change with `hash_key[K]()%cap` (generic-call mis-inference, kind-8 named after the fn
— must NOT be touched) and full A==B. The comparison-reject at typecheck ~2619 is the
reject template once the type is correct. Related: [[bug93-contains-method-resolution-open]]
(BUG#53 reject-gate family), [[inline-match-arm-unsupported]].
Positive coverage locked in: `t_optunwrapadd(112)` (the CORRECT unwrapped counterpart).
