# p15-t09: Actor Codegen Integration

## Purpose
Wire the AXIOM compiler to generate correct runtime calls for actor syntax — `spawn`, `send`, `receive`, `become` — transforming actor declarations into `ax_actor_spawn()`, `ax_actor_send()`, and state machine handler functions.

## Context
Phases 15-t01 through t08 implement the runtime. This task implements the compiler side: how `actor MyActor { fn handle(msg: Msg) {...} }` in AXIOM source becomes `ax_actor_spawn(my_actor_handler, ...)` in generated code. The codegen bridges the language semantics to the runtime API.

## Inputs
- Actor declarations from TypedAST (p04)
- Message type definitions from TypeTable
- AIR builder (p09) — emit OpActorSpawn, OpActorSend opcodes
- Runtime headers: `runtime/actor.h`, `runtime/msgqueue.h`

## Outputs
- `codegen/actor_codegen.go` — actor syntax → runtime call codegen
- New AIR opcodes: `OpActorSpawn`, `OpActorSend`, `OpActorSelf`
- Generated handler function: `void _AX_MyModule_MyActor_handler(AxActor*, void*, uint32_t)`

## Dependencies
- p09-t01: air-instruction-set — add actor opcodes to opcode table
- p15-t01: actor-struct — `ax_actor_spawn()` signature
- p15-t05: actor-message-queue — `ax_actor_send()` signature
- p04: type checker — actor declaration semantic validation

## Subsystems Affected
- AIR instruction set: new actor opcodes
- C backend (p10-t10): lower OpActorSpawn to `ax_actor_spawn()` call
- Native backend (p11): lower OpActorSpawn to CALL instruction

## Detailed Requirements

New AIR opcodes:
```go
const (
    OpActorSpawn  = 0x0601  // %id = actor_spawn handler_fn, init_state, supervisor
    OpActorSend   = 0x0602  // actor_send %actor_id, %msg, type_id
    OpActorSelf   = 0x0603  // %id = actor_self (current actor's ID)
    OpActorBecome = 0x0604  // actor_become %new_state (update handler state)
)
```

Actor declaration lowering:
```axiom
actor Counter:
    var count: i32 = 0
    fn handle(msg: CounterMsg):
        match msg:
            Increment -> count += 1
            GetCount(reply_to) -> send reply_to, count
```

Generated C:
```c
// Handler function (compiled by AXIOM)
void _AX_main_Counter_handler(AxActor* self, void* msg, uint32_t msg_type) {
    CounterState* state = (CounterState*)self->state_data;
    switch (msg_type) {
    case MSG_Increment: state->count++; break;
    case MSG_GetCount: {
        GetCountMsg* m = (GetCountMsg*)msg;
        ax_actor_send(m->reply_to, &state->count, MSG_Count, sizeof(i32));
        break;
    }
    }
}

// Spawn call
uint64_t counter_id = ax_actor_spawn(_AX_main_Counter_handler, &init_state,
                                      sizeof(CounterState), supervisor_id);
```

`spawn MyActor(args)` expression type: `ActorRef[MyActor]`.

`send actor_ref, Msg(args)` statement: calls `ax_actor_send()`.

## Implementation Steps

1. Create `codegen/actor_codegen.go`.
2. Add actor opcodes to AIR instruction set (p09-t01 extension).
3. Lower `actor` declaration to handler function + state struct.
4. Lower `spawn` expression to `OpActorSpawn` → `ax_actor_spawn()` call.
5. Lower `send` statement to `OpActorSend` → `ax_actor_send()` call.
6. Lower `self` keyword to `OpActorSelf`.
7. Wire C backend to emit `ax_actor_spawn()` etc. for actor opcodes.
8. Write codegen integration tests.

## Test Plan
- `TestActorCodegenSpawn`: `spawn Counter()` → emits `ax_actor_spawn` call
- `TestActorCodegenSend`: `send counter, Increment` → emits `ax_actor_send`
- `TestActorCodegenHandler`: Counter handler handles Increment message correctly
- `TestActorCodegenE2E`: spawned counter actor counts correctly

## Validation Checklist
- [ ] Handler function signature matches `AxHandlerFn` typedef
- [ ] State struct generated with correct field types
- [ ] spawn expression yields ActorRef[T] type
- [ ] send to wrong actor type → compile-time type error

## Acceptance Criteria
- `spawn Counter() → send counter, Increment → send counter, GetCount(self)` pipeline works

## Definition of Done
- [ ] `codegen/actor_codegen.go` implemented
- [ ] End-to-end actor test passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Type mismatch on send (wrong message type) | Enforce ActorRef[T] send type at type-check time |
| Handler fn pointer type unsafe | Use explicit typedef; verify at codegen time |

## Future Follow-up Tasks
- Actor hot-swap (become with new handler) for live upgrades
- Distribution: remote actor refs across network
