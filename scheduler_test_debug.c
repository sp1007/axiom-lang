#include <stddef.h>
int ax_actor_step(void* actor);
int ax_actor_is_running(void* actor);
int ax_actor_has_messages(void* actor);

int ax_actor_step_impl(void* actor_ptr) {
    return ax_actor_step(actor_ptr);
}
int ax_actor_is_running_impl(void* actor_ptr) {
    return ax_actor_is_running(actor_ptr);
}
int ax_actor_has_messages_impl(void* actor_ptr) {
    return ax_actor_has_messages(actor_ptr);
}

#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_AxGlobalState;
struct ax_Segment;
struct ax_SegmentList;
struct ax_FreeList;
struct ax_ActorHeap;
struct ax_FreeSlot;
struct ax_AxHeader;
struct ax_runq;
struct ax_worker;
struct ax_scheduler;
struct ax_AxActorTable;
struct ax_AxMsgQueue;
struct ax_AxRestartConfig;
struct ax_AxActor;
struct ax_AxMessage;
struct ax_SchedulerStats;

/* Type definitions */
struct ax_AxGlobalState {
    ax_bool is_rt_initialized;
    ax_bool g_sched_initialized;
    ax_i64 padding;
    ax_i64 g_slab_used;
    struct ax_Segment* g_free_pool;
    struct ax_Segment* g_slab;
    struct ax_ActorHeap* global_heap;
    void* g_actor_table;
    void* g_sched;
    ax_u32 oom_written;
    ax_u8 oom_digit;
};

struct ax_Segment {
    ax_u8* base;
    ax_u8* bump;
    ax_u8* limit;
    ax_i32 sclass;
    struct ax_Segment* next;
    ax_u32 magic;
};

struct ax_SegmentList {
    struct ax_Segment* active;
    struct ax_Segment* retired;
    ax_i64 count;
};

struct ax_FreeList {
    struct ax_FreeSlot* head;
    ax_i64 count;
};

struct ax_ActorHeap {
    ax_u64 actor_id;
    ax_u32 magic;
    ax_u32 padding;
    struct ax_SegmentList seg_0;
    struct ax_SegmentList seg_1;
    struct ax_SegmentList seg_2;
    struct ax_SegmentList seg_3;
    struct ax_SegmentList seg_4;
    struct ax_SegmentList seg_5;
    struct ax_SegmentList seg_6;
    struct ax_SegmentList seg_7;
    struct ax_SegmentList seg_8;
    struct ax_SegmentList seg_9;
    struct ax_FreeList free_0;
    struct ax_FreeList free_1;
    struct ax_FreeList free_2;
    struct ax_FreeList free_3;
    struct ax_FreeList free_4;
    struct ax_FreeList free_5;
    struct ax_FreeList free_6;
    struct ax_FreeList free_7;
    struct ax_FreeList free_8;
    struct ax_FreeList free_9;
    ax_u64 total_allocated;
    ax_u64 total_freed;
    ax_u64 alloc_count;
    ax_u64 free_count;
};

struct ax_FreeSlot {
    struct ax_FreeSlot* next;
};

struct ax_AxHeader {
    ax_u32 gen_id;
    ax_u32 flags;
};

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

struct ax_scheduler {
    struct ax_worker workers[256];
    ax_u32 worker_count;
    ax_i32 running;
    ax_u64 total_submitted;
};

struct ax_AxActorTable {
    struct ax_AxActor* actors;
    ax_u64 next_id;
    ax_u32 actor_count;
};

struct ax_AxMsgQueue {
    struct ax_AxMessage* head;
    struct ax_AxMessage* tail;
    ax_u64 msg_count;
    ax_u64 pending;
};

struct ax_AxRestartConfig {
    ax_i32 policy;
    ax_u32 max_restarts;
    ax_u32 window_ms;
    ax_u32 restart_count;
};

struct ax_AxActor {
    ax_u64 id;
    ax_i32 state;
    struct ax_AxMsgQueue mailbox;
    void* handler;
    void* init_fn;
    void* stop_fn;
    void* state_data;
    ax_u64 state_size;
    struct ax_ActorHeap* heap;
    ax_u64 supervisor_id;
    struct ax_AxRestartConfig restart;
    ax_u64 msgs_processed;
    ax_u32 flags;
};

struct ax_AxMessage {
    struct ax_AxMessage* next;
    ax_u64 sender;
    ax_u32 msg_type;
    ax_u32 size;
};

struct ax_SchedulerStats {
    ax_u32 worker_count;
    ax_u64 total_submitted;
    ax_u64 total_executed;
    ax_u64 total_steals;
};


/* Global variables */
extern const ax_i64 ax_SEGMENT_SIZE;
const ax_i64 ax_SEGMENT_SIZE = 65536;
extern const ax_i64 ax_MAX_SEGMENTS;
const ax_i64 ax_MAX_SEGMENTS = 4096;
extern const ax_u32 ax_SEGMENT_MAGIC;
const ax_u32 ax_SEGMENT_MAGIC = 0xAF5E6000;
extern const ax_u32 ax_ACTOR_HEAP_MAGIC;
const ax_u32 ax_ACTOR_HEAP_MAGIC = 0xAC704EA0;

