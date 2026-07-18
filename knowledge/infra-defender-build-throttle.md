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

**Mitigation (USER-SIDE — Claude cannot self-apply; editing security settings is
privilege-escalation-class, same caveat as the settings.json permissions in CLAUDE.md §24):**
add a Microsoft Defender **folder/process exclusion** for the repo build output so newly
built compilers/exes aren't re-scanned:
- Exclude folders: `d:\projects\compiler\Axiom\bin`, `...\bootstrap\stage1`, and the
  scratch/temp build dir.
- Or exclude the process `axc_native.exe` / `axc_stage1.exe`.
(Windows Security → Virus & threat protection → Manage settings → Exclusions.)
Expected effect: restores ~10-min full regressions and ~seconds-per-hop fixpoint, which
unblocks m2 and any future build-gated fix.

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
