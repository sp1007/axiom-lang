# M6 measurement protocol: compare compilers over a DISTRIBUTION OF LAYOUTS, not one binary.
#
# WHY THIS EXISTS (measured 2026-07-30). The layout channel is REAL, it is LARGE, and it is
# SHAPE-DEPENDENT. Two experiments, both on identical instruction streams shifted only by upstream
# code size, and they disagree in magnitude -- which is itself the point.
#
# EXPERIMENT A (hot loop inline in `main`, bin/bench/align):
#     loop body @ mod64=54  ->  134.8 ms   (133.8 / 134.9 / 135.1 / 135.4)
#     loop body @ mod64=38  ->  153.1 ms   (152.0 / 153.5 / 153.8)
#     loop body @ mod64= 6  ->  167.3 ms
#   Clean separation, ~1.2% scatter WITHIN a layout group, and a 24% spread ACROSS groups.
#
# EXPERIMENT B (same loop moved into its own function `hot()`, this script's shape):
#     mod64=53 -> 58.8    mod64=5 -> 57.8    mod64=21 -> 57.65
#   Only a 4% spread, and it does NOT separate cleanly by address -- one binary at mod64=5 gave
#   59.3 while three others at the SAME address gave 57.0/57.4/57.5.
#
# ⚠️ So do NOT quote a single "layout spread" figure. In A the layout term dwarfs every peephole
# win ever measured on this project; in B it is comparable to run-to-run noise. What is established
# is that a single-binary A/B between two compiler versions can be dominated by which layout each
# happened to draw -- which is how peephole 1e "cost fib 5%", how 1f "cost callloop 7%", and how
# xorshift once got 4.6% slower from a change that REMOVED instructions. Three phantoms, one
# channel. This script measures the term instead of pretending it is absent.
#
# ⛔ IT IS NOT FIXABLE BY ALIGNING ANYTHING, and that is settled rather than suspected. In A the
# FASTEST layout is the LEAST aligned (mod64=54) and the SLOWEST is the nearly-64-aligned one
# (mod64=6); in B the ordering by alignment is different again. An alignment pass must pick a
# boundary, and the optimum is not at a boundary and not even consistent between shapes. The
# loop-header-alignment hypothesis is REFUTED here, before any compiler code was written for it --
# 16-byte FUNCTION-ENTRY alignment (7ac52f5) is still worth keeping, since it removes the mod-16
# component, but there is nothing further to align toward.
#
# ⇒ The remedy is the PROTOCOL, not the compiler. Measure each shape at N controlled layout offsets
# and compare MEDIANS between compiler versions, reporting the spread alongside so a delta smaller
# than the spread is never read as a result. Zero compiler risk, and it is what actually blocks the
# M6-codegen milestone: a gate whose input carries an unquantified layout term cannot be passed or
# failed honestly.
#
# The offsets are produced by a `padfn` of varying body length called once from main -- once, so
# dead-function elimination cannot prune it, and outside the hot loop so it does not perturb the
# measured work. Function-entry alignment means each extra chunk shifts the entry by a multiple of
# 16, which is exactly the axis that matters (mod 16 is already fixed by the alignment pass).
#
# USAGE
#   scripts\perf_layout_dist.ps1 -Compiler bin\axc_native.exe -Shape callloop
#   scripts\perf_layout_dist.ps1 -Compiler bin\axc_pre1f.exe  -Shape callloop
# then compare the two MEDIAN lines. A difference smaller than the reported spread is NOT a result.
param(
    [string]$Compiler = "bin\axc_native.exe",
    [string]$Shape = "callloop",
    [int]$Variants = 8,
    [int]$BestOf = 9,
    [string]$Opt = "-O3"
)
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

# Hot-loop bodies, matching scripts/perf_suite.ps1 so numbers stay comparable to the suite.
$bodies = @{}
$bodies["callloop"] = @"
fn work(a: i64, b: i64, c: i64) -> i64:
    return a + b * 2 + c * 3

