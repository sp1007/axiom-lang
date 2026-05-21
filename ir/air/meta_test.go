package air

import "testing"

// ---------------------------------------------------------------------------
// AIHintTable tests
// ---------------------------------------------------------------------------

func TestAIHintTable_ZeroSentinel(t *testing.T) {
	ht := NewAIHintTable()
	if got := ht.Get(0); got != nil {
		t.Errorf("Get(0) should return nil (sentinel), got %v", got)
	}
	if ht.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for empty table", ht.Len())
	}
}

func TestAIHintTable_AddAndGet(t *testing.T) {
	ht := NewAIHintTable()
	hints := []AIHint{
		{Kind: AIHintAssertPure, Data: "no side effects"},
		{Kind: AIHintSuggestVec, Data: "vectorize inner loop"},
	}
	idx := ht.Add(hints)
	if idx == 0 {
		t.Fatal("Add should never return 0 (reserved sentinel)")
	}
	got := ht.Get(idx)
	if len(got) != 2 {
		t.Fatalf("Get(%d) returned %d hints, want 2", idx, len(got))
	}
	if got[0].Kind != AIHintAssertPure || got[1].Kind != AIHintSuggestVec {
		t.Errorf("hint kinds mismatch: %v", got)
	}
	if ht.Len() != 1 {
		t.Errorf("Len() = %d, want 1", ht.Len())
	}
}

func TestAIHintTable_OutOfRange(t *testing.T) {
	ht := NewAIHintTable()
	if got := ht.Get(999); got != nil {
		t.Errorf("Get(999) should return nil for out-of-range, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// MetaTable tests
// ---------------------------------------------------------------------------

func TestMetaTable_SetAndGet(t *testing.T) {
	mt := NewMetaTable()
	meta := AirMeta{
		SourceFile: 1,
		SourceLine: 42,
		SourceCol:  8,
		OwnerInfo:  OwnerHeap,
		AIHints:    0,
	}
	mt.Set(5, meta)

	got := mt.Get(5)
	if got == nil {
		t.Fatal("Get(5) returned nil")
	}
	if got.SourceLine != 42 {
		t.Errorf("SourceLine = %d, want 42", got.SourceLine)
	}
	if got.OwnerInfo != OwnerHeap {
		t.Errorf("OwnerInfo = %d, want %d (OwnerHeap)", got.OwnerInfo, OwnerHeap)
	}
}

func TestMetaTable_GetMissing(t *testing.T) {
	mt := NewMetaTable()
	if got := mt.Get(99); got != nil {
		t.Errorf("Get(99) should return nil for missing entry, got %v", got)
	}
}

func TestMetaTable_Overwrite(t *testing.T) {
	mt := NewMetaTable()
	mt.Set(0, AirMeta{SourceLine: 10})
	mt.Set(0, AirMeta{SourceLine: 20})
	got := mt.Get(0)
	if got == nil || got.SourceLine != 20 {
		t.Errorf("overwrite failed: got %v", got)
	}
	if mt.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after overwrite", mt.Len())
	}
}

func TestMetaTable_Delete(t *testing.T) {
	mt := NewMetaTable()
	mt.Set(3, AirMeta{SourceLine: 100})
	if mt.Len() != 1 {
		t.Fatalf("Len() = %d before delete, want 1", mt.Len())
	}
	mt.Delete(3)
	if mt.Get(3) != nil {
		t.Error("Get(3) should return nil after Delete")
	}
	if mt.Len() != 0 {
		t.Errorf("Len() = %d after delete, want 0", mt.Len())
	}
}

func TestMetaTable_Len(t *testing.T) {
	mt := NewMetaTable()
	mt.Set(0, AirMeta{SourceLine: 1})
	mt.Set(1, AirMeta{SourceLine: 2})
	mt.Set(2, AirMeta{SourceLine: 3})
	if mt.Len() != 3 {
		t.Errorf("Len() = %d, want 3", mt.Len())
	}
}

func TestOwnerInfoConstants(t *testing.T) {
	if OwnerNone != 0 || OwnerStack != 1 || OwnerHeap != 2 || OwnerArena != 3 {
		t.Errorf("OwnerInfo constants wrong: none=%d stack=%d heap=%d arena=%d",
			OwnerNone, OwnerStack, OwnerHeap, OwnerArena)
	}
}

func TestAIHintKindValues(t *testing.T) {
	// Ensure iota ordering.
	if AIHintAssertPure != 0 || AIHintSuggestSoA != 1 || AIHintExplain != 2 || AIHintSuggestVec != 3 {
		t.Errorf("AIHintKind iota values wrong: pure=%d soa=%d explain=%d vec=%d",
			AIHintAssertPure, AIHintSuggestSoA, AIHintExplain, AIHintSuggestVec)
	}
}
