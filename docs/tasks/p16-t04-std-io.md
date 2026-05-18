# p16-t04: std.io — File and Stream I/O

## Purpose
Implement the AXIOM standard I/O library for file reading/writing, standard streams (stdin/stdout/stderr), and buffered I/O — both synchronous and async variants.

## Context
`std.io` is one of the most commonly used stdlib modules. It wraps OS file descriptors with AXIOM's ownership model (File closes automatically when dropped) and integrates with the async runtime for non-blocking I/O.

## Inputs
- I/O event loop from p15-t08
- OS file system calls (POSIX: open/read/write/close; Win32: CreateFile/ReadFile/WriteFile)
- AXIOM ownership model for automatic file close (destructor via CTGC)

## Outputs
- `stdlib/io/file.ax` — File type with read/write/seek/close
- `stdlib/io/stream.ax` — stdin/stdout/stderr as streams
- `stdlib/io/buffered.ax` — BufReader/BufWriter
- `stdlib/io/async.ax` — async read/write using I/O event loop

## Dependencies
- p15-t08: io-event-loop — async read/write
- p06-t05: ctgc-destroy-injection — auto-close file on scope exit
- p14-t01: axalloc — buffered I/O buffers

## Detailed Requirements

```axiom
# stdlib/io/file.ax
type File:
    var fd: i32  # OS file descriptor

    fn open(path: str, mode: FileMode) -> Result[File, IOError]
    fn read(mut self, buf: []u8) -> Result[u32, IOError]
    fn read_all(mut self) -> Result[str, IOError]
    fn read_line(mut self) -> Result[str, IOError]
    fn write(mut self, buf: []u8) -> Result[u32, IOError]
    fn write_str(mut self, s: str) -> Result[u32, IOError]
    fn seek(mut self, pos: SeekPos) -> Result[u64, IOError]
    fn flush(mut self) -> Result[void, IOError]
    fn close(mut self)  # called by CTGC destructor

type FileMode:
    Read, Write, Append, ReadWrite, Create, CreateNew, Truncate

# stdlib/io/stream.ax
var stdin:  InputStream
var stdout: OutputStream
var stderr: OutputStream

fn print(s: str)         # write to stdout
fn println(s: str)       # write to stdout + newline
fn eprint(s: str)        # write to stderr
fn eprintln(s: str)      # write to stderr + newline

# stdlib/io/buffered.ax
type BufReader:
    var inner: File
    var buf: []u8
    var pos: u32
    var filled: u32

    fn new(f: File) -> BufReader
    fn read_line(mut self) -> Result[str, IOError]
    fn lines(mut self) -> LineIter

type BufWriter:
    var inner: File
    var buf: []u8
    var pos: u32

    fn new(f: File) -> BufWriter
    fn write(mut self, s: str) -> Result[void, IOError]
    fn flush(mut self) -> Result[void, IOError]

# stdlib/io/async.ax
async fn read_file(path: str) -> Result[str, IOError]
async fn write_file(path: str, content: str) -> Result[void, IOError]
```

Error type:
```axiom
type IOError:
    NotFound(path: str)
    PermissionDenied
    AlreadyExists
    Interrupted
    UnexpectedEOF
    Other(msg: str)
```

## Implementation Steps

1. Create `stdlib/io/file.ax` wrapping POSIX open/read/write.
2. Wire CTGC: File has destructor `close()` injected at scope exit.
3. Create `stdlib/io/stream.ax` with stdin/stdout/stderr globals.
4. Create `stdlib/io/buffered.ax` — BufReader with 8KB default buffer.
5. Create `stdlib/io/async.ax` — async read/write using p15-t08 event loop.
6. Write tests using temporary files.

## Test Plan
- `TestFileReadWrite`: write string → read back → equal
- `TestBufReaderLines`: read multi-line file → correct line count
- `TestFileNotFound`: open nonexistent file → Result::Err(NotFound)
- `TestAutoClose`: file auto-closed when leaving scope (verified via fd table)
- `TestAsyncRead`: async read of file → future resolves with content

## Validation Checklist
- [ ] File closed automatically by CTGC destructor
- [ ] IOError variants cover all common OS errors
- [ ] BufReader handles partial reads correctly
- [ ] async variants use I/O event loop (not blocking syscall)

## Acceptance Criteria
- Read 1MB file in < 5ms with BufReader

## Definition of Done
- [ ] All io modules implemented
- [ ] All tests pass including async variant

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Double-close if user calls close() + CTGC also closes | Set fd = -1 after close; skip if fd < 0 |

## Future Follow-up Tasks
- Directory walking (`stdlib/io/fs`)
- Network I/O (`stdlib/net`)
