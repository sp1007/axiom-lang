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

## Where it goes: nowhere, it just accumulates

Both follow-up questions are now answered by measurement.

**Peak does NOT scale with optimization or relocation count.** `-O0` and `-O1` land within
33 MB of each other on nearly identical reloc counts:

| level | peak | relocs | syms |
|---|---|---|---|
| `-O0` | 7,833 MB | 15,314 | 2,446 |
| `-O1` | 7,866 MB | 15,279 | 2,446 |

So the ~30 MB that `-O1` adds is the entire margin between building and not building. The
build is on a knife edge, and which side it lands on is close to arbitrary.

**It is not a link-time spike — it is monotonic accumulation.** Sampling working set across a
whole build:

| point in run | working set |
|---|---|
| 10% | 525 MB |
| 25% | 3,816 MB |
| 50% | 6,594 MB |
| 75% | 7,295 MB |
| 90% | 7,638 MB |
| max | **7,865 MB** |

Memory climbs from start to finish and **never drops**. Peak is therefore approximately
everything the compiler ever allocated: it frees essentially nothing across a build. The
linker is where the OOM surfaces only because the linker runs last, at the top of the ramp —
not because the linker is the consumer.

That is a coherent picture rather than a mystery, and it points at the compiler`s standing
posture of not freeing (`-ctgc-free` is off by default everywhere). The cost of that posture is
now measured: ~8 GB, and the ceiling has been reached.

**Where to start:** the 10% -> 25% segment, which jumps 525 MB -> 3.8 GB, is the single
largest step and worth attributing first. Reducing retention anywhere buys headroom
immediately, since the margin that decides success is ~30 MB out of ~7,900.
