# RFC 0026 — Function Inlining & Self-Recursion→Loop (M6 perf)

- Status: **P1 + P1.5 + OP_INDEX SHIPPED** — pure-scalar single-block inliner (`d64a68d`, hotloop
  2.57x) + scalar-field/multi-field getter inlining (`f286cac9`, getter 2.89x) + scalar array-element
  getter inlining (`343fa03b`, 418/418). The getter family is complete (field + index). See
  [[m6-perf-gate-fib-benchmark]]. Groundwork in `8d05f96`. P2 (control-flow inliner for collatz;
  accumulator-recursion→loop for fib) remains — the real fib/collatz-gate closers, each a dedicated
  B==C-gated session.
- Author: autopilot (2026-07-18), per user direction [[autopilot-direction-2026-07-18]]
- Affected: `bootstrap/stage1/ssa_opt.ax` (optimizer pipeline), AIR only. No syntax,
  no ABI, no linker, no runtime changes.

## 1. Motivation

The M6 gate (`docs/tasks/milestones.md`) requires native code within **≤1.05x of clang -O2**
on Fib(40). Measured baseline (daily driver, `scripts/bench_perf.sh`, best-of-5):

| bench | AXIOM -O1 | clang -O2 | ratio |
|---|---|---|---|
| fib(42) | 2442 ms | 899 ms | **2.72x** |
| collatz | 170 ms | 53 ms | **3.19x** |

From clang's disassembly ([[m6-perf-gate-fib-benchmark]]): the fib gap is almost entirely
**call/frame overhead** — clang emits ONE `call` per activation by turning the double
recursion into a loop, while AXIOM emits two real calls plus prologue/epilogue each time.
The two highest-leverage, broadly-applicable transforms AXIOM is missing:

1. **Function inlining** of small callees — removes call/frame overhead everywhere (not
   just fib), exposes constants across the call boundary for the existing fold/CSE passes.
2. **Self-recursion→loop** — rewrites a self-tail-recursive (or accumulator-recursive)
   function into an in-function loop, matching clang's fib transform.

## 2. Design

AIR is a flat 16-byte-instruction list (`AirInst{opcode,type_id,dest,src1,src2}`) grouped
into `BasicBlock`s that reference `[instrs_start, instrs_len)` ranges (`air.ax`). `OP_CALL`
carries the callee **symbol index** in `type_id`, the result vreg in `dest`, and an
`extras`-offset arg list in `src2`. Both passes run **module-level, before** the existing
per-function pipeline in `SsaOptimizer.run` (`ssa_opt.ax:1798`), so their output is cleaned
up by the downstream fold→simplify→copy-prop→cse→dce loop.

### 2a′. P2 control-flow inliner — technical plan (investigated 2026-07-18, NOT yet built)
To inline a MULTI-block callee (e.g. `collatz_len` with its `while`), the linear machinery
(shipped: single-block, param-copy + offset-rename + return→copy + per-block f.insts/block_instrs
rebuild) must be extended with CFG surgery:
1. **Block-ID remap**: fresh callee block ids = `max_block_id(caller)+1 + callee_block`. Clone all
   callee blocks; remap every terminator's target (OP_JUMP `src1`, OP_BRANCH targets) to the new ids.
2. **Split the caller block at the call**: `[pre-call insts]` end with a JUMP to the callee entry;
   a new continuation block holds `[post-call insts]`. Callee `OP_RETURN v` → `copy call.dest=v'` +
   JUMP to the continuation.
3. **⚠️ CFG REBUILD IS THE TRAP**: `block_succs`/`block_preds` are built ONLY via explicit
   `add_edge(src,dst)` calls during AIR construction (air.ax:355-385) — there is NO recompute
   function, and the passes (compute_loop_depths/liveness/licm) READ the stored edges directly. A
   control-flow inline must call a NEW `recompute_cfg(f)` that rebuilds succs/preds from each block's
   terminator. **B==C depends on edge ORDER** matching what add_edge produced (succs in the builder's
   emission order; preds in block-visit order). `recompute_cfg` MUST reproduce that exact ordering or
   a -O2-built compiler diverges. Verify by asserting recompute_cfg == the as-built CFG on every
   function before trusting it for inlined ones.
