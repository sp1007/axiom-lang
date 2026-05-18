# p17-t08: Build System (axiom.toml Integration)

## Purpose
Implement the AXIOM build system that reads `axiom.toml` project manifests, manages multi-file builds, handles build profiles (debug/release), and integrates with the package manager and incremental compiler.

## Context
`axc build` needs to understand project structure: which files to compile, in what order, with what flags, linking which libraries. The build system reads `axiom.toml`, resolves the build graph, invokes the compiler for each module, and links the final executable — replacing ad-hoc `axc compile` invocations.

## Inputs
- `axiom.toml` project manifest
- Source files in `src/` directory
- Dependency packages from `axpkg install`
- Build profiles: `[profile.debug]`, `[profile.release]`

## Outputs
- `tools/build/build.go` — build system driver
- `axc build` subcommand (with `--release`, `--debug`, `--target` flags)
- `target/` directory for build artifacts

## Dependencies
- p17-t04: package-manager — resolves and downloads dependencies
- p17-t05: incremental-compilation — skip unchanged files
- p12-t05: incremental-linker — final link step
- p11-t01: target-triple — target platform selection

## Detailed Requirements

`axiom.toml` build configuration:
```toml
[package]
name = "myapp"
version = "1.0.0"

[build]
src = "src/"
entry = "src/main.ax"
output = "myapp"

[profile.debug]
opt_level = 0
debug_info = true

[profile.release]
opt_level = 3
debug_info = false
lto = true

[features]
networking = ["axiom-http"]

[dependencies]
axiom-http = { version = "^2.0", optional = true }
```

```go
type BuildConfig struct {
    Package  PackageConfig
    Build    BuildSection
    Profiles map[string]ProfileConfig
    Deps     map[string]DepSpec
}

type BuildContext struct {
    Config  BuildConfig
    Profile ProfileConfig
    Target  Target
    Outdir  string
}

func Build(ctx BuildContext) error
func (ctx *BuildContext) collectSources() []string
func (ctx *BuildContext) buildDependencyOrder(sources []string) [][]string  // toposorted
func (ctx *BuildContext) compileModule(path string) (string, error)  // returns .o path
func (ctx *BuildContext) linkFinal(objects []string) error
```

Build artifact layout:
```
target/
  debug/
    myapp           (debug executable)
    *.o             (object files)
  release/
    myapp           (optimized executable)
```

## Implementation Steps

1. Create `tools/build/build.go`.
2. Implement `axiom.toml` parser (TOML format).
3. Implement source collection: glob `src/**/*.ax`.
4. Implement topological sort of modules by import dependencies.
5. Implement parallel compilation: compile independent modules concurrently.
6. Implement profile selection (debug/release flag sets).
7. Implement `target/` layout management.
8. Add `axc build` subcommand.

## Test Plan
- `TestBuildSimple`: single-file project → executable built
- `TestBuildMultiFile`: 3-file project → all compiled and linked
- `TestBuildRelease`: --release → O3 flags used
- `TestBuildIncremental`: change one file, rebuild → only changed file recompiled
- `TestBuildTarget`: --target=x86_64-linux → cross-compiled output

## Validation Checklist
- [ ] Parallel compilation doesn't cause race conditions
- [ ] Target directory created if absent
- [ ] Debug and release profiles use correct flags
- [ ] Build fails cleanly if source has errors (no corrupt artifacts)

## Acceptance Criteria
- `axc build` on the AXIOM stdlib builds all modules successfully

## Definition of Done
- [ ] `tools/build/build.go` implemented
- [ ] Multi-file project build test passes

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Circular import → infinite toposort loop | Detect cycle in toposort, report error with cycle path |
| Parallel build races on shared state | Each compiler invocation uses own temp directory |

## Future Follow-up Tasks
- `axc run` — build + execute in one command
- Custom build scripts (`build.ax` run before compilation)
