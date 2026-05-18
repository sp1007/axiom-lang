# p16-t20: `std.crypto` Implementation

## Purpose
Implement the standard library cryptography module providing SHA-256, SHA-512, and ChaCha20 implementations in AXIOM. These are essential for the package manager's lock file verification (`axiom.lock` uses SHA-256) and general-purpose hashing.

## Context
Plan §Phase 9 lists: _"`std/crypto.ax` — SHA-256, SHA-512, ChaCha20"_. The package manager (p17-t04) requires SHA-256 for `tree_hash` verification in `axiom.lock`.

## Inputs
- AXIOM array/slice types from std.collections (p16-t03)
- AXIOM integer arithmetic
- NIST test vectors for SHA-256/512

## Outputs
- `std/crypto/sha256.ax` — SHA-256 implementation
- `std/crypto/sha512.ax` — SHA-512 implementation
- `std/crypto/chacha20.ax` — ChaCha20 stream cipher
- `std/crypto/hmac.ax` — HMAC construction
- Test files with NIST test vectors

## Dependencies
- p16-t01: std-testing-assert — test framework
- p16-t02: std-string — byte array handling
- p16-t03: std-collections — Array[u8] type
- p08-t10: e2e-compliance-tests — compiler can build .ax files

## Subsystems Affected
- Package manager (p17-t04): uses SHA-256 for tree_hash
- Standard library: crypto module

## Detailed Requirements

### SHA-256 API
```axiom
pub fn sha256(data: Seq[u8]) -> Array[u8, 32]
pub fn sha256_hex(data: Seq[u8]) -> string

pub struct Sha256Hasher:
    fn new() -> Sha256Hasher
    fn update(mut self, data: Seq[u8])
    fn finalize(self) -> Array[u8, 32]
```

### SHA-512 API
```axiom
pub fn sha512(data: Seq[u8]) -> Array[u8, 64]
```

### ChaCha20 API
```axiom
pub fn chacha20_encrypt(key: Array[u8, 32], nonce: Array[u8, 12], plaintext: Seq[u8]) -> Seq[u8]
pub fn chacha20_decrypt(key: Array[u8, 32], nonce: Array[u8, 12], ciphertext: Seq[u8]) -> Seq[u8]
```

### HMAC API
```axiom
pub fn hmac_sha256(key: Seq[u8], message: Seq[u8]) -> Array[u8, 32]
```

### Implementation Notes
- Pure AXIOM implementation (no FFI to C crypto libraries)
- Constant-time comparison for hash equality (prevent timing attacks)
- No dynamic allocation in hash core (fixed buffers on stack)

## Implementation Steps

1. Implement SHA-256 per FIPS 180-4.
2. Implement SHA-512 per FIPS 180-4.
3. Implement HMAC per RFC 2104.
4. Implement ChaCha20 per RFC 8439.
5. Write tests using NIST test vectors.
6. Implement `sha256_hex` for human-readable output.

## Test Plan

- `TestSHA256Empty`: SHA-256("") → known hash
- `TestSHA256ABC`: SHA-256("abc") → known hash
- `TestSHA256Long`: SHA-256 of 1M bytes → known hash
- `TestSHA512NIST`: NIST test vectors
- `TestChaCha20RFC`: RFC 8439 test vectors
- `TestHMACSHA256`: RFC 4231 test vectors
- `TestConstantTimeCompare`: equal and unequal hashes take similar time

## Acceptance Criteria

- All NIST/RFC test vectors pass
- `sha256("abc")` produces `ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad`

## Definition of Done

- [ ] `std/crypto/sha256.ax` implemented and tested
- [ ] `std/crypto/sha512.ax` implemented and tested
- [ ] `std/crypto/chacha20.ax` implemented and tested
- [ ] All test vectors pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Side-channel attacks | Use constant-time operations for comparisons; document limitations |
| Integer overflow in hash computation | Use explicit u32/u64 wrapping arithmetic |

## Future Follow-up Tasks

- p17-t04: Package manager uses sha256 for tree_hash
- Future: TLS implementation using these primitives