/* Function prototypes */
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
void* ax_get_global_state_internal(void);
void* ax_mmap(void* addr, ax_u64 length, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset);
ax_i32 ax_munmap(void* addr, ax_u64 length);
static void ax_print_hex_digit(ax_u8 digit);
static void ax_print_hex_freestanding(ax_u32 val);
struct ax_AxGlobalState* ax_get_global_state(void);
ax_i64* ax_std_mem_alloc_get_slab_used(void);
struct ax_Segment** ax_std_mem_alloc_get_free_pool(void);
struct ax_Segment* ax_std_mem_alloc_get_slab(void);
ax_i64 ax_ax_size_class_size(ax_i32 sc);
ax_i32 ax_ax_size_class_for(ax_i64 user_size);
struct ax_SegmentList* ax_ActorHeap_ax_get_segment_list(struct ax_ActorHeap* heap, ax_i32 sc);
struct ax_FreeList* ax_ActorHeap_ax_get_free_list(struct ax_ActorHeap* heap, ax_i32 sc);
void ax_FreeList_ax_free_list_push(struct ax_FreeList* list, ax_u8* block);
ax_u8* ax_FreeList_ax_free_list_pop(struct ax_FreeList* list);
static void ax_ax_os_alloc_report_error(ax_u32 err);
void* ax_ax_os_alloc(ax_i64 size);
void ax_ax_os_free(void* ptr, ax_i64 size);
void ax_ax_segment_manager_init(void);
void ax_ax_segment_manager_shutdown(void);
static struct ax_Segment* ax_alloc_segment_meta(void);
static void ax_Segment_free_segment_meta(struct ax_Segment* seg);
struct ax_Segment* ax_ax_segment_acquire(ax_i32 sc);
void ax_Segment_ax_segment_release(struct ax_Segment* seg);
struct ax_Segment* ax_SegmentList_ax_segment_get_active(struct ax_SegmentList* list, ax_i32 sc);
void ax_SegmentList_ax_segment_list_release_all(struct ax_SegmentList* list);
ax_u8* ax_Segment_ax_segment_bump_alloc(struct ax_Segment* seg, ax_i32 sc);
ax_u8* ax_ax_large_alloc(ax_i64 user_size);
void ax_ax_large_free(ax_u8* user_ptr, ax_i64 user_size);
struct ax_ActorHeap* ax_ax_actor_heap_create(ax_u64 actor_id);
void ax_ActorHeap_ax_actor_heap_destroy(struct ax_ActorHeap* heap);
ax_u8* ax_ActorHeap_ax_actor_alloc(struct ax_ActorHeap* heap, ax_i64 user_size);
void ax_ActorHeap_ax_actor_free(struct ax_ActorHeap* heap, ax_u8* user_ptr);
ax_u8* ax_ax_numa_alloc(ax_i64 size, ax_i32 node_id);
void ax_runq_init(struct ax_runq* self);
ax_i32 ax_runq_push(struct ax_runq* self, ax_u64 id);
ax_u64 ax_runq_pop(struct ax_runq* self);
ax_u64 ax_runq_steal(struct ax_runq* self);
ax_i32 ax_runq_empty(const struct ax_runq* self);
ax_i32 ax_scheduler_init(struct ax_scheduler* self, ax_u32 worker_count);
ax_i32 ax_scheduler_submit(struct ax_scheduler* self, ax_u64 actor_id);
static struct ax_AxGlobalState* ax_get_state(void);
void ax_ax_actor_system_init(void);
void* ax_ax_actor_lookup(ax_u64 id);
void ax_AxMsgQueue_ax_msgq_init(struct ax_AxMsgQueue* q);
void ax_AxMsgQueue_ax_msgq_push(struct ax_AxMsgQueue* q, struct ax_AxMessage* msg);
struct ax_AxMessage* ax_AxMsgQueue_ax_msgq_pop(struct ax_AxMsgQueue* q);
ax_i32 ax_ax_actor_send(ax_u64 target, ax_u64 sender, ax_u32 msg_type, void* payload, ax_u32 size);
ax_u64 ax_ax_actor_spawn(void* handler, void* init_data, ax_u64 data_size);
ax_i32 ax_actor_step(void* actor_ptr);
ax_i32 ax_actor_is_running(void* actor_ptr);
ax_i32 ax_actor_has_messages(void* actor_ptr);
static ax_bool ax_has_active_actors(void);
ax_i32 ax_scheduler_run(struct ax_scheduler* self);
static void ax_scheduler_worker_loop(struct ax_scheduler* self, struct ax_worker* w);
void ax_scheduler_shutdown(struct ax_scheduler* self);
void ax_scheduler_stats(struct ax_scheduler* self, struct ax_SchedulerStats* stats);
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
static void ax_test_print_str(ax_string s);
static void ax_test_runq_basic(void);
static void ax_test_runq_steal(void);
static void ax_test_scheduler_lifecycle(void);
ax_i32 ax_main_usr(void);
ax_i64 ax_std_string_len(ax_string p0);
ax_i64 ax_std_string_char_count(ax_string p0);
ax_string ax_std_string_trim(ax_string p0);
ax_string ax_std_string_to_upper(ax_string p0);
ax_string ax_std_string_to_lower(ax_string p0);


void* ax_mmap(void* addr, ax_u64 length, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset) {
    ax_i64 res = syscall(((ax_u64)(9)), ((ax_u64)(addr)), length, ((ax_u64)(prot)), ((ax_u64)(flags)), ((ax_u64)(((ax_i64)(fd)))), ((ax_u64)(offset)));
    return ((void*)(res));
}

ax_i32 ax_munmap(void* addr, ax_u64 length) {
    ax_i64 res = syscall(((ax_u64)(11)), ((ax_u64)(addr)), length, ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
    return ((ax_i32)(res));
}

static void ax_print_hex_digit(ax_u8 digit) {
    void* h_err = GetStdHandle(((ax_u32)(0xFFFFFFF4)));
    ax_u8 c = ((ax_u8)('0'));
    if ((digit < ((ax_u8)(10)))) {
        c = (digit + ((ax_u8)('0')));
    } else {
        {
            c = ((digit - ((ax_u8)(10))) + ((ax_u8)('A')));
        }
    }
    ax_u32 written = ((ax_u32)(0));
    WriteFile(h_err, ((void*)(&(c))), ((ax_u32)(1)), ((void*)(((ax_u32*)(&(written))))), ((void*)(NULL)));
}

static void ax_print_hex_freestanding(ax_u32 val) {
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(28))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(24))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(20))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(16))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(12))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(8))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)(((val >> ((ax_u32)(4))) & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)((val & ((ax_u32)(15))))));
    ax_print_hex_digit(((ax_u8)('\n')));
}

struct ax_AxGlobalState* ax_get_global_state(void) {
    return ((struct ax_AxGlobalState*)(ax_get_global_state_internal()));
}

ax_i64* ax_std_mem_alloc_get_slab_used(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    return ((ax_i64*)((((ax_i64)(state)) + ((ax_i64)(16)))));
}

struct ax_Segment** ax_std_mem_alloc_get_free_pool(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    return ((struct ax_Segment**)((((ax_i64)(state)) + ((ax_i64)(24)))));
}

struct ax_Segment* ax_std_mem_alloc_get_slab(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    struct ax_Segment** p_g_slab = ((struct ax_Segment**)((((ax_i64)(state)) + ((ax_i64)(32)))));
    if (((*((struct ax_Segment**)(p_g_slab))) == ((struct ax_Segment*)(NULL)))) {
        if (1) {
            (*((struct ax_Segment**)(p_g_slab))) = ((struct ax_Segment*)(VirtualAlloc(((void*)(NULL)), ((ax_u64)(196608)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)))));
        }
    }
    return (*((struct ax_Segment**)(p_g_slab)));
}

ax_i64 ax_ax_size_class_size(ax_i32 sc) {
    if ((sc == 0)) {
        return 8;
    }
    if ((sc == 1)) {
        return 16;
    }
    if ((sc == 2)) {
        return 32;
    }
    if ((sc == 3)) {
        return 64;
    }
    if ((sc == 4)) {
        return 128;
    }
    if ((sc == 5)) {
        return 256;
    }
    if ((sc == 6)) {
        return 512;
    }
    if ((sc == 7)) {
        return 1024;
    }
    if ((sc == 8)) {
        return 2048;
    }
    if ((sc == 9)) {
        return 4096;
    }
    return 0;
}

