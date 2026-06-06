# scripts/harness.ps1
# AXIOM Compiler Test Harness
#
# USAGE
#   scripts\harness.ps1                    # all tests, stage 1, air mode
#   scripts\harness.ps1 -Stage 2           # test axc_stage2_native.exe
#   scripts\harness.ps1 -Mode run          # compile + execute binaries
#   scripts\harness.ps1 -Mode build        # compile only with native backend
#   scripts\harness.ps1 -Mode err          # only error-expectation tests
#   scripts\harness.ps1 -Filter codegen    # tests whose path matches "codegen"
#   scripts\harness.ps1 -Update            # regenerate .expected files from actual output
#   scripts\harness.ps1 -List              # print test inventory, do not run
#
# TEST CONVENTIONS
#   tests/**/*.ax                          all test sources (err_* and valid_* etc.)
#   err_*.ax                               compilation must fail (non-zero exit)
#   *.expected                             in -Mode run: compile+exec, compare stdout
#   *.pattern                              regex matched against compiler error output
#   # HARNESS: SKIP                        first-line comment skips the test entirely
#   # HARNESS: NORUN                       first-line comment skips execute step

param(
    [ValidateRange(0,3)][int]$Stage = 1,
    [ValidateSet("air","build","run","err","all")][string]$Mode = "air",
    [string]$Filter = "",
    [switch]$Update,
    [switch]$List,
    [switch]$Verbose,
    [switch]$NoColor,
    [switch]$CompareRef,   # In air mode: compare dump-air output against stage 0 reference
    [int]$TimeoutSec = 15  # Per-test timeout in seconds (0 = no timeout)
)

$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

# ── color helpers ─────────────────────────────────────────────────────────────
function Write-C($text, $color) {
    if ($NoColor) { Write-Host $text } else { Write-Host $text -ForegroundColor $color }
}

# ── select compiler binary ────────────────────────────────────────────────────
$compilerBinaries = @{
    0 = "bin\axc.exe"
    1 = "bin\axc_stage1.exe"
    2 = "bin\axc_stage2_native.exe"
    3 = "bin\axc_stage3_native.exe"
}
$compiler = $compilerBinaries[$Stage]
if (-not (Test-Path $compiler)) {
    Write-Error "Stage $Stage compiler not found: $compiler"
    Write-Error "Build it first with scripts\native_triple_build.ps1 or scripts\fast_native_cycle.ps1"
    exit 1
}

# Stage 0 reference (for AIR comparison mode)
$refCompiler = "bin\axc.exe"
$hasRef = (Test-Path $refCompiler) -and ($Stage -gt 0)

# ── discover test files ───────────────────────────────────────────────────────
$excludeNames = @("tmp_", "torture_", "axiom_gui_", "axiom_compliance_", "axiom_game_",
                  "axiom_ai_", "axiom_web_", "axiom_distributed_", "axiom_lowlevel_",
                  "axiom_devops_", "axiom_functional_", "axiom_security_", "axiom_ffi_",
                  "axiom_robotics_", "axiom_compiler_", "axiom_dbms_", "axiom_multimedia_",
                  "axiom_hft_", "axiom_dsl_", "axiom_scientific_")

$allTests = Get-ChildItem -Path tests -Filter "*.ax" -Recurse | Where-Object {
    $n = $_.Name
    $skip = $false
    foreach ($pat in $excludeNames) {
        if ($n.StartsWith($pat)) { $skip = $true; break }
    }
    # Also skip concurrency tests in non-run modes
    if ($_.FullName -match "\\concurrency\\") { $skip = $true }
    # Skip scratch
    if ($_.FullName -match "\\scratch\\") { $skip = $true }
    -not $skip
}

if ($Filter) {
    $allTests = $allTests | Where-Object { $_.FullName -like "*$Filter*" }
}

$allTests = $allTests | Sort-Object FullName

