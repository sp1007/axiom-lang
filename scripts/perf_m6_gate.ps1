# The M6-codegen gate, read HONESTLY: median-over-layouts on BOTH sides of the ratio.
#
# The gate (user decision D1, 2026-07-29) is: AXIOM within 15% of a hand-written NASM floor in the
# SAME algorithmic shape. scripts/perf_suite.ps1 computes that ratio from ONE AXIOM binary and ONE
# asm binary, and on 2026-07-30 that was shown to be unsound: identical instruction streams shifted
# by multiples of 16 bytes measured 134.8 / 153.1 / 167.3 ms on one shape, and a single-binary A/B
# credited peephole 1f with a callloop regression whose true value is exactly zero (medians 143.5 vs
# 143.5). The FIXED asm-floor binary has itself been seen to swing 569 -> 514 ms between sessions,
# which is the same channel showing up on the denominator.
#
# A ratio of two single draws therefore carries two independent layout terms. This script removes
# both: it builds N layout variants of each side, takes the MEDIAN of each, and reports the ratio
# together with both spreads. A ratio whose distance from 1.15 is smaller than the spreads is NOT a
# verdict, and the script says so rather than printing a pass/fail.
#
# HOW THE LAYOUTS ARE VARIED, and why the two sides differ:
#   asm side   -- `times N nop` injected at the top of .text. Exact, byte-precise, and it shifts
#                 every subsequent address by N. This is the clean instrument.
#   AXIOM side -- a `padfn` of varying body length, called once from main so dead-function
#                 elimination cannot prune it and placed outside the hot loop so it does not
#                 perturb the measured work. Indirect, because there is no way to inject raw bytes
#                 through the compiler. Function-entry alignment (7ac52f5) means each increment
#                 moves the entry by a multiple of 16, which is the axis that matters -- the mod-16
#                 component is already fixed by that pass.
#
# ⛔ Do NOT "fix" the spread with a loop-alignment pass. Refuted 2026-07-30: the FASTEST layout
# measured was the LEAST aligned (mod64=54) and the SLOWEST was the nearly-64-aligned one (mod64=6),
# and the ordering by alignment differs between shapes. There is no boundary to align toward. See
# the header of scripts/perf_layout_dist.ps1.
#
# 2026-07-30: arrwalk is now in the DEFAULT $Shapes list (it was already implemented as a shape --
# commit 3ef26f0 -- but left out of the default array, and a stale "still has no distribution
# reading" TODO survived in knowledge/ docs after the fact landed; both are now corrected). Re-run
# at 10 layouts/side, best-of-9, to confirm the recorded number still reproduces:
#   arrwalk   1.092x +- 1.7% PASS by 5.1%  (vs. recorded 1.087x +- 0.5% PASS by 5.5% -- consistent)
#   xorshift  0.995x +- 0.8% PASS by 13.4% (control, unchanged -- confirms nothing else regressed)
# startup floor at time of that run: 10.8 ms.
param(
    [string]$Compiler = "bin\axc_native.exe",
    [string[]]$Shapes = @("callloop", "arrwalk", "fib", "xorshift"),
    [int]$Variants = 6,
    [int]$BestOf = 9,
    [string]$Opt = "-O3",
    [double]$Gate = 1.15
)
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

$nasm = $null
foreach ($cand in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $c = Get-Command $cand -ErrorAction SilentlyContinue
    if ($c) { $nasm = $c.Source; break }
}
if (-not $nasm) { Write-Host "nasm not found -- this script cannot compute the gate without a floor."; exit 1 }
if (-not (Test-Path $Compiler)) { Write-Host "compiler not found: $Compiler"; exit 1 }

# --- AXIOM hot bodies. `hot()` is the measured function; main only calls it. ---
$ax = @{}
$ax["callloop"] = @"
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
# arrwalk needs a module-level global, which sits in the shape body above `main` -- the template
# splices the body between padfn and main, so a global declaration there is module scope as usual.
# The fill loop (65536 iterations) is inside `hot()` exactly as in perf_suite.ps1 and is not the
# measurement; the walk is 40M dependent loads.
$ax["arrwalk"] = @"
mut tbl: [i64; 65536]

