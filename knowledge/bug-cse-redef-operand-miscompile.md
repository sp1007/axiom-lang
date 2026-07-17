---
name: bug-cse-redef-operand-miscompile
description: "FIXED: CSE reused a common subexpression across a redefinition of its operand registers → silent -O1 miscompile of `x = x op x` (e.g. pow_i64 returned 0)."
metadata: 
  node_type: memory
  type: project
  originSessionId: 36f1f15d-9c5e-4868-8984-73755a93e817
---

✅ **FIXED `f65addb`** (A==B AND hand-built **B==C `338b9d3c`**, 175/175). Core optimizer correctness bug in `cse_func` (`bootstrap/stage1/ssa_opt.ax:1436`).

**Symptom:** at **-O1 only** (-O0 correct), repeated straight-line self-assignment `b = b * b; b = b * b` left `b` stuck after the FIRST op (2→4→4→4 instead of 4→16→256). Broke any `x = x <op> x`. Found via std.math `pow_i64(2,10)` returning 0 during a probe.

**Root cause:** `cse_func` matched common subexpressions on `(opcode, src1, src2)` ALONE, assuming strict SSA. **AIR is NOT strict SSA** — a mutable local reuses ONE register — so `b=b*b; b=b*b` lowers to `t1=MUL b,b; b=COPY t1; t2=MUL b,b; b=COPY t2`. The two MULs are register-identical even though `b` was reassigned between them; CSE rewrote the 2nd MUL to `COPY t1` (stale) → 2nd squaring became a no-op.

**Fix:** scan CSE candidates DOWNWARD from the current inst and STOP at the first inst that redefines `src1` or `src2` (redef check runs BEFORE the match test, so a self-referential `b = <op> b` overwriting its own source is never reused). Legitimate CSE (same expr, no intervening redef) still fires — verified `a*c` twice = 30.

**LESSON:** any optimizer keyed on register identity (CSE/GVN/LICM/copy-prop) MUST account for the non-SSA mutable-local register reuse in AIR — an operand can be redefined between two textually-identical expressions. Probe pattern that catches it: `x = x op x` repeated straight-line at -O1 vs -O0. Oracle `t_cseredef(12)`. This bug class had escaped ~27+ prior feature-combo probes (they didn't repeat self-ops in a straight line).
