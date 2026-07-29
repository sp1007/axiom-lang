# RFC 0036 — M6-opt: self-tail-recursion to loop (and why the accumulator transform is NOT proposed)

> # ⛔⛔ WITHDRAWN — **THE PROPOSED TRANSFORM IS ALREADY IMPLEMENTED**
>
> Withdrawn within the hour, by the pricing experiment §6 demanded. AXIOM **already** converts a
> self-tail-call into a jump to the entry block. Emitted `sumto` (driver `42F49C73`, `-O3`)
> contains **no `call` at all**:
> ```
> push %rbp ; mov %rsp,%rbp ; push %rbx ; sub $0x28,%rsp   <- prologue ONCE
> cmp  %rdx,%rax ; je base
> lea  -0x1(%rax),%rdx          <- n-1
> mov  %rcx,%rbx ; add %rax,%rbx <- acc+n
> mov  %rdx,%rax ; mov %rbx,%rcx <- reassign the parameters
> jmp  <entry>                   <- JUMP BACK, no call
> ```
> That is §3's design, already shipped. The suite even names it: **`t_selfrec`, `t_selfrec2`** —
> which I should have read *before* writing this RFC. Same failure recorded repeatedly in
> `knowledge/`: **reconcile a proposed item against what has already shipped before designing it.**
> The pricing gave it away — AXIOM measured **0.355x** of a NASM floor written to model "the shape
> AXIOM emits", i.e. nearly 3x FASTER than its own supposed shape, which is only possible if the
> shape assumption was wrong.
>
> ### What the experiment DID establish (the useful residue)
> On the new tail-recursive shape, 20M calls, paired 3 alternating rounds:
>
> | variant | ms | |
> |---|---|---|
> | NASM real-recursion (what I wrongly assumed AXIOM emits) | 63.7 | not relevant |
> | **AXIOM as it compiles today** | **22.6** | already loop-form |
> | **NASM loop floor** (ideal transform output) | **16.2** | |
>
> ⇒ **AXIOM is 1.40x of the loop floor on a tail-recursive shape.** The gap is NOT a missing opt
> pass — it is the parameter shuffle (`mov %rdx,%rax ; mov %rbx,%rcx`) plus a prologue the
> hand-written form does not need. That is **M6-codegen** territory, in the same family as the
> copy folds shipped this session, and it is a *newly measured* gap on a shape the suite never had.
>
> ⇒ **Real next step for M6-opt**: not this transform. Either (a) close the parameter-shuffle gap
> on tail-recursive functions (codegen, cheap, now measurable), or (b) revisit the accumulator
> transform for genuinely non-tail recursion like `fib` — with §2's objections still standing.
>
> The design below is kept **only** as the record of what was proposed and why it was wrong to
> propose. Do not implement it.

- Status: **WITHDRAWN 2026-07-30 — already implemented; see the banner above**
- Author: autopilot, 2026-07-30
- Approved in advance by the user (standing decision D4, `knowledge/user-decisions-2026-07-29.md`):
  self-written + self-approved RFCs are authorized; the fixpoint gate still binds.
- Related: `knowledge/m6-perf-baseline.md`, `knowledge/session-handoff-2026-07-30a.md`
  (D1 split M6 into **M6-codegen**, now met, and **M6-opt**, this RFC), RFC 0034 (IR pipeline).

## 1. Motivation, and the number that defines the ceiling

D1 split the M6 milestone in two because measurement forced it. On `fib`, clang is ~1.5x faster
than a **hand-written NASM floor in the shape AXIOM emits**, and the reason is not codegen at all:
`objdump` shows clang turns the second recursive call into a loop (accumulator transform), so it
executes roughly **half the calls**. A perfect backend still lands ~1.55x off clang.

M6-codegen (≤15% of the same-shape floor) is now met on all four shapes, with the floors
themselves validated rather than assumed. Everything left on `fib` is **algorithmic**, which is
exactly what M6-opt was carved out to address.

## 2. Scope: self-tail-recursion ONLY

**Proposed:** a direct self-tail-call becomes a jump to the function's entry block, reusing the
frame. `f(...)` in tail position where the callee is `f` itself.

**NOT proposed — and this is the substance of the RFC.** The transform clang applies to `fib` is
the **accumulator transform**: rewriting `fib(n-1) + fib(n-2)` so one arm becomes iterative. That
requires proving the combining operator associative, synthesizing an accumulator parameter, and
reassociating a recurrence. Three reasons to decline it now:

