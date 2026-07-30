---
name: axiom-autopilot
description: Autonomous execution loop for the AXIOM self-hosting compiler. Selects the optimal next task from the backlog (memory OPEN bugs → RFC follow-ups → milestone tasks), investigates, implements minimally, runs the fixpoint+regression gate, and auto-commits — WITHOUT asking for confirmation — then loops until the backlog is empty, after which it derives the next direction from the master plan. TRIGGER on "tiếp tục" / "continue" / "làm tiếp" / "chạy tự động" / "autopilot" / "next task" / "keep going" / "do the plan" / "hết task thì tự tìm hướng", or any request to run the project autonomously without step-by-step confirmation. Also re-triggers for "re-run / resume / update / improve the compiler autonomously".
---

# AXIOM Autopilot — Autonomous Compiler-Dev Loop

Purpose: run the AXIOM compiler project the way the user wants — **auto-decide the optimal approach and execute it, no confirmation prompts, grind the backlog to zero, then pick the next direction from the master plan.** This packages `do_plan.md` + the user's standing feedback (autonomous work, auto-commit, fixpoint-async gate) into a repeatable loop.

## Standing user directives (from memory — treat as law)
- **Do not ask for confirmation.** Choose the optimal direction and execute it. The user observes little; act.
- **Auto-run shell commands.** The gate/probe/build workflow (fast_fixpoint.ps1, regression_repros.sh, per-program compile/run, git) is shell-driven — run these bash/PowerShell commands automatically, never pausing to ask. (Enforcement caveat: the actual prompt-suppression is a `.claude/settings.json` `permissions.allow` setting the user must apply once — auto mode forbids Claude from self-granting permission rules. If prompts still appear, note it and keep going; do NOT reach for `-ExecutionPolicy Bypass` or other flagged security bypasses — run scripts directly via `& scripts/x.ps1`.) See CLAUDE.md §24 "Execution autonomy".
- **Auto-commit + push to `main`** after a change is GREEN (see the gate). Commit message ends with the Co-Authored-By trailer.
- **Fixpoint-async rule:** frontend-only change ⇒ commit after regression is GREEN, fixpoint can settle asynchronously. Backend/ABI/linker change ⇒ fixpoint (B==C) is MANDATORY *before* commit.
- **Auto-compact** context when near full; keep going across the summary.
- Prefer language ergonomics the user values (inference/sugar over boilerplate) when a design choice is open.

