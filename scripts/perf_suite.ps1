# scripts/perf_suite.ps1
# M6 performance suite: AXIOM native vs clang -O2 across FOUR code shapes.
#
# WHY THIS EXISTS (see knowledge/m6-perf-baseline.md): the M6 gate was measured on Fib(40)
# ALONE, which is recursion/call-dominated. Three optimization attempts (cmp-imm fusion, lea
# folding, allocator preference reorder) all came back neutral-or-worse on that single number,
# and the conclusion drawn was "the gap is structural, needs a full allocator rewrite". That
# conclusion is only actionable if we know WHICH shapes carry the tax. A single benchmark
# cannot distinguish "our register allocator is bad everywhere" from "our CALL sequence is bad".
#
# Each shape isolates one axis:
#   fib       recursion + call-heavy      (call ABI, prologue, values live across calls)
#   xorshift  serial ALU, no calls, no memory (instruction selection + loop shape)
#   arrwalk   dependent-index array walk  (addressing modes, load/store selection)
#   callloop  hot NON-recursive call      (pure call overhead, callee is noinline on both sides)
#
# FAIRNESS RULES:
#   - AXIOM i64/u64 vs C int64_t/uint64_t (never C `long` -- 32-bit on Windows).
#   - Every kernel carries a serial dependency so neither compiler can close-form,
#     vectorize, or delete the loop.
#   - callloop's C callee is __attribute__((noinline)) so we compare call-vs-call, not
#     call-vs-inlined.
#   - Correctness is pinned per shape: a build whose exit code disagrees with the C build
#     is reported BAD and its timing is not trusted.
#
# THE ASM FLOOR COLUMN: clang -O2 is not a pure codegen reference -- on fib it applies an
# accumulator transform that turns the second recursive call into a loop, so it runs about
# HALF the calls AXIOM does. Measuring only against clang therefore conflates two gaps. Where
# a `asm` source is supplied it is hand-written x86-64 in the SAME shape AXIOM emits, giving
# the practical floor for that algorithm on this CPU, and splitting the gap in two:
#     AXIOM / asm   = code-generation quality  (reachable by allocator/selector work)
#     asm   / clang = missing optimization     (reachable only by a real opt pass)
# Needs NASM; if nasm is absent the column is simply skipped.
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root
$bench = Join-Path $root "bin\bench"
New-Item -ItemType Directory -Force $bench | Out-Null

$axc  = Join-Path $root "bin\axc_native.exe"
# best-of-9, NOT best-of-5. Calibrated 2026-07-29c: at best-of-5 the memory-bound arrwalk shape
# swung 398-428 ms run to run, which read as a consistent 2-4% regression from a change that
# best-of-9 showed to be flat (413.8 / 411.8 / 412.7 across three builds). That phantom nearly
# caused a real 4.4% win to be thrown away. Do not lower this.
$runs = 9

# ---------------------------------------------------------------- sources
$srcs = @{}

$srcs["fib"] = @{
  ax = @"
fn fib(n: i64) -> i64:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)
fn main() -> i64:
    return fib(40) % 256
"@
  c = @"
#include <stdint.h>
int64_t fib(int64_t n){ if(n<2) return n; return fib(n-1)+fib(n-2); }
int main(){ return (int)(fib(40) % 256); }
"@
}

# Serial xorshift: three dependent shifts per iteration. No calls, no memory, no closed form.
$srcs["xorshift"] = @{
  ax = @"
fn main() -> i64:
    mut x: u64 = 88172645463325252
    mut i: u64 = 0
    while i < 120000000:
        x = x ^ (x << 13)
        x = x ^ (x >> 7)
        x = x ^ (x << 17)
        i = i + 1
    return (x & 255) as i64
"@
  c = @"
#include <stdint.h>
int main(){
    uint64_t x = 88172645463325252ULL;
    for (uint64_t i = 0; i < 120000000ULL; i++) {
        x ^= x << 13;
        x ^= x >> 7;
        x ^= x << 17;
    }
    return (int)(x & 255);
}
"@
}

