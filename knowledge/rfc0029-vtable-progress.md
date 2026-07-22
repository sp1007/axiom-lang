---
name: rfc0029-vtable-progress
description: "RFC 0029 interface vtable dynamic dispatch — ✅ FEATURE-COMPLETE 770dfc3. Dispatch shipped af26946; T→I coercion now fires at EVERY value site (call-arg/let/return/struct-field/assign/array-elem) + conformance rejects at let/return/field-init. Interface usable as a VALUE type (Box[Interface] shape). Shared static vtables = WON'T-DO (quantified bad trade). A==B 2FC562C7, 474/474."
metadata:
  type: project
---

# ✅ FEATURE-COMPLETE 770dfc3 (2026-07-22) — coercion family closed

`af26946` shipped dispatch but boxed T→I **only at call-arg**. Every other site stored the
raw struct and dispatched on its bytes → SIGSEGV. `770dfc3` closes the family, making an
interface usable as a **value type** (the `Box[Interface]` shape `std/log`'s `Logger.sink`
needs):

| Site | Before | After |
|---|---|---|
| call-arg `f(Square(..))` | ✅ worked | ✅ (delegates to shared helper) |
| `let s: Shape = Square(..)` | SIGSEGV | ✅ 36 |
| `return <struct>` from `-> Iface` | SIGSEGV | ✅ 48 |
| interface-typed struct FIELD | SIGSEGV | ✅ 49 |
| `s = <struct>` reassign | SIGSEGV | ✅ 75 (t_ifaceassign) |
| `[Iface; N]` array literal | SIGSEGV | ✅ 43 (t_ifacearray) |
| `Vec[Iface].push` | ✅ (call-arg path) | ✅ 43 |
| module-level `mut g: Shape` | clean REJECT | unchanged (RFC 0017 storage gap, NOT a miscompile) |

**air_builder:** `coerce_struct_to_interface(target, src_node, reg)` = the one general helper;
wired at var-decl, return (`self.fb.ret_type`), ctor field-init (via `coerce_field_to_interface`),
simple-var assign, field assign, array-literal elements.
**typecheck:** `check_iface_conformance(target, src_node)` wired at let / return / ctor-field-init
(call-arg keeps its inline check) — a non-implementing struct is a clean REJECT instead of a
null-vtable-slot dispatch (BUG#53 class). Oracles t_ifacenoconf / t_ifacefieldconf.

## ❌ Shared static vtables — WON'T-DO (not merely "deferred")
Quantified: current per-value box `{data, m0..m(N-1)}` costs 1 alloc + (N+1) stores per
coercion. A shared static `{m0..m(N-1)}` global + `{data, vtable_ptr}` box costs 1 alloc +
2 stores. **For N=1 (the common interface — Shape) that saves ZERO; for N=2 (LogSink) one
store.** Against that it ADDS: (a) a third load on EVERY dispatch (extra indirection),
(b) a module-level (T,I)→global vtable registry = global mutable state (CLAUDE.md §3),
(c) main-entry init sequencing (RFC 0017 machinery) that escalates the gate A==B → B==C.
Trading per-dispatch runtime cost + global state for ~zero coercion-time savings is a bad
trade at realistic interface sizes. Revisit ONLY if profiling shows large-N interfaces
dominating. This closes the follow-up rather than leaving it open.

## Remaining (NOT part of RFC 0029)
`std/log` `Box[LogSink]` rewrite is blocked on the **aspirational-dialect rewrite** milestone,
not on vtables: `std/log.ax` uses `enum {}` braces, `impl X {}`, `match self { A => .. }` —
none of which parse in the real grammar. RFC 0029's consumer story is proven instead by
**t_ifaceconsumer(46)**, a real-grammar Logger-with-a-sink-field exercising let + return +
field-init + call-arg coercion + polymorphic dispatch through a field. The std module rewrite
stays its own backlog item.

# ✅ SHIPPED af26946 (2026-07-19b) — closes BUG#71

Dynamic dispatch through an interface-typed value WORKS. Actual impl matched the §8c plan
but with an even simpler representation than the runtime-init global vtables: **per-value
INLINE vtable** — the interface value is an 8-byte pointer to a heap box
`{data, m0..m(N-1)}` where the method code-addresses are stored INLINE at the coercion site
(OP_FUNC_ADDR &T.method_k → OP_SET_FIELD). No global vtable, no runtime-init, no synthesized
global symbols — even simpler than planned. Coercion = coerce_interface_arg (wired beside
coerce_float_arg in the call-arg loop). Dispatch = interception in lower_call_expr reading
data (field 0) + method ptr (field slot+1) then OP_CALL callee_reg. typecheck:
interface_method_ret_type resolves the result type from the interface method contract; the
BUG#71 reject is gone. interface size set 8/8; box layout = ensure_iface_box_type (synthesized
struct reusing step-1 iface_methods slot order). Gate A==B (inert on self-build), 463/463.
**Follow-ups:** ✅ structural-conformance diagnostic SHIPPED `7858aa5` (struct missing an
interface method → clean REJECT instead of null-pointer dispatch; struct_has_method +
interface_missing_method in typecheck; oracle t_ifacenoconf; A==B 6bf63abd, 464/464). STILL
optional: dedup vtables into shared statics (per-value alloc is slightly wasteful); std/log
`Box[LogSink]` rewrite as the real consumer; conformance at let/return coercion sites (only
call-arg covered so far — the demonstrated hole).

