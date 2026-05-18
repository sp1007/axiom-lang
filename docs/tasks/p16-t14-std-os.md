# p16-t14: std.os — OS Abstraction Layer

## Purpose
Implement the OS abstraction layer providing platform-agnostic access to filesystem metadata, paths, signals, and OS-specific constants, normalizing differences between Linux, macOS, and Windows.

## Context
`std.os` provides path manipulation (forward/backward slashes), filesystem metadata (stat), signal handling, and OS detection. It's the foundation for portable AXIOM programs that run on multiple platforms without `#[cfg]` clutter in user code.

## Inputs
- OS APIs: stat/lstat, mkdir, rename, remove (POSIX); GetFileAttributes, CreateDirectory (Windows)
- Platform detection at compile time via Target from p11-t01

## Outputs
- `stdlib/os/path.ax` — Path, PathBuf (platform-aware)
- `stdlib/os/fs.ax` — filesystem operations (stat, mkdir, rename, remove)
- `stdlib/os/signal.ax` — signal handling
- `stdlib/os/platform.ax` — compile-time platform constants

## Dependencies
- p11-t01: target-triple — Target for compile-time platform detection
- p16-t04: std-io — IOError type reuse
- p16-t07: std-process — signal handling integration

## Detailed Requirements

```axiom
# stdlib/os/path.ax
type PathBuf:
    var inner: str

    fn new(s: str) -> PathBuf
    fn from(s: str) -> PathBuf
    fn push(mut self, component: str)
    fn parent(self) -> Option[PathBuf]
    fn file_name(self) -> Option[str]
    fn extension(self) -> Option[str]
    fn with_extension(self, ext: str) -> PathBuf
    fn join(self, other: str) -> PathBuf
    fn to_str(self) -> str
    fn exists(self) -> bool
    fn is_file(self) -> bool
    fn is_dir(self) -> bool

const SEPARATOR: str = "/" or "\\" (platform)

# stdlib/os/fs.ax
type FileMetadata:
    var size:     u64
    var modified: SystemTime
    var created:  SystemTime
    var is_file:  bool
    var is_dir:   bool
    var is_symlink: bool
    var mode:     u32   # Unix permission bits

fn metadata(path: str) -> Result[FileMetadata, IOError]
fn create_dir(path: str) -> Result[void, IOError]
fn create_dir_all(path: str) -> Result[void, IOError]
fn remove_file(path: str) -> Result[void, IOError]
fn remove_dir(path: str) -> Result[void, IOError]
fn rename(from: str, to: str) -> Result[void, IOError]
fn copy(from: str, to: str) -> Result[u64, IOError]
fn read_dir(path: str) -> Result[DirIter, IOError]
fn temp_dir() -> PathBuf

# stdlib/os/signal.ax
type Signal: SIGINT, SIGTERM, SIGHUP, SIGUSR1, SIGUSR2

fn trap_signal(sig: Signal, handler: fn(Signal))

# stdlib/os/platform.ax
const IS_LINUX:   bool = (compile-time)
const IS_MACOS:   bool = (compile-time)
const IS_WINDOWS: bool = (compile-time)
const OS_NAME:    str  = "linux" | "macos" | "windows"
```

Path separator: normalized on all platforms to `/` internally; converted to `\` on Windows for system calls.

## Implementation Steps

1. Create `stdlib/os/path.ax` — PathBuf with platform separator handling.
2. Create `stdlib/os/fs.ax` — wrapping POSIX stat/mkdir/rename/unlink.
3. Implement `read_dir()` — DirIter wrapping opendir/readdir.
4. Create `stdlib/os/signal.ax` — sigaction wrapper.
5. Create `stdlib/os/platform.ax` — compile-time constants.
6. Write tests on each platform.

## Test Plan
- `TestPathJoin`: "a/b".join("c") = "a/b/c"
- `TestPathParent`: "/a/b/c".parent() = Some("/a/b")
- `TestFsMetadata`: stat existing file → correct size
- `TestFsCreateDir`: create + stat new dir → is_dir = true
- `TestFsRemove`: create + remove file → exists() = false
- `TestSignalTrap`: SIGINT handler called on Ctrl-C

## Validation Checklist
- [ ] Path separator correct per platform
- [ ] create_dir_all creates missing parents
- [ ] DirIter skips "." and ".."
- [ ] Signal handler registered as async-signal-safe

## Acceptance Criteria
- `axc` tool uses std.os.path for all path manipulation

## Definition of Done
- [ ] All os modules implemented
- [ ] Tests pass on Linux (Windows/macOS as follow-up)

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Windows path separator normalization | Normalize to "/" in PathBuf; convert to "\\" only at OS boundary |

## Future Follow-up Tasks
- Filesystem watching (inotify/FSEvents/ReadDirectoryChanges)
- Symlink creation and resolution
