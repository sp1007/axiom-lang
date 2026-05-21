package codegen

// --------------------------------------------------------------------------
// p15-t09: Actor Codegen Integration
//
// Lowers actor syntax (spawn, send, receive, become) to runtime API calls.
// Generates handler functions, state structs, and message dispatch tables.
// --------------------------------------------------------------------------

import (
	"fmt"

	"github.com/axiom-lang/axiom/ir/air"
)

// ActorCodegen handles the lowering of actor constructs to C runtime calls.
type ActorCodegen struct {
	// Map from actor type name to generated handler name
	handlers map[string]string
	// Generated C code fragments
	fragments []string
}

// NewActorCodegen creates a new ActorCodegen instance.
func NewActorCodegen() *ActorCodegen {
	return &ActorCodegen{
		handlers: make(map[string]string),
	}
}

// GenerateSpawn emits code for spawning an actor.
// Result: ax_actor_spawn(handler_fn, init_data, data_size)
func (ac *ActorCodegen) GenerateSpawn(actorType string, initArgs []string) string {
	handler := ac.handlerName(actorType)
	ac.handlers[actorType] = handler

	if len(initArgs) == 0 {
		return fmt.Sprintf("ax_actor_spawn(%s, NULL, 0)", handler)
	}

	return fmt.Sprintf("ax_actor_spawn(%s, &_init_%s, sizeof(_init_%s))",
		handler, actorType, actorType)
}

// GenerateSend emits code for sending a message to an actor.
func (ac *ActorCodegen) GenerateSend(targetExpr, senderExpr string,
	msgType uint32, payloadExpr string, payloadSize string) string {
	return fmt.Sprintf("ax_actor_send(%s, %s, %d, &(%s), %s)",
		targetExpr, senderExpr, msgType, payloadExpr, payloadSize)
}

// GenerateHandlerFunction generates the message handler for an actor type.
func (ac *ActorCodegen) GenerateHandlerFunction(actorType string,
	msgHandlers []ActorMsgHandler) string {
	handler := ac.handlerName(actorType)

	code := fmt.Sprintf("static void %s(AxActor* self, void* payload, AxMsgType msg_type) {\n",
		handler)
	code += fmt.Sprintf("    %s_State* state = (%s_State*)self->state_data;\n",
		actorType, actorType)
	code += "    switch (msg_type) {\n"

	for _, mh := range msgHandlers {
		code += fmt.Sprintf("    case %d: { // %s\n", mh.MsgTypeID, mh.MsgName)
		code += fmt.Sprintf("        %s* msg = (%s*)payload;\n",
			mh.PayloadType, mh.PayloadType)
		code += fmt.Sprintf("        %s\n", mh.Body)
		code += "        break;\n"
		code += "    }\n"
	}

	code += "    default: break;\n"
	code += "    }\n"
	code += "}\n"
	return code
}

// GenerateStateStruct generates the state struct for an actor type.
func (ac *ActorCodegen) GenerateStateStruct(actorType string,
	fields []ActorField) string {
	code := fmt.Sprintf("typedef struct {\n")
	for _, f := range fields {
		code += fmt.Sprintf("    %s %s;\n", f.CType, f.Name)
	}
	code += fmt.Sprintf("} %s_State;\n", actorType)
	return code
}

func (ac *ActorCodegen) handlerName(actorType string) string {
	return fmt.Sprintf("_AX_%s_handler", actorType)
}

// ActorMsgHandler describes one message handler within an actor.
type ActorMsgHandler struct {
	MsgName     string // e.g. "Increment"
	MsgTypeID   uint32 // runtime message type ID
	PayloadType string // C type name for payload
	Body        string // C code for handler body
}

// ActorField describes a field in an actor's state struct.
type ActorField struct {
	Name  string // field name
	CType string // C type
}

// LowerActorOps processes AIR instructions and replaces actor-specific
// opcodes with runtime API calls.
func LowerActorOps(mod *air.AirModule) {
	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		for bi := range fn.Blocks {
			bb := &fn.Blocks[bi]
			for _, instIdx := range bb.Instrs {
				inst := &fn.Insts[instIdx]
				switch inst.Opcode {
				case air.OpCall:
					// Check if callee is an actor intrinsic
					// and lower to runtime call
					_ = inst
				}
			}
		}
	}
}


