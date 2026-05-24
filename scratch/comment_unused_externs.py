import os
import re

files = [
    "bootstrap/stage1/air_builder.ax",
    "bootstrap/stage1/cgen.ax",
    "bootstrap/stage1/ctgc.ax",
    "bootstrap/stage1/escape.ax",
    "bootstrap/stage1/intern.ax",
    "bootstrap/stage1/mono.ax",
    "bootstrap/stage1/ownership.ax",
    "bootstrap/stage1/parser.ax",
    "bootstrap/stage1/resolver.ax",
    "bootstrap/stage1/ssa_opt.ax",
    "bootstrap/stage1/wasm.ax",
    "bootstrap/stage1/main_air.ax",
    "bootstrap/stage1/typecheck.ax",
    "bootstrap/stage1/typetable.ax"
]

def clean_file(path):
    print(f"Cleaning {path}...")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
        
    new_content = content
    
    # 1. Comment out unused extern declarations
    new_content = re.sub(
        r'^(extern\s+"C"\s+fn\s+printf\b)',
        r'// \1',
        new_content,
        flags=re.MULTILINE
    )
    new_content = re.sub(
        r'^(extern\s+"C"\s+fn\s+puts\b)',
        r'// \1',
        new_content,
        flags=re.MULTILINE
    )
    new_content = re.sub(
        r'^(extern\s+"C"\s+fn\s+fprintf\b)',
        r'// \1',
        new_content,
        flags=re.MULTILINE
    )
    new_content = re.sub(
        r'^(extern\s+"C"\s+fn\s+snprintf\b)',
        r'// \1',
        new_content,
        flags=re.MULTILINE
    )
    new_content = re.sub(
        r'^(extern\s+"C"\s+fn\s+sprintf\b)',
        r'// \1',
        new_content,
        flags=re.MULTILINE
    )
    
    # 2. Specifically for wasm.ax:
    if "wasm.ax" in path:
        # Replace snprintf with ax_snprintf_local
        new_content = re.sub(r'\bsnprintf\b', 'ax_snprintf_local', new_content)
        # Replace fprintf with ax_fprintf_local
        new_content = re.sub(r'\bfprintf\b', 'ax_fprintf_local', new_content)
        
    if new_content != content:
        with open(path, "w", encoding="utf-8") as f:
            f.write(new_content)
        print(f"  Modified {path}")

for f in files:
    if os.path.exists(f):
        clean_file(f)
