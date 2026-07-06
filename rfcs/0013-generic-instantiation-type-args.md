# RFC 0013 — Track generic instantiation type-args on `StructInfo`

- **Status:** Accepted, SHIPPED (`4fb74f4`, 2026-07-06). Fixpoint verified (A==B),
  regression suite 95/95 (`t_hashi64` + new `t_nestedgen` rows green). Section 3's
  alternative (b) (self-describing mangled format) remains open for RFC 0011 P5
  (generics through the `.lib`/`.dll` interface), where the mangled string is still
  the only source of truth and the same ambiguity is latent but out of this RFC's
  shipped scope.
- **Author:** self-host team
- **Tracking:** [[bug64-vec-big-aggregate-element]] BUG#65 (memory), follow-on to BUG#64
  (shipped `b08867b`) and BUG#66 (shipped `73f396d`).
- **Related:** `bootstrap/stage1/typetable.ax` (`TypeTable`, `StructInfo`,
  `register_struct`, `get_generic_args`, `get_generic_args_from_mangled`),
  `bootstrap/stage1/mono.ax` (`instantiate_function`).

---

## 1. Motivation

Nested generic instantiation (`Vec[Vec[T]]`, `MV[MV[i32]]`, and by extension any
generic struct whose own type parameter is itself a generic instance) mis-resolves
its type arguments, causing method calls on the "inner" value (`v0.get(0)` where
`v0 = outer.get(0)`) to fall back to indirect field+call dispatch and segfault.

