package codegen

import (
	"fmt"
	"strings"
)

// --------------------------------------------------------------------------
// p12-t01: Symbol Name Mangling
//
// Deterministic symbol mangling for AXIOM functions and types.
// Scheme: _AX_<module>_<name>_<typesig>
//
// Type encoding: i8=b, i16=s, i32=i, i64=l, u8=B, u16=S, u32=I, u64=L,
//                f32=f, f64=d, bool=o, str=t, void=v, ptr=p
// --------------------------------------------------------------------------

// typeCharMap maps TypeID to its mangling character.
var typeCharMap = map[uint32]byte{
	0:  'v', // void
	1:  'v', // void alias
	2:  'o', // bool
	3:  'i', // i32
	4:  'l', // i64
	5:  'd', // f64
	6:  'b', // i8
	7:  's', // i16
	8:  'B', // u8
	9:  'S', // u16
	10: 'I', // u32
	11: 'L', // u64
	12: 'f', // f32
	13: 't', // str
	14: 'p', // ptr
}

// charTypeMap maps mangling characters back to TypeID.
var charTypeMap map[byte]uint32

func init() {
	charTypeMap = make(map[byte]uint32, len(typeCharMap))
	for id, ch := range typeCharMap {
		if existing, exists := charTypeMap[ch]; !exists || id < existing {
			charTypeMap[ch] = id
		}
	}
}

// Mangle produces a mangled symbol name for a function.
// module: module path (e.g., "math")
// name: function name (e.g., "add")
// params: parameter TypeIDs
// ret: return TypeID
//
// Returns: "_AX_math_add_ii_i" for math::add(i32, i32) -> i32
func Mangle(module, name string, params []uint32, ret uint32) string {
	var b strings.Builder
	b.WriteString("_AX_")
	b.WriteString(sanitize(module))
	b.WriteByte('_')
	b.WriteString(sanitize(name))
	b.WriteByte('_')

	// Encode parameter types
	if len(params) == 0 {
		b.WriteByte('v') // void params → no args
	} else {
		for _, p := range params {
			b.WriteByte(encodeType(p))
		}
	}

	b.WriteByte('_')
	b.WriteByte(encodeType(ret))

	return b.String()
}

// MangleGeneric produces a mangled name for a generic monomorphization.
func MangleGeneric(module, name string, typeArgs []uint32, params []uint32, ret uint32) string {
	var b strings.Builder
	b.WriteString("_AX_")
	b.WriteString(sanitize(module))
	b.WriteByte('_')
	b.WriteString(sanitize(name))

	// Type arguments: _T<typechar>...
	if len(typeArgs) > 0 {
		b.WriteString("_T")
		for _, ta := range typeArgs {
			b.WriteByte(encodeType(ta))
		}
	}

	b.WriteByte('_')
	if len(params) == 0 {
		b.WriteByte('v')
	} else {
		for _, p := range params {
			b.WriteByte(encodeType(p))
		}
	}

	b.WriteByte('_')
	b.WriteByte(encodeType(ret))

	return b.String()
}

// MangleMethod produces a mangled name for a method.
func MangleMethod(module, typeName, method string, params []uint32, ret uint32) string {
	var b strings.Builder
	b.WriteString("_AX_")
	b.WriteString(sanitize(module))
	b.WriteByte('_')
	b.WriteString(sanitize(typeName))
	b.WriteByte('_')
	b.WriteString(sanitize(method))
	b.WriteByte('_')

	if len(params) == 0 {
		b.WriteByte('v')
	} else {
		for _, p := range params {
			b.WriteByte(encodeType(p))
		}
	}

	b.WriteByte('_')
	b.WriteByte(encodeType(ret))

	return b.String()
}

// MangleResult holds the demangled components of a mangled symbol.
type MangleResult struct {
	Module string
	Name   string
	Params []uint32
	Ret    uint32
}

// Demangle parses a mangled symbol name back into its components.
func Demangle(mangled string) (MangleResult, error) {
	if !strings.HasPrefix(mangled, "_AX_") {
		return MangleResult{}, fmt.Errorf("not an AXIOM mangled name: %q", mangled)
	}

	rest := mangled[4:] // strip "_AX_"
	parts := strings.Split(rest, "_")

	if len(parts) < 4 {
		return MangleResult{}, fmt.Errorf("malformed mangled name: %q (need at least 4 parts)", mangled)
	}

	result := MangleResult{
		Module: parts[0],
		Name:   parts[1],
	}

	// Parse parameters (second-to-last part)
	paramStr := parts[len(parts)-2]
	if paramStr != "v" {
		for _, ch := range []byte(paramStr) {
			if tid, ok := charTypeMap[ch]; ok {
				result.Params = append(result.Params, tid)
			}
		}
	}

	// Parse return type (last part)
	retStr := parts[len(parts)-1]
	if len(retStr) > 0 {
		if tid, ok := charTypeMap[retStr[0]]; ok {
			result.Ret = tid
		}
	}

	return result, nil
}

// IsMangled returns true if the symbol name is a mangled AXIOM name.
func IsMangled(name string) bool {
	return strings.HasPrefix(name, "_AX_")
}

// encodeType returns the single-char encoding for a TypeID.
func encodeType(typeID uint32) byte {
	if ch, ok := typeCharMap[typeID]; ok {
		return ch
	}
	return 'x' // unknown type
}

// sanitize replaces characters that are invalid in linker symbols.
func sanitize(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
