# RFC 0040 — `--verbose`: one gate for compiler output, silent by default

- Status: proposed
- Author: AXIOM compiler team
- Created: 2026-09-05
- Affects: driver CLI surface (`main_air.ax`), `print_helpers.ax`; no IR opcode, ABI, linker or backend change
- Decides: D1 decision 4 (user, 2026-08-07)
- Related: RFC 0037 (method-name resolution rank), RFC 0038 (`print` contract) — same "matching by spelling" defect class

---

## 1. Motivation

**The compiler is not silent.** Compiling a five-line hello-world prints, unprompted:

```
[Debug] read_file_content len=140
[Debug] Lexing...
[Debug] Lexing finished, tokens len=19
[Debug] Parsing...
[Debug] Parsing finished, nodes count=8
[Debug] Resolving...
[Debug] Resolving finished
[Debug] Typechecking...
[Debug] Typechecking finished
fn @3() -> t3:
  block_0:  ; entry exit
    %1: t12 = iconst %51
    ...
```

A production compiler prints diagnostics and nothing else. `cc hello.c` is silent on success.
This output is not merely untidy — it is **the compiler's normal-path behaviour on every build**,
so it is part of the product surface, and CLAUDE.md §8 makes diagnostics a product feature.

### 1.1 The real defect is the gate, not the volume

There is already a silencer, `print_helpers.ax:114 is_verbose_debug(s: str) -> bool`, consulted from
three places (`:191`, `:237`, `:392`). It decides whether a line is debug output **by looking at the
spelling of the string being printed**:

```axiom
if s_ptr[start_idx] == '[' as u8:
    let ch = s_ptr[start_idx + 1]
    if ch == 'M' or ch == 'T' or ch == 'C':      // [MONO-DEBUG], [TC-DEBUG], [CGEN-DEBUG]
        return true                               // silence
    ...
    if ch == 'D':                                 // [Debug ...] / [DEBUG ...]
        if check_prefix(p, n, "[Debug] Stage"):      return false   // but NOT this one
        if check_prefix(p, n, "[Debug] Finished"):   return false
        if check_prefix(p, n, "[Debug] Creating"):   return false
        ...                                       // 19 such exceptions in total
        return true
```

So "is this line debug output?" is answered by a **19-entry string whitelist**. The consequences are
not hypothetical:

- **The 19 whitelisted lines print unconditionally** — there is no flag, anywhere, that turns them
  off. That is the output quoted above.
- **The rule is invisible at the call site.** A `[Debug]` line is silenced or not depending on its
  *wording*, decided in a different file. `parser.ax:171` already carries a comment recording that a
  line leaked because `is_verbose_debug` skipped leading spaces and then saw `t`, not `[`.
  `main_air.ax:635` carries another explaining that the prefix "is what routes a line through
  is_verbose_debug()'s silencer".
- **Rewording a message changes whether it is printed.** Renaming `[Debug] Lexing...` to
  `[Debug] Tokenizing...` silently silences it. Nothing fails; the information just disappears.

This is the **same defect class as RFC 0037 rank 2/3 and P4**: behaviour keyed on *how a thing is
spelled* rather than on *what it is*. The user's D1 decision 3 already ruled that class out for
symbol resolution — "key on identity, not on spelling". This RFC applies the same ruling to output.

### 1.2 Why not an environment variable

Rejected by D1 decision 4, and the reason is §19: `AXIOM_DEBUG=1` is implicit state. It is
undiscoverable (`--help` cannot list it), it leaks between shells, it is invisible in a build log,
and it makes a build non-reproducible in the one way that matters most — *the same command produces
different output depending on ambient state*.

---

## 2. Design

### 2.1 The CLI surface

Add one flag to the driver's option loop (`main_air.ax`, the `elif opt == ...` chain at `:1105-1130`,
alongside `--no-stdlib`, `--time`, `-O0`):

```
--verbose, -v      print compiler progress and internal stage output to stderr
```

