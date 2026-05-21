/*
 * p15-t05: Actor Message Queue — Implementation
 */

#include "msgqueue.h"
#include <string.h>

/* --------------------------------------------------------------------------
 * Priority Message Queue
 * -------------------------------------------------------------------------- */

void ax_pmq_init(AxPriorityMsgQueue* pq, uint32_t max_batch) {
    ax_msgq_init(&pq->normal);
    ax_msgq_init(&pq->priority);
    pq->total = 0;
    pq->dropped = 0;
    pq->max_batch = max_batch > 0 ? max_batch : 64;
}

void ax_pmq_push(AxPriorityMsgQueue* pq, AxMessage* msg, int priority) {
    if (!pq || !msg) return;
    pq->total++;

    if (priority >= AX_MSG_PRIORITY_HIGH) {
        ax_msgq_push(&pq->priority, msg);
    } else {
        ax_msgq_push(&pq->normal, msg);
    }
}

AxMessage* ax_pmq_pop(AxPriorityMsgQueue* pq) {
    if (!pq) return NULL;

    /* Priority first */
    AxMessage* msg = ax_msgq_pop(&pq->priority);
    if (msg) return msg;

    return ax_msgq_pop(&pq->normal);
}

int ax_pmq_empty(const AxPriorityMsgQueue* pq) {
    if (!pq) return 1;
    return ax_msgq_empty(&pq->normal) && ax_msgq_empty(&pq->priority);
}

uint64_t ax_pmq_pending(const AxPriorityMsgQueue* pq) {
    if (!pq) return 0;
    return pq->normal.pending + pq->priority.pending;
}

/* --------------------------------------------------------------------------
 * Dead Letter Queue
 * -------------------------------------------------------------------------- */

void ax_dlq_init(AxDeadLetterQueue* dlq) {
    ax_msgq_init(&dlq->queue);
    dlq->count = 0;
}

void ax_dlq_push(AxDeadLetterQueue* dlq, AxMessage* msg) {
    if (!dlq || !msg) return;
    ax_msgq_push(&dlq->queue, msg);
    dlq->count++;
}

AxMessage* ax_dlq_pop(AxDeadLetterQueue* dlq) {
    if (!dlq) return NULL;
    AxMessage* msg = ax_msgq_pop(&dlq->queue);
    if (msg) dlq->count--;
    return msg;
}
