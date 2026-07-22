---
name: bug-iface-variant-payload-no-vtable-box
description: OPEN — a concrete struct in an interface-typed VARIANT payload (Some/Ok/user sum) is stored raw with no vtable box, so the method call jumps through a wild pointer (SIGSEGV)
metadata:
  type: project
---

**Status: OPEN.** Found by probing 2026-07-23. Repro:
`bin/known_fail_iface_variant_payload.ax` (expect 9, get SIGSEGV / exit 139).

```
interface Shape:  fn area(self: Self) -> i64
struct Sq: side: i64 ; fn area(self: Sq) -> i64: return self.side * self.side

let o: Option[Shape] = Some(Sq(side: 3))
o.unwrap().area()        // SIGSEGV
```

## Scope — the whole variant-payload family, nothing else

| position | result |
|---|---|
| `Option[Shape] = Some(Sq(..))` | **SEGV** |
| `Result[Shape, i64] = Ok(Sq(..))` | **SEGV** |
| user sum `Wrap = Box(Shape)`, `Box(Sq(..))` | **SEGV** |
| `fn take(o: Option[Shape])`, `take(Some(Sq(..)))` | **SEGV** |
| **tuple element**, `let t: (Shape, i64) = (Sq(..), 7)` | **SEGV** |
| `Vec[Shape]` + `push(Sq(..))` | ok |
| `HashMap[i64, Shape]` + `insert(1, Sq(..))` | ok |
| `mut s: Shape = Sq(..)` then `s = Rec(..)` (reassign) | ok |
| generic struct field `Box[Shape](v: Sq(..))` | ok |
| array literal `[Sq(..), Sq(..)]` typed as shapes | ok |
| `struct Holder: s: Shape` | ok |
| `let s: Shape = Sq(..)` | ok |

**The tuple case was PREDICTED and then found (2026-07-23).** Reasoning that the root cause is
an enumerated site list with an omission, tuple-element construction is the same kind of
aggregate-payload position as a variant payload — so it should fail the same way. It does, and
the same workaround fixes it (`let s: Shape = Sq(..)` then `(s, 7)` returns 16). That is
confirmation that the defect is the missing site, not anything specific to `Some`/`Ok`.

Probed alongside it and CLEAN: HashMap values, interface reassignment, generic struct fields,
generic fn params, array literals. So the omission is narrow — aggregate CONSTRUCTOR payloads
(variant and tuple) — rather than "containers in general".

**Workaround:** coerce through a typed local first — `let s: Shape = Sq(side: 3)` then
`Some(s)` returns 9 correctly. The box gets built at the `let`, and wrapping a value that is
already an interface is fine.

## Root cause

`air_builder.ax:1744 coerce_struct_to_interface` is the RFC 0029 T→I coercion, and its own
comment enumerates where it is called: *"call-arg, let-binding, return, struct-field init,
assignment."* **Variant-constructor payload is not in that list.** Seven call sites exist; none
is the `Some`/`Ok`/`Err`/user-variant path at `air_builder.ax:~2165`, which lowers the payload
with a bare `payload_reg = self.lower_expr(payload_idx)`.

This is the same failure shape as the RFC 0031 root set: **an enumerated list of sites that
missed one, where the missing entry fails as a wild dispatch rather than a wrong value.**

## Why it is NOT a one-line fix

The obvious patch — call `coerce_struct_to_interface` on `payload_reg` — has nothing to pass as
the target type. The variant ctor lowers **bottom-up** and never sees the declared type: it
computes `box_ty` FROM the payload's own type (`register_option(payload_t)`), so
`Some(Sq(..))` builds an `Option[Sq]` box and the declaration's `Option[Shape]` never reaches
it. At the `let` site the existing coercion call does run, but bails because the target kind is
`TYPE_KIND_OPTION`, not `TYPE_KIND_INTERFACE`.

Two candidate fix points, both real work:

