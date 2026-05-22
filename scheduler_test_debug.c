#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_runq;
struct ax_worker;
struct ax_SchedulerStats;
struct ax_scheduler;

/* Type definitions */
struct ax_runq {
    ax_u64 buffer[4096];
    ax_u64 top;
    ax_u64 bottom;
};
struct ax_worker {
    ax_u32 id;
    struct ax_runq runq;
    ax_u64 tasks_executed;
    ax_u64 steals_attempted;
    ax_u64 steals_succeeded;
    ax_i32 running;
};
struct ax_SchedulerStats {
    ax_u32 worker_count;
    ax_u64 total_submitted;
    ax_u64 total_executed;
    ax_u64 total_steals;
};
struct ax_scheduler {
    struct ax_worker workers[256];
    ax_u32 worker_count;
    ax_i32 running;
    ax_u64 total_submitted;
};

/* Function prototypes */
void* ax_actor_lookup(ax_u64 id);
ax_i32 ax_actor_step(void* actor);
ax_i32 ax_actor_is_running(void* actor);
ax_i32 ax_actor_has_messages(void* actor);
void ax_runq_init(struct ax_runq* self);
ax_i32 ax_runq_push(struct ax_runq* self, ax_u64 id);
ax_u64 ax_runq_pop(struct ax_runq* self);
ax_u64 ax_runq_steal(struct ax_runq* self);
ax_i32 ax_runq_empty(const struct ax_runq* self);
ax_i32 ax_scheduler_init(struct ax_scheduler* self, ax_u32 worker_count);
ax_i32 ax_scheduler_submit(struct ax_scheduler* self, ax_u64 actor_id);
ax_i32 ax_scheduler_run(struct ax_scheduler* self);
static void ax_scheduler_worker_loop(struct ax_scheduler* self, struct ax_worker* w);
void ax_scheduler_shutdown(struct ax_scheduler* self);
void ax_scheduler_stats(const struct ax_scheduler* self, struct ax_SchedulerStats* stats);
static void ax_test_runq_basic(void);
static void ax_test_runq_steal(void);
static void ax_test_scheduler_lifecycle(void);
ax_i32 ax_main_usr(void);


void ax_runq_init(struct ax_runq* self) {
    void* p = ((void*)(&(self->buffer)));
    memset(p, ((ax_i32)(0)), (((ax_u64)(4096)) * ((ax_u64)(8))));
    self->top = ((ax_u64)(0));
    self->bottom = ((ax_u64)(0));
}

ax_i32 ax_runq_push(struct ax_runq* self, ax_u64 id) {
    ax_u64 b = self->bottom;
    ax_u64 t = self->top;
    if (((b - t) >= ((ax_u64)(4096)))) {
        return (-((ax_i32)(1)));
    }
    ax_bounds_check((ax_u64)(((ax_i64)((b % ((ax_u64)(4096)))))), (ax_u64)(4096));
    ((self->buffer)[((ax_i64)((b % ((ax_u64)(4096)))))]) = id;
    self->bottom = (b + ((ax_u64)(1)));
    return ((ax_i32)(0));
}

ax_u64 ax_runq_pop(struct ax_runq* self) {
    ax_u64 b = self->bottom;
    if ((b == ((ax_u64)(0)))) {
        return ((ax_u64)(0));
    }
    b = (b - ((ax_u64)(1)));
    self->bottom = b;
    ax_u64 t = self->top;
    if ((t <= b)) {
        return (ax_bounds_check((ax_u64)(((ax_i64)((b % ((ax_u64)(4096)))))), (ax_u64)(4096)), (self->buffer)[((ax_i64)((b % ((ax_u64)(4096)))))]);
    }
    if ((t == b)) {
        self->bottom = (t + ((ax_u64)(1)));
        self->top = (t + ((ax_u64)(1)));
        return (ax_bounds_check((ax_u64)(((ax_i64)((b % ((ax_u64)(4096)))))), (ax_u64)(4096)), (self->buffer)[((ax_i64)((b % ((ax_u64)(4096)))))]);
    }
    self->bottom = t;
    return ((ax_u64)(0));
}

ax_u64 ax_runq_steal(struct ax_runq* self) {
    ax_u64 t = self->top;
    ax_u64 b = self->bottom;
    if ((t >= b)) {
        return ((ax_u64)(0));
    }
    ax_u64 id = (ax_bounds_check((ax_u64)(((ax_i64)((t % ((ax_u64)(4096)))))), (ax_u64)(4096)), (self->buffer)[((ax_i64)((t % ((ax_u64)(4096)))))]);
    self->top = (t + ((ax_u64)(1)));
    return id;
}

