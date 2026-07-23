---
name: compiler-selfhost-peak-memory-8gb
description: RESOLVED 2026-07-23 (5bfd5c7) — self-host peak 7.7GB->1.2GB. Root cause was clone_subtree_from leaking one ~2MB src buffer per cloned token during monomorphization; the fix frees spent intermediates
metadata:
  type: project
---

## ✅ RESOLVED 2026-07-23 (`5bfd5c7`) — peak 7,722 MB → 1,231 MB (−84%)

The headline defect below is FIXED. Root cause was **not** per-node retention,
mono clones, or the type table — it was a single leak in `clone_subtree_from`
(`ast.ax`). It rebuilt the destination tree's monolithic `src` string by
`self.src = std.string.concat(old_src, tok_text)` **once per cloned token** and
never freed `old_src`. With `src` starting at the full ~2 MB self-host source and
thousands of tokens cloned across all generic instantiations, each append leaked
a fresh ~2 MB buffer.

**How it was found (decisive, one instrumented build each):** a temporary
allocator probe at `__ax_runtime_shutdown` printed the small-vs-large split. Of
~8 GB live at shutdown, the small pool held only **268 MB** (slab 4099/16384 — far
from the 1 GB cap, so the get_str-copy small-leak theory was wrong) and **~7.76 GB
was ~3,936 LIVE large buffers averaging ~1.96 MB each** — almost exactly the
1.98 MB whole-program `src`. A size histogram put 3,936 of them in the 512 KB–4 MB
bucket. "~2 MB buffer allocated per cloned token" pointed straight at the concat.

**Fix:** free each spent intermediate concat buffer; `concat` already copies the
full prior contents forward so every token offset still resolves and nothing else
holds a pointer into the old buffer. A new `AstTree.orig_src` field guards the
tree's ORIGINAL caller-owned src (possibly not heap-allocated) so it is never
freed — the guard fires exactly once, on the first concat. Memory-only; emitted
bytes unchanged (fixpoint **A==B = E4E10E2D…**, regression **518/518**).

**Lesson:** the extensive prior analysis in this note (below) concluded "per-node
footprint, ~1.9 KB/source-byte, no O(n²)" and sent the reader to instrument
per-AST-node in typecheck. That framing was a **dead end** — the real cost was a
handful (~4 K) of whole-source-sized buffers from ONE quadratic string-append
loop in monomorphization, invisible to per-node reasoning. The synthetic
"1.9 KB/byte, linear" measurements were misleading because small test programs
have few generic instantiations, so the concat leak stayed small and looked like
a uniform per-byte constant. **The decisive move was measuring the allocation
SIZE distribution, not reasoning about what attaches to nodes.** Everything below
predates the fix; keep for the measurement methodology, not the (wrong) handoff.

---


**Measured 2026-07-23.** A healthy `-O1 -self-link` build of `tmp_concatenated_air.ax`
(1.9 MB of source) peaks at **~7,865 MB working set**. That is the build that SUCCEEDS.

That single number explains a blocker that took four reframings to corner
([[bug-iface-variant-payload-no-vtable-box]]): adding one call site to `lower_call_expr` makes
the build die with `AXIOM RUNTIME PANIC: Out of memory`. It is not a call-site ceiling, not an
optimizer pass, and not an allocator cap — **the build already runs at the edge of available
memory, and anything that pushes it further falls off.**

Ruled out along the way, each by measurement:

- **Not an artificial allocator cap.** `std/mem/alloc.ax` declares `SEGMENT_SIZE 65536` and
  `MAX_SEGMENTS 4096` (which would be a 256 MB cap), and the AXIOM allocator **never references
  `MAX_SEGMENTS`** — nothing enforces it on the self-hosted path, which is why the build sails
  past 256 MB to 7.9 GB.

  **Correction to an earlier draft of this note:** I first wrote that it is a dead constant and
  should be deleted. It is not dead. `runtime/axalloc/*.c` uses it for real — `static
  AxiomSegment axiom_segment_slab[AXIOM_MAX_SEGMENTS]` — so the C runtime genuinely caps at
  4096 segments, and the `.ax` copies (`std/mem/alloc.ax:22`, `bootstrap/runtime/axalloc.ax:20`)
  exist to mirror it. The real finding is a **divergence**: the cap is enforced in the C
  allocator and absent in the AXIOM one. Do not "clean up" the constant; the two allocators
  disagree, and that is the thing worth reconciling.
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

