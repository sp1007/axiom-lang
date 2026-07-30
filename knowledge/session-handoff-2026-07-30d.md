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

## ⏳ IN FLIGHT when this was written — check these first
Two sub-agents were dispatched and had not reported yet. If a resumed session finds no commits from
them, **re-dispatch rather than assume**:
1. **`axiom-implementer`: arrwalk layout-distribution reading** in `scripts/perf_m6_gate.ps1` (needs
   a global array, so the `hot()` wrapper template must be extended). Last recorded debt of the M6
   measurement protocol. Zero compiler risk. Constraints handed to it: copy the NASM floor verbatim,
   report the startup floor next to every number, verdict from **SE of the mean propagated into the
   ratio** (not raw spread), and re-run xorshift as a control (must still read ≈0.995x).
2. **`axiom-investigator`: probe unswept surfaces** into `bin/probe4/`. Brief was to exploit the
   pattern behind all three of 07-30c's finds: **a surface where the suite exercises only ONE type
   class — then vary the class** (float in XMM vs GPR; str/bytes >8 B). Directions given:
   Option/Result payloads of non-i64, generic struct fields at f32/f64/str, str through interfaces,
   arrays of floats/structs, globals × aggregates × float, f32 casts/mixed-width.

## Next, after those land
See [[BACKLOG]]. The two D1-class items still waiting on the **user**, deliberately not decided:
**fib is undecidable by the M6 gate as written** (its ratio's denominator is bimodal, spread 17.2%/
13.7% > the whole 15% margin) and **`let x: u8 = 300` keeps 300 with no diagnostic**
([[question-out-of-range-narrow-int-literal]]; recommendation if asked: REJECT).