fn hot() -> i64:
    mut acc: i64 = 1
    mut i: i64 = 0
    while i < 60000000:
        acc = work(acc, i, 7) & 1048575
        i = i + 1
    return acc & 255
"@
$bodies["xorshift"] = @"
fn hot() -> i64:
    mut x: i64 = 88172645463325252
    mut i: i64 = 0
    while i < 200000000:
        x = x ^ (x * 8192)
        x = x ^ (x / 131072)
        x = x ^ (x * 4194304)
        i = i + 1
    return x & 255
"@
$bodies["fib"] = @"
fn fib(n: i64) -> i64:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

fn hot() -> i64:
    return fib(32) % 256
"@

if (-not $bodies.ContainsKey($Shape)) {
    Write-Host "unknown shape '$Shape'. known: $($bodies.Keys -join ', ')"
    exit 1
}
if (-not (Test-Path $Compiler)) { Write-Host "compiler not found: $Compiler"; exit 1 }

$out = Join-Path $root "bin\bench\layout_$Shape"
New-Item -ItemType Directory -Force $out | Out-Null

$built = @()
for ($k = 0; $k -lt $Variants; $k++) {
    $pad = ""
    for ($j = 0; $j -lt ($k + 1); $j++) { $pad += "    q = q + $($j + 1)`n" }
    @"
fn padfn(q0: i64) -> i64:
    mut q: i64 = q0
$pad    return q

$($bodies[$Shape])

fn main() -> i64:
    mut g: i64 = 0
    if padfn(0) == 999999:
        g = 1
    return hot() + g
"@ | Out-File -Encoding ascii (Join-Path $out "v$k.ax")
    & $Compiler build (Join-Path $out "v$k.ax") -o (Join-Path $out "v$k.exe") -self-link $Opt > $null 2>&1
    if (Test-Path (Join-Path $out "v$k.exe")) { $built += $k }
}
if ($built.Count -eq 0) { Write-Host "no variant built"; exit 1 }

# Startup floor, so a small workload is never reported as if it were all compute.
@"
fn main() -> i64:
    return 0
"@ | Out-File -Encoding ascii (Join-Path $out "nop.ax")
& $Compiler build (Join-Path $out "nop.ax") -o (Join-Path $out "nop.exe") -self-link -O1 > $null 2>&1

function BestOfN($exe, $n) {
    $b = [double]::MaxValue
    for ($i = 0; $i -lt $n; $i++) {
        $t = Measure-Command { & $exe | Out-Null }
        if ($t.TotalMilliseconds -lt $b) { $b = $t.TotalMilliseconds }
    }
    return $b
}

$floor = 0.0
if (Test-Path (Join-Path $out "nop.exe")) { $floor = BestOfN (Join-Path $out "nop.exe") $BestOf }

Write-Host ("=== layout distribution: {0}  shape={1}  {2}  best-of-{3} ===" -f $Compiler, $Shape, $Opt, $BestOf)
Write-Host ("startup floor: {0:F1} ms (subtracted below)" -f $floor)

$samples = @()
foreach ($k in $built) {
    $ms = (BestOfN (Join-Path $out "v$k.exe") $BestOf) - $floor
    $sz = (Get-Item (Join-Path $out "v$k.exe")).Length
    $samples += $ms
    Write-Host ("  v{0}  exe={1,7}B  {2,8:F1} ms" -f $k, $sz, $ms)
}

$sorted = $samples | Sort-Object
$n = $sorted.Count
$median = if ($n % 2 -eq 1) { $sorted[[int](($n - 1) / 2)] } else { ($sorted[$n / 2 - 1] + $sorted[$n / 2]) / 2.0 }
$min = $sorted[0]; $max = $sorted[$n - 1]
$spread = 0.0
if ($min -gt 0) { $spread = ($max - $min) / $min * 100.0 }
Write-Host ("MEDIAN {0:F1} ms   min {1:F1}   max {2:F1}   LAYOUT SPREAD {3:F1}%  (n={4})" -f $median, $min, $max, $spread, $n)
Write-Host "Compare MEDIANs between compilers. A delta below the layout spread is not a result."
