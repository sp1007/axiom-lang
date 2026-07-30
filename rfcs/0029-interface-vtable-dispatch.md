# RFC 0029 — Interface vtable dynamic dispatch

- Status: DRAFT (2026-07-19c) — investigation complete, design decisions made. User greenlit
  (2026-07-19c backlog, the real remaining chunk of "Self + vtable"). Implementation is a
  dedicated codegen+type-system session.
- Depends on: typetable.ax (interface method table + vtable registry), typecheck.ax
  (impl-matching, fat-pointer typing), air_builder.ax + x86_selector.ax (fat-pointer repr +
  indirect dispatch), resolver.ax (interface method resolution).
- Closes: BUG#71 (method call through an interface-typed value segfaults → currently rejected).
- Unblocks: std-module rewrite (std/log's `Box[LogSink]`, std/iter, etc. need `Box[Interface]`).

## 1. Motivation

AXIOM has `interface X:` declarations (parsed to NODE_INTERFACE_DECL with method-signature
children) and a distinct `TYPE_KIND_INTERFACE` (size already 16 — fat-pointer capable), but
**interfaces carry no runtime representation**: no method table, no vtable, no fat pointer.
Calling a method through an interface-TYPED value (`s: Shape; s.area()`) has no concrete
receiver at compile time → BUG#71 rejects/segfaults it. This blocks all open polymorphism:
`Box[LogSink]`, heterogeneous collections, plugin-style APIs, and the aspirational std modules
(std/log, std/iter, std/net) that are written against `Box[Interface]`.

Self-return, Self-as-param, and static/UFCS method dispatch all already work — **dynamic
dispatch through an interface value is the one missing piece.**

## 2. Current state (investigation 2026-07-19c)

- `parse_interface_decl` (parser.ax:1177) parses methods as NODE_METHOD_SIG children — the
  signatures are in the AST but nowhere else.
- `register_interface` (typetable.ax:485) stores ONLY a name_id + kind (no method list); the
  comment explicitly says "No method-signature side table … nothing else to store here."
- `TYPE_KIND_INTERFACE` entries have size 16, align 8 — already sized for a `{data, vtable}`
  fat pointer.
- Methods are free functions resolved by name+receiver-type (static). There is no indirect
  call opcode used for method dispatch today (OP_CALL takes a resolved symbol).

### ⚠️ Correction (2026-07-19b) — the "reuse RFC 0028's `.rodata` reloc" premise is FALSE
This RFC (§3c/§5/§8-step-3) assumed RFC 0028 would build a `.rodata` **code-address relocation**
that 0029 could reuse. **RFC 0028 SHIPPED as a balanced COMPARE-TREE** (`emit_bsearch_range`,
air_builder.ax:3163 — recursive OP_ICONST/OP_EQ/OP_LT/OP_BRANCH), which needs **no rodata table
and no relocation at all** (the RFC 0028 spec itself notes "the compare-tree needs no new
opcode"; there is no `OP_JUMP_TABLE` in the codebase). So there is **no 0028 prerequisite to do
first** — the vtable session must build the `.rodata` **code-address relocation from scratch**
(a static array of function-pointer entries in `.rodata`, each a code-address reloc, on BOTH
COFF and ELF). Note the existing `OP_GLOBAL_ADDR` path already emits *data*-address rip-relative
relocations (air_builder.ax ~L792); the new need is a relocation whose target is a FUNCTION
symbol/offset — closer to how `OP_FUNC_ADDR` (BUG#49 fn-pointers) resolves a code address. Start
step 3 from `OP_FUNC_ADDR`'s reloc emission, not from a (non-existent) 0028 rodata table.

## 3. Design (decisions made per §20 — safest minimal)

### 3a. Interface method table (typetable)
Add a side table: for each `TYPE_KIND_INTERFACE`, an ordered `Vec[MethodSig]` (name_id +
param types + ret type) built from the NODE_METHOD_SIG children. Method **slot index** = its
position in this ordered list — the vtable layout. Stored as a new `TypeTable.iface_methods`
registry (a single-instance TypeTable field → NO per-element struct size change, avoiding the
StructField/StructInfo size-machinery risk).

### 3b. Impl-matching (typecheck)
AXIOM uses STRUCTURAL conformance (no explicit `impl Interface for T` — matches the existing
duck-typed method model, RFC 0002). A struct `T` implements interface `I` iff for every method
in I, `T` has a method of that name with a compatible signature (self + params). Checked at the
COERCION site (assigning a `T` value to an `I`-typed slot: let/param/return/`Box[I]`). On
success, ensure a vtable for (T, I) exists; on failure, a clear diagnostic (BUG#53 convention).

### 3c. Vtable construction
For each (concrete struct T, interface I) pair that is actually coerced, emit a static vtable
in `.rodata`: an array of function pointers to T's methods in I's slot order. Deduplicated by
(T, I). Each slot is a code-address relocation (same reloc machinery RFC 0028 needs — build
0028's `.rodata` code-address relocation FIRST, then 0029 reuses it, on both COFF and ELF).

### 3d. Fat-pointer representation
An interface-typed value is a 16-byte `{ data: ptr, vtable: ptr[fn] }`. Coercing `T → I`:
`data = &t` (address of the struct; T already by-address under RFC 0001 aggregate=reference),
`vtable = &vtable_T_I`. This fits the existing 16-byte interface type slot (§2). Passing/return
uses the existing 16-byte-aggregate ABI path (like `str`).

### 3e. Dynamic dispatch codegen
`iface.m(args)` where iface : I → look up m's slot index k in I's method table; emit
`fptr = vtable[k]; call fptr(data, args...)` — the receiver is `data`. Needs an indirect-call
lowering (`OP_CALL_INDIRECT` on a reg holding the fn ptr; x86 `call rax`). Add `OP_CALL_INDIRECT`
if absent (fn-pointer calls already work — BUG#49 — so an indirect-call path likely exists;
reuse it).

## 4. Alternatives

- **Closed sum-type dispatch** (tag + `match`) — already available via sum types; gives no OPEN
  polymorphism (must know all impls), so it does not substitute for interfaces. Rejected.
- **Monomorphize per concrete type** (generics instead of dynamic dispatch) — works when the
  type is statically known, already supported; interfaces are for when it is NOT. Complementary.

## 5. Drawbacks / risk

- Largest type-system + codegen surface of the greenlit features: method table + structural
  impl-matching + vtable emission (with code-address relocations) + fat-pointer ABI + indirect
  dispatch. Monolithic — no feature value until all pieces land, so it CANNOT be shipped in
  inert increments; it needs a focused session, not autopilot ticks.
- Shares the `.rodata` code-address relocation need with RFC 0028 — do 0028's reloc kind first.
- Backend/ABI change → **B==C mandatory + full regression + -O2 acceptance + ELF (Linux) parity**.

## 6. Migration / compatibility

No change to existing static/UFCS dispatch (unaffected). Interface-typed values become usable
where they were rejected (BUG#71). No syntax change (structural conformance, no `impl` block).
`Box[I]` becomes a fat pointer. ABI: interface values join the 16-byte-aggregate convention.

## 7. Gate (before commit, when implemented)

Structural-conformance diagnostic tested (missing method → clean reject, not segfault). Dynamic
dispatch oracle: ≥2 distinct structs behind one interface param, each dispatching to its own
method (proves the vtable is per-concrete-type, not miscompiled to one impl). `Box[I]` in a Vec
(heterogeneous). B==C fixpoint; full regression; -O2-built-compiler regression; ELF/Linux smoke
(the compiler self-hosts on Linux now — verify the vtable relocations resolve there too).

## 8a. Implementation progress (2026-07-19b focused session)

- **✅ Step 1 SHIPPED** (`feat(rfc0029): step 1`, A==B `B540FB12`, 462/462): `TypeTable.iface_methods`
  (U32VecVec) + `register_interface_methods`/`interface_method_list`; each interface's `extra` =
  (iface_methods index + 1). Ordered method-name list built from NODE_METHOD_SIG children in Phase-0
  typecheck. Inert (nothing consumes it yet). `new_type_table` alloc switched to `size_of` (was
  hardcoded 120) so the added 6th Vec field can't desync the allocation.
- **Step 3 reloc scoping (read-only):** the reloc PRIMITIVES exist — `OP_FUNC_ADDR`
  (x86_selector.ax:1853) → `MACH_MOV_IMM` vreg=3 → `lea reg,[rip+func]` + `RELOC_PC32` on a **.text**
  symbol; `OP_GLOBAL_ADDR` → vreg=4 → same on a **.data** symbol. What's MISSING for a vtable: a
  **static data blob whose individual slots are function-symbol relocations** (a `.rodata`/`.data`
  array where slot k = &method_k, resolved by the linker). That is new object-file emission on BOTH
  COFF (x86_coff.ax) and ELF (elf emitter) — the architecturally-significant, B==C-critical piece.
  Start it from the globals/.data emission path (RFC 0017) extended to write a per-offset function
  relocation, mirroring how `OP_FUNC_ADDR` names a .text symbol. **⚠️ CHECKPOINT: linker/object-file
  relocation changes (§16) need user visibility before proceeding — high-risk, hard-to-reverse.**

## 8b. ⭐ DESIGN REFINEMENT (2026-07-19b) — runtime-init vtables, NO linker changes

The original §3c/§8-step-3 plan (static `.rodata` vtable with per-slot **function-symbol
relocations**) requires new COFF+ELF object-file emission — the architecturally-significant §16
risk. **This is avoidable.** RFC 0017's runtime global-init (air_builder.ax:4598, "at the top of
`main`, evaluate each global's init_node and OP_STORE it into the global's storage") already runs
arbitrary init code before user code. So build vtables at **runtime-init** instead of statically:

- Each (concrete T, interface I) pair gets a **global byte array** `vtable_T_I` of `8*N` bytes
  (N = methods in I), zero-initialized in `.data` (RFC 0017 aggregate globals — already supported).
- At main-entry init, for each method slot k: `OP_FUNC_ADDR &T.method_k` → `OP_STORE` into
  `vtable_T_I + 8*k` (via `OP_GLOBAL_ADDR vtable_T_I`). **All three are proven primitives**
  (`OP_FUNC_ADDR` BUG#49, `OP_GLOBAL_ADDR`+`OP_STORE` RFC 0017). No new opcode, **no relocation
  machinery, NO COFF/ELF changes, NO §16 linker work** — the entire feature becomes typecheck +
  air_builder, gateable at B==C without touching the linker. This SUPERSEDES §3c/§5/§8-step-3's
  static-reloc plan and removes the linker checkpoint.
- Indirect dispatch reuses the existing `OP_CALL` with a `callee_reg` (air_builder.ax:2106-2168,
  the BUG#49 fn-pointer-through-a-variable path): load `vtable` (fat-ptr word 1), load slot k,
  `OP_CALL callee_reg=slot(data, args...)`.

Cost vs static: a few stores at startup (negligible) + the vtable lives in writable `.data` not
`.rodata` (acceptable). Trade a one-time init for zero linker risk — clearly right per §20.

## 8. Implementation order (dedicated session) — REVISED per §8b

1. ✅ Interface method table in typetable (build from NODE_METHOD_SIG); inert (A==B). — DONE `B540FB12`.
2. Structural impl-matching + the missing-method diagnostic at coercion sites; inert-ish.
3. **[SUPERSEDED by §8b — no linker step]** ~~`.rodata` vtable emission + code-address reloc~~ →
   instead: **synthesize a vtable global + runtime-init** per (T,I). See the ready-to-execute plan
   below.
4. Fat-pointer T→I coercion + 16-byte-aggregate ABI wiring.
5. `iface.m()` indirect dispatch codegen (reuse the `OP_CALL callee_reg` fn-pointer path).
6. Gate as §7. Then rewrite std/log (the first real consumer) as the end-to-end proof.

## 8c. Ready-to-execute plan (runtime-init, execution-ready 2026-07-19b)

**Gate profile:** the compiler's OWN source has no dynamic interface dispatch (BUG#71 rejected it),
so EVERY new code path below is inert on the self-build → the compiler binary stays byte-identical
→ **gate = A==B** (not B==C) + a user-program oracle `t_ifacedispatch` + full regression. Same
favorable profile as the negative-match feature. Guard every new path so it only fires for a real
interface coercion/dispatch (`entry.kind == TYPE_KIND_INTERFACE`), keeping self-build inert.

Building blocks CONFIRMED to exist (2026-07-19b investigation):
- Method table + slot order: `TypeTable.iface_methods` / `interface_method_list` (step 1, shipped).
- Function address into a reg: `OP_FUNC_ADDR` (air_builder.ax:537; x86_selector.ax:1853). BUG#49.
- Global storage + runtime store: `AirGlobal` (air.ax:231) + main-entry init loop
  (air_builder.ax:4598) + `OP_GLOBAL_ADDR`/`OP_STORE`. RFC 0017.
- Indirect call: `OP_CALL` with a non-zero `callee_reg` (air_builder.ax:2106-2168). BUG#49.
- 16-byte fat-pointer value: mirror `str` (`{ptr,len}` 16B, x86_selector.ax:1077 by-address agg ABI).

Concrete steps for the focused session:
1. **Fat-pointer sizing (P1):** in `register_interface` (typetable.ax) set user-interface entries to
   `size:16, align:8` (built-ins already are). ⚠️ this ALONE changes the "pass struct to interface
   param" behavior (case c of BUG#71) — must land WITH P2 in the same gate.
2. **Coercion T→I (P2):** at call-arg lowering (and let/return) in air_builder, when the target type
   is `TYPE_KIND_INTERFACE` and the source value is a concrete struct T: build a 16-byte fat pointer
   `{ data = &struct, vtable = &vtable_T_I }` (struct is already by-address under RFC 0001, so `data`
   = the struct's address value). Ensure `vtable_T_I` exists (P3). Emit as a 16B aggregate value.
3. **Vtable synth + init (P3):** maintain a (T,I)→global registry on the module builder. First time a
   pair is coerced, push an `AirGlobal` of `size = 8*N` (zero init), and record N `OP_FUNC_ADDR`
   stores to run at main-init: for slot k, resolve T's method whose name = `iface_methods[I][k]`,
   `OP_FUNC_ADDR &T.method_k` → `OP_STORE [OP_GLOBAL_ADDR vtable_T_I + 8*k]`. Hook the existing
   main-entry init loop (air_builder.ax:4598) to also emit these vtable-init stores.
4. **Dispatch (P4):** replace the BUG#71 reject (typecheck.ax ~3495) — for `iface.m(args)` where
   `iface : I`: slot `k = index of m in interface_method_list(I)`; load `vtable = word1(iface)`
   (offset 8), `fptr = load [vtable + 8*k]`, `data = word0(iface)` (offset 0); `OP_CALL callee_reg=fptr`
   with `(data, args...)`. Structural conformance (step 2) = check every slot resolves to a real T
   method at P3; missing → clean diagnostic (BUG#53 convention).
5. **Oracle + gate:** `t_ifacedispatch` = ≥2 distinct structs behind one interface param, each
   dispatching to its own method (proves per-concrete-type vtables, not one impl). A==B + full
   regression + -O2 acceptance + ELF/Linux smoke. Then rewrite std/log's `Box[LogSink]` as proof.

---

## 9. AMENDMENT (2026-07-31): dispatch must coerce ARGUMENTS to the declared parameter types

**Status:** accepted, implemented. Extends §8's dispatch step; changes no representation.

### 9.1 Motivation

Shipped dispatch honoured the interface contract for the RETURN type only. Arguments were lowered
raw: `air_builder.lower_call_expr`'s dispatch branch did `iargs[n] = self.lower_expr(ia)` with no
coercion, so **every dispatch argument that needed a conversion instruction was passed
unconverted**. Measured (`bin/probe4/`):

| form | before |
|---|---|
| static `s.c32(1.5)` | correct |
| dynamic `i.c32(v)`, `v: f32` (no conversion needed) | correct |
| **dynamic `i.c32(1.5)`** (f64 literal → f32 param) | **0.0 delivered** |
| **dynamic `i.c32(2.0+2.0)`** (folded f64 expr) | **0.0 delivered** |
| **dynamic `i.c64(v)`**, `v: f32` → f64 param | **wrong** |
| dynamic, f32 param in slot 1 / 2 / after i64 / after f64 | wrong in every position |

"Exactly 0.0, not garbage" is the tell: the low 32 bits of the double `1.5` (`0x3FF8000000000000`)
are zero, and the callee's `movss` reads exactly those. Accept-then-miscompile, BUG#53 class, and a
violation of the contract this RFC already states.

### 9.2 Where the existing mechanism lives (and why dispatch could not reach it)

Call-argument float coercion is **not** a typecheck concern in this compiler. `typecheck.ax`'s
expected-type threading is gated to `NODE_IDENT`/`NODE_INDEX_EXPR` callees, so a *method* call
(`NODE_FIELD_EXPR` callee) already infers its arguments with `expected = TYPE_UNKNOWN` — yet static
method calls are correct, because `air_builder.coerce_float_arg` emits `OP_CAST` after reading the
parameter types off the **resolved callee symbol**.

Dynamic dispatch has no callee symbol by construction: `coerce_float_arg` returns immediately at
`fn_sym == 0`. The missing ingredient was never an accessor — it was the *signature* at the one site
that already performs this job.

### 9.3 Design

1. `typecheck.interface_method_sig` — the single walk that already resolves an interface method's
   arity and return type — additionally yields the DECLARED parameter types (optional out-param;
   `self` included at index 0 so the list lines up with a FUNC type and with air_builder's argument
   index, which counts the receiver). Same walk, same predicate: parameter types and parameter count
   can never be produced by two different traversals.
2. **Phase 2.5** of `run_type_checker` publishes those signatures on the interface's type entry as
   `TYPE_KIND_FUNC` type ids, slot-aligned with the method-name list from step 1. It enumerates
   interface SYMBOLS (not `NODE_INTERFACE_DECL` nodes) and delegates every extraction to
   `interface_method_sig`. Runs after struct/alias registration (so a struct-typed parameter
   resolves) and before inference (so every later dispatch site sees a populated table).
3. `TypeTable.iface_method_sigs` is pushed **in lockstep** with `iface_methods` inside
   `register_interface_methods`, addressed by the same `entry.extra` index, so the name list and the
   signature list cannot desynchronize.
4. `air_builder.coerce_float_arg` is split: `coerce_float_arg(fn_sym, ...)` resolves the symbol and
   delegates to `coerce_float_arg_ft(fn_type, ...)`, which holds the rule. The dispatch branch calls
   `coerce_float_arg_ft` with the interface method's published FUNC type. **One rule, two entry
   points** — a static and a dynamic call cannot diverge on what a float argument means.

No new AIR opcode (`OP_CAST` is exactly what the static path already emits), no change to the
interface box layout, vtable slot layout, calling convention, ABI, linker, or syntax.

### 9.4 Single source of truth

`NODE_INTERFACE_DECL` remains the sole authority. The type-table entry is *derived*, by exactly one
function, from exactly one walk. This is deliberate: four bugs in this compiler
(`376af08`, `76de988`, `0bf34ee`, and this one) trace to two copies of one mechanism drifting apart,
so the fix adds a consumer, never a second producer. Staleness bound: the FUNC types are resolved
once at Phase 2.5 and nothing downstream mutates an interface declaration (monomorphization clones
functions, not interfaces).

### 9.5 Alternatives considered (and rejected)

- **Thread expected types at the typecheck dispatch site only.** Rejected on evidence: it creates a
  SECOND argument-coercion mechanism in a different phase for a problem the backend already solves
  for static calls, and it still cannot fix `i.c64(v)` with `v: f32` — threading an expected type
  does not rebuild an ident's value (the codebase states this at `typecheck.ax`'s Option/Result
  width reject: *"Threading an expected type does not REBUILD an ident's value, so a coercion is
  impossible here"*). Two of the broken rows would have survived.
- **Walk `NODE_INTERFACE_DECL` from air_builder.** Rejected: a second walk of the mechanism this RFC
  is trying to keep single, plus frontend type resolution (`infer_node`) in the backend.
- **Use any implementor's method symbol at the dispatch site.** Rejected as unsound: two unrelated
  structs may declare same-named methods with different float parameter types, and conformance
  checking currently proves only arity and return type — this could emit a *wrong* conversion, which
  is worse than none.

### 9.6 Drawbacks

- One more derived table on `TypeTable` (memory: one `U32Vec` per declared interface).
- Coercion is currently f32↔f64 only, inherited verbatim from the static path. Int-literal→float and
  int narrowing at method arguments are NOT coerced — dispatch now has exactly the same coverage as
  a static call, no more (see §9.7).

### 9.7 Known gaps left open on purpose (parity with static calls, not new debt)

- `s.f64int(3)` — an int literal at an `f64` parameter — is broken for **static** method calls too
  (`bin/probe4/f4.ax` returns 253 before and after this change). It is a distinct defect in
  `coerce_float_arg`'s f32/f64-only predicate and is tracked separately.
- Interface-typed parameters of an interface method are not boxed at the dispatch site (the
  `coerce_interface_arg` analogue). Untested territory, deliberately not bundled here.
- E3030 (out-of-range integer literal) does not fire at dispatch arguments, because no expected type
  is threaded there — identical to static method calls, whose range check is gated to `NODE_IDENT`
  free-function callees.

### 9.8 Migration / compatibility

No source-level change. Programs that were silently miscompiled now compute the declared conversion.
Self-build is inert (the compiler's own source declares no interfaces), but `air_builder.ax` is
touched, so the gate is **B==C** plus full regression at default and `-O0`.
