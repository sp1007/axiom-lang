---
name: probe-batch-clean-2026-07-18-b
description: "Probe batch (2026-07-18, post condition-reject fix 216d0a4): 6 feature-crosses all CLEAN, O0==O1. Confirms mature plateau — don't re-probe these."
metadata:
  node_type: memory
  type: project
---

# Probe batch 2026-07-18-b — all CLEAN (driver `1C2E3D6A`)

Ran after shipping the string/aggregate/Option condition-reject fix (`216d0a4`, 395/395).
Backlog scan: every file tagged "OPEN" in `knowledge/` is actually FIXED (stale
descriptions, bodies confirm closure) — **no genuinely-open confirmed bugs**. So refilled
via proactive probing.

Six feature-crosses, built + run at **-O0 AND -O1**, all matched oracle, **O0==O1**:
1. single-field sum `match` with payload extraction (Circle/Square/Nil) → 81
2. unsigned bitwise on boundary `u32::MAX >> 26 & 63` → 42
3. nested `for i in 0..4` × `continue` skipping diagonal → 90
4. 15 simultaneously-live temps (register spill) sum → 120
5. two-call recursion `fib(11)` → 89
6. signed→u8 wraparound round-trip `(-56 as u8) as i64` → 100

**CONCLUSION: compiler robust; plateau holds.** Combined with the prior clean sweeps
([[probe-batch-1012-clean-2026-07-18]], the 8-area sweep in STATE), the accessible
silent-miscompile veins are exhausted. Remaining backlog is design-level pending user
direction (m2 fn-pointer type in the type system = RFC; milestones M4 compliance suite,
M6 ELF export/perf gate/Mach-O). Don't re-probe the crosses above.

Pitfall re-confirmed: match arms use **bare patterns** (`Wrap(x):`), NOT `case`;
multi-field variant `Rect(i64,i64)` is REJECTED by design (BUG#81, backlog); loops are
`for i in a..b:` with `mut x: T = v` (no `let`) for mutables.
