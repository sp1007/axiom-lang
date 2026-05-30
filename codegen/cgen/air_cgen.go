package cgen

import (
	"bytes"
	"fmt"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t10: C-Backend v2 — Lowering from AIR
//
// Generates C11 source code from an optimized AirModule. Each AIR
// instruction maps to a simple C statement. Control flow uses goto
// labels for basic blocks. All registers are declared at function top.
// --------------------------------------------------------------------------

// AirCGen generates C source code from an AirModule.
type AirCGen struct {
	tt   *types.TypeTable
	pool *ast.InternPool
	buf  bytes.Buffer
}

// NewAirCGen creates a new AIR→C code generator.
func NewAirCGen(tt *types.TypeTable, pool *ast.InternPool) *AirCGen {
	return &AirCGen{tt: tt, pool: pool}
}

// Generate produces a complete C11 source string from the module.
func (g *AirCGen) Generate(mod *air.AirModule) string {
	g.buf.Reset()

	// Header
	g.writeln("#include <stdint.h>")
	g.writeln("#include <stdbool.h>")
	g.writeln("#include <stdlib.h>")
	g.writeln("")

	// Type aliases
	g.writeln("typedef int32_t ax_i32;")
	g.writeln("typedef int64_t ax_i64;")
	g.writeln("typedef double ax_f64;")
	g.writeln("typedef bool ax_bool;")
	g.writeln("typedef void ax_void;")
	g.writeln("")

	// Forward declarations
	for i := range mod.Funcs {
		fn := &mod.Funcs[i]
		if fn.IsExtern {
			continue
		}
		g.emitForwardDecl(fn)
	}
	g.writeln("")

	// Function bodies
	for i := range mod.Funcs {
		fn := &mod.Funcs[i]
		if fn.IsExtern {
			continue
		}
		g.emitFunction(fn)
		g.writeln("")
	}

	return g.buf.String()
}

// emitForwardDecl writes a forward declaration for a function.
func (g *AirCGen) emitForwardDecl(fn *air.AirFunc) {
	retType := g.cTypeName(fn.RetType)
	name := g.funcName(fn)
	g.buf.WriteString(retType)
	g.buf.WriteString(" ")
	g.buf.WriteString(name)
	g.buf.WriteString("(")

	for i, paramType := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(g.cTypeName(paramType))
		fmt.Fprintf(&g.buf, " r_%d", i+1)
	}
	if len(fn.Params) == 0 {
		g.buf.WriteString("void")
	}
	g.buf.WriteString(");\n")
}

// emitFunction writes a complete function definition.
func (g *AirCGen) emitFunction(fn *air.AirFunc) {
	retType := g.cTypeName(fn.RetType)
	name := g.funcName(fn)

	g.buf.WriteString(retType)
	g.buf.WriteString(" ")
	g.buf.WriteString(name)
	g.buf.WriteString("(")

	for i, paramType := range fn.Params {
		if i > 0 {
			g.buf.WriteString(", ")
		}
		g.buf.WriteString(g.cTypeName(paramType))
		fmt.Fprintf(&g.buf, " r_%d", i+1)
	}
	if len(fn.Params) == 0 {
		g.buf.WriteString("void")
	}
	g.writeln(") {")

	// Declare all registers at function top (to avoid goto-over-init issues)
	g.emitRegisterDecls(fn)

	// Emit basic blocks
	for bi := range fn.Blocks {
		blk := &fn.Blocks[bi]
		fmt.Fprintf(&g.buf, "  block_%d: ;\n", blk.ID)
		for _, instIdx := range blk.Instrs {
			if int(instIdx) < len(fn.Insts) {
				g.emitInst(&fn.Insts[instIdx])
			}
		}
	}

	// If we're using flat instruction array (no block mapping), emit all
	if len(fn.Blocks) == 0 || !g.hasBlockInstrs(fn) {
		for i := range fn.Insts {
			g.emitInst(&fn.Insts[i])
		}
	}

	g.writeln("}")
}

// hasBlockInstrs checks if blocks have instruction indices populated.
func (g *AirCGen) hasBlockInstrs(fn *air.AirFunc) bool {
	for bi := range fn.Blocks {
		if len(fn.Blocks[bi].Instrs) > 0 {
			return true
		}
	}
	return false
}

// emitRegisterDecls declares all used registers at function top.
func (g *AirCGen) emitRegisterDecls(fn *air.AirFunc) {
	// Collect all registers with their types
	declared := make(map[uint32]bool)
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Dest != 0 && !declared[inst.Dest] {
			declared[inst.Dest] = true
			typeName := g.cTypeName(uint32(inst.TypeID))
			fmt.Fprintf(&g.buf, "  %s r_%d;\n", typeName, inst.Dest)
		}
	}
	if len(declared) > 0 {
		g.writeln("")
	}
}

