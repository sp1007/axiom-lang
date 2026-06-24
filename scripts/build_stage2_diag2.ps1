# Build stage2 with the air_builder [B26] FIELD_EXPR-decision diagnostic, then
# compile t_strlen and t_strmulti with it to capture the self-call signature.
$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root
"STAGE2-DIAG2 BUILD START $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
& .\bin\axc_stage1.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage2_diag2.exe -self-link -O1 *> compiler_stage2_diag2.log
"STAGE2-DIAG2 BUILD DONE exit=$LASTEXITCODE $(Get-Date -Format 'HH:mm:ss')"
if (Test-Path bin\axc_stage2_diag2.exe) {
    Copy-Item axiom_temp.obj bin\obj_stage2_diag2.obj -Force
    "obj saved"
    & .\bin\axc_stage2_diag2.exe build bin\t_strlen.ax -o bin\t_strlen_s2d2.exe -self-link -O1 > bin\s2d2_strlen.out.txt 2>&1
    "=== stage2-diag2 t_strlen [B26] lines: ==="
    (Get-Content bin\s2d2_strlen.out.txt) -replace "`0","" | Select-String -Pattern '\[B26' -SimpleMatch | ForEach-Object { $_.Line }
} else {
    "BUILD FAILED"
}
