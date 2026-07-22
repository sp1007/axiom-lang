---
name: probe-iface-store-coercion-sites-complete
description: All REACHABLE struct->interface store/assignment coercion sites are covered after the array-assign fix; deref-assign is a latent gap guarded by a parse limitation
metadata:
  type: project
---

**Probed 2026-07-23 after the array/Vec element-assign fix (`BE15094A`).** Swept the
struct->interface coercion sites on the STORE / assignment side (the write direction, as
opposed to the construction side tracked in [[bug-iface-variant-payload-no-vtable-box]]).

Reachable store positions, all now CORRECT:

| position | status |
|---|---|
| `obj.field = Sq(..)` (struct field) | ok (pre-existing) |
| `arr[i] = Sq(..)` (`[Shape;N]`) | fixed `BE15094A` |
| `v[i] = Sq(..)` (`Vec[Shape]`) | fixed `BE15094A` (same branch) |
| `outer.inner.s = Sq(..)` (nested field) | ok |
| `holder.arr[i] = Sq(..)` (array field of struct) | ok (via the array fix) |
| `m.insert(k, Sq(..))` then re-insert (HashMap value) | ok |

So the reachable store frontier is complete. No new miscompile here.

**One latent site, currently unreachable:** the `NODE_DEREF_EXPR` branch of `lower_assign`
(`air_builder.ax:~4516`) computes `type_id` as the pointee type but does NOT call
`coerce_struct_to_interface`. So `*p = Sq(..)` where `p: ptr[Shape]` would store a raw struct —
the same bug the array branch had. It does not bite because **pointer-to-interface is not
supported syntactically**: `*p = x` fails to parse (`expected expression`), and even `&s` /
`(*p).area()` on an interface local produces multiple type errors. The coercion site is dead
code behind that limitation.

If pointer-to-interface is ever implemented, the deref-assign branch needs the same one-line
`coerce_struct_to_interface(type_id, rhs_idx, val_reg)` the array branch got. Nothing to do
until then; recorded so it is not a surprise.

The remaining OPEN interface-coercion gap is Option/Result CONSTRUCTION (`Some(Sq(..))`), which
is the one site whose fix does not have the interface target locally in hand — see the variant
bug note.
