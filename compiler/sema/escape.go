package sema

import (
	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/types"
)

// EscapeReport summarizes escape analysis results for a function.
type EscapeReport struct {
	AllocatesOnHeap  []uint32 // symIDs of variables that escape to heap
	AllocatesOnStack []uint32 // symIDs of variables that stay on stack
}

// EscapeAnalysis determines whether each allocation escapes its declaring
// scope (heap-allocated) or stays within scope (stack-allocated).
// Stack allocation eliminates heap overhead and makes CTGC automatic.
type EscapeAnalysis struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable

	// sizeThreshold: values larger than this (bytes) are always heap-allocated.
	// Default: 1024 bytes (1KB).
	sizeThreshold uint32
}

// NewEscapeAnalysis creates a new EscapeAnalysis pass.
func NewEscapeAnalysis(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable) *EscapeAnalysis {
	return &EscapeAnalysis{
		ast:           tree,
		intern:        intern,
		symtable:      st,
		types:         tt,
		sizeThreshold: 1024, // 1KB default
	}
}

// SetSizeThreshold configures the threshold above which values are always heap-allocated.
func (ea *EscapeAnalysis) SetSizeThreshold(bytes uint32) {
	ea.sizeThreshold = bytes
}

// AnalyzeFunction performs escape analysis for a single function, using a
// pre-built ConnectionGraph. Sets FlagEscapesToHeap on AST nodes that escape.
// Returns an EscapeReport summarizing the results.
func (ea *EscapeAnalysis) AnalyzeFunction(funcNodeIdx uint32, cg *ConnectionGraph) EscapeReport {
	report := EscapeReport{}

	// Walk the function's children to find VarDecl nodes
	ea.analyzeNode(funcNodeIdx, cg, &report)

	return report
}

func (ea *EscapeAnalysis) analyzeNode(nodeIdx uint32, cg *ConnectionGraph, report *EscapeReport) {
	node := &ea.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeVarDecl {
		symID := node.Payload
		if symID != 0 {
			escapes := false

			// Check escape via ConnectionGraph
			if cgNodeID, ok := cg.NodeOfSym(symID); ok {
				escapes = cg.Escapes(cgNodeID)
			}

			// Size threshold check: large values always go to heap
			if !escapes && ea.sizeThreshold > 0 {
				if sym := ea.symtable.SymbolAt(symID); sym.TypeID != 0 {
					typeID := types.TypeID(sym.TypeID)
					if int(typeID) < ea.types.Count() {
						entry := ea.types.Entry(typeID)
						if entry.Size > ea.sizeThreshold {
							escapes = true
						}
					}
				}
			}

			if escapes {
				ea.ast.SetFlags(nodeIdx, ast.FlagEscapesToHeap)
				report.AllocatesOnHeap = append(report.AllocatesOnHeap, symID)
			} else {
				report.AllocatesOnStack = append(report.AllocatesOnStack, symID)
			}
		}
	}

	// Recurse into children
	child := node.FirstChild
	for child != 0 {
		ea.analyzeNode(child, cg, report)
		child = ea.ast.Nodes[child].NextSibling
	}
}

// AnalyzeAll performs escape analysis on all functions in the AST.
// Returns a map from function symID to EscapeReport.
func (ea *EscapeAnalysis) AnalyzeAll(cgs map[uint32]*ConnectionGraph) map[uint32]EscapeReport {
	reports := make(map[uint32]EscapeReport)
	ea.analyzeAllNode(0, cgs, reports)
	return reports
}

func (ea *EscapeAnalysis) analyzeAllNode(nodeIdx uint32, cgs map[uint32]*ConnectionGraph, reports map[uint32]EscapeReport) {
	node := &ea.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeFuncDecl {
		funcSym := node.Payload
		if cg, ok := cgs[funcSym]; ok {
			report := ea.AnalyzeFunction(nodeIdx, cg)
			reports[funcSym] = report
		}
	}

	child := node.FirstChild
	for child != 0 {
		ea.analyzeAllNode(child, cgs, reports)
		child = ea.ast.Nodes[child].NextSibling
	}
}
