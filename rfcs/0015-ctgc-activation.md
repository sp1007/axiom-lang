# RFC 0015 — CTGC / Ownership / Escape activation (BUG#69)

- **Status:** P1 + P2 Implemented — OwnershipChecker live as a mutability-only
  checker (E4002); EscapeAnalyser now ACTIVE and marking escaping locals with
  `SYM_FLAG_ESCAPES` (2026-07-16, A==B `184E35B4`); move-checking removed by design
  (no move semantics); CTGC-free (P3) still deferred (needs borrow/alias analysis)
- **Author:** self-host team
- **Tracking:** BUG#69 (discovered 2026-07-06 while investigating why RFC 0014
  drop-glue never fired)
- **Related:** RFC 0010 (aggregate value/alias semantics — the crux constraint),
  RFC 0014 (`drop(self)` — BLOCKED behind this), `ownership.ax`, `escape.ax`,
  `ctgc.ax`, air_builder (`OP_DESTROY`, `FLAG_ESCAPES_TO_HEAP` readers)
- **Blocks:** RFC 0014 (drop-glue), any real compile-time GC / move / escape story

---

## 1. Problem — three passes that have never run

`OwnershipChecker` (`ownership.ax`), `EscapeAnalyser` (`escape.ax`), and
`CtgcInjector` (`ctgc.ax`) are **complete no-ops** and always have been. Each
pass's entry traversal opens with:

```
if node_idx == 0xffffffff as u32 or node_idx == 0 as u32:
    return
```

But AST **node index 0 is the real `NODE_PROGRAM` root** (`ast.ax::new_ast_tree`,
"Append root program node at index 0"). The passes are entered with node 0
(`check()` → `check_node(0)`, `run()` → `traverse_and_inject(0, 0)`,
`run()` → `traverse_nodes(0)`), so the guard fires on the very first call and the
traversal returns immediately, every time. The value `0` is overloaded: it is
both the true root AND `NULL_IDX` (the "no child/sibling" sentinel), and these
three passes conflated them.

**Consequences:**
- The driver prints `[Debug] Running Ownership Checker...` / `... Escape
  Analyser...` / `... CTGC Injector...` — **misleading**: nothing is checked,
  marked, or injected.
- No compile-time GC has ever run for any AXIOM program. Every local
  struct/sum/option/result/generic-inst leaks (CTGC was meant to free exactly the
  owning block via `OP_DESTROY`→`ax_free`, but never did).
- No move/ownership enforcement has ever run — single-owner is spec intent only.
- `FLAG_ESCAPES_TO_HEAP` is read (air_builder.ax, ctgc.ax) but never set by a real
  pass → always reads false.

Independently confirmed: `let b = a; let c = a` (two copies of a struct) compiles
clean, though a running `check_move` would flag `E4001 use of moved value`.

## 2. Why the one-line guard fix is NOT the fix

Removing `or node_idx == 0` (safe in isolation — all recursive calls already
filter `!= 0` at the call site) makes the passes actually run. Measured
experimentally (patched, built, **reverted, not committed**):

- Fixpoint A==B still holds (compiler self-builds).
- **Regression 0/98** — every build fails in OwnershipChecker. Core runtime/stdlib
  is always bundled (`result/alloc/scheduler/runtime/os/string/io/collections`)
  and already "violates" the move rule in dozens of places.
- Scale: building the whole self-host compiler through itself yields **260 errors,
  258 of them `E4001 use of moved value`**, 2 `E4002`. A sampled E4001
  (`parser.ax`, `self.append_child(node, body)` reusing `node`) is a **completely
  normal read-only reuse** of a struct variable — not a real ownership transfer.

## 3. Root constraint — AXIOM has NO move semantics (RFC 0010)

`check_move` (ownership.ax:89) marks **any** aggregate-typed identifier as `MOVED`
after a single use as var-decl-init / assign-rhs / call-arg — i.e. it assumes
move-only aggregates. But **RFC 0010 established the opposite**: AXIOM aggregates
have alias / shared-view semantics, and the self-host compiler actively *depends*
on aliasing (the "read-after-mutate same buffer" pattern, RFC 0010 §9). There is
no move operator and no move semantics in the language today.

⟹ The move-checker is not "buggy in details" — its entire premise contradicts the
language as it exists. Activating it as-written turns almost all valid code into
false compile errors (the 258 E4001). This is why BUG#69 is a **language-model
decision**, not a guard fix (CLAUDE.md §13: ownership-model changes require an RFC).

## 4. Decision points (must be resolved before any activation)

