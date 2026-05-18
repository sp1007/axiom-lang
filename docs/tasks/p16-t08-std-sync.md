# p16-t08: std.sync — Synchronization Primitives

## Purpose
Implement thread-safe synchronization primitives for AXIOM programs that operate outside the actor model: Mutex, RwLock, Atomic types, Channel, and Once.

## Context
While AXIOM encourages the actor model for concurrency, programs sometimes need direct shared-state synchronization (e.g., for low-level data structures, FFI integration, or performance-critical paths). `std.sync` provides safe wrappers around OS primitives.

## Inputs
- OS threading primitives: pthread_mutex, pthread_rwlock (POSIX); CRITICAL_SECTION (Windows)
- C11 atomics: `_Atomic` types
- AXIOM ownership model: Mutex[T] owns its protected value

## Outputs
- `stdlib/sync/mutex.ax` — Mutex[T], MutexGuard[T]
- `stdlib/sync/rwlock.ax` — RwLock[T]
- `stdlib/sync/atomic.ax` — AtomicI32, AtomicI64, AtomicBool, AtomicPtr
- `stdlib/sync/channel.ax` — Channel[T] (bounded + unbounded)
- `stdlib/sync/once.ax` — Once (run-once initialization)

## Dependencies
- p14-t04: axalloc-actor-heap — channel message allocation
- p15-t05: actor-message-queue — channel implementation reuse

## Detailed Requirements

```axiom
# stdlib/sync/mutex.ax
type Mutex[T]:
    var inner: T
    var raw: RawMutex   # platform pthread_mutex_t or CRITICAL_SECTION

    fn new(val: T) -> Mutex[T]
    fn lock(mut self) -> MutexGuard[T]       # blocks until locked
    fn try_lock(mut self) -> Option[MutexGuard[T]]

type MutexGuard[T]:  # RAII guard — unlocks on scope exit (via CTGC)
    fn get(self) -> T
    fn get_mut(mut self) -> *T

# stdlib/sync/atomic.ax
type AtomicI32:
    fn new(val: i32) -> AtomicI32
    fn load(self) -> i32                     # SeqCst
    fn store(mut self, val: i32)
    fn fetch_add(mut self, delta: i32) -> i32
    fn fetch_sub(mut self, delta: i32) -> i32
    fn compare_exchange(mut self, expected: i32, new: i32) -> Result[i32, i32]

# stdlib/sync/channel.ax
type Channel[T]:
    fn new(capacity: u32) -> (Sender[T], Receiver[T])
    fn new_unbounded() -> (Sender[T], Receiver[T])

type Sender[T]:
    fn send(self, val: T) -> Result[void, ChannelError]
    fn try_send(self, val: T) -> Result[void, ChannelError]

type Receiver[T]:
    fn recv(self) -> Result[T, ChannelError]
    fn try_recv(self) -> Result[T, ChannelError]
    async fn recv_async(self) -> Result[T, ChannelError]

type ChannelError: Closed, Full, Empty

# stdlib/sync/once.ax
type Once:
    fn new() -> Once
    fn call_once(mut self, f: fn() -> void)  # runs f exactly once
    fn is_completed(self) -> bool
```

Channel implementation: bounded uses ring buffer + semaphore; unbounded uses MPSC queue from p15-t05.

Atomic: wraps C11 `_Atomic` via extern.

## Implementation Steps

1. Create `stdlib/sync/mutex.ax` — Mutex[T] wrapping pthread_mutex_t.
2. Implement MutexGuard with CTGC-based automatic unlock.
3. Create `stdlib/sync/atomic.ax` — extern wrappers for C11 atomics.
4. Create `stdlib/sync/channel.ax` — bounded ring buffer + MPSC.
5. Create `stdlib/sync/once.ax` — pthread_once or C11 atomic flag.
6. Write concurrency tests with multiple threads.

## Test Plan
- `TestMutexBasic`: lock, modify, unlock → correct value
- `TestMutexContention`: 16 threads increment shared counter → exact 16000
- `TestAtomicFetchAdd`: 16 threads × 1000 fetch_add → total = 16000
- `TestChannelSendRecv`: send 1000 messages → all received in order
- `TestChannelClosed`: recv after close → ChannelError::Closed
- `TestOnceCallOnce`: 16 threads race on call_once → fn runs exactly once

## Validation Checklist
- [ ] MutexGuard unlocks on scope exit (verified with lock count)
- [ ] Atomic operations use SeqCst by default
- [ ] Channel sender/receiver drop closes channel
- [ ] Once::call_once is safe to call concurrently

## Acceptance Criteria
- 16-thread counter increment with Mutex reaches exactly 16×1000

## Definition of Done
- [ ] All sync types implemented
- [ ] All concurrency tests pass under `-race`

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Deadlock if MutexGuard not released | CTGC ensures unlock; document no-recursive-lock constraint |

## Future Follow-up Tasks
- Condvar (condition variable) for Mutex
- Barrier for rendezvous synchronization
