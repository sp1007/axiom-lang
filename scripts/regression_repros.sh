#!/usr/bin/env bash
# Regression gate for the known-bug repros (chain #24..#31 + Family C / RFC 0001).
# Run after ANY codegen/optimizer/ABI change before trusting the next step.
# Each case: build with the given compiler (default bin/axc_stage1.exe) at -O1,
# run, and compare stdout (CMP=out) or exit code (CMP=exit) to the expected value.
#
# NOTE: on this Windows box, Windows Defender realtime adds ~50s per self-link
# build (CPU ~0, pure scan). That is NOT a hang — keep the per-build timeout >=120s.
set -u
cd "$(dirname "$0")/.."
AXC="${AXC:-bin/axc_stage1.exe}"
TIMEOUT="${TIMEOUT:-150}"
pass=0; fail=0; failed=""

# rows: name | cmp(out|exit) | expected
rows=(
  "tsp|exit|7"
  "tsp2|exit|9"
  "tsp3|exit|12"
  "tstruct_abi|out|A=7 B=12 C=15 D=6 E=99 "
  "t_movrr|out|072 137 229 "
  "t_modrm|out|229"
  "t_cp2|out|7"
  "t_cpaddr|exit|7"
  "t_cse|exit|98"
  "t_param5|out|A38"
  "t_strip|out|a.b len exit print"
  "t_alias|out|A=1 B=0 "
  "t_mutstr|exit|9"
  "t_method|exit|24"
  "t_enum_np|exit|6"
  "t_enum|exit|42"
  "t_adt2|exit|104"
  "t_adt3|exit|19"
  "t_builtin_opt|exit|15"
  "t_litinfer|exit|111"
  "t_udiv|exit|5"
  "t_ucmp|exit|6"
  "t_f32|exit|5"
  "t_switch|exit|4"
  "t_cassign|exit|12"
  "t_unsafe|exit|20"
  "t_pow|exit|59"
  "t_defer|exit|7"
  "t_mask|exit|15"
  "t_opover|exit|44"
  "t_bytes|exit|69"
  "t_u128|exit|91"
  "t_opmix|exit|13"
  "t_mul128|exit|7"
  "t_fcmp|exit|15"
  # t_farg returns 300, but bash $? is 8-bit: 300 & 0xFF = 44. 65a17b8 set 300
  # (verified under PowerShell, full 32-bit exit) which can never pass HERE.
  "t_farg|exit|44"
  "t_fret|exit|75"
  "t_math|exit|127"
  "t_bignum|exit|63"
  "t_bn256|exit|31"
  "t_bf128|exit|15"
  "t_bndyn|exit|99"
  "t_mathx|exit|28"
  "t_b44|exit|54"
  "t_erf|exit|15"
  "t_gamma|exit|63"
  "t_complex|exit|63"
  "t_vec|exit|63"
  "t_stats|exit|63"
  "t_numtheory|exit|127"
  "t_combinatorics|exit|127"
  "t_interpolation|exit|127"
  "t_geometry|exit|127"
  "t_probability|exit|127"
  "t_numtheory2|exit|127"
  "t_coordinates|exit|127"
  "t_color|exit|127"
  "t_statx|exit|127"
  "t_signal|exit|127"
  "t_ml|exit|127"
  "t_mathfill|exit|127"
  "t_bitmath|exit|127"
  "t_millerrabin|exit|127"
  "t_special|exit|127"
  "t_numerical|exit|127"
  "t_geometry2|exit|127"
  "t_colorhsl|exit|127"
  "t_rng|exit|127"
  "t_decimal|exit|127"
  "t_quaternion|exit|127"
  "t_matrix|exit|127"
  "t_vec4|exit|127"
  "t_fspill|exit|78"
  "t_fctor|exit|40"
  "t_ffree|exit|110"
  "fmeth|exit|110"
  "fsmall|exit|37"
  "t_trecip|exit|8"
  "t_ieeerem|exit|8"
  "t_zeta|exit|5"
  "t_quatrot|exit|8"
  "t_fft|exit|18"
  "fp0i|exit|5"
  "fpBi|exit|5"
  "fpf|exit|14"
  "t_nmf|exit|9"
  "t_numeric|exit|5"
  "t_sort|exit|4"
  "t_qsort|exit|6"
  "t_sort2|exit|5"
  "t_func|exit|6"
  "t_lambda|exit|5"
  "t_modcollide|exit|101"
  "t_hashi64|exit|42"
  "t_nestedgen|exit|42"
  "t_optstruct|exit|40"
  "t_optnested|exit|42"
  "t_structeq|exit|42"
  "t_arrlit|exit|15"
  "t_closurecap|exit|6"
  "t_genericstruct2|exit|40"
  "t_mutparam|exit|22"
  "t_structarrfield|exit|119"
  "t_retstructfield|exit|11"
  "t_nested3field|exit|13"
  "t_arroption|exit|110"
  "t_vecoption|exit|110"
  "t_ovcollide|exit|37"
  "t_variantstruct|exit|17"
  "t_underscore|exit|27"
  "t_arrelemaddr|exit|66"
  "t_fldelemaddr|exit|150"
  "t_forrange|exit|28"
  "t_forcollect|exit|56"
  "t_forcontinue|exit|52"
  "t_forstruct|exit|129"
  "t_forsum|exit|46"
  "t_mfvariant|exit|137"
  "t_strvariant|exit|44"
  "t_selftype|exit|13"
  "t_gensumctor|exit|54"
  "t_optmethod|exit|42"
  "t_strconcat|exit|15"
  "t_globals|exit|113"
  "sc1|exit|42"
  "sc2|exit|42"
  "scv|exit|3"
  "scw|exit|5"
  "scstress|exit|32"
  "t_matchret|exit|30"
  "t_missingret|reject|"
  "t_partialmatch|reject|"
  "t_undefname|reject|"
  "t_retmismatch|reject|"
  "t_strnumop|reject|"
  "t_argstrnum|reject|"
  "t_cmpstrnum|reject|"
  "t_forvec|exit|60"
  "t_rune|exit|65"
  "t_bytesview|exit|204"
  "t_forbytes|exit|198"
  "t_forstr|exit|15"
  "t_strslice|exit|142"
  "t_strmethod|exit|147"
  "t_strsplit|exit|46"
  "t_strutf8|exit|74"
  "t_tostr|exit|88"
  "t_tostru64|exit|19"
  "t_tostrf|exit|91"
  "t_tohex|exit|123"
  "t_toradix|exit|99"
  "t_numufcs|exit|11"
  "t_numufcsneg|reject|"
  "t_numufcswide|exit|7"
  "t_strviews|exit|244"
  "t_geninstfield|exit|44"
  "t_cseredef|exit|12"
  "t_hashiter|exit|15"
  "t_hashcontainskey|exit|7"
  "t_crosstypereject|reject|"
  "t_vecoptmatch|exit|42"
  "t_genfieldufcs|exit|4"
  "t_sumdispatch|exit|31"
  "t_optunwrapadd|exit|112"
  "t_optarithreject|reject|"
  "t_arrret|exit|70"
  "t_arrargmismatch|reject|"
  "t_ctorarrfield|exit|110"
  "t_freefncollision|exit|15"
  "t_freefncollision_arity|exit|105"
  "t_inlinearmstmt|reject|"
  "t_optstrmatch|exit|10"
  "t_gentree|exit|15"
  "t_variantshadow|reject|"
  "t_nomethod|reject|"
  "t_nonexhenum|reject|"
  "t_exhaustenum|exit|3"
  "t_arityfew|reject|"
  "t_aritymany|reject|"
  "t_arityok|exit|21"
  "t_badfield|reject|"
  "t_goodfield|exit|5"
  "t_vecindex|exit|7"
  "t_vecset|exit|42"
  "t_scalarmember|reject|"
  "t_floatneg|exit|74"
  "t_constshift|exit|35"
  "t_methodmany|reject|"
  "t_methodfew|reject|"
  "t_methodok|exit|40"
  "t_forgotunwrap|reject|"
  "t_unwrapok|exit|42"
  "t_staticcall|exit|41"
  "t_globalinit|exit|210"
  "t_bracelit|reject|"
  "t_optarray|exit|40"
  "t_resultarray|exit|40"
  "t_optarrayiter|exit|123"
  "t_optarraymethod|exit|40"
  "t_structoptfield|exit|42"
  "t_vecvec|exit|42"
  "t_hashvecval|exit|42"
  "t_optresult|exit|42"
  "t_vecstructopt|exit|42"
  "t_matchshadow|exit|42"
  "t_freefnfind|exit|42"
  "t_freefncontains|exit|42"
  "t_f32boundary|exit|42"
  "t_f32aggregate|exit|42"
  "t_f32generic|exit|42"
  "t_f32argcoerce|exit|42"
  "t_f32binop|exit|42"
  # feature-combo probe oracles (banked 2026-07-12, batch 7)
  "t_genpairopt|exit|77"
  "t_arrstructidx|exit|102"
  "t_negdivmod|exit|25"
  "t_scshortcirc|exit|12"
  "t_genfnopt|exit|63"
  "t_strfieldcat|exit|5"
  "t_resstrmatch|exit|3"
  "t_treeoptchild|exit|15"
  "t_nestbreakcont|exit|16"
  "t_hashstrkey|exit|30"
  "t_u64cmp|exit|7"
  "t_optveciter|exit|6"
  "t_optvecnest|exit|42"
  "t_vectupmix|exit|47"
  "t_foldu32wrap|exit|205"
  "t_foldi8wrap|exit|94"
  "t_u32fieldwrap|exit|205"
  "t_genswaphet|exit|42"
  "t_deepnestmut|exit|42"
  "t_ucmphighbit|exit|42"
  "t_vecofvec|exit|42"
  "t_deepopt|exit|99"
  # RFC 0017 P2 — aggregate (struct/array/tuple) module-level globals
  "t_globstruct|exit|30"
  "t_globarr|exit|80"
  "t_globnested|exit|40"
  "t_globarriter|exit|119"
  "t_globbig|exit|65"
  "t_globpass|exit|42"
  # RFC 0017 P2 — aggregate-global interaction surface
  "t_globarrstruct|exit|118"
  "t_globmethod|exit|42"
  "t_globgeneric|exit|42"
  "t_globmixed|exit|40"
  "t_globi32arr|exit|130"
  "t_globstrfield|exit|10"
  # RFC 0017 — no-initializer globals default to a zeroed .data slot (parser desync fix)
  "t_globnoinit|exit|7"
  "t_globstructnoinit|exit|42"
  "t_globarrnoinit|exit|40"
  # RFC 0017 P2 — aggregate-global write paths (user-code block-copy, not just init)
  "t_globwholeassign|exit|42"
  "t_globg2g|exit|42"
  "t_globreturn|exit|42"
  "t_globrmwloop|exit|20"
  "t_globarrfn|exit|50"
  "t_globopt|exit|42"
  "t_globresult|exit|42"
  "t_globsum|exit|42"
  "t_globstr|exit|42"
  "t_globstridx|exit|42"
  # RFC 0017 P2 — str/16B-inline & pointer-repr embedded inside aggregate globals (deep crosses)
  "t_globarrstr|exit|15"
  "t_globoptstr|exit|8"
  "t_globarrstructstr|exit|20"
  "t_globarropt|exit|22"
  "t_globinitorder|exit|40"
  # RFC 0022 — tuple literal expressions (anonymous struct + .N access)
  "t_tuple|exit|66"
  # RFC 0022 P2 — tuple TYPE annotations: annotated let / return / param / global
  "t_tuple2|exit|84"
  # RFC 0022 P3 — tuple destructuring `let (a, b) = EXPR` (single-eval + wildcard + mut)
  "t_tuple3|exit|50"
  # RFC 0022 P4 — chained tuple field access `t.N.M` (float-selector split)
  "t_tuple4|exit|10"
  # RFC 0022 — tuple-typed struct field (ctor-field element coercion)
  "t_tupfield|exit|33"
  # assert() builtin now works on the native path (symbol map + msg-arg synthesis)
  "t_assert|exit|7"
  # RFC 0022 — tuple cross-feature coverage (probe-banked 2026-07-13)
  "t_tupret|exit|117"
  "t_tupstr|exit|12"
  "t_tupnest|exit|5"
  "t_tupmethod|exit|42"
  "t_tuparr|exit|9"
  "t_tupcall|exit|66"
  # inline lambda -> generic fn/method (infer U from lambda return type)
  "t_lambdagen|exit|42"
  "t_lambdamap|exit|42"
  "t_lambdamap2|exit|1"
  "t_lambdazip|exit|42"
  # Vec higher-order methods (map/filter/fold) with inline lambdas
  "t_vecmap|exit|45"
  "t_vecfilter|exit|42"
  "t_vecfold|exit|42"
  "t_vecmapbool|exit|2"
  "t_vecchain|exit|50"
  "t_vecany|exit|1"
  "t_vecall|exit|42"
  "t_vecfind|exit|42"
  # Vec predicate/order combinators (count/position/take_while/skip_while/reverse)
  "t_veccount|exit|42"
  "t_vecposition|exit|42"
  "t_vectakewhile|exit|12"
  "t_vecskipwhile|exit|25"
  "t_vecreverse|exit|42"
  # closure zero-capture scan no longer flags global variant constructors
  "t_lambdaflatmap|exit|42"
  "t_lambdavariant|exit|7"
  # lambda body may read a field / call a method on its own parameter
  "t_lambdafield|exit|24"
  "t_lambdamethod|exit|42"
  # lambda/HOF chaining coverage (free-fn call, Option chain, nested Vec, map->bool)
  "t_lambdafreefn|exit|42"
  "t_optfindmap|exit|42"
  "t_vecnestedfold|exit|42"
  "t_vecmapall|exit|42"
  # Result combinators with inline lambdas (3 type params)
  "t_resmap|exit|42"
  "t_resmaperr|exit|42"
  # RFC 0023 if-expressions (scalar/elif/nested/str/struct + lambda body; rejects)
  "t_ifexpr|exit|42"
  "t_ifexprelif|exit|42"
  "t_ifexprnest|exit|42"
  "t_ifexprstr|exit|3"
  "t_ifexprstruct|exit|42"
  "t_ifexprlambda|exit|42"
  "t_ifexproptnone|exit|42"
  "t_ifexprusersum|exit|42"
  "t_ifexprfield|exit|42"
  "t_ifexprvecpush|exit|42"
  "t_ifexpridx|exit|42"
  "t_ifexprret|exit|42"
  "t_ifexprnoelse|reject|"
  "t_ifexprmismatch|reject|"
  # aggregate Vec element + field access inside a fold lambda
  "t_vecstructfold|exit|42"
  # if-expr result as a match scrutinee (Some/None) + 3-level nesting
  "t_ifexprmatchscrut|exit|42"
  "t_ifexprdeepnest|exit|42"
  # recursion returning a struct accumulator (by-ref aggregate) + mutual recursion
  "t_recurstruct|exit|42"
  "t_mutualrecur|exit|42"
  # generic free-fn called with explicit type-args (enclosing type param): f[T]()
  "t_genfnexpltypearg|exit|42"
  # tuple type flowing through a generic fn param (mono substitution + name mangling)
  "t_gentuple|exit|42"
  "t_tupctor|exit|60"
  # tuple LITERAL pushed through a generic method arg: Vec[(i64,i64)].push((10,20))
  "t_vectup|exit|42"
  # Vec whose element is a tuple containing a generic param: Vec[(i64,T)] in a generic fn
  "t_vecgentup|exit|1"
  # generic enumerate end-to-end: Vec[T] -> Vec[(i64,T)], push (i,elem), read .0/.1
  "t_vecenum|exit|63"
  # guard: concrete-tuple Vec inside a generic body (must stay working)
  "t_vecconctup|exit|1"
  # doubly-nested generic tuple element: Vec[(i64,(i64,T))] in a generic fn (is_generic __tup recursion)
  "t_nestedgentup|exit|30"
  # Vec.enumerate stdlib HOF: yields (index, element) tuples (unblocked by aa419a2)
  "t_vecenumerate|exit|63"
  # Vec.partition stdlib HOF: (matching, non-matching) tuple-of-Vecs return
  "t_vecpartition|exit|62"
  # Vec.zip stdlib HOF: element-wise pairs Vec[(T,U)] (free-fn-shadow overload hole fixed)
  "t_veczip|exit|62"
  # user free-fn overload: concrete 1-arg vs generic-first-param 2-arg (no over-match)
  "t_useroverload|exit|42"
  # tuple-literal RHS coerced on ASSIGNMENT to a tuple-typed lvalue (global reassign)
  "t_tupassign|exit|42"
  # array-literal RHS coerced on ASSIGNMENT to an [i64;N] lvalue (element width)
  "t_arrassign|exit|42"
  # variant-ctor RHS (Some(tuple)) coerced on ASSIGNMENT to an Option[tuple] lvalue
  "t_optassign|exit|42"
  # tuple as a HashMap VALUE: insert + get().unwrap() + read both fields
  "t_hmtupval|exit|42"
  # chained Vec HOF: map (double) then fold (sum)
  "t_hofchain|exit|28"
  # value-if producing a tuple, bound via let-destructure
  "t_ifexprtupdestr|exit|42"
  # Vec accessor methods (is_empty / extend)
  "t_vecisempty|exit|42"
  "t_vecextend|exit|42"
  # RFC 0015 P2 — EscapeAnalyser now ACTIVE (crash fixed); escaping + non-escaping locals
  "t_escape|exit|41"
  # RFC 0015 P2 — non-escaping aggregate local in a loop with an early-return path
  "t_escapeloop|exit|42"
  # RFC 0015 P3 — compile-time free oracle (also run under -ctgc-free by ctgc_free_check.sh)
  "t_ctgcfree|exit|42"
  # RFC 0014 — drop-glue oracle: no drop without the flag (0); drop fires 42x under
  # -ctgc-free (see ctgc_free_check.sh)
  "t_drop|exit|0"
  # RFC 0024 block strings """...""" (len/multi-line/escapes/byte/embedded-quotes)
  "t_blockstr|exit|5"
  "t_blockstrml|exit|5"
  "t_blockstresc|exit|42"
  "t_blockstrbyte|exit|65"
  "t_blockstrquote|exit|11"
  "elfglob|exit|55"
  "t_parse|exit|40"
  "t_shapehof|exit|33"
  "t_castwidth|exit|15"
  "t_divpow2|exit|16"
  "t_immfold|exit|42"
  "t_noneinfer|exit|42"
  "t_uninferreject|reject|"
  "t_recstructreject|reject|"
  "t_matchnonsum|reject|"
  "t_letstrmismatch|reject|"
  "t_genfloatret|exit|42"
  "t_genmethodfloat|exit|42"
  "t_floatcvt|exit|10"
  "t_licm|exit|44"
  "t_loopcall|exit|140"
  "t_loopstruct|exit|9"
  "t_licmhoist|exit|235"
  "t_loopcross|exit|10"
  "t_licmchain|exit|123"
  "t_licmunroll|exit|18"
)

