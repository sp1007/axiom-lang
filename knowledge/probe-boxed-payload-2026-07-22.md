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

## OPEN #2 — a TUPLE payload in a USER sum: field 1 reads 0

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

24-byte (3-element) tuple behaves the same: `3 + 0 + 5 = 8`. A **struct** payload in the
same position is CORRECT, so this is tuple-specific, and `Option[(i64,i64)]` is now
correct too, so it is specifically the USER-sum path. Prime suspect:
`lower_variant_construct`'s `is_multifield` branch treats a flagged synth struct payload
by iterating the CALL ARGS as fields (`arg[i] -> field i`). A tuple payload is ONE arg
that is itself a flagged synth struct, so field 0 gets it and field 1 is never written.
**Verify that reading before changing it** — RFC 0019 multi-field variants depend on the
same branch.

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
