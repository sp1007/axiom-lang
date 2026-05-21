package air

import "sync"

// ---------------------------------------------------------------------------
// AI Hints — metadata for the AI semantic layer.
// ---------------------------------------------------------------------------

// AIHintKind classifies the kind of AI hint attached to an instruction.
type AIHintKind uint8

const (
	AIHintAssertPure AIHintKind = iota // assertion: function is pure
	AIHintSuggestSoA                   // suggestion: convert AoS → SoA layout
	AIHintExplain                      // explanation / rationale annotation
	AIHintSuggestVec                   // suggestion: vectorize this loop
)

// AIHint is a single AI-layer annotation attached to an instruction.
type AIHint struct {
	Kind AIHintKind
	Data string // freeform payload (explanation text, etc.)
}

// AIHintTable stores all AIHint values, indexed by a uint32 handle.
// Handle 0 is reserved to mean "no hints".
type AIHintTable struct {
	entries [][]AIHint // index 0 is unused (sentinel)
}

// NewAIHintTable creates an empty hint table with the zero-sentinel reserved.
func NewAIHintTable() *AIHintTable {
	return &AIHintTable{
		entries: [][]AIHint{nil}, // index 0 = no hints
	}
}

// Add appends a list of hints and returns the table index.
func (t *AIHintTable) Add(hints []AIHint) uint32 {
	idx := uint32(len(t.entries))
	t.entries = append(t.entries, hints)
	return idx
}

// Get retrieves hints by index. Returns nil for index 0 or out-of-range.
func (t *AIHintTable) Get(idx uint32) []AIHint {
	if idx == 0 || int(idx) >= len(t.entries) {
		return nil
	}
	return t.entries[idx]
}

// Len returns the number of hint groups (excluding the zero sentinel).
func (t *AIHintTable) Len() int {
	if len(t.entries) <= 1 {
		return 0
	}
	return len(t.entries) - 1
}

// ---------------------------------------------------------------------------
// AirMeta — per-instruction metadata (source location, ownership, AI hints).
// ---------------------------------------------------------------------------

// AirMeta holds metadata for a single AIR instruction. It is stored
// out-of-band in MetaTable rather than inside AirInst to keep the
// instruction struct frozen at 16 bytes.
type AirMeta struct {
	SourceFile uint32 // interned file path
	SourceLine uint32 // 1-based line number
	SourceCol  uint16 // 1-based column number
	OwnerInfo  uint8  // 0=none, 1=stack, 2=heap, 3=arena
	AIHints    uint32 // index into AIHintTable (0 = no hints)
}

// OwnerInfo constants.
const (
	OwnerNone  uint8 = 0
	OwnerStack uint8 = 1
	OwnerHeap  uint8 = 2
	OwnerArena uint8 = 3
)

// ---------------------------------------------------------------------------
// MetaTable — maps instruction indices to their metadata.
// ---------------------------------------------------------------------------

// MetaTable provides a sparse mapping from instruction index (uint32)
// to AirMeta. Not every instruction needs metadata, so a map is used
// instead of a flat array to save memory for large functions.
//
// MetaTable is safe for concurrent reads after all Set calls are done.
// It is NOT safe for concurrent Set + Get.
type MetaTable struct {
	mu      sync.RWMutex
	entries map[uint32]*AirMeta
}

// NewMetaTable creates an empty MetaTable.
func NewMetaTable() *MetaTable {
	return &MetaTable{
		entries: make(map[uint32]*AirMeta),
	}
}

// Set attaches metadata to an instruction index. Overwrites any
// previous metadata for the same index.
func (m *MetaTable) Set(instIdx uint32, meta AirMeta) {
	m.mu.Lock()
	m.entries[instIdx] = &meta
	m.mu.Unlock()
}

// Get retrieves the metadata for an instruction index, or nil if none.
func (m *MetaTable) Get(instIdx uint32) *AirMeta {
	m.mu.RLock()
	v := m.entries[instIdx]
	m.mu.RUnlock()
	return v
}

// Len returns the number of entries in the table.
func (m *MetaTable) Len() int {
	m.mu.RLock()
	n := len(m.entries)
	m.mu.RUnlock()
	return n
}

// Delete removes metadata for the given instruction index.
func (m *MetaTable) Delete(instIdx uint32) {
	m.mu.Lock()
	delete(m.entries, instIdx)
	m.mu.Unlock()
}
