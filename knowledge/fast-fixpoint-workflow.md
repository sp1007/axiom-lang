---
name: fast-fixpoint-workflow
description: "scripts/fast_fixpoint.ps1 — ~9s native-backend fixpoint gate (seed from axc_native, A→B→C, check B==C) instead of the hours-long gcc axc_stage1 path. Use for iterating on backend/codegen changes."
metadata: 
  node_type: memory
  type: project
  originSessionId: d798c715-682a-4cc0-8591-878e9b3d4235
---

**`scripts/fast_fixpoint.ps1` (added 2026-07-05, commit 6ef4176) — the fast gate
for backend/codegen iteration.** Regenerates the concat, then seeds from the native
daily-driver `bin/axc_native.exe`: axc_native builds A, A builds B, B builds C (all
`build ... -self-link -O1`), and checks **B==C** SHA-256 bit-identical. ~9s total
(each self-build ~4.5s) vs. the gcc `axc_stage1.exe` path which took 10000s+
(machine-sleep inflated; active time is only seconds — stage3 by contrast ran 4.3s).

**Reading the result correctly (IMPORTANT):**
- **A≠B is EXPECTED** after any codegen change — A is built by the OLD axc_native
  (old rules) from NEW source; B is built by A (new rules). They legitimately differ
  at every site the change touches. This is the one-time bootstrap transition, NOT a
  failure.
- **B==C is the REAL fixpoint** — both are built under the new rules, so equality
  means the new backend reproduces itself. That is the commit gate for backend/ABI
  changes (per [[feedback-fixpoint-async-rule]]).
- Also sanity-check that **B builds a small program correctly** (not just that it
  self-reproduces) — a compiler can be a stable fixpoint yet miscompile inputs.

**After a good fix:** `cp bin/axc_fpC.exe bin/axc_native.exe` to keep the daily
driver current, then clean `bin/axc_fp{A,B,C}.exe`.

**Gotcha — `ax_printf_local` silently eats `[D...` debug prints (found 2026-07-06,
[[bug74-generic-struct-inferred-ctor-args]]):** `print_helpers.ax::is_verbose_debug`
suppresses ANY string starting with `[` whose second char is `D`/`T`/`M`/`C`/`f` unless it
exactly matches one of a hardcoded `"[Debug] Stage"`/`"[Debug] Running"`/etc. whitelist —
temporary debug prints like `"[DBG] ..."` or `"[DBG-CALL] ..."` print NOTHING, which looks
EXACTLY like "this code path never executes" (the same symptom as a real dormant-code bug
like [[bug69-ctgc-ownership-escape-noop]]). **Always prefix ad-hoc debug prints with
something that does NOT start with `[D`/`[T`/`[M`/`[C`/`[f`** — e.g. `"XTRACE ..."` — and
confirm the print fires before concluding a branch is unreachable.

**Gotchas on this Windows box:**
- Run PowerShell scripts as `& scripts/x.ps1` (NOT `powershell -ExecutionPolicy
  Bypass ...`, which the auto-mode classifier denies as "Security Weaken").
- `bin/*.exe` is gitignored; `bin/*.ax` are tracked regression repros. Don't commit
  build artifacts. The bash `scripts/regression_repros.sh` writes exes to `/tmp` —
  a NATIVE Windows axc writes `/tmp` to a DIFFERENT place than Git Bash reads, so it
  reports spurious exit-127 "failures"; run repros via PowerShell with `bin/`-local
  output paths and absolute exe paths instead (`.exe` extension, no piping the exe
  to `| Out-Null` — use `& $abs *> $null` then read `$LASTEXITCODE`).