- **Default: silent.** With no flag the compiler prints **diagnostics only** (errors, warnings, and
  anything a user explicitly asked for, e.g. `--time`, `--emit-*`).
- `--verbose` is **additive and idempotent**; repeating it is not an error and does not stack.
- `-v` is the short form. It is *not* `--version`; the driver has no `--version` today, and this RFC
  reserves `-V`/`--version` for that so the collision cannot be introduced later by accident.

### 2.2 The gate

Replace the spelling whitelist with **one explicit level**, owned by the driver and consulted by the
print helpers:

```axiom
// print_helpers.ax
pub const OUT_QUIET:   u8 = 0    // diagnostics only (default)
pub const OUT_VERBOSE: u8 = 1    // + stage progress, internal dumps

pub fn set_output_level(level: u8)   // called ONCE by the driver after parsing argv
pub fn output_level() -> u8
```

`is_verbose_debug` is **deleted**, along with all 19 string comparisons and `check_prefix`'s use for
this purpose. Its three call sites (`:191`, `:237`, `:392`) become a level test.

**A line's category is decided by which function prints it, not by what it says.** Concretely:

| helper | prints when | used for |
|---|---|---|
| `diag_*` (existing diagnostic path) | always | `error[E….]`, warnings — never gated |
| `verbose_printf` / `verbose_puts` (new names for the gated helpers) | `output_level() >= OUT_VERBOSE` | `[Debug] Lexing...`, `[MONO-DEBUG]`, IR dumps |

Migration is mechanical: a call site that today prints a `[Debug]`/`[MONO-DEBUG]`/`[TC-DEBUG]`/
`[CGEN-DEBUG]`/`[field_offset]` line calls the verbose helper; everything else calls the plain one.
The bracket prefixes may then be kept purely as *human-readable* tags — **nothing reads them any
more**, which is the entire point.

### 2.3 Where verbose output goes

**stderr.** Progress is not the program's product; the product is the object file or the executable.
Sending it to stdout makes `axc … --emit-c -o -` unusable and makes every stdout-comparing test
(CLAUDE.md §7.1) depend on the verbosity level. Diagnostics already go to stderr, so this is also the
smaller change. ⚠️ This is a **behaviour change for the 19 currently-unconditional lines**, which go
to stdout today — see §5.

---

## 3. Alternatives considered

1. **Keep the whitelist, add the missing entries.** Rejected: it grows the defect rather than
   removing it, and every future message must remember to register itself in a table in another file.
   The failure mode is silent in both directions (a leaked line, or a lost line).
2. **`AXIOM_DEBUG` env var.** Rejected by D1 decision 4 and §19 — implicit state, undiscoverable.
3. **Per-subsystem flags** (`--verbose=mono,typecheck`). Deferred, not rejected: the level is
   deliberately an ordered `u8` so `OUT_VERBOSE=1` can later grow `OUT_TRACE=2`, and a future RFC can
   add a mask without changing any call site written under this one. Doing it now would be designing
   for a need nobody has measured (§19: avoid overly generic designs early).
4. **Delete the debug output entirely.** Rejected: it is the primary tool for diagnosing the compiler,
   and this project's own bug history (the B3/B6 SEGV was localized by reading exactly these
   `[Debug] Typechecking...` lines to find where the crash landed) is the argument for keeping it
   behind a flag rather than removing it.

---

## 4. Drawbacks

- **Every gated call site must be touched once.** This is a wide, shallow change — the risk is not
  difficulty but *volume*, and a missed site fails **silently** (a line keeps printing, or stops).
  Mitigated by §6: a test asserting byte-exact empty stdout on a successful compile catches every
  missed site at once, which the whitelist could never do.
- **`--time` and `--emit-*` output must be explicitly classified** as user-requested (always printed),
  not as progress. They are asked for by a flag, so they are product output.
- The self-image prints these lines while compiling itself; making them conditional changes nothing
  about the bytes emitted, but §5 must confirm that.