if ($List) {
    Write-C "=== AXIOM Test Inventory ===" "Cyan"
    $errCount = 0; $validCount = 0
    foreach ($f in $allTests) {
        $rel = (Resolve-Path -Relative $f.FullName).TrimStart(".\")
        $tag = if ($f.Name.StartsWith("err_")) { "[err]  " } else { "[valid]" }
        $hasExp = Test-Path ([System.IO.Path]::ChangeExtension($f.FullName, ".expected"))
        $expTag = if ($hasExp) { " +expected" } else { "" }
        Write-Host "  $tag $rel$expTag"
        if ($f.Name.StartsWith("err_")) { $errCount++ } else { $validCount++ }
    }
    Write-Host ""
    Write-Host "  Total: $($allTests.Count)  (valid: $validCount, err: $errCount)"
    exit 0
}

# ── temp dir ──────────────────────────────────────────────────────────────────
$tmpDir = ".test-tmp"
if (-not (Test-Path $tmpDir)) { New-Item -ItemType Directory -Path $tmpDir | Out-Null }

# ── result tracking ───────────────────────────────────────────────────────────
$nPass = 0; $nFail = 0; $nSkip = 0; $nUpdate = 0
$failures = [System.Collections.Generic.List[string]]::new()
$sw_total = [System.Diagnostics.Stopwatch]::StartNew()

# ── helpers ───────────────────────────────────────────────────────────────────

function Read-FirstLine($path) {
    try { (Get-Content $path -TotalCount 1) } catch { "" }
}

function Normalize-Output($lines) {
    ($lines -join "`n").Replace("`r`n", "`n").Replace("`r", "`n").Trim()
}

# Run an executable with arguments and an optional wall-clock timeout.
# Returns hashtable: ExitCode, Output (string[]), TimedOut (bool)
# Uses a temp file to avoid stdout/stderr deadlocks.
function Exec-Timed([string]$Exe, [string]$CmdArgs, [int]$Timeout = 0) {
    $outFile = [System.IO.Path]::GetTempFileName()
    # Wrap in cmd.exe so we can redirect both stdout and stderr to one file
    $cmdLine = "`"$Exe`" $CmdArgs"
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName        = "cmd.exe"
    $psi.Arguments       = "/c `"$cmdLine > `"$outFile`" 2>&1`""
    $psi.UseShellExecute = $false
    $psi.CreateNoWindow  = $true
    $p = [System.Diagnostics.Process]::Start($psi)
    $timedOut = $false
    if ($Timeout -gt 0) {
        if (-not $p.WaitForExit($Timeout * 1000)) {
            $timedOut = $true
            try { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue } catch {}
            $p.WaitForExit(3000) | Out-Null
        }
    } else {
        $p.WaitForExit()
    }
    $ec    = if ($timedOut) { -1 } else { $p.ExitCode }
    $lines = @()
    try { if (Test-Path $outFile) { $lines = Get-Content $outFile -Encoding UTF8 } } catch {}
    Remove-Item $outFile -Force -ErrorAction SilentlyContinue
    return @{ ExitCode = $ec; Output = $lines; TimedOut = $timedOut }
}

# Returns hashtable: Status=(pass|fail|skip|updated), Reason, Expected, Actual, Output
function Invoke-AirTest($file, $rel) {
    $r = Exec-Timed $compiler "dump-air `"$rel`" -O1" $TimeoutSec
    if ($r.TimedOut) {
        return @{ Status = "fail"; Reason = "timeout after ${TimeoutSec}s (possible stdlib import hang)" }
    }
    if ($r.ExitCode -ne 0) {
        return @{ Status = "fail"; Reason = "dump-air failed (exit $($r.ExitCode))"; Output = $r.Output }
    }

    if ($CompareRef) {
        if (-not $hasRef) { return @{ Status = "skip"; Reason = "no stage 0 reference" } }
        $ref = Exec-Timed $refCompiler "dump-air `"$rel`" -O1" $TimeoutSec
        if ($ref.TimedOut -or $ref.ExitCode -ne 0) {
            return @{ Status = "skip"; Reason = "axc.exe dump-air unsupported or timed out" }
        }
        $refStr = Normalize-Output $ref.Output
        $actStr = Normalize-Output $r.Output
        if ($refStr -ne $actStr) {
            return @{ Status = "fail"; Reason = "AIR output mismatch vs stage 0 reference";
                      Expected = $refStr; Actual = $actStr }
        }
    }
    return @{ Status = "pass" }
}

