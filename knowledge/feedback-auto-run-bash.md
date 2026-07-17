---
name: feedback-auto-run-bash
description: User wants build/test/git shell commands auto-run without prompts; enforcement is a settings.json permission the user must apply (auto mode blocks self-granting).
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 4a94d91d-722b-4c4c-9f37-689d97291884
---

User directive (2026-07-11, via `/harness`): **tự động thực hiện các bash command, không cần hỏi** — auto-run the gate/probe/build shell commands (fast_fixpoint.ps1, regression_repros.sh, per-program compile/run, git) without pausing for confirmation. Extends [[feedback-autonomous]] and [[feedback-auto-commit]].

**Why:** the whole AXIOM self-host loop is shell-driven; stopping to confirm each command defeats the "run autonomously" intent.

**How to apply:** run routine build/test/git shell commands automatically. BUT the actual prompt-suppression is enforced by `.claude/settings.json` `permissions.allow` (add `"Bash"` + `"PowerShell"`) or session bypass-mode — and the **auto-mode safety classifier BLOCKS Claude from editing permission rules itself** (self-granting = privilege escalation). So: document the directive (done — CLAUDE.md §24 "Execution autonomy" + axiom-autopilot skill, commit `2973032`), tell the user to apply the settings change once, and if prompts still appear just note it and keep going. **Never** work around prompts with `-ExecutionPolicy Bypass` or other flagged security-mitigation bypasses (also classifier-blocked) — run scripts directly via `& scripts/x.ps1`. Note: `.claude/settings.json` had an invalid `permissions.allow=["All"]` token (does nothing) — the valid tokens are `"Bash"`/`"PowerShell"`.