ax_i32 ax_ax_size_class_for(ax_i64 user_size) {
    ax_i64 total = (user_size + 8);
    if ((total <= 8)) {
        return 0;
    }
    if ((total <= 16)) {
        return 1;
    }
    if ((total <= 32)) {
        return 2;
    }
    if ((total <= 64)) {
        return 3;
    }
    if ((total <= 128)) {
        return 4;
    }
    if ((total <= 256)) {
        return 5;
    }
    if ((total <= 512)) {
        return 6;
    }
    if ((total <= 1024)) {
        return 7;
    }
    if ((total <= 2048)) {
        return 8;
    }
    if ((total <= 4096)) {
        return 9;
    }
    return 10;
}

struct ax_SegmentList* ax_ActorHeap_ax_get_segment_list(struct ax_ActorHeap* heap, ax_i32 sc) {
    if ((sc == 0)) {
        return ((struct ax_SegmentList*)((((ax_i64)(heap)) + 16)));
    }
    if ((sc == 1)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 24)));
    }
    if ((sc == 2)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 48)));
    }
    if ((sc == 3)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 72)));
    }
    if ((sc == 4)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 96)));
    }
    if ((sc == 5)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 120)));
    }
    if ((sc == 6)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 144)));
    }
    if ((sc == 7)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 168)));
    }
    if ((sc == 8)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 192)));
    }
    if ((sc == 9)) {
        return ((struct ax_SegmentList*)(((((ax_i64)(heap)) + 16) + 216)));
    }
    return ((struct ax_SegmentList*)(NULL));
}

struct ax_FreeList* ax_ActorHeap_ax_get_free_list(struct ax_ActorHeap* heap, ax_i32 sc) {
    ax_i64 base = ((((ax_i64)(heap)) + 16) + 240);
    if ((sc == 0)) {
        return ((struct ax_FreeList*)(base));
    }
    if ((sc == 1)) {
        return ((struct ax_FreeList*)((base + 16)));
    }
    if ((sc == 2)) {
        return ((struct ax_FreeList*)((base + 32)));
    }
    if ((sc == 3)) {
        return ((struct ax_FreeList*)((base + 48)));
    }
    if ((sc == 4)) {
        return ((struct ax_FreeList*)((base + 64)));
    }
    if ((sc == 5)) {
        return ((struct ax_FreeList*)((base + 80)));
    }
    if ((sc == 6)) {
        return ((struct ax_FreeList*)((base + 96)));
    }
    if ((sc == 7)) {
        return ((struct ax_FreeList*)((base + 112)));
    }
    if ((sc == 8)) {
        return ((struct ax_FreeList*)((base + 128)));
    }
    if ((sc == 9)) {
        return ((struct ax_FreeList*)((base + 144)));
    }
    return ((struct ax_FreeList*)(NULL));
}

void ax_FreeList_ax_free_list_push(struct ax_FreeList* list, ax_u8* block) {
    struct ax_FreeSlot* slot = ((struct ax_FreeSlot*)((((ax_i64)(block)) + 8)));
    slot->next = list->head;
    list->head = slot;
    list->count = (list->count + 1);
}

ax_u8* ax_FreeList_ax_free_list_pop(struct ax_FreeList* list) {
    struct ax_FreeSlot* slot = list->head;
    if ((slot == ((struct ax_FreeSlot*)(NULL)))) {
        return ((ax_u8*)(NULL));
    }
    list->head = slot->next;
    list->count = (list->count - 1);
    return ((ax_u8*)((((ax_i64)(slot)) - 8)));
}

static void ax_ax_os_alloc_report_error(ax_u32 err) {
    void* h_err = GetStdHandle(((ax_u32)(0xFFFFFFF4)));
    ax_string msg = (ax_string){.ptr=(const ax_u8*)"ax_os_alloc FAILED! Error code: ", .len=32};
    ax_u32 written = ((ax_u32)(0));
    WriteFile(h_err, ((void*)(msg.ptr)), ((ax_u32)(ax_str_len(msg))), ((void*)(((ax_u32*)(&(written))))), ((void*)(NULL)));
    ax_print_hex_freestanding(err);
}

void* ax_ax_os_alloc(ax_i64 size) {
    if (1) {
        void* res = VirtualAlloc(((void*)(NULL)), ((ax_u64)(size)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)));
        if ((res == ((void*)(NULL)))) {
            ax_ax_os_alloc_report_error(GetLastError());
        }
        return res;
        ax_free(res);
    } else {
        {
            return ax_mmap(((void*)(NULL)), ((ax_u64)(size)), ((ax_i32)(3)), ((ax_i32)(0x22)), ((ax_i32)((-1))), ((ax_i64)(0)));
        }
    }
}

void ax_ax_os_free(void* ptr, ax_i64 size) {
    if ((ptr == ((void*)(NULL)))) {
        return;
    }
    if (1) {
        VirtualFree(ptr, ((ax_u64)(0)), ((ax_u32)(0x8000)));
    } else {
        {
            ax_munmap(ptr, ((ax_u64)(size)));
        }
    }
}

void ax_ax_segment_manager_init(void) {
    ax_i64* slab_used_ptr = ax_std_mem_alloc_get_slab_used();
    (*((ax_i64*)(slab_used_ptr))) = 0;
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    (*((struct ax_Segment**)(free_pool_ptr))) = ((struct ax_Segment*)(NULL));
    struct ax_Segment* slab_ptr = ax_std_mem_alloc_get_slab();
    memset(((ax_u8*)(slab_ptr)), ((ax_u8)(0)), ((ax_i64)(196608)));
}

void ax_ax_segment_manager_shutdown(void) {
    ax_i64* slab_used_ptr = ax_std_mem_alloc_get_slab_used();
    struct ax_Segment* slab_ptr = ax_std_mem_alloc_get_slab();
    ax_i64 i = ((ax_i64)(0));
    while ((i < (*((ax_i64*)(slab_used_ptr))))) {
        struct ax_Segment* seg = ((struct ax_Segment*)((((ax_i64)(slab_ptr)) + (i * ((ax_i64)(sizeof(struct ax_Segment)))))));
        if ((seg->magic == ax_SEGMENT_MAGIC)) {
            ax_ax_os_free(((void*)(seg->base)), ax_SEGMENT_SIZE);
            seg->magic = ((ax_u32)(0));
        }
        i = (i + 1);
    }
    (*((ax_i64*)(slab_used_ptr))) = 0;
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    (*((struct ax_Segment**)(free_pool_ptr))) = ((struct ax_Segment*)(NULL));
}