fn hot() -> i64:
    mut i: i64 = 0
    while i < 65536:
        tbl[i] = (i * 2654435761 + 12345) & 65535
        i = i + 1
    mut idx: i64 = 0
    mut acc: i64 = 0
    mut k: i64 = 0
    while k < 40000000:
        idx = tbl[idx]
        acc = acc + idx
        k = k + 1
    return acc & 255
"@
$ax["fib"] = @"
fn fib(n: i64) -> i64:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

fn hot() -> i64:
    return fib(40) % 256
"@
# VERBATIM from perf_suite.ps1 apart from the main->hot rename. Do NOT paraphrase these bodies.
# A first version of this script wrote the shifts as `x * 8192` / `x / 128` on i64 and measured
# 1.50x against the floor -- entirely a construction error, not a codegen result: `x / 128` is a
# SIGNED division rounding toward zero, which is not the floor's `shr`, so the two sides were no
# longer the same algorithm. The floor is only a floor if the program matches it exactly.
$ax["xorshift"] = @"
fn hot() -> i64:
    mut x: u64 = 88172645463325252
    mut i: u64 = 0
    while i < 120000000:
        x = x ^ (x << 13)
        x = x ^ (x >> 7)
        x = x ^ (x << 17)
        i = i + 1
    return (x & 255) as i64
"@

# --- NASM floors, copied verbatim from perf_suite.ps1 so the two scripts agree. ---
$asm = @{}
$asm["callloop"] = @"
default rel
section .text
global main
main:
    mov     rax, 1
    xor     rcx, rcx
    mov     r8, 60000000
.loop:
    lea     rax, [rax + rcx*2]
    add     rax, 21
    and     rax, 1048575
    inc     rcx
    cmp     rcx, r8
    jb      .loop
    and     rax, 255
    ret
"@
$asm["arrwalk"] = @"
default rel
section .bss
tbl:    resq 65536
section .text
global main
main:
    lea     r9, [tbl]
    mov     r10, 2654435761
    xor     rax, rax
.fill:
    mov     rdx, rax
    imul    rdx, r10
    add     rdx, 12345
    and     rdx, 65535
    mov     [r9 + rax*8], rdx
    inc     rax
    cmp     rax, 65536
    jb      .fill
    xor     rax, rax
    xor     rcx, rcx
    mov     r8, 40000000
.walk:
    mov     rax, [r9 + rax*8]
    add     rcx, rax
    dec     r8
    jnz     .walk
    mov     rax, rcx
    and     rax, 255
    ret
"@
$asm["fib"] = @"
default rel
section .text
global fib_hand
global main
fib_hand:
    cmp     rcx, 2
    jl      .base
    push    rbx
    push    rsi
    sub     rsp, 40
    mov     rbx, rcx
    dec     rcx
    call    fib_hand
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib_hand
    add     rax, rsi
    add     rsp, 40
    pop     rsi
    pop     rbx
    ret
.base:
    mov     rax, rcx
    ret
main:
    sub     rsp, 40
    mov     rcx, 40
    call    fib_hand
    and     rax, 255
    add     rsp, 40
    ret
"@
$asm["xorshift"] = @"
default rel
section .text
global main
main:
    mov     rax, 88172645463325252
    mov     rcx, 120000000
.loop:
    mov     rdx, rax
    shl     rdx, 13
    xor     rax, rdx
    mov     rdx, rax
    shr     rdx, 7
    xor     rax, rdx
    mov     rdx, rax
    shl     rdx, 17
    xor     rax, rdx
    dec     rcx
    jnz     .loop
    and     rax, 255
    ret
"@

