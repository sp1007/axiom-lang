# RFC 0039 — Annotation-inferred struct literals

- **Status:** Implemented (2026-08-07) — phase 1 as specified in §2.2
  - Parser: `parser.ax:295-311` (prefix `(` + `ident` + `:` ⇒ `NODE_STRUCT_LIT` placeholder in
    the callee slot, then the ordinary named-argument call-args parser).
  - Typecheck: `typecheck.ax:4277-4306` fills the placeholder from the annotated binding's
    `SYM_STRUCT`; `typecheck.ax:5012-5028` reports **`error[E3034]`** when it was never filled.
  - The placeholder is bound by **symbol identity**, not spelling: the annotation's payload is
    only a symbol index when the resolver actually resolved it, so the symbol's `name_id` is
    checked against the annotation's text before binding. This is a direct application of the
    lesson from P6 (`74eab1d`), where an unguarded payload read landed on an unrelated symbol.
  - Gate: **A == B = `84A13E958B59D2A1022C860C8E4637E81716BA65A66BFDDFD096034E4DB3FF68`**,
    regression **685/685** at default and `-O0`.
  - Oracles: `bin/t_structlitinfer.ax` (value-checked, with tuple and grouped-expression
    controls) and `bin/t_structlitinfer_noctx.ax`, plus a `structlitinfer-ecode` block that
    pins **exactly one** E3034 so a cascade cannot creep back in.
- **Author:** AXIOM compiler team
- **Created:** 2026-08-06
- **Affects:** parser, typecheck (frontend only — no IR, ABI, backend or linker change)
- **Gate:** A == B

---

## 1. Motivation

A struct literal today must repeat the type name even when the binding already states it:

```axiom
struct TmpStruct:
    a: i64
    b: f64

let c: TmpStruct = TmpStruct(a: 64, b: 64)
//     ^^^^^^^^^   ^^^^^^^^^ said twice
```

A user reading `examples/bigfloat128.ax` sees

```axiom
let onehalf = U128(lo: HALF, hi: 1 as u64)     // examples/bigfloat128.ax:99
```

and reasonably concludes that the type name belongs *somewhere*, then writes the annotated form
with the name in the annotation instead of on the literal:

```axiom
let c: TmpStruct = (a: 64, b: 64)      // rejected today
```

This is a real user report (2026-08-06). The rejection is currently delivered as
`error: unexpected token at offset 142042` — a byte offset into the concatenated stdlib buffer,
followed by 8 cascading parse errors and a raw token dump (see `bin/probe11/s1.ax`). Even if this
RFC were declined, that diagnostic must be fixed; see §7.

The compiler already holds the annotated type at the point the initializer is checked, so the
information needed to accept the shorter form is present and unused.

**Measured scope of the status quo:** 3575 struct literals across `std/` and `bootstrap/stage1/`,
every one of them naming its type. Zero use the bare form. There is not a single binding in the
repository that carries *both* an annotation and a struct literal — i.e. the redundant form this
RFC removes is one nobody writes, and the concise form is one nobody can write.

## 2. Design

### 2.1 Rule

When a **struct-typed context supplies an expected type**, a parenthesized named-field list

```
( field₁: expr₁ , field₂: expr₂ , … )
```

is a literal of that expected type. It is exactly equivalent to writing `T(field₁: expr₁, …)` where
`T` is the expected type, and is checked by the identical code path — same field-presence rules,
same ordering freedom, same coercions, same diagnostics.

### 2.2 Contexts that supply an expected type

Accepted in this RFC (phase 1) — the annotated `let`/`mut` binding only:

```axiom
let  c: TmpStruct = (a: 64, b: 64)
mut  d: TmpStruct := (a: 64, b: 64)
```

**Deliberately NOT in phase 1**, though each is a natural extension: function arguments, `return`
position, struct field initializers, array elements, and assignment to an already-declared binding.
Phase 1 is limited to the one context the user actually hit, so the change is small enough to be
reviewed against the 679-row regression baseline in a single commit. Extending it is a follow-up
that reuses the same `check_annotated_target` plumbing that RFC 0006 §6.4 already established for
E3030/E3031/E3032, and should be a separate RFC amendment with its own gate.

### 2.3 What stays an error

Without an expected type there is nothing to infer, and the form is rejected with a diagnostic that
says so:

```axiom
let e = (a: 64, b: 64)
//      ^ error[E3034]: cannot infer the struct type of this literal
//        help: name the type, as in `TmpStruct(a: 64, b: 64)`, or annotate the binding
```

Also unchanged: `T(...)` remains legal everywhere it is legal today. This RFC **adds** a form; it
removes nothing and changes no existing program's meaning.

### 2.4 Ambiguity with tuples

AXIOM's tuple syntax is positional — `(1, 2)`. The inferred struct literal is distinguished by
`ident :` after the open parenthesis, which no tuple element can begin with. `(a: 64)` is therefore
never a 1-tuple, and a parenthesized expression `(a)` is untouched because it has no `:`. The
lookahead is one token past an identifier, decided in the parser's prefix position.

