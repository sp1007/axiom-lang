---
name: infra-defender-build-throttle
description: "Windows Defender real-time scanning throttles AXIOM builds hard (~80s-1.5min PER test compile on 2026-07-18, turning a ~10-min full regression into ~5hr). It is the binding constraint on all build-gated work. Mitigation is user-side: add a Defender exclusion for the repo build dirs."
metadata:
  node_type: memory
  type: project
---

# Windows Defender throttles the build/gate loop (workflow-binding)

**Observed 2026-07-18:** each regression test compile (bundle ~123k stdlib + codegen +
self-link → new .exe) took **~80s–1.5min** with Defender real-time protection scanning
every freshly-emitted exe. A full ~396-test regression that normally runs ~10 min took an
estimated **~5 hours** at this rate. This is NOT a compiler slowdown — the driver hash was
unchanged and every test PASSed; it is pure AV scan latency on each new binary.

**Impact:** this is the binding constraint on ALL build-gated progress right now —
fixpoint hops, full regression, and any change that needs the A==B / B==C gate. The one
remaining bounded bug (m2 call-non-fn, strategy prepped in
[[bug-malformed-input-robustness-cluster]]) is blocked on it: the fix is understood but
can't be gated in reasonable wall-time on a Defender-throttled box.

**Mitigation — RESOLVED 2026-07-18.** User added Defender folder exclusions for
`d:\projects\compiler\Axiom\bin` + `...\bootstrap\stage1` → build to those dirs dropped
from ~99s to **~1s**. ⚠️ BUT the regression harness built test exes to hardcoded `/tmp`
(= `C:\Users\sp\AppData\Local\Temp`, NOT excluded) → still ~99s each. FIXED in
`regression_repros.sh` (`61768b5`): build dir is now `$REGTMP` (default `/tmp`, unchanged
for CI/Linux); on this box run **`REGTMP=bin/_regtmp bash scripts/regression_repros.sh`**
(bin/ is excluded → ~1s/build, full 396-test run in minutes). `bin/_regtmp/` is gitignored.
(USER-SIDE note kept for reference: Claude cannot self-apply Defender settings — same
privilege-escalation caveat as CLAUDE.md §24. Windows Security → Virus & threat protection
→ Manage settings → Exclusions.)

🐞 **SEPARATE gotcha found same session (`61768b5`):** the harness DEFAULT compiler was the
**stale `bin/axc_stage1.exe`** (bootstrap seed, hash `448cd6c7`, predates recent aggregate
fixes), NOT the daily driver `axc_native.exe` (`1c2e3d6a`). A plain `bash
scripts/regression_repros.sh` therefore tested the OLD compiler → spurious struct/Option
SIGSEGV "failures" (an old compiler genuinely miscompiling aggregates the current driver
handles). Default is now `AXC=bin/axc_native.exe` (the gate compiler per build_native.ps1:67).
LESSON: always gate on `axc_native`; if you see aggregate-only failures, FIRST check which
compiler ran + reproduce standalone with `bin/axc_native.exe` before suspecting a regression.

**Operating note until then:** run regressions/fixpoints in the BACKGROUND (they finish
async and notify), commit additive/independently-verified changes without waiting hours,
and reserve build-gated compiler changes for when the box is quiet. Prior sessions saw the
same pattern under browser/load pressure ([[next-step-16-fnptr-shipped]] env note: trivial
self-link `return 5` took 84s vs ~5-10s normal).

⚠️ Concurrency trap (hit 2026-07-18): running TWO regressions at once (e.g. a redundant
re-run alongside the first) makes them contend on the shared build artifacts
(`bin/*.exe`, `tmp_concatenated_air.ax`) → spurious `FAIL <t> (build produced no exe)`
that is NOT a real regression. Also, `TaskStop` on a git-bash regression may leave orphan
child bash/axc processes still looping and spawning compiles — verify with
`Get-CimInstance Win32_Process -Filter "Name='bash.exe'"` (check CreationDate/parent) and
kill the whole tree before re-running. Only ever run ONE regression at a time.

## Concurrency trap, concrete instance (2026-07-22)

Running `scripts/exe_size_check.sh` and `scripts/elf_linux_check.sh` WHILE
`regression_repros.sh` was running produced a **spurious `t_strsplit` failure**
(`508 passed, 1 failed`). Re-running `t_strsplit` alone gave the expected 46 five times out
of five, and a clean solo regression was green.

Root cause was not merely "two suites at once": `exe_size_check.sh` defaulted its scratch
directory to `${REGTMP:-bin/_regtmp}` — the *same* directory the regression suite uses. It
now defaults to `bin/_sizetmp` (`$SIZETMP` to override).

**Rule:** run ONE build-driven suite at a time, and give every new suite its own scratch
directory. A red gate that arrives while another suite is running is suspect until
reproduced solo — but reproduce it, never assume it away.
