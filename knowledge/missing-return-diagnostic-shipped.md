---
name: missing-return-diagnostic-shipped
description: "SHIPPED: typecheck now diagnoses a non-void fn that can fall off the end without return (was silent garbage-return). Root cause behind HashSet.contains miscompile."
metadata: 
  node_type: memory
  type: project
  originSessionId: 289d388b-4eef-403b-8501-2b93051b32f5
---

✅ **Missing-return diagnostic SHIPPED** (`ebca3d3`, frontend, A==B fixpoint `C992AB2F`, regression **131/131**).

**Problem (BUG#53 family):** AXIOM has NO implicit last-expression return. A non-void fn whose body could reach the end without `return` was silently accepted and "returned" the garbage leftover in RAX. Concretely:
- `fn f(n) -> i64: n + 1` returned `n`, not `n+1`.
- `match: Some(_): true / None: false` (block or inline) as the tail dropped the arm value and fell off the end → returned `true` for ANY input. This was the **root cause of the HashSet.contains miscompile** (fixed in stdlib first, `867d901` [[inline-match-arm-unsupported]]).

**Fix:** `stmt_terminates` (conservative structural terminator analysis) in `bootstrap/stage1/typecheck.ax`, called from the `NODE_FUNC_DECL` handler. Terminates iff: `return`; if/elif/**else** with all branches terminating; `match` with ≥1 arm and all arms terminating; **loops and trailing call-exprs are treated as terminating** (may be `while true` / noreturn `panic`) so valid code is NEVER rejected (accept false negatives, never false positives). Only concrete non-void ret types checked (unresolved generics skipped). The whole self-build passes cleanly → compiler's own code returns on all paths.

**Harness:** added a reusable `reject` cmp mode to `scripts/regression_repros.sh` (expect build failure / no exe) — use it for future BUG#53-style reject oracles. Oracles: `t_matchret`(exit 30, match w/ explicit returns — must NOT false-positive) + `t_missingret`(reject, bare tail expr).

✅ **Follow-up SHIPPED — exhaustiveness-aware terminator** (`4d6971c`, A==B `292D1C08`, **132/132**). A non-exhaustive `match` as the fn tail (uncovered variant, no catch-all, nothing after) fell off the end → garbage; missing-return previously assumed all matches exhaustive & missed it. `match_is_confidently_partial` resolves the scrutinee variant set (Option/Result/sum/generic-inst) and sets `FLAG_MATCH_PARTIAL` **only when CERTAIN** a variant is uncovered w/ no wildcard/binding — any uncertainty ⇒ assumed total (never false-positive). `stmt_terminates` treats FLAG_MATCH_PARTIAL as non-terminating. AXIOM still allows non-exhaustive match to fall through at runtime (NOT a standalone exhaustiveness error); partial-match-with-explicit-fallthrough still OK (169). Oracle `t_partialmatch`[reject].

**Known residual (low prio):** inline match arms (`Pattern: expr`) still PARSE via `parse_match_arm`'s own path (line ~633), unlike `if`/`while` inline suites which `parse_block` rejects (BUG#53). Bare-value inline arms are now caught by missing-return when used in value position; call-body inline arms (`Some(v): assert(...)` in `collections_test.ax`) still work. Could reject inline arms for consistency, but would need rewriting those tests — deferred.
