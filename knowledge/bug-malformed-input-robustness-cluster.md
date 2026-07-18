---
name: bug-malformed-input-robustness-cluster
description: "Cluster (probe batch 9, 2026-07-18): 4 malformed programs the compiler mis-handled. FIXED→reject: m1 (self-recursive struct crash), m2 (calling a non-fn value → segfault), m4 (match non-sum), m5 (str↔numeric let literal). One RARE residual m2b OPEN: a local var SHADOWING a fn, then called, still segfaults (guard spares same-named fns). All BUG#53-class: REJECT with a diagnostic."
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

## ✅ m2 — FIXED 2026-07-18 (A==B `EC01FECA`, 398/398, oracle t_callnonfn, reject mode)
Calling a plain value variable `x(3)` (x: i64) now REJECTS: `error: '<name>' is not a function;
it cannot be called`. Fix = a new `elif self.tree.nodes.data[callee].kind == NODE_IDENT` branch in
the NODE_CALL_EXPR `if callee_type != TYPE_UNKNOWN` elif-chain (typecheck.ax ~L3886, AFTER
FUNC/STRUCT/SUM/GENERIC_INST/OPTION/RESULT), gated: (1) callee symbol kind ∈ {SYM_VAR, SYM_PARAM,
SYM_CONST} (a value, not a type/ctor), AND (2) **RECEIVER-AGNOSTIC guard** — NO `SYM_FUNC` of that
name exists anywhere (scan `symbols` with `name_matches_method`). The guard is what made attempt-1
land: the 5 false-positive sites all attribute to `len`, a bare ident that resolves to a struct
`.len` FIELD (also SYM_VAR) whose name coincides with the stdlib `fn len(s: str)` — since a function
`len` exists, those defer to normal resolution untouched. `x(3)` (no `fn x` anywhere) is rejected.
A genuine fn pointer types as TYPE_KIND_FUNC and is handled in the branch above, so first-class
fn-values are unaffected. Self-build-safe (A==B). **CLUSTER CLOSED** (m1/m2/m4/m5 all fixed).

### 🟡 m2b — PARTIAL FIX shipped (arity-mismatch shadow REJECTS; arity-MATCH shadow still open)
Boundary probe after the m2 fix surfaced an adjacent silent SIGSEGV. A local variable SHADOWING a
same-named function, then called:
```
fn helper() -> i64: return 9
fn main() -> i32:
    let helper: i64 = 5      # local shadows the fn
    return helper(0) as i32  # helper takes 0 args, called with 1
```
**Trace nailed the mechanism (2026-07-18, build+trace of the compiler source):** the receiver-
agnostic guard's 5 false-positives are `name=len symkind=SYM_VAR declkind=13 (NODE_VAR_DECL)` — i.e.
the compiler's own `let len = std.string.len(s)` LOCALS. On the native path the qualified
`std.string.len` collapses to a bare `len` IDENT whose typecheck-resolved callee is the local, yet
air_builder RE-BINDS it to `fn len(s: str)` by name+args and it works. So `decl_node` kind does NOT
separate them (both NODE_VAR_DECL) — the real discriminator is **arity**: the 5 `len` sites are
1-arg and `fn len` is 1-param.
**SHIPPED (arity-aware guard):** reject only when NO same-named `SYM_FUNC` has a matching param
count (`fpc == argc` or `fpc == argc+1` for a self-method form). `helper(0)` (1 arg vs 0-param fn)
→ REJECT ✓; the 5 `len` sites (1 arg / 1-param) → spared, A==B `22DA5200`, 399/399. Oracle
`t_shadowcall`.
**STILL OPEN (deeper, rarer):** an arity-MATCHING shadow — `fn twice(n): …; let twice: i64 = 5;
twice(20)` — is still accept-then-SIGSEGV: the arity guard conservatively spares it (a fn `twice`
with matching arity exists), but air_builder here binds to the LOCAL (segfault), NOT the function
(unlike the qualified-collapse `len` case which binds to the fn). The clean fix needs typecheck to
know whether the callee was ORIGINALLY a qualified name that collapsed (bind to fn) vs a genuine
bare shadow (reject) — info partly lost after import-strip. This is the "infer_node/resolver
hardening" the m2 note always flagged as deep. LOW priority (deliberate fn-name/local collision
with matching arity — very rare). Do not chase without mapping air_builder's bare-call binding. LOW priority: requires a deliberate
fn-name/local-name collision — very rare in real code. Probe-banked repro `/tmp/pa_shadow.ax`.
Boundary CONFIRMED elsewhere: const-call `K(1)` and param-call `n(2)` (no same-named fn) both
REJECT correctly; single-primitive alias `type UserId = i64` is NOT mis-rejected (separate
narrowing, commit 66a3f48).

