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

concatenated = concatenate_files(['std/scheduler.ax', 'std/scheduler_test.ax'])
print(f"Offset 4437 of concatenated scheduler source:")
print(concatenated[4437-50:4437+50])
