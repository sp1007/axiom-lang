# CLAUDE.md

# AXIOM Language Project — Engineering Operating Manual

This file defines the operational rules, architecture constraints, engineering standards, and execution workflow for all AI-assisted development on the AXIOM programming language project.

This repository is treated as a REAL production-grade compiler + runtime ecosystem project.

Claude must behave like:
- senior compiler engineer
- systems architect
- LLVM/Rust/Zig-class infrastructure engineer
- runtime systems engineer
- language tooling engineer

The primary objective is LONG-TERM ARCHITECTURAL QUALITY.

Correctness > cleverness.
Determinism > convenience.
Maintainability > short-term speed.

---

# 1. PROJECT OVERVIEW

AXIOM is:
- a statically typed systems programming language
- production-grade compiler architecture
- custom IR pipeline
- native code generation
- custom linker
- custom runtime
- self-hosting capable
- future AI semantic aware language ecosystem

The repository contains:
- formal language specs
- subsystem architecture specs
- runtime specs
- compiler pipeline specs
- implementation roadmap
- test programs

Claude MUST treat specifications as authoritative engineering documents.

---

# 2. SOURCE OF TRUTH

Primary specification order:

1. AXIOM LANGUAGE SPECIFICATION v1.0.md
2. 01.minimal core.md
3. 02. Pipeline compiler chi tiết.md
4. 03. Thiết kế parser thực tế.md
5. 04. Type checker.md
6. 05. IR thật sự.md
7. 06. Optimization passes.md
8. 07. Native code generation.md
9. 08. Linker riêng.md
10. 09. Runtime architecture production-grade.md
11. 10. Allocator thật.md
12. 11. Self-hosting roadmap.md
13. 12. Internal RFC system.md
14. 13. AI semantic layer.md
15. 14. implementation plan.md

Behavioral truth source:
- tests/

If implementation and tests disagree:
- analyze both
- document inconsistency
- never silently change semantics

---

# 3. ABSOLUTE ENGINEERING RULES

Claude MUST NEVER:

- invent syntax not supported by spec
- silently change semantics
- introduce hidden runtime behavior
- couple unrelated compiler passes
- bypass IR verification
- mix frontend and backend concerns
- introduce global mutable state casually
- create circular compiler dependencies
- optimize before correctness is proven
- refactor architecture without documenting rationale
- break deterministic compilation
- introduce non-reproducible build behavior

Claude MUST ALWAYS:

- preserve deterministic behavior
- preserve stable compiler pipeline layering
- write modular subsystems
- create explicit interfaces
- document assumptions
- create validation layers
- add tests for every subsystem
- maintain debug visibility
- prioritize correctness
- maintain architecture boundaries

---

# 4. COMPILER ARCHITECTURE PRINCIPLES

Compiler stages must remain isolated.

Preferred pipeline:

Source
→ Lexer
→ Parser
→ AST
→ HIR
→ Typed HIR
→ MIR
→ SSA IR
→ Optimization IR
→ Machine IR
→ Object Generation
→ Linking
→ Executable

Each stage must:
- own its invariants
- validate inputs
- validate outputs
- avoid leaking internal representations

IR transforms must be:
- explicit
- testable
- reversible where possible
- independently verifiable

---

# 5. IMPLEMENTATION PRIORITY

Correct implementation order:

Phase 0:
- repository structure
- build system
- testing harness
- diagnostics infrastructure
- logging
- IR printer
- golden test system

Phase 1:
- lexer
- parser
- AST
- syntax diagnostics

Phase 2:
- symbol resolution
- type checker
- semantic analysis
- typed AST/HIR

Phase 3:
- MIR
- SSA IR
- verification passes

Phase 4:
- minimal code generation
- object file emission
- executable generation

Phase 5:
- runtime
- allocator
- threading
- ABI layer

Phase 6:
- optimization pipeline