## 3. Implementation

Grounded in the current tree; every line reference verified 2026-08-06.

**Key fact that shapes the whole design:** `NODE_STRUCT_LIT` (`ast.ax:53`) exists but **the parser
never constructs it** — `grep 'add_node(NODE_STRUCT_LIT'` returns nothing. `U128(lo: …, hi: …)` is
parsed as an ordinary **call expression with named arguments**, and it is *typecheck* that
recognizes the callee as a `SYM_STRUCT` and turns the call into a construction (`typecheck.ax:732`,
`:794`, `:5621`, `:6178`). `NODE_STRUCT_LIT` survives only as a *classification* input in
`escape.ax:137,279` and `air_builder.ax:562`.

⇒ The implementation must therefore **not** invent a second construction path. It rewrites the bare
form into the shape the existing path already consumes:

1. **Parser.** In prefix position, on `(` followed by `ident` `:`, parse the named-field list and
   emit the *same node shape* a call with named arguments produces, but with the callee left
   unresolved (a sentinel). Do not add a new node kind; do not duplicate the field-list parser.
2. **Typecheck.** At the annotated-binding site, when the initializer is that sentinel shape, fill
   the callee from the annotation's `SYM_STRUCT` and then **fall through to the existing
   constructor-call checking code unchanged**. If no expected type is available, emit `E3034`.

This keeps the entire field-checking, coercion and layout story in one place. Writing a parallel
checker for the bare form would recreate the exact "one rule, two copies" shape that produced the
interface-return miscompile, the `lower_int_lit`/`lower_float_lit` divergence, and the
`resolve_method_overload`/`resolve_free_call_overload` arity split — the single most expensive
recurring defect class in this project's history.

## 4. Alternatives considered

- **Zig-style `.{ … }` anonymous literal.** Explicitly marks "infer my type", removing all tuple
  ambiguity. Rejected for phase 1: it introduces new punctuation the language does not otherwise
  use, and the user's report is specifically about the annotated form reading naturally without a
  sigil. Worth revisiting if §2.2 is ever widened to argument position, where the expected type is
  less visually adjacent.
- **Require the type name always (status quo).** One way to write a thing, zero ambiguity, no RFC.
  This remains a defensible position — the argument against it is that the annotation is already
  mandatory-looking to users and the duplication is pure noise. If this alternative is chosen, the
  §7 diagnostic work is still required.
- **Infer from the *field names* alone**, with no expected type, by finding the unique struct with
  that field set. Rejected: it makes adding a field to an unrelated struct a spooky action at a
  distance that can silently re-target an existing literal, and it is not decidable in the presence
  of two structs with identical field sets.

## 5. Drawbacks

- Two spellings for one construction. Mitigated by §2.3: the bare form is legal *only* where the
  type is stated adjacently, so it never reduces local readability.
- The parser gains one token of lookahead in prefix position.
- Widening §2.2 later means auditing every expected-type context; doing that piecemeal risks the
  same partial coverage that RFC 0006 §6.4 had to go back and close for method arguments.

## 6. Migration and compatibility

Purely additive. No existing program changes meaning, so no migration is needed and the 3575
existing literals are untouched. Frontend-only ⇒ **A == B**, and the compiler's own sources use no
bare literals, so its self-image is unaffected.

## 7. Diagnostics (required regardless of this RFC's outcome)

The current rejection of `let c: TmpStruct = (a: 64, b: 64)` is, verbatim:

```
  tokens[32928]: kind=67 offset=142042
error: unexpected token at offset 142042
error: expected newline at offset 142042
error: expected expression nud at offset 142042
… 9 parse error(s)
```

Four separate CLAUDE.md §8 violations: a byte offset into the **concatenated stdlib+user** buffer
rather than `file:line:col` in the user's own 9-line file; nine cascading errors from one mistake;
a raw token dump with numeric token kinds on a user-facing path; and `nud`, an internal
Pratt-parser term, in user-visible text. Repro committed at `bin/probe11/s1.ax`; tracked in
`knowledge/BACKLOG.md` §3b.

## 8. Testing

Per CLAUDE.md §7.1, oracles assert on **stdout** via `println("<UTF-8 string>: ", value)`, not on
the exit code.

- `bin/t_structlitinfer.ax` — the accepted form, including a field requiring int→f64 coercion
  (`b: 64` into an `f64` field, the exact case from the user report), a field read back through
  `c.a` as a `for` bound, and the equivalence check that the bare and named forms produce identical
  values. One line of output, non-ASCII text.
- `bin/t_structlitinfer_noctx.ax` — `let e = (a: 64, b: 64)` ⇒ `reject|` row, pinning `E3034`.
- Both at `-O0` and `-O1`.
- The 679-row baseline must hold, and `bin/probe11/s2.ax` (the named form) must keep working —
  the point of §3 is that it goes through the *same* code, so a regression there would mean the
  implementation forked the path after all.
