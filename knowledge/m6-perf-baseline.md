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

## Profiled root causes — VERIFIED against fib's ACTUAL disassembly
fib is at **0x140011dd7** in `fib_ax_o2.exe` (found via self-call detection: two `call 0x140011dd7`
at e12/e2a; AXIOM emits NO per-fn symbols). Its real hot body (⚠️ NOT the runtime syscall wrappers I
first extrapolated from — those legitimately push 8 regs; fib does not):
```
push rbp; mov rsp,rbp; push rbx; push rsi; push rdi; push r12; sub $0x20,rsp   ; 5 callee-saved
mov rcx,rbx ; mov rbx,rsi                     ; (C) redundant double-copy of n
mov $0x2,rax ; cmp rax,rsi ; jl base          ; (A) cmp materialises the const
jmp else                                       ; (D) jcc + unconditional jmp (no fallthrough)
mov $0x1,rax ; mov rsi,rdi ; sub rax,rdi ; call fib   ; (B) n-1 in 3 insns, should be lea
mov rax,rdi ; mov $0x2,rax ; mov rsi,r12 ; sub rax,r12 ; call fib  ; (A)+(B) again for n-2
add rax,rdi ; mov rdi,rax ; <epilogue>
```
⭐ **CORRECTION to the first-pass ranking:** #2 below (width-masks) is **INERT for fib** — fib is
all-`i64`, and `emit_wrap_to_width` (x86_selector.ax:887) masks ONLY i8/i16/i32 (size 1/2/4); i64/u64
get size 0 → no mask. So mask-elimination shrinks narrow-int code + binary size but does **NOT** move
the Fib metric. Don't start there for the perf gate. Ranked by ROI **for fib**:

A. **`cmp` materialises its constant** (`mov $2,rax; cmp rax,rsi`, 2 sites) → `cmp $2,rsi`. Cleanest,
   clearly correct, in the hot loop. **Fix = selector: OP_LT/compare with an ICONST operand → emit
   the imm form.**
B. **No `lea`/`dec` for `reg±const`** (`mov $1,rax; mov rsi,rdi; sub rax,rdi`, 2 sites = ~4 wasted
   insns) → `lea -0x1(rsi),rdi`. **Fix = selector: OP_ISUB/OP_IADD with a small ICONST operand →
   lea (or inc/dec).** Highest raw insn savings in the loop.
C. **Redundant copies** (`mov rcx,rbx; mov rbx,rsi` → `mov rcx,rsi`; `mov rax,rdi` after a call).
   **Fix = copy-coalescing / dead-move peephole over machine IR** (may overlap ssa_opt/regalloc).
D. **Branch shape**: every `if` = `jcc taken; jmp fallthrough` instead of laying the fallthrough
   block next → drop the extra `jmp`. **Fix = block layout in the emitter.**
E. **5 callee-saved vs clang's 2** — fib allocates callee-saved regs (rbx/rsi/rdi/r12) where volatile
   would avoid push/pop. **Fix = prefer volatile/caller-saved regs in the allocator when no value is
   live across a call that needs them** — biggest but hardest (regalloc), do LAST.

(Width-masks — historical #2, keep for narrow-int/binary-size, NOT the perf gate: pervasive
`mov $C; mov $0xff; and` materialises the mask + masks constants that already fit. Fix = immediate
`and $imm` + const-fold `mov $C; and $M`→`mov $(C&M)`. Localized, but inert on fib.)

## How to proceed (each = isolated, measured, reversible; backend → B==C MANDATORY before commit)
Start with **A (immediate cmp)** then **B (lea for reg±const)** — both are selector peepholes,
clearly correct, and directly in fib's hot loop. Re-run `perf_fib.ps1` + full regression + B==C after
EACH; attribute each delta (do NOT batch). Then C (copy-coalescing), D (branch layout), E (regalloc
volatile preference). Mask-elimination is a separate binary-size win, not the perf gate.

## Reality check
2.59x → 1.05x is a LARGE multi-session program (clang has decades of codegen tuning). Realistic
near-term goal: knock down the systemic taxes (#2,#3) to get under ~1.5x, then reassess whether the
5% gate is reachable without a full optimizing backend (loop/tail-call/inlining maturity) or should
be renegotiated with the user. Honest framing beats chasing 5% blindly. See [[session-state-2026-07-24e]].
