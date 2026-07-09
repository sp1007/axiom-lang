---
name: axiom-fixpoint-gate
description: The deterministic self-host verification gate for the AXIOM compiler. Use before committing ANY compiler change — runs the fast fixpoint (A==B for frontend, hand-built B==C for backend/linker), full regression, and oracle spot-checks, and tells you exactly which criterion applies. TRIGGER whenever verifying a compiler change, "run the gate", "check fixpoint", "is it self-hosting", "verify before commit", or after editing bootstrap/stage1/*.ax.
---

# AXIOM Fixpoint Gate

The self-hosting compiler must reproduce itself byte-for-byte. This skill is the exact, deterministic gate. Get the criterion right — using A==B for a backend change is a false gate.

## 0. Rebuild the daily driver if source changed
`bin/axc_native.exe` is the seed for the fast fixpoint AND the `AXC=` regression driver. If you changed `bootstrap/stage1/*.ax` and want `AXC=` regression to reflect it, rebuild it first — otherwise you test stale code. (The fast fixpoint script seeds from the current `bin/axc_native.exe` and builds the NEW concat, so it already exercises new source into A→B.)

## 1. Regenerate the concatenated source
`scripts/regen_concat.ps1` rebuilds `bootstrap/stage1/tmp_concatenated_air.ax` from the per-file modules. Never hand-edit the concat. (The fast_fixpoint script calls this for you.)

## 2. Fast fixpoint (~9s)
Run: `powershell scripts/fast_fixpoint.ps1`
Chain: `axc_native (seed) --build new source--> A`, then `A --build new source--> B`. It prints SHA-256 of A and B.

**Pick the criterion by change class:**
- **Frontend-only** (lexer / parser / typecheck / air_builder / ssa_opt when it doesn't change emitted bytes): expect **A == B**. A mismatch is RED.
- **Backend / codegen / regalloc / emitter / coff / elf / linker / ABI** (changes the bytes the compiler emits for itself): **A != B is the CORRECT transition** (A was built by the old backend, B by the new one). The real gate is **B == C**, built by hand:
  ```
  bin/axc_fpA.exe  build bootstrap/stage1/tmp_concatenated_air.ax -o bin/axc_fpB.exe -self-link -O1
  bin/axc_fpB.exe  build bootstrap/stage1/tmp_concatenated_air.ax -o bin/axc_fpC.exe -self-link -O1
  ```
  Then compare SHA-256 of `axc_fpB.exe` and `axc_fpC.exe`. **B == C ⇒ GREEN.**
  (Get-FileHash on Windows / sha256sum under bash.)

## 3. Full regression
`AXC=bin/axc_native.exe bash scripts/regression_repros.sh` — every test must pass. Know the baseline count from the handoff memory (e.g. 114 tests + short-circuit oracles). A drop is RED.

## 4. Oracle spot-checks
For each new/affected `bin/t_*.ax`:
`bin/axc_native.exe build bin/t_x.ax -o /tmp/x.exe -O1; ./x.exe; echo $?` — and also at `-O0`.
- Exit codes truncate to 8 bits in bash (921→153). For full 32-bit values use PowerShell `$LASTEXITCODE`.
- If `-O0` passes but `-O1` fails (or vice-versa), the fault is in `ssa_opt` / block ordering / liveness — dump AIR at both levels and diff.

## 5. Promote & record
On GREEN: if the daily driver should advance, copy the verified B (== C) over `bin/axc_native.exe`. Record the new fixpoint hash + baseline count in the handoff memory.

## Commit timing (fixpoint-async rule)
- **Frontend-only:** commit after regression GREEN; the fixpoint may settle asynchronously.
- **Backend/ABI/linker:** **B==C is MANDATORY before commit.**

## Red flags (RED even if some tests pass)
- Hash varies between runs of the same inputs ⇒ non-deterministic build. Unacceptable.
- A test was edited to pass ⇒ invalid; revert and fix the code.
- New unresolved-symbol / segfault only under `-self-link` ⇒ backend miscompiles its own emitter; bisect via dump-air.
