---
name: bug-two-array-payload-instantiations
description: "FIXED bf67c7d (A==B 44688E18, 508/508). TWO Option-with-ARRAY-payload instantiations sharing a LENGTH SIGSEGV'd because mono's get_type_name_recursive read TypeEntry.name_id — which holds the array LENGTH — as an intern id, so every array of that length mangled to the same name and the second instantiation reused the first's layout. Third instance of the name_id footgun."
metadata:
  node_type: memory
  type: project
---

# ✅ FIXED `bf67c7d` — two array-payload instantiations collided in the mangler

Found 2026-07-22 while building a verification sweep for [[bug-array-typearg-mono-payload]];
closed 2026-07-22 (same family, same footgun).

## Root cause — a THIRD `name_id`-is-not-a-name site

`get_type_name_recursive` (mono.ax:275) returned `pool.get(entry.name_id)` for any entry
with a non-zero `name_id`, **before** reaching its own structural branches further down —
which already built correct names like `arr_<len>_<elem>` and were simply unreachable.
For an ARRAY, `name_id` is the **LENGTH**, so the name came out as whatever string happened
to be interned at that id. Traced decisively:

```
MANGLEDBG _AX_std_Box__T arg0=482 kind0=3      ([i64; 2])
MANGLEDBG _AX_std_Box__T arg0=484 kind0=3      ([i32; 2])
```

Both `[i32;2]` and `[i64;2]` mangled to `Box__T` — intern id 2 was the generic param name
`"T"`. Two instantiations differing only in element type therefore shared ONE monomorphized
type, and the second silently reused the first's layout.

The fix excludes ARRAY (length) and RESULT (err type id) from the early name return, so both
fall through to the structural branches, and gives OPTION a name of its own — it has none, so
every `Option[..]` type argument previously mangled to the shared fallback `"type"`.

## Why the earlier hypotheses looked refuted

The two theories recorded against this crash were *"same length → same mangled name"* and
*"same length + differing total size"*. Both were rejected because the measured table showed
same-length pairs that worked (`[i64;2]` + `[P;2]`, `[P;2]` + `[Q;2]`).

**They were rejected for the wrong reason.** The collision really is keyed by length; it is
only OBSERVABLE when the merged layouts disagree. `[i64;2]` and `[P;2]` are both 16 bytes
with an 8-byte element, so sharing one instantiation is harmless. The crashing pairs are
exactly those where a narrow-scalar element merged with a struct element and the two halves
disagreed about aggregate-vs-scalar handling.

⚠️ **Lesson: a hypothesis that explains the failures but "predicts collisions that don't
crash" is not refuted — silent-but-harmless is a real outcome of aliasing.** Ask whether the
non-failing rows would be *observable* before discarding the theory. Tracing the mangled name
directly settled in one build what the behaviour table could not settle at all.

## Two related gaps closed with it (both silent miscompiles)

1. **Generic ctor element width.** A bare int-literal array defaults to `[i32; N]`, so
   `let b: Box[[i64;2]] = Box(v: [5,6])` inferred the type argument from the VALUE and built
   `Box[[i32;2]]` — 4-byte stores read back at the 8-byte stride, so the sum came out 6.
   `try_instantiate_struct_ctor` now takes the `expected` type and binds its params from that
   instantiation's field types, which coerces the literal at the same time.
2. **`Box[[i32; 3]](v: ..)` was a parse error** while the same spelling in TYPE position
   worked. The `[` led now parses a bracketed argument as a TYPE when it contains a `;` at
   bracket depth 1 — an array literal never can, so `f[[1,2][0]]` still parses as an
   expression.

## Gate + oracles

A==B `44688E182D4C0084`, **508/508**. `t_twoarraypayload`(104) covers four same-length
array payloads (i32 / 8-byte struct / i16 / 16-byte struct); `t_arrayctorgeneric`(80) covers
both spellings, nested array type args, a two-param generic, non-merging of same-length
different-element instantiations, and the array-literal-as-index guard. Both SIGSEGV on the
pre-fix compiler and are identical across O0–O3.

## Also closed as stale on the same pass

The *"narrow-element array payload `Option[[i32;3]]` still returns 8"* residual recorded
against [[bug-array-typearg-mono-payload]] was **already fixed by `82d0565`** and covered by
`t_arrpayloadwidth`(166); it measured 48 (correct) before any change in this session. The
note was stale, not an open bug.

Related: [[bug-array-typearg-mono-payload]], [[probe-boxed-payload-2026-07-22]].
