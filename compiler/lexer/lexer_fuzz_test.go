package lexer

import "testing"

// FuzzLexer feeds random byte sequences into the lexer to find panics,
// out-of-bounds reads, or infinite loops. The lexer must NEVER panic on
// any input — it must always return a valid token slice (possibly with
// TokenError diagnostics).
func FuzzLexer(f *testing.F) {
	// Seed corpus with interesting patterns
	seeds := []string{
		"",
		"fn main():\n    return 0\n",
		"let x: i32 = 42",
		"0xFF 0b1010 0o777 3.14e-6",
		`"hello\nworld"`,
		"'a' '\\n' '\\u{1F600}'",
		"// comment\n",
		"if x > 0:\n    if y < 0:\n        z()\n",
		"import std.fs { read, write }",
		"struct Foo:\n    pub mut x: i32\n",
		"match x:\n    Ok(v): v\n    Err(e): panic(e)\n",
		"spawn worker(isolate(data))",
		"x := y + z ** w",
		"let result: Result[i32, Error] = Ok(42)",
		"pub async fn fetch(url: string) -> Future[string]:",
		"unsafe:\n    let p: *mut i32 = alloc()\n",
		"\t\t\t\t", // tabs
		"   \n   \n   \n",
		"@#$%^&*()!~`",
		"\x00\x01\x02\xff\xfe\xfd",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic
		tokens, lt, _ := Lex(data)

		// Basic invariants
		if tokens == nil {
			t.Fatal("Lex() returned nil token slice")
		}
		if len(tokens) == 0 {
			t.Fatal("Lex() returned empty token slice (must have at least EOF)")
		}
		if tokens[len(tokens)-1].Kind != TokenEOF {
			t.Fatal("last token must be EOF")
		}
		if lt == nil {
			t.Fatal("Lex() returned nil LineTable")
		}
	})
}
