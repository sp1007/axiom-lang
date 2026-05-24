with open('compiler/lexer/lexer.go', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

for idx, line in enumerate(lines[:100]):
    print(f"{idx+1}: {line.strip()}")
