---
name: probe-boxed-payload-2026-07-22
description: "Probe of the boxed-aggregate-payload surface found three pre-existing silent miscompiles. #2 (tuple payload in a USER sum) FIXED 6667c68 — the variant ctor never got an expected type, so int literals stayed i32. STILL OPEN: #1 Option[[T;N]] array-payload element reads (NOT the same cause), #3 RFC 0019 multi-field variant bound to ONE name (SEGFAULT). Also open: variant-ctor arity is unchecked in both directions."
metadata:
  node_type: memory
  type: project
---

# Probe of the boxed-payload surface — 3 found, 1 fixed, 2 open

Run 2026-07-22 immediately after [[bug-opt-tuple16-deref-caller-clobber]] shipped
(`3319bdc`), on the theory that a region which just yielded one silent miscompile is
worth sweeping. It was: three more, all **pre-existing** — every one reproduces
identically on a compiler built from the pre-fix source (`44a1274`), so none is fallout
from the fix.

Bonus confirmation: the fix ALSO repaired `Result[(i64,i64), i64]` with a 16-byte `Ok`
tuple payload (43 → 50), which the `t_optupclobber` oracle does not cover.

## Clean on this sweep (do NOT re-probe)

`Result` Ok/Err 16B tuple payload · `Option[(f64,f64)]` · plain arrays · plain tuples ·
user sum with a **struct** payload · RFC 0019 `Pair(x, y)` two-name pattern.

## ✅ MOSTLY FIXED #1 — array payload element reads (i32-element residual OPEN)

**Fix `7bb826f`** — an array literal used as a variant-ctor payload is inferred TWICE:
once with the concrete expected array type, then again with UNKNOWN. A trace showed the
same node id twice — `expected=481 elem=4` (i64) then `expected=0 elem=3` (i32). The
hint-less re-visit re-registered `[i32; N]` and that write landed LAST, so elements stored
4 bytes wide while reads used the declared stride. `NODE_TUPLE_EXPR` already had a sticky
guard for exactly this double-visit; arrays had none. Added the analogue (recover the
element hint from the type a prior visit pinned on the node). Fixes user-sum AND builtin-
Option payloads, for indexed / single-element / loop reads. A==B `EB02B23D`, 497/497,
oracle `t_arrpayload(232)`.

A plain annotated `let v: [i64;3] = [..]` is visited once and was never affected — that
asymmetry is what made the bug look payload-specific.

### 🔴 RESIDUAL still OPEN — narrow-ELEMENT array payload (and why the i64 case only LOOKS fixed)

```axiom
let a: Option[[i32; 3]] = Some([3, 40, 5])
match a:
    Some(arr):
        return (arr[0] + arr[1] + arr[2]) as i64   // still 8, want 48
```

Here the declared and inferred types already AGREE (both `[i32;3]`), so the double-visit
is not the cause. **Traced 2026-07-22 to a payload-type recovery failure**, and the trace
also corrected an over-claim about the fix above:

| payload | pl_type recovered | bindsym | result |
|---|---|---|---|
| `[i32;3]` | **0 → defaulted to i64** | 0 | 8 ✗ |
| `[i64;3]` | **0 → defaulted to i64** | 0 | 48 ✓ **by accident** |
| `(i64,i64)` | 287 | 287 | 43 ✓ genuinely |

For an ARRAY payload, `lower_match_tagged`'s pl_type chain finds nothing —
`find_variant_info` yields no payload for the monomorphized Option-as-SUM (kind 6), and
the binding-symbol fallback is empty because the TYPECHECKER never typed the pattern var
(`bindsym=0`). pl_type therefore hits its final `= 4` (i64) default. `OP_INDEX` then sees
`insttype=0` with a base register typed i64 (not an array), so it cannot recover an
element type and its stride falls back to 8.

**So `[i64;3]` is correct only because the 8-byte default stride happens to equal i64's
size.** The `7bb826f` sticky fix was still necessary — it repaired the STORE side, which
was writing 4-byte elements — but it did not repair payload-type recovery. Any element
whose size is not 8 (i32/i16/u8/f32) is still read at the wrong stride.

Real fix = make the monomorphized Option/Result SUM carry its array payload_type (the
tuple payload already does, which is why tuples work), so the typechecker types the
pattern var and both the DEREF and the index stride follow. That touches Option
monomorphization/registration — self-host-critical, so it wants a dedicated gated
session, not a tack-on.

## ✅ FIXED #2 — a TUPLE payload in a USER sum read field 1 as 0

