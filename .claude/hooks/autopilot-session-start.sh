#!/usr/bin/env sh
# SessionStart hook — AXIOM autopilot continuous-supervision rule.
# Emits a reminder (added to the session context) so that EVERY new session
# arms the persistent 5-minute Monitor heartbeat and never idle-hibernates.
# The hook cannot call Claude tools itself; it surfaces the directive so the
# agent acts on it deterministically at session start. Non-fatal if it fails.
cat <<'EOF'
[AUTOPILOT — CONTINUOUS SUPERVISION | CLAUDE.md §24]
New session detected. Before/around your first substantive action:
1. Check running tasks. If NO persistent autopilot Monitor is armed, arm ONE now:
   Monitor(persistent:true, command:'while true; do sleep 300; echo "[autopilot-tick] $(date -u +%H:%M:%S)"; done')
   It emits a heartbeat line every 5 minutes.
2. On each [autopilot-tick] AND whenever you would otherwise go idle / finish a
   turn: run Phase 0 (git status, in-flight work, OPEN bugs in MEMORY.md), then
   AUTO-SELECT the highest-value backlog task (in-flight > OPEN bugs > RFC
   follow-ups > docs task queues > proactive probing) and execute it via the
   `axiom-autopilot` skill — no confirmation prompts.
3. NEVER idle-hibernate. Keep the loop self-perpetuating. Only TaskStop the
   monitor when the user explicitly says to stop.
EOF
