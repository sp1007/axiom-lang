package lexer

import (
	"testing"
	"unsafe"
)

func TestTokenSize(t *testing.T) {
	const want = 8
	got := unsafe.Sizeof(Token{})
	if got != want {
		t.Fatalf("Token size = %d bytes, want %d bytes. "+
			"Token layout is FROZEN. Do not add fields without an RFC.", got, want)
	}
}

func TestTokenFieldOffsets(t *testing.T) {
	var tok Token
	base := uintptr(unsafe.Pointer(&tok))
	kindOff := uintptr(unsafe.Pointer(&tok.Kind)) - base
	lenOff := uintptr(unsafe.Pointer(&tok.Len)) - base
	offOff := uintptr(unsafe.Pointer(&tok.Offset)) - base
	if kindOff != 0 {
		t.Errorf("Token.Kind offset = %d, want 0", kindOff)
	}
	if lenOff != 2 {
		t.Errorf("Token.Len offset = %d, want 2", lenOff)
	}
	if offOff != 4 {
		t.Errorf("Token.Offset offset = %d, want 4", offOff)
	}
}

func TestTokenZeroValue(t *testing.T) {
	var tok Token
	if tok.Kind != 0 {
		t.Errorf("Token zero value Kind = %d, want 0", tok.Kind)
	}
	if tok.Len != 0 {
		t.Errorf("Token zero value Len = %d, want 0", tok.Len)
	}
	if tok.Offset != 0 {
		t.Errorf("Token zero value Offset = %d, want 0", tok.Offset)
	}
}
