# p10-t01: Optimization Pipeline Manager

## Purpose
Implement the optimization pipeline manager that orchestrates optimization passes, runs them in the correct order, manages fixpoint iteration, and enforces verification between passes.

## Context
The pipeline manager is the conductor of the optimization orchestra. It defines three optimization tiers (O0, O1, O2/O3), orders passes correctly (constant folding before DCE, escape analysis before CTGC), and runs the AIR verifier after each pass to catch bugs early. A pass returns true if it made any changes; the pipeline iterates until no pass makes changes (fixpoint).

## Inputs
- `AirModule` — the module to optimize
- Optimization level flag (`-O0`, `-O1`, `-O2`, `-O3`)
- `OptPass` interface implementations from subsequent tasks

## Outputs
- `ir/opt/pipeline.go` — OptPipeline, OptPass interface
- Optimized `AirModule`

## Dependencies
- p09-t04: air-verifier — runs between passes
- p09-t01: air-instruction-set — passes operate on AirInst

## Subsystems Affected
- All optimization passes: registered and orchestrated here
- AIR verifier: invoked after each pass
- Build pipeline: called from `axc build` before codegen

## Detailed Requirements

1. `OptPass` interface:
   ```go
   type OptPass interface {
       Name() string
       Run(module *AirModule) bool  // returns true if any change made
   }
   ```
2. `OptPipeline` struct: `passes []OptPass, verifier *Verifier, level OptLevel`
3. `OptLevel` enum: `O0 = 0, O1 = 1, O2 = 2, O3 = 3`
4. Pass ordering (all levels):
   - O0: no passes (emit as-is)
   - O1: ConstantFolding → DCE
   - O2: O1 + FunctionInlining → ConstantFolding → DCE → EscapeAnalysis → CTGCOnAIR → LoopRegionDetection
   - O3: O2 + Vectorization → SoATransform → CTGCOnAIR (again)
5. Fixpoint loop for each pass group: run passes until no changes (max 10 iterations to prevent infinite loop).
6. After each pass: run `verifier.Verify(func)` for all functions; if errors found, panic with "pass X produced invalid AIR".
7. `PassStats`: track iterations per pass, time per pass, changes per pass. Print with `--stats` flag.
8. `NewPipeline(level OptLevel, verifier *Verifier) *OptPipeline`
9. `Run(module *AirModule)` — main entry point.

## Implementation Steps

1. Create `ir/opt/pipeline.go` with OptPass interface and OptPipeline.
2. Implement `Run()` with per-level pass lists and fixpoint loop.
3. Implement verification after each pass.
4. Implement PassStats collection.
5. Add `--stats` output in axc build.
6. Register all passes (implementations added in subsequent tasks).
7. Write unit tests: `TestPipelineO0`, `TestPipelineO1`, `TestPipelineFixpoint`.

## Test Plan

- `TestPipelineO0`: O0 makes no changes to any function
- `TestPipelineO1ConstFold`: simple arithmetic → folded by O1
- `TestPipelineVerification`: deliberately produce bad AIR after "pass" → pipeline panics
- `TestPipelineFixpoint`: idempotent O1 → terminates in 1 iteration

## Validation Checklist

- [ ] O0 makes no changes
- [ ] O1 runs exactly ConstFold + DCE
- [ ] O2 runs all passes in correct order
- [ ] Verifier runs after each pass
- [ ] Fixpoint terminates in ≤ 10 iterations
- [ ] PassStats available with --stats

## Acceptance Criteria

- `axc build -O2 hello.ax` runs O2 pipeline without errors
- `axc build -O0 hello.ax` produces correct output with no optimization

## Definition of Done

- [ ] `ir/opt/pipeline.go` implemented
- [ ] OptPass interface defined
- [ ] Unit tests pass

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Pass ordering produces incorrect optimization | Document ordering rationale in comments; test order matters |
| Infinite fixpoint loop | Max 10 iterations hard limit with warning |

## Future Follow-up Tasks

- p10-t02 through p10-t09: individual pass implementations registered here
- p10-t10: cbackend-v2-from-air called after optimization pipeline
