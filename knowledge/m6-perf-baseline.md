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

## Allocator reorder (prefer-volatiles) — TRIED + REVERTED 2026-07-24e (measured NEUTRAL)
The greenlit "allocator work" first attempt: `get_allocatable_gprs` reordered per-ABI to try all
VOLATILE GPRs before callee-saved (so non-call-spanning values avoid prologue pushes). Built,
**B==C `B06CDC7F`**. Result: **NEUTRAL — fib prologue UNCHANGED (still 5 pushes), self-host driver
byte-identical in size (delta 0)**, fib -O3 907 vs 936 ms = noise. Reverted. Reason: fib's
callee-saved regs hold values LIVE ACROSS A CALL, which already forbid volatiles
(graph_coloring_alloc:867) → they MUST be callee-saved regardless of preference order; the reorder
only reaches non-spanning values, which weren't the bottleneck.

## ⭐⭐⭐ THREE attempts, one conclusion: the gap is STRUCTURAL, not tweak-shaped
| attempt | result |
|---|---|
| opt A (cmp→imm fusion) | shipped, correct, ~5% -O3, **ratio-neutral** |
| opt B (lea for reg±const) | **REGRESSED** (regalloc spilled n), reverted |
| allocator prefer-volatiles reorder | **NEUTRAL** (bottleneck is elsewhere), reverted |
fib's real cost = redundant copies (`mov rcx,rbx; mov rbx,rsi`) + values kept live across calls +
greedy-colouring quality. These are coupled and resist incremental change: coalescing the copies
(opt-B class) perturbs the greedy allocator into spills; reordering preferences doesn't reach
call-spanning values. **Closing 2.58x→1.05x needs a real allocator+selection overhaul** (SSA-based
or iterated-coalescing allocation, rematerialization of `base±c`, live-range splitting) — a large
multi-session rewrite of self-host-critical code, NOT peepholes. **Recommend renegotiating M6 to a
reachable near-term milestone (≤2x clang) OR an explicit decision to fund the overhaul.**

## ⭐⭐⭐⭐ 2026-07-29 — THE GAP IS NOW DECOMPOSED (user idea: compare against hand-written ASM)
The "structural, needs an allocator rewrite" conclusion below was drawn from ONE benchmark
against ONE reference, and it was **over-general**. Two measurements settle it.

**(1) A four-shape suite (`scripts/perf_suite.ps1`)** — fib is an OUTLIER, not the norm:

| shape | what it isolates | AXIOM -O3 vs clang -O2 |
|---|---|---|
| fib | recursion / non-leaf call | **2.46x** |
| xorshift | serial ALU, no calls, no memory | 1.37x |
| arrwalk | dependent-index array walk | 1.15x |
| callloop | hot NON-recursive call (noinline both sides) | **1.01x** |

Call overhead per se is NOT the tax (callloop is at parity). Everything except fib is 1.0–1.4x.

**(2) A hand-written NASM floor** (`fib_hand.asm`, embedded in perf_suite.ps1). clang does NOT
compile fib as written — it applies an **accumulator transform that turns the second recursive
call into a loop**, so it executes ~HALF the calls AXIOM does (its `fib` has ONE `call` and a
back-edge; verify with objdump before disbelieving). Comparing to clang therefore conflates two
different gaps. The NASM reference uses the SAME naive double recursion AXIOM emits:

| build | Fib(40) | vs clang |
|---|---|---|
| clang -O2 (accumulator loop) | 347–355 ms | 1.00x |
| **hand NASM (naive double recursion)** | **541–564 ms** | **~1.55x** |
| AXIOM -O3 (before this session) | 853 ms | 2.46x |
| AXIOM -O3 (after the regalloc fix) | 817 ms | **2.30x** |

⇒ **codegen gap ≈ 1.5x** (817/541; reachable by backend work)
⇒ **missing-optimization gap ≈ 1.5x** (541/355; reachable ONLY by an opt pass)
(Absolute ms drift ±5% run to run — quote the two RATIOS, and re-measure both columns in the
same run. The product is stable at ~2.3x.)

⭐ xorshift is the cleanest secondary datum: **asm floor 222 ms ≈ clang 220 ms**. There, clang has
NO transform we lack, so its whole 1.35x is codegen — our tight-loop code is ~1.33x off the
hand-written floor. That is the honest size of the "instruction selection + allocator" debt on
loop code, and it is much smaller than the 2.3x fib headline suggested.

