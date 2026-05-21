/*
 * p15-t06: Async/Await Runtime — Implementation
 */

#include "async.h"
#include <stdlib.h>
#include <string.h>

#ifdef _WIN32
  #include <windows.h>
#else
  #include <time.h>
#endif

/* --------------------------------------------------------------------------
 * Future Lifecycle
 * -------------------------------------------------------------------------- */

AxFuture* ax_future_new(uint32_t type_tag) {
    AxFuture* f = (AxFuture*)calloc(1, sizeof(AxFuture));
    if (!f) return NULL;
    f->state = AX_FUTURE_PENDING;
    f->type_tag = type_tag;
    return f;
}

void ax_future_resolve(AxFuture* f, void* result, size_t size) {
    if (!f || f->state != AX_FUTURE_PENDING) return;

    if (result && size > 0) {
        f->result = malloc(size);
        if (f->result) {
            memcpy(f->result, result, size);
            f->result_size = size;
        }
    }
    f->state = AX_FUTURE_READY;
}

void ax_future_reject(AxFuture* f, void* error) {
    if (!f || f->state != AX_FUTURE_PENDING) return;
    f->error = error;
    f->state = AX_FUTURE_ERROR;
}

void ax_future_cancel(AxFuture* f) {
    if (!f || f->state != AX_FUTURE_PENDING) return;
    f->state = AX_FUTURE_CANCELLED;
}

int ax_future_is_ready(const AxFuture* f) {
    return f && f->state == AX_FUTURE_READY;
}

int ax_future_poll(AxFuture* f) {
    if (!f) return 0;
    return f->state != AX_FUTURE_PENDING;
}

void* ax_future_await(AxFuture* f) {
    if (!f) return NULL;
    /*
     * In a real implementation, this would:
     * 1. Suspend the current actor's coroutine
     * 2. Register a waker callback
     * 3. Yield back to the scheduler
     * 4. Resume when the future is resolved
     *
     * For now: busy-wait (only for testing).
     */
    while (f->state == AX_FUTURE_PENDING) {
        /* yield to scheduler in real impl */
    }
    return f->result;
}

void ax_future_free(AxFuture* f) {
    if (!f) return;
    if (f->result) free(f->result);
    free(f);
}

/* --------------------------------------------------------------------------
 * Timer
 * -------------------------------------------------------------------------- */

uint64_t ax_time_now_ns(void) {
#ifdef _WIN32
    LARGE_INTEGER freq, counter;
    QueryPerformanceFrequency(&freq);
    QueryPerformanceCounter(&counter);
    return (uint64_t)(counter.QuadPart * 1000000000ULL / freq.QuadPart);
#else
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
#endif
}

AxFuture* ax_timer_after(uint64_t delay_ms, uint64_t actor_id) {
    AxFuture* f = ax_future_new(0);
    if (!f) return NULL;
    f->waiter_actor = actor_id;
    /*
     * In a real implementation, this would register the timer
     * with the IO event loop (p15-t08) which fires it after delay_ms.
     * For now, immediately resolve (stub).
     */
    uint64_t dummy = delay_ms;
    ax_future_resolve(f, &dummy, sizeof(dummy));
    return f;
}
