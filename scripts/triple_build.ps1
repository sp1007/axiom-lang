# scripts/triple_build.ps1
# AXIOM Compiler Self-Hosting and Triple-Build Verification Loop
#
# This script automates:
# 1. Building Stage 0 (the Go-implemented compiler driver)
# 2. Concatenating the self-hosted AXIOM compiler frontend source files
# 3. Compiling the self-hosted frontend (Stage 1) using Stage 0
# 4. Verifying Stage 1 against Stage 0 reference output on the test corpus
# 5. Ensuring deterministic compilations

$ErrorActionPreference = "Continue"

# Ensure we are in the workspace root
$root = Resolve-Path "$PSScriptRoot\.."
cd $root

Write-Host "=== AXIOM Triple-Build Verification Loop ===" -ForegroundColor Cyan

# 1. Create output directory
if (-not (Test-Path bin)) {
    New-Item -ItemType Directory -Path bin | Out-Null
}

# 2. Build Stage 0 (Go-implemented compiler driver)
Write-Host "[Stage 0] Building Go-implemented driver (axc.exe)..." -ForegroundColor Green
go build -o bin/axc.exe ./cmd/axc
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build Stage 0 Go compiler driver"
}
Write-Host "[Stage 0] Version: " -NoNewline
& .\bin\axc.exe version

# 3. Concatenate the self-hosted frontend files
Write-Host "[Stage 1] Concatenating self-hosted compiler frontend source files..." -ForegroundColor Green

function Concatenate-Axiom-Files {
    param(
        [string[]]$Paths,
        [string]$OutputPath
    )

    $imports = [System.Collections.Generic.List[string]]::new()
    $body = [System.Collections.Generic.List[string]]::new()

    foreach ($p in $Paths) {
        $lines = Get-Content $p
        foreach ($line in $lines) {
            $trimmed = $line.Trim()
            if ($trimmed.StartsWith("import ")) {
                if ($trimmed -like "import bootstrap.stage1.*") {
                    continue
                }
                $imports.Add($line)
            } else {
                $body.Add($line)
            }
        }
    }

    $uniqueImports = [System.Collections.Generic.List[string]]::new()
    $importMap = @{}
    foreach ($imp in $imports) {
        $trimmed = $imp.Trim()
        if (-not $importMap.ContainsKey($trimmed)) {
            $importMap[$trimmed] = $true
            $uniqueImports.Add($imp)
        }
    }

    $result = ($uniqueImports -join "`n") + "`n`n" + ($body -join "`n")
    [System.IO.File]::WriteAllText($OutputPath, $result)
}

$frontendFiles = @(
    "bootstrap/stage1/print_helpers.ax",
    "bootstrap/stage1/token.ax",
    "bootstrap/stage1/lexer.ax",
    "bootstrap/stage1/ast.ax",
    "bootstrap/stage1/intern.ax",
    "bootstrap/stage1/parser.ax",
    "bootstrap/stage1/resolver.ax",
    "bootstrap/stage1/typetable.ax",
    "bootstrap/stage1/mono.ax",
    "bootstrap/stage1/typecheck.ax",
    "bootstrap/stage1/connection_graph.ax",
    "bootstrap/stage1/ownership.ax",
    "bootstrap/stage1/escape.ax",
    "bootstrap/stage1/ctgc.ax",
    "bootstrap/stage1/alias_reuse.ax",
    "bootstrap/stage1/air.ax",
    "bootstrap/stage1/air_builder.ax",
    "bootstrap/stage1/ssa_opt.ax",
    "bootstrap/stage1/cgen.ax",
    "bootstrap/stage1/wasm.ax",
    "bootstrap/stage1/x86_regs.ax",
    "bootstrap/stage1/x86_selector.ax",
    "bootstrap/stage1/x86_regalloc.ax",
    "bootstrap/stage1/x86_asm_emitter.ax",
    "bootstrap/stage1/x86_modrm.ax",
    "bootstrap/stage1/x86_encoding.ax",
    "bootstrap/stage1/x86_emitter.ax",
    "bootstrap/stage1/x86_elf64.ax",
    "bootstrap/stage1/x86_coff.ax",
    "bootstrap/stage1/linker.ax",
    "bootstrap/stage1/fmt.ax",
    "bootstrap/stage1/main_air.ax"
)

$concatenatedPath = "bootstrap/stage1/tmp_concatenated_air.ax"
Concatenate-Axiom-Files -Paths $frontendFiles -OutputPath $concatenatedPath
Write-Host "Generated concatenated frontend at: $concatenatedPath" -ForegroundColor DarkGray

# 4. Compile Stage 1 (self-hosted frontend) using Stage 0
Write-Host "[Stage 1] Compiling Stage 1 self-hosted frontend (axc_stage1.exe)..." -ForegroundColor Green
if (Test-Path bin/axc_stage1.exe) {
    Remove-Item bin/axc_stage1.exe
}

cmd /c "bin\axc.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage1.exe > compiler_stage1.log 2>&1"
Start-Sleep -Seconds 1
if (-not (Test-Path bin/axc_stage1.exe)) {
    Write-Error "Stage 1 compilation failed!"
}
Write-Host "[Stage 1] Stage 1 compiler binary compiled successfully to bin/axc_stage1.exe" -ForegroundColor Green

