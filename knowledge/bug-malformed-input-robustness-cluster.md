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

## m1 — self-recursive struct by value → COMPILER CRASHES (worst; highest priority)
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

## m2 — calling a non-function → accept-then-SEGFAULT
```
fn main() -> i64:
    let x = 5
    return x(3)     # x is i64, not callable
```
Built an exe; runtime **segfault (139)**. FIX: at NODE_CALL typecheck, if the callee resolves to a
value whose type is NOT a function/function-pointer type, REJECT "value of type 'i64' is not
callable". WATCH: genuine fn-pointer/closure values ARE callable (TYPE_KIND_FUNC) — exclude those.

## m4 — match on a non-sum with variant patterns → accept-then-SEGFAULT
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
