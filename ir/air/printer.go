// Package air — human-readable text printer for AIR.
//
// The printer emits a compact, deterministic text format suitable for
// golden tests, debugging, and IR dumps. It mirrors the assembly-like
// syntax common in compiler IRs (LLVM IR, Cranelift CLIF, etc.).
package air

import (
	"fmt"
	"io"
	"strings"
)

// ---------------------------------------------------------------------------
// PrintFunc — write a single function in text form.
// ---------------------------------------------------------------------------

// PrintFunc writes a human-readable text representation of fn to w.
// The output is deterministic and suitable for snapshot testing.
func PrintFunc(w io.Writer, fn *AirFunc) {
	// Function header: fn @<name>(params...):
	fmt.Fprintf(w, "fn @%d(", fn.Name)
	for i, p := range fn.Params {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "t%d", p)
	}
	fmt.Fprint(w, ")")
	if fn.RetType != 0 {
		fmt.Fprintf(w, " -> t%d", fn.RetType)
	}
	fmt.Fprintln(w, ":")

	for bi := range fn.Blocks {
		blk := &fn.Blocks[bi]
		printBlock(w, fn, blk)
	}
}

// ---------------------------------------------------------------------------
// PrintModule — write all functions in a module.
// ---------------------------------------------------------------------------

// PrintModule writes all functions in mod to w, separated by blank lines.
func PrintModule(w io.Writer, mod *AirModule) {
	for i := range mod.Funcs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		PrintFunc(w, &mod.Funcs[i])
	}
}

// ---------------------------------------------------------------------------
// SprintFunc — convenience wrapper returning a string.
// ---------------------------------------------------------------------------

// SprintFunc returns the text representation of fn as a string.
func SprintFunc(fn *AirFunc) string {
	var sb strings.Builder
	PrintFunc(&sb, fn)
	return sb.String()
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func printBlock(w io.Writer, fn *AirFunc, blk *BasicBlock) {
	// Block label with annotations.
	fmt.Fprintf(w, "  block_%d:", blk.ID)
	if blk.IsEntry && blk.IsExit {
		fmt.Fprint(w, "  ; entry exit")
	} else if blk.IsEntry {
		fmt.Fprint(w, "  ; entry")
	} else if blk.IsExit {
		fmt.Fprint(w, "  ; exit")
	}
	fmt.Fprintln(w)

	for _, instIdx := range blk.Instrs {
		if int(instIdx) >= len(fn.Insts) {
			fmt.Fprintf(w, "    ; ERROR: instruction index %d out of range\n", instIdx)
			continue
		}
		inst := fn.Insts[instIdx]
		if inst.Opcode == OpNop {
			continue // skip NOPs
		}
		fmt.Fprint(w, "    ")
		printInst(w, &inst)
		fmt.Fprintln(w)
	}
}

func printInst(w io.Writer, inst *AirInst) {
	switch inst.Opcode {
	case OpReturn:
		printReturn(w, inst)
	case OpJump:
		printJump(w, inst)
	case OpBranch:
		printBranch(w, inst)
	default:
		printGeneric(w, inst)
	}
}

func printReturn(w io.Writer, inst *AirInst) {
	if inst.Src1 != 0 {
		fmt.Fprintf(w, "ret %%%d", inst.Src1)
	} else {
		fmt.Fprint(w, "ret")
	}
}

func printJump(w io.Writer, inst *AirInst) {
	fmt.Fprintf(w, "jump block_%d", inst.Src1)
}

func printBranch(w io.Writer, inst *AirInst) {
	// branch %cond block_true block_false
	// Src1 = condition, Src2 = true target, Dest = false target
	fmt.Fprintf(w, "branch %%%d block_%d block_%d", inst.Src1, inst.Src2, inst.Dest)
}

func printGeneric(w io.Writer, inst *AirInst) {
	mnemonic := inst.Opcode.Mnemonic()

	hasDest := inst.Dest != 0
	hasSrc1 := inst.Src1 != 0
	hasSrc2 := inst.Src2 != 0
	isBinary := inst.Opcode.IsBinaryALU()

	if hasDest {
		// Value instruction: %Dest: tN = mnemonic ...
		if inst.TypeID != 0 {
			fmt.Fprintf(w, "%%%d: t%d = %s", inst.Dest, inst.TypeID, mnemonic)
		} else {
			fmt.Fprintf(w, "%%%d = %s", inst.Dest, mnemonic)
		}
	} else {
		// Void instruction: mnemonic ...
		fmt.Fprint(w, mnemonic)
	}

	// Operands
	if isBinary && hasSrc1 && hasSrc2 {
		fmt.Fprintf(w, " %%%d, %%%d", inst.Src1, inst.Src2)
	} else if hasSrc1 && hasSrc2 {
		fmt.Fprintf(w, " %%%d, %%%d", inst.Src1, inst.Src2)
	} else if hasSrc1 {
		fmt.Fprintf(w, " %%%d", inst.Src1)
	} else if hasSrc2 {
		fmt.Fprintf(w, " %%%d", inst.Src2)
	}
}
