---
name: bug53-single-line-if-miscompile
description: "BUG#53 FIXED (8d4b89b): parse errors now HALT before codegen; inline `if COND: <stmt>` rejected cleanly (no miscompile); `;` statement separator is now first-class. Multi-line if still preferred."
metadata:
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

**STATUS: FIXED 2026-07-04 (commit 8d4b89b).** Root cause was TWO bugs: (1) the driver
(main_air) never checked `parser.diags_count`, so a malformed parse flowed into codegen;
(2) `parse_block` requires TK_INDENT, which an inline suite lacks → `expect(TK_INDENT)`
soft-fails, desyncs the parser, and codegen emits a call to an empty symbol
("Unresolved external symbol ''"). **Fix:** main_air aborts (exit 1) when
`parser.diags_count > 0` BEFORE codegen (§8 — no more accept-then-miscompile). Bonus:
found `;` same-line statement separator (`putchar(x); fflush(null)`, used by t_cp2/
t_movrr/t_modrm/t_alias) raised a spurious "expected newline" that the old swallow-diags
path hid — the new halt would (correctly) abort those valid programs, so `expect_newline`
now consumes `;` as a newline-equivalent, making it first-class. Inline `if COND: <stmt>`
is now REJECTED cleanly (not miscompiled); to SUPPORT it as real syntax would need an RFC
(single-statement suite in parse_block) — deferred. Original report below.

**BUG#53 (was OPEN, compiler correctness):** the inline form

    if COND: return X          # or any single-line `if COND: <stmt>`

is **silently miscompiled** by the self-hosted stage1 compiler. The parser ACCEPTS
it (no diagnostic), but codegen emits corrupt code. Symptom when such code is added
to the frontend and self-built: the resulting compiler is broken — mis-parses valid
source (e.g. "expected expression nud at offset <beyond EOF>") then `OOM / RUNTIME
PANIC: Out of memory`. Discovered 2026-07-04 building RFC 0011 P4 inc2: a new
`iface_type_token` used 17 `if x: return "lit"` lines → n1 self-build produced a
corrupt binary; n2 crashed. **This form appears NOWHERE else in bootstrap/stage1**
(grep `^\s+if [^:]+: \S` = only that file), which is exactly why it was never
exercised before.

**Why:** parser likely produces a malformed inline-block AST (or attaches the stmt
wrong) that codegen then mistranslates; not yet root-caused in parser vs codegen.

**How to apply:** in ALL `.ax` sources, write ifs multi-line —
    if COND:
        return X
Never `if COND: <statement>` on one line — it is now a hard parse error (exit 1), no
longer a miscompile, but still not supported syntax. (`if COND: // comment` with NO
statement is fine.) When a self-build suddenly breaks after adding frontend code that
itself is never called on the self-build path, suspect a miscompiled *construct* (like
this) — bisect by the unusual syntax, not the logic. **General lesson that outlived this
bug:** the driver now halts on any parse diagnostic, so latent false-positive diagnostics
become build failures — if you add a construct the parser mis-diagnoses (like `;` was),
fix the parser, don't weaken the halt. To SUPPORT inline suites as real syntax would need
an RFC (single-statement block in parse_block). Related self-host-hazard class:
[[next-step-15-selfhost-status]] (TRẦN STAGE0), [[bug51-hunt-progress]].