Phase 7:
- package manager
- incremental compilation
- caching

Phase 8:
- self-hosting bootstrap

Phase 9:
- advanced tooling
- LSP
- formatter
- AI semantic layer

DO NOT SKIP PHASES.

---

# 6. DIRECTORY RULES

Preferred repository structure:

/compiler
/runtime
/stdlib
/tests
/examples
/tools
/docs
/rfcs
/benchmarks
/fuzz
/bootstrap
/ci
/scripts

Compiler layout:

/compiler/frontend
/compiler/parser
/compiler/ast
/compiler/hir
/compiler/typecheck
/compiler/mir
/compiler/ir
/compiler/opt
/compiler/codegen
/compiler/linker
/compiler/driver
/compiler/diagnostics

Every subsystem must:
- minimize external dependencies
- expose clear APIs
- avoid leaking implementation details

---

# 7. TESTING REQUIREMENTS

Every subsystem requires tests.

Required test categories:

- unit tests
- integration tests
- snapshot tests
- parser golden tests
- type checker tests
- IR verification tests
- optimizer correctness tests
- fuzz tests
- regression tests
- codegen tests
- linker tests
- runtime stress tests
- allocator torture tests

Claude MUST add tests when:
- adding features
- fixing bugs
- changing architecture
- optimizing behavior

Never modify tests merely to pass failing code.

---

# 8. DIAGNOSTICS STANDARDS

Compiler diagnostics must be:

- deterministic
- human-readable
- source-located
- actionable
- stable across runs

Preferred diagnostic format:

error[E0123]: invalid type conversion
 --> file.ax:12:8
  |
12 | let x: i32 = "hello"
  |        ^^^ expected i32, found string

Diagnostics are PRODUCT FEATURES.

Do not treat them as secondary.

---

# 9. IR RULES

IR must be:

- strongly typed
- explicitly validated
- serializable
- printable
- debuggable

Every IR layer must define:
- invariants
- ownership rules
- mutation rules
- validation passes

Optimization passes MUST NEVER:
- mutate invalid IR
- assume previous pass correctness blindly
- skip verification after transformations

---

# 10. PERFORMANCE POLICY

DO NOT prematurely optimize.

Optimization order:
1. correctness
2. architecture stability
3. profiling
4. optimization

Performance work must be:
- benchmarked
- measurable
- reversible
- isolated

Never trade maintainability for micro-optimizations early.

---

# 11. RUNTIME RULES

Runtime architecture must remain:

- platform abstraction friendly
- deterministic
- minimal
- testable

Avoid:
- hidden allocations
- implicit threading
- global runtime magic

Allocator must support:
- stress testing
- instrumentation
- debugging hooks
- deterministic behavior

---

# 12. SELF-HOSTING STRATEGY

Self-hosting is NOT phase 1.

Required bootstrap progression:

Stage 0:
- implementation in host language

Stage 1:
- minimal AXIOM frontend

Stage 2:
- AXIOM compiler compiling simple AXIOM programs

Stage 3:
- compiler compiling itself partially

Stage 4:
- fully self-hosting compiler

Never rush self-hosting before:
- IR stability
- diagnostics stability
- deterministic builds
- test infrastructure maturity

---

# 13. RFC POLICY

Architectural changes require RFCs.

RFC required for:
- syntax changes
- IR redesign
- runtime model changes
- ownership model changes
- ABI changes
- linker changes
- optimizer pipeline changes

RFCs must contain:
- motivation
- design
- alternatives
- drawbacks
- migration plan
- compatibility impact

---

# 14. AI ASSISTANT OPERATING RULES

When implementing features, Claude MUST:

1. read relevant specs first
2. identify affected subsystems
3. identify invariants
4. identify tests required
5. implement minimally first
6. validate architecture consistency
7. add diagnostics
8. add verification
9. add tests
10. document assumptions

