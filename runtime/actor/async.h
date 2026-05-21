/*
 * p15-t06: Async/Await Runtime Support
 *
 * Stackless coroutine infrastructure for async functions.
 * Each async function compiles to a state machine that yields
 * at each await point.
 */

#ifndef AXIOM_RUNTIME_ASYNC_H
#define AXIOM_RUNTIME_ASYNC_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Future / Promise
 * -------------------------------------------------------------------------- */

typedef enum {
    AX_FUTURE_PENDING   = 0,
    AX_FUTURE_READY     = 1,
    AX_FUTURE_CANCELLED = 2,
    AX_FUTURE_ERROR     = 3,
} AxFutureState;

typedef struct AxFuture {
    AxFutureState   state;
    void*           result;        /* result value when READY */
    size_t          result_size;
    void*           error;         /* error value when ERROR */
    uint32_t        type_tag;      /* result type tag */

    /* Continuation: state machine for stackless coroutine */
    void*           coroutine;     /* pointer to coroutine state */
    int             resume_point;  /* which await point to resume at */

    /* Waker: who to notify when result is ready */
    uint64_t        waiter_actor;  /* actor waiting for this future */
    struct AxFuture* next;         /* linked list for pending futures */
} AxFuture;

/** Create a new pending future. */
AxFuture* ax_future_new(uint32_t type_tag);

/** Set the future's result (transitions to READY). */
void ax_future_resolve(AxFuture* f, void* result, size_t size);

/** Set the future's error (transitions to ERROR). */
void ax_future_reject(AxFuture* f, void* error);

/** Cancel the future. */
void ax_future_cancel(AxFuture* f);

/** Check if the future is ready. */
int ax_future_is_ready(const AxFuture* f);

/** Poll the future. Returns 1 if ready, 0 if pending. */
int ax_future_poll(AxFuture* f);

/** Block the current actor until the future is ready. */
void* ax_future_await(AxFuture* f);

/** Free a future. */
void ax_future_free(AxFuture* f);

/* --------------------------------------------------------------------------
 * Timer
 * -------------------------------------------------------------------------- */

typedef struct {
    uint64_t    deadline_ns;   /* absolute deadline in nanoseconds */
    AxFuture*   future;        /* future to resolve when timer fires */
    uint64_t    actor_id;      /* actor that created this timer */
} AxTimer;

/** Create a timer that resolves after delay_ms milliseconds. */
AxFuture* ax_timer_after(uint64_t delay_ms, uint64_t actor_id);

/** Get current time in nanoseconds (monotonic). */
uint64_t ax_time_now_ns(void);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_ASYNC_H */