4. **Terminator encoding (found 2026-07-18, x86_selector.ax:1563-1569):** `OP_JUMP` target block id
   = `src1` (1 succ). `OP_BRANCH` cond = `src1` (vreg), TRUE-target (JCC-NE) = `src2`, FALSE-target
   (fall-through) = `dest` (2 succs). `OP_RETURN` = exit (no succ). So `recompute_cfg` per block:
   JUMP→succ[src1]; BRANCH→succ[src2,dest] with UNIQUE append (BRANCH where src2==dest = 1 succ);
   preds built by scanning source blocks in block order (matches add_edge's build order). The
   then/else succ ORDER (src2-first vs dest-first) must be validated empirically against B==C — flip
   if a -O2-built compiler diverges.
5. Gate: the split, then recompute_cfg, then the full B==C + -O2 regression + collatz benchmark.
This is milestone-scale (~200 lines + the ordering-exact recompute_cfg); do it in a dedicated session.
NOTE: `recompute_cfg` cannot be validated/committed standalone — unconditionally recomputing an
identical CFG is pure overhead (§10 dead code) — so it must land TOGETHER with the multi-block inliner
that calls it. All technical unknowns are now resolved; the implementation is mechanical + gate-heavy.

### 2a. Inlining (`inline_module`)
For each `OP_CALL` to a **directly-resolvable, non-extern, non-recursive** callee whose
body is **small** (≤ `INLINE_INST_THRESHOLD`, initially ~12 real insts) and **structurally
simple** (single basic block, no nested calls in the first increment):
1. Clone the callee's instruction range into the caller at the call site.
2. **Rename** every `dest` vreg to a fresh caller vreg (offset by the caller's current max
   vreg + 1); rewrite `src1`/`src2` operands through the same rename map.
3. **Bind params→args**: the callee's `params[i]` vreg is rewritten to the caller's i-th
   argument vreg (read from the call's `extras` list).
4. **Return→copy**: each `OP_RETURN v` in the clone becomes `dest_of_call = copy v'`
   (single-block callee ⇒ exactly one return in the first increment; multi-return deferred).
5. Delete the original `OP_CALL`. Rebuild the caller's block instr ranges.

**Guards (first increment, conservative):** skip if callee is recursive/mutually-recursive,
has >1 block, contains any `OP_CALL`, has aggregate/16B params or return (by-address ABI —
defer), is `is_async`, or is variadic/intrinsic. Skip the whole pass on `is_large` callers.

### 2b. Self-recursion→loop (`selfrec_to_loop`)
Only for a function whose **every** recursive call to itself is in **tail position**
(the call's result is returned directly, no post-call arithmetic): rewrite
`return self(a,b)` into `param0=a; param1=b; jump entry`. The accumulator form
(`return n * fib(n-1)`, NOT tail) is **out of scope** for this RFC — that needs an
accumulator-introduction transform (clang does it; higher risk) and is deferred to a P2.
fib itself is accumulator-recursive, so the fib ratio is primarily closed by **inlining**
in this RFC; the pure tail-self-recursion transform benefits the many tail-recursive
helpers in the compiler/stdlib.

## 3. Alternatives considered
- **Immediate-operand folding** (`CMP/ADD reg,imm`) — attempted twice, both **reverted**:
  broke B==C on sub-64-bit spilled operands ([[m6-perf-gate-fib-benchmark]]). Orthogonal;
  not revisited here.
- **Leaf prologue/epilogue trimming & rsp-churn removal** — tried; the CPU stack engine
  makes them ~free (code-size win only). Not a runtime lever.
- **Full CPS/trampoline recursion elimination** — far larger, unnecessary for the gate.

## 4. Drawbacks / risks
- Inlining in flat AIR requires correct vreg renaming + block-range rebuild; a bug can
  silently miscompile and **pass A-only while breaking B==C** (the exact failure mode of the
  reverted imm-fold attempts). Mitigation: hard gate (§6), tiny threshold, single-block-only
  first increment, revert-on-red.
- Code-size growth if the threshold is too high. Mitigation: conservative threshold + only
  callees with a single call-return shape.

## 5. Compatibility / migration
Pure AIR optimization, semantics-preserving. No source, ABI, IR-layout, or linker change.
No migration needed. Both passes are **off unless** the optimizer level is ≥ their gate
(inlining ≥ -O1; recursion→loop ≥ -O2, matching the structural-pass tier) and each is
independently revertible by deleting its call in `SsaOptimizer.run`.

## 6. Verification gate (NON-NEGOTIABLE — backend change)
Per CLAUDE.md §24 + [[feedback-fixpoint-async-rule]], a codegen/optimizer change ships ONLY
after:
1. **B==C fixpoint** — build compiler with A, use it to build B, use B to build C, require
   B==C (self-host stability; A==B is NOT sufficient for backend changes).
2. **Full regression** `REGTMP=bin/_regtmp AXC=bin/axc_native.exe bash scripts/regression_repros.sh`
   → 401/401 + the `-O2/-O3` opt-level guard block.
3. **Benchmark** `bash scripts/bench_perf.sh` — record the ratio delta; a change that does
   not improve (or regresses) the ratio is reverted.
4. New oracle(s) crossing the transform (e.g. `t_inline*`, `t_selfrec*`) added to the gate.

If any criterion is RED, **revert cleanly** and record the precise blocker in
[[m6-perf-gate-fib-benchmark]] — a well-characterized blocker is valid progress.
