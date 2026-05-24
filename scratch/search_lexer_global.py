with open('compiler/lexer/lexer.go', 'r', encoding='utf-8') as fh:
    content = fh.read()

import re
matches = re.finditer(r'var\s+\w+\s+=.*', content)
for m in matches:
    print(m.group(0))
