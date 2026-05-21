/*
 * p14-t03: AxAlloc Free-List Sharding
 *
 * Thread-safe sharded free lists for concurrent allocation.
 * Each shard has its own lock, reducing contention.
 * Shard selection uses thread-ID hashing for locality.
 */

#ifndef AXIOM_AXALLOC_SHARD_H
#define AXIOM_AXALLOC_SHARD_H

#include "size_classes.h"

#ifdef _WIN32
  #include <windows.h>
  typedef CRITICAL_SECTION ax_mutex_t;
  #define AX_MUTEX_INIT(m)    InitializeCriticalSection(&(m))
  #define AX_MUTEX_DESTROY(m) DeleteCriticalSection(&(m))
  #define AX_MUTEX_LOCK(m)    EnterCriticalSection(&(m))
  #define AX_MUTEX_UNLOCK(m)  LeaveCriticalSection(&(m))
#else
  #include <pthread.h>
  typedef pthread_mutex_t ax_mutex_t;
  #define AX_MUTEX_INIT(m)    pthread_mutex_init(&(m), NULL)
  #define AX_MUTEX_DESTROY(m) pthread_mutex_destroy(&(m))
  #define AX_MUTEX_LOCK(m)    pthread_mutex_lock(&(m))
  #define AX_MUTEX_UNLOCK(m)  pthread_mutex_unlock(&(m))
#endif

#ifdef __cplusplus
extern "C" {
#endif

#define NUM_SHARDS 8

/* A single shard: one free list per size class + a lock */
typedef struct {
    FreeList  lists[NUM_SIZE_CLASSES];
    ax_mutex_t lock;
    uint64_t   alloc_count;
    uint64_t   free_count;
} AllocShard;

/* Sharded allocator: NUM_SHARDS shards */
typedef struct {
    AllocShard shards[NUM_SHARDS];
} ShardedAllocator;

/** Initialize a sharded allocator. */
void ax_shard_init(ShardedAllocator* sa);

/** Destroy a sharded allocator. */
void ax_shard_destroy(ShardedAllocator* sa);

/** Get the shard for the current thread. */
AllocShard* ax_shard_for_thread(ShardedAllocator* sa);

/** Thread-safe alloc from the appropriate shard's free list. */
void* ax_shard_alloc(ShardedAllocator* sa, SizeClass sc);

/** Thread-safe free into the appropriate shard's free list. */
void ax_shard_free(ShardedAllocator* sa, void* block, SizeClass sc);

#ifdef __cplusplus
}
#endif

#endif /* AXIOM_AXALLOC_SHARD_H */
