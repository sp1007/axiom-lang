# RFC 0019 — Multi-field variant payloads (`Rect(i64, i64)`)

- **Status:** Accepted (design); implementation pending (next task).
- **Author:** autopilot
- **Supersedes the reject in:** BUG#81 ([[bug81-multifield-variant-payload]]) — currently a variant with >1 payload field is rejected with "wrap them in a struct".
- **Requires (CLAUDE.md §13):** representation change → this RFC before implementing. Backend/ABI → **B==C fixpoint mandatory before commit**.

## Motivation

Tuple-like variants (`Rect(i64, i64)`, `Line(Point, Point)`) are a natural, common
ADT shape. Today they are rejected; the user must hand-wrap the fields in a named
struct — exactly the boilerplate the project's ergonomics direction avoids
([[feedback-ergonomics]]). This RFC makes multi-field variant payloads work.

## Chosen design — desugar to a synthesized struct payload

Rather than change `VariantInfo` (single `payload_type: u32`) into a payload **list**
— which would ripple through `find_variant_info`, sum box **sizing**, mono, and every
self-host use of sums (high fixpoint risk) — we **desugar** an N-field payload to a
single **synthesized struct**:

```
type Shape = Rect(i64, i64) | Circle(i64)
```
is treated exactly as if the user had written:
```
struct __payload_Shape_Rect:      # synthesized, N fields _0.._(N-1)
    _0: i64
    _1: i64
type Shape = Rect(__payload_Shape_Rect) | Circle(i64)
```

This **reuses the proven single-struct-payload path** (BUG#58; `t_variantstruct`), which
is validated to work for a 16-byte (2×i64) struct payload (probe `sp.ax` → 12). So:
- `VariantInfo.payload_type` stays a single `u32` = the synthesized struct's type id →
  **no VariantInfo shape change, no sum-sizing change** (the struct's own size, computed
  by the existing struct sizing, is the payload size the box is sized for).
- No new opcode, no ABI change beyond what struct-payload variants already do.

### Touch points (precise)

1. **typecheck** `pre_infer` sum registration — [typecheck.ax] ~L1443-1462 (the current
   reject site). When `type_expr_node.next_sibling != 0` (≥2 payload types):
   - Collect all N payload type ids (walk `first_child` → `.next_sibling`).
   - `register_struct` a synthetic struct named e.g. `__payload_<Sum>_<Variant>` with
     fields `_0.._(N-1)` of those types (dedup by name like other registrations).
   - Set the variant's `payload_type` = that struct type id. **Remove the reject.**
   - The variant symbol's constructor arity is N (already parsed as N type-children).

2. **air_builder constructor** `lower_variant_construct` — [air_builder.ax] ~L2344-2387.
   Today stores ONE `payload_arg_idx` at box field 1. For a synthesized-struct payload
   with N call args (`Rect(3,4)`): allocate the struct (OP_ALLOC struct type), store
   `arg[i]` into struct field `i` (OP_SET_FIELD), then store the struct pointer into the
   box payload slot (field 1) — same as a user single-struct payload. Needs the call
   site (~L1630-1655) to pass the **arg list**, not just `arg[0]`, when arity > 1.

3. **air_builder match** `lower_match` variant-pattern arm — [air_builder.ax] ~L2654-2715.
   Today binds ONE `payload_bind_sym` (first child) to box field 1. For an N-binding
   pattern (`Rect(w, h)`): read the struct payload (box field 1) once, then bind
   `binding[i]` → struct field `i` (OP_GET_FIELD with the field's type). Walk all
   NODE_BINDING_PAT children, not just the first.

4. **typecheck match binding types** — the pattern vars `w`, `h` get the synthesized
   struct's field types (i64, i64) so field access / arithmetic type-checks.

## Drawbacks / alternatives

- **(alt) multi-slot box** (VariantInfo → payload list; store fields at box slots 1..N;
  size box = tag + max(sum of variant field sizes)). More "direct" (no struct
  indirection) but changes the fixpoint-sensitive `VariantInfo` + sum sizing → higher
  self-host risk. Rejected for v1 in favor of the desugar's lower blast radius.
- Extra indirection: the desugar adds one heap struct box per multi-field variant value
  (same cost the user's manual `wrap-in-struct` workaround already pays). Acceptable.

## Migration / fixpoint

- Purely additive: programs that got the BUG#81 reject now compile. No existing valid
  program changes (single-payload variants untouched; the synth path only triggers when
  `next_sibling != 0`). No self-host source declares a multi-field variant (grep to
  confirm before building) → self-codegen for existing code is unaffected, but the
  constructor/match changes touch shared lowering → treat as backend: **hand-build C
  from B, require B==C bit-identical** + full regression + a new oracle.

## Test

Oracle `t_mfvariant`: `type Shape = Rect(i64,i64) | Circle(i64)`; `Rect(3,4)` →
`match { Rect(w,h): w*h ; Circle(r): r }` → **12**. Add a 3-field and a mixed-type case.
Keep `tests/sema/err_multifield_variant.ax` only if we decide to still reject some
subcase; otherwise remove it (feature now supported).
