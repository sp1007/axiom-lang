---
name: bug-array-typearg-mono-payload
description: "FIXED 82d0565 (A==B 29BAF82C, 500/500): monomorphizing Option/Result with a STRUCTURAL ARRAY type argument records the variant payload_type as 0, so the pattern var is untyped and OP_INDEX falls back to stride 8. Root = register_array OVERLOADS name_id to hold the LENGTH, so the monomorphizer substituted the length as if it were an intern id. Fix = mark an array type unnameable and reuse the existing __tup extra_idx stash. Oracle t_arrpayloadwidth(166)."
metadata:
  node_type: memory
  type: project
---

# ✅ FIXED `82d0565` — array type argument lost the mono variant payload type

The last survivor of the 2026-07-22 boxed-payload sweep
([[probe-boxed-payload-2026-07-22]]), now closed. The symptom and the two refuted
hypotheses below are kept because they are what the investigation had to walk through.

## Symptom

```axiom
let a: Option[[i32; 3]] = Some([3, 40, 5])
match a:
    Some(arr):
        return (arr[0] + arr[1] + arr[2]) as i64   // 8, want 48
```

`arr[0]=3, arr[1]=5, arr[2]=0` against data `[3,40,5]` — reads at **stride 8** instead of 4.

## Chain of causation (all traced, not inferred)

1. Monomorphizing `Option[[i32;3]]` records the `Some` variant's `payload_type` as **0**:
   `TCPL scrut=482 sumextra=12 found=1 pre=0 post=0` — the variant IS found, its payload
   is simply empty.
2. The typechecker therefore never types the pattern var (`bindsym=0`).
3. `lower_match_tagged`'s `pl_type` chain finds nothing (find_variant_info empty,
   binding-symbol fallback empty) and hits its final `pl_type = 4` (i64) default:
   `MPAT scrut=482 scrutkind=6 pre=0 final=4 bindsym=0`.
4. `OP_INDEX` then sees `insttype=0` over a base register typed i64 rather than an array,
   cannot recover an element type, and its stride falls back to 8:
   `IDX insttype=0 basety=4 basekind=0 elem=0 size=8`.

## What distinguishes the working cases

| type argument | payload_type recorded | result |
|---|---|---|
| `P` (named struct) | 52 | ✓ |
| `(i64, i64)` (tuple — registered as a NAMED `__tup2` struct) | 287 | ✓ |
| `[i32; 3]` (structural array type expression) | **0** | ✗ |
| `[i64; 3]` (structural array) | **0** | ✓ **only by accident** — the 8-byte default stride equals i64's size |

So: type arguments that are **named** types substitute correctly through the
monomorphization path, which clones the type-alias AST with the generic param replaced and
re-runs `pre_infer_type_alias`. A **structural** array type expression does not survive
that substitution.

## Two hypotheses I formed and REFUTED — do not retry

1. *"The payload lives on a sum record SHARED across instantiations, so it is
   last-write-wins."* **False.** A program using `Option[(i64,i64)]` and `Option[[i64;3]]`
   together gets DISTINCT sumtypes records (`sumextra` 12 and 13) and returns the correct
   49. Each instantiation has its own record; the array one is simply registered empty.
2. *"Aliasing the array to a name would work around it."* **Not available** —
   `type Arr3 = [i32; 3]` is a PARSE ERROR; a type alias to an array type is unsupported.

## Actual root cause — `name_id` is overloaded as the array LENGTH

`register_array` (typetable.ax) stores `name_id: length, extra: elem_id`. **An array type
has no name at all.** The monomorphizer substitutes a generic param by writing the concrete
type's NAME into the node, so for `[i32; 3]` it wrote the *length* `3` as if it were an
intern id — a meaningless name resolving to nothing, hence the empty payload_type.

The fix mechanism already existed: tuple synth-structs have the same
"registered-but-not-name-resolvable" problem and are handled by stashing the exact type_id
in `extra_idx`, which the typechecker recovers **generically** whenever the name failed to
resolve. Only the PRODUCER-side guard was missing. `82d0565` marks an array type unnameable
so it uses the same stash — no new mechanism.

**Note the fix direction guessed above was wrong** (writing the payload from `args` in
`finish_generic_instantiation`). It would have papered over the substitution rather than
repairing it. A first attempt keyed on `entry.name_id == 0` also failed — never fired,
precisely because of the length overload — and that dead end is what exposed the real cause.

Gate: A==B `29BAF82C` + full regression 500/500. Oracle `t_arrpayloadwidth(166)` covers
i32/u8/i16 payloads alongside i64, so the previously-accidental i64 case now sits beside
widths that exercise the real path.

**Footgun worth remembering: `TypeEntry.name_id` does NOT always hold a name.** For an
ARRAY it is the length. Any code doing name-based reasoning over arbitrary type entries
must exclude arrays.

Related: [[probe-boxed-payload-2026-07-22]], [[bug-opt-tuple16-deref-caller-clobber]].
