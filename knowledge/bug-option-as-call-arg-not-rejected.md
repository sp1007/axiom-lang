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

## Fix direction (dedicated gated session — NOT yet attempted)
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
