package codegen_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStage1NativeLink(t *testing.T) {
	// 1. Concatenate all self-hosted compiler frontend files including native codegen & linker
	workspaceDir := "../.." // relative to tests/codegen
	printHelpersPath := filepath.Join(workspaceDir, "bootstrap/stage1/print_helpers.ax")
	tokenPath := filepath.Join(workspaceDir, "bootstrap/stage1/token.ax")
	lexerPath := filepath.Join(workspaceDir, "bootstrap/stage1/lexer.ax")
	astPath := filepath.Join(workspaceDir, "bootstrap/stage1/ast.ax")
	internPath := filepath.Join(workspaceDir, "bootstrap/stage1/intern.ax")
	parserPath := filepath.Join(workspaceDir, "bootstrap/stage1/parser.ax")
	resolverPath := filepath.Join(workspaceDir, "bootstrap/stage1/resolver.ax")
	typetablePath := filepath.Join(workspaceDir, "bootstrap/stage1/typetable.ax")
	monoPath := filepath.Join(workspaceDir, "bootstrap/stage1/mono.ax")
	typecheckPath := filepath.Join(workspaceDir, "bootstrap/stage1/typecheck.ax")
	connectionGraphPath := filepath.Join(workspaceDir, "bootstrap/stage1/connection_graph.ax")
	ownershipPath := filepath.Join(workspaceDir, "bootstrap/stage1/ownership.ax")
	escapePath := filepath.Join(workspaceDir, "bootstrap/stage1/escape.ax")
	ctgcPath := filepath.Join(workspaceDir, "bootstrap/stage1/ctgc.ax")
	aliasReusePath := filepath.Join(workspaceDir, "bootstrap/stage1/alias_reuse.ax")
	ssaOptPath := filepath.Join(workspaceDir, "bootstrap/stage1/ssa_opt.ax")
	airPath := filepath.Join(workspaceDir, "bootstrap/stage1/air.ax")
	airBuilderPath := filepath.Join(workspaceDir, "bootstrap/stage1/air_builder.ax")
	cgenPath := filepath.Join(workspaceDir, "bootstrap/stage1/cgen.ax")
	wasmPath := filepath.Join(workspaceDir, "bootstrap/stage1/wasm.ax")
	x86RegsPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_regs.ax")
	x86SelectorPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_selector.ax")
	x86RegallocPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_regalloc.ax")
	x86AsmEmitterPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_asm_emitter.ax")
	x86ModrmPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_modrm.ax")
	x86EncodingPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_encoding.ax")
	x86EmitterPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_emitter.ax")
	x86CoffPath := filepath.Join(workspaceDir, "bootstrap/stage1/x86_coff.ax")
	x86Elf64Path := filepath.Join(workspaceDir, "bootstrap/stage1/x86_elf64.ax")
	linkerPath := filepath.Join(workspaceDir, "bootstrap/stage1/linker.ax")
	fmtPath := filepath.Join(workspaceDir, "bootstrap/stage1/fmt.ax")
	mainPath := filepath.Join(workspaceDir, "bootstrap/stage1/main_air.ax")

	sourceBytes, err := concatenateAxiomFiles(
		printHelpersPath, tokenPath, lexerPath, astPath, internPath, parserPath, resolverPath, typetablePath, monoPath, typecheckPath,
		connectionGraphPath, ownershipPath, escapePath, ctgcPath, aliasReusePath,
		airPath, airBuilderPath, ssaOptPath, cgenPath, wasmPath,
		x86RegsPath, x86SelectorPath, x86RegallocPath, x86AsmEmitterPath,
		x86ModrmPath, x86EncodingPath, x86EmitterPath, x86Elf64Path, x86CoffPath,
		linkerPath, fmtPath, mainPath,
	)
	if err != nil {
		t.Fatalf("failed to concatenate stage1 files: %v", err)
	}

	// 2. Compile using C Backend to generate axc_stage1_selfhosted.exe
	tmpDir, err := os.MkdirTemp("", "axiom-native-bootstrap-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "axc_stage1_selfhosted")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	t.Logf("Compiling self-hosted compiler frontend to %s...", binPath)
	if err := compileCBackendIgnoringAtDiagnostics(t, sourceBytes, binPath, mainPath); err != nil {
		t.Fatalf("failed to compile self-hosted stage1 compiler: %v", err)
	}

	// Get absolute paths so they are independent of the command working directory
	absWorkspaceDir, err := filepath.Abs(workspaceDir)
	if err != nil {
		t.Fatalf("failed to get absolute workspace path: %v", err)
	}

	srcAxFile, err := filepath.Abs(filepath.Join(absWorkspaceDir, "tests/air/001_return_const.ax"))
	if err != nil {
		t.Fatalf("failed to get absolute source file path: %v", err)
	}

	outputExeName := "return_const_nativelinked"
	if filepath.Separator == '\\' {
		outputExeName += ".exe"
	}
	outputExe, err := filepath.Abs(filepath.Join(tmpDir, outputExeName))
	if err != nil {
		t.Fatalf("failed to get absolute output executable path: %v", err)
	}

	// 3. Copy ax_runtime.dll to tmpDir so that the loader can find it on Windows
	dllSrcPath := filepath.Join(absWorkspaceDir, "ax_runtime.dll")
	dllDestPath := filepath.Join(tmpDir, "ax_runtime.dll")
	dllBytes, err := os.ReadFile(dllSrcPath)
	if err == nil {
		t.Logf("Copying %s to %s for loader search path...", dllSrcPath, dllDestPath)
		_ = os.WriteFile(dllDestPath, dllBytes, 0644)
	} else {
		t.Logf("Warning: could not read ax_runtime.dll from %s: %v", dllSrcPath, err)
	}

	t.Logf("Running compiled self-hosted binary to build %s with direct native self-linking...", outputExe)
	
	// Create scratch dir if it doesn't exist
	err = os.MkdirAll(filepath.Join(absWorkspaceDir, "scratch"), 0755)
	if err != nil {
		t.Fatalf("failed to create scratch dir: %v", err)
	}

	cmd := exec.Command(binPath, "build", srcAxFile, "-o", outputExe)
	cmd.Dir = absWorkspaceDir // Set working directory to the repository root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run the self-hosted compiler
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run self-hosted compiler:\nErr: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}
	
	t.Logf("Self-hosted compiler ran successfully.\nStdout: %s\nStderr: %s", stdout.String(), stderr.String())

	// 4. Run the natively compiled & linked executable return_const_nativelinked
	t.Logf("Running natively-compiled and self-linked binary: %s", outputExe)
	runCmd := exec.Command(outputExe)
	var runStdout, runStderr bytes.Buffer
	runCmd.Stdout = &runStdout
	runCmd.Stderr = &runStderr
	
	err = runCmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("failed to run output executable: %v", err)
		}
	} else {
		exitCode = 0
	}
	
	t.Logf("Output binary exit code: %d, stdout: %q, stderr: %q", exitCode, runStdout.String(), runStderr.String())
	
	// Since 001_return_const.ax returns 42, we expect exit code 42
	if exitCode != 42 {
		t.Errorf("expected exit code 42, got %d", exitCode)
	}
}
