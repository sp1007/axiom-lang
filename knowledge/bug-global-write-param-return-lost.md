---
name: bug-global-write-param-return-lost
description: "OPEN (high-value, pre-existing) — a function with (a param) AND (a non-void return) AND (a module-global write) LOSES/corrupts the global write. Remove ANY one of the three → correct. O0-independent (not DCE). Found 2026-07-24 while trying to ship RFC 0033 Phase C threads (a thread entry is fn(ptr[void])->u32 that writes shared state → hits this). Blocks real OS threads. Needs a dedicated backend session (selector/regalloc of OP_STORE→OP_GLOBAL_ADDR with a live param + return)."
metadata:
  node_type: memory
  type: project
---

**OPEN, HIGH-VALUE, PRE-EXISTING silent miscompile.** Found 2026-07-24 while implementing
RFC 0033 Phase C (OS threads). NOT caused by the concurrency work (that touches parser/
typecheck/OP_AWAIT/linker-names only; this is OP_STORE/OP_GLOBAL_ADDR codegen). Repro banked at
`bin/known_fail_global_param_return.ax` (kept OUT of regression until fixed).

## Minimal trigger (the TRIPLE)
A function that simultaneously has all three loses its module-global write:
1. at least one **parameter**, AND
2. a **non-void return** value, AND
3. a **module-global write** (`g = ...` where `g` is a top-level `mut`).

Remove ANY one → correct. Proven by minimal programs (all `-self-link`, both O0 and O1):
- `fn f(): g=42`                    (no param, void)      → 42 ✓  (T10)
- `fn f(arg: ptr[void]): g=42`      (param, VOID return)  → 42 ✓  (T20)
- `fn f() -> u32: g=42; return 0`   (NO param, return)    → 42 ✓  (T21)
- `fn f(n: i64) -> i64: g=42; return 0`  (param+return+global) → **0 ✗**  (T27)
- `fn f(n: i64) -> u32: g=42; return 0`  (same, narrow ret)    → **0 ✗**  (T26)
- `fn f(n:i64)->u32: g=99; return n` then `g+r`               → **8 ✗** (T28, corruption not just 0)
Return width (i64 vs u32) does NOT matter — T27 (wide) also fails. So it is NOT narrow-return.

## Not DCE
Fails identically at **O0** (T22), where there is no dead-code elimination. OP_STORE is in
`has_side_effect` (ssa_opt.ax:41) anyway. So the store is not being optimized away — it is
mis-selected / mis-allocated.

## AIR evidence (`dump-air -O1` of the failing `worker`)
```
fn @3(t4) -> t4:
  block_0:
    %1: t4 = copy %1          ; param
    %2: t4 = iconst %42
    %3 = cast %2
    %4: t4 = globaddr %43     ; &g
    %4: t4 = store %3         ; store %3 to [%4] — dest REUSES the globaddr vreg %4
    %5: t4 = iconst
    %6 = cast %5
    ret %6
```
The store reuses `%4` (the globaddr result) as its own dest vreg. Suspicion: in regalloc/emit,
with a param (%1) and the return value (%6) also live, the register holding the global address
(%4) is clobbered before/at the store, so the store writes to the wrong location (or the value/
address regs get swapped → the T28 "8" corruption). The store's address vs value operand field
mapping (dest=address, src1=value) under param pressure is the place to look:
x86_selector.ax OP_STORE (:2051) + OP_GLOBAL_ADDR (:1861) + emit_param_prologue (:2348) + regalloc.

## Related fn-pointer crashes (same region, likely same root)
- `let f = worker; f(null as ptr[void])` where worker is `ptr[void]->u32` → **SIGSEGV** (T8).
  But `let f = wf; f(&x)` (real addr) and `apply(wf)` (fn-typed param) both work. And a thread
  spawned via `CreateThread(...,worker,...)` with an EMPTY worker also SIGSEGVs (T3) — the OS
  calls the passed address and faults, suggesting either the address handed to CreateThread is
  wrong or the thread-entry ABI differs. These may or may not be the same bug as the store issue;
  investigate together.

## Impact on RFC 0033 Phase C (threads)
A Windows thread entry is exactly `fn(ptr[void]) -> u32` that writes shared state — it hits both
the store-triple bug and the fn-ptr/thread-entry crash. So **real OS threads are BLOCKED** on
this. Phase A (async/await) + Phase B (timer) shipped fine; Phase C deferred to a dedicated
backend session that fixes this miscompile first. See [[rfc0033-concurrency-v1]].

## Severity
HIGH for a silent miscompile: `fn accumulate(x: i64) -> i64: total = total + x; return total`
(param + return + global) is an ordinary pattern. It escapes regression because RFC 0017 module
globals (recent) aren't combined with param+return functions in any existing oracle. Add oracles
covering global-write × {param} × {return} once fixed.
