# scripts/regen_concat.ps1
# Regenerate ONLY the concatenated frontend (bootstrap/stage1/tmp_concatenated_air.ax)
# from the committed per-file sources. Shared by build_native.ps1 (which does not want
# to also rebuild the gcc stage1 that rebuild_stage1.ps1 forces). Keeps the frontend
# file list in one place.

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
$len = (Get-Item $concatenatedPath).Length
Write-Host "[concat] OK: $len bytes" -ForegroundColor Green
