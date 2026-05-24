$paths = @(
    "bootstrap/stage1/token.ax",
    "bootstrap/stage1/lexer.ax",
    "bootstrap/stage1/ast.ax",
    "bootstrap/stage1/intern.ax",
    "bootstrap/stage1/parser.ax",
    "bootstrap/stage1/resolver.ax",
    "bootstrap/stage1/typecheck.ax",
    "bootstrap/stage1/air.ax",
    "bootstrap/stage1/air_builder.ax",
    "bootstrap/stage1/x86_regs.ax",
    "bootstrap/stage1/x86_selector.ax",
    "bootstrap/stage1/x86_regalloc.ax",
    "bootstrap/stage1/x86_asm_emitter.ax",
    "bootstrap/stage1/x86_modrm.ax",
    "bootstrap/stage1/x86_encoding.ax",
    "bootstrap/stage1/x86_emitter.ax",
    "bootstrap/stage1/x86_elf64.ax",
    "bootstrap/stage1/x86_coff.ax",
    "bootstrap/stage1/linker.ax",
    "bootstrap/stage1/main_air.ax"
)
$imports = @()
$body = @()
foreach ($p in $paths) {
    $content = Get-Content $p -Raw
    $lines = $content -split "\r?\n"
    foreach ($line in $lines) {
        $trimmed = $line.Trim()
        if ($trimmed.StartsWith("import ")) {
            $imports += $line
        } else {
            $body += $line
        }
    }
}
$uniqueImports = $imports | Sort-Object -Unique
$result = ($uniqueImports -join "`n") + "`n`n" + ($body -join "`n")
Set-Content -Path bootstrap/stage1/tmp_concatenated_air.ax -Value $result -NoNewline
Write-Host "Concatenation complete!"
