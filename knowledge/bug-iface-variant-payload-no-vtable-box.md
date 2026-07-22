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
| `Vec[Shape]` + `push(Sq(..))` | ok |
| `struct Holder: s: Shape` | ok |
| `let s: Shape = Sq(..)` | ok |

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

Related: [[rfc0029-vtable-progress]], [[dfe-elf-runtime-is-in-program]] (same "enumerated site
list missed one" shape).
