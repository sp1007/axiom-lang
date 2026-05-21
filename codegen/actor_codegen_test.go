package codegen_test

import (
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen"
)

func TestActorCodegen_Spawn(t *testing.T) {
	ac := codegen.NewActorCodegen()
	code := ac.GenerateSpawn("Counter", nil)
	if !strings.Contains(code, "ax_actor_spawn") {
		t.Errorf("expected ax_actor_spawn, got %q", code)
	}
	if !strings.Contains(code, "_AX_Counter_handler") {
		t.Errorf("expected handler name, got %q", code)
	}
}

func TestActorCodegen_SpawnWithArgs(t *testing.T) {
	ac := codegen.NewActorCodegen()
	code := ac.GenerateSpawn("Worker", []string{"42"})
	if !strings.Contains(code, "_init_Worker") {
		t.Errorf("expected init struct, got %q", code)
	}
}

func TestActorCodegen_Send(t *testing.T) {
	ac := codegen.NewActorCodegen()
	code := ac.GenerateSend("target_id", "self->id", 1, "msg", "sizeof(msg)")
	if !strings.Contains(code, "ax_actor_send") {
		t.Errorf("expected ax_actor_send, got %q", code)
	}
}

func TestActorCodegen_Handler(t *testing.T) {
	ac := codegen.NewActorCodegen()
	handlers := []codegen.ActorMsgHandler{
		{MsgName: "Increment", MsgTypeID: 1, PayloadType: "int32_t", Body: "state->count++;"},
		{MsgName: "GetCount", MsgTypeID: 2, PayloadType: "void", Body: "/* reply */"},
	}
	code := ac.GenerateHandlerFunction("Counter", handlers)
	if !strings.Contains(code, "Counter_State") {
		t.Error("expected state type")
	}
	if !strings.Contains(code, "case 1:") {
		t.Error("expected case 1")
	}
	if !strings.Contains(code, "case 2:") {
		t.Error("expected case 2")
	}
}

func TestActorCodegen_StateStruct(t *testing.T) {
	ac := codegen.NewActorCodegen()
	fields := []codegen.ActorField{
		{Name: "count", CType: "int64_t"},
		{Name: "name", CType: "const char*"},
	}
	code := ac.GenerateStateStruct("Counter", fields)
	if !strings.Contains(code, "Counter_State") {
		t.Error("expected Counter_State")
	}
	if !strings.Contains(code, "int64_t count") {
		t.Error("expected count field")
	}
}
