---
name: bug-tuple-generic-payload-unwrap-open
description: "✅ CLUSTER FULLY CLOSED. variant/Option/Result ctor half=1aff7ca, generic method/fn-arg half (Vec[(i64,i64)].push)=3d2aab0. Both halves = SAME element-width coercion gap ({i32,i32} 8B into 16B slot), NOT deep-mono (historical hypothesis below WRONG/superseded). Fix pattern: thread expected __tup to the arg + sticky NODE_TUPLE_EXPR coercion. (was: PARTIALLY FIXED ccecc6a: mono type-substitution of a __tup type (id[T], layout+mangle) correct + unwrap no longer segfaults. STILL OPEN: tuple LITERAL in Some(...)/Ok(...)/generic-arg not coerced to expected (i64,i64) → builds {i32,i32}, .1==0. Coercion machinery exists (typecheck.ax:4067), just thread expected to the ctor-arg."
metadata: 
  node_type: memory
  type: project
  originSessionId: 514f834a-ce6c-4fa8-8b55-78462a04139e
---

# ✅✅ CLUSTER FULLY CLOSED — tuple as generic type-argument payload

## ✅ UPDATE 2026-07-15 `3d2aab0` — generic method/fn-arg half FIXED (A==B `11B62D1E`, 312/312)
The **remaining Vec-push half is now correct** — and it was the SAME element-width
coercion gap as the variant-ctor half, NOT the deep-mono unwrap issue hypothesized all
over the historical notes below (those are SUPERSEDED). The investigator confirmed the
mono/box/unwrap path is fully correct: pre-fix, `v.push((10 as i64, 20 as i64))` and
pre-annotated `let t:(i64,i64)=(10,20); v.push(t)` BOTH returned 42; only the bare tuple
LITERAL `v.push((10,20))` built `{i32,i32}` 8B into the 16B element slot.
- **Root:** in the generic-call path the tuple arg is inferred during arg_types collection
  with `expected=UNKNOWN` (int-literals default i32); by the time `inferred[T]` is the
  canonical `__tup{i64,i64}`, the arg was already typed narrow.
- **Fix (typecheck.ax, will-instantiate `else:` branch ~L3325):** after `inferred[]` is
  final/concrete, walk decl params in lockstep with call args (skip param0=self for a
  method call); for any param typed as a bare generic `T` whose `inferred[gk]` is a `__tup`
  struct and whose arg node is `NODE_TUPLE_EXPR`, call `infer_node(arg, inferred[gk])`. That
  hits the existing element-coercion branch (widen to i64) and PINS the node; the sticky
  guard (`1aff7ca`) preserves it against the later UNKNOWN re-visit in the 2nd arg loop.
  Frontend-only ACCEPT → A==B. Prior post-pass failed only because it predated the guard.
- Oracle **t_vectup(42)**: two-push `Vec[(i64,i64)]` summing all fields, O0==O1. Same fix
  pattern likely applies to HashMap/array-of-Vec tuple-literal elements (not yet probed).

## ✅ UPDATE 2026-07-14 `1aff7ca` — variant/Option/Result ctor payload FIXED (A==B `B5934FA9`, 311/311)
The **whole variant-ctor half of this cluster is now correct.** Root was NOT deep-mono —
it was the tuple LITERAL building as default `{i32,i32}` (8B) and a SECOND, hint-less
re-inference overwriting the coerced type as the last write. Two-part frontend fix:
1. `try_instantiate_variant_call` now takes the caller's `expected` and threads its
   positional generic arg down as the payload arg's expected-type hint (so the existing
   NODE_TUPLE_EXPR element-coercion at typecheck.ax fires → tuple builds `{i64,i64}`).
