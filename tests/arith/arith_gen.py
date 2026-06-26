#!/usr/bin/env python3
# Differential arithmetic test generator for AXIOM (RFC 0006).
#
# Emits a self-contained, self-checking AXIOM program: for each generated
# expression it computes `got` in a chosen result type, then compares
# `(got as i64)` against an expected i64 constant computed here by a trusted
# fixed-width oracle. Comparing in the i64 domain avoids needing u64 literals
# or unsigned printing in the generated program.
#
# Oracle semantics (the intended AXIOM numeric spec):
#   - integers are fixed-width two's complement, wrapping on overflow
#   - signed div/mod truncate toward zero; unsigned use the unsigned rep
#   - `as T` = truncating two's-complement cast (sign/zero-extend on widen)
#   - shifts: logical for unsigned, arithmetic for signed; amount in [0,W-1]
#
# Usage: python3 arith_gen.py <N> <seed> > arith_test.ax
import sys, random, struct

# (name, width_bits, signed)
ITYPES = [
    ("i8", 8, True), ("i16", 16, True), ("i32", 32, True), ("i64", 64, True),
    ("u8", 8, False), ("u16", 16, False), ("u32", 32, False), ("u64", 64, False),
]
TINFO = {n: (w, s) for (n, w, s) in ITYPES}

def mask(w):  return (1 << w) - 1

def wrap(val, ty):
    """Reduce a Python int to type `ty`'s mathematical value (two's complement)."""
    w, s = TINFO[ty]
    m = val & mask(w)
    if s and m >= (1 << (w - 1)):
        m -= (1 << w)
    return m

def as_cast(val, ty):
    """`val as ty` — truncating two's-complement cast."""
    return wrap(val, ty)

def to_i64(val):
    return wrap(val, "i64")

def trunc_div(a, b):
    # truncate toward zero (C/Rust/Go semantics), unlike Python //
    q = abs(a) // abs(b)
    return -q if (a < 0) != (b < 0) else q

def trunc_mod(a, b):
    return a - trunc_div(a, b) * b

def rand_val(ty, rng):
    w, s = TINFO[ty]
    edge = [0, 1, 2, mask(w), mask(w) - 1]
    if s:
        edge += [(1 << (w - 1)) - 1, -(1 << (w - 1)), -1, -2]
    if rng.random() < 0.35:
        v = rng.choice(edge)
    else:
        v = rng.randint(-(1 << (w - 1)) if s else 0, mask(w) if not s else (1 << (w - 1)) - 1)
    return wrap(v, ty)

def lit(val):
    # AXIOM integer literal; negatives via unary minus
    return f"({val})" if val >= 0 else f"(0 - {abs(val)})"

ARITH = ["+", "-", "*", "/", "%", "&", "|", "^", "<<", ">>"]

# ---- float oracle ----
def f32_round(x):
    return struct.unpack('<f', struct.pack('<f', x))[0]

def fbits(val, ty):
    """IEEE bit pattern of `val` in float type `ty`, as a SIGNED int matching
    AXIOM's `(&g as ptr[iN])[0]` reinterpret (i64 for f64, i32 for f32→i64)."""
    if ty == "f64":
        return struct.unpack('<q', struct.pack('<d', val))[0]
    return struct.unpack('<i', struct.pack('<f', f32_round(val)))[0]  # sign-ext to i64 OK

def frand(rng):
    # exactly-representable f32/f64 values (int + {0,.25,.5,.75}); finite, no NaN/inf
    base = rng.randint(-1000, 1000)
    frac = rng.choice([0.0, 0.25, 0.5, 0.75, -0.25, -0.5])
    edge = [0.0, 1.0, -1.0, 2.0, 0.5, -0.5, 1024.0, -1024.0]
    return rng.choice(edge) if rng.random() < 0.25 else float(base) + frac

def flit(v):
    s = repr(v)
    if "." not in s and "e" not in s and "E" not in s:
        s += ".0"
    return f"({s})" if v >= 0 else f"(0.0 - {repr(abs(v))})"

