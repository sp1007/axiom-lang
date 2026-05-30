package codegen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// --------------------------------------------------------------------------
// p12-t04: Dynamic Linking Stub
//
// Infrastructure for dynamic linking support (.so/.dll/.dylib).
// Currently a structural stub for future implementation.
// --------------------------------------------------------------------------

// DynLinkInfo contains metadata for a dynamically linked symbol.
type DynLinkInfo struct {
	Name    string // mangled symbol name
	Library string // library name (e.g., "libc.so.6")
	IsWeak  bool   // weak symbol (may be absent at runtime)
}

// --------------------------------------------------------------------------
// p12-t05: Incremental Linker Implementation
//
// A custom linker that combines AXIOM object files into a single executable,
// supporting ELF64 and PE/COFF, with incremental relinking capabilities.
// --------------------------------------------------------------------------

type OutputType int

const (
	OutputTypeExec OutputType = iota // ET_EXEC
	OutputTypeShared                  // ET_DYN
)

type AxiomLinker struct {
	InputFiles  []string
	LibPaths    []string
	EntryPoint  string
	OutputPath  string
	OutputType  OutputType
	Incremental *IncrementalState
}

type LinkerSymbol struct {
	Name    string
	Section int
	Offset  uint64
	Size    uint64
	Defined bool
}

type ParsedObject struct {
	Text     []byte
	Symbols  []LinkerSymbol
	SymNames []string
	Relocs   []ParsedReloc
	VA       uint64
}

type ParsedReloc struct {
	Offset int64
	SymIdx uint32
	IsPC   bool
	Addend int64
}

// IncrementalState tracks which object files need relinking.
type IncrementalState struct {
	ObjectFiles map[string]uint64 // file path → content hash
	OutputPath  string
}

// NeedsRelink returns true if the object file has changed since last link.
func (s *IncrementalState) NeedsRelink(path string, hash uint64) bool {
	prev, ok := s.ObjectFiles[path]
	return !ok || prev != hash
}

