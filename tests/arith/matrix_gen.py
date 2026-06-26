#!/usr/bin/env python3
# Structured matrix arithmetic test for AXIOM (RFC 0006) — systematic coverage of
# the 10 numeric types (i8 i16 i32 i64 u8 u16 u32 u64 f32 f64): single-type ops,
# all 100 ordered type-pair casts (both directions), and mixed arithmetic.
#
# Unlike arith_gen.py (random fuzz), this is DETERMINISTIC and labelled, so a
# failure points at an exact (type/op/cast) combination. Self-checking: prints
# `F <id> <got> <exp>  # <label>` per mismatch + a PASS/FAIL summary; exit=fails.
#
# Oracle reused from arith_gen.py. Run: bash tests/arith/run_matrix.sh
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from arith_gen import (TINFO, ITYPES, FTYPES, wrap, as_cast, to_i64,
                       trunc_div, trunc_mod, lit, flit, fbits, f32_round, mask)

INT = [t[0] for t in ITYPES]
ALL = INT + FTYPES

def is_float(t): return t in ("f32", "f64")

def fround(t, x): return f32_round(x) if t == "f32" else float(x)

def repvals(ty):
    """A few representative values of `ty` (in range), incl. edges."""
    if is_float(ty):
        return [0.0, 1.5, -2.25, 1024.0]
    w, s = TINFO[ty]
    if s:
        return [0, 7, -3, (1 << (w - 1)) - 1, -(1 << (w - 1))]
    return [0, 7, mask(w), mask(w) // 2]

def src_val(ty, v):
    """AXIOM literal source for value v of type ty (as a typed expr)."""
    if is_float(ty):
        return f"({flit(v)} as {ty})"
    return f"({lit(v)} as {ty})"

# A test case = (label, axiom_expr_src, result_type, expected_i64)
def cases():
    cs = []
    def add(label, src, rty, expv):
        cs.append((label, src, rty, expv))

    # ---- 1. single-type integer arithmetic + bitwise ----
    for ty in INT:
        w, s = TINFO[ty]
        a, b = (7, 3) if (not s or w > 8) else (7, 3)
        av, bv = wrap(a, ty), wrap(b, ty)
        A, B = src_val(ty, av), src_val(ty, bv)
        add(f"{ty}_add", f"({A} + {B})", ty, to_i64(wrap(av + bv, ty)))
        add(f"{ty}_sub", f"({A} - {B})", ty, to_i64(wrap(av - bv, ty)))
        add(f"{ty}_mul", f"({A} * {B})", ty, to_i64(wrap(av * bv, ty)))
        add(f"{ty}_div", f"({A} / {B})", ty, to_i64(trunc_div(av, bv) if s else (av & mask(w)) // (bv & mask(w))))
        add(f"{ty}_mod", f"({A} % {B})", ty, to_i64(trunc_mod(av, bv) if s else (av & mask(w)) % (bv & mask(w))))
        add(f"{ty}_and", f"({A} & {B})", ty, to_i64(wrap((av & mask(w)) & (bv & mask(w)), ty)))
        add(f"{ty}_or",  f"({A} | {B})", ty, to_i64(wrap((av & mask(w)) | (bv & mask(w)), ty)))
        add(f"{ty}_xor", f"({A} ^ {B})", ty, to_i64(wrap((av & mask(w)) ^ (bv & mask(w)), ty)))
        add(f"{ty}_shl", f"({A} << (2 as {ty}))", ty, to_i64(wrap(av << 2, ty)))
        sh = (av & mask(w)) >> 1 if not s else (av >> 1)
        add(f"{ty}_shr", f"({A} >> (1 as {ty}))", ty, to_i64(wrap(sh, ty)))
        # comparisons -> bool (compare in i64 domain: bool true=1)
        add(f"{ty}_lt", f"({B} < {A})", "i64", 1)
        add(f"{ty}_gt", f"({A} > {B})", "i64", 1)
        add(f"{ty}_eq", f"({A} == {A})", "i64", 1)
        add(f"{ty}_ne", f"({A} != {B})", "i64", 1)

    # ---- 2. single-type float arithmetic ----
    for ty in FTYPES:
        av, bv = 5.5, 2.0
        A, B = src_val(ty, av), src_val(ty, bv)
        rnd = (lambda z: f32_round(z)) if ty == "f32" else (lambda z: z)
        add(f"{ty}_fadd", f"({A} + {B})", ty, fbits(rnd(av + bv), ty))
        add(f"{ty}_fsub", f"({A} - {B})", ty, fbits(rnd(av - bv), ty))
        add(f"{ty}_fmul", f"({A} * {B})", ty, fbits(rnd(av * bv), ty))
        add(f"{ty}_fdiv", f"({A} / {B})", ty, fbits(rnd(av / bv), ty))

    # ---- 3. all 100 ordered type-pair casts (value-preserving where in range) ----
    for t1 in ALL:
        for t2 in ALL:
            v = repvals(t1)[1]  # a small representative value (7 or 1.5)
            src = f"(({src_val(t1, v)}) as {t2})"
            if is_float(t2):
                # cast to float: int->float = float(value); float->float = round
                fv = fround(t2, v if is_float(t1) else wrap(v, t1))
                add(f"cast_{t1}_to_{t2}", src, t2, fbits(fv, t2))
            else:
                # cast to int: float->int truncates toward zero; int->int wraps
                if is_float(t1):
                    iv = int(v)  # trunc toward zero
                else:
                    iv = wrap(v, t1)
                add(f"cast_{t1}_to_{t2}", src, t2, to_i64(wrap(iv, t2)))

    # ---- 4. mixed arithmetic via explicit cast to a common type ----
    pairs = [("i8", "i32"), ("i16", "i64"), ("u8", "u64"), ("u16", "u32"),
             ("i32", "i64"), ("u32", "i64"), ("i8", "i16"), ("u8", "u32")]
    for (t1, t2) in pairs:
        v1, v2 = repvals(t1)[1], repvals(t2)[1]
        # compute in t2
        a = wrap(as_cast(wrap(v1, t1), t2), t2)
        b = wrap(v2, t2)
        src = f"(({src_val(t1, v1)} as {t2}) + {src_val(t2, v2)})"
        add(f"mix_{t1}_{t2}_add", src, t2, to_i64(wrap(a + b, t2)))
    # int -> float mixed (explicit)
    for (it, ft) in [("i32", "f32"), ("u64", "f64"), ("i8", "f64"), ("u8", "f32")]:
        iv = repvals(it)[1]
        fv = 2.5
        rnd = (lambda z: f32_round(z)) if ft == "f32" else (lambda z: z)
        a = rnd(float(wrap(iv, it)))
        b = rnd(fv)
        src = f"(({src_val(it, iv)} as {ft}) + {src_val(ft, fv)})"
        add(f"mix_{it}_{ft}_add", src, ft, fbits(rnd(a + b), ft))

    # ---- 5. overflow / wrap (two's complement, well-defined per RFC 0006 §6) ----
    def wrapcase(label, ty, a, op, b):
        w, s = TINFO[ty]
        A, B = src_val(ty, wrap(a, ty)), src_val(ty, wrap(b, ty))
        if op == "+": r = a + b
        elif op == "-": r = a - b
        else: r = a * b
        add(f"wrap_{label}", f"({A} {op} {B})", ty, to_i64(wrap(r, ty)))
    wrapcase("u8_max_inc", "u8", 255, "+", 1)        # -> 0
    wrapcase("u8_zero_dec", "u8", 0, "-", 1)         # -> 255
    wrapcase("i8_max_inc", "i8", 127, "+", 1)        # -> -128
    wrapcase("i8_min_dec", "i8", -128, "-", 1)       # -> 127
    wrapcase("u16_max_inc", "u16", 65535, "+", 1)    # -> 0
    wrapcase("u32_zero_dec", "u32", 0, "-", 1)       # -> 4294967295
    wrapcase("i32_max_inc", "i32", (1 << 31) - 1, "+", 1)   # -> -2147483648
    wrapcase("u8_mul", "u8", 200, "*", 2)            # -> 144
    wrapcase("i16_mul_ovf", "i16", 1234, "*", 56)    # 69104 -> wraps i16

    return cs

CHUNK = 80

def main():
    cs = cases()
    out = []
    out.append('// AUTO-GENERATED by matrix_gen.py — structured numeric matrix test (RFC 0006)')
    out.append(f'// {len(cs)} cases over 10 types. F lines = mismatches; P=/F= summary; exit=fails.')
    out.append('extern "C" fn putchar(c: i32) -> i32')
    out.append('extern "C" fn fflush(s: ptr[void]) -> i32')
    out.append('')
    out.append('fn emit_i64(v: i64):')
    out.append('    mut x := v')
    out.append('    if x == 0 as i64:')
    out.append("        putchar('0' as i32)")
    out.append('        return')
    out.append('    if x < 0 as i64:')
    out.append("        putchar('-' as i32)")
    out.append('        x = 0 as i64 - x')
    out.append('    mut buf := @alloc(32) as ptr[u8]')
    out.append('    mut k := 0 as i64')
    out.append('    while x > 0 as i64:')
    out.append("        buf[k] = ('0' as u8) + (x % 10 as i64) as u8")
    out.append('        k = k + 1 as i64')
    out.append('        x = x / 10 as i64')
    out.append('    while k > 0 as i64:')
    out.append('        k = k - 1 as i64')
    out.append('        putchar(buf[k] as i32)')
    out.append('    @free(buf as ptr[u8])')
    out.append('')
    out.append('fn report(id: i64, got: i64, exp: i64):')
    out.append("    putchar('F' as i32)")
    out.append("    putchar(' ' as i32)")
    out.append('    emit_i64(id)')
    out.append("    putchar(' ' as i32)")
    out.append('    emit_i64(got)')
    out.append("    putchar(' ' as i32)")
    out.append('    emit_i64(exp)')
    out.append("    putchar(10 as i32)")
    out.append('')

    nchunks = (len(cs) + CHUNK - 1) // CHUNK
    idx = 0
    for c in range(nchunks):
        out.append(f'fn chunk_{c}() -> i64:')
        out.append('    mut fails := 0 as i64')
        for _ in range(CHUNK):
            if idx >= len(cs):
                break
            label, src, rty, expv = cs[idx]
            if is_float(rty):
                bptr = "ptr[i64]" if rty == "f64" else "ptr[i32]"
                out.append(f'    mut g{idx}: {rty} = {src}   // {label}')
                out.append(f'    let bits{idx}: i64 = ((&g{idx} as {bptr})[0]) as i64')
                out.append(f'    if bits{idx} != ({expv} as i64):')
                out.append('        fails = fails + 1 as i64')
                out.append(f'        report({idx} as i64, bits{idx}, ({expv} as i64))')
            else:
                out.append(f'    let g{idx}: {rty} = {src}   // {label}')
                out.append(f'    if (g{idx} as i64) != ({expv} as i64):')
                out.append('        fails = fails + 1 as i64')
                out.append(f'        report({idx} as i64, (g{idx} as i64), ({expv} as i64))')
            idx += 1
        out.append('    return fails')
        out.append('')

    out.append('pub fn main() -> i32:')
    out.append('    mut total := 0 as i64')
    for c in range(nchunks):
        out.append(f'    total = total + chunk_{c}()')
    out.append("    putchar('P' as i32)")
    out.append("    putchar('=' as i32)")
    out.append(f'    emit_i64(({len(cs)} as i64) - total)')
    out.append("    putchar(' ' as i32)")
    out.append("    putchar('F' as i32)")
    out.append("    putchar('=' as i32)")
    out.append('    emit_i64(total)')
    out.append("    putchar(10 as i32)")
    out.append('    fflush(null as ptr[void])')
    out.append('    mut code := total')
    out.append('    if code > 255 as i64:')
    out.append('        code = 255 as i64')
    out.append('    return code as i32')
    sys.stdout.write("\n".join(out) + "\n")

if __name__ == "__main__":
    main()
