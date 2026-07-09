---
name: axiom-bug-probe
description: Proactive feature-combination probing to surface silent-miscompile bugs in the AXIOM compiler when the explicit backlog is empty. Batches small programs that cross language features (generics × options × aggregates × globals × control-flow × strings), runs them, and compares exit codes to an oracle — a wrong answer or crash is a candidate bug. TRIGGER when asked to "hunt bugs", "find bugs", "probe the compiler", "stress the compiler", or when the autopilot backlog is empty and needs refilling.
---

# AXIOM Bug Probe

When the explicit backlog (OPEN bugs, RFC follow-ups) is empty, refill it by *provoking* the compiler. This method found BUG#83–88. The compiler is robust, so yield drops toward zero after ~15–18 batches — that itself is a signal (stop and re-plan).

## Method
1. **Pick a feature-cross.** Combine 2–4 features that rarely appear together, e.g.:
   - generic struct × Option/Result payload × aggregate ≥16B
   - module-level global × read-modify-write across functions
   - array/slice of Option × None construction
   - `str` concat/compare × operator overload resolution
   - short-circuit `and`/`or` inside `if`/`while` conditions with side effects
   - `for i in a..b` loop var × nested loops × break/continue
2. **Write a tiny program with a known answer** (the oracle). Make the expected exit code a distinctive number (e.g. 66, 150) so an accidental 0/garbage stands out.
3. **Build + run at -O0 AND -O1:**
   `bin/axc_native.exe build probe.ax -o probe.exe -O1; ./probe.exe; echo $?`
   Divergence between -O0 and -O1 is a strong signal (ssa_opt / regalloc / liveness).
4. **Compare to oracle.** Wrong exit, crash, hang, or compile-accept-then-miscompile ⇒ candidate.
   - Exit codes truncate to 8 bits in bash; for values >255 read PowerShell `$LASTEXITCODE`.
5. **Shrink** the candidate to the minimal repro, then hand to `axiom-investigator` for root-cause.

## Triage of a candidate
- **Silent wrong answer / crash, language clearly intends the behavior** ⇒ real bug ⇒ implement the fix.
- **Behavior the language does NOT intend, no clear user intent** ⇒ REJECT with a diagnostic (BUG#53 convention), don't silently miscompile.
- **"Bug" that matches a documented design invariant** (e.g. aggregate = reference semantics, RFC 0001) ⇒ NOT a bug. Check specs/memory before filing.

## Pitfalls (from past sessions)
- Rename functions when shrinking a repro — free-function overload resolves on arg[0] only, so a same-named stdlib fn can mask the real cause (BUG#80).
- `ax_printf_local` can swallow a `[D`/`[T` prefix; use a distinct XTRACE marker when debug-printing.
- Build the combo matrix explicitly (type × size × container) — it is the key to isolating aggregate/repr bugs (BUG#74/77).

## Output
A ranked list of candidates, each with: minimal repro, observed vs oracle, -O0 vs -O1 behavior, and a first guess at the owning pipeline stage. Feed the top candidate into the autopilot loop.