1. **Does AXIOM adopt move semantics?**
   - **No (recommended, status quo):** then `OwnershipChecker`'s move-checking is
     wrong by design and must be removed or gutted, NOT merely retuned. Only its
     mutability check (`E4002`, assign to non-`mut`) is potentially meaningful —
     and even that must be audited (the 2 E4002 sites may be checker offset bugs
     or genuinely missing `mut`).
   - **Yes:** a full move/borrow system is a multi-month RFC of its own and
     conflicts head-on with RFC 0010's alias dependence. Out of scope here.
2. **What is CTGC allowed to free, given aliasing?** RFC 0010 §9 proved the
   compiler holds aliases into vectors it later mutates. A CTGC that frees a
   block still aliased elsewhere = real UAF (the same class as BUG#51/#52, only
   currently "safe" because free never runs). CTGC free MUST be gated by a *sound*
   escape/alias analysis, which does not exist yet.
3. **Is EscapeAnalyser's heuristic sound?** It currently (e.g.) treats every
   `NODE_CALL_EXPR` argument as escaping. Never validated against real code.

## 5. Recommended phased plan

Do **not** flip all three passes at once. Sequence by risk, gate each on
`fast_fixpoint` + full `regression_repros` on the native driver:

- **P0 (this RFC):** framing + decisions above. Add in-code `// no-op — RFC 0015`
  notes at each guard so future readers are not misled (comment-only, zero
  behavior change).
- **P1 — mutability-only OwnershipChecker — ✅ DONE (2026-07-06).** Decision
  4.1 taken = "no move semantics". `check_move` gutted to a no-op; the guard
  fixed (`node_idx == 0xffffffff` only) so the pass runs; only `E4002`
  immutable-assign is enforced. Building the whole compiler through the active
  checker surfaced exactly the 2 expected E4002 sites — both **false positives**:
  assignments to `mut` PARAMETERS (`fn substr(s, mut start, mut end): start = 0`).
  Root cause: the resolver's `NODE_PARAM_DECL` handler dropped the parser's
  `FLAG_IS_MUT`, so `mut` params never carried `SYM_FLAG_MUT`. Fixed in
  resolver.ax (mirror `NODE_VAR_DECL`); `SYM_FLAG_MUT` is read ONLY by this
  checker, so zero codegen effect. Result: 0 false positives across stdlib +
  compiler, fixpoint A==B, regression 102/102. Tests: `bin/t_mutparam.ax`
  (positive), `tests/sema/err_immutable_assign.ax` (reject). AXIOM now enforces
  `let` immutability — assigning to a non-`mut` binding is a compile error.