1. **It does not fit AXIOM's IR contract yet.** Reassociating `+` over i64 is legal (wrapping
   two's complement is associative), but the same pass over `f32`/`f64` is **not** — float
   addition is not associative, and this project has already been bitten by a rule applied on one
   binop branch and forgotten on its sibling ([[bug-f32-compare-float-literal]],
   [[bug-negative-literal-compare-o0]]). A transform whose correctness depends on the operand
   type needs the type discipline designed first, not bolted on.
2. **`fib` is the only shape that would benefit.** Of the four benchmark shapes, three have no
   recursion at all. Building a recurrence-reassociation pass to move one benchmark is the
   "optimize for the benchmark" trap §10 warns about.
3. **Self-tail-recursion is a strict prerequisite anyway.** The accumulator form is only useful
   once the resulting self-tail-call is turned into a loop. Shipping the simpler transform first
   makes the harder one measurable in isolation, which is the only way this project has
   successfully priced anything (see §6).

## 3. Design

At the AIR level, after inlining and before SSA optimization:

- **Detect.** An `OP_CALL` whose callee symbol is the enclosing function, whose result feeds
  directly into `OP_RETURN` with no intervening instruction that observes it, and which is not
  inside a `defer` or destroy region.
- **Transform.** Assign the argument registers to the parameter registers, then `OP_JUMP` to the
  entry block. No frame teardown, no `call`, no `ret`.
- **Ordering hazard.** Arguments must be evaluated into temporaries **before** any parameter is
  overwritten, or `f(b, a)` corrupts itself. This is the same shape as the copy-pair problem
  peephole 1c solves at the machine level; here it must be handled in AIR because parameter
  registers are the destination.

## 4. Interaction with what already ships

- **CTGC / `defer`**: a tail call is only a tail call if nothing runs after it. `flush_defers`
  emits work at `OP_RETURN`, so a function with pending defers or CTGC destroys at the return
  point **must be excluded** — the frame is still live. Gate on the same predicate
  `lower_return` uses.
- **Drop glue (RFC 0014)**: same exclusion; a value needing drop keeps the frame live.
- **Debug info**: collapsing frames changes stack traces. Acceptable for `-O1`+; the transform
  must be **off at `-O0`**, which also keeps the new `-O0` regression sweep meaningful.

## 5. Drawbacks

- Stack traces lose recursion depth (mitigated: `-O0` unaffected).
- A new AIR transform is new surface in the self-host-critical path. The fixpoint gate covers
  correctness, but this is the component §10 says not to trade maintainability for.
- **It may buy nothing.** `fib` as written is not tail-recursive, so this RFC alone does **not**
  speed up any current benchmark shape. That is stated plainly rather than buried: the win is
  enabling, and the enabling step must be measured before the enabled one is built.

## 6. Measurement plan — mandatory, and shaped by this session's failures

Five separate perf conclusions were overturned today by measurement. The rules that produced the
corrections are non-negotiable here:

- **Price it before building it**, with a hand-written NASM variant of a *tail-recursive* shape —
  not `fib`. Pricing on the wrong shape produced "confident zeros" twice
  (immediate-folding, register coalescing).
- **A new benchmark shape is required**: none of the four existing ones is tail-recursive, so the
  suite currently **cannot** measure this. Add a tail-recursive shape with its own NASM floor
  BEFORE writing the pass.
- **Paired, alternating, ≥2 rounds**, comparing the DELTA between two builds. Function alignment
  is shipped, so placement noise is bounded, but a single run remains untrustworthy.
- **Prove the pass FIRES** by dumping the instruction stream, not by reading the emitted binary.
  Peephole 1d passed every gate while matching nothing.

## 7. Gate

AIR-level transform affecting codegen ⇒ `A != B` is expected and **`B == C` is mandatory** before
commit, plus full regression at `-O1` **and** the `-O0` sweep, plus ELF / ctgc / exe_size /
lib_collision / so_export. Oracles: a tail-recursive function at a depth that would overflow the
stack without the transform (proves it fired), one with a pending `defer` (proves the exclusion
holds), and one at `-O0` (proves it is off there).

## 8. Alternatives considered

- **Do nothing.** M6-codegen is met; M6-opt is a separate milestone with no deadline. Legitimate,
  and preferable to shipping the accumulator transform speculatively.
- **General tail-call optimization** (mutual/indirect tail calls). Larger, needs ABI thought, and
  self-recursion is the case that actually appears in this codebase.
- **Accumulator transform first.** Rejected in §2.

## 9. Recommendation

Adopt §2's narrow scope. **Do not start with code** — start with §6's new tail-recursive benchmark
shape and its NASM floor. If that pricing shows the transform is worth less than ~5% on a shape
built to favour it, this RFC should be closed unimplemented, and that outcome is a success of the
method rather than a failure of the work.
