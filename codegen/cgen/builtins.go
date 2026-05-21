package cgen

// --------------------------------------------------------------------------
// builtins.go — Built-in function recognition for C code generation.
//
// When the C backend encounters a call to a known built-in function,
// it emits a direct call to the corresponding C runtime function
// (defined in runtime/ax_stdlib.h) instead of the default mangled name.
//
// This bridges the gap between AXIOM source-level function names
// and the real C runtime implementations.
// --------------------------------------------------------------------------

import (
	"fmt"
	"strings"
)

// BuiltinKind classifies a built-in function by its dispatch behavior.
type BuiltinKind int

const (
	BuiltinNone   BuiltinKind = iota // Not a built-in
	BuiltinDirect                     // Direct 1:1 mapping to C function
	BuiltinTyped                      // Dispatched by argument type (e.g. println)
)

// BuiltinInfo describes a recognized built-in function.
type BuiltinInfo struct {
	Kind    BuiltinKind
	CName   string // C function name (for Direct)
	// For Typed builtins, CName is a prefix: ax_println_{str,i64,f64,bool}
}

// builtinTable maps AXIOM function names to their C runtime equivalents.
var builtinTable = map[string]BuiltinInfo{
	// ---------- Print / Format ----------
	"print":   {Kind: BuiltinTyped, CName: "ax_print"},
	"println": {Kind: BuiltinTyped, CName: "ax_println"},
	"eprint":  {Kind: BuiltinTyped, CName: "ax_eprint"},
	"eprintln": {Kind: BuiltinTyped, CName: "ax_eprintln"},

	// ---------- Assertions ----------
	"assert":    {Kind: BuiltinDirect, CName: "ax_assert_axiom"},
	"assert_eq": {Kind: BuiltinTyped, CName: "ax_assert_eq"},

	// ---------- String Operations ----------
	"str_len":         {Kind: BuiltinDirect, CName: "ax_str_len"},
	"str_char_count":  {Kind: BuiltinDirect, CName: "ax_str_char_count"},
	"str_contains":    {Kind: BuiltinDirect, CName: "ax_str_contains"},
	"str_starts_with": {Kind: BuiltinDirect, CName: "ax_str_starts_with"},
	"str_ends_with":   {Kind: BuiltinDirect, CName: "ax_str_ends_with"},
	"str_index_of":    {Kind: BuiltinDirect, CName: "ax_str_index_of"},
	"str_trim":        {Kind: BuiltinDirect, CName: "ax_str_trim"},
	"str_slice":       {Kind: BuiltinDirect, CName: "ax_str_slice"},
	"str_concat":      {Kind: BuiltinDirect, CName: "ax_str_concat"},
	"str_eq":          {Kind: BuiltinDirect, CName: "ax_str_eq"},
	"std.string.slice": {Kind: BuiltinDirect, CName: "ax_str_slice"},
	"std.string.len":   {Kind: BuiltinDirect, CName: "ax_str_len"},

	// ---------- Conversions ----------
	"to_str": {Kind: BuiltinTyped, CName: "ax_"},  // ax_{i64,f64,bool}_to_str

	// ---------- Math ----------
	"abs":   {Kind: BuiltinTyped, CName: "ax_abs"},
	"min":   {Kind: BuiltinTyped, CName: "ax_min"},
	"max":   {Kind: BuiltinTyped, CName: "ax_max"},
	"clamp": {Kind: BuiltinDirect, CName: "ax_clamp_i64"},
	"pow":   {Kind: BuiltinDirect, CName: "ax_pow"},
	"pow_i": {Kind: BuiltinDirect, CName: "ax_pow_i64"},
	"gcd":   {Kind: BuiltinDirect, CName: "ax_gcd"},
	"lcm":   {Kind: BuiltinDirect, CName: "ax_lcm"},
	"sqrt":  {Kind: BuiltinDirect, CName: "sqrt"},
	"sin":   {Kind: BuiltinDirect, CName: "sin"},
	"cos":   {Kind: BuiltinDirect, CName: "cos"},
	"tan":   {Kind: BuiltinDirect, CName: "tan"},
	"log":   {Kind: BuiltinDirect, CName: "log"},
	"exp":   {Kind: BuiltinDirect, CName: "exp"},
	"floor": {Kind: BuiltinDirect, CName: "floor"},
	"ceil":  {Kind: BuiltinDirect, CName: "ceil"},
	"round": {Kind: BuiltinDirect, CName: "round"},

	// ---------- Memory ----------
	"size_of":  {Kind: BuiltinDirect, CName: "sizeof"},
	"align_of": {Kind: BuiltinDirect, CName: "_Alignof"},
	"memcpy":   {Kind: BuiltinDirect, CName: "memcpy"},

	// ---------- Vec ----------
	"vec_new":    {Kind: BuiltinDirect, CName: "ax_vec_new"},
	"vec_push":   {Kind: BuiltinDirect, CName: "ax_vec_push"},
	"vec_pop":    {Kind: BuiltinDirect, CName: "ax_vec_pop"},
	"vec_get":    {Kind: BuiltinDirect, CName: "ax_vec_get"},
	"vec_set":    {Kind: BuiltinDirect, CName: "ax_vec_set"},
	"vec_len":    {Kind: BuiltinDirect, CName: "ax_vec_len"},
	"vec_clear":  {Kind: BuiltinDirect, CName: "ax_vec_clear"},
	"vec_free":   {Kind: BuiltinDirect, CName: "ax_vec_free"},

	// ---------- Arena ----------
	"arena_new":       {Kind: BuiltinDirect, CName: "ax_arena_new"},
	"arena_alloc":     {Kind: BuiltinDirect, CName: "ax_arena_alloc"},
	"arena_reset":     {Kind: BuiltinDirect, CName: "ax_arena_reset"},
	"arena_destroy":   {Kind: BuiltinDirect, CName: "ax_arena_destroy"},
	"arena_remaining": {Kind: BuiltinDirect, CName: "ax_arena_remaining"},
	"arena_used":      {Kind: BuiltinDirect, CName: "ax_arena_used"},

	// ---------- Process ----------
	"exit":  {Kind: BuiltinDirect, CName: "exit"},
	"abort": {Kind: BuiltinDirect, CName: "abort"},
	"panic": {Kind: BuiltinDirect, CName: "ax_panic"},
}

// LookupBuiltin checks if funcName is a recognized built-in.
func LookupBuiltin(funcName string) (BuiltinInfo, bool) {
	info, ok := builtinTable[funcName]
	return info, ok
}

// EmitBuiltinCall generates the C code for a built-in function call.
// Returns the call expression string, or empty string if not a built-in.
func EmitBuiltinCall(funcName string, args []string) string {
	info, ok := builtinTable[funcName]
	if !ok {
		return ""
	}

	switch info.Kind {
	case BuiltinDirect:
		return fmt.Sprintf("%s(%s)", info.CName, strings.Join(args, ", "))

	case BuiltinTyped:
		// For typed builtins, the codegen should append the type suffix.
		// Default to _str for now (type-dispatch will be improved
		// when the type system is wired through to codegen).
		suffix := "_str"
		return fmt.Sprintf("%s%s(%s)", info.CName, suffix, strings.Join(args, ", "))

	default:
		return ""
	}
}

// IsBuiltin returns true if funcName is a recognized built-in.
func IsBuiltin(funcName string) bool {
	_, ok := builtinTable[funcName]
	return ok
}