ax_i32 ax_runq_empty(const struct ax_runq* self) {
    if ((self->top >= self->bottom)) {
        return ((ax_i32)(1));
    }
    return ((ax_i32)(0));
}

ax_i32 ax_scheduler_init(struct ax_scheduler* self, ax_u32 worker_count) {
    if (((worker_count == ((ax_u32)(0))) || (worker_count > ((ax_u32)(256))))) {
        return (-((ax_i32)(1)));
    }
    void* p = ((void*)(self));
    ax_u64 sz = sizeof(struct ax_scheduler);
    memset(p, ((ax_i32)(0)), ((ax_u64)(sz)));
    self->worker_count = worker_count;
    self->running = ((ax_i32)(0));
    ax_u32 i = ((ax_u32)(0));
    while ((i < worker_count)) {
        ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256));
        ((self->workers)[((ax_i64)(i))]).id = i;
        ax_runq_init(((ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256)), &(((self->workers)[((ax_i64)(i))]).runq))));
        ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256));
        ((self->workers)[((ax_i64)(i))]).running = ((ax_i32)(0));
        i = (i + ((ax_u32)(1)));
    }
    return ((ax_i32)(0));
}

ax_i32 ax_scheduler_submit(struct ax_scheduler* self, ax_u64 actor_id) {
    if ((actor_id == ((ax_u64)(0)))) {
        return (-((ax_i32)(1)));
    }
    ax_u32 target = ((ax_u32)((self->total_submitted % ((ax_u64)(self->worker_count)))));
    ax_i32 res = ax_runq_push(((ax_bounds_check((ax_u64)(((ax_i64)(target))), (ax_u64)(256)), &(((self->workers)[((ax_i64)(target))]).runq))), actor_id);
    if ((res == ((ax_i32)(0)))) {
        self->total_submitted = (self->total_submitted + ((ax_u64)(1)));
    }
    return res;
}

ax_i32 ax_scheduler_run(struct ax_scheduler* self) {
    self->running = ((ax_i32)(1));
    ax_u32 i = ((ax_u32)(0));
    while ((i < self->worker_count)) {
        ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256));
        ((self->workers)[((ax_i64)(i))]).running = ((ax_i32)(1));
        i = (i + ((ax_u32)(1)));
    }
    ax_u32 j = ((ax_u32)(0));
    while ((j < self->worker_count)) {
        ax_scheduler_worker_loop(self, ((ax_bounds_check((ax_u64)(((ax_i64)(j))), (ax_u64)(256)), &(((self->workers)[((ax_i64)(j))])))));
        j = (j + ((ax_u32)(1)));
    }
    return ((ax_i32)(0));
}

static void ax_scheduler_worker_loop(struct ax_scheduler* self, struct ax_worker* w) {
    while ((w->running != ((ax_i32)(0)))) {
        ax_u64 id = ax_runq_pop(&(w->runq));
        if ((id == ((ax_u64)(0)))) {
            ax_u32 i = ((ax_u32)(0));
            ax_i32 stop_loop = ((ax_i32)(0));
            while (((i < self->worker_count) && (stop_loop == ((ax_i32)(0))))) {
                if ((i != w->id)) {
                    w->steals_attempted = (w->steals_attempted + ((ax_u64)(1)));
                    ax_u64 stolen_id = ax_runq_steal(((ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256)), &(((self->workers)[((ax_i64)(i))]).runq))));
                    if ((stolen_id != ((ax_u64)(0)))) {
                        id = stolen_id;
                        w->steals_succeeded = (w->steals_succeeded + ((ax_u64)(1)));
                        stop_loop = ((ax_i32)(1));
                    }
                }
                i = (i + ((ax_u32)(1)));
            }
        }
        if ((id == ((ax_u64)(0)))) {
            return;
        }
        void* actor_ptr = ax_actor_lookup(id);
        if ((actor_ptr != ((void*)(NULL)))) {
            while ((ax_actor_step(actor_ptr) != ((ax_i32)(0)))) {
                w->tasks_executed = (w->tasks_executed + ((ax_u64)(1)));
            }
            if (((ax_actor_is_running(actor_ptr) != ((ax_i32)(0))) && (ax_actor_has_messages(actor_ptr) != ((ax_i32)(0))))) {
                ax_runq_push(&(w->runq), id);
            }
        }
    }
}

