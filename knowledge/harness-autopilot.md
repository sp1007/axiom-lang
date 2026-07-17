---
name: harness-autopilot
description: "An autonomous agent harness (autopilot) is set up in .claude/ — triggers on \"tiếp tục\"/continue to run the compiler backlog autonomously without confirmation prompts."
metadata: 
  node_type: memory
  type: project
  originSessionId: f482e19c-5639-4d54-a2ca-f24d4c2c01aa
---

Harness built 2026-07-09 (user asked: run autonomously, stop asking for confirmation, grind backlog then re-plan). Lives in the repo under `.claude/`:

- **Agents** (`.claude/agents/`): `axiom-investigator` (read-only repro/root-cause + bug probing), `axiom-implementer` (scoped fixes), `axiom-verifier` (fixpoint+regression gate). All `model: opus`.
- **Skills** (`.claude/skills/`): `axiom-autopilot` (the orchestrator loop), `axiom-fixpoint-gate` (exact A==B vs B==C gate commands), `axiom-bug-probe` (proactive feature-combo probing).
- **CLAUDE.md §24** registers the pointer + change log.

**How it works:** on "tiếp tục"/continue, the `axiom-autopilot` skill runs Phase 0 (read CLAUDE.md + MEMORY.md + handoff, `git status`) → select next task (priority: in-flight → OPEN bugs → RFC follow-ups → `docs/next-step-*.md` queues → probing → milestones) → investigate → implement minimally → gate → auto-commit per the [[feedback-fixpoint-async-rule]] → loop; when empty, re-plan from milestones.

Encodes standing feedback: [[feedback-autonomous]], [[feedback-auto-commit]], [[feedback-fixpoint-async-rule]], [[feedback_compact]]. Sub-agent pattern (not TeamCreate — team tools unavailable in this runtime); spawn investigators in parallel for independent candidates, keep implementation+gating sequential (self-host context is heavy). Open backlog snapshot: [[backlog-open-items]].
