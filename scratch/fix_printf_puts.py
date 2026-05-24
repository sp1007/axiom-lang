import os
import re

frontend_files = [
    "bootstrap/stage1/token.ax",
    "bootstrap/stage1/lexer.ax",
    "bootstrap/stage1/ast.ax",
    "bootstrap/stage1/intern.ax",
    "bootstrap/stage1/parser.ax",
    "bootstrap/stage1/resolver.ax",
    "bootstrap/stage1/typetable.ax",
    "bootstrap/stage1/mono.ax",
    "bootstrap/stage1/typecheck.ax",
    "bootstrap/stage1/connection_graph.ax",
    "bootstrap/stage1/ownership.ax",
    "bootstrap/stage1/escape.ax",
    "bootstrap/stage1/ctgc.ax",
    "bootstrap/stage1/alias_reuse.ax",
    "bootstrap/stage1/air.ax",
    "bootstrap/stage1/air_builder.ax",
    "bootstrap/stage1/ssa_opt.ax",
    "bootstrap/stage1/cgen.ax",
    "bootstrap/stage1/wasm.ax",
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
    "bootstrap/stage1/fmt.ax",
    "bootstrap/stage1/main_air.ax"
]

print_helpers = """fn print_to_file(stream: ptr[void], s: str):
    fputs(s, stream)

fn print_i64_to_file(stream: ptr[void], val: i64):
    if val == 0:
        print_to_file(stream, "0")
        return
        
    mut v := val
    mut is_neg := false
    if v < 0:
        is_neg = true
        v = -v
        
    mut buf := @alloc(32) as ptr[u8]
    mut len := 0
    while v > 0:
        let digit = (v % 10) as u8
        buf[len] = '0' as u8 + digit
        len = len + 1
        v = v / 10
        
    if is_neg:
        buf[len] = '-' as u8
        len = len + 1
        
    // Reverse buffer
    mut i := 0
    while i < len / 2:
        let tmp = buf[i]
        buf[i] = buf[len - 1 - i]
        buf[len - 1 - i] = tmp
        i = i + 1
        
    buf[len] = 0 as u8
    
    // Convert to str and print
    let s = std.string.slice(buf as str, 0, len as i64)
    print_to_file(stream, s)
    @free(buf)

fn ax_fprintf_local(stream: ptr[void], fmt: str, a1: i64, a2: i64, a3: i64, a4: i64, a5: i64, a6: i64, a7: i64, a8: i64) -> i32:
    let s_ptr = fmt as ptr[u8]
    let s_len = std.string.len(fmt)
    mut arg_idx := 0
    mut i := 0 as i64
    mut start := 0 as i64
    
    while i < s_len:
        if s_ptr[i] == '%' as u8:
            if i > start:
                print_to_file(stream, std.string.slice(fmt, start, i))
                
            if i + 1 < s_len:
                if s_ptr[i + 1] == '%' as u8:
                    print_to_file(stream, "%")
                    i = i + 1
                elif s_ptr[i + 1] == 'd' as u8:
                    mut val := 0 as i64
                    if arg_idx == 0:
                        val = a1
                    elif arg_idx == 1:
                        val = a2
                    elif arg_idx == 2:
                        val = a3
                    elif arg_idx == 3:
                        val = a4
                    elif arg_idx == 4:
                        val = a5
                    elif arg_idx == 5:
                        val = a6
                    elif arg_idx == 6:
                        val = a7
                    elif arg_idx == 7:
                        val = a8
                    
                    print_i64_to_file(stream, val)
                    arg_idx = arg_idx + 1
                    i = i + 1
                elif s_ptr[i + 1] == 's' as u8:
                    mut val_ptr_raw := 0 as i64
                    if arg_idx == 0:
                        val_ptr_raw = a1
                    elif arg_idx == 1:
                        val_ptr_raw = a2
                    elif arg_idx == 2:
                        val_ptr_raw = a3
                    elif arg_idx == 3:
                        val_ptr_raw = a4
                    elif arg_idx == 4:
                        val_ptr_raw = a5
                    elif arg_idx == 5:
                        val_ptr_raw = a6
                    elif arg_idx == 6:
                        val_ptr_raw = a7
                    elif arg_idx == 7:
                        val_ptr_raw = a8
                    
                    if val_ptr_raw != 0:
                        let c_ptr = val_ptr_raw as ptr[u8]
                        mut len := 0 as i64
                        while c_ptr[len] != 0 as u8:
                            len = len + 1
                        print_to_file(stream, std.string.slice(c_ptr as str, 0, len))
                    arg_idx = arg_idx + 1
                    i = i + 1
                start = i + 1
        i = i + 1
        
    if i > start:
        print_to_file(stream, std.string.slice(fmt, start, i))
        
    return 0"""

def process_file(path):
    print(f"Processing {path}...")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    lines = content.splitlines()
    new_lines = []
    
    is_cgen = "cgen.ax" in path

    for line in lines:
        trimmed = line.strip()
        # Comment out extern C declarations for printf, puts, sprintf, snprintf, fprintf
        is_extern_to_remove = False
        if trimmed.startswith('extern "C" fn'):
            # check function name
            match = re.search(r'extern\s+"C"\s+fn\s+([a-zA-Z0-9_]+)\b', trimmed)
            if match:
                fn_name = match.group(1)
                if fn_name in ["printf", "puts", "sprintf", "snprintf", "fprintf"]:
                    is_extern_to_remove = True
        
        if is_extern_to_remove:
            print(f"  Commenting out: {line}")
            new_lines.append("// " + line)
            if is_cgen:
                # Add our print helpers right here in cgen.ax
                new_lines.append(print_helpers)
        else:
            # Replace active calls
            # Use regex to find call boundaries securely but simply.
            # We want to replace puts(...) and printf(...) and fprintf(...)
            # But only when not part of declaration or comment
            if not trimmed.startswith("//") and not trimmed.startswith("#") and not trimmed.startswith('extern "C" fn'):
                # Replacement for puts( -> ax_puts_local(
                line = re.sub(r'\bputs\b(?=\s*\()', "ax_puts_local", line)
                # Replacement for printf( -> ax_printf_local(
                line = re.sub(r'\bprintf\b(?=\s*\()', "ax_printf_local", line)
                # Replacement for fprintf( -> ax_fprintf_local(
                line = re.sub(r'\bfprintf\b(?=\s*\()', "ax_fprintf_local", line)
                # Replacement for snprintf( -> ax_snprintf_local(
                line = re.sub(r'\bsnprintf\b(?=\s*\()', "ax_snprintf_local", line)
            new_lines.append(line)

    new_content = "\n".join(new_lines) + "\n"
    with open(path, "w", encoding="utf-8") as f:
        f.write(new_content)

for p in frontend_files:
    if os.path.exists(p):
        process_file(p)
    else:
        print(f"File not found: {p}")
