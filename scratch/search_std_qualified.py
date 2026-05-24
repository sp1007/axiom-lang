import os

def search_std(dir_path):
    for root, dirs, files in os.walk(dir_path):
        for f in files:
            if f.endswith('.ax'):
                full_path = os.path.join(root, f)
                with open(full_path, 'r', encoding='utf-8') as fh:
                    lines = fh.readlines()
                    for idx, line in enumerate(lines):
                        if 'std.mem.alloc' in line:
                            # print only if it's not a simple import line
                            if not line.strip().startswith('import std.mem.alloc') or '{' in line:
                                print(f"{f}:{idx+1}: {line.strip()}")

search_std('std')
