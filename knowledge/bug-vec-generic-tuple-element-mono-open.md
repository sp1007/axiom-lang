---
name: bug-vec-generic-tuple-element-mono-open
description: "✅ FIXED `aa419a2` (2026-07-16, A==B `13F28B29`, 321/321). `Vec[(i64,T)]` segfault. ROOT: a __tup synth-struct has no name-resolvable identity, but `is_generic` did NOT recurse into struct fields → `is_generic(__tup{i64,T})`=false → finish_generic_instantiation (gates on is_generic(args)) baked a bogus concrete Vec[generic __tup] and bound T→__tup{i64,T} (param bound to a tuple containing itself); receiver stayed generic → push never bound → backend degraded to getfld+call → SIGSEGV. FIX: is_generic recurses into __tup fields (scoped to __tup) → generic tuple reports generic → instantiation DEFERS until element concrete. Oracles t_vecgentup(1)/t_vecenum(63)/t_vecconctup(1). UNBLOCKS Vec HOF enumerate/zip/partition."
metadata:
  node_type: memory
  type: project
  originSessionId: 8e6e1303-fa98-42a9-9be6-cb389f8aac2b
---

# ✅ FIXED `aa419a2` — `Vec[(i64, T)]` (tuple-with-generic-element as Vec element) segfault

## ✅ RESOLUTION 2026-07-16 `aa419a2` (A==B `13F28B29`, 321/321)
The REAL root turned out to be simpler than the mono-substitution chain traced below (that
chain was a downstream SYMPTOM). A tuple synth-struct `__tup{...}` has NO name-resolvable
identity — its element types ARE its signature — but `is_generic` (typecheck.ax:110) did not
recurse into struct fields, so `is_generic(__tup{i64, T})` returned **false** (treated as
concrete). Consequence: `finish_generic_instantiation` (typecheck.ax:2342, which gates on
`is_generic(args)`) baked a BOGUS concrete `Vec[generic __tup]` instance and bound the element
param `T → __tup{i64,T}` (a param bound to a tuple containing itself — confirmed by an [STSH]
subst trace showing `subst[0] name=T -> type=386` where 386=`__tup{i64,T}`). That bogus concrete
instance left the receiver generic, so `out.push` was never bound (flag 2048 unset) and the
backend degraded to an indirect `getfld+call` → SIGSEGV.
**FIX (typecheck.ax `is_generic`, ~line 139-160):** recurse into `__tup` struct fields — if any
field is generic, the tuple is generic. Scoped to `__tup` (starts_with "__tup") so ordinary
named structs are unaffected → **A==B holds** (the compiler's own generic-tuple usage, if any,
reproduces). With the fix, `Vec[generic __tup]` correctly DEFERS (registers a generic-inst) until
the element is concrete; then `Vec[(i64,i64)]` resolves and `push` binds to a direct static call.
- **A FAILED first attempt** (reverted): re-canonicalizing the `__tup`'s fields via `subst` at
  the mono stash (mono.ax:161) — the active subst there maps `T → 386` (the generic tuple
  itself), so it produced `{i64, 386}` (worse). The subst there is NOT concrete; the real fix
  had to be upstream at the is_generic gate that PREVENTS the bogus binding.
- Oracles: **t_vecgentup(1)** (probe_gt_b), **t_vecenum(63)** (generic enumerate end-to-end,
  probe_enum2), **t_vecconctup(1)** guard (probe_gt_c). Also verified probe_gt_a=106.
- **UNBLOCKS Vec HOF enumerate/zip/partition/unzip** (they build `Vec[(idx,T)]`/`(A,B)`).
  (`zip` still needs the `t_lambdazip` name-collision handled separately.)

## (historical, superseded) 🔴 OPEN — segfaults through mono

Found 2026-07-15 while probing feasibility of a stdlib `Vec.enumerate` (which builds
`Vec[(i64, T)]`). DISTINCT from the just-closed tuple×generic-payload cluster
([[bug-tuple-generic-payload-unwrap-open]], `3d2aab0`/`1aff7ca`) and from `ccecc6a`
(tuple-as-WHOLE-generic-param). This is a tuple NESTED as the element type argument of a
`Vec[...]` generic-inst, where the tuple itself contains an unsubstituted generic param.

