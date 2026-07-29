# scripts/perf_asm_variants.ps1
# PRICE THE CODEGEN BACKLOG, in assembly, before writing any compiler code.
#
# The measured fib gap splits into ~1.5x codegen + ~1.5x missing optimization
# (knowledge/m6-perf-baseline.md). "1.5x codegen" is still not actionable: it does not say
# WHICH of our codegen habits costs what. Three past optimization attempts were implemented
# first and measured afterwards -- one regressed, two were neutral.
#
# So: take the hand-written floor and re-introduce AXIOM's codegen deficiencies ONE AT A TIME,
# in NASM. Each step's delta is that deficiency's price, measured on this CPU, with no compiler
# risk and no gate to run. Only then do we know what is worth implementing.
#
#   V0  the floor: early exit BEFORE the frame, no rbp, 2 callee-saved, lea/dec
#   V1  V0 + frame set up before the early-exit test   -> prices SHRINK WRAPPING
#   V2  V1 + rbp frame pointer                          -> prices FRAME-POINTER OMISSION
#   V3  V2 + a 3rd callee-saved reg + the param copy chain -> prices COALESCING
#   V4  V3 + mov/sub instead of lea/dec + jcc/jmp shape -> prices SELECTION + BLOCK LAYOUT
#
# V4 should land near AXIOM's own number. If it does, the gap is fully explained and the
# per-step deltas ARE the prioritized backlog.
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root
$bench = Join-Path $root "bin\bench"
New-Item -ItemType Directory -Force $bench | Out-Null

$nasm = $null
foreach ($cand in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $c = Get-Command $cand -ErrorAction SilentlyContinue
    if ($c) { $nasm = $c.Source; break }
}
if (-not $nasm) { Write-Error "nasm not found"; exit 1 }

$variants = [ordered]@{}

# ---- V0: the floor (same shape AXIOM emits, written the way a human would) ----
$variants["V0 floor"] = @"
default rel
section .text
global fib
global main
fib:
    cmp     rcx, 2
    jl      .base
    push    rbx
    push    rsi
    sub     rsp, 40
    mov     rbx, rcx
    dec     rcx
    call    fib
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib
    add     rax, rsi
    add     rsp, 40
    pop     rsi
    pop     rbx
    ret
.base:
    mov     rax, rcx
    ret
"@

