---
name: bug-malformed-input-robustness-cluster
description: "Cluster (probe batch 9, 2026-07-18): 4 malformed programs the compiler mis-handled. m1 (self-recursive struct crash), m4 (match non-sum), m5 (str↔numeric let literal) all FIXED→reject. Only m2 (calling a non-fn → segfault) still OPEN (needs callable-form enumeration). All BUG#53-class: should REJECT with a diagnostic."
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

## m2 — calling a non-function → accept-then-SEGFAULT (OPEN — HARD, type-based reject is UNSOUND)
❌ **Attempt 1 (2026-07-18, REVERTED):** reject when ident-callee payload is `SYM_VAR`/`SYM_PARAM`
with a resolved type not `TYPE_KIND_FUNC` → over-rejected **5 valid compiler-source sites** → A!=B.

🔬 **Attempt 2 enumeration (2026-07-18, traced, then reverted — NO fix shipped):** instrumented the
NODE_CALL_EXPR fall-through (after the FUNC/STRUCT/SUM/GENERIC_INST/OPTION/RESULT/FIELD_EXPR elif
chain, `typecheck.ax` ~L3676) with a NON-fatal trace over every NODE_IDENT callee whose
`callee_type` isn't handled. Findings on the compiler's own source:
  - Exactly **5 sites**, all identical: `callee_type=4` (**i64**, `ekind=PRIMITIVE`, unwrap-through-
    ptr/ref still PRIMITIVE), `symkind=0` (**SYM_VAR**), attributed `name=len`.
  - BUT there is **no literal `len(` free-call anywhere** in source (bootstrap/stage1 or the inlined
    stdlib in tmp_concatenated) → the `name=len` payload is a **mis-attribution**: `infer_node(callee)`
    returns **i64 spuriously** for these valid callees (the SAME inference imprecision that bit m5,
    where infer_node returned i64 for str-returning CALLs). The callee.payload does not reliably
    identify the callee here.
  - Control: `let f = add; f(20,22)` → 42 and does NOT trip the trace → a fn-name-assigned fn-pointer
    is correctly typed `TYPE_KIND_FUNC` (handled at ~L3678). So fn-pointers are NOT the false-positive
    source; imprecise callee inference is.
**CONCLUSION: a type-kind-based "not callable" reject at typecheck is UNSOUND** — `callee_type` (from
`infer_node(callee)`) returns i64 for ≥5 valid call sites, indistinguishable from the genuine bug
`let x=5; x(3)` (also i64 SYM_VAR). Neither symbol-kind nor callee_type separates them. Also note the
TypeChecker struct has **no `tokens`/source field**, so pinning offsets needs extra plumbing.
**A real fix needs either** (a) a proper fn-pointer TYPE in the type system so callable values are
reliably kinded (RFC-level), or (b) hardening `infer_node`'s callee path so it never spuriously
returns i64 for a valid call (then a scalar-callee reject becomes safe). Both are larger than a
bounded bugfix and self-host-risky. **Deferred as a hard blocker** (needs the fn-ptr-type/inference
design). Priority: low — a rare hand-written typo, and the safe fix is a design change.

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

## ✅ m5 — FIXED 2026-07-18 (`6ab7c00`, A==B `F81E2A77`, oracle t_letstrmismatch, reject mode)
Was: `let x: i64 = "hello"` accepted → garbage (exit 52, str {ptr,len} repr read as int).
Now REJECTS: "type mismatch in `let`: a string and a numeric type are not compatible". Fix in the
NODE_VAR_DECL handler (typecheck.ax ~L2720): when an explicit annotation is present, reject if the
RHS **literal node kind** is NODE_STRING_LIT into a numeric-scalar T, or NODE_INT_LIT/NODE_FLOAT_LIT
into a str/bytes T. ⚠️ KEY LESSON: a first attempt gated on `inferred` (the RHS's inferred type)
OVER-REJECTED — infer_node reports **i64 for some str-returning CALLs**, so 3 valid `let s: str =
<call>` sites in the compiler's OWN source were rejected → A!=B (B build failed with 3 spurious
errors). Gating on the literal node kind (unambiguous type, no infer dependency) = zero over-rejection.
Regression 372/372, self-build green. LIMITATION: catches str↔numeric *literal* mismatches only; a
str *variable* assigned to a numeric annotation (`let x:i64 = some_str`) is not caught (would need a
reliable assignability predicate — infer_node too imprecise today). The reported bug (a literal) is
covered. **CLUSTER now: m1✅ m4✅ m5✅ fixed; only m2 OPEN** (call-non-fn, needs callable-form enum).

## m5 (original report) — annotated let with an incompatible RHS type → accept-then-MISCOMPILE (garbage)
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

## Second malformed-input pass — 2026-07-18 (7 bad programs, 3 new FIXES)
A fresh reject-vs-accept batch beyond the original m1–m6:
- ✅ **undefined name** (`return foo` where foo undeclared) → was accept+garbage, now REJECT
  "undefined name 'X'". Resolver forced payload=0 on unresolved + typecheck reject. A==B `D79B62DE`,
  oracle t_undefname. Full write-up [[bug-undefined-name-accept]]. (Cleanly rejectable unlike m2 —
  the RESOLVER with live scopes is the reliable oracle.)
- ✅ **return-type str↔numeric mismatch** (`return "hello"` from `-> i64`) → was accept+garbage, now
  REJECT. Literal-gated (return-path analog of m5). A==B `20DCA173`, oracle t_retmismatch.
- ✅ **binop str↔numeric mismatch** (`"hi" + 5`) → was pointer+int garbage, now REJECT. Both-definite
  str+numeric; str+str concat & numeric+numeric untouched. A==B `6F638209`, oracle t_strnumop.
- **NOT bugs:** missing struct field (`P(x:5)` when P has x,y) → fields ZERO-FILLED (verified 0 even
  after stack noise), safe partial-init (Go-like), leave it. Assign-to-immutable already REJECTS.
  Wrong generic arg count `Pair[i64]` → B inferred from field (lenient-correct, = m6). Duplicate
  struct-literal field (`P(x:5,x:7)`) → accepted first-wins; minor typo, low value, LEFT open.

## Gate & priority
All FRONTEND (typecheck/struct-layout) → A==B. Each is a REJECT (adds a diagnostic + `diags_count`
bump so the driver halts before codegen — same mechanism as the existing rejects, e.g. variant-shadow
typecheck.ax:2157). Suggested order: m1 (compiler crash, bounded, near-zero over-rejection risk) →
m2 (bounded) → m4 (bounded) → m5 (highest value, needs the coercion predicate, gate hardest). Each
needs full regression + self-build to confirm no valid program (incl. the compiler's own source) is
newly rejected. Add a `reject`-mode oracle per fix (harness supports it, cf. t_uninferreject).