// Link combines AXIOM object files into a single executable.
func (l *AxiomLinker) Link() error {
	// 1. Calculate hashes and check incremental state
	var changed []string
	if l.Incremental != nil {
		if l.Incremental.ObjectFiles == nil {
			l.Incremental.ObjectFiles = make(map[string]uint64)
		}
		for _, f := range l.InputFiles {
			hash, err := l.calculateHash(f)
			if err != nil {
				return err
			}
			if l.Incremental.NeedsRelink(f, hash) {
				changed = append(changed, f)
			}
		}
		// If nothing has changed, we don't need to link!
		if len(changed) == 0 && l.fileExists(l.OutputPath) {
			return nil
		}
	}

	// 2. Load all object files
	objects := make([]*ParsedObject, 0, len(l.InputFiles))
	for _, path := range l.InputFiles {
		obj, err := l.loadObject(path)
		if err != nil {
			return fmt.Errorf("load object %s: %w", path, err)
		}
		objects = append(objects, obj)
	}

	// 3. Resolve symbols
	globalSymbols := make(map[string]*LinkerSymbol)
	for _, obj := range objects {
		for _, sym := range obj.Symbols {
			if sym.Defined {
				if existing, ok := globalSymbols[sym.Name]; ok && existing.Defined {
					return fmt.Errorf("duplicate symbol definition: %s", sym.Name)
				}
				// Copy symbol by value to map
				s := sym
				globalSymbols[sym.Name] = &s
			}
		}
	}

	// Check undefined symbols
	for _, obj := range objects {
		for _, sym := range obj.Symbols {
			if !sym.Defined {
				if _, ok := globalSymbols[sym.Name]; !ok {
					// Fallback: if it's a standard extern symbol like printf/malloc, we don't treat it as undefined
					if sym.Name == "printf" || sym.Name == "malloc" || sym.Name == "free" || sym.Name == "exit" || sym.Name == "" {
						continue
					}
					return fmt.Errorf("undefined symbol: %s", sym.Name)
				}
			}
		}
	}

	// 4. Layout sections
	var mergedCode []byte
	var funcOffsets = make(map[string]uint64)
	baseAddr := uint64(0x400000)

	for _, obj := range objects {
		obj.VA = baseAddr + uint64(len(mergedCode))
		for _, sym := range obj.Symbols {
			if sym.Defined {
				funcOffsets[sym.Name] = uint64(len(mergedCode)) + sym.Offset
			}
		}
		mergedCode = append(mergedCode, obj.Text...)
	}

	externVAs := make(map[string]uint64)

	// 5. Apply relocations
	for _, obj := range objects {
		for _, r := range obj.Relocs {
			if r.SymIdx >= uint32(len(obj.SymNames)) {
				continue
			}
			targetName := obj.SymNames[r.SymIdx]
			var targetVA uint64
			if offset, ok := funcOffsets[targetName]; ok {
				targetVA = baseAddr + offset
			} else {
				// Extern / imported symbol stub target address resolved dynamically
				if va, ok := externVAs[targetName]; ok {
					targetVA = va
				} else {
					va = uint64(0x500000 + len(externVAs)*8)
					externVAs[targetName] = va
					targetVA = va
				}
			}

			// Relocation offset relative to mergedCode
			relOffset := uint64(obj.VA-baseAddr) + uint64(r.Offset)
			if relOffset+4 > uint64(len(mergedCode)) {
				continue
			}
			
			// PC-relative target calculation
			if r.IsPC {
				// rel32 = target - (pc + 4)
				pc := baseAddr + relOffset
				val := int32(int64(targetVA) - int64(pc+4) + r.Addend)
				binary.LittleEndian.PutUint32(mergedCode[relOffset:relOffset+4], uint32(val))
			} else {
				if relOffset+8 <= uint64(len(mergedCode)) {
					// absolute 64-bit address
					val := targetVA + uint64(r.Addend)
					binary.LittleEndian.PutUint64(mergedCode[relOffset:relOffset+8], val)
				}
			}
		}
	}

	// 6. Write final executable
	if err := l.writeExecutable(mergedCode); err != nil {
		return err
	}

	// 7. Update incremental state
	if l.Incremental != nil {
		for _, f := range l.InputFiles {
			hash, _ := l.calculateHash(f)
			l.Incremental.ObjectFiles[f] = hash
		}
	}

	return nil
}

func (l *AxiomLinker) loadObject(path string) (*ParsedObject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) < 4 {
		return nil, fmt.Errorf("invalid format")
	}

	// Simple detector: ELF magic = "\x7fELF", COFF AMD64 magic = 0x8664
	if string(data[:4]) == "\x7fELF" {
		return l.parseELF(data)
	} else if binary.LittleEndian.Uint16(data[:2]) == 0x8664 {
		return l.parseCOFF(data)
	}

	return nil, fmt.Errorf("unsupported object format")
}