# ---- V1: frame BEFORE the test -- every base-case call now pays push/pop ----
$variants["V1 +no-shrinkwrap"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbx
    push    rsi
    sub     rsp, 40
    cmp     rcx, 2
    jl      .base
    mov     rbx, rcx
    dec     rcx
    call    fib
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib
    add     rax, rsi
    add     rsp, 40
    pop     rsi
    pop     rbx
    ret
.base:
    mov     rax, rcx
    add     rsp, 40
    pop     rsi
    pop     rbx
    ret
"@

# ---- V2: + rbp frame pointer ----
$variants["V2 +rbp frame"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    sub     rsp, 32
    cmp     rcx, 2
    jl      .base
    mov     rbx, rcx
    dec     rcx
    call    fib
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib
    add     rax, rsi
    add     rsp, 32
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rcx
    add     rsp, 32
    pop     rsi
    pop     rbx
    pop     rbp
    ret
"@

# ---- V3: + a 3rd callee-saved register + the redundant param copy chain ----
$variants["V3 +copychain/3csr"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    push    rdi
    sub     rsp, 40
    mov     rax, rcx
    mov     rbx, rax
    cmp     rbx, 2
    jl      .base
    mov     rcx, rbx
    dec     rcx
    call    fib
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib
    add     rax, rsi
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rbx
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
"@

# ---- V3b: + mov/sub instead of lea/dec (constant materialised into a register first).
#      Isolates INSTRUCTION SELECTION from the block layout change in V4. ----
$variants["V3b +mov/sub sel"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    push    rdi
    sub     rsp, 40
    mov     rax, rcx
    mov     rbx, rax
    cmp     rbx, 2
    jl      .base
    mov     rax, 1
    mov     rsi, rbx
    sub     rsi, rax
    mov     rcx, rsi
    call    fib
    mov     rsi, rax
    mov     rax, 2
    mov     rdi, rbx
    sub     rdi, rax
    mov     rcx, rdi
    call    fib
    add     rsi, rax
    mov     rax, rsi
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rbx
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
"@

# ---- V3c: V3b but with the constant folded into the SUB's immediate operand.
#      This is the SAFE half of the selection fix: it removes the MOV_IMM and its vreg
#      (pressure goes DOWN) while KEEPING the `mov dst,base` copy, so it cannot repeat the
#      2026-07-24e lea regression (which lengthened base's live range and provoked a spill).
#      The delta V3b->V3c is what immediate-folding alone buys; V3c->V3 is the extra that
#      only `lea`/`dec` can buy, i.e. the price of taking the riskier step. ----
$variants["V3c +imm-fold sub"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    push    rdi
    sub     rsp, 40
    mov     rax, rcx
    mov     rbx, rax
    cmp     rbx, 2
    jl      .base
    mov     rsi, rbx
    sub     rsi, 1
    mov     rcx, rsi
    call    fib
    mov     rsi, rax
    mov     rdi, rbx
    sub     rdi, 2
    mov     rcx, rdi
    call    fib
    add     rsi, rax
    mov     rax, rsi
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rbx
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
"@

# ---- V3d: V3c but the argument is computed DIRECTLY INTO rcx (the call-arg register)
#      instead of into a temp vreg that is then copied. No lea, no dec -- still `sub imm`.
#      V3c->V3d isolates COALESCING WITH A PRECOLORED REGISTER; V3d->V3 is then all that
#      `lea`/`dec` itself is worth. This is the measurement that decides whether to build a
#      coalescer (George-Appel, precolored nodes) or a selection change. ----
$variants["V3d +arg coalesce"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    push    rdi
    sub     rsp, 40
    mov     rax, rcx
    mov     rbx, rax
    cmp     rbx, 2
    jl      .base
    mov     rcx, rbx
    sub     rcx, 1
    call    fib
    mov     rsi, rax
    mov     rcx, rbx
    sub     rcx, 2
    call    fib
    add     rsi, rax
    mov     rax, rsi
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rbx
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
"@

# ---- V4: V3b + the jcc-then-unconditional-jmp block layout (no fallthrough).
#      This is a faithful transcription of what AXIOM actually emits today. ----
$variants["V4 =AXIOM shape"] = @"
default rel
section .text
global fib
global main
fib:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    push    rdi
    sub     rsp, 40
    mov     rax, rcx
    mov     rbx, rax
    cmp     rbx, 2
    jl      .base
    jmp     .else
.else:
    mov     rax, 1
    mov     rsi, rbx
    sub     rsi, rax
    mov     rcx, rsi
    call    fib
    mov     rsi, rax
    mov     rax, 2
    mov     rdi, rbx
    sub     rdi, rax
    mov     rcx, rdi
    call    fib
    add     rsi, rax
    mov     rax, rsi
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rbx
    add     rsp, 40
    pop     rdi
    pop     rsi
    pop     rbx
    pop     rbp
    ret
"@

$mainStub = @"

main:
    sub     rsp, 40
    mov     rcx, 40
    call    fib
    and     rax, 255
    add     rsp, 40
    ret
"@

$runs = 7
Write-Host "=== fib(40) codegen deficiency pricing (NASM, best of $runs) ===" -ForegroundColor Cyan
Write-Host ("{0,-22} {1,9} {2,9} {3}" -f "variant","ms","delta","exit")

$prev = $null
$results = [ordered]@{}
foreach ($name in $variants.Keys) {
    $safe = ($name -replace '[^A-Za-z0-9]','_')
    $src = Join-Path $bench "v_$safe.asm"
    $obj = Join-Path $bench "v_$safe.obj"
    $exe = Join-Path $bench "v_$safe.exe"
    ($variants[$name] + $mainStub) | Out-File -Encoding ascii $src
    Remove-Item -Force $obj,$exe -ErrorAction SilentlyContinue
    & $nasm -f win64 $src -o $obj 2>&1 | Out-Null
    if (-not (Test-Path $obj)) { Write-Host ("{0,-22} ASSEMBLE-FAIL" -f $name) -ForegroundColor Red; continue }
    gcc $obj -o $exe 2>&1 | Out-Null
    if (-not (Test-Path $exe)) { Write-Host ("{0,-22} LINK-FAIL" -f $name) -ForegroundColor Red; continue }

    & $exe | Out-Null; $ec = $LASTEXITCODE
    $best = [double]::MaxValue
    for ($i=0; $i -lt $runs; $i++) {
        $t = Measure-Command { & $exe | Out-Null }
        if ($t.TotalMilliseconds -lt $best) { $best = $t.TotalMilliseconds }
    }
    $delta = if ($prev) { "{0,9:+0.0;-0.0}" -f ($best - $prev) } else { "{0,9}" -f "-" }
    $ok = if ($ec -eq 203) { "203 ok" } else { "BAD($ec)" }
    $color = if ($ec -ne 203) { "Red" } else { "Gray" }
    Write-Host ("{0,-22} {1,9:F1} {2} {3}" -f $name, $best, $delta, $ok) -ForegroundColor $color
    $results[$name] = $best
    $prev = $best
}

Write-Host ""
Write-Host "Each delta is what that ONE codegen habit costs on fib(40). Implement in delta order," -ForegroundColor DarkGray
Write-Host "largest first; anything whose delta is noise is not worth compiler risk." -ForegroundColor DarkGray
