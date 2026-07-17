---
name: bug-malformed-input-robustness-cluster
description: "OPEN cluster (probe batch 9, 2026-07-18): 4 malformed programs the compiler mis-handles instead of cleanly rejecting — a self-recursive struct CRASHES the compiler, and calling a non-fn / matching a non-sum / str→i64 assignment are all silently accepted then segfault or return garbage. All BUG#53-class: should REJECT with a diagnostic."
metadata:
  node_type: memory
  type: project
---

# Malformed-input robustness cluster — compiler crash / accept-then-miscompile (OPEN, 4 items)

Found by the FIRST malformed-input probe pass this session (all prior probing used VALID programs
to hunt miscompiles; this pass feeds BAD programs and checks reject-vs-crash). All 4 are the BUG#53
convention: a silent accept-then-crash/miscompile must become a clean REJECT with a diagnostic.
None emit any diagnostic today. Repro files in `/tmp/pb9/{m1,m2,m4,m5}.ax` (regenerate from below).

## ✅ m1 — FIXED 2026-07-18 (A==B `DBA48355`, oracle t_recstructreject, reject mode)
Was: compiler crashed in codegen. Now REJECTS: "recursive struct field of its own type has
infinite size; use `ptr[...]` for the recursive field". Fix in `pre_infer_struct` (typecheck.ax
~L2262): when a field's `type_id == sym.type_id` (the struct's own type, set by the earlier
pre-registration pass), emit the diagnostic + bump `diags_count` (driver halts before codegen).
`ptr[S]` resolves to a distinct POINTER type so it's correctly allowed; self-build OK (the compiler's
own recursive structs all use ptr), valid linked-list + recursive-SUM Tree still work, regression
green. LIMITATION: catches DIRECT self-reference only; an indirect by-value cycle (A has field B, B
has field A) is not yet detected — follow-up (needs a small cycle walk over struct field types).

## m1 (original report) — self-recursive struct by value → COMPILER CRASHES
```
struct S:
    x: S        # infinite size: S contains an S by value
    n: i64
fn main() -> i64:
    let s = S(x: S(x: S(n: 0), n: 1), n: 2)
    return s.n
```
The compiler accepts it through typecheck and **crashes during codegen** (reached `[codegen] func
140/163`, then exit 127 — a compiler abort, no diagnostic). A compiler must NEVER crash on input.
FIX: at struct registration / layout (`builder_type_size_and_align` or struct-decl typecheck),
detect a field whose type is the struct itself (directly, and ideally through a by-value cycle
A→B→A) and REJECT: "recursive struct 'S' has infinite size; use `ptr[S]` for the recursive field".
Direct self-reference is the safe bounded first cut (the compiler's own structs never self-reference
by value — they use ptr — so self-build is unaffected). Indirect cycles = follow-up.

