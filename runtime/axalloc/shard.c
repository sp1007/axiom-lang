/*
 * p14-t03: AxAlloc Free-List Sharding — Implementation
 */

#include "shard.h"
#include <string.h>

/* --------------------------------------------------------------------------
 * Thread ID Hashing
 * -------------------------------------------------------------------------- */

#ifdef _WIN32
  #include <processthreadsapi.h>
  static uint32_t get_thread_hash(void) {
      return (uint32_t)GetCurrentThreadId();
  }
#else
  #include <pthread.h>
  static uint32_t get_thread_hash(void) {
      return (uint32_t)(uintptr_t)pthread_self();
  }
#endif

static uint32_t hash_to_shard(uint32_t tid) {
    /* FNV-1a-like mixing */
    tid ^= tid >> 16;
    tid *= 0x45d9f3b;
    tid ^= tid >> 16;
    return tid & (NUM_SHARDS - 1);
}

/* --------------------------------------------------------------------------
 * Init / Destroy
 * -------------------------------------------------------------------------- */

void ax_shard_init(ShardedAllocator* sa) {
    memset(sa, 0, sizeof(ShardedAllocator));
    for (int i = 0; i < NUM_SHARDS; i++) {
        AX_MUTEX_INIT(sa->shards[i].lock);
    }
}

void ax_shard_destroy(ShardedAllocator* sa) {
    for (int i = 0; i < NUM_SHARDS; i++) {
        AX_MUTEX_DESTROY(sa->shards[i].lock);
    }
}

/* --------------------------------------------------------------------------
 * Shard Selection
 * -------------------------------------------------------------------------- */

AllocShard* ax_shard_for_thread(ShardedAllocator* sa) {
    uint32_t tid = get_thread_hash();
    uint32_t idx = hash_to_shard(tid);
    return &sa->shards[idx];
}

/* --------------------------------------------------------------------------
 * Thread-Safe Alloc / Free
 * -------------------------------------------------------------------------- */

void* ax_shard_alloc(ShardedAllocator* sa, SizeClass sc) {
    if (sc >= NUM_SIZE_CLASSES) return NULL;

    AllocShard* shard = ax_shard_for_thread(sa);
    AX_MUTEX_LOCK(shard->lock);

    void* block = ax_free_list_pop(&shard->lists[sc]);
    if (block) {
        shard->alloc_count++;
    }

    AX_MUTEX_UNLOCK(shard->lock);
    return block;
}

void ax_shard_free(ShardedAllocator* sa, void* block, SizeClass sc) {
    if (!block || sc >= NUM_SIZE_CLASSES) return;

    AllocShard* shard = ax_shard_for_thread(sa);
    AX_MUTEX_LOCK(shard->lock);

    ax_free_list_push(&shard->lists[sc], block);
    shard->free_count++;

    AX_MUTEX_UNLOCK(shard->lock);
}
