---
name: bug-array-typearg-mono-payload
description: "OPEN, fully diagnosed, ready to implement: monomorphizing Option/Result with a STRUCTURAL ARRAY type argument records the variant payload_type as 0, so the pattern var is untyped and OP_INDEX falls back to stride 8. Named type args (struct, and tuples via __tupN) work. Fix = set the payload from the concrete arg instead of relying on AST substitution."
metadata:
  node_type: memory
  type: project
---

# OPEN — array type argument loses the mono variant payload type

The last survivor of the 2026-07-22 boxed-payload sweep
([[probe-boxed-payload-2026-07-22]]). **Diagnosis is complete**; only the implementation
is left, and it is deliberately deferred because it touches generic monomorphization.

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

## Fix direction

Set the monomorphized variant's `payload_type` from the CONCRETE type argument directly,
rather than relying on the cloned AST's type expression to re-resolve. `finish_generic_
instantiation` already has `args` in hand and already calls `set_sum_generic_args` for a
fresh type-alias instantiation — the payload write belongs beside it. That is preferable to
teaching AST substitution to synthesize array type expressions.

Fixing this also removes the accidental correctness of `[i64;3]`: with a real payload type
the DEREF and the index stride both follow the declared element size, so `[u8;N]`,
`[i16;N]`, `[f32;N]` start working too.

**Gate when implemented:** generic monomorphization is self-host-critical → A==B plus the
FULL regression on the newly built compiler, and check `t_arrpayload` still passes for the
right reason rather than the accidental one.

Related: [[probe-boxed-payload-2026-07-22]], [[bug-opt-tuple16-deref-caller-clobber]].
