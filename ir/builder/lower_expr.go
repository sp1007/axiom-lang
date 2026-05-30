package builder

import (
	"strconv"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
)

// lowerExpr translates an expression AST node into AIR instructions
// and returns the SSA register holding the result value.
// Returns 0 for void expressions (e.g., function calls with void return).
func (fl *funcLowering) lowerExpr(idx uint32, node *ast.AstNode) uint32 {
	switch node.Kind {
	case ast.NodeIntLit:
		return fl.lowerIntLit(node)
	case ast.NodeFloatLit:
		return fl.lowerFloatLit(node)
	case ast.NodeBoolLit:
		return fl.lowerBoolLit(node)
	case ast.NodeNilLit:
		return fl.lowerNilLit()
	case ast.NodeStringLit:
		return fl.lowerStringLit(node)
	case ast.NodeCharLit:
		return fl.lowerCharLit(node)
	case ast.NodeIdent:
		return fl.lowerIdent(node)
	case ast.NodeBinaryExpr:
		return fl.lowerBinaryExpr(idx, node)
	case ast.NodeUnaryExpr:
		return fl.lowerUnaryExpr(idx, node)
	case ast.NodeCallExpr:
		return fl.lowerCallExpr(idx, node)
	case ast.NodeFieldExpr:
		return fl.lowerFieldExpr(idx, node)
	case ast.NodeIndexExpr:
		return fl.lowerIndexExpr(idx, node)
	case ast.NodeCastExpr:
		return fl.lowerCastExpr(idx, node)
	case ast.NodeDerefExpr:
		return fl.lowerDerefExpr(idx, node)
	case ast.NodeSpawnExpr:
		return fl.lowerSpawnExpr(idx, node)
	case ast.NodeAwaitExpr:
		return fl.lowerAwaitExpr(idx, node)
	case ast.NodeStructLit:
		return fl.lowerStructLit(idx, node)
	case ast.NodeArrayLit:
		return fl.lowerArrayLit(idx, node)
	default:
		// Unknown expression → NOP, return 0
		return 0
	}
}

// lowerIntLit emits an integer constant.
func (fl *funcLowering) lowerIntLit(node *ast.AstNode) uint32 {
	reg := fl.fb.FreshReg()
	text := fl.mb.tree.TokenText(node.TokenIdx)
	val, _ := strconv.ParseInt(string(text), 0, 64)

	typeID := uint16(3) // default i32
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIConst,
		TypeID: typeID,
		Dest:   reg,
		Src1:   uint32(val),
	})
	return reg
}

// lowerFloatLit emits a float constant.
func (fl *funcLowering) lowerFloatLit(node *ast.AstNode) uint32 {
	reg := fl.fb.FreshReg()
	text := fl.mb.tree.TokenText(node.TokenIdx)
	val, _ := strconv.ParseFloat(string(text), 64)

	typeID := uint16(10) // default f64
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	// Store the float bits in Src1 (truncated to uint32 for f32, lossy for f64)
	// Full float encoding requires Extras; for MVP, store truncated bits
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpFConst,
		TypeID: typeID,
		Dest:   reg,
		Src1:   uint32(val), // truncated representation
	})
	return reg
}

// lowerBoolLit emits a boolean constant (true=1, false=0).
func (fl *funcLowering) lowerBoolLit(node *ast.AstNode) uint32 {
	reg := fl.fb.FreshReg()
	text := fl.mb.tree.TokenText(node.TokenIdx)
	val := uint32(0)
	if string(text) == "true" {
		val = 1
	}
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIConst,
		TypeID: uint16(11), // bool
		Dest:   reg,
		Src1:   val,
	})
	return reg
}

// lowerNilLit emits a nil constant.
func (fl *funcLowering) lowerNilLit() uint32 {
	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIConst,
		TypeID: 0,
		Dest:   reg,
		Src1:   0,
	})
	return reg
}

// lowerStringLit emits a string literal load.
func (fl *funcLowering) lowerStringLit(node *ast.AstNode) uint32 {
	reg := fl.fb.FreshReg()
	// String literals are stored as interned string IDs
	text := fl.mb.tree.TokenText(node.TokenIdx)
	strID := fl.mb.intern.Intern(text)

	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIConst,
		TypeID: uint16(12), // string
		Dest:   reg,
		Src1:   strID,
	})
	return reg
}

