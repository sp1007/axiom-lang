---
name: session-state-2026-07-24e
description: "STATE 2026-07-24e — RFC 0015 P3 RETURN-PATH FREE SHIPPED (driver 12DBE1D8, regression 530/530, ctgc 16/16). Freeable locals live on a return path are now freed on that path too, closing the last CTGC completeness gap. No AST temp needed — lower_return already materialises ret_val before its pre-OP_RETURN hook. Ordered-descent ctgc traversal makes it sound."
metadata:
  node_type: memory
  type: project
---

**ĐỌC ĐẦU TIÊN.** HEAD after this session, daily driver `bin/axc_native.exe` = **`12DBE1D8`**,
regression **530/530**, `ctgc_free_check.sh` **16/16**, fast fixpoint **A==B `12DBE1D8`**.
Continues [[session-state-2026-07-24d]] + [[session-state-2026-07-24c]]. User: "giải quyết dứt
điểm #8" = RFC 0015 P3 return-path free.

## What shipped: CTGC now frees freeable locals on RETURN paths (not only at fall-through)
Before: a freeable (non-escaping, single-owner) local live at a `return` **leaked** — CTGC only
injected `OP_DESTROY` at block fall-through, and the injected destroys sit AFTER the return in the
child list = dead code once the return jumps out. Confirmed empirically first: `early()` with two
explicit returns → `drop_count=0` under default-on free, while an identical fall-through control →
`drop_count=2` (so the local WAS freeable; the only reason it leaked was the return).

**The RFC anticipated an AST return-value-temp transform; it turned out unnecessary.** `lower_return`
(air_builder.ax) already computes `ret_val = lower_expr(first_child)` and THEN runs `flush_defers()`
right before `OP_RETURN` — the exact "materialise value, then clean up" hook RFC 0006 `defer` uses.
So return-path frees ride the same point: **no synthetic temp, no new symbol.**

Two edits (both frontend/AIR-level → inert on self-host, freeable=0 → A==B):
1. **ctgc.ax** — block traversal rewritten from the old two-pass (collect-all-up-front, then recurse)
   to a **single ordered-descent pass**: recurse into each child, then push a freeable var onto
   `active_vars` only AFTER descending past its decl. This is the soundness crux — at a
   `NODE_RETURN_STMT` the WHOLE stack is provably initialised, so freeing it never touches a
   not-yet-declared local. New `elif NODE_RETURN_STMT` branch appends `NODE_DESTROY_STMT` for the
   entire live stack (LIFO) as TRAILING children of the return node. Factored the freeable predicate
   into `is_freeable_local()` (byte-identical to the old inline check → fall-through set + order
   unchanged). report_only stays byte-identical (appends nothing).
2. **air_builder.lower_return** — after materialising `ret_val` (handles a void `return` whose
   first_child is now a DESTROY, via a `has_val` check), lowers the trailing `NODE_DESTROY_STMT`
   children (each → `lower_destroy` → `OP_DESTROY` + drop-glue) BEFORE `flush_defers`, matching the
   fall-through ordering (block destroys run before function-tail defers).

## Why it is sound (relies on ONE existing guarantee)
CTGC only ever frees locals the EscapeAnalyser left non-escaping. A `return`-reachable local's memory
IS an escape edge → it is escape-marked → never freeable → never freed (pinned by `t_ctgcfreeesc`:
`return owned` aggregate stays un-dropped). So the return value NEVER aliases a freeable local; after
`ret_val` is read into a register, freeing the non-escaping remainder is safe. `return r.field` (scalar
copy) keeps `r` freeable and frees it after the copy — correct. No double-free: a return terminates its
path, fall-through is a mutually-exclusive path, so each local is freed on exactly one exit path (verified
across loop / nested-scope cases).

## Gate (GREEN)
- Fast fixpoint **A==B `12DBE1D8`** (inert on self-host — freeable=0).
- `ctgc_free_check.sh` **16/16** — added `t_ctgcretfree` (off=0 leak / on=2 return-path drop; take()'s
  returned aggregate stays un-dropped → on=2 not 3, pins escape exclusion) + `t_ctgcretorder` (off=0 /
  on=3, ordered-descent: local declared AFTER an early return NOT freed there, no crash).
- Full regression **530/530** (added `t_ctgcretfree|exit|2` on the default-on path).
- Extra ad-hoc validation (not banked): loop-return (4 drops, no double-free), nested enclosing-scope
  return (3 drops), 100-iter alloc-churn value-correctness (14850%256=2, no UAF).

## Next (loop stays alive)
No OPEN bugs. RFC 0015 fully closed except the **`break`/`continue` path leak** (same safe-leak class,
current-iteration local not freed on a loop-exit-via-break) — untouched, documented in RFC §8. Remaining
heavy backlog unchanged: M4 compliance / M6 perf+Mach-O / p13 ARM64 / p14 allocator size-classes — all
need user direction. Otherwise: proactive probing (`axiom-bug-probe`).
