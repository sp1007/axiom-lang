---
name: rfc0019-multifield-variant-shipped
description: "RFC 0019 SHIPPED — multi-field variant payloads `Rect(i64,i64)` now work (desugar to synthesized single-struct payload). Closes BUG#81. A==B==C fixpoint, backend."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

✅ **RFC 0019 SHIPPED — multi-field variant payloads work.** Closes BUG#81 [[bug81-multifield-variant-payload]] (was a clean reject; now fully supported).

**What:** `type Shape = Rect(i64,i64) | Tri(i64,i64,i64)` — construct `Rect(3,4)` with N args, `match { Rect(w,h): ... }` binds ALL N fields (previously only field[0]; field[1+] silently dropped → 0/segfault).

**Design = desugar to synthesized struct** (RFC `rfcs/0019-*.md`, chosen for lowest fixpoint risk): a multi-field payload is registered as ONE synth struct `__mfv_<Sum>_<Variant>` with positional fields `_f0.._f(N-1)`. `VariantInfo.payload_type` stays a single `u32` (= that struct) → **VariantInfo shape + sum sizing UNCHANGED**. The constructed value is **byte-identical to a user single-struct payload** (struct pointer at box field 1).

**4 touch points:**
1. `typecheck.ax` pre_infer sum reg (~L1443): when `type_expr_node.next_sibling != 0`, walk all N payload types → `register_struct` synth struct → `payload_type` = it. REMOVED the BUG#81 reject.
2. `typecheck.ax` NODE_MATCH_ARM (~L1903): multi-field pattern binds each var to its struct FIELD type (not the struct), so `w*h` type-checks.
3. `air_builder.ax` `lower_variant_construct`: >1 call arg + struct payload → OP_ALLOC the synth struct, OP_SET_FIELD arg[i]→field i, store struct ptr into box field 1. Single arg keeps the old path. (Disambiguated by ARG COUNT.)
4. `air_builder.ax` `lower_match` variant arm: >1 binding + struct payload → read struct ptr (box field 1), bind pattern child[i]→struct field i. Single binding keeps old path. (Disambiguated by BINDING COUNT.)

**Gate:** backend/lowering, but guarded — no self-host/std source declares a multi-field variant, so self-codegen unchanged → **A==B==C bit-identical** (`8a2a0e8b`, was `f5b8041c`). Regression **125/125**. Oracle `t_mfvariant` (exit 137: Rect(3,4)=12 + Tri(1,2,5) field-order a*100+b*10+c=125). Also verified: -O0/-O1, 3-field, mixed i32/i64, in arrays + for-loops. Obsolete `tests/sema/err_multifield_variant.ax` → replaced by `valid_multifield_variant.ax` (17).

**Also verified working:** aggregate payload FIELDS `Line(Point,Point)` → 33 at -O0/-O1. Feature robust across scalar/mixed/aggregate fields, 2-3 fields, arrays+for-loops, across-fn, inside Option.

**EXTENSION (same iteration) — single `str`/>8B primitive payload FIXED.** `Text(str)` used to return 0 (a 16-byte str stored inline in the box's fixed 8-byte payload slot overflowed — root: `register_sum_type` hardcodes box size 16 = tag8+payload8, typetable.ax:297). Fix: extend the synth-wrap to a SINGLE payload whose type is a `TYPE_KIND_PRIMITIVE` with `size > 8` (i.e. `str`), so the payload becomes an 8-byte struct pointer that fits. Disambiguation switched from arg/binding-COUNT to a **flag on the synth struct's TypeEntry** (`flags` bit 0, set in typecheck; checked in air_builder constructor + match + typecheck match-binding-types) — unifies multi-field and single-str-wrap. No box-sizing change. Oracle `t_strvariant` (44). A==B==C `909d1e07`, regression 126/126. No test/std/compiler declares a single str payload (grep) → self-codegen unaffected. Closes the "str/>8B variant payload" backlog gap.
