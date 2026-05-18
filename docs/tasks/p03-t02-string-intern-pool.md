# p03-t02: String Intern Pool

## Purpose
Implement a string interning pool in `compiler/ast/intern.go` that deduplicates identifier strings across the entire compilation unit. Rather than storing identifier text as `string` values (which copies bytes and increases GC pressure), the intern pool stores each unique string once in a backing byte arena and returns a `uint32` ID. All subsequent references to the same string use the same ID. This enables O(1) identifier equality comparisons (compare IDs), reduces memory usage, and is essential for the symbol table (p04-t01) which keys on intern IDs.

## Context
In a typical source file, the same identifier (e.g., `i`, `x`, `result`, `main`) appears dozens or hundreds of times. Without interning, each occurrence would require allocating a new `string` or holding a reference into the original source buffer. With interning, each unique string is stored once, and all occurrences share the same `uint32` ID. The intern pool is populated by the name resolution pass (p04-t04) as it walks the AST. The pool uses open-addressing hash map with FNV-1a hashing and a flat `[]byte` backing arena. The pool is per-compilation-unit (not global) to allow parallel compilation of multiple units.

## Inputs
- `compiler/ast/tree.go` from p03-t01 — `AstTree` which holds `Source []byte`
- No external dependencies beyond Go standard library

## Outputs
- `compiler/ast/intern.go` — `InternPool` struct with `Intern()` and `Get()` methods
- `compiler/ast/intern_test.go` — unit tests

## Dependencies
- p03-t01: ast-node-definitions — the `compiler/ast` package context

## Subsystems Affected
- `compiler/ast/`: InternPool lives here
- `compiler/sema/`: Symbol table (p04-t01) uses intern IDs as keys
- `compiler/parser/`: Parser interns identifier tokens during parsing

## Detailed Requirements

1. **`InternPool` struct**:
   ```go
   // InternPool interns []byte slices, returning a stable uint32 ID.
   // Each unique byte slice is stored exactly once in the arena.
   // IDs start at 1; ID 0 is reserved as "no string" / empty.
   type InternPool struct {
       arena   []byte            // backing storage for all strings
       table   []internEntry     // open-addressing hash table
       count   int               // number of unique strings stored
   }

   type internEntry struct {
       hash   uint32 // FNV-1a hash of the string
       start  uint32 // byte offset in arena
       length uint16 // byte length
       id     uint32 // the assigned intern ID (1-based)
   }
   ```

2. **API**:
   ```go
   // NewInternPool creates an intern pool with the given initial capacity.
   func NewInternPool(initialCap int) *InternPool

   // Intern returns the ID for s, interning it if not already present.
   // s may be a slice into the source buffer; the pool copies the bytes.
   // Returns 0 only for empty s (empty strings are not interned).
   func (p *InternPool) Intern(s []byte) uint32

   // InternString is like Intern but accepts a Go string.
   func (p *InternPool) InternString(s string) uint32

   // Get returns the string for the given ID.
   // Panics if id is 0 or out of range (programming error).
   func (p *InternPool) Get(id uint32) string

   // GetBytes returns the bytes for the given ID without allocating.
   func (p *InternPool) GetBytes(id uint32) []byte

   // Len returns the number of unique strings interned.
   func (p *InternPool) Len() int
   ```

3. **FNV-1a hash**:
   ```go
   func fnv1a(s []byte) uint32 {
       h := uint32(2166136261) // FNV offset basis
       for _, b := range s {
           h ^= uint32(b)
           h *= 16777619 // FNV prime
       }
       return h
   }
   ```

4. **Open-addressing hash table** — linear probing:
   ```go
   func (p *InternPool) Intern(s []byte) uint32 {
       if len(s) == 0 { return 0 }

       h := fnv1a(s)
       mask := uint32(len(p.table) - 1)
       idx := h & mask

       for {
           e := &p.table[idx]
           if e.id == 0 {
               // empty slot: insert new entry
               return p.insert(idx, s, h)
           }
           // Check if existing entry matches
           stored := p.arena[e.start : e.start+uint32(e.length)]
           if e.hash == h && bytes.Equal(stored, s) {
               return e.id
           }
           idx = (idx + 1) & mask // linear probe
       }
   }
   ```

