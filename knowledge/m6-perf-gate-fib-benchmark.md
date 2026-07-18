---
name: m6-perf-gate-fib-benchmark
description: "M6 milestone perf gate MEASURED 2026-07-17: AXIOM native Fib(42) is 2.44x slower than clang -O2 (2379ms vs 974ms) — the M6 gate (<=1.05x slower) is NOT met. The ELF-binary half of M6 shipped this session; the perf half is an open optimizer effort. Not a bounded tick task; needs prioritization."
metadata:
  node_type: memory
  type: project
  originSessionId: 3228306b-52d7-4378-bb1c-a0b6cef57eba
---

# M6 perf gate — Fib benchmark (measured 2026-07-17)

## ✅ 2026-07-18 (autopilot) — RFC 0026 P2 CONTROL-FLOW (multi-block) INLINER SHIPPED (`6b40dec`, B==C `4138AEB5`, 421/421)
`inline_multiblock_func` (ssa_opt.ax) inlines a small MULTI-block callee (if/else, while) — clones
the callee's blocks into the caller: split the caller block at the call, clone each callee block with
fresh ids + vreg offset in **reverse-post-order**, remap terminator targets, `return v` → `copy
call.dest=v'; jump cont`. Gated -O1 after single-block `inline_func`; scalar params/return +
whitelisted pure-scalar/control-flow body, no nested calls/memory/aggregates.
- **Groundwork `recompute_cfg(f)`** (air.ax): rebuilds block_succs/block_preds from each block's
  terminator, reproducing add_edge's EXACT ordering (succ [true=src2, false=dest]; preds
  ascending-source-index). The RFC called edge-ordering "THE TRAP" — PROVEN exact by wiring it
  UNCONDITIONALLY on every function and confirming a bit-identical B==C + 418/418 (a true no-op iff
  ordering-exact) BEFORE building the inliner. Reusable for any future CFG-rewriting pass.
- **THE bring-up bug (classic §4 hazard):** cloning callee blocks in INDEX order put a merge/exit
  block physically BEFORE its predecessors (e.g. `floorf`'s exit = block index 1), placing a value
  USE physically before its conditional redef → the physical-order value/regalloc passes miscompiled
  it **once inlined** (t_mathfill 127→126, t_fft 18→10 at -O1) **while B==C still held**. Standalone
  the same layout survives; inlined into a bigger fn it breaks. Fix = **clone in RPO** (`inl_rpo_dfs`
  DFS from callee entry, remap targets via RPO position map). LESSON reinforced: **B==C is necessary
  but NOT sufficient — the full USER-program regression is what caught this** (self-build never hit
  the float-fn shape). Always run the full regression, not just fixpoint.
- **Perf:** collatz UNCHANGED at -O1 (its inner while-loop dominates; call removal negligible, and
  strength-reduction is -O2-only). The clear P2 win is **call-dominated** multi-block code: new bench
  `cfhot` (200M calls to `classify()` w/ 2 ifs) = 2511→2205ms = **12% faster** with the inliner (main
  has 0 calls vs 1). Bench A/B via a temporary `if false` gate on the inliner call. **fib still needs
  accumulator-recursion→loop** (self-recursive → not inlinable) = the remaining P2 lever.
- Oracle `t_inlinecf` (exit 17: if/else `absdiff` + while `sumloop` inlined). Gate cmd unchanged.

## ✅ 2026-07-18 (autopilot) — RFC 0026 P1 INLINER SHIPPED (B==C, 407/407, hotloop 2.57x)
The pure-scalar single-block inliner is IMPLEMENTED, gated, and shipped in `ssa_opt.ax`
(`inline_func` + helpers, wired in `SsaOptimizer.run` at level≥1 before the per-fn passes).
**Gate GREEN:** B==C fixpoint bit-identical (`aae2ea1f…`), full regression **407/407** (incl.
compliance @ -O2/-O3), daily driver promoted. **Measured win** (`benchmarks/hotloop` = hot loop
calling `sq`): **185.5ms → 72.3ms = 2.57x faster** (17.5x→7.3x vs clang). fib (2.77x) / collatz
(3.21x) unchanged — correctly skipped (recursive / multi-block). Oracle `t_inline|exit|140`.
**TWO bugs found + fixed during bring-up (both classic AXIOM gotchas):**
1. `mut nn := cin` ALIASES the callee's original AirInst (aggregates are by-reference) → mutating
   the clone corrupted the CALLEE's body (dump showed a leaf `sq` with a dangling `ret %2`). Fix:
   build a FRESH `AirInst(...)` via constructor, never mutate an aliased copy.
2. `AirFunc.params.data[i]` holds the param's TYPE id, NOT its vreg (the param vreg is `i+1`,
   air_builder.ax:288-295). Original param→arg mapping compared vregs to type ids → args never
   wired, inlined body read garbage offsets. Fix: emit an explicit param-copy `base+(i+1) = arg`
   (type = `params.data[i]`) at the site, then offset ALL callee vregs by `base` uniformly (also
   isolates param mutation, so the no-mutation guard was dropped).
