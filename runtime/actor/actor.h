/*
 * p15-t01: AxActor — Core Actor Struct and Lifecycle
 *
 * Defines the fundamental actor structure for AXIOM's concurrency model.
 * Each actor has: isolated heap, mailbox, state machine, handler function.
 */

#ifndef AXIOM_RUNTIME_ACTOR_H
#define AXIOM_RUNTIME_ACTOR_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/* --------------------------------------------------------------------------
 * Actor ID
 * -------------------------------------------------------------------------- */

typedef uint64_t AxActorID;

#define AX_ACTOR_ID_NONE 0
#define AX_MAX_ACTORS    65536

/* --------------------------------------------------------------------------
 * Actor State Machine
 * -------------------------------------------------------------------------- */

typedef enum {
    AX_ACTOR_SPAWNING = 0,
    AX_ACTOR_RUNNING  = 1,
    AX_ACTOR_STOPPING = 2,
    AX_ACTOR_DEAD     = 3,
} AxActorState;

/* --------------------------------------------------------------------------
 * Message
 * -------------------------------------------------------------------------- */

typedef uint32_t AxMsgType;

#define AX_MSG_USER     0x0000  /* user-defined message */
#define AX_MSG_STOP     0x0001  /* stop request */
#define AX_MSG_LINK     0x0002  /* link notification */
#define AX_MSG_EXIT     0x0003  /* linked actor exited */
#define AX_MSG_TIMEOUT  0x0004  /* timeout fired */

typedef struct AxMessage {
    struct AxMessage* next;     /* intrusive linked list */
    AxActorID         sender;   /* sender actor ID */
    AxMsgType         type;     /* message type tag */
    uint32_t          size;     /* payload size in bytes */
    /* payload follows inline after this header */
} AxMessage;

/** Get the payload pointer from a message. */
static inline void* ax_msg_payload(AxMessage* msg) {
    return (char*)msg + sizeof(AxMessage);
}

/* --------------------------------------------------------------------------
 * Message Queue (MPSC: multiple producers, single consumer)
 * -------------------------------------------------------------------------- */

typedef struct {
    AxMessage* head;       /* consumer reads from head */
    AxMessage* tail;       /* producers append to tail */
    uint64_t   msg_count;  /* total messages received */
    uint64_t   pending;    /* current pending count */
} AxMsgQueue;

/** Initialize a message queue. */
void ax_msgq_init(AxMsgQueue* q);

/** Push a message (thread-safe for producers). */
void ax_msgq_push(AxMsgQueue* q, AxMessage* msg);

/** Pop a message (single consumer only). */
AxMessage* ax_msgq_pop(AxMsgQueue* q);

/** Check if the queue is empty. */
int ax_msgq_empty(const AxMsgQueue* q);

/* --------------------------------------------------------------------------
 * Handler Function
 * -------------------------------------------------------------------------- */

struct AxActor; /* forward declaration */

/** Actor message handler. Called once per message. */
typedef void (*AxHandlerFn)(struct AxActor* self, void* payload,
                            AxMsgType msg_type);

/** Actor init callback. Called once when actor starts. */
typedef void (*AxInitFn)(struct AxActor* self);

/** Actor stop callback. Called once when actor stops. */
typedef void (*AxStopFn)(struct AxActor* self);

/* --------------------------------------------------------------------------
 * Restart Policy
 * -------------------------------------------------------------------------- */

typedef enum {
    AX_RESTART_NEVER     = 0,  /* don't restart on crash */
    AX_RESTART_ALWAYS    = 1,  /* always restart */
    AX_RESTART_ON_CRASH  = 2,  /* restart only on crash, not normal exit */
} AxRestartPolicy;

typedef struct {
    AxRestartPolicy policy;
    uint32_t        max_restarts;    /* max restarts in window */
    uint32_t        window_ms;       /* time window for max_restarts */
    uint32_t        restart_count;   /* current restart count */
} AxRestartConfig;

/* --------------------------------------------------------------------------
 * Actor Struct
 * -------------------------------------------------------------------------- */

typedef struct AxActor {
    AxActorID       id;
    AxActorState    state;
    AxMsgQueue      mailbox;
    AxHandlerFn     handler;
    AxInitFn        init_fn;
    AxStopFn        stop_fn;
    void*           state_data;        /* user-defined state */
    size_t          state_size;
    AxActorID       supervisor_id;     /* supervising actor */
    AxRestartConfig restart;
    uint64_t        msgs_processed;
    uint32_t        flags;
} AxActor;

/* Actor flags */
#define AX_ACTOR_FLAG_SYSTEM  (1 << 0)  /* system actor (not user-created) */
#define AX_ACTOR_FLAG_LINKED  (1 << 1)  /* linked to supervisor */

/* --------------------------------------------------------------------------
 * Actor Lifecycle API
 * -------------------------------------------------------------------------- */

/**
 * Spawn a new actor. Returns the actor ID, or AX_ACTOR_ID_NONE on failure.
 */
AxActorID ax_actor_spawn(AxHandlerFn handler, void* init_data,
                         size_t data_size);

/**
 * Send a message to an actor. Returns 0 on success, -1 on failure.
 */
int ax_actor_send(AxActorID target, AxActorID sender,
                  AxMsgType type, const void* payload, uint32_t size);

/**
 * Request an actor to stop gracefully.
 */
void ax_actor_stop(AxActorID id);

/**
 * Process one message from the actor's mailbox.
 * Returns 1 if a message was processed, 0 if mailbox empty.
 */
int ax_actor_step(AxActor* actor);

/**
 * Look up an actor by ID. Returns NULL if not found.
 */
AxActor* ax_actor_lookup(AxActorID id);

/**
 * Initialize the global actor table. Call once at runtime init.
 */
void ax_actor_table_init(void);

/**
 * Destroy all actors and free the actor table.
 */
void ax_actor_table_destroy(void);

/**
 * Get the number of live actors.
 */
uint32_t ax_actor_count(void);

/**
 * Check if the actor is in AX_ACTOR_RUNNING state.
 */
int ax_actor_is_running(AxActor* actor);

/**
 * Check if the actor has pending messages.
 */
int ax_actor_has_messages(AxActor* actor);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_ACTOR_H */
