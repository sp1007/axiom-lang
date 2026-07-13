# RFC 0024 — Block (triple-quoted) string literals `"""..."""`

Status: Accepted — implemented (lexer + string-value extraction; frontend, A==B)
Author: autopilot
Related: RFC 0020 (string type system), the M4 compliance suite (`tests/axiom_compliance_suite.ax`)
which uses `"""..."""`.

## 1. Motivation

AXIOM has only single-line `"..."` strings — a normal string terminates at a newline
(lexer.ax:299, "unterminated"). Multi-line text (embedded docs, SQL, templates, test
fixtures) requires awkward `"\n"` concatenation. The M4 compliance suite already uses
`"""..."""`. This RFC adds a multi-line **block string** literal.

## 2. Design

A **block string** opens with three double-quotes `"""` and closes at the next
**unescaped** `"""`. Between the delimiters:

* **Newlines are allowed** (this is the point) and are part of the content verbatim.
* **Interior single `"` and `""` are literal** — only a run of three closes the literal.
* **Escape sequences are processed** exactly as in normal strings (`\n \t \r \" \\`),
  so an escaped quote `\"` does not prematurely close the block.
* Content is **verbatim** between the delimiters — **no** indentation dedenting and **no**
  leading/trailing-newline stripping (the simplest, most predictable semantics; a dedent
  policy can be a later addition without breaking this).

Examples:
```
let msg = """line one
line two"""            // len 17 (the newline is one byte)
let q   = """say "hi" ok"""   // interior double-quotes are literal, len 11
let e   = """a\tb"""   // escapes processed: 'a', TAB, 'b' -> len 3
let z   = """"""       // empty block string, len 0
```

## 3. Implementation (frontend only)

Three sites, all mechanical; the compiler itself uses no block strings so self-codegen is
unchanged (fast fixpoint **A==B**, hash 21ED17B1…):

1. **lexer.ax `scan_string`** — on `"`, if the next two bytes are also `"`, enter block
   mode: skip the opening `"""`, then consume bytes (including `\n`) until an unescaped
   `"""`, skipping `\`-escapes so `\"` can't close early. The token text is the whole
   `"""..."""` (delimiters included, like normal strings).
2. **print_helpers.ax `unescape_string_literal`** (the native byte/length source, called by
   every x86 emitter) — a token that begins and ends with three `"` strips 3 delimiter
   chars each side (else the existing 1). Unambiguous: only a block token starts with three
   literal quotes (a normal string escapes any interior quote).
3. **typecheck.ax `strip_quotes`** — same 3-vs-1 strip for the comptime string value.

The runtime `{ptr,len}` bytes and length are derived from `unescape_string_literal`'s
output (x86_coff.ax:431-432 etc.), so fixing (2) fixes both value and length on the native
path. `lower_string_lit` interns the raw token unchanged.

## 4. Alternatives / drawbacks

* **Dedent (strip common leading indentation)** — ergonomic for indented multi-line blocks
  but requires an indentation policy and interacts with AXIOM's significant-indentation
  lexer. Deferred; verbatim is a strict subset a dedent mode can extend.
* **Raw (no escape processing)** — rejected for MVP; matching normal-string escapes is more
  consistent and lets `\"` appear in a block.
* Interior line numbers for diagnostics are not advanced across a block string's newlines
  (minor; the whole block is one token). Acceptable.

## 5. Compatibility

Purely additive: `"""` was previously two empty strings `""` `""` adjacent (a parse
oddity) or a lex of `""` then a new string — no valid program relied on it. Normal strings
are untouched (verified: `"hi".len == 2`). No stdlib changes.

## 6. Tests

Oracles: t_blockstr (len 5), t_blockstrml (multi-line len 5), t_blockstresc (escape, 42),
t_blockstrbyte (first byte 'A'=65), t_blockstrquote (embedded quotes len 11). All O0==O1.
Regression green.