FOPS = ["+", "-", "*", "/"]
FTYPES = ["f32", "f64"]

def gen_fexpr(rng):
    """Return (axiom_src, expected_bits_i64, result_type) for a float expr."""
    rty = rng.choice(FTYPES)
    op = rng.choice(FOPS)
    av = frand(rng); bv = frand(rng)
    if op == "/":
        while bv == 0.0:
            bv = frand(rng)
    # operands cast to result type
    rnd = f32_round if rty == "f32" else (lambda z: z)
    a = rnd(av); b = rnd(bv)
    if op == "+": res = a + b
    elif op == "-": res = a - b
    elif op == "*": res = a * b
    else: res = a / b
    res = rnd(res)  # round result to f32 when applicable
    a_src = f"({flit(av)} as {rty})"
    b_src = f"({flit(bv)} as {rty})"
    src = f"({a_src} {op} {b_src})"
    return src, fbits(res, rty), rty

def gen_mixed_expr(rng):
    """Mixed int/float operands, each cast to a float result type. Tests int->float
    conversion + float op. (Run AFTER float arithmetic is fixed — RFC 0006.)"""
    rty = rng.choice(FTYPES)
    op = rng.choice(FOPS)
    rnd = f32_round if rty == "f32" else (lambda z: z)

    def operand():
        if rng.random() < 0.5:
            ity = rng.choice([t[0] for t in ITYPES])
            iv = rand_val(ity, rng)
            src = f"(({lit(iv)} as {ity}) as {rty})"
            return src, rnd(float(iv))
        fv = frand(rng)
        return f"({flit(fv)} as {rty})", rnd(fv)

    a_src, a = operand()
    b_src, b = operand()
    if op == "/":
        while b == 0.0:
            b_src, b = operand()
    if op == "+": res = a + b
    elif op == "-": res = a - b
    elif op == "*": res = a * b
    else: res = a / b
    res = rnd(res)
    return f"({a_src} {op} {b_src})", fbits(res, rty), rty

def gen_imix_expr(rng):
    """IMPLICIT mixed: an int operand (NOT cast to float) combined with an f64
    operand — tests implicit int->float promotion in a binary op (e.g. a/3.0).
    Result is f64. (Run AFTER float arithmetic + promotion is fixed — RFC 0006.)"""
    op = rng.choice(FOPS)
    # int operand (no float cast), value promoted to float by the oracle
    ity = rng.choice([t[0] for t in ITYPES])
    iv = rand_val(ity, rng)
    int_src = f"({lit(iv)} as {ity})"
    fv = frand(rng)
    flt_src = f"({flit(fv)} as f64)"
    # random operand order (int OP float) or (float OP int)
    if rng.random() < 0.5:
        a, b = float(iv), fv
        a_src, b_src = int_src, flt_src
    else:
        a, b = fv, float(iv)
        a_src, b_src = flt_src, int_src
    if op == "/":
        if b == 0.0:
            b = 1.0; b_src = "(1.0 as f64)"
    if op == "+": res = a + b
    elif op == "-": res = a - b
    elif op == "*": res = a * b
    else: res = a / b
    return f"({a_src} {op} {b_src})", fbits(res, "f64"), "f64"

