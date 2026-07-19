---
name: rfc0029-vtable-progress
description: "RFC 0029 interface vtable dynamic dispatch — ✅ SHIPPED af26946 (closes BUG#71). `fn f(s: Shape): s.area()` dispatches through a per-value inline vtable: interface value = 8-byte ptr to heap box {data, m0..m(N-1)}, built at T→I coercion (OP_ALLOC+OP_SET_FIELD+OP_FUNC_ADDR), dispatched via OP_GET_FIELD+OP_CALL callee_reg. NO linker/reloc changes. A==B c8910d77, 463/463, oracle t_ifacedispatch(37), multi-method+args(51), O0-O3 clean. Optional follow-ups: shared static vtables, conformance diagnostic, std/log Box[LogSink]."
metadata:
  type: project
---

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
**Optional follow-ups (NOT shipped):** dedup vtables into shared statics (per-value alloc is
slightly wasteful), structural-conformance diagnostic (missing method → clean reject vs
find_struct_method_sym returning 0), std/log `Box[LogSink]` rewrite as the real consumer.

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
