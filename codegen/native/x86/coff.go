package x86

import (
	"encoding/binary"
)

// --------------------------------------------------------------------------
// p12-t02: PE/COFF Object File Emitter (Windows)
//
// Produces a COFF object file (.obj) compatible with the Windows PE format.
// Used when the target OS is Windows.
// --------------------------------------------------------------------------

// COFF constants
const (
	IMAGE_FILE_MACHINE_AMD64 = 0x8664
	IMAGE_SCN_CNT_CODE       = 0x00000020
	IMAGE_SCN_MEM_EXECUTE    = 0x20000000
	IMAGE_SCN_MEM_READ       = 0x40000000
	IMAGE_SYM_CLASS_EXTERNAL = 2
	IMAGE_SYM_DTYPE_FUNCTION = 0x20
)

// COFFWriter builds a COFF object file.
type COFFWriter struct {
	textData []byte
	syms     []ELF64Sym // reuse symbol struct
}

// NewCOFFWriter creates a new COFF writer.
func NewCOFFWriter() *COFFWriter {
	return &COFFWriter{}
}

// SetText sets the .text section contents.
func (w *COFFWriter) SetText(code []byte) {
	w.textData = code
}

// AddSymbol adds a symbol to the symbol table.
func (w *COFFWriter) AddSymbol(sym ELF64Sym) {
	w.syms = append(w.syms, sym)
}

// Serialize produces the COFF object file bytes.
func (w *COFFWriter) Serialize() []byte {
	// COFF header: 20 bytes
	// Section header: 40 bytes (.text)
	// Section data: textData
	// Symbol table: 18 bytes per entry
	// String table: 4 bytes (size) + names

	headerSize := 20
	sectionHeaderSize := 40
	numSections := 1

	textOff := headerSize + sectionHeaderSize*numSections
	textSize := len(w.textData)

	symtabOff := textOff + textSize
	symEntSize := 18
	symCount := len(w.syms)
	symtabSize := symCount * symEntSize

	// String table (for long symbol names)
	strtab := buildCOFFStringTable(w.syms)
	strtabOff := symtabOff + symtabSize

	totalSize := strtabOff + len(strtab)
	buf := make([]byte, totalSize)

	// COFF Header
	binary.LittleEndian.PutUint16(buf[0:], IMAGE_FILE_MACHINE_AMD64)
	binary.LittleEndian.PutUint16(buf[2:], uint16(numSections))
	binary.LittleEndian.PutUint32(buf[8:], uint32(symtabOff))
	binary.LittleEndian.PutUint32(buf[12:], uint32(symCount))
	// buf[16:18] = 0 (optional header size)
	// buf[18:20] = 0 (characteristics)

	// Section header (.text)
	off := headerSize
	copy(buf[off:off+8], ".text\x00\x00\x00")
	binary.LittleEndian.PutUint32(buf[off+16:], uint32(textSize))
	binary.LittleEndian.PutUint32(buf[off+20:], uint32(textOff))
	binary.LittleEndian.PutUint32(buf[off+36:], IMAGE_SCN_CNT_CODE|IMAGE_SCN_MEM_EXECUTE|IMAGE_SCN_MEM_READ)

	// Section data
	copy(buf[textOff:], w.textData)

	// Symbol table
	off = symtabOff
	for _, sym := range w.syms {
		nameBytes := []byte(sym.Name)
		if len(nameBytes) <= 8 {
			copy(buf[off:off+8], nameBytes)
		} else {
			// Long name: store offset in string table
			binary.LittleEndian.PutUint32(buf[off:], 0)
			binary.LittleEndian.PutUint32(buf[off+4:], uint32(findCOFFString(strtab, sym.Name)))
		}
		binary.LittleEndian.PutUint32(buf[off+8:], uint32(sym.Value))
		binary.LittleEndian.PutUint16(buf[off+12:], uint16(sym.Section))
		binary.LittleEndian.PutUint16(buf[off+14:], IMAGE_SYM_DTYPE_FUNCTION)
		buf[off+16] = IMAGE_SYM_CLASS_EXTERNAL
		off += symEntSize
	}

	// String table
	copy(buf[strtabOff:], strtab)

	return buf
}

func buildCOFFStringTable(syms []ELF64Sym) []byte {
	// String table starts with 4-byte size field
	buf := make([]byte, 4)
	for _, sym := range syms {
		if len(sym.Name) > 8 {
			buf = append(buf, []byte(sym.Name)...)
			buf = append(buf, 0)
		}
	}
	binary.LittleEndian.PutUint32(buf[0:], uint32(len(buf)))
	return buf
}

func findCOFFString(strtab []byte, name string) int {
	nameBytes := []byte(name)
	for i := 4; i < len(strtab); i++ {
		match := true
		for j := 0; j < len(nameBytes) && i+j < len(strtab); j++ {
			if strtab[i+j] != nameBytes[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return 4
}