## Crisp localization (3 probes, all -O0, daily driver `bin/axc_native.exe` @ `3d2aab0`)
- **A — tuple w/ generic element, NO Vec:** `fn wrap[T](x:T)->(i64,T): return (99,x)` → **106 ✓**.
  ⇒ mono of a tuple type `(i64,T)` by itself is FINE (`ccecc6a` covers it).
- **C — CONCRETE tuple Vec inside a GENERIC body:** `fn mk2[T](x:T)->i64: mut out:Vec[(i64,i64)]=Vec.new(); out.push((99,7)); return out.len` → **1 ✓**.
  ⇒ generic-body context + `Vec[(i64,i64)]` (concrete tuple element) is FINE.
- **B — the bug:** `fn mk[T](x:T)->Vec[(i64,T)]: mut out:Vec[(i64,T)]=Vec.new(); out.push((99,x)); return out` then `mk(7).len` → **0xC0000005 ACCESS VIOLATION (segfault)**.
  Reading only `.len` (no tuple field access) still crashes ⇒ the Vec aggregate returned by
  `mk` is CORRUPT, not merely a bad field offset.
- Also: the direct end-to-end `enumerate_probe[T](src:Vec[T])->Vec[(i64,T)]` (push `(i,src.data[i])`
  in a loop, read `.0/.1`) segfaults identically.

## Precise scope
The crash is SPECIFICALLY: a `Vec` (or presumably any user generic type) whose type
ARGUMENT is a tuple type that contains a generic param `T`, when that outer fn is
monomorphized. The concrete-tuple Vec works (C); the generic-in-tuple-element Vec crashes (B).

## ✅ ROOT CAUSE FULLY PINNED 2026-07-15 (instrumentation build, then reverted — tree clean @ 803a806/97A86F33)
Did the instrumentation cycle (3 trace builds via bin/axc_instr.exe, all reverted). **Exact
mechanism + fix location now known.** The fix is a careful mono change (a focused implementation
task, NOT yet done — deferred as a clean next task, not a tail-of-session rush).

**THE CHAIN (evidence-backed):**
1. Clone `mk[T=i64]`'s local `out: Vec[(i64,T)]` — its Vec type-ARGUMENT node is a
   **NODE_TYPE_EXPR (kind 46), NOT a NODE_TUPLE_TYPE** — carrying the TEMPLATE's **generic
   `__tup{i64,T}` id (52) in `extra_idx`**. (GTRACE at the NODE_GENERIC_TYPE arg loop:
   `base=Vec argkind=46 arg_type=52 is_gen=1`.)
2. `infer_node` resolves that NODE_TYPE_EXPR via the substituted-`__tup` recovery at
   **typecheck.ax:4019-4025** (`if result_type==UNKNOWN and is_substituted: result_type =
   node.extra_idx`) — returning the GENERIC `__tup` 52 DIRECTLY. **`register_tuple_type` is
   NEVER called** (ungated ENTER trace = 0 hits, even for working concrete-tuple programs) —
   so the tuple is NOT re-canonicalized to a concrete `__tup{i64,i64}`.
3. Root: **mono `substitute_type_params` (mono.ax:70) remaps type nodes BY NAME** via
   `lookup_subst(name_id)`. A `__tup` has **NO name**, so the stashed `extra_idx=52` (generic
   `__tup{i64,T}`) is **never remapped** when `T→i64`. The forward stash (mono.ax:148-162) only
   handles substituting a param TO a `__tup`; it does NOT re-map an EXISTING generic-`__tup`
   extra_idx whose field types contain gen params.
4. So `out` = `Vec[generic __tup 52]`. At `out.push(...)`, arg_types[0]=Vec[52], and the push
   generic-call sees `is_generic=1` → **has_generic_arg=1 → DEFER** (typecheck.ax:3310-3324) →
   the call is never instantiated/bound to the push instance. (VTRACE confirmed: push call in
   the clone has arg[0]/arg[1]/inferred[0] all is_gen=1, has_generic_arg=1.)
5. Unbound method (flag 2048 unset) → air_builder emits **degraded `getfld+call`**
   (air_builder.ax:1588) → executes the Vec `data` heap pointer as code → **SIGSEGV**.

