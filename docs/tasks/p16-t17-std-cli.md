# p16-t17: std.cli — Command-Line Argument Parsing

## Purpose
Implement a declarative command-line argument parser for AXIOM programs, supporting flags, positional arguments, subcommands, and auto-generated help text.

## Context
`std.cli` enables AXIOM programs to parse `--flag value` and `command subcommand` style CLI arguments with minimal boilerplate. The parser auto-generates `--help` output. The `axc` tool itself uses `std.cli` to parse compiler flags.

## Inputs
- `std.process.args()` from p16-t07
- `std.string` from p16-t02
- `std.result` from p16-t12

## Outputs
- `stdlib/cli/cli.ax` — Command, Arg, Flag, Parser

## Dependencies
- p16-t07: std-process — args()
- p16-t12: std-result — Result[T, CliError]

## Detailed Requirements

```axiom
# stdlib/cli/cli.ax

type ArgKind: Flag, Option, Positional, Subcommand

type Arg:
    var name:     str
    var short:    Option[str]    # "-v" or None
    var long:     Option[str]    # "--verbose" or None
    var help:     str
    var required: bool
    var default:  Option[str]
    var kind:     ArgKind

type Command:
    var name:        str
    var version:     str
    var description: str
    var args:        Array[Arg]
    var subcommands: Array[Command]

    fn new(name: str) -> Command
    fn version(mut self, v: str) -> Command   # builder
    fn about(mut self, desc: str) -> Command
    fn arg(mut self, a: Arg) -> Command
    fn subcommand(mut self, cmd: Command) -> Command
    fn parse(self, args: []str) -> Result[ParsedArgs, CliError]
    fn print_help(self)

type ParsedArgs:
    var matched_subcommand: Option[str]
    var values: HashMap[str, str]
    var flags:  HashSet[str]

    fn get(self, name: str) -> Option[str]
    fn get_required(self, name: str) -> Result[str, CliError]
    fn is_set(self, flag: str) -> bool
    fn subcommand(self) -> Option[(str, ParsedArgs)]

type CliError:
    UnknownArg(str)
    MissingRequired(str)
    InvalidValue(arg: str, val: str, expected: str)
    UsageError(str)
```

Usage example:
```axiom
let cli = Command.new("myapp")
    .version("1.0.0")
    .about("My AXIOM application")
    .arg(Arg{name: "input", kind: Positional, required: true, help: "input file"})
    .arg(Arg{name: "verbose", short: "-v", long: "--verbose", kind: Flag, help: "verbose output"})
    .arg(Arg{name: "output", short: "-o", long: "--output", kind: Option, default: Some("-"), help: "output file"})

let parsed = cli.parse(std.process.args_skip_program())?
```

Auto-generated `--help` output:
```
myapp 1.0.0
My AXIOM application

USAGE:
    myapp [OPTIONS] <input>

ARGS:
    <input>    input file

OPTIONS:
    -v, --verbose    verbose output
    -o, --output     output file [default: -]
    -h, --help       print help
```

## Implementation Steps

1. Create `stdlib/cli/cli.ax`.
2. Implement `Command` builder pattern.
3. Implement parser: scan args[], match flags and positionals.
4. Implement `--help` auto-generation.
5. Implement subcommand routing.
6. Wire into `axc` CLI as first consumer.
7. Write tests with various arg combinations.

## Test Plan
- `TestFlagParsing`: `--verbose` → is_set("verbose") = true
- `TestOptionParsing`: `-o output.txt` → get("output") = Some("output.txt")
- `TestPositional`: `myapp file.ax` → get("input") = Some("file.ax")
- `TestMissingRequired`: missing required arg → CliError::MissingRequired
- `TestHelpOutput`: --help → help text printed, exit 0
- `TestSubcommand`: `myapp build --release` → subcommand = ("build", {release: true})

## Validation Checklist
- [ ] Unknown flags produce helpful error, not panic
- [ ] --help always works even with missing required args
- [ ] Short and long flag forms both work
- [ ] Subcommand parsing chains correctly

## Acceptance Criteria
- `axc compile --help` produces readable help output

## Definition of Done
- [ ] `stdlib/cli/cli.ax` implemented
- [ ] `axc` CLI uses std.cli for argument parsing
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Combined flags `-vrf` parsing | Split into individual flag chars in parser |

## Future Follow-up Tasks
- Shell completion generation (bash/zsh)
- Config file integration (merge args with config file values)
