---
name: lesson-bash-grep-not-powershell-selectstring
description: "Tooling lesson (2026-07-24e): reading AXIOM native-compiler TRACE output (ax_printf_local debug prints) through PowerShell `... | Select-String TAG` SILENTLY DROPS matches — it reported 0 hits for traces bash `grep` showed firing 14x. Corrupted a multi-layer bug diagnosis with false 'not reached' conclusions. ALWAYS read trace output with `bash ... | grep`."
metadata:
  node_type: memory
  type: feedback
---

When instrumenting the self-hosted compiler with `ax_printf_local("TAG ...")` debug traces and
reading them back, **use bash `grep`, never PowerShell `Select-String`.**

**Why:** During the [[bug-option-as-call-arg-not-rejected]] diagnosis, running
`& bin/axc_fpA.exe build x.ax ... 2>&1 | Select-String "TAG"` returned **zero matches** for a trace
that was actually firing — a bash `... 2>&1 | grep -c TAG` on the exact same build showed **14 hits**.
PowerShell's `Select-String` over a native exe's piped stdout drops/mis-handles the output (encoding
or stream buffering of the raw byte stream), producing **false negatives**. This is invisible and
insidious: it looks like "the code path was never reached", which sent a bug diagnosis down 5+ false
layers (concluding a call "doesn't reach" loops it actually does reach).

**How to apply:**
- Read compiler trace output with **`bash -c '... 2>&1 | grep "TAG"'`** (or the Bash tool), or
  redirect to a file and `grep` it. Cross-check counts with `grep -c`.
- Treat any PowerShell `Select-String` "no matches" on native-exe output as UNTRUSTED — re-verify
  with bash grep before concluding "not reached / didn't fire".
- Fixpoint hashes / `Get-FileHash` in PowerShell are fine (they read files, not piped native stdout);
  this only bites piped stdout filtering of the compiler's own prints.

**Why:** a whole diagnostic session's conclusions can be silently inverted by this, wasting many
build cycles. See the LAYER-8 correction in [[bug-option-as-call-arg-not-rejected]].
