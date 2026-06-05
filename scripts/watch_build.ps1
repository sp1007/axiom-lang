# scripts/watch_build.ps1
# Chay ngam: ghi bao cao trang thai build ra file moi 10 phut.
# Cach dung: Start-Job { d:\projects\compiler\Axiom\scripts\watch_build.ps1 }
# Xem bao cao:  Get-Content d:\projects\compiler\Axiom\build_status.log -Tail 30

param([int]$IntervalSeconds = 600)

$root  = "d:\projects\compiler\Axiom"
$log   = "$root\build_status.log"

while ($true) {
    $ts   = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $proc = Get-Process -Name "axc_stage1" -ErrorAction SilentlyContinue
    $s2   = Test-Path "$root\bin\axc_stage2_native.exe"
    $s3   = Test-Path "$root\bin\axc_stage3_native.exe"

    $line = "[$ts]"

    if ($proc) {
        $cpu = [math]::Round($proc.CPU)
        $ram = [math]::Round($proc.WorkingSet64 / 1MB)
        $line += " BUILD=RUNNING PID=$($proc.Id) CPU=${cpu}s RAM=${ram}MB"
    } else {
        $line += " BUILD=STOPPED"
    }

    $line += " Stage2=$(if ($s2) { 'DONE' } else { 'PENDING' })"
    $line += " Stage3=$(if ($s3) { 'DONE' } else { 'PENDING' })"

    Add-Content -Path $log -Value $line
    Write-Host $line

    if ($s2 -and $s3) {
        Add-Content -Path $log -Value "[$ts] === BOTH STAGES COMPLETE ==="
        Write-Host "Both stages done - stopping watcher."
        break
    }

    Start-Sleep -Seconds $IntervalSeconds
}
