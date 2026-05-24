with open('C:/Users/sp/.gemini/antigravity/brain/06276bd4-3781-4952-af76-f0fce47b0d4b/.system_generated/tasks/task-250.log', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

for idx, line in enumerate(lines):
    if 'failed to compile AXIOM scheduler' in line:
        print(f"Line {idx+1}:")
        for i in range(max(0, idx - 10), min(len(lines), idx + 20)):
            print(f"{i+1}: {lines[i].strip()}")
