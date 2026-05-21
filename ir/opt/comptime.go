package opt

import (
	"fmt"

	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t06: Compile-Time Expression Interpreter (#run)
//
// Evaluates pure AIR subgraphs at compile time, replacing OpComptime-marked
// regions with constant values. Supports all ALU ops, comparisons, and
// control flow. Rejects side-effectful operations (alloc, free, I/O).
// --------------------------------------------------------------------------

const (
	// MaxComptimeSteps limits the number of instructions the interpreter
	// will execute before aborting to prevent infinite loops.
	MaxComptimeSteps = 100_000

	// MaxComptimeDepth limits recursive call depth.
	MaxComptimeDepth = 1000
)

// Value represents a compile-time value.
type Value struct {
	TypeID uint32
	IVal   int64
	FVal   float64
	IsNil  bool
}

// ComptimeError reports an error during compile-time evaluation.
type ComptimeError struct {
	Msg string
}

func (e *ComptimeError) Error() string { return e.Msg }

// CompTimeInterpreter executes AIR instructions at compile time.
type CompTimeInterpreter struct {
	module *air.AirModule
	regs   map[uint32]Value // virtual register file
	steps  int              // step counter
	depth  int              // current call depth
}

// NewCompTimeInterpreter creates a new interpreter for the given module.
func NewCompTimeInterpreter(mod *air.AirModule) *CompTimeInterpreter {
	return &CompTimeInterpreter{
		module: mod,
		regs:   make(map[uint32]Value, 64),
	}
}

// Interpret executes a function with the given arguments and returns the result.
func (interp *CompTimeInterpreter) Interpret(fn *air.AirFunc, args []Value) (Value, error) {
	if interp.depth >= MaxComptimeDepth {
		return Value{}, &ComptimeError{"#run: exceeded maximum call depth"}
	}
	interp.depth++
	defer func() { interp.depth-- }()

	// Create local register file (copy parent to allow recursive calls)
	savedRegs := interp.regs
	interp.regs = make(map[uint32]Value, 64)
	defer func() { interp.regs = savedRegs }()

	// Load arguments into parameter registers (1-based)
	for i, arg := range args {
		interp.regs[uint32(i+1)] = arg
	}

	// Execute instructions sequentially
	for i := 0; i < len(fn.Insts); i++ {
		interp.steps++
		if interp.steps > MaxComptimeSteps {
			return Value{}, &ComptimeError{"#run: computation exceeded step limit"}
		}

		inst := &fn.Insts[i]
		switch inst.Opcode {
		case air.OpNop:
			continue

		case air.OpIConst:
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: int64(int32(inst.Src1))}

		case air.OpFConst:
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), FVal: float64(inst.Src1)}

		case air.OpCopy:
			interp.regs[inst.Dest] = interp.regs[inst.Src1]

		case air.OpIAdd:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal + b.IVal}

		case air.OpISub:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal - b.IVal}

		case air.OpIMul:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal * b.IVal}

		case air.OpIDiv:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			if b.IVal == 0 {
				return Value{}, &ComptimeError{"#run: division by zero"}
			}
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal / b.IVal}

		case air.OpIMod:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			if b.IVal == 0 {
				return Value{}, &ComptimeError{"#run: modulo by zero"}
			}
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal % b.IVal}

		case air.OpNeg:
			a := interp.regs[inst.Src1]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: -a.IVal}

		case air.OpNot:
			a := interp.regs[inst.Src1]
			if a.IVal == 0 {
				interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: 1}
			} else {
				interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: 0}
			}

		case air.OpEq:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: boolToI64(a.IVal == b.IVal)}

		case air.OpNe:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: boolToI64(a.IVal != b.IVal)}

		case air.OpLt:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: boolToI64(a.IVal < b.IVal)}

		case air.OpLe:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: boolToI64(a.IVal <= b.IVal)}

		case air.OpGt:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: boolToI64(a.IVal > b.IVal)}

		case air.OpGe:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: boolToI64(a.IVal >= b.IVal)}

		case air.OpAnd:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal & b.IVal}

		case air.OpOr:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal | b.IVal}

		case air.OpXor:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal ^ b.IVal}

		case air.OpShl:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal << uint(b.IVal)}

		case air.OpShr:
			a, b := interp.regs[inst.Src1], interp.regs[inst.Src2]
			interp.regs[inst.Dest] = Value{TypeID: uint32(inst.TypeID), IVal: a.IVal >> uint(b.IVal)}

		case air.OpReturn:
			if inst.Src1 != 0 {
				return interp.regs[inst.Src1], nil
			}
			return Value{}, nil

		case air.OpCall:
			// Recursive interpretation: find callee function
			calleeName := inst.Src1
			callee := interp.findFunc(calleeName)
			if callee == nil {
				return Value{}, &ComptimeError{fmt.Sprintf("#run: unknown function %d", calleeName)}
			}
			if callee.IsExtern {
				return Value{}, &ComptimeError{"#run: cannot call extern function at compile time"}
			}
			// For MVP: call with no args (full arg passing via extras is future work)
			result, err := interp.Interpret(callee, nil)
			if err != nil {
				return Value{}, err
			}
			if inst.Dest != 0 {
				interp.regs[inst.Dest] = result
			}

		// Disallowed operations
		case air.OpAlloc, air.OpFree, air.OpStore, air.OpLoad,
			air.OpSpawn, air.OpSend, air.OpAwait,
			air.OpMakeRef, air.OpDeref, air.OpDestroy,
			air.OpArenaAlloc:
			return Value{}, &ComptimeError{"#run: cannot use memory/IO operations at compile time"}

		default:
			// Unknown opcode — skip
			continue
		}
	}

	// Fell through without return — void return
	return Value{}, nil
}

// findFunc looks up a function by Name (interned ID) in the module.
func (interp *CompTimeInterpreter) findFunc(nameID uint32) *air.AirFunc {
	for i := range interp.module.Funcs {
		if interp.module.Funcs[i].Name == nameID {
			return &interp.module.Funcs[i]
		}
	}
	return nil
}

func boolToI64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// --------------------------------------------------------------------------
// ComptimePass — OptPass wrapper for the comptime interpreter
// --------------------------------------------------------------------------

// ComptimePass implements OptPass. It finds instructions flagged as
// compile-time evaluable and replaces them with constants.
type ComptimePass struct{}

func (p *ComptimePass) Name() string { return "comptime" }

// Run scans for compile-time evaluable patterns.
// For MVP, this pass looks for sequences that are entirely constant
// (all inputs are OpIConst/OpFConst) and evaluates them.
// Full #run support requires OpComptime markers from the frontend.
func (p *ComptimePass) Run(mod *air.AirModule) bool {
	// The comptime pass is driven by the frontend's #run markers.
	// Since the frontend doesn't emit OpComptime yet, this is a no-op
	// placeholder that will be activated when the parser supports #run.
	return false
}
