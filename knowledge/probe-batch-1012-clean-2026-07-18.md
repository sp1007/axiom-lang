---
name: probe-batch-1012-clean-2026-07-18
description: "Proactive bug-probe batches (2026-07-18, ~22 feature-cross programs across O0-O3) found ZERO real compiler bugs — every apparent mismatch was a test-authoring error (wrong comment char, wrong sum syntax, wrong cwd, miscalculated oracle). Compiler robust on these crosses. Banked t_optvecnest oracle. Signal to stop probing and re-plan toward milestones."
metadata:
  node_type: memory
  type: project
---

# Bug-probe session 2026-07-18 — 3 batches, 0 real bugs (compiler robust)

Ran after closing malformed-input m5 + deferring m2. ~22 small feature-cross programs, each with a
distinctive oracle exit code, built+run at -O0/-O1 (batch 3 also -O2/-O3). **No real bugs.** Every
initial "mismatch" was a MY-side test error, not a compiler fault:
- Used `#` for inline comments → `#` is `TK_HASH` (attribute syntax), **AXIOM comments are `//`**.
- Used `sum Name:` / `struct S: field: S` sum syntax → correct sum form is `type T = A(..) | B(..)`
  with direct recursion (compiler boxes it); recursive-by-value struct is (correctly) rejected.
- Inline match-arm `Some(v): stmt` → unsupported (known, [[inline-match-arm-unsupported]]); arm body
  must be an indented block.
- A batch loop `cd /tmp/probe` broke the relative `./bin/axc_native.exe` path → spurious NOEXE.
- p4 oracle miscalculated: `while i<5 and side(i)` with `side(0)` returning `0>0`==false → loop never
  runs (short-circuit correct). got=10 was RIGHT.

Feature crosses that all compiled+ran CORRECTLY (O0..O3 where tested):
- generic struct holding str/struct field, cross-fn read (Wrap[str], Box[Pair])
- array-of-Option iterated + matched; Option[struct] unwrap; **Option[Vec[i64]]** nested generic
  unwrap+iterate (banked as oracle **t_optvecnest**, exit 42)
- module-level `mut` global RMW across fns; Result[struct,str] Ok/Err match
- recursive `type Tree = Node(Tree,Tree) | Leaf(i64)` sum; nested `for i in a..b` × break/continue
- str concat in loop + `.len`; str `==` equality dispatch chain; u8 wraparound; f64 arith→i64 trunc
- tuple return + destructure; generic `dup[T]->(T,T)`; Pair[A,B] two-param; 16B struct return via branch
- optimizer stress: loop × fn-call × sum-of-squares stable across -O0/-O1/-O2/-O3

**Conclusion:** yield is ~zero (compiler mature at 373 regression tests). Per the bug-probe skill this
is the signal to STOP probing and re-plan from milestones. Remaining backlog is design-level:
m2 (fn-ptr type / callee-inference hardening — [[bug-malformed-input-robustness-cluster]]) and the
larger items in [[backlog-open-items]] (need user direction). Don't redo these crosses next session.