void ax_scheduler_shutdown(struct ax_scheduler* self) {
    self->running = ((ax_i32)(0));
    ax_u32 i = ((ax_u32)(0));
    while ((i < self->worker_count)) {
        ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256));
        ((self->workers)[((ax_i64)(i))]).running = ((ax_i32)(0));
        i = (i + ((ax_u32)(1)));
    }
}

void ax_scheduler_stats(const struct ax_scheduler* self, struct ax_SchedulerStats* stats) {
    if ((stats == ((struct ax_SchedulerStats*)(NULL)))) {
        return;
    }
    stats->worker_count = self->worker_count;
    stats->total_submitted = self->total_submitted;
    stats->total_executed = ((ax_u64)(0));
    stats->total_steals = ((ax_u64)(0));
    ax_u32 i = ((ax_u32)(0));
    while ((i < self->worker_count)) {
        stats->total_executed = (stats->total_executed + (ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256)), (self->workers)[((ax_i64)(i))]).tasks_executed);
        stats->total_steals = (stats->total_steals + (ax_bounds_check((ax_u64)(((ax_i64)(i))), (ax_u64)(256)), (self->workers)[((ax_i64)(i))]).steals_succeeded);
        i = (i + ((ax_u32)(1)));
    }
}

static void ax_test_runq_basic(void) {
    struct ax_runq q = ((struct ax_runq){.top=0, .bottom=0});
    ax_runq_init(&(q));
    ax_assert_axiom((ax_runq_empty(&(q)) == 1), AX_STR("(ax_runq_empty(&(q)) == 1)"));
    ax_assert_axiom((ax_runq_push(&(q), ((ax_u64)(101))) == 0), AX_STR("(ax_runq_push(&(q), ((ax_u64)(101))) == 0)"));
    ax_assert_axiom((ax_runq_push(&(q), ((ax_u64)(102))) == 0), AX_STR("(ax_runq_push(&(q), ((ax_u64)(102))) == 0)"));
    ax_assert_axiom((ax_runq_push(&(q), ((ax_u64)(103))) == 0), AX_STR("(ax_runq_push(&(q), ((ax_u64)(103))) == 0)"));
    ax_assert_axiom((ax_runq_empty(&(q)) == 0), AX_STR("(ax_runq_empty(&(q)) == 0)"));
    ax_assert_axiom((ax_runq_pop(&(q)) == ((ax_u64)(103))), AX_STR("(ax_runq_pop(&(q)) == ((ax_u64)(103)))"));
    ax_assert_axiom((ax_runq_pop(&(q)) == ((ax_u64)(102))), AX_STR("(ax_runq_pop(&(q)) == ((ax_u64)(102)))"));
    ax_assert_axiom((ax_runq_pop(&(q)) == ((ax_u64)(101))), AX_STR("(ax_runq_pop(&(q)) == ((ax_u64)(101)))"));
    ax_assert_axiom((ax_runq_empty(&(q)) == 1), AX_STR("(ax_runq_empty(&(q)) == 1)"));
    ax_assert_axiom((ax_runq_pop(&(q)) == ((ax_u64)(0))), AX_STR("(ax_runq_pop(&(q)) == ((ax_u64)(0)))"));
    puts((const char*)((ax_string){.ptr=(const ax_u8*)"  PASS: test_runq_basic", .len=23}).ptr);
}

static void ax_test_runq_steal(void) {
    struct ax_runq q = ((struct ax_runq){.top=0, .bottom=0});
    ax_runq_init(&(q));
    ax_assert_axiom((ax_runq_push(&(q), ((ax_u64)(201))) == 0), AX_STR("(ax_runq_push(&(q), ((ax_u64)(201))) == 0)"));
    ax_assert_axiom((ax_runq_push(&(q), ((ax_u64)(202))) == 0), AX_STR("(ax_runq_push(&(q), ((ax_u64)(202))) == 0)"));
    ax_assert_axiom((ax_runq_push(&(q), ((ax_u64)(203))) == 0), AX_STR("(ax_runq_push(&(q), ((ax_u64)(203))) == 0)"));
    ax_assert_axiom((ax_runq_steal(&(q)) == ((ax_u64)(201))), AX_STR("(ax_runq_steal(&(q)) == ((ax_u64)(201)))"));
    ax_assert_axiom((ax_runq_steal(&(q)) == ((ax_u64)(202))), AX_STR("(ax_runq_steal(&(q)) == ((ax_u64)(202)))"));
    ax_assert_axiom((ax_runq_pop(&(q)) == ((ax_u64)(203))), AX_STR("(ax_runq_pop(&(q)) == ((ax_u64)(203)))"));
    ax_assert_axiom((ax_runq_empty(&(q)) == 1), AX_STR("(ax_runq_empty(&(q)) == 1)"));
    puts((const char*)((ax_string){.ptr=(const ax_u8*)"  PASS: test_runq_steal", .len=23}).ptr);
}