**⚠️ NUANCE (redirects the fix search):** `register_tuple_type` (typecheck.ax:2162) was
**NEVER called** during the whole compile (ungated ENTER trace = 0 hits) — even for WORKING
concrete-tuple programs (pb2). So tuple types are resolved from **cached node_types / the
`extra_idx` recovery path (typecheck.ax:4022-4025)**, and register_tuple_type's canonical dedup
is bypassed at clone time. ⇒ the fix likely CANNOT live in register_tuple_type; it must be at
(a) the mono stash (mono.ax:160-162, re-canonicalize a generic `__tup` before stashing), or
(b) the recovery point (typecheck.ax:4022, if the stashed `__tup` is generic, re-resolve). Still
UNPINNED: the exact PROVENANCE of the clone's Vec-arg NODE_TYPE_EXPR carrying extra_idx=52 —
i.e. which node/subst-entry sets it (mk's `out` tuple should be a NODE_TUPLE_TYPE, yet the clone
shows a NODE_TYPE_EXPR; something collapses it to a stashed-type node). Needs 1-2 more trace
builds (trace the stash at mono.ax:161-162 printing type_id + whether its `__tup` fields are
generic; and where mk's `out` annotation node kind/extra_idx is set) BEFORE editing.

**THE FIX (mono.ax substitute_type_params):** when a node's `extra_idx` stashes a `__tup`
whose field types contain gen params (or more generally, when descending a type node that
resolves to a generic `__tup`), SUBSTITUTE the `__tup`'s field types via `subst` and
re-canonicalize to the concrete `__tup{i64,i64}`, then re-stash that concrete id in `extra_idx`.
mono has `self.typetable` access; it must replicate register_tuple_type's dedup (scan for a
`__tup` struct whose fields == the substituted field types; else register_struct a fresh one).
**CRITICAL:** the produced concrete `__tup` id MUST equal the one the tuple LITERAL `(99,x)`
produces (register_tuple_type dedup by field type-ids guarantees this IF mono uses the same
dedup) — otherwise the same divergence recurs. Verify with dump-air that the clone's push call
becomes a DIRECT static call (not getfld+call).
**Gate:** frontend-only ⇒ **A==B**. Oracles: probe_gt_b→1, probe_enum2 correct; guards
b8 Vec[T]→1, b13 Vec[Pair[i64,T]]→1, probe_gt_c→1. **Est. moderate**; the risk is the
canonicalization-id-match (get it wrong → still diverges or a NEW dup). Do it as a focused task.

## 🔬 EARLIER REFINED DIAGNOSIS 2026-07-15 (investigator, dump-air + bisection) — superseded by the ROOT CAUSE above
Much stronger than the original hypothesis below. Findings:
- **Crash is INSIDE `push`, a DEGRADED INDIRECT DISPATCH.** dump-air of mono'd `mk[T=i64]`
  shows `%7 = getfld %3; %8 = call %7` (reads a field off the Vec and `call`s it as a fn
  pointer → executes the `data` heap pointer as code → SIGSEGV). The working concrete case
  emits a DIRECT static `call` to the push instance. air_builder only emits getfld+call when
  the `out.push` FIELD_EXPR has **flag 2048 UNSET** (air_builder.ax:1588) = typecheck FAILED
  to bind the method. push itself IS instantiated fine (`iconst 16` element size — size is
  NOT the problem, aggregate-return is NOT the problem).
- **Root = tuple type-id CANONICALIZATION DIVERGENCE.** push instance expects Vec of
  `__tup` id **t443**; the clone's `out` carries a DIFFERENT `__tup{i64,i64}` id **t449**
  (+Vec-ref t450). Two ids for the same structural `{i64,i64}` ⇒ method-bind by receiver
  element type fails ⇒ air degrades. `register_tuple_type` (typecheck.ax:2162) dedups by
  element TYPE-IDs, so the divergence means one side did NOT go through it with concrete
  `[i64,i64]`.
- **Why `__tup` specifically (not named generic struct):** a `__tup` has NO resolvable symbol
  name (identity is purely structural). A named generic struct element (`Vec[Pair[i64,T]]`,
  probe b13) resolves via its symbol/mangled-name and **WORKS**. `Vec[T]` bare (b8) WORKS.
  Only `Vec[(i64,T)]`/`Vec[(T,T)]` fail. The mono `__tup` extra_idx stash (mono.ax:148-162)
  canonicalizes a type-param substituted DIRECTLY to a `__tup`, but does NOT fire for a tuple
  nested as a Vec type-argument whose OWN element `T` is the thing substituted.
- **b15 warmup PROVES first-encounter ordering:** a program that registers a CONCRETE
  `Vec[(i64,i64)].push` BEFORE calling `mk[T=i64]` → **CORRECT**. Once the canonical
  `__tup{i64,i64}`+`Vec[__tup]`+`push[__tup]` already exist, the clone reuses them and binds.
