---
name: bug-const-shift-fold-open
description: "FIXED (fold disabled): `const << const` / `const >> const` miscompiled to 0 at -O1 + codegen crash. eval_binary now refuses to fold shifts. Micro-cause still open."
metadata: 
  node_type: memory
  type: project
  originSessionId: b9556362-e358-4881-8a42-df3ba54b80bb
---

✅ **`const << const` / `const >> const` -O1 miscompile FIXED** — 2026-07-11.
`641f0b1`, daily-driver A==B **`4448C745`**, 159/159. (Found bug-probe round 5.)

**Symptom.** A shift with BOTH operands compile-time int constants folded to a
garbage `iconst` at -O1 (whole expr → 0 → uninitialized-register garbage), and
crashed native codegen on some `>>`. -O0 + every non-both-const shift (var<<var,
var<<const, const<<var) were correct.

**Fix.** `ssa_opt.eval_binary` now `return false` for OP_SHL/OP_SHR → the const-fold
never fires; shifts lower at runtime (correct). Negligible lost optimization (folding
two literals). Oracle `t_constshift`(35). A==B (no compiler code relied on the fold).

✅ **CLOSED 2026-07-24 as correctness-resolved (fold-restoration DECLINED, by design).**
Re-verified on driver `23C1E261`: `5 << 1` + `160 >> 2` = 50 (correct at -O0/-O1). The bug's
CORRECTNESS is fully resolved by the disabled fold — shifts lower at runtime, always right.
The remaining "micro-cause" is purely about RE-ENABLING an optimization that folds two integer
LITERALS in a shift. Per CLAUDE.md §10 (do not prematurely optimize; correctness > cleverness),
restoring it is DECLINED: the failure mode was a SILENT wrong value, the saved work is folding
`5 << 1` → `10` (which real code writes as `10`), and ssa_opt is un-instrumentable by printf
(needs multi-rebuild dump-air bisection) — a bad risk/reward trade. Not reopening without a
concrete real-code motivation. Details of the micro-cause kept below for whoever revisits.

⚠️ **Micro-cause (root not pinned, fold merely disabled — kept for reference).** Deep probing
found: the fold produces `iconst` with `src1==0` (value 0) even though
`eval_binary(SHL,5,1)` reads as 10 and var-shifts work; disabling the fold cleanly
fixes it. So the defect is in how the folded shift constant is APPLIED/materialized
(fold_func apply ~239-250) or an interaction with a later pass — NOT the shift
arithmetic. NOTE: `ax_printf_local` debug prints inserted into `ssa_opt` did NOT
appear in stdout even null-stripped (debug-print path in that module seems inert) —
use pass-disabling BISECTION or `dump-air` at intermediate stages, not printf, to
trace ssa_opt next time. A proper fix would restore the fold with correct
apply/materialization.

Sibling from the same round-5 probe: float negation `-x` (BUG A) SHIPPED `d60c5dc`
[[bug-accept-then-miscompile-cluster-0711]].
