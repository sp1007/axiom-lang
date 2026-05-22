package wasm

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

// WasmBackend translates AXIOM Intermediate Representation (AIR) into WebAssembly Text format (WAT).
type WasmBackend struct {
	OptLevel opt.OptLevel
	Pool     *ast.InternPool
	Table    *types.TypeTable
}

// NewWasmBackend creates a new WebAssembly backend.
func NewWasmBackend() *WasmBackend {
	return &WasmBackend{
		OptLevel: opt.O2,
	}
}

// Compile compiles the given AirModule to WAT format string.
func (b *WasmBackend) Compile(mod *air.AirModule) (string, error) {
	// Step 1: Run optimization pipeline first
	pipeline := opt.DefaultPipeline(b.OptLevel, false)
	pipeline.Run(mod)

	var sb strings.Builder

	// Write WAT Module header
	sb.WriteString("(module\n")

	// Declare linear memory
	sb.WriteString("  (memory (export \"memory\") 1)\n\n")

	// Collect and declare all external functions as Wasm imports
	for _, fn := range mod.Funcs {
		if fn.IsExtern {
			name := b.resolveSymName(fn.Name)
			// Map parameters
			paramsStr := ""
			for i, p := range fn.Params {
				paramsStr += fmt.Sprintf(" (param %s)", mapWasmType(uint16(p)))
				_ = i
			}
			// Map return
			retStr := ""
			if fn.RetType != 0 && types.TypeID(fn.RetType) != types.TypeVoid {
				retStr = fmt.Sprintf(" (result %s)", mapWasmType(uint16(fn.RetType)))
			}
			fmt.Fprintf(&sb, "  (import \"env\" \"%s\" (func $%s%s%s))\n", name, name, paramsStr, retStr)
		}
	}

	// Always import/define malloc and free if not already declared by externs
	hasMalloc := false
	hasFree := false
	for _, fn := range mod.Funcs {
		if fn.IsExtern {
			name := b.resolveSymName(fn.Name)
			if name == "malloc" {
				hasMalloc = true
			}
			if name == "free" {
				hasFree = true
			}
		}
	}
	if !hasMalloc {
		sb.WriteString("  (import \"env\" \"malloc\" (func $malloc (param i32) (result i32)))\n")
	}
	if !hasFree {
		sb.WriteString("  (import \"env\" \"free\" (func $free (param i32)))\n")
	}

	sb.WriteString("\n")

	// Step 2: Compile non-external functions
	for _, fn := range mod.Funcs {
		if fn.IsExtern {
			continue
		}

		funcWat, err := b.compileFunc(&fn)
		if err != nil {
			return "", err
		}
		sb.WriteString(funcWat)
		sb.WriteString("\n")
	}

	// Close WAT module
	sb.WriteString(")\n")

	return sb.String(), nil
}

func (b *WasmBackend) resolveSymName(symID uint32) string {
	if symID == 0 {
		return "main"
	}
	if symID == 4294967295 {
		return "malloc"
	}
	if symID == 4294967294 {
		return "free"
	}
	if b.Pool != nil && int(symID) <= b.Pool.Len() {
		name := b.Pool.Get(symID)
		if len(name) > 0 {
			if name == "main" || name == "printf" || name == "malloc" || name == "free" {
				return name
			}
			if strings.HasPrefix(name, "_AX_") {
				return name
			}
			return "_AX_" + name
		}
	}
	return fmt.Sprintf("_AX_f%d", symID)
}

func mapWasmType(typeID uint16) string {
	switch types.TypeID(typeID) {
	case types.TypeI64, types.TypeU64:
		return "i64"
	case types.TypeF32:
		return "f32"
	case types.TypeF64:
		return "f64"
	case types.TypeVoid:
		return ""
	default:
		// Integers i8..i32, bool, usize, isize, pointers map to i32 in wasm32
		return "i32"
	}
}

