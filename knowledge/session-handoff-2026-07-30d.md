---
name: session-handoff-2026-07-30d
description: "HANDOFF 2026-07-30d — token-economy harness change shipped (5507dd1): sub-agent delegation is now the DEFAULT execution mode, and the real token sink was MEASURED and closed (MEMORY.md = 175 KB ≈ 87k tokens, ~25k per session just to orient ⇒ new compact knowledge/BACKLOG.md). No compiler source touched. Two sub-agent tasks were IN FLIGHT when this was written."
metadata:
  node_type: memory
  type: project
---

# HANDOFF 2026-07-30d — token economy: delegate by default, stop paying 25k tokens to orient

Short handoff by design — it is a harness/docs session, not a compiler session. Compiler state is
unchanged from [[session-handoff-2026-07-30c]]; read [[BACKLOG]] first for live state.

## State of the tree
- **HEAD `5507dd1`** (docs/harness only). Compiler source **untouched** ⇒ no gate owed.
- Driver `bin/axc_native.exe` = **A==B `105B623C`**, baseline **593/593** — both unchanged.
- Untracked and deliberately left alone: `.claude/settings.json`, `bin/probe3/`.

## What shipped and why
The user asked a **third** time for "auto-`/clear` after each task to save tokens". `/clear` is still
not callable — not a tool, not any hook (Stop/SessionEnd/PreCompact), not a settings threshold, not
`/loop`/MCP. Repeating that answer a third time would have been useless, so instead: **measure where
the tokens actually go.**

⭐ **The measurement is the finding.** `knowledge/MEMORY.md` is **175 KB ≈ 87k tokens** — over the
read cap, and **one truncated page costs ~25k tokens per session purely to orient, before any work
starts.** That single read dwarfs anything `/clear` would have recovered. The index had grown into
the detail store: its "one-line hook per memory" lines are full paragraphs.

Two rules, neither needing a user keystroke (CLAUDE.md §24 "Token economy"):
1. **Delegate a WHOLE task to a sub-agent by default** — fresh context per task, only the report
   returns ⇒ functionally a per-task `/clear`. Inline is now the *exception*, for self-host-critical
   work that needs the orchestrator's accumulated diagnosis. ⚠️ Never delegate a two-tool errand;
   each sub-agent re-reads its own orientation.
2. **`knowledge/MEMORY.md` is Grep-only.** Orientation = new compact **`knowledge/BACKLOG.md`**
   (~2k tokens) + newest `session-handoff-*.md`.

⚠️ **`BACKLOG.md` is a POINTER file and must never hold facts of its own.** `MEMORY.md` stays the
detail store and **wins any conflict**. Two copies of one truth is exactly the defect class that
produced the interface-return miscompile (two copies of one walk that drifted apart) — so that
constraint is written into both files, not just remembered here.

Applied to `CLAUDE.md` §24 + changelog, `.claude/skills/axiom-autopilot/SKILL.md` Phase 0 (orient
from BACKLOG, never MEMORY wholesale) and Phase 3 (delegate by default), and a pointer line at the
top of `MEMORY.md`.

