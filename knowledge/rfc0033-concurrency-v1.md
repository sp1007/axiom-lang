---
name: rfc0033-concurrency-v1
description: "RFC 0033 concurrency v1 (async/await/spawn/timer/thread). Phase A SHIPPED 81e6f29 — async/await v1 correctness: await produced NO code (OP_AWAIT no selector case) → returned 0/garbage on every platform; top-level async fn didn't parse. Fixed parser+typecheck+selector under synchronous-cooperative v1 semantics. A==B==C 91AE46F2, 535/535, oracle t_awaitv1=42. Phases B (timer) / C (threads) pending."
metadata:
  node_type: memory
  type: project
---

**ĐỌC ĐẦU TIÊN cho concurrency.** User: "dứt điểm Async/spawn-await cross-platform" +
"timer too" + "threads too". Full concurrency toolkit, cross-platform. Done as RFC 0033
(`rfcs/0033-concurrency-model-v1.md`) with phased delivery.

## Ground truth (investigated 2026-07-24)
Two ORTHOGONAL axes, conflating them is what broke things:
1. **async/await** = value/coroutine axis. Was BROKEN on every platform.
2. **spawn/scheduler** = actor/task axis. WORKS (single-threaded cooperative).

The old "finish async = build preemptive runtime" framing is RFC-scale. RFC 0033 instead
defines a coherent **v1 = synchronous cooperative** that is correct + deterministic +
identical on Win COFF / Linux ELF, with a forward-compatible path to preemptive v2.

### Confirmed defects (before fix)
- `await expr` → **0/garbage on EVERY platform**: `OP_AWAIT` had NO case in x86_selector
  (only OP_SPAWN at x86_selector.ax:1868) → dest reg never written. Measured: `await fetch()`
  where fetch()==42 exited **0**.
- top-level `async fn` **did not parse** ("expected top level declaration") — `async` was
  only handled in `parse_method_sig` (interface methods), not the top-level dispatcher.
- `typecheck.ax` had ZERO async/await handling → `await` node type unset.
- `spawn handler(...)` WORKS (OP_SPAWN → `ax_actor_spawn(handler,0,0)` → actor id u64;
  discards the call-args, supplies handler ptr via RIP-relative lea). `std/scheduler.ax:492`
  uses `spawn supervisor_handler(...)` and scheduler.ax is bundled into EVERY program incl.
  the compiler self-build → spawn **cannot be rejected**.

## Phase A — SHIPPED `81e6f29`, A==B==C `91AE46F2`, 535/535
async/await v1 = synchronous: `async fn` is a normal fn (FLAG_IS_ASYNC = forward-compat
marker, no runtime diff); `await e` evaluates e to completion and yields its value (type =
e's type, no Future[T] wrapper); `await` allowed in any fn (v1 doesn't require async context).
3 edits, all inert on self-build (compiler uses no `await`/top-level `async fn`):
- **parser.ax** top-level dispatcher: `elif TK_ASYNC` (+ `pub async fn` in the TK_PUB branch)
  → parse_func_decl + set FLAG_IS_ASYNC.
- **typecheck.ax** infer_node: `elif NODE_AWAIT_EXPR` → result_type = infer_node(child).
  (Left NODE_SPAWN_EXPR UNTOUCHED — it works, and inferring the call-child could ripple the
  self-build's node_types / add diagnostics on scheduler.ax's spawn.)
- **x86_selector.ax**: `elif OP_AWAIT` → `MACH_MOV dest,src1` — target-independent identity
  forward (awaited expr already ran in lowering). No ABI branch → cross-platform.
OP_AWAIT correctly NOT in `has_side_effect` (ssa_opt.ax:40) — it's a pure move; the child
OP_CALL keeps side effects so the awaited expr always runs. Oracle `t_awaitv1`(42): nested
await + pub async fn + await-in-plain-fn, O0 & O1.

## Phase B — timer (PENDING). Key findings that reframe "timer already exists":
- `std/time.ax` **does NOT compile** as written — uses `impl Duration { ... }` brace blocks;
  the compiler has **NO `impl` parser** (grep TK_IMPL/parse_impl = 0). So Instant/SystemTime/
  sleep are nominal only. Phase B needs a **grammar-clean** timer (free fns, indent bodies).
- Linker `get_dll_for_symbol` (linker.ax:662) hard-lists kernel32 imports; **`Sleep`,
  `QueryPerformanceCounter`, `CreateThread`, `WaitForSingleObject`, `GetSystemTimeAsFileTime`
  are NOT in it** → misroute to ucrtbase.dll → unresolved. Must be ADDED (linker change → B==C).
- `ax_time_now_ns`/`ax_timer_after` ARE in the ax_runtime.dll valid list (linker.ax:651) → on
  Windows they resolve to the prebuilt ax_runtime.dll; on Linux (no ax_runtime.dll) need a
  bundled syscall impl (the historic "ax_time_now_ns missing on ELF" gap).
- Design: blocking timer only in v1 (Win `Sleep`/kernel32, Linux `nanosleep` syscall 35).
  Async timer (`sleep_async` suspend/resume) = v2 reactor (epoll/IOCP), deferred.

## Phase C — threads (SHIPPED for Windows). std/thread.ax + oracle t_threadv1(42).
`std/thread.ax`: `thread_spawn(entry: fn(ptr[void])->u32, arg) -> handle` + `thread_join` +
`threads_supported()`. Windows = REAL preemptive OS threads via kernel32 CreateThread/
WaitForSingleObject (linker imports shipped in Phase B). Oracle `bin/t_threadv1.ax`: spawn a
thread that writes a module global, join, read it back = 42 at O0 AND O1 → the thread runs to
completion and shared memory is visible. **No compiler change** (library + oracle + Phase-B
linker imports). Linux: `thread_spawn` returns null (deferred — freestanding ELF, no libc, heap
still blocked [[rfc0009-p3-elf-linux-target-wip]]; needs clone(2)+mmap stack+exit trampoline).

### ⚠️⚠️ The "Phase C blocked by a miscompile" scare was a FALSE ALARM (name collision)
Earlier this session I mis-diagnosed threads as blocked by a "param + non-void return + module-
global write loses the write" miscompile (committed `6d6eba5`, since CORRECTED). It was NOT a
miscompile — every failing minimal test was named `worker`, which collides with the bundled
`std/scheduler.ax` `struct worker`; `call worker` mis-links to a stdlib symbol so the user fn
never runs. Every PASSING test used a different name (setit/dbl/wf). The name was the real
variable. Renaming → correct; threads work. See [[bug-user-fn-stdlib-struct-name-collision]].
**Lesson: hold the symbol NAME fixed when minimizing — a name colliding with bundled stdlib is
invisible in a feature matrix. Disassembly (main calling the wrong address) broke the false
theory.** The bogus `bug-global-write-param-return-lost.md` + `known_fail_global_param_return.ax`
were deleted.

## Gate cmd (unchanged): `& scripts/fast_fixpoint.ps1` → cp fpB→axc_native →
`REGTMP=bin/_regtmp AXC=bin/axc_native.exe bash scripts/regression_repros.sh`.
Backend change (selector/linker) → B==C mandatory before commit; frontend → A==B.