func (b *WasmBackend) compileFunc(fn *air.AirFunc) (string, error) {
	var sb strings.Builder
	name := b.resolveSymName(fn.Name)

	// Function signature
	fmt.Fprintf(&sb, "  (func $%s (export \"%s\")", name, name)
	for i, p := range fn.Params {
		fmt.Fprintf(&sb, " (param $p%d %s)", i+1, mapWasmType(uint16(p)))
	}
	if fn.RetType != 0 && types.TypeID(fn.RetType) != types.TypeVoid {
		fmt.Fprintf(&sb, " (result %s)", mapWasmType(uint16(fn.RetType)))
	}
	sb.WriteString("\n")

	// Declare control flow dispatcher state
	sb.WriteString("    (local $state i32)\n")

	// Identify and declare all virtual registers/locals used
	regs := make(map[uint32]string)
	for _, inst := range fn.Insts {
		if inst.Dest != 0 {
			regs[inst.Dest] = mapWasmType(inst.TypeID)
		}
	}
	for r, t := range regs {
		fmt.Fprintf(&sb, "    (local $r%d %s)\n", r, t)
	}

	if len(fn.Blocks) == 0 {
		sb.WriteString("    return\n  )\n")
		return sb.String(), nil
	}

	// Initialize state machine dispatcher to entry block (always block 0)
	sb.WriteString("    (local.set $state (i32.const 0))\n\n")

	// Emit Loop-Switch flat block dispatcher
	sb.WriteString("    (block $outer\n")
	sb.WriteString("      (loop $loop\n")

	// Open all blocks from N down to 0 hierarchically
	for i := len(fn.Blocks) - 1; i >= 0; i-- {
		fmt.Fprintf(&sb, "        (block $b_%d\n", fn.Blocks[i].ID)
	}

	// Emit the br_table dispatcher inside the innermost block
	sb.WriteString("          (br_table")
	for i := 0; i < len(fn.Blocks); i++ {
		fmt.Fprintf(&sb, " $b_%d", fn.Blocks[i].ID)
	}
	// Fallback to the entry block
	fmt.Fprintf(&sb, " $b_0 (local.get $state))\n")

	// Close each block, then emit its instructions
	for i := 0; i < len(fn.Blocks); i++ {
		blk := &fn.Blocks[i]
		fmt.Fprintf(&sb, "        ) ;; end $b_%d\n", blk.ID)
		fmt.Fprintf(&sb, "        ;; --- block_%d body ---\n", blk.ID)

		for _, instIdx := range blk.Instrs {
			if int(instIdx) >= len(fn.Insts) {
				continue
			}
			inst := &fn.Insts[instIdx]
			if inst.Opcode == air.OpNop {
				continue
			}

			b.lowerInst(&sb, fn, inst)
		}
	}

	// Close loop and outer block
	sb.WriteString("      )\n")
	sb.WriteString("    )\n")

	// Ensure there is a return fallback for functions with return value
	if fn.RetType != 0 && types.TypeID(fn.RetType) != types.TypeVoid {
		fmt.Fprintf(&sb, "    (%s.const 0)\n", mapWasmType(uint16(fn.RetType)))
	}
	sb.WriteString("  )\n")

	return sb.String(), nil
}

