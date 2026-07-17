---
name: bug-lambda-generic-fntype-infer-open
description: "✅ FIXED 2cc67ed: inline lambda → generic fn/method (type-param inferred from fn(T)->U param) now monomorphizes. Was: unresolved external ax_<name>. 2-part fix in typecheck (NODE_CLOSURE_EXPR in infer_node + NODE_FUNC_TYPE in infer_generic_type_args)."
metadata: 
  node_type: memory
  type: project
  originSessionId: 514f834a-ce6c-4fa8-8b55-78462a04139e
---

# 🔧 Generic type-param inferred from fn-typed param + inline lambda → unresolved external

Found probing (2026-07-13) closures × generics.

## Symptom
```
fn apply2[T, U](x: T, f: fn(T) -> U) -> U:
    return f(x)
fn main() -> i64:
    return apply2(20, |x: i64| -> i64 x + 22)   // Linker: Unresolved external 'ax_apply2'
```
Also `Option[i64].map(|x:i64| -> i64 x+22)` → `Unresolved external 'ax_map'`.

## Bisection (WORKS vs FAILS)
- ✅ named fn arg: `apply2(20, add22)` / `o.map(add22)` = 42.
- ✅ bound lambda: `let f: fn(i64)->i64 = |..|..; o.map(f)` = 42.
- ✅ `filter` (single type-param `[T]`, no `U`) with inline lambda = 42.
- ❌ ONLY: inline lambda passed where the callee has a SECOND type-param `U` inferred
  from the fn param's RETURN type (`map[T,U]`, `apply2[T,U]`). Not method-specific
  (free fn fails too).

## Root cause — TWO parts (verified 2026-07-13 by implementing part 2 alone; insufficient)
1. **`infer_node` (typecheck.ax:2396) has NO `NODE_CLOSURE_EXPR` (kind 35) branch**
   (grep: zero refs to NODE_CLOSURE_EXPR in typecheck.ax). The func_type registration
   at 2218-2221 is inside `pre_infer_func_signature` — fn DECLARATIONS, not closures.
   So `infer_node(inline-closure-arg)` at the generic-call site (typecheck.ax:3120)
   returns **UNKNOWN(0)**. ⇒ `arg_types[i]=0` for the lambda; the defer gate
   (3174-3182: `has_unresolved` + a generic/unknown arg) fires ⇒ template NEVER
   instantiated ⇒ call lowers to unmangled `ax_<name>` ⇒ unresolved external. A NAMED
   fn arg yields a concrete fn type ⇒ no defer ⇒ instantiates (U i32-defaulted but the
   8-byte return passes through fine).
2. **`infer_generic_type_args` (typecheck.ax:577) has NO `NODE_FUNC_TYPE` branch**
   (kind 50) — even once the closure IS concretely typed, `U` inside `fn(T)->U` can't
   be structurally inferred.

## Fix (frontend; needs BOTH; A==B gate)
- Part 1 (the blocker): add `NODE_CLOSURE_EXPR` handling to `infer_node` — compute the
  closure's fn type from its explicit param annotations + return annotation (mirror
  pre_infer_func_signature's param/ret collection), `register_function`, set_node_type,
  return it. **This is core closure typing — verify it doesn't conflict with air_builder
  lambda-lifting and doesn't perturb self-compile (A==B).** Higher risk → dedicated pass.
- Part 2 (ready): NODE_FUNC_TYPE branch in infer_generic_type_args — normalize arg_type
  ptr/ref → if TYPE_KIND_FUNC read `funcs.data[extra]`={params,ret}; recurse each param
  child ↔ fi.params[i]; ret child (last when flags&1) ↔ fi.ret. (Children of a
  NODE_FUNC_TYPE = param type nodes then, if flags&1, the ret type node.) I wrote+tested
  this alone on 2026-07-13: compiled clean, c3f/c3i still 42, but c3h/c3b STILL
  unresolved (because part 1 leaves arg_type=0) ⇒ reverted; land it together with part 1.

## Status: ✅ FIXED `2cc67ed` (2026-07-13, frontend, A==B `B2311928…`, 257/257).
Both parts landed together in typecheck.ax:
- Part 1: `infer_node` NODE_CLOSURE_EXPR → `result_type = symbols[node.payload].type_id`
  (the lifted closure fn's signature, registered by the pre_infer pre-pass).
- Part 2: `infer_generic_type_args` NODE_FUNC_TYPE branch (normalize ptr/ref → TYPE_KIND_FUNC
  → recurse param children ↔ fi.params[i], ret child (last when flags&1) ↔ fi.ret).
Oracles: bin/t_lambdagen.ax (free generic `apply2[T,U]`), bin/t_lambdamap.ax (`Option.map`).
Both were `Unresolved external` before; named-fn / bound-lambda always worked.