## ✅ Both dispatched sub-agents have reported (updated in place — see the rule below)
1. **arrwalk layout-distribution reading — the task was ALREADY SHIPPED** in `3ef26f0`; only cosmetics
   were left (arrwalk missing from the script's default `$Shapes`, plus the stale TODO itself). Closed
   by `a2e04ca`: re-verified **arrwalk 1.092x ± 1.7% PASS by 5.1%**, xorshift control **0.995x ± 0.8%**
   unchanged, startup floor 10.8 ms.
   ⛔ **I dispatched already-finished work** because I trusted 07-30c's "remaining to do" line without
   running `git log` on the target file first. Now a Phase 1 rule (`4cec761`): **cross-check a TODO
   against git BEFORE dispatching**, and — for the writing side — **when a later commit closes a TODO,
   edit the file holding that TODO in the same commit.** Under delegate-by-default a wasted dispatch is
   a whole wasted context, not just a few minutes.
2. **probe4 — THREE more silent miscompiles**, all re-verified by me directly rather than taken on the
   agent's word. Details + control matrices + the clean-swept surface list live in [[BACKLOG]]; not
   duplicated here (pointer discipline). Headline: `let a: f64 = 3` yields **0.0** because
   `lower_int_lit` emits `OP_ICONST` carrying a *float* `type_id` while its sibling `lower_float_lit`
   does it correctly — **two copies of one mechanism drifting apart, for the third time today.**
   ⚠️ It also falsified a written claim: `knowledge/bugs.md:1015-1019` asserted *"`let x: f64 = 3` →
   OK"*. It was never verified and it is false.

## ✅ probe4 bug #1 SHIPPED — `76de988`, `A==B==C 824807E2`, **597/597** (new baseline)
Verified by me row by row, not accepted on report. `let a: f64 = 3` now yields 3.0 (`h1` 42 at -O0
**and** -O1; oracle `t_intlitfloatctx` 42 at -O0/-O2; the f32 twin `i2` still 42).
- ⭐ The fix **merged the drifted copies instead of adding a third**: `emit_float_const` is now the
  single emit site for a float constant, `lower_float_lit` gave up its private copy to call it, and
  `lower_int_lit` routes into it when the adopted `type_id` is 9/10.
- **The §9 invariant is the durable product**, not the one-line fix: `verify_air_const_types` rejects
  `OP_ICONST` carrying a float `type_id` (and the symmetric FCONST case, ignoring type_id 0 = unset),
  called from `lower_func` for every function. **Calibrated** — with the fix disabled the build prints
  `internal error: OP_ICONST carries a float type_id … inst #0 … type_id=10` and exits 1 with no
  output file. RFC 0006 gained §7.1 stating it as a *requirement*; `bugs.md:1019`'s false claim was
  corrected in place. No new RFC — RFC 0006 already governed this and only lacked the invariant.
- ⭐ **The agent did not overclaim its scope, and I checked**: it reported four positions still broken
  because the literal's node type stays INTEGER there (nothing float to materialize) — f64 struct
  field, f64 fn param, f64 method param, `let c: f64 = 3 + 1`. Measured `i1`→2, `g7`→100, `g13`→3,
  matching its claims exactly. Every **f32** twin passes because `typecheck.ax:5208` hints `TYPE_F32`
  only ⇒ **the drifted-copy class AGAIN**. That is now BACKLOG task 0, with oracle rows **4, 8, 9, 11
  reserved** so it lands without renumbering. It has real fixpoint exposure ⇒ full gate.

⇒ **The drifted-copy defect class has now produced FOUR bugs in two sessions** (interface return type,
generic explicit type args, `lower_int_lit` vs `lower_float_lit`, and the `TYPE_F32`-only hint). "One
walk, one predicate" is now written into the brief for every fix in this family.

## ⏳ IN FLIGHT — fix for probe4 bug #2 (dispatch drops argument coercion)
Dispatched with the measured control matrix and an explicit instruction **not to add a second walk
over `NODE_INTERFACE_DECL`** — extend `interface_method_sig` (which already walks it correctly) with a
parameter-type accessor and use it at `typecheck.ax:4336`. Frontend ⇒ **A==B**, baseline **597/597**
including the `-O0` pass. If no commit exists, re-dispatch rather than assume.

## ⭐ LESSON — the "transient flake" was OUR OWN concurrency, and calling it a flake was the error
The arrwalk agent hit 592/593 (`t_localtuplenoinit@-O0`), passed on rerun, and filed it as a transient
flake. It was not. Measured: that test gives **exit 12, six times, byte-identical binaries** — and 12
is exactly what the suite expects. The real cause was **two sub-agents compiling into `bin/` at the
same time** (an exe being rewritten while another process launches it), i.e. a consequence of my own
parallel dispatch — which delegate-by-default makes the *common* case, not a one-off.
    Two rules added to CLAUDE.md §24: **serialize anything that writes `bin/` or runs the suite**
(parallelise only read-only investigation), and ⛔ **never file an unreproducible failure as a
"flake"** — §3 makes determinism absolute, so a one-off is a **bug report until attributed to a named
cause**. "It passed on rerun" is a claim about the rerun, the same shape as "its caller copes with the
degenerate value", which hid a real miscompile for six days.

## Next, after those land
See [[BACKLOG]]. The two D1-class items still waiting on the **user**, deliberately not decided:
**fib is undecidable by the M6 gate as written** (its ratio's denominator is bimodal, spread 17.2%/
13.7% > the whole 15% margin) and **`let x: u8 = 300` keeps 300 with no diagnostic**
([[question-out-of-range-narrow-int-literal]]; recommendation if asked: REJECT).
