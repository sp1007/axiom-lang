with open('compiler/sema/resolver.go', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

for idx, line in enumerate(lines):
    if 'registerGenericTemplate' in line:
        print(f"{idx+1}: {line.strip()}")
