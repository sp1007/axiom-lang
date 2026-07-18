---
name: ctgc-p3-scoping-2026-07-18
description: "CTGC P3 (general compile-time free) scoping (2026-07-18 investigation): general-free activation is TOO RISKY for a single session (un-enumerated missed borrow-locals UAF-segfault a -ctgc-free self-build); the inert marking substrate already shipped (P2.1). The one safe direction-aligned increment = a dump-only freeable-set REPORT flag. Kickoff note for the future dedicated activation session."
metadata:
  type: project
---

# CTGC P3 / RFC 0014 drop-glue — scoping (2026-07-18, read-only investigation)

Direction #3 from [[autopilot-direction-2026-07-18]] = "RFC 0015 P3 (CTGC-free) + RFC 0014
drop-glue = ATTEMPT with tight gate (dedicated session)". This note scopes it.

## Current state (file:line)
- **Escape analyser ACTIVE, marks an INERT flag.** `escape.ax:107` sets
  `symbols.data[sym_idx].flags |= SYM_FLAG_ESCAPES` (4096, defined `resolver.ax:49`).
  Ownership-origin soundness `escape.ax:156-166`: a local `is_owning` ONLY if init is
  `NODE_STRUCT_LIT`/`NODE_ARRAY_LIT` or a `NODE_CALL_EXPR` whose callee is `SYM_STRUCT` (ctor);
  everything else (ident/INDEX/FIELD/plain call) → escape-edge → never freed. Scalar-flow
  refinement `escape.ax:256-279`.
- **The flag has ONE consumer, gated OFF by default.** Read only at `ctgc.ax:63`. The whole
  CTGC pass is a no-op unless `-ctgc-free`: `ctgc.ax:34` `if not self.free_enabled: return`
  (flag wired `main_air.ax:836-837,1026`). So escape marking is fixpoint-neutral (A==B) today.
- **Free is DOUBLY narrowed:** (1) injected only at block FALL-THROUGH, never before `return`
  (`ctgc.ax:98-105`, UAF-avoidance); (2) backend `air_builder.ax::lower_destroy` (L4190) emits
  `OP_CALL drop`+`OP_DESTROY` ONLY if `resolve_drop_method(sym.type_id) != 0` (L4225-4243).
  The compiler declares no `drop`, so a `-ctgc-free` self-build frees NOTHING → byte-identical.
  True current scope = "opt-in `-ctgc-free`, only for user types declaring `drop(self)`" (RFC 0014
  `c149872`). **General free of every non-escaping owned local = DEFERRED** (rfcs/0015 §5, §8).

## The blocker (why activation is not tick-sized)
The self-host compiler is saturated with **borrow-locals aliasing long-lived vectors** that must
NEVER be freed: `let node = self.tree.nodes.data[node_idx]` (INDEX-init borrow into the AST vector),
`let sym = symtable.symbols.data[i]`, `let entry = typetable.entries.data[id]`. Under AXIOM
aggregate=reference semantics, freeing one frees a slot INSIDE the live vector → catastrophic UAF.
`escape.ax:156-166` defends against this, **but rfcs/0015-ctgc-activation.md:289-292 records the
escape analysis still MISSES some of the compiler's pervasive aliasing** — a `-ctgc-free` self-build
UAF-segfaults. That residual, **un-enumerated** missed-borrow set is the blocker. Closing it = either
(a) audit/close more aliasing holes (open-ended), or (b) a module whitelist (hard: native path is a
single concatenated tree with no AST-level module boundary). Both multi-session; the mandated gate
(`-ctgc-free` self-build reproduces fixpoint, revert-on-red, no partial commit) goes RED.

## Decision: DEFER activation; ship the inert REPORT precursor instead
The "add inert borrow-edge marking first" idea from the direction doc **already shipped as P2.1** —
no inert marking work remains. The ONE safe, gate-able, direction-aligned increment is a
**dump-only freeable-set REPORT** (`-ctgc-free-report`): traverse like the free pass but PRINT each
freeable local instead of injecting any `OP_DESTROY`. Zero codegen change → default build
byte-identical (A==B trivially). It converts the blocker's "un-enumerated missed borrows" into an
**inspectable artifact** (run against the concatenated self-host source, see everything the current
escape analysis would free) — exactly RFC §5 steps (d)/(f) "measure real regression blast radius".
That is the correct kickoff for the future dedicated activation session.

Gate scripts for the eventual activation: `scripts/ctgc_free_check.sh` (`CTGC_FREE_OK`), oracles
`bin/t_ctgcfree.ax` (42), `bin/t_drop.ax`, `bin/t_escape.ax`. See [[bug69-ctgc-ownership-escape-noop]],
[[rfc0014-drop-glue-blocked]], [[backlog-open-items]].
