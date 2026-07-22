---
name: probe-boxed-payload-2026-07-22
description: "OPEN x3 — probe of the boxed-aggregate-payload surface right after fixing the 16B-tuple deref clobber found three MORE pre-existing silent miscompiles: Option[[T;N]] array-payload element reads, a TUPLE payload in a USER sum (field 1 = 0), and an RFC 0019 multi-field variant bound to ONE name (SEGFAULT). All verified identical on the pre-fix backend."
metadata:
  node_type: memory
  type: project
---

# Probe of the boxed-payload surface — 3 new OPEN bugs

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

## OPEN #1 — `Option[[T; N]]` array payload: element reads are wrong

**NOT the same root cause as #2** (checked): an explicitly-i64 array literal
(`Some([3 as i64, ...])`) IS correct, but unlike the tuple case, declaring the payload
`[i32; 3]` does **not** make plain literals work — it still returns 8. So an element-width
component exists, plus a second mechanism that is still unidentified.

```axiom
fn main() -> i64:
    let a: Option[[i64; 3]] = Some([3, 40, 5])
    match a:
        Some(arr):
            return arr[0] + arr[1] + arr[2]   // 8, want 48  (element 1 reads 0)
        None:
            return 0
```

Reading `arr[1]` **alone** returns **5** — element 2's value — so the wrong value is
context-dependent (register allocation), not a fixed zero. A loop-indexed read gives the
same 8. A 2-element array is equally wrong (3, want 43). Needs no function call; plain
arrays outside a payload are fine. Smells like the payload base address being off by an
element, or the array payload being treated as inline where it is a reference.

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

**Fix `4b8f2c1`** — new `coerce_variant_ctor_payload` re-infers the single payload arg
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

## OPEN #3 — RFC 0019 multi-field variant bound to ONE name: SEGFAULT

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
name is presumably meant to be unsupported — but it is an accept-then-SEGFAULT, i.e.
BUG#53 class. Correct outcome is a clean diagnostic at the pattern, the same convention
used for inline match arms and non-sum scrutinees. This is the cheapest of the three
(frontend reject, A==B gate).

## Method note

The old-vs-new comparison is what makes these safe to report: build the previous commit's
`tmp_concatenated_air.ax` with the current driver to get a pre-fix BACKEND, then run both.
Without that step a probe run straight after a backend change cannot distinguish "found a
bug" from "caused a bug".

Related: [[bug-opt-tuple16-deref-caller-clobber]], [[rfc0019-multifield-variant-shipped]],
[[backlog-open-items]].
