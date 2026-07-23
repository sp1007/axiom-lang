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

## Opt A — concrete implementation recipe (investigated 2026-07-24e, ready to execute)
- **Insertion point:** end of `select_all` (x86_selector.ax) — a peephole there applies to ALL
  backends at once (coff/asm/elf all call `select_all` → liveness → regalloc → emit; see
  x86_coff.ax:762). Run it on the `MachInstVec` BEFORE `compute_liveness`, so vregs are still
  virtual and fusing also drops register pressure.
- **Pattern (fib's 2 cmp sites):** the selector emits `MOV_IMM v, c` IMMEDIATELY followed by
  `MACH_CMP dst=<reg>, src1=OPND_VREG(v)` (the const is the SRC1 operand of the cmp; dst holds the
  compared reg). Both the branch-fusion path (fib's `if`) and `select_comparison` produce a bare
  `MACH_CMP`, so a peephole on `MACH_CMP` catches both.
- **Fusion:** if `inst[i]` is `MOV_IMM` with dst vreg `v` and a `c` that fits imm32, and `inst[i+1]`
  is a `MACH_CMP` whose `src1.kind==OPND_VREG && src1.vreg==v`, AND `v` is used EXACTLY ONCE and
  defined EXACTLY ONCE across the whole fn (mirror `const_shift_amount`'s non-SSA guards — AIR/mach
  IR is non-SSA, a vreg can have multiple defs), then set `cmp.src1 = OPND_IMM(c)` and drop the
  MOV_IMM (rebuild the vec without it, or set it to `MACH_NOP` which the emitters already skip —
  x86_emitter.ax:153 / x86_asm_emitter.ax:143). The encoder already supports `MACH_CMP` with
  `OPND_IMM` src1 (x86_selector.ax:1402 emits exactly that form for the setcc path).
- **Guard:** imm must fit signed imm32 (cmp r64,imm32 sign-extends); bail otherwise. Do NOT fuse if
  `v` has >1 use or >1 def. Adjacency (i+1) keeps it trivially safe (no redef between def and use).
- **Gate:** B==C MANDATORY (backend), full regression, then `perf_fib.ps1` for the delta. Opt B
  (lea for reg±const) is a follow-up — its MOV_IMM is NOT adjacent to the sub (`mov $1; mov rsi,rdi;
  sub`), so it needs the single-use form without the adjacency shortcut.

## Opt A — SHIPPED 2026-07-24e (`30d03b0`, driver `AFA6529F`)
`fuse_cmp_immediate` (x86_selector.ax, runs at end of `select_all` before `fuse_compare_branch`).
fib's `n<2` is now `cmp $0x2,%rsi` (was `mov $2,rax; cmp rax,rsi`). Gate: **B==C `AFA6529F`** (A≠B
expected), regression 530/530, ctgc 16/16, fib=203. **Measured delta: ~5% absolute -O3 (975→920 ms),
gap ratio UNCHANGED (2.58x) — noise-dominated** (clang itself varied 356–376 ms run-to-run; opt A
removes 1 insn from an ~18-insn loop). ⭐ **Lesson: a single-instruction peephole will NOT dent a
2.58x gap** — the measurable wins must be structural. The n-2 `mov $2` was correctly NOT fused (it
feeds a `sub`, i.e. opt B, not a cmp).

## Opt B — PROTOTYPED + REVERTED 2026-07-24e (measured fib REGRESSION)
`fuse_lea_const` (post-selection 3-window `MOV_IMM v,c; MOV dst,base; SUB/ADD dst,v` → `LEA
dst,[base±c]`) was built, B==C-verified (`59A48AF4`), and emitted the intended `lea -0x1(%rsi),%rax`
for `n-1`/`n-2` — but **MEASURED SLOWER**: -O3 922→**974 ms**, ratio 2.59x→**2.73x**. Root cause in
the disasm: removing the `mov dst,base` copy lengthened `base`'s (=`n`) live range, so the register
allocator **spilled `n` to `-0x18(%rbp)` and reloaded it before each `lea`** — the memory traffic
outweighed the one saved instruction (the prologue also dropped 5→3 callee-saved but that didn't
compensate). Reverted. ⭐⭐ **LESSON: instruction count is NOT the perf metric — measured time is.** A
post-selection peephole that cuts insns can lose badly to its interaction with a naive linear-scan
allocator (longer live ranges → spills). A real `lea`/copy win must be **co-designed with the
allocator** (e.g. teach it to rematerialize `base±c` or avoid the spill), not bolted on after
selection. Do NOT re-attempt the naive post-selection lea peephole.

## Revised direction (the honest one)
The 3 remaining "instruction-shaving" peepholes (lea, copy-coalescing, branch fallthrough) all risk
the same regalloc backfire, AND opt A proved even a clean 1-insn win is noise on a 2.58x gap. The gap
is **structural**: fib spills `n` and shuffles registers because the **linear-scan allocator + the
non-SSA MOV-heavy selection** are far from clang's. Closing it needs allocator/selection MATURITY
(better copy handling, rematerialization, fewer spills), not more peepholes — a large program. **The
≤5% gate is almost certainly not reachable without that.** Recommend to the user: either commit to
the allocator work (multi-session, high-risk, B==C each step) or **renegotiate M6's target** (e.g.
"within 2x of clang" as a nearer milestone). Opt A stays (correct, banked). Next low-risk M6 step if
continuing: broaden `perf_fib.ps1` into a small suite (iterative loop, array sum) so allocator work
is measured across shapes, not just fib.

## (obsolete) Opt B — original plan: lea for reg±const
fib still has, at BOTH recursion sites, `mov $1,rax; mov rsi,rdi; sub rax,rdi` (n-1) and
`mov $2,rax; mov rsi,r12; sub rax,r12` (n-2) = 3 insns each → `lea -0x1(%rsi),%rdi` (1 insn). Saves
~4 insns/loop (vs opt A's 1). **Best done IN the selector's OP_ISUB/OP_IADD lowering** (not a
post-peephole — the MOV_IMM isn't adjacent to the sub): if src2 is defined by an OP_ICONST with a
value fitting signed imm32, emit `LEA dst,[src1 ± c]` instead of `mov dst,src1; sub/add dst,cvreg`.
**SAFE because arithmetic flags are never consumed** in this compiler (comparisons are separate
OP_LT/OP_EQ → cmp; OP_IADD/ISUB are pure value ops), so LEA (which sets no flags) is a valid
substitute. Need a `const_value_of_vreg(fn, vreg)` helper (mirror `const_shift_amount`'s
single-def non-SSA guard). B==C-gate + perf_fib.ps1. Then C (redundant-copy coalescing: `mov
rcx,rbx; mov rbx,rsi`→`mov rcx,rsi`), D (branch fallthrough), E (regalloc volatile preference).

## Reality check
2.59x → 1.05x is a LARGE multi-session program (clang has decades of codegen tuning). Realistic
near-term goal: knock down the systemic taxes (#2,#3) to get under ~1.5x, then reassess whether the
5% gate is reachable without a full optimizing backend (loop/tail-call/inlining maturity) or should
be renegotiated with the user. Honest framing beats chasing 5% blindly. See [[session-state-2026-07-24e]].
