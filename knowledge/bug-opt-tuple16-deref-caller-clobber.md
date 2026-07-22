---
name: bug-opt-tuple16-deref-caller-clobber
description: "OPEN silent miscompile: reading a field of a match-bound EXACTLY-16-BYTE tuple payload (Option[(i64,i64)]) makes OP_DEREF take the str-style 16-byte inline copy and zero one adjacent 8-byte stack slot in the CALLER. Repro bin/repro_optuple_clobber.ax. One fix attempt held B==C but regressed working cases — REVERTED."
metadata:
  node_type: memory
  type: project
---

# OPEN — 16-byte tuple payload deref clobbers a caller stack slot

Found 2026-07-22 while probing the B4 width-coerce residual (which turned out to be
already closed by A1 `171ea83`). This is a **new, unrelated, silent miscompile**, and
the first OPEN bug since the "no OPEN bugs remain" note.

## Minimal repro — `bin/repro_optuple_clobber.ax` (NOT registered in the suite; it FAILS)

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
    return keep          // returns 0 — `keep` was destroyed
```

Deterministic at **-O0/-O1/-O2/-O3** on driver `57738F36`, so it is in AIR building or
the selector, not the optimizer.

## Scoping — all measured, do NOT re-derive

| Variation | Result |
|---|---|
| `Option[(i64,i64)]` payload (**16B**) | **BROKEN** |
| `Option[(i64,i64,i64)]` payload (24B) | OK |
| `Option[P]`, `struct P{x,y}` (also 16B!) | OK |
| plain `(i64,i64)` param, no Option | OK |
| `Option[i64]` scalar payload | OK |
| ONE call | OK |
| TWO calls | BROKEN (a third destroys nothing further) |
| inline `match` in main, no call at all | OK |
| callee ignores the param / only `is_some()` | OK |
| callee binds `t` but reads NO field | OK |
| callee reads `t.0` **or** `t.1` (either alone) | BROKEN |

Exactly **one 8-byte slot** is zeroed, the one adjacent to `a` in the frame layout —
which local that is follows the stack layout, not declaration order (declaring locals
before `a` moves the damage to a different one; in one arrangement nothing live is hit).
This is why the failure first looked like "the second Option variable is broken" and
then like "the second call returns None" — both are downstream of the same single
stray 8-byte write.

## Root cause (located, not yet fixed)

[x86_selector.ax:2253-2259](bootstrap/stage1/x86_selector.ax#L2253-L2259), `OP_DEREF`:

```axiom
mut deref_is_agg := false
if type_size != 16 as u32 and type_is_aggregate(sel.table, type_id):
    deref_is_agg = true
```

`deref_is_agg` forces the 8-byte "aggregates are held by reference" load. It is gated on
`type_size != 16`, using the size as a proxy for "this is a `str`" (str is the one value
genuinely 16 bytes INLINE). That proxy silently excludes any aggregate whose `entry.size`
is *exactly* 16 → a 2-element tuple payload takes the str-style 16-byte inline copy and
writes 16 bytes into an 8-byte destination home.

The 16B/24B split was **predicted from this reading and then confirmed**, so the mechanism
is right. What is NOT yet explained: a 16-byte *user struct* payload (`P{x,y}`) is fine,
so `get_register_type(sel, inst.dest)` must yield the concrete `__tup` for tuples but
something non-concrete for user structs — plausibly a consequence of `a5c410f` (air_builder
preferring the concrete bound-var payload type). **Explain that asymmetry before fixing.**

## Fix attempt 2026-07-22 — REVERTED (B==C held, behaviour regressed)

Replaced the size proxy with an explicit `is_str_val` discriminator (set at `type_id == 12`
and at the `ptr[str]`/`ref[str]` recovery), gating `deref_is_agg` on `not is_str_val`.

- **B==C fixpoint HELD** — `0F36BEBE74275F54` (A≠B is expected for a backend change; the
  criterion is B==C).
- But the new compiler returned **0** for cases that were CORRECT before (24B tuple, 16B
  user struct, the survivor-bitmap probe). Strictly worse → reverted per revert-on-red.

**Lesson: the `type_size != 16` guard is load-bearing for paths not yet mapped.** B==C
proved the backend reproduces itself; it said nothing about correctness — the same lesson
already banked for the RPO-inliner ("B==C necessary NOT sufficient, full user regression
is the real gate"). Next attempt must enumerate every producer of a 16-byte `type_size`
reaching this site BEFORE narrowing the guard, and must run the full regression on the
newly built compiler, not just the fixpoint.

Related: [[bug-unannotated-some-aggregate-match]] (a5c410f, the concrete-payload-type
preference this interacts with), [[backlog-open-items]], [[fast-fixpoint-workflow]].
