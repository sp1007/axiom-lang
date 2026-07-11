# RFC 0021 — `Vec[T]` index operator `v[i]`

Status: Accepted — implement (frontend + AIR lowering, self-host A==B)
Author: autopilot
Related: RFC 0018 (for-in iteration), `std/collections.ax` (Vec), the BUG#53
accept-then-miscompile convention.

## 1. Motivation

`Vec[T]` is AXIOM's growable array. Today its elements are reachable only via
`v.get(i) -> Option[T]` (bounds-checked) or `for x in v` (RFC 0018 P2). The natural
`v[i]` subscript was **silently miscompiled**: typecheck's `NODE_INDEX_EXPR` typed
only pointer/slice/array/str operands, so a `Vec` fell through untyped and
air_builder emitted `OP_INDEX` on the Vec's *address* — reading the `{data,len,cap}`
header (`c[1]` returned the length; wider element types segfaulted) instead of
`vec.data[i]`. That miscompile was closed by a REJECT (commit `d94aa8e`). This RFC
turns the reject into the expected, ergonomic behavior: `v[i]` loads `vec.data[i]`.

The user prioritizes ergonomics over boilerplate; `v[i]` is a near-universal
expectation and removes the `.get(i).unwrap()` ceremony for the common
already-in-bounds access.

## 2. Design

`v[i]` on a value of type `Vec[T]` evaluates to the element `T` at index `i`. It is
**exactly** `vec.data[i]`: load the `data` pointer field, then index it — the same
two-instruction sequence RFC 0018 P2's for-in-vec lowering already emits
(`air_builder.ax` ~3688: `OP_GET_FIELD data` then `OP_INDEX`). Semantics:

- **Type:** result type is the Vec's element type `T`, resolved via
  `get_generic_args(vec_type)[0]` (works for a monomorphic Vec STRUCT
  `_AX_std_Vec__i64` and a generic-context `GENERIC_INST`), mirroring the for-in-vec
  element-type binding.
- **Bounds:** like Rust/C and AXIOM's existing raw pointer/array `p[i]`, `v[i]` is
  **unchecked** and O(1). Out-of-range is undefined (reads past `data`). Bounds-safe
  access remains `v.get(i) -> Option[T]`. (A checked `v.at(i)`/panic variant can be a
  later addition; keeping `[]` unchecked matches the existing `ptr[T]`/array `[]`.)
- **Element width:** the element load reuses `OP_INDEX` with the element type, which
  already handles scalar / 16-byte / by-address aggregate elements correctly (the
  for-in-vec path relies on the same instruction for arbitrary `T`).
- **Lvalue:** `v[i] = x` (index as an assignment target) is **out of scope** here —
  this RFC covers the read (rvalue) form only; write-through-subscript is a follow-up.

## 3. Alternatives considered

1. **Keep the REJECT (status quo after `d94aa8e`).** Correct and safe, but leaves the
   ergonomic gap and diverges from every array-like type already supporting `[]`.
2. **Make `v[i]` bounds-checked (return `Option[T]` / panic).** Rejected as the `[]`
   default: it would make `[]` semantically different from the existing raw `ptr`/array
   `[]` (unchecked, O(1)) and hide a branch in a hot path. Checked access stays `.get`.
3. **A general `Index` trait/operator overload.** Rejected for now: AXIOM has no
   operator-trait system for indexing; this RFC hard-wires the one built-in collection
   (`Vec`) exactly as for-in already special-cases it. A general `Index` trait is a
   larger, separate design.

## 4. Drawbacks

- Adds an unchecked footgun (out-of-range `v[i]`), same class as the existing
  `ptr[T]`/array `[]`. Documented; `.get` remains for safety.
- `[]` now has two lowerings (pointer/array direct vs Vec data-load), selected by the
  operand's static type — a small asymmetry, contained to `lower_index_expr`.

## 5. Migration / compatibility

Fully additive. The compiler and all current programs never index a `Vec` with `[]`
(they use `.data[i]` / `.get` / for-in), so no existing code changes and the
self-host fixpoint is **A==B** (the new lowering branch is only reachable for a Vec
operand, of which the compiler source has none). Turns a previously-rejected program
into a working one — no behavior change for valid programs.

## 6. Implementation

- **typecheck** (`NODE_INDEX_EXPR`): when the operand type's base name is `Vec`, set
  `result_type = get_generic_args(col_type)[0]` (reject only if the element type is
  indeterminate). Replaces the interim REJECT.
- **air_builder** (`lower_index_expr`): when the operand's `node_types` entry is a Vec
  (base name `Vec`), `fl_resolve_field(data)` → `OP_GET_FIELD` to get the `data`
  pointer, then `OP_INDEX` on it; otherwise the existing direct `OP_INDEX`.
- Oracle `t_vecindex` flips from `reject` to `exit 7` (`c[0]==70,c[1]==80,c[2]==90`).

## 7. Verification

Fast fixpoint **A==B** (no self-codegen change) + full regression. Oracle exercises a
scalar element `Vec[i64]`; the for-in-vec path already validates wider element types
through the identical `OP_GET_FIELD`+`OP_INDEX` sequence.
