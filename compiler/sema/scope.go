package sema

// ScopeKind identifies what language construct created this scope.
type ScopeKind uint8

const (
	ScopeGlobal   ScopeKind = iota // module-level scope
	ScopeFunction                  // function body
	ScopeBlock                     // if/elif/else/for/match/lock/arena block
	ScopeClosure                   // closure body
	ScopeLoop                      // for/in loop (owns loop variable)
)

// scopeEntry is a slot in the open-addressing hash map.
type scopeEntry struct {
	nameID    uint32 // 0 = empty slot
	symbolIdx uint32 // index into SymbolTable.Symbols
}

// Scope is a single lexical scope with O(1) name lookup.
// Implemented as an open-addressing hash table with linear probing.
type Scope struct {
	Kind     ScopeKind
	ParentID uint32 // index of parent scope (0 for global)
	Depth    uint32 // nesting depth (0 = global)

	entries  []scopeEntry // open-addressing hash table
	count    uint32       // number of occupied slots
	capacity uint32       // length of entries slice (always power of 2)
}

// init initializes the scope with a given power-of-2 capacity.
func (s *Scope) init(capacity uint32) {
	s.capacity = capacity
	s.entries = make([]scopeEntry, capacity)
	s.count = 0
}

// hashFNV1a computes a simple FNV-1a hash of a uint32.
func hashFNV1a(v uint32) uint32 {
	hash := uint32(2166136261)
	hash ^= v & 0xFF
	hash *= 16777619
	hash ^= (v >> 8) & 0xFF
	hash *= 16777619
	hash ^= (v >> 16) & 0xFF
	hash *= 16777619
	hash ^= (v >> 24) & 0xFF
	hash *= 16777619
	return hash
}

// put inserts a new nameID -> symbolIdx mapping.
// Assumes nameID is not already in the scope (caller must check).
func (s *Scope) put(nameID uint32, symbolIdx uint32) {
	// Grow if load factor > 75%
	if s.count*4 > s.capacity*3 {
		s.grow()
	}

	s.insert(nameID, symbolIdx)
	s.count++
}

// insert performs the actual open-addressing insertion without updating count.
func (s *Scope) insert(nameID uint32, symbolIdx uint32) {
	mask := s.capacity - 1
	idx := hashFNV1a(nameID) & mask

	for {
		if s.entries[idx].nameID == 0 {
			s.entries[idx].nameID = nameID
			s.entries[idx].symbolIdx = symbolIdx
			return
		}
		// Move to next slot (linear probing)
		idx = (idx + 1) & mask
	}
}

// get looks up a nameID in the scope.
func (s *Scope) get(nameID uint32) (uint32, bool) {
	if s.capacity == 0 {
		return 0, false
	}

	mask := s.capacity - 1
	idx := hashFNV1a(nameID) & mask
	startIdx := idx

	for {
		entry := s.entries[idx]
		if entry.nameID == 0 {
			return 0, false // empty slot means not found
		}
		if entry.nameID == nameID {
			return entry.symbolIdx, true
		}
		idx = (idx + 1) & mask
		if idx == startIdx {
			break // table is completely full (shouldn't happen due to load factor limit)
		}
	}
	return 0, false
}

// grow doubles the capacity and rehashes all entries.
func (s *Scope) grow() {
	oldEntries := s.entries
	
	newCap := s.capacity * 2
	if newCap == 0 {
		newCap = 8 // fallback if initially 0
	}
	
	s.capacity = newCap
	s.entries = make([]scopeEntry, newCap)

	for _, entry := range oldEntries {
		if entry.nameID != 0 {
			s.insert(entry.nameID, entry.symbolIdx)
		}
	}
}
