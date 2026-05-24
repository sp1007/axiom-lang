import os
import re

files_to_process = [
    "bootstrap/stage1/x86_asm_emitter.ax",
    "bootstrap/stage1/cgen.ax",
    "bootstrap/stage1/linker.ax",
    "bootstrap/stage1/air_builder.ax",
    "bootstrap/stage1/main_air.ax"
]

complex_replacements = [
    (r'\bself\.tree\.src\.ptr\b', 'get_str_ptr(self.tree.src)'),
    (r'\bself\.tree\.src\s+as\s+ptr\s*\[\s*u8\s*\]', 'get_str_ptr(self.tree.src)'),
]

str_variables = [
    "dst", "src", "sym_name", "cc_str", "byte_reg", "addr_str", "fn_name",
    "name", "struct_name", "mangled_name", "param_type", "ret_type", "t_name", "s_name",
    "print_buf", "formatted", "output_file_nt", "temp_obj_nt", "cmd_buf",
    "path", "arg_text", "text", "tok_str", "inner_callee_name", "op_token", "op_text"
]

def process_file(path):
    print(f"Processing casts in {path}...")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    new_content = content
    
    # 1. Apply complex replacements first
    for pattern, replacement in complex_replacements:
        new_content, count = re.subn(pattern, replacement, new_content)
        if count > 0:
            print(f"  Replaced complex: {pattern} -> {replacement} ({count} times)")
            
    # 2. Apply regular variable replacements for .ptr and as ptr[u8]
    for var in str_variables:
        # Match pattern like \bvar\.ptr\b
        pattern_ptr = r'\b' + var + r'\.ptr\b'
        replacement = f"get_str_ptr({var})"
        new_content, count1 = re.subn(pattern_ptr, replacement, new_content)
        
        # Match pattern like \bvar\s+as\s+ptr\s*\[\s*u8\s*\]
        pattern_as = r'\b' + var + r'\s+as\s+ptr\s*\[\s*u8\s*\]'
        new_content, count2 = re.subn(pattern_as, replacement, new_content)
        
        if count1 + count2 > 0:
            print(f"  Replaced {var}: {count1} .ptr, {count2} as ptr[u8] -> {replacement}")

    if new_content != content:
        with open(path, "w", encoding="utf-8") as f:
            f.write(new_content)

for p in files_to_process:
    if os.path.exists(p):
        process_file(p)
    else:
        print(f"File not found: {p}")
