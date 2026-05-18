# p15-t05: Actor Message Queue

## Purpose
Implement the per-actor message queue (mailbox) as a lock-free MPSC (multi-producer, single-consumer) queue enabling multiple senders to enqueue messages to one actor safely.

## Context
Each actor has one mailbox. Multiple other actors (and threads) may send messages simultaneously (multi-producer), but only the actor itself dequeues and processes messages (single consumer). A lock-free MPSC queue is optimal: producers use CAS, consumer is wait-free.

## Inputs
- Message envelopes: `{type_id, payload_ptr, payload_size, is_isolated}`
- `AxIsolated` from p15-t04 for isolated message type
- Actor heap from p14-t04 for message payload allocation

## Outputs
- `runtime/msgqueue.h` — `AxMsgQueue` struct
- `runtime/msgqueue.c` — MPSC queue implementation
- `ax_msgq_enqueue()`, `ax_msgq_dequeue()` API

## Dependencies
- p15-t01: actor-struct — embeds AxMsgQueue in AxActor
- p15-t04: isolated-type-runtime — Isolated messages in queue
- p14-t04: axalloc-actor-heap — message payload memory

## Subsystems Affected
- Scheduler (p15-t02): checks mailbox non-empty before removing actor from run queue
- Actor step (p15-t01): calls `ax_msgq_dequeue()` in message loop

## Detailed Requirements

```c
typedef struct AxMsg {
    uint32_t  type_id;
    uint32_t  flags;          // AX_MSG_ISOLATED = 0x1
    void*     payload;
    size_t    payload_size;
    AxIsolated isolated;      // valid if flags & AX_MSG_ISOLATED
} AxMsg;

typedef struct AxMsgNode {
    AxMsg            msg;
    _Atomic(struct AxMsgNode*) next;
} AxMsgNode;

typedef struct AxMsgQueue {
    _Atomic(AxMsgNode*) head;   // consumer reads from head
    _Atomic(AxMsgNode*) tail;   // producers append to tail
    _Atomic uint32_t    count;
    uint32_t            max_size;  // 0 = unlimited
} AxMsgQueue;

// Thread-safe (any thread)
int  ax_msgq_enqueue(AxMsgQueue* q, AxMsg msg);

// Single-consumer (actor thread only)
int  ax_msgq_dequeue(AxMsgQueue* q, AxMsg* out);

bool ax_msgq_empty(AxMsgQueue* q);
uint32_t ax_msgq_count(AxMsgQueue* q);
void ax_msgq_init(AxMsgQueue* q, uint32_t max_size);
void ax_msgq_destroy(AxMsgQueue* q);
```

MPSC algorithm (Michael-Scott queue variant):
- `enqueue`: allocate node, set next=NULL, CAS tail->next from NULL to node, update tail.
- `dequeue`: read head->next; if non-NULL, copy msg, free old head, update head.
- Consumer never does CAS; only reads and pointer updates.

Message memory:
- Payload copied into actor's own heap on enqueue (from sender's allocator).
- Exception: `Isolated` messages — payload NOT copied (segments transferred).

Backpressure: if `count >= max_size`, `enqueue` returns -1 (queue full).

## Implementation Steps

1. Create `runtime/msgqueue.h` and `runtime/msgqueue.c`.
2. Implement MPSC queue with atomic operations.
3. Implement `ax_msgq_enqueue()` with payload copy (non-isolated) or segment transfer (isolated).
4. Implement `ax_msgq_dequeue()` — wait-free for consumer.
5. Implement backpressure.
6. Wire into AxActor struct.
7. Write concurrency stress tests.

## Test Plan
- `TestMsgQueueBasic`: enqueue + dequeue single message
- `TestMsgQueueMPSC`: 16 producers × 1 consumer, 10K messages each → no loss
- `TestMsgQueueIsolated`: isolated message → no copy, segments transferred
- `TestMsgQueueBackpressure`: full queue → enqueue returns -1
- `TestMsgQueueEmpty`: empty check before dequeue

## Validation Checklist
- [ ] No messages lost under concurrent 16-producer test
- [ ] Consumer never uses CAS (wait-free consumer)
- [ ] Isolated messages: no memcpy of payload
- [ ] Backpressure at max_size boundary

## Acceptance Criteria
- 10M messages/sec through single actor mailbox on 4 cores

## Definition of Done
- [ ] `runtime/msgqueue.c` implemented
- [ ] MPSC stress test passes (no lost messages after 1M enqueue/dequeue)

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| ABA problem in Michael-Scott queue | Use split-reference counting or hazard pointers for node reclaim |
| Message node allocator contention | Use per-producer free list for node allocation |

## Future Follow-up Tasks
- Priority message queue (for system messages like Stop)
- Dead letter queue: undeliverable messages logged to supervisor