for row in "${rows[@]}"; do
  IFS='|' read -r name cmp want <<< "$row"
  src="bin/${name}.ax"
  [ -f "$src" ] || { echo "SKIP $name (no $src)"; continue; }
  out_exe="/tmp/reg_${name}.exe"
  rm -f "$out_exe"
  timeout "$TIMEOUT" "$AXC" build "$src" -o "$out_exe" -O1 >/dev/null 2>&1
  if [ "$cmp" = "reject" ]; then
    # Expect the compiler to REJECT this program (diagnostic, no exe emitted).
    if [ ! -f "$out_exe" ]; then echo "PASS $name (rejected)"; pass=$((pass+1)); else echo "FAIL $name (expected rejection, got exe)"; fail=$((fail+1)); failed="$failed $name"; fi
    continue
  fi
  if [ ! -f "$out_exe" ]; then echo "FAIL $name (build produced no exe)"; fail=$((fail+1)); failed="$failed $name"; continue; fi
  got_out=$("$out_exe" 2>/dev/null); got_exit=$?
  if [ "$cmp" = "exit" ]; then got="$got_exit"; else got="$got_out"; fi
  if [ "$got" = "$want" ]; then
    echo "PASS $name ($cmp=$got)"; pass=$((pass+1))
  else
    echo "FAIL $name ($cmp: got '$got' want '$want')"; fail=$((fail+1)); failed="$failed $name"
  fi
