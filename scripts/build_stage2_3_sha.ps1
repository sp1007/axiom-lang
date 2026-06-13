# scripts/build_stage2_3_sha.ps1
# Phase 3 of next-step-14: using the freshly-rebuilt (fixed) bin/axc_stage1.exe,
# build Stage 2 native, then Stage 3 native, then compare SHA-256.
# Assumes bin/axc_stage1.exe already contains the str codegen fixes
# (run scripts/rebuild_stage1.ps1 first). ASCII only for PS 5.1 compatibility.

$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

$src = "bootstrap\stage1\tmp_concatenated_air.ax"

Write-Host "[stage2] Building bin/axc_stage2_native.exe from fixed stage1 ..." -ForegroundColor Green
if (Test-Path bin\axc_stage2_native.exe) { Remove-Item bin\axc_stage2_native.exe }
cmd /c "bin\axc_stage1.exe build $src -o bin\axc_stage2_native.exe -self-link -O1 > compiler_stage2_native.log 2>&1"
if (-not (Test-Path bin\axc_stage2_native.exe)) {
    Write-Host "[stage2] FAILED -- see compiler_stage2_native.log" -ForegroundColor Red
    Get-Content compiler_stage2_native.log -Tail 30
    exit 1
}
$s2 = (Get-Item bin\axc_stage2_native.exe).Length
Write-Host "[stage2] OK: $s2 bytes" -ForegroundColor Green

Write-Host "[stage3] Building bin/axc_stage3_native.exe from stage2 ..." -ForegroundColor Green
if (Test-Path bin\axc_stage3_native.exe) { Remove-Item bin\axc_stage3_native.exe }
cmd /c "bin\axc_stage2_native.exe build $src -o bin\axc_stage3_native.exe -self-link -O1 > compiler_stage3_native.log 2>&1"
if (-not (Test-Path bin\axc_stage3_native.exe)) {
    Write-Host "[stage3] FAILED -- see compiler_stage3_native.log" -ForegroundColor Red
    Get-Content compiler_stage3_native.log -Tail 30
    exit 1
}
$s3 = (Get-Item bin\axc_stage3_native.exe).Length
Write-Host "[stage3] OK: $s3 bytes" -ForegroundColor Green

$hash2 = (Get-FileHash bin\axc_stage2_native.exe -Algorithm SHA256).Hash
$hash3 = (Get-FileHash bin\axc_stage3_native.exe -Algorithm SHA256).Hash
Write-Host "Stage2 SHA-256: $hash2" -ForegroundColor Cyan
Write-Host "Stage3 SHA-256: $hash3" -ForegroundColor Cyan

if ($hash2 -eq $hash3) {
    Write-Host "SUCCESS: Stage2 == Stage3, bit-identical self-hosting verified." -ForegroundColor Green
    exit 0
} else {
    Write-Host "MISMATCH: Stage2 != Stage3." -ForegroundColor Red
    exit 1
}
