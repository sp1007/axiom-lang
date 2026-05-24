import os
import re

files_to_process = [
    "bootstrap/stage1/air_builder.ax",
    "bootstrap/stage1/cgen.ax",
    "bootstrap/stage1/linker.ax",
    "bootstrap/stage1/main_air.ax",
    "bootstrap/stage1/wasm.ax",
    "bootstrap/stage1/x86_asm_emitter.ax",
    "bootstrap/stage1/x86_coff.ax",
    "bootstrap/stage1/x86_elf64.ax",
    "bootstrap/stage1/typecheck.ax",
    "bootstrap/stage1/typetable.ax"
]

def wrap_fputs(line):
    if "extern" in line or "fn " in line:
        return line
    match = re.search(r'\bfputs\s*\((.*)\)', line)
    if not match:
        return line
    args_str = match.group(1)
    in_quotes = False
    comma_idx = -1
    for idx, char in enumerate(args_str):
        if char == '"':
            in_quotes = not in_quotes
        elif char == ',' and not in_quotes:
            comma_idx = idx
            break
            
    if comma_idx != -1:
        lhs = args_str[:comma_idx].strip()
        rhs = args_str[comma_idx+1:].strip()
        if not lhs.startswith("get_str_ptr"):
            return line.replace(match.group(0), f"fputs(get_str_ptr({lhs}), {rhs})")
    return line

def wrap_fopen(line):
    if "extern" in line or "fn " in line:
        return line
    match = re.search(r'\bfopen\s*\((.*)\)', line)
    if not match:
        return line
    args_str = match.group(1)
    in_quotes = False
    comma_idx = -1
    for idx, char in enumerate(args_str):
        if char == '"':
            in_quotes = not in_quotes
        elif char == ',' and not in_quotes:
            comma_idx = idx
            break
            
    if comma_idx != -1:
        lhs = args_str[:comma_idx].strip()
        rhs = args_str[comma_idx+1:].strip()
        if not rhs.startswith("get_str_ptr"):
            return line.replace(match.group(0), f"fopen({lhs}, get_str_ptr({rhs}))")
    return line

def wrap_atof(line):
    if "extern" in line or "fn " in line:
        return line
    match = re.search(r'\batof\s*\(([^)]+)\)', line)
    if not match:
        return line
    arg = match.group(1).strip()
    if not arg.startswith("get_str_ptr"):
        return line.replace(match.group(0), f"atof(get_str_ptr({arg}))")
    return line

def process_file(path):
    print(f"Processing extern C signatures in {path}...")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    new_content = content
    
    # 1. Replace extern declarations exactly
    new_content = new_content.replace(
        'extern "C" fn fopen(filename: ptr[void], mode: str) -> ptr[void]',
        'extern "C" fn fopen(filename: ptr[void], mode: ptr[u8]) -> ptr[void]'
    )
    new_content = new_content.replace(
        'extern "C" fn fopen(filename: ptr[u8], mode: str) -> ptr[void]',
        'extern "C" fn fopen(filename: ptr[u8], mode: ptr[u8]) -> ptr[void]'
    )
    new_content = new_content.replace(
        'extern "C" fn fputs(s: str, stream: ptr[void]) -> i32',
        'extern "C" fn fputs(s: ptr[u8], stream: ptr[void]) -> i32'
    )
    new_content = new_content.replace(
        'extern "C" fn atof(s: str) -> f64',
        'extern "C" fn atof(s: ptr[u8]) -> f64'
    )
    
    # 2. Process calls line-by-line
    lines = new_content.splitlines()
    processed_lines = []
    for line in lines:
        if not line.strip().startswith("//"):
            line = wrap_fopen(line)
            line = wrap_fputs(line)
            line = wrap_atof(line)
        processed_lines.append(line)
        
    new_content = "\n".join(processed_lines) + "\n"

    if new_content != content:
        with open(path, "w", encoding="utf-8") as f:
            f.write(new_content)
        print(f"  Modified {path}")

for p in files_to_process:
    if os.path.exists(p):
        process_file(p)
    else:
        print(f"File not found: {p}")