static struct ax_Segment* ax_alloc_segment_meta(void) {
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    if (((*((struct ax_Segment**)(free_pool_ptr))) != ((struct ax_Segment*)(NULL)))) {
        struct ax_Segment* seg = (*((struct ax_Segment**)(free_pool_ptr)));
        (*((struct ax_Segment**)(free_pool_ptr))) = seg->next;
        return seg;
        ax_free(seg);
    }
    ax_i64* slab_used_ptr = ax_std_mem_alloc_get_slab_used();
    if (((*((ax_i64*)(slab_used_ptr))) >= ((ax_i64)(4096)))) {
        return ((struct ax_Segment*)(NULL));
    }
    struct ax_Segment* slab_ptr = ax_std_mem_alloc_get_slab();
    struct ax_Segment* seg = ((struct ax_Segment*)((((ax_i64)(slab_ptr)) + ((*((ax_i64*)(slab_used_ptr))) * ((ax_i64)(sizeof(struct ax_Segment)))))));
    (*((ax_i64*)(slab_used_ptr))) = ((*((ax_i64*)(slab_used_ptr))) + 1);
    return seg;
    ax_free(seg);
}

static void ax_Segment_free_segment_meta(struct ax_Segment* seg) {
    memset(((ax_u8*)(seg)), ((ax_u8)(0)), sizeof(struct ax_Segment));
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    seg->next = (*((struct ax_Segment**)(free_pool_ptr)));
    (*((struct ax_Segment**)(free_pool_ptr))) = seg;
}

struct ax_Segment* ax_ax_segment_acquire(ax_i32 sc) {
    struct ax_Segment* seg = ax_alloc_segment_meta();
    if ((seg == ((struct ax_Segment*)(NULL)))) {
        return ((struct ax_Segment*)(NULL));
    }
    ax_u8* mem = ((ax_u8*)(ax_ax_os_alloc(ax_SEGMENT_SIZE)));
    if ((mem == ((ax_u8*)(NULL)))) {
        ax_Segment_free_segment_meta(seg);
        return ((struct ax_Segment*)(NULL));
    }
    seg->base = mem;
    seg->bump = mem;
    seg->limit = ((ax_u8*)((((ax_i64)(mem)) + ax_SEGMENT_SIZE)));
    seg->sclass = sc;
    seg->next = ((struct ax_Segment*)(NULL));
    seg->magic = ax_SEGMENT_MAGIC;
    return seg;
    ax_free(seg);
}

void ax_Segment_ax_segment_release(struct ax_Segment* seg) {
    if ((seg == ((struct ax_Segment*)(NULL)))) {
        return;
    }
    if ((seg->magic != ax_SEGMENT_MAGIC)) {
        return;
    }
    ax_ax_os_free(((void*)(seg->base)), ax_SEGMENT_SIZE);
    seg->magic = ((ax_u32)(0));
    ax_Segment_free_segment_meta(seg);
}

struct ax_Segment* ax_SegmentList_ax_segment_get_active(struct ax_SegmentList* list, ax_i32 sc) {
    ax_i64 block_size = ax_ax_size_class_size(sc);
    if ((list->active != ((struct ax_Segment*)(NULL)))) {
        if (((((ax_i64)(list->active->limit)) - ((ax_i64)(list->active->bump))) >= block_size)) {
            return list->active;
        }
    }
    if ((list->active != ((struct ax_Segment*)(NULL)))) {
        list->active->next = list->retired;
        list->retired = list->active;
        list->count = (list->count + 1);
    }
    struct ax_Segment* seg = ax_ax_segment_acquire(sc);
    list->active = seg;
    return seg;
    ax_free(seg);
}

void ax_SegmentList_ax_segment_list_release_all(struct ax_SegmentList* list) {
    if ((list->active != ((struct ax_Segment*)(NULL)))) {
        ax_Segment_ax_segment_release(list->active);
        list->active = ((struct ax_Segment*)(NULL));
    }
    struct ax_Segment* seg = list->retired;
    while ((seg != ((struct ax_Segment*)(NULL)))) {
        struct ax_Segment* next = seg->next;
        ax_Segment_ax_segment_release(seg);
        seg = next;
    }
    list->retired = ((struct ax_Segment*)(NULL));
    list->count = 0;
}

ax_u8* ax_Segment_ax_segment_bump_alloc(struct ax_Segment* seg, ax_i32 sc) {
    if ((seg == ((struct ax_Segment*)(NULL)))) {
        return ((ax_u8*)(NULL));
    }
    if ((sc >= 10)) {
        return ((ax_u8*)(NULL));
    }
    ax_i64 block_size = ax_ax_size_class_size(sc);
    if (((((ax_i64)(seg->bump)) + block_size) > ((ax_i64)(seg->limit)))) {
        return ((ax_u8*)(NULL));
    }
    ax_u8* block = seg->bump;
    seg->bump = ((ax_u8*)((((ax_i64)(seg->bump)) + block_size)));
    return block;
    ax_free(block);
}

ax_u8* ax_ax_large_alloc(ax_i64 user_size) {
    ax_i64 total = (16 + user_size);
    ax_i64 page_aligned = (((total + 4095) / 4096) * 4096);
    ax_u8* block = ((ax_u8*)(ax_ax_os_alloc(page_aligned)));
    if ((block == ((ax_u8*)(NULL)))) {
        return ((ax_u8*)(NULL));
    }
    ax_u64* p_total = ((ax_u64*)(block));
    (*((ax_u64*)(p_total))) = ((ax_u64)(page_aligned));
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)((((ax_i64)(block)) + 8)));
    hdr->gen_id = ((ax_u32)(1));
    hdr->flags = ((ax_u32)(10));
    return ((ax_u8*)((((ax_i64)(block)) + 16)));
}

void ax_ax_large_free(ax_u8* user_ptr, ax_i64 user_size) {
    if ((user_ptr == ((ax_u8*)(NULL)))) {
        return;
    }
    void* block = ((void*)((((ax_i64)(user_ptr)) - 16)));
    ax_u64* p_total = ((ax_u64*)((((ax_i64)(user_ptr)) - 16)));
    ax_i64 total = ((ax_i64)((*((ax_u64*)(p_total)))));
    ax_ax_os_free(block, total);
}

struct ax_ActorHeap* ax_ax_actor_heap_create(ax_u64 actor_id) {
    ax_i64 size = ((ax_i64)(sizeof(struct ax_ActorHeap)));
    ax_i64 page_aligned = (((size + 4095) / 4096) * 4096);
    struct ax_ActorHeap* heap = ((struct ax_ActorHeap*)(ax_ax_os_alloc(page_aligned)));
    if ((heap == ((struct ax_ActorHeap*)(NULL)))) {
        return ((struct ax_ActorHeap*)(NULL));
    }
    heap->actor_id = actor_id;
    heap->magic = ax_ACTOR_HEAP_MAGIC;
    return heap;
    ax_free(heap);
}

