# Paired, alternating pricing of peephole 1f (guarded) against the HEAD reference.
#
# Discipline mandated by knowledge/m6-perf-baseline.md: a single run has 8-10% variance, so
# measure REF and CAND in ALTERNATING pairs inside one session and report the sign of EACH pair.
# best-of-N per sample removes scheduler noise; the pair structure removes drift.
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

function BestOfN($exe, $n) {
    $b = [double]::MaxValue
    for ($i = 0; $i -lt $n; $i++) {
        $t = Measure-Command { & $exe | Out-Null }
        if ($t.TotalMilliseconds -lt $b) { $b = $t.TotalMilliseconds }
    }
    return $b
}

$ref = "$root\bin\zz_trl_ref.exe"
$cand = "$root\bin\zz_trl_1fg.exe"
$nop = "$root\bin\zz_nop.exe"

$nopMs = BestOfN $nop 9
Write-Host ("startup floor (trivial program, best-of-9): {0:F1} ms" -f $nopMs)

for ($p = 0; $p -lt 4; $p++) {
    $r = BestOfN $ref 9
    $c = BestOfN $cand 9
    $rc = $r - $nopMs
    $cc = $c - $nopMs
    $d = 0.0
    if ($rc -gt 0) { $d = ($cc - $rc) / $rc * 100.0 }
    Write-Host ("pair {0}: REF {1:F1} ms (compute {2:F1})  1FG {3:F1} ms (compute {4:F1})  compute delta {5:F1}%" -f $p, $r, $rc, $c, $cc, $d)
}
