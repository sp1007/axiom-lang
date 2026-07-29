# RFC 0036 §6, mandated first step: price self-tail-recursion BEFORE writing the pass.
#
# None of the four existing benchmark shapes is tail-recursive, so the suite currently cannot
# measure this at all. New shape: sumto(n, acc) accumulates 1..n with the recursive call in TAIL
# position. 20M total calls, split as depth 10000 x 2000 outer iterations so the un-transformed
# version does not blow the stack (10000 frames x ~48B is well under the default).
#
#   AX     AXIOM as it compiles today (real recursion)
#   REC    NASM floor in the shape AXIOM emits -- real calls, minimal frame
#   LOOP   NASM floor in the shape the transform WOULD produce -- jump to entry, no call
#
# The value of the whole RFC is REC - LOOP. If that gap is small, RFC 0036 §9 says close it
# unimplemented. AX - REC is the usual codegen gap and is NOT what is being priced here.
$ErrorActionPreference = "Continue"
$root = "d:\projects\compiler\Axiom"
$out  = Join-Path $root "bin\bench\tailrec"
New-Item -ItemType Directory -Force $out | Out-Null
$nasm = $null
foreach ($c in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $g = Get-Command $c -ErrorAction SilentlyContinue
    if ($g) { $nasm = $g.Source; break }
}
if (-not $nasm) { Write-Host "nasm not found"; exit 1 }

@"
fn sumto(n: i64, acc: i64) -> i64:
    if n == 0:
        return acc
    return sumto(n - 1, acc + n)

fn main() -> i64:
    mut t: i64 = 0
    mut i: i64 = 0
    while i < 2000:
        t = t + sumto(10000, 0)
        i = i + 1
    return t % 256
"@ | Out-File -Encoding ascii (Join-Path $out "tailrec.ax")

# REC: same shape AXIOM emits -- a real call per step, rbp frame (measured faster than none on
# fib, 2026-07-30), 2 callee-saved.
$REC = @"
default rel
section .text
global sumto
global main
align 16
sumto:
    test    rcx, rcx
    jz      .base
    push    rbp
    mov     rbp, rsp
    sub     rsp, 32
    lea     rax, [rdx + rcx]
    dec     rcx
    mov     rdx, rax
    call    sumto
    add     rsp, 32
    pop     rbp
    ret
.base:
    mov     rax, rdx
    ret
align 16
main:
    push    rbx
    push    rsi
    sub     rsp, 40
    xor     rbx, rbx
    mov     rsi, 2000
.outer:
    mov     rcx, 10000
    xor     rdx, rdx
    call    sumto
    add     rbx, rax
    dec     rsi
    jnz     .outer
    mov     rax, rbx
    mov     rcx, 256
    xor     rdx, rdx
    div     rcx
    mov     rax, rdx
    add     rsp, 40
    pop     rsi
    pop     rbx
    ret
"@

# LOOP: exactly what the RFC's transform produces -- arguments reassigned, jump to entry,
# one frame for the whole recursion.
$LOOP = @"
default rel
section .text
global sumto
global main
align 16
sumto:
align 16
.top:
    test    rcx, rcx
    jz      .base
    lea     rdx, [rdx + rcx]
    dec     rcx
    jmp     .top
.base:
    mov     rax, rdx
    ret
align 16
main:
    push    rbx
    push    rsi
    sub     rsp, 40
    xor     rbx, rbx
    mov     rsi, 2000
.outer:
    mov     rcx, 10000
    xor     rdx, rdx
    call    sumto
    add     rbx, rax
    dec     rsi
    jnz     .outer
    mov     rax, rbx
    mov     rcx, 256
    xor     rdx, rdx
    div     rcx
    mov     rax, rdx
    add     rsp, 40
    pop     rsi
    pop     rbx
    ret
"@

$exes = @{}
foreach ($v in @(@{n="REC"; s=$REC}, @{n="LOOP"; s=$LOOP})) {
  $asm = Join-Path $out "$($v.n).asm"; $obj = Join-Path $out "$($v.n).obj"; $exe = Join-Path $out "$($v.n).exe"
  $v.s | Out-File -Encoding ascii $asm
  Remove-Item -Force $obj,$exe -ErrorAction SilentlyContinue
  & $nasm -f win64 $asm -o $obj 2>&1 | Out-Null
  if (Test-Path $obj) { gcc $obj -o $exe 2>&1 | Out-Null }
  if (-not (Test-Path $exe)) { Write-Host "BUILD FAILED $($v.n)"; exit 1 }
  $exes[$v.n] = $exe
}
& (Join-Path $root "bin\axc_native.exe") build (Join-Path $out "tailrec.ax") -o (Join-Path $out "AX.exe") -O3 2>&1 | Out-Null
if (-not (Test-Path (Join-Path $out "AX.exe"))) { Write-Host "AXIOM BUILD FAILED"; exit 1 }
$exes["AX"] = Join-Path $out "AX.exe"

# every variant must agree on the answer, else the comparison is meaningless
$res = @{}
foreach ($k in @("REC","LOOP","AX")) { & $exes[$k] | Out-Null; $res[$k] = $LASTEXITCODE }
Write-Host ("exit codes: REC=$($res['REC']) LOOP=$($res['LOOP']) AX=$($res['AX'])")
if ($res["REC"] -ne $res["LOOP"] -or $res["REC"] -ne $res["AX"]) { Write-Host "MISMATCH -- not comparable" -ForegroundColor Red; exit 1 }
Write-Host "all three agree" -ForegroundColor Green

function BestOf9($exe){ $b=[double]::MaxValue; for($i=0;$i -lt 9;$i++){ $t=Measure-Command { & $exe | Out-Null }; if($t.TotalMilliseconds -lt $b){$b=$t.TotalMilliseconds} }; return $b }
$acc = @{ REC=@(); LOOP=@(); AX=@() }
for ($r=1; $r -le 3; $r++) {
  $ord = if ($r % 2 -eq 1) { @("REC","LOOP","AX") } else { @("AX","LOOP","REC") }
  foreach ($k in $ord) { $acc[$k] += BestOf9 $exes[$k] }
}
$mr=($acc["REC"]|Measure-Object -Minimum).Minimum
$ml=($acc["LOOP"]|Measure-Object -Minimum).Minimum
$ma=($acc["AX"]|Measure-Object -Minimum).Minimum
Write-Host ""
Write-Host ("{0,-6} {1,9}" -f "variant","best ms")
Write-Host ("{0,-6} {1,9:F1}" -f "REC",$mr)
Write-Host ("{0,-6} {1,9:F1}" -f "LOOP",$ml)
Write-Host ("{0,-6} {1,9:F1}" -f "AX",$ma)
Write-Host ""
Write-Host ("VALUE OF THE TRANSFORM (REC -> LOOP): {0:F1}%" -f (($ml/$mr-1)*100))
Write-Host ("AXIOM vs its own shape (AX / REC):    {0:F3}x" -f ($ma/$mr))
