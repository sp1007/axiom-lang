# scripts/fast_native_cycle.ps1
# Fast development build loop: Stage 2 + Stage 3 via native backend only.
# Skips the slow Stage 1 GCC build. Requires axc_stage1.exe to already exist.
# Use this for iterative development; run native_triple_build.ps1 only when
# axc_stage1.exe needs to be rebuilt from scratch.

$ErrorActionPreference = "Continue"

$root = Resolve-Path "$PSScriptRoot\.."
cd $root

Write-Host "=== AXIOM Fast Native Cycle (Stage 2 + Stage 3) ===" -ForegroundColor Cyan

if (-not (Test-Path bin/axc_stage1.exe)) {
    Write-Error "bin/axc_stage1.exe not found. Run scripts/native_triple_build.ps1 first."
    exit 1
}

# --- Stage 2: axc_stage1 --native--> axc_stage2_native ---
Write-Host "[Stage 2] Building Stage 2 via native backend..." -ForegroundColor Green
if (Test-Path bin/axc_stage2_native.exe) { Remove-Item bin/axc_stage2_native.exe }

$sw = [System.Diagnostics.Stopwatch]::StartNew()
cmd /c "bin\axc_stage1.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage2_native.exe -self-link -O1 > compiler_stage2_native.log 2>&1"
$sw.Stop()

if (-not (Test-Path bin/axc_stage2_native.exe)) {
    Write-Error "Stage 2 failed! Check compiler_stage2_native.log"
    Get-Content compiler_stage2_native.log | Select-Object -Last 20
    exit 1
}
Write-Host ("[Stage 2] OK in {0:F1}s -> bin/axc_stage2_native.exe" -f $sw.Elapsed.TotalSeconds) -ForegroundColor Green

# --- Stage 3: axc_stage2 --native--> axc_stage3_native ---
Write-Host "[Stage 3] Building Stage 3 via native backend..." -ForegroundColor Green
if (Test-Path bin/axc_stage3_native.exe) { Remove-Item bin/axc_stage3_native.exe }

$sw = [System.Diagnostics.Stopwatch]::StartNew()
cmd /c "bin\axc_stage2_native.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage3_native.exe -self-link -O1 > compiler_stage3_native.log 2>&1"
$sw.Stop()

if (-not (Test-Path bin/axc_stage3_native.exe)) {
    Write-Error "Stage 3 failed! Check compiler_stage3_native.log"
    Get-Content compiler_stage3_native.log | Select-Object -Last 20
    exit 1
}
Write-Host ("[Stage 3] OK in {0:F1}s -> bin/axc_stage3_native.exe" -f $sw.Elapsed.TotalSeconds) -ForegroundColor Green

# --- Reproducible Build Verification ---
Write-Host "[Verify] SHA-256 comparison..." -ForegroundColor Green
$hash2 = (Get-FileHash bin/axc_stage2_native.exe -Algorithm SHA256).Hash
$hash3 = (Get-FileHash bin/axc_stage3_native.exe -Algorithm SHA256).Hash
Write-Host "  Stage 2: $hash2" -ForegroundColor Cyan
Write-Host "  Stage 3: $hash3" -ForegroundColor Cyan

if ($hash2 -eq $hash3) {
    Write-Host "SUCCESS: Stage 2 == Stage 3 (100% bit-identical)" -ForegroundColor Green
    exit 0
} else {
    Write-Error "FAILURE: Stage 2 and Stage 3 differ - non-deterministic codegen!"
    exit 1
}
