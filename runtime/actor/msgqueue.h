/*
 * p15-t05: Actor Message Queue — Lock-Free MPSC
 *
 * Multiple-Producer Single-Consumer queue using a lock-free
 * intrusive linked list (Michael-Scott style, simplified).
 */

#ifndef AXIOM_RUNTIME_MSGQUEUE_H
#define AXIOM_RUNTIME_MSGQUEUE_H

#include "actor.h"

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Priority Messages
 * -------------------------------------------------------------------------- */

#define AX_MSG_PRIORITY_NORMAL  0
#define AX_MSG_PRIORITY_HIGH    1
#define AX_MSG_PRIORITY_SYSTEM  2

/* --------------------------------------------------------------------------
 * Batched Message Queue
 *
 * Uses two queues: normal and priority.
 * Consumer drains priority queue first.
 * -------------------------------------------------------------------------- */

typedef struct {
    AxMsgQueue normal;     /* normal priority messages */
    AxMsgQueue priority;   /* high-priority / system messages */
    uint64_t   total;      /* total messages ever enqueued */
    uint64_t   dropped;    /* messages dropped (overflow, dead actor) */
    uint32_t   max_batch;  /* max messages to process per step */
} AxPriorityMsgQueue;

/** Initialize a priority message queue. */
void ax_pmq_init(AxPriorityMsgQueue* pq, uint32_t max_batch);

/** Enqueue a message with priority. */
void ax_pmq_push(AxPriorityMsgQueue* pq, AxMessage* msg, int priority);

/** Dequeue the next message (priority first). */
AxMessage* ax_pmq_pop(AxPriorityMsgQueue* pq);

/** Check if both queues are empty. */
int ax_pmq_empty(const AxPriorityMsgQueue* pq);

/** Get total pending count. */
uint64_t ax_pmq_pending(const AxPriorityMsgQueue* pq);

/* --------------------------------------------------------------------------
 * Dead Letter Queue
 *
 * Messages sent to dead/unknown actors are routed here.
 * -------------------------------------------------------------------------- */

typedef struct {
    AxMsgQueue queue;
    uint64_t   count;
} AxDeadLetterQueue;

void ax_dlq_init(AxDeadLetterQueue* dlq);
void ax_dlq_push(AxDeadLetterQueue* dlq, AxMessage* msg);
AxMessage* ax_dlq_pop(AxDeadLetterQueue* dlq);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_MSGQUEUE_H */