function Check-ErrorPattern($file, $outLines) {
    $patFile = [System.IO.Path]::ChangeExtension($file.FullName, ".pattern")
    if (-not (Test-Path $patFile)) { return @{ Status = "pass" } }
    $pattern = (Get-Content $patFile -Raw).Trim()
    $outStr  = Normalize-Output $outLines
    if ($outStr -match $pattern) { return @{ Status = "pass" } }
    return @{ Status = "fail"; Reason = "error output did not match .pattern";
              Expected = $pattern; Actual = $outStr }
}

function Invoke-BuildTest($file, $rel, $isErr) {
    $tmpExe = "$tmpDir\build_test.exe"
    if (Test-Path $tmpExe) { Remove-Item $tmpExe -Force }

    $r = Exec-Timed $compiler "build `"$rel`" -o `"$tmpExe`" -self-link -O1" $TimeoutSec
    if ($r.TimedOut) {
        return @{ Status = "fail"; Reason = "compile timeout after ${TimeoutSec}s" }
    }

    if ($isErr) {
        if ($r.ExitCode -ne 0) { return Check-ErrorPattern $file $r.Output }
        return @{ Status = "fail"; Reason = "expected compiler error, compilation succeeded" }
    }
    if ($r.ExitCode -ne 0) {
        return @{ Status = "fail"; Reason = "compilation failed (exit $($r.ExitCode))"; Output = $r.Output }
    }
    return @{ Status = "pass" }
}

function Invoke-RunTest($file, $rel, $isErr) {
    if ($isErr) { return Invoke-BuildTest $file $rel $isErr }

    $first = Read-FirstLine $file.FullName
    if ($first -match "HARNESS:\s*NORUN") {
        return Invoke-BuildTest $file $rel $false
    }

    $tmpExe = "$tmpDir\run_test.exe"
    if (Test-Path $tmpExe) { Remove-Item $tmpExe -Force }

    $comp = Exec-Timed $compiler "build `"$rel`" -o `"$tmpExe`" -self-link -O1" $TimeoutSec
    if ($comp.TimedOut) { return @{ Status = "fail"; Reason = "compile timeout" } }
    if ($comp.ExitCode -ne 0) {
        return @{ Status = "fail"; Reason = "compilation failed"; Output = $comp.Output }
    }

    $run = Exec-Timed $tmpExe "" 10   # 10s run timeout
    $actual = Normalize-Output $run.Output
    if ($run.TimedOut) { return @{ Status = "fail"; Reason = "run timeout (infinite loop?)" } }

    $expFile = [System.IO.Path]::ChangeExtension($file.FullName, ".expected")

    if ($Update) {
        $dir = Split-Path $expFile
        if (-not (Test-Path $dir)) { New-Item -ItemType Directory $dir | Out-Null }
        [System.IO.File]::WriteAllText($expFile, $actual)
        return @{ Status = "updated" }
    }

    if (-not (Test-Path $expFile)) {
        return @{ Status = "skip"; Reason = "no .expected file (run with -Update to create)" }
    }

    $expected = (Get-Content $expFile -Raw).Replace("`r`n", "`n").Replace("`r", "`n").Trim()
    if ($actual -eq $expected) { return @{ Status = "pass" } }
    return @{ Status = "fail"; Reason = "output mismatch"; Expected = $expected; Actual = $actual }
}

function Invoke-ErrTest($file, $rel) {
    $tmpExe = "$tmpDir\err_test.exe"
    if (Test-Path $tmpExe) { Remove-Item $tmpExe -Force }
    $r = Exec-Timed $compiler "build `"$rel`" -o `"$tmpExe`" -self-link -O1" $TimeoutSec
    if ($r.TimedOut) { return @{ Status = "fail"; Reason = "timeout (compiler hung on bad input)" } }
    if ($r.ExitCode -ne 0) { return Check-ErrorPattern $file $r.Output }
    return @{ Status = "fail"; Reason = "expected compiler error, compilation succeeded" }
}