---

# RFC 0029 — interface vtable dynamic dispatch (implementation history)

Greenlit feature (BUG#71's real fix). Closes: method call through an interface-typed value
(`fn f(s: Shape): s.area()`), currently a clean reject. Unblocks `Box[Interface]` (std/log, iter, net).

## ✅ Shipped this session
- **Step 1 — interface method table** (`B540FB12`, A==B, 462/462): `TypeTable.iface_methods`
  (U32VecVec) + `register_interface_methods` / `interface_method_list`; each TYPE_KIND_INTERFACE
  entry's `extra` = (iface_methods index + 1), 0 = none. Ordered method-name list = vtable slot
  layout, built from NODE_METHOD_SIG children in Phase-0 typecheck (typecheck.ax ~L2061). Inert.
  `new_type_table` alloc switched hardcoded 120 → `size_of[TypeTable]` (6th Vec field added safely).

## ⭐ Design breakthrough (removes the linker risk)
Original plan: static `.rodata` vtable with per-slot function-symbol relocations = new COFF+ELF
object-file emission (§16, architecturally significant, the reason this was a "checkpoint"). **This
is avoidable.** Build vtables at **runtime-init** instead, reusing only PROVEN primitives:
- vtable_T_I = a global `8*N`-byte array (zero init, RFC 0017 aggregate global).
- at main-entry init (air_builder.ax:4598 loop), per slot k: `OP_FUNC_ADDR &T.method_k` (BUG#49) →
  `OP_STORE` into `OP_GLOBAL_ADDR vtable_T_I + 8*k` (RFC 0017).
- dispatch: `OP_CALL callee_reg=fptr` (the BUG#49 fn-ptr-through-variable path, air_builder.ax:2106).
- interface value = 16-byte fat pointer `{data, vtable}`, mirror `str`'s 16B by-address agg ABI.
**No linker/COFF/ELF/reloc changes.** Feature = typecheck + air_builder only.

## Gate profile = A==B (favorable)
The compiler's own source has NO dynamic interface dispatch (BUG#71 rejected it), so every new path
is inert on the self-build → compiler binary byte-identical → gate is **A==B** (like negative-match),
NOT B==C. Guard each new path on `entry.kind == TYPE_KIND_INTERFACE` to keep self-build inert. Prove
correctness with a user-program oracle `t_ifacedispatch` (≥2 structs behind one interface param, each
dispatching its own method) + full regression + -O2 + ELF smoke.

## Remaining P1-P4 (execution-ready — full detail in rfcs/0029 §8c)
1. **P1 sizing:** register_interface (typetable.ax:490) → user interfaces `size:16, align:8`. Must
   land WITH P2 (alone it breaks BUG#71 case-c "pass struct to interface param").
2. **P2 coercion T→I:** at call-arg (and let/return) lowering, target kind INTERFACE + source struct
   → build 16B `{data=&struct, vtable=&vtable_T_I}`. struct already by-address (RFC 0001).
3. **P3 vtable synth+init:** module-builder (T,I)→global registry; first coercion pushes an
   AirGlobal(size=8*N) + N init stores hooked into the main-entry init loop.
4. **P4 dispatch:** replace BUG#71 reject (typecheck.ax ~L3495). typecheck: type `iface.m()` by
   finding m in the interface's NODE_INTERFACE_DECL AST (read its return type_expr — no need to store
   sigs in the table, the AST is reachable via the interface symbol). air_builder: slot k =
   `interface_method_list(I)` index of m; `vtable=word1(iface)`, `fptr=[vtable+8k]`, `data=word0`,
   `OP_CALL callee_reg=fptr (data, args...)`.

## Building blocks CONFIRMED (2026-07-19b investigation — don't re-investigate)
- OP_FUNC_ADDR: air_builder.ax:537, x86_selector.ax:1853 (lea rip + RELOC_PC32).
- OP_GLOBAL_ADDR/OP_STORE + main-entry global-init loop: air_builder.ax:4598.
- indirect OP_CALL callee_reg: air_builder.ax:2106-2168.
- 16B fat-ptr agg ABI: mirror str, x86_selector.ax:1077.
- AirGlobal struct: air.ax:231. TypeTable side-tables pattern: typetable.ax.

Related: [[bug71-interface-dynamic-dispatch]] (the reject this closes), [[self-type-return-fixed]].
