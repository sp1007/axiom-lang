/*
 * p15-t07: Supervisor Tree — Implementation
 */

#include "supervisor.h"
#include "async.h"
#include <stdlib.h>
#include <string.h>

/* --------------------------------------------------------------------------
 * Internal supervisor state storage
 * -------------------------------------------------------------------------- */

#define MAX_SUPERVISORS 256
static AxSupervisor g_supervisors[MAX_SUPERVISORS];
static int g_supervisor_count = 0;

/* Supervisor message handler */
static void supervisor_handler(AxActor* self, void* payload, AxMsgType type) {
    AxSupervisor* sup = ax_supervisor_lookup(self->id);
    if (!sup) return;

    if (type == AX_MSG_EXIT) {
        /* Child died — apply restart strategy */
        AxActorID child_id = *(AxActorID*)payload;
        ax_supervisor_handle_child_exit(sup, child_id, 1);
    }
}

/* --------------------------------------------------------------------------
 * Create
 * -------------------------------------------------------------------------- */

AxActorID ax_supervisor_create(AxSupStrategy strategy,
                               uint32_t max_restarts,
                               uint32_t window_ms) {
    if (g_supervisor_count >= MAX_SUPERVISORS) return AX_ACTOR_ID_NONE;

    AxActorID id = ax_actor_spawn(supervisor_handler, NULL, 0);
    if (id == AX_ACTOR_ID_NONE) return AX_ACTOR_ID_NONE;

    AxSupervisor* sup = &g_supervisors[g_supervisor_count++];
    memset(sup, 0, sizeof(AxSupervisor));
    sup->id = id;
    sup->strategy = strategy;
    sup->restart_intensity = max_restarts;
    sup->restart_window_ms = window_ms;
    sup->window_start_ns = ax_time_now_ns();

    return id;
}

/* --------------------------------------------------------------------------
 * Child Management
 * -------------------------------------------------------------------------- */

int ax_supervisor_add_child(AxSupervisor* sup, const AxChildSpec* spec) {
    if (!sup || !spec || sup->child_count >= AX_MAX_CHILDREN) return -1;

    uint32_t idx = sup->child_count;
    sup->specs[idx] = *spec;
    sup->children[idx] = AX_ACTOR_ID_NONE;
    sup->child_count++;
    return 0;
}

int ax_supervisor_start_children(AxSupervisor* sup) {
    if (!sup) return -1;

    for (uint32_t i = 0; i < sup->child_count; i++) {
        AxChildSpec* spec = &sup->specs[i];
        AxActorID child = ax_actor_spawn(spec->handler, spec->init_data,
                                          spec->data_size);
        if (child == AX_ACTOR_ID_NONE) return -1;

        sup->children[i] = child;

        /* Set the child's supervisor */
        AxActor* actor = ax_actor_lookup(child);
        if (actor) {
            actor->supervisor_id = sup->id;
            actor->flags |= AX_ACTOR_FLAG_LINKED;
            actor->restart.policy = spec->restart;
            actor->restart.max_restarts = spec->max_restarts;
            actor->restart.window_ms = spec->window_ms;
        }
    }
    return 0;
}

/* --------------------------------------------------------------------------
 * Child Exit Handling
 * -------------------------------------------------------------------------- */

static int find_child_index(AxSupervisor* sup, AxActorID child_id) {
    for (uint32_t i = 0; i < sup->child_count; i++) {
        if (sup->children[i] == child_id) return (int)i;
    }
    return -1;
}

static void restart_child(AxSupervisor* sup, uint32_t idx) {
    AxChildSpec* spec = &sup->specs[idx];
    AxActorID child = ax_actor_spawn(spec->handler, spec->init_data,
                                      spec->data_size);
    sup->children[idx] = child;

    AxActor* actor = ax_actor_lookup(child);
    if (actor) {
        actor->supervisor_id = sup->id;
        actor->flags |= AX_ACTOR_FLAG_LINKED;
    }
}

void ax_supervisor_handle_child_exit(AxSupervisor* sup, AxActorID child_id,
                                     int exit_reason) {
    if (!sup) return;
    (void)exit_reason;

    /* Check restart intensity */
    uint64_t now = ax_time_now_ns();
    uint64_t window_ns = (uint64_t)sup->restart_window_ms * 1000000ULL;
    if (now - sup->window_start_ns > window_ns) {
        sup->current_restarts = 0;
        sup->window_start_ns = now;
    }

    sup->current_restarts++;
    if (sup->current_restarts > sup->restart_intensity) {
        /* Too many restarts — escalate to parent supervisor */
        return;
    }

    int idx = find_child_index(sup, child_id);
    if (idx < 0) return;

    switch (sup->strategy) {
    case AX_STRATEGY_ONE_FOR_ONE:
        restart_child(sup, (uint32_t)idx);
        break;

    case AX_STRATEGY_ONE_FOR_ALL:
        for (uint32_t i = 0; i < sup->child_count; i++) {
            if (sup->children[i] != AX_ACTOR_ID_NONE) {
                ax_actor_stop(sup->children[i]);
            }
            restart_child(sup, i);
        }
        break;

    case AX_STRATEGY_REST_FOR_ONE:
        for (uint32_t i = (uint32_t)idx; i < sup->child_count; i++) {
            if (sup->children[i] != AX_ACTOR_ID_NONE) {
                ax_actor_stop(sup->children[i]);
            }
            restart_child(sup, i);
        }
        break;
    }
}

/* --------------------------------------------------------------------------
 * Shutdown
 * -------------------------------------------------------------------------- */

void ax_supervisor_shutdown(AxSupervisor* sup) {
    if (!sup) return;
    for (uint32_t i = 0; i < sup->child_count; i++) {
        if (sup->children[i] != AX_ACTOR_ID_NONE) {
            ax_actor_stop(sup->children[i]);
        }
    }
    ax_actor_stop(sup->id);
}

AxSupervisor* ax_supervisor_lookup(AxActorID id) {
    for (int i = 0; i < g_supervisor_count; i++) {
        if (g_supervisors[i].id == id) return &g_supervisors[i];
    }
    return NULL;
}
