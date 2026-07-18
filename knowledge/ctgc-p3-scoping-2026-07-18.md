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

## ⭐ MEASUREMENT RESULT (2026-07-18, via `-ctgc-free-report` `d18bc4f`) — blocker now ENUMERATED + root-caused
Ran the shipped report against the whole self-host source:
`bin/axc_native.exe build bootstrap/stage1/tmp_concatenated_air.ax -o … -self-link -O1 -ctgc-free-report`.
**The freeable set is only SIX locals across the ENTIRE compiler** (not "pervasive"):
`v_exp`, `exp` (both `main_air.ax` ~1634/1666), `init_inst` (`ssa_opt.ax:1108`), `subst`
(`mono.ax:454`), `dst_alloc` (`x86_regalloc.ax:1202`), `sup` (`std/scheduler.ax:495`, concurrency).
All are ctor-literal locals (ModuleExport / AirInst / TypeSubstVec / RegAllocation / AxSupervisor).

**Root cause of the `-ctgc-free` self-build UAF is now PRECISE, not vague:** at least
`v_exp`/`exp` are immediately `(&mod.exports).push(v_exp)` — a **container store that ESCAPES**,
yet the escape analyser marks them freeable. Under activation, block-end free would destroy a
`ModuleExport` still aliased inside `mod.exports`, UAF'd later at link/emit. `init_inst` is the same
class (AirInst ctor then inserted into the function's inst list). So the blocker = **the escape
analyser does not treat "a ctor local passed by value into a container `.push()`/store method" as
escaping** (escape.ax:156-166 handles struct/array/index/field origins + assignments, but NOT
by-value method-arg-into-container). That is a **bounded escape-analysis improvement** (mark a local
escaping when it flows as a by-value arg to a `push`/insert/store method whose receiver outlives it),
MUCH more tractable than "pervasive aliasing". `subst`/`dst_alloc` are likely function-local scratch
(safe to free) — the container-store trio is the danger.

**Next dedicated activation session, concrete plan:** (1) extend the escape analyser to add an
escape-edge when a local is a by-value argument to a container mutator (`push`/`insert`/`set`), re-run
the report → the freeable set should shrink to only the truly-local scratch; (2) only then flip
`-ctgc-free` on and require the `ctgc_free_check.sh` self-build fixpoint + full -O2 regression, revert
on red. The report is the regression-free way to re-measure the set after each escape-analysis change.

## ⭐⭐ STEP (1) SHIPPED — container-store escape fix, freeable set 6 → 1 (A==B `48E17C7B`, 434/434, CTGC_FREE_OK)
Root of the missed-escape was NOT a subtle aliasing case — it was a plain **hole in escape.ax's
statement dispatch**: a bare call statement (`(&mod.exports).push(v_exp)`) parses as an unwrapped
`NODE_CALL_EXPR` at statement position (no NODE_EXPR_STMT wrapper — parser.ax:944 returns the expr
directly), and `analyze_stmt` had NO case for it → fell to the `else` branch which recursed into the
call's children **as statements** (`analyze_stmt`), so the argument idents were visited by a function
with no NODE_IDENT case and never flowed to the escape node. Fix = one `elif kind == NODE_CALL_EXPR:`
in `analyze_stmt` routing through `self.analyze_expr(stmt_idx, self.escape_node_idx)` (escape.ax, after
NODE_RETURN_STMT) — analyze_expr's existing CALL_EXPR case already flows callee + every arg to escape.
**Verified via the report: freeable set 6 → 1.** Dropped: `v_exp`,`exp` (push→mod.exports),
`init_inst` (AirInst→insert), `subst`, `sup`. Remaining: `dst_alloc` (x86_regalloc.ax:1202) — genuine
function-local `RegAllocation` scratch, passed to no call and stored in no container. Inert on the
default build (SYM_FLAG_ESCAPES only consumed by the gated-off free pass) → A==B `48E17C7B`, 434/434,
and the 9 `-ctgc-free` oracles still free their true ctor locals (CTGC_FREE_OK). **This closes the
"un-enumerated pervasive aliasing" framing entirely: the real blocker was 5 container-store escapes
from ONE dispatch hole, now fixed.** Remaining before activation: audit whether `dst_alloc` (and any
future freeable) is truly single-owner, then the activation gate.

## DIAGNOSTIC (2026-07-18, no source change) — `-ctgc-free` self-build now REPRODUCES fixpoint
Built the compiler with itself under `-ctgc-free` twice (B then C): both hops exit 0, no
abort/panic, and **B==C == the normal fixpoint hash `48E17C7B`** (byte-identical). The identity is
because the compiler's ONLY freeable local is `dst_alloc` (a `RegAllocation`), which declares no
`drop(self)`, and the backend narrowing (`air_builder.ax:4225`, `resolve_drop_method != 0`) emits
`OP_DESTROY` only for drop-typed locals -- nothing is actually freed -- identical binary. So
`-ctgc-free` is currently a **safe no-op on the self-host build** (matches RFC 0014 `c149872`). The
RFC's "`-ctgc-free` self-build UAF-segfaults" referred to GENERAL free (removing the drop-narrowing).
That true-activation step is now **bounded to a single compiler local**: removing the
`air_builder.ax:4225` narrowing would make the compiler free exactly `dst_alloc` on self-build (plus
each regression program's own freeable set -- re-measure with `-ctgc-free-report` before flipping).
**Next dedicated session** = (a) audit `dst_alloc` single-ownership + run `-ctgc-free-report` over
every regression program to enumerate their freeable sets, (b) remove the drop-narrowing, (c) gate =
`-ctgc-free` self-build B==C + FULL regression built by the `-ctgc-free` compiler + ctgc_free_check,
revert-on-red. High enough risk (changes free behavior for ALL user programs) that it is its own
deliberate increment, not a tail-of-session change -- but now de-risked from "pervasive" to "one
compiler local + measurable per-program sets".

## ⭐⭐⭐ 2nd escape gap FIXED — freeable set 1 → 0 (A==B `11EB77BD`, 434/434, CTGC_FREE_OK)
Auditing the last freeable local (`dst_alloc`, x86_regalloc.ax:1202) found it is NOT single-owner:
at x86_regalloc.ax:1208 it is REASSIGNED `dst_alloc = alloc` where `alloc = allocs[v]` (an INDEX
borrow into the `allocs` vector). Under AXIOM aggregate=reference semantics that makes `dst_alloc`
ALIAS `allocs[v]`, so freeing it would free the borrowed slot -> UAF. The escape analyser missed it
because it keyed "owning" off the INIT (a `RegAllocation(...)` ctor) and ignored the later
reassignment to a borrow. **Second gap, same family as the first.** Fix: in `analyze_stmt`'s
NODE_ASSIGN_STMT ident-LHS branch, after flowing RHS->LHS, if the RHS is NON-owning
(`expr_is_owning` false: index/field/ident/plain-call) add an escape edge LHS->escape_node — an owned
local reassigned to a borrow now escapes. Also refactored the VAR_DECL owning-origin check into the
shared `expr_is_owning(init_idx)` helper (used by both VAR_DECL and ASSIGN). **Verified via report:
freeable set 1 -> 0** — the escape analyser is now provably conservative-sound on the entire self-host
build (no local it deems freeable aliases longer-lived memory). Inert (A==B `11EB77BD`, 434/434) and
the 9 `-ctgc-free` oracles STILL free their legit single-owner ctor locals (CTGC_FREE_OK: t_drop 42x,
t_ctgcfree 42x) — the reassign rule is precise, not over-marking.

**Net position after this session:** the two escape holes that caused the historic `-ctgc-free`
self-build UAF (container-store escape + reassign-to-borrow) are BOTH closed; the compiler's freeable
set is 0 (nothing to wrongly free), so general-free activation on the self-host build is now provably
safe (a no-op). Real user-program CTGC free still works (oracles free clean locals). The remaining
"activation" lever (remove the drop-narrowing air_builder.ax:4225 so non-drop-typed clean locals also
free) is now LOW-risk but delivers no self-host benefit (compiler freeable=0); worth it only to make
`-ctgc-free` free non-drop user aggregates — a future opt-in, gated the same way (re-measure each
program's freeable set with the report, full regression under the activated compiler, revert-on-red).

## VALIDATION (2026-07-18, dump-only stress batch, no source change) — escape analysis robustly sound
Ran `-ctgc-free-report` over 5 aggregate/container/escape stress programs. The escape analyser marked
as freeable ONLY genuine single-owner local scratch, correctly excluding every stored/passed/aliased
local (all exit codes correct): (p1) ctor `stored` assigned into `arr[0]` -> freeable {} (non-ident
LHS escape); (p2) ctor `x` assigned into `o.inner` -> {} ; (p3) ctor `item` passed to `consume(item)`
-> {} (call-arg escape, the container-store fix generalizing to any by-value arg); (p4) ctor `a`
aliased by `let b = a` -> {} (a reaches escape transitively through non-owning b); (p5) ctor `s` read
only via a scalar field, never stored/passed/returned -> freeable {s} (the one correct positive).
No residual escape gaps in these patterns — corroborates that the container-store + reassign-to-borrow
fixes leave the analysis conservative-sound across the common aggregate-escape shapes, not just the
self-host source. (Batch was scratch, not banked — t_ctgcescape already guards the primary path.)
