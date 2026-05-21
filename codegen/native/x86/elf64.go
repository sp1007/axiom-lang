package x86

import (
	"encoding/binary"
)

// --------------------------------------------------------------------------
// p11-t12: ELF64 Object File Emitter
//
// Produces a complete ELF64 relocatable object file (.o) from the
// emitted machine code. Includes .text, .symtab, .strtab, and
// .rela.text sections.
// --------------------------------------------------------------------------

// ELF64 constants
const (
	ELF_MAGIC      = "\x7fELF"
	ELFCLASS64     = 2
	ELFDATA2LSB    = 1
	EV_CURRENT     = 1
	ET_REL         = 1  // relocatable
	EM_X86_64      = 62
	SHT_NULL       = 0
	SHT_PROGBITS   = 1
	SHT_SYMTAB     = 2
	SHT_STRTAB     = 3
	SHT_RELA       = 4
	SHF_ALLOC      = 0x02
	SHF_EXECINSTR  = 0x04
	STB_LOCAL      = 0
	STB_GLOBAL     = 1
	STT_FUNC       = 2
	STT_SECTION    = 3
)

// ELF64Sym represents an ELF64 symbol table entry.
type ELF64Sym struct {
	Name    string
	Value   uint64
	Size    uint64
	Binding uint8
	Type    uint8
	Section uint16
}

// ELF64Writer builds an ELF64 relocatable object file.
type ELF64Writer struct {
	textData []byte      // .text section contents
	syms     []ELF64Sym  // symbols
	relocs   []Relocation // relocations for .text
}

// NewELF64Writer creates a new ELF64 writer.
func NewELF64Writer() *ELF64Writer {
	return &ELF64Writer{}
}

// SetText sets the .text section contents.
func (w *ELF64Writer) SetText(code []byte) {
	w.textData = code
}

// AddSymbol adds a symbol to the symbol table.
func (w *ELF64Writer) AddSymbol(sym ELF64Sym) {
	w.syms = append(w.syms, sym)
}

// AddRelocation adds a relocation entry.
func (w *ELF64Writer) AddRelocation(r Relocation) {
	w.relocs = append(w.relocs, r)
}

