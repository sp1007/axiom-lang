#define AX_FREESTANDING_RUNTIME
#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_AxRef;
struct ax_AxHeader;
struct ax_AxGlobalState;
struct ax_Segment;
struct ax_SegmentList;
struct ax_FreeList;
struct ax_ActorHeap;
struct ax_FreeSlot;
struct ax_AxRunQueue;
struct ax_AxScheduler;
struct ax_AxWorker;
struct ax_AxSchedulerStats;

/* Type definitions */
struct ax_AxRef {
    void* ptr;
    ax_u64 gen_id;
};

struct ax_AxHeader {
    ax_u64 gen_id;
    ax_u64 size;
};

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

struct ax_AxRunQueue {
    ax_u64* buffer;
    ax_u64 top;
    ax_u64 bottom;
};

struct ax_AxScheduler {
    struct ax_AxWorker* workers;
    ax_u32 worker_count;
    ax_i32 running;
    ax_u64 total_submitted;
};

struct ax_AxWorker {
    ax_u32 id;
    struct ax_AxRunQueue runq;
    ax_u64 tasks_executed;
    ax_u64 steals_attempted;
    ax_u64 steals_succeeded;
    ax_i32 running;
};

struct ax_AxSchedulerStats {
    ax_u32 worker_count;
    ax_u64 total_submitted;
    ax_u64 total_executed;
    ax_u64 total_steals;
};


/* Global variables */
extern ax_u64 ax_g_program_name;
ax_u64 ax_g_program_name = ((ax_u64)(0));
extern ax_u64 ax_g_ax_global_state;
ax_u64 ax_g_ax_global_state = ((ax_u64)(0));
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
void* ax_sys_mmap(void* addr, ax_u64 len, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset);
ax_i32 ax_sys_munmap(void* addr, ax_u64 len);
ax_i64 ax_sys_write(ax_i32 fd, void* buf, ax_u64 count);
void ax_sys_exit(ax_i32 code);
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
void ax_ax_set_program_name(ax_u8* name);
static ax_i64 ax_my_strlen(ax_u8* s);
static ax_bool ax_my_strcmp(ax_u8* s1, ax_string s2);
void ax_ax_panic(ax_string msg);
void ax_ax_assert_axiom(ax_bool cond, ax_string msg);
void* ax_ax_get_global_state_internal(void);
ax_i64 ax_ax_compiler_intrinsic(ax_u8* name, void* p1, void* p2, void* p3);
void* ax_AxRef_ax_deref(struct ax_AxRef ref);
void ax_ax_invalidate(void* p);
ax_bool ax_AxRef_ax_ref_valid(struct ax_AxRef ref);
struct ax_AxRef ax_ax_make_ref(void* p);
void* sys_mmap(void* addr, ax_u64 len, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset);
ax_i32 sys_munmap(void* addr, ax_u64 len);
void sys_exit(ax_i32 code);
void ax_panic(ax_string msg);
static void ax_ax_panic_str(ax_string msg);
void* ax_get_global_state_internal(void);
struct ax_AxGlobalState* ax_get_global_state(void);
ax_i64* ax_axalloc_get_slab_used(void);
struct ax_Segment** ax_axalloc_get_free_pool(void);
struct ax_Segment* ax_axalloc_get_slab(void);
struct ax_ActorHeap** ax_axalloc_get_global_heap(void);
static struct ax_ActorHeap* ax_get_global_heap(void);
ax_u64 ax_ax_size_class_size(ax_i32 sc);
ax_i32 ax_ax_size_class_for(ax_i64 user_size);
struct ax_SegmentList* ax_ActorHeap_ax_get_segment_list(struct ax_ActorHeap* heap, ax_i32 sc);
struct ax_FreeList* ax_ActorHeap_ax_get_free_list(struct ax_ActorHeap* heap, ax_i32 sc);
void ax_FreeList_ax_free_list_push(struct ax_FreeList* list, ax_u8* block);
ax_u8* ax_FreeList_ax_free_list_pop(struct ax_FreeList* list);
void* ax_ax_os_alloc(ax_i64 size);
void ax_ax_os_free(void* ptr_val, ax_i64 size);
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
void* ax_ax_alloc(ax_i64 size);
void ax_ax_free(void* p);
void* ax_ax_realloc(void* p, ax_i64 new_size);
ax_i64 ax_ax_alloc_size(void* p);
ax_u8* ax_ax_numa_alloc(ax_i64 size, ax_i32 node_id);
void ax_AxRunQueue_ax_runq_init(struct ax_AxRunQueue* q);
ax_i32 ax_AxRunQueue_ax_runq_push(struct ax_AxRunQueue* q, ax_u64 id);
ax_u64 ax_AxRunQueue_ax_runq_pop(struct ax_AxRunQueue* q);
ax_u64 ax_AxRunQueue_ax_runq_steal(struct ax_AxRunQueue* q);
ax_i32 ax_AxRunQueue_ax_runq_empty(struct ax_AxRunQueue* q);
ax_i32 ax_AxScheduler_ax_scheduler_init(struct ax_AxScheduler* sched, ax_u32 worker_count);
ax_i32 ax_AxScheduler_ax_scheduler_submit(struct ax_AxScheduler* sched, ax_u64 actor_id);
ax_i32 ax_AxScheduler_ax_scheduler_run(struct ax_AxScheduler* sched);
void ax_AxScheduler_ax_scheduler_stats(struct ax_AxScheduler* sched, struct ax_AxSchedulerStats* stats);
void ax_AxScheduler_ax_scheduler_shutdown(struct ax_AxScheduler* sched);
static void ax_test_print_str(ax_string s);
static void ax_test_allocator(void);
static void ax_test_genref(void);
ax_i32 ax_main_usr(void);


