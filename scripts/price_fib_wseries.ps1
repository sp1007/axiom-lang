# W-series for fib, same method that explained xorshift on 2026-07-29f: start from the
# same-shape floor and add AXIOM's actual instruction sequence until the time is reproduced.
#
#   V3  = prologue-first + rbp                  (AXIOM's FRAME shape, tight body)
#   V4  = V3 + AXIOM's four extra reg-reg movs  (a literal transcription of ax_fib)
#
# If V4 lands near AXIOM's ~608 ms, those copies are the cost and the gap is explained.
# If V4 stays near V3's ~493 ms, the copies are free (as the 2026-07-29e coalescing
# experiment concluded) and the cost is something neither variant models -- which would be
# the more interesting answer, because nothing else differs.
$ErrorActionPreference = "Continue"
$out = "d:\projects\compiler\Axiom\bin\bench\price"
New-Item -ItemType Directory -Force $out | Out-Null
$nasm = $null
foreach ($c in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $g = Get-Command $c -ErrorAction SilentlyContinue
    if ($g) { $nasm = $g.Source; break }
}
if (-not $nasm) { Write-Host "nasm not found"; exit 1 }

$V3 = @"
default rel
section .text
global fib_hand
global main
align 16
fib_hand:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    sub     rsp, 32
    cmp     rcx, 2
    jl      .base
    mov     rbx, rcx
    dec     rcx
    call    fib_hand
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib_hand
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
align 16
main:
    sub     rsp, 40
    mov     rcx, 40
    call    fib_hand
    and     rax, 255
    add     rsp, 40
    ret
"@

# Literal transcription of ax_fib as emitted (driver 522BEA6B, -O3), including the four
# reg-reg copies the floor does not have: rcx->rax->rbx on entry, rax->rcx before each
# call, and rsi->rax on the way out.
$V4 = @"
default rel
section .text
global fib_hand
global main
align 16
fib_hand:
    push    rbp
    mov     rbp, rsp
    push    rbx
    push    rsi
    sub     rsp, 32
    mov     rax, rcx
    mov     rbx, rax
    cmp     rbx, 2
    jl      .base
    lea     rax, [rbx-1]
    mov     rcx, rax
    call    fib_hand
    mov     rsi, rax
    lea     rax, [rbx-2]
    mov     rcx, rax
    call    fib_hand
    add     rsi, rax
    mov     rax, rsi
    add     rsp, 32
    pop     rsi
    pop     rbx
    pop     rbp
    ret
.base:
    mov     rax, rbx
    add     rsp, 32
    pop     rsi
    pop     rbx
    pop     rbp
    ret
align 16
main:
    sub     rsp, 40
    mov     rcx, 40
    call    fib_hand
    and     rax, 255
    add     rsp, 40
    ret
"@

$set = @( @{n="V3_shape"; s=$V3}, @{n="V4_axiom_literal"; s=$V4} )
foreach ($v in $set) {
  $asm = Join-Path $out "$($v.n).asm"; $obj = Join-Path $out "$($v.n).obj"; $exe = Join-Path $out "$($v.n).exe"
  $v.s | Out-File -Encoding ascii $asm
  Remove-Item -Force $obj,$exe -ErrorAction SilentlyContinue
  & $nasm -f win64 $asm -o $obj 2>&1 | Out-Null
  if (Test-Path $obj) { gcc $obj -o $exe 2>&1 | Out-Null }
  if (-not (Test-Path $exe)) { Write-Host "BUILD FAILED: $($v.n)"; exit 1 }
  & $exe | Out-Null
  if ($LASTEXITCODE -ne 203) { Write-Host "WRONG RESULT $($v.n): exit=$LASTEXITCODE"; exit 1 }
}
Write-Host "V3 and V4 build and return 203" -ForegroundColor Green

function BestOf9($exe){ $b=[double]::MaxValue; for($i=0;$i -lt 9;$i++){ $t=Measure-Command { & $exe | Out-Null }; if($t.TotalMilliseconds -lt $b){$b=$t.TotalMilliseconds} }; return $b }

$names = @("V3_shape","V4_axiom_literal")
$axiom = "d:\projects\compiler\Axiom\bin\bench\fib_ax.exe"
$acc = @{}; foreach ($n in $names) { $acc[$n] = @() }; $acc["AXIOM"] = @()
for ($r=1; $r -le 4; $r++) {
  $ord = if ($r % 2 -eq 1) { @("V3_shape","V4_axiom_literal") } else { @("V4_axiom_literal","V3_shape") }
  foreach ($n in $ord) { $acc[$n] += BestOf9 (Join-Path $out "$n.exe") }
  $acc["AXIOM"] += BestOf9 $axiom
}
Write-Host ""
Write-Host ("{0,-20} {1,9}" -f "variant","best ms")
foreach ($n in @("V3_shape","V4_axiom_literal","AXIOM")) {
  Write-Host ("{0,-20} {1,9:F1}" -f $n, ($acc[$n] | Measure-Object -Minimum).Minimum)
}
