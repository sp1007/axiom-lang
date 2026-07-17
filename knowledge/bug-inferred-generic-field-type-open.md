---
name: bug-inferred-generic-field-type-open
description: "CLOSED. Layer 1 (annotated Box[i64] field) FIXED 08f6c11. 'Layer 2' was MISDIAGNOSED — generic monomorphization is correct; the reject came from bare-int defaulting to i32 + numeric-UFCS exact-match. Fixed a02db69 (narrow-int UFCS widening)."
metadata:
  node_type: memory
  type: project
  originSessionId: 4812f945-8f5c-44ec-92ce-e746238ea5f1
---

# ✅ CLOSED — generic-inst field type monomorphization

## ✅ Layer 1 FIXED `08f6c11` (A==B, oracle t_geninstfield=44)
Field read off a GENERIC_INST with CONCRETE args (`Box[i64]` annotated/explicit)
returned the TEMPLATE placeholder `T`, not `i64`. Fixed at the typecheck field-
inference site (~3729): substitute the generic-param field type via the instance's
concrete args (guarded so mono/concrete structs are a no-op → A==B preserved).

## ✅ "Layer 2" — MISDIAGNOSED, then CLOSED `a02db69`
Prior sessions thought `let b = Box(v:42); b.v.to_str()` rejecting meant the
constructor failed to stamp `T=i64` onto the inferred instance. **That was wrong.**
Investigator (2026-07-12) proved generic mono is CORRECT: `let a = Box(v:100 as i64)`
+ `let c = Box(v:2.5)` coexist and both `.v.to_str()` work (exit 6). The moment the
field value is a concrete i64/f64 the fully-inferred `Box(v:…)` works.

Real cause = two independent, NON-generic facts:
1. A bare int literal with no expected type defaults to **i32** (typecheck NODE_INT_LIT
   else-branch ~3561). So `Box(v:42)` infers field `T=i32` — correctly.
2. The numeric-UFCS method-form `x.to_str()` required an **exact** first-param match
   (method_ret_type ~1096 / air_builder dispatch ~1598). The `std.string.to_str`
   overloads take `i64`/`f64`, so an i32 receiver matched neither → clean reject.
   The free-CALL form `to_str(x)` worked because ordinary arg passing coerces i32→i64.
   Non-generic `let x=42; x.to_str()` rejected identically — nothing generic about it.

## Fix `a02db69` — numeric UFCS widening (frontend-only, A==B `943C595F`)
Widened BOTH gates so a narrower integer receiver (i8/i16/i32/u8/u16/u32) matches an
i64 first-param free fn. **Gates-only, NO receiver cast**: every narrow-int producer
already normalizes its register to full 64-bit width (x86_selector emit_wrap_to_width
/ emit_load_extend), so the receiver value equals `recv as i64` already — passing it
directly is exactly the coercion the call-form does. Deliberately **NOT** f32→f64: an
f32 in an XMM reg is not an f64 (needs a real cvtss2sd), which would miscompile with
no emitted cast. Oracle `t_numufcswide` (exit 7). Regression 181/181.

STILL niche-open (separate, low value): `u64` decimal overload, uppercase hex — see
[[backlog-open-items]].

## Related
[[bug92-generic-recursive-multifield-open]] (FIXED). [[string-utf8-default]].