function BestOfN($exe, $n) {
    & $exe | Out-Null              # warm the page cache; result discarded
    $b = [double]::MaxValue
    for ($i = 0; $i -lt $n; $i++) {
        $t = Measure-Command { & $exe | Out-Null }
        if ($t.TotalMilliseconds -lt $b) { $b = $t.TotalMilliseconds }
    }
    return $b
}
# MEAN, not median, and deliberately. The per-layout distributions are BIMODAL -- fib's asm floor
# alternates between ~460 ms and ~522 ms with nothing varying but nop padding -- and a median over a
# handful of bimodal samples is unstable: with n=6 split 4/2 versus 2/4 between clusters, the median
# jumps between clusters while the mean moves smoothly. The gate question is also better served by
# the mean: "what does an arbitrary build of this shape cost", not "what does the middle build cost".
function Mean($a) {
    if ($a.Count -eq 0) { return 0.0 }
    $s = 0.0
    foreach ($x in $a) { $s += $x }
    return $s / $a.Count
}
# Standard error of the mean. This is the uncertainty of the ESTIMATOR, which shrinks as 1/sqrt(n),
# and it is what a verdict must be compared against -- not the raw min-max spread, which describes
# the population and does not shrink no matter how many layouts are sampled. Using the raw spread
# made every shape INDETERMINATE by construction.
function Sem($a) {
    $n = $a.Count
    if ($n -lt 2) { return 0.0 }
    $m = Mean $a
    $ss = 0.0
    foreach ($x in $a) { $ss += ($x - $m) * ($x - $m) }
    return [Math]::Sqrt($ss / ($n - 1)) / [Math]::Sqrt($n)
}
function SpreadPct($a) {
    $s = $a | Sort-Object
    if ($s.Count -lt 2 -or $s[0] -le 0) { return 0.0 }
    return ($s[$s.Count - 1] - $s[0]) / $s[0] * 100.0
}

$out = Join-Path $root "bin\bench\m6gate"
New-Item -ItemType Directory -Force $out | Out-Null

Write-Host ("=== M6-codegen gate, mean over {0} layouts per side (best-of-{1}), 2-SE verdict ===" -f $Variants, $BestOf)
Write-Host ("compiler: {0}   opt: {1}   gate: <= {2}x of the NASM floor" -f $Compiler, $Opt, $Gate)

# Startup floor, measured once. Both sides pay it, but it must be removed BEFORE the ratio --
# otherwise a shape whose compute is small has its ratio dragged toward 1.0 by shared overhead.
@"
fn main() -> i64:
    return 0
"@ | Out-File -Encoding ascii "$out\nop.ax"
& $Compiler build "$out\nop.ax" -o "$out\nop.exe" -self-link -O1 > $null 2>&1
$floor = 0.0
if (Test-Path "$out\nop.exe") { $floor = BestOfN "$out\nop.exe" $BestOf }
Write-Host ("startup floor: {0:F1} ms (subtracted from every sample)" -f $floor)
Write-Host ""

