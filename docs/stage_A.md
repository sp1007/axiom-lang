# AXIOM Toolchain — Stage A Task Specifications

This document defines the task specifications, inputs, outputs, and success criteria to implement AXIOM's Dynamic Import resolution (PE/COFF IAT & ELF64 PLT/GOT) and instruction selector Direct System Calls.

---

## 🎯 Task 1: Direct Binary Syscall Selector Lowering
* **Objective**: Teach AXIOM's instruction selector (`x86_selector.ax`) and machine encoder (`x86_encoding.ax`) to emit direct binary `syscall` instructions for standard low-level kernel FFI APIs (`sys_write`, `sys_exit`, `sys_mmap`, `sys_munmap`) on x86-64 Linux targets, bypassing standard dynamic C libraries completely.

### 📋 Task Details:
1. **Machine Opcode**:
   - Define `pub const MACH_SYSCALL: u16 = 23` in `x86_selector.ax` and `x86_encoding.ax`.
2. **Byte Serializer**:
   - Add `x86_encode_syscall(buf: ptr[ByteVec])` emitting the `0F 05` instruction bytes in `x86_encoding.ax`.
3. **Instruction Lowering**:
   - Update `x86_selector.ax`'s `OP_CALL` lowering logic: If compiling for Linux/sysv ABI and the FFI function called is a core OS system symbol (e.g. `write`, `exit`, `mmap`, `munmap`), lower the call directly to a syscall sequence:
     - Load the system call number in `RAX`.
     - Map arguments to parameter registers (`RDI`, `RSI`, `RDX`, `R10`, `R8`, `R9`).
     - Emit `MACH_SYSCALL`.

### 🏁 Task Criteria:
* **Input**: An AXIOM source code invoking standard syscall operations (e.g. FFI `write` or `exit`).
* **Output**: A relocatable object file containing raw `0F 05` (`syscall`) instruction bytes instead of `E8` (`call`) relocations pointing to dynamic libc.
* **Success Indicator**: A compiled native program runs cleanly on Linux without linking dynamic libc wrappers.

---

## 🎯 Task 2: PE/COFF Windows Import Address Table (IAT) Solver
* **Objective**: Upgrade AXIOM's custom linker (`linker.ax`) to construct PE/COFF dynamic Import tables, enabling self-linked AXIOM executables to dynamically bind to standard Windows API DLLs (like `kernel32.dll` and `ucrtbase.dll`) at load-time.

### 📋 Task Details:
1. **Dynamic Import Table Structure**:
   - Design an `.idata` section generator building PE/COFF standard structures:
     - Import Directory Table (listings of DLL descriptors).
     - Import Lookup Table (ILT) and Import Address Table (IAT) mapping function names/ordinals.
     - Hint/Name Table (containing the raw function name strings).
2. **Dynamic Relocation Resolution**:
   - When the linker parses an external dynamic symbol (e.g. `VirtualAlloc`), map the dynamic reference to the corresponding resolved IAT slot address using relative ModR/M addressing (`call qword ptr [rip + displacement]`).

### 🏁 Task Criteria:
* **Input**: COFF relocatable object files (`.obj`) carrying unresolved external DLL symbol relocations.
* **Output**: A fully compliant Windows PE executable (`.exe`) containing MZ/PE directories and a valid `.idata` segment.
* **Success Indicator**: The generated `.exe` runs successfully from the command line, dynamic symbols are bound by the Windows OS loader, and `hello_custom.exe` runs flawlessly.

---

## 🎯 Task 3: ELF64 Linux PLT/GOT Relocation Resolver
* **Objective**: Upgrade the custom linker (`linker.ax`) to generate Procedure Linkage Table (PLT) and Global Offset Table (GOT) segments for ELF64 executables, supporting shared library FFI linkages.

### 📋 Task Details:
1. **GOT/PLT Segment Generator**:
   - Build GOT/PLT segments mapping dynamic symbol slots.
   - Patch `R_X86_64_PLT32` relative branches pointing directly to PLT jump slots.

### 🏁 Task Criteria:
* **Input**: Relocatable ELF64 object files (`.o`) carrying PLT relative branches.
* **Output**: A static or dynamically-linked ELF64 executable carrying active `.got` and `.plt` headers.
* **Success Indicator**: Executable successfully verified via host tools (`readelf -d` and execution).