### The cost is per-TOKEN/NODE, not per-declaration

Two sources of opposite shape, marginal cost taken over the ~264 MB floor:

| source | bytes | peak | marginal | rate |
|---|---|---|---|---|
| 1500 tiny functions | 108,092 | 474 MB | 210 MB | **1.99 KB per source byte** |
| 1 function, 1500 statements | 73,257 | 400 MB | 136 MB | **1.91 KB per source byte** |

Within 4% of each other, and both match the 1.89–2.02 range from the size sweep. Maximising
the number of functions — and with it symbols, signatures, and types — costs **no more per byte**
than putting the same statements in a single function.

So the memory is **not** in symbol tables, type tables, or per-declaration structures. It is
spread uniformly across the token/AST representation. Given ~1.9 KB per source BYTE and a
token averaging a handful of bytes, the per-node cost is on the order of kilobytes — which for
an AST node is the anomaly worth chasing.

**Sharpest available handoff:** look at what is allocated per AST node and per token during
typecheck, not at the tables. `node_types` is already cleared (4 bytes per node, grows by
doubling, frees correctly), so it is something else attached to nodes.

### `-ctgc-free` buys nothing here — measured

The obvious lever is the shipped compile-time-GC free pass, since the diagnosis is "nothing is
freed". It does not help:

| build | peak |
|---|---|
| default | 7,865 MB |
| `-ctgc-free` | 7,866 MB |

Identical. That is not a surprise in hindsight — the escape analysis already reported
**freeable = 0** on the whole self-host build ([[ctgc-p3-scoping-2026-07-18]]), meaning every
owning local in the compiler escapes and CTGC has nothing it can legally free. This measurement
turns that analyzer claim into an observed fact, and closes the cheapest-looking route.

So the headroom has to come from structure, not from switching on an existing pass: freeing
intermediate phase data once a phase is done, or shrinking per-node footprint.

### Already checked and NOT the cause (do not repeat)

- **`typecheck.ax` does not itself allocate much.** Zero `get_token_text` calls, 7 `alloc_str`,
  10 `std.string.concat`, against 36 `@free`. The phase spends its memory in what it invokes,
  not in string churn in that file.
- **`node_types` is not a leak.** It grows by doubling (`typecheck.ax:1804`), copies, and
  `@free`s the old buffer. Correct as written.
- **`AstNode` is 24 bytes** (`ast.ax:121`) in a flat `NodeVec`. A 2 MB source is a few MB of
  nodes — the AST itself is nowhere near the budget.
- **Binding a struct VALUE out of an array does NOT allocate.** This was the scariest
  hypothesis, because `let node = tree.nodes.data[i]` is the compiler`s most pervasive idiom
  and AXIOM aggregates are reference semantics (RFC 0001 §5) — if each such binding heap-copied
  24 bytes and never freed, that alone would explain everything. Tested directly with a program
  doing **20,000,000** such bindings in a loop: **peak 4 MB**. If each allocated and leaked it
  would have been ~480 MB. The idiom is clean.

So the remaining suspects are the structures the phase builds up rather than the code that
walks them: the type table (`register_option`/`register_result` linear-scan and push), generic
monomorphization, and whatever the AST retains per node. Instrumenting is awkward — printing
per-node from inside typecheck is itself enough to change the picture, and adding a call to a
hot lowering function is what started this whole investigation. Prefer sampling working set
against `--time` (as done above) over in-process counters.
