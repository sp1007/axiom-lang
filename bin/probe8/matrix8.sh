#!/bin/sh
cd /d/projects/compiler/Axiom
for n in "$@"; do
  line="$n"
  for O in "-O0" "" "-O1"; do
    lbl=$O; [ -z "$lbl" ] && lbl="(default)"
    out="bin/probe8/$n.m.exe"; rm -f "$out"
    log=$(bin/axc_native.exe build "bin/probe8/$n.ax" -o "$out" $O 2>&1)
    if [ ! -f "$out" ]; then
      line="$line | $lbl=REJECT"
    else
      ./"$out" >/dev/null 2>&1
      line="$line | $lbl=$?"
    fi
  done
  echo "$line"
done
