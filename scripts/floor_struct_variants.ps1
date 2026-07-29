# Structural variant study for the three floors fib's V0 defect was never checked against.
# fib's floor turned out NOT to be the fastest expression of its algorithm (a frame-pointer
# variant was 6% faster), which flattered AXIOM. Loop-body alignment was already ruled out for
# the other three; this tries STRUCTURALLY different formulations, which is what found fib's.
#
#   xorshift  A = current floor (dec/jnz)          B = unrolled x2 (half the loop overhead)
#   arrwalk   A = current floor (dec/jnz)          B = unrolled x2
#   callloop  A = current floor (inc/cmp/jb = 3)   B = countdown (dec/jnz = 2 instructions)
#
# callloop's B is the strongest candidate: its floor spends three instructions on loop control
# where two suffice, the same class of miss as fib's V0.
$ErrorActionPreference = "Continue"
$out = "d:\projects\compiler\Axiom\bin\bench\struct"
New-Item -ItemType Directory -Force $out | Out-Null
$nasm = $null
foreach ($c in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $g = Get-Command $c -ErrorAction SilentlyContinue
    if ($g) { $nasm = $g.Source; break }
}
if (-not $nasm) { Write-Host "nasm not found"; exit 1 }

$src = @{}

$src["xorshift_A"] = @"
default rel
section .text
global main
align 16
main:
    mov     rax, 88172645463325252
    mov     rcx, 120000000
align 16
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

# same 120M steps, two steps per iteration -> loop control amortised over two
$src["xorshift_B"] = @"
default rel
section .text
global main
align 16
main:
    mov     rax, 88172645463325252
    mov     rcx, 60000000
align 16
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

$arrwalk_head = @"
default rel
section .bss
tbl:    resq 65536
section .text
global main
align 16
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
"@

$src["arrwalk_A"] = $arrwalk_head + "`n" + @"
    mov     r8, 40000000
align 16
.walk:
    mov     rax, [r9 + rax*8]
    add     rcx, rax
    dec     r8
    jnz     .walk
    mov     rax, rcx
    and     rax, 255
    ret
"@

$src["arrwalk_B"] = $arrwalk_head + "`n" + @"
    mov     r8, 20000000
align 16
.walk:
    mov     rax, [r9 + rax*8]
    add     rcx, rax
    mov     rax, [r9 + rax*8]
    add     rcx, rax
    dec     r8
    jnz     .walk
    mov     rax, rcx
    and     rax, 255
    ret
"@

$src["callloop_A"] = @"
default rel
section .text
global main
align 16
main:
    mov     rax, 1
    xor     rcx, rcx
    mov     r8, 60000000
align 16
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

# countdown: r8 counts down and doubles as the loop test, so loop control is dec+jnz (2)
# instead of inc+cmp+jb (3). rcx still walks 0..N-1 so the arithmetic is identical.
$src["callloop_B"] = @"
default rel
section .text
global main
align 16
main:
    mov     rax, 1
    xor     rcx, rcx
    mov     r8, 60000000
align 16
.loop:
    lea     rax, [rax + rcx*2]
    add     rax, 21
    and     rax, 1048575
    inc     rcx
    dec     r8
    jnz     .loop
    and     rax, 255
    ret
"@

$want = @{ xorshift_A=61; xorshift_B=61; arrwalk_A=0; arrwalk_B=0; callloop_A=1; callloop_B=1 }
$exes = @{}
foreach ($k in $src.Keys) {
  $asm = Join-Path $out "$k.asm"; $obj = Join-Path $out "$k.obj"; $exe = Join-Path $out "$k.exe"
  $src[$k] | Out-File -Encoding ascii $asm
  Remove-Item -Force $obj,$exe -ErrorAction SilentlyContinue
  & $nasm -f win64 $asm -o $obj 2>&1 | Out-Null
  if (Test-Path $obj) { gcc $obj -o $exe 2>&1 | Out-Null }
  if (-not (Test-Path $exe)) { Write-Host "BUILD FAILED $k"; exit 1 }
  & $exe | Out-Null
  if ($LASTEXITCODE -ne $want[$k]) { Write-Host "WRONG RESULT $k exit=$LASTEXITCODE want=$($want[$k])"; exit 1 }
  $exes[$k] = $exe
}
Write-Host "all six variants build and produce the correct result" -ForegroundColor Green

function BestOf9($exe){ $b=[double]::MaxValue; for($i=0;$i -lt 9;$i++){ $t=Measure-Command { & $exe | Out-Null }; if($t.TotalMilliseconds -lt $b){$b=$t.TotalMilliseconds} }; return $b }
Write-Host ""
Write-Host ("{0,-10} {1,10} {2,10} {3,9}" -f "shape","A (floor)","B (variant)","B vs A")
foreach ($s in @("xorshift","arrwalk","callloop")) {
  $pa=@(); $pb=@()
  for ($r=1; $r -le 3; $r++) {
    if ($r % 2 -eq 1) { $pa += BestOf9 $exes["${s}_A"]; $pb += BestOf9 $exes["${s}_B"] }
    else              { $pb += BestOf9 $exes["${s}_B"]; $pa += BestOf9 $exes["${s}_A"] }
  }
  $ma=($pa|Measure-Object -Minimum).Minimum; $mb=($pb|Measure-Object -Minimum).Minimum
  Write-Host ("{0,-10} {1,10:F1} {2,10:F1} {3,8:F2}%" -f $s,$ma,$mb,(($mb/$ma-1)*100))
}
