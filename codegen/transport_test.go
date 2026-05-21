package codegen_test

import (
	"testing"

	"github.com/axiom-lang/axiom/codegen"
)

func TestLocalTransport_RegisterResolve(t *testing.T) {
	lt := codegen.NewLocalTransport()
	if err := lt.Register("counter", 42); err != nil {
		t.Fatal(err)
	}

	addr, err := lt.Resolve("counter")
	if err != nil {
		t.Fatal(err)
	}
	if addr.ActorID != 42 {
		t.Errorf("expected 42, got %d", addr.ActorID)
	}
	if !addr.IsLocal() {
		t.Error("expected local")
	}
}

func TestLocalTransport_SendLocal(t *testing.T) {
	lt := codegen.NewLocalTransport()
	lt.Register("worker", 1)

	target := codegen.ActorAddr{NodeID: 0, ActorID: 1}
	if err := lt.Send(target, 0, nil); err != nil {
		t.Errorf("send failed: %v", err)
	}
}

func TestLocalTransport_SendRemoteFails(t *testing.T) {
	lt := codegen.NewLocalTransport()
	target := codegen.ActorAddr{NodeID: 1, ActorID: 1}
	if err := lt.Send(target, 0, nil); err == nil {
		t.Error("expected error for remote send")
	}
}

func TestLocalTransport_SendUnknownFails(t *testing.T) {
	lt := codegen.NewLocalTransport()
	target := codegen.ActorAddr{NodeID: 0, ActorID: 999}
	if err := lt.Send(target, 0, nil); err == nil {
		t.Error("expected error for unknown actor")
	}
}

func TestLocalTransport_Unregister(t *testing.T) {
	lt := codegen.NewLocalTransport()
	lt.Register("temp", 5)
	lt.Unregister("temp")

	_, err := lt.Resolve("temp")
	if err == nil {
		t.Error("expected error after unregister")
	}
}

func TestLocalTransport_ResolveNotFound(t *testing.T) {
	lt := codegen.NewLocalTransport()
	_, err := lt.Resolve("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}
