/*
 * p15-t08: I/O Event Loop — Implementation (Stub/Portable)
 *
 * Stub implementation using select() for portability.
 * Real backends (epoll/kqueue/IOCP) will be added per-platform.
 */

#include "ioloop.h"
#include <stdlib.h>
#include <string.h>

/* --------------------------------------------------------------------------
 * Init / Destroy
 * -------------------------------------------------------------------------- */

int ax_ioloop_init(AxIOLoop* loop) {
    if (!loop) return -1;
    memset(loop, 0, sizeof(AxIOLoop));
    loop->backend_fd = -1;
    loop->running = 0;
    return 0;
}

void ax_ioloop_destroy(AxIOLoop* loop) {
    if (!loop) return;
    loop->running = 0;
    loop->pending_count = 0;
}

/* --------------------------------------------------------------------------
 * Register / Unregister
 * -------------------------------------------------------------------------- */

int ax_ioloop_register(AxIOLoop* loop, const AxIORequest* req) {
    if (!loop || !req) return -1;
    if (loop->pending_count >= AX_MAX_IO_EVENTS) return -1;

    loop->pending[loop->pending_count++] = *req;
    return 0;
}

int ax_ioloop_unregister(AxIOLoop* loop, int fd) {
    if (!loop) return -1;

    for (uint32_t i = 0; i < loop->pending_count; i++) {
        if (loop->pending[i].fd == fd) {
            /* Remove by shifting */
            for (uint32_t j = i; j < loop->pending_count - 1; j++) {
                loop->pending[j] = loop->pending[j + 1];
            }
            loop->pending_count--;
            return 0;
        }
    }
    return -1;
}

/* --------------------------------------------------------------------------
 * Poll (Stub: immediately resolve all pending)
 * -------------------------------------------------------------------------- */

int ax_ioloop_poll(AxIOLoop* loop, int timeout_ms) {
    if (!loop) return -1;
    (void)timeout_ms;

    /*
     * Stub implementation: resolve all pending I/O futures immediately.
     * A real implementation would:
     * 1. Call epoll_wait / kevent / select
     * 2. For each ready fd, resolve the associated future
     * 3. Re-enqueue the actor for scheduling
     */
    int processed = 0;
    for (uint32_t i = 0; i < loop->pending_count; i++) {
        AxIORequest* req = &loop->pending[i];
        if (req->future && !ax_future_is_ready(req->future)) {
            /* Simulate completion: resolve with bytes_transferred = buf_size */
            size_t result = req->buf_size;
            ax_future_resolve(req->future, &result, sizeof(result));
            processed++;
        }
    }

    loop->events_processed += (uint64_t)processed;
    return processed;
}

/* --------------------------------------------------------------------------
 * Run / Stop
 * -------------------------------------------------------------------------- */

void ax_ioloop_run(AxIOLoop* loop) {
    if (!loop) return;
    loop->running = 1;

    while (loop->running && loop->pending_count > 0) {
        ax_ioloop_poll(loop, 100);
    }
}

void ax_ioloop_stop(AxIOLoop* loop) {
    if (!loop) return;
    loop->running = 0;
}

/* --------------------------------------------------------------------------
 * Async I/O
 * -------------------------------------------------------------------------- */

AxFuture* ax_io_read_async(AxIOLoop* loop, int fd, void* buf, size_t size,
                           uint64_t actor_id) {
    AxFuture* f = ax_future_new(0);
    if (!f) return NULL;
    f->waiter_actor = actor_id;

    AxIORequest req;
    req.fd = fd;
    req.events = AX_IO_READ;
    req.future = f;
    req.buffer = buf;
    req.buf_size = size;
    req.actor_id = actor_id;

    ax_ioloop_register(loop, &req);
    return f;
}

AxFuture* ax_io_write_async(AxIOLoop* loop, int fd, const void* buf,
                            size_t size, uint64_t actor_id) {
    AxFuture* f = ax_future_new(0);
    if (!f) return NULL;
    f->waiter_actor = actor_id;

    AxIORequest req;
    req.fd = fd;
    req.events = AX_IO_WRITE;
    req.future = f;
    req.buffer = (void*)buf;
    req.buf_size = size;
    req.actor_id = actor_id;

    ax_ioloop_register(loop, &req);
    return f;
}
