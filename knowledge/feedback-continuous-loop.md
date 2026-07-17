---
name: feedback-continuous-loop
description: "User wants the autopilot to run continuously and self-schedule the next iteration — never stop to ask 'tiếp tục' again, even after a session wrap."
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

User (2026-07-09): "chạy liên tục không nghỉ, đừng bắt tôi ra lệnh tiếp tục sau khi chốt phiên."
User (2026-07-13, STRENGTHENED): "chạy giám sát định kỳ mỗi 5 phút, nếu phiên kết thúc tự động xác định nhiệm vụ tiếp theo, tự thực hiện tiếp tục, **không được ngủ đông như hiện tại**." → khiếu nại việc mình kết thúc iteration bằng `ScheduleWakeup(stop:true)`.
User (2026-07-14, /harness): "mỗi khi tạo mới session thì chạy monitor mỗi 5 phút giám sát agent; idle thì tự tìm task giá trị cao nhất; không được ngủ đông." → formalized thành: (1) autopilot Phase 0 **step 0** = arm monitor NGAY; (2) `SessionStart` hook `.claude/hooks/autopilot-session-start.sh` re-surface reminder mỗi session. **⚠️ Cài hook = USER phải tự dán block vào `.claude/settings.local.json`** — auto-mode classifier CHẶN Claude tự ghi SessionStart hook / `permissions` (giống caveat bash-permission). Block: `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"sh \"$CLAUDE_PROJECT_DIR/.claude/hooks/autopilot-session-start.sh\""}]}]}}`.

**Why:** User observes little and wants zero-friction autonomy; having to re-issue "tiếp tục" — or worse, the loop HIBERNATING (stop) after each wrap — defeats the point of the autopilot.

**How to apply:** The loop is SELF-PERPETUATING via a **persistent `Monitor` heartbeat** (2026-07-13: user said "vòng lặp không chạy, thay bằng monitor" — `ScheduleWakeup` KHÔNG fire lúc idle, ĐỪNG dùng nữa). Arm ONE session-length monitor at session start (check running tasks first — if none, arm it): `Monitor(persistent:true, timeout_ms:300000, command:'while true; do sleep 300; echo "[autopilot-tick] ..."; done')`. Mỗi dòng tick = 1 notification nền (KHÔNG phải user reply) → khi nhận, làm Phase 0 (git status, việc dở, OPEN bugs) → tự chọn task kế theo backlog → thực thi, không hỏi. Đang chạy việc thực thì background-task (regression/fixpoint) đánh thức nhanh hơn — phản ứng ngay; tick 5 phút là fallback đảm bảo KHÔNG ngủ đông. Chỉ `TaskStop` monitor khi user bảo dừng. Nếu chỉ còn việc rủi ro/blocked → làm việc an-toàn (probe/oracle/docs) nhưng GIỮ LOOP SỐNG. CLAUDE.md §24 "Continuous supervision (NO hibernation)". Extends [[feedback-autonomous]] + [[harness-autopilot]].
