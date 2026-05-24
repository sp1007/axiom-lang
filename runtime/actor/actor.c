/*
 * p15-t01: AxActor — Core Implementation
 *
 * Actor lifecycle: spawn, send, step, stop.
 * Global actor table with simple ID-indexed lookup.
 */

#include "actor.h"
#include <stdlib.h>
#include <string.h>

/* --------------------------------------------------------------------------
 * Global Actor Table
 * -------------------------------------------------------------------------- */

static AxActor g_actors[AX_MAX_ACTORS];
static uint32_t g_actor_count = 0;
static AxActorID g_next_id = 1;

void ax_actor_table_init(void) {
    memset(g_actors, 0, sizeof(g_actors));
    g_actor_count = 0;
    g_next_id = 1;
}

void ax_actor_table_destroy(void) {
    for (uint32_t i = 0; i < AX_MAX_ACTORS; i++) {
        if (g_actors[i].state != AX_ACTOR_DEAD && g_actors[i].id != 0) {
            /* Drain mailbox */
            AxMessage* msg;
            while ((msg = ax_msgq_pop(&g_actors[i].mailbox)) != NULL) {
                free(msg);
            }
            if (g_actors[i].state_data && g_actors[i].state_size > 0) {
                free(g_actors[i].state_data);
            }
            g_actors[i].state = AX_ACTOR_DEAD;
            g_actors[i].id = 0;
        }
    }
    g_actor_count = 0;
}

uint32_t ax_actor_count(void) {
    return g_actor_count;
}

/* --------------------------------------------------------------------------
 * Message Queue (simple non-thread-safe version for now)
 * -------------------------------------------------------------------------- */

void ax_msgq_init(AxMsgQueue* q) {
    q->head = NULL;
    q->tail = NULL;
    q->msg_count = 0;
    q->pending = 0;
}

void ax_msgq_push(AxMsgQueue* q, AxMessage* msg) {
    msg->next = NULL;
    if (q->tail) {
        q->tail->next = msg;
    } else {
        q->head = msg;
    }
    q->tail = msg;
    q->msg_count++;
    q->pending++;
}

AxMessage* ax_msgq_pop(AxMsgQueue* q) {
    if (!q->head) return NULL;
    AxMessage* msg = q->head;
    q->head = msg->next;
    if (!q->head) {
        q->tail = NULL;
    }
    msg->next = NULL;
    q->pending--;
    return msg;
}

int ax_msgq_empty(const AxMsgQueue* q) {
    return q->head == NULL;
}

/* --------------------------------------------------------------------------
 * Actor Lookup
 * -------------------------------------------------------------------------- */

static int id_to_slot(AxActorID id) {
    return (int)(id % AX_MAX_ACTORS);
}

AxActor* ax_actor_lookup(AxActorID id) {
    if (id == AX_ACTOR_ID_NONE) return NULL;
    int slot = id_to_slot(id);
    if (g_actors[slot].id == id && g_actors[slot].state != AX_ACTOR_DEAD) {
        return &g_actors[slot];
    }
    return NULL;
}

/* --------------------------------------------------------------------------
 * Spawn
 * -------------------------------------------------------------------------- */

AxActorID ax_actor_spawn(AxHandlerFn handler, void* init_data,
                         size_t data_size) {
    if (!handler) return AX_ACTOR_ID_NONE;
    if (g_actor_count >= AX_MAX_ACTORS) return AX_ACTOR_ID_NONE;

    AxActorID id = g_next_id++;
    int slot = id_to_slot(id);

    /* Find free slot (linear probe if collision) */
    int attempts = 0;
    while (g_actors[slot].id != 0 && g_actors[slot].state != AX_ACTOR_DEAD) {
        slot = (slot + 1) % AX_MAX_ACTORS;
        if (++attempts >= AX_MAX_ACTORS) return AX_ACTOR_ID_NONE;
    }

    AxActor* actor = &g_actors[slot];
    memset(actor, 0, sizeof(AxActor));
    actor->id = id;
    actor->state = AX_ACTOR_SPAWNING;
    actor->handler = handler;
    ax_msgq_init(&actor->mailbox);

    /* Copy init data if provided */
    if (init_data && data_size > 0) {
        actor->state_data = malloc(data_size);
        if (actor->state_data) {
            memcpy(actor->state_data, init_data, data_size);
            actor->state_size = data_size;
        }
    }

    actor->restart.policy = AX_RESTART_NEVER;
    actor->restart.max_restarts = 3;
    actor->restart.window_ms = 5000;

    /* Transition to RUNNING */
    actor->state = AX_ACTOR_RUNNING;
    if (actor->init_fn) {
        actor->init_fn(actor);
    }

    g_actor_count++;

    /* Send initial startup message to trigger handler execution */
    ax_actor_send(id, AX_ACTOR_ID_NONE, AX_MSG_USER, init_data, data_size);

    /* Submit to the global work-stealing scheduler */
    struct AxScheduler;
    extern struct AxScheduler* ax_get_global_scheduler(void);
    struct AxScheduler* sched = ax_get_global_scheduler();
    if (sched) {
        extern int ax_scheduler_submit(struct AxScheduler* sched, AxActorID actor_id);
        ax_scheduler_submit(sched, id);
    }

    return id;
}

/* --------------------------------------------------------------------------
 * Send
 * -------------------------------------------------------------------------- */

int ax_actor_send(AxActorID target, AxActorID sender,
                  AxMsgType type, const void* payload, uint32_t size) {
    AxActor* actor = ax_actor_lookup(target);
    if (!actor) return -1;

    /* Allocate message + payload inline */
    AxMessage* msg = (AxMessage*)malloc(sizeof(AxMessage) + size);
    if (!msg) return -1;

    msg->next = NULL;
    msg->sender = sender;
    msg->type = type;
    msg->size = size;

    if (payload && size > 0) {
        memcpy(ax_msg_payload(msg), payload, size);
    }

    ax_msgq_push(&actor->mailbox, msg);
    return 0;
}

/* --------------------------------------------------------------------------
 * Step (process one message)
 * -------------------------------------------------------------------------- */

int ax_actor_step(AxActor* actor) {
    if (!actor || actor->state != AX_ACTOR_RUNNING) return 0;

    AxMessage* msg = ax_msgq_pop(&actor->mailbox);
    if (!msg) return 0;

    /* Handle system messages */
    if (msg->type == AX_MSG_STOP) {
        actor->state = AX_ACTOR_STOPPING;
        if (actor->stop_fn) {
            actor->stop_fn(actor);
        }
        actor->state = AX_ACTOR_DEAD;
        g_actor_count--;
        free(msg);
        return 1;
    }

    /* Dispatch to user handler */
    actor->handler(actor, ax_msg_payload(msg), msg->type);
    actor->msgs_processed++;

    free(msg);
    return 1;
}

/* --------------------------------------------------------------------------
 * Stop
 * -------------------------------------------------------------------------- */

void ax_actor_stop(AxActorID id) {
    ax_actor_send(id, AX_ACTOR_ID_NONE, AX_MSG_STOP, NULL, 0);
}

int ax_actor_is_running(AxActor* actor) {
    return actor && actor->state == AX_ACTOR_RUNNING;
}

int ax_actor_has_messages(AxActor* actor) {
    return actor && actor->mailbox.head != NULL;
}

