import os
import re

def search_qualified_types(dir_path):
    # Regex to find std.something in type contexts
    pattern = re.compile(r'(:\s*|->\s*|ptr\s*\[)\s*std\.\w+')
    for root, dirs, files in os.walk(dir_path):
        for f in files:
            if f.endswith('.ax'):
                full_path = os.path.join(root, f)
                with open(full_path, 'r', encoding='utf-8') as fh:
                    lines = fh.readlines()
                    for idx, line in enumerate(lines):
                        if pattern.search(line) or 'std.' in line and ('ptr[' in line or 'struct' in line):
                            print(f"{f}:{idx+1}: {line.strip()}")

search_qualified_types('std')