static void ax_test_scheduler_lifecycle(void) {
    ax_u64 sz = sizeof(struct ax_scheduler);
    struct ax_scheduler* sched = ((struct ax_scheduler*)(malloc(sz)));
    ax_assert_axiom((sched != ((struct ax_scheduler*)(NULL))), AX_STR("(sched != ((struct ax_scheduler*)(NULL)))"));
    ax_assert_axiom((ax_scheduler_init(sched, ((ax_u32)(4))) == 0), AX_STR("(ax_scheduler_init(sched, ((ax_u32)(4))) == 0)"));
    ax_assert_axiom((sched->worker_count == ((ax_u32)(4))), AX_STR("(sched->worker_count == ((ax_u32)(4)))"));
    ax_assert_axiom((sched->running == 0), AX_STR("(sched->running == 0)"));
    ax_u64 i = ((ax_u64)(1));
    while ((i <= ((ax_u64)(8)))) {
        ax_assert_axiom((ax_scheduler_submit(sched, i) == 0), AX_STR("(ax_scheduler_submit(sched, i) == 0)"));
        i = (i + ((ax_u64)(1)));
    }
    ax_assert_axiom((sched->total_submitted == ((ax_u64)(8))), AX_STR("(sched->total_submitted == ((ax_u64)(8)))"));
    ax_assert_axiom((ax_runq_empty(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == 0), AX_STR("(ax_runq_empty(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == 0)"));
    ax_assert_axiom((ax_runq_empty(((ax_bounds_check((ax_u64)(3), (ax_u64)(256)), &(((sched->workers)[3]).runq)))) == 0), AX_STR("(ax_runq_empty(((ax_bounds_check((ax_u64)(3), (ax_u64)(256)), &(((sched->workers)[3]).runq)))) == 0)"));
    ax_assert_axiom((ax_runq_pop(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == ((ax_u64)(5))), AX_STR("(ax_runq_pop(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == ((ax_u64)(5)))"));
    ax_assert_axiom((ax_runq_pop(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == ((ax_u64)(1))), AX_STR("(ax_runq_pop(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == ((ax_u64)(1)))"));
    ax_assert_axiom((ax_runq_empty(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == 1), AX_STR("(ax_runq_empty(((ax_bounds_check((ax_u64)(0), (ax_u64)(256)), &(((sched->workers)[0]).runq)))) == 1)"));
    struct ax_SchedulerStats stats = ((struct ax_SchedulerStats){.worker_count=0, .total_submitted=0, .total_executed=0, .total_steals=0});
    ax_scheduler_stats(sched, &(stats));
    ax_assert_axiom((stats.worker_count == ((ax_u32)(4))), AX_STR("(stats.worker_count == ((ax_u32)(4)))"));
    ax_assert_axiom((stats.total_submitted == ((ax_u64)(8))), AX_STR("(stats.total_submitted == ((ax_u64)(8)))"));
    ax_assert_axiom((stats.total_executed == ((ax_u64)(0))), AX_STR("(stats.total_executed == ((ax_u64)(0)))"));
    ax_scheduler_shutdown(sched);
    free(((void*)(sched)));
    puts((const char*)((ax_string){.ptr=(const ax_u8*)"  PASS: test_scheduler_lifecycle", .len=32}).ptr);
}

ax_i32 ax_main_usr(void) {
    puts((const char*)((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native Scheduler unit tests...", .len=44}).ptr);
    ax_test_runq_basic();
    ax_test_runq_steal();
    ax_test_scheduler_lifecycle();
    return puts((const char*)((ax_string){.ptr=(const ax_u8*)"All AXIOM-native Scheduler tests passed!", .len=40}).ptr);
    return 0;
}

/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}
