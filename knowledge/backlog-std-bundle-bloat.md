---
name: backlog-std-bundle-bloat
description: "USER BACKLOG (do later, requested 2026-07-24): the native compile path bundles the WHOLE std library into every compiled program → exe bloat. User has seen the -dfe report but no default processing. = RFC 0031 DFE default-on decision. Investigate + make the default path not ship unused stdlib."
metadata:
  node_type: memory
  type: project
---

**USER TASK (2026-07-24, "thực hiện sau" = do later):** "hình như compiler đang nhét cả thư
viện std vào chương trình được biên dịch, tôi đã thấy báo cáo nhưng chưa thấy kết quả xử lý"
— the compiler stuffs the whole std library into each compiled program; user saw the report
but no processing result.

## What this is
This is the **RFC 0031 Dead-Function-Elimination default-on** question, plus the
`concatenate_stdlib` bundling model (main_air.ax:398). The native `-self-link` path
**always concatenates a fixed stdlib set** (result / mem.alloc / scheduler / runtime / os /
string / io / collections) into every program, then compiles it all. Unused functions are
pruned only by `-dfe` (**opt-in**, `9ce3d69`) — the default build keeps them → the exe carries
the entire bundled stdlib. Measured baseline: `return 42` = **77,824 B**, `.text` 96.7%,
~75 KB fixed overhead (see [[rfc0031-dead-function-elimination]] context notes).

The "report" the user saw is almost certainly `-dfe-report` / `-dfe` printing pruned function
names — it lists what WOULD be removed but the default path does not remove it.

## Why it was left opt-in (the real blocker)
Per [[rfc0031-dead-function-elimination]]: ALL risk is in the ROOT SET — a wrong root turns a
call into a wild jump (SIGSEGV), not just a bigger file. `-dfe` was proven by self-compiling
the compiler with 138/993 of its own functions pruned to a byte-identical binary, AND by
running the full regression + ELF suites under `AXEXTRA=-dfe`. So the machinery is trustworthy;
flipping default-on is a **USER decision** (the note says "bật mặc định là quyết định của USER").
User is now effectively asking for that. Also ties to [[feedback-no-exe-bloat]].

## Task when picked up
1. Re-confirm `-dfe` is still green on HEAD (regression + ELF under `AXEXTRA=-dfe`).
2. Decide default-on vs a smarter always-on prune. Flipping `-dfe` default-on is the direct
   answer; the ELF runtime-in-program caveat ([[dfe-elf-runtime-is-in-program]]) must be
   re-verified (COFF green does NOT imply ELF green — it shipped 4/12 SIGSEGV once).
3. Measure exe size before/after on a few programs; report the delta to the user (they want to
   SEE the processing result, not just the report).
4. Backend/linker-adjacent → B==C gate + full regression + ELF check before commit.

Related: [[rfc0031-dead-function-elimination]], [[dfe-elf-runtime-is-in-program]],
[[dfe-prune-list-vs-inlining]], [[feedback-no-exe-bloat]], [[lib-consumption-cwd-and-export-traps]].