// lowerCharLit emits a char literal.
func (fl *funcLowering) lowerCharLit(node *ast.AstNode) uint32 {
	reg := fl.fb.FreshReg()
	text := fl.mb.tree.TokenText(node.TokenIdx)
	val := uint32(0)
	if len(text) > 0 {
		// Remove quotes if present
		s := string(text)
		if len(s) >= 3 && s[0] == '\'' && s[len(s)-1] == '\'' {
			runes := []rune(s[1 : len(s)-1])
			if len(runes) > 0 {
				val = uint32(runes[0])
			}
		}
	}
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIConst,
		TypeID: uint16(13), // char8
		Dest:   reg,
		Src1:   val,
	})
	return reg
}

// lowerIdent resolves an identifier to its SSA register.
func (fl *funcLowering) lowerIdent(node *ast.AstNode) uint32 {
	// The Ident node stores its interned name ID in Payload
	nameID := node.Payload
	if nameID == 0 {
		// Fallback: intern from token text
		text := fl.mb.tree.TokenText(node.TokenIdx)
		nameID = fl.mb.intern.Intern(text)
	}

	// Only look up in fl.locals if the symbol is a local variable, parameter, or constant
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.Kind != sema.SymVar && sym.Kind != sema.SymParam && sym.Kind != sema.SymConst {
			return 0
		}
	}

	if reg, ok := fl.locals[nameID]; ok {
		return reg
	}

	// If not in locals, it may be a global or function reference.
	// For MVP, return 0 (unresolved).
	return 0
}

// lowerBinaryExpr lowers a binary operation: lhs op rhs
func (fl *funcLowering) lowerBinaryExpr(idx uint32, node *ast.AstNode) uint32 {
	// First child = LHS, second child = RHS
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}
	lhsNode := fl.mb.tree.Node(child)
	lhsReg := fl.lowerExpr(child, lhsNode)

	rhsIdx := lhsNode.NextSibling
	if rhsIdx == ast.NullIdx {
		return lhsReg
	}
	rhsNode := fl.mb.tree.Node(rhsIdx)
	rhsReg := fl.lowerExpr(rhsIdx, rhsNode)

	// Map operator token to AIR opcode
	opToken := fl.mb.tree.TokenText(node.TokenIdx)
	opcode := mapBinaryOp(string(opToken))

	// Determine result type
	typeID := fl.nodeType(node)
	if typeID == 0 {
		isCmp := opcode == air.OpEq || opcode == air.OpNe || opcode == air.OpLt || opcode == air.OpLe || opcode == air.OpGt || opcode == air.OpGe
		if isCmp {
			typeID = uint16(types.TypeBool)
		} else {
			typeID = fl.getRegType(lhsReg)
		}
	}
	if typeID == 0 {
		typeID = 3 // default i32
	}

	// Adjust opcode to float if the result type is float
	if types.TypeID(typeID).IsFloat() {
		switch opcode {
		case air.OpIAdd:
			opcode = air.OpFAdd
		case air.OpISub:
			opcode = air.OpFSub
		case air.OpIMul:
			opcode = air.OpFMul
		case air.OpIDiv:
			opcode = air.OpFDiv
		}
	}

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: opcode,
		TypeID: typeID,
		Dest:   reg,
		Src1:   lhsReg,
		Src2:   rhsReg,
	})
	return reg
}

// mapBinaryOp maps an operator token to an AIR opcode.
func mapBinaryOp(op string) air.Opcode {
	switch op {
	case "+":
		return air.OpIAdd
	case "-":
		return air.OpISub
	case "*":
		return air.OpIMul
	case "/":
		return air.OpIDiv
	case "%":
		return air.OpIMod
	case "==":
		return air.OpEq
	case "!=":
		return air.OpNe
	case "<":
		return air.OpLt
	case "<=":
		return air.OpLe
	case ">":
		return air.OpGt
	case ">=":
		return air.OpGe
	case "and", "&&":
		return air.OpAnd
	case "or", "||":
		return air.OpOr
	case "^":
		return air.OpXor
	case "<<":
		return air.OpShl
	case ">>":
		return air.OpShr
	default:
		return air.OpNop // unknown operator
	}
}

// lowerUnaryExpr lowers a unary operation: op expr
func (fl *funcLowering) lowerUnaryExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}
	cn := fl.mb.tree.Node(child)
	operandReg := fl.lowerExpr(child, cn)

	opToken := fl.mb.tree.TokenText(node.TokenIdx)
	opcode := air.OpNop
	switch string(opToken) {
	case "-":
		opcode = air.OpNeg
	case "not", "!":
		opcode = air.OpNot
	case "~":
		opcode = air.OpNot
	}

	typeID := uint16(3) // default
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: opcode,
		TypeID: typeID,
		Dest:   reg,
		Src1:   operandReg,
	})
	return reg
}