**NEXT (higher bench leverage, deferred):** inline callees WITH control flow (multi-block clone +
block-id remap) to catch collatz; accumulator-recursion→loop for fib. Both larger, same B==C gate.

### ✅ P1.5 SHIPPED 2026-07-18 — scalar-field getter inlining (B==C, 412/412, getter 2.89x)
Getter inlining is now IMPLEMENTED. `SsaOptimizer.run` + `inline_func` now take the `TypeTable`
(threaded from main_air.ax:1085). `OP_GET_FIELD` added to the whitelist, gated in `inl_is_inlinable`
to a **single-level SCALAR field read of a PARAM**: `src1` must be a param vreg (1..params.len),
struct type = `params.data[src1-1]`, and the field must satisfy `not field_is_aggregate and not
field_is_pointer_sum and field_size<=8` (via TypeTable). In the clone, `OP_GET_FIELD.src2` (field
INDEX) is kept verbatim (not a vreg); `src1` is offset-renamed. The inlined struct-ptr resolves its
type via the param-copy's `type_id`. **Gate GREEN:** B==C `f286cac9`, regression **412/412**, daily
driver promoted. **Win:** `benchmarks/getter` (hot loop calling `getx`) **185.6ms → 64.2ms = 2.89x**
(O0 no-inline vs O1 inline); hotloop/fib/collatz unchanged. Oracle `t_inline4|exit|42`. The scoping
note below is kept for history. Also covers COMPUTED multi-field accessors for free (GET_FIELD +
arithmetic already whitelisted: `area(r)=r.w*r.h`, oracle `t_inline5|exit|112`). Safety gate verified
by negative probe (str-field 16B + Option-field pointer-sum getters correctly SKIPPED, `t_inline6|exit|47`).
### ✅ OP_INDEX SHIPPED 2026-07-18 — scalar array-element getter inlining (B==C `343fa03b`, 418/418)
`fn at(a:[i64;N],i)=a[i]` now inlines. `OP_INDEX` added to the whitelist, gated in `inl_is_inlinable`
to a resolved SCALAR element: `type_id != 0` (element type is carried there, x86_selector.ax:542) and
`not type_is_aggregate(table, type_id)` and `type_size_and_align(...) <= 8`. The clone needed NO change
(OP_INDEX isn't GET_FIELD → both operands rename normally; type_id copied verbatim). Oracle
`t_inline8|exit|60`. **The getter family is now COMPLETE: pure-arith + scalar field + computed
multi-field + array-element.** (Array-value params + indexing confirmed working; earlier "language
gap" was a test int-literal-i32-vs-i64-param mismatch.)
**M6 gate-closer ROI map (2026-07-18, CORRECTED — supersedes an earlier overstated note):**
- **Div/mod-by-2^k strength reduction is ALREADY SHIPPED for i64/u64 incl. the SIGNED path**
  (`38a9d81`, [[perf-div-pow2-strength-reduction]]). collatz uses i64 → it ALREADY benefits
  (was the 10.5x→2.89x win). So this is NOT a remaining collatz lever. The only strength-reduction
  gap left is SUB-64-BIT signed (i8/i16/i32), gated out by `sel_type_is_8byte_int` (dirty-upper
  hazard) — niche, collatz doesn't need it.
- **collatz remaining gap (~3.2x):** loop quality + clang inlining `collatz_len` (which enables
  cross-boundary register allocation / tighter merged code — [[perf-div-pow2-strength-reduction]]
  line 49-50). So the **P2 control-flow inliner DOES help collatz** (not just call removal — it
  unlocks optimization of the merged loop). It's genuinely relevant, though the inner loop still
  bounds the total win — expect partial closure, not ≤1.05x alone.
- **fib remaining gap (~2.6x):** ACCUMULATOR-RECURSION→LOOP (fib self-recurses → the linear inliner
  skips it; needs the recurrence transform). This is the fib-specific closer.
**Ranked next M6 levers:** (1) P2 control-flow inliner (helps collatz + general call-heavy code; full
plan in RFC 0026 §2a′); (2) fib accumulator→loop; (3) sub-64-bit signed div/mod strength reduction
(niche). All independent, hard, dedicated B==C sessions.

**(historical) P1.5 candidate — getter inlining (`fn read_x(p)=return p.x`), scoped 2026-07-18 (do NOT rush):**
adding `OP_GET_FIELD` to the whitelist is MORE than one opcode — verified constraints:
- `OP_GET_FIELD.src1` = struct-ptr VREG (rename it); `.src2` = FIELD INDEX, NOT a vreg — must be
  EXCLUDED from renaming (like ICONST/FCONST), else `field_offset(type,src2)` reads a bogus index.
- Correctness depends on `get_register_type(src1)` returning the right struct type. When inlined,
  src1 = the param-copy target whose `type_id` = `params.data[i]` (the struct type) — so the type
  resolves IF the param-copy carries the struct type id (it does). Verify this holds.
- The selector's GET_FIELD has size==16 / aggregate / pointer-sum(Option/Result/sum-field) branches
  (x86_selector.ax:2099-2114). To stay safe, restrict getter inlining to SCALAR fields (size≤8, not
  aggregate, not pointer-sum) in the first cut; aggregate/16B/opt fields are a further increment.
- Same for `OP_INDEX`/`OP_SET_FIELD`/`OP_FIELD_ADDR` if ever added — check which operand is a vreg.
This is a real aggregate/memory increment (P1 was pure-scalar precisely to avoid this), so it needs
its own careful pass + oracles (scalar getter, nested, wrong-type guard) + full B==C gate. NOT a
tail-of-session change.
- ⚠️ **ARCHITECTURAL PREREQ (found 2026-07-18):** gating to scalar fields needs `field_size`/
  `field_is_aggregate`/`field_is_pointer_sum`, which require the `TypeTable` — but `SsaOptimizer.run(self, m)`
  only receives the `AirModule`, NOT the TypeTable. So getter-inlining first needs the TypeTable threaded
  into the optimizer (signature change to `run` + its call site in main_air.ax ~L1085). Without it the
  inliner can't tell a scalar field from a 16B/aggregate/pointer-sum one → would miscompile. Do the
  plumbing as step 0 of the getter-inlining increment.

## 2026-07-18 (autopilot) — INLINER IMPLEMENTATION SPEC + cost/benefit gate (staged, not shipped)
Fully designed the flat-AIR inliner (read air.ax/ssa_opt.ax/x86_selector.ax). **Key finding that
gates it:** the value passes (copy_prop/cse/dce, `ssa_opt.ax`) scan `f.insts` in PHYSICAL order =
execution order, so inlining CANNOT append — it needs a per-block **rebuild** of `f.insts` +
`f.block_instrs` (blocks index inst-indices via `block_instrs`, `x86_selector.ax:2514`; extras hold
VREGS not inst-indices, so reindexing is safe). Enumerated hazards + the safe design:
- **Rename by offset**: `rename(v)= v==0?0 : base+v`, `base = max_reg_id(caller)+1`, advance by
  `max_reg_id(callee)+1` per site. Params map DIRECTLY to caller arg vregs
  (`caller.extras[arg_start+1+i]`, arg_start=call.src2) — NO param-copy — but ONLY if the callee
  never mutates a param (guard: no inst `dest==params[i]`), else the arg vreg is clobbered.
- **Return→copy**: `OP_RETURN v` → `OP_COPY(dest=call.dest, src1=rename(v), type_id=callee.ret_type)`
  (type known-correct; the one synthesized inst).
- **ICONST/FCONST src1/src2 are immediates, NOT vregs** — don't rename. To avoid ALL other operand-
  kind/type/memory hazards, WHITELIST the callee body to pure scalar ops (ICONST/FCONST/0x02xx
  arithmetic/OP_COPY/OP_MOVE/OP_RETURN); bail on OP_LOAD/STORE/GET_FIELD/INDEX/CALL/branches/globals.
- **Guards**: callee non-extern, non-async, single block, whitelist-only body, no param mutation,
  scalar ret, ≤~12 insts, non-recursive; call site direct (src1==0), resolvable sym, argc==paramc;
  caller `blocks_are_contiguous` (block order == physical order) else bail.
- **Insertion point**: new `inline_module(m)` called in `SsaOptimizer.run` BEFORE the per-fn loop
  (`ssa_opt.ax:1798`); downstream fold/copy_prop/dce clean up.

⚠️ **COST/BENEFIT — why NOT shipped this pass**: the SAFE (whitelist) form is **perf-NEUTRAL on the
benchmarks** — fib is self-recursive (skipped), collatz_len is multi-block/loop (skipped), so neither
improves; it would FAIL RFC 0026's "revert if no ratio gain" criterion. Moving fib/collatz needs the
RISKY forms: (a) inline callees WITH control flow (multi-block clone + block-id remap + CFG rebuild —
much larger), or (b) **accumulator-recursion→loop** (fib's shape: `return n*fib(n-1)` — needs an
accumulator-introduction transform, clang does it, high risk). Both are multi-session backend efforts
requiring the full B==C + -O2 regression + bench gate with revert-on-red. **Next dedicated session:**
implement (b) accumulator→loop for fib specifically (highest bench leverage) OR the control-flow
inliner, using this spec. Do NOT ship the pure-arithmetic-only inliner alone (perf-neutral churn).

## 2026-07-18 (autopilot) — committed benchmark harness + RFC 0026 (inlining + recursion→loop)
User prioritized M6 perf ([[autopilot-direction-2026-07-18]]). Shipped the **reusable, committed
perf harness** (was previously ad-hoc): `scripts/bench_perf.sh` builds `benchmarks/{fib,collatz}.ax`
with AXIOM -O1 + `.c` refs with clang -O2, times best-of-N, prints ratio. **Re-measured baseline on
current daily driver `22DA5200`:** fib(42) **2.72x** (2442ms vs 899ms), collatz **3.19x** (170ms vs
53ms) — both exit-correct vs clang (40, 161). (2.44x→2.72x vs the 2026-07-17 number = measurement
variance + driver drift, same ballpark.) **RFC 0026** (`rfcs/0026-inlining-and-recursion-to-loop.md`)
designs the two chosen levers: (1) module-level **inliner** of small single-block non-recursive
callees inserted before the per-fn loop at `ssa_opt.ax:1798` (vreg-rename + param→arg bind +
return→copy in flat AIR); (2) **tail-self-recursion→loop** (accumulator form like fib is P2/deferred —
fib's ratio is closed mainly by inlining). GATE = B==C + 401/401 regression + -O2 guard + bench delta
+ new t_inline*/t_selfrec* oracles; revert-on-red, record blocker here. Harness cmd: `bash scripts/bench_perf.sh`.

`docs/tasks/milestones.md` **M6 (Native x86-64)** gate = "ELF binary without GCC, **Fib(40)
≤ 5% slower than clang -O2**". This session shipped the ELF-binary half (`--target linux`);
this note records the perf half.

## Measurement (Windows, best-of-3 wall-clock via Python perf_counter)
| | Fib(42) | ratio |
|---|---|---|
| AXIOM native `-O1` | **2379 ms** | **2.44x** |
| clang `-O2` | 974 ms | 1.00x (ref) |

Both correct (fib(42)=267914296; AXIOM exit 40 = 296 & 0xFF, the 8-bit-truncated `%1000`).
**M6 perf gate (≤1.05x) is NOT met — AXIOM is ~2.44x slower on recursive fib.**

## Why (from clang's disasm)
clang -O2 aggressively transforms `fib`: it turns the double-recursion into a **loop** — only
ONE `call fib` per activation inside a `ja fib+0x20` loop (accumulating the fib(n-2) chain
iteratively), plus tight register use (edi/esi, no stack traffic in the hot path). AXIOM emits
straightforward recursive codegen (two real calls + prologue/epilogue per activation). fib is
call-dominated, so the gap is mostly **call/frame overhead + missed recursion→loop + regalloc**.

## Finding 2026-07-17 (commit `34ec05f`): rsp churn is NOT the bottleneck
First concrete M6 attempt shipped: **dropped the redundant per-call win64 shadow-space
`sub rsp,0x20`/`add rsp,0x20`** (`compute_frame` already reserves 32B of outgoing shadow at
the bottom of every win64 frame — the per-call adjust was a DOUBLE reservation for <=4-arg
calls; skip when `shadow_size<=32`, win64-only, SysV untouched). Removed 2 instrs/call site incl.
BOTH fib recursive calls. **Result: code size 2330→2188 KB (-6.1%) but runtime FLAT (+0.1% on
fib(42), within noise).** Lesson: the CPU **stack engine** renames rsp adjustments for ~free, so
micro-removing `sub/add rsp` is a code-size/cleanliness win, NOT a hot-path speedup. The fib gap
is register-shuffle overhead (redundant `mov rbx,rcx; mov rsi,rbx` copy chains → forces 4
callee-saved pushes) + call/frame mechanics — **future M6 effort must target regalloc COALESCING
(item #1), not instruction micro-removal.** Gate was clean: B==C `D6AF9DC7`, 340/340, ELF 10/10.

## Investigation 2026-07-17: next lever = register-pressure reduction (imm folding + copy-prop)
Traced fib's redundant `mov rax,2; mov rcx,rax; cmp rsi,rcx` (and `mov rbx,rcx; mov rsi,rbx`
param copy chains). Root: constants/params are materialized into vregs and compared/used
register-to-register; `select_comparison` (x86_selector.ax:764) always emits `CMP vreg,vreg`,
never `CMP reg,imm`. The extra vregs raise register pressure → forces the 4 callee-saved pushes
in fib's prologue. **The real M6 win is cutting register pressure** (fewer live vregs → fewer
spills/callee-saves → genuinely faster, unlike the shadow-sub which the stack engine made free).
Two angles: (a) **immediate-operand folding** — `CMP/ADD/SUB reg,imm32` when an operand is a
constant (emitter already supports OPND_IMM in MACH_CMP, e.g. x86_selector.ax:1285); (b)
**copy-propagation/coalescing** of the vreg copy chains.
**CORRECTNESS HAZARD (why this is NOT a quick tick fix):** `2 as i64` lowers to
`CAST(ICONST 2)`, so folding must chase the const THROUGH the cast — but `300 as u8` is
`CAST(ICONST 300)` whose true value is 44 (truncated). Folding the raw ICONST there silently
miscompiles, and the B==C self-host gate may NOT catch it (compiler may never hit that pattern
while user code does). Safe folding needs type-aware cast handling: only chase non-narrowing
casts, or verify the ICONST value is unchanged in the cast's dest type. `const_shift_amount`
(x86_selector.ax:902) sidesteps this by requiring v in [0,64); a general folder can't.
→ Needs a deliberate, type-checked design pass + user prioritization; do NOT bolt on blind.

### ATTEMPT 2026-07-17 (immediate-compare folding) — REVERTED, gate caught a miscompile
Implemented `const_cmp_imm32` (chase RHS to non-neg ICONST <2^31 through copies + casts with
dest>=4 bytes, skip narrowing) and emitted `CMP reg,imm32` in `select_comparison`. **Correct for
fib + 3 oracles when compiled by A** (fib freed r12 — prologue 5→4 pushes, real pressure drop, the
`cmp rdx,0x2` immediate form confirmed). BUT **B (compiler built BY the folding codegen) was
BROKEN**: it rejected the compiler source with 135 spurious "operator '+/&/*' not defined for
Option/Result" errors — the TYPECHECKER's `tc_is_opt_res`/kind comparisons got mis-evaluated. B!=C,
reverted cleanly (back to `D6AF9DC7`). **Root cause (unconfirmed, prime suspect): `CMP reg,imm`
differs from `CMP reg,reg` for SUB-64-BIT operands with dirty upper bits and/or a SPILLED first
operand** — fib (small, 64-bit i64, no spills) worked; the compiler's kind/tag compares (u8/u16/u32
values, spilled in huge functions) broke. LESSONS for the next attempt: (1) verify the emitter's
MACH_CMP-with-OPND_IMM encoding for a MEMORY (spilled) first operand AND that it emits the correct
operand WIDTH (a 64-bit `cmp` of a u8/u16 reg with dirty upper bits is wrong); (2) restrict the fold
to comparisons whose operand type is provably 64-bit / clean, OR add a width to the MACH_CMP; (3)
ALWAYS run the full B==C fixpoint — A-only correctness (oracles+fib) passed while B was broken. The
gate worked exactly as designed. This is why immediate-folding is "needs careful backend work", not
a tick task.

## SHIPPED 2026-07-17 (session): 2 broad codegen wins + biggest lever found
Two safe, broadly-applicable optimizations shipped (each B==C + 341/341):
- **`a4784ee` compare-branch fusion** — `CMP;SETCC;MOVZX;TEST;JCC-NE` -> `CMP;JCC cc` when the bool
  is a throwaway (contiguous window, single-use). Removes 3 hot-path insts + 1 vreg per conditional
  (ALL `if <cmp>`/`while <cmp>`). fib 3.02x -> 2.57x. Peephole `fuse_compare_branch` in select_all.
- **`50796cc` move-coalescing** (biased graph coloring + copy-edge elision) — coalesces move-related
  vregs so redundant `mov r,r` self-annihilate; drops the interference edge for a copy pair that only
  touches at the copy point. fib 2.57x -> 2.40x. `move_partner[]` in x86_regalloc.ax.
- **is_param liveness shorten — TRIED+REVERTED**: the `MOV vreg,argreg` snapshot is force-extended to
  end-of-function (pre-CFG-liveness hack); shortening it coalesces the param copy (fib 4->3 pushes) BUT
  broke t_optresult (the hack protects a real param-liveness case). Also: removing ONE push is ~free
  (stack engine), didn't move fib wall-clock. Reverted.

**BIGGEST lever found = div/mod-by-pow2 strength reduction** (idiv->shift): a FAIR benchmark
(collatz, clang can't constant-fold) exposed a **10.5x** gap, reducible to **3.0x**. Correct for all
user code; signed path breaks self-host — full diagnosis + partial fixes in
[[perf-div-pow2-strength-reduction]]. This, not fib micro-opts, is the high-value arithmetic gap.

LESSON reinforced: fib's residual ~2.4x is call-count (clang recursion->loop, 1 call/activation vs 2)
— RFC-scale, narrow. loopsum/fib are POOR benchmarks (clang closed-forms loopsum, transforms fib);
use div/branch/pointer-chasing benchmarks clang can't exotic-transform to measure real codegen.

## Scope / next steps (NOT a tick task)
Closing a 2.44x gap to clang is a **large, open-ended optimizer effort**, and CLAUDE.md §10
("DO NOT prematurely optimize"; correctness → stability → profiling → optimization; work must be
benchmarked/measurable/reversible/isolated) says to prioritize it deliberately, not grind blind.
Candidate high-leverage items (benefit ALL code, not just fib), roughly ordered:
1. **Register allocation quality** — reduce spill/reload + redundant `mov`s in the hot path
   (compare AXIOM's fib body to clang's ~6-instruction loop). Biggest general win.
2. **Leaf/small-function prologue-epilogue trimming** — avoid unneeded frame setup / callee-save
   pushes when a function doesn't need them (clang's fib pushes only rsi/rdi it actually uses).
3. **Inlining** small functions (none today?) — removes call overhead broadly.
4. **Tail-call / self-recursion → loop** — matches clang's fib transform; higher effort, narrower.
Each needs: a benchmark harness (this fib + a broader suite), an isolated pass, and A==B/B==C +
regression after. Recommend a dedicated perf session with the user prioritizing (perf vs. other
milestones). Bench recipe: build fib(42) `-O1`, time best-of-3 vs `clang -O2` (clang is in
msys2 `/c/msys64/ucrt64/bin`). NB PowerShell `&` needs `.\`/abs paths; `bc` absent (use python).
