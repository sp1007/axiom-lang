---
name: session-handoff-2026-07-30c
description: "HANDOFF 2026-07-30c — peephole 1f is SHIPPED and GREEN (d454c49, B==C B71DF82A, 581/581). The RED was ABI, not float pressure. NEXT TARGET identified and evidenced: LOOP-HEADER alignment — a pure +16-byte code shift costs callloop 7%, which is what makes the M6-codegen milestone unmeasurable."
metadata:
  node_type: memory
  type: project
---

# HANDOFF 2026-07-30c — 1f shipped GREEN; next target is LOOP alignment

## State of the tree
- **HEAD = `d454c49`**, pushed to `main`. Clean except two things that are NOT mine and should be
  left alone: `CLAUDE.md` (user's context-hygiene edit) and untracked `.claude/settings.json`.
- **Daily driver `bin/axc_native.exe` = B == C `B71DF82A`**, built from HEAD source.
- **BASELINE = 581 / 581, 0 failed.** Was 578; `t_floatparamchain` adds 3 rows (main list at -O1
  plus the O2 and O3 sweep). Below 581 is RED.
- `bin/axc_pre1f.exe` is a reference compiler built from `HEAD~1` source, kept for paired pricing.

## What shipped
**Peephole 1f `collapse_copy_chain`** — the non-adjacent copy-chain collapse that was left RED and
uncommitted by handoff 07-30b. Guard added, gate GREEN, priced, committed, pushed.

Gate (backend change ⇒ `A != B` expected, `B == C` is the criterion): **B==C `B71DF82A`**,
regression **581/581**, ELF 12/12, ctgc 16/16, exe_size 4/4, so_export OK, lib_collision 6/6.

## ⭐ The RED was ABI ordering — BOTH suspected mechanisms were wrong
07-30b left two candidates: register class, or float pressure exposing a latent spill bug. Neither.

**Float pressure was ruled out by measurement**, and cheaply: a function with 18 simultaneously-live
f64 values compiles correctly on HEAD, and four f64 params with a trivial body are fine under 1f.
The break needs 4+ register-passed float params AND enough pressure to spill a param vreg. Two
one-file probes settled what two sessions of reasoning had not.

**The disassembly then named the mechanism outright:**

    movsd %xmm0,%xmm2      ; v1 spilled, through the float spill scratch
    movsd %xmm1,%xmm2      ; v2 spilled, same scratch
    movsd %xmm2,%xmm9      ; v3 := p2 -- reads a CLOBBERED %xmm2

The float spill scratch is **XMM2** (2026-07-29e), and on win64 **XMM2 is the third float argument
register**. `emit_param_prologue` avoids the collision structurally by snapshotting all four
incoming arg registers into fresh vregs BEFORE reading any — its own comment says exactly that. 1f
folded the snapshots away and deleted the protection.

**Guard: the def must not read `OPND_PHYS`.** A `MOV vT,PHYS(r)` is ABI establishment, not a value
copy; `counts[vT]==2` proves nothing about `r`, which is a second invisible source owned by the
calling convention. Chosen over float-ness or param-ness on three counts — it names the mechanism,
it costs none of the win (sumto's two winning folds have vreg-to-vreg defs even though its winning
DESTINATION is a parameter, so a param guard would have cost half), and it closes the integer form
of the same hazard too.

The float side of `param_idx_of_vreg` has **no backstop** — it lists RCX/RDX/R8/R9 but not
XMM0–XMM3. Recorded as a latent hole, deliberately not fixed: [[bug-float-arg-reg-unprotected]].

## ⭐ Pricing lesson: subtract a MEASURED startup floor
Re-priced after the guard: `t_tailrecloop` **21.1 → 18.2 ms wall, −13.8%, 4/4 pairs same sign** —
matching the pre-guard −14.2%, so the guard cost nothing.

⚠️ A first bash attempt reported a confident **−5%, 5/5 pairs same sign, and it was garbage**: the
samples were ~55 ms of which ~55 ms was process startup. It was caught only by measuring a trivial
program (`return 0`) in the same harness and seeing it cost the same 55 ms. **Any perf harness must
report the startup floor next to the number.** `scripts/price_1f_paired.ps1` does; it exists for
that reason. Real startup floor here is **10.3 ms**, so a 20 ms sample is ~half overhead.

## ⛔⭐⭐⭐ NEXT TARGET (evidenced, not guessed): LOOP-HEADER alignment
1f trades **tailrec −13.8%** for **callloop +5.7…+8.9%** (4/4 pairs same sign, paired against
`axc_pre1f.exe`). The callloop cost is **pure LAYOUT, proven by observation**:

- Normalized objdump: 43 differing instruction lines out of 6318, ALL in lines 3301–5078.
- callloop's hot loop is at line ~5900. The tail from there on is **byte-identical instruction
  text** in both builds.
- Its address moved `0x140006418 → 0x140006428` = **exactly +16 bytes**.

Unchanged code cannot get slower — only its address changed. **Third instance of this exact
signature** (peephole 1e, unguarded 1f, guarded 1f).

⭐⭐ 16-byte FUNCTION-ENTRY alignment already shipped (`7ac52f5`) and does not help here. `main`'s
entry IS 16-aligned (`0x1400063d0`, preceded by pad NOPs) while the loop header is `0x1400063f9` —
**mod 16 = 9, not aligned at all.** AXIOM emits no loop-header alignment; `-falign-loops` is
standard in clang/gcc. So "align loop headers" was the obvious next move.

## ⛔⛔⛆ AND IT IS REFUTED — do not build a loop-alignment pass
I wrote the above into this handoff as the evidenced next target, then tested it before writing any
compiler code. **It does not survive.**

First, arithmetic that should have been done immediately: **a +16-byte shift preserves address
mod 16.** With function entries 16-aligned, a loop header's mod-16 offset is *invariant* across
these builds by construction. So a 16-byte loop-align pass provably cannot explain the callloop
delta, let alone fix it.

Then the measurement, identical code at controlled layout offsets (`bin/bench/align`):

    loop body @ mod64=54  ->  134.8 ms   (133.8 / 134.9 / 135.1 / 135.4)
    loop body @ mod64=38  ->  153.1 ms   (152.0 / 153.5 / 153.8)
    loop body @ mod64= 6  ->  167.3 ms

**The FASTEST layout is the LEAST aligned one, and the SLOWEST is the nearly-64-aligned one.** An
alignment pass has to pick a boundary; the optimum is not at a boundary. Refuted.

⚠️ **And do not quote that 24% as a constant.** Re-running with the loop moved into its own
function gave only a **4% spread that does not separate cleanly by address** (one binary at mod64=5
read 59.3 ms while three others at the SAME address read 57.0/57.4/57.5). The layout term is real
and shape-dependent, not a fixed tax. Both experiments are recorded in the header of
`scripts/perf_layout_dist.ps1`.

## ⇒ ACTUAL next target: fix the PROTOCOL, not the compiler
`scripts/perf_layout_dist.ps1` (new, this session) builds one shape at N controlled layout offsets
and reports **median + layout spread**, so a delta smaller than the spread is never read as a
result. Zero compiler risk. Usage:

    scripts\perf_layout_dist.ps1 -Compiler bin\axc_native.exe -Shape callloop
    scripts\perf_layout_dist.ps1 -Compiler bin\axc_pre1f.exe  -Shape callloop

Next concrete steps, in order:
1. Add NASM-floor shapes to the same distribution treatment so the M6 gate ratio itself becomes
   median-over-layouts instead of one draw. **This is what unblocks the milestone** — the current
   gate reads one binary per side and its input carries an unquantified layout term.
2. ~~Re-decide 1f's callloop cost under the new protocol.~~ **DONE — see below.**
3. Only then revisit whether any codegen work is owed on callloop at all. (Per below: **none is.**)

## ⭐⭐⭐ RE-PRICED OVER LAYOUTS: 1f is a PURE WIN, and the callloop cost was a phantom
Both 1f numbers re-measured with 8 controlled layout offsets per compiler, medians compared.

**callloop, on the EXACT shape that produced the original +7%** (loop inline in `main`):

    pre-1f  samples: 136.3 134.1 153.2 150.7 152.0 152.5 135.0 135.0   MEDIAN 143.5  spread 14.3%
    HEAD    samples: 152.0 156.1 133.7 133.8 134.0 134.9 152.6 154.1   MEDIAN 143.5  spread 16.8%

**Identical medians — the cost is EXACTLY ZERO.** And the mechanism is visible in the raw samples:
there are two clusters (~134 and ~152) and the compilers assign them to *opposite* variants, which
is what a ±16-byte shift does. The set of layouts and their costs is unchanged; only which layout
each build draws changed. The earlier **+5.7…+8.9% with 4/4 pairs the same sign was 100% a layout
draw.**

**tailrec — SURVIVES:**

    pre-1f  MEDIAN 21.4 ms  (21.0 .. 21.8, spread 4.0%)
    HEAD    MEDIAN 18.4 ms  (17.7 .. 19.0, spread 7.4%)

**−14.0%**, matching the single-binary −13.8%, and the two sample sets **do not overlap at all**
(`max(HEAD)=19.0 < min(pre-1f)=21.0`). Clean separation ⇒ attributable to the change.

⇒ **Peephole 1f: tailrec −14%, callloop 0. A pure win.** The M6 "callloop got worse" concern is
withdrawn.

## ⛔⭐⭐⭐ THE METHOD LESSON — pairing does NOT remove layout bias
This is the most transferable thing in this handoff. `knowledge/m6-perf-baseline.md` has said for
sessions: "variance is 8–10% per run, so measure PAIRED and ALTERNATING." That rule is **not wrong
but misattributed, and it is not sufficient.**

- Within one binary, best-of-9 is repeatable to **~1%**. The big term is not run-to-run noise.
- The big term is **deterministic layout sensitivity**, resampled every time a binary is rebuilt.
- Pairing/alternating removes **drift**. It cannot remove layout bias, because it compares **the
  same two binaries** over and over. **Same-sign consistency across pairs is therefore NOT evidence
  that an effect belongs to the compiler change** — 4/4 identical-sign pairs said "+7%" for an
  effect whose true value is 0.

Three past "results" were this same phantom: 1e "cost fib 5.1%", xorshift "+4.6% slower from a
change that removed instructions", and 1f "cost callloop 7%". **Every future perf claim must be a
median over layouts with the spread reported next to it.**

⚠️ The 07-30a lesson still applies to anything determinism-flavoured: measure on the **DIFFERENCE
BETWEEN TWO BUILDS**, never on one build. Judging a determinism fix by whether one number improved
is the category error that already caused one unjust revert.

## Current M6-codegen standing (ONE suite run — do not bank it)
`scripts/perf_suite.ps1`: fib **1.12x** ✅, xorshift **0.99x** ✅, arrwalk **1.08x** ✅,
callloop **1.17x** ❌ vs the hand-written NASM floor (gate is ≤1.15x, user decision D1).

⚠️ A single `perf_suite` run is explicitly untrustworthy in this project — the FIXED asm-floor
binary has itself swung 569→514 ms between runs. Treat the above as a pointer, not a result.
Notably fib is back inside the gate: the unguarded 1f's fib +14.5% was itself layout, and the guard
removed the prologue folds that caused the shift.

## Method notes worth keeping
- A reference compiler from `git show HEAD~1:<file>` is cheap and decisive; restore the working
  source and regen immediately after building it, so the tree never lingers dirty.
- Normalized objdump diff (strip addresses, `s/0x[0-9a-f]+/HEX/g`) turns "did my change touch this
  function?" into a yes/no. Watch out: that normalization also erases the constants you might be
  grepping for — locate the hot loop in the RAW dump, not the normalized one.
- The bench binaries have no per-function symbols, so whole-stream diff + line-range reasoning is
  the attribution tool available.