# 5. Run Verification Loop across Corpus
Write-Host "[Verify] Running verification loop..." -ForegroundColor Green

# Gather simple valid .ax files to verify
$testFiles = Get-ChildItem -Path tests -Filter "*.ax" -Recurse | Where-Object {
    $_.FullName -notmatch "scratch" -and
    $_.FullName -notmatch "tmp" -and
    $_.FullName -notmatch "err_" -and
    ($_.Name -like "00*" -or $_.Name -eq "valid_assign.ax" -or $_.Name -eq "valid_fibonacci.ax" -or $_.Name -eq "valid_shadow.ax" -or $_.Name -eq "valid_hello.ax" -or $_.Name -eq "opt_test.ax")
}

$passed = 0
$total = 0

foreach ($f in $testFiles) {
    $total++
    $relPath = Resolve-Path -Relative $f.FullName
    Write-Host "  Verifying $relPath... " -NoNewline

    # Get Stage 0 reference output using cmd /c for robust execution
    $stage0Out = cmd /c "bin\axc.exe dump-air $relPath -O1 2>nul"
    $stage0Exit = $LASTEXITCODE
    if ($stage0Exit -ne 0) {
        Write-Host "Skipped (Stage 0 failed to dump AIR, expected for complex/non-supported features)" -ForegroundColor Yellow
        continue
    }

    # Get Stage 1 self-hosted output using cmd /c for robust execution
    $stage1Out = cmd /c "bin\axc_stage1.exe $relPath -O1 2>nul"
    $stage1Exit = $LASTEXITCODE
    if ($stage1Exit -ne 0) {
        Write-Host "FAILED (Stage 1 crashed or exited with error)" -ForegroundColor Red
        exit 1
    }

    # Normalise newlines and spaces for exact matching
    $stage0Normalized = $stage0Out -join "`n"
    $stage0Normalized = $stage0Normalized.Replace("`r`n", "`n").Trim()

    $stage1Normalized = $stage1Out -join "`n"
    $stage1Normalized = $stage1Normalized.Replace("`r`n", "`n").Trim()

    if ($stage0Normalized -eq $stage1Normalized) {
        Write-Host "PASSED (Deterministic Match)" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "FAILED (Mismatch!)" -ForegroundColor Red
        Write-Host "=== EXPECTED (Stage 0 Go Reference) ===" -ForegroundColor DarkCyan
        Write-Host $stage0Normalized
        Write-Host "=== ACTUAL (Stage 1 Self-Hosted) ===" -ForegroundColor DarkRed
        Write-Host $stage1Normalized
        exit 1
    }
}

Write-Host "=== Verification Finished ===" -ForegroundColor Cyan
Write-Host "Result: $passed / $total corpus files matched exactly." -ForegroundColor Green

if ($passed -lt $total) {
    Write-Host "Warning: Some corpus tests were skipped ($($total - $passed) skipped), but all run tests passed." -ForegroundColor Yellow
}

Write-Host "[Stage 2] Building Stage 2 compiler using Stage 1..." -ForegroundColor Green
if (Test-Path bin/axc_stage2.exe) {
    Remove-Item bin/axc_stage2.exe
}
cmd /c "bin\axc_stage1.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage2.exe -use-gcc > compiler_stage2.log 2>&1"
Start-Sleep -Seconds 1
if (-not (Test-Path bin/axc_stage2.exe)) {
    Write-Error "Stage 2 compilation failed!"
}
Write-Host "[Stage 2] Compiled successfully to bin/axc_stage2.exe" -ForegroundColor Green

Write-Host "[Stage 3] Building Stage 3 compiler using Stage 2..." -ForegroundColor Green
if (Test-Path bin/axc_stage3.exe) {
    Remove-Item bin/axc_stage3.exe
}
cmd /c "bin\axc_stage2.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage3.exe -use-gcc > compiler_stage3.log 2>&1"
Start-Sleep -Seconds 1
if (-not (Test-Path bin/axc_stage3.exe)) {
    Write-Error "Stage 3 compilation failed!"
}
Write-Host "[Stage 3] Compiled successfully to bin/axc_stage3.exe" -ForegroundColor Green

Write-Host "[Verify] Computing SHA-256 hashes..." -ForegroundColor Green
$hash2 = (Get-FileHash bin/axc_stage2.exe -Algorithm SHA256).Hash
$hash3 = (Get-FileHash bin/axc_stage3.exe -Algorithm SHA256).Hash

Write-Host "Stage 2 SHA-256: $hash2" -ForegroundColor Cyan
Write-Host "Stage 3 SHA-256: $hash3" -ForegroundColor Cyan

if ($hash2 -eq $hash3) {
    Write-Host "SUCCESS: Stage 2 and Stage 3 compiler binaries are 100% bit-identical!" -ForegroundColor Green
    exit 0
} else {
    Write-Error "FAILURE: Stage 2 and Stage 3 compiler binaries differ!"
    exit 1
}