void* ax_sys_mmap(void* addr, ax_u64 len, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset) {
    if (1) {
        return VirtualAlloc(addr, len, ((ax_u32)(0x3000)), ((ax_u32)(0x04)));
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(9)), ((ax_u64)(addr)), len, ((ax_u64)(prot)), ((ax_u64)(flags)), ((ax_u64)(fd)), ((ax_u64)(offset)));
            return ((void*)(res));
        }
    }
}

ax_i32 ax_sys_munmap(void* addr, ax_u64 len) {
    if (1) {
        ax_i32 res = VirtualFree(addr, ((ax_u64)(0)), ((ax_u32)(0x8000)));
        if ((res != 0)) {
            return 0;
        }
        return (-1);
    } else {
        {
            ax_i64 res = syscall(((ax_u64)(11)), ((ax_u64)(addr)), len, ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            return ((ax_i32)(res));
        }
    }
}

ax_i64 ax_sys_write(ax_i32 fd, void* buf, ax_u64 count) {
    if (1) {
        ax_u32 std_handle = ((ax_u32)(0xFFFFFFF4));
        if ((fd == 1)) {
            std_handle = ((ax_u32)(0xFFFFFFF5));
        }
        void* h = GetStdHandle(std_handle);
        ax_u32 written = ((ax_u32)(0));
        ax_i32 res = WriteFile(h, buf, ((ax_u32)(count)), ((void*)(&(written))), ((void*)(NULL)));
        if ((res != 0)) {
            return ((ax_i64)(written));
        }
        return (-1);
    } else {
        {
            return syscall(((ax_u64)(1)), ((ax_u64)(fd)), ((ax_u64)(buf)), count, ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

void ax_sys_exit(ax_i32 code) {
    if (1) {
        ExitProcess(((ax_u32)(code)));
    } else {
        {
            syscall(((ax_u64)(60)), ((ax_u64)(code)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

void ax_ax_set_program_name(ax_u8* name) {
    ax_g_program_name = ((ax_u64)(name));
}

static ax_i64 ax_my_strlen(ax_u8* s) {
    if ((s == ((ax_u8*)(NULL)))) {
        return 0;
    }
    ax_i64 len = ((ax_i64)(0));
    while (((*((ax_u8*)(((ax_u8*)((((ax_i64)(s)) + len)))))) != ((ax_u8)(0)))) {
        len = (len + 1);
    }
    return len;
}

static ax_bool ax_my_strcmp(ax_u8* s1, ax_string s2) {
    if ((s1 == ((ax_u8*)(NULL)))) {
        return AX_FALSE;
    }
    ax_i64 i = ((ax_i64)(0));
    while ((i < s2.len)) {
        ax_u8 c1 = (*((ax_u8*)(((ax_u8*)((((ax_i64)(s1)) + i))))));
        ax_u8 c2 = (*((ax_u8*)(((ax_u8*)((((ax_i64)(s2.ptr)) + i))))));
        if ((c1 != c2)) {
            return AX_FALSE;
        }
        i = (i + 1);
    }
    return ((*((ax_u8*)(((ax_u8*)((((ax_i64)(s1)) + s2.len)))))) == ((ax_u8)(0)));
}

void ax_ax_panic(ax_string msg) {
    ax_string prefix = (ax_string){.ptr=(const ax_u8*)"\nAXIOM PANIC: ", .len=14};
    if (1) {
        void* h_err = GetStdHandle(((ax_u32)(0xFFFFFFF4)));
        ax_u32 written = ((ax_u32)(0));
        WriteFile(h_err, ((void*)(prefix.ptr)), ((ax_u32)(prefix.len)), ((void*)(&(written))), ((void*)(NULL)));
        if ((ax_g_program_name != ((ax_u64)(0)))) {
            ax_u8* prog_name = ((ax_u8*)(ax_g_program_name));
            WriteFile(h_err, ((void*)(prog_name)), ((ax_u32)(ax_my_strlen(prog_name))), ((void*)(&(written))), ((void*)(NULL)));
            WriteFile(h_err, ((void*)((ax_string){.ptr=(const ax_u8*)": ", .len=2}.ptr)), ((ax_u32)(2)), ((void*)(&(written))), ((void*)(NULL)));
        }
        WriteFile(h_err, ((void*)(msg.ptr)), ((ax_u32)(msg.len)), ((void*)(&(written))), ((void*)(NULL)));
        WriteFile(h_err, ((void*)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}.ptr)), ((ax_u32)(1)), ((void*)(&(written))), ((void*)(NULL)));
        ExitProcess(((ax_u32)(101)));
    } else {
        {
            syscall(((ax_u64)(1)), ((ax_u64)(2)), ((ax_u64)(prefix.ptr)), ((ax_u64)(prefix.len)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            if ((ax_g_program_name != ((ax_u64)(0)))) {
                syscall(((ax_u64)(1)), ((ax_u64)(2)), ax_g_program_name, ((ax_u64)(ax_my_strlen(((ax_u8*)(ax_g_program_name))))), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
                syscall(((ax_u64)(1)), ((ax_u64)(2)), ((ax_u64)((ax_string){.ptr=(const ax_u8*)": ", .len=2}.ptr)), ((ax_u64)(2)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            }
            syscall(((ax_u64)(1)), ((ax_u64)(2)), ((ax_u64)(msg.ptr)), ((ax_u64)(msg.len)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            syscall(((ax_u64)(1)), ((ax_u64)(2)), ((ax_u64)((ax_string){.ptr=(const ax_u8*)"\n", .len=1}.ptr)), ((ax_u64)(1)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
            syscall(((ax_u64)(60)), ((ax_u64)(101)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

void ax_ax_assert_axiom(ax_bool cond, ax_string msg) {
    if ((!cond)) {
        ax_ax_panic(msg);
    }
}

void* ax_ax_get_global_state_internal(void) {
    if ((ax_g_ax_global_state == ((ax_u64)(0)))) {
        if (1) {
            void* block = VirtualAlloc(((void*)(NULL)), ((ax_u64)(4096)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)));
            ax_g_ax_global_state = ((ax_u64)(block));
        } else {
            {
                ax_i64 res = syscall(((ax_u64)(9)), ((ax_u64)(0)), ((ax_u64)(4096)), ((ax_u64)(3)), ((ax_u64)(0x22)), ((ax_u64)(0xFFFFFFFFFFFFFFFF)), ((ax_u64)(0)));
                ax_g_ax_global_state = ((ax_u64)(res));
            }
        }
    }
    return ((void*)(ax_g_ax_global_state));
}

ax_i64 ax_ax_compiler_intrinsic(ax_u8* name, void* p1, void* p2, void* p3) {
    if (ax_my_strcmp(name, (ax_string){.ptr=(const ax_u8*)"is_windows", .len=10})) {
        if (1) {
            return 1;
        }
        return 0;
    }
    if (ax_my_strcmp(name, (ax_string){.ptr=(const ax_u8*)"atomic_load", .len=11})) {
        ax_i64* p = ((ax_i64*)(p1));
        return (*((ax_i64*)(p)));
    }
    if (ax_my_strcmp(name, (ax_string){.ptr=(const ax_u8*)"atomic_store", .len=12})) {
        ax_i64* p = ((ax_i64*)(p1));
        (*((ax_i64*)(p))) = ((ax_i64)(p2));
        return 0;
    }
    if (ax_my_strcmp(name, (ax_string){.ptr=(const ax_u8*)"atomic_cas", .len=10})) {
        ax_i64* p = ((ax_i64*)(p1));
        ax_i64 expected = ((ax_i64)(p2));
        ax_i64 desired = ((ax_i64)(p3));
        if (((*((ax_i64*)(p))) == expected)) {
            (*((ax_i64*)(p))) = desired;
            return 1;
        }
        return 0;
    }
    return 0;
}

void* ax_AxRef_ax_deref(struct ax_AxRef ref) {
    if ((ref.ptr == ((void*)(NULL)))) {
        ax_ax_panic((ax_string){.ptr=(const ax_u8*)"null pointer dereference", .len=24});
    }
    struct ax_AxHeader* h = ((struct ax_AxHeader*)((((ax_i64)(ref.ptr)) - ((ax_i64)(16)))));
    if ((h->gen_id != ref.gen_id)) {
        ax_ax_panic((ax_string){.ptr=(const ax_u8*)"GenerationalID mismatch: use-after-free detected", .len=48});
    }
    return ref.ptr;
}

void ax_ax_invalidate(void* p) {
    if ((p == ((void*)(NULL)))) {
        return;
    }
    struct ax_AxHeader* h = ((struct ax_AxHeader*)((((ax_i64)(p)) - ((ax_i64)(16)))));
    h->gen_id = ((ax_u64)(0));
}

ax_bool ax_AxRef_ax_ref_valid(struct ax_AxRef ref) {
    if ((ref.ptr == ((void*)(NULL)))) {
        return AX_FALSE;
    }
    struct ax_AxHeader* h = ((struct ax_AxHeader*)((((ax_i64)(ref.ptr)) - ((ax_i64)(16)))));
    return (h->gen_id == ref.gen_id);
}

struct ax_AxRef ax_ax_make_ref(void* p) {
    if ((p == ((void*)(NULL)))) {
        return ((struct ax_AxRef){.ptr=((void*)(NULL)), .gen_id=((ax_u64)(0))});
    }
    struct ax_AxHeader* h = ((struct ax_AxHeader*)((((ax_i64)(p)) - ((ax_i64)(16)))));
    return ((struct ax_AxRef){.ptr=p, .gen_id=h->gen_id});
}

static void ax_ax_panic_str(ax_string msg) {
    ax_ax_panic(msg);
}

struct ax_AxGlobalState* ax_get_global_state(void) {
    return ((struct ax_AxGlobalState*)(ax_ax_get_global_state_internal()));
}

ax_i64* ax_axalloc_get_slab_used(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    return ((ax_i64*)((((ax_i64)(state)) + ((ax_i64)(16)))));
}

struct ax_Segment** ax_axalloc_get_free_pool(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    return ((struct ax_Segment**)((((ax_i64)(state)) + ((ax_i64)(24)))));
}

struct ax_Segment* ax_axalloc_get_slab(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    struct ax_Segment** p_g_slab = ((struct ax_Segment**)((((ax_i64)(state)) + ((ax_i64)(32)))));
    if (((*((struct ax_Segment**)(p_g_slab))) == ((struct ax_Segment*)(NULL)))) {
        if (1) {
            (*((struct ax_Segment**)(p_g_slab))) = ((struct ax_Segment*)(VirtualAlloc(((void*)(NULL)), ((ax_u64)(196608)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)))));
        } else {
            {
                (*((struct ax_Segment**)(p_g_slab))) = ((struct ax_Segment*)(ax_sys_mmap(((void*)(NULL)), ((ax_u64)(196608)), ((ax_i32)(3)), ((ax_i32)(0x22)), ((ax_i32)((-1))), ((ax_i64)(0)))));
            }
        }
        if (((*((struct ax_Segment**)(p_g_slab))) == ((struct ax_Segment*)(NULL)))) {
            ax_ax_panic_str((ax_string){.ptr=(const ax_u8*)"axalloc_get_slab FAILED", .len=23});
        }
    }
    return (*((struct ax_Segment**)(p_g_slab)));
}

struct ax_ActorHeap** ax_axalloc_get_global_heap(void) {
    struct ax_AxGlobalState* state = ax_get_global_state();
    return ((struct ax_ActorHeap**)((((ax_i64)(state)) + ((ax_i64)(40)))));
}

static struct ax_ActorHeap* ax_get_global_heap(void) {
    struct ax_ActorHeap** p_global_heap = ax_axalloc_get_global_heap();
    if (((*((struct ax_ActorHeap**)(p_global_heap))) == ((struct ax_ActorHeap*)(NULL)))) {
        ax_ax_segment_manager_init();
        (*((struct ax_ActorHeap**)(p_global_heap))) = ax_ax_actor_heap_create(((ax_u64)(1)));
    }
    return (*((struct ax_ActorHeap**)(p_global_heap)));
}

ax_u64 ax_ax_size_class_size(ax_i32 sc) {
    if ((sc == 0)) {
        return ((ax_u64)(8));
    }
    if ((sc == 1)) {
        return ((ax_u64)(16));
    }
    if ((sc == 2)) {
        return ((ax_u64)(32));
    }
    if ((sc == 3)) {
        return ((ax_u64)(64));
    }
    if ((sc == 4)) {
        return ((ax_u64)(128));
    }
    if ((sc == 5)) {
        return ((ax_u64)(256));
    }
    if ((sc == 6)) {
        return ((ax_u64)(512));
    }
    if ((sc == 7)) {
        return ((ax_u64)(1024));
    }
    if ((sc == 8)) {
        return ((ax_u64)(2048));
    }
    if ((sc == 9)) {
        return ((ax_u64)(4096));
    }
    return ((ax_u64)(0));
}

ax_i32 ax_ax_size_class_for(ax_i64 user_size) {
    ax_i64 total = (user_size + ((ax_i64)(16)));
    if ((total <= ((ax_i64)(8)))) {
        return 0;
    }
    if ((total <= ((ax_i64)(16)))) {
        return 1;
    }
    if ((total <= ((ax_i64)(32)))) {
        return 2;
    }
    if ((total <= ((ax_i64)(64)))) {
        return 3;
    }
    if ((total <= ((ax_i64)(128)))) {
        return 4;
    }
    if ((total <= ((ax_i64)(256)))) {
        return 5;
    }
    if ((total <= ((ax_i64)(512)))) {
        return 6;
    }
    if ((total <= ((ax_i64)(1024)))) {
        return 7;
    }
    if ((total <= ((ax_i64)(2048)))) {
        return 8;
    }
    if ((total <= ((ax_i64)(4096)))) {
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
    struct ax_FreeSlot* slot = ((struct ax_FreeSlot*)((((ax_i64)(block)) + 16)));
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
    return ((ax_u8*)((((ax_i64)(slot)) - 16)));
}

void* ax_ax_os_alloc(ax_i64 size) {
    if (1) {
        void* res = VirtualAlloc(((void*)(NULL)), ((ax_u64)(size)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)));
        if ((res == ((void*)(NULL)))) {
            ax_ax_panic_str((ax_string){.ptr=(const ax_u8*)"VirtualAlloc failed", .len=19});
        }
        return res;
        ax_free(res);
    } else {
        {
            void* block = ax_sys_mmap(((void*)(NULL)), ((ax_u64)(size)), ((ax_i32)(3)), ((ax_i32)(0x22)), ((ax_i32)((-1))), ((ax_i64)(0)));
            if ((block == ((void*)(NULL)))) {
                ax_ax_panic_str((ax_string){.ptr=(const ax_u8*)"ax_os_alloc FAILED", .len=18});
            }
            return block;
            ax_free(block);
        }
    }
}

void ax_ax_os_free(void* ptr_val, ax_i64 size) {
    if ((ptr_val == ((void*)(NULL)))) {
        return;
    }
    if (1) {
        VirtualFree(ptr_val, ((ax_u64)(0)), ((ax_u32)(0x8000)));
    } else {
        {
            ax_sys_munmap(ptr_val, ((ax_u64)(size)));
        }
    }
}

void ax_ax_segment_manager_init(void) {
    (*((ax_i64*)(ax_axalloc_get_slab_used()))) = 0;
    (*((struct ax_Segment**)(ax_axalloc_get_free_pool()))) = ((struct ax_Segment*)(NULL));
    struct ax_Segment* slab = ax_axalloc_get_slab();
    memset(((ax_u8*)(slab)), ((ax_u8)(0)), ((ax_i64)(196608)));
}

void ax_ax_segment_manager_shutdown(void) {
    ax_i64* slab_used_ptr = ax_axalloc_get_slab_used();
    struct ax_Segment* slab = ax_axalloc_get_slab();
    ax_i64 i = ((ax_i64)(0));
    while ((i < (*((ax_i64*)(slab_used_ptr))))) {
        struct ax_Segment* seg = ((struct ax_Segment*)((((ax_i64)(slab)) + (i * ((ax_i64)(sizeof(struct ax_Segment)))))));
        if ((seg->magic == ax_SEGMENT_MAGIC)) {
            ax_ax_os_free(((void*)(seg->base)), ax_SEGMENT_SIZE);
            seg->magic = ((ax_u32)(0));
        }
        i = (i + 1);
    }
    (*((ax_i64*)(slab_used_ptr))) = 0;
    (*((struct ax_Segment**)(ax_axalloc_get_free_pool()))) = ((struct ax_Segment*)(NULL));
}

static struct ax_Segment* ax_alloc_segment_meta(void) {
    struct ax_Segment** free_pool_ptr = ax_axalloc_get_free_pool();
    if (((*((struct ax_Segment**)(free_pool_ptr))) != ((struct ax_Segment*)(NULL)))) {
        struct ax_Segment* seg = (*((struct ax_Segment**)(free_pool_ptr)));
        (*((struct ax_Segment**)(free_pool_ptr))) = seg->next;
        return seg;
        ax_free(seg);
    }
    ax_i64* slab_used_ptr = ax_axalloc_get_slab_used();
    if (((*((ax_i64*)(slab_used_ptr))) >= ((ax_i64)(4096)))) {
        return ((struct ax_Segment*)(NULL));
    }
    struct ax_Segment* slab = ax_axalloc_get_slab();
    struct ax_Segment* seg = ((struct ax_Segment*)((((ax_i64)(slab)) + ((*((ax_i64*)(slab_used_ptr))) * ((ax_i64)(sizeof(struct ax_Segment)))))));
    (*((ax_i64*)(slab_used_ptr))) = ((*((ax_i64*)(slab_used_ptr))) + 1);
    return seg;
    ax_free(seg);
}

static void ax_Segment_free_segment_meta(struct ax_Segment* seg) {
    memset(((ax_u8*)(seg)), ((ax_u8)(0)), sizeof(struct ax_Segment));
    struct ax_Segment** free_pool_ptr = ax_axalloc_get_free_pool();
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
    ax_i64 block_size = ((ax_i64)(ax_ax_size_class_size(sc)));
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
    ax_i64 block_size = ((ax_i64)(ax_ax_size_class_size(sc)));
    if (((((ax_i64)(seg->bump)) + block_size) > ((ax_i64)(seg->limit)))) {
        return ((ax_u8*)(NULL));
    }
    ax_u8* block = seg->bump;
    seg->bump = ((ax_u8*)((((ax_i64)(seg->bump)) + block_size)));
    return block;
    ax_free(block);
}

ax_u8* ax_ax_large_alloc(ax_i64 user_size) {
    ax_i64 total = (((ax_i64)(24)) + user_size);
    ax_i64 page_aligned = (((total + ((ax_i64)(4095))) / ((ax_i64)(4096))) * ((ax_i64)(4096)));
    ax_u8* block = ((ax_u8*)(ax_ax_os_alloc(((ax_i64)(page_aligned)))));
    if ((block == ((ax_u8*)(NULL)))) {
        return ((ax_u8*)(NULL));
    }
    ax_i64* p_total = ((ax_i64*)(block));
    (*((ax_i64*)(p_total))) = page_aligned;
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)((((ax_i64)(block)) + 8)));
    hdr->gen_id = ((ax_u64)(1));
    hdr->size = ((ax_u64)(user_size));
    return ((ax_u8*)((((ax_i64)(block)) + 24)));
}

void ax_ax_large_free(ax_u8* user_ptr, ax_i64 user_size) {
    if ((user_ptr == ((ax_u8*)(NULL)))) {
        return;
    }
    void* block = ((void*)((((ax_i64)(user_ptr)) - 24)));
    ax_i64* p_total = ((ax_i64*)((((ax_i64)(user_ptr)) - 24)));
    ax_i64 total = ((ax_i64)((*((ax_i64*)(p_total)))));
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
    hdr->gen_id = ((ax_u64)(1));
    hdr->size = ((ax_u64)(sz));
    ax_u64 block_size = ax_ax_size_class_size(sc);
    heap->total_allocated = (heap->total_allocated + block_size);
    heap->alloc_count = (heap->alloc_count + ((ax_u64)(1)));
    return ((ax_u8*)((((ax_i64)(block)) + 16)));
}

void ax_ActorHeap_ax_actor_free(struct ax_ActorHeap* heap, ax_u8* user_ptr) {
    if ((heap == ((struct ax_ActorHeap*)(NULL)))) {
        return;
    }
    if ((user_ptr == ((ax_u8*)(NULL)))) {
        return;
    }
    ax_u8* block = ((ax_u8*)((((ax_i64)(user_ptr)) - 16)));
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)(block));
    ax_i32 sc = ax_ax_size_class_for(((ax_i64)(hdr->size)));
    hdr->gen_id = ((ax_u64)(0));
    if (((sc == 10) || (sc >= 10))) {
        ax_i64* p_total = ((ax_i64*)((((ax_i64)(user_ptr)) - 16)));
        ax_i64 user_size = ((*((ax_i64*)(p_total))) - ((ax_i64)(16)));
        heap->total_freed = (heap->total_freed + ((ax_u64)(user_size)));
        heap->free_count = (heap->free_count + ((ax_u64)(1)));
        ax_ax_large_free(user_ptr, ((ax_i64)(0)));
        return;
    }
    ax_u64 block_size = ((ax_u64)(ax_ax_size_class_size(sc)));
    heap->total_freed = (heap->total_freed + block_size);
    heap->free_count = (heap->free_count + ((ax_u64)(1)));
    struct ax_FreeList* free_list = ax_ActorHeap_ax_get_free_list(heap, sc);
    ax_FreeList_ax_free_list_push(free_list, block);
}

void* ax_ax_alloc(ax_i64 size) {
    ax_i64 sz = size;
    if ((sz < ((ax_i64)(1)))) {
        sz = ((ax_i64)(1));
    }
    struct ax_ActorHeap* heap = ax_get_global_heap();
    ax_u8* p = ax_ActorHeap_ax_actor_alloc(heap, sz);
    if ((p == ((ax_u8*)(NULL)))) {
        ax_ax_panic_str((ax_string){.ptr=(const ax_u8*)"ax_alloc: out of memory", .len=23});
    }
    return ((void*)(p));
}

void ax_ax_free(void* p) {
    if ((p == ((void*)(NULL)))) {
        return;
    }
    struct ax_ActorHeap* heap = ax_get_global_heap();
    ax_ActorHeap_ax_actor_free(heap, ((ax_u8*)(p)));
}

void* ax_ax_realloc(void* p, ax_i64 new_size) {
    if ((p == ((void*)(NULL)))) {
        return ax_ax_alloc(new_size);
    }
    struct ax_AxHeader* block = ((struct ax_AxHeader*)((((ax_i64)(p)) - ((ax_i64)(16)))));
    ax_i64 old_size = ((ax_i64)(block->size));
    ax_i32 old_sc = ax_ax_size_class_for(old_size);
    ax_i32 new_sc = ax_ax_size_class_for(new_size);
    if (((old_sc == new_sc) && (old_sc != 10))) {
        block->size = ((ax_u64)(new_size));
        return p;
    }
    void* new_ptr = ax_ax_alloc(new_size);
    if ((new_ptr == ((void*)(NULL)))) {
        return ((void*)(NULL));
    }
    ax_i64 copy_size = old_size;
    if ((new_size < old_size)) {
        copy_size = new_size;
    }
    memcpy(((ax_u8*)(new_ptr)), ((ax_u8*)(p)), copy_size);
    ax_ax_free(p);
    return new_ptr;
    ax_free(new_ptr);
}

ax_i64 ax_ax_alloc_size(void* p) {
    if ((p == ((void*)(NULL)))) {
        return ((ax_i64)(0));
    }
    struct ax_AxHeader* block = ((struct ax_AxHeader*)((((ax_i64)(p)) - ((ax_i64)(16)))));
    return ((ax_i64)(block->size));
}

ax_u8* ax_ax_numa_alloc(ax_i64 size, ax_i32 node_id) {
    ax_i64 aligned_size = (((size + 63) / 64) * 64);
    return ((ax_u8*)(ax_ax_alloc(aligned_size)));
}

void ax_AxRunQueue_ax_runq_init(struct ax_AxRunQueue* q) {
    q->top = ((ax_u64)(0));
    q->bottom = ((ax_u64)(0));
    q->buffer = ((ax_u64*)(ax_ax_alloc((((ax_i64)(4096)) * ((ax_i64)(8))))));
}

ax_i32 ax_AxRunQueue_ax_runq_push(struct ax_AxRunQueue* q, ax_u64 id) {
    ax_u64 b = q->bottom;
    ax_u64 t = __atomic_load_n(&(q->top), __ATOMIC_SEQ_CST);
    if (((b - t) >= ((ax_u64)(4096)))) {
        return (-1);
    }
    ax_i64 idx = ((ax_i64)((b & ((ax_u64)(4095)))));
    ax_u64* p = ((ax_u64*)((((ax_i64)(q->buffer)) + (idx * ((ax_i64)(8))))));
    (*((ax_u64*)(p))) = id;
    __atomic_store_n(&(q->bottom), (b + ((ax_u64)(1))), __ATOMIC_SEQ_CST);
    return 0;
}

ax_u64 ax_AxRunQueue_ax_runq_pop(struct ax_AxRunQueue* q) {
    ax_u64 b = q->bottom;
    if ((b == ((ax_u64)(0)))) {
        return ((ax_u64)(0));
    }
    b = (b - ((ax_u64)(1)));
    __atomic_store_n(&(q->bottom), b, __ATOMIC_SEQ_CST);
    ax_u64 t = __atomic_load_n(&(q->top), __ATOMIC_SEQ_CST);
    if ((t > b)) {
        __atomic_store_n(&(q->bottom), t, __ATOMIC_SEQ_CST);
        return ((ax_u64)(0));
    }
    ax_i64 idx = ((ax_i64)((b & ((ax_u64)(4095)))));
    ax_u64* p = ((ax_u64*)((((ax_i64)(q->buffer)) + (idx * ((ax_i64)(8))))));
    ax_u64 id = (*((ax_u64*)(p)));
    if ((t == b)) {
        if ((!__sync_bool_compare_and_swap(&(q->top), t, (t + ((ax_u64)(1)))))) {
            id = ((ax_u64)(0));
        }
        __atomic_store_n(&(q->bottom), (t + ((ax_u64)(1))), __ATOMIC_SEQ_CST);
    }
    return id;
}

ax_u64 ax_AxRunQueue_ax_runq_steal(struct ax_AxRunQueue* q) {
    ax_u64 t = __atomic_load_n(&(q->top), __ATOMIC_SEQ_CST);
    ax_u64 b = __atomic_load_n(&(q->bottom), __ATOMIC_SEQ_CST);
    if ((t >= b)) {
        return ((ax_u64)(0));
    }
    ax_i64 idx = ((ax_i64)((t & ((ax_u64)(4095)))));
    ax_u64* p = ((ax_u64*)((((ax_i64)(q->buffer)) + (idx * ((ax_i64)(8))))));
    ax_u64 id = (*((ax_u64*)(p)));
    if (__sync_bool_compare_and_swap(&(q->top), t, (t + ((ax_u64)(1))))) {
        return id;
    }
    return ((ax_u64)(0));
}

ax_i32 ax_AxRunQueue_ax_runq_empty(struct ax_AxRunQueue* q) {
    if ((__atomic_load_n(&(q->top), __ATOMIC_SEQ_CST) >= __atomic_load_n(&(q->bottom), __ATOMIC_SEQ_CST))) {
        return 1;
    }
    return 0;
}

ax_i32 ax_AxScheduler_ax_scheduler_init(struct ax_AxScheduler* sched, ax_u32 worker_count) {
    sched->worker_count = worker_count;
    sched->running = 0;
    sched->total_submitted = ((ax_u64)(0));
    ax_i64 worker_size = ((ax_i64)(80));
    sched->workers = ((struct ax_AxWorker*)(ax_ax_alloc((((ax_i64)(256)) * worker_size))));
    ax_i64 i = ((ax_i64)(0));
    while ((i < ((ax_i64)(256)))) {
        struct ax_AxWorker* w = ((struct ax_AxWorker*)((((ax_i64)(sched->workers)) + (i * worker_size))));
        w->id = ((ax_u32)(i));
        w->tasks_executed = ((ax_u64)(0));
        w->steals_attempted = ((ax_u64)(0));
        w->steals_succeeded = ((ax_u64)(0));
        w->running = 0;
        ax_AxRunQueue_ax_runq_init(&(w->runq));
        i = (i + 1);
    }
    return 0;
}

ax_i32 ax_AxScheduler_ax_scheduler_submit(struct ax_AxScheduler* sched, ax_u64 actor_id) {
    sched->total_submitted = (sched->total_submitted + ((ax_u64)(1)));
    struct ax_AxWorker* w = sched->workers;
    return ax_AxRunQueue_ax_runq_push(&(w->runq), actor_id);
}

ax_i32 ax_AxScheduler_ax_scheduler_run(struct ax_AxScheduler* sched) {
    sched->running = 1;
    struct ax_AxWorker* w = sched->workers;
    w->running = 1;
    return 0;
}

void ax_AxScheduler_ax_scheduler_stats(struct ax_AxScheduler* sched, struct ax_AxSchedulerStats* stats) {
    stats->worker_count = sched->worker_count;
    stats->total_submitted = sched->total_submitted;
    struct ax_AxWorker* w = sched->workers;
    stats->total_executed = w->tasks_executed;
    stats->total_steals = w->steals_succeeded;
}

void ax_AxScheduler_ax_scheduler_shutdown(struct ax_AxScheduler* sched) {
    sched->running = 0;
    struct ax_AxWorker* w = sched->workers;
    w->running = 0;
}

static void ax_test_print_str(ax_string s) {
    ax_sys_write(1, ((void*)(s.ptr)), ((ax_u64)(s.len)));
}

static void ax_test_allocator(void) {
    void* p1 = ax_ax_alloc(((ax_i64)(10)));
    ax_assert_axiom((p1 != ((void*)(NULL))), AX_STR("(p1 != ((void*)(NULL)))"));
    struct ax_AxHeader* hdr1 = ((struct ax_AxHeader*)((((ax_i64)(p1)) - ((ax_i64)(16)))));
    ax_assert_axiom((hdr1->gen_id == ((ax_u64)(1))), AX_STR("(hdr1->gen_id == ((ax_u64)(1)))"));
    ax_assert_axiom((hdr1->size == ((ax_u64)(10))), AX_STR("(hdr1->size == ((ax_u64)(10)))"));
    void* p2 = ax_ax_realloc(p1, ((ax_i64)(20)));
    ax_assert_axiom((p2 != ((void*)(NULL))), AX_STR("(p2 != ((void*)(NULL)))"));
    struct ax_AxHeader* hdr2 = ((struct ax_AxHeader*)((((ax_i64)(p2)) - ((ax_i64)(16)))));
    ax_assert_axiom((hdr2->size == ((ax_u64)(20))), AX_STR("(hdr2->size == ((ax_u64)(20)))"));
    ax_ax_free(p2);
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_allocator\n", .len=23});
}

static void ax_test_genref(void) {
    void* p = ax_ax_alloc(((ax_i64)(32)));
    struct ax_AxRef ref = ax_ax_make_ref(p);
    ax_assert_axiom((ref.ptr == p), AX_STR("(ref.ptr == p)"));
    ax_assert_axiom((ref.gen_id == ((ax_u64)(1))), AX_STR("(ref.gen_id == ((ax_u64)(1)))"));
    void* p_checked = ax_AxRef_ax_deref(ref);
    ax_assert_axiom((p_checked == p), AX_STR("(p_checked == p)"));
    ax_assert_axiom(ax_AxRef_ax_ref_valid(ref), AX_STR("ax_AxRef_ax_ref_valid(ref)"));
    ax_ax_invalidate(p);
    ax_assert_axiom((!ax_AxRef_ax_ref_valid(ref)), AX_STR("(!ax_AxRef_ax_ref_valid(ref))"));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_genref\n", .len=20});
}

ax_i32 ax_main_usr(void) {
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native Freestanding Runtime unit tests...\n", .len=56});
    ax_test_allocator();
    ax_test_genref();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"All AXIOM-native Freestanding Runtime tests passed!\n", .len=52});
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
