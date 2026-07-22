---
name: bug-two-array-payload-instantiations
description: "OPEN, pre-existing: TWO distinct Option-with-ARRAY-payload instantiations in one program SIGSEGV. Either half alone is correct. Verified to reproduce on the pre-82d0565 compiler, so it is not fallout from the array type-arg fix. Repro bin/repro_twoarraypayload.ax (not registered — it crashes)."
metadata:
  node_type: memory
  type: project
---

# OPEN — two array-payload Option instantiations crash

Found 2026-07-22 while building a verification sweep for [[bug-array-typearg-mono-payload]].

## Minimal repro — `bin/repro_twoarraypayload.ax`

```axiom
struct P:
    x: i32
    y: i32

fn main() -> i64:
    mut r := 0 as i64
    let n: Option[Option[[i32; 2]]] = Some(Some([3, 40]))
    match n:
        Some(inner):
            match inner:
                Some(arr):
                    r = r + (arr[0] + arr[1]) as i64
                None:
                    r = r + 0
        None:
            r = r + 0
    let s: Option[[P; 2]] = Some([P(x: 1, y: 2), P(x: 3, y: 40)])
    match s:
        Some(arr):
            r = r + (arr[1].x + arr[1].y) as i64
        None:
            r = r + 0
    return r          // want 86; actual SIGSEGV
```

Deterministic (3/3 runs, and two builds are byte-identical). **Either half alone is
correct**: the nested `Option[Option[[i32;2]]]` returns 43 on its own, and the
`Option[[P;2]]` returns 43 on its own.

## Not a regression

Reproduces identically on a compiler built from the pre-`82d0565` source, so it is NOT
fallout from the array type-arg monomorphization fix. Both instantiations are
Option-with-an-ARRAY-payload, which points at their registration/mono interacting rather
than at either shape being wrong.

## Everything else in the sweep is CLEAN (banked as `t_arraygenerics`, exit 110)

generic fn with an array param · `Vec[[i32;3]]` element · `HashMap[str,[i32;3]]` value ·
`Result[[i32;3],i64]` Ok payload · `[(i64,i64);2]` array-of-tuples · array return value ·
`for x in arr` over an array payload · `Option[[P;2]]` struct elements. A parser gap was
also noted: an explicit array type ARG in a ctor call (`Box[[i32;3]](v: …)`) is a parse
error, while the same syntax in TYPE position (`let b: Box[[i32;3]] = Box(v: …)`) works —
clean reject, two workarounds, low priority.

## ⚠️ How this was nearly mis-reported — 8-bit exit codes, again

The first bisect produced a chain of "failures" (270→14, 275→19, 318→62, 366→110) that
were **entirely my own arithmetic error**: bash exit codes are 8 bits, and 270 & 0xFF = 14.
Every one of those was CORRECT. Only the SIGSEGV survived re-checking.

This pitfall is already banked ("keep oracles 0..255") and I walked into it anyway. When a
probe's expected value can exceed 255, either keep the accumulator small or subtract — and
treat a "wrong" result that differs from the expectation by exactly 256 as the wraparound
until proven otherwise. A crash (139) is the one signal that cannot be an exit-code
artifact, which is why the bisect was redone using crash-vs-no-crash alone.

Related: [[bug-array-typearg-mono-payload]], [[probe-boxed-payload-2026-07-22]].