1. **Thread the expected type into ctor lowering** so the payload coerces at construction.
   Correct and general; touches the lowering contract.
2. **Rebuild at the binding site** when target `Option[I]` meets source `Option[Concrete]`.
   More local, but it means unboxing and reboxing an already-built variant.

## The existing width reject does not and cannot catch it

`typecheck.ax:4971` rejects Option/Result payload **width** mismatches, but (a) it is scoped to
call arguments only, and (b) it compares payload SIZES. An interface value is an 8-byte box
pointer and `Sq` is 8 bytes, so nothing looks wrong — and a 16-byte payload struct segfaults
too, since the block never runs for the declaration site. Sizes are the wrong invariant here:
the two payloads agree on size and disagree completely on meaning.

An interim reject (turning the crash into a diagnostic, as was done for the Option-tuple case)
would need a NEW check at the declaration site. Worth doing, but `typecheck.ax` records several
over-rejection attempts that broke the self-build, so it needs its own gated session rather
than a drive-by.

## Fix ATTEMPTED and reverted 2026-07-23

Tried candidate 1 in its cheapest form: at the variant ctor, read the type typecheck assigned
to the CALL NODE (`node_types[idx]`), and if it is `Option`/`Result` of an interface, coerce the
payload before `box_ty` is chosen.

**It OOMs the self-build.** The seed compiler dies with `AXIOM RUNTIME PANIC: Out of memory`
(`OOM size requested: 24`) building the compiler source, which uses `Option`/`Result` heavily —
the new path fires far more often than the failing shape and blows up type registration.
Reverted; the tree is back at `1FD4C07F…`.

## The real blocker: the fix SITE cannot hold any more code

The OOM turned out to have nothing to do with the fix logic. Bisected by shrinking the added
code to nothing:

| what was added inside `if variant_ctor != 0:` | result |
|---|---|
| full coercion (read declared type, coerce payload) | OOM |
| same, without the `let we = entries.data[..]` aggregate copy | OOM |
| only `let want_ty = self.mb.node_types[idx]`, unused | OOM |
| only `if variant_ctor == 99: ax_puts_local(..)` — inert, touches nothing | **OOM** |
| that same inert statement in `coerce_struct_to_interface` instead | **builds fine** |
| the inert `ax_puts_local` at the TOP of `lower_call_expr` instead | OOM |
| `mut probe_unused := idx` — a statement with NO call — at the top | **builds fine** |

**The rule, and it fits every row above: `lower_call_expr` cannot take ONE MORE CALL SITE.**
A statement with no call is fine anywhere in the function; a statement containing a call OOMs
the seed compiler regardless of where in the function it sits, while the same call in a
neighbouring function is harmless. `OOM size requested: 24` on the 1.9 MB compiler source.

That also explains the row that first looked like a counterexample: `let want_ty =
node_types[idx]` contains no visible call, but array indexing emits an implicit
`ax_bounds_check` call — so it is a call site too.

A probe counting how often the interface condition matched reported **0 hits** — the coercion
never ran even once, and the build still OOMed. The failure is entirely about the call site
existing, not about anything it does.

This is a capacity/miscompile bug at exactly the site the fix has to go, and it is the actual
blocker. It resembles the documented "stage0 ceiling" (`read_file_content` in main_air.ax
carries a comment about a stage0 inference bug that caps how much code stage1 can hold), but
this one is stage1-side and function-local.

**Consequence for anyone picking this up.** The fix needs a call — `coerce_struct_to_interface`
— at a site that cannot accept one. Three ways out, in increasing order of effort:

1. ~~**Put the call in a different function.**~~ **TRIED AND REFUTED 2026-07-23.** Implemented
   the tuple half in `lower_tuple_lit` — a different, smaller function — calling the existing
   `coerce_field_to_interface`. It OOMs too.

   The earlier control that suggested this route was misleading: an inert `if x == 0xFFFFFFFF`
   in `coerce_struct_to_interface` built fine, but an unreachable branch in a cold function
   adds essentially nothing. A real call on a hot path (every tuple literal) is a different
   test, and it fails. **The ceiling is NOT function-local — it is the global ~30 MB margin.**
