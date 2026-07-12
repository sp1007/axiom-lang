# RFC 0022 — Tuple literal expressions

- Status: Accepted (P1 + P2 implemented)
- Author: autopilot
- Date: 2026-07-12
- Affects: lexer (none), parser, typechecker, AIR lowering
- Supersedes/relates: RFC 0019 (multi-field variant payload — synth-struct machinery reused here), RFC 0017 (global storage — the dead `TYPE_KIND_TUPLE` branch this RFC makes reachable)

## Motivation

`TYPE_KIND_TUPLE` (typetable kind 5) exists and is listed in every "is this an
aggregate?" check (`typecheck.check_module_global`, `air_builder` global/param
aggregate tests), yet **no tuple value can be written**: the parser has only
`NODE_TUPLE_PAT` (destructuring patterns), no tuple *expression*. A probe on
2026-07-12 confirmed `let p = (1, 2)` and `p.0` are hard parse errors, so the
tuple branch of `check_module_global` is unreachable dead code and no tuple
oracle has ever existed. Aspirational stdlib (`std/iter.ax`) already assumes
tuples. This RFC makes tuples a first-class, constructible value.

## Design

A tuple is an **anonymous struct** with positionally-named fields `_f0`, `_f1`,
…, `_fN-1`. This reuses the exact synth-struct machinery RFC 0019 introduced for
multi-field variant payloads (`register_struct(name_id, fields)` computes layout,
size, and alignment; existing `OP_ALLOC` / `OP_SET_FIELD` / `OP_GET_FIELD`
lowering handles construction and access). No new type-layout, ABI, or
codegen concept is introduced.

### Surface syntax

- **Literal (NUD on `(`):** after parsing the first parenthesized expression, a
  `,` promotes the group to a tuple. `(a, b)` and `(a, b, c)` are N-tuples;
  `(a,)` (trailing comma) is a 1-tuple; `(a)` with no comma stays a grouping
  expression (unchanged). A trailing comma before `)` is permitted.
- **Field access (`.N`):** `t.0`, `t.1`, … The lexer already tokenizes `t.0` as
  `IDENT DOT INT_LIT` (a bare `.` that is not `.*`/`..` emits `TK_DOT`; the float
  rule only fires while already lexing digits), so **no lexer change**. The `.`
  LED accepts an `INT_LIT` selector and rewrites it to the field name `_f<N>`,
  which resolves against the synth struct's fields by the existing name-based
  field resolution.

### Safety for self-hosting

Both parser additions only **accept previously-rejected** syntax: a comma
immediately inside `(…)` in prefix position, and an integer after `.`, were both
hard parse errors before. The self-hosting compiler source therefore cannot
contain either construct, so adding them cannot change how any existing valid
program (including the compiler itself) parses — the fixpoint **A==B** is
preserved by construction. This mirrors the array-literal NUD (RFC/BUG#70).

### Typing and lowering

- `infer_node(NODE_TUPLE_EXPR)`: infer each element type, `register_struct` a
  synth struct `__tupN` with fields `_f0.._fN-1`, and return that type id (same
  register-per-inference behavior as `NODE_ARRAY_LIT`; duplicate layout-identical
  entries across passes are harmless and deterministic).
- `lower_tuple_lit`: `OP_ALLOC` the struct, then `OP_SET_FIELD` each child
  positionally (identical to `lower_struct_lit` minus the `NODE_NAMED_ARG`
  unwrap).
- Field access flows through the existing struct-field path (`extra_idx` = field
  index), so all aggregate rules (16-byte by-address repr per BUG#77, reference
  semantics per RFC 0001, aggregate return/param) apply unchanged.

## Scope

**P1 (this RFC):** tuple literal construction and `.N` access as local values,
plus everything that composes for free through the struct type (assignment,
passing/returning as an aggregate, nesting in other aggregates).

**P2 (implemented):** tuple *type annotations* `(T0, T1, …)` in `parse_type_expr`
(new `NODE_TUPLE_TYPE`, resolved to the tuple struct), which transitively enables
annotated `let`, tuple **params**, `-> (T..)` **returns**, and tuple **globals**.
Two supporting mechanisms were required for these to be *correct* (not just parse):

- **Canonicalization.** Tuple types are structural: `register_tuple_type` reuses an
  existing `__tup` struct with identical ordered element types instead of minting a
  new one, so a literal and a matching annotation are the *same* type id. The lookup
  is name-prefix-guarded to `__tup` so it never aliases RFC 0019's `__mfv_` variant
  structs (which also use `_f0..` field names).
- **Element coercion by expected tuple type.** When an expected tuple type of the
  same arity is in scope, each element is inferred with the matching field type as
  its hint. Without this, `(10, 20)` infers `{i32,i32}` (8 bytes) and would be stored
  into an annotated `{i64,i64}` slot (16 bytes), reading the second field as garbage.
  A tuple **argument** whose type still differs from a tuple **param** after coercion
  (e.g. an unannotated `let t = (10,15)` — genuinely `{i32,i32}` — passed to an
  `(i64,i64)` param) is **rejected** with an actionable diagnostic (BUG#53
  convention: reject rather than silently miscompile); the fix is to annotate the
  local or pass the literal directly.

**Still deferred:** tuple destructuring in `let`/`match` (the pattern side already
has `NODE_TUPLE_PAT` but is not wired to expressions) and chained `.N.M` (below).

**Known P1 limitation — chained `.N.M`:** the lexer tokenizes `t.0.0` as
`IDENT DOT FLOAT_LIT(0.0)` because the `0.0` run lexes as a float, so *chained*
tuple field access does not parse. Single-level `.N` on a value is fine. The
workaround is an intermediate binding (`let a = t.0` then `a.0`). A proper fix
(splitting a `FLOAT_LIT` selector `0.1` into `._f0._f1`, mirroring how Rust
resolved the same ambiguity) is deferred to a later phase.

## Drawbacks / alternatives

- **Anonymous-struct desugaring vs. a distinct `TYPE_KIND_TUPLE`:** using the
  struct kind means tuples inherit struct behavior for free but never exercise
  the `TYPE_KIND_TUPLE` code paths. Those remain reserved for a future distinct
  representation if structural typing (tuple-of-same-shape interchangeability)
  is ever required; today AXIOM structs are nominal, and P1 does not promise
  structural equivalence between two independently-written `(i64, i64)` literals.
- Per-inference struct registration adds a few duplicate type-table entries;
  negligible and deterministic (matches `NODE_ARRAY_LIT`).

## Migration

Purely additive. No existing syntax changes meaning.
