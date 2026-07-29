# RFC 0034 — IR pipeline evolution: Typed SSA, Memory SSA, Sea of Nodes, E-Graph

- Status: **DRAFT / partially recommended-against**
- Author: proposed by the project owner 2026-07-29, assessed the same day
- Affects: `air.ax`, `air_builder.ax`, `ssa_opt.ax`, `x86_selector.ax`, `x86_regalloc.ax`
- Gate impact: every stage touches codegen ⇒ **B==C mandatory per stage**

## 1. Motivation

The proposed target pipeline is:

```
AST → HIR → Typed SSA → Memory SSA → Sea of Nodes → E-Graph → InstSel → RegAlloc → Machine IR → PE/ELF
```

Today AXIOM runs: `AST → typecheck → AIR (flat, non-SSA vregs) → ssa_opt → x86_selector →
x86_regalloc → COFF/ELF`. The proposal is a serious modernization and three of its four
mid-end layers are worth real consideration. This RFC assesses each against **measured**
evidence gathered on 2026-07-29 (see `knowledge/m6-perf-baseline.md` and
`scripts/perf_asm_variants.ps1`), not against general reputation.

## 2. What the measurements say the compiler actually needs

The fib(40) gap vs clang decomposes into ~1.5x codegen × ~1.5x missing optimization. Pricing
each codegen deficiency in hand-written NASM gives, per fib(40) run:

| deficiency | cost | layer that fixes it |
|---|---|---|
| `mov/sub` instead of `lea`/`dec` | +115 ms | instruction selection |
| copy chains + an extra callee-saved reg | +101 ms | register coalescing |
| `jcc`+`jmp` instead of fallthrough | +42 ms | block layout |
| no shrink wrapping | +0.5 ms | — (not worth building) |
| rbp frame pointer | −17 ms | — (not worth building) |

**None of the top three is fixed by Sea of Nodes or an E-Graph.** They are fixed by better
selection, coalescing, and layout. This must anchor the priority discussion: a mid-end rewrite
does not, by itself, close the measured gap.

## 3. Assessment per layer

### 3.1 Typed SSA — **RECOMMENDED (highest structural value)**

The single most valuable item, and it is already implicitly demanded by the existing code:

- Every machine-IR peephole currently pays a **non-SSA tax**. `fuse_cmp_immediate` must scan the
  whole function to prove a vreg has exactly 2 references; `const_shift_amount` needs an explicit
  single-def guard; `bug-cse-redef-operand-miscompile` was caused precisely by a redefinition that
  SSA would have made impossible to express.
- The register allocator cannot do **live-range splitting** without SSA-ish structure, and
  splitting is what lets a value be volatile on one path and callee-saved on another.
- Recommended construction algorithm: **Braun et al. 2013, "Simple and Efficient Construction of
  SSA Form"** — on-the-fly, no dominance frontiers, no separate dominator-tree pass. It is the
  right fit for a self-hosted compiler because it is ~300 lines and easy to verify.
- Destruction: **Boissinot et al.** (or Sreedhar method III) for correct parallel-copy handling.

Cost: large but bounded. Risk: high (self-host critical) ⇒ must land behind a verifier
(`air_verify` extended with SSA dominance checks) and B==C per step.

### 3.2 Memory SSA — **DEFER until measured**

Buys store-to-load forwarding, dead-store elimination, and load hoisting in LICM. AXIOM already
has a connection graph and escape analysis that answer some of the same queries. No current
measurement shows memory optimization as a bottleneck (`arrwalk` is at 1.15x of clang, the best
non-trivial shape we have). **Revisit only when a benchmark demands it.** Adopting it before
Typed SSA is not possible anyway — Memory SSA is defined on top of SSA.

### 3.3 Sea of Nodes — **RECOMMENDED AGAINST**

This is the one item I would not adopt, for reasons specific to this project rather than taste:

1. **It collides with the project's central gate.** CLAUDE.md makes deterministic, reproducible,
   byte-identical output non-negotiable, and the B==C fixpoint is how every backend change is
   verified. Sea of Nodes deliberately *discards* instruction order and recomputes a schedule;
   scheduling is heuristic and famously sensitive to small changes. Determinism is achievable
   but it becomes something we must actively defend at every edit, in exchange for optimizations
   we can also get from CFG+SSA.
