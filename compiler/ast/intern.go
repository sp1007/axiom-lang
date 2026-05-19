package ast

import (
	"bytes"
	"fmt"
)

// InternPool interns []byte slices, returning a stable uint32 ID.
// Each unique byte slice is stored exactly once in the arena.
// IDs start at 1; ID 0 is reserved as "no string" / empty.
type InternPool struct {
	arena []byte        // backing storage for all interned strings
	table []internEntry // open-addressing hash table (power-of-2 size)
	ids   []internEntry // indexed by id-1; for O(1) reverse lookup
	count int           // number of unique strings stored
}

type internEntry struct {
	hash   uint32 // FNV-1a hash
	start  uint32 // byte offset in arena
	length uint16 // byte length
	id     uint32 // assigned intern ID (1-based)
}

// NewInternPool creates an intern pool with the given initial capacity hint.
func NewInternPool(initialCap int) *InternPool {
	// Table size must be a power of 2
	size := 64
	for size < initialCap*2 {
		size <<= 1
	}
	return &InternPool{
		arena: make([]byte, 0, initialCap*8),
		table: make([]internEntry, size),
		ids:   make([]internEntry, 0, initialCap),
	}
}

// WellKnownIDs contains pre-interned IDs for common strings.
type WellKnownIDs struct {
	Main    uint32
	Init    uint32
	String_ uint32 // "string"
	Bool    uint32
	I32     uint32
	I64     uint32
	F64     uint32
}

// NewInternPoolWithWellKnown creates an intern pool and pre-interns common strings.
func NewInternPoolWithWellKnown(initialCap int) (*InternPool, WellKnownIDs) {
	p := NewInternPool(initialCap)
	return p, WellKnownIDs{
		Main:    p.InternString("main"),
		Init:    p.InternString("init"),
		String_: p.InternString("string"),
		Bool:    p.InternString("bool"),
		I32:     p.InternString("i32"),
		I64:     p.InternString("i64"),
		F64:     p.InternString("f64"),
	}
}

// Intern returns the ID for s, interning it if not already present.
// s may be a slice into the source buffer; the pool copies the bytes.
// Returns 0 for empty s (empty strings are not interned).
func (p *InternPool) Intern(s []byte) uint32 {
	if len(s) == 0 {
		return 0
	}
	if len(s) > 65535 {
		panic(fmt.Sprintf("InternPool: string too long: %d bytes", len(s)))
	}

	h := fnv1a(s)
	mask := uint32(len(p.table) - 1)
	idx := h & mask

	for {
		e := &p.table[idx]
		if e.id == 0 {
			// Empty slot: insert new entry
			return p.insert(idx, s, h)
		}
		// Check if existing entry matches
		if e.hash == h && e.length == uint16(len(s)) {
			stored := p.arena[e.start : e.start+uint32(e.length)]
			if bytes.Equal(stored, s) {
				return e.id
			}
		}
		idx = (idx + 1) & mask // linear probe
	}
}

// InternString is like Intern but accepts a Go string.
func (p *InternPool) InternString(s string) uint32 {
	return p.Intern([]byte(s))
}

// Get returns the string for the given ID.
// Panics if id is 0 or out of range (programming error).
func (p *InternPool) Get(id uint32) string {
	return string(p.GetBytes(id))
}

// GetBytes returns the bytes for the given ID without allocating a new string.
// The returned slice references the arena — do not modify it.
func (p *InternPool) GetBytes(id uint32) []byte {
	if id == 0 {
		panic("InternPool.GetBytes: id 0 is reserved")
	}
	if int(id) > len(p.ids) {
		panic(fmt.Sprintf("InternPool.GetBytes: id %d out of range (max=%d)", id, len(p.ids)))
	}
	e := p.ids[id-1]
	return p.arena[e.start : e.start+uint32(e.length)]
}

// Len returns the number of unique strings interned.
func (p *InternPool) Len() int {
	return p.count
}

// insert writes a new entry into the hash table and returns the new ID.
func (p *InternPool) insert(slot uint32, s []byte, h uint32) uint32 {
	// Check load factor before inserting
	if float64(p.count+1)/float64(len(p.table)) > 0.7 {
		p.grow()
		return p.Intern(s) // re-intern after resize
	}

	// Store bytes in arena
	start := uint32(len(p.arena))
	p.arena = append(p.arena, s...)

	// Assign ID (1-based)
	id := uint32(p.count + 1)
	p.count++

	entry := internEntry{hash: h, start: start, length: uint16(len(s)), id: id}
	p.table[slot] = entry
	p.ids = append(p.ids, entry)

	return id
}

// grow doubles the table size and rehashes all entries.
func (p *InternPool) grow() {
	newSize := len(p.table) * 2
	newTable := make([]internEntry, newSize)
	mask := uint32(newSize - 1)
	for _, e := range p.table {
		if e.id == 0 {
			continue
		}
		idx := e.hash & mask
		for newTable[idx].id != 0 {
			idx = (idx + 1) & mask
		}
		newTable[idx] = e
	}
	p.table = newTable
}

// fnv1a computes the FNV-1a hash of s.
func fnv1a(s []byte) uint32 {
	h := uint32(2166136261) // FNV offset basis
	for _, b := range s {
		h ^= uint32(b)
		h *= 16777619 // FNV prime
	}
	return h
}
