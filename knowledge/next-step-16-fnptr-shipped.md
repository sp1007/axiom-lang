---
name: next-step-16-fnptr-shipped
description: "BUG#49 function pointers shipped + fixpoint-verified; std.numerical"
metadata: 
  node_type: memory
  type: project
  originSessionId: ed12f2e7-f4ab-41c0-9bb3-940acfa7aaec
---

**2026-07-03 — Function pointers (BUG#49) SHIPPED + fixpoint-verified.** Bare
function values (`let f = add`, `apply(add, …)`, `fn(f64)->f64` params) now work
O0+O1 through native self-link.

Commits on main: `964fcba` (fn-ptr) + `3493c11` (std.numerical generic).
Self-host fixpoint converged: stage3==stage4 = `19276a2c…` (stage2=`4bc1b9c9`,
normal — stage2 built by gcc-stage1, converges from stage3). Full detail =
BUG#49 in [[knowledge/bugs.md]].

Mechanism (6 files): air.ax OP_FUNC_ADDR(0x030D, src1=symbol-index IMMEDIATE);
air_builder lower_ident SYM_FUNC; cgen; x86_selector (MACH_MOV_IMM vreg=3 +
capture target→R11 before arg marshaling + MACH_CALL_INDIRECT); x86_emitter
(call r/m + **RIP-relative lea/RELOC_PC32 — ASLR keystone**, absolute movabs
breaks under DYNAMICBASE). Two -O1 optimizer fixes in ssa_opt: copy-prop guard
OP_FUNC_ADDR (src1 is immediate, not vreg — like BUG#15) + DCE count
indirect-call target (else func_addr deleted as dead → crash 139).

**Coherent fn-ptr stdlib layer shipped (all oracle O0+O1, library-only):**
- std.numerical #25: nm_integrate/simpson/diff/bisection/newton over fn(f64)->f64.
- std.sort: insertion sort_i64/f64, comparator quicksort sort_quick_i64/f64
  (recursive fn-ptr), binary_search_i64, find_i64/all_i64.
- std.functional: map_i64/f64, filter_i64, reduce_i64/f64, count_i64.
Session commits (main): 964fcba fn-ptr · 3493c11 numerical · 00fc256 sort ·
a163fc4 bugs-fix · 72ee361 cleanup · 2203bea+0cd586d RFC0008(+P1 blueprint) ·
0161c74 qsort · 8a02a25 f64-qsort+bsearch · 464cf98 functional.

CORRECTED 2026-07-03: the "no-import program → STATUS_ENTRYPOINT_NOT_FOUND
(0xC0000139)" symptom was NOT a real no-import compiler bug. It only appeared
while the box was severely loaded (trivial self-link builds taking 84-112s from
browser/thrash pressure) — a build-under-resource-starvation artifact (malformed
PE from the self-linker under memory pressure). On the healthy box, no-import
programs (plain `return 5`, direct call, local fn-ptr) all return correctly at
O0+O1 via bash/cmd/PowerShell. Do NOT chase a "no-import" bug — check system load
first if you see STATUS_ENTRYPOINT_NOT_FOUND. (Robustness follow-up MAYBE worth
it: make the self-linker fail loudly instead of emitting a bad PE if an alloc/
write fails under pressure.)
Still-valid trap notes: git-bash `$?` is 8-bit (0xC0000139&0xFF=57), use
PowerShell $LASTEXITCODE; never edit a bash script while it is running.

**Float keystones CONFIRMED FIXED 2026-07-03** (correct `Name(field: v)` syntax,
extern called, isolated build; O0 and O1). All three fixed by earlier commits
(476b352 float-vreg-GPR-pool + BUG#44 XMM callee-save); bugs.md was just stale.
Records corrected in commit `a163fc4` (BUG#46/#48 → ✅).
- BUG#45: `mk()->V4` 4×f64 (32B) by-value return → 10 ✓ (docs were correct).
- BUG#46: VERBATIM original st_variance (separate st_mean call, read m across
  variance loop) over [2,4,4,4,5,5,7,9] → variance 4 ✓.
- BUG#48: 10 f64 `sq(c_i)` live at once (forces >2 spills past 8 XMM) → sum 110 ✓
  (8-live → 204 ✓). Original verbatim 254/255 file not relocated, but mechanism
  stress-verified.
**No open backend keystone blocks work.** The float infra is healthy end-to-end.

**Next language keystone drafted:** RFC 0008 (closures with capture), commit
`2203bea` — follows BUG#49. MVP = separate closure type {code_ptr, env_ptr},
capture-by-value POD-only, move-only (sidesteps the [[bignum-ctgc-conflict]]
owned-field drop-glue gap); reuses BUG#49 indirect-call path + implicit env arg 0.
Implementation is P1 (parser lambda) → P2 (lower) → P3 (fixpoint). Not started.

**Repo cleanup done** (commit `72ee361`): purged ~1GB of committed/untracked build
junk (logs, dumps, obj/asm, 478 scratch .exe, 61 scratch bin/*.ax); kept toolchain
bin/axc.exe + bin/axc_stage1.exe; widened .gitignore. Working tree 1.1G → 113M.

**Open observation (needs a HEALTHY machine to diagnose):** `import std.sort`
(and a no-op `import std.sort`) failed to run cleanly — the fn-ptr call version
segfaulted (139) and the no-op version didn't finish codegen in 300s. BUT the
whole math library is tested ONLY via self-contained mirror repros, never via
`import` (std.math/fft/quaternion/numerical all follow this) — so direct import
of these free-function math modules is an untested/unsupported path in general,
NOT a defect introduced by std.sort/std.numerical (both are validated per the
same mirror convention, O0+O1). Likely a general import-bundler item. Revisit on
a healthy box: first check whether importing a PRE-EXISTING math module (e.g.
std.math) also fails — if so it's pre-existing, not tonight's code.

**Environment note:** builds were ~10-20× slow this session (trivial self-link
`return 5` took 84s vs normal ~5-10s) due to heavy background load (browsers +
a stale mathver.exe from a 2026-06-29 session). Every 2-min timeout tonight was
this, not a compiler hang. Run builds sequentially + isolated; expect slowness
until the box is quiet.

**Test-harness lessons (cost me a false alarm this session):**
- Struct literal is `Name(field: v, …)` (PARENS), NOT `Name{…}` (braces) — braces
  parse wrong → garbage codegen → segfault 139, easily mistaken for a real bug.
- A minimal program that declares but never CALLS an extern can still segfault at
  load (empty/dead import). Always CALL an extern (e.g. putchar) in float/struct
  probes, and build ISOLATED (contention makes builds time out / look hung).
See [[next-step-15-selfhost-status]].
