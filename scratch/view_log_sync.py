with open('C:/Users/sp/.gemini/antigravity/brain/06276bd4-3781-4952-af76-f0fce47b0d4b/.system_generated/tasks/task-250.log', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

found = False
for idx, line in enumerate(lines):
    if 'failed to compile AXIOM sync' in line:
        found = True
        print(f"Line {idx+1}:")
        for i in range(max(0, idx - 5), min(len(lines), idx + 25)):
            print(f"{i+1}: {lines[i].strip()}")
