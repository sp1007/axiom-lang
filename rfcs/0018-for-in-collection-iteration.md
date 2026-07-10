# RFC 0018 — `for x in <collection>` element iteration

Status: Accepted (P1 arrays + P2 Vec[T] shipped)
Author: autopilot
Related: BUG#87 (for-range loop var), BUG#72 (slice reject convention)

## Motivation

`for i in a..b` (range) works, but `for x in <collection>` — the ergonomic
element-iteration form every modern language has — is unimplemented. Before this
RFC the frontend *rejected* it cleanly (typecheck emits a diagnostic for
ARRAY/STRUCT/SUM/GENERIC_INST iterees; the raw `lower_for` non-range stub would
otherwise treat the collection's address as a numeric loop bound and hang).

The user values removing boilerplate: `for x in arr` instead of
`for i in 0..N: let x = arr[i]`. This RFC introduces real element iteration,
starting with the cleanest, fully-static case.

## Design

### P1 — fixed arrays (this RFC, shipped)

`for x in arr` where `arr` has a fixed-array type `[T; N]`
(`TYPE_KIND_ARRAY`, `entry.name_id == N` compile-time length,
`entry.extra == T` element type).

Semantics — desugars to an index loop over `0..N`, binding `x` to `arr[i]`
each iteration:

```
for x in arr:        # arr: [T; N]
    <body>
```
lowers to (conceptually):
```
i := 0
while i < N:
    x := arr[i]      # OP_INDEX, element type T
    <body>
    i := i + 1
```

- **Typecheck** (`typecheck.ax`, NODE_FOR_STMT): when the iteree is a
  `TYPE_KIND_ARRAY`, bind the loop variable's `type_id` to the element type
  `extra` (instead of the default `TYPE_I32`), and do NOT emit the
  "not supported" diagnostic. STRUCT/SUM/GENERIC_INST iterees remain rejected
  (that is P2).
- **Lowering** (`air_builder.ax`, `lower_for`): detect the collection case via
  `self.mb.node_types[range_expr]` being a `TYPE_KIND_ARRAY`. Evaluate the
  collection base register once, use an *internal* index counter (distinct from
  the loop variable) counting `0..name_id`, and at the top of the body emit
  `OP_INDEX(base, i)` with the element type into a fresh register that the loop
  variable is bound to (`local_map_put`). No `OP_COPY` — the loop var maps
  directly to the per-iteration `OP_INDEX` result, which naturally handles both
  scalar (by-value) and aggregate (by-address, reference semantics per RFC 0001
  §5) elements.

Element access reuses the existing `OP_INDEX` machinery (same as `arr[i]`), so
no new opcode and no ABI change. The internal counter register is never exposed
to the program, so the loop variable holds the *element*, matching BUG#87's rule
that the loop var is read via `local_map_get(sym_idx)`.

### P2 — Vec[T] (shipped)

`for x in vec` where `vec: Vec[T]`. A Vec is `{ data: ptr[T], len: i64, cap: i64 }`,
so unlike a fixed array the length is a RUNTIME field and elements live behind the
`data` pointer. `lower_for` loads `vec.len` (i64) once as the bound, and each
iteration loads `vec.data` and `OP_INDEX`es it with an i64 counter, binding the loop
var to the element (scalars by-value, aggregates by-address per RFC 0001 §5).

- **Typecheck** (`typecheck.ax`, NODE_FOR_STMT): a Vec iteree resolves to a mono
  STRUCT (`_AX_std_Vec__i64`) or GENERIC_INST; recognized via `extract_base_type_name`
  (strips module qualifier + generic args -> "Vec"). Loop var bound to the sole
  generic arg (element type). HashMap/other structs still rejected.
- **Lowering** (`air_builder.ax`): new `fl_resolve_field` resolves the data/len
  field indices+types; the Vec branch of `lower_for` emits the field-loads + index.
- **Test**: oracle `t_forvec` (element sum 60). Compiler source uses no for-in-Vec,
  so self-codegen is byte-identical (A==B).

### P2 — future

- `HashMap` / `HashSet` / string iteration (needs an iterator convention / bucket
  walk). Still rejected.
- Multi-dimensional / nested-array iteration.

## Drawbacks / alternatives

- Alternative: a general iterator protocol (`next()`/`Option`) up front. Rejected
  for P1 as premature — fixed arrays need none of that machinery and cover the
  common static case with zero ABI surface. P2 can layer the protocol on top.

## Migration / compatibility

Purely additive: programs that previously got the reject diagnostic now compile.
No existing valid program changes behavior. Backend emits only existing opcodes,
so self-codegen for code that does not use array iteration is byte-identical
(A==B expected); the compiler source itself does not yet use this form.

## Test

Oracle `t_forcollect`: sum a fixed array via `for x in arr` and exit with the
sum; independent range-based oracle must still pass.
