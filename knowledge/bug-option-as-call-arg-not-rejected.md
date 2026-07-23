---
name: bug-option-as-call-arg-not-rejected
description: "OPEN bug (probe-found 2026-07-24e): passing an Option[T]/Result value where the inner type T is expected AT A CALL ARGUMENT is silently accepted-then-miscompiled (reads the box raw → garbage) instead of rejected 'expected T, found Option[T] (missing .unwrap()?)'. The arithmetic version was already rejected (313cb51); call-args are the uncovered sibling site."
metadata:
  node_type: memory
  type: project
---

## OPEN — Option/Result passed to a `T` param is accepted-then-miscompiled (should reject)
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
