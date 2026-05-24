with open('std/scheduler.ax', 'r', encoding='utf-8') as fh:
    scheduler = fh.read()

with open('std/scheduler_test.ax', 'r', encoding='utf-8') as fh:
    test = fh.read()

# Simulate concatenation imports and body as in concatenateAxiomFiles
# Let's just do a basic concatenation or check what's at offset 715 of std/scheduler.ax itself first
print(f"Offset 715 of std/scheduler.ax:")
print(scheduler[715-50:715+50])