# ── run all tests ─────────────────────────────────────────────────────────────
Write-C "" "White"
Write-C "=== AXIOM Test Harness ===" "Cyan"
Write-Host "  Compiler : $compiler"
Write-Host "  Mode     : $Mode"
Write-Host "  Tests    : $($allTests.Count)"
if ($Filter) { Write-Host "  Filter   : $Filter" }
if ($Update)  { Write-Host "  Updating golden files" }
Write-C "" "White"

foreach ($f in $allTests) {
    $rel    = (Resolve-Path -Relative $f.FullName).TrimStart(".\")
    $isErr  = $f.Name.StartsWith("err_")
    $first  = Read-FirstLine $f.FullName
    $label  = $rel.PadRight(56)

    if ($first -match "HARNESS:\s*SKIP") {
        Write-C "  SKIP  $label  (# HARNESS: SKIP)" "DarkGray"
        $nSkip++
        continue
    }

    # Mode filtering
    $shouldRun = switch ($Mode) {
        "air"   { -not $isErr }
        "build" { $true }
        "run"   { $true }
        "err"   { $isErr }
        "all"   { $true }
    }
    if (-not $shouldRun) {
        $nSkip++
        continue
    }

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $result = switch ($Mode) {
        "air"   { Invoke-AirTest $f $rel }
        "build" { Invoke-BuildTest $f $rel $isErr }
        "run"   { Invoke-RunTest $f $rel $isErr }
        "err"   { Invoke-ErrTest $f $rel }
        "all"   { if ($isErr) { Invoke-ErrTest $f $rel } else { Invoke-RunTest $f $rel $false } }
    }
    $sw.Stop()
    $ms = [math]::Round($sw.Elapsed.TotalMilliseconds)

    $timeTag = if ($ms -ge 1000) { " ($([math]::Round($ms/1000,1))s)" } else { " (${ms}ms)" }

    switch ($result.Status) {
        "pass"    { Write-C "  PASS  $label$timeTag" "Green";  $nPass++ }
        "updated" { Write-C "  UPDT  $label$timeTag" "Cyan";   $nUpdate++ }
        "skip"    {
            $skipReason = if ($result.Reason) { "  <- $($result.Reason)" } else { "" }
            Write-C "  SKIP  $label$skipReason" "DarkYellow"
            $nSkip++
        }
        "fail"    {
            Write-C "  FAIL  $label  <- $($result.Reason)" "Red"
            $nFail++
            $failures.Add($rel)
            if ($Verbose) {
                if ($result.Output) {
                    Write-C "        [output]" "DarkRed"
                    $result.Output | Select-Object -Last 10 | ForEach-Object { Write-Host "          $_" }
                }
                if ($result.Expected) {
                    Write-C "        [expected]" "DarkGray"
                    $result.Expected.Split("`n") | Select-Object -First 5 | ForEach-Object { Write-Host "          $_" }
                    Write-C "        [actual]" "DarkGray"
                    $result.Actual.Split("`n") | Select-Object -First 5 | ForEach-Object { Write-Host "          $_" }
                }
            }
        }
    }
}

$sw_total.Stop()

# ── summary ───────────────────────────────────────────────────────────────────
Write-C "" "White"
Write-C "=== Results ===" "Cyan"
Write-C ("  PASS   : " + $nPass)   "Green"
if ($nUpdate -gt 0) { Write-C ("  UPDATED: " + $nUpdate) "Cyan" }
if ($nFail -gt 0) {
    Write-C ("  FAIL   : " + $nFail)   "Red"
} else {
    Write-C ("  FAIL   : " + $nFail)   "Green"
}
Write-C ("  SKIP   : " + $nSkip)   "DarkYellow"
Write-C ("  TOTAL  : " + ($nPass + $nFail + $nSkip + $nUpdate)) "White"
Write-C ("  TIME   : " + [math]::Round($sw_total.Elapsed.TotalSeconds, 1) + "s") "DarkGray"

if ($nFail -gt 0) {
    Write-C "" "White"
    Write-C "Failed:" "Red"
    foreach ($f in $failures) { Write-C "  - $f" "Red" }
    exit 1
}
exit 0
