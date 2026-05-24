import os

def concatenate_files(paths):
    imports = []
    body = []
    for p in paths:
        with open(p, 'r', encoding='utf-8') as fh:
            content = fh.read()
        lines = content.split('\n')
        for line in lines:
            trimmed = line.strip()
            if trimmed.startswith('import '):
                imports.append(line)
            else:
                body.append(line)
    
    unique_imports = []
    seen = set()
    for imp in imports:
        trimmed = imp.strip()
        if trimmed not in seen:
            seen.add(trimmed)
            unique_imports.append(imp)
            
    result = '\n'.join(unique_imports) + '\n\n' + '\n'.join(body)
    return result

concatenated = concatenate_files(['std/result.ax', 'std/sync.ax', 'std/sync_test.ax'])
lines = concatenated.split('\n')
for i in range(115, 145):
    print(f"{i+1}: {lines[i]}")