void ax_ActorHeap_ax_actor_heap_destroy(struct ax_ActorHeap* heap) {
    if ((heap == ((struct ax_ActorHeap*)(NULL)))) {
        return;
    }
    ax_i32 sc = 0;
    while ((sc < 10)) {
        struct ax_SegmentList* list = ax_ActorHeap_ax_get_segment_list(heap, sc);
        ax_SegmentList_ax_segment_list_release_all(list);
        sc = (sc + 1);
    }
    ax_i64 size = ((ax_i64)(sizeof(struct ax_ActorHeap)));
    ax_i64 page_aligned = (((size + 4095) / 4096) * 4096);
    ax_ax_os_free(((void*)(heap)), page_aligned);
}

ax_u8* ax_ActorHeap_ax_actor_alloc(struct ax_ActorHeap* heap, ax_i64 user_size) {
    ax_i64 sz = user_size;
    if ((sz < ((ax_i64)(1)))) {
        sz = ((ax_i64)(1));
    }
    if ((heap == ((struct ax_ActorHeap*)(NULL)))) {
        return ((ax_u8*)(NULL));
    }
    ax_i32 sc = ax_ax_size_class_for(sz);
    if ((sc == 10)) {
        ax_u8* ptr_val = ax_ax_large_alloc(sz);
        if ((ptr_val != ((ax_u8*)(NULL)))) {
            heap->total_allocated = (heap->total_allocated + ((ax_u64)(sz)));
            heap->alloc_count = (heap->alloc_count + ((ax_u64)(1)));
        }
        return ptr_val;
        ax_free(ptr_val);
    }
    struct ax_FreeList* free_list = ax_ActorHeap_ax_get_free_list(heap, sc);
    ax_u8* block = ax_FreeList_ax_free_list_pop(free_list);
    if ((block == ((ax_u8*)(NULL)))) {
        struct ax_SegmentList* seg_list = ax_ActorHeap_ax_get_segment_list(heap, sc);
        struct ax_Segment* seg = ax_SegmentList_ax_segment_get_active(seg_list, sc);
        if ((seg == ((struct ax_Segment*)(NULL)))) {
            return ((ax_u8*)(NULL));
        }
        block = ax_Segment_ax_segment_bump_alloc(seg, sc);
        if ((block == ((ax_u8*)(NULL)))) {
            seg = ax_SegmentList_ax_segment_get_active(seg_list, sc);
            if ((seg == ((struct ax_Segment*)(NULL)))) {
                return ((ax_u8*)(NULL));
            }
            block = ax_Segment_ax_segment_bump_alloc(seg, sc);
            if ((block == ((ax_u8*)(NULL)))) {
                return ((ax_u8*)(NULL));
            }
        }
    }
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)(block));
    hdr->gen_id = ((ax_u32)(1));
    hdr->flags = ((ax_u32)(sc));
    ax_i64 block_size = ax_ax_size_class_size(sc);
    heap->total_allocated = (heap->total_allocated + ((ax_u64)(block_size)));
    heap->alloc_count = (heap->alloc_count + ((ax_u64)(1)));
    return ((ax_u8*)((((ax_i64)(block)) + 8)));
}

void ax_ActorHeap_ax_actor_free(struct ax_ActorHeap* heap, ax_u8* user_ptr) {
    if ((heap == ((struct ax_ActorHeap*)(NULL)))) {
        return;
    }
    if ((user_ptr == ((ax_u8*)(NULL)))) {
        return;
    }
    ax_u8* block = ((ax_u8*)((((ax_i64)(user_ptr)) - 8)));
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)(block));
    ax_i32 sc = ((ax_i32)((hdr->flags & ((ax_u32)(15)))));
    hdr->gen_id = ((ax_u32)(0));
    if (((sc == 10) || (sc >= 10))) {
        ax_u64* p_total = ((ax_u64*)((((ax_i64)(user_ptr)) - 16)));
        ax_u64 user_size = ((*((ax_u64*)(p_total))) - ((ax_u64)(16)));
        heap->total_freed = (heap->total_freed + user_size);
        heap->free_count = (heap->free_count + ((ax_u64)(1)));
        ax_ax_large_free(user_ptr, ((ax_i64)(0)));
        return;
    }
    ax_i64 block_size = ax_ax_size_class_size(sc);
    heap->total_freed = (heap->total_freed + ((ax_u64)(block_size)));
    heap->free_count = (heap->free_count + ((ax_u64)(1)));
    struct ax_FreeList* free_list = ax_ActorHeap_ax_get_free_list(heap, sc);
    ax_FreeList_ax_free_list_push(free_list, block);
}

ax_u8* ax_ax_numa_alloc(ax_i64 size, ax_i32 node_id) {
    return ((ax_u8*)(ax_ax_os_alloc(size)));
}

void ax_runq_init(struct ax_runq* self) {
    self->top = ((ax_u64)(0));
    self->bottom = ((ax_u64)(0));
}

ax_i32 ax_runq_push(struct ax_runq* self, ax_u64 id) {
    ax_u64 b = self->bottom;
    ax_u64 t = __atomic_load_n(&(self->top), __ATOMIC_SEQ_CST);
    if (((b - t) >= ((ax_u64)(4096)))) {
        return ((ax_i32)((-1)));
    }
    ax_bounds_check((ax_u64)(((ax_i64)((b % ((ax_u64)(4096)))))), (ax_u64)(4096));
    ((self->buffer)[((ax_i64)((b % ((ax_u64)(4096)))))]) = id;
    __atomic_store_n(&(self->bottom), (b + ((ax_u64)(1))), __ATOMIC_SEQ_CST);
    return ((ax_i32)(0));
}

ax_u64 ax_runq_pop(struct ax_runq* self) {
    ax_u64 b = self->bottom;
    if ((b == ((ax_u64)(0)))) {
        return ((ax_u64)(0));
    }
    b = (b - ((ax_u64)(1)));
    __atomic_store_n(&(self->bottom), b, __ATOMIC_SEQ_CST);
    ax_u64 t = __atomic_load_n(&(self->top), __ATOMIC_SEQ_CST);
    if ((t > b)) {
        __atomic_store_n(&(self->bottom), t, __ATOMIC_SEQ_CST);
        return ((ax_u64)(0));
    }
    ax_u64 id = (ax_bounds_check((ax_u64)(((ax_i64)((b % ((ax_u64)(4096)))))), (ax_u64)(4096)), (self->buffer)[((ax_i64)((b % ((ax_u64)(4096)))))]);
    if ((t == b)) {
        if ((!__sync_bool_compare_and_swap(&(self->top), t, (t + ((ax_u64)(1)))))) {
            id = ((ax_u64)(0));
        }
        __atomic_store_n(&(self->bottom), (t + ((ax_u64)(1))), __ATOMIC_SEQ_CST);
    }
    return id;
}

