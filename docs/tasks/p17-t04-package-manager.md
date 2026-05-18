# p17-t04: Package Manager (axpkg)

## Purpose
Implement the AXIOM package manager (`axpkg`) for declaring dependencies, resolving versions, downloading packages, and building project dependency trees.

## Context
AXIOM needs a package ecosystem. `axpkg` manages packages declared in `axiom.toml`, resolves semantic versioning constraints, downloads from a registry (initially a git-based registry), and produces a lockfile for reproducible builds.

## Inputs
- `axiom.toml` — project manifest (name, version, dependencies)
- `axiom.lock` — lockfile with pinned versions and content hashes
- Package registry (initially GitHub-based git tags)

## Outputs
- `tools/pkg/axpkg.go` — package manager
- `axiom.toml` schema
- `axiom.lock` lockfile format
- `axpkg` CLI: `add`, `remove`, `update`, `install`, `publish`

## Dependencies
- p16-t09: std-json — manifest/lockfile parsing (TOML via C library)
- p16-t07: std-process — git subprocess for package download
- p16-t06: std-net — HTTP for registry API

## Detailed Requirements

`axiom.toml` format:
```toml
[package]
name = "myapp"
version = "1.0.0"
authors = ["Author Name <email>"]
license = "MIT"

[dependencies]
axiom-http = "^2.1"
axiom-json = "1.0.0"

[dev-dependencies]
axiom-mock = "^0.5"
```

`axiom.lock` format (JSON):
```json
{
  "packages": [
    {
      "name": "axiom-http",
      "version": "2.3.1",
      "source": "registry+https://pkg.axiom-lang.org",
      "checksum": "sha256:abc123..."
    }
  ]
}
```

Version resolution algorithm:
1. Parse all `axiom.toml` dependency constraints.
2. Fetch available versions from registry.
3. Run SAT-based resolver (PubGrub algorithm) for compatible version set.
4. Write resolved versions to `axiom.lock`.
5. Download missing packages to `~/.axiom/cache/`.

```go
type Package struct {
    Name    string
    Version semver.Version
    Source  string
    Hash    string  // SHA256
}

func Resolve(manifest Manifest, lockfile Lockfile) ([]Package, error)
func Download(pkg Package, cacheDir string) error
func Install(dir string) error  // download all locked deps
```

## Implementation Steps

1. Create `tools/pkg/axpkg.go`.
2. Implement TOML parser for `axiom.toml` (via C libtomk or pure-Go library).
3. Implement semantic versioning constraint parser (`^2.1` → `>=2.1.0, <3.0.0`).
4. Implement PubGrub resolver (or simpler greedy-latest for MVP).
5. Implement download: git clone tag or HTTP tarball + SHA256 verify.
6. Implement lockfile generation and read.
7. Implement `axpkg install`, `axpkg add`, `axpkg update` CLI commands.

## Test Plan
- `TestResolveSimple`: one dependency, one version → resolved
- `TestResolveConflict`: incompatible constraints → error with explanation
- `TestLockfileGeneration`: resolve → lockfile written with correct hashes
- `TestInstall`: lockfile present → packages downloaded to cache
- `TestChecksumVerify`: corrupted download → error with hash mismatch

## Validation Checklist
- [ ] Reproducible: same axiom.toml + axiom.lock → same build always
- [ ] Checksum verified before using downloaded package
- [ ] Version resolution considers transitive dependencies
- [ ] --offline flag: use only cached packages

## Acceptance Criteria
- `axpkg add axiom-http@2` updates axiom.toml and axiom.lock correctly

## Definition of Done
- [ ] `tools/pkg/axpkg.go` implemented
- [ ] Install and resolve tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Diamond dependency version conflict | PubGrub produces human-readable conflict explanation |
| Registry unavailable | --offline fallback using cache; lockfile always present for CI |

## Future Follow-up Tasks
- Private registry support
- `axpkg publish` to upload packages
- Workspace support (monorepo with multiple packages)
