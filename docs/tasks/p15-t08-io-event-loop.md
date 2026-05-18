# p15-t08: I/O Event Loop

## Purpose
Implement a non-blocking I/O event loop (epoll on Linux, IOCP on Windows, kqueue on macOS) that integrates with the async/await runtime to wake actor tasks when I/O is ready.

## Context
AXIOM's async model requires that I/O operations (file read, network send/recv) don't block OS threads. Instead, when an actor awaits I/O, the thread returns to the scheduler. When I/O completes (detected by the event loop), the awaiting task is re-enqueued. This enables one worker thread to serve thousands of concurrent I/O operations.

## Inputs
- `AxFuture` from p15-t06 — future to resolve when I/O completes
- OS event APIs: `epoll` (Linux), `kqueue` (macOS/BSD), `IOCP` (Windows)
- File descriptors and socket handles

## Outputs
- `runtime/ioloop.c` — I/O event loop
- `ax_io_read_async()`, `ax_io_write_async()`, `ax_io_connect_async()` API

## Dependencies
- p15-t06: async-await-runtime — `ax_future_resolve()` to wake tasks
- p15-t02: scheduler — re-enqueue tasks after I/O completion

## Subsystems Affected
- Async networking/file I/O (stdlib, p16)
- Scheduler: I/O loop runs as a dedicated thread (not a worker)

## Detailed Requirements

```c
typedef struct AxIOLoop {
    int       epoll_fd;     // Linux: epoll; macOS: kqueue; Windows: IOCP handle
    pthread_t thread;       // dedicated I/O loop thread
    _Atomic int running;
} AxIOLoop;

typedef struct AxIORequest {
    int        fd;
    void*      buffer;
    size_t     size;
    AxFuture*  future;      // resolved with bytes_read/written when done
    uint32_t   events;      // EPOLLIN, EPOLLOUT
} AxIORequest;

// Initialize event loop (called from ax_runtime_init)
int ax_ioloop_init(AxIOLoop* loop);
void ax_ioloop_run(AxIOLoop* loop);   // blocks; call in dedicated thread
void ax_ioloop_stop(AxIOLoop* loop);

// Register async I/O (returns future)
AxFuture* ax_io_read_async(AxIOLoop* loop, int fd, void* buf, size_t size, ActorHeap* heap);
AxFuture* ax_io_write_async(AxIOLoop* loop, int fd, const void* buf, size_t size, ActorHeap* heap);

// Internal: I/O completion handler
void ax_ioloop_on_ready(AxIOLoop* loop, int fd, uint32_t events);
```

Event loop implementation (Linux):
```c
void ax_ioloop_run(AxIOLoop* loop) {
    struct epoll_event events[64];
    while (loop->running) {
        int n = epoll_wait(loop->epoll_fd, events, 64, 10 /* ms */);
        for (int i = 0; i < n; i++) {
            AxIORequest* req = events[i].data.ptr;
            ssize_t bytes = read(req->fd, req->buffer, req->size);
            ax_future_resolve(req->future, &bytes);
        }
    }
}
```

Platform detection: `#ifdef __linux__` → epoll; `#ifdef __APPLE__` → kqueue; `#ifdef _WIN32` → IOCP.

## Implementation Steps

1. Create `runtime/ioloop.c` with platform-specific backends.
2. Implement `ax_ioloop_init()` — create epoll/kqueue/IOCP descriptor.
3. Start dedicated I/O thread in `ax_runtime_init()`.
4. Implement `ax_io_read_async()` — register fd with event loop, return future.
5. Implement `ax_ioloop_on_ready()` — perform I/O, resolve future.
6. Implement `ax_ioloop_stop()` — signal thread to exit.
7. Write tests with socketpair (localhost loopback).

## Test Plan
- `TestIOLoopRead`: async read from pipe → future resolved with bytes
- `TestIOLoopWrite`: async write to pipe → future resolved
- `TestIOLoopConcurrent`: 100 concurrent async reads on 100 fds → all complete
- `TestIOLoopShutdown`: stop loop → no pending futures remain

## Validation Checklist
- [ ] Event loop runs in dedicated thread (not a worker)
- [ ] Future resolved in I/O thread, task re-enqueued in scheduler
- [ ] No OS thread blocked waiting for I/O
- [ ] Platform backend selected at compile time

## Acceptance Criteria
- 10K concurrent TCP connections handled with 4 worker threads

## Definition of Done
- [ ] `runtime/ioloop.c` implemented for Linux (epoll)
- [ ] Async read/write tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Buffer lifetime: task re-enqueues before I/O completes | Buffer stays alive in actor heap until future resolved |
| epoll_wait error (EINTR) | Retry on EINTR; log other errors |

## Future Follow-up Tasks
- TLS async support (via OpenSSL BIO with non-blocking fd)
- UDP async read/write
- kqueue and IOCP backends
