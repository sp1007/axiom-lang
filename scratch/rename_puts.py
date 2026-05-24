import re

filepath = r"d:\projects\compiler\Axiom\bootstrap\stage1\main_air.ax"

with open(filepath, "r", encoding="utf-8") as f:
    content = f.read()

# Replace puts( with ax_puts_local(
new_content = re.sub(r"\bputs\(", "ax_puts_local(", content)

with open(filepath, "w", encoding="utf-8") as f:
    f.write(new_content)

print("Renamed all puts calls to ax_puts_local successfully!")
