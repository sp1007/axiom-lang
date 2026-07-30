---
name: lesson-exit-code-8bit-masking
description: "METHOD LESSON 2026-07-30 — process exit codes are masked to 8 bits, so a program returning 300 reports 44. Every oracle and probe in this project communicates via exit code, so any expected value >= 256 is unreadable and any observed value is ambiguous mod 256. Compare IN-PROGRAM and return a small sentinel."
metadata:
  node_type: memory
  type: feedback
---

# Exit codes are masked to 8 bits — never read a value >= 256 through one

Every oracle and probe in this project reports through the process exit code
(`regression_repros.sh` compares `exit|<n>`). **The OS masks that to the low 8 bits.** A program
that returns 300 exits with **44**. A program that returns 256 exits with **0**.

## How it bit, concretely (2026-07-30)
While pinning [[bug-generic-explicit-typearg-float-literal]] I probed `pick[u8](300, 1)` and read
the resulting **44** as "the u8 literal was narrowed, 300 & 255 = 44". From that I concluded that
integer literals ARE coerced on the explicit-type-argument path, and therefore that only a float
clause was missing.

**All of it was wrong.** The value is really **300** — the literal is not narrowed at all — and 44
was the exit code of `return 300`. Verified by comparing INSIDE the program instead:

    let b = pick[u8](300, 1) as i64
    if b == 300:
        return 42        // this is what fires
    if b == 44:
        return 43
    return 99

The same holds for the plain non-generic `let x: u8 = 300`, so it was never generic-specific.

Cost: one wrong hypothesis committed to memory, one retraction, and a bad assertion in an oracle
that then failed and looked like a regression in a correct fix. The oracle's own guard rows caught
it, which is the only reason it did not ship as "the fix broke integers".

## The rule
- **Any expected value must be < 256.** Prefer the project's existing convention: compute
  in-program, compare in-program, and `return 42` (or a small distinct code per failing check, as
  `t_methfloatret` and `t_genexplicitfloatarg` do — `return 1`, `return 2`, … identify WHICH row
  failed).
- **Never infer a value from an exit code** when the true value could exceed 255. An observed `44`
  means "the value is 44 **mod 256**" — it is 44, or 300, or 556.
- **A masked exit code can imitate exactly the arithmetic you are testing for.** Truncation,
  wrapping, and masking all produce `v & 0xFF`, so an exit code cannot distinguish "the compiler
  narrowed it" from "the OS masked it". That is what made this so convincing.
- Watch for the sneaky one: a test expecting 256 or 512 reads as **0**, which many harnesses treat
  as success.

Related: [[lesson-bash-grep-not-powershell-selectstring]] — same category of failure, where the
measuring instrument silently altered what was being measured.
