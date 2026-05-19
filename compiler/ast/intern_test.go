package ast

import (
	"fmt"
	"testing"
)

func TestInternBasic(t *testing.T) {
	p := NewInternPool(16)
	id1 := p.Intern([]byte("hello"))
	id2 := p.Intern([]byte("hello"))
	if id1 != id2 {
		t.Fatalf("same string got different IDs: %d vs %d", id1, id2)
	}
	if id1 == 0 {
		t.Fatal("id must not be 0")
	}
}

func TestInternDifferentStrings(t *testing.T) {
	p := NewInternPool(16)
	id1 := p.Intern([]byte("foo"))
	id2 := p.Intern([]byte("bar"))
	if id1 == id2 {
		t.Fatalf("different strings got same ID: %d", id1)
	}
}

func TestInternEmpty(t *testing.T) {
	p := NewInternPool(16)
	id := p.Intern([]byte{})
	if id != 0 {
		t.Fatalf("empty string must return 0, got %d", id)
	}
}

func TestInternGet(t *testing.T) {
	p := NewInternPool(16)
	id := p.Intern([]byte("axiom"))
	got := p.Get(id)
	if got != "axiom" {
		t.Fatalf("Get(%d) = %q, want %q", id, got, "axiom")
	}
}

func TestInternGrow(t *testing.T) {
	p := NewInternPool(4)
	ids := make(map[string]uint32)
	for i := 0; i < 200; i++ {
		s := fmt.Sprintf("var_%d", i)
		id := p.InternString(s)
		if id == 0 {
			t.Fatalf("got id=0 for %q", s)
		}
		ids[s] = id
	}
	// Verify all are still correct after resizes
	for s, want := range ids {
		got := p.InternString(s)
		if got != want {
			t.Errorf("after grow: %q: id changed from %d to %d", s, want, got)
		}
	}
	if p.Len() != 200 {
		t.Fatalf("Len=%d, want 200", p.Len())
	}
}

func TestInternLen(t *testing.T) {
	p := NewInternPool(16)
	p.Intern([]byte("a"))
	p.Intern([]byte("b"))
	p.Intern([]byte("a")) // duplicate
	if p.Len() != 2 {
		t.Fatalf("Len=%d, want 2", p.Len())
	}
}

func TestInternGetBytesNoAlloc(t *testing.T) {
	p := NewInternPool(16)
	id := p.Intern([]byte("hello"))
	b1 := p.GetBytes(id)
	b2 := p.GetBytes(id)
	// Same underlying slice reference
	if &b1[0] != &b2[0] {
		t.Error("GetBytes should return same slice into arena")
	}
}

func TestInternStringWrapper(t *testing.T) {
	p := NewInternPool(16)
	id1 := p.InternString("test")
	id2 := p.Intern([]byte("test"))
	if id1 != id2 {
		t.Fatalf("InternString and Intern disagree: %d vs %d", id1, id2)
	}
}

func TestWellKnownIDs(t *testing.T) {
	p, wk := NewInternPoolWithWellKnown(16)
	if wk.Main == 0 {
		t.Error("Main ID must not be 0")
	}
	if p.Get(wk.Main) != "main" {
		t.Errorf("Main = %q, want %q", p.Get(wk.Main), "main")
	}
	if wk.I32 == 0 {
		t.Error("I32 ID must not be 0")
	}
	if p.Get(wk.I32) != "i32" {
		t.Errorf("I32 = %q, want %q", p.Get(wk.I32), "i32")
	}
}

func TestInternGetPanicOnZero(t *testing.T) {
	p := NewInternPool(16)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on Get(0)")
		}
	}()
	p.Get(0)
}

func TestInternManyCollisions(t *testing.T) {
	// Force many entries into a small table to stress linear probing
	p := NewInternPool(4)
	for i := 0; i < 100; i++ {
		s := fmt.Sprintf("%c", 'a'+i%26)
		p.InternString(s + fmt.Sprintf("%d", i))
	}
	if p.Len() != 100 {
		t.Fatalf("expected 100, got %d", p.Len())
	}
	// Verify all IDs resolve correctly
	for id := uint32(1); id <= uint32(p.Len()); id++ {
		s := p.Get(id)
		if len(s) == 0 {
			t.Errorf("Get(%d) returned empty string", id)
		}
	}
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

func BenchmarkInternInsert(b *testing.B) {
	p := NewInternPool(256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.InternString(fmt.Sprintf("var_%d", i))
	}
}