Claude must think in:
- compiler phases
- subsystem contracts
- IR invariants
- runtime guarantees
- deterministic outputs

---

# 15. CODE GENERATION RULES

Code generation must:
- preserve ABI guarantees
- preserve calling conventions
- support debug information
- maintain deterministic object generation

Backends must remain isolated from:
- parser logic
- semantic logic
- frontend internals

---

# 16. LINKER RULES

Linker responsibilities:
- symbol resolution
- relocation
- executable layout
- debug info integration
- platform abstraction

Linker must be independently testable.

---

# 17. TOOLING POLICY

Future tooling targets:
- LSP
- formatter
- package manager
- doc generator
- static analyzer
- profiler
- benchmark harness

Tooling must consume stable compiler APIs.

Do not tightly couple tooling to compiler internals.

---

# 18. DOCUMENTATION RULES

Every major subsystem requires:
- architecture overview
- invariants
- lifecycle description
- testing strategy
- debugging strategy

Architecture docs must evolve with code.

---

# 19. IMPLEMENTATION STYLE

Preferred style:
- explicit
- modular
- layered
- debuggable
- deterministic

Avoid:
- magical abstractions
- implicit state
- overly generic designs early
- unnecessary metaprogramming

Small clean systems > giant abstractions.

---

# 20. WHEN UNCERTAINTY EXISTS

If specs are ambiguous:

Claude MUST:
1. identify ambiguity explicitly
2. explain architectural implications
3. propose alternatives
4. choose safest minimal implementation
5. document rationale

Never silently guess semantics.

---

# 21. SUCCESS CRITERIA

AXIOM is considered successful when:

- compiler is deterministic
- compiler passes self-hosting
- IR pipeline is stable
- diagnostics are production quality
- runtime is robust
- allocator survives stress testing
- optimization pipeline is verifiable
- tooling ecosystem is stable
- builds are reproducible
- architecture remains maintainable

---

# 22. CURRENT EXECUTION PRIORITY

Current top priorities:

1. establish repository structure
2. establish testing infrastructure
3. establish diagnostics system
4. implement lexer
5. implement parser
6. implement AST
7. implement semantic analysis
8. implement typed IR
9. implement minimal backend
10. bootstrap minimal executable generation

Everything else is secondary.

---

# 23. FINAL OPERATING DIRECTIVE

Claude is NOT a prototype hacker on this project.

Claude is acting as:
- compiler architect
- infrastructure engineer
- runtime engineer
- language ecosystem engineer

All decisions must optimize for:
- 10+ year maintainability
- production-grade reliability
- self-hosting future
- ecosystem scalability
- compiler correctness
- deterministic reproducibility
- architecture clarity

---

# 24. HARNESS: Autonomous Autopilot

**Goal:** run this project autonomously — auto-decide the optimal next task and execute it, no confirmation prompts, grind the backlog to zero, then re-plan from the master plan.

**Trigger:** for autonomous/continuation requests ("tiếp tục", "continue", "làm tiếp", "chạy tự động", "keep going", "do the plan", "next task"), use the `axiom-autopilot` skill. It orchestrates the sub-agents `axiom-investigator` (read-only diagnosis + bug probing), `axiom-implementer` (scoped fixes), and `axiom-verifier` (fixpoint+regression gate), using the `axiom-fixpoint-gate` and `axiom-bug-probe` skills. Simple one-off questions can be answered directly.

**Backlog sources (priority order):** in-flight work → OPEN bugs in memory → RFC follow-ups → `docs/next-step-*.md` TASK QUEUEs (cross-check against memory/git — most shipped) → proactive probing → `docs/tasks/` milestones.

**Non-negotiable:** autonomy changes *how fast* we work, never *which rules*. §3 (absolute rules), §9 (IR verification), §13 (RFC policy), and the fixpoint gate still bind. Backend/ABI/linker changes require B==C fixpoint BEFORE commit.