---

## 5. Compatibility and migration

**Breaking, in a visible and intended way:** a successful compile that printed ~9 lines to stdout now
prints nothing. Anything parsing that output breaks. Audit before landing:

- `scripts/regression_repros.sh` — rows with `cmp=out` compare **program** stdout, not compiler
  stdout, so they are unaffected; but the diagnostic-text rows (`diagloc-*`, `input-halt`, `emit-c`)
  capture **compiler** output and must be re-checked.
- `scripts/fast_fixpoint.ps1` and the build scripts — they redirect compiler output; confirm none
  greps it for a `[Debug]` line.
- `--emit-c` (`emit-c` row) writes the C text; confirm it goes to the file/stdout it is asked for and
  is not interleaved with progress.

**Determinism (§3) is strengthened, not weakened:** output becomes a function of the command line
alone. Today a message's visibility depends on its wording, which is a hidden input.

---

## 6. Testing strategy

Gate class: **frontend/driver only ⇒ A==B**, plus full regression at default and `-O0`.

Oracles in the §7.1 stdout form:

1. **`t_verbose_silent`** — compile a trivial program with no flag, capture the **compiler's** stdout,
   and assert it is empty. This is the test the whitelist made impossible, and it is the one that
   catches every missed migration site at once.
   Expected line: `Trình biên dịch im lặng khi biên dịch thành công: 0` (0 = bytes on stdout).
2. **`t_verbose_flag`** — same program with `--verbose`; assert progress **is** produced (non-zero
   byte count on stderr), so the flag is proven to do something and cannot be "fixed" by deleting the
   output.
   Expected line: `Cờ --verbose bật lại tiến trình: 42`
3. **A diagnostic is NEVER gated** — compile a program with a type error, no `--verbose`, and assert
   the `error[E….]` line and its `--> file:line` still appear. This pins the rule that silence
   applies to progress only. Reuses the existing `diagloc-*` fixtures.

**Calibration is mandatory** (measure on today's driver before the change, so the tests are known to
be capable of failing): oracle 1 must FAIL today (stdout is ~9 lines, not empty) and oracle 2 must
FAIL today (`--verbose` is not a recognized flag). Record both measurements in the commit.

---

## 7. Measured findings, and one remaining question

### 7.1 ⚠️ Unknown options are SILENTLY IGNORED — measured, not assumed

The driver's option chain (`main_air.ax:1058-1169`) is a bare `if/elif` ladder that ends at
`-ctgc-free-report` with **no terminal `else`**. An argument matching no branch falls through and the
loop simply advances. Consequences for this RFC:

- **Oracle 2's calibration as first drafted would have been meaningless.** `axc foo.ax --verbose`
  *already* "succeeds" today — silently doing nothing. A test asserting "`--verbose` is accepted"
  would pass before the feature exists. Oracle 2 must therefore assert the **effect** (progress
  output appears on stderr), never the acceptance of the flag. §6 above is written that way; this
  paragraph records *why*, so a later session cannot simplify it back into a flag-acceptance test.

- **This is a defect in its own right, and it is the accept-then-miscompile class (BUG#53) applied
  to the CLI.** `axc foo.ax --no-stdib` (typo) compiles **with the standard library**, reports
  success, and never mentions that the flag it was given meant nothing. Same for a mistyped `-O2`,
  `--target`, or `--shared`. The user gets a build that silently is not the build they asked for.

  **Not fixed by this RFC** — it is a separate change with its own breakage surface (every script
  passing a flag this compiler does not know would begin to fail, which is exactly the point, but it
  must be audited separately). Filed as a follow-up. It is noted here because implementing
  `--verbose` on top of a ladder that ignores typos would ship a flag whose most common failure mode
  — misspelling it — is invisible.

### 7.2 Still open

- `--time` output: stdout or stderr? It is user-requested, so it is product output, but it is also
  progress-shaped. Proposed: stdout, unchanged, and explicitly out of scope for the gate.
