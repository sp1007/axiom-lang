# p16-t24: `std.quantum` — Stub (Bool-Based Simulator)

## Purpose
Implement the `std.quantum` stub module as a bool-based quantum simulator. This satisfies compliance tests 097–099 without requiring real QPU hardware. Per plan §Issue 4: _"Real QPU: Phase 10+ (marked [Future])"_.

## Context
Plan §Phase 9 lists: _"`std/quantum.ax` — stub (bool-based simulator)"_. Plan §Issue 4 resolves: _"Implement `std.quantum` as a stub module: `quantum.alloc_qbit()` → returns a `qbit` (internally a `bool` with random initial value)"_.

AIR includes quantum opcodes (`qalloc`, `qgate`, `qmeasure`) which map to calls into this stub module.

## Inputs
- AIR quantum opcodes from p09-t01
- Compliance tests 097–099
- `std.random` from p16-t16

## Outputs
- `std/quantum.ax` — stub quantum module
- Tests

## Dependencies
- p16-t01: std-testing-assert — test framework
- p16-t16: std-random — random number generation for probabilistic simulation

## Detailed Requirements

### API Surface

```axiom
pub type Qbit:
    _value: bool        // internal: simulated quantum state
    _collapsed: bool    // true after measurement

pub fn alloc_qbit() -> Qbit:
    return Qbit { _value: random_bool(), _collapsed: false }

pub fn H(mut q: Qbit):
    // Hadamard gate stub: 50/50 coin flip
    if not q._collapsed:
        q._value = random_bool()

pub fn X(mut q: Qbit):
    // Pauli-X (NOT) gate
    q._value = not q._value

pub fn CNOT(control: Qbit, mut target: Qbit):
    if control._value:
        target._value = not target._value

pub fn measure(mut q: Qbit) -> bool:
    q._collapsed = true
    return q._value
```

### C-Backend Mapping
- AIR `qalloc` → `_AX_std_quantum_alloc_qbit()`
- AIR `qgate %q, H` → `_AX_std_quantum_H(&q)`
- AIR `qmeasure %q` → `_AX_std_quantum_measure(&q)`

## Implementation Steps

1. Create `std/quantum.ax` with `Qbit` type.
2. Implement `alloc_qbit()` using `std.random`.
3. Implement gates: `H`, `X`, `CNOT`.
4. Implement `measure()`.
5. Verify compliance tests 097–099 pass.

## Test Plan

- `TestQbitAlloc`: `alloc_qbit()` returns valid Qbit
- `TestHGate`: after H, measure produces both true and false over 1000 trials
- `TestXGate`: `X(q)` flips value
- `TestCNOT`: control=true → target flipped; control=false → target unchanged
- `TestMeasureCollapse`: second measure returns same value as first

## Acceptance Criteria

- Compliance tests 097–099 pass
- AIR quantum opcodes lower to stub calls without error

## Definition of Done

- [ ] `std/quantum.ax` implemented
- [ ] Compliance tests 097–099 pass
- [ ] AIR quantum opcodes mapped to stub

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Non-deterministic test results | Use seeded PRNG for tests; probabilistic assertions (>40% and <60% over 1000 trials) |
