package fmt

import (
	"bytes"
	"testing"
)

func TestFormatter_Idempotency(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "simple function",
			src: `fn main() {
    let x = 10
    let y = 20
    let z = x + y
    println(z)
}
`,
		},
		{
			name: "nested conditions",
			src: `fn process(val: i32) -> bool {
    if val > 0 {
        return true
    } else {
        return false
    }
}
`,
		},
		{
			name: "comments and operators",
			src: `// Leading module comment
import std.io

fn main() {
    let x: i32 = -5            // Negative number
    let y = x * 2 + 10          // Complex arithmetic
    if x < 0 and not false {
        std.io.println("negative")
    }
}
`,
		},
	}

	formatter := NewFormatter()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			firstPass, err := formatter.Format([]byte(tc.src))
			if err != nil {
				t.Fatalf("first format failed: %v", err)
			}
			secondPass, err := formatter.Format(firstPass)
			if err != nil {
				t.Fatalf("second format failed: %v", err)
			}
			if !bytes.Equal(firstPass, secondPass) {
				t.Errorf("formatter is not idempotent!\nFirst pass:\n%s\nSecond pass:\n%s", string(firstPass), string(secondPass))
			}
		})
	}
}

func TestFormatter_Indentation(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "tabs to spaces",
			src: `fn main() {
	let x = 1
	if x == 1 {
		println("one")
	}
}`,
			want: `fn main() {
    let x = 1
    if x == 1 {
        println("one")
    }
}
`,
		},
		{
			name: "arbitrary spaces correction",
			src: `fn main() {
  let x = 1
    if x == 1 {
       println("one")
    }
}`,
			want: `fn main() {
    let x = 1
    if x == 1 {
        println("one")
    }
}
`,
		},
	}

	formatter := NewFormatter()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := formatter.Format([]byte(tc.src))
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("indentation incorrect.\nGot:\n%q\nWant:\n%q", string(got), tc.want)
			}
		})
	}
}

func TestFormatter_OperatorSpacing(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "binary operators",
			src:  `let x=1+2*3/4-5`,
			want: `let x = 1 + 2 * 3 / 4 - 5
`,
		},
		{
			name: "unary operators",
			src:  `let x = - 5 ; let y = ! true ; let z = ~ val`,
			want: `let x = -5; let y = !true; let z = ~val
`,
		},
		{
			name: "type annotation and comma spacing",
			src:  `fn foo ( a :i32 , b : string )`,
			want: `fn foo(a: i32, b: string)
`,
		},
	}

	formatter := NewFormatter()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := formatter.Format([]byte(tc.src))
			if err != nil {
				t.Fatalf("format failed: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("operator/punctuation spacing incorrect.\nGot:\n%q\nWant:\n%q", string(got), tc.want)
			}
		})
	}
}

func TestFormatter_ImportSorting(t *testing.T) {
	src := `import std.io
import local.mod
import thirdparty.pkg
import std.fs
import .relative
import std.net`

	want := `import std.fs
import std.io
import std.net

import thirdparty.pkg

import .relative
import local.mod
`

	formatter := NewFormatter()
	got, err := formatter.Format([]byte(src))
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if string(got) != want {
		t.Errorf("import sorting incorrect.\nGot:\n%q\nWant:\n%q", string(got), want)
	}
}

func TestFormatter_InlineCommentAlignment(t *testing.T) {
	src := `fn main() {
    let x = 10 // first comment
    let longer_variable = 20 // second comment
    let y = 30     // third comment
}`

	want := `fn main() {
    let x = 10                          // first comment
    let longer_variable = 20            // second comment
    let y = 30                          // third comment
}
`

	formatter := NewFormatter()
	got, err := formatter.Format([]byte(src))
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if string(got) != want {
		t.Errorf("inline comment alignment incorrect.\nGot:\n%q\nWant:\n%q", string(got), want)
	}
}

func TestFormatter_BlankLines(t *testing.T) {
	src := `fn a() {}



fn b() {
    let x = 1


    let y = 2
}`

	want := `fn a() {}

fn b() {
    let x = 1

    let y = 2
}
`

	formatter := NewFormatter()
	got, err := formatter.Format([]byte(src))
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if string(got) != want {
		t.Errorf("blank lines collapsing/separation incorrect.\nGot:\n%q\nWant:\n%q", string(got), want)
	}
}
