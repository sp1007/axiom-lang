package opt

import (
	"fmt"
	"time"

	"github.com/axiom-lang/axiom/ir/air"
)

// --------------------------------------------------------------------------
// p10-t01: Optimization Pipeline Manager
//
// Orchestrates a sequence of optimization passes over an AirModule.
// Each pass implements the OptPass interface. The pipeline runs passes
// in order, optionally verifying AIR invariants after each pass.
// --------------------------------------------------------------------------

// OptLevel controls which optimization passes are enabled.
type OptLevel uint8

const (
	O0 OptLevel = iota // no optimization
	O1                 // basic: constant folding + DCE
	O2                 // standard: O1 + inlining + escape analysis + CTGC + loop region
	O3                 // aggressive: O2 + vectorization + SoA transform
)

// String returns the human-readable name for the optimization level.
func (l OptLevel) String() string {
	switch l {
	case O0:
		return "O0"
	case O1:
		return "O1"
	case O2:
		return "O2"
	case O3:
		return "O3"
	default:
		return fmt.Sprintf("O?(%d)", l)
	}
}

// OptPass is the interface that all optimization passes must implement.
// Each pass receives the full AirModule and returns true if any changes
// were made (triggering a potential re-run of the pipeline).
type OptPass interface {
	// Name returns the human-readable name of this pass.
	Name() string

	// Run executes the pass on the given module.
	// Returns true if any instructions were modified.
	Run(mod *air.AirModule) bool
}

// PassStats records execution statistics for a single pass run.
type PassStats struct {
	PassName   string
	Iteration  int
	Changed    bool
	Duration   time.Duration
	VerifyErrs int
}

// PipelineStats records aggregate statistics for the full pipeline run.
type PipelineStats struct {
	Level      OptLevel
	Iterations int
	Passes     []PassStats
	TotalTime  time.Duration
}

// OptPipeline manages an ordered list of optimization passes.
type OptPipeline struct {
	passes       []OptPass
	level        OptLevel
	verify       bool // run verifier after each pass
	maxIter      int  // maximum fixpoint iterations
}

// NewPipeline creates an optimization pipeline for the given level.
// If verify is true, the AIR verifier runs after each pass to catch
// correctness regressions immediately.
func NewPipeline(level OptLevel, verify bool) *OptPipeline {
	return &OptPipeline{
		level:   level,
		verify:  verify,
		maxIter: 10, // sensible default: prevent infinite fixpoint loops
	}
}

// AddPass appends a pass to the pipeline.
func (p *OptPipeline) AddPass(pass OptPass) {
	p.passes = append(p.passes, pass)
}

// Level returns the optimization level of this pipeline.
func (p *OptPipeline) Level() OptLevel {
	return p.level
}

// SetMaxIterations sets the maximum number of fixpoint iterations.
func (p *OptPipeline) SetMaxIterations(n int) {
	if n > 0 {
		p.maxIter = n
	}
}

// Run executes the pipeline on the module.
// It runs all passes in order. If any pass reports a change, the entire
// pipeline is re-run from the start (fixpoint iteration) until no pass
// makes changes or maxIter is reached.
func (p *OptPipeline) Run(mod *air.AirModule) PipelineStats {
	if p.level == O0 || len(p.passes) == 0 {
		return PipelineStats{Level: p.level}
	}

	start := time.Now()
	stats := PipelineStats{Level: p.level}

	for iter := 0; iter < p.maxIter; iter++ {
		anyChanged := false
		stats.Iterations = iter + 1

		for _, pass := range p.passes {
			passStart := time.Now()
			changed := pass.Run(mod)
			passDur := time.Since(passStart)

			ps := PassStats{
				PassName:  pass.Name(),
				Iteration: iter,
				Changed:   changed,
				Duration:  passDur,
			}

			// Verify after pass if enabled
			if p.verify && changed {
				verifyErrs := 0
				for i := range mod.Funcs {
					errs := air.Verify(&mod.Funcs[i])
					verifyErrs += len(errs)
				}
				ps.VerifyErrs = verifyErrs
			}

			stats.Passes = append(stats.Passes, ps)

			if changed {
				anyChanged = true
			}
		}

		if !anyChanged {
			break // fixpoint reached
		}
	}

	stats.TotalTime = time.Since(start)
	return stats
}

// DefaultPipeline creates a pipeline with the default passes for the given level.
// Passes are registered in the standard order defined by the AXIOM optimization spec.
func DefaultPipeline(level OptLevel, verify bool) *OptPipeline {
	p := NewPipeline(level, verify)

	switch level {
	case O1:
		// O1: basic optimizations
		p.AddPass(&ConstantFoldingPass{})
		p.AddPass(&DCEPass{})

	case O2:
		// O2: O1 + advanced
		p.AddPass(&InliningPass{})
		p.AddPass(&CopyPropagationPass{})
		p.AddPass(&ConstantFoldingPass{})
		p.AddPass(&DCEPass{})
		p.AddPass(&LoopRegionPass{})
		p.AddPass(&LoopUnrollPass{})
		p.AddPass(&GVNPass{})
		p.AddPass(&CopyPropagationPass{})
		p.AddPass(&DCEPass{})
		// Future: p.AddPass(&EscapeAnalysisPass{})
		// Future: p.AddPass(&CTGCPass{})

	case O3:
		// O3: O2 + aggressive
		p.AddPass(&InliningPass{})
		p.AddPass(&CopyPropagationPass{})
		p.AddPass(&ConstantFoldingPass{})
		p.AddPass(&DCEPass{})
		p.AddPass(&LoopRegionPass{})
		p.AddPass(&LoopUnrollPass{})
		p.AddPass(&GVNPass{})
		p.AddPass(&CopyPropagationPass{})
		p.AddPass(&DCEPass{})
		// Future: all O2 passes + vectorization + SoA
	}

	return p
}
