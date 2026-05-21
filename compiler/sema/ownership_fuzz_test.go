package sema_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// runFullSemaPipeline runs the complete semantic pipeline on src bytes.
// Returns all diagnostics without panicking.
func runFullSemaPipeline(src []byte) []diagnostics.Diagnostic {
	toks, _, _ := lexer.Lex(src)
	pool := ast.NewInternPool(16)
	tree, parseDiags := parser.Parse(toks, src, pool)
	_ = parseDiags

	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	lazy := sema.NewLazyResolver(st, tt, nil)
	nr := sema.NewNameResolver(tree, pool, st, tt, lazy)
	nr.Resolve()

	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	ie.Infer()

	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tcDiags := tc.Check()

	oc := sema.NewOwnershipChecker(tree, pool, st, tt)
	ocDiags := oc.Check()

	ea := sema.NewEscapeAnalysis(tree, pool, st, tt)
	_ = ea // escape analysis reads CG; with no pre-built CG it's a no-op for fuzz

	moved := oc.Moved()
	ctgc := sema.NewCTGCPass(tree, st, moved)
	ctgc.InjectDestroys(0)

	var allDiags []diagnostics.Diagnostic
	allDiags = append(allDiags, tcDiags...)
	allDiags = append(allDiags, ocDiags...)
	return allDiags
}

// FuzzOwnershipChecker exercises the full semantic pipeline with random inputs.
// Goal: no panics on any input.
func FuzzOwnershipChecker(f *testing.F) {
	// Inline seed corpus — representative ownership patterns
	seeds := []string{
		`fn main():
    let x = 42
`,
		`fn main():
    let x = 42
    let y = x
    let z = x
`,
		`fn main():
    mut x = 5
    x = 10
`,
		`fn make() -> i32:
    let x = 42
    return x
`,
		`fn add(a: i32, b: i32) -> i32:
    return a + b

fn main():
    let r = add(1, 2)
`,
		`struct Foo:
    x: i32

fn main():
    let f = Foo{x: 1}
`,
		`type Color = Red | Green | Blue

fn main():
    let c = Color.Red
`,
		`interface Printable:
    fn print(self: Self) -> string

fn main():
    let x = 42
`,
		`fn main():
    let x = 42
    if x > 0:
        let y = x
    return x
`,
		`fn main():
    let x = true
    let y = false
    let z = x and y
`,
	}

	// Load seed corpus from test files
	for _, dir := range []string{"tests/sema", "tests/generics"} {
		absDir := filepath.Join(findProjectRoot(), dir)
		entries, err := os.ReadDir(absDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if filepath.Ext(e.Name()) == ".ax" {
				data, err := os.ReadFile(filepath.Join(absDir, e.Name()))
				if err == nil && len(data) > 0 {
					f.Add(data)
				}
			}
		}
	}

	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, src []byte) {
		// Limit input size to prevent OOM
		if len(src) > 10_000 {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ownership pipeline panic: %v\n%s", r, debug.Stack())
			}
		}()

		runFullSemaPipeline(src)
	})
}

// findProjectRoot walks up from the test file to find the project root.
func findProjectRoot() string {
	// Start from the working directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	// Walk up looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// TestOwnershipNoPanicOnEmpty verifies the pipeline doesn't panic on empty input.
func TestOwnershipNoPanicOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on empty input: %v", r)
		}
	}()
	runFullSemaPipeline([]byte{})
}

// TestOwnershipNoPanicOnGarbage verifies no panic on random garbage.
func TestOwnershipNoPanicOnGarbage(t *testing.T) {
	garbage := [][]byte{
		[]byte("@#$%^&*()"),
		[]byte("\x00\x01\x02\x03"),
		[]byte("fn "),
		[]byte("struct struct struct"),
		[]byte("let let let"),
		[]byte("fn main():\n    let x = 42\n    let y = x\n    let z = y\n    let w = z"),
		[]byte("fn a():\n    let x = 1\n"),
		[]byte(""),
		[]byte("   "),
		[]byte("fn main():\n    42"),
	}

	for i, g := range garbage {
		t.Run(fmt.Sprintf("garbage_%d", i), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic on garbage input %d: %v\n%s", i, r, debug.Stack())
				}
			}()
			runFullSemaPipeline(g)
		})
	}
}

// TestOwnershipValidPrograms verifies that valid programs produce no ownership errors.
func TestOwnershipValidPrograms(t *testing.T) {
	validPrograms := []struct {
		name string
		src  string
	}{
		{"simple_let", "fn main():\n    let x = 42\n"},
		{"mutable_assign", "fn main():\n    mut x = 5\n    x = 10\n"},
		{"function_call", "fn add(a: i32, b: i32) -> i32:\n    return a + b\n\nfn main():\n    let r = add(1, 2)\n"},
		{"return_value", "fn make() -> i32:\n    let x = 42\n    return x\n"},
		{"bool_expr", "fn main():\n    let x = true\n    let y = false\n"},
		{"string_lit", "fn main():\n    let s = \"hello\"\n"},
		{"multiple_lets", "fn main():\n    let a = 1\n    let b = 2\n    let c = 3\n"},
		{"nested_calls", "fn id(x: i32) -> i32:\n    return x\n\nfn main():\n    let r = id(42)\n"},
	}

	for _, tt := range validPrograms {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %v\n%s", r, debug.Stack())
				}
			}()
			diags := runFullSemaPipeline([]byte(tt.src))
			// Filter only ownership errors (code 4001, 4002)
			var ownershipErrors int
			for _, d := range diags {
				if d.Code >= 4001 && d.Code <= 4099 {
					ownershipErrors++
				}
			}
			if ownershipErrors > 0 {
				t.Errorf("valid program %q got %d ownership errors", tt.name, ownershipErrors)
				for _, d := range diags {
					if d.Code >= 4001 && d.Code <= 4099 {
						t.Logf("  [%d] %s", d.Code, d.Message)
					}
				}
			}
		})
	}
}

// TestOwnershipSeedCorpusNoPanic loads all .ax test files and verifies no panics.
func TestOwnershipSeedCorpusNoPanic(t *testing.T) {
	root := findProjectRoot()
	dirs := []string{
		filepath.Join(root, "tests", "sema"),
		filepath.Join(root, "tests", "generics"),
	}

	count := 0
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if filepath.Ext(e.Name()) != ".ax" {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			t.Run(e.Name(), func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic on %s: %v\n%s", e.Name(), r, debug.Stack())
					}
				}()
				runFullSemaPipeline(data)
			})
			count++
		}
	}

	if count == 0 {
		t.Skip("no seed corpus files found")
	}
	t.Logf("tested %d seed corpus files without panics", count)
}