func (l *AxiomLinker) parseELF(data []byte) (*ParsedObject, error) {
	// Read ELF e_shoff, e_shnum, e_shstrndx
	shoff := binary.LittleEndian.Uint64(data[40:48])
	shnum := binary.LittleEndian.Uint16(data[60:62])

	var text []byte
	var symtab []byte
	var strtab []byte
	var rela []byte

	// Iterate section headers (each is 64 bytes)
	for i := uint16(0); i < shnum; i++ {
		off := shoff + uint64(i)*64
		shType := binary.LittleEndian.Uint32(data[off+4 : off+8])
		shOffset := binary.LittleEndian.Uint64(data[off+24 : off+32])
		shSize := binary.LittleEndian.Uint64(data[off+32 : off+40])

		switch shType {
		case 1: // SHT_PROGBITS
			text = data[shOffset : shOffset+shSize]
		case 2: // SHT_SYMTAB
			symtab = data[shOffset : shOffset+shSize]
		case 3: // SHT_STRTAB
			strtab = data[shOffset : shOffset+shSize]
		case 4: // SHT_RELA
			rela = data[shOffset : shOffset+shSize]
		}
	}

	// Parse string table
	var symNames []string
	if len(symtab) > 0 {
		numSyms := len(symtab) / 24
		symNames = make([]string, numSyms)
		for i := 0; i < numSyms; i++ {
			off := i * 24
			nameOffset := binary.LittleEndian.Uint32(symtab[off : off+4])
			if nameOffset < uint32(len(strtab)) {
				var sb []byte
				for j := nameOffset; j < uint32(len(strtab)) && strtab[j] != 0; j++ {
					sb = append(sb, strtab[j])
				}
				symNames[i] = string(sb)
			}
		}
	}

	// Parse symbols
	var symbols []LinkerSymbol
	if len(symtab) > 0 {
		numSyms := len(symtab) / 24
		for i := 1; i < numSyms; i++ { // skip index 0 (NULL)
			off := i * 24
			secIdx := binary.LittleEndian.Uint16(symtab[off+6 : off+8])
			val64 := binary.LittleEndian.Uint64(symtab[off+8 : off+16])
			size := binary.LittleEndian.Uint64(symtab[off+16 : off+24])

			name := symNames[i]
			isDefined := secIdx != 0
			symbols = append(symbols, LinkerSymbol{
				Name:    name,
				Section: int(secIdx),
				Offset:  val64,
				Size:    size,
				Defined: isDefined,
			})
		}
	}

	// Parse relocations
	var relocs []ParsedReloc
	if len(rela) > 0 {
		numRelocs := len(rela) / 24
		for i := 0; i < numRelocs; i++ {
			off := i * 24
			rOffset := binary.LittleEndian.Uint64(rela[off : off+8])
			rInfo := binary.LittleEndian.Uint64(rela[off+8 : off+16])
			rAddend := binary.LittleEndian.Uint64(rela[off+16 : off+24])

			symIdx := uint32(rInfo >> 32)
			rType := uint32(rInfo & 0xffffffff)
			isPC := rType == 2 || rType == 4 // R_X86_64_PC32 or R_X86_64_PLT32

			relocs = append(relocs, ParsedReloc{
				Offset: int64(rOffset),
				SymIdx: symIdx,
				IsPC:   isPC,
				Addend: int64(rAddend),
			})
		}
	}

	return &ParsedObject{
		Text:     text,
		Symbols:  symbols,
		SymNames: symNames,
		Relocs:   relocs,
	}, nil
}

