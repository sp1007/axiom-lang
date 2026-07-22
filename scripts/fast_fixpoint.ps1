# scripts/fast_fixpoint.ps1
# FAST native-backend fixpoint gate for iterative backend work.
# Seeds from the daily-driver bin/axc_native.exe (native-built, ~4.5s self-build)
# instead of the slow gcc bin/axc_stage1.exe (1200x slower). Chain:
#   axc_native (seed) --build NEW source--> A
#   A               --build NEW source--> B
# If SHA-256(A) == SHA-256(B) the new backend reproduces itself => fixpoint.
# Total ~9s vs. hours for the gcc path. Run scripts/fast_native_cycle.ps1 (or the
# full triple build) only for a final belt-and-suspenders check before shipping.
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
cd $root

if (-not (Test-Path bin/axc_native.exe)) {
    Write-Error "bin/axc_native.exe not found (daily driver). Build it first."
    exit 1
}

Write-Host "=== AXIOM Fast Fixpoint (axc_native seed) ===" -ForegroundColor Cyan
$src = "bootstrap/stage1/tmp_concatenated_air.ax"

# Delete the hop outputs BEFORE building. Without this the Test-Path guards below
# are satisfied by LEFTOVERS from the previous run: a build that fails outright
# (parse error, OOM) leaves the old fpA/fpB in place, every check passes, and the
# script prints "SUCCESS: A == B" for a compiler that was never built. The tell is
# subtle -- A and B agree because they are the SAME STALE FILE -- and it has now
# cost two separate sessions real time. Deleting first turns that into an honest
# "hop1 (A) failed".
Remove-Item bin/axc_fpA.exe,bin/axc_fpB.exe,bin/axc_fpC.exe -ErrorAction SilentlyContinue

# Regenerate the concatenated source from clean per-file modules.
& "$PSScriptRoot\regen_concat.ps1" | Out-Null

$sw=[System.Diagnostics.Stopwatch]::StartNew()
& bin/axc_native.exe build $src -o bin/axc_fpA.exe -self-link -O1 > fpA.log 2>&1
$sw.Stop()
if (-not (Test-Path bin/axc_fpA.exe)) { Write-Error "hop1 (A) failed"; Get-Content fpA.log -Tail 20; exit 1 }
Write-Host ("[A] {0:F1}s" -f $sw.Elapsed.TotalSeconds) -ForegroundColor Green

$sw=[System.Diagnostics.Stopwatch]::StartNew()
& bin/axc_fpA.exe build $src -o bin/axc_fpB.exe -self-link -O1 > fpB.log 2>&1
$sw.Stop()
if (-not (Test-Path bin/axc_fpB.exe)) { Write-Error "hop2 (B) failed"; Get-Content fpB.log -Tail 20; exit 1 }
Write-Host ("[B] {0:F1}s" -f $sw.Elapsed.TotalSeconds) -ForegroundColor Green

$seed=(Get-FileHash bin/axc_native.exe -Algorithm SHA256).Hash
$a=(Get-FileHash bin/axc_fpA.exe -Algorithm SHA256).Hash
$b=(Get-FileHash bin/axc_fpB.exe -Algorithm SHA256).Hash
Write-Host "  A: $a" -ForegroundColor Cyan
Write-Host "  B: $b" -ForegroundColor Cyan
if ($a -eq $seed) { Write-Host "  note: A == seed (source is inert vs the driver, or nothing was rebuilt)" -ForegroundColor DarkYellow }
if ($a -eq $b) { Write-Host "SUCCESS: A == B (fixpoint)" -ForegroundColor Green; exit 0 }
else { Write-Error "FAILURE: A != B (non-deterministic / broken backend)"; exit 1 }
