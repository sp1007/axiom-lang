// Package lexer implements the AXIOM source lexer.
// It converts raw UTF-8 source bytes into a flat []Token slice
// with no string allocations (zero-copy design).
package lexer

import "unsafe"

// Token represents a single lexical token.
// Layout is FROZEN at 8 bytes. Do not add fields without an RFC.
// Zero-copy design: the token does not own its text.
// Use src[tok.Offset : tok.Offset+uint32(tok.Len)] to recover the text.
//
// FROZEN: do not modify without RFC
type Token struct {
	Kind   TokenKind // 1 byte (uint8)
	_      uint8     // 1 byte padding (reserved, must remain zero)
	Len    uint16    // 2 bytes: length of token text in source bytes
	Offset uint32    // 4 bytes: byte offset of token start in source
}

// TokenKind is the discriminant of a Token.
// Must fit in uint8 (max 255 values). See token_kind.go for the full enum (p02-t01).
type TokenKind uint8

// Compile-time size assertion — ensures Token is exactly 8 bytes.
// If this line produces a compile error, Token has been modified incorrectly.
var _ = [1]struct{}{}[8-unsafe.Sizeof(Token{})]
var _ = [1]struct{}{}[unsafe.Sizeof(Token{})-8]
