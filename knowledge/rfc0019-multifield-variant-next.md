---
name: rfc0019-multifield-variant-next
description: "NEXT TASK — implement RFC 0019 multi-field variant payloads (Rect(i64,i64)) via desugar-to-synthesized-struct. Design done+committed (f2d7137); approach validated. Precise plan + touch points inside. Backend → B==C mandatory."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

🎯 **NEXT CONCRETE TASK: implement RFC 0019** (`rfcs/0019-multi-field-variant-payload.md`, committed `f2d7137`). Unblocks BUG#81 [[bug81-multifield-variant-payload]] (currently multi-field variants are REJECTED).

**Approach (validated):** desugar an N-field variant payload to ONE synthesized struct → reuse the proven single-struct-payload path (BUG#58). **VariantInfo shape + sum sizing UNCHANGED** (lowest fixpoint risk). Validated the target repr: a `Rect(RectP)` variant with `struct RectP { w:i64, h:i64 }` compiles + matches → 12 (probe `sp.ax`).

**All-or-nothing:** do NOT land a partial. Removing the typecheck reject WITHOUT the constructor+match changes re-introduces the BUG#81 silent miscompile (2nd+ field reads 0 / segfault). Land all three pieces + gate together.

**Touch points (verified line refs at f2d7137):**
1. **typecheck** `pre_infer` sum reg — `bootstrap/stage1/typecheck.ax` ~L1443-1462 (reject site). When `type_expr_node.next_sibling != 0`: walk `first_child`→`.next_sibling` to collect N payload types; `register_struct` a synth struct `__payload_<Sum>_<Variant>` fields `_0.._(N-1)`; set variant `payload_type` = that struct id; REMOVE the reject. (Check `register_struct` API in typetable.ax / how structs are registered in typecheck.)
2. **air_builder constructor** `lower_variant_construct` — `air_builder.ax` ~L2344-2387 (single payload → box field 1 via OP_SET_FIELD). For N args: OP_ALLOC the synth struct, OP_SET_FIELD arg[i]→struct field i, then store struct ptr into box field 1. Call sites ~L1638/1652 must pass the ARG LIST (currently only `arg[0]`/`callee.next_sibling`).
3. **air_builder match** `lower_match` variant arm — `air_builder.ax` ~L2654-2715. Currently binds ONE `payload_bind_sym` (pat first_child) to box field 1 (OP_GET_FIELD src2=1, payload_type). For N bindings: read struct payload (field 1) once, then bind pattern child[i] → struct field i (OP_GET_FIELD field-i, field type). Walk ALL NODE_BINDING_PAT children.
4. **typecheck match binding types** — give pattern vars `w,h` the synth struct field types so `p.w`/arithmetic type-checks. (May be implicit once payload_type is the struct.)

**Gate:** backend/lowering → **A!=B expected, hand-build B==C** (`fast_fixpoint` A/B then B→C, require B==C bit-identical) + full regression + oracle `t_mfvariant` (Rect(3,4)→match→12; add 3-field + mixed-type). Grep confirmed NO self-host/std source declares a multi-field variant (self-codegen unaffected). Field offsets for a struct payload are handled by `field_offset` (x86_selector.ax L229). Box sized by `type_size_and_align` on the struct payload (OP_ALLOC L1648).

**Key files:** typecheck.ax, air_builder.ax, typetable.ax (register_struct). See [[rfc0018-for-in-array-shipped]] for the recent for-loop work; [[inline-match-arm-unsupported]] for probe lessons.
