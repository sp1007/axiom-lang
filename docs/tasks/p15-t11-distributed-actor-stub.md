# p15-t11: Distributed Actor Transport (Stub)

## Purpose
Implement a stub distributed actor transport layer that provides location-transparent messaging between actors. In the MVP, all actors run locally and the "distributed" transport is just an in-process message relay. This establishes the API surface that a real TCP/QUIC transport will implement later.

## Context
Plan §1.2 mentions: _"Distributed Actor Transport (TCP/QUIC, location transparency)"_. Plan §12.7 confirms this should be stubbed: _"Distributed runtime: local-only messaging"_.

The stub ensures the actor API is location-transparent from the start: `actor.send(msg)` works identically whether the target is local or remote. The transport interface is designed for future TCP/QUIC implementation.

## Inputs
- Actor struct from p15-t01
- Message queue from p15-t05
- Channel infrastructure from p15-t05

## Outputs
- `runtime/actors/transport.go` — `Transport` interface + local stub
- `runtime/actors/registry.go` — actor address registry
- Tests

## Dependencies
- p15-t05: actor-message-queue — messaging infrastructure
- p15-t03: actor-system-init — actor system lifecycle

## Subsystems Affected
- Actor runtime: all message sends route through transport
- Future: TCP/QUIC transport replaces the stub

## Detailed Requirements

### Transport Interface

```go
type Transport interface {
    Send(target ActorAddr, msg Message) error
    Register(actor *Actor) ActorAddr
    Unregister(addr ActorAddr)
    Resolve(name string) (ActorAddr, error)
}

type ActorAddr struct {
    NodeID  uint64  // 0 = local node
    ActorID uint64
}

// LocalTransport is the stub implementation (in-process only).
type LocalTransport struct {
    actors map[uint64]*Actor
    names  map[string]uint64
    mu     sync.RWMutex
}
```

### Local Stub Behavior

- `Send()`: look up actor by ID in local map, push to mailbox
- `Register()`: assign unique ID, store in map
- `Resolve()`: look up by name
- NodeID always 0 (local)

### Future Extension Points

The interface is designed so a `TCPTransport` or `QUICTransport` can implement the same interface:
- `Send()` serializes the message and sends over the network
- `Resolve()` queries a distributed registry
- `ActorAddr` with `NodeID != 0` indicates a remote actor

## Implementation Steps

1. Create `runtime/actors/transport.go` with `Transport` interface.
2. Implement `LocalTransport` as the stub.
3. Create `runtime/actors/registry.go` for name → actor mapping.
4. Wire into actor system init (p15-t03).
5. Write tests for local send/receive.

## Test Plan

- `TestLocalSend`: send message to local actor → received in mailbox
- `TestLocalResolve`: register actor with name → resolve returns correct addr
- `TestLocalUnregister`: unregistered actor → send returns error

## Acceptance Criteria

- All actor messaging works through the Transport interface
- Switching from LocalTransport to a future TCPTransport requires no actor code changes

## Definition of Done

- [ ] `runtime/actors/transport.go` implemented
- [ ] `runtime/actors/registry.go` implemented
- [ ] Tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Interface too narrow for real distributed systems | Design based on Erlang's distribution protocol; extend as needed |

## Future Follow-up Tasks

- Future RFC: TCP/QUIC transport implementation
- Future: distributed actor discovery and clustering
