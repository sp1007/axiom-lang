import os

def count_lines(dir_path):
    for root, dirs, files in os.walk(dir_path):
        for f in files:
            if f.endswith('.ax'):
                full_path = os.path.join(root, f)
                with open(full_path, 'r', encoding='utf-8') as fh:
                    lines = fh.readlines()
                    print(f"{f}: {len(lines)}")

count_lines('std')