ax_u64 ax_runq_steal(struct ax_runq* self) {
    ax_u64 t = __atomic_load_n(&(self->top), __ATOMIC_SEQ_CST);
    ax_u64 b = __atomic_load_n(&(self->bottom), __ATOMIC_SEQ_CST);
    if ((t >= b)) {
        return ((ax_u64)(0));
    }
    ax_u64 id = (ax_bounds_check((ax_u64)(((ax_i64)((t % ((ax_u64)(4096)))))), (ax_u64)(4096)), (self->buffer)[((ax_i64)((t % ((ax_u64)(4096)))))]);
    if (__sync_bool_compare_and_swap(&(self->top), t, (t + ((ax_u64)(1))))) {
        return id;
    }
    return ((ax_u64)(0));
}

ax_i32 ax_runq_empty(const struct ax_runq* self) {
    ax_u64 t = __atomic_load_n(&(self->top), __ATOMIC_SEQ_CST);
    ax_u64 b = __atomic_load_n(&(self->bottom), __ATOMIC_SEQ_CST);
    if ((t >= b)) {
        return ((ax_i32)(1));
    }
    return ((ax_i32)(0));
}

ax_i32 ax_scheduler_init(struct ax_scheduler* self, ax_u32 worker_count) {
    if (((worker_count == ((ax_u32)(0))) || (worker_count > ((ax_u32)(256))))) {
        return ((ax_i32)((-1)));
    }
    self->worker_count = worker_count;
    self->running = ((ax_i32)(0));
    ax_u32 i = ((ax_u32)(0));
    while ((i < worker_count)) {
        struct ax_worker* w = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(i)) * ((ax_i64)(sizeof(struct ax_worker)))))));
        w->id = i;
        struct ax_runq* q = ((struct ax_runq*)((((ax_i64)(w)) + ((ax_i64)(8)))));
        ax_runq_init(q);
        w->running = ((ax_i32)(0));
        i = (i + ((ax_u32)(1)));
    }
    return ((ax_i32)(0));
}

ax_i32 ax_scheduler_submit(struct ax_scheduler* self, ax_u64 actor_id) {
    if ((actor_id == ((ax_u64)(0)))) {
        return ((ax_i32)((-1)));
    }
    ax_u32 target = ((ax_u32)((self->total_submitted % ((ax_u64)(self->worker_count)))));
    struct ax_worker* w = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(target)) * ((ax_i64)(sizeof(struct ax_worker)))))));
    struct ax_runq* q = ((struct ax_runq*)((((ax_i64)(w)) + ((ax_i64)(8)))));
    ax_i32 res = ax_runq_push(q, actor_id);
    if ((res == ((ax_i32)(0)))) {
        self->total_submitted = (self->total_submitted + ((ax_u64)(1)));
    }
    return res;
}

static struct ax_AxGlobalState* ax_get_state(void) {
    return ax_get_global_state();
}

void ax_ax_actor_system_init(void) {
    struct ax_AxGlobalState* state = ax_get_state();
    void** p_g_actor_table = ((void**)((((ax_i64)(state)) + ((ax_i64)(48)))));
    struct ax_AxActorTable* actor_table = ((struct ax_AxActorTable*)((*((void**)(p_g_actor_table)))));
    if ((actor_table != ((struct ax_AxActorTable*)(NULL)))) {
        return;
    }
    ax_i64 table_size = ((ax_i64)(sizeof(struct ax_AxActorTable)));
    struct ax_AxActorTable* new_table = ((struct ax_AxActorTable*)(ax_ax_os_alloc(table_size)));
    new_table->next_id = ((ax_u64)(1));
    new_table->actor_count = ((ax_u32)(0));
    ax_i64 actors_size = (((ax_i64)(1024)) * ((ax_i64)(sizeof(struct ax_AxActor))));
    new_table->actors = ((struct ax_AxActor*)(ax_ax_os_alloc(actors_size)));
    memset(((ax_u8*)(new_table->actors)), ((ax_u8)(0)), actors_size);
    (*((void**)(p_g_actor_table))) = ((void*)(new_table));
    ax_i64 sz = ((ax_i64)(sizeof(struct ax_scheduler)));
    struct ax_scheduler* new_sched = ((struct ax_scheduler*)(ax_ax_os_alloc(sz)));
    ax_scheduler_init(new_sched, ((ax_u32)(1)));
    void** p_g_sched = ((void**)((((ax_i64)(state)) + ((ax_i64)(56)))));
    (*((void**)(p_g_sched))) = ((void*)(new_sched));
    ax_bool* p_g_sched_initialized = ((ax_bool*)((((ax_i64)(state)) + ((ax_i64)(1)))));
    (*((ax_bool*)(p_g_sched_initialized))) = AX_TRUE;
}

void* ax_ax_actor_lookup(ax_u64 id) {
    struct ax_AxGlobalState* state = ax_get_state();
    void** p_g_actor_table = ((void**)((((ax_i64)(state)) + ((ax_i64)(48)))));
    struct ax_AxActorTable* actor_table = ((struct ax_AxActorTable*)((*((void**)(p_g_actor_table)))));
    if (((actor_table == ((struct ax_AxActorTable*)(NULL))) || (id == ((ax_u64)(0))))) {
        return ((void*)(NULL));
    }
    ax_i64 slot = ((ax_i64)((id % ((ax_u64)(1024)))));
    ax_i64 attempts = ((ax_i64)(0));
    while ((attempts < ((ax_i64)(1024)))) {
        struct ax_AxActor* actor = ((struct ax_AxActor*)((((ax_i64)(actor_table->actors)) + (((slot + attempts) % ((ax_i64)(1024))) * ((ax_i64)(sizeof(struct ax_AxActor)))))));
        if (((actor->id == id) && (actor->state != 3))) {
            return ((void*)(actor));
        }
        attempts = (attempts + ((ax_i64)(1)));
    }
    return ((void*)(NULL));
}

void ax_AxMsgQueue_ax_msgq_init(struct ax_AxMsgQueue* q) {
    q->head = ((struct ax_AxMessage*)(NULL));
    q->tail = ((struct ax_AxMessage*)(NULL));
    q->msg_count = ((ax_u64)(0));
    q->pending = ((ax_u64)(0));
}

void ax_AxMsgQueue_ax_msgq_push(struct ax_AxMsgQueue* q, struct ax_AxMessage* msg) {
    msg->next = ((struct ax_AxMessage*)(NULL));
    if ((q->tail == ((struct ax_AxMessage*)(NULL)))) {
        q->head = msg;
        q->tail = msg;
    } else {
        {
            q->tail->next = msg;
            q->tail = msg;
        }
    }
    q->msg_count = (q->msg_count + ((ax_u64)(1)));
    q->pending = (q->pending + ((ax_u64)(1)));
}

