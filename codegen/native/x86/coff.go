package x86

import (
	"encoding/binary"
)

// --------------------------------------------------------------------------
// p12-t02: PE/COFF Object File Emitter (Windows)
//
// Produces a valid PE-COFF relocatable object file (.obj) for Windows targets.
// --------------------------------------------------------------------------

// COFF constants
const (
	IMAGE_FILE_MACHINE_AMD64 = 0x8664

	// Section characteristics
	IMAGE_SCN_CNT_CODE               = 0x00000020 // Section contains code.
	IMAGE_SCN_CNT_INITIALIZED_DATA   = 0x00000040 // Section contains initialized data.
	IMAGE_SCN_CNT_UNINITIALIZED_DATA = 0x00000080 // Section contains uninitialized data.
	IMAGE_SCN_LNK_INFO               = 0x00000200 // Section contains comments or other information.
	IMAGE_SCN_LNK_REMOVE             = 0x00000800 // Section will not become part of the image.
	IMAGE_SCN_LNK_COMDAT             = 0x00001000 // Section contains COMDAT data.
	IMAGE_SCN_ALIGN_1BYTES           = 0x00100000
	IMAGE_SCN_ALIGN_2BYTES           = 0x00200000
	IMAGE_SCN_ALIGN_4BYTES           = 0x00300000
	IMAGE_SCN_ALIGN_8BYTES           = 0x00400000
	IMAGE_SCN_ALIGN_16BYTES          = 0x00500000
	IMAGE_SCN_ALIGN_32BYTES          = 0x00600000
	IMAGE_SCN_ALIGN_64BYTES          = 0x00700000
	IMAGE_SCN_MEM_DISCARDABLE        = 0x02000000 // Section can be discarded.
	IMAGE_SCN_MEM_EXECUTE            = 0x20000000 // Section can be executed as code.
	IMAGE_SCN_MEM_READ               = 0x40000000 // Section can be read.
	IMAGE_SCN_MEM_WRITE              = 0x80000000 // Section can be written.

	// Storage class constants
	IMAGE_SYM_CLASS_NULL     = 0
	IMAGE_SYM_CLASS_EXTERNAL = 2
	IMAGE_SYM_CLASS_STATIC   = 3

	// Relocation type constants
	IMAGE_REL_AMD64_ADDR64 = 1 // 64-bit absolute address.
	IMAGE_REL_AMD64_REL32  = 4 // 32-bit PC-relative address.
)

type COFFSection struct {
	Name            [8]byte
	VirtSize        uint32
	VirtAddr        uint32 // 0 for .obj
	RawSize         uint32
	RawDataPtr      uint32
	RelocsPtr       uint32
	LineNumsPtr     uint32 // 0
	NumRelocs       uint16
	NumLineNums     uint16 // 0
	Characteristics uint32

	// Internal fields for building the object file
	Data   []byte
	Relocs []COFFReloc
}

type COFFSymbol struct {
	Name         [8]byte // If <= 8 chars. Otherwise, 0 in first 4 bytes, offset in next 4 bytes.
	Value        uint32
	SectionNum   int16 // 1-based index, 0 = undefined external, -1 = absolute
	Type         uint16
	StorageClass uint8
	NumAux       uint8

	NameStr string // string version of name
}

type COFFReloc struct {
	VirtAddr    uint32
	SymTableIdx uint32
	Type        uint16 // IMAGE_REL_AMD64_REL32=4, IMAGE_REL_AMD64_ADDR64=1
}

type COFFWriter struct {
	Sections []COFFSection
	Symbols  []COFFSymbol
	Strings  []byte
}

// NewCOFFWriter creates a new COFF writer.
func NewCOFFWriter() *COFFWriter {
	return &COFFWriter{
		Sections: []COFFSection{},
		Symbols:  []COFFSymbol{},
		Strings:  []byte{0, 0, 0, 0}, // Start string table with 4-byte size field
	}
}

// AddSection adds a section to the object file.
// Returns the 1-based section index.
func (w *COFFWriter) AddSection(name string, flags uint32, data []byte) int {
	var sName [8]byte
	copy(sName[:], name)

	section := COFFSection{
		Name:            sName,
		Characteristics: flags,
		Data:            data,
		RawSize:         uint32(len(data)),
	}
	w.Sections = append(w.Sections, section)
	return len(w.Sections) // 1-based index for COFF
}

// AddSymbol adds a symbol to the symbol table.
// Returns the 0-based symbol index.
func (w *COFFWriter) AddSymbol(name string, sectionIdx int, offset uint32, external bool) int {
	sym := COFFSymbol{
		Value:      offset,
		SectionNum: int16(sectionIdx),
		Type:       0x20, // IMAGE_SYM_DTYPE_FUNCTION
		NameStr:    name,
	}
	if external {
		sym.StorageClass = IMAGE_SYM_CLASS_EXTERNAL
	} else {
		sym.StorageClass = IMAGE_SYM_CLASS_STATIC
	}

	w.Symbols = append(w.Symbols, sym)
	return len(w.Symbols) - 1 // 0-based index
}

// AddReloc adds a relocation to a section.
func (w *COFFWriter) AddReloc(sectionIdx, offset, symIdx int, relocType uint16) {
	if sectionIdx > 0 && sectionIdx <= len(w.Sections) {
		reloc := COFFReloc{
			VirtAddr:    uint32(offset),
			SymTableIdx: uint32(symIdx),
			Type:        relocType,
		}
		w.Sections[sectionIdx-1].Relocs = append(w.Sections[sectionIdx-1].Relocs, reloc)
		w.Sections[sectionIdx-1].NumRelocs++
	}
}