## (historical) m2 — the two failed attempts before the receiver-agnostic guard
❌ **Attempt 1 (2026-07-18, REVERTED):** reject when ident-callee payload is `SYM_VAR`/`SYM_PARAM`
with a resolved type not `TYPE_KIND_FUNC` → over-rejected **5 valid compiler-source sites** → A!=B.
(Root of the false positives: `len` field vs `fn len` — the receiver-agnostic function-existence
guard added in the 2026-07-18 fix is exactly what distinguishes them.)

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
**A real fix needs (b) hardening `infer_node`'s callee path** so it never spuriously returns i64 for
a valid call (then a scalar-callee reject becomes safe). ⚠️ **CORRECTION 2026-07-18:** the earlier
"option (a) = add a proper fn-pointer TYPE (RFC-level)" is a DEAD END — fn-pointers are ALREADY
shipped (BUG#49, `964fcba`, [[next-step-16-fnptr-shipped]]) and are correctly kinded `TYPE_KIND_FUNC`
(the `let f=add; f(20,22)` control proves it). So there is NO fn-ptr-type design gap and NO RFC to
write; do NOT draft a fn-pointer RFC (nearly done in error this session). The ONLY viable path is (b):
find WHY `infer_node(callee)` returns i64 for the 5 valid sites (mis-attributed `name=len`) and make
that path return the true type / a reliable "is-callable" answer — likely at the RESOLVER (live
scopes were the reliable oracle for [[bug-undefined-name-accept]]) rather than typecheck's imprecise
infer_node. This is an implementation-hardening task (needs build + A==B gate + full regression;
self-host-risky), NOT a design change. **Deferred** — priority low (rare hand-written typo `let x=5;
x(3)`), blocked on the infer_node/resolver hardening + a quiet box for the gate.

🎯 **Concrete fix strategy (2026-07-18 static read, NOT yet built — start the quiet-box attempt HERE):**
Verified in code, so the next attempt skips re-deriving:
- **Imprecision root CONFIRMED:** `typecheck.ax:4234-4235` — for a `NODE_IDENT`, infer_node does
  `result_type = symbols[node.payload].type_id`. For the 5 valid sites `node.payload` is a
  *mis-resolved* `SYM_VAR` (type i64), so ANY reject keyed on `infer_node(callee)` / `callee.payload`
  over-rejects. Do NOT key the reject on infer_node(callee) — that path is fundamentally the wrong
  oracle here.
- **Authoritative callability oracle is the CALL handler's OWN dispatch, not the callee's payload:**
  free calls resolve via `resolve_free_call_overload(callee.payload, first_arg, argc)` at
  `typecheck.ax:3353` (returns the real `SYM_FUNC` idx `ov_pick`); method calls resolve via the
  `NODE_FIELD_EXPR` path (`~3368`, flag 2048). The genuine bug `let x=5; x(3)` is the case where the
  callee is a `NODE_IDENT` and NEITHER (a) `resolve_free_call_overload` yields a valid `SYM_FUNC`
  NOR (b) the bound symbol's own type_id is `TYPE_KIND_FUNC` (a real fn-ptr value like `let f=add`).
- **Proposed predicate (reject non-callable):** at the NODE_CALL_EXPR handler, for a `NODE_IDENT`
  callee, reject ONLY when: `resolve_free_call_overload(...) == 0`/not-a-SYM_FUNC **AND** the callee's
  bound symbol is a `SYM_VAR`/`SYM_PARAM` whose `type_id` kind is a scalar PRIMITIVE (not
  `TYPE_KIND_FUNC`, not a ptr/ref-to-func). The "no SYM_FUNC of this name resolves" clause is exactly
  what attempt-1 was missing → its over-rejection. VERIFY the 5 sites: instrument
  `resolve_free_call_overload`'s return for them — hypothesis is it DOES resolve them to a real
  SYM_FUNC (so they'd be correctly spared), and only `x(3)` fails all avenues.
- **Gate:** frontend reject → A==B + full regression + confirm the 5 compiler-source sites still
  build (they were the canaries). Oracle `t_callnonfn` (reject mode). Self-host-risky.

  ❌ **CORRECTION 2026-07-18 (static read of resolve_free_call_overload, typecheck.ax:990-1015):**
  the proposed predicate above is WRONG — `resolve_free_call_overload` is ONLY an overload
  disambiguator, not a callability oracle. If the callee symbol has no `next_overload` (line 993)
  it returns `sym_idx` **unchanged**, regardless of whether that symbol is a SYM_FUNC or a SYM_VAR.
  So for `let x=5; x(3)` it returns the SYM_VAR untouched — it cannot distinguish callable from
  non-callable. Do NOT key the reject on it.
  🔬 **Refined open question for the NEXT attempt:** the 5 valid sites have callee.payload = a
  SYM_VAR (i64, mis-attributed `name=len`) yet compile CORRECTLY, so air_builder must dispatch them
  by some path that IGNORES the stale callee.payload (probably re-resolving the callee ident by NAME
  to a SYM_FUNC). The real discriminator is therefore "**does a SYM_FUNC with the callee ident's
  name exist/resolve in scope?**" (yes for the 5 → spare; no for `x` → reject). To implement safely
  you must FIRST instrument air_builder's NODE_CALL_EXPR dispatch to see exactly how the 5 sites
  resolve their function (by name? via a separate lookup table?), then replicate that lookup in the
  typecheck reject. This needs an instrumented build+trace cycle (now cheap again — gate is fast
  post-Defender-exclusion, see [[infra-defender-build-throttle]]), NOT more static reading.
  Priority stays LOW (rare hand-typo) — do only if a session has spare low-risk budget.

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

## Third malformed-input pass — 2026-07-18 (str↔numeric family closed + 3 OPEN candidates)
Extended the str↔numeric mismatch reject to ALL common contexts (each literal-gated on the operand's
str/int/float LITERAL node kind, so real coercions & value forms are untouched; all A==B, regression
green): **let** (m5), **return** (t_retmismatch), **arithmetic binop** `"a"+5` (t_strnumop),
**comparison** `s==5` (t_cmpstrnum), **call argument** `f("x")` (t_argstrnum), **assignment**
`x="s"` (t_assignstrnum). Also **index-a-scalar-variable** `x[0]` where x:i64 → reject (t_scalaridx,
GATED on collection==NODE_IDENT — infer_node returns i64 imprecisely for CALL collections, m2/m5 trap).

### ✅ ALL candidates FIXED — str↔numeric mismatch family CLOSED across 8 paths
- ✅ **m3 — struct ctor field mismatch FIXED `cec104e`** (A==B `47BC1E70`, 393/393, oracle
  t_ctorfieldstrnum): `P(x: "hello")` where `P.x: i64` → reject. `check_ctor_field_str_num` walks
  fields+args lockstep (like try_instantiate_struct_ctor), gated same-tree + literal.
- ✅ **m2 — array literal mixed elements FIXED `b6359bd`** (A==B `F89A4EE7`, 392/392, oracle
  t_arrmixstrnum): `[1, 2, "three"]` → reject. Element-loop str↔numeric literal check.
- ✅ **m5agg — FIXED `78245e5`** (A==B `B9D819F6`, 391/391, oracle t_retagg): `return <aggregate
  ident>` (struct/sum/array/Option/Result) from a numeric-returning fn → reject. Gated on the return
  expr being a bare NODE_IDENT (reliable declared type). Only the RETURN path done; assign/let
  aggregate↔scalar left for a follow-up if a repro surfaces.
- NOT-bugs from this batch: `mut x:i64 = 5; x = 10` (valid), method-arg mismatch already parse/rejects.

## NON-bugs / design-tension (4th bad-input batch 2026-07-18) — do NOT re-probe
- **Non-bool `if`/`while` condition** (`if x` where x:i64; `if s` where s:str): ACCEPTED (truthy —
  nonzero/non-null = true). The spec task-doc (`AXIOM SPECIFICATION`.../`docs/tasks/p04-t06`:58) says
  "condition must be bool" (code 3011), BUT the compiler's OWN source uses **11 truthy `if <i64>`
  conditions** (verified by trace) — the implementation RELIES on truthy int conditions. Enforcing
  bool-only would break self-build; it needs migrating those 11 sites + is a design decision (bool-only
  vs C-like truthy). **Deferred — needs user/spec decision, not an autonomous change.** A NARROWER
  future option: reject only str/aggregate/Option conditions (never meaningfully truthy; the 11
  compiler sites are all int so it'd be self-build-safe) — catches `if <Option>` (forgot .is_some())
  and `if <str>`. ✅ **SHIPPED 2026-07-18 `216d0a4`** (A==B `1C2E3D6A`, 396th oracle t_condagg):
  helper `check_cond_type` rejects string/bytes or aggregate-kind (struct/sum/Option/Result/array)
  conditions at NODE_IF/ELIF/WHILE; scalar-truthy preserved so self-build stays safe. The bool-only
  full enforcement (migrating the 11 truthy-int sites) remains the deferred design decision above.
- **Literal out of range for its type** (`let x:u8 = 300` → 44; `let x:u8 = -5` → 251): ACCEPTED,
  DETERMINISTIC WRAP (two's-complement). Consistent with AXIOM's narrow-int wrap semantics (cf.
  [[bug-const-fold-narrow-int-wrap]]). Rust rejects these as a strictness choice; AXIOM wrapping is a
  valid alternative design. NOT a miscompile. Leave unless a strictness policy is decided.

## Gate & priority
All FRONTEND (typecheck/struct-layout) → A==B. Each is a REJECT (adds a diagnostic + `diags_count`
bump so the driver halts before codegen — same mechanism as the existing rejects, e.g. variant-shadow
typecheck.ax:2157). Suggested order: m1 (compiler crash, bounded, near-zero over-rejection risk) →
m2 (bounded) → m4 (bounded) → m5 (highest value, needs the coercion predicate, gate hardest). Each
needs full regression + self-build to confirm no valid program (incl. the compiler's own source) is
newly rejected. Add a `reject`-mode oracle per fix (harness supports it, cf. t_uninferreject).
