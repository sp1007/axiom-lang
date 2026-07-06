# RFC 0015 — CTGC / Ownership / Escape activation (BUG#69)

- **Status:** Draft (2026-07-06) — problem framing + phased plan; NO activation yet
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
- **P1 — mutability-only OwnershipChecker (optional):** if decision 4.1 = "no
  move semantics", strip `check_move`'s move-marking entirely; keep ONLY the
  `E4002` immutable-assignment check; fix the guard; verify 0 false positives
  across the bundled stdlib + compiler before enabling. If even `E4002` produces
  false positives, defer the whole pass.
- **P2 — sound escape analysis:** redesign EscapeAnalyser to *conservatively*
  mark a local aggregate as non-escaping ONLY when it provably never (a) is
  returned, (b) is stored into a longer-lived aggregate/global, (c) has its
  address taken across a call that can realloc the source, or (d) is aliased.
  Validate `FLAG_ESCAPES_TO_HEAP` is set correctly on a corpus before any
  consumer trusts it.
- **P3 — CTGC free, whitelisted:** only after P2 is sound, inject `OP_DESTROY`
  for provably-non-escaping, non-aliased local aggregates — starting with a
  narrow whitelist (e.g. a single leaf module), measuring real regression blast
  radius, expanding gradually. This is the prerequisite RFC 0014 (`drop(self)`)
  actually needs.

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

Draft — problem framed, decisions enumerated, plan sequenced. No code activation.
The guard-fix + RFC 0014 wiring experiment is fully reverted. Next concrete step
is a decision on §4.1 (move semantics: no, recommended), then P1 scoping.
