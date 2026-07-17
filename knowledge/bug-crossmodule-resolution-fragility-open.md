---
name: bug-crossmodule-resolution-fragility-open
description: "✅ FIXED `a9cfd91` (2026-07-16, A==B `7D682EB4`, 323/323). Cross-module same-name mis-link (t_modcollide 101→2 when Vec.partition bundled). Root (NOT emission — that was a wrong mid-diagnosis): air_builder's match-by-name UFCS method-dispatch loop ran even for an ALREADY-resolved (flag 2048) FIELD_EXPR callee, re-picking the FIRST same-named SYM_FUNC by symbol order (lib_a.close 1619) for a module-qualified `lib_b.close` call (correct payload 1622). Fix=skip UFCS re-dispatch when callee flag 2048 set. Also SHIPPED Vec.partition. (zip still blocked — different free-fn-shadow flavor.)"
metadata:
  node_type: memory
  type: project
  originSessionId: 8e6e1303-fa98-42a9-9be6-cb389f8aac2b
---

# ✅ FIXED `a9cfd91` — cross-module same-name mis-link (air_builder UFCS re-dispatch)

## ✅ RESOLUTION 2026-07-16 `a9cfd91` (A==B `7D682EB4`, 323/323, t_modcollide=101 WITH partition)
The whole multi-day trail below converged. Trace-verified across the FULL pipeline (all
instrumentation reverted): registration ✓ (distinct symbols lib_a=1619/lib_b=1622), call-site
resolution ✓ (resolve_field returns each), mangling ✓, and BODY emission ✓ (bodies emit as
`ax_close__m1619`=x+1 and `ax_close__m1622`=x+100). The [MTRACE]/emission "bypass" mid-diagnosis
was WRONG — the hidden stdlib close is `File.close` (`ax_File_close`), which is why MTRACE gated
on fn_name=="close" saw 0. The REAL bug: at AIR call lowering, `main`'s reloc for BOTH close
calls targeted sym 1619 ([RELOC] trace). air_builder's **match-by-name UFCS method-dispatch loop**
(air_builder.ax ~1660) is a REPAIR for calls the resolver could not bind — receiver-type-agnostic,
picks the FIRST same-named `SYM_FUNC` by symbol order. It ran EVEN for the already-resolved
(flag 2048) `lib_b.close` FIELD_EXPR (correct payload 1622), setting is_method_call=true +
method_sym_idx=1619 (lib_a, lower index), so air_builder used 1619 for the lib_b call too
(AIRDBG: `callee_payload=1622 is_method=1 method_sym=1619`). partition only shifted indices enough
to make the receiver node_type match and the loop fire (enumerate didn't).
- **FIX (air_builder.ax, the `if not resolved:` UFCS block ~1632):** guard it with
  `not callee_resolved` where `callee_resolved = (callee_node.flags & 2048) != 0`. A resolved
  FIELD_EXPR is authoritative — a real method call is already handled by the flag-2048 branch
  above (sets is_method_call there), and a module-qualified free call carries the exact payload.
  A FIRST attempt guarding on "receiver is a SYM_MODULE" did NOT trigger (receiver-symbol detection
  subtlety); the flag-2048 guard is cleaner and correct.
- **A==B held** (7D682EB4; the compiler's own module-qualified calls now lower via the resolved
  payload) ⇒ implies B==C by determinism. Regression 323/323, t_modcollide=101 WITH partition.
- **Also SHIPPED Vec.partition** `partition[T](self, pred) -> (Vec[T], Vec[T])` (oracle
  t_vecpartition=62). **zip still blocked** — that's the free-fn-SHADOW flavor (user bare `fn zip`
  vs stdlib Vec.zip, BUG#80 [[bug-freefn-stdlib-collision-noarg]]), NOT this module-qualified path.

## (superseded trail) 🟡 OPEN — cross-module same-name resolution breaks when a stdlib symbol is added

Found 2026-07-16 adding `Vec.partition[T](self, pred) -> (Vec[T], Vec[T])` to std/collections.ax.

## Symptom (reproducible, deterministic)
- Oracle **t_modcollide** (`lib_a.close(0) + lib_b.close(0)`, lib_a.close=x+1, lib_b.close=x+100,
  want **101**) → returns **2** (both resolve to lib_a.close = 0+1 twice) when `partition` is in
  the bundled stdlib. Remove `partition` → back to 101.
- **`Vec.enumerate` (also added, also generic, also tuple-involved) did NOT break it.** So it is
  NOT a pure symbol-table-index shift; it is specific to `partition`'s shape. partition differs by:
  returns a **tuple of aggregates** `(Vec[T], Vec[T])` AND has a **fn-typed param** `fn(T)->bool`.

## ✅ TRACE-VERIFIED 2026-07-16 (instrumentation, reverted) — bug is CODEGEN EMISSION, not resolution
Ran 3 trace builds (partition re-added, all reverted). RULED OUT my first hypothesis
(overload-dedup pointer-instability — it was WRONG):
- **Registration CORRECT:** [RTRACE] at resolver.ax:542 shows THREE `close` symbols —
  282 (hidden STDLIB close, decl 14523, tree A), 1619 (lib_a, decl 1, tree B), 1622 (lib_b,
  decl 1, tree C). Cross-tree dedup correctly returns dedup=0 (distinct symbols); the
  tree-pointer eq works fine (pointers stable, NOT invalidated). So 1619≠1622, distinct.
- **Call-site resolution CORRECT:** [FTRACE] at lazy_resolver_resolve_field shows
  `lib_a.close → symbol_idx 1619`, `lib_b.close → symbol_idx 1622`. Exactly right.
- **BUT the `__m` uniquing is BYPASSED at codegen:** [MTRACE] at x86_resolve_sym_name
  (x86_regs.ax:198, the fn that emits `ax_<name>__m<sym_idx>` for MODDUP symbols) **NEVER
  FIRES for `close`** (0 hits) during the self-link compile — even though fn_name resolves
  to "close" and there's no earlier early-return for it. ⇒ the two `close` bodies/calls are
  emitted via a path that does NOT apply the MODDUP `__m` suffix → they emit bare `ax_close`
  → LINK-MERGE (both calls bind to lib_a's body, x+1) → result 2. Same failure family as the
  free-fn bare-mangle residual [[bug-freefn-stdlib-collision-noarg]] (`f7bc186`), but on the
  MODULE path and only surfacing at certain symbol-index values (partition shifts indices).
- **Why partition specifically:** partition shifts symbol INDICES; without it, close bodies
  apparently DO go through the `__m` path (or don't collide) → 101. With it, they bypass it →
  bare `ax_close` collision → 2. The next question is WHERE the self-link path emits the
  close body/call label (bypassing x86_resolve_sym_name) and why that path skips MODDUP.

## NEXT-SESSION plan (corrected + precise)
CORRECTION: `x86_resolve_sym_name` (x86_regs.ax:198) IS the self-link mangle and IS called
widely — x86_selector.ax:1493 (call targets), x86_coff.ax:354 (body labels),
x86_asm_emitter.ax:167/501. So MTRACE=0 at its MODDUP branch (line ~245) means for the two
`close` symbols EITHER `fn_name = pool.get(sym.name_id)` is NOT "close" at mangle time (the
symbol's name_id may be a pre-qualified/module name), OR an early-return fires before line 245
(main/assert/memcpy/free/alloc/malloc — none is close, so unlikely), OR the guard at
x86_regs.ax:204 rejects the kind. Precise next trace:
1. Instrument x86_regs.ax:206 (right after `let fn_name = pool.get(sym.name_id)`): for
   sym_idx ∈ {282, and the two lib closes}, print sym_idx, fn_name, sym.kind, sym.flags. This
   shows what name/flags the close symbols actually carry at mangle time — the crux.
2. If MODDUP flag (2048) is UNSET on the lib closes at codegen (but the resolver set it), find
   where flags get reset/lost between resolve and emit. If fn_name isn't "close", find where
   the pre-qualified name is stamped. Fix = ensure the MODDUP `__m` uniquing applies to these
   symbols (single source of truth).
3. Backend/codegen change ⇒ gate = **B==C** (not just A==B), + t_modcollide=101 WITH partition.

## Additional trace findings 2026-07-16 (more layers ruled out — emission path still not pinned)
- [MTRACE] x86_resolve_sym_name MODDUP branch (x86_regs.ax) never fires for fn_name=="close".
- [CALLTRACE] x86_selector.ax:1483 `OP_CALL` with `inst.src1==0` branch (`callee_sym_idx =
  inst.type_id`) never fires for the close calls either (0 hits). ⇒ the two `close` CALLS are
  emitted via a DIFFERENT path: either `inst.src1 != 0` (another call form) in the selector, OR
  the target name is computed at MIR-emission time in x86_asm_emitter.ax (167/501/596/618/634) /
  x86_coff.ax (354/485/497) rather than the selector. Next session: instrument ALL those call/
  body-label emission points (print sym_idx + emitted name) to find which one handles the close
  bodies+calls and whether it applies MODDUP. This bug is MULTI-LAYER codegen (deeper than the
  frontend fixes) — several emission paths; needs methodical per-path tracing, a fresh session.

## (superseded) Mechanism map (read-only, resolver.ax @ 9e7abe3)
- Cross-module same-name fns (both modules define `close`) are kept distinct at REGISTRATION by an
  overload-chain dedup keyed on `(decl_node, owning-tree-pointer)`: resolver.ax:535-566. `close` in
  lib_a and lib_b share the SAME decl_node index (each is the first fn of a single-fn module tree),
  so they MUST be distinguished by `self.symbol_trees.data[curr_idx] == self.current_tree` (pointer
  eq). The later one gets `SYM_FLAG_MODDUP` → backend emits unique `ax_close__m<idx>` (BUG#50 fix).
- Module-qualified CALL resolution (`lib_a.close`) goes via `lazy_resolver_resolve_field`
  (resolver.ax:1067): find ModuleInfo by name_id → walk `mod.exports` for the field name → return
  `exp.symbol_idx`. This path looks ROBUST (keyed by module→export, not symbol order).
- ∴ likely culprit = the REGISTRATION-time overload dedup (542): when `partition` perturbs module
  load order / AstTree allocation, the `symbol_trees[curr_idx] == current_tree` pointer-eq (or the
  decl_node match) mis-fires for `lib_b.close`, so it dedups onto lib_a's symbol and BOTH module
  exports end up pointing at lib_a's close symbol_idx. Alt: partition's tuple-of-aggregates return
  or fn-param triggers a mono/type registration that reallocates the trees vector (invalidating the
  cached tree pointers used by the eq check).

## Next-session plan (needs runtime trace — self-host-critical resolver, DO NOT guess-edit)
1. Re-add partition; trace at resolver.ax:542 for name=="close": print curr_idx, decl_node,
   symbol_trees[curr_idx] vs current_tree (pointers), and whether it returns curr_idx (dedup) or
   pushes a new MODDUP symbol — WITH and WITHOUT partition. Confirms whether the dedup mis-fires.
2. Trace `mod.exports` for lib_a/lib_b: does lib_b's `close` export symbol_idx == lib_a's? If a
   tree-pointer is stale (trees vector reallocated by partition's registration), the fix is to key
   the dedup on a STABLE tree id (module index / tree-id), not a raw `ptr[AstTree]` that can move.
3. If confirmed pointer-instability: replace the `ptr[AstTree]` equality with a stable per-module id.

## Blocks / relation
- Blocks **Vec.partition** (works standalone — probe 62 O0/O1 — but unsafe to bundle) and likely
  **Vec.zip** (also collides, via the free-fn-shadow flavor [[bug-freefn-stdlib-collision-noarg]]).
- Determinism concern (CLAUDE.md §3: "never break deterministic compilation") — adding an unrelated
  symbol changing resolution outcome is the red flag. Repro NOTE lives in std/collections.ax (`9e7abe3`).