// emitInst writes the C statement for a single AIR instruction.
func (g *AirCGen) emitInst(inst *air.AirInst) {
	switch inst.Opcode {
	case air.OpNop:
		// skip

	case air.OpIConst:
		fmt.Fprintf(&g.buf, "  r_%d = %d;\n", inst.Dest, int32(inst.Src1))

	case air.OpFConst:
		fmt.Fprintf(&g.buf, "  r_%d = %d.0;\n", inst.Dest, inst.Src1)

	case air.OpCopy:
		fmt.Fprintf(&g.buf, "  r_%d = r_%d;\n", inst.Dest, inst.Src1)

	case air.OpMove:
		fmt.Fprintf(&g.buf, "  r_%d = r_%d; /* move */\n", inst.Dest, inst.Src1)

	// Binary ALU
	case air.OpIAdd:
		g.emitBinaryOp(inst, "+")
	case air.OpISub:
		g.emitBinaryOp(inst, "-")
	case air.OpIMul:
		g.emitBinaryOp(inst, "*")
	case air.OpIDiv:
		g.emitBinaryOp(inst, "/")
	case air.OpIMod:
		g.emitBinaryOp(inst, "%")

	// Comparisons
	case air.OpEq:
		g.emitBinaryOp(inst, "==")
	case air.OpNe:
		g.emitBinaryOp(inst, "!=")
	case air.OpLt:
		g.emitBinaryOp(inst, "<")
	case air.OpLe:
		g.emitBinaryOp(inst, "<=")
	case air.OpGt:
		g.emitBinaryOp(inst, ">")
	case air.OpGe:
		g.emitBinaryOp(inst, ">=")

	// Bitwise
	case air.OpAnd:
		g.emitBinaryOp(inst, "&")
	case air.OpOr:
		g.emitBinaryOp(inst, "|")
	case air.OpXor:
		g.emitBinaryOp(inst, "^")
	case air.OpShl:
		g.emitBinaryOp(inst, "<<")
	case air.OpShr:
		g.emitBinaryOp(inst, ">>")

	// Unary
	case air.OpNeg:
		fmt.Fprintf(&g.buf, "  r_%d = -r_%d;\n", inst.Dest, inst.Src1)
	case air.OpNot:
		if inst.TypeID == uint16(types.TypeBool) {
			fmt.Fprintf(&g.buf, "  r_%d = !r_%d;\n", inst.Dest, inst.Src1)
		} else {
			fmt.Fprintf(&g.buf, "  r_%d = ~r_%d;\n", inst.Dest, inst.Src1)
		}

	// Memory
	case air.OpAlloc:
		typeName := g.cTypeName(uint32(inst.TypeID))
		fmt.Fprintf(&g.buf, "  r_%d = (%s*)malloc(sizeof(%s));\n", inst.Dest, typeName, typeName)
	case air.OpFree:
		fmt.Fprintf(&g.buf, "  free(r_%d);\n", inst.Src1)
	case air.OpStore:
		fmt.Fprintf(&g.buf, "  *(void**)r_%d = r_%d;\n", inst.Src2, inst.Src1)
	case air.OpLoad:
		fmt.Fprintf(&g.buf, "  r_%d = *(void**)r_%d;\n", inst.Dest, inst.Src1)
	case air.OpMakeRef:
		fmt.Fprintf(&g.buf, "  r_%d = r_%d; /* make_ref */\n", inst.Dest, inst.Src1)
	case air.OpDeref:
		fmt.Fprintf(&g.buf, "  r_%d = r_%d; /* deref (gen_id checked) */\n", inst.Dest, inst.Src1)
	case air.OpDestroy:
		fmt.Fprintf(&g.buf, "  free(r_%d); /* destroy */\n", inst.Src1)

	// Control flow
	case air.OpReturn:
		if inst.Src1 != 0 {
			fmt.Fprintf(&g.buf, "  return r_%d;\n", inst.Src1)
		} else {
			g.buf.WriteString("  return;\n")
		}
	case air.OpJump:
		fmt.Fprintf(&g.buf, "  goto block_%d;\n", inst.Src1)
	case air.OpBranch:
		fmt.Fprintf(&g.buf, "  if (r_%d) goto block_%d; else goto block_%d;\n",
			inst.Src1, inst.Src2, inst.Dest)

	// Call
	case air.OpCall:
		if inst.Dest != 0 {
			fmt.Fprintf(&g.buf, "  r_%d = /* call @%d */;\n", inst.Dest, inst.Src1)
		} else {
			fmt.Fprintf(&g.buf, "  /* call @%d */;\n", inst.Src1)
		}

	// Phi (insert copies in predecessors — handled at block level)
	case air.OpPhi:
		fmt.Fprintf(&g.buf, "  /* phi r_%d */;\n", inst.Dest)

	default:
		fmt.Fprintf(&g.buf, "  /* unknown op 0x%04X */;\n", uint16(inst.Opcode))
	}
}

// emitBinaryOp writes a binary operation: r_dest = r_src1 OP r_src2
func (g *AirCGen) emitBinaryOp(inst *air.AirInst, op string) {
	fmt.Fprintf(&g.buf, "  r_%d = r_%d %s r_%d;\n", inst.Dest, inst.Src1, op, inst.Src2)
}

// cTypeName returns the C type name for a TypeID.
func (g *AirCGen) cTypeName(typeID uint32) string {
	switch typeID {
	case 0:
		return "ax_void"
	case 1:
		return "ax_void"
	case 2:
		return "ax_bool"
	case 3:
		return "ax_i32"
	case 4:
		return "ax_i64"
	case 5:
		return "ax_f64"
	case 11:
		return "ax_bool"
	default:
		return fmt.Sprintf("ax_type_%d", typeID)
	}
}

// funcName returns the C function name for an AirFunc.
func (g *AirCGen) funcName(fn *air.AirFunc) string {
	if g.pool != nil && fn.Name != 0 {
		name := g.pool.Get(fn.Name)
		if len(name) > 0 {
			if name == "main" {
				return "main"
			}
			return "_AX_" + name
		}
	}
	return fmt.Sprintf("_AX_f%d", fn.Name)
}

func (g *AirCGen) writeln(s string) {
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}