5. **`insert()` helper — grows table if load factor > 0.7**:
   ```go
   func (p *InternPool) insert(slot uint32, s []byte, h uint32) uint32 {
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
       // Write entry
       p.table[slot] = internEntry{hash: h, start: start, length: uint16(len(s)), id: id}
       return id
   }
   ```

6. **`grow()` — double the table size and rehash**:
   ```go
   func (p *InternPool) grow() {
       newSize := len(p.table) * 2
       newTable := make([]internEntry, newSize)
       mask := uint32(newSize - 1)
       for _, e := range p.table {
           if e.id == 0 { continue }
           idx := e.hash & mask
           for newTable[idx].id != 0 {
               idx = (idx + 1) & mask
           }
           newTable[idx] = e
       }
       p.table = newTable
   }
   ```

7. **`NewInternPool(initialCap int) *InternPool`**:
   ```go
   func NewInternPool(initialCap int) *InternPool {
       // table size must be a power of 2
       size := 64
       for size < initialCap*2 {
           size <<= 1
       }
       return &InternPool{
           arena: make([]byte, 0, initialCap*8),
           table: make([]internEntry, size),
       }
   }
   ```

8. **`Get(id uint32) string`** — linear scan to find entry with matching ID:
   ```go
   func (p *InternPool) Get(id uint32) string {
       if id == 0 { panic("InternPool.Get: id 0 is reserved") }
       return string(p.GetBytes(id))
   }

   func (p *InternPool) GetBytes(id uint32) []byte {
       if id == 0 { panic("InternPool.GetBytes: id 0 is reserved") }
       // Entries are assigned IDs in insertion order (1-based).
       // Build a reverse index on first use, or use a separate []uint32 of arena offsets.
       // Simpler: maintain a separate slice indexed by id.
       // See implementation note below.
   }
   ```
   Implementation note: To make `Get()` O(1), maintain a separate `ids []internEntry` slice (indexed by `id-1`) alongside the hash table:
   ```go
   type InternPool struct {
       arena   []byte
       table   []internEntry
       ids     []internEntry // indexed by id-1; for O(1) reverse lookup
       count   int
   }
   ```
   In `insert()`, append the new entry to `p.ids` at position `p.count` (which equals `id-1`).

9. **`internEntry` length constraint**: `uint16` for length means strings up to 65535 bytes. Identifiers and keywords are always short (< 1000 bytes). Add a check in `Intern()`:
   ```go
   if len(s) > 65535 {
       panic(fmt.Sprintf("InternPool: string too long: %d bytes", len(s)))
   }
   ```

10. **Well-known intern IDs**: Pre-intern common keywords at construction time so their IDs are stable:
    ```go
    // WellKnownIDs contains pre-interned IDs for common strings.
    // These IDs are stable across compilation units with the same pool.
    type WellKnownIDs struct {
        Main      uint32
        Init      uint32
        String_   uint32 // "string" (builtin type)
        Bool      uint32
        // ... etc
    }

    func NewInternPoolWithWellKnown(initialCap int) (*InternPool, WellKnownIDs) {
        p := NewInternPool(initialCap)
        return p, WellKnownIDs{
            Main:   p.InternString("main"),
            Init:   p.InternString("init"),
            String_: p.InternString("string"),
            Bool:   p.InternString("bool"),
        }
    }
    ```

## Implementation Steps

1. Create `compiler/ast/intern.go` with `InternPool`, `internEntry`, and `WellKnownIDs` structs.

2. Implement `fnv1a()` hash function.

3. Implement `NewInternPool(initialCap int) *InternPool` and `NewInternPoolWithWellKnown()`.

4. Implement `Intern(s []byte) uint32` with open-addressing linear probe.

5. Implement `insert()` with load factor check and `grow()` for resizing.

6. Implement `Get(id uint32) string` and `GetBytes(id uint32) []byte` using the `ids` slice for O(1) lookup.

7. Implement `InternString(s string) uint32` as a thin wrapper over `Intern`.

8. Implement `Len() int`.

9. Create `compiler/ast/intern_test.go` with all tests below.

10. Run `go test ./compiler/ast/` — all tests pass.

## Test Plan

Write `compiler/ast/intern_test.go`:

```go
func TestInternBasic(t *testing.T) {
    p := NewInternPool(16)
    id1 := p.Intern([]byte("hello"))
    id2 := p.Intern([]byte("hello"))
    if id1 != id2 { t.Fatalf("same string got different IDs: %d vs %d", id1, id2) }
    if id1 == 0 { t.Fatal("id must not be 0") }
}

func TestInternDifferentStrings(t *testing.T) {
    p := NewInternPool(16)
    id1 := p.Intern([]byte("foo"))
    id2 := p.Intern([]byte("bar"))
    if id1 == id2 { t.Fatal("different strings got same ID") }
}

func TestInternEmpty(t *testing.T) {
    p := NewInternPool(16)
    id := p.Intern([]byte{})
    if id != 0 { t.Fatalf("empty string must return 0, got %d", id) }
}

func TestInternGet(t *testing.T) {
    p := NewInternPool(16)
    id := p.Intern([]byte("axiom"))
    got := p.Get(id)
    if got != "axiom" { t.Fatalf("Get(%d) = %q, want %q", id, got, "axiom") }
}

func TestInternGrow(t *testing.T) {
    p := NewInternPool(4)
    // Insert enough entries to trigger multiple resizes
    ids := make(map[string]uint32)
    for i := 0; i < 200; i++ {
        s := fmt.Sprintf("var_%d", i)
        id := p.InternString(s)
        if id == 0 { t.Fatalf("got id=0 for %q", s) }
        ids[s] = id
    }
    // Verify all are still correct after resizes
    for s, want := range ids {
        got := p.InternString(s)
        if got != want { t.Errorf("after grow: %q: id changed from %d to %d", s, want, got) }
    }
    if p.Len() != 200 { t.Fatalf("Len=%d, want 200", p.Len()) }
}

func TestInternLen(t *testing.T) {
    p := NewInternPool(16)
    p.Intern([]byte("a"))
    p.Intern([]byte("b"))
    p.Intern([]byte("a")) // duplicate
    if p.Len() != 2 { t.Fatalf("Len=%d, want 2", p.Len()) }
}

func TestInternGetBytes_NoAlloc(t *testing.T) {
    p := NewInternPool(16)
    id := p.Intern([]byte("hello"))
    b1 := p.GetBytes(id)
    b2 := p.GetBytes(id)
    // Same underlying slice (no copy)
    if &b1[0] != &b2[0] { t.Error("GetBytes should return same slice") }
}

func TestWellKnownIDs(t *testing.T) {
    p, wk := NewInternPoolWithWellKnown(16)
    if wk.Main == 0 { t.Error("Main ID must not be 0") }
    if p.Get(wk.Main) != "main" { t.Error("Main ID resolves to wrong string") }
}

func BenchmarkInternLookup(b *testing.B) {
    p := NewInternPool(256)
    s := []byte("some_identifier")
    p.Intern(s)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        p.Intern(s)
    }
}
```

## Validation Checklist
- [ ] Same `[]byte` content always returns same ID
- [ ] Different content always returns different IDs
- [ ] Empty slice returns ID=0
- [ ] `Get(id)` returns correct string
- [ ] `GetBytes(id)` returns slice into arena (no allocation)
- [ ] Table grows correctly when load factor > 0.7
- [ ] After growth, all previously interned IDs still resolve correctly
- [ ] `Len()` counts unique strings only
- [ ] `WellKnownIDs` pre-intern common strings
- [ ] `go test ./compiler/ast/` passes all tests

## Acceptance Criteria
- 200 unique strings interned: all `Get()` calls return correct values
- `BenchmarkInternLookup` shows sub-100ns per lookup
- No heap allocations on `Intern()` hit (existing string): verified with `benchmem`
- `go test -race ./compiler/ast/` passes

## Definition of Done
- [ ] `compiler/ast/intern.go` committed
- [ ] `compiler/ast/intern_test.go` committed with all tests + benchmark
- [ ] All tests pass
- [ ] Lint passes

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Hash collisions cause incorrect ID assignment | Test with strings that share FNV-1a prefix; add collision test |
| Arena grows without bound for large programs | Pre-allocate generously; arena is freed when compilation unit is done |
| `grow()` called too often with small initial cap | Use `NewInternPool(256)` as default; tune based on real programs |
| `uint16` length limit for very long strings | Check and panic; identifiers are always short |

## Future Follow-up Tasks
- p03-t04: Parser calls `pool.Intern(tree.TokenText(identIdx))` for each identifier
- p04-t01: Symbol table uses intern IDs as keys (no string comparison)
- p04-t04: Name resolver uses intern IDs for scope lookups
