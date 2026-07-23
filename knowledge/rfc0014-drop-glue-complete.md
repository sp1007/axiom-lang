---
name: rfc0014-drop-glue-complete
description: "RFC 0014 drop-glue COMPLETE 2026-07-24 (fixpoint 70D2EEBC, 526/526, ctgc-free 14/14). P0/P2 mechanism was already shipped 2026-07-16; this session closed the two open success-criteria: P1 (copy-by-value of a `drop`-typed value REJECTED with E4003 in ownership.ax — the double-free guard, §5.2/§6-P1/§9-#3) and P3 (bignum auto-free) which is FORMALLY SCOPED OUT with a precise root cause. The 'BLOCKED on bug69' backlog note was STALE — bug69 CTGC activation shipped P1+P2+P3 long ago."
metadata:
  node_type: memory
  type: project
---

# RFC 0014 drop-glue — implementable surface COMPLETE (2026-07-24, HEAD after this session)

User: "giải quyết dứt điểm RFC 0014 drop-glue". Driver `axc_native` = **`70D2EEBC`**,
fixpoint A==B, regression **526/526** (+t_dropcopy reject, +t_dropcopyok 42),
ctgc-free 14/14, t_drop still 42×/-ctgc-free · 0/default.

## The backlog note was STALE, not blocked
`backlog-open-items.md` said "RFC 0014 drop-glue — BLOCKED on bug69 (needs escape/ctgc
real analysis)". **bug69 CTGC activation shipped P1+P2+P3 + container free-glue long ago**
([[bug69-ctgc-ownership-escape-noop]], [[ctgc-p3-scoping-2026-07-18]]). The drop-glue
MECHANISM (`resolve_drop_method` + `lower_destroy` calling `Type.drop(self)` then
`OP_DESTROY`) was already live behind opt-in `-ctgc-free` since `2026-07-16`. What was
actually open were two of the RFC's own §9 success criteria: P1 and P3.

## P1 — double-free guard SHIPPED (the genuine remaining work)
`ownership.ax`: an implicit copy-by-value of a `drop`-typed VALUE local is now REJECTED
with **error[E4003]**. AXIOM has no explicit move, so `let b = a` / `mut b := a` / `b = a`
(bare-IDENT rhs) where `a`'s non-pointer type declares `drop(self)` would alias the owned
resource into two live locals that both drop it — a double-drop (or, masked by the escape
analyser marking both as escaping, a silently-skipped drop). Like Rust's `Drop` types not
being `Copy`. Pieces:
- `scan_has_drop()` — one-time scan in `check()`, sets `self.has_drop_types`. When false
  (every current build — the compiler + bundled stdlib declare NO drop) the per-site check
  short-circuits → **zero cost, self-host byte-identical (A==B `70D2EEBC`)**.
- `type_has_drop(rec_type)` — mirrors `air_builder.ax::resolve_drop_method` (mangled-name
  match `match_mangled_method_raw_bytes`, since an inline `fn drop(self)` carries a dotted
  mangled name `resolve_method_sym` misses). Guards: rec_type must NOT be a pointer/ref (a
  `ptr[T]` copy is a pointer copy, not a value copy, and pointer locals are never CTGC-freed),
  and the drop method's unwrapped self-param must equal rec_type.
- `check_drop_copy(stmt, rhs)` — wired into `NODE_VAR_DECL` (init) + `NODE_ASSIGN_STMT` (rhs).
  Only a bare `NODE_IDENT` rhs of a `SYM_VAR`/`SYM_PARAM` fires; a ctor-call rhs (`Res(id:i)`)
  or a plain call is fine → no over-rejection.
Oracles: `bin/t_dropcopy` (reject; the OLD driver compiled it → exit 7, proving the reject
is genuinely new) + `bin/t_dropcopyok` (42) + `tests/sema/err_drop_copy.ax`. Closes
§5.2 / §6-P1 / §9-#3.

## P3 — bignum auto-free FORMALLY SCOPED OUT (precise root cause, not a punt)
`BigUint.drop` freeing `limbs` **cannot work with scope-exit CTGC**:
1. **No bignum value is ever "owned."** `escape.ax::expr_is_owning` (L130–145) marks a
   local owning ONLY if init is a direct `NODE_STRUCT_LIT`/`NODE_ARRAY_LIT` or a ctor call
   whose callee resolves to `SYM_STRUCT`. bignum builds every `BigUint` via helper FUNCTIONS
   (`bu_alloc`/`bu_new`/`bu_from_u64`/`bu_shl`…) → callee is `SYM_FUNC` → non-owning → escapes
   → `drop` never fires for ANY bignum value. (Verified by reading expr_is_owning.)
2. **The leaks are reassignment/temporary leaks, not scope-death leaks.** `bu_shl`:
   `mut r := a` then `r = bu_shl1(r)` orphans the previous buffer at the reassignment; CTGC
   only frees at scope exit, so it wouldn't reclaim those anyway.
3. **`bu_shl`'s `mut r := a` is exactly the copy P1 now (correctly) rejects** — so a
   `BigUint.drop` cannot even coexist with the current bignum source.
Closing bignum's leak needs **interprocedural return-value ownership** (a helper's returned
value = a fresh owned resource — RFC-scale, §6-P4) or explicit `bu_free`/arena (§7). bignum
is a pure, non-self-host lib (§10: not blocking) → deferred as a language-level increment.

## Lesson
The "BLOCKED on bug69" framing had been true in 2026-07-06 and was never revised after CTGC
activation shipped — **re-read the blocker's own memory before trusting a backlog label**;
here bug69 was fully closed and the real remaining work was two success-criteria, one of
which (P3) is architecturally impossible in the RFC's scope and had to be closed with a root
cause, not implemented. Related: [[bug69-ctgc-ownership-escape-noop]],
[[ctgc-p3-scoping-2026-07-18]], [[bignum-ctgc-conflict]], [[backlog-open-items]].