Traced root cause (not guessed — confirmed via AIR dump on the minimal repro
`scratch/mvmv.ax`, see BUG#65 in memory `bug64-vec-big-aggregate-element.md`):

1. When a generic **struct** is monomorphized, `mono.instantiate_function` calls
   `typetable.register_struct(mangled_id, ...)` (mono.ax:421). The resulting
   `TypeEntry.kind` is `TYPE_KIND_STRUCT` (kind=1) — **not** `TYPE_KIND_GENERIC_INST`
   (kind=8) — even though the struct is a concrete instantiation of a generic
   template.
2. `get_generic_args(type_id)` branches on `entry.kind`. For a real
   `TYPE_KIND_GENERIC_INST` entry it reads the exact args vector from
   `TypeTable.generic_insts` (typetable.ax, `register_generic_inst`). For a
   monomorphized-struct `TYPE_KIND_STRUCT` entry, there is no such vector, so it
   falls back to `get_generic_args_from_mangled`, which **re-derives** the args by
   splitting the type's mangled name string on every `"__"`.
3. Mangling is recursive: `MV[MV[i32]]`'s mangled name is built by concatenating the
   *already-mangled* name of its own type argument (`MV__i32`) behind the same
   `"__"` delimiter used to separate sibling arguments — yielding
   `"_AX_std_MV__MV__i32"`. The string-splitting fallback cannot tell "one nested
   argument `MV[i32]`" apart from "two sibling arguments `MV`, `i32`" — the format
   has no length-prefix or bracket to mark nesting boundaries. It picks the wrong
   split, `T` gets bound to a garbage/default type, and the chained method call
   resolves against the wrong template.

This is a genuine mangling-format ambiguity (an architectural gap, not a one-line
bug), which is why BUG#65 was deliberately left open rather than patched under
schedule pressure — per CLAUDE.md §13 (RFC required for changes with this kind of
representational ambiguity) and §20 (don't silently guess semantics).

## 2. Design

**Do not change the mangled-name string format.** It is load-bearing for
`.lib`/`.dll` interface compatibility (RFC 0011) and symbol matching; changing its
delimiter scheme is the higher-risk option (b) noted in memory and is explicitly
rejected here.

Instead: stop relying on **string round-tripping** to recover type-args for any
struct that the compiler itself monomorphized. The concrete args vector already
exists in memory at the moment of instantiation — thread it through and store it
directly, once, on the `StructInfo` record:

```
pub struct StructInfo:
    fields: StructFieldVec
    generic_args: U32Vec        // NEW. Empty (data=null,len=0) for a plain,
                                 // non-generic struct.
```

`StructInfoVec.push`'s growth path uses
`@compiler_intrinsic("size_of")[StructInfo]()` (typetable.ax:106) rather than a
hardcoded byte count, so adding a field here does **not** hit the same landmine as
`TypeTable` itself (whose `new_type_table` uses a hardcoded `@alloc(120)` sized to
its exact current field list — that struct must NOT gain fields without also
updating the magic number and auditing every other assumption of its layout; this
RFC does not touch `TypeTable`'s own shape at all).

`mono.ax::instantiate_function` (mono.ax:395-422) already receives the exact
concrete `args: U32Vec` for this instantiation as a parameter. At the existing
`register_struct` call site (mono.ax:421), also record `args` onto the new
`StructInfo.generic_args` field for that struct entry (`register_struct` gains an
optional args param, or a follow-up `set_struct_generic_args(struct_type_id, args)`
setter — either is fine; prefer whichever keeps `register_struct`'s existing
non-generic callers unchanged).

`typetable.ax::get_generic_args(type_id)` gains a new first branch: if
`entry.kind == TYPE_KIND_STRUCT` and `structs.data[entry.extra].generic_args.len >
0`, return that vector directly. Only fall through to
`get_generic_args_from_mangled` for structs with no recorded args — i.e. types that
arrived via a `.lib`/`.dll` interface import (RFC 0011), where the mangled string
really is the only source of truth and the ambiguity, while still latent, is out of
this RFC's scope (import interfaces don't yet carry nested-generic structs at all —
tracked separately under RFC 0011 P5).

## 3. Alternatives considered

- **(a) chosen above.**
- **(b) Self-describing mangled format** (length-prefix each argument, or bracket
  syntax `Vec<Vec<i32>>` instead of flat `__`): fixes the ambiguity at the root and
  would also fix the `.lib` interface import case, but touches every consumer of
  mangled names (symbol matching, DLL export/import, static lib interface parsing)
  — much larger blast radius for a bug whose current known trigger (generic struct
  nested in another generic struct) is comparatively rare. Deferred; worth
  revisiting once RFC 0011 P5 (generics through the `.lib` interface) is tackled,
  since that will need *some* answer to nested-generic naming anyway.
- **Do nothing / leave as a documented landmine:** rejected — the bug is real,
  reproducible, and will resurface as soon as any stdlib collection nests (e.g. a
  future `HashMap[K, Vec[V]]`).

## 4. Migration / compatibility impact

Purely additive: new struct field (safe, dynamic-sized vec growth), new
early-return branch in `get_generic_args` gated on the field being non-empty.
No change to mangled-name format, ABI, or existing `.lib`/`.dll` interfaces.
Existing non-generic structs get `generic_args = (null, 0, 0)` and take the
existing string-fallback path unchanged (in practice they never reach it, since
`get_generic_args` is only called for `TYPE_KIND_GENERIC_INST`/`STRUCT` entries
being treated as generic).

## 5. Testing plan

- Un-skip / add `scratch/mvmv.ax`'s repro as a committed regression test (e.g.
  `bin/t_nestedgen.ax` following the `bin/t_*.ax` + `scripts/regression_repros.sh`
  row convention used for BUG#66's `t_hashi64`).
- Full `scripts/fast_fixpoint.ps1` gate (A==B) before commit — this changes
  `typetable.ax`/`mono.ax`, both pervasive to the compiler's own self-build.
- Full `scripts/regression_repros.sh` (94/94 baseline) must stay green.
- Explicitly re-test `Vec[BigStruct]` (BUG#64) and `HashMap[K,i64]` (BUG#66) repros
  to confirm no regression in adjacent generic/collection paths.

## 6. Drawbacks

- Adds a small amount of memory (one `U32Vec`, 24 bytes on this ABI) per struct
  entry, including non-generic ones (empty vec, so effectively 3 zeroed
  words) — negligible next to `TypeEntry`/`StructField` vectors already carried.
  Considered acceptable; not worth the complexity of a separate sparse side-table.