2. **Sticky coercion** in the NODE_TUPLE_EXPR branch: on an UNKNOWN re-visit, if the node
   was already pinned to a concrete `__tup`, preserve it (don't re-register `{i32,i32}`).
   The double-inference (once `expected=__tup`, once `expected=UNKNOWN`) was the real
   silent killer — proven via a per-visit trace (node visited twice, 2nd with expected=0).
- **NOW CORRECT (all 42/60 at O0 AND O1):** `Some((20,22))`+unwrap; `Ok((3,4))`+unwrap;
  **direct `let o: Option[(i64,i64)] = Some(..); o.unwrap()`** (Facet B — was logged as a
  deep-mono SEGFAULT; root was the element width all along, unwrap mono is FINE once the
  tuple is 16B); **match-bound `Some(pr): pr.0+pr.1`** (was pr.1==pr.0); 3-element tuples.
- Oracle **t_tupctor(60)**.
- **STILL OPEN (separate, deeper):** `Vec[(i64,i64)].push((10,12))` → 2nd elt still 0
  (returns 30 not 42). This is the GENERIC-FN-ARG path where coercion is deliberately
  skipped for `is_generic_call` (mono clones push's body). My fix covers variant CTORS,
  not generic fn args. Prior Vec re-inference attempt was reverted (destabilizing) — leave
  for a dedicated deep-mono session. Repro /tmp/tup/vec.ax.

## (historical) 🟡 PARTIALLY FIXED — tuple as generic type-argument payload

## ✅ UPDATE 2026-07-13 `ccecc6a` — CORE mono-substitution root FIXED (A==B, 310/310)
The **generic type-parameter substitution of a `__tup` type is now correct**. Root was
in **mono.ax `substitute_type_params`**: it wrote the concrete type's NAME into the mono'd
param/return type node and the checker re-resolved by name — but tuple synth-structs have
NO symbol and every N-tuple shares the bare name `__tupN`, so resolution → UNKNOWN (scalar
`t0` param) → field `.1` aliased `.0`. Also `get_type_name_recursive` returned bare `__tupN`
so `id[(i32,i32)]`/`id[(i64,i64)]` mangled to the SAME instance (2nd call reused 1st). Two
fixes: (1) stash the exact concrete type_id in the substituted node's `extra_idx`, checker
uses it as a fallback when name resolution fails (guarded to `__tup`; every name-resolvable
substitution byte-identical); (2) encode element type names into a tuple struct's mangled
name. Oracle **t_gentuple(42)**.
- **NOW CORRECT:** `id[T](v:T)->T` with T=(i64,i64) literal/annotated/i64-var; two distinct
  tuple shapes in one program; named struct through generic unchanged (218).
- **Side effect:** `Option[(i64,i64)].unwrap()` NO LONGER SEGFAULTS (Facet B 139→wrong-value) —
  the mono'd `ptr[T].*` now sizes T=__tup correctly enough to not crash.

## 🔴 STILL OPEN — tuple LITERAL not coerced to expected element type in Some/Ok payload
**NOT a box-ABI bug — a plain element-coercion gap.** `Some((10,20))` with return type
`Option[(i64,i64)]` builds the payload tuple as `{i32,i32}` (8B, int-literal default i32)
instead of coercing to `{i64,i64}` (16B). AIR of `mk()` (`/tmp/rem.ax` dump): `setfld` uses
`t3`(=i32); box stores an 8B pointer to that 8B tuple. unwrap correctly reads `{i64,i64}` from
the pointer, so `.1` (offset 8) reads past the 8B tuple → 0. **PROOF it's coercion, not ABI:**
explicit `Some((10 as i64, 20 as i64))` (`/tmp/rem2.ax`) → **50 CORRECT.** So the deep unwrap
path is FINE post-`ccecc6a`; the only gap is the literal's element width.
- **Fix entry point:** the tuple-literal coercion-by-expected-type machinery ALREADY EXISTS at
  `typecheck.ax:4067` (NODE_TUPLE_EXPR: if `expected` is a `__tup` of same arity, infer each
  element with the matching field type). It works for annotated `let`/param/return/global.
  The gap: the payload arg of a **variant/Option ctor call** (`Some(...)`/`Ok(...)`) is inferred
  with expected=UNKNOWN, so the tuple stays `{i32,i32}`. Thread the expected payload element
  type down: when a `Some(x)`/`Ok(x)`/user-variant ctor call is inferred with a known expected
  `Option[__tup]`/sum type, infer the payload arg `x` with expected = the variant's payload
  type. Same gap for `Vec[(i64,i64)].push((10,20))` (generic fn arg, coercion deliberately
  skipped for `is_generic_call`). ⚠️ Self-host-critical Option/variant inference; a prior
  generic-arg re-inference attempt was REVERTED — thread `expected` at the ctor-arg site, do
  NOT re-infer post-mono. Frontend A==B expected.
Everything below documents the ORIGINAL investigation (pre-fix), largely superseded by the
above precise diagnosis.

---
Found by probing (2026-07-13) the freshly-shipped RFC 0022 tuple surface × generics.

## Symptom
```
let o: Option[(i64, i64)] = Some((20, 22))
let p = o.unwrap()
return p.0 + p.1      // SEGFAULT (exit 139), oracle=42
```
- Segfaults at BOTH -O0 and -O1 (⇒ not an ssa_opt/regalloc bug — a lowering/mono bug).
- **Even `p.0` alone segfaults** ⇒ the pointer returned by `unwrap` is bad, NOT an offset/size issue.
- **Even explicit `Some((20 as i64, 22 as i64))` segfaults** ⇒ NOT the int-literal `{i32,i32}` element-coercion issue (that's the separate deferred Vec-of-tuples coercion bug).
- **`Result[(i64,i64), i64]` = `Ok((20,22))` fails identically.**

## Bisection (what WORKS vs FAILS)
- ✅ `is_some()` on the tuple-payload Option works (returns 1) ⇒ **box construction is fine**.
- ✅ Identical-size **named struct** payload works: `struct P2{a:i64,b:i64}; Option[P2]=Some(P2(...)); o.unwrap()` = 42. ⇒ the general aggregate-Option-payload path is correct.
- ✅ Tuple as a plain **fn return** works: `fn mk()->(i64,i64)` + `let (x,y)=mk()` = correct (probe pt1=117).
- ❌ Only **tuple-as-Option/Result-payload + unwrap** segfaults.

## Key evidence
Box-construction AIR for tuple payload is **byte-identical** to the working named-struct case:
```
%1: t24 = alloc            // tuple aggregate (16B)
... setfld x2 ...
%6: t25 = alloc            // box
%6: t4  = store %1         // store 8B pointer-to-aggregate into box  (SAME as struct)
%8 = getfld %7 ; %9 = call %8   // unwrap
%11 = getfld %10 ; %12 = getfld %10
```
The `Some()` box-store path (air_builder.ax:1709-1721) stores the aggregate as an 8-byte
pointer for both tuple and struct — so the box holds a good pointer in both. Ruled OUT:
`type_is_pointer_repr` (x86_selector.ax:340) correctly returns **false** for the `__tup`
kind-1 struct (only kind 6/11/12 or named Option/Result are pointer-repr).

∴ Fault is in the **monomorphized `unwrap` body specialized on the `__tup` return type**
(or the tuple aggregate's alloc/size in that mono context) — dump-air elides the stdlib
mono body so it wasn't traced further. Candidate: how mono resolves/sizes an anon-tuple
(`__tup`-named synth struct) as a generic RETURN type vs a user-named struct.

## Cluster
Same root family as the deferred **Vec[(i64,i64)].push((10,20))** miscompile (see
[[backlog-open-items]]): "tuple used as a generic type argument in the mono/box path."
A prior attempt to post-fix the Vec case (re-inference pass) was **destabilizing and
reverted** (O0=32/O1=127). Treat this cluster as DEEP-MONO, needs careful trace-debug of
the monomorphizer's tuple-type handling — NOT a quick air_builder patch. Do it in a
dedicated session, not a routine autopilot tick.

## Refined diagnosis (2026-07-13, read-only trace)
- `unwrap` = **std/result.ax:38-54**. `ax_sum_layout_is_pointer()` is a **constant TRUE**
  (`mov eax,1; ret`, x86_coff.ax:631-637) ⇒ EVERY `Option[T]` is pointer-repr; the
  pointer branch (lines 40-46) is the ONLY live path: `return (raw64[0] as ptr[T]).*`
  where `raw64 = (&self) as ptr[u64]`. (The `size<=8` / large branches, 47-54, are dead.)
- ∴ scalar `Option[i64]` (g11=42 ✓), named-struct `Option[P2]` (pt9d=42 ✓), and tuple
  (FAULT) all execute the **same** `(raw64[0] as ptr[T]).*`. Since scalar & named-struct
  work, the box/self plumbing is correct; the fault is **purely the monomorphizer's
  substitution of an anon-tuple (`__tup` synth struct) for the type parameter `T`** —
  the `as ptr[T]` / `.*` (deref-and-return-by-value of a `__tup` aggregate) mono is wrong.
- Ruled out branch-specificity: **8B tuple `(i32,i32)` (g9) AND 24B `(i64,i64,i64)` (g10)
  both fault** — not a size-branch issue (both hit the same pointer branch anyway).
- Next-session entry point: compare how mono resolves `ptr[T].*` / `@compiler_intrinsic
  ("size_of")[T]` when T binds to a `__tup`-named kind-1 struct vs a user-named kind-1
  struct (mono.ax). Likely the tuple's type-arg binding isn't found/sized in the mono
  type-substitution table (registered late/locally by register_tuple_type).

## ALSO via match (2026-07-13, cleaner repro — wrong value, no crash)
```
let o: Option[(i64,i64)] = Some((20, 22))
match o:
    Some(pr):
        return pr.0 + pr.1   // returns 40 (=20+20), oracle 42
    None:
        return 0
```
`pr.1` misreads as `pr.0`'s value (40 = 20+20) → the payload extracted from the box
is only ~8 bytes / mis-sized, so the second tuple element is lost. This DETERMINISTIC
wrong-value (no segfault) is the best debugging entry point — set a known distinctive
pair like (20,22) and watch pr.1. Confirms the bug is in payload EXTRACTION sizing for a
`__tup` type across BOTH unwrap (std/result.ax `ptr[T].*`) and match (lower_match_tagged
payload deref) — i.e. a shared `__tup`-as-payload size/repr miscompute in the tagged
extraction path, not specific to unwrap. Match on non-tuple payloads (sum mixed payloads,
Option[scalar], Option[named-struct]) is CORRECT — only tuple payload misreads.

## CONSOLIDATION 2026-07-13j (Vec[(i64,i64)] IS the SAME bug, not a separate "coercion" bug)
Re-probed the `Vec[(i64,i64)].push` case with a proper harness (distinguish build-reject from
run value). Findings:
- `bare tuple` local `(10 as i64,20 as i64)` + `._f0/._f1` = **42 ✓** (tuples fine standalone).
- named-**struct** Vec element `Vec[P].push(P(a,b))` = **42 ✓**.
- `Vec[(i64,i64)]`: `v.get(0).unwrap()._f0` = **77 ✓** but `._f1` = **77 ✗** (reads _f0's value;
  want 88). `v[0]` even DIVERGES O0=52/O1=127. Sum `_f0+_f1+12` = 32 (=10+10+12).
- **Explicit `(10 as i64,20 as i64)` STILL fails** ⇒ NOT literal-coercion. The backlog listed
  "Vec[(i64,i64)].push coercion" and this unwrap-segfault as two items — **they are ONE bug**:
  `__tup`-as-generic-payload extraction loses the 2nd field (`_f1` mis-offsets to 0 / reads _f0).
  `_f0` works only because its offset is 0 by coincidence ⇒ the unwrapped element's type is NOT
  resolved to `__tup{i64,i64}` at field-access lowering, so ALL field offsets default to 0.
- Same symptom as the match case above (pr.1==pr.0). Entry point unchanged: mono substitution of
  a `__tup` synth-struct for a type param — the extracted-payload type/layout isn't the `__tup`.
- Repro files: /tmp/tg/{a,b,e,f,g}.ax (f=_f0 ok 77, g=_f1 wrong 77). Vec case exit≈32; O0/O1
  diverge on `v[0]`. DEEP-MONO, dedicated session (do NOT hotfix — prior Vec re-inference revert).

## TWO FACETS (2026-07-13j, annotation probe) — do NOT treat as one bounded fix
- **Facet A — Vec `v.get(i).unwrap()` = WRONG VALUE, annotation FIXES it.** `let t = v.get(0).unwrap(); t._f1` reads `_f0`'s value (frontend type-prop: t not inferred as `__tup`, offsets→0). **`let t: (i64,i64) = v.get(0).unwrap()` → CORRECT (88).** So a user workaround exists, and this facet is FRONTEND return-type inference (t not resolved to `__tup` through get→Option→unwrap; same class as the BUG#61 generic-return recovery block typecheck.ax:3424). `let (a,b)=v.get(0).unwrap()` (destructure) ALSO misreads (b==a).
- **Facet B — direct `o.unwrap()` on `Option[(i64,i64)]` = SEGFAULT (139), annotation does NOT fix it.** `let t: (i64,i64) = o.unwrap()` still 139. This is the deep-mono unwrap-body issue already documented above (the mono'd `ptr[T].*` for a `__tup`). Named-struct `Option[P].unwrap()` works unannotated (165).
- **⚠️ RISK for the fixer:** naively teaching get→unwrap to infer `__tup` (fix Facet A) may route the Vec case INTO Facet B's segfaulting mono unwrap path (annotation "works" possibly via a different path than inference would produce). Fix Facet B (the mono `__tup` payload deref) FIRST, then Facet A becomes safe. Repro: /tmp/tg/{h(annot=88),i(destructure=77),j(direct seg),k(annot direct seg),m(struct ok)}.ax.

## Repro files
/tmp/pt9.ax (Option), pt9a (explicit i64), pt9c (fn-return payload), pt9h (Result),
pt9f (p.0 only), pt9g (p.1 only) — all segfault. pt9b (is_some), pt9d (struct payload) pass.

## Also noticed (separate, minor)
`match o: case Some(pr): ...` where `pr` binds a tuple → **parse error** (offset in bundled
stdlib) — match-binding a tuple payload is likely a separate parser gap (RFC 0022 P5 area:
pattern binding). Not chased.