struct ax_AxMessage* ax_AxMsgQueue_ax_msgq_pop(struct ax_AxMsgQueue* q) {
    struct ax_AxMessage* msg = q->head;
    if ((msg == ((struct ax_AxMessage*)(NULL)))) {
        return ((struct ax_AxMessage*)(NULL));
    }
    q->head = msg->next;
    if ((q->head == ((struct ax_AxMessage*)(NULL)))) {
        q->tail = ((struct ax_AxMessage*)(NULL));
    }
    q->pending = (q->pending - ((ax_u64)(1)));
    return msg;
    ax_free(msg);
}

ax_i32 ax_ax_actor_send(ax_u64 target, ax_u64 sender, ax_u32 msg_type, void* payload, ax_u32 size) {
    struct ax_AxActor* actor_ptr = ((struct ax_AxActor*)(ax_ax_actor_lookup(target)));
    if ((actor_ptr == ((struct ax_AxActor*)(NULL)))) {
        return ((ax_i32)((-1)));
    }
    ax_i64 total_msg_size = (((ax_i64)(sizeof(struct ax_AxMessage))) + ((ax_i64)(size)));
    struct ax_AxMessage* msg = ((struct ax_AxMessage*)(ax_ActorHeap_ax_actor_alloc(actor_ptr->heap, total_msg_size)));
    if ((msg == ((struct ax_AxMessage*)(NULL)))) {
        return ((ax_i32)((-1)));
    }
    msg->next = ((struct ax_AxMessage*)(NULL));
    msg->sender = sender;
    msg->msg_type = msg_type;
    msg->size = size;
    if (((payload != ((void*)(NULL))) && (size > ((ax_u32)(0))))) {
        ax_u8* payload_dest = ((ax_u8*)((((ax_i64)(msg)) + ((ax_i64)(sizeof(struct ax_AxMessage))))));
        memcpy(payload_dest, ((ax_u8*)(payload)), ((ax_i64)(size)));
    }
    ax_AxMsgQueue_ax_msgq_push(((struct ax_AxMsgQueue*)((((ax_i64)(actor_ptr)) + ((ax_i64)(16))))), msg);
    return ((ax_i32)(0));
}

ax_u64 ax_ax_actor_spawn(void* handler, void* init_data, ax_u64 data_size) {
    struct ax_AxGlobalState* state = ax_get_state();
    void** p_g_actor_table = ((void**)((((ax_i64)(state)) + ((ax_i64)(48)))));
    struct ax_AxActorTable* actor_table = ((struct ax_AxActorTable*)((*((void**)(p_g_actor_table)))));
    if ((actor_table == ((struct ax_AxActorTable*)(NULL)))) {
        ax_ax_actor_system_init();
    }
    struct ax_AxActorTable* table = ((struct ax_AxActorTable*)((*((void**)(p_g_actor_table)))));
    if ((handler == ((void*)(NULL)))) {
        return ((ax_u64)(0));
    }
    ax_u64 id = table->next_id;
    table->next_id = (id + ((ax_u64)(1)));
    ax_i64 slot = ((ax_i64)((id % ((ax_u64)(1024)))));
    ax_i64 attempts = ((ax_i64)(0));
    ax_i64 found_slot = ((ax_i64)((-1)));
    while ((attempts < ((ax_i64)(1024)))) {
        ax_i64 cur_slot = ((slot + attempts) % ((ax_i64)(1024)));
        struct ax_AxActor* actor = ((struct ax_AxActor*)((((ax_i64)(table->actors)) + (cur_slot * ((ax_i64)(sizeof(struct ax_AxActor)))))));
        if (((actor->id == ((ax_u64)(0))) || (actor->state == 3))) {
            found_slot = cur_slot;
            break;
        }
        attempts = (attempts + ((ax_i64)(1)));
    }
    if ((found_slot == ((ax_i64)((-1))))) {
        return ((ax_u64)(0));
    }
    struct ax_AxActor* actor = ((struct ax_AxActor*)((((ax_i64)(table->actors)) + (found_slot * ((ax_i64)(sizeof(struct ax_AxActor)))))));
    memset(((ax_u8*)(actor)), ((ax_u8)(0)), sizeof(struct ax_AxActor));
    actor->id = id;
    actor->state = 0;
    actor->handler = handler;
    struct ax_ActorHeap* heap = ax_ax_actor_heap_create(id);
    actor->heap = heap;
    if (((init_data != ((void*)(NULL))) && (data_size > ((ax_u64)(0))))) {
        ax_u8* state_mem = ax_ActorHeap_ax_actor_alloc(heap, ((ax_i64)(data_size)));
        if ((state_mem != ((ax_u8*)(NULL)))) {
            memcpy(state_mem, ((ax_u8*)(init_data)), ((ax_i64)(data_size)));
            actor->state_data = ((void*)(state_mem));
            actor->state_size = data_size;
        }
    }
    actor->state = 1;
    table->actor_count = (table->actor_count + ((ax_u32)(1)));
    ax_ax_actor_send(id, ((ax_u64)(0)), ((ax_u32)(0)), init_data, ((ax_u32)(data_size)));
    ax_bool* p_g_sched_initialized = ((ax_bool*)((((ax_i64)(state)) + ((ax_i64)(1)))));
    if ((*((ax_bool*)(p_g_sched_initialized)))) {
        void** p_g_sched = ((void**)((((ax_i64)(state)) + ((ax_i64)(56)))));
        struct ax_scheduler* sched = ((struct ax_scheduler*)((*((void**)(p_g_sched)))));
        ax_scheduler_submit(sched, id);
    }
    return id;
}

static ax_bool ax_has_active_actors(void) {
    struct ax_AxGlobalState* state = ax_get_state();
    void** p_g_actor_table = ((void**)((((ax_i64)(state)) + ((ax_i64)(48)))));
    struct ax_AxActorTable* actor_table = ((struct ax_AxActorTable*)((*((void**)(p_g_actor_table)))));
    if ((actor_table == ((struct ax_AxActorTable*)(NULL)))) {
        return AX_FALSE;
    }
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(1024)))) {
        struct ax_AxActor* actor = ((struct ax_AxActor*)((((ax_i64)(actor_table->actors)) + (i * ((ax_i64)(sizeof(struct ax_AxActor)))))));
        if (((actor->id != ((ax_u64)(0))) && (actor->state == 1))) {
            if ((actor->mailbox.pending > ((ax_u64)(0)))) {
                return AX_TRUE;
            }
        }
        i = (i + ((ax_i64)(1)));
    }
    return AX_FALSE;
}

