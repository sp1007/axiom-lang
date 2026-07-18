# RFC 0026 — Function Inlining & Self-Recursion→Loop (M6 perf)

- Status: **P1 + P1.5 + OP_INDEX + P2 CONTROL-FLOW INLINER + §2b TAIL-SELF-RECURSION→LOOP SHIPPED** (§2b: B==C `ECABD5EF`, 427/427, tailrec 4.9x) — pure-scalar single-block
  inliner (`d64a68d`, hotloop 2.57x) + scalar-field/multi-field getter inlining (`f286cac9`, getter
  2.89x) + scalar array-element getter inlining (`343fa03b`, 418/418) + **P2 multi-block (control-flow)
  inliner** (this session, B==C `4138AEB5`, 421/421, cfhot 12% win). The getter family is complete
  (field + index). See [[m6-perf-gate-fib-benchmark]]. Groundwork in `8d05f96`. P2 built the
  `recompute_cfg` groundwork (§2a′, ordering PROVEN exact via unconditional-rebuild B==C) + the
  RPO-ordered block clone. **Remaining P2: accumulator-recursion→loop for fib — DEFERRED by
  determination 2026-07-18** (fib is TREE-recursive `fib(n-1)+fib(n-2)`, not tail → needs an
  accumulator/recurrence-introduction transform whose call reassociation is only sound for
  provably-pure callees → high B==C self-host risk for narrow single-benchmark ROI; §10 don't force.
  Not "todo" — deferred until a dedicated, user-prioritized perf session; rationale at top of
  [[m6-perf-gate-fib-benchmark]]). collatz gains from P2 only at -O2
  (its inner loop dominates at -O1; call removal is negligible there) — the clear P2 win is on
  call-dominated multi-block code (cfhot: `classify()` with 2 `if`s, 200M calls).
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

