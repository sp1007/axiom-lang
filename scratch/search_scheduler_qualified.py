with open('std/scheduler.ax', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

for idx, line in enumerate(lines):
    if 'std.mem.alloc' in line:
        print(f"{idx+1}: {line.strip()}")
