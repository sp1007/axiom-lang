package sema

// ModuleStatus indicates the lazy loading state of a module.
type ModuleStatus uint8

const (
	ModuleUnloaded ModuleStatus = iota // import seen, not yet accessed
	ModuleLoading                      // currently resolving (cycle detection)
	ModuleLoaded                       // all accessed fields resolved
)

// ModuleInfo tracks the resolution state of an imported module.
type ModuleInfo struct {
	NameID   uint32            // interned module name
	Status   ModuleStatus      // current load status
	FilePath string            // source file path (for multi-file projects)
	Exports  map[uint32]uint32 // nameID -> symbolIdx of exported symbols
	AstRoot  uint32            // root AST node index of the module's file
}
