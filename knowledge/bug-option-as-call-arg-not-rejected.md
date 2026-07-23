---
name: bug-option-as-call-arg-not-rejected
description: "OPEN bug (probe-found 2026-07-24e): passing an Option[T]/Result value where the inner type T is expected AT A CALL ARGUMENT is silently accepted-then-miscompiled (reads the box raw → garbage) instead of rejected 'expected T, found Option[T] (missing .unwrap()?)'. The arithmetic version was already rejected (313cb51); call-args are the uncovered sibling site."
metadata:
  node_type: memory
  type: project
---

## ✅ FIXED 2026-07-24e (ident-arg form) — driver `178da3ca`, 532/532, A==B==C
**Root (found only after LAYER 8 corrected the PowerShell tooling artifact):** an annotated
`Option[i64]` local resolves to a **SUM-kind** type (`TYPE_KIND_SUM`=6), NOT `TYPE_KIND_OPTION`(11) —
which is why every earlier reject (checking kind 11/12) silently missed it. And at the 4616 arg loop
`node_types[arg]` is UNSET (not "pre-coerced"), so the arg's true type must be read from its SYMBOL.
**Fix (typecheck.ax, in the per-arg BUG#53 loop ~4620):** for an IDENT arg whose param is a
PRIMITIVE scalar, read `symbols[arg.payload].type_id`; if its kind is OPTION/RESULT/**SUM**, reject
"argument is an Option/Result/sum value but the parameter expects a plain value (missing .unwrap()?)".
Passing an aggregate where a scalar is expected is always a type error → cannot over-reject (self-build
clean). Gate: A==B==C `178da3ca`, regression **532/532** (+`t_optargreject`), positives clean
(`f(o.unwrap())`, Option→Option param), t_gentree/t_rectreesum intact.
**REMAINING (follow-up): the CALL-RESULT-arg form** `f(v.get(0))` (o1/o2) is NOT yet caught — the
arg is a NODE_CALL_EXPR with no symbol, so it needs the callee's RETURN type (also SUM-kind for an
Option-returning fn). Add a parallel branch: for a non-ident arg, if `infer_node(anode,UNKNOWN)` (or
the callee return type) is OPTION/RESULT/SUM and the param is a primitive scalar, reject. Verify with
**bash grep** (not Select-String — see [[lesson-bash-grep-not-powershell-selectstring]]).

## (history) OPEN — Option/Result passed to a `T` param is accepted-then-miscompiled (should reject)
**Found by proactive probing on driver `03F808DE`, 2026-07-24e** (while probing recursive-sum
variants — the original `sum(v.get(i))` gave 127/O0 vs 112/O1, isolated to this).
```
fn f(x: i64) -> i64: return x + 1
fn main() -> i64:
    let o: Option[i64] = Some(41)
    return f(o)            // Option[i64] passed to an i64 param — SILENTLY ACCEPTED, returns garbage (9)
```
| repro | shape | got | want |
|---|---|---|---|
| o3 | `f(o)` Option[i64]→i64 param | 9 | REJECT |
| o1 | `f(v.get(0))` (get→Option[i64]) → i64 | 41 | REJECT |
| o2 | `g(v.get(0))` Option[P]→P (struct) | 8 | REJECT |
Also miscompiles at O0 (and the Vec/get form diverged O0=127 vs O1=112). The callee receives the
8-byte Option/Result BOX pointer reinterpreted as `T` → reads tag/pointer as the value → garbage.

## Scope / not-a-bug boundary
- Only when the PARAM type is a concrete `T` and the ARG is `Option[T]`/`Result[...]` (the exact
  "forgot `.unwrap()`" mistake). Passing `Option[T]` to an `Option[T]` param is correct (must NOT
  reject). Generic params (`fn f[U](x: U)`) accept anything — don't reject there.
- Correct code is unaffected: `f(o.unwrap())` works; `v.get(i).unwrap()` works (r3c=32).

## Distinct from the shipped Option-arith reject
`313cb51` [[bug-option-arith-miscompile-open]] rejects `opt + 1` (arithmetic on an un-unwrapped
Option). This is the sibling site: a CALL ARGUMENT. Same root class (an Option/Result value used
where its inner T is expected, no auto-unwrap in AXIOM), different site — the arg-vs-param coercion
check doesn't flag the Option→T mismatch.

## ⚠️ ATTEMPTED 2026-07-24e — reject did NOT fire; reverted. KEY finding for next time:
Added the reject in the per-arg BUG#53 loop (typecheck.ax ~4645, beside the str/num-literal + the
interface-conformance rejects) gated on `param kind ∉ {OPTION,RESULT,GENERIC,GENERIC_INST,INTERFACE}`
and `arg type kind ∈ {OPTION,RESULT}`, reading the arg type from `node_types[anode]` (fallback
`infer_node`). Result: **the reject never fired** (o1/o2/o3 still ran), while positives were clean
(ok1 `f(o.unwrap())`=42, ok2 Option→Option param=42) and the compiler self-built (no over-reject).
⭐ **ROOT: `node_types[anode]` for the arg is ALREADY COERCED to the param type** (i64), not the
Option — by the time this loop runs, an earlier inference/coercion pass has "helpfully" erased the
Option→i64 mismatch (which is EXACTLY the miscompile: the arg is silently treated as i64). So the
check saw i64, not Option, and `infer_node(anode)` re-derived the coerced type too (and had
side-effects — the exit codes shifted). Reverted (driver stays `03F808DE`).
**Next approach:** do NOT trust `node_types[anode]` here. Read the arg's TRUE type — for an IDENT arg,
look up its symbol's DECLARED `type_id` (`symbols[node.payload].type_id`); for a call arg
(`v.get(0)`), the callee's return type. OR, better, find the COERCION site that silently turns an
Option arg into the param type and make IT reject instead of coerce (grep the arg-coercion path that
runs before this loop — likely near `infer_node(arg, pt)` at typecheck.ax:568 or the generic-arg
coercion). That site has the pre-coercion type and is the true home for the diagnostic.

## ⚠️ ATTEMPT 2 (same session) — coercion-site reject ALSO did not fire; reverted
Added the inverse reject at the REGULAR-call arg-coercion loop (typecheck.ax ~5140): for a plain
concrete param (`not param_is_optlike/arr/tup`, param kind ∉ {GENERIC,GENERIC_INST,INTERFACE}),
reject when `at` (the arg type from `infer_node(arg, pexp)` with `pexp` UNKNOWN for a plain param)
is OPTION/RESULT. Rationale: at THIS site `pexp` is UNKNOWN for a plain param, so `at` should be the
TRUE type. Result: **still no reject — o3 unchanged (9)**, positives + t_gentree/q1 clean, self-build
clean. So `at` is NOT OPTION here either, OR the simple `f(o)` call never reaches this loop (there
may be an earlier/simpler resolved-free-call path that consumes args before line 5043).
⭐ **TWO obvious arg-check sites ruled out** (4645 downstream loop = pre-coerced; 5140 coercion loop
= arg type not visible / path maybe not reached). **The arg's Option type is masked/erased before any
arg-check the two loops see.** NEXT = INSTRUMENT: add a temporary trace print of `callee`, the taken
call-branch, and the arg's inferred type for `f(o)` (o3) to find WHERE the simple-ident Option arg is
processed and where its type becomes non-Option. Likely candidates: the overload-resolution path
(`resolve_free_call_overload`, ~848/1234) may bind + coerce before the 5043 loop; or an even earlier
NODE_CALL_EXPR fast path. Fix belongs wherever the arg's true Option type is still visible. This is a
trace-driven debugging task, not a blind edit — two blind edits have now failed identically.

## 🔬 TRACE DIAGNOSIS 2026-07-24e (4 instrumented iterations) — the reject can't live at any arg site
Instrumented the regular-call arg loop (typecheck.ax:5043) and compiled o3 (`f(o)`, `o:Option[i64]`,
`f(x:i64)`). Findings:
1. The loop IS reached (2065× during o3's build — stdlib calls dominate).
2. **Across ALL 2065 arg-checks, NO arg ever infers as Option/Result** (`atkind` is never 11/12).
   For an ident→i64-param call the trace shows `at=4 (i64), atkind=0` — so `infer_node(arg, UNKNOWN)`
   for `o` returns **i64, not Option[i64]**.
3. **NO ident arg has an Option/Result SYMBOL type** either (a guard printing only when
   `symbols[arg.payload].type_id` is OPTION/RESULT never fired for o3).
⇒ By the time ANY call-arg check runs, `o`'s Option-ness is GONE — not in the inferred arg type, not
in the symbol type as seen here. So NO arg-site reject can work; both earlier edits were doomed.
**Two live hypotheses for the fresh session (need one more targeted trace of the `f(o)` call
specifically — print the CALLEE name + the o-symbol's type_id at its decl):**
(a) `let o: Option[i64] = Some(41)` mis-types the SYMBOL `o` as i64 (the annotation/`Some` ctor path
    collapses Option→inner at the let-binding) — then everything downstream is consistently i64 and
    the 16B box is read as i64 = the miscompile. If so, the fix is at the let-binding type, and it may
    even make the value CORRECT (not just rejected). CHECK FIRST: does `let o: Option[i64] = Some(41)`
    give `o` an Option symbol type? (probe `o.is_some()` / dump the symbol.)
(b) `f(o)` reaches a DIFFERENT call path (overload resolution binds it) that unwraps before 5043.
The o3 traces are dominated by stdlib; isolate `f(o)` by printing the callee name (`fnm`) so the ONE
line for the user's `f` is findable. Do NOT edit blind again — 3 blind/site edits have failed
identically. This is now a let-binding-vs-inference question, answerable with one more scoped trace.

## ✅ ROOT LOCATION DECIDED (no more blind edits needed at arg-loop): the FREE-CALL OVERLOAD path
Cross-referencing the trace with the `ok2` positive settles it: `ok2` = `takesopt(o)` with
`takesopt(o: Option[i64])` returns 42, so `o`'s SYMBOL genuinely IS `Option[i64]` (it matches as an
Option). Yet `OPTIONSYM` never fired for `f(o)` in o3 → therefore **`f(o)` never reaches the 5043
arg-coercion loop at all**; the simple resolved free-call is bound by the OVERLOAD-RESOLUTION path
(`resolve_free_call_overload` / the free-call handling ~848–1360, and `param_arg_match`), which
accepts the `Option[i64]`→`i64` arg without an unwrap check. **THE FIX BELONGS THERE**: in the
free-call arg-vs-param matching, when a param is a concrete non-Option/Result/generic type and the
arg's resolved (symbol/callee-return) type is OPTION/RESULT, reject "missing .unwrap()". Trace the ONE
`f(o)` call through `param_arg_match` / the overload binder to place it exactly. This is a scoped,
now-well-targeted fix — the arg-coercion loop (5043) and the downstream validation loop (4645) are
both DEAD ENDS (proven: f(o) doesn't reach 5043; arg is pre-coerced at 4645).

## 🔬 LAYER 5 (2026-07-24e) — `param_arg_match` is overload-only; single-fn `f(o)` has NO arg check
Read `param_arg_match` (typecheck.ax:1239) and `resolve_free_call_overload` (1250): param_arg_match
is used ONLY to disambiguate OVERLOADS, and `resolve_free_call_overload` returns early at **:1253**
(`if next_overload == 0: return sym_idx`) for a NON-overloaded function like `f`. So for a single `f`,
param_arg_match is never called, and the overload path does no arg-type validation. Combined with the
earlier layers (f(o) doesn't reach the 5043 arg-coercion loop; the 4645 loop is pre-coerced), the
conclusion is: **a simple single-function call `f(o)` gets essentially NO arg-vs-param type validation
for the Option→T case** — the NODE_CALL_EXPR handler infers the result type and lowers, and nothing
rejects the mismatch. That is the accept-then-miscompile.
**Therefore the fix must ADD an arg-vs-param check on the general call path** (not piggyback an
existing one). Best next step for the dedicated session: instrument the NODE_CALL_EXPR handler to
print the CALLEE NAME (`fnm`) for every processed call while building o3, find the ONE `f` line, and
see which branch handles its args (the FUNC branch at ~4595 that contains the 5043 loop DOES run for
resolved funcs — so either f(o)'s `callee_type` isn't FUNC there, or o's node_type is already i64 by
then). The check itself: in that branch, for each arg vs `fi.params.data[i]`, if the param is a
concrete non-Option/Result/generic/interface type and the arg's TRUE type (symbol decl type for an
ident; callee return type for a call) is OPTION/RESULT → reject. Must read the TRUE type, since
node_types is pre-coerced. Over-reject caught by the 531 regression + fixpoint. **This is a
dedicated, uninterrupted debug session — five tick-driven layers have peeled the onion but the exact
branch that consumes f(o)'s args (and where the Option type vanishes) still needs one callee-name
trace to nail. Do it in one focused sitting, not fragments.**

## 🔬 LAYER 6 — f(o) reaches NONE of the 4 obvious sites; needs a NODE_CALL_EXPR-entry trace
A callee-name-filtered trace (`fnm == "f"/"g"`) at the FUNC-branch arg loop (4616) **never fired**
for o3/o2 — so `f(o)`/`g(v.get(0))` do NOT reach that loop either. Net, FOUR sites are now
definitively ruled out (with trace evidence): (1) downstream arg loop 4645 (pre-coerced),
(2) arg-coercion loop 5043 (never reached — no OPTIONSYM), (3) FUNC-branch loop 4616 (never reached
— no FDBG), (4) param_arg_match (overload-only, single fn returns early at 1253). ⇒ **The call's
arg-vs-param validation for a simple `f(o)` happens on NONE of the obvious paths** — possibly it is
simply never validated (the miscompile), and the arg's node_type is set to i64 somewhere in ident
inference / the let-binding before any call check.
**START HERE next session (systematic, not guess-and-check):** add a trace at the ENTRY of the
`NODE_CALL_EXPR` case in `infer_node` printing the callee name (`intern.get(symbols[callee.payload].
name_id)`) + `callee_type` + `entry.kind` for EVERY call; build o3; find the `f` line; that reveals
which branch consumes it (or that it falls through unvalidated). THEN, separately, trace the
LET-BINDING of `o` (`let o: Option[i64] = Some(41)`) to confirm `o`'s symbol type_id is Option (ok2
implies yes) vs i64. The fix is wherever the Option→i64 identity is first accepted. Six tick-driven
layers have RULED OUT the easy answers; the remaining work is a single systematic sitting.

## 🔬 LAYER 7 — handler mapped: infer_node:4089 (2336 is `stmt_terminates`, a red herring)
The real NODE_CALL_EXPR type-checker is `infer_node` at **typecheck.ax:4089**. For `f(o)`: callee is
NODE_IDENT (4096) → `resolve_free_call_overload` returns `f` unchanged (single fn) →
`is_generic_call=false` (4113, `f` non-generic). Since the 5043 arg-coercion loop is gated on `not
is_generic_call`, `f(o)` SHOULD reach it — yet the OPTIONSYM guard never fired there. So the narrow
remaining mystery is ONE of: (a) `callee_type` for `f` is not resolved to `TYPE_KIND_FUNC` at the
5037 check (so `fp_data` stays null → the arg-type checks are skipped entirely), or (b) `o`'s
node/symbol type is already i64 by the time the loop reads it. **ONE trace from 4089 resolves it:**
print, for `fnm=="f"`, the `callee_type` + `entries[callee_type].kind` right where the FUNC branch is
decided (~4560-4595), and the arg's symbol type_id. If callee_type isn't FUNC → the whole per-arg
validation is skipped for simple free calls = the bug's home (add the check + the FUNC dispatch there,
or reject when a non-generic ident callee resolves to FUNC but fp check is bypassed). This is now a
~10-minute focused trace, fully set up. STOP fragmenting across ticks.

## ⚠️⚠️ LAYER 8 — META-ERROR FOUND: earlier "didn't fire" traces were POWERSHELL ARTIFACTS
CRITICAL: reading native trace output through PowerShell `Select-String` (`... 2>&1 | Select-String
"TAG"`) **silently drops matches** — it reported ZERO hits for traces that a **bash `grep`** showed
firing 14×. So several LAYER 2–7 conclusions ("f(o) doesn't reach the 5043 loop", "OPTIONSYM never
fired", "FUNC branch not entered") are **UNRELIABLE / likely FALSE**. ⭐ **ALWAYS read compiler trace
output with `bash ... | grep`, NEVER PowerShell Select-String.**
**Re-established on solid (bash-grep) ground:** (1) `f`'s `callee_type` **IS** `TYPE_KIND_FUNC`
(kind=2, `CALLTRACE`), so the FUNC branch at typecheck.ax:4568 **IS** entered for `f(o)`; (2) the
per-arg loop at ~4616 **IS** reached. So the fix DOES belong at that loop (LAYER-1 site was right;
layers 2–7 chased the artifact). REMAINING (clean re-trace, bash grep only): at the 4616 loop, an
`ARGDBG` for `fnm=="f"` showed `nodetype=0` (node_types UNSET there — so it is NOT "pre-coerced to
i64" as LAYER 1 guessed; it's simply unresolved at that point) and an ambiguous `symtype/symkind` read
(needs the arg node's payload→symbol mapping re-derived cleanly — the `pt=4` i64-param line is the
`f(o)` one; confirm whether its arg symbol type is the Option). **NEXT SESSION (bash grep, ~20 min):**
at the 4616 loop, for the arg, resolve its TRUE type via `infer_node(anode, TYPE_UNKNOWN)` freshly (or
the arg-ident symbol's decl type) and print it with bash-grep verification; when the param is a plain
concrete type and the arg's true type is OPTION/RESULT, reject. The machinery to reject is trivial
(sits beside the str/num-literal reject at 4625) — only the arg's-true-type read must be gotten right,
and it must be VERIFIED via bash grep, not Select-String.

## Fix direction (dedicated gated session — original notes)
In the call-argument type check (typecheck.ax, where each arg type is checked against the param
type — near the existing width/variant-arg checks, cf. the `try_instantiate_variant_call` /
arg-coercion path), add: if the param type is a concrete non-generic `T` and the arg's resolved
type is `TYPE_KIND_OPTION`/`TYPE_KIND_RESULT` whose inner == `T` (or simply kind OPTION/RESULT while
the param is a non-Option/Result concrete type), emit `error: expected <T>, found Option/Result
(missing .unwrap()?)` and count a diag. Mirror the arith reject's precision to avoid over-reject:
only fire when the param is concrete + non-generic + NOT itself an Option/Result. Gate: A==B
(frontend reject-only; inert on self-host unless the compiler does this — it shouldn't) + full
regression (watch for over-reject on any stdlib/compiler call passing an Option to an Option param)
+ oracles `t_optargreject` (reject) and a positive `t_optargok` (`.unwrap()` still compiles+runs).
Repro `/tmp/probe3/o1.ax`,`o2.ax`,`o3.ax`.
