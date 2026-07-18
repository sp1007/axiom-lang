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

## Batch 4 (arithmetic/bitwise) — found the ONE real bug this session
A later arithmetic/bitwise batch (bitwise/shift/signed-div-mod/narrow-int-wrap/comparison-chain)
DID surface a real backend bug: **narrow-int const-fold wrap** ([[bug-const-fold-narrow-int-wrap]],
fixed `633913E9`). Two testing-pitfall reminders it also re-taught (keep oracles simple):
- **8-bit exit truncation** (documented): oracles 462/512/1111 read back as 206/0/87 — always keep
  oracle exit codes < 256.
- **bash clamps NEGATIVE Windows exit codes to 127**: `return -2` shows `$?`==127 in bash (looks like
  a bug!) but the RAW OS code is -2 (correct — verified via PowerShell `$LASTEXITCODE`). Signed div/mod
  and negative returns are all CORRECT. **Keep probe oracles NON-NEGATIVE**, or read the raw code via
  PowerShell. Chased a phantom "a3 signed-arith bug" (127) before realizing it was this clamp.

**Conclusion:** yield is ~zero for VALID-feature crosses (compiler mature); the arithmetic batch shows
edge-value/overflow crosses are the more productive probe angle. Per the bug-probe skill this
is the signal to STOP probing and re-plan from milestones. Remaining backlog is design-level:
m2 (fn-ptr type / callee-inference hardening — [[bug-malformed-input-robustness-cluster]]) and the
larger items in [[backlog-open-items]] (need user direction). Don't redo these crosses next session.
