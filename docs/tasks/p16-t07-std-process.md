# p16-t07: std.process — Process and OS Interaction

## Purpose
Implement process management, environment access, command execution, and OS interaction for AXIOM programs.

## Context
`std.process` provides the primitives for AXIOM programs to interact with the OS: spawn child processes, read environment variables, access command-line arguments, and exit. These are needed for tools, build scripts, and system programs written in AXIOM.

## Inputs
- OS process APIs: fork/exec/waitpid (POSIX), CreateProcess (Windows)
- Environment variable access: getenv/setenv
- `args()` access: main's argc/argv preserved by runtime

## Outputs
- `stdlib/process/process.ax` — process spawning and management
- `stdlib/process/env.ax` — environment variables
- `stdlib/process/args.ax` — command-line argument access

## Dependencies
- p15-t03: actor-system-init — runtime preserves argc/argv
- p16-t04: std-io — stdin/stdout/stderr piping for child processes

## Detailed Requirements

```axiom
# stdlib/process/process.ax
type Command:
    var program: str
    var args: Array[str]
    var env: HashMap[str, str]
    var cwd: Option[str]
    var stdin:  Pipe
    var stdout: Pipe
    var stderr: Pipe

    fn new(program: str) -> Command
    fn arg(mut self, a: str) -> Command    # builder pattern
    fn args(mut self, a: []str) -> Command
    fn env(mut self, k: str, v: str) -> Command
    fn cwd(mut self, dir: str) -> Command
    fn spawn(self) -> Result[Child, IOError]
    fn output(self) -> Result[Output, IOError]  # wait for completion

type Child:
    var pid: u32
    fn wait(mut self) -> Result[ExitStatus, IOError]
    fn kill(mut self) -> Result[void, IOError]

type Output:
    var stdout: str
    var stderr: str
    var status: ExitStatus

type ExitStatus:
    var code: i32
    fn success(self) -> bool

fn exit(code: i32) -> !   # never returns
fn abort() -> !

# stdlib/process/env.ax
fn env_var(name: str) -> Option[str]
fn set_env_var(name: str, val: str)
fn env_vars() -> HashMap[str, str]

# stdlib/process/args.ax
fn args() -> []str   # command-line arguments (including argv[0])
fn args_skip_program() -> []str  # argv[1..]
```

POSIX implementation of `Command.spawn()`:
```c
// fork + exec with optional pipe setup
pid_t pid = fork();
if (pid == 0) {
    // child: dup2 pipes, execvp(program, args)
}
// parent: record pid in Child
```

Windows: `CreateProcess` with `STARTUPINFO` for pipe setup.

## Implementation Steps

1. Create `stdlib/process/process.ax` and C shims for fork/exec.
2. Implement `Command` builder with pipe configuration.
3. Implement `Child.wait()` — waitpid (POSIX) or WaitForSingleObject (Windows).
4. Create `stdlib/process/env.ax` — getenv wrapper.
5. Create `stdlib/process/args.ax` — read preserved argc/argv from runtime.
6. Write tests spawning `echo` as subprocess.

## Test Plan
- `TestCommandOutput`: `Command.new("echo").arg("hello").output()` → stdout = "hello\n"
- `TestCommandExit`: command with non-zero exit → ExitStatus.code != 0
- `TestEnvVar`: set env var → get env var → same value
- `TestArgs`: args() returns slice starting with program name
- `TestExit`: exit(0) terminates process (tested in subprocess)

## Validation Checklist
- [ ] Child process fd cleanup (no fd leaks after spawn)
- [ ] exit() calls ax_runtime_shutdown before _exit()
- [ ] env_vars() returns complete environment (not just selected keys)

## Acceptance Criteria
- `axc` tool itself uses `Command` to invoke system linker

## Definition of Done
- [ ] `stdlib/process/process.ax` implemented
- [ ] subprocess tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Signal delivery to child on Kill() varies by platform | Use SIGKILL (POSIX); TerminateProcess (Windows) |

## Future Follow-up Tasks
- Async process waiting (futures-based waitpid)
- Process groups and job control
