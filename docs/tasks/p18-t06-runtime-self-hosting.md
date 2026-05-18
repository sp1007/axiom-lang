# p18-t06: Runtime Self-Hosting (AxAlloc + Scheduler in AXIOM)

## Purpose
Port the runtime components (AxAlloc, panic handler, and scheduler stub) from C to AXIOM. This eliminates the last C dependency in the AXIOM toolchain, completing the full self-hosting goal where the compiler, runtime, and standard library are all written in AXIOM.

## Context
Plan §7 Stage 3 specifies: _"Runtime (AxAlloc, M:N scheduler) written in AXIOM. Zero dependency on libc for core runtime."_ This is the final step before declaring AXIOM fully self-hosted. The runtime must be compilable by the self-hosted compiler (p18-t04) and produce a standalone binary that doesn't link against libc for core operations.

## Inputs
- C runtime from p07 (axalloc.c, genref.c, panic.c)
- Production AxAlloc from p14 (size classes, segments)
- Scheduler from p15 (M:N work-stealing)
- Self-hosted compiler from p18-t04
- `extern "C"` FFI for syscall access

## Outputs
- `bootstrap/runtime/axalloc.ax` — allocator in AXIOM
- `bootstrap/runtime/genref.ax` — generational reference check in AXIOM
- `bootstrap/runtime/panic.ax` — panic handler in AXIOM
- `bootstrap/runtime/scheduler_stub.ax` — basic scheduler in AXIOM
- `bootstrap/runtime/syscall.ax` — raw syscall wrappers (Linux: mmap, munmap, write, exit)

## Dependencies
- p18-t04: stage4-full-compiler — self-hosted compiler works
- p18-t05: triple-build-verification — verification infrastructure
- p14-t02: axalloc-segment-manager — segment design to port
- p15-t02: scheduler — scheduler design to port

## Subsystems Affected
- Runtime: entirely rewritten in AXIOM
- Build system: no more C compilation step for runtime
- Linking: runtime is compiled AXIOM code, not C object files

## Detailed Requirements

### 1. Syscall Layer

To eliminate libc dependency, the runtime must make raw syscalls:

```axiom
// bootstrap/runtime/syscall.ax
// Linux x86-64 syscalls
extern fn syscall(num: usize, a1: usize, a2: usize, a3: usize, a4: usize, a5: usize, a6: usize) -> isize

pub fn sys_mmap(addr: usize, len: usize, prot: i32, flags: i32, fd: i32, offset: usize) -> usize:
    return syscall(9, addr, len, prot as usize, flags as usize, fd as usize, offset) as usize

pub fn sys_munmap(addr: usize, len: usize) -> i32:
    return syscall(11, addr, len, 0, 0, 0, 0) as i32

pub fn sys_write(fd: i32, buf: *u8, count: usize) -> isize:
    return syscall(1, fd as usize, buf as usize, count, 0, 0, 0)

pub fn sys_exit(code: i32):
    syscall(60, code as usize, 0, 0, 0, 0, 0)
```

### 2. AxAlloc in AXIOM

Port the C allocator to AXIOM, maintaining the same ABI:

```axiom
pub struct AxHeader:
    gen_id: u64    // 63 bits gen + 1 bit is_free

pub struct AxAlloc:
    segments: Array[*AxSegment, 30]  // one per size class
    
    pub fn alloc(mut self, size: usize) -> *u8
    pub fn free(mut self, ptr: *u8)
```

### 3. Generational Reference Check

```axiom
pub fn ax_deref_check(ptr: *u8, expected_gen: u64):
    let header = (ptr as *AxHeader) - 1
    if header.gen_id != expected_gen:
        ax_panic("Generational reference mismatch: use-after-free detected")
```

### 4. Panic Handler

```axiom
pub fn ax_panic(msg: string):
    sys_write(2, msg.ptr, msg.len)
    sys_write(2, "\n".ptr, 1)
    sys_exit(101)
```

### 5. Phased Approach

1. **Phase A**: Port panic handler (simplest, no memory allocation)
2. **Phase B**: Port genref check (simple comparison)
3. **Phase C**: Port AxAlloc MVP (malloc wrapper using mmap)
4. **Phase D**: Port production AxAlloc (size classes, bump pointer)
5. **Phase E**: Port scheduler stub (thread creation via clone syscall)

## Implementation Steps

1. Create `bootstrap/runtime/` directory.
2. Write `syscall.ax` with raw Linux syscall wrappers.
3. Port `panic.c` → `panic.ax` (simplest component).
4. Port `genref.c` → `genref.ax`.
5. Port `axalloc.c` → `axalloc.ax` (MVP: mmap-based).
6. Test: compile a program with the AXIOM runtime → runs correctly.
7. Run triple-build with AXIOM runtime → hashes match.
8. Port scheduler stub (thread creation).

## Test Plan

- `TestAxAllocAXIOM`: allocate and free 1000 objects → no crash
- `TestGenRefAXIOM`: gen_id mismatch → panic with correct message
- `TestPanicAXIOM`: panic handler prints message and exits 101
- `TestSyscallMmap`: mmap/munmap cycle → memory accessible
- `TestTripleBuildWithAXIOMRuntime`: triple-build verification passes

## Validation Checklist

- [ ] All compliance tests pass with AXIOM runtime
- [ ] No libc dependency in the final binary (verified with `ldd`)
- [ ] Triple-build verification passes
- [ ] Memory safety: AddressSanitizer equivalent checks pass

## Acceptance Criteria

- `axc build hello.ax` using AXIOM runtime produces working binary
- `ldd` shows no libc linkage (Linux)
- Triple-build hash match

## Definition of Done

- [ ] `bootstrap/runtime/` directory with all .ax files
- [ ] All compliance tests pass
- [ ] Triple-build passes with AXIOM runtime
- [ ] No C runtime files needed

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Syscall differences across platforms | Start with Linux x86-64 only; add platform layers later |
| Self-referential allocation (allocator needs memory to run) | Bootstrap with a static buffer; use mmap for initial segments |
| Performance regression from AXIOM vs C runtime | Profile and optimize hot paths; AXIOM compiler should match C perf |

## Future Follow-up Tasks

- Production release v1.0.0: fully self-hosted toolchain
- Future: Windows and macOS syscall layers
- Future: Full M:N scheduler in AXIOM