```axiom
type Shape = Pair((i64, i64)) | Empty

fn main() -> i64:
    let a: Shape = Pair((3, 40))
    match a:
        Pair(t):
            return t.0 + t.1     // 3, want 43
        Empty:
            return 0
```

24-byte (3-element) tuple behaves the same: `3 + 0 + 5 = 8`. Note the signature: field 1
is zero while fields 0 **and 2** are correct. A **struct** payload in the same position is
CORRECT, and `Option[(i64,i64)]` is correct since `3319bdc`, so this is specifically the
USER-sum path with a tuple payload.

### Root cause (found by TRACING, after two refuted structural guesses)

A non-generic user sum's variant constructor **never received an expected type for its
payload argument**. `try_instantiate_variant_call` threads a hint only INSIDE its
generic-parameter binding loop, so `Pair((3, 40))` inferred its tuple literal with no
expectation at all — and a bare int literal defaults to i32. That built an 8-byte
{i32,i32} value for a 16-byte payload slot, so the match arm read field 1 past the end.

The tell that identified it as a width gap rather than a lowering bug: declaring the
payload `(i32, i32)` made the *same* program correct. Explicit `(3 as i64, 40 as i64)`
also worked.

**Fix `6667c68`** — new `coerce_variant_ctor_payload` re-infers the single payload arg
with the variant's DECLARED payload type, mirroring the expected-type threading already
done for struct-ctor fields (`6132b15`) and function params (`f509506`). Flagged RFC 0019
synth structs (multi-field / str wrap) and generic payloads are excluded.
A==B `7764E7F9`, 494/494, oracle `t_sumtupctor(102)`.

### Two structural hypotheses that were REFUTED on the way — do not retry them

1. *"`V((A,B))` and `V(A,B)` desugar to the same type, so nothing can tell them apart."*
   **False.** RFC 0019 multi-field payloads are synth structs named `__mfv_<sum>_<variant>`
   (typecheck.ax ~L2553) and tuples are `__tup<N>` (~L2715); the naming was deliberately
   chosen not to collide.
2. *"`lower_variant_construct`'s `is_multifield` branch iterates call ARGS as fields, so a
   single tuple arg fills only field 0."* **False.** That branch is gated on the payload
   struct's flag bit 0, and `do_wrap` only wraps-and-flags for MULTIPLE payload type exprs
   or a single >8-byte PRIMITIVE. A tuple payload stays an UNFLAGGED `__tup<N>`.

Both readings looked convincing; both were wrong. What settled it was tracing `CTOR` /
`MPAT` / `GETF`: the GET_FIELD sequences for the broken tuple payload and the working
struct payload were **byte-for-byte identical in shape** (same offsets, sizes, branches),
which ruled out the whole read path and pointed at construction. **Trace before theorising
in this area.**

Still unchecked anywhere: arity. `V((i64,i64))` accepts `V(3, 40)` (→ 6) and
`V(i64,i64)` accepts `V((3,40))` (→ 127), both silently wrong.

## ✅ FIXED #3 — RFC 0019 multi-field variant bound to ONE name SEGFAULTed

```axiom
type Shape = Pair(i64, i64) | Empty

fn main() -> i64:
    let a: Shape = Pair(3, 40)
    match a:
        Pair(t):            // one binding for a two-field variant
            return t.0 + t.1
        Empty:
            return 0        // SIGSEGV 139
```

The supported form `Pair(x, y)` works (43). Binding a multi-field variant to a single
name typed the name as field 0 and let the body read `.0`/`.1` off a scalar.

**Fix `1146b75`** — an RFC 0019 synthesized payload binds one name per FIELD, so the
pattern's arity is now checked against the variant's and a mismatch is a clean diagnostic.
A==B `FBC2AA3F`, 496/496, oracles `t_mfvarity`(reject) + `t_mfvarityok`(58).

Comparing counts is sound ONLY because the two payload shapes are distinguishable — the
flagged `__mfv_…` synth struct vs the unflagged `__tup…`. Refuted hypothesis 1 below had
assumed they collapsed to one type and shelved this reject as unimplementable; that was
wrong. The guard oracle pins every legitimate shape: `Pair(x, y)`, an unflagged tuple
payload bound to ONE name, the str wrap `Text(msg)`, a plain scalar payload, and a
payload-less variant.

## Method note

The old-vs-new comparison is what makes these safe to report: build the previous commit's
`tmp_concatenated_air.ax` with the current driver to get a pre-fix BACKEND, then run both.
Without that step a probe run straight after a backend change cannot distinguish "found a
bug" from "caused a bug".

Related: [[bug-opt-tuple16-deref-caller-clobber]], [[rfc0019-multifield-variant-shipped]],
[[backlog-open-items]].