### The M6 decision this forces
**The ≤5% gate is unreachable by allocator/selector work — PROVEN, not estimated.** A perfect
backend lands at 564 ms = **1.63x clang**. Reaching 1.05x additionally requires the accumulator/
tail-recursion transform, i.e. an optimizer RFC, and then near-perfect codegen on top. Recommend
splitting M6 into two independently measurable gates:
- **M6-codegen**: AXIOM within ~15% of the hand-ASM floor per shape (today: fib 1.46x; the other
  three shapes need their own asm floors to be scored — xorshift's is already in the suite).
- **M6-opt**: the accumulator/tail-recursion pass, scored vs clang. Bigger ROI than the allocator
  (1.63x vs 1.46x) and does not touch the allocator's self-host-critical code.

⭐ **Method lesson**: when a reference compiler beats you by a lot, FIRST check it is running the
same algorithm. A hand-written asm floor separates "our codegen is bad" from "they applied a
transform we don't have" — three prior sessions burned peepholes without knowing which they faced.

## ✅ SHIPPED 2026-07-29 — precise PARAMETER liveness (B==C `1D9E3AE2`, 553/553)
The first backend win that is not a peephole. `compute_liveness` pinned every parameter snapshot
(`MOV pv <- PHYS(arg reg)`) live to the LAST instruction of its function — a conservatism that
**predated the CFG-aware dataflow** (RFC 0016 P2') sitting 50 lines below it, which already extends
anything live across a back-edge (incl. an RFC 0026 tail-recursion re-entry, whose snapshot sits
before `.L_b_0`). Consequences of the pin: every param interfered with every later value, spanned
every call, was therefore forced into a **callee-saved** register, and its copy chain could never be
coalesced by the `move_partner` bias. fib pushed/popped `%rbx` and `%r12` on ~331M calls for values
dead two instructions in. Fix = end the interval at the def and let the dataflow grow it (step 7
only ever GROWS intervals, so this cannot under-approximate). fib: 5 callee-saved → 3, **853→817 ms,
2.46x→2.30x**. Non-fib shapes unchanged, as expected.

### ⚠️ It exposed a LATENT RCX-clobber bug — the pin was load-bearing by accident
Removing the pin turned `t_fft` into a SIGSEGV and `t_foldu32wrap` into a wrong answer. Root cause
was NOT the liveness change: **x86 shifts by a variable count use the `_cl` form**, so the selector
emits `MOV rcx <- count` before the shift, destroying whatever else lived in RCX — and the allocator
denied RCX only to the shift's own DESTINATION vreg, never to values merely **live across** it. With
params pinned, params always spanned a call and so were always callee-saved and never in RCX; the
moment liveness got precise, a pointer param landed in RCX across `n >> 1` and the function wrote
through a wrecked pointer. Fix = mirror the existing IDIV/RAX-RDX span rule: every interval strictly
containing a variable-count shift gets `forbid_rcx` (immediate-count shifts excluded — they never
touch RCX, and including them would cost a register in every shift-heavy function).
Oracles: `t_shiftrcxclobber` (42; SIGSEGV on the pre-fix build — calibrated, not assumed) and
`t_paramlive` (42; the four shapes the pin was insuring: param read after a call / across a
back-edge / through tail-recursion / only in a branch target).

⭐⭐ **Lessons.** (1) A conservatism that "has always been there" may be **hiding** a real bug rather
than preventing one — removing it is how you find out, so do it behind the full gate. (2) **B==C is
not correctness**: the first (buggy) build was a clean B==C fixpoint and still miscompiled t_fft;
the 553-oracle regression is what caught it. (3) Reassuringly, B converged to the SAME hash
`1D9E3AE2` when seeded from both the trusted baseline and the buggy intermediate — seed-independence
is a cheap extra signal that a fixpoint is real.

## ⭐⭐⭐⭐⭐ 2026-07-29b — THE CODEGEN BACKLOG IS NOW **PRICED** (`scripts/perf_asm_variants.ps1`)
Rather than implement-then-measure (which cost three sessions), take the hand-written NASM floor
and re-introduce AXIOM's codegen habits ONE AT A TIME in assembly. Each delta is that habit's
price, measured, with zero compiler risk. fib(40), best of 7 (V0 reruns at 537–541 ms ⇒ noise ±3):

| variant | ms | delta | verdict |
|---|---|---|---|
| V0 floor (early exit, no rbp, 2 csr, lea/dec) | 537.7 | – | |
| V1 + frame BEFORE the early-exit test | 538.2 | **+0.5** | **shrink wrapping is WORTHLESS** |
| V2 + rbp frame pointer | 521.1 | **−17.1** | **frame-pointer omission is WORTHLESS** (even slightly faster) |
| V3 + copy chain + 3rd callee-saved | 622.3 | **+101.2** | ← register **coalescing** |
| V3b + `mov/sub` instead of `lea`/`dec` | 737.6 | **+115.3** | ← instruction **selection** (biggest) |
| V4 + `jcc`+`jmp` instead of fallthrough | 779.9 | **+42.3** | ← block **layout** |

V4 is a faithful transcription of what AXIOM emits and lands at 780 vs AXIOM's 817 (~4.5%
residual) ⇒ **the codegen gap is fully explained**, and the sum of the three real items (~258 ms)
is essentially all of it.

⛔ **DO NOT IMPLEMENT shrink wrapping or frame-pointer omission.** Both measure zero. push/pop of
a hot stack are nearly free on modern x86 (dedicated stack engine, L1-hot lines); the intuition
that "half of fib's 331M calls pay a useless prologue" is arithmetically true and
performance-irrelevant. This was predicted at 10–20% and measured at 0.1% — the estimate was
wrong by two orders of magnitude. Note clang does not shrink-wrap fib either.

⭐ **Priority (measured, not guessed): (1) selection `lea`/`dec` for `reg±const`, (2) register
coalescing, (3) block layout.** ⚠️ The 2026-07-24e `lea` attempt REGRESSED — but the pricing shows
the `lea` FORM is worth ~115 ms, so that failure was about *where* the transform lived (a
post-selection peephole that lengthened a live range and provoked a spill), not about its value.
Do it inside the selector's OP_IADD/OP_ISUB lowering, co-designed with coalescing.

⭐⭐ **METHOD** — this is the reusable win: an assembly harness prices an optimization BEFORE any
compiler code is written. Two of five candidates turned out to be worth nothing, which no amount
of reasoning about instruction counts revealed. Price first, then implement.
See also [[rfc0034-ir-pipeline-evolution]] for how this reprioritizes a mid-end redesign.

### Refinement after two more variants (V3c/V3d) — it is COPIES, not arithmetic
| step | delta | reading |
|---|---|---|
| V3b → V3c (fold const into the SUB immediate) | −0.5 / −13 ms (2 runs) | **noise — immediate-folding is NOT worth it** |
| V3c → V3d (compute the arg straight into `rcx`) | −38.5 ms | coalescing with a PRECOLORED reg |
| V3d → V3 (`lea`/`dec` + result already in rax) | −53 ms | more copy elimination |
The model that fits every row: **≈20–27 ms per instruction removed from the recursive path**
(≈331M calls). So the target is total hot-path instruction count, and almost every removable
instruction is a `mov`. That also retires the old "immediate operand folding" idea
([[perf-immediate-operand-folding-inert]]) on VALUE grounds, independently of its cast-guard bug.

## ✅ SHIPPED 2026-07-29c — block layout: thread jump-to-jump, drop fallthrough jumps
**fib 817 → 654 ms (−20%), 2.30x → 1.87x clang, 1.51x → 1.24x of the ASM floor.** Largest single
win in this milestone's history, from ~90 lines. `thread_and_drop_jumps` (x86_selector.ax, runs
last in `select_all`). The block lowering emitted every `if` as `JCC taken; JMP fallthrough`, and
the fallthrough target was usually a block whose whole body was ANOTHER `JMP`:
```
cmp; jl base; jmp L_X; <else body>; ... base: ...; L_X: jmp <else body>
```
i.e. two unconditional jumps to reach the instruction that was already next. Two local rewrites:
(1) **thread** a JMP/JCC whose target block's first real instruction is an unconditional JMP,
retargeting to that JMP's destination (depth-capped for `L: jmp L`); (2) **drop** a JMP whose
target label is reachable by passing only LABELs, since control already arrives there. Labels are
zero-byte pseudo-instructions, so "only LABELs in between" really is "the next instruction".
Conditional jumps are only ever RETARGETED, never deleted, so no path is created or destroyed;
LABELs are never removed, so jump tables (RFC 0028) that reference them still resolve.
Gate: B==C `958F3BF5`, regression 553/553, ctgc 16/16, exe-size 4/4, ELF 12/12.
⭐ Priced at +42/+72/+108 ms for ONE extra jump; delivered 163 ms because our code had TWO.
⭐ Blocks orphaned by threading are left in place (a few dead bytes); dead-block removal is a
separate, later cleanup — it changes nothing observable.

### Remaining measured backlog for fib (now 1.24x of floor)
Everything left is **register coalescing**: the param copy chain (`mov rax,rcx; mov rax→rbx`),
the arg temp→`rcx` copies, and the final `mov rax,rsi`. Worth ~90–100 ms combined. The named
algorithm is **George & Appel iterated coalescing with precolored nodes**; today there is only a
one-partner, first-wins colouring BIAS (`move_partner`), which cannot coalesce into precoloured
registers at all. That is the next item — and it is also what RFC 0034 §3.1 says Typed SSA would
make tractable.

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