// lowerCallExpr lowers a function call: fn(args...)
func (fl *funcLowering) lowerCallExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	// First child is the callee expression
	calleeNode := fl.mb.tree.Node(child)
	if calleeNode.Kind == ast.NodeIdent {
		symIdx := calleeNode.Payload
		if symIdx != 0 && int(symIdx) < len(fl.mb.symbols.Symbols) {
			sym := fl.mb.symbols.SymbolAt(symIdx)
			if sym.Kind == sema.SymStruct {
				return fl.lowerStructConstructorCall(idx, node, sym.TypeID)
			}
		}
	}
	calleeReg := fl.lowerExpr(child, calleeNode)

	// Remaining children are arguments
	var tempArgs []uint32
	arg := calleeNode.NextSibling
	for arg != ast.NullIdx {
		argNode := fl.mb.tree.Node(arg)
		argReg := fl.lowerExpr(arg, argNode)
		tempArgs = append(tempArgs, argReg)
		arg = argNode.NextSibling
	}

	argStart := fl.fb.EmitExtra(uint32(len(tempArgs)))
	for _, argReg := range tempArgs {
		fl.fb.EmitExtra(argReg)
	}

	// Determine return type
	typeID := uint16(0)
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpCall,
		TypeID: typeID,
		Dest:   reg,
		Src1:   calleeReg,
		Src2:   argStart,
	})
	return reg
}

// lowerFieldExpr lowers a field access: expr.field
func (fl *funcLowering) lowerFieldExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	objNode := fl.mb.tree.Node(child)
	objReg := fl.lowerExpr(child, objNode)

	// The field index is typically stored in ExtraIdx or Payload
	fieldIdx := node.ExtraIdx

	typeID := fl.nodeType(node)

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpGetField,
		TypeID: typeID,
		Dest:   reg,
		Src1:   objReg,
		Src2:   fieldIdx,
	})
	return reg
}

// lowerIndexExpr lowers an index operation: expr[idx]
func (fl *funcLowering) lowerIndexExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	arrNode := fl.mb.tree.Node(child)
	arrReg := fl.lowerExpr(child, arrNode)

	idxExpr := arrNode.NextSibling
	idxReg := uint32(0)
	if idxExpr != ast.NullIdx {
		idxNode := fl.mb.tree.Node(idxExpr)
		idxReg = fl.lowerExpr(idxExpr, idxNode)
	}

	typeID := fl.nodeType(node)

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpIndex,
		TypeID: typeID,
		Dest:   reg,
		Src1:   arrReg,
		Src2:   idxReg,
	})
	return reg
}

// lowerCastExpr lowers a type cast: expr as Type
func (fl *funcLowering) lowerCastExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	srcNode := fl.mb.tree.Node(child)
	srcReg := fl.lowerExpr(child, srcNode)

	typeID := fl.nodeType(node)

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpCast,
		TypeID: typeID,
		Dest:   reg,
		Src1:   srcReg,
	})
	return reg
}

// lowerDerefExpr lowers a pointer dereference: expr.*
func (fl *funcLowering) lowerDerefExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	ptrNode := fl.mb.tree.Node(child)
	ptrReg := fl.lowerExpr(child, ptrNode)

	typeID := fl.nodeType(node)

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpDeref,
		TypeID: typeID,
		Dest:   reg,
		Src1:   ptrReg,
	})
	return reg
}

// lowerSpawnExpr lowers a spawn expression. MVP: synchronous.
func (fl *funcLowering) lowerSpawnExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	targetNode := fl.mb.tree.Node(child)
	targetReg := fl.lowerExpr(child, targetNode)

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpSpawn,
		Dest:   reg,
		Src1:   targetReg,
	})
	return reg
}

// lowerAwaitExpr lowers an await expression. MVP: synchronous.
func (fl *funcLowering) lowerAwaitExpr(idx uint32, node *ast.AstNode) uint32 {
	child := node.FirstChild
	if child == ast.NullIdx {
		return 0
	}

	futureNode := fl.mb.tree.Node(child)
	futureReg := fl.lowerExpr(child, futureNode)

	typeID := uint16(0)
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	reg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpAwait,
		TypeID: typeID,
		Dest:   reg,
		Src1:   futureReg,
	})
	return reg
}

