---
name: bug-opt-tuple16-deref-caller-clobber
description: "FIXED 2026-07-22 (A==B==C 907958ED, 493/493): silent miscompile where reading a field of a match-bound EXACTLY-16-BYTE tuple payload (Option[(i64,i64)]) made OP_DEREF take the str-style 16-byte inline copy and zero one adjacent 8-byte stack slot in the CALLER. Root = OP_DEREF used dest SIZE as a proxy for is-str; the real discriminator is the SOURCE being a tagged box. Oracle t_optupclobber(110)."
metadata:
  node_type: memory
  type: project
---

# ✅ FIXED — 16-byte tuple payload deref clobbered a caller stack slot

Found 2026-07-22 while probing the B4 width-coerce residual (which turned out to be
already closed by A1 `171ea83`). A new, unrelated silent miscompile.

## Symptom

```axiom
fn use_it(o: Option[(i64, i64)]) -> i64:
    match o:
        Some(t):
            return t.0 + t.1
        None:
            return 99

fn main() -> i64:
    let a: Option[(i64, i64)] = Some((3, 40))
    let keep = 7 as i64
    let r1 = use_it(a)
    let r2 = use_it(a)
    return keep          // returned 0 — `keep` was destroyed
```

Deterministic at -O0..-O3, so never the optimizer.

## Scoping — measured, do NOT re-derive

| Variation | Result |
|---|---|
| `Option[(i64,i64)]` payload (**16B**) | **BROKEN** |
| `Option[(i64,i64,i64)]` payload (24B) | OK |
| `Option[P]`, `struct P{x,y}` (also 16B) | OK |
| plain `(i64,i64)` param, no Option | OK |
| `Option[i64]` scalar payload | OK |
| ONE call | OK · TWO calls | BROKEN (a third adds nothing) |
| inline `match`, no call at all | OK |
| callee ignores param / only `is_some()` / binds `t` without reading a field | OK |
| callee reads `t.0` **or** `t.1` (either alone) | BROKEN |

Exactly **one 8-byte slot** was zeroed — the one adjacent to `a` in the frame layout,
which follows stack layout, not declaration order. That is why the failure looked in
turn like "the second Option variable is broken" and "the second call returns None":
both were downstream of one stray 8-byte write. With a single call the write landed on
a dead temp and stayed invisible, which is why two calls are needed to see it.

## Root cause

[x86_selector.ax](bootstrap/stage1/x86_selector.ax) `OP_DEREF`. `deref_is_agg` forces the
8-byte "aggregates are held by reference" load, and was gated on `type_size != 16`, using
the destination SIZE as a proxy for "this is a `str`" (str being the one genuinely
16-byte-INLINE value). The proxy silently excluded any aggregate whose `entry.size` is
*exactly* 16, so a 2-element tuple payload took the str-style 16-byte inline copy and
wrote 16 bytes into an 8-byte destination home.

**The real discriminator is the SOURCE type, not the destination size.** Extracting a
payload out of a tagged box (sum/option/result) always reads the box's 8-byte payload
SLOT, which holds a *reference* for any aggregate payload — `lower_variant_construct`
stores exactly one pointer there, for a user struct and an RFC 0019 / tuple synth struct
alike. That is a different situation from `p.*` where `p: ptr[Struct]`, which genuinely
reads the struct's bytes inline. Fix = one added condition:

```axiom
if type_is_pointer_repr(sel.table, sel.pool, src_type) and type_is_aggregate(sel.table, type_id):
    deref_is_agg = true
```

The existing `type_size != 16` guards are left alone. A `str` payload is unaffected:
air_builder unwraps a single-str-field synth struct back to str (type id 12), which is
not an aggregate.

**A==B==C `907958ED5460DE41`, 493/493.** All three hops identical means the change is
inert on the compiler's own self-build (compiler source never derefs a 16-byte aggregate
payload out of a box) while fixing user programs — a low-risk shape.

## The asymmetry that had to be explained first (16B struct was FINE)

A 16-byte user struct payload was unaffected, which made the bug look payload-specific
rather than size-specific. Armchair reasoning could not settle it; a temporary trace at
the `OP_DEREF` site did, in one build. The broken program had exactly ONE trace line the
two working controls lacked:

```
DRFX tid=287 kind=1 esize=16 src=415 skind=6 sextra=12 agg=0 size=16   <- tuple payload, src kind 6 = SUM
DRFX tid=42  kind=1 esize=16 src=439 skind=9 sextra=42 agg=0 size=16   <- present in ALL THREE, src kind 9 = POINTER
```

Two distinct 16-byte-struct derefs reach this site: one out of a **box** (must load 8) and
one out of a genuine **`ptr[Struct]`** (must load 16). The `tid=42` lines are why the first
fix attempt regressed working code.

## Failed first attempt — the lesson that mattered

Attempt #1 replaced the size proxy with an `is_str_val` discriminator gating
`deref_is_agg` on `not is_str_val`. **B==C HELD (`0F36BEBE`) and the compiler was still
strictly worse** — 24-byte tuples and 16-byte user structs that had been correct started
returning 0, because it swept up the legitimate `ptr[Struct]` inline derefs. Reverted.

**B==C proves the backend reproduces itself; it says nothing about correctness.** Same
lesson already banked for the RPO inliner. The gate that caught it was running the repro
set plus the full regression *on the newly built compiler* — do that before trusting any
selector change, and prefer ADDING a narrow condition over loosening an existing guard
whose load-bearing cases you have not enumerated.

Second lesson: when a guard's intent is "is this X", trace what actually reaches it rather
than deducing it. The trace prefix must be UNBRACKETED (`DRFX`, not `[D…`) or
`is_verbose_debug` (print_helpers.ax) swallows it silently.

Related: [[bug-unannotated-some-aggregate-match]], [[backlog-open-items]],
[[fast-fixpoint-workflow]].
