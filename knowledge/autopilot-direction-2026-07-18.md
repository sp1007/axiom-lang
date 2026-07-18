---
name: autopilot-direction-2026-07-18
description: "User-chosen strategic direction (2026-07-18) for autonomous execution: M4=rewrite compliance suite in real grammar, next milestone=M6 perf, CTGC P3=attempt with tight gate. Autopilot is authorized to pick+execute without further confirmation."
metadata:
  type: project
---

User (2026-07-18) surfaced the 3 user-gated design decisions once, chose, and granted
standing autonomy ("bạn chỉ việc lựa chọn phương án tối ưu và thực hiện, không cần chờ tôi quyết định nữa").

## Decisions
1. **M4 (compliance) — REWRITE suite in real grammar.** `tests/axiom_compliance_suite.ax`
   is aspirational-dialect (=> match-arm, impl Trait, std.gpu/quantum/net, .length()) that
   the impl does NOT parse. Do NOT implement the spec dialect. Instead rewrite the suite using
   the REAL grammar (match `Pattern:`, `.len`, structural methods, block strings, generics/HOF
   already shipped). Redefine M4 = "compliance with the real language surface". Target the core
   groups (~60 tests) exercising shipped features. See [[m4-compliance-suite-spec-vs-impl-gap]].
2. **Next big milestone = M6 perf.** After M4, autonomous loop pursues Fib 2.44x → ≤1.05x vs
   clang -O2: recursion→loop transform + inlining, benchmark-gated (§10), reversible/isolated.
   Chosen over macOS/Mach-O and async runtime. See [[m6-perf-gate-fib-benchmark]].
3. **RFC 0015 P3 (CTGC-free) + RFC 0014 drop-glue = ATTEMPT with tight gate** (dedicated session):
   borrow-edge tracking (INDEX/FIELD-init = never-free) + module whitelist + **B==C fixpoint +
   full -O2 regression BEFORE every commit + revert-on-red, no partial commit** (UAF risk on
   self-host). See [[bug69-ctgc-ownership-escape-noop]] / [[backlog-open-items]].

## Order
M4 (immediate directive "thực hiện dứt điểm m4") → then M6 perf → CTGC P3 as a later dedicated
session. Between milestones, keep the loop alive per CLAUDE.md §24 (Monitor heartbeat, no hibernation).