## m2 — calling a non-function → accept-then-SEGFAULT (STILL OPEN; attempt 1 over-rejected)
❌ **Attempt 1 (2026-07-18, REVERTED):** rejected when the ident-callee's payload is `SYM_VAR`/
`SYM_PARAM` (kinds 0/6) with a resolved `type_id` whose kind is not `TYPE_KIND_FUNC`. It rejected m2
correctly AND a simple `let f = add; f(..)` still worked — BUT the fixpoint gate caught an
OVER-REJECTION: compiling the compiler's OWN source, A rejected **5 valid call sites** with "value of
a non-function type is not callable" → A!=B (B build failed). So the compiler has ≥5 valid callable
forms whose var/param type is NOT directly `TYPE_KIND_FUNC` — likely higher-order-function PARAMS
(`fn foo(f: fn(i64)->i64)` then `f(x)`) or fn-pointers stored as a POINTER-to-func / a distinct
fn-type representation. NEXT attempt MUST first enumerate those 5 forms (temp-trace the rejected
call sites: print the callee sym kind + type_id + that type's entry.kind) and BROADEN the "callable"
predicate to include them (follow POINTER/REF to a func; accept whatever kind a `fn(...)->...` param
type actually has), THEN reject only genuinely non-callable scalar/struct/etc. values. Gate = A==B
(the over-rejection shows up as B-build-fails-on-compiler-source). Lower priority than it looked —
the callable-form enumeration is the real work.

## m2 (original report) — calling a non-function → accept-then-SEGFAULT
```
fn main() -> i64:
    let x = 5
    return x(3)     # x is i64, not callable
```
Built an exe; runtime **segfault (139)**. FIX: at NODE_CALL typecheck, if the callee resolves to a
value whose type is NOT a function/function-pointer type, REJECT "value of type 'i64' is not
callable". WATCH: genuine fn-pointer/closure values ARE callable (TYPE_KIND_FUNC) — exclude those.

## ✅ m4 — FIXED 2026-07-18 (A==B `B9F66834`, oracle t_matchnonsum, reject mode)
Was: `match an_i64: Some(v):` accepted → segfault (payload never binds). Now REJECTS: "cannot match
a value of a non-sum type against a variant pattern (Some/None/Ok/Err/...)". Fix in the NODE_MATCH_ARM
handler (typecheck.ax ~L2773): if the pattern is a `NODE_VARIANT_PAT` and the scrutinee's type kind
is not OPTION/RESULT/SUM/GENERIC_INST (the sum-like kinds), emit the diagnostic + bump diags_count.
Verified: scalar binding/int-literal match arms still allowed (`match x: 1: ... y: ...`), and valid
Option/Result/multi-field-variant/Tree-sum matches still work; self-build OK, regression green.

## m4 (original report) — match on a non-sum with variant patterns → accept-then-SEGFAULT
```
fn main() -> i64:
    let x = 5           # i64
    match x:
        Some(v): ...    # Some/None on a non-Option scrutinee
        None: ...
```
Built; runtime **segfault (139)**. FIX: in match typecheck, if the scrutinee type is not a
SUM/OPTION/RESULT (or a type whose variants include the arm patterns), REJECT "cannot match value of
type 'i64' against variant pattern 'Some'". Relatedly, a non-exhaustive/foreign-variant arm should
already reject (cf. t_nonexhenum / accept-then-miscompile cluster) — this is the non-sum scrutinee
gap.

## m5 — annotated let with an incompatible RHS type → accept-then-MISCOMPILE (garbage)
```
fn main() -> i64:
    let x: i64 = "hello"    # str assigned to i64
    return x                # garbage (52; `x + 100` gave 32, inconsistent => reinterpreted bytes)
```
Built; `x` is garbage (the str repr reinterpreted as i64). FIX: a `let name: T = rhs` must check the
RHS type is assignable to T. OVER-REJECTION RISK (why this is nuanced): the checker must still allow
the language's real coercions — int-literal → any int width, float-literal → f32/f64, and any other
intended implicit conversions. Only reject genuinely-incompatible pairs (str↔int, aggregate↔scalar,
etc.). Mirror the coercion predicate the assignment/return paths already use; do NOT hand-roll a new
one. This is the highest-value (common real typo) but also the highest over-rejection risk — gate
carefully (A==B + full regression; a too-strict check breaks self-build).

## NOT a bug
- m6 `Pair[i64](a:1, b:2)` on `struct Pair[A,B]` → returns p.a=1 CORRECTLY: B was inferred from the
  field `b:2`. Lenient (only one of two type args given) but correct. Leave it.

## Gate & priority
All FRONTEND (typecheck/struct-layout) → A==B. Each is a REJECT (adds a diagnostic + `diags_count`
bump so the driver halts before codegen — same mechanism as the existing rejects, e.g. variant-shadow
typecheck.ax:2157). Suggested order: m1 (compiler crash, bounded, near-zero over-rejection risk) →
m2 (bounded) → m4 (bounded) → m5 (highest value, needs the coercion predicate, gate hardest). Each
needs full regression + self-build to confirm no valid program (incl. the compiler's own source) is
newly rejected. Add a `reject`-mode oracle per fix (harness supports it, cf. t_uninferreject).