# Dependent-index walk over a 64K-element global array: the next index is derived from the
# value just loaded, so neither compiler can vectorize or prefetch it away.
$srcs["arrwalk"] = @{
  ax = @"
mut tbl: [i64; 65536]

fn main() -> i64:
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
  c = @"
#include <stdint.h>
int64_t tbl[65536];
int main(){
    for (int64_t i = 0; i < 65536; i++) tbl[i] = (i * 2654435761 + 12345) & 65535;
    int64_t idx = 0, acc = 0;
    for (int64_t k = 0; k < 40000000; k++) { idx = tbl[idx]; acc += idx; }
    return (int)(acc & 255);
}
"@
}

# Pure call overhead: a small non-recursive callee invoked in a hot loop, with a serial
# dependency through the accumulator. C callee is noinline so both sides really call.
$srcs["callloop"] = @{
  ax = @"
fn work(a: i64, b: i64, c: i64) -> i64:
    return a + b * 2 + c * 3

fn main() -> i64:
    mut acc: i64 = 1
    mut i: i64 = 0
    while i < 60000000:
        acc = work(acc, i, 7) & 1048575
        i = i + 1
    return acc & 255
"@
  c = @"
#include <stdint.h>
__attribute__((noinline)) int64_t work(int64_t a, int64_t b, int64_t c){ return a + b*2 + c*3; }
int main(){
    int64_t acc = 1;
    for (int64_t i = 0; i < 60000000; i++) acc = work(acc, i, 7) & 1048575;
    return (int)(acc & 255);
}
"@
}

# --- hand-written ASM floors (same shape AXIOM emits, minimal frame) ---
# fib: naive DOUBLE recursion (not clang's accumulator loop), no rbp chain, 2 callee-saved.
$srcs["fib"].asm = @"
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

# xorshift: the tightest possible form of the same serial dependency -- one scratch
# register, dec/jnz loop control. Pure ALU floor, no calls, no memory.
$srcs["xorshift"].asm = @"
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

# arrwalk: the hot half is a SERIAL pointer chase (`idx = tbl[idx]`), so the floor is about
# dependency-chain length, not instruction count. Everything loop-invariant is hoisted (the
# table base) and the index scaling lives in the addressing mode, which is what keeps the chain
# at load-latency: idx -> load -> idx. Note RIP-relative addressing cannot carry an index
# register, so the base genuinely has to be materialised into one -- once, before the loop.
# The fill loop runs 65536 times against the walk's 40M and is not the measurement.
$srcs["arrwalk"].asm = @"
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

# callloop: work() is INLINED here on purpose. AXIOM already inlines it (confirmed in the
# disassembly -- the emitted loop body contains no call), so a floor that kept the call would
# not be measuring the same program, and the AXIOM/asm ratio would silently credit the backend
# for an inlining decision made upstream. With it inlined the constant `c*3` folds to 21 and
# `b*2` folds into the LEA, which is exactly the codegen quality this column is meant to bound.
$srcs["callloop"].asm = @"
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

$order = @("fib","xorshift","arrwalk","callloop")

# NASM is optional: without it the asm floor column is skipped, everything else still runs.
$nasm = $null
foreach ($cand in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $c = Get-Command $cand -ErrorAction SilentlyContinue
    if ($c) { $nasm = $c.Source; break }
}
if (-not $nasm) { Write-Host "(nasm not found -- asm floor column skipped)" -ForegroundColor DarkGray }

function Time-Exe($exe) {
    & $exe | Out-Null; $ec = $LASTEXITCODE
    $best = [double]::MaxValue
    for ($i=0; $i -lt $runs; $i++) {
        $t = Measure-Command { & $exe | Out-Null }
        if ($t.TotalMilliseconds -lt $best) { $best = $t.TotalMilliseconds }
    }
    return @{ ms=$best; ec=$ec }
}

