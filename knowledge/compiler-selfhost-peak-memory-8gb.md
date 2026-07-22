---
name: compiler-selfhost-peak-memory-8gb
description: A successful -O1 self-link peaks at ~7.9 GB working set, close enough to the machine limit that one extra call site in a hot function tips it into OOM
metadata:
  type: project
---

**Measured 2026-07-23.** A healthy `-O1 -self-link` build of `tmp_concatenated_air.ax`
(1.9 MB of source) peaks at **~7,865 MB working set**. That is the build that SUCCEEDS.

That single number explains a blocker that took four reframings to corner
([[bug-iface-variant-payload-no-vtable-box]]): adding one call site to `lower_call_expr` makes
the build die with `AXIOM RUNTIME PANIC: Out of memory`. It is not a call-site ceiling, not an
optimizer pass, and not an allocator cap — **the build already runs at the edge of available
memory, and anything that pushes it further falls off.**

Ruled out along the way, each by measurement:

- **Not an artificial allocator cap.** `std/mem/alloc.ax` declares
  `SEGMENT_SIZE 65536` and `MAX_SEGMENTS 4096` (which would be 256 MB), but **`MAX_SEGMENTS` is
  never referenced anywhere** — a dead constant that reads like a cap and is not one. Worth
  deleting or wiring up; as written it will mislead the next reader exactly as it misled me.
- **Not an optimizer pass.** `--time` shows codegen completing (1868 ms, 15,281 relocs,
  2,447 syms, object written) before the OOM in stage 5, self-linking.
- **Not source size.** Call-free statements can be added freely; only statements carrying a
  call tip it over — consistent with each call site multiplying relocations, which is what the
  linker holds in memory.

## Why this matters beyond one bug

The compiler is **one small change away from being unable to build itself** on this machine.
Any feature that adds a call to a hot lowering function hits the same wall, and the symptom is
an OOM at link time with no connection to the change that caused it. Every future contributor
meets it as a mystery.

Treat the ~8 GB as the headline defect. The RFC 0029 interface fix is blocked by it, but so is
anything else of comparable size, and the honest fix is to reduce self-link memory rather than
to route individual features around it.

Cheap first questions for whoever takes it: what does the linker retain for 15,281
relocations and 2,447 symbols that could be streamed or freed; and does peak scale with
relocation count (measure at `-O0`, which succeeds, against `-O1`).
