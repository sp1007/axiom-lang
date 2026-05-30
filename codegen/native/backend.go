package native

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/codegen/native/x86"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
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
	Pool     *ast.InternPool
	Table    *types.TypeTable
}

// NewNativeBackend creates a new native backend for the given target.
func NewNativeBackend(target Target) *NativeBackend {
	return &NativeBackend{
		Target:   target,
		OptLevel: opt.O1,
	}
}

// Compile generates native object code from an AirModule.
// Returns the raw bytes of the object file (ELF64 for Linux, PE/COFF for Windows, Mach-O for macOS).
func (b *NativeBackend) Compile(mod *air.AirModule) ([]byte, error) {
	// Step 1: Run optimization pipeline
	pipeline := opt.DefaultPipeline(b.OptLevel, false)
	pipeline.Run(mod)

	// Step 2: For each function, run the codegen pipeline
	abi := x86.NewABI(b.Target.ABI.String())

	switch b.Target.BinaryFormat() {
	case BinPE:
		coff := x86.NewCOFFWriter()
		var allCode []byte
		var funcOffsets []uint32
		var funcSyms []x86.ELF64Sym
		var funcRelocs [][]x86.Relocation

		// Map to keep track of added symbol indices by their name string
		symIndices := make(map[string]int)

		// First register all extern functions as undefined external symbols
		for fi := range mod.Funcs {
			fn := &mod.Funcs[fi]
			if fn.IsExtern {
				name := b.resolveSymName(fn.Name)
				symIdx := coff.AddSymbol(name, 0, 0, true)
				symIndices[name] = symIdx
			}
		}

		for fi := range mod.Funcs {
			fn := &mod.Funcs[fi]
			if fn.IsExtern {
				continue
			}

			code, sym, relocs := b.compileFunc(fn, abi)
			funcOffsets = append(funcOffsets, uint32(len(allCode)))
			funcSyms = append(funcSyms, sym)
			funcRelocs = append(funcRelocs, relocs)
			allCode = append(allCode, code...)
		}

		textSecIdx := coff.AddSection(".text", x86.IMAGE_SCN_CNT_CODE|x86.IMAGE_SCN_MEM_EXECUTE|x86.IMAGE_SCN_MEM_READ, allCode)
		for i, sym := range funcSyms {
			symIdx := coff.AddSymbol(sym.Name, textSecIdx, funcOffsets[i], sym.Binding == x86.STB_GLOBAL)
			symIndices[sym.Name] = symIdx
		}

		// Add relocations for external calls
		for i, relocs := range funcRelocs {
			funcOffset := funcOffsets[i]
			for _, r := range relocs {
				targetName := b.resolveSymName(r.SymName)
				symIdx, exists := symIndices[targetName]
				if !exists {
					symIdx = coff.AddSymbol(targetName, 0, 0, true)
					symIndices[targetName] = symIdx
				}

				var relocType uint16 = x86.IMAGE_REL_AMD64_REL32
				if r.Kind == x86.RelocAbs64 {
					relocType = x86.IMAGE_REL_AMD64_ADDR64
				}

				relocOffset := int(funcOffset) + r.Offset
				coff.AddReloc(textSecIdx, relocOffset, symIdx, relocType)
			}
		}

		return coff.Serialize(), nil

	case BinMachO:
		macho := x86.NewMachOWriter()
		for fi := range mod.Funcs {
			fn := &mod.Funcs[fi]
			if fn.IsExtern {
				continue
			}

			code, sym, _ := b.compileFunc(fn, abi)
			macho.SetText(code) // For MVP: single function per object
			macho.AddSymbol(sym)
		}
		return macho.Serialize(), nil

	default: // BinELF
		elf := x86.NewELF64Writer()
		for fi := range mod.Funcs {
			fn := &mod.Funcs[fi]
			if fn.IsExtern {
				continue
			}

			code, sym, _ := b.compileFunc(fn, abi)
			elf.SetText(code) // For MVP: single function per object
			elf.AddSymbol(sym)
		}
		return elf.Serialize(), nil
	}
}

