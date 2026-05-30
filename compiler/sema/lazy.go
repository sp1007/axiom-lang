package sema

import (
	"fmt"
	"runtime/debug"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// ModuleLoader is a callback that populates a module's exports from its AST.
// In the real compiler, this runs the first phase of name resolution on the module.
type ModuleLoader func(m *ModuleInfo, st *SymbolTable, tt *types.TypeTable) error

// LazyResolver handles on-demand loading of imported modules.
type LazyResolver struct {
	modules  map[uint32]*ModuleInfo // nameID -> module info
	symtable *SymbolTable
	types    *types.TypeTable
	loader   ModuleLoader
}

// NewLazyResolver creates a new LazyResolver.
func NewLazyResolver(st *SymbolTable, tt *types.TypeTable, loader ModuleLoader) *LazyResolver {
	if loader == nil {
		// Default no-op loader if none provided
		loader = func(m *ModuleInfo, st *SymbolTable, tt *types.TypeTable) error { return nil }
	}
	lr := &LazyResolver{
		modules:  make(map[uint32]*ModuleInfo),
		symtable: st,
		types:    tt,
		loader:   loader,
	}
	st.LazyResolver = lr
	return lr
}

// GetModules returns the map of registered ModuleInfo structures.
func (lr *LazyResolver) GetModules() map[uint32]*ModuleInfo {
	return lr.modules
}

// FindModuleOfSymbol returns the NameID of the module that exported the symbol.
func (lr *LazyResolver) FindModuleOfSymbol(symIdx uint32) uint32 {
	for nameID, mod := range lr.modules {
		for _, idx := range mod.Exports {
			if idx == symIdx {
				return nameID
			}
		}
	}
	return 0
}

// RegisterImport registers an imported module in the current scope without loading its contents.
func (lr *LazyResolver) RegisterImport(nameID uint32, filePath string, astRoot uint32, declNode uint32) (uint32, *diagnostics.Diagnostic) {
	// If the module is already defined in scope, reuse the symbol
	if symIdx, found := lr.symtable.Resolve(nameID); found {
		if lr.symtable.SymbolAt(symIdx).Kind == SymModule {
			return symIdx, nil
		}
	}

	// Define the module symbol in the current scope
	symIdx, diag := lr.symtable.Define(nameID, SymModule, 0, declNode)
	if diag != nil {
		return 0, diag
	}

	if _, ok := lr.modules[nameID]; !ok {
		lr.modules[nameID] = &ModuleInfo{
			NameID:   nameID,
			Status:   ModuleUnloaded,
			FilePath: filePath,
			Exports:  make(map[uint32]uint32),
			AstRoot:  astRoot,
		}
	}

	return symIdx, nil
}

// ResolveField resolves a field access like `moduleName.fieldName`.
// It triggers module loading if the module is currently unloaded.
func (lr *LazyResolver) ResolveField(moduleNameID uint32, fieldNameID uint32, pos diagnostics.Pos) (uint32, *diagnostics.Diagnostic) {
	mod, ok := lr.modules[moduleNameID]
	if !ok {
		return 0, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     2002, // undefined module
			Pos:      pos,
			Message:  "module not imported",
		}
	}

	// Cycle detection
	if mod.Status == ModuleLoading {
		moduleName := ""
		if lr.symtable != nil && lr.symtable.intern != nil {
			moduleName = lr.symtable.intern.Get(moduleNameID)
		}
		var loading []string
		if lr.symtable != nil && lr.symtable.intern != nil {
			for nameID, m := range lr.modules {
				if m.Status == ModuleLoading {
					loading = append(loading, lr.symtable.intern.Get(nameID))
				}
			}
		}
		debug.PrintStack()
		fmt.Printf("[CIRCULAR IMPORT TRACE] Active resolution stack: %v\n", loading)
		fmt.Printf("[ALL REGISTERED MODULES]:\n")
		for id, m := range lr.modules {
			fmt.Printf("  - %s: status=%v\n", lr.symtable.intern.Get(id), m.Status)
		}
		return 0, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     2003, // circular import
			Pos:      pos,
			Message:  fmt.Sprintf("circular import detected: module '%s' is already being resolved", moduleName),
		}
	}

	if mod.Status == ModuleUnloaded {
		mod.Status = ModuleLoading
		err := lr.loader(mod, lr.symtable, lr.types)
		if err != nil {
			mod.Status = ModuleUnloaded
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     2004, // module load error
				Pos:      pos,
				Message:  fmt.Sprintf("failed to load module: %v", err),
			}
		}
		mod.Status = ModuleLoaded
	}

	// Lookup field
	symIdx, found := mod.Exports[fieldNameID]
	if !found {
		fmt.Printf("[DEBUG RESOLVE FIELD FAILED] moduleNameID=%d fieldNameID=%d Name=%s Status=%d\n", moduleNameID, fieldNameID, lr.symtable.intern.Get(fieldNameID), mod.Status)
		fmt.Printf("Exports in module %s (ID=%d):\n", lr.symtable.intern.Get(moduleNameID), moduleNameID)
		for k, v := range mod.Exports {
			fmt.Printf("  k=%d Name=%s v=%d\n", k, lr.symtable.intern.Get(k), v)
		}
		return 0, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     2005, // undefined field
			Pos:      pos,
			Message:  "field not found in module",
		}
	}

	// Mark module symbol as used
	if modSymIdx, ok := lr.symtable.Resolve(moduleNameID); ok {
		lr.symtable.MarkUsed(modSymIdx)
	}
	lr.symtable.MarkUsed(symIdx)

	return symIdx, nil
}

// PreloadModule loads a module immediately on demand.
func (lr *LazyResolver) PreloadModule(nameID uint32) error {
	mod, ok := lr.modules[nameID]
	if !ok {
		return fmt.Errorf("module not registered")
	}
	if mod.Status == ModuleUnloaded {
		mod.Status = ModuleLoading
		err := lr.loader(mod, lr.symtable, lr.types)
		if err != nil {
			mod.Status = ModuleUnloaded
			return err
		}
		mod.Status = ModuleLoaded
	}
	return nil
}


// CheckUnusedImports returns diagnostics for all modules that were imported but never accessed.
func (lr *LazyResolver) CheckUnusedImports(intern *ast.InternPool) []diagnostics.Diagnostic {
	var diags []diagnostics.Diagnostic
	for _, mod := range lr.modules {
		if mod.Status == ModuleUnloaded {
			name := intern.Get(mod.NameID)
			
			// Find the symbol to get its DeclNode
			declNode := uint32(0)
			if symIdx, ok := lr.symtable.Resolve(mod.NameID); ok {
				sym := lr.symtable.SymbolAt(symIdx)
				declNode = sym.DeclNode
			}

			// In a real implementation we would map declNode to a Pos
			pos := diagnostics.Pos{} 
			_ = declNode // Use declNode here to suppress unused warning for this mock logic

			diags = append(diags, diagnostics.Diagnostic{
				Severity: diagnostics.SeverityWarning,
				Code:     2006, // unused import
				Pos:      pos,
				Message:  fmt.Sprintf("unused import: '%s'", name),
			})
		}
	}
	return diags
}
