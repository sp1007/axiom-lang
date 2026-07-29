# Are the xorshift / arrwalk / callloop floors actually the fastest expression of their
# algorithm? fib's V0 floor turned out not to be, and it flattered AXIOM by 5%. The variable
# none of the three specifies is LOOP-BODY alignment: the function entry is aligned but the
# hot backward-branch target lands wherever the prologue leaves it.
#
# Each shape is built twice -- as written today, and with `align 16` on the loop label only --
# and compared paired. A faster aligned variant means the floor is too slow and every AXIOM
# ratio against it is flattering.
$ErrorActionPreference = "Continue"
$out = "d:\projects\compiler\Axiom\bin\bench\floorchk"
New-Item -ItemType Directory -Force $out | Out-Null
$nasm = $null
foreach ($c in @("nasm", "$env:LOCALAPPDATA\bin\NASM\nasm.exe", "C:\Program Files\NASM\nasm.exe")) {
    $g = Get-Command $c -ErrorAction SilentlyContinue
    if ($g) { $nasm = $g.Source; break }
}
if (-not $nasm) { Write-Host "nasm not found"; exit 1 }

function Body($name, $alignLoop) {
  $al = if ($alignLoop) { "align 16`n" } else { "" }
  if ($name -eq "xorshift") {
    return @"
default rel
section .text
global main
align 16
main:
    mov     rax, 88172645463325252
    mov     rcx, 120000000
$al.loop:
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
  } elseif ($name -eq "arrwalk") {
    return @"
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
    mov     r8, 40000000
$al.walk:
    mov     rax, [r9 + rax*8]
    add     rcx, rax
    dec     r8
    jnz     .walk
    mov     rax, rcx
    and     rax, 255
    ret
"@
  } else {
    return @"
default rel
section .text
global main
align 16
main:
    mov     rax, 1
    xor     rcx, rcx
    mov     r8, 60000000
$al.loop:
    lea     rax, [rax + rcx*2]
    add     rax, 21
    and     rax, 1048575
    inc     rcx
    cmp     rcx, r8
    jb      .loop
    and     rax, 255
    ret
"@
  }
}

$want = @{ xorshift=61; arrwalk=0; callloop=1 }
$exes = @{}
foreach ($n in @("xorshift","arrwalk","callloop")) {
  foreach ($v in @(@{t="plain"; a=$false}, @{t="loopalign"; a=$true})) {
    $tag = "$($n)_$($v.t)"
    $asm = Join-Path $out "$tag.asm"; $obj = Join-Path $out "$tag.obj"; $exe = Join-Path $out "$tag.exe"
    Body $n $v.a | Out-File -Encoding ascii $asm
    Remove-Item -Force $obj,$exe -ErrorAction SilentlyContinue
    & $nasm -f win64 $asm -o $obj 2>&1 | Out-Null
    if (Test-Path $obj) { gcc $obj -o $exe 2>&1 | Out-Null }
    if (-not (Test-Path $exe)) { Write-Host "BUILD FAILED $tag"; exit 1 }
    & $exe | Out-Null
    if ($LASTEXITCODE -ne $want[$n]) { Write-Host "WRONG RESULT $tag exit=$LASTEXITCODE want=$($want[$n])"; exit 1 }
    $exes[$tag] = $exe
  }
}
Write-Host "all six build and return the right exit code" -ForegroundColor Green

function BestOf9($exe){ $b=[double]::MaxValue; for($i=0;$i -lt 9;$i++){ $t=Measure-Command { & $exe | Out-Null }; if($t.TotalMilliseconds -lt $b){$b=$t.TotalMilliseconds} }; return $b }
Write-Host ""
Write-Host ("{0,-10} {1,10} {2,11} {3,9}" -f "shape","plain ms","loopalign","delta%")
foreach ($n in @("xorshift","arrwalk","callloop")) {
  $p=@(); $q=@()
  for ($r=1; $r -le 3; $r++) {
    if ($r % 2 -eq 1) { $p += BestOf9 $exes["$($n)_plain"]; $q += BestOf9 $exes["$($n)_loopalign"] }
    else              { $q += BestOf9 $exes["$($n)_loopalign"]; $p += BestOf9 $exes["$($n)_plain"] }
  }
  $mp=($p|Measure-Object -Minimum).Minimum; $mq=($q|Measure-Object -Minimum).Minimum
  Write-Host ("{0,-10} {1,10:F1} {2,11:F1} {3,8:F2}%" -f $n,$mp,$mq,(($mq/$mp-1)*100))
}
