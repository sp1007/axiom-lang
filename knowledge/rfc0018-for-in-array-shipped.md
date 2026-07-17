---
name: rfc0018-for-in-array-shipped
description: "RFC 0018 P1 SHIPPED — `for x in <fixed-array>` element iteration (typecheck binds loop var to elem type; lower_for desugars to indexed OP_INDEX loop). A==B fixpoint D6DBA89D, 121/121."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

✅ `96dd586` — RFC 0018 P1: `for x in <fixed-array>` element iteration SHIPPED.

Closes the top OPEN feature from [[backlog-open-items]] ("for x in <collection>") for the **fixed-array** case. Before: typecheck REJECTED array iteration (raw `lower_for` stub treated the array address as a numeric bound → hang; loop var was a bare 0-based counter that never fetched elements).

**Fix (2 files):**
- `typecheck.ax` NODE_FOR_STMT: a `TYPE_KIND_ARRAY` iteree binds the loop var's `type_id` to the array element type (`entry.extra`) instead of default i32; no longer emits the reject diagnostic. STRUCT/SUM/GENERIC_INST (Vec/HashMap) STILL rejected = **P2**.
- `air_builder.ax` `lower_for`: detect array iteree via `self.mb.node_types[range_expr]` → `TYPE_KIND_ARRAY` (`name_id`=compile-time len, `extra`=elem type). Uses an **internal i32 index counter** (iter_reg, NOT bound to the loop-var symbol); at body top emits `OP_INDEX(base, i)` with elem type into a fresh reg and `local_map_put(name_id, elem_reg)` — no OP_COPY, so scalar elems by-value / aggregates by-address (RFC 0001 §5). Range + integer-bound (`for i in n`) paths kept **byte-identical** (counter_type=type_id, local_map_get==iter_reg).

**Gate:** frontend+IR-lowering, A==B fast fixpoint (**new daily-driver hash `D6DBA89D`**, was `6676E76D`); regression **121/121**. Oracle `t_forcollect` exit **56** (sum-of-elems 26 + max-elem 30; index-sum would be 6 → proves element not index binding).

**Follow-up fixes (found by probing the new path):**
- **BUG#89** `continue` in a `for` loop hung → [[bug89-for-continue-increment]] (`3e26928`).
- **struct-element field access:** `for p in <array-of-struct>` read every field at offset 0 (`p.y` returned `p.x`) — silent miscompile. Typecheck set the loop-var type AFTER inferring the body, so `p.y`'s field index defaulted to 0. Fixed by binding the loop-var element type BEFORE body inference (collection case only; range loops keep post-body assignment for byte-identical codegen). Frontend A==B `F5B8041C`. Oracle `t_forstruct` exit 129.

✅ **P2 Vec[T] SHIPPED** (`d364b6a`, A==B `14825575`, **133/133**, oracle `t_forvec`=60). `for x in vec` where vec: Vec[T]. Vec = `{data:ptr[T], len:i64, cap:i64}` so length is a RUNTIME field & elements live behind `data`: `lower_for` loads `vec.len` (i64) as bound once, each iter loads `vec.data` + `OP_INDEX` with an **i64** counter → loop var = element (scalars by-value / structs by-address, field access works). **KEY LESSON:** a monomorphic Vec value is a **mono STRUCT** `_AX_std_Vec__i64` (kind 1), NOT a GENERIC_INST — check BOTH kinds. New helper `extract_base_type_name` strips module qualifier + generic args → "Vec". New backend helper `fl_resolve_field` resolves data/len field idx+type. Correct ctor form = **`Vec[i64].new()`** (NOT `collections.new_vec[i64]()` — that mis-monomorphizes → segfault). Still P2-future: HashMap/HashSet/string iteration.

**Still OPEN / P2-future:** HashMap / HashSet / string iteration (needs iterator convention / bucket walk). RFC file `rfcs/0018-for-in-collection-iteration.md`.

**Lesson:** ARRAY TypeEntry packs metadata cleverly — `name_id`=length, `extra`=element type (NOT a name). Fixed arrays are the zero-ABI-surface iteration case: length is a compile-time constant, element access reuses existing OP_INDEX.