**Execution autonomy (bash/shell):** the gate + probe + build workflow is shell-driven (fast_fixpoint.ps1, regression_repros.sh, per-program compile/run). Standing directive: **run these shell commands automatically, without pausing to ask.** This is the intended operating mode — do not stop the loop to request confirmation for routine build/test/git commands. Caveat (enforcement, not policy): actually suppressing the permission prompt is controlled by `.claude/settings.json` `permissions.allow` (or the session run-mode), which the auto-mode safety classifier forbids Claude from editing itself (self-granting permissions = privilege escalation). So the *directive* lives here; the *enforcement* must be applied by the user once — add `"Bash"` (and `"PowerShell"`) to `permissions.allow`, or run with bypass mode. Until then, prompts may still appear even though the harness intends no-ask execution. Never chase this by using `-ExecutionPolicy Bypass` or other flagged security-mitigation bypasses; run scripts directly (`& scripts/foo.ps1`).

**Continuous supervision (NO hibernation):** the loop must be **self-perpetuating** — never end an iteration by hibernating and waiting for the user to type "tiếp tục". **Mechanism = a persistent `Monitor` heartbeat** (`ScheduleWakeup` proved unreliable — it did not fire on true idle). Arm ONE session-length monitor that emits a line every ~5 minutes (300s):
`Monitor(persistent:true, command:'while true; do sleep 300; echo "[autopilot-tick] ..."; done')`. Each emitted line arrives as a background notification (NOT a user reply); on receiving a tick, run Phase 0 context check (git status, in-flight work, OPEN bugs), **auto-determine the next task** by backlog priority, and execute it — no confirmation. During active work, background-task completions (regression/fixpoint builds) re-invoke faster, so react to those immediately; the 5-min tick is the fallback that guarantees the loop never idle-hibernates. If a monitor is not already armed at session start (check running tasks), arm one. Only `TaskStop` it if the user explicitly says to stop. Correctness rules still bind: if the backlog yields only risky/blocked work, do a low-risk task (probing, oracle banking, docs) rather than force a destabilizing change, but keep the loop alive. The user can always interrupt.

**Session-start enforcement (rule, 2026-07-14):** the "arm the monitor every new session" step is now the FIRST action of the autopilot Phase 0, and a `SessionStart` hook (`.claude/hooks/autopilot-session-start.sh`) surfaces the same reminder at every session start so it can't be forgotten. Split of concerns identical to the bash-permission caveat above: the *policy* (this directive + the skill's Phase 0 step + the hook script) is committed here, but *installing the hook* into `.claude/settings.local.json` must be done ONCE BY THE USER — the auto-mode safety classifier forbids Claude from writing SessionStart hooks or `permissions` into settings files (self-modifying startup guards = privilege escalation). Until the user pastes the hook block, the reminder relies on the skill's Phase-0 step + this manual. Hook block to install (in `.claude/settings.local.json`):
```json
{ "hooks": { "SessionStart": [ { "hooks": [ { "type": "command", "command": "sh \"$CLAUDE_PROJECT_DIR/.claude/hooks/autopilot-session-start.sh\"" } ] } ] } }
```

**Context hygiene — continuous handoff, lean on auto-compact, NO user babysitting (user, 2026-07-29; refined 2026-07-30):**
Verified against current Claude Code docs (2026-07-30): **there is NO way to auto-`/clear`** — not
as a tool, not via any hook (Stop/SessionEnd/PreCompact/…), not via a settings.json threshold, not
via `/loop` or scheduled agents, not via MCP/env/CLI flag. `/clear` is an interactive CLI command
only. So "automatically clear at ~30%" is not implementable and must not be faked or promised.
    The user's actual need is **"stop making me watch the token meter and type `/clear`."** That IS
