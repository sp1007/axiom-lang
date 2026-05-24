with open('compiler/sema/resolver.go', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

found = False
for idx, line in enumerate(lines):
    if 'func (nr *NameResolver) resolveType' in line:
        found = True
        start = idx
    if found:
        print(f"{idx+1}: {line.strip()}")
        if line.startswith('}'):
            break