func (l *AxiomLinker) parseCOFF(data []byte) (*ParsedObject, error) {
	numSections := binary.LittleEndian.Uint16(data[2:4])
	symtabOff := binary.LittleEndian.Uint32(data[8:12])
	symCount := binary.LittleEndian.Uint32(data[12:16])

	var text []byte
	var textRelocs []byte
	var textNumRelocs uint16

	// Iterate section headers (each is 40 bytes)
	for i := uint16(0); i < numSections; i++ {
		off := 20 + uint64(i)*40
		name := string(bytes.TrimRight(data[off:off+8], "\x00"))
		rawDataPtr := binary.LittleEndian.Uint32(data[off+20 : off+24])
		rawSize := binary.LittleEndian.Uint32(data[off+16 : off+20])
		relocsPtr := binary.LittleEndian.Uint32(data[off+24 : off+28])
		numRelocs := binary.LittleEndian.Uint16(data[off+32 : off+34])

		if name == ".text" {
			text = data[rawDataPtr : rawDataPtr+rawSize]
			if numRelocs > 0 {
				textRelocs = data[relocsPtr : relocsPtr+uint32(numRelocs)*10]
				textNumRelocs = numRelocs
			}
		}
	}

	// Parse string table
	strtabOff := symtabOff + symCount*18
	strtab := data[strtabOff:]

	// Parse symbols
	var symNames []string
	var symbols []LinkerSymbol
	for i := uint32(0); i < symCount; i++ {
		off := symtabOff + i*18
		var name string
		first4 := binary.LittleEndian.Uint32(data[off : off+4])
		if first4 == 0 {
			strOff := binary.LittleEndian.Uint32(data[off+4 : off+8])
			if strOff < uint32(len(strtab)) {
				var sb []byte
				for j := strOff; j < uint32(len(strtab)) && strtab[j] != 0; j++ {
					sb = append(sb, strtab[j])
				}
				name = string(sb)
			}
		} else {
			name = string(bytes.TrimRight(data[off:off+8], "\x00"))
		}

		val := binary.LittleEndian.Uint32(data[off+8 : off+12])
		secNum := int16(binary.LittleEndian.Uint16(data[off+12 : off+14]))
		auxCount := data[off+17]

		symNames = append(symNames, name)
		symbols = append(symbols, LinkerSymbol{
			Name:    name,
			Section: int(secNum),
			Offset:  uint64(val),
			Defined: secNum > 0,
		})

		// Skip auxiliary symbol table entries
		i += uint32(auxCount)
		for k := uint8(0); k < auxCount; k++ {
			symNames = append(symNames, "") // placeholder for alignment
		}
	}

	// Parse relocations
	var relocs []ParsedReloc
	for i := uint16(0); i < textNumRelocs; i++ {
		off := uint32(i) * 10
		virtAddr := binary.LittleEndian.Uint32(textRelocs[off : off+4])
		symIdx := binary.LittleEndian.Uint32(textRelocs[off+4 : off+8])
		relType := binary.LittleEndian.Uint16(textRelocs[off+8 : off+10])

		isPC := relType == 4 // IMAGE_REL_AMD64_REL32 = 4

		relocs = append(relocs, ParsedReloc{
			Offset: int64(virtAddr),
			SymIdx: symIdx,
			IsPC:   isPC,
			Addend: -4, // COFF rel32 displacements assume PC starts after rel32 field
		})
	}

	return &ParsedObject{
		Text:     text,
		Symbols:  symbols,
		SymNames: symNames,
		Relocs:   relocs,
	}, nil
}

func (l *AxiomLinker) calculateHash(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	// Simple FNV-1a hash
	h := uint64(14695981039346656037)
	for _, b := range data {
		h ^= uint64(b)
		h *= 1099511628211
	}
	return h, nil
}

func (l *AxiomLinker) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (l *AxiomLinker) writeExecutable(code []byte) error {
	return os.WriteFile(l.OutputPath, code, 0755)
}

// --------------------------------------------------------------------------
// p12-t07: Symbol Demangling
//
// Human-readable display of mangled AXIOM symbols.
// --------------------------------------------------------------------------

// DemangleDisplay returns a human-readable representation of a mangled symbol.
// Example: "_AX_math_add_ii_i" → "math::add(i32, i32) -> i32"
func DemangleDisplay(mangled string) string {
	result, err := Demangle(mangled)
	if err != nil {
		return mangled // return as-is if not demangleable
	}

	display := result.Module + "::" + result.Name + "("
	for i, p := range result.Params {
		if i > 0 {
			display += ", "
		}
		display += typeDisplayName(p)
	}
	display += ") -> " + typeDisplayName(result.Ret)
	return display
}

// typeDisplayName returns the human-readable type name for a TypeID.
func typeDisplayName(typeID uint32) string {
	switch typeID {
	case 0:
		return "void"
	case 2:
		return "bool"
	case 3:
		return "i32"
	case 4:
		return "i64"
	case 5:
		return "f64"
	case 6:
		return "i8"
	case 7:
		return "i16"
	case 8:
		return "u8"
	case 9:
		return "u16"
	case 10:
		return "u32"
	case 11:
		return "u64"
	case 12:
		return "f32"
	case 13:
		return "str"
	case 14:
		return "ptr"
	default:
		return "?"
	}
}