// Serialize produces the COFF object file bytes.
func (w *COFFWriter) Serialize() []byte {
	// Build the string table and map offsets for all symbols
	strtab := make([]byte, 4)
	strOffsets := make(map[string]uint32)
	for _, sym := range w.Symbols {
		if len(sym.NameStr) > 8 {
			strOffsets[sym.NameStr] = uint32(len(strtab))
			strtab = append(strtab, []byte(sym.NameStr)...)
			strtab = append(strtab, 0)
		}
	}
	binary.LittleEndian.PutUint32(strtab[0:4], uint32(len(strtab)))

	numSections := len(w.Sections)
	headerSize := 20
	sectionHeaderSize := 40

	currentOffset := headerSize + numSections*sectionHeaderSize

	// 1. Assign raw data pointers
	for i := range w.Sections {
		if len(w.Sections[i].Data) > 0 {
			w.Sections[i].RawDataPtr = uint32(currentOffset)
			currentOffset += len(w.Sections[i].Data)
		} else {
			w.Sections[i].RawDataPtr = 0
		}
	}

	// 2. Assign relocation table pointers
	for i := range w.Sections {
		if len(w.Sections[i].Relocs) > 0 {
			w.Sections[i].RelocsPtr = uint32(currentOffset)
			currentOffset += len(w.Sections[i].Relocs) * 10
		} else {
			w.Sections[i].RelocsPtr = 0
		}
	}

	symtabOff := currentOffset
	symCount := len(w.Symbols)
	strtabOff := symtabOff + symCount*18

	totalSize := strtabOff + len(strtab)
	buf := make([]byte, totalSize)

	// Write COFF Header
	binary.LittleEndian.PutUint16(buf[0:2], IMAGE_FILE_MACHINE_AMD64)
	binary.LittleEndian.PutUint16(buf[2:4], uint16(numSections))
	binary.LittleEndian.PutUint32(buf[4:8], 0)
	binary.LittleEndian.PutUint32(buf[8:12], uint32(symtabOff))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(symCount))
	binary.LittleEndian.PutUint16(buf[16:18], 0)
	binary.LittleEndian.PutUint16(buf[18:20], 0)

	// Write Section Headers
	for i, sec := range w.Sections {
		off := headerSize + i*sectionHeaderSize
		copy(buf[off:off+8], sec.Name[:])
		binary.LittleEndian.PutUint32(buf[off+8:off+12], sec.VirtSize)
		binary.LittleEndian.PutUint32(buf[off+12:off+16], sec.VirtAddr)
		binary.LittleEndian.PutUint32(buf[off+16:off+20], sec.RawSize)
		binary.LittleEndian.PutUint32(buf[off+20:off+24], sec.RawDataPtr)
		binary.LittleEndian.PutUint32(buf[off+24:off+28], sec.RelocsPtr)
		binary.LittleEndian.PutUint32(buf[off+28:off+32], sec.LineNumsPtr)
		binary.LittleEndian.PutUint16(buf[off+32:off+34], sec.NumRelocs)
		binary.LittleEndian.PutUint16(buf[off+34:off+36], sec.NumLineNums)
		binary.LittleEndian.PutUint32(buf[off+36:off+40], sec.Characteristics)
	}

	// Write Section Data
	for _, sec := range w.Sections {
		if sec.RawDataPtr > 0 {
			copy(buf[sec.RawDataPtr:], sec.Data)
		}
	}

	// Write Relocation Tables
	for _, sec := range w.Sections {
		if sec.RelocsPtr > 0 {
			relOff := sec.RelocsPtr
			for _, r := range sec.Relocs {
				binary.LittleEndian.PutUint32(buf[relOff:relOff+4], r.VirtAddr)
				binary.LittleEndian.PutUint32(buf[relOff+4:relOff+8], r.SymTableIdx)
				binary.LittleEndian.PutUint16(buf[relOff+8:relOff+10], r.Type)
				relOff += 10
			}
		}
	}

	// Write Symbol Table
	symOff := symtabOff
	for _, sym := range w.Symbols {
		symBuf := w.serializeSymbol(&sym, strOffsets)
		copy(buf[symOff:], symBuf)
		symOff += 18
	}

	// Write String Table
	copy(buf[strtabOff:], strtab)

	return buf
}

func (w *COFFWriter) serializeSymbol(sym *COFFSymbol, stringTableOffset map[string]uint32) []byte {
	buf := make([]byte, 18)
	nameBytes := []byte(sym.NameStr)
	if len(nameBytes) <= 8 {
		copy(buf[0:8], nameBytes)
	} else {
		offset := stringTableOffset[sym.NameStr]
		binary.LittleEndian.PutUint32(buf[0:4], 0)
		binary.LittleEndian.PutUint32(buf[4:8], offset)
	}
	binary.LittleEndian.PutUint32(buf[8:12], sym.Value)
	binary.LittleEndian.PutUint16(buf[12:14], uint16(sym.SectionNum))
	binary.LittleEndian.PutUint16(buf[14:16], sym.Type)
	buf[16] = sym.StorageClass
	buf[17] = sym.NumAux
	return buf
}
