/*
 * p15-t02: M:N Work-Stealing Scheduler
 *
 * Maps N actors onto M OS worker threads using work-stealing deques.
 * Each worker has a local run queue. When empty, steals from a random
 * victim's queue.
 */

#ifndef AXIOM_RUNTIME_SCHEDULER_H
#define AXIOM_RUNTIME_SCHEDULER_H

#include "actor.h"

#ifdef __cplusplus
extern "C" {
#endif

#define AX_MAX_WORKERS    256
#define AX_RUNQ_SIZE      4096

/* --------------------------------------------------------------------------
 * Run Queue (Chase-Lev Work-Stealing Deque)
 *
 * Owner pushes/pops from bottom. Stealers take from top.
 * -------------------------------------------------------------------------- */

typedef struct {
    AxActorID  buffer[AX_RUNQ_SIZE];
    uint64_t   top;       /* steal end (atomic) */
    uint64_t   bottom;    /* owner end (atomic) */
} AxRunQueue;

/** Initialize a run queue. */
void ax_runq_init(AxRunQueue* q);

/** Push an actor ID (owner only). Returns 0 on success. */
int ax_runq_push(AxRunQueue* q, AxActorID id);

/** Pop an actor ID (owner only). Returns AX_ACTOR_ID_NONE if empty. */
AxActorID ax_runq_pop(AxRunQueue* q);

/** Steal an actor ID (any thread). Returns AX_ACTOR_ID_NONE if empty. */
AxActorID ax_runq_steal(AxRunQueue* q);

/** Check if run queue is empty. */
int ax_runq_empty(const AxRunQueue* q);

/* --------------------------------------------------------------------------
 * Worker Thread
 * -------------------------------------------------------------------------- */

typedef struct {
    uint32_t     id;
    AxRunQueue   runq;
    uint64_t     tasks_executed;
    uint64_t     steals_attempted;
    uint64_t     steals_succeeded;
    int          running;
} AxWorker;

/* --------------------------------------------------------------------------
 * Scheduler
 * -------------------------------------------------------------------------- */

typedef struct AxScheduler {
    AxWorker   workers[AX_MAX_WORKERS];
    uint32_t   worker_count;
    int        running;
    uint64_t   total_submitted;
} AxScheduler;

/** Initialize the scheduler with the given number of worker threads. */
int ax_scheduler_init(AxScheduler* sched, uint32_t worker_count);

/** Submit an actor to be scheduled for execution. */
int ax_scheduler_submit(AxScheduler* sched, AxActorID actor_id);

/** Start the scheduler (spawns worker threads). */
int ax_scheduler_run(AxScheduler* sched);

/** Shutdown the scheduler (joins all worker threads). */
void ax_scheduler_shutdown(AxScheduler* sched);

/** Get scheduler statistics. */
typedef struct {
    uint32_t worker_count;
    uint64_t total_submitted;
    uint64_t total_executed;
    uint64_t total_steals;
} AxSchedulerStats;

void ax_scheduler_stats(const AxScheduler* sched, AxSchedulerStats* stats);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_SCHEDULER_H */
