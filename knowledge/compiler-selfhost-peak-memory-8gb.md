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

## Attributed: TYPECHECK holds ~72% of it

Correlating the working-set series against `--time` phase boundaries (total 7,900 ms):

| phase | window | memory across it | share |
|---|---|---|---|
| concat/lex/parse/resolve | 0–663 ms | 2 MB → ~450 MB | 6% |
| **typecheck** | **663–3770 ms** | **~450 MB → ~5,715 MB** | **~67%** |
| air-build + ssa-opt | 3770–4276 ms | 5,715 → ~7,191 MB | ~19% |
| codegen | 4276–6382 ms | 7,191 → 7,311 MB | ~2% |
| self-link | 6382–7264 ms | 7,373 → 7,865 MB | ~6% |

**Typecheck is the consumer**, at roughly 5.3 GB of the 7.9 GB, and it is also the slowest
phase (3,107 ms of 7,900). Codegen — which handles 15,279 relocations and 2,447 symbols — adds
only about 120 MB. The linker adds ~500 MB.

This is worth stating loudly because it is the opposite of where the investigation kept
pointing. The OOM message names the linker, the failing statement was in the lowering code, and
both earlier versions of this note sent readers to `linker.ax` and to the optimizer. The actual
budget is spent in `typecheck.ax`, three phases earlier, and nothing about the crash says so.

**Where to start:** `typecheck`. Roughly 5 GB is retained there across a build that never
frees. Since the margin between success and OOM is ~30 MB out of ~7,900, even a small reduction
in typecheck retention buys real headroom — and it is where the headroom is.

### It is LINEAR with a huge constant — there is no O(n^2) to hunt

Slope analysis on generated sources of doubling size (same shape, `-O1 -self-link`):

| source | bytes | peak | marginal cost |
|---|---|---|---|
| m250 | 23,172 | 308 MB | — |
| m500 | 46,422 | 355 MB | 2.02 KB RAM per source byte |
| m1000 | 92,922 | 444 MB | 1.91 KB per byte |
| m2000 | 186,922 | 622 MB | 1.89 KB per byte |

The slope is **constant across three doublings**, so consumption is **linear in source size**.
There is no quadratic structure to find, which rules out the most tempting class of
explanation and saves whoever picks this up from hunting one.

What is wrong is the **constant: roughly 1.9 KB of RAM per BYTE of source, about a 1,900x
blowup**, plus a ~264 MB floor for any program at all. Extrapolating the slope to the compiler`s
2,012,779 bytes predicts ~4.1 GB; the measured peak is 7.9 GB, so the compiler`s own source
costs about 1.9x the synthetic rate — denser code (generics, monomorphization, more types), not
a different asymptotic class.

This fits the monotonic never-dropping trajectory exactly: the compiler retains essentially
every structure it builds, at a uniform high per-byte cost. **So the work is reducing retention
and per-node footprint, not finding an algorithmic blowup.**

### Already checked and NOT the cause (do not repeat)

- **`typecheck.ax` does not itself allocate much.** Zero `get_token_text` calls, 7 `alloc_str`,
  10 `std.string.concat`, against 36 `@free`. The phase spends its memory in what it invokes,
  not in string churn in that file.
- **`node_types` is not a leak.** It grows by doubling (`typecheck.ax:1804`), copies, and
  `@free`s the old buffer. Correct as written.

So the remaining suspects are the structures the phase builds up rather than the code that
walks them: the type table (`register_option`/`register_result` linear-scan and push), generic
monomorphization, and whatever the AST retains per node. Instrumenting is awkward — printing
per-node from inside typecheck is itself enough to change the picture, and adding a call to a
hot lowering function is what started this whole investigation. Prefer sampling working set
against `--time` (as done above) over in-process counters.
