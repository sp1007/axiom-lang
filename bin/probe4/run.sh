#!/bin/sh
# usage: run.sh file.ax ...
for f in "$@"; do
  b=$(basename "$f" .ax)
  for O in 0 1; do
    out="bin/probe4/$b.O$O.exe"
    log=$(bin/axc_native.exe build "$f" -o "$out" -O$O 2>&1)
    if [ $? -ne 0 ]; then
      echo "$b -O$O COMPILE-FAIL: $(echo "$log" | tail -3 | tr '\n' ' ')"
      continue
    fi
    "$out" >/dev/null 2>&1
    echo "$b -O$O exit=$?"
  done
done
