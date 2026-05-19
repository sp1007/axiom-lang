# AXIOM Test Data

This directory contains test data files for the AXIOM compiler test suites.
It is **not** a Go package — it contains `.ax` source files, expected outputs,
and golden test snapshots used by the Go test harness.

## Structure

```
tests/
├── lexer/    — lexer test inputs and expected token streams
├── parser/   — parser test inputs and expected AST snapshots
└── sema/     — semantic analysis test inputs and expected diagnostics
```
