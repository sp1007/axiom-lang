package codegen

// --------------------------------------------------------------------------
// p15-t11: Distributed Actor Transport (Stub)
//
// Location-transparent messaging via a Transport interface.
// MVP is local-only in-process relay. Interface is designed for
// future TCP/QUIC transport backends.
// --------------------------------------------------------------------------

// ActorAddr represents a location-transparent actor address.
type ActorAddr struct {
	NodeID  uint32 // cluster node (0 = local)
	ActorID uint64 // actor ID within the node
}

// IsLocal returns true if the actor is on this node.
func (a ActorAddr) IsLocal() bool {
	return a.NodeID == 0
}

// Transport defines the interface for actor message delivery.
type Transport interface {
	// Send a message to an actor address.
	Send(target ActorAddr, msgType uint32, payload []byte) error

	// Register a local actor with a name for discovery.
	Register(name string, id uint64) error

	// Unregister a named actor.
	Unregister(name string) error

	// Resolve a name to an actor address.
	Resolve(name string) (ActorAddr, error)
}

// LocalTransport is the in-process transport (MVP).
type LocalTransport struct {
	actors map[uint64]bool     // registered actor IDs
	names  map[string]uint64   // name → actor ID
}

// NewLocalTransport creates a new local-only transport.
func NewLocalTransport() *LocalTransport {
	return &LocalTransport{
		actors: make(map[uint64]bool),
		names:  make(map[string]uint64),
	}
}

func (lt *LocalTransport) Send(target ActorAddr, msgType uint32, payload []byte) error {
	if !target.IsLocal() {
		return ErrRemoteNotSupported
	}
	if !lt.actors[target.ActorID] {
		return ErrActorNotFound
	}
	// In a real implementation, this would enqueue to the actor's mailbox
	// via the runtime API. For the stub, we just verify the actor exists.
	_ = msgType
	_ = payload
	return nil
}

func (lt *LocalTransport) Register(name string, id uint64) error {
	lt.actors[id] = true
	lt.names[name] = id
	return nil
}

func (lt *LocalTransport) Unregister(name string) error {
	if id, ok := lt.names[name]; ok {
		delete(lt.actors, id)
		delete(lt.names, name)
	}
	return nil
}

func (lt *LocalTransport) Resolve(name string) (ActorAddr, error) {
	id, ok := lt.names[name]
	if !ok {
		return ActorAddr{}, ErrActorNotFound
	}
	return ActorAddr{NodeID: 0, ActorID: id}, nil
}

// Sentinel errors
type transportError string

func (e transportError) Error() string { return string(e) }

const (
	ErrRemoteNotSupported = transportError("remote transport not supported")
	ErrActorNotFound      = transportError("actor not found")
)
