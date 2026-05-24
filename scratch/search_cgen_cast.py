with open('codegen/cgen/exprs.go', 'r', encoding='utf-8') as fh:
    lines = fh.readlines()

found = False
for idx, line in enumerate(lines):
    if 'case ast.NodeCastExpr' in line:
        found = True
        start = idx
    if found:
        print(f"{idx+1}: {line.strip()}")
        if 'case' in line and idx > start + 5:
            break
