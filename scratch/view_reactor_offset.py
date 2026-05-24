with open('std/reactor.ax', 'r', encoding='utf-8') as fh:
    reactor = fh.read()

print(f"Offset 4437 of std/reactor.ax:")
print(reactor[4437-50:4437+50])
