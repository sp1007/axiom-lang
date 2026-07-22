---
name: bug-small-alloc-256mb-segment-slab-cap
description: FIXED — the small-object allocator capped at 256 MB; proven root cause of the "OOM size requested: 24" self-build failures, cap raised to 1 GB
metadata:
  type: project
---

**Status: FIXED 2026-07-23** — cap raised 4096 → 16384 segments (256 MB → 1 GB).

**Proven, not inferred.** A two-stage bootstrap experiment: (1) build a compiler carrying the
raised cap — succeeds; (2) promote it as the seed and re-apply a fix that had OOMed in every
location tried — **it compiles and reaches fixpoint**. A first attempt at this conflated the
compiler being BUILT with the compiler doing the BUILDING and produced a false refutation; the
seed carries the cap, so it must be rebuilt first.

Original analysis follows. This is the real cause of the
`AXIOM RUNTIME PANIC: Out of memory` that blocks [[bug-iface-variant-payload-no-vtable-box]],
and it is not what the surrounding evidence suggested.

## The cap

`std/mem/alloc.ax:375`, inside `alloc_segment_meta()`:

```
    let slab_used_ptr = std_mem_alloc_get_slab_used()
    if slab_used_ptr.* >= 4096 as i64:
        return null as ptr[Segment]
```

4096 segments × `SEGMENT_SIZE` 64 KB = **256 MB, a hard ceiling on SMALL allocations**. Past it
`ax_segment_acquire` returns null, `malloc` panics, and the message is
`OOM size requested: 24` — a tiny request failing while the process holds gigabytes.

## Why every other reading was misleading

- **The process peaks at 7.9 GB**, so "out of memory" looks like genuine exhaustion. It is not:
  large allocations (`ax_large_alloc`, >4096 bytes) go straight to the OS and are uncapped, so
  the process grows freely while the SMALL-object pool is stuck at 256 MB.
- **The machine has headroom.** 15.9 GB physical, 8.0 GB free, and a commit limit of 37.9 GB
  with 12.7 GB in use — roughly 25 GB of commit available. A 64 KB `VirtualAlloc` was never
  going to fail. That mismatch is what pointed here.
- **`MAX_SEGMENTS` looked unused and is not the enforcement.** The constant at
  `std/mem/alloc.ax:22` is genuinely never referenced; the limit is a **hardcoded `4096`** at
  line 375. So an earlier note in this repo saying "nothing enforces it" was wrong in the way
  that mattered: the value is enforced, just not through the named constant. That is precisely
  the drift a named constant exists to prevent, and it cost several probes.

## Why the margin is so thin

`~30 MB` decides whether the self-build succeeds — see
[[compiler-selfhost-peak-memory-8gb]]. That now makes sense: the build sits just under 4096
segments, and any change that adds small-object traffic on a hot path (one more call site
anywhere, in any function) tips it past 256 MB.

It also explains why every attempted fix location failed identically, and why an inert
statement in a cold function was fine — the inert branch allocates nothing.

## DANGER: five coupled sites, four of them magic numbers

Do not simply raise the `4096`. The slab is a fixed region sized to match it, and the two are
not derived from each other:

| site | value | role |
|---|---|---|
| `std/mem/alloc.ax:22` | `MAX_SEGMENTS = 4096` | declared, **never referenced** |
| `std/mem/alloc.ax:375` | literal `4096` | the cap that actually fires |
| `std/mem/alloc.ax:160` | `196608` | slab `VirtualAlloc` size (Windows) |
| `std/mem/alloc.ax:166` | `196608` | slab `mmap` size (Linux) |
| `std/mem/alloc.ax:352` | `196608` | slab `@memset` size at init |
| `runtime/axalloc/*.c` | `AXIOM_MAX_SEGMENTS 4096` | C runtime`s `axiom_segment_slab[]` |

**FIXED for the four `.ax` sites (2026-07-23).** All now derive from one constant: the cap reads
`>= MAX_SEGMENTS`, and the three region sizes read
`MAX_SEGMENTS * @compiler_intrinsic("size_of")[Segment]()`. Byte-for-byte the same values today
(48 x 4096 = 196,608), and the drift is now structurally impossible. Note the old code sized
the region with a literal while INDEXING it with `size_of(Segment)` (line 378) — so the two
could already have disagreed if `Segment` ever gained a field. They now cannot.

Changing the cap is consequently a **one-line edit**, which is what the fix directions below
assume. `runtime/axalloc/*.c` still keeps its own copy and would need the matching change.

`Segment` is 48 bytes (3 pointers + i32 + pointer + u32, padded), and **48 × 4096 = 196,608
exactly**. So raising the segment count without raising all three `196608` sites writes
descriptor #4097 past the end of the mapped region — silent heap corruption, in the allocator,
on both platforms.

That coupling is the deeper defect. One capacity decision is spelled out in five places across
two languages, four of them as bare literals, and the one place with a NAME is the one nothing
reads.

## Fix directions

1. **Raise the cap.** Mechanically trivial and unblocks the interface bug immediately. The
   segment slab is a fixed array, so raising 4096 costs slab descriptors, not committed heap —
   segments are still acquired lazily via `ax_os_alloc`. Needs the allocator gate
   (`CLAUDE.md` §11: deterministic, instrumented, stress-tested) and a decision on the new
   value.
2. **Use `MAX_SEGMENTS` at line 375** regardless of what value is chosen, so the constant, the
   `.ax` mirrors, and `runtime/axalloc/*.c` (which really does size
   `axiom_segment_slab[AXIOM_MAX_SEGMENTS]`) cannot drift apart again.
3. **Reduce small-object retention** in typecheck, which holds ~67% of the budget. The
   principled fix, and much larger.

(1) + (2) together are a small, well-scoped change that would unblock a shipped-feature crash.
Worth doing deliberately rather than as a drive-by, because it is the allocator.
