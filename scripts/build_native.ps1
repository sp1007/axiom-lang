# scripts/build_native.ps1
# Fast native daily-driver rebuild (BUG#51 unlocked this: a native-built compiler
# self-builds the 790-fn frontend in ~8s vs ~2h50m for the stage0/gcc-built one).
#
# Default (fast, ~20s): regenerate the concatenated frontend, then use the EXISTING
# native driver (bin/axc_native.exe) to self-build a new one, self-build once more,
# and promote only if the two are SHA-identical (self-host fixpoint). This is the
# gate for backend/linker changes together with scripts/regression_repros.sh.
#
# -Bootstrap (slow, ~3h): cold start with no trusted native driver. Runs
# rebuild_stage1.ps1 (stage0 -> gcc-built stage1, correct but slow) and uses that
# to mint the first native driver, then proceeds as above.
#
# NOTE: bin/*.exe are gitignored; this script (re)produces them from committed
# sources. The stage0/gcc chain (rebuild_stage1.ps1) remains the trusted root.

param([switch]$Bootstrap)

$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

$concat = "bootstrap/stage1/tmp_concatenated_air.ax"

# 1. Regenerate the concatenated frontend from the current (committed) sources.
#    Reuse rebuild_stage1.ps1's concat step by dot-sourcing its function only would
#    also rebuild the gcc stage1; instead inline the same file list + concat.
Write-Host "[native] Regenerating $concat ..." -ForegroundColor Green
& "$PSScriptRoot\regen_concat.ps1"
if (-not (Test-Path $concat)) { Write-Host "[native] concat FAILED" -ForegroundColor Red; exit 1 }

# 2. Pick the builder. Prefer the existing fast native driver.
#    cmd /c needs backslash program paths; keep a bs-variant for invocation.
$builder = "bin/axc_native.exe"
$builderBs = "bin\axc_native.exe"
if ($Bootstrap -or -not (Test-Path $builder)) {
    Write-Host "[native] Bootstrap: building gcc stage1 (slow, correct root) ..." -ForegroundColor Yellow
    & "$PSScriptRoot\rebuild_stage1.ps1"
    if (-not (Test-Path "bin/axc_stage1.exe")) { Write-Host "[native] stage1 FAILED" -ForegroundColor Red; exit 1 }
    Write-Host "[native] Minting first native driver via gcc stage1 (this is the slow ~3h build) ..." -ForegroundColor Yellow
    cmd /c "bin\axc_stage1.exe build $concat -o bin\axc_native.exe -self-link -O1 > compiler_native_boot.log 2>&1"
    $builderBs = "bin\axc_native.exe"
    if (-not (Test-Path $builder)) { Write-Host "[native] native mint FAILED -- see compiler_native_boot.log" -ForegroundColor Red; exit 1 }
}

# 3. Self-build twice and check the SHA fixpoint.
Write-Host "[native] Self-build n1 (via $builder) ..." -ForegroundColor Green
cmd /c "$builderBs build $concat -o bin\axc_native_n1.exe -self-link -O1 > compiler_native_n1.log 2>&1"
if (-not (Test-Path "bin/axc_native_n1.exe")) { Write-Host "[native] n1 FAILED -- see compiler_native_n1.log" -ForegroundColor Red; exit 1 }

Write-Host "[native] Self-build n2 (via n1) ..." -ForegroundColor Green
cmd /c "bin\axc_native_n1.exe build $concat -o bin\axc_native_n2.exe -self-link -O1 > compiler_native_n2.log 2>&1"
if (-not (Test-Path "bin/axc_native_n2.exe")) { Write-Host "[native] n2 FAILED -- see compiler_native_n2.log" -ForegroundColor Red; exit 1 }

$h1 = (Get-FileHash bin/axc_native_n1.exe -Algorithm SHA256).Hash
$h2 = (Get-FileHash bin/axc_native_n2.exe -Algorithm SHA256).Hash
if ($h1 -ne $h2) {
    Write-Host "[native] NO FIXPOINT (n1 != n2) -- NOT promoting. n1=$h1 n2=$h2" -ForegroundColor Red
    exit 1
}

# 4. Promote.
Copy-Item bin/axc_native_n2.exe bin/axc_native.exe -Force
$len = (Get-Item bin/axc_native.exe).Length
Remove-Item bin/axc_native_n1.exe, bin/axc_native_n2.exe -ErrorAction SilentlyContinue
Write-Host "[native] FIXPOINT OK -- promoted bin/axc_native.exe ($len bytes, sha=$h1)" -ForegroundColor Green
Write-Host "[native] Gate next: AXC=bin/axc_native.exe bash scripts/regression_repros.sh" -ForegroundColor Cyan
