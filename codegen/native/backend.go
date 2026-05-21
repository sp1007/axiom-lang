package native

import (
	"github.com/axiom-lang/axiom/codegen/native/x86"
	"github.com/axiom-lang/axiom/ir/air"
	"github.com/axiom-lang/axiom/ir/opt"
)

// --------------------------------------------------------------------------
// p11-t15: Native Backend Integration
//
// Orchestrates the full native code generation pipeline:
// AIR → optimize → select → liveness → regalloc → spill → frame → emit → ELF
// --------------------------------------------------------------------------

// NativeBackend coordinates native code generation for a target.
type NativeBackend struct {
	Target   Target
	OptLevel opt.OptLevel
}

// NewNativeBackend creates a new native backend for the given target.
func NewNativeBackend(target Target) *NativeBackend {
	return &NativeBackend{
		Target:   target,
		OptLevel: opt.O1,
	}
}

// Compile generates native object code from an AirModule.
// Returns the raw bytes of the object file (ELF64 for Linux, etc.).
func (b *NativeBackend) Compile(mod *air.AirModule) ([]byte, error) {
	// Step 1: Run optimization pipeline
	pipeline := opt.DefaultPipeline(b.OptLevel, false)
	pipeline.Run(mod)

	// Step 2: For each function, run the codegen pipeline
	abi := x86.NewABI(b.Target.ABI.String())
	elf := x86.NewELF64Writer()

	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		if fn.IsExtern {
			continue
		}

		code, sym := b.compileFunc(fn, abi)
		elf.SetText(code) // For MVP: single function per object
		elf.AddSymbol(sym)
	}

	return elf.Serialize(), nil
}

// compileFunc runs the full codegen pipeline for a single function.
func (b *NativeBackend) compileFunc(fn *air.AirFunc, abi x86.ABI) ([]byte, x86.ELF64Sym) {
	// Step 1: Instruction selection (AIR → MachInst with VRegs)
	machInsts := x86.Select(fn)

	// Step 2: Liveness analysis
	intervals := x86.ComputeLiveness(machInsts)

	// Step 3: Register allocation
	availRegs := x86.AllocatableGPRs()
	allocResult := x86.LinearScanAlloc(intervals, availRegs)

	// Step 4: Compute stack frame
	calleeSaved := computeUsedCalleeSaved(allocResult, abi)
	frame := x86.ComputeFrame(calleeSaved, allocResult.SpillCount, 0)

	// Step 5: Insert spill code
	machInsts = x86.InsertSpillCode(machInsts, allocResult.Allocs, &frame)

	// Step 6: Emit machine code
	emitter := x86.NewEmitter(allocResult.Allocs)
	emitter.EmitFunction(machInsts, &frame)

	// Build symbol
	sym := x86.ELF64Sym{
		Name:    "main", // TODO: resolve from intern pool
		Value:   0,
		Size:    uint64(emitter.CodeSize()),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1, // .text
	}

	return emitter.Code, sym
}

// computeUsedCalleeSaved determines which callee-saved registers were
// actually used by the register allocator.
func computeUsedCalleeSaved(result x86.RegAllocResult, abi x86.ABI) []x86.PhysReg {
	calleeSaved := abi.CalleeSavedRegs()
	usedSet := make(map[x86.PhysReg]bool)

	for _, alloc := range result.Allocs {
		if !alloc.Spilled {
			usedSet[alloc.Phys] = true
		}
	}

	var used []x86.PhysReg
	for _, reg := range calleeSaved {
		if usedSet[reg] {
			used = append(used, reg)
		}
	}
	return used
}