// resolveSymName resolves a symbol ID to a string, using InternPool if available.
func (b *NativeBackend) resolveSymName(symID uint32) string {
	if symID == 0 {
		return "main"
	}
	if symID == 4294967295 {
		return "ax_alloc"
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

// compileFunc runs the full codegen pipeline for a single function.
func (b *NativeBackend) compileFunc(fn *air.AirFunc, abi x86.ABI) ([]byte, x86.ELF64Sym, []x86.Relocation) {
	// Step 1: Instruction selection (AIR → MachInst with VRegs)
	machInsts := x86.Select(fn, abi, b.Table)

	// Step 2 & 3: Register allocation with split GPR & XMM classes
	allocResult := b.allocateRegisters(fn, machInsts, abi)

	// Step 4: Compute stack frame
	calleeSaved := computeUsedCalleeSaved(allocResult, abi)
	frame := x86.ComputeFrame(calleeSaved, allocResult.SpillCount, 0)

	// Step 5: Insert spill code
	machInsts = x86.InsertSpillCode(machInsts, allocResult.Allocs, &frame)

	// Step 6: Emit machine code
	emitter := x86.NewEmitter(allocResult.Allocs)
	emitter.EmitFunction(machInsts, &frame)

	// Build symbol
	name := b.resolveSymName(fn.Name)
	sym := x86.ELF64Sym{
		Name:    name,
		Value:   0,
		Size:    uint64(emitter.CodeSize()),
		Binding: x86.STB_GLOBAL,
		Type:    x86.STT_FUNC,
		Section: 1, // .text
	}

	return emitter.Code, sym, emitter.Relocs
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

func allocatableCalleeSaved(abi x86.ABI) []x86.PhysReg {
	var regs []x86.PhysReg
	for _, r := range abi.CalleeSavedRegs() {
		if r != x86.RBP && r != x86.RSP {
			regs = append(regs, r)
		}
	}
	return regs
}

// CompileAsm compiles an AirModule directly into NASM, FASM, or WinAsm (MASM) assembly text format.
func (b *NativeBackend) CompileAsm(mod *air.AirModule, format string) (string, error) {
	// Step 1: Run optimization pipeline
	pipeline := opt.DefaultPipeline(b.OptLevel, false)
	pipeline.Run(mod)

	var sb strings.Builder
	abi := x86.NewABI(b.Target.ABI.String())

	// Emit assembler-specific header
	if format == "fasm" {
		sb.WriteString("; Generated by AXIOM Compiler (FASM format)\n")
		sb.WriteString("format ELF64 executable\n")
		sb.WriteString("segment readable executable\n\n")
	} else if format == "winasm" {
		sb.WriteString("; Generated by AXIOM Compiler (WinAsm / MASM format)\n")
		sb.WriteString(".code\n\n")
	} else {
		// Default to NASM
		sb.WriteString("; Generated by AXIOM Compiler (NASM format)\n")
		sb.WriteString("bits 64\n")
		sb.WriteString("section .text\n\n")
	}

	// Declare external symbols
	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		if fn.IsExtern {
			name := b.resolveSymName(fn.Name)
			if format == "fasm" {
				fmt.Fprintf(&sb, "extrn %s\n", name)
			} else if format == "winasm" {
				fmt.Fprintf(&sb, "EXTERN %s:PROC\n", name)
			} else {
				fmt.Fprintf(&sb, "extern %s\n", name)
			}
		}
	}
	sb.WriteString("\n")

	// Declare global symbols
	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		if fn.IsExtern {
			continue
		}
		name := b.resolveSymName(fn.Name)
		if format == "fasm" {
			fmt.Fprintf(&sb, "public %s\n", name)
		} else if format == "winasm" {
			fmt.Fprintf(&sb, "PUBLIC %s\n", name)
		} else {
			fmt.Fprintf(&sb, "global %s\n", name)
		}
	}
	sb.WriteString("\n")

	// Compile each function to assembly text
	for fi := range mod.Funcs {
		fn := &mod.Funcs[fi]
		if fn.IsExtern {
			continue
		}

		// Run the backend code selection & register allocation pipeline
		machInsts := x86.Select(fn, abi, b.Table)
		allocResult := b.allocateRegisters(fn, machInsts, abi)
		calleeSaved := computeUsedCalleeSaved(allocResult, abi)
		frame := x86.ComputeFrame(calleeSaved, allocResult.SpillCount, 0)
		machInsts = x86.InsertSpillCode(machInsts, allocResult.Allocs, &frame)

		// Emit ASM text using the AsmEmitter
		asmEmitter := x86.NewAsmEmitter(allocResult.Allocs, format)
		fnName := b.resolveSymName(fn.Name)
		asmText := asmEmitter.EmitFunction(fnName, machInsts, &frame, func(id uint32) string {
			return b.resolveSymName(id)
		})
		sb.WriteString(asmText)
		sb.WriteString("\n")
	}

	if format == "winasm" {
		sb.WriteString("END\n")
	}

	return sb.String(), nil
}

func (b *NativeBackend) isFloatVReg(fn *air.AirFunc, vreg uint32) bool {
	if vreg == 0 {
		return false
	}
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Dest == vreg {
			return types.TypeID(inst.TypeID).IsFloat()
		}
	}
	return false
}