// Serialize produces the complete ELF64 object file as bytes.
func (w *ELF64Writer) Serialize() []byte {
	// Build string table
	strtab := buildStringTable(w.syms)

	// Section layout:
	// 0: NULL
	// 1: .text
	// 2: .symtab
	// 3: .strtab
	// 4: .rela.text (if relocations exist)
	numSections := uint16(4)
	if len(w.relocs) > 0 {
		numSections = 5
	}

	// ELF header: 64 bytes
	ehdrSize := 64
	// Section headers: 64 bytes each
	shdrSize := 64
	shdrTableSize := int(numSections) * shdrSize

	// Compute section offsets
	textOff := ehdrSize
	textSize := len(w.textData)

	symtabOff := textOff + textSize
	// Align to 8 bytes
	if symtabOff%8 != 0 {
		symtabOff += 8 - symtabOff%8
	}

	// Symbol table: 24 bytes per entry + 1 NULL entry
	symEntSize := 24
	symCount := 1 + len(w.syms) // NULL + user symbols
	symtabSize := symCount * symEntSize

	strtabOff := symtabOff + symtabSize
	strtabSize := len(strtab.data)

	relaOff := strtabOff + strtabSize
	relaEntSize := 24
	relaSize := len(w.relocs) * relaEntSize

	shdrOff := relaOff + relaSize
	if len(w.relocs) == 0 {
		shdrOff = strtabOff + strtabSize
	}
	// Align to 8
	if shdrOff%8 != 0 {
		shdrOff += 8 - shdrOff%8
	}

	totalSize := shdrOff + shdrTableSize
	buf := make([]byte, totalSize)

	// Write ELF header
	copy(buf[0:4], ELF_MAGIC)
	buf[4] = ELFCLASS64
	buf[5] = ELFDATA2LSB
	buf[6] = EV_CURRENT
	// buf[7..15] = 0 (padding)
	binary.LittleEndian.PutUint16(buf[16:], ET_REL)
	binary.LittleEndian.PutUint16(buf[18:], EM_X86_64)
	binary.LittleEndian.PutUint32(buf[20:], EV_CURRENT)
	// e_entry = 0 (relocatable, no entry point)
	// e_phoff = 0 (no program headers)
	binary.LittleEndian.PutUint64(buf[40:], uint64(shdrOff))  // e_shoff
	binary.LittleEndian.PutUint16(buf[52:], uint16(ehdrSize)) // e_ehsize
	binary.LittleEndian.PutUint16(buf[58:], uint16(shdrSize)) // e_shentsize
	binary.LittleEndian.PutUint16(buf[60:], numSections)       // e_shnum
	binary.LittleEndian.PutUint16(buf[62:], 3)                 // e_shstrndx = .strtab

	// Write .text section
	copy(buf[textOff:], w.textData)

	// Write symbol table
	off := symtabOff
	// NULL entry (24 zero bytes)
	off += symEntSize

	for _, sym := range w.syms {
		nameIdx := strtab.lookup(sym.Name)
		binary.LittleEndian.PutUint32(buf[off:], uint32(nameIdx))
		buf[off+4] = (sym.Binding << 4) | sym.Type
		buf[off+5] = 0 // st_other
		binary.LittleEndian.PutUint16(buf[off+6:], sym.Section)
		binary.LittleEndian.PutUint64(buf[off+8:], sym.Value)
		binary.LittleEndian.PutUint64(buf[off+16:], sym.Size)
		off += symEntSize
	}

	// Write string table
	copy(buf[strtabOff:], strtab.data)

	// Write relocations
	if len(w.relocs) > 0 {
		off = relaOff
		for _, r := range w.relocs {
			binary.LittleEndian.PutUint64(buf[off:], uint64(r.Offset))
			// r_info: symbol index in high 32 bits, type in low 32 bits
			rType := uint32(1) // R_X86_64_64 by default
			switch r.Kind {
			case RelocPC32:
				rType = 2 // R_X86_64_PC32
			case RelocPLT32:
				rType = 4 // R_X86_64_PLT32
			case RelocAbs64:
				rType = 1 // R_X86_64_64
			}
			binary.LittleEndian.PutUint64(buf[off+8:], uint64(rType)|uint64(r.SymName)<<32)
			binary.LittleEndian.PutUint64(buf[off+16:], uint64(r.Addend))
			off += relaEntSize
		}
	}

	// Write section headers
	off = shdrOff

	// Section 0: NULL
	off += shdrSize

	// Section 1: .text
	textNameIdx := strtab.lookup(".text")
	binary.LittleEndian.PutUint32(buf[off:], uint32(textNameIdx))
	binary.LittleEndian.PutUint32(buf[off+4:], SHT_PROGBITS)
	binary.LittleEndian.PutUint64(buf[off+8:], SHF_ALLOC|SHF_EXECINSTR)
	binary.LittleEndian.PutUint64(buf[off+24:], uint64(textOff))
	binary.LittleEndian.PutUint64(buf[off+32:], uint64(textSize))
	binary.LittleEndian.PutUint64(buf[off+48:], 16) // alignment
	off += shdrSize

	// Section 2: .symtab
	symtabNameIdx := strtab.lookup(".symtab")
	binary.LittleEndian.PutUint32(buf[off:], uint32(symtabNameIdx))
	binary.LittleEndian.PutUint32(buf[off+4:], SHT_SYMTAB)
	binary.LittleEndian.PutUint64(buf[off+24:], uint64(symtabOff))
	binary.LittleEndian.PutUint64(buf[off+32:], uint64(symtabSize))
	binary.LittleEndian.PutUint32(buf[off+40:], 3) // sh_link → .strtab
	binary.LittleEndian.PutUint32(buf[off+44:], 1) // sh_info → first global
	binary.LittleEndian.PutUint64(buf[off+48:], 8) // alignment
	binary.LittleEndian.PutUint64(buf[off+56:], uint64(symEntSize))
	off += shdrSize

	// Section 3: .strtab
	strtabNameIdx := strtab.lookup(".strtab")
	binary.LittleEndian.PutUint32(buf[off:], uint32(strtabNameIdx))
	binary.LittleEndian.PutUint32(buf[off+4:], SHT_STRTAB)
	binary.LittleEndian.PutUint64(buf[off+24:], uint64(strtabOff))
	binary.LittleEndian.PutUint64(buf[off+32:], uint64(strtabSize))
	binary.LittleEndian.PutUint64(buf[off+48:], 1) // alignment
	off += shdrSize

	// Section 4: .rela.text (optional)
	if len(w.relocs) > 0 {
		relaNameIdx := strtab.lookup(".rela.text")
		binary.LittleEndian.PutUint32(buf[off:], uint32(relaNameIdx))
		binary.LittleEndian.PutUint32(buf[off+4:], SHT_RELA)
		binary.LittleEndian.PutUint64(buf[off+24:], uint64(relaOff))
		binary.LittleEndian.PutUint64(buf[off+32:], uint64(relaSize))
		binary.LittleEndian.PutUint32(buf[off+40:], 2)  // sh_link → .symtab
		binary.LittleEndian.PutUint32(buf[off+44:], 1)  // sh_info → .text section
		binary.LittleEndian.PutUint64(buf[off+48:], 8)  // alignment
		binary.LittleEndian.PutUint64(buf[off+56:], uint64(relaEntSize))
	}

	return buf
}

// stringTable manages the ELF string table (.strtab).
type stringTable struct {
	data   []byte
	offsets map[string]int
}

func buildStringTable(syms []ELF64Sym) stringTable {
	st := stringTable{
		data:    []byte{0}, // first byte is always NULL
		offsets: make(map[string]int),
	}

	// Add section names
	for _, name := range []string{".text", ".symtab", ".strtab", ".rela.text"} {
		st.add(name)
	}

	// Add symbol names
	for _, sym := range syms {
		st.add(sym.Name)
	}

	return st
}

func (st *stringTable) add(s string) {
	if _, ok := st.offsets[s]; ok {
		return
	}
	st.offsets[s] = len(st.data)
	st.data = append(st.data, []byte(s)...)
	st.data = append(st.data, 0) // null terminator
}

func (st *stringTable) lookup(s string) int {
	if off, ok := st.offsets[s]; ok {
		return off
	}
	return 0
}
