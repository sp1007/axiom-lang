// Package builder provides the AIR construction API.
// It translates typed AST / semantic graph nodes into AIR instructions
// and basic blocks, maintaining SSA invariants during construction.
//
// The builder walks the typed AstTree from the semantic analysis phase
// and emits AIR instructions into basic blocks using the AirFuncBuilder.
//
// Architecture:
//
//	AstTree + SymbolTable + TypeTable
//	         │
//	         ▼
//	    ModuleBuilder     (top-level: iterates functions)
//	         │
//	         ▼
//	    FuncLowering      (per-function: manages blocks, scopes, locals)
//	    ├── ExprLowering  (expressions → AIR values)
//	    └── StmtLowering  (statements → AIR instructions + control flow)
package builder

import (
	"sort"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/ir/air"
)

// ModuleBuilder translates an entire compilation unit (AstTree) into an AirModule.
type ModuleBuilder struct {
	tree    *ast.AstTree
	symbols *sema.SymbolTable
	ttable  *types.TypeTable
	intern  *ast.InternPool
	module  *air.AirModule
}

// NewModuleBuilder creates a builder for the given compilation context.
func NewModuleBuilder(
	tree *ast.AstTree,
	symbols *sema.SymbolTable,
	ttable *types.TypeTable,
	intern *ast.InternPool,
) *ModuleBuilder {
	return &ModuleBuilder{
		tree:    tree,
		symbols: symbols,
		ttable:  ttable,
		intern:  intern,
		module:  &air.AirModule{},
	}
}

// Build walks all top-level function declarations and lowers them to AIR.
// Returns the completed AirModule.
func (mb *ModuleBuilder) Build() *air.AirModule {
	// 1. Build main tree
	mb.buildTree(mb.tree)

	// 2. Build all loaded module trees
	if mb.symbols != nil && mb.symbols.LazyResolver != nil {
		var modKeys []uint32
		for k := range mb.symbols.LazyResolver.GetModules() {
			modKeys = append(modKeys, k)
		}
		sort.Slice(modKeys, func(i, j int) bool {
			return modKeys[i] < modKeys[j]
		})
		for _, k := range modKeys {
			mod := mb.symbols.LazyResolver.GetModules()[k]
			if mod.AstTree != nil {
				mb.buildTree(mod.AstTree)
			}
		}
	}

	return mb.module
}

func (mb *ModuleBuilder) buildTree(tree *ast.AstTree) {
	oldTree := mb.tree
	mb.tree = tree
	defer func() { mb.tree = oldTree }()

	root := tree.Node(0) // NodeProgram
	child := root.FirstChild
	for child != ast.NullIdx {
		node := tree.Node(child)
		if node.Kind == ast.NodeFuncDecl {
			fn := mb.lowerFunc(child, node)
			if fn != nil {
				mb.module.Funcs = append(mb.module.Funcs, *fn)
			}
		} else if node.Kind == ast.NodeStructDecl {
			sChild := node.FirstChild
			for sChild != ast.NullIdx {
				sNode := tree.Node(sChild)
				if sNode.Kind == ast.NodeFuncDecl {
					fn := mb.lowerFunc(sChild, sNode)
					if fn != nil {
						mb.module.Funcs = append(mb.module.Funcs, *fn)
					}
				}
				sChild = sNode.NextSibling
			}
		}
		child = node.NextSibling
	}
}

// lowerFunc creates an AirFunc from a NodeFuncDecl AST node.
func (mb *ModuleBuilder) lowerFunc(idx uint32, node *ast.AstNode) *air.AirFunc {
	// Skip extern declarations (no body)
	if node.Flags&ast.FlagIsExtern != 0 {
		return nil
	}

	// Extract function info from symbol table
	nameID := uint32(0)
	retTypeID := uint32(0)
	var paramTypeIDs []uint32

	symIdx := node.Payload
	if symIdx != 0 && int(symIdx) < len(mb.symbols.Symbols) {
		sym := mb.symbols.SymbolAt(symIdx)
		nameID = sym.NameID
		if sym.TypeID != 0 {
			entry := mb.ttable.Entry(types.TypeID(sym.TypeID))
			if entry.Kind == types.KindFunction {
				fi := mb.ttable.FuncInfo(types.TypeID(sym.TypeID))
				retTypeID = uint32(fi.Return)
				for _, p := range fi.Params {
					paramTypeIDs = append(paramTypeIDs, uint32(p))
				}
			}
		}
	}

	// If no symbol found, use token text for the name
	if nameID == 0 {
		text := mb.tree.TokenText(node.TokenIdx)
		nameID = mb.intern.Intern(text)
	}

	fb := air.NewAirFuncBuilder(nameID, retTypeID)
	fl := newFuncLowering(mb, fb, paramTypeIDs)

	// Register function parameters as SSA values
	fl.registerParams(idx, node)

	// Find the body block and lower it
	bodyIdx := mb.findBody(node)
	if bodyIdx != ast.NullIdx {
		fl.lowerBlock(bodyIdx)
	}

	// Ensure the function ends with a return if missing
	fl.ensureReturn()

	fn := fb.Build()
	fn.SymID = symIdx
	fn.Params = paramTypeIDs
	fn.IsAsync = node.Flags&ast.FlagIsAsync != 0
	return fn
}

// findBody locates the body block among a function's children.
func (mb *ModuleBuilder) findBody(node *ast.AstNode) uint32 {
	var bodyIdx uint32 = ast.NullIdx
	child := node.FirstChild
	for child != ast.NullIdx {
		cn := mb.tree.Node(child)
		if cn.Kind == ast.NodeBlock {
			bodyIdx = child
		}
		child = cn.NextSibling
	}
	return bodyIdx
}
