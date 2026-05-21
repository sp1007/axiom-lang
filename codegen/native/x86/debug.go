package x86

// --------------------------------------------------------------------------
// p11-t13: DWARF Line Info (Stub)
//
// Generates .debug_line section data following DWARF4 format for
// source-level debugging. Maps machine code offsets to source file
// locations (file, line, column).
// --------------------------------------------------------------------------

// LineEntry records a mapping from machine code offset to source location.
type LineEntry struct {
	PC     uint64
	File   uint16
	Line   uint32
	Column uint16
}

// DWARFLineWriter builds the .debug_line section.
type DWARFLineWriter struct {
	Files   []string    // source file names
	Entries []LineEntry // PC → source mappings
}

// NewDWARFLineWriter creates a new DWARF line info writer.
func NewDWARFLineWriter() *DWARFLineWriter {
	return &DWARFLineWriter{}
}

// AddFile registers a source file and returns its index.
func (w *DWARFLineWriter) AddFile(filename string) uint16 {
	w.Files = append(w.Files, filename)
	return uint16(len(w.Files))
}

// EmitEntry records a PC → source location mapping.
func (w *DWARFLineWriter) EmitEntry(pc uint64, file uint16, line uint32, col uint16) {
	w.Entries = append(w.Entries, LineEntry{
		PC: pc, File: file, Line: line, Column: col,
	})
}

// Serialize produces the .debug_line section bytes.
// TODO: Full DWARF4 encoding with line number program opcodes.
func (w *DWARFLineWriter) Serialize() []byte {
	// Stub: return empty section for now
	// Full implementation requires DWARF state machine encoding
	return nil
}

// --------------------------------------------------------------------------
// p11-t14: .axmeta Section Writer (Stub)
//
// Writes a custom .axmeta section containing AXIOM-specific metadata:
// type information, AI semantic hints, and optimization annotations.
// --------------------------------------------------------------------------

// AxMetaSection contains AXIOM metadata for an object file.
type AxMetaSection struct {
	ModuleName string
	Version    uint32
	TypeCount  uint32
	FuncCount  uint32
}

// SerializeAxMeta produces the .axmeta section bytes.
// Format: "AXMT" magic + version + counts + JSON payload.
func SerializeAxMeta(meta *AxMetaSection) []byte {
	if meta == nil {
		return nil
	}

	buf := make([]byte, 16)
	copy(buf[0:4], "AXMT") // magic
	buf[4] = byte(meta.Version)
	buf[5] = byte(meta.Version >> 8)
	buf[6] = byte(meta.Version >> 16)
	buf[7] = byte(meta.Version >> 24)
	buf[8] = byte(meta.TypeCount)
	buf[9] = byte(meta.TypeCount >> 8)
	buf[10] = byte(meta.TypeCount >> 16)
	buf[11] = byte(meta.TypeCount >> 24)
	buf[12] = byte(meta.FuncCount)
	buf[13] = byte(meta.FuncCount >> 8)
	buf[14] = byte(meta.FuncCount >> 16)
	buf[15] = byte(meta.FuncCount >> 24)

	return buf
}
