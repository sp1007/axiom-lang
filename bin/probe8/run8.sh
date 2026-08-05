#!/bin/sh
# usage: sh bin/probe8/run8.sh <name>   (name without .ax)
cd /d/projects/compiler/Axiom
n="$1"
for O in O0 O1; do
  out="bin/probe8/$n.$O.exe"
  rm -f "$out"
  log=$(bin/axc_native.exe build "bin/probe8/$n.ax" -o "$out" -$O 2>&1)
  rc=$?
  if [ ! -f "$out" ]; then
    echo "$n  -$O  BUILD-FAIL(rc=$rc): $(echo "$log" | grep -i -m2 'error\|Error' | tr '\n' ' ')"
  else
    ./"$out" >/dev/null 2>&1
    echo "$n  -$O  exit=$?"
  fi
done
