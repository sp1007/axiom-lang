package codegen_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axiom-lang/axiom/codegen/cgen"
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestAxiomRuntimePorts(t *testing.T) {
	workspaceDir := "../.." // relative to tests/codegen
	runtimeDir := filepath.Join(workspaceDir, "runtime")

	t.Run("Scheduler", func(t *testing.T) {
		memAllocPath := filepath.Join(workspaceDir, "std/mem/alloc.ax")
		allocPath := filepath.Join(workspaceDir, "std/scheduler.ax")
		testPath := filepath.Join(workspaceDir, "std/scheduler_test.ax")

		sourceBytes, err := concatenateAxiomFiles(memAllocPath, allocPath, testPath)
		if err != nil {
			t.Fatalf("failed to concatenate scheduler files: %v", err)
		}

		tmpDir, err := os.MkdirTemp("", "axiom-scheduler-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		binPath := filepath.Join(tmpDir, "scheduler_test")
		if filepath.Separator == '\\' {
			binPath += ".exe"
		}

		if err := compileCBackendWithActor(t, sourceBytes, binPath, runtimeDir); err != nil {
			t.Fatalf("failed to compile AXIOM scheduler: %v", err)
		}

		// Copy generated C file for debugging
		if cBytes, err := os.ReadFile(binPath + ".c"); err == nil {
			_ = os.WriteFile("../../scheduler_test_debug.c", cBytes, 0644)
		}

		cmd := exec.Command(binPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("running AXIOM scheduler test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		output := stdout.String()
		t.Logf("Scheduler Test Output:\n%s", output)

		if !strings.Contains(output, "All AXIOM-native Scheduler tests passed!") {
			t.Errorf("expected test success message, got output:\n%s", output)
		}
	})

	t.Run("Reactor", func(t *testing.T) {
		allocPath := filepath.Join(workspaceDir, "std/reactor.ax")
		testPath := filepath.Join(workspaceDir, "std/reactor_test.ax")

		sourceBytes, err := concatenateAxiomFiles(allocPath, testPath)
		if err != nil {
			t.Fatalf("failed to concatenate reactor files: %v", err)
		}

		tmpDir, err := os.MkdirTemp("", "axiom-reactor-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		binPath := filepath.Join(tmpDir, "reactor_test")
		if filepath.Separator == '\\' {
			binPath += ".exe"
		}

		if err := compileCBackendWithActor(t, sourceBytes, binPath, runtimeDir); err != nil {
			t.Fatalf("failed to compile AXIOM reactor: %v", err)
		}

		cmd := exec.Command(binPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("running AXIOM reactor test failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		}

		output := stdout.String()
		t.Logf("Reactor Test Output:\n%s", output)

		if !strings.Contains(output, "All AXIOM-native Reactor tests passed!") {
			t.Errorf("expected test success message, got output:\n%s", output)
		}
	})
}

func compileCBackendWithActor(t *testing.T, source []byte, outPath string, runtimeDir string) error {
	// Lex
	tokens, _, lexDiags := lexer.Lex(source)
	if hasErrorsIgnoringAt(lexDiags) {
		return fmt.Errorf("lex errors: %v", lexDiags)
	}

	// Parse
	intern := ast.NewInternPool(256)
	tree, parseDiags := parser.Parse(tokens, source, intern)
	if hasErrors(parseDiags) {
		return fmt.Errorf("parse errors: %v", parseDiags)
	}

	// Print AST
	astFile, _ := os.Create("ast.txt")
	ast.Print(astFile, tree, intern)
	astFile.Close()

	// Semantic analysis
	symbols := sema.NewSymbolTable(intern)
	table := types.NewTypeTable()



	resolver := sema.NewNameResolver(tree, intern, symbols, table, nil)
	if errs := resolver.Resolve(); hasErrors(errs) {
		opts := diagnostics.DefaultFormatOptions()
		opts.UseColor = false
		formatted := diagnostics.FormatDiagnostics(errs, source, "tmp_port_compile.ax", opts)
		return fmt.Errorf("name resolution errors:\n%s", formatted)
	}

	mono := sema.NewMonomorphizer(tree, intern, symbols, table)
	infer := sema.NewInferenceEngine(tree, symbols, table, mono)
	if errs := infer.Infer(); hasErrors(errs) {
		opts := diagnostics.DefaultFormatOptions()
		opts.UseColor = false
		formatted := diagnostics.FormatDiagnostics(errs, source, "tmp_port_compile.ax", opts)
		return fmt.Errorf("type inference errors:\n%s", formatted)
	}

	tc := sema.NewTypeChecker(tree, intern, symbols, table, infer)
	if errs := tc.Check(); hasErrors(errs) {
		opts := diagnostics.DefaultFormatOptions()
		opts.UseColor = false
		formatted := diagnostics.FormatDiagnostics(errs, source, "tmp_port_compile.ax", opts)
		return fmt.Errorf("type check errors:\n%s", formatted)
	}

	if err := runCTGCAndOwnership(tree, intern, symbols, table, infer); err != nil {
		return err
	}

	// Transpile to C
	pipeline := cgen.NewPipeline(table, intern, symbols, tree)
	cSrcPath := outPath + ".c"
	cFile, err := os.Create(cSrcPath)
	if err != nil {
		return err
	}
	defer cFile.Close()

	// Write helper functions to bridge void* parameters to C runtime AxActor* parameters
	cFile.WriteString("struct AxActor;\n")
	cFile.WriteString("int ax_actor_step(struct AxActor* actor);\n")
	cFile.WriteString("int ax_actor_is_running(struct AxActor* actor);\n")
	cFile.WriteString("int ax_actor_has_messages(struct AxActor* actor);\n\n")

	cFile.WriteString("int ax_actor_step_impl(void* actor_ptr) {\n")
	cFile.WriteString("    return ax_actor_step((struct AxActor*)actor_ptr);\n")
	cFile.WriteString("}\n")
	cFile.WriteString("int ax_actor_is_running_impl(void* actor_ptr) {\n")
	cFile.WriteString("    return ax_actor_is_running((struct AxActor*)actor_ptr);\n")
	cFile.WriteString("}\n")
	cFile.WriteString("int ax_actor_has_messages_impl(void* actor_ptr) {\n")
	cFile.WriteString("    return ax_actor_has_messages((struct AxActor*)actor_ptr);\n")
	cFile.WriteString("}\n\n")

	if err := pipeline.GenerateC(cFile); err != nil {
		return err
	}
	cFile.Close()

	// Build extra C sources list, excluding duplicates if they are implemented in AXIOM
	extraSrcs := []string{
		filepath.Join(runtimeDir, "axalloc", "axalloc.c"),
		filepath.Join(runtimeDir, "panic", "panic.c"),
		filepath.Join(runtimeDir, "ax_assert.c"),
		filepath.Join(runtimeDir, "ax_collections.c"),
		filepath.Join(runtimeDir, "ax_math.c"),
		filepath.Join(runtimeDir, "ax_print.c"),
		filepath.Join(runtimeDir, "ax_string_ops.c"),
		filepath.Join(runtimeDir, "actor", "actor.c"),
		filepath.Join(runtimeDir, "actor", "msgqueue.c"),
		filepath.Join(runtimeDir, "actor", "async.c"),
		filepath.Join(runtimeDir, "actor", "runtime_init.c"),
		filepath.Join(runtimeDir, "actor", "supervisor.c"),
		filepath.Join(runtimeDir, "actor", "isolated.c"),
	}

	if !strings.Contains(outPath, "scheduler") {
		extraSrcs = append(extraSrcs, filepath.Join(runtimeDir, "actor", "scheduler.c"))
	}
	if !strings.Contains(outPath, "reactor") {
		extraSrcs = append(extraSrcs, filepath.Join(runtimeDir, "actor", "ioloop.c"))
	}

	// Compile C with full actor libraries
	err = pipeline.CompileCWithOptions(outPath, cSrcPath, cgen.CompileOptions{
		IncludeDirs: []string{runtimeDir},
		ExtraSrcs:   extraSrcs,
		Debug:       true,
	})
	return err
}
