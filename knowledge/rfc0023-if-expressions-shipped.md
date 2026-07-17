---
name: rfc0023-if-expressions-shipped
description: "SHIPPED 4c3ee6b: RFC 0023 if/elif/else EXPRESSIONS (value-producing). `let x = if c: a else: b`, mandatory else, works in lambda bodies. NUD-only so statement-if untouched."
metadata: 
  node_type: memory
  type: project
  originSessionId: 1e902df0-dc42-4a78-ba53-850e359b6f93
---

**SHIPPED `4c3ee6b`** (fast fixpoint A==B `97AAA1A7…`, 290/290). RFC doc `89ecb80`
(`rfcs/0023-if-expressions.md`). Design/impl derived autonomously during autopilot.

**Feature:** `if C0: V0 (elif Ci: Vi)* else: VE` as a value-producing expression.
`else` MANDATORY. Each branch body = a single inline expression (not a block).

**Why NUD-only:** reached solely via the Pratt NUD for `TK_IF` in EXPRESSION position
([parser.ax parse_nud](../../../../../d--projects-compiler-Axiom/bootstrap/stage1/parser.ax)).
Statement-`if` is dispatched separately in `parse_stmt` (TK_IF → parse_if_stmt) BEFORE
parse_nud, so statement-if is untouched and there's zero grammar ambiguity. The inline
`if C: STMT` at statement start is still rejected (BUG#53) — orthogonal.

**Pipeline:**
- ast: `NODE_IF_EXPR = 71`, flat children `cond0, then0, [condi, theni,...], else`.
- typecheck (infer_node NODE_IF_EXPR): conds→bool; value branches inferred against an
  ANCHOR (caller's `expected` if concrete, else the else-branch type) so int literals
  coerce; a value branch whose concrete type ≠ anchor → REJECT "incompatible types"
  (BUG#53 no-silent-miscompile). Missing else → clean parse diagnostic.
- air (`lower_if_expr`): N-branch value diamond generalizing the RFC 0016 P3 short-circuit
  lowering — ONE result vreg `r` written (OP_COPY type=result_type) on every path, read at
  the merge block. Correct under CFG-aware liveness [[rfc0016-p2prime-cfg-liveness]].

**Verified:** scalar / str (16B inline) / by-address struct results, elif chains, nesting,
`return if…`, and the motivating lambda body `v.map(|n| if n>0: n else: 0)`. All O0==O1.
Compiler source uses NO if-expr → self-codegen unchanged → A==B (not A!=B), so the A==B
fixpoint is the gate; new backend path verified by oracles. Oracles t_ifexpr/elif/nest/
str/struct/lambda + t_ifexprnoelse/mismatch (reject).

**Branch-type guard (final `f82d3f7`, A==B `EFCA0F86…`, 292/292):** the value diamond writes
one fixed-width result slot, so plain-scalar/str/struct branch WIDTHS must match. Guard in
infer_node NODE_IF_EXPR:
- Box-repr branches (Option/Result/**user-sum**/generic-inst) → SKIP width check; their box
  lowering round-trips through the result slot. `if c: Some(x) else: None`=42, `if c: Circle(7)
  else: Empty`=42, `if c: Red(1) else: Blue(42)`=42 all WORK (O0==O1). Oracle t_ifexprusersum
  (accept=42, INDENTED match arms), t_ifexproptnone(42).
- plain primitive/str/struct → width-mismatch reject (`if c: 42 else: "no"` = 8 vs 16). Uses
  `builder_type_size_and_align` (already used in typecheck for global sizing — pure type query).

**⚠️ FALSE-ALARM LESSON (important):** commit `9162628` briefly REJECTED user-sum branches on a
MISDIAGNOSIS — a probe harness `build && run; echo $?` reported the build-REJECT exit code 1 of
an unrelated **inline match arm** (`Pat: stmt`, a separately-unsupported form — [[inline-match-arm-unsupported]])
as if it were a miscompiled RUN value. Reverted in `f82d3f7` after re-testing with a harness
that greps the build log for `error:` and only runs on a produced exe. **RULE: in probes,
NEVER conflate build-reject with a run value — check the log + that the exe exists BEFORE
running; don't reuse a stale exe path across cases.** `build` DOES reject inline match arms
correctly (same as dump-air); there was never a user-sum lowering bug.

**Deferred (RFC §3, niche):** block-valued if (last-expr-of-block), ternary `?:`. Related
[[feedback-ergonomics]] (kills mut+stmt-if boilerplate).
