---
name: bug91-generic-sum-ctor-inference-open
description: "BUG#91 FIXED — an unannotated user generic-sum match (`let p = One(42)` for `Opt[T]`) silently lowered to NOTHING (returned 0): scrutinee typed as GENERIC_INST, but lower_match/find_variant_info only handled kind-6 SUM. Fixed via resolve_sum_type."
metadata: 
  node_type: memory
  type: project
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

✅ **BUG#91 FIXED** — unannotated user generic-sum constructor + match now works.

**Symptom:** `type Opt[T] = One(T) | Non; let p = One(42 as i64); match p { One(a): a; Non: 0 }` returned **0** (expected 42). Annotated `let p: Opt[i64] = One(42)` worked. Affected generic single- AND multi-field variants; found by probing generics × ADT.

**Root cause (NOT inference — LOWERING):** `try_instantiate_variant_call` DID infer `T=i64` and typed `p` as a **GENERIC_INST** `Opt[i64]` (kind 8), not a kind-6 SUM. But `lower_match` set `is_sum` only for `TYPE_KIND_SUM`, and `find_variant_info` early-returned for non-SUM — so for a generic_inst scrutinee the whole `match` lowered to NOTHING (dump-air: construct then bare `ret`, no dispatch). The annotated path monomorphizes to a real kind-6 sum, which is why it worked. (Construction was fine all along — the payload 42 was stored.)

**Fix (backend, air_builder.ax):** new `resolve_sum_type(type_id)` sees through a GENERIC_INST to the base user sum (a generic instance's `name_id` equals its base's — same shortcut `builder_type_size` uses). `find_variant_info` and `lower_match`'s is_sum check both go through it. Payload `OP_GET_FIELD` still uses the template `payload_type` (`T`), but an 8-byte scalar loads fine; generic MULTI-field variants also work (the RFC 0019 flagged synth struct carries them).

**Gate:** backend/lowering, guarded (only adds handling for a previously-no-op case) → A==B==C bit-identical (`3b704a37`, was `39695e6b`); regression **128/128**. Oracle `t_gensumctor` (54 = unannotated `One(42)`→42 + `Both(3,4)`→12). Self-host unaffected (compiler didn't rely on unannotated generic-inst sum matches).

**Note:** payload type stays `T` (template) not substituted to the concrete arg — OK for scalar/pointer payloads (8-byte load). A future edge case: a generic sum whose payload monomorphizes to a differently-SIZED aggregate might need T→concrete substitution in the match GET_FIELD. Not hit by current tests.