2. **It collides with the debuggability rules.** §9 requires IR that is printable, serializable
   and verifiable; §18 requires debugging strategy per subsystem. Sea of Nodes is notoriously
   hard to read and to diff — and diffing IR is how several bugs in `knowledge/` were actually
   found (`dump-air` output is cited repeatedly as the thing that cracked them).
3. **The industry data point runs the other way.** V8 moved Turbofan *off* Sea of Nodes to the
   CFG-based Turboshaft (2023), citing exactly maintainability, debuggability and compile time.
   HotSpot C2 keeps it, but C2 is also the part of HotSpot most often described as unmaintainable.
4. **It does not buy what we measured.** GVN and code motion — the headline wins — are available
   on CFG+SSA with a standard GVN pass and LICM (which AXIOM already has, RFC 0025).

Recommendation: keep an explicit **CFG + Typed SSA** mid-end. If global code motion is wanted
later, add a scheduler pass over SSA rather than changing the IR's identity.

### 3.4 E-Graph — **CONDITIONALLY LATER, and only in the bounded form**

Equality saturation removes rewrite phase-ordering problems, which is genuinely attractive for a
peephole/algebraic layer. But full saturation has unbounded compile time, and AXIOM's compile
speed is already a stated goal.

If adopted, adopt it the way **Cranelift** did: an *acyclic*, non-saturating e-graph over the
mid-end with a cost-model extraction, deliberately restricted so compile time stays linear-ish.
Note the implementation cost is real for a self-hosted compiler — union-find + hashcons +
rebuilding, written in AXIOM, plus a cost model and extraction pass.

Prerequisite: Typed SSA. Priority: **after** the measured codegen items are done.

### 3.5 Instruction selection — **the proposal understates this stage**

The pipeline lists "Instruction Selection" as a single box, but the measurements say it is the
**largest single line item** (+115 ms). Today the selector expands one AIR op into a fixed macro,
which is why `n-1` costs 4 instructions instead of `lea -1(%rbx),%rcx`. The named fix is
**BURS / bottom-up rewrite with a dynamic-programming cost model** (Aho–Ganapathi–Tjiang; `iburg`
style), or Cranelift's **ISLE** pattern DSL. This should be sequenced *with* Typed SSA, not after
the mid-end layers.

⚠️ History to respect: a `lea` peephole was tried post-selection and **regressed** (it removed a
copy, lengthened a live range, and the allocator spilled). The pricing above says the `lea` form
itself is worth ~115 ms, so the lesson is about *where* the transform belongs (in selection,
co-designed with coalescing), not about whether it is worth doing.

## 4. Recommended sequencing

| stage | content | why this order |
|---|---|---|
| **S0 (now)** | selection `lea`/`dec` + register coalescing + block layout | measured +258 ms; no architecture change; each independently gated |
| **S1** | **Typed SSA** for AIR (Braun construction, Boissinot destruction) + SSA verifier | removes the non-SSA tax; unlocks live-range splitting and real coalescing |
| **S2** | BURS/ISLE-style selection over the SSA IR | the largest measured item, done properly |
| **S3** | accumulator / tail-recursion-modulo-associativity pass | the other ~1.5x; independent of all the above |
| **S4** | Memory SSA — **only if** a benchmark demands it | no current evidence |
| **S5** | bounded acyclic E-Graph for the algebraic layer | needs S1; compile-time budget must be set first |
| **—** | Sea of Nodes | **not recommended**, see §3.3 |

## 5. Drawbacks

- S1 is a rewrite of self-host-critical code. Every step must hold B==C and 553+ oracles. Expect
  the fixpoint to break repeatedly during development; that is the gate working, not a setback.
- This sequencing delays the mid-end modernization behind unglamorous backend work. The
  justification is purely that the backend work is what the measurements price highest.

## 6. Migration constraints (non-negotiable)

1. B==C fixpoint before every commit that touches codegen.
2. The AIR text printer and verifier must be extended in the SAME change that introduces SSA —
   an IR we cannot print is an IR we cannot debug (§9, §18).
3. No stage may be merged while regression is below the current oracle count.
4. Each stage must be independently revertible; no batching of S0 items (each has its own delta
   to attribute).