- **P2 — sound escape analysis:** redesign EscapeAnalyser to *conservatively*
  mark a local aggregate as non-escaping ONLY when it provably never (a) is
  returned, (b) is stored into a longer-lived aggregate/global, (c) has its
  address taken across a call that can realloc the source, or (d) is aliased.
  Validate `FLAG_ESCAPES_TO_HEAP` is set correctly on a corpus before any
  consumer trusts it.

  **P2 activation attempt 2026-07-12 — NOT landed, reverted; two blockers found
  (de-risks the next attempt):**

  1. **Flag collision (solved).** `FLAG_ESCAPES_TO_HEAP = 128` (node-flag space)
     is the SAME value as `SYM_FLAG_MOVED = 128` (sym-flag space). escape.ax's
     original `sym.flags |= FLAG_ESCAPES_TO_HEAP` therefore sets `SYM_FLAG_MOVED`,
     which `ownership.ax:79` reads and emits `E4001 use of moved value` on every
     use — re-creating the 258-error flood. And its `decl_node.flags |=
     FLAG_ESCAPES_TO_HEAP` drives air_builder's heap boxing (lower_var_decl ~3143 →
     lower_ownership_aware ~2419: alloc + store). So **both** original flag writes
     change behavior. Fix: a dedicated non-colliding `SYM_FLAG_ESCAPES = 4096`
     (free in sym-flag space; no current consumer) that P3 reads instead. This part
     works.

  2. **~~Merely RUNNING the pass breaks the fixpoint~~ — RESOLVED 2026-07-16: it
     was a CRASH, not non-determinism.** The 2026-07-12 conclusion ("latent
     allocation-order / address-dependent codegen non-determinism", "period-2
     oscillation `8AF4B46C ⇄ A70D6243`") was **wrong** — a *stale-artifact phantom*.
     When enabled, the escape pass **segfaults** (exit 139) in `run()`. The fixpoint
     script's hop-2 (`A build src -o axc_fpB`) therefore never writes `axc_fpB.exe`,
     leaving the *previous* run's binary in place; the reported "B" hash was whatever
     stale file survived, so it appeared to flip-flop between the last two builds — a
     fake 2-cycle. Confirmed by (a) the hop-2 log ending abruptly at "Running Escape
     Analyser...", (b) `axc_fpB.exe`'s mtime never advancing, (c) a direct run on a
     2-function toy segfaulting identically. **Crash root cause:** `traverse_nodes`'
     per-function `let old_cg = self.curr_cg; self.curr_cg = new_connection_graph();
     … ; self.curr_cg = old_cg` save/restore was written for value-copy semantics,
     but AXIOM aggregates **alias** (RFC 0010). `old_cg` aliases the `curr_cg` field
     instead of snapshotting it, so after the last function frees its graph and does
     `self.curr_cg = old_cg` (a self-aliasing no-op), `curr_cg` holds **dangling**
     buffer pointers; `run()`'s final `free_connection_graph(self.curr_cg)` iterates
     the dangling `adj_out`/`adj_in` arrays and `@free`s garbage → wild free → SIGSEGV.
     **Fix (escape.ax):** functions are never nested, so the save/restore is both
     pointless and broken — dropped it; each `NODE_FUNC_DECL` builds a fresh graph,
     frees it, and resets `curr_cg` to a valid empty graph, so neither the next
     function nor `run()`'s final free ever sees dangling buffers. With the crash
     fixed and the flag write disabled, the pass runs over the *entire self-host
     compiler* and **A==B holds** (`F100027D`) — there was never any codegen
     non-determinism. **P2 landed 2026-07-16** (`184E35B4`, 327/327): the escaping
     marking is now written to `SYM_FLAG_ESCAPES = 4096` (SYM-flag space, no 128
     collision), consumed by nothing yet, so it stays fixpoint-neutral. Oracle
     `t_escape` (exit 41). Remaining: P3 must add borrow/alias tracking (see below)
     before any CTGC consumer trusts the flag.
- **P3 — CTGC free, whitelisted (OPEN, high-risk):** only after P2 is sound, inject
  `OP_DESTROY` for provably-non-escaping, non-aliased local aggregates — starting with
  a narrow whitelist (e.g. a single leaf module), measuring real regression blast
  radius, expanding gradually. This is the prerequisite RFC 0014 (`drop(self)`) needs.
  **Concrete blocker identified 2026-07-16:** activating CTGC free *unconditionally*
  will UAF-crash the self-host. `ctgc.ax` frees every block-local of struct/sum/opt/
  result/generic-inst type whose `SYM_FLAG_ESCAPES` bit is clear — but the self-host
  compiler is full of **borrow** locals like `let node = tree.nodes.data[i]` that,
  under alias semantics (RFC 0010 §9), point *into* a longer-lived vector rather than
  owning fresh memory. Freeing one corrupts the AST → catastrophic UAF.

  **Ownership-origin soundness substrate landed 2026-07-16 (`1A220D0B`, A==B, 327/327).**
  `analyze_stmt` now treats a local as OWNING fresh, single-owner memory **only** when
  its initialiser is a direct aggregate construction (`NODE_STRUCT_LIT`/`NODE_ARRAY_LIT`);
  any other init (an ident = alias, `INDEX_EXPR`/`FIELD_EXPR` = borrow, or a call = may
  return a borrow, no interprocedural analysis yet) flows the local to the escape node,
  so `escapes()=true` and P3 will refuse to free it. Combined with the existing
  return/call-arg/field-store flow this also transitively covers `let y = x` aliasing
  (y is escape-marked as an ident-init, and x→y flow propagates). Verified by a
  temporary per-function dump over the whole bundled compiler: the freeable set is
  small and conservative (e.g. a 19-local function yields 1 freeable, a 22-local one
  yields 0). Still inert (no consumer) → fixpoint-neutral.

  **P3 activation prototyped + reverted 2026-07-16 (findings, not shipped).** A
  flag-gated (`-ctgc-free`, OFF by default) activation was built and tested end-to-end,
  then reverted. What it established:
  1. **The free machinery works.** With the flag on, `ctgc.ax` (guard fixed, reading
     `SYM_FLAG_ESCAPES`) injects `NODE_DESTROY_STMT` → `lower_destroy` → `OP_DESTROY` →
     `ax_free`, and test programs compile and run correctly. The flag-off path is
     byte-identical (fixpoint A==B `3E099646`), so the mechanism is self-host-safe.
  2. **Before-return destroy injection is a UAF and was removed.** `ctgc.ax` originally
     inserted destroys *before* a `return`, but the return value is often computed from
     an otherwise-freeable local (`return tmp.v + 1`) → free-then-read. The sound rule:
     free only at block **fall-through**; a local live on a return path simply leaks
     (safe), never UAFs. Freeing on return paths needs the return value materialised into
     a temp first (destroy after it is computed) — deferred AST transform.
  3. **The blocker that makes P3 free *nothing* useful yet: struct construction is a
     `CALL_EXPR`, not `NODE_STRUCT_LIT`.** `Name(field: val)` parses as a call
     (`NODE_STRUCT_LIT` is effectively unused). So the P2.1 "owns iff STRUCT_LIT/ARRAY_LIT"
     rule conservatively marks **every** struct construction as a borrow/escape → across
     the whole bundled compiler, 709/709 var-decls came out escaping, 0 freeable. Sound
     (over-approximates escaping) but inert. **P3's real prerequisite is therefore:
     recognise a constructor call — a `CALL_EXPR` whose callee ident resolves to a struct
     TYPE symbol — as OWNING fresh memory (freeable), while a call to an ordinary function
     stays conservatively escaping (may return a borrow; no interprocedural analysis).**
     This must be done carefully: a single mis-classified borrow-returning call = UAF.
  4. **Even with ctor-detection, the coarse flow neutralises it — a second refinement is
     needed (prototyped 2026-07-16, reverted).** Adding ctor-call ownership (finding 3) was
     built and dump-verified, but freed almost nothing extra: the flow analysis marks a
     local as escaping on *any* use that reaches the escape node, INCLUDING a **scalar**
     field read that feeds a returned value. E.g. `let owned = Box(v: 10); return owned.v`
     — `owned.v` is an i64 *copy*, so `owned` itself does not escape and is safe to free,
     yet `analyze_expr` flows `owned` → escape node and marks it escaping (freeable=0).
     So P3 also needs: when a `FIELD_EXPR`/`INDEX_EXPR` reads a **scalar** (non-aggregate)
     field/element of a local, do NOT flow the *base* local to the destination (only the
     scalar value leaves; the aggregate stays). Only an aggregate-typed field read, or a
     use of the whole local, should propagate escape. This is a sound, targeted refinement
     but couples with ctor-detection and the free consumer — hence P3 is best built and
     validated as ONE unit, not a chain of individually-inert substrate commits.

  **Remaining for P3 (dedicated session — build + validate as one unit):** (a) ownership-
  origin: classify ctor-calls as owning (finding 3); (b) flow refinement: scalar field/
  index reads don't escape the base aggregate (finding 4); (c) the return-value-temp
  transform so return paths free too (finding 2); (d) opt-in flag or module whitelist to
  bound blast radius; (e) fix `ctgc.ax`'s `node_idx == 0` guard + retarget its flag read to
  `SYM_FLAG_ESCAPES`; (f) validate no new UAF via the native allocator's deterministic-fail
  on the full regression corpus. The `OP_DESTROY` backend path needs no work — verified ready.

## 6. Alternatives

- **Leave all three as no-ops (status quo):** honest option; costs the leak of
  every local aggregate (already the reality for the entire project's life). Pairs
  with P0 (make the no-op explicit in code). Lowest risk.
- **Delete the three passes outright:** removes misleading output and dead code,
  but discards the CTGC/drop-glue foundation the roadmap wants. Rejected — keep
  the infrastructure, document it as inert.
- **Arena/never-free allocation for AST/symbol buffers (RFC 0010 §9 option c):**
  sidesteps CTGC for the compiler's own hot buffers; orthogonal but relevant to
  decision 4.2.

## 7. Success criteria (for a future activation, not this RFC)

- Each activated pass: `fast_fixpoint` A==B AND full regression GREEN on the
  native driver, with **zero** false-positive diagnostics on the bundled
  stdlib + self-host compiler.
- CTGC free (P3): no new UAF — verified by the native allocator's
  deterministic-fail behavior (RFC 0010 §1) on the full regression corpus.
- RFC 0014 drop-glue unblocked only after P3 is stable.

## 8. Status

- **P1 Implemented** (2026-07-06) — mutability-only OwnershipChecker (E4002).
- **P2 Implemented** (2026-07-16, `184E35B4`, 327/327 regression, A==B). The escape
  pass is now **active** (previously it segfaulted the moment it was enabled — the
  2026-07-12 "period-2 non-determinism" was a stale-artifact phantom; see §5 P2.2 for
  the crash root cause and fix). It builds a per-function ConnectionGraph, computes
  transitive escape to the escape node, and marks escaping locals with the new
  non-colliding `SYM_FLAG_ESCAPES = 4096`. No consumer reads that flag yet, so the
  activation is fixpoint-neutral. Oracle `bin/t_escape.ax` (exit 41).
- **P3 Open (high-risk)** — CTGC free. Blocked on adding borrow/alias tracking to the
  escape analysis + a whitelist (see §5 P3): unconditional activation UAF-crashes the
  self-host because it would free borrow-locals aliasing the AST vectors. Needs a
  dedicated session. Unblocks RFC 0014 (drop-glue).