### 2a′. P2 control-flow inliner — SHIPPED (this session)
IMPLEMENTED as `inline_multiblock_func` (ssa_opt.ax) + `recompute_cfg` (air.ax), gated at -O1 after
the single-block `inline_func`. Confirmed self-host-safe: B==C `4138AEB5`, 421/421 (incl. -O2/-O3),
oracle `t_inlinecf` (exit 17). The plan below held, with ONE addition found during bring-up:
**callee blocks MUST be cloned in reverse-post-order, not the callee's index order.** A callee's
merge/exit block can sit physically BEFORE its predecessors (e.g. `floorf`'s exit is block index 1);
cloning that verbatim puts a value USE physically before its (conditional) redefinition, which the
physical-order value/regalloc passes miscompile once inlined into a larger function (standalone it
happens to survive; inlined it broke t_mathfill/t_fft at -O1 while B==C still held — the exact
"passes A/self-build but breaks user code" hazard §4 warns about). Fix: RPO DFS from the callee entry
(`inl_rpo_dfs`), clone in that order, remap terminator targets through the RPO position map. The
`recompute_cfg` ordering (below) was PROVEN exact independently by wiring it unconditionally on every
function and confirming a bit-identical B==C + 418/418 (a true no-op iff ordering-exact).

Original plan (kept for reference):
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

### 2b. Self-recursion→loop (`selfrec_to_loop`) — ✅ SHIPPED (B==C `ECABD5EF`, 427/427, tailrec 4.9x)
IMPLEMENTED as `selfrec_to_loop` + `sr_is_tail_block` (ssa_opt.ax), gated -O1, run LAST (after the
value passes), + the `emit_param_prologue`-before-`.L_b_0` move (option B, x86_selector.ax). Detects
`%r = call self(args); ret %r` (adjacent, block-terminating, direct, argc==nparams); rewrites to
two-phase staged `temp_i = arg_i` then `param_i = temp_i` (staging is REQUIRED for swap-style tail
calls like `gcd(b, a%b)`), then `jump block_0`; adds the back-edge via `recompute_cfg`. THREE things
that mattered in bring-up: (1) **option B** (param prologue before the entry label) is mandatory —
else `jump block_0` re-materializes params from arg registers → infinite loop; (2) run it **LAST** —
copy_prop would collapse `param=copy temp` (params have no other AIR def) and rewrite the entry's
param reads to the loop-only temp → garbage on first entry; (3) **no contiguity guard** — the tail
block commonly sits at a LOW index but is emitted LAST (`if-return / else-tailcall`), so
`inl_blocks_contiguous` is false; the rebuild re-lays-out block_instrs in block order and (running
last, no physical-order pass follows) that's fine. Non-tail (`n*fact(n-1)`) and mixed tail/non-tail
(ack) correctly untouched. Perf A/B: `benchmarks/tailrec` (deep tail sumto in a loop) OFF 1214ms →
ON 246ms = **4.9x**. Oracle `t_selfrec` (accumulator + swap gcd, exit 67). Accumulator-recursion for
fib (`return n*fib(n-1)`, tree-recursive) remains out of scope (needs a recurrence transform).

Original plan:
Only for a function whose **every** recursive call to itself is in **tail position**
(the call's result is returned directly, no post-call arithmetic): rewrite
`return self(a,b)` into `param0=a; param1=b; jump entry`. The accumulator form
(`return n * fib(n-1)`, NOT tail) is **out of scope** for this RFC — that needs an
accumulator-introduction transform (clang does it; higher risk) and is deferred to a P2.
fib itself is accumulator-recursive, so the fib ratio is primarily closed by **inlining**
in this RFC; the pure tail-self-recursion transform benefits the many tail-recursive
helpers in the compiler/stdlib.

**FEASIBILITY (2026-07-18, read-only) — one hazard CLEAR, a SECOND hazard is the real blocker:**
- ✅ **Frame prologue is safe.** `emit_function` (x86_asm_emitter.ax:512-527) emits `fn_name:` →
  frame prologue (push rbp / sub rsp / callee-saves) → THEN the blocks. The entry block's label
  `.L_b_0:` is after the frame prologue, so `jump .L_b_0` does NOT re-run frame setup.
- ❌ **The PARAM prologue IS re-run — this is the blocker.** `select_all` (x86_selector.ax:2494-2510)
  emits, per block: `MACH_LABEL(.L_b_{id})` FIRST, then (only for `bi==0`) `emit_param_prologue`
  (the `MOV pvreg = arg_reg` snapshot that materializes params from the incoming ABI arg registers).
  So the param materialization lives INSIDE block 0, AFTER its label. A `jump .L_b_0` re-executes it
  → every param vreg is reset to the ORIGINAL incoming arg register → the tail-recursion's updated
  args are clobbered → infinite loop with the original arguments. So jumping to block 0 is WRONG.

**Two ways to fix (both need a full B==C + -O2 regression gate — codegen change):**
- **(B, cleanest) Move `emit_param_prologue` to BEFORE the first block label** (into the true
  prologue region of `emit_function`/`select_all`, before the `bi==0` label push). Then params
  materialize exactly once on entry and block 0 is re-entrant → `jump .L_b_0` is safe. General
  improvement; risk = it changes emitted code for EVERY function (verify B==C holds + the `is_main`
  runtime-init ordering at bi==0 is untouched).
- **(A) Insert an empty pre-header as the new block 0** (carries the param prologue) that jumps to
  the old entry (now block 1); tail-calls target the old entry, skipping the param prologue. Needs a
  block-renumber (+1) + terminator remap (the P2 machinery), + recompute_cfg.

The AIR transform itself is then simple (detect `%r=call self; ret %r`; stage args→temps→param
vregs; replace with a jump to the re-entrant header; add back-edge; recompute_cfg). Gate at -O2 per
§5 + a tail-recursive benchmark. **Do option B first** (self-contained codegen move) as a SEPARATE
gated commit, THEN the AIR transform on top. Deferred to a dedicated session under non-Defender
build conditions. (This corrects an earlier note that checked only the frame prologue.)

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
