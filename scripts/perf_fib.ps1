# scripts/perf_fib.ps1
# M6 performance gate: Fib(40) wall-time, AXIOM native vs clang/gcc.
# The M6 milestone target is AXIOM within 5% of clang -O2. This harness makes the gap
# REPRODUCIBLE so optimization work is measured, not guessed (CLAUDE.md §10: profile
# before optimize; perf work must be benchmarked + reversible).
#
# FAIR comparison: AXIOM `i64` vs C `int64_t` (NOT `long` -- `long` is 32-bit on Windows,
# which would be an unfair 32-vs-64-bit race). Same recursive algorithm, same width.
# Best-of-N wall time (min = least noise). Correctness pinned: all must exit 203
# (fib(40)=102334155; 102334155 % 256 = 203).
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root
$bench = Join-Path $root "bin\bench"
New-Item -ItemType Directory -Force $bench | Out-Null

# --- sources ---
$axsrc = Join-Path $bench "fib.ax"
@"
fn fib(n: i64) -> i64:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)
fn main() -> i64:
    return fib(40) % 256
"@ | Out-File -Encoding ascii $axsrc
$csrc = Join-Path $bench "fib64.c"
@"
#include <stdint.h>
int64_t fib(int64_t n){ if(n<2) return n; return fib(n-1)+fib(n-2); }
int main(){ return (int)(fib(40) % 256); }
"@ | Out-File -Encoding ascii $csrc

$axc = Join-Path $root "bin\axc_native.exe"
$runs = 7
$oracle = 203

function Build-And-Time($name, $exe, $buildBlock) {
    & $buildBlock 2>&1 | Out-Null
    if (-not (Test-Path $exe)) { Write-Host ("{0,-12} BUILD-FAIL" -f $name) -ForegroundColor Red; return $null }
    & $exe | Out-Null; $ec = $LASTEXITCODE
    $best = [double]::MaxValue
    for ($i=0; $i -lt $runs; $i++) {
        $t = Measure-Command { & $exe | Out-Null }
        if ($t.TotalMilliseconds -lt $best) { $best = $t.TotalMilliseconds }
    }
    $ok = if ($ec -eq $oracle) { "ok" } else { "BAD($ec)" }
    Write-Host ("{0,-12} {1,8:F1} ms  exit={2}" -f $name, $best, $ok)
    return @{ ms=$best; ec=$ec }
}

Write-Host "=== M6 Fib(40) perf gate (i64 vs int64_t) ===" -ForegroundColor Cyan
$axo2 = Build-And-Time "axiom -O2" (Join-Path $bench "fib_ax_o2.exe") { & $axc build $axsrc -o (Join-Path $bench "fib_ax_o2.exe") -O2 }
$axo3 = Build-And-Time "axiom -O3" (Join-Path $bench "fib_ax_o3.exe") { & $axc build $axsrc -o (Join-Path $bench "fib_ax_o3.exe") -O3 }
$cl   = Build-And-Time "clang -O2" (Join-Path $bench "fib_clang.exe") { clang -O2 $csrc -o (Join-Path $bench "fib_clang.exe") }
$gc   = Build-And-Time "gcc -O2"   (Join-Path $bench "fib_gcc.exe")   { gcc   -O2 $csrc -o (Join-Path $bench "fib_gcc.exe") }

if ($axo3 -and $cl) {
    $ratio = $axo3.ms / $cl.ms
    Write-Host ""
    Write-Host ("AXIOM -O3 / clang -O2 = {0:F2}x  ({1:+0.0;-0.0}% vs clang)" -f $ratio, (($ratio-1)*100)) -ForegroundColor Yellow
    if ($gc) { Write-Host ("AXIOM -O3 / gcc   -O2 = {0:F2}x" -f ($axo3.ms/$gc.ms)) }
    # M6 gate: within 5% of clang.
    if ($ratio -le 1.05) { Write-Host "M6_PERF_OK (within 5% of clang)" -ForegroundColor Green }
    else { Write-Host ("M6_PERF_GAP ({0:F2}x -- target <=1.05x)" -f $ratio) -ForegroundColor DarkYellow }
}
