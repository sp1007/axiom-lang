---
name: session-state-2026-07-24c
description: "STATE 2026-07-24c — RFC 0015 P3 CTGC compile-time free FLIPPED TO ON BY DEFAULT. Driver AF899BDF, regression 528/528. The 'high-risk, needs borrow/alias tracking' framing was STALE — that shipped long ago; the flip was inert on the self-host (freeable set = 0) so it just worked. Added -no-ctgc-free opt-out."
metadata:
  node_type: memory
  type: project
---

**ĐỌC ĐẦU TIÊN.** HEAD after this session, daily driver `bin/axc_native.exe` = **`AF899BDF`**,
regression **528/528** GREEN (rebuilt by the default-on compiler). User (chose from menu):
"RFC 0015 P3 — CTGC free" → "Flip default-on (có gate)".

## What shipped: `-ctgc-free` OFF→ON by default (RFC 0015 P3 activation)
Compile-time free is now the DEFAULT. Every provably non-escaping, single-owner local aggregate
is freed at block fall-through (general non-drop `OP_DESTROY` + container free-glue RFC 0027
path C/D + drop-glue RFC 0014). Two source edits + two gate-script fixes:
1. **main_air.ax:801** `mut ctgc_free := false` → `:= true`. Added `-no-ctgc-free` opt-out
   (main_air.ax option loop) as the escape hatch (debugging / a program whose aliasing the
   escape analysis under-approximates → would become a UAF instead of a safe leak).
2. **scripts/ctgc_free_check.sh** — the "off" column now builds with `-no-ctgc-free` (was plain
   `-O1`, which is now default-ON, so both columns would have been ON and the `t_drop 0|42`
   distinction lost). This ALSO validates the new opt-out flag.
3. **scripts/regression_repros.sh:527** — `t_drop|exit|0` → `42` (drop(self) now fires on a
   plain build; the off-behavior stays pinned by ctgc_free_check.sh via `-no-ctgc-free`).

## Why it "just worked" (the 2026-07-24b 'HIGH-RISK / needs borrow-alias tracking' note was STALE)
The borrow/alias tracking (P2 sound escape) + general-free activation + container free-glue +
RFC 0027 path D user-container free had ALL shipped 2026-07-16…07-22 behind the opt-in flag.
The self-host **freeable set = 0** (two escape holes closed earlier: container-store escape
`f873948` + reassign-to-borrow `68d2c78` — see [[ctgc-p3-scoping-2026-07-18]]), so the default-on
compiler compiles the compiler source IDENTICALLY to default-off ⇒ the flip is inert on the
self-host. The only remaining thing was the product decision to flip the default; the machinery
was done. ⭐ **Lesson: a "remaining task" note can describe a blocker that was already dismantled
by later sessions — reconcile the task against what actually shipped (grep the flag wiring + run
the gate) before treating its risk framing as current.**

## Gate (GREEN, revert-on-red satisfied)
- **Fast fixpoint A==B `AF899BDF`** (scripts/fast_fixpoint.ps1) — activation inert on self-host.
- **Full regression 528/528** built by the default-on compiler (`AXC=bin/axc_fpB.exe
  REGTMP=bin/_regtmp bash scripts/regression_repros.sh` → `REGRESSION_OK`). Only diff caught was
  `t_drop` 0→42 = the INTENDED drop side-effect (row updated).
- **ctgc_free_check.sh 14/14 CTGC_FREE_OK** — incl. the sharp negative oracles `t_ctgcescape`(33,
  Vec-stored ctor NOT freed), `t_ctgcfreeesc`(16, returned aggregate NOT freed), `t_ctgcborrow`(42,
  borrowed pointer field NOT freed). Both directions verified: off=0 via `-no-ctgc-free`, on=42.
- ⚠️ Self-inflicted flake: editing `regression_repros.sh` WHILE the bash run was reading it shifted
  byte offsets → a bogus "syntax error near line 719" at the summary (results still valid). Do NOT
  edit a running bash script; re-run clean instead. `bash -n` confirmed both scripts parse.

## NOT done / next
- **ELF path not re-verified under default-on** (`scripts/elf_linux_check.sh`). CTGC free is
  target-independent (same air_builder injection + `ax_free`), and self-host freeable=0 keeps ELF
  self-builds inert, but ELF test programs now free too — a follow-up `elf_linux_check.sh` run
  would confirm. Low risk (COFF 528/528 green, shared free path).
- **Return-path free still leaks (safe)** — locals live on a `return` path are not freed (needs a
  return-value-temp AST transform; ctgc.ax:111-118). Documented completeness gap, not a bug.
- Remaining heavy backlog unchanged: RFC 0009 P3 ELF `.so` (PIC wall), M4/M6 milestones.
  See [[session-state-2026-07-24b]].