done

# --- Optimizer level guard (-O2/-O3) --------------------------------------
# The rows above all build at -O1. This block guards the -O2/-O3 structural
# loop passes (LICM [RFC 0025 model-A] + strength_reduction; loop_unroll stays
# disabled): a loop program must (a) NOT crash the compiler and (b) produce the
# same result at -O2/-O3 as at -O1. Regression for bug-o2-o3-loop-compile-crash
# and RFC 0025 (sound LICM re-enablement).
# Each entry crosses a different AIR pattern with the loop optimizers:
#   t_licm       plain induction-var sum loop
#   t_loopcall   function call inside the loop body (caller-save/clobber)
#   t_loopstruct struct field read-modify-write inside the loop (aggregate)
#   t_licmhoist  genuinely-invariant sub-expr -> LICM MUST hoist, stay correct
#   t_loopcross  loop-crossing value used in the condition (RFC 0016 liveness)
#   t_licmchain  two-link invariant chain -> both hoist (copy_prop before LICM)
#   t_licmunroll const-trip accumulator loop -> sound full unroll (loop-carried)
opt_rows=(
  "t_foldu32wrap|205"
  "t_licm|44"
  "t_loopcall|140"
  "t_loopstruct|9"
  "t_licmhoist|235"
  "t_loopcross|10"
  "t_licmchain|123"
  "t_licmunroll|18"
)
for opt in O2 O3; do
  for orow in "${opt_rows[@]}"; do
    IFS='|' read -r oname owant <<< "$orow"
    osrc="bin/${oname}.ax"
    [ -f "$osrc" ] || continue
    oe="/tmp/reg_${oname}_${opt}.exe"; rm -f "$oe"
    timeout "$TIMEOUT" "$AXC" build "$osrc" -o "$oe" -${opt} >/dev/null 2>&1
    if [ ! -f "$oe" ]; then
      echo "FAIL ${oname}@-${opt} (compiler produced no exe — crash?)"; fail=$((fail+1)); failed="$failed ${oname}@-${opt}"
    else
      "$oe" >/dev/null 2>&1; ge=$?
      if [ "$ge" = "$owant" ]; then echo "PASS ${oname}@-${opt} (exit=$ge)"; pass=$((pass+1)); else echo "FAIL ${oname}@-${opt} (exit: got '$ge' want '$owant')"; fail=$((fail+1)); failed="$failed ${oname}@-${opt}"; fi
    fi
  done
done

echo "=== regression: $pass passed, $fail failed ==="
[ -n "$failed" ] && echo "FAILED:$failed"
[ "$fail" -eq 0 ] && echo "REGRESSION_OK" || echo "REGRESSION_FAIL"
