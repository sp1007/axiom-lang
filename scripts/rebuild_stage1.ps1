# scripts/rebuild_stage1.ps1
# Regenerate concatenated frontend (picks up committed x86 codegen fixes +
# uncommitted air_builder.ax field-expr fix) and rebuild stage1 via the C backend (ASCII only).
# This is the fix for the stale-stage1 problem: stage1 must contain the str
# codegen fixes so the stage2 it generates is correct.

$ErrorActionPreference = "Continue"
$root = Resolve-Path "$PSScriptRoot\.."
Set-Location $root

function Concatenate-Axiom-Files {
    param([string[]]$Paths, [string]$OutputPath)
    $imports = [System.Collections.Generic.List[string]]::new()
    $body = [System.Collections.Generic.List[string]]::new()
    foreach ($p in $Paths) {
        $lines = Get-Content $p
        foreach ($line in $lines) {
            $trimmed = $line.Trim()
            if ($trimmed.StartsWith("import ")) {
                if ($trimmed -like "import bootstrap.stage1.*") { continue }
                $imports.Add($line)
            } else {
                $body.Add($line)
            }
        }
    }
    $uniqueImports = [System.Collections.Generic.List[string]]::new()
    $importMap = @{}
    foreach ($imp in $imports) {
        $trimmed = $imp.Trim()
        if (-not $importMap.ContainsKey($trimmed)) {
            $importMap[$trimmed] = $true
            $uniqueImports.Add($imp)
        }
    }
    $result = ($uniqueImports -join "`n") + "`n`n" + ($body -join "`n")
    [System.IO.File]::WriteAllText($OutputPath, $result)
}

$frontendFiles = @(
    "bootstrap/stage1/print_helpers.ax", "bootstrap/stage1/token.ax",
    "bootstrap/stage1/lexer.ax", "bootstrap/stage1/ast.ax",
    "bootstrap/stage1/intern.ax", "bootstrap/stage1/parser.ax",
    "bootstrap/stage1/resolver.ax", "bootstrap/stage1/typetable.ax",
    "bootstrap/stage1/mono.ax", "bootstrap/stage1/typecheck.ax",
    "bootstrap/stage1/connection_graph.ax", "bootstrap/stage1/ownership.ax",
    "bootstrap/stage1/escape.ax", "bootstrap/stage1/ctgc.ax",
    "bootstrap/stage1/alias_reuse.ax", "bootstrap/stage1/air.ax",
    "bootstrap/stage1/air_builder.ax", "bootstrap/stage1/ssa_opt.ax",
    "bootstrap/stage1/cgen.ax", "bootstrap/stage1/wasm.ax",
    "bootstrap/stage1/x86_regs.ax", "bootstrap/stage1/x86_selector.ax",
    "bootstrap/stage1/x86_regalloc.ax", "bootstrap/stage1/x86_asm_emitter.ax",
    "bootstrap/stage1/x86_modrm.ax", "bootstrap/stage1/x86_encoding.ax",
    "bootstrap/stage1/x86_emitter.ax", "bootstrap/stage1/x86_elf64.ax",
    "bootstrap/stage1/x86_coff.ax", "bootstrap/stage1/linker.ax",
    "bootstrap/stage1/fmt.ax", "bootstrap/stage1/lsp.ax",
    "bootstrap/stage1/main_air.ax"
)

$concatenatedPath = "bootstrap/stage1/tmp_concatenated_air.ax"
Write-Host "[concat] Regenerating $concatenatedPath ..." -ForegroundColor Green
Concatenate-Axiom-Files -Paths $frontendFiles -OutputPath $concatenatedPath

Write-Host "[stage1] Building bin/axc_stage1.exe via C backend (-O1) ..." -ForegroundColor Green
if (Test-Path bin/axc_stage1.exe) { Remove-Item bin/axc_stage1.exe }
cmd /c "bin\axc.exe build bootstrap\stage1\tmp_concatenated_air.ax -o bin\axc_stage1.exe -O1 > compiler_stage1.log 2>&1"
if (-not (Test-Path bin/axc_stage1.exe)) {
    Write-Host "[stage1] FAILED -- see compiler_stage1.log" -ForegroundColor Red
    Get-Content compiler_stage1.log -Tail 30
    exit 1
}
$s1len = (Get-Item bin/axc_stage1.exe).Length
Write-Host "[stage1] OK: $s1len bytes" -ForegroundColor Green