Write-Host "=== M6 perf suite: AXIOM -O3 vs clang -O2 (best of $runs) ===" -ForegroundColor Cyan
Write-Host ("{0,-10} {1,9} {2,9} {3,9} {4,8} {5,8}  {6}" -f "shape","axiom ms","asm ms","clang ms","vs asm","vs clang","exit")

$results = @{}
foreach ($name in $order) {
    $axsrc = Join-Path $bench "$name.ax"
    $csrc  = Join-Path $bench "$name.c"
    $srcs[$name].ax | Out-File -Encoding ascii $axsrc
    $srcs[$name].c  | Out-File -Encoding ascii $csrc

    $axexe = Join-Path $bench "${name}_ax.exe"
    $clexe = Join-Path $bench "${name}_clang.exe"
    Remove-Item -Force $axexe,$clexe -ErrorAction SilentlyContinue

    & $axc build $axsrc -o $axexe -O3 2>&1 | Out-Null
    clang -O2 $csrc -o $clexe 2>&1 | Out-Null

    if (-not (Test-Path $axexe)) { Write-Host ("{0,-10} AXIOM-BUILD-FAIL" -f $name) -ForegroundColor Red; continue }
    if (-not (Test-Path $clexe)) { Write-Host ("{0,-10} CLANG-BUILD-FAIL" -f $name) -ForegroundColor Red; continue }

    # optional hand-written asm floor
    $asmMs = $null
    if ($nasm -and $srcs[$name].ContainsKey("asm")) {
        $asmsrc = Join-Path $bench "${name}_hand.asm"
        $asmobj = Join-Path $bench "${name}_hand.obj"
        $asmexe = Join-Path $bench "${name}_hand.exe"
        $srcs[$name].asm | Out-File -Encoding ascii $asmsrc
        Remove-Item -Force $asmobj,$asmexe -ErrorAction SilentlyContinue
        & $nasm -f win64 $asmsrc -o $asmobj 2>&1 | Out-Null
        if (Test-Path $asmobj) { gcc $asmobj -o $asmexe 2>&1 | Out-Null }
        if (Test-Path $asmexe) { $h = Time-Exe $asmexe; $asmMs = $h.ms; $asmEc = $h.ec }
    }

    $a = Time-Exe $axexe
    $c = Time-Exe $clexe
    $ratio = $a.ms / $c.ms
    $agree = if ($a.ec -eq $c.ec) { "$($a.ec) ok" } else { "MISMATCH ax=$($a.ec) c=$($c.ec)" }
    if ($asmMs -and $asmEc -ne $a.ec) { $agree = "$agree ASM-MISMATCH($asmEc)" }
    $color = if ($a.ec -ne $c.ec) { "Red" } elseif ($ratio -le 1.30) { "Green" } elseif ($ratio -le 2.0) { "Yellow" } else { "DarkYellow" }
    $asmTxt   = if ($asmMs) { "{0,9:F1}" -f $asmMs } else { "{0,9}" -f "-" }
    $vsAsmTxt = if ($asmMs) { "{0,7:F2}x" -f ($a.ms / $asmMs) } else { "{0,8}" -f "-" }
    Write-Host ("{0,-10} {1,9:F1} {2} {3,9:F1} {4} {5,7:F2}x  {6}" -f $name, $a.ms, $asmTxt, $c.ms, $vsAsmTxt, $ratio, $agree) -ForegroundColor $color
    $results[$name] = @{ ratio=$ratio; agree=($a.ec -eq $c.ec) }
}

Write-Host ""
Write-Host "Interpretation: 'vs asm' is the code-generation gap we can close with backend work;" -ForegroundColor DarkGray
Write-Host "'vs clang' also contains whatever high-level transform clang applied that we do not" -ForegroundColor DarkGray
Write-Host "have. When vs-asm is near 1.0 the backend is at its floor and the rest is an opt pass." -ForegroundColor DarkGray