- **LEADING HYPOTHESIS (needs 1 instrumentation build to confirm):** the clone's `out` retains
  the TEMPLATE's GENERIC-FIELDED `__tup{i64,T}` (registered during template typecheck of
  `mk[T]`) instead of re-registering `__tup{i64,i64}` with the substituted concrete element →
  `is_generic(Vec[t449])` TRUE → the generic-call `has_generic_arg` DEFER path
  (typecheck.ax:3296-3324) fires → the `out.push` call is never bound to the push instance →
  air degrades. Alternative: a pure duplicate `__tup` id that register_tuple_type dedup missed
  (annotation registered via a path that bypasses dedup).
- **EXACT TRACE STILL NEEDED (1 build, revert after):** temp dump at typecheck.ax:~3300
  (print `inferred[]` ids + `is_generic()` for the `out.push` call in the clone) AND inside
  `register_tuple_type` (print input `elem_types[]` + returned id for BOTH the annotation-tuple
  and the literal-tuple). That selects: (a) stale-generic-field defer → fix = re-register the
  clone's tuple annotation with substituted concrete elements (mono substitute descends into
  NODE_TUPLE_TYPE and forces re-canonicalization, analogous to the extra_idx stash but for a
  nested tuple type-arg); vs (b) dup-id → fix in register_tuple_type / the generic-inst
  type-arg resolution dedup. Candidate edit points: mono.ax substitute_type_params walk
  (207-213) + stash (148-162); typecheck.ax NODE_TUPLE_TYPE infer (~4741) / register_tuple_type
  (2162) / generic-method bind (3325-3412).
- **Fix surface = FRONTEND-ONLY ⇒ gate A==B** (not B==C). Oracles after fix: probe_gt_b→1,
  probe_enum2 (generic enumerate) correct+noncrash; GUARDS stay green: b8 `Vec[T]`→1,
  b13 `Vec[Pair[i64,T]]`→1, probe_gt_c (concrete tuple Vec)→1.
- **Convention:** clear user intent ⇒ FIX (not reject). Deep-mono self-host-critical; do the
  instrumentation cycle in a focused session, do NOT guess-edit.

## Why it's deep (do NOT hotfix — dedicated session)
`substitute_type_params` (mono.ax:70) is NAME-based on type nodes: for `Vec[(i64,T)]` it
recurses into the `NODE_TUPLE_TYPE` child and substitutes `T`→`i64` textually, so the
annotation should read `Vec[(i64,i64)]`. The `__tup` `extra_idx` stash (mono.ax:148-162)
only fires for a `NODE_TYPE_EXPR`/`NODE_IDENT` node being substituted TO a `__tup` — NOT for
a tuple node nested as a Vec type-argument. Suspect: the clone's `Vec.new()`/`push` instance
gets an element type whose `__tup{i64,i64}` is not (re)registered / sized in the instance
context → wrong element stride → `Vec.new`/`push`/return builds/copies a corrupt 24B Vec
aggregate (`{data,len,cap}`), so even `.len` faults. Needs a trace of: (1) how the clone's
`Vec[(i64,T)]` local resolves its element type post-subst; (2) whether `register_tuple_type`
re-runs for the tuple annotation in the clone; (3) how `Vec.new[E]`/`push[E]` mono computes
element size when E is the (substituted) tuple. Multi-build iteration; self-host-critical
(Vec+generics). Same caution class as [[bug-tuple-generic-payload-unwrap-open]] historical.

## Blocks
Vec HOF **enumerate** (`Vec[T]→Vec[(i64,T)]`), **zip** (`→Vec[(A,B)]`), **partition/unzip**
(tuple-of-Vecs return) — all construct a Vec whose element is a generic-param-bearing tuple.
So the whole tuple-returning HOF sub-cluster stays BLOCKED on this. (`zip` is ALSO blocked by
a name collision with oracle `t_lambdazip`'s user `fn zip` — BUG#80 class.) Non-tuple HOFs
(map/filter/fold/any/all/find/count/position/take_while/skip_while/reverse) already shipped.

## Repro files (scratch/, untracked)
scratch/probe_gt_a.ax (106 ✓), probe_gt_c.ax (1 ✓), probe_gt_b.ax (segfault),
probe_enum2.ax (generic enumerate end-to-end, segfault), probe_enum.ax (concrete, 63 ✓).