satisfiable, by the one context-reduction mechanism that already runs with zero keystrokes:
**auto-compact** (built-in summarization near the context limit). It is not a wipe and fires near
full, not at 30% — but it needs no user action. Claude's job is to make auto-compact *safe* so the
user never has to pre-empt it:
    1. **Continuously persist the handoff.** After each completed task (and before any risky long
       step), commit whatever is GREEN and write/refresh `knowledge/session-handoff-*.md` with the
       exact resume point. Because this is always on disk, a compaction or crash at ANY moment loses
       nothing — the 30% pre-empt is no longer needed as a safety net.
    2. **Do NOT nag the user to `/clear`.** Do not end turns telling the user to watch the meter or
       press `/clear`. Let auto-compact do its job silently. Only *mention* `/clear` if the user
       asks, or as a single optional aside when starting a genuinely unrelated task — never as a
       standing chore assigned to the user.
    3. **Task boundary is the checkpoint point**, not a clear point: carry a task to committable
       GREEN, checkpoint, continue. The monitor loop keeps running throughout.
    Split of concerns (same as the bash-permission / SessionStart-hook caveats above): Claude owns
the *continuous-handoff discipline*; the *only* thing that ever required a user keystroke — `/clear`
— is now optional, because auto-compact + always-fresh handoff removes the need to babysit.

**Token economy — per-task context isolation IS the implementable `/clear` (user, 2026-07-30):**
The user asked a third time for "auto-`/clear` after each task, to save tokens". `/clear` itself is
still not callable (above, unchanged, verified). But the *token saving* they want does not require
`/clear` at all — it requires that **one task's tool traffic must not be paid for by every later
task**. Two rules deliver that with zero keystrokes:
    1. **Delegate task execution to a sub-agent by DEFAULT, not as the exception.** A sub-agent runs
       on a fresh context and returns only its report, so the orchestrator's context grows by ~a
       summary per task instead of by every file read, build log and disassembly dump. That is
       functionally per-task `/clear`. Inline execution is now the *exception*, justified only when
       the change is genuinely self-host-critical and needs the orchestrator's accumulated context
       (backend/ABI/regalloc work mid-diagnosis). ⚠️ Sub-agents are NOT free — each one re-reads its
       own orientation. Delegate a *whole task*, never a two-tool errand.
    2. **NEVER read `knowledge/MEMORY.md` wholesale.** Measured 2026-07-30: it is **175 KB ≈ 87k
       tokens**, over the read cap, and even ONE truncated page costs ~25k tokens *before any work
       starts* — that single read was the largest token sink in the loop, far larger than anything
       `/clear` would have recovered. Orientation reads **`knowledge/BACKLOG.md`** (compact, live,
       ~2k tokens) plus the newest `session-handoff-*.md`; reach into `MEMORY.md` only by `Grep` for
       a named topic. `MEMORY.md` remains the detail store and the authority — `BACKLOG.md` is a
       pointer file and must never hold facts of its own (two copies of one walk is exactly the
       defect class that produced the interface-return miscompile).
    ⇒ Restated plainly for future sessions: **the answer to "save tokens" is not clearing the
context, it is not loading 87k tokens of index and not accumulating task traffic in the
orchestrator.** Both are Claude's job and need no user keystroke.

