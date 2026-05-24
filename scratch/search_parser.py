with open('compiler/parser/parser.go', 'r', encoding='utf-8') as fh:
    content = fh.read()

import re
matches = re.findall(r'func \(p \*Parser\) parseType\w*\(.*', content)
for m in matches:
    print(m)
