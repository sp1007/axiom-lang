---
name: bug-user-fn-stdlib-struct-name-collision
description: "MITIGATED 2026-07-29 (548/548, A==B C432EA9E) — a top-level user FUNCTION whose name matches a bundled stdlib STRUCT silently mis-linked (`fn worker` aliased `struct worker` from std/scheduler.ax, so the user fn was never called). Now a clean REJECT with an actionable diagnostic instead of a wrong answer. Full disambiguation (per-module symbol namespacing) remains task-cross-library-name-collision."
metadata:
  node_type: memory
  type: project
---

**MITIGATED (silent miscompile → clean reject) 2026-07-29.** Found 2026-07-24 during RFC 0033
Phase C (threads). 5th hole in the free-fn/stdlib collision family
([[bug-freefn-stdlib-collision-noarg]], whose 4 fixed holes were all fn-vs-fn): this one is
**user FUNCTION vs bundled stdlib STRUCT name**.

## Symptom (pre-fix)
`fn worker(n: i64) -> i64: ...` + `worker(5)` → main's `call worker` linked to a DIFFERENT
symbol, so the user `worker` was never called. The user body (incl. any global write) was
correct but dead. Wrong value / crash depending on what the mis-linked target does.

## Root — nailed 2026-07-29
`SymbolTable.define` (resolver.ax ~L509) keeps **ONE namespace per scope** and merges only
**FUNC-with-FUNC** into an overload chain (~L542). Every OTHER cross-kind clash falls through to
`return prev.symbol_idx  // Already defined, return previous` (~L574) — it hands back the
EXISTING symbol. `pre_define_top_levels` hoists ALL top-level decls in textual order over the
**single concatenated** stdlib+user tree (stdlib first, user source appended last), so by the
time the user's `fn worker` is hoisted, `struct worker` (std/scheduler.ax:113, bundled into every
program) already owns the name. `NODE_FUNC_DECL.payload` is then set to the STRUCT's symbol
index (resolver.ax:876) and every call resolves through that.

Note the asymmetry that made it silent: `resolve_type` (L611) filters by kind and skips non-type
symbols, but the value/call path has no such filter, and one scope entry can hold only one
symbol — so there is no way to reach the shadowed function at all.

## Fix shipped (mitigation, not the full cure)
typecheck.ax `infer_node` NODE_FUNC_DECL branch: if the decl's resolved symbol kind is
`SYM_STRUCT | SYM_INTERFACE | SYM_TYPE_ALIAS | SYM_VARIANT` (i.e. the resolver aliased it), emit
an actionable diagnostic and count a type error. Exact detector — a FUNC_DECL whose resolved
symbol is not a function is *always* the aliasing bug. Frontend-only ⇒ **A==B `C432EA9E`**,
regression **548/548**.

**Why reject and not namespace-mangle:** the real cure (per-module symbol namespacing +
linker-level collision diagnostic) is ABI/linker scope and is filed as
[[task-cross-library-name-collision]] — it needs an RFC and a B==C gate. Converting a silent
wrong answer into a compile error is the safe minimal step (§20) and is exactly the "collision
diagnostic instead of silent first-wins" the user asked for. Verified inert by construction
BEFORE implementing: the intersection of bundled struct names (27) with bundled fn names (206),
and of compiler struct names (113) with compiler fn names (778), is EMPTY in every direction.

## ⭐ The reject immediately caught a live instance in our own suite
`bin/t_spawnsmoke.ax` (the RFC 0032 P3 OP_SPAWN oracle) declared `fn worker` — its spawn handler
had been aliased to `struct worker` the whole time. It **passed anyway**, because the oracle only
asserts the binary runs and returns 5 and never dispatches a message to the handler. Renamed to
`zqspawnhandler` (intent preserved, now actually spawning the user function). Lesson: an oracle
that never exercises the symbol it names can pass straight through a mis-link.

## ⭐ CRITICAL LESSON — minimization fooled by a co-varying name
First mis-diagnosed as a "param + non-void return + module-global-write miscompile" because every
FAILING minimal test happened to be named `worker` while every PASSING one was named
`setit`/`dbl`/`wf`. The name was the real independent variable. **When minimizing, hold the SYMBOL
NAME fixed (or vary it explicitly as its own axis)** — cf. [[bug80-free-call-overload-collision]].

## Oracles
- `t_fnstructcollide` (reject) — `fn worker` + bundled `struct worker`.
- `t_fnstructok` (exit 105) — over-rejection guard: distinct fn name, a real fn-vs-fn OVERLOAD
  set (1-arg + 2-arg, the adjacent path in `define`), user struct + free-fn method.
(`bin/known_fail_worker_name_collision.ax` deleted — superseded by the reject oracle.)

Related: [[bug-freefn-stdlib-collision-noarg]], [[task-cross-library-name-collision]],
[[bug80-free-call-overload-collision]], [[rfc0033-concurrency-v1]].