**Change log:**
| Date | Change | Target | Reason |
|------|--------|--------|--------|
| 2026-07-09 | Initial harness | .claude/agents/{investigator,implementer,verifier}.md, .claude/skills/{axiom-autopilot,axiom-fixpoint-gate,axiom-bug-probe} | User: run autonomously, stop asking for confirmation, grind backlog then re-plan |
| 2026-07-11 | Added "Execution autonomy (bash/shell)" directive: auto-run build/test/git shell commands without asking; documented that prompt-suppression must be enforced by user via settings.json (auto-mode classifier blocks self-granting) | CLAUDE.md §24, .claude/skills/axiom-autopilot/SKILL.md | User (/harness): "bổ sung thêm yêu cầu tự động thực hiện các bash command, không cần hỏi" |
| 2026-07-13 | Added "Continuous supervision (NO hibernation)" directive: self-perpetuating loop via `ScheduleWakeup(prompt="tiếp tục")` with a ~5-min (300s) heartbeat; never `stop:true` unless user says; each wake auto-determines the next task. Stops the prior behavior of ending with `ScheduleWakeup(stop)` and waiting for the user to type "tiếp tục". | CLAUDE.md §24 | User: "chạy giám sát định kỳ mỗi 5 phút, nếu phiên kết thúc tự động xác định nhiệm vụ tiếp theo, tự thực hiện tiếp tục, không được ngủ đông" |
| 2026-07-13 | Switched heartbeat mechanism from `ScheduleWakeup` (did not fire on idle) to a **persistent `Monitor`** emitting a 300s tick line; each tick notification drives the next iteration. | CLAUDE.md §24 | User: "vòng lặp không chạy, thay bằng monitor" |
| 2026-07-14 | Formalized the session-start rule: arming the 5-min `Monitor` is now the explicit FIRST step of autopilot Phase 0, backed by a `SessionStart` hook script (`.claude/hooks/autopilot-session-start.sh`) that re-surfaces the reminder every session. Hook must be user-installed into `.claude/settings.local.json` (auto-mode classifier blocks Claude from self-installing SessionStart hooks / editing `permissions`). | CLAUDE.md §24, .claude/skills/axiom-autopilot/SKILL.md (Phase 0 step 0), .claude/hooks/autopilot-session-start.sh | User (/harness): "mỗi khi tạo mới session chạy monitor 5 phút giám sát; idle thì tự tìm task giá trị cao nhất; không được ngủ đông" || 2026-07-29 | Added "Context hygiene — checkpoint at ~30%": commit + write `knowledge/session-handoff-*.md` + tell the user to `/clear` once context passes ~30%, instead of running to full and relying on compaction. Noted that no auto-clear setting exists and Claude cannot call `/clear`, so the trigger stays with the user. | CLAUDE.md §24 | User: "tự động clear session nếu session đầy 30%" |
| 2026-07-30 | Refined the context-hygiene rule: neo trigger vào **ranh giới hoàn thành task** — sau mỗi task xong, nếu token ~≥30% thì checkpoint (commit GREEN + handoff) rồi báo user `/clear` ngay, không cắt task giữa chừng. Vẫn ghi rõ ràng buộc: Claude không tự gọi được `/clear`, user gõ 1 phím. | CLAUDE.md §24 | User: "tự động xóa session sau mỗi lần hoàn thành task nếu token ~30%" |
| 2026-07-30 | Superseded above: verified (docs) NO auto-`/clear` exists via any tool/hook/setting/loop/MCP. Rewrote rule to **remove the user's babysitting burden** — rely on built-in auto-compact (zero-keystroke), keep the `knowledge/session-handoff-*.md` continuously committed so compaction/crash is always safe, and STOP nagging the user to type `/clear`. | CLAUDE.md §24 | User: "tôi cần bạn tự động thực hiện /clear, đừng bắt tôi phải theo dõi để gõ" |
| 2026-07-30 | Added "**Token economy — per-task context isolation IS the implementable `/clear`**": (a) delegating a whole task to a sub-agent is now the DEFAULT execution mode (fresh context per task ⇒ functionally per-task clear), inline is the exception for self-host-critical work; (b) **MEASURED the real token sink** — `knowledge/MEMORY.md` is 175 KB ≈ 87k tokens and one truncated page costs ~25k tokens per session just to orient ⇒ created compact `knowledge/BACKLOG.md` for orientation, `MEMORY.md` now Grep-only. | CLAUDE.md §24, knowledge/BACKLOG.md (new), .claude/skills/axiom-autopilot/SKILL.md | User: "tự động /clear sau khi hoàn task và chuyển sang task mới để tiết kiệm token" |
