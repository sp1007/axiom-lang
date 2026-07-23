---
name: bug-iface-conformance-signature-not-checked
description: OPEN — interface conformance checks method NAME-presence only, not signature; a same-named method with wrong arity/return-type is accepted and miscompiles (arg dropped, or garbage read for extra params)
metadata:
  type: project
---

**Status: OPEN, found by probing 2026-07-23.** Interface conformance
(`check_iface_conformance` -> `interface_missing_method`, `typecheck.ax:1662/1617`) verifies
that the struct has a method of the right NAME, and nothing else. Param count, param types, and
return type are not checked. A struct with a same-named but signature-incompatible method is
accepted as conforming, and the vtable then dispatches through the mismatched signature.

Three severities, all accept-then-miscompile (they should REJECT):

| case | struct method vs interface | result |
|---|---|---|
| **F2** wrong return type | iface `area()->i64`, struct `area()->bool` | accepted; dispatches to the bool fn |
| **F3** fewer params | iface `scale(self,f)`, struct `scale(self)` | accepted; the argument is silently DROPPED (`scale(3)` ignores 3) |
| **F3b** more params | iface `scale(self)`, struct `scale(self,extra)` | accepted; `extra` is read as GARBAGE — undefined, nondeterministic |

F3b is the worst: calling a 2-param impl through a 1-param interface slot reads an uninitialized
argument register, so the result is undefined (observed 7 = 5 + whatever was in the register).
F3 is a wrong-answer (dropped arg). F2 is a type-safety hole (wrong return type flows on).

**Root cause is exact and narrow.** `interface_missing_method(struct_type, iface_type)` returns
the name-id of the first interface method the struct lacks BY NAME, or 0. It never compares the
found method's signature to the interface's. So "present with any signature" == "conforms".

**Fix direction (dedicated session — reject-path change):** extend the conformance check to
compare, for each interface method, the struct method's param count, param types, and return
type against the interface declaration, and emit a diagnostic on mismatch (analogous to the
existing "does not implement interface method 'x'" error). Gate carefully: `typecheck.ax`
records several over-rejection attempts that broke the self-build, and the compiler's own
interface/method usage must still pass. The signature data is available — `method_ret_type`
(`typecheck.ax:1685`) already walks a method's type entry for its return type, and the same
`f_entry` carries params.

Not a coercion-site bug (that family is [[bug-iface-variant-payload-no-vtable-box]] /
[[probe-iface-store-coercion-sites-complete]]); this is the ORTHOGONAL conformance-checking
axis. RFC 0029's "structural-conformance diagnostic" (`7858aa5`) shipped the name-presence half;
the signature half was never done.

## Implementation recipe (read-only scouting 2026-07-23)

The fix is more tractable than it first looks, once two subtleties are named:

1. **Skip the `self`/receiver param.** The interface declares `self: Self`, the struct declares
   `self: Sq` — the receiver types DIFFER and that is correct conformance. Comparing only the
   NON-self params plus the return type catches all three failing cases (F2 return, F3 fewer, F3b
   more) and sidesteps the Self-vs-concrete tangle entirely.

2. **Type comparison — no general `types_equal` exists.** But `is_method_compatible`
   (`typecheck.ax:1389`), which resolves method-to-receiver, compares unwrapped type_ids
   directly (`unwrapped_param == unwrapped_rec`) and works — evidence that the type table
   canonicalizes enough (`register_option`/`register_result` dedup by content, primitives are
   singletons) that raw type_id equality is usable for a first cut. If a false-reject on
   "equivalent types, different ids" turns up in the gate, THAT is the signal to build a
   structural predicate; do not build one pre-emptively.

**Raw material is in hand:** the method`s full param list and return live in the `FuncInfo`
(`fi = self.types.funcs.data[f_entry.extra]`, `fi.params`), reached exactly as
`is_method_compatible` does. So the recipe is: in `interface_missing_method` (or a sibling),
for each interface method, find the struct`s same-named method, and compare
`fi.params[1..]` + return by unwrapped type_id; diagnose on mismatch.

**Gate discipline:** self-build fixpoint FIRST — if the compiler`s own interface impls trip a
false reject, that is the over-rejection this file`s neighbours warn about; revert-on-red, do
not iterate more than once before declaring it needs the structural predicate.
