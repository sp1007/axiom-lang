---
name: self-type-return-fixed
description: "Self type: `-> Self` return on a free-function method now resolves to the self/receiver type, so method chaining on a Self-returning result works (was segfault). Frontend A==B."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

✅ **Self type `-> Self` return now works for method chaining.** Was a silent segfault.

**Symptom:** `fn inc(self: Counter) -> Self` — a SINGLE call `c.inc()` and field access `c.inc().n` worked, but calling a METHOD on a Self-returning result (`c.inc().inc()`, or 2-stmt `let d=c.inc(); d.inc()`) SEGFAULTED. Found probing the "Self type" backlog item (which was ~90% already working: Self as param, single return, field access all fine).

**Root cause:** `Self` in a free-function return type resolved to `TYPE_UNKNOWN` — the resolver (resolver.ax:907) defines `Self` only in INTERFACE_DECL scope, but AXIOM methods are FREE FUNCTIONS with a `self` param (RFC 0002 untyped-self), so there is no interface scope. `pre_infer_func_signature` inferred `fi.ret = UNKNOWN`; `method_ret_type` returns `fi.ret`, so `c.inc()` was untyped → the next `.inc()` couldn't resolve/dispatch → segfault. (Field access tolerated the unknown type; method dispatch did not.)

**Fix (frontend, typecheck.ax `pre_infer_func_signature` ~L1580):** when the first param is named `self` and the return annotation text is `"Self"`, set `ret_type = param_types[0]` (the self/receiver type) instead of inferring the (unresolved) `Self` ident. Guarded on `first_is_self` so a stray `Self` in a non-method fn isn't mis-resolved.

**Gate:** frontend-only, **A==B `39695e6b`** (no compiler/std fn returns `-> Self` — only std/io.ax uses `self: Self` as an interface PARAM, untouched — so self-codegen unchanged). Regression **127/127**. Oracle `t_selftype` (13: `c.inc().inc().add(10).inc().n` builder chain).

**Not covered:** `Self` as a PARAM type in a free fn (`fn f(self: Self)`) still unresolved (works for free-fn calls/field-access but not as a method receiver); interface `Self` params (io.ax) unchanged (no vtable dispatch yet, [[bug71-interface-dynamic-dispatch]]). Enough for builder-pattern method chaining; unblocks std-module rewrites that return Self.
