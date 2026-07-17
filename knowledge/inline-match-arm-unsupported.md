---
name: inline-match-arm-unsupported
description: "NOT-A-BUG: inline match arms `Pattern: stmt` aren't supported; block form works. Diagnostic polished fa9e76a (one clean error/arm, no cascade). for-in-array-of-sum + match is solid. Also: stale /tmp exe masked a failed build."
metadata: 
  node_type: memory
  type: reference
  originSessionId: 73f7537d-461e-4ce6-91c3-169b6cb570f7
---

**NOT A BUG — inline match arms are unsupported syntax.**

```
match s:
    Circle(r): return r      # ❌ inline arm → "expected expression nud" cascade (37 errors)
    Point: return 0 as i32
```
```
match s:
    Circle(r):               # ✅ block form (the norm everywhere in repo)
        return r
    Point:
        return 0 as i32
```
Consistent with [[bug53-single-line-if-miscompile]] (inline `if COND: <stmt>` also rejected — colon introduces a block). Every oracle/stdlib match uses block arms.

**Validated:** `for sh in <array-of-sum>` + `match` + calling a fn on the element all work correctly (block-form probe returned the exact expected value, deterministic ×3). RFC 0018 for-loop work [[rfc0018-for-in-array-shipped]] is solid for sum-typed array elements too.

**Two lessons for probing:**
1. **Don't use inline match/if arms in probe programs** — they produce a confusing 37-error "expected expression nud" cascade at the stdlib/program junction offset (~97.6KB), which looks like a deep/nondeterministic compiler bug but is just unsupported surface syntax.
2. **Stale `/tmp/<name>.exe` masks failed builds.** `build ... -o /tmp/x.exe; [ -f /tmp/x.exe ] && x.exe` runs a PRIOR exe if the current build failed → a stale/wrong exit code. Always `rm -f` the output first, or check the build's exit status, before trusting the run.

✅ **Diagnostic polish DONE** (`fa9e76a`, frontend-only, A==B `BC276EF6`, 142/142). `parse_match_arm` (parser.ax:633) now detects a STATEMENT keyword (return/if/while/for/let/mut/break/continue/match) at the start of an inline arm body and emits **ONE** actionable diagnostic per offending arm ("match arm body must be an indented block; ... put the statement(s) on the next line, indented"), then skips to NEWLINE/DEDENT so the next arm parses cleanly — no more 10+ error cascade. The legit inline VALUE-expression arm (`Some(x): 0`) and block-form arms are untouched. Reject-oracle `bin/t_inlinearmstmt.ax`.

**Still NOT supported (by design, not done):** inline STATEMENT arms as a real feature. SUPPORTING them (`Ok(v): return v` desugaring to a block) would be an ergonomic win [[feedback-ergonomics]] but is a **syntax change → RFC required** (CLAUDE.md §13) and contradicts the BUG#53 inline-colon precedent. Deferred to a design decision, not a quick fix.