2. ~~**Spend an existing call site.**~~ Follows the same fate: there is no per-function budget
   to trade, only total headroom.
3. **Fix the memory ceiling.** Was the blocker for the compile — now DONE
   ([[bug-small-alloc-256mb-segment-slab-cap]], cap 256 MB -> 1 GB). Compilation of a fix at
   this site no longer OOMs.

## Progress 2026-07-23: tuple half FIXED, variant half still open for a DIFFERENT reason

With the cap raised, the coercion fix compiles. The two halves then diverged:

- **Tuple element — FIXED and shipped** (`lower_tuple_lit`, oracle `t_ifacetuple`). typecheck
  assigns the tuple node its declared anonymous-struct type, whose fields carry the real
  element types, so `coerce_field_to_interface` per element works.
- **Variant payload — compiles but STILL SEGFAULTS at runtime.** The same shape of fix at the
  `Some`/`Ok` ctor reads `node_types[idx]` for the ctor call node and it is **not**
  `Option[Shape]` — the coercion condition never matches (an earlier `[IFPROBE]` counter showed
  0 hits), so nothing is boxed and the raw struct is still stored. Reverted; driver back at
  `3CD8B786`.

**The real distinction, now isolated:** typecheck gives a TUPLE literal its declared aggregate
type but gives a VARIANT constructor call the payload-DERIVED type, not the annotation`s
`Option[Shape]`. So the declared interface type genuinely is not reachable from the ctor node
bottom-up, which is what makes this candidate 2 (rebuild at the binding site, or thread the
`let` target type into ctor lowering) rather than the one-liner the tuple case turned out to
be. The memory ceiling was masking this the whole time — the fix could never even be RUN to
observe that it reads the wrong type.

**And it is an `-O1` limit, not an intrinsic one.** The exact source that OOMs at `-O1` builds
**fine at `-O0`**. So there is no hard ceiling on call sites. The coercion design can therefore
be validated at `-O0` before anyone fights the limit.

**CORRECTION — it is the LINKER, not an optimizer pass.** I first concluded "some `-O1` pass
degrades superlinearly" and wrote that down. Running the repro with `--time` falsifies it:

```
[codegen] reloc table done: 15281 relocs, 2447 syms, writing object...
[time] codegen: 1868 ms
[Debug] Stage 4: Finished native code generation.
[Debug] Stage 5: Starting self-linking...
AXIOM RUNTIME PANIC: Out of memory
```

Codegen COMPLETES and writes its object. The OOM is in **stage 5, self-linking**. `-O1` matters
only because it changes the object handed to the linker (relocation and symbol counts), not
because an optimizer pass misbehaves.

So route 3 starts in `linker.ax`, not in the optimizer pipeline — a much narrower search, and a
different file from the one my previous note sent people to. The lesson is the cheap one: one
`--time` run replaced a plausible theory I had already committed.

Do NOT start by designing the coercion. Two attempts have already failed for a reason that has
nothing to do with their design.

**Gate trap hit while doing this, worth its own warning:** `fast_fixpoint.ps1` reported
`SUCCESS: A == B` **twice** for a source that did not compile at all. A failed hop leaves the
previous run's `fpA/fpB` in place, the `Test-Path` guards pass, and the two hashes agree
because they are the same stale file. The tell is `A == seed`. The script now deletes the hop
outputs before building and prints a note when `A == seed`, so this fails honestly instead.
Never trust a fixpoint result without confirming the hop output was actually produced.

Related: [[rfc0029-vtable-progress]], [[dfe-elf-runtime-is-in-program]] (same "enumerated site
list missed one" shape).