// lowerStructLit lowers a struct literal: TypeName{field: value, ...}
func (fl *funcLowering) lowerStructLit(idx uint32, node *ast.AstNode) uint32 {
	typeID := uint16(0)
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	// Allocate stack slot for the struct
	structReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpAlloc,
		TypeID: typeID,
		Dest:   structReg,
	})

	// Initialize fields
	child := node.FirstChild
	fieldIdx := uint32(0)
	for child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		if cn.Kind == ast.NodeNamedArg {
			// Named field: field: value
			valChild := cn.FirstChild
			if valChild != ast.NullIdx {
				valNode := fl.mb.tree.Node(valChild)
				valReg := fl.lowerExpr(valChild, valNode)
				fl.fb.Emit(air.AirInst{
					Opcode: air.OpSetField,
					Src1:   structReg,
					Src2:   fieldIdx,
					Dest:   valReg,
				})
			}
			fieldIdx++
		}
		child = cn.NextSibling
	}

	return structReg
}

// lowerArrayLit lowers an array literal: [expr, ...]
func (fl *funcLowering) lowerArrayLit(idx uint32, node *ast.AstNode) uint32 {
	typeID := uint16(0)
	if node.Payload != 0 && int(node.Payload) < len(fl.mb.symbols.Symbols) {
		sym := fl.mb.symbols.SymbolAt(node.Payload)
		if sym.TypeID != 0 {
			typeID = uint16(sym.TypeID)
		}
	}

	arrReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpAlloc,
		TypeID: typeID,
		Dest:   arrReg,
	})

	// Store elements
	child := node.FirstChild
	elemIdx := uint32(0)
	for child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		elemReg := fl.lowerExpr(child, cn)

		idxReg := fl.fb.FreshReg()
		fl.fb.Emit(air.AirInst{
			Opcode: air.OpIConst,
			TypeID: uint16(3),
			Dest:   idxReg,
			Src1:   elemIdx,
		})

		fl.fb.Emit(air.AirInst{
			Opcode: air.OpStore,
			Src1:   elemReg,
			Src2:   idxReg,
			Dest:   arrReg,
		})

		elemIdx++
		child = cn.NextSibling
	}

	return arrReg
}

func (fl *funcLowering) lowerStructConstructorCall(idx uint32, node *ast.AstNode, typeID uint32) uint32 {
	structReg := fl.fb.FreshReg()
	fl.fb.Emit(air.AirInst{
		Opcode: air.OpAlloc,
		TypeID: uint16(typeID),
		Dest:   structReg,
	})

	callee := node.FirstChild
	if callee == ast.NullIdx {
		return structReg
	}
	child := fl.mb.tree.Node(callee).NextSibling
	fieldIdx := uint32(0)
	for child != ast.NullIdx {
		cn := fl.mb.tree.Node(child)
		if cn.Kind == ast.NodeNamedArg {
			valChild := cn.FirstChild
			if valChild != ast.NullIdx {
				valNode := fl.mb.tree.Node(valChild)
				valReg := fl.lowerExpr(valChild, valNode)
				fl.fb.Emit(air.AirInst{
					Opcode: air.OpSetField,
					Src1:   structReg,
					Src2:   fieldIdx,
					Dest:   valReg,
				})
			}
			fieldIdx++
		} else {
			valReg := fl.lowerExpr(child, cn)
			fl.fb.Emit(air.AirInst{
				Opcode: air.OpSetField,
				Src1:   structReg,
				Src2:   fieldIdx,
				Dest:   valReg,
			})
			fieldIdx++
		}
		child = cn.NextSibling
	}

	return structReg
}

func (fl *funcLowering) nodeType(node *ast.AstNode) uint16 {
	if node.Payload == 0 {
		return 0
	}
	// If it's an identifier or parameter reference, look up symbol
	if node.Kind == ast.NodeIdent {
		if int(node.Payload) < len(fl.mb.symbols.Symbols) {
			sym := fl.mb.symbols.SymbolAt(node.Payload)
			return uint16(sym.TypeID)
		}
		return 0
	}
	// Otherwise, for expression nodes, Payload is directly the TypeID!
	return uint16(node.Payload)
}

func (fl *funcLowering) getRegType(reg uint32) uint16 {
	if reg == 0 {
		return 0
	}
	insts := fl.fb.Insts()
	for i := range insts {
		if insts[i].Dest == reg {
			return insts[i].TypeID
		}
	}
	return 0
}

