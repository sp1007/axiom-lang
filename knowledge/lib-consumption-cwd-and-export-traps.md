---
name: lib-consumption-cwd-and-export-traps
description: Two traps when testing .lib/.dll consumption — the compiler resolves std/ and imports relative to CWD, and only #[export] pub fn is linkable via -l
metadata:
  type: reference
---

Two independent traps, both of which produce a **segfault** rather than a build failure, and
both of which cost real time when hit together.

## 1. Run the compiler from the REPOSITORY ROOT

`std/*.ax` and any `import`ed module are resolved relative to the **current directory**. The
same applies to `import mylib` — the module file must sit in the CWD, which for a build needing
the stdlib means the repository root.

**FIXED 2026-07-23** (`36365fb`). This used to print eight anonymous `Error: open failed:`
lines with no filename and then emit a ~2 KB executable that segfaulted, so the user met a
crash rather than a build failure and the real cause sat several screens above the symptom.
Now it names the file, states the cause, and halts:

```
error: cannot open source file:
std/result.ax
note: the compiler resolves std/*.ax and imported modules relative to the CURRENT DIRECTORY
error: 1 stdlib source file(s) could not be read; aborting before code generation
```

Guarded by the `input-halt` row in `regression_repros.sh`, which pins all three properties —
non-zero exit, the diagnostic names the file, and no output is emitted. `tests/ffi/README.md`
documented `cd tests/ffi` plus relative paths and therefore could not work as written; fixed to
use root-relative paths.

## 2. `-l` links `#[export] pub fn`, not bare `pub fn`

`tests/ffi/axmath.ax` uses `#[export] pub fn ax_add`, and the consumer declares
`extern "C" fn ax_ax_add` — codegen's `ax_` prefix plus the source name. A bare `pub fn` is not
in the archive symbol index, so the `-l` path gives:

```
Linker Error: Unresolved external symbol 'ax_ax_bset'
```

and then **still emits an executable**, which segfaults on the first call.

## The consequence that actually mattered

Trap 2 made it look as though a non-`#[export]` `pub` function is simply not consumable from a
`.lib` — which would have meant the `--staticlib` no-pruning guard in
[[rfc-0031-dead-function-elimination]] was unnecessary. It is not: the **import-driven path
does consume it**. Verified directly —

```
library mylib ; pub fn pub_only(x: i64) -> i64: return x + 20
import mylib  ; fn main() -> i64: return mylib.pub_only(22)
axc build app.ax --auto-lib -self-link -O1   ->   exit 42
```

So a `pub` function that no `#[export]` names IS a legitimate entry point of a static library,
pruning it WOULD break that consumer, and the guard is empirically justified rather than merely
conservative. Two consumption paths with different symbol requirements is the thing to hold
onto: testing only `-l` would have produced the wrong conclusion.

## Note on the shared failure mode

Both traps emit a broken executable after reporting errors. Worth treating as a robustness
item: a build that reports `open failed` or `Unresolved external symbol` should not go on to
produce output, because the symptom the user actually sees is a segfault at run time rather
than a failed build.
