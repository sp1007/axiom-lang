---
name: dfe-prune-list-vs-inlining
description: Reading -dfe's prune list without accounting for the inliner produces false "missing root" alarms; at -O1+ a dropped callee is usually correct
metadata:
  type: feedback
---

**Do not read `-dfe -dfe-report`'s prune list as "these functions were reachable and got
dropped."** At `-O1` and above the inliner has already deleted call edges. A small callee
inlined into its only caller has no remaining edge, is genuinely dead, and pruning it is
correct — the program still returns the right answer.

**Why:** probing `--shared` + `-dfe` I saw `ax_helper` pruned while `lib_triple`, which calls
it, survived as an export root. That looks exactly like "export roots are seeded but their
call edges are not followed" — a wild call in a shipped DLL. It was not. Re-running with a
helper too large to inline showed both functions retained, in the executable and the DLL, at
`-O0` and `-O1`. The control case gave it away: `main → mid → leaf` pruned `leaf` and still
returned the correct `14 * 3 = 42`.

**How to apply.** When a `-dfe` prune looks wrong, before reporting anything:

1. **Make the callee non-inlinable** (a loop plus a couple of branches is enough) and re-run.
   If it survives, there was never a missing root.
2. **Check the program's actual result.** A genuinely missing root is a SIGSEGV or a wrong
   answer, not merely a name in a list. The list alone is not evidence of anything.
3. **Compare `-O0` against `-O1`.** A prune that only happens at `-O1` is the inliner.

The corresponding oracle rule, learned the same way: `t_dfeexport` originally made every
exported function a LEAF, so it proved roots survive but never that the walk continues
THROUGH one. It now includes `exported_via_helper → deep_helper`, with `deep_helper`
deliberately too big to inline — otherwise the oracle would test the inliner rather than the
root set. Related: [[dfe-elf-runtime-is-in-program]].

**Verified as part of the same probe** (previously untested end to end): `--shared` with
`-dfe` keeps the export directory intact, keeps `IMAGE_FILE_DLL` set, retains every exported
function, and correctly retains their transitive callees while dropping unreferenced ones.
