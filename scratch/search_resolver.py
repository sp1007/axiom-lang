with open('compiler/sema/resolver.go', 'r', encoding='utf-8') as fh:
    content = fh.read()

import re
matches = re.finditer(r'func \(nr \*NameResolver\) .*', content)
for m in matches:
    print(m.group(0))