foreach ($shape in $Shapes) {
    if (-not $ax.ContainsKey($shape)) { Write-Host "skip unknown shape $shape"; continue }

    # ---- AXIOM side ----
    $axs = @()
    for ($k = 0; $k -lt $Variants; $k++) {
        $pad = ""
        for ($j = 0; $j -lt ($k + 1); $j++) { $pad += "    q = q + $($j + 1)`n" }
        @"
fn padfn(q0: i64) -> i64:
    mut q: i64 = q0
$pad    return q

$($ax[$shape])

fn main() -> i64:
    mut g: i64 = 0
    if padfn(0) == 999999:
        g = 1
    return hot() + g
"@ | Out-File -Encoding ascii "$out\${shape}_ax$k.ax"
        & $Compiler build "$out\${shape}_ax$k.ax" -o "$out\${shape}_ax$k.exe" -self-link $Opt > $null 2>&1
        if (Test-Path "$out\${shape}_ax$k.exe") { $axs += ((BestOfN "$out\${shape}_ax$k.exe" $BestOf) - $floor) }
    }

    # ---- ASM floor side ----
    $asms = @()
    for ($k = 0; $k -lt $Variants; $k++) {
        $shift = $k * 16
        $body = $asm[$shape]
        # Inject the shift right after `section .text`, before any label, so every address moves.
        $inj = if ($shift -gt 0) { "section .text`n    times $shift nop" } else { "section .text" }
        $body = $body -replace "section \.text", $inj
        $body | Out-File -Encoding ascii "$out\${shape}_fl$k.asm"
        & $nasm -f win64 "$out\${shape}_fl$k.asm" -o "$out\${shape}_fl$k.obj" 2>$null
        if (Test-Path "$out\${shape}_fl$k.obj") {
            # gcc, matching perf_suite.ps1 -- it supplies the CRT entry that calls main, so the
            # floor binary pays the same startup as the AXIOM one and $floor cancels on both sides.
            & gcc "$out\${shape}_fl$k.obj" -o "$out\${shape}_fl$k.exe" 2>&1 | Out-Null
            if (Test-Path "$out\${shape}_fl$k.exe") { $asms += ((BestOfN "$out\${shape}_fl$k.exe" $BestOf) - $floor) }
        }
    }

    if ($axs.Count -eq 0 -or $asms.Count -eq 0) {
        Write-Host ("{0,-10} INCOMPLETE (axiom n={1}, floor n={2})" -f $shape, $axs.Count, $asms.Count)
        continue
    }

    $mAx = Mean $axs
    $mFl = Mean $asms
    $sAx = SpreadPct $axs
    $sFl = SpreadPct $asms
    $ratio = if ($mFl -gt 0) { $mAx / $mFl } else { 0.0 }

    Write-Host ("{0,-10} axiom mean {1,7:F1} ms (layout spread {2,4:F1}%)   floor mean {3,7:F1} ms (spread {4,4:F1}%)   ratio {5:F3}x" -f $shape, $mAx, $sAx, $mFl, $sFl, $ratio)
    Write-Host ("           axiom samples: {0}" -f (($axs | ForEach-Object { "{0:F1}" -f $_ }) -join " "))
    Write-Host ("           floor samples: {0}" -f (($asms | ForEach-Object { "{0:F1}" -f $_ }) -join " "))

    # Uncertainty of the RATIO, propagated from each side's standard error of the mean. A verdict
    # requires the gate to sit more than 2 SE away (~95%); inside that band the honest answer is
    # "not resolved at this many layouts", and the fix is more variants, not a louder claim.
    $relAx = if ($mAx -gt 0) { (Sem $axs) / $mAx } else { 0.0 }
    $relFl = if ($mFl -gt 0) { (Sem $asms) / $mFl } else { 0.0 }
    $relSE = [Math]::Sqrt($relAx * $relAx + $relFl * $relFl)
    $ciPct = 2.0 * $relSE * 100.0
    $margin = ($Gate - $ratio) / $Gate * 100.0

    Write-Host ("           ratio {0:F3}x +- {1:F1}% (2 SE, n={2}/{3})   gate {4}x" -f $ratio, $ciPct, $axs.Count, $asms.Count, $Gate)
    if ([Math]::Abs($margin) -lt $ciPct) {
        Write-Host ("           INDETERMINATE: {0:F1}% from the gate, inside the +-{1:F1}% confidence band. Add variants." -f [Math]::Abs($margin), $ciPct) -ForegroundColor Yellow
    } elseif ($ratio -le $Gate) {
        Write-Host ("           PASS by {0:F1}%, outside the +-{1:F1}% band" -f $margin, $ciPct) -ForegroundColor Green
    } else {
        Write-Host ("           MISS by {0:F1}%, outside the +-{1:F1}% band" -f [Math]::Abs($margin), $ciPct) -ForegroundColor Red
    }
    Write-Host ""
}