def gen_expr(rng):
    """Return (axiom_src, expected_i64, result_type)."""
    rty = rng.choice([t[0] for t in ITYPES])
    w, s = TINFO[rty]
    op = rng.choice(ARITH)

    # operands are values of (possibly different) types, cast to rty
    aty = rng.choice([t[0] for t in ITYPES])
    bty = rng.choice([t[0] for t in ITYPES])
    av = rand_val(aty, rng)
    bv = rand_val(bty, rng)

    # operands cast into result type
    a = as_cast(av, rty)
    b = as_cast(bv, rty)

    a_src = f"({lit(av)} as {rty})"
    b_src = f"({lit(bv)} as {rty})"

    if op in ("/", "%"):
        # avoid div-by-zero and signed INT_MIN/-1 overflow: use a safe positive divisor
        bv2 = rng.randint(1, min(mask(w), 1000))
        b = as_cast(bv2, rty)
        if b == 0:
            b = 1
        b_src = f"({lit(bv2)} as {rty})"
        if op == "/":
            res = trunc_div(a, b) if s else (a & mask(w)) // (b & mask(w))
        else:
            res = trunc_mod(a, b) if s else (a & mask(w)) % (b & mask(w))
        res = wrap(res, rty)
    elif op in ("<<", ">>"):
        sh = rng.randint(0, w - 1)
        b_src = f"({sh} as {rty})"
        if op == "<<":
            res = wrap(a << sh, rty)
        else:
            if s:
                res = wrap(a >> sh, rty)               # arithmetic (a is signed value)
            else:
                res = wrap((a & mask(w)) >> sh, rty)   # logical
    elif op == "&":
        res = wrap((a & mask(w)) & (b & mask(w)), rty)
    elif op == "|":
        res = wrap((a & mask(w)) | (b & mask(w)), rty)
    elif op == "^":
        res = wrap((a & mask(w)) ^ (b & mask(w)), rty)
    elif op == "+":
        res = wrap(a + b, rty)
    elif op == "-":
        res = wrap(a - b, rty)
    elif op == "*":
        res = wrap(a * b, rty)

    src = f"({a_src} {op} {b_src})"
    return src, to_i64(res), rty

CHUNK = 100  # expressions per function (keep functions small for regalloc)

def main():
    n = int(sys.argv[1]) if len(sys.argv) > 1 else 1000
    seed = int(sys.argv[2]) if len(sys.argv) > 2 else 12345
    mode = sys.argv[3] if len(sys.argv) > 3 else "int"   # int | float
    rng = random.Random(seed)

    out = []
    out.append('// AUTO-GENERATED by arith_gen.py — differential arithmetic test (RFC 0006)')
    out.append(f'// N={n} seed={seed}. exit code = min(fails,255); "FAIL" lines list mismatches.')
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

    nchunks = (n + CHUNK - 1) // CHUNK
    idx = 0
    for c in range(nchunks):
        out.append(f'fn chunk_{c}() -> i64:')
        out.append('    mut fails := 0 as i64')
        for _ in range(CHUNK):
            if idx >= n:
                break
            if mode in ("float", "mixed", "imix"):
                if mode == "float":
                    src, exp, rty = gen_fexpr(rng)
                elif mode == "mixed":
                    src, exp, rty = gen_mixed_expr(rng)
                else:
                    src, exp, rty = gen_imix_expr(rng)
                bptr = "ptr[i64]" if rty == "f64" else "ptr[i32]"
                out.append(f'    mut g{idx}: {rty} = {src}')
                out.append(f'    let bits{idx}: i64 = ((&g{idx} as {bptr})[0]) as i64')
                out.append(f'    if bits{idx} != ({exp} as i64):')
                out.append('        fails = fails + 1 as i64')
                out.append(f'        report({idx} as i64, bits{idx}, ({exp} as i64))')
            else:
                src, exp, rty = gen_expr(rng)
                out.append(f'    let g{idx}: {rty} = {src}')
                out.append(f'    if (g{idx} as i64) != ({exp} as i64):')
                out.append('        fails = fails + 1 as i64')
                out.append(f'        report({idx} as i64, (g{idx} as i64), ({exp} as i64))')
            idx += 1
        out.append('    return fails')
        out.append('')

    out.append('pub fn main() -> i32:')
    out.append('    mut total := 0 as i64')
    for c in range(nchunks):
        out.append(f'    total = total + chunk_{c}()')
    out.append("    putchar('P' as i32)")
    out.append("    putchar('=' as i32)")
    out.append(f'    emit_i64(({n} as i64) - total)')
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
