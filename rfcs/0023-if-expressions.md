# RFC 0023 — `if`/`elif`/`else` expressions

Status: Accepted — implement (parser NUD + typecheck branch-unify + AIR value-diamond; backend-touching → B==C gate)
Author: autopilot
Related: RFC 0016 P3 (short-circuit `and`/`or` value-diamond lowering — the codegen
model reused here), RFC 0022 (tuple literals), the BUG#53 accept-then-miscompile
convention (inline statement-suite reject).

## 1. Motivation

AXIOM has no conditional *expression*. To pick a value by condition today the user
must introduce a mutable binding and a statement-`if`:

```
mut label := "small"
if n > 10:
    label = "big"
```

or extract a helper `fn`. Both are boilerplate the user explicitly dislikes
([[feedback-ergonomics]]). Every mainstream language offers either `if/else`
expressions (Rust, Kotlin, Scala) or a ternary (`?:`, C/JS). The gap is felt most
sharply now that higher-order methods (`Vec.map/filter/fold`, RFC-less stdlib) invite
inline lambdas — the natural `v.map(|n| if n > 0: n else: 0)` currently fails to parse
(`expected expression nud`), because `if` has no expression form.

This RFC adds `if COND: A elif COND: B else: C` as a value-producing expression.

## 2. Design

### 2.1 Syntax

An **if-expression** is the inline, single-expression-per-branch form:

```
if COND: EXPR (elif COND: EXPR)* else: EXPR
```

* Every branch body is a single value **expression** (not a block, not a statement).
* The `else` branch is **mandatory** — an expression must always produce a value;
  a missing `else` would leave the value undefined when no arm matches. (This is the
  key difference from statement-`if`, where `else` is optional because the statement
  produces no value.)
* `elif` chains are allowed and desugar to nested else-if.
* Branch bodies may themselves be any expression, including a nested if-expression and
  the short-circuit operators.

### 2.2 Disambiguation from statement-`if`

Statement-`if` is unchanged: in **statement position** (`parse_stmt` sees `TK_IF`),
`if COND:` is followed by a newline + INDENT block (`parse_if_stmt`, parser.ax:723).

The if-**expression** is reached only through the Pratt **NUD** for `TK_IF`, i.e. when
`if` appears in **expression position** (`= if …`, a call/return argument, a lambda
body, an operand). The NUD parses the inline form: after each `:` it calls
`parse_expr_with_prec` for a single expression rather than `parse_block`. Because the
two entry points are distinct (statement dispatch vs. expression NUD), there is no
grammar ambiguity: a leading `if` at statement start is always the statement; an `if`
anywhere a value is expected is always the expression.

Consequence: the inline `if COND: STMT` form at statement start remains rejected
(BUG#53) — that reject is about a *statement* body and is orthogonal.

### 2.3 AST

New node `NODE_IF_EXPR`. Children, in order:
`cond0, then0, [cond1, then1, …], else_expr` — i.e. a flat list of
(condition, value) pairs followed by the trailing else value. (Reusing the flat layout
keeps `elif` uniform and avoids a distinct elif-clause node in expression form.) A flag
distinguishes it from `NODE_IF_STMT` for downstream passes; using a separate node kind
is cleaner and avoids touching every existing `NODE_IF_STMT` consumer.

### 2.4 Typecheck

* Infer each branch value's type.
* All `then_i` and the `else` must **unify** to a common result type `R` (the numeric
  literal-coercion rules already used for `let`/return/call args apply, so
  `if c: 1 else: 2` is `i64` under the default-int rule and `if c: 1 else: x` coerces
  the literal to `x`'s type). If two concrete branch types disagree, **reject** with an
  actionable diagnostic (`if-expression branches have incompatible types: R0 vs R1`).
* Each `cond_i` must be `bool` (same check as statement-`if`/`while`).
* `else` is required; a NUD that fails to find `else` emits
  `if-expression requires an else branch (it must produce a value)`.
* The node's result type is `R`.

### 2.5 AIR lowering (backend-touching)

Lower to a **value-producing diamond**, modelled directly on the RFC 0016 P3
short-circuit `and`/`or` lowering (which already yields a value from a CFG diamond and
is CFG-aware-liveness-safe):

1. Allocate one result temp `r` of type `R`.
2. For each `(cond_i, then_i)`: lower `cond_i`; conditional-branch to `then_i`'s block
   or the next test block. In `then_i`'s block, lower `then_i`, `OP_COPY`/store into
   `r`, then jump to the merge block.
3. Final `else`: lower into `r`, jump to merge.
4. Merge block: `r` holds the value.

`R` may be a scalar, a 16-byte `str`, or a by-address aggregate; the store into `r`
reuses the same width-aware copy the existing branch-value / return lowering uses, so
aggregate results ride the established machinery. Because this creates blocks *between*
sub-expressions, it depends on the CFG-aware liveness already shipped (RFC 0016 P2',
[[rfc0016-p2prime-cfg-liveness]]) — no new dataflow work required.

This is a **backend change** (new value-diamond emission): A!=B is expected; the gate is
a hand-built **B==C** fixpoint plus full regression and O0==O1 oracle spot-checks, per
the fixpoint-async rule.

## 3. Alternatives

* **Ternary `c ? a : b`** — terser but a second conditional syntax; AXIOM's block style
  favours keyword forms (`and`/`or`/`not` over `&&`/`||`/`!`), so `if/else` expressions
  are more idiomatic and reuse existing keywords/typecheck.
* **Block-valued `if` (last expression of a block is its value)** — more powerful but a
  much larger change (every block becomes an expression; interacts with the
  statement/expression boundary and reference semantics). Deferred; the inline form
  covers the ergonomic 90% and is a strict subset a future block-value form can extend.
* **`match` as the only conditional expression** — `match` already works but is verbose
  for a boolean pick; if-expressions are the common case.

## 4. Drawbacks

* A second parse path for `if`. Mitigated by the clean statement-vs-expression split.
* Mandatory `else` may surprise users coming from statement-`if`; the diagnostic makes
  the rule explicit.
* Backend value-diamond emission adds codegen surface (hence the B==C gate).

## 5. Migration / compatibility

Purely additive. Existing statement-`if` is untouched; no existing program changes
meaning (an inline `if` in expression position previously failed to parse, so there is
no code relying on the old behavior). No stdlib rewrite required.

## 6. Implementation plan

1. Lexer: none (`if`/`elif`/`else` already tokens).
2. Parser: register a NUD for `TK_IF` producing `NODE_IF_EXPR` (inline branches,
   mandatory `else`). Statement dispatch unchanged.
3. Typecheck: `infer_node` case for `NODE_IF_EXPR` — cond=bool, branch-unify, result R,
   diagnostics for missing-else / incompatible branches.
4. AIR: `lower_if_expr` value-diamond (model on short-circuit lowering).
5. Oracles: scalar pick, numeric-literal coercion, nested if-expr, `elif` chain,
   if-expr as a lambda body (`v.map(|n| if n>0: n else: 0)`), aggregate (`str`) result.
6. Gate: B==C fixpoint + regression.
