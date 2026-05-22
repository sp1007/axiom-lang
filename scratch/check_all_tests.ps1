# scratch/check_all_tests.ps1
$ErrorActionPreference = "Continue"
$files = Get-ChildItem -Path tests -Filter "*.ax" -Recurse
Write-Host "Found $($files.Count) AXIOM files to check..." -ForegroundColor Cyan

$passed = 0
$failed = 0

foreach ($f in $files) {
    if ($f.FullName -match "scratch" -or $f.FullName -match "tmp") {
        continue
    }
    $rel = Resolve-Path -Relative $f.FullName
    Write-Host "Checking $rel... " -NoNewline
    
    # Run axc check
    $out = & .\bin\axc.exe check $f.FullName 2>&1
    $exitCode = $LASTEXITCODE
    
    if ($exitCode -eq 0) {
        Write-Host "PASSED" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host $out -ForegroundColor DarkRed
        $failed++
    }
}

Write-Host "Finished: $passed passed, $failed failed." -ForegroundColor Cyan