## Phase 0 — Context check (start of every run)
0. **Arm the continuous-supervision heartbeat FIRST (every new session).** Check running tasks; if NO persistent autopilot `Monitor` is armed, arm exactly ONE now:
   `Monitor(persistent:true, command:'while true; do sleep 300; echo "[autopilot-tick] $(date -u +%H:%M:%S)"; done')`.
   It emits a heartbeat line every 5 minutes. Each `[autopilot-tick]` — and any moment you would otherwise finish a turn and go idle — re-runs Phase 0 → Phase 1 and executes the next task WITHOUT asking. **Never idle-hibernate.** Only `TaskStop` the monitor when the user explicitly says to stop. (A SessionStart hook at `.claude/hooks/autopilot-session-start.sh` surfaces this same reminder; if the hook is installed the reminder arrives automatically, but arming the monitor is still THIS step's job — the hook cannot call tools.) See CLAUDE.md §24 "Continuous supervision".
1. Read `CLAUDE.md` (operating manual) + **`knowledge/BACKLOG.md`** (compact orientation: fixpoint hash, baseline count, OPEN items, refuted ideas) + the newest `knowledge/session-handoff-*.md`. ⛔ **NEVER read `knowledge/MEMORY.md` wholesale** — it is 175 KB ≈ 87k tokens and one truncated page costs ~25k tokens before any work starts; reach into it **only by `Grep`** for a named topic. It stays the detail store and the authority on any conflict (CLAUDE.md §24 "Token economy").
2. `git status` + `git log --oneline -5`. Determine: clean tree? uncommitted in-flight work? on `main` or a branch?
3. If there is uncommitted in-flight work from a prior session, resume it (gate → commit) before starting anything new.

## Phase 1 — Select the next task (deterministic priority)
Pick the single highest-value task, in this order:
1. **In-flight / regressions** — anything half-done or a broken gate. Finish/fix first.
2. **OPEN bugs in memory** — entries tagged OPEN (e.g. BUG#82 globals, `for x in <collection>` iteration). These are confirmed, high-value.
3. **RFC follow-ups** — unfinished phases of shipped RFCs (`rfcs/*.md` with P2/P3 pending) — e.g. RFC 0009-P3 (ELF export), RFC 0014 (drop-glue, blocked on BUG#69), RFC 0015-P2/P3 (CTGC escape/ctgc).
4. **`docs/next-step-*.md` TASK QUEUEs** — these hold larger feature queues (esp. `next-step-15.md` + `next-step-15-sub-1.md`). Genuinely-open items there: multi-field variant `Rect(i64,i64)` (currently only *rejected* per BUG#81 → real support is backlog), str/>8B variant payload, `Self` type (RFC), rewriting aspirational std modules (iter/json/log/net/fmt) to grammar, jump-table dispatch optimization (RFC). **Cross-check every item against MEMORY.md + git before doing it — most of these queues are already shipped (RFC 0002–0007, ADT v1); do not redo completed work.**
5. **Backlog refill** — if 1–4 are empty, run the `axiom-bug-probe` skill (proactive feature-combo probing) to surface new silent-miscompile bugs.
6. **Master-plan tasks** — `docs/tasks/pXX-tYY-*.md` (see `docs/tasks/milestones.md`). Most early phases are done (compiler self-hosts); target the next unmet milestone gate.

State the chosen task in one line and proceed — do not ask which one.

## Phase 2 — Investigate
For a bug/feature, spawn `axiom-investigator` (read-only) to reproduce, localize to `file:line`, check the spec/RFC (strange behavior is often intended design — do NOT "fix" design), and propose a minimal fix + oracle. Spawn **multiple investigators in parallel** only for independent candidates. For a task whose design is already clear from memory, skip to Phase 3.

## Phase 3 — Implement
**Delegate to `axiom-implementer` BY DEFAULT** (CLAUDE.md §24 "Token economy"): a sub-agent runs on a fresh context and returns only its report, so one task's file reads, build logs and disassembly dumps are not paid for by every later task — that is functionally a per-task `/clear`. Implement **inline only as the exception**, when the change is genuinely self-host-critical and needs the orchestrator's accumulated diagnosis (backend/ABI/regalloc mid-investigation). ⚠️ Delegate a WHOLE task, never a two-tool errand — each sub-agent re-reads its own orientation. Rules: minimal first, respect pipeline layering, add an oracle test, RFC for any syntax/IR/ABI/linker change. Never edit `tmp_concatenated_air.ax` directly (it is generated).

## Phase 4 — Gate (invoke the `axiom-fixpoint-gate` skill, or spawn `axiom-verifier`)
- Rebuild the daily driver if source changed.
- Frontend-only ⇒ fast fixpoint **A==B** + full regression.
- Backend/linker/ABI ⇒ **A!=B is expected**; hand-build and require **B==C**, + full regression + oracles.
- RED ⇒ loop back to Phase 3 with the evidence. Never commit RED.

## Phase 5 — Commit (auto, no confirmation)
On GREEN, commit per the fixpoint-async rule and push to `main`. Update the relevant memory file (bug status → FIXED, new fixpoint hash, baseline count) and add a one-line pointer in `MEMORY.md`. If the daily driver changed, note the new hash.

## Phase 6 — Loop / re-plan
Go back to Phase 1. When Phases 1–4 of selection are all empty (no OPEN bugs, no RFC follow-ups, probing yields nothing after ~15+ batches), **derive the next direction**: read `docs/tasks/milestones.md` + `AXIOM SPECIFICATION/14. implementation plan.md`, identify the nearest unmet milestone gate, decompose it into the next concrete task, write it to memory as the new focus, and continue the loop. Use the `Plan` built-in agent for this survey if helpful.

## Guardrails (do not violate even in autopilot)
- Determinism, pipeline isolation, and the RFC policy from CLAUDE.md are absolute. Autonomy speeds *how* you work, not *what rules* you follow.
- Do not weaken tests to pass. Do not silently change semantics. Do not skip the backend fixpoint gate.
- Destructive/outward actions beyond commit/push-to-main (force-push, history rewrite, deleting others' work) still require explicit user approval.
- If you hit a true hard blocker (needs a product decision only the user can make), record it in memory and move to the next backlog item rather than stalling.

## Test scenarios
- **Normal:** user says "tiếp tục" → Phase 0 finds clean tree at fixpoint + BUG#82 in-flight → resume BUG#82 → gate B==C → commit → pick next OPEN item → loop.
- **Backlog empty:** no OPEN bugs, probing yields nothing → Phase 6 reads milestones, finds nearest unmet gate, decomposes it into a task, continues.
- **RED gate:** implement → verifier RED (regression fails) → loop to implementer with evidence → re-gate → GREEN → commit.