func (b *NativeBackend) is16ByteVReg(fn *air.AirFunc, vreg uint32) bool {
	if vreg == 0 {
		return false
	}
	for i := range fn.Insts {
		inst := &fn.Insts[i]
		if inst.Dest == vreg {
			t := inst.TypeID
			if t != 0 && b.Table != nil {
				entry := b.Table.Entry(types.TypeID(t))
				if entry.Size == 16 {
					return true
				}
			}
		}
	}
	if vreg >= 1 && vreg <= uint32(len(fn.Params)) {
		t := fn.Params[vreg-1]
		if t != 0 && b.Table != nil {
			entry := b.Table.Entry(types.TypeID(t))
			if entry.Size == 16 {
				return true
			}
		}
	}
	return false
}

func (b *NativeBackend) allocateRegisters(fn *air.AirFunc, machInsts []x86.MachInst, abi x86.ABI) x86.RegAllocResult {
	intervals := x86.ComputeLiveness(machInsts)

	var gprIntervals []x86.LiveInterval
	var xmmIntervals []x86.LiveInterval
	var struct16Intervals []x86.LiveInterval
	for _, iv := range intervals {
		if b.is16ByteVReg(fn, iv.VReg) {
			struct16Intervals = append(struct16Intervals, iv)
		} else if b.isFloatVReg(fn, iv.VReg) {
			xmmIntervals = append(xmmIntervals, iv)
		} else {
			gprIntervals = append(gprIntervals, iv)
		}
	}

	gprRegs := allocatableCalleeSaved(abi)
	xmmRegs := x86.AllocatableXMMs()

	gprAlloc := x86.GraphColoringAlloc(gprIntervals, gprRegs)
	xmmAlloc := x86.GraphColoringAlloc(xmmIntervals, xmmRegs)

	allocResult := x86.RegAllocResult{
		Allocs: make(map[uint32]x86.RegAllocation),
	}
	for k, v := range gprAlloc.Allocs {
		v.Is16 = b.is16ByteVReg(fn, k)
		allocResult.Allocs[k] = v
	}

	spillCount := gprAlloc.SpillCount
	for _, iv := range struct16Intervals {
		allocResult.Allocs[iv.VReg] = x86.RegAllocation{
			VReg:     iv.VReg,
			Phys:     x86.RegNone,
			Spilled:  true,
			SpillIdx: spillCount,
			Is16:     true,
		}
		spillCount += 2
	}

	for k, v := range xmmAlloc.Allocs {
		if v.Spilled {
			v.SpillIdx += spillCount
		}
		v.Is16 = b.is16ByteVReg(fn, k)
		allocResult.Allocs[k] = v
	}
	allocResult.SpillCount = spillCount + xmmAlloc.SpillCount

	return allocResult
}

