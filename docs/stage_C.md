# AXIOM Toolchain — Stage C Task Specifications: Direct Syscalls, Autonomous stdlib & Advanced Frontend Features

This document defines the technical task specifications, structural design, inputs, outputs, and success criteria for implementing Stage C of AXIOM's self-hosted compiler pipeline.

Stage C addresses the final barriers to complete runtime independence and advanced language capability in AXIOM:
1. **Direct Binary Syscall Lowering** to bypass dynamic C libraries completely on x86_64 Linux targets.
2. **100% C-Independent Standard Library (`std/io.ax` & `std/os.ax`)** utilizing raw system calls.
3. **Struct, Type Alias, Sum Type, and Match Expression Support** in the self-hosted Stage 1 compiler frontend.

---

## 🎯 Task 1: Direct Binary Syscall Lowering
* **Objective**: Teach AXIOM's instruction selector (`x86_selector.ax`) and machine encoder (`x86_encoding.ax`) to emit direct binary `syscall` instructions for standard low-level kernel APIs (`sys_write`, `sys_exit`, `sys_mmap`, `sys_munmap`) on x86-64 Linux targets, bypassing standard dynamic C libraries completely.

### 📋 Technical Design:
1. **Syscall Machine Opcode**:
   - Define `pub const MACH_SYSCALL: u16 = 23` in `x86_selector.ax` and `x86_encoding.ax`.
2. **Byte Serializer**:
   - Implement `x86_encode_syscall(buf: ptr[ByteVec])` in `x86_encoding.ax` to emit the raw `0F 05` assembly instruction bytes.
3. **Instruction Lowering**:
   - Update `x86_selector.ax`'s `OP_CALL` lowering logic: If compiling for Linux/sysv ABI and the FFI function called matches an OS system call (e.g. `write`, `exit`, `mmap`, `munmap`), lower the call directly to a syscall sequence:
     - Load the system call number in `RAX`.
     - Map arguments to parameter registers (`RDI`, `RSI`, `RDX`, `R10`, `R8`, `R9`).
     - Emit `MACH_SYSCALL`.

### 🏁 Task Criteria:
* **Input**: An AXIOM source code invoking standard syscall operations.
* **Output**: A relocatable object file containing raw `0F 05` (`syscall`) instruction bytes instead of relocations pointing to dynamic libc.
* **Success Indicator**: A compiled native program runs cleanly on Linux without linking dynamic libc wrappers.

---

## 🎯 Task 2: 100% C-Independent Standard Library (`std/io.ax` & `std/os.ax`)
* **Objective**: Rewrite AXIOM's core I/O and OS standard library modules to use AXIOM-native segmented allocation and direct OS system calls, eliminating dynamic libc FFI dependencies.

### 📋 Technical Design:
1. **Segmented Memory Allocator Integration**:
   - Integrate `std/mem/alloc.ax` directly with standard output/input streams, replacing standard malloc/free calls.
2. **Standard I/O Streams (`std/io.ax`)**:
   - Rewrite `print`, `println`, `puts`, and file descriptor reading/writing to call AXIOM-native `sys_write` and `sys_read` direct syscall wrappers.
3. **OS Interface (`std/os.ax`)**:
   - Rewrite environment variable parsing, command line argument retrieval, and process exit routines to interact directly with the OS stack/registers without libc.

### 🏁 Task Criteria:
* **Input**: Standard library dependencies on raw FFI calls (`fopen`, `fclose`, `printf`, `puts`, `malloc`, `free`).
* **Output**: Pure, self-contained AXIOM standard library modules calling direct binary system calls under `unsafe` blocks.
* **Success Indicator**: Programs like `valid_hello.ax` compile, link, and print stdout natively using only custom allocators and direct system calls.

---

## 🎯 Task 3: Struct, Sum Type, and Match Expression Frontend Support
* **Objective**: Port advanced language features—struct declarations, type aliases, sum types (variants), and match expressions—from the Go driver (Stage 0) into the self-hosted Stage 1 compiler frontend (`bootstrap/stage1/`).

### 📋 Technical Design:
1. **Extend Parser (`bootstrap/stage1/parser.ax`)**:
   - Support `struct` declarations with fields and methods, `type` declarations containing sum type variants (e.g. `type Option[T] = Some(T) | None`), and `match` expressions.
2. **Extend Name Resolver (`bootstrap/stage1/resolver.ax`)**:
   - Handle scoped symbol resolution for match pattern bindings, struct types, and sum type variant constructors.
3. **Extend Type Checker (`bootstrap/stage1/typecheck.ax`)**:
   - Register and validate `TypeTable` entries for struct fields and sum types.
   - Implement type inference for match expressions and decision-tree pattern matching exhaustiveness checking.

### 🏁 Task Criteria:
* **Input**: An AXIOM source compiling complex structs, sum types, and match expressions.
* **Output**: Stage 1 compiler-generated SSA intermediate representation (AIR) containing fully lowered struct methods and discriminant-based tagged union match branches.
* **Success Indicator**: Extended frontend features successfully bootstrap via the triple-build verification loop.
