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

⭐⭐ **The key new fact: 16-byte FUNCTION-ENTRY alignment already shipped (`7ac52f5`), and it does
not help here.** So the sensitivity is not function entry — it is the hot LOOP's offset inside the
function crossing a 32/64-byte boundary (uop-cache / DSB line). AXIOM emits **no loop-header
alignment** at all; `-falign-loops` is standard in clang/gcc for precisely this.

This is the highest-value next task because it is not merely a few percent: **it is what makes the
M6-codegen milestone unmeasurable.** Every peephole for three sessions has had its real effect
masked or faked by ±7–14% of layout noise.

⚠️ **Price it the way function-entry alignment had to be priced** (the 07-30a lesson, learned by
reverting a good change on a bad measurement): a determinism fix must be measured on the
**DIFFERENCE BETWEEN TWO BUILDS**, never on one build. Concretely — take two compilers that differ
only by an upstream code-size perturbation (`axc_pre1f` vs HEAD is exactly such a pair, +16 bytes),
and check that loop alignment shrinks the callloop delta between them toward zero. Measuring
"aligned vs unaligned on one build" is the mistake that already cost one revert.

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
