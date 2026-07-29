# Re-price the two fib candidates dismissed on 2026-07-29 (shrink-wrapping +0.5ms,
# drop-rbp-frame -17ms), under the paired-alternating method. Four NASM variants of the
# SAME double-recursive fib, differing only in the two features:
#
#   V0  shrink-wrapped, no rbp   = the current floor
#   V1  shrink-wrapped, rbp
#   V2  prologue-first, no rbp
#   V3  prologue-first, rbp      = the shape AXIOM actually emits
#
# If V3 lands near AXIOM's own time, these two features explain the gap and are worth
# implementing; V1 and V2 then separate the individual contributions.
$ErrorActionPreference = "Continue"
$root = "d:\projects\compiler\Axiom"
$out  = Join-Path $root "bin\bench\price"
New-Item -ItemType Directory -Force $out | Out-Null

$nasm = $null
foreach ($c in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $g = Get-Command $c -ErrorAction SilentlyContinue
    if ($g) { $nasm = $g.Source; break }
}
if (-not $nasm) { Write-Host "nasm not found"; exit 1 }

# body shared by all four; only prologue placement and rbp differ
function Src($shrink, $rbp) {
  $pro = if ($rbp) { "    push    rbp`n    mov     rbp, rsp`n    push    rbx`n    push    rsi`n    sub     rsp, 32" }
         else      { "    push    rbx`n    push    rsi`n    sub     rsp, 40" }
  $epi = if ($rbp) { "    add     rsp, 32`n    pop     rsi`n    pop     rbx`n    pop     rbp" }
         else      { "    add     rsp, 40`n    pop     rsi`n    pop     rbx" }
  $core = @"
    mov     rbx, rcx
    dec     rcx
    call    fib_hand
    mov     rsi, rax
    lea     rcx, [rbx-2]
    call    fib_hand
    add     rax, rsi
$epi
    ret
"@
  if ($shrink) {
    # base case tested BEFORE the prologue: it costs cmp/jl/mov/ret and nothing else
    $body = "    cmp     rcx, 2`n    jl      .base`n$pro`n$core`n.base:`n    mov     rax, rcx`n    ret"
  } else {
    # prologue unconditionally first: the base case must also tear it down
    $body = "$pro`n    cmp     rcx, 2`n    jl      .base`n$core`n.base:`n    mov     rax, rcx`n$epi`n    ret"
  }
  return @"
default rel
section .text
global fib_hand
global main
align 16
fib_hand:
$body
align 16
main:
    sub     rsp, 40
    mov     rcx, 40
    call    fib_hand
    and     rax, 255
    add     rsp, 40
    ret
"@
}

$variants = @(
  @{ n="V0_shrink_norbp"; s=$true;  r=$false },
  @{ n="V1_shrink_rbp";   s=$true;  r=$true  },
  @{ n="V2_pro1st_norbp"; s=$false; r=$false },
  @{ n="V3_pro1st_rbp";   s=$false; r=$true  }
)

foreach ($v in $variants) {
  $asm = Join-Path $out "$($v.n).asm"; $obj = Join-Path $out "$($v.n).obj"; $exe = Join-Path $out "$($v.n).exe"
  Src $v.s $v.r | Out-File -Encoding ascii $asm
  Remove-Item -Force $obj,$exe -ErrorAction SilentlyContinue
  & $nasm -f win64 $asm -o $obj 2>&1 | Out-Null
  if (Test-Path $obj) { gcc $obj -o $exe 2>&1 | Out-Null }
  if (-not (Test-Path $exe)) { Write-Host "BUILD FAILED: $($v.n)"; exit 1 }
  & $exe | Out-Null
  if ($LASTEXITCODE -ne 203) { Write-Host "WRONG RESULT $($v.n): exit=$LASTEXITCODE (want 203)"; exit 1 }
}
Write-Host "all four variants build and return 203" -ForegroundColor Green

function BestOf9($exe){ $b=[double]::MaxValue; for($i=0;$i -lt 9;$i++){ $t=Measure-Command { & $exe | Out-Null }; if($t.TotalMilliseconds -lt $b){$b=$t.TotalMilliseconds} }; return $b }

# interleave all four per round so machine drift hits every variant equally
$acc = @{}
foreach ($v in $variants) { $acc[$v.n] = @() }
for ($r=1; $r -le 4; $r++) {
  $order = if ($r % 2 -eq 1) { $variants } else { $variants[($variants.Count-1)..0] }
  foreach ($v in $order) { $acc[$v.n] += BestOf9 (Join-Path $out "$($v.n).exe") }
}
Write-Host ""
Write-Host ("{0,-18} {1,9} {2,9}" -f "variant","best ms","vs V0")
$base = ($acc["V0_shrink_norbp"] | Measure-Object -Minimum).Minimum
foreach ($v in $variants) {
  $m = ($acc[$v.n] | Measure-Object -Minimum).Minimum
  Write-Host ("{0,-18} {1,9:F1} {2,8:F2}%" -f $v.n, $m, (($m/$base-1)*100))
}
