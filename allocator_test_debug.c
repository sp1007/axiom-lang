#define AX_EMIT_MAIN
#include "ax_runtime.h"
#include "ax_stdlib.h"

/* Forward declarations */
struct ax_AxHeader;
struct ax_FreeSlot;
struct ax_FreeList;
struct ax_Segment;
struct ax_SegmentList;
struct ax_ActorHeap;
struct ax_AxGlobalState;

/* Type definitions */
struct ax_AxHeader {
    ax_u32 gen_id;
    ax_u32 flags;
};
struct ax_FreeSlot {
    struct ax_FreeSlot* next;
};
struct ax_FreeList {
    struct ax_FreeSlot* head;
    ax_i64 count;
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
void* ax_mmap(void* addr, ax_u64 length, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset);
ax_i32 ax_munmap(void* addr, ax_u64 length);
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
ax_i64 syscall(ax_u64 num, ax_u64 a1, ax_u64 a2, ax_u64 a3, ax_u64 a4, ax_u64 a5, ax_u64 a6);
static void ax_test_print_str(ax_string s);
static void ax_test_size_class_for(void);
static void ax_test_free_list_push_pop(void);
static void ax_test_segment_bump_alloc(void);
static void ax_test_actor_heap_lifecycle_and_alloc(void);
ax_i32 ax_main_usr(void);


void* ax_mmap(void* addr, ax_u64 length, ax_i32 prot, ax_i32 flags, ax_i32 fd, ax_i64 offset) {
    return ((void*)(NULL));
}

ax_i32 ax_munmap(void* addr, ax_u64 length) {
    return ((ax_i32)(0));
}

struct ax_AxGlobalState* ax_get_global_state(void) {
    void* addr = ((void*)(0x50000000));
    struct ax_AxGlobalState* state = ((struct ax_AxGlobalState*)(VirtualAlloc(addr, ((ax_u64)(4096)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)))));
    if ((state == ((struct ax_AxGlobalState*)(NULL)))) {
        return ((struct ax_AxGlobalState*)(0x50000000));
    }
    return state;
    ax_free(state);
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
    if ((state->g_slab == ((struct ax_Segment*)(NULL)))) {
        state->g_slab = ((struct ax_Segment*)(VirtualAlloc(((void*)(NULL)), (((ax_u64)(4096)) * ((ax_u64)(48))), ((ax_u32)(0x3000)), ((ax_u32)(0x04)))));
    }
    return state->g_slab;
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

void* ax_ax_os_alloc(ax_i64 size) {
    if (1) {
        return VirtualAlloc(((void*)(NULL)), ((ax_u64)(size)), ((ax_u32)(0x3000)), ((ax_u32)(0x04)));
    } else {
        {
            return ax_mmap(((void*)(NULL)), ((ax_u64)(size)), ((ax_i32)(3)), ((ax_i32)(0x22)), (-((ax_i32)(1))), ((ax_i64)(0)));
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
    (*(slab_used_ptr)) = 0;
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    (*(free_pool_ptr)) = ((struct ax_Segment*)(NULL));
    struct ax_Segment* slab_ptr = ax_std_mem_alloc_get_slab();
    memset(((ax_u8*)(slab_ptr)), ((ax_u8)(0)), (((ax_i64)(4096)) * ((ax_i64)(sizeof(struct ax_Segment)))));
}

void ax_ax_segment_manager_shutdown(void) {
    ax_i64* slab_used_ptr = ax_std_mem_alloc_get_slab_used();
    struct ax_Segment* slab_ptr = ax_std_mem_alloc_get_slab();
    ax_i64 i = ((ax_i64)(0));
    while ((i < (*(slab_used_ptr)))) {
        struct ax_Segment* seg = ((struct ax_Segment*)((((ax_i64)(slab_ptr)) + (i * ((ax_i64)(sizeof(struct ax_Segment)))))));
        if ((seg->magic == ax_SEGMENT_MAGIC)) {
            ax_ax_os_free(((void*)(seg->base)), ax_SEGMENT_SIZE);
            seg->magic = ((ax_u32)(0));
        }
        i = (i + 1);
    }
    (*(slab_used_ptr)) = 0;
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    (*(free_pool_ptr)) = ((struct ax_Segment*)(NULL));
}

static struct ax_Segment* ax_alloc_segment_meta(void) {
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    if (((*(free_pool_ptr)) != ((struct ax_Segment*)(NULL)))) {
        struct ax_Segment* seg = (*(free_pool_ptr));
        (*(free_pool_ptr)) = seg->next;
        return seg;
        ax_free(seg);
    }
    ax_i64* slab_used_ptr = ax_std_mem_alloc_get_slab_used();
    if (((*(slab_used_ptr)) >= ((ax_i64)(4096)))) {
        return ((struct ax_Segment*)(NULL));
    }
    struct ax_Segment* slab_ptr = ax_std_mem_alloc_get_slab();
    struct ax_Segment* seg = ((struct ax_Segment*)((((ax_i64)(slab_ptr)) + ((*(slab_used_ptr)) * ((ax_i64)(sizeof(struct ax_Segment)))))));
    (*(slab_used_ptr)) = ((*(slab_used_ptr)) + 1);
    return seg;
    ax_free(seg);
}

static void ax_Segment_free_segment_meta(struct ax_Segment* seg) {
    memset(((ax_u8*)(seg)), ((ax_u8)(0)), sizeof(struct ax_Segment));
    struct ax_Segment** free_pool_ptr = ax_std_mem_alloc_get_free_pool();
    seg->next = (*(free_pool_ptr));
    (*(free_pool_ptr)) = seg;
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
    if (((seg == ((struct ax_Segment*)(NULL))) || (seg->magic != ax_SEGMENT_MAGIC))) {
        return;
    }
    ax_ax_os_free(((void*)(seg->base)), ax_SEGMENT_SIZE);
    seg->magic = ((ax_u32)(0));
    ax_Segment_free_segment_meta(seg);
}

struct ax_Segment* ax_SegmentList_ax_segment_get_active(struct ax_SegmentList* list, ax_i32 sc) {
    if (((list->active != ((struct ax_Segment*)(NULL))) && (list->active->bump < list->active->limit))) {
        return list->active;
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
    if (((seg == ((struct ax_Segment*)(NULL))) || (sc >= 10))) {
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
    ax_i64 total = (8 + user_size);
    ax_i64 page_aligned = (((total + 4095) / 4096) * 4096);
    ax_u8* block = ((ax_u8*)(ax_ax_os_alloc(page_aligned)));
    if ((block == ((ax_u8*)(NULL)))) {
        return ((ax_u8*)(NULL));
    }
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)(block));
    hdr->gen_id = ((ax_u32)(1));
    hdr->flags = ((ax_u32)(10));
    return ((ax_u8*)((((ax_i64)(block)) + 8)));
}

void ax_ax_large_free(ax_u8* user_ptr, ax_i64 user_size) {
    if ((user_ptr == ((ax_u8*)(NULL)))) {
        return;
    }
    ax_u8* block = ((ax_u8*)((((ax_i64)(user_ptr)) - 8)));
    ax_i64 total = (8 + user_size);
    ax_i64 page_aligned = (((total + 4095) / 4096) * 4096);
    ax_ax_os_free(((void*)(block)), page_aligned);
}

struct ax_ActorHeap* ax_ax_actor_heap_create(ax_u64 actor_id) {
    ax_i64 size = ((ax_i64)(sizeof(struct ax_ActorHeap)));
    ax_i64 page_aligned = (((size + 4095) / 4096) * 4096);
    struct ax_ActorHeap* heap = ((struct ax_ActorHeap*)(ax_ax_os_alloc(page_aligned)));
    if ((heap == ((struct ax_ActorHeap*)(NULL)))) {
        return ((struct ax_ActorHeap*)(NULL));
    }
    memset(((ax_u8*)(heap)), ((ax_u8)(0)), page_aligned);
    heap->actor_id = actor_id;
    heap->magic = ax_ACTOR_HEAP_MAGIC;
    return heap;
    ax_free(heap);
}

void ax_ActorHeap_ax_actor_heap_destroy(struct ax_ActorHeap* heap) {
    if (((heap == ((struct ax_ActorHeap*)(NULL))) || (heap->magic != ax_ACTOR_HEAP_MAGIC))) {
        return;
    }
    ax_i32 sc = 0;
    while ((sc < 10)) {
        struct ax_SegmentList* list = ax_ActorHeap_ax_get_segment_list(heap, sc);
        ax_SegmentList_ax_segment_list_release_all(list);
        sc = (sc + 1);
    }
    heap->magic = ((ax_u32)(0));
    ax_i64 size = ((ax_i64)(sizeof(struct ax_ActorHeap)));
    ax_i64 page_aligned = (((size + 4095) / 4096) * 4096);
    ax_ax_os_free(((void*)(heap)), page_aligned);
}

ax_u8* ax_ActorHeap_ax_actor_alloc(struct ax_ActorHeap* heap, ax_i64 user_size) {
    if (((heap == ((struct ax_ActorHeap*)(NULL))) || (heap->magic != ax_ACTOR_HEAP_MAGIC))) {
        return ((ax_u8*)(NULL));
    }
    ax_i32 sc = ax_ax_size_class_for(user_size);
    if ((sc == 10)) {
        ax_u8* ptr_val = ax_ax_large_alloc(user_size);
        if ((ptr_val != ((ax_u8*)(NULL)))) {
            heap->total_allocated = (heap->total_allocated + ((ax_u64)(user_size)));
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
    if ((((heap == ((struct ax_ActorHeap*)(NULL))) || (user_ptr == ((ax_u8*)(NULL)))) || (heap->magic != ax_ACTOR_HEAP_MAGIC))) {
        return;
    }
    ax_u8* block = ((ax_u8*)((((ax_i64)(user_ptr)) - 8)));
    struct ax_AxHeader* hdr = ((struct ax_AxHeader*)(block));
    ax_i32 sc = ((ax_i32)((hdr->flags & ((ax_u32)(15)))));
    hdr->gen_id = ((ax_u32)(0));
    if (((sc == 10) || (sc >= 10))) {
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

static void ax_test_print_str(ax_string s) {
    if (1) {
        void* h = GetStdHandle(((ax_u32)(0xFFFFFFF5)));
        ax_u32 written = ((ax_u32)(0));
        WriteFile(h, ((ax_u8*)(s.ptr)), ((ax_u32)(ax_str_len(s))), ((void*)(&(written))), ((void*)(NULL)));
    } else {
        {
            syscall(((ax_u64)(1)), ((ax_u64)(1)), ((ax_u64)(((ax_u8*)(s.ptr)))), ((ax_u64)(ax_str_len(s))), ((ax_u64)(0)), ((ax_u64)(0)), ((ax_u64)(0)));
        }
    }
}

static void ax_test_size_class_for(void) {
    ax_assert_axiom((ax_ax_size_class_for(0) == 0), AX_STR("(ax_ax_size_class_for(0) == 0)"));
    ax_assert_axiom((ax_ax_size_class_for(1) == 1), AX_STR("(ax_ax_size_class_for(1) == 1)"));
    ax_assert_axiom((ax_ax_size_class_for(8) == 1), AX_STR("(ax_ax_size_class_for(8) == 1)"));
    ax_assert_axiom((ax_ax_size_class_for(9) == 2), AX_STR("(ax_ax_size_class_for(9) == 2)"));
    ax_assert_axiom((ax_ax_size_class_for(56) == 3), AX_STR("(ax_ax_size_class_for(56) == 3)"));
    ax_assert_axiom((ax_ax_size_class_for(57) == 4), AX_STR("(ax_ax_size_class_for(57) == 4)"));
    ax_assert_axiom((ax_ax_size_class_for(4088) == 9), AX_STR("(ax_ax_size_class_for(4088) == 9)"));
    ax_assert_axiom((ax_ax_size_class_for(4089) == 10), AX_STR("(ax_ax_size_class_for(4089) == 10)"));
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_size_class_for\n", .len=28});
}

static void ax_test_free_list_push_pop(void) {
    ax_u8* raw = ((ax_u8*)(ax_ax_os_alloc(256)));
    ax_assert_axiom((raw != ((ax_u8*)(NULL))), AX_STR("(raw != ((ax_u8*)(NULL)))"));
    ax_u8* block0 = raw;
    ax_u8* block1 = ((ax_u8*)((((ax_i64)(raw)) + 32)));
    ax_u8* block2 = ((ax_u8*)((((ax_i64)(raw)) + 64)));
    struct ax_FreeList list = ((struct ax_FreeList){.head=((struct ax_FreeSlot*)(NULL)), .count=0});
    ax_FreeList_ax_free_list_push(&(list), block0);
    ax_FreeList_ax_free_list_push(&(list), block1);
    ax_FreeList_ax_free_list_push(&(list), block2);
    ax_assert_axiom((list.count == 3), AX_STR("(list.count == 3)"));
    ax_u8* b2 = ax_FreeList_ax_free_list_pop(&(list));
    ax_u8* b1 = ax_FreeList_ax_free_list_pop(&(list));
    ax_u8* b0 = ax_FreeList_ax_free_list_pop(&(list));
    ax_assert_axiom((b2 == block2), AX_STR("(b2 == block2)"));
    ax_assert_axiom((b1 == block1), AX_STR("(b1 == block1)"));
    ax_assert_axiom((b0 == block0), AX_STR("(b0 == block0)"));
    ax_assert_axiom((list.count == 0), AX_STR("(list.count == 0)"));
    ax_u8* b_null = ax_FreeList_ax_free_list_pop(&(list));
    ax_assert_axiom((b_null == ((ax_u8*)(NULL))), AX_STR("(b_null == ((ax_u8*)(NULL)))"));
    ax_ax_os_free(((void*)(raw)), 256);
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_free_list_push_pop\n", .len=32});
}

static void ax_test_segment_bump_alloc(void) {
    ax_ax_segment_manager_init();
    struct ax_Segment* seg = ax_ax_segment_acquire(2);
    ax_assert_axiom((seg != ((struct ax_Segment*)(NULL))), AX_STR("(seg != ((struct ax_Segment*)(NULL)))"));
    ax_u8* b1 = ax_Segment_ax_segment_bump_alloc(seg, 2);
    ax_assert_axiom((b1 != ((ax_u8*)(NULL))), AX_STR("(b1 != ((ax_u8*)(NULL)))"));
    ax_u8* b2 = ax_Segment_ax_segment_bump_alloc(seg, 2);
    ax_assert_axiom((b2 != ((ax_u8*)(NULL))), AX_STR("(b2 != ((ax_u8*)(NULL)))"));
    ax_assert_axiom(((((ax_i64)(b2)) - ((ax_i64)(b1))) == 32), AX_STR("((((ax_i64)(b2)) - ((ax_i64)(b1))) == 32)"));
    ax_Segment_ax_segment_release(seg);
    ax_ax_segment_manager_shutdown();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_segment_bump_alloc\n", .len=32});
}

static void ax_test_actor_heap_lifecycle_and_alloc(void) {
    ax_ax_segment_manager_init();
    struct ax_ActorHeap* heap = ax_ax_actor_heap_create(((ax_u64)(42)));
    ax_assert_axiom((heap != ((struct ax_ActorHeap*)(NULL))), AX_STR("(heap != ((struct ax_ActorHeap*)(NULL)))"));
    ax_assert_axiom((heap->actor_id == ((ax_u64)(42))), AX_STR("(heap->actor_id == ((ax_u64)(42)))"));
    ax_assert_axiom((heap->magic == ax_ACTOR_HEAP_MAGIC), AX_STR("(heap->magic == ax_ACTOR_HEAP_MAGIC)"));
    ax_u8* p1 = ax_ActorHeap_ax_actor_alloc(heap, 4);
    ax_assert_axiom((p1 != ((ax_u8*)(NULL))), AX_STR("(p1 != ((ax_u8*)(NULL)))"));
    ax_u8* p2 = ax_ActorHeap_ax_actor_alloc(heap, 20);
    ax_assert_axiom((p2 != ((ax_u8*)(NULL))), AX_STR("(p2 != ((ax_u8*)(NULL)))"));
    ax_u8* p3 = ax_ActorHeap_ax_actor_alloc(heap, 5000);
    ax_assert_axiom((p3 != ((ax_u8*)(NULL))), AX_STR("(p3 != ((ax_u8*)(NULL)))"));
    ax_assert_axiom((heap->total_allocated == ((((ax_u64)(16)) + ((ax_u64)(32))) + ((ax_u64)(5000)))), AX_STR("(heap->total_allocated == ((((ax_u64)(16)) + ((ax_u64)(32))) + ((ax_u64)(5000))))"));
    ax_assert_axiom((heap->alloc_count == ((ax_u64)(3))), AX_STR("(heap->alloc_count == ((ax_u64)(3)))"));
    ax_ActorHeap_ax_actor_free(heap, p1);
    ax_ActorHeap_ax_actor_free(heap, p2);
    ax_ActorHeap_ax_actor_free(heap, p3);
    struct ax_AxHeader* hdr1 = ((struct ax_AxHeader*)((((ax_i64)(p1)) - 8)));
    ax_assert_axiom((hdr1->gen_id == ((ax_u32)(0))), AX_STR("(hdr1->gen_id == ((ax_u32)(0)))"));
    struct ax_AxHeader* hdr2 = ((struct ax_AxHeader*)((((ax_i64)(p2)) - 8)));
    ax_assert_axiom((hdr2->gen_id == ((ax_u32)(0))), AX_STR("(hdr2->gen_id == ((ax_u32)(0)))"));
    ax_ActorHeap_ax_actor_heap_destroy(heap);
    ax_ax_segment_manager_shutdown();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"  PASS: test_actor_heap_lifecycle_and_alloc\n", .len=44});
}

ax_i32 ax_main_usr(void) {
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"Running AXIOM-native Allocator unit tests...\n", .len=45});
    ax_test_size_class_for();
    ax_test_free_list_push_pop();
    ax_test_segment_bump_alloc();
    ax_test_actor_heap_lifecycle_and_alloc();
    ax_test_print_str((ax_string){.ptr=(const ax_u8*)"All AXIOM-native Allocator tests passed!\n", .len=41});
    return 0;
}

/* Entry point wrapper */
ax_i32 ax_main(void) {
    return ax_main_usr();
}
