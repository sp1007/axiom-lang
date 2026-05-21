/*
 * p15-t08: I/O Event Loop
 *
 * Non-blocking I/O event loop integrated with async/await.
 * Uses platform-specific backends (epoll/kqueue/IOCP).
 */

#ifndef AXIOM_RUNTIME_IOLOOP_H
#define AXIOM_RUNTIME_IOLOOP_H

#include "async.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define AX_MAX_IO_EVENTS 1024

typedef enum {
    AX_IO_READ  = 1,
    AX_IO_WRITE = 2,
    AX_IO_ERROR = 4,
    AX_IO_CLOSE = 8,
} AxIOEvent;

typedef struct {
    int         fd;            /* file descriptor / socket handle */
    uint32_t    events;        /* requested events mask */
    AxFuture*   future;        /* future to resolve on event */
    void*       buffer;        /* read/write buffer */
    size_t      buf_size;      /* buffer size */
    uint64_t    actor_id;      /* owning actor */
} AxIORequest;

typedef struct {
    int         backend_fd;    /* epoll/kqueue fd */
    int         running;
    uint64_t    events_processed;
    AxIORequest pending[AX_MAX_IO_EVENTS];
    uint32_t    pending_count;
} AxIOLoop;

/** Initialize the I/O event loop. */
int ax_ioloop_init(AxIOLoop* loop);

/** Register an I/O request. */
int ax_ioloop_register(AxIOLoop* loop, const AxIORequest* req);

/** Unregister an fd. */
int ax_ioloop_unregister(AxIOLoop* loop, int fd);

/** Run one iteration of the event loop (poll + dispatch). */
int ax_ioloop_poll(AxIOLoop* loop, int timeout_ms);

/** Run the event loop until stopped. */
void ax_ioloop_run(AxIOLoop* loop);

/** Stop the event loop. */
void ax_ioloop_stop(AxIOLoop* loop);

/** Destroy the event loop. */
void ax_ioloop_destroy(AxIOLoop* loop);

/* --------------------------------------------------------------------------
 * Async I/O Operations
 * -------------------------------------------------------------------------- */

/** Async read: returns a future that resolves when data is available. */
AxFuture* ax_io_read_async(AxIOLoop* loop, int fd, void* buf, size_t size,
                           uint64_t actor_id);

/** Async write: returns a future that resolves when write completes. */
AxFuture* ax_io_write_async(AxIOLoop* loop, int fd, const void* buf,
                            size_t size, uint64_t actor_id);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_IOLOOP_H */
