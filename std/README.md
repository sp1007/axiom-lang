# AXIOM Standard Library

This directory contains AXIOM source files (`.ax`), **not** Go source code.

Do NOT place `.go` files here. The AXIOM standard library modules are written
in AXIOM itself and will be compiled by the AXIOM compiler during the build
process.

## Structure

```
std/
├── testing/     — test framework and assertions
├── string/      — string manipulation
├── collections/ — generic data structures
├── io/          — I/O abstractions
├── math/        — mathematical functions
├── net/         — networking
├── process/     — process management
├── sync/        — synchronization primitives
├── json/        — JSON serialization
├── time/        — time and duration
├── fmt/         — formatting and printing
├── result/      — Result and Option types
├── log/         — structured logging
├── os/          — OS interface
├── iter/        — iterators
├── random/      — random number generation
├── cli/         — CLI argument parsing
├── ffi/         — foreign function interface
├── crypto/      — cryptographic primitives
└── mem/         — memory utilities
```
