#!/usr/bin/env bash
# RFC 0035 P2 — cross-library symbol-collision oracle.
#
# The property under test is the one the whole RFC exists for: two SEPARATELY COMPILED
# libraries that each define a function of the same name must both be callable from one
# program, and each call must reach its own library's body.
#
# This cannot live in regression_repros.sh, which builds a single `bin/<name>.ax`. It needs
# three sources, `--auto-lib`, and two child compilations. It is also the ONLY thing that can
# judge this change: the self-host build uses no libraries, so it is inert, and RFC 0035 §7bis
# records what that inertness is worth -- an earlier increment of this same RFC reached a
# clean A==B and 557/557 while being a structural no-op. A green gate is not evidence here;
# `30` is.
#
# The sources are generated in the REPOSITORY ROOT and removed again, which is ugly and
# deliberate. Imports resolve relative to the current directory, and the compiler must run
# from the root to find std/*.ax, so a library imported as `axlibcola` has to be
# `./axlibcola.ax`. The tidier `bin/libcol/` layout was tried first and does not work:
# `import bin.libcol.libcola` compiles the libraries correctly but the CALL
# `bin.libcol.libcola.helper()` fails with "undefined name 'bin'". That is a PRE-EXISTING
# limitation of multi-segment non-std module paths -- the shipped compiler fails it
# identically -- and is not something this RFC introduced or needs to fix here.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_native.exe}"
# Own scratch directory for outputs, per the contention trap in
# [[infra-defender-build-throttle]]. Only the .ax/.lib inputs must live in the root.
TMP="bin/_libcoltmp"
mkdir -p "$TMP"
pass=0; fail=0

cleanup() {
  rm -f axlibcola.ax axlibcolb.ax axappcol.ax \
        axlibcola.lib axlibcolb.lib axlibcola.lib.manifest axlibcolb.lib.manifest
}
trap cleanup EXIT

cat > axlibcola.ax <<'EOF'
pub fn helper() -> i32:
    return 10
EOF

cat > axlibcolb.ax <<'EOF'
pub fn helper() -> i32:
    return 20
EOF

cat > axappcol.ax <<'EOF'
import axlibcola
import axlibcolb

fn main() -> i32:
    return axlibcola.helper() + axlibcolb.helper()
EOF

# Stale `.lib`s MUST go first. The staleness manifest hashes only the library SOURCE, so a
# `.lib` produced by a compiler with a different mangling scheme is considered fresh and is
# not rebuilt -- the flag day RFC 0035 §7 calls out. Without this the oracle would report a
# failure that a rebuild fixes, or worse, pass on a stale pair.
rm -f axlibcola.lib axlibcolb.lib axlibcola.lib.manifest axlibcolb.lib.manifest
rm -f "$TMP/axappcol.exe"

"$AXC" build axappcol.ax -o "$TMP/axappcol.exe" --auto-lib -O1 > "$TMP/build.log" 2>&1
if [ ! -f "$TMP/axappcol.exe" ]; then
  echo "FAIL twolib_build (no exe; see $TMP/build.log)"
  grep -i "^error" "$TMP/build.log" | head -5
  fail=$((fail+1))
else
  "$TMP/axappcol.exe" >/dev/null 2>&1; got=$?
  if [ "$got" = "30" ]; then
    echo "PASS twolib_call (exit=30: each call reached its own library)"
    pass=$((pass+1))
  else
    # 20 or 40 means both calls bound to ONE definition -- the silent first-match mis-link
    # this RFC was written from. Before P2 this case did not even link: the app called
    # ax_helper__m<sym_idx>, a name no library ever defined.
    echo "FAIL twolib_call (exit: got '$got' want '30')"
    fail=$((fail+1))
  fi
fi

# The definition side: each library must emit its own qualified symbol. Checked on the
# artifact rather than inferred from the exit code, because a correct 30 could in principle
# come from some other mechanism, and because this is the assertion that the qualifier is
# derived from the module name rather than from any per-compilation index.
for m in axlibcola axlibcolb; do
  if [ ! -f "$m.lib" ]; then
    echo "FAIL ${m}_symbol (no .lib produced)"; fail=$((fail+1)); continue
  fi
  if strings -a "$m.lib" | grep -q "^ax_${m}_helper$"; then
    echo "PASS ${m}_symbol (ax_${m}_helper)"
    pass=$((pass+1))
  else
    echo "FAIL ${m}_symbol (expected ax_${m}_helper; got: $(strings -a "$m.lib" | grep '^ax_.*helper' | tr '\n' ' '))"
    fail=$((fail+1))
  fi
done

# Runtime/ABI names must NOT be qualified (RFC 0035 §4): they are bound BY NAME with no call
# edge, so qualifying one turns a working binding into an unresolved external. This asserts
# the exclusion actually holds on a real library rather than trusting the predicate.
if strings -a axlibcola.lib | grep -q "^ax_sum_layout_is_pointer$"; then
  echo "PASS abi_names_unqualified (ax_sum_layout_is_pointer kept its fixed name)"
  pass=$((pass+1))
else
  echo "FAIL abi_names_unqualified (runtime shim was renamed)"
  fail=$((fail+1))
fi

# The interface still advertises the LOCAL name: qualification is a link-time contract and
# must not leak into the source-level API, or `axlibcola.helper` would stop resolving.
if strings -a axlibcola.lib | grep -q "^F helper 0 -> i32$"; then
  echo "PASS iface_local_name (F helper 0 -> i32)"
  pass=$((pass+1))
else
  echo "FAIL iface_local_name (iface no longer exports the local name)"
  fail=$((fail+1))
fi

echo "lib_collision_check: $pass passed, $fail failed"
[ "$fail" = "0" ] || exit 1