ax_i32 ax_scheduler_run(struct ax_scheduler* self) {
    self->running = ((ax_i32)(1));
    ax_u32 i = ((ax_u32)(0));
    while ((i < self->worker_count)) {
        struct ax_worker* w = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(i)) * ((ax_i64)(sizeof(struct ax_worker)))))));
        w->running = ((ax_i32)(1));
        i = (i + ((ax_u32)(1)));
    }
    while (ax_has_active_actors()) {
        ax_u32 j = ((ax_u32)(0));
        while ((j < self->worker_count)) {
            struct ax_worker* w = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(j)) * ((ax_i64)(sizeof(struct ax_worker)))))));
            struct ax_runq* q = ((struct ax_runq*)((((ax_i64)(w)) + ((ax_i64)(8)))));
            ax_u64 id = ax_runq_pop(q);
            if ((id != ((ax_u64)(0)))) {
                void* actor_ptr = ax_ax_actor_lookup(id);
                if ((actor_ptr != ((void*)(NULL)))) {
                    while ((ax_actor_step(actor_ptr) != ((ax_i32)(0)))) {
                        w->tasks_executed = (w->tasks_executed + ((ax_u64)(1)));
                    }
                    if (((ax_actor_is_running(actor_ptr) != ((ax_i32)(0))) && (ax_actor_has_messages(actor_ptr) != ((ax_i32)(0))))) {
                        ax_runq_push(q, id);
                    }
                }
            }
            j = (j + ((ax_u32)(1)));
        }
    }
    return ((ax_i32)(0));
}

static void ax_scheduler_worker_loop(struct ax_scheduler* self, struct ax_worker* w) {
    struct ax_runq* q = ((struct ax_runq*)((((ax_i64)(w)) + ((ax_i64)(8)))));
    while ((w->running != ((ax_i32)(0)))) {
        ax_u64 id = ax_runq_pop(q);
        if ((id == ((ax_u64)(0)))) {
            ax_u32 i = ((ax_u32)(0));
            ax_i32 stop_loop = ((ax_i32)(0));
            while (((i < self->worker_count) && (stop_loop == ((ax_i32)(0))))) {
                if ((i != w->id)) {
                    w->steals_attempted = (w->steals_attempted + ((ax_u64)(1)));
                    struct ax_worker* w_other = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(i)) * ((ax_i64)(sizeof(struct ax_worker)))))));
                    struct ax_runq* q_other = ((struct ax_runq*)((((ax_i64)(w_other)) + ((ax_i64)(8)))));
                    ax_u64 stolen_id = ax_runq_steal(q_other);
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
        void* actor_ptr = ax_ax_actor_lookup(id);
        if ((actor_ptr != ((void*)(NULL)))) {
            while ((ax_actor_step(actor_ptr) != ((ax_i32)(0)))) {
                w->tasks_executed = (w->tasks_executed + ((ax_u64)(1)));
            }
            if (((ax_actor_is_running(actor_ptr) != ((ax_i32)(0))) && (ax_actor_has_messages(actor_ptr) != ((ax_i32)(0))))) {
                ax_runq_push(q, id);
            }
        }
    }
}

void ax_scheduler_shutdown(struct ax_scheduler* self) {
    self->running = ((ax_i32)(0));
    ax_u32 i = ((ax_u32)(0));
    while ((i < self->worker_count)) {
        struct ax_worker* w = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(i)) * ((ax_i64)(sizeof(struct ax_worker)))))));
        w->running = ((ax_i32)(0));
        i = (i + ((ax_u32)(1)));
    }
}

void ax_scheduler_stats(struct ax_scheduler* self, struct ax_SchedulerStats* stats) {
    if ((stats == ((struct ax_SchedulerStats*)(NULL)))) {
        return;
    }
    stats->worker_count = self->worker_count;
    stats->total_submitted = self->total_submitted;
    stats->total_executed = ((ax_u64)(0));
    stats->total_steals = ((ax_u64)(0));
    ax_u32 i = ((ax_u32)(0));
    while ((i < self->worker_count)) {
        struct ax_worker* w = ((struct ax_worker*)((((ax_i64)(self)) + (((ax_i64)(i)) * ((ax_i64)(sizeof(struct ax_worker)))))));
        stats->total_executed = (stats->total_executed + w->tasks_executed);
        stats->total_steals = (stats->total_steals + w->steals_succeeded);
        i = (i + ((ax_u32)(1)));
    }
}

static void ax_test_print_str(ax_string s) {
    if (1) {
        void* h = GetStdHandle(((ax_u32)(0xFFFFFFF5)));
        ax_u32 written = ((ax_u32)(0));
        WriteFile(h, ((void*)(s.ptr)), ((ax_u32)(ax_str_len(s))), ((void*)(&(written))), ((void*)(NULL)));
    } else {
        {
            syscall(((ax_u64)(1)), ((ax_u64)(1)), ((ax_u64)(((ax_u8*)(s.ptr)))), ((ax_u64)(ax_str_len(s))), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
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
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_runq_basic\n", .len=24});
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
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_runq_steal\n", .len=24});
}

static void ax_test_scheduler_lifecycle(void) {
    ax_u64 sz = sizeof(struct ax_scheduler);
    struct ax_scheduler* sched = ((struct ax_scheduler*)(ax_alloc(sz)));
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
    ax_free(((ax_u8*)(sched)));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_scheduler_lifecycle\n", .len=33});
}

ax_i32 ax_main_usr(void) {
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native Scheduler unit tests...\n", .len=45});
    ax_test_runq_basic();
    ax_test_runq_steal();
    ax_test_scheduler_lifecycle();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"All AXIOM-native Scheduler tests passed!\n", .len=41});
    return 0;
}

/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}

/* Bridge allocator functions for C runtime integration */
struct ax_ActorHeap;
struct ax_ActorHeap* ax_ax_actor_heap_create(ax_u64 actor_id);
void ax_ActorHeap_ax_actor_heap_destroy(struct ax_ActorHeap* heap);
ax_u8* ax_ActorHeap_ax_actor_alloc(struct ax_ActorHeap* heap, ax_i64 user_size);
void ax_ActorHeap_ax_actor_free(struct ax_ActorHeap* heap, ax_u8* user_ptr);

void* ax_actor_heap_create(unsigned long long actor_id) {
    return (void*)ax_ax_actor_heap_create((ax_u64)actor_id);
}
void ax_actor_heap_destroy(void* heap) {
    ax_ActorHeap_ax_actor_heap_destroy((struct ax_ActorHeap*)heap);
}
void* ax_actor_alloc(void* heap, size_t user_size) {
    return (void*)ax_ActorHeap_ax_actor_alloc((struct ax_ActorHeap*)heap, (ax_i64)user_size);
}
void ax_actor_free(void* heap, void* user_ptr) {
    ax_ActorHeap_ax_actor_free((struct ax_ActorHeap*)heap, (ax_u8*)user_ptr);
}
