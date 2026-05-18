# p17-t05: Incremental Compilation

## Purpose
Implement incremental compilation that re-compiles only changed files and their dependents, reducing rebuild times from O(project size) to O(changed files).

## Context
For large AXIOM projects, full recompilation is too slow. Incremental compilation uses a dependency graph: when file A changes, only A and files that import A (transitively) are recompiled. Content hashing ensures correctness — only files with changed content trigger recompilation.

## Inputs
- Content hash (SHA256) of each source file
- Import dependency graph (file → files it imports)
- Previous compilation artifacts (cached .o files)
- Changed files (detected by comparing hashes)

## Outputs
- `tools/incremental/incr.go` — incremental compilation driver
- `build.db` — compilation database (SQLite-lite or JSON) storing hashes + artifacts
- `axc build --incremental` flag

## Dependencies
- p03: parser — imports declared at top of file (for dependency graph)
- p11-t15: native-backend-integration — produces cacheable .o artifacts
- p12-t05: incremental-linker — incremental link step

## Detailed Requirements

```go
type BuildCache struct {
    Files   map[string]FileRecord  // path → {hash, artifact_path, deps}
    Version string
}

type FileRecord struct {
    ContentHash  string    // SHA256 of source
    ArtifactPath string    // path to cached .o
    Imports      []string  // direct import dependencies
    CompiledAt   time.Time
}

type IncrementalBuilder struct {
    Cache   BuildCache
    CacheDB string  // path to build.db
}

func (b *IncrementalBuilder) Build(files []string) error
func (b *IncrementalBuilder) dirtyFiles(files []string) []string
func (b *IncrementalBuilder) transitiveDependers(changed []string) []string
func (b *IncrementalBuilder) compileFile(path string) error
func (b *IncrementalBuilder) linkAll() error
```

Algorithm:
1. Compute SHA256 of all source files.
2. Compare against cache: find files with changed hashes → dirty set.
3. Compute transitive closure of dependents: all files that (transitively) import a dirty file.
4. Recompile dirty set + transitive dependents.
5. Link: reuse cached .o for unchanged files.
6. Update cache with new hashes and artifact paths.

Cache storage: JSON file `~/.axiom/buildcache/<project-hash>.json`.

Dependency extraction: parse only the import declarations (first N lines), not full parse.

## Implementation Steps

1. Create `tools/incremental/incr.go`.
2. Implement content hash computation (SHA256 via crypto/sha256).
3. Implement cache read/write (JSON serialization).
4. Implement import dependency extraction (fast regex on import lines).
5. Implement dirty file detection and transitive closure.
6. Implement build invocation for dirty files only.
7. Implement cache update after successful compilation.
8. Add `axc build --incremental` flag.

## Test Plan
- `TestIncrNoChange`: no changes → zero files recompiled
- `TestIncrOneChange`: change file A → A and its importers recompiled
- `TestIncrCacheInvalid`: corrupt cache → full recompile (graceful recovery)
- `TestIncrSpeedup`: 100-file project, 1 change → < 5% of full build time

## Validation Checklist
- [ ] Changed files always recompiled (no stale artifacts)
- [ ] Unchanged files never recompiled (hash match)
- [ ] Transitive dependents correctly computed
- [ ] Cache file atomic write (no corruption on crash)

## Acceptance Criteria
- 100-file project with 1 change: incremental build < 500ms vs 10s full build

## Definition of Done
- [ ] `tools/incremental/incr.go` implemented
- [ ] Incremental speedup test demonstrates improvement

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Cache stale after compiler version change | Include compiler version in cache key; invalidate on version change |
| Race condition: parallel build updates cache | Atomic JSON write via temp file + rename |

## Future Follow-up Tasks
- Fine-grained function-level incremental (only recompile changed functions)
- Distributed build cache (share cache across CI machines)
