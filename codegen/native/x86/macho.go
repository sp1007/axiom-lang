package x86

import (
	"encoding/binary"
)

// --------------------------------------------------------------------------
// p12-t03: Mach-O Object File Emitter (macOS)
//
// Produces a Mach-O object file (.o) for macOS/Darwin targets.
// Used when the target OS is macOS.
// --------------------------------------------------------------------------

// Mach-O constants
const (
	MH_MAGIC_64         = 0xFEEDFACF
	MH_OBJECT           = 0x1
	CPU_TYPE_X86_64     = 0x01000007
	CPU_SUBTYPE_ALL     = 3
	LC_SEGMENT_64       = 0x19
	LC_SYMTAB           = 0x02
	S_REGULAR           = 0x0
	S_ATTR_PURE_INSTRUCTIONS = 0x80000000
	S_ATTR_SOME_INSTRUCTIONS = 0x00000400
	N_SECT              = 0x0E
	N_EXT               = 0x01
)

// MachOWriter builds a Mach-O 64-bit object file.
type MachOWriter struct {
	textData []byte
	syms     []ELF64Sym
}

// NewMachOWriter creates a new Mach-O writer.
func NewMachOWriter() *MachOWriter {
	return &MachOWriter{}
}

// SetText sets the .text section contents.
func (w *MachOWriter) SetText(code []byte) {
	w.textData = code
}

// AddSymbol adds a symbol.
func (w *MachOWriter) AddSymbol(sym ELF64Sym) {
	w.syms = append(w.syms, sym)
}

// Serialize produces the Mach-O object file bytes.
func (w *MachOWriter) Serialize() []byte {
	// Mach-O header: 32 bytes
	// LC_SEGMENT_64: 72 bytes + section64: 80 bytes
	// LC_SYMTAB: 24 bytes
	// Text data
	// Symbol table (nlist_64: 16 bytes each)
	// String table

	headerSize := 32
	segCmdSize := 72 + 80 // segment + one section
	symtabCmdSize := 24
	loadCmdsSize := segCmdSize + symtabCmdSize

	textOff := headerSize + loadCmdsSize
	textSize := len(w.textData)

	// Align to 8 bytes
	symtabOff := textOff + textSize
	if symtabOff%8 != 0 {
		symtabOff += 8 - symtabOff%8
	}

	nlistSize := 16
	symCount := len(w.syms)
	symtabSize := symCount * nlistSize

	strtabOff := symtabOff + symtabSize
	strtab := buildMachOStringTable(w.syms)
	strtabSize := len(strtab)

	totalSize := strtabOff + strtabSize
	buf := make([]byte, totalSize)

	// Mach-O Header
	binary.LittleEndian.PutUint32(buf[0:], MH_MAGIC_64)
	binary.LittleEndian.PutUint32(buf[4:], CPU_TYPE_X86_64)
	binary.LittleEndian.PutUint32(buf[8:], CPU_SUBTYPE_ALL)
	binary.LittleEndian.PutUint32(buf[12:], MH_OBJECT)
	binary.LittleEndian.PutUint32(buf[16:], 2) // ncmds
	binary.LittleEndian.PutUint32(buf[20:], uint32(loadCmdsSize))
	// buf[24:28] = flags = 0
	// buf[28:32] = reserved = 0

	// LC_SEGMENT_64
	off := headerSize
	binary.LittleEndian.PutUint32(buf[off:], LC_SEGMENT_64)
	binary.LittleEndian.PutUint32(buf[off+4:], uint32(segCmdSize))
	// segname = "" (16 zero bytes)
	binary.LittleEndian.PutUint64(buf[off+24:], 0) // vmaddr
	binary.LittleEndian.PutUint64(buf[off+32:], uint64(textSize)) // vmsize
	binary.LittleEndian.PutUint64(buf[off+40:], uint64(textOff)) // fileoff
	binary.LittleEndian.PutUint64(buf[off+48:], uint64(textSize)) // filesize
	binary.LittleEndian.PutUint32(buf[off+64:], 1) // nsects
	off += 72

	// section64 (__text, __TEXT)
	copy(buf[off:off+16], "__text\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	copy(buf[off+16:off+32], "__TEXT\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	binary.LittleEndian.PutUint64(buf[off+32:], 0)  // addr
	binary.LittleEndian.PutUint64(buf[off+40:], uint64(textSize))
	binary.LittleEndian.PutUint32(buf[off+48:], uint32(textOff))
	binary.LittleEndian.PutUint32(buf[off+52:], 4) // align = 2^4 = 16
	binary.LittleEndian.PutUint32(buf[off+60:], S_ATTR_PURE_INSTRUCTIONS|S_ATTR_SOME_INSTRUCTIONS)
	off += 80

	// LC_SYMTAB
	binary.LittleEndian.PutUint32(buf[off:], LC_SYMTAB)
	binary.LittleEndian.PutUint32(buf[off+4:], uint32(symtabCmdSize))
	binary.LittleEndian.PutUint32(buf[off+8:], uint32(symtabOff))
	binary.LittleEndian.PutUint32(buf[off+12:], uint32(symCount))
	binary.LittleEndian.PutUint32(buf[off+16:], uint32(strtabOff))
	binary.LittleEndian.PutUint32(buf[off+20:], uint32(strtabSize))

	// Text data
	copy(buf[textOff:], w.textData)

	// Symbol table (nlist_64)
	off = symtabOff
	for _, sym := range w.syms {
		nameOff := findMachOString(strtab, sym.Name)
		binary.LittleEndian.PutUint32(buf[off:], uint32(nameOff))
		buf[off+4] = N_SECT | N_EXT  // n_type
		buf[off+5] = 1                // n_sect (1-based, __text)
		binary.LittleEndian.PutUint64(buf[off+8:], sym.Value) // n_value
		off += nlistSize
	}

	// String table
	copy(buf[strtabOff:], strtab)

	return buf
}

func buildMachOStringTable(syms []ELF64Sym) []byte {
	buf := []byte{0} // first byte null
	for _, sym := range syms {
		// Mach-O prepends underscore
		buf = append(buf, '_')
		buf = append(buf, []byte(sym.Name)...)
		buf = append(buf, 0)
	}
	return buf
}

func findMachOString(strtab []byte, name string) int {
	target := "_" + name
	for i := 1; i < len(strtab); i++ {
		if i+len(target) <= len(strtab) {
			match := true
			for j := 0; j < len(target); j++ {
				if strtab[i+j] != target[j] {
					match = false
					break
				}
			}
			if match {
				return i
			}
		}
	}
	return 0
}
