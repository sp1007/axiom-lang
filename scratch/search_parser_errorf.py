with open('compiler/parser/parser.go', 'r', encoding='utf-8') as fh:
    content = fh.read()

import re
matches = re.finditer(r'func \(p \*Parser\) errorf\(.*', content)
for m in matches:
    print(m.group(0))
