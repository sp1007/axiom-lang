with open('compiler/sema/resolver.go', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

for idx in range(120, 165):
    print(f"{idx+1}: {lines[idx].strip()}")
