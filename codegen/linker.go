package codegen

// --------------------------------------------------------------------------
// p12-t04: Dynamic Linking Stub
//
// Infrastructure for dynamic linking support (.so/.dll/.dylib).
// Currently a structural stub for future implementation.
// --------------------------------------------------------------------------

// DynLinkInfo contains metadata for a dynamically linked symbol.
type DynLinkInfo struct {
	Name      string // mangled symbol name
	Library   string // library name (e.g., "libc.so.6")
	IsWeak    bool   // weak symbol (may be absent at runtime)
}

// --------------------------------------------------------------------------
// p12-t05: Incremental Linker Stub
//
// Infrastructure for incremental linking that relinks only changed modules.
// Currently a structural stub for future implementation.
// --------------------------------------------------------------------------

// IncrementalState tracks which object files need relinking.
type IncrementalState struct {
	ObjectFiles map[string]uint64 // file path → content hash
	OutputPath  string
}

// NeedsRelink returns true if the object file has changed since last link.
func (s *IncrementalState) NeedsRelink(path string, hash uint64) bool {
	prev, ok := s.ObjectFiles[path]
	return !ok || prev != hash
}

// --------------------------------------------------------------------------
// p12-t07: Symbol Demangling
//
// Human-readable display of mangled AXIOM symbols.
// --------------------------------------------------------------------------

// DemangleDisplay returns a human-readable representation of a mangled symbol.
// Example: "_AX_math_add_ii_i" → "math::add(i32, i32) -> i32"
func DemangleDisplay(mangled string) string {
	result, err := Demangle(mangled)
	if err != nil {
		return mangled // return as-is if not demangleable
	}

	display := result.Module + "::" + result.Name + "("
	for i, p := range result.Params {
		if i > 0 {
			display += ", "
		}
		display += typeDisplayName(p)
	}
	display += ") -> " + typeDisplayName(result.Ret)
	return display
}

// typeDisplayName returns the human-readable type name for a TypeID.
func typeDisplayName(typeID uint32) string {
	switch typeID {
	case 0:
		return "void"
	case 2:
		return "bool"
	case 3:
		return "i32"
	case 4:
		return "i64"
	case 5:
		return "f64"
	case 6:
		return "i8"
	case 7:
		return "i16"
	case 8:
		return "u8"
	case 9:
		return "u16"
	case 10:
		return "u32"
	case 11:
		return "u64"
	case 12:
		return "f32"
	case 13:
		return "str"
	case 14:
		return "ptr"
	default:
		return "?"
	}
}