func (b *WasmBackend) lowerInst(sb *strings.Builder, fn *air.AirFunc, inst *air.AirInst) {
	switch inst.Opcode {
	case air.OpIConst:
		t := mapWasmType(inst.TypeID)
		fmt.Fprintf(sb, "        (%s.const %d)\n", t, int32(inst.Src1))
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpFConst:
		t := mapWasmType(inst.TypeID)
		fmt.Fprintf(sb, "        (%s.const %d)\n", t, int32(inst.Src1)) // truncated / integer bitwise representable
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpCopy, air.OpMove:
		// Initial param copies
		if inst.Src1 <= uint32(len(fn.Params)) && inst.Src1 > 0 {
			fmt.Fprintf(sb, "        (local.get $p%d)\n", inst.Src1)
		} else {
			fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		}
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpIAdd, air.OpISub, air.OpIMul, air.OpIDiv, air.OpIMod,
		air.OpFAdd, air.OpFSub, air.OpFMul, air.OpFDiv:
		t := mapWasmType(inst.TypeID)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src2)

		var op string
		switch inst.Opcode {
		case air.OpIAdd, air.OpFAdd:
			op = "add"
		case air.OpISub, air.OpFSub:
			op = "sub"
		case air.OpIMul, air.OpFMul:
			op = "mul"
		case air.OpIDiv:
			if types.TypeID(inst.TypeID).IsUnsigned() {
				op = "div_u"
			} else {
				op = "div_s"
			}
		case air.OpFDiv:
			op = "div"
		case air.OpIMod:
			if types.TypeID(inst.TypeID).IsUnsigned() {
				op = "rem_u"
			} else {
				op = "rem_s"
			}
		}
		fmt.Fprintf(sb, "        %s.%s\n", t, op)
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpEq, air.OpNe, air.OpLt, air.OpLe, air.OpGt, air.OpGe:
		// Comparison results are always boolean i32 in Wasm
		// Try to infer src type from src1 type in the function
		srcType := "i32"
		for _, inlineInst := range fn.Insts {
			if inlineInst.Dest == inst.Src1 {
				srcType = mapWasmType(inlineInst.TypeID)
				break
			}
		}

		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src2)

		var op string
		switch inst.Opcode {
		case air.OpEq:
			op = "eq"
		case air.OpNe:
			op = "ne"
		case air.OpLt:
			if srcType == "f32" || srcType == "f64" {
				op = "lt"
			} else {
				op = "lt_s"
			}
		case air.OpLe:
			if srcType == "f32" || srcType == "f64" {
				op = "le"
			} else {
				op = "le_s"
			}
		case air.OpGt:
			if srcType == "f32" || srcType == "f64" {
				op = "gt"
			} else {
				op = "gt_s"
			}
		case air.OpGe:
			if srcType == "f32" || srcType == "f64" {
				op = "ge"
			} else {
				op = "ge_s"
			}
		}
		fmt.Fprintf(sb, "        %s.%s\n", srcType, op)
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpAnd, air.OpOr, air.OpXor, air.OpShl, air.OpShr:
		t := mapWasmType(inst.TypeID)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src2)

		var op string
		switch inst.Opcode {
		case air.OpAnd:
			op = "and"
		case air.OpOr:
			op = "or"
		case air.OpXor:
			op = "xor"
		case air.OpShl:
			op = "shl"
		case air.OpShr:
			if types.TypeID(inst.TypeID).IsUnsigned() {
				op = "shr_u"
			} else {
				op = "shr_s"
			}
		}
		fmt.Fprintf(sb, "        %s.%s\n", t, op)
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpNeg:
		t := mapWasmType(inst.TypeID)
		if t == "f32" || t == "f64" {
			fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
			fmt.Fprintf(sb, "        %s.neg\n", t)
		} else {
			// Integer negation: 0 - x
			fmt.Fprintf(sb, "        (%s.const 0)\n", t)
			fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
			fmt.Fprintf(sb, "        %s.sub\n", t)
		}
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpNot:
		t := mapWasmType(inst.TypeID)
		// Bitwise not: x xor -1
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		if t == "i64" {
			fmt.Fprintf(sb, "        (i64.const -1)\n")
		} else {
			fmt.Fprintf(sb, "        (i32.const -1)\n")
		}
		fmt.Fprintf(sb, "        %s.xor\n", t)
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpAlloc, air.OpArenaAlloc:
		// Map type size
		size := uint32(8)
		if inst.TypeID != 0 && b.Table != nil {
			entry := b.Table.Entry(types.TypeID(inst.TypeID))
			if entry.Size != 0 {
				size = entry.Size
			}
		}
		fmt.Fprintf(sb, "        (i32.const %d)\n", size)
		fmt.Fprintf(sb, "        (call $malloc)\n")
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpFree, air.OpDestroy:
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (call $free)\n")

	case air.OpLoad:
		t := mapWasmType(inst.TypeID)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        %s.load\n", t)
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpStore:
		// Get type of Src1 (value to store)
		t := "i32"
		for _, inlineInst := range fn.Insts {
			if inlineInst.Dest == inst.Src1 {
				t = mapWasmType(inlineInst.TypeID)
				break
			}
		}
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src2) // address
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1) // value
		fmt.Fprintf(sb, "        %s.store\n", t)

	case air.OpGEP:
		// Dest = Src1 + Src2 * ElementSize
		elemSize := uint32(8)
		if inst.TypeID != 0 && b.Table != nil {
			entry := b.Table.Entry(types.TypeID(inst.TypeID))
			if entry.Size != 0 {
				elemSize = entry.Size
			}
		}
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src2)
		fmt.Fprintf(sb, "        (i32.const %d)\n", elemSize)
		fmt.Fprintf(sb, "        i32.mul\n")
		fmt.Fprintf(sb, "        i32.add\n")
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpGetField:
		// Field offset
		offset := inst.Src2 * 8
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (i32.const %d)\n", offset)
		fmt.Fprintf(sb, "        i32.add\n")
		t := mapWasmType(inst.TypeID)
		fmt.Fprintf(sb, "        %s.load\n", t)
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	case air.OpSetField:
		// Src1.field[Src2] = Dest
		offset := inst.Src2 * 8
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1) // address base
		fmt.Fprintf(sb, "        (i32.const %d)\n", offset)
		fmt.Fprintf(sb, "        i32.add\n") // address offset
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Dest) // value to store
		// Get Wasm type of value
		t := "i32"
		for _, inlineInst := range fn.Insts {
			if inlineInst.Dest == inst.Dest {
				t = mapWasmType(inlineInst.TypeID)
				break
			}
		}
		fmt.Fprintf(sb, "        %s.store\n", t)

	case air.OpCall:
		// Arguments via Extras
		argStart := inst.Src2
		argCount := uint32(0)
		if argStart < uint32(len(fn.Extras)) {
			argCount = fn.Extras[argStart]
		}
		for j := uint32(0); j < argCount; j++ {
			argReg := fn.Extras[argStart+1+j]
			fmt.Fprintf(sb, "        (local.get $r%d)\n", argReg)
		}
		callee := b.resolveSymName(inst.Src1)
		fmt.Fprintf(sb, "        (call $%s)\n", callee)
		if inst.Dest != 0 {
			fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)
		}

	case air.OpJump:
		fmt.Fprintf(sb, "        (local.set $state (i32.const %d))\n", inst.Src1)
		fmt.Fprintf(sb, "        (br $loop)\n")

	case air.OpBranch:
		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		fmt.Fprintf(sb, "        (if\n")
		fmt.Fprintf(sb, "          (then\n")
		fmt.Fprintf(sb, "            (local.set $state (i32.const %d))\n", inst.Src2)
		fmt.Fprintf(sb, "          )\n")
		fmt.Fprintf(sb, "          (else\n")
		fmt.Fprintf(sb, "            (local.set $state (i32.const %d))\n", inst.Dest)
		fmt.Fprintf(sb, "          )\n")
		fmt.Fprintf(sb, "        )\n")
		fmt.Fprintf(sb, "        (br $loop)\n")

	case air.OpReturn:
		if inst.Src1 != 0 {
			fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		}
		fmt.Fprintf(sb, "        return\n")

	case air.OpCast, air.OpZExt, air.OpSExt, air.OpTrunc:
		// Check src and dst types
		dstT := mapWasmType(inst.TypeID)
		srcT := "i32"
		for _, inlineInst := range fn.Insts {
			if inlineInst.Dest == inst.Src1 {
				srcT = mapWasmType(inlineInst.TypeID)
				break
			}
		}

		fmt.Fprintf(sb, "        (local.get $r%d)\n", inst.Src1)
		if srcT == "i64" && dstT == "i32" {
			fmt.Fprintf(sb, "        i32.wrap_i64\n")
		} else if srcT == "i32" && dstT == "i64" {
			if inst.Opcode == air.OpZExt {
				fmt.Fprintf(sb, "        i64.extend_i32_u\n")
			} else {
				fmt.Fprintf(sb, "        i64.extend_i32_s\n")
			}
		}
		fmt.Fprintf(sb, "        (local.set $r%d)\n", inst.Dest)

	default:
		// OpNop or unsupported opcodes just translate to Wasm nop
		sb.WriteString("        nop\n")
	}
}
