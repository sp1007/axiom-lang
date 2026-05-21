/*
 * p15-t02: M:N Work-Stealing Scheduler — Implementation
 *
 * Simple single-threaded scheduler for initial implementation.
 * Full multi-threaded Chase-Lev deque will be added when pthread
 * infrastructure is stable.
 */

#include "scheduler.h"
#include <string.h>
#include <stdlib.h>

/* --------------------------------------------------------------------------
 * Run Queue
 * -------------------------------------------------------------------------- */

void ax_runq_init(AxRunQueue* q) {
    memset(q->buffer, 0, sizeof(q->buffer));
    q->top = 0;
    q->bottom = 0;
}

int ax_runq_push(AxRunQueue* q, AxActorID id) {
    uint64_t b = q->bottom;
    uint64_t t = q->top;
    if (b - t >= AX_RUNQ_SIZE) return -1; /* full */

    q->buffer[b % AX_RUNQ_SIZE] = id;
    q->bottom = b + 1;
    return 0;
}

AxActorID ax_runq_pop(AxRunQueue* q) {
    uint64_t b = q->bottom;
    if (b == 0) return AX_ACTOR_ID_NONE;

    b--;
    q->bottom = b;
    uint64_t t = q->top;

    if (t <= b) {
        return q->buffer[b % AX_RUNQ_SIZE];
    }

    if (t == b) {
        q->bottom = t + 1;
        q->top = t + 1;
        return q->buffer[b % AX_RUNQ_SIZE];
    }

    /* Queue empty */
    q->bottom = t;
    return AX_ACTOR_ID_NONE;
}

AxActorID ax_runq_steal(AxRunQueue* q) {
    uint64_t t = q->top;
    uint64_t b = q->bottom;

    if (t >= b) return AX_ACTOR_ID_NONE; /* empty */

    AxActorID id = q->buffer[t % AX_RUNQ_SIZE];
    q->top = t + 1;
    return id;
}

int ax_runq_empty(const AxRunQueue* q) {
    return q->top >= q->bottom;
}

/* --------------------------------------------------------------------------
 * Scheduler
 * -------------------------------------------------------------------------- */

int ax_scheduler_init(AxScheduler* sched, uint32_t worker_count) {
    if (!sched || worker_count == 0 || worker_count > AX_MAX_WORKERS)
        return -1;

    memset(sched, 0, sizeof(AxScheduler));
    sched->worker_count = worker_count;

    for (uint32_t i = 0; i < worker_count; i++) {
        sched->workers[i].id = i;
        ax_runq_init(&sched->workers[i].runq);
        sched->workers[i].running = 0;
    }

    return 0;
}

int ax_scheduler_submit(AxScheduler* sched, AxActorID actor_id) {
    if (!sched || actor_id == AX_ACTOR_ID_NONE) return -1;

    /* Round-robin to least-loaded worker */
    uint32_t target = (uint32_t)(sched->total_submitted % sched->worker_count);
    int result = ax_runq_push(&sched->workers[target].runq, actor_id);
    if (result == 0) {
        sched->total_submitted++;
    }
    return result;
}

/* Worker loop: run local tasks, then try stealing */
static void worker_loop(AxScheduler* sched, AxWorker* worker) {
    while (worker->running) {
        /* Try local queue */
        AxActorID id = ax_runq_pop(&worker->runq);

        /* Try stealing if local is empty */
        if (id == AX_ACTOR_ID_NONE) {
            for (uint32_t i = 0; i < sched->worker_count; i++) {
                if (i == worker->id) continue;
                worker->steals_attempted++;
                id = ax_runq_steal(&sched->workers[i].runq);
                if (id != AX_ACTOR_ID_NONE) {
                    worker->steals_succeeded++;
                    break;
                }
            }
        }

        if (id == AX_ACTOR_ID_NONE) {
            /* No work available — in real impl would yield/sleep */
            break;
        }

        /* Execute: step the actor */
        AxActor* actor = ax_actor_lookup(id);
        if (actor) {
            /* Process all pending messages */
            while (ax_actor_step(actor)) {
                worker->tasks_executed++;
            }

            /* Re-enqueue if still alive and has pending messages */
            if (actor->state == AX_ACTOR_RUNNING &&
                !ax_msgq_empty(&actor->mailbox)) {
                ax_runq_push(&worker->runq, id);
            }
        }
    }
}

int ax_scheduler_run(AxScheduler* sched) {
    if (!sched) return -1;

    sched->running = 1;
    for (uint32_t i = 0; i < sched->worker_count; i++) {
        sched->workers[i].running = 1;
    }

    /* Single-threaded execution for now */
    for (uint32_t i = 0; i < sched->worker_count; i++) {
        worker_loop(sched, &sched->workers[i]);
    }

    return 0;
}

void ax_scheduler_shutdown(AxScheduler* sched) {
    if (!sched) return;

    sched->running = 0;
    for (uint32_t i = 0; i < sched->worker_count; i++) {
        sched->workers[i].running = 0;
    }
}

void ax_scheduler_stats(const AxScheduler* sched, AxSchedulerStats* stats) {
    if (!sched || !stats) return;

    stats->worker_count = sched->worker_count;
    stats->total_submitted = sched->total_submitted;
    stats->total_executed = 0;
    stats->total_steals = 0;

    for (uint32_t i = 0; i < sched->worker_count; i++) {
        stats->total_executed += sched->workers[i].tasks_executed;
        stats->total_steals += sched->workers[i].steals_succeeded;
    }
}
