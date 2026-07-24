# RFC 0033 — Concurrency Model v1 (async / await / spawn / timer / thread)

- Status: **Active** (Phase A shipped)
- Author: autopilot
- Supersedes/relates: the actor/scheduler subsystem (`std/scheduler.ax`), `std/time.ax`,
  `std/reactor.ax`, RFC 0009 P3 (Linux ELF target).

## 1. Motivation

AXIOM has surface syntax for `async fn`, `await`, and `spawn`, plus a bundled actor
scheduler and time primitives — but the pieces were incoherent and partly broken:

- `await expr` **produced no code** (`OP_AWAIT` had no instruction-selector case) → it
  returned garbage/0 on **every** platform. Confirmed: `await fetch()` where `fetch()`
  returns 42 exited **0**.
- Top-level `async fn` **did not parse** (`async` was only handled on interface method
  signatures) → the `tests/generics/async_*.ax` programs could not compile.
- `spawn handler(...)` already worked (creates an actor via `ax_actor_spawn`, returns its
  id) and is used by `std/scheduler.ax` itself, so it is bundled into every program
  (including the compiler self-build) and **cannot be rejected**.

The prior framing ("finish async" = build a preemptive suspending runtime) is RFC-scale and
risky. This RFC instead **defines a coherent, correct, cross-platform v1** whose semantics
are deterministic and identical on Windows COFF and Linux ELF, and lays out the phased path
to a real preemptive v2.

## 2. The two orthogonal models

AXIOM concurrency has **two independent axes**; conflating them is what broke things.

1. **Value/coroutine axis — `async` / `await`.** About *computing a value* that may be
   produced asynchronously. In v1 this degrades to **synchronous cooperative** evaluation.
2. **Task/actor axis — `spawn` / message send / scheduler.** About *independent units of
   execution* that communicate by messages. This is the real concurrency primitive and is
   unchanged.

These do not need to interoperate in v1 (you do not `await` a `spawn`).

## 3. Semantics — v1

### 3.1 `async fn`
`async fn f(...) -> T` is **a normal function**. The `async` modifier is accepted at
top level and on method signatures, recorded as `FLAG_IS_ASYNC` (forward-compatibility
metadata), and imposes **no runtime difference** in v1. Rationale: it lets programs be
written in their eventual form; a v2 that makes `async fn` return a real `Future[T]` can do
so without a syntax change.

### 3.2 `await expr`
`await expr` **evaluates `expr` to completion synchronously and yields its value.** The type
of `await expr` is the type of `expr` (no `Future[T]` wrapper exists in v1; `await` on a
plain value/call is the identity on the value). `await` is permitted in any function (v1
does not require an `async` enclosing context — synchronous evaluation is always valid).
Codegen: `OP_AWAIT dest = src1` — a target-independent register move.

> Forward-compat: in a v2 preemptive runtime, `await` becomes a suspension point but keeps
> the same **value** semantics, so programs correct under v1 stay correct under v2.

### 3.3 `spawn handler(...)`
Unchanged. Creates an actor from the handler function pointer (`ax_actor_spawn`) and returns
its **actor id** (`u64`). Actual execution/concurrency is driven by the cooperative
scheduler (`scheduler.run()` steps actors on the calling thread). Message-based; actors do
not have a function return value, so `spawn` is **not** an awaitable future in v1.

### 3.4 Timer
Time is **orthogonal to async** and already cross-platform in `std/time.ax`:
`Instant.now()` (monotonic), `SystemTime.now()` (wall clock), and **blocking `sleep(Duration)`**
(Windows `Sleep`, Linux `nanosleep`). v1 ships the **blocking** timer only. An *async* timer
(`sleep_async` that suspends and resumes after N ms) requires the suspension machinery of a
reactor/event-loop (`std/reactor.ax`, epoll/IOCP) and is **v2**.

### 3.5 Thread
A real OS-thread primitive (`Thread.spawn(entry, arg) -> handle`, `join(handle)`) is
**Phase C** (see §5). Windows has a direct implementation via `kernel32`
`CreateThread`/`WaitForSingleObject`. Linux ELF is constrained: the freestanding ELF target
has no libc and its heap allocator is still blocked (RFC 0009 P3), so shared-heap Linux
threads are deferred; compute-only Linux threads via the raw `clone(2)` syscall are the
eventual path.

## 4. Cross-platform guarantee

Everything in v1 is deterministic and identical on Windows and Linux:
- `async`/`await` lower to ordinary computation + a register move — **no threads, no OS
  calls** → byte-for-byte behaviour parity.
- `spawn`/scheduler is a single-threaded cooperative loop — no OS threads.
- `sleep`/`Instant`/`SystemTime` already branch on `is_windows` to the right OS primitive.

## 5. Phases

- **Phase A — async/await v1 correctness (SHIPPED).** Parse top-level `async fn`; type
  `await e` as `e`'s type; `OP_AWAIT → mov`. Inert on the compiler self-build (it uses no
  `await`/top-level `async fn`) → A==B==C. Oracle `t_awaitv1`.
- **Phase B — timer.** Confirm blocking `sleep`/`Instant` compile & run on both targets;
  add a runtime oracle. No new runtime.
- **Phase C — threads (SHIPPED for Windows; Linux deferred).** `std/thread.ax`:
  `thread_spawn(entry: fn(ptr[void])->u32, arg) -> handle` + `thread_join(handle)` +
  `threads_supported()`. Windows uses real preemptive OS threads via kernel32
  `CreateThread`/`WaitForSingleObject` (linker imports added in Phase B). Oracle `t_threadv1`
  spawns a thread that writes a module global, joins, and reads it back = 42 (O0 & O1) — proving
  the thread runs to completion and shared memory is visible. No compiler change (library +
  oracle + the Phase B linker imports). Linux: `thread_spawn` returns null (deferred — freestanding
  ELF, no libc, heap still blocked per RFC 0009 P3; real Linux threads need `clone(2)` + mmap'd
  child stack + exit-syscall trampoline). ⚠️ The earlier "Phase C blocked by a param+return+global
  miscompile" was WRONG — it was a `worker` name-collision with the bundled `std/scheduler.ax`
  `struct worker` (see `knowledge/bug-user-fn-stdlib-struct-name-collision.md`); real threads work
  with a distinctively-named entry. Do NOT name a thread entry `worker`.
- **Phase D (v2, deferred) — preemptive async.** `Future[T]`, real suspension at `await`,
  reactor-driven `sleep_async`/IO, scheduler multi-threading (`worker_loop` already
  scaffolded). RFC-scale; separate effort.

## 6. Drawbacks / alternatives

- **v1 `await` is "fake" async.** Accepted: it is *correct* (right value, deterministic,
  cross-platform) and forward-compatible. A silent wrong-value `await` (the prior state) is
  strictly worse than a correct synchronous one.
- **Rejecting `spawn`/`await` as unimplemented** was considered and is **impossible**:
  `std/scheduler.ax` uses `spawn` and is bundled into the self-build.
- **Requiring `await` only inside `async fn`** was considered and deferred: it adds a
  reject path with over-reject risk for zero v1 benefit (synchronous eval is always valid).

## 7. Test / gate

Backend change (`x86_selector`) but inert on self-host → **B==C** required and expected
equal to A. Full regression + oracle `t_awaitv1` (exit = the awaited value). Frontend parts
(parser/typecheck) also inert → A==B.
