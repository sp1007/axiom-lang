---
name: m6-perf-baseline
description: "M6 perf milestone (Fib(40) <=5% of clang) — the FOCUS after RFC 0015 closed. Reproducible baseline: AXIOM -O3 = 2.59x clang -O2 (i64-fair). Profiled the gap to 4 systemic codegen taxes; prioritized optimization backlog. Harness scripts/perf_fib.ps1."
metadata:
  node_type: memory
  type: project
---

**M6 perf gate = FOCUS (user-chosen 2026-07-24e).** Target: AXIOM within 5% of clang -O2 on
Fib(40). Reproducible harness: **`scripts/perf_fib.ps1`** (i64-vs-`int64_t`, best-of-7, pins exit
203, prints `M6_PERF_OK`/`M6_PERF_GAP`). Re-run it after EVERY codegen change to measure the delta.

## Baseline (2026-07-24e, driver `12DBE1D8`)
| build | Fib(40) best | vs clang |
|---|---|---|
| gcc -O2 | ~264 ms | 0.70x |
| clang -O2 | ~376 ms | 1.00x |
| **AXIOM -O3** | **~975 ms** | **2.59x (159% slower)** |
| AXIOM -O2 | ~1350 ms | 3.6x |
(Absolute ms drift with machine load; the RATIO is the gate. Earlier 1.89x figure used C `long`
= 32-bit on Windows = unfair; the i64-fair `int64_t` baseline is 2.59x.)

## Profiled root causes (objdump of the bundled binary — AXIOM has NO per-fn symbols, single
`.text` blob; compare against clang/gcc `fib` which DO have symbols). Ranked by expected ROI:
1. **Heavy unconditional prologue/epilogue.** Every AXIOM fn pushes up to 8 callee-saved regs
   (`rbp,rbx,rsi,rdi,r12,r13,r14,r15`) and pops them, regardless of how many it uses (~2). clang's
   `fib` prologue = `push rsi; push rdi; sub $0x28` (2 regs). On a fn called ~331M times this
   push/pop tax likely dominates. **Fix = save only callee-saved regs the fn actually allocates**
   (x86_regalloc / x86_emitter frame setup). Biggest, but touches regalloc/frame → careful B==C.
2. **Register-materialized immediates + redundant width-masks.** Pervasive pattern
   `mov $C,%rax; mov $0xff,%rcx; and %rcx,%rax` — (a) the mask is materialised into a scratch reg
   instead of using the immediate form `and $0xff,%dst`; (b) constants that already fit are masked
   anyway (`10 & 255`, `65 & 255` = dead). i32 literals also get `and $0xffffffff`. **Fix = peephole:
   emit `and $imm,%dst` (imm32) directly, and constant-fold `mov $C; and $M` → `mov $(C&M)`; drop the
   mask when the value provably fits the type width.** Localized emitter/selector peephole — the
   cleanest first win. Verify it's not narrowing-semantics-load-bearing before dropping.
3. **No `lea` for add/sub-with-constant.** `n - 1` = `mov $1; and ...; sub` instead of clang's
   `lea -0x1(%rdi),%ecx`. **Fix = selector recognises reg±smallconst → lea/dec/inc.**
4. **Excessive register shuffling / stack round-trips.** Long `mov` chains + spill/reload of live
   values across the call (visible in the syscall wrappers: 7 stores then 7 reloads back-to-back).
   `cmp` also materialises its constant instead of `cmp $imm,%reg`. **Fix = a copy-coalescing /
   dead-move peephole pass over machine IR** (may overlap with ssa_opt / regalloc).

## How to proceed (each = isolated, measured, reversible; backend → B==C MANDATORY before commit)
Start with **#2 (immediate-AND + const-mask-fold)** — smallest blast radius, clearly correct,
measurable on perf_fib.ps1, and it also shrinks the binary (helps RFC 0031/0030 goals). Then #3
(lea), then #1 (lean prologue — biggest but riskiest). Re-run perf_fib.ps1 + full regression + B==C
after each. Do NOT batch them — attribute each delta.

## Reality check
2.59x → 1.05x is a LARGE multi-session program (clang has decades of codegen tuning). Realistic
near-term goal: knock down the systemic taxes (#2,#3) to get under ~1.5x, then reassess whether the
5% gate is reachable without a full optimizing backend (loop/tail-call/inlining maturity) or should
be renegotiated with the user. Honest framing beats chasing 5% blindly. See [[session-state-2026-07-24e]].
