with open('std/scheduler.ax', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

for idx, line in enumerate(lines):
    if 'self as i64' in line or 'self as ptr' in line:
        print(f"{idx+1}: {line.strip()}")
