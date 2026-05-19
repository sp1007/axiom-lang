package ast

import (
	"testing"
	"unsafe"
)

func TestAstNodeSize(t *testing.T) {
	const want = 24
	got := unsafe.Sizeof(AstNode{})
	if got != want {
		t.Fatalf("AstNode size = %d bytes, want %d bytes. "+
			"AstNode layout is FROZEN. Do not add fields without an RFC.", got, want)
	}
}

func TestAstNodeFieldOffsets(t *testing.T) {
	var n AstNode
	base := uintptr(unsafe.Pointer(&n))
	check := func(name string, got, want uintptr) {
		t.Helper()
		if got != want {
			t.Errorf("AstNode.%s offset = %d, want %d", name, got, want)
		}
	}
	check("Kind", uintptr(unsafe.Pointer(&n.Kind))-base, 0)
	check("Flags", uintptr(unsafe.Pointer(&n.Flags))-base, 2)
	check("TokenIdx", uintptr(unsafe.Pointer(&n.TokenIdx))-base, 4)
	check("FirstChild", uintptr(unsafe.Pointer(&n.FirstChild))-base, 8)
	check("NextSibling", uintptr(unsafe.Pointer(&n.NextSibling))-base, 12)
	check("Payload", uintptr(unsafe.Pointer(&n.Payload))-base, 16)
	check("ExtraIdx", uintptr(unsafe.Pointer(&n.ExtraIdx))-base, 20)
}

func TestNodeKindCount(t *testing.T) {
	// Ensure NodeKindCount fits in uint8 (NodeKind is uint8)
	if NodeKindCount > 255 {
		t.Fatalf("NodeKindCount = %d exceeds uint8 max (255)", NodeKindCount)
	}
}

func TestFlagConstants(t *testing.T) {
	// All flag constants must be distinct powers of 2
	flags := []uint16{
		FlagIsPub, FlagIsMut, FlagIsAsync, FlagIsExtern,
		FlagIsSink, FlagIsLent, FlagIsPacked, FlagEscapesToHeap,
		FlagUsesArena, FlagIsGeneric, FlagIsMoved,
	}
	seen := map[uint16]bool{}
	for _, f := range flags {
		if f == 0 {
			t.Error("flag must not be zero")
		}
		if f&(f-1) != 0 {
			t.Errorf("flag 0x%04x is not a power of 2", f)
		}
		if seen[f] {
			t.Errorf("duplicate flag value 0x%04x", f)
		}
		seen[f] = true
	}
}

func TestNodeKindZeroIsInvalid(t *testing.T) {
	if NodeInvalid != 0 {
		t.Fatalf("NodeInvalid = %d, want 0 (sentinel value)", NodeInvalid)
	}
}
