/*
 * p15-t07: Supervisor Tree
 *
 * Erlang-inspired supervision strategies for actor lifecycle management.
 * Supervisors monitor child actors and restart them according to policy.
 */

#ifndef AXIOM_RUNTIME_SUPERVISOR_H
#define AXIOM_RUNTIME_SUPERVISOR_H

#include "actor.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define AX_MAX_CHILDREN 256

/* --------------------------------------------------------------------------
 * Supervision Strategy
 * -------------------------------------------------------------------------- */

typedef enum {
    AX_STRATEGY_ONE_FOR_ONE  = 0,  /* restart only the failed child */
    AX_STRATEGY_ONE_FOR_ALL  = 1,  /* restart all children if one fails */
    AX_STRATEGY_REST_FOR_ONE = 2,  /* restart failed child + all after it */
} AxSupStrategy;

/* --------------------------------------------------------------------------
 * Child Specification
 * -------------------------------------------------------------------------- */

typedef struct {
    AxHandlerFn     handler;
    void*           init_data;
    size_t          data_size;
    AxRestartPolicy restart;
    uint32_t        max_restarts;
    uint32_t        window_ms;
} AxChildSpec;

/* --------------------------------------------------------------------------
 * Supervisor
 * -------------------------------------------------------------------------- */

typedef struct {
    AxActorID       id;                        /* supervisor's own actor ID */
    AxSupStrategy   strategy;
    AxActorID       children[AX_MAX_CHILDREN]; /* child actor IDs */
    AxChildSpec     specs[AX_MAX_CHILDREN];    /* child specifications */
    uint32_t        child_count;
    uint32_t        restart_intensity;         /* max restarts in window */
    uint32_t        restart_window_ms;         /* time window */
    uint32_t        current_restarts;          /* restarts in current window */
    uint64_t        window_start_ns;           /* start of current window */
} AxSupervisor;

/** Create a new supervisor actor with the given strategy. */
AxActorID ax_supervisor_create(AxSupStrategy strategy,
                               uint32_t max_restarts,
                               uint32_t window_ms);

/** Add a child spec to the supervisor. */
int ax_supervisor_add_child(AxSupervisor* sup, const AxChildSpec* spec);

/** Start all children. */
int ax_supervisor_start_children(AxSupervisor* sup);

/** Handle a child exit notification. */
void ax_supervisor_handle_child_exit(AxSupervisor* sup, AxActorID child_id,
                                     int exit_reason);

/** Stop the supervisor and all its children. */
void ax_supervisor_shutdown(AxSupervisor* sup);

/** Look up a supervisor by its actor ID. */
AxSupervisor* ax_supervisor_lookup(AxActorID id);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_RUNTIME_SUPERVISOR_H */
