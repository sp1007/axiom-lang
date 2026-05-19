package sema

import (
	"fmt"
	"math"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/types"
)

// ComptimeValue represents a value evaluated at compile time.
type ComptimeValue struct {
	Kind     types.TypeID
	IntVal   int64
	FloatVal float64
	StrVal   string
	BoolVal  bool
}

// String returns a human-readable representation of the compile-time value.
func (v ComptimeValue) String() string {
	switch v.Kind {
	case types.TypeI8, types.TypeI16, types.TypeI32, types.TypeI64, types.TypeISize:
		return fmt.Sprintf("%d", v.IntVal)
	case types.TypeU8, types.TypeU16, types.TypeU32, types.TypeU64, types.TypeUSize:
		return fmt.Sprintf("%d", v.IntVal)
	case types.TypeF32, types.TypeF64:
		return fmt.Sprintf("%g", v.FloatVal)
	case types.TypeBool:
		if v.BoolVal {
			return "true"
		}
		return "false"
	case types.TypeString:
		return fmt.Sprintf("%q", v.StrVal)
	default:
		return fmt.Sprintf("comptime(%d)", v.Kind)
	}
}

// ComptimeEvaluator evaluates constant expressions at compile time.
// This is the MVP stub for `#run` and `const` initializers.
type ComptimeEvaluator struct {
	ast      *ast.AstTree
	intern   *ast.InternPool
	symtable *SymbolTable
	types    *types.TypeTable
	// consts maps symbol indices to their evaluated compile-time values.
	consts map[uint32]ComptimeValue
	// evaluating tracks symbols currently being evaluated to detect cycles.
	evaluating map[uint32]bool
	errors     []diagnostics.Diagnostic
}

// NewComptimeEvaluator creates a new compile-time constant evaluator.
func NewComptimeEvaluator(tree *ast.AstTree, intern *ast.InternPool, st *SymbolTable, tt *types.TypeTable) *ComptimeEvaluator {
	return &ComptimeEvaluator{
		ast:        tree,
		intern:     intern,
		symtable:   st,
		types:      tt,
		consts:     make(map[uint32]ComptimeValue),
		evaluating: make(map[uint32]bool),
	}
}

// Errors returns collected diagnostics.
func (ce *ComptimeEvaluator) Errors() []diagnostics.Diagnostic {
	return ce.errors
}

// Consts returns the evaluated constant values.
func (ce *ComptimeEvaluator) Consts() map[uint32]ComptimeValue {
	return ce.consts
}

func (ce *ComptimeEvaluator) errorf(nodeIdx uint32, code int, format string, args ...any) {
	ce.errors = append(ce.errors, diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     uint32(code),
		Message:  fmt.Sprintf(format, args...),
		Pos:      diagnostics.Pos{},
	})
}

// EvalConsts walks the AST and evaluates all const declarations.
func (ce *ComptimeEvaluator) EvalConsts() []diagnostics.Diagnostic {
	if ce.ast == nil || ce.ast.NodeCount() == 0 {
		return ce.errors
	}
	ce.evalNode(0)
	return ce.errors
}

func (ce *ComptimeEvaluator) evalNode(nodeIdx uint32) {
	node := &ce.ast.Nodes[nodeIdx]

	if node.Kind == ast.NodeConstDecl {
		ce.evalConstDecl(nodeIdx)
	}

	child := node.FirstChild
	for child != 0 {
		ce.evalNode(child)
		child = ce.ast.Nodes[child].NextSibling
	}
}

func (ce *ComptimeEvaluator) evalConstDecl(nodeIdx uint32) {
	node := &ce.ast.Nodes[nodeIdx]
	symIdx := node.Payload

	// Find the initializer expression (skip type child if present)
	var initNode uint32
	child := node.FirstChild
	for child != 0 {
		childNode := &ce.ast.Nodes[child]
		if childNode.Kind != ast.NodeTypeExpr && childNode.Kind != ast.NodeGenericParams {
			initNode = child
			break
		}
		child = childNode.NextSibling
	}

	if initNode == 0 {
		// No initializer — can't evaluate at compile time
		ce.errorf(nodeIdx, 1500, "const declaration requires an initializer")
		return
	}

	// Check for cycles
	if ce.evaluating[symIdx] {
		ce.errorf(nodeIdx, 1501, "cyclic const reference detected")
		return
	}
	ce.evaluating[symIdx] = true

	val, err := ce.Eval(initNode)
	delete(ce.evaluating, symIdx)

	if err != nil {
		ce.errors = append(ce.errors, *err)
		return
	}

	ce.consts[symIdx] = val
}

// Eval evaluates a single expression node at compile time.
// Returns the value or a diagnostic if evaluation is not possible.
func (ce *ComptimeEvaluator) Eval(nodeIdx uint32) (ComptimeValue, *diagnostics.Diagnostic) {
	node := &ce.ast.Nodes[nodeIdx]

	switch node.Kind {
	case ast.NodeIntLit:
		return ce.evalIntLit(nodeIdx)
	case ast.NodeFloatLit:
		return ce.evalFloatLit(nodeIdx)
	case ast.NodeStringLit:
		return ce.evalStringLit(nodeIdx)
	case ast.NodeBoolLit:
		return ce.evalBoolLit(nodeIdx)
	case ast.NodeIdent:
		return ce.evalConstRef(nodeIdx)
	case ast.NodeCallExpr:
		// Binary operators are represented as call expressions in some parsers,
		// but in AXIOM the stub parser doesn't produce binary ops yet.
		// For now, reject function calls in comptime context.
		d := &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  "cannot evaluate function call at compile time",
			Pos:      diagnostics.Pos{},
		}
		return ComptimeValue{}, d
	default:
		d := &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  fmt.Sprintf("cannot evaluate expression of kind %d at compile time", node.Kind),
			Pos:      diagnostics.Pos{},
		}
		return ComptimeValue{}, d
	}
}

func (ce *ComptimeEvaluator) evalIntLit(nodeIdx uint32) (ComptimeValue, *diagnostics.Diagnostic) {
	text := ce.tokenText(nodeIdx)
	val := parseInt64(text)
	return ComptimeValue{Kind: types.TypeI64, IntVal: val}, nil
}

func (ce *ComptimeEvaluator) evalFloatLit(nodeIdx uint32) (ComptimeValue, *diagnostics.Diagnostic) {
	text := ce.tokenText(nodeIdx)
	val := parseFloat64(text)
	return ComptimeValue{Kind: types.TypeF64, FloatVal: val}, nil
}

func (ce *ComptimeEvaluator) evalStringLit(nodeIdx uint32) (ComptimeValue, *diagnostics.Diagnostic) {
	text := ce.tokenText(nodeIdx)
	// Strip quotes
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}
	return ComptimeValue{Kind: types.TypeString, StrVal: text}, nil
}

func (ce *ComptimeEvaluator) evalBoolLit(nodeIdx uint32) (ComptimeValue, *diagnostics.Diagnostic) {
	text := ce.tokenText(nodeIdx)
	return ComptimeValue{Kind: types.TypeBool, BoolVal: text == "true"}, nil
}

func (ce *ComptimeEvaluator) evalConstRef(nodeIdx uint32) (ComptimeValue, *diagnostics.Diagnostic) {
	node := &ce.ast.Nodes[nodeIdx]
	symIdx := node.Payload
	if symIdx == 0 || int(symIdx) >= len(ce.symtable.Symbols) {
		d := &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  "cannot evaluate unresolved reference at compile time",
			Pos:      diagnostics.Pos{},
		}
		return ComptimeValue{}, d
	}

	sym := ce.symtable.SymbolAt(symIdx)
	if sym.Kind != SymConst {
		d := &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  fmt.Sprintf("'%s' is not a compile-time constant", string(ce.intern.Get(sym.NameID))),
			Pos:      diagnostics.Pos{},
		}
		return ComptimeValue{}, d
	}

	// Check if already evaluated
	if val, ok := ce.consts[symIdx]; ok {
		return val, nil
	}

	// Need to evaluate the const's declaration
	if ce.evaluating[symIdx] {
		d := &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1501,
			Message:  "cyclic const reference detected",
			Pos:      diagnostics.Pos{},
		}
		return ComptimeValue{}, d
	}

	ce.evaluating[symIdx] = true
	declNode := sym.DeclNode
	ce.evalConstDecl(declNode)
	delete(ce.evaluating, symIdx)

	if val, ok := ce.consts[symIdx]; ok {
		return val, nil
	}

	d := &diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     1500,
		Message:  "failed to evaluate const reference",
		Pos:      diagnostics.Pos{},
	}
	return ComptimeValue{}, d
}

// IntArith performs integer arithmetic with overflow checking.
func IntArith(op string, a, b int64) (int64, *diagnostics.Diagnostic) {
	switch op {
	case "+":
		result := a + b
		if (b > 0 && result < a) || (b < 0 && result > a) {
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     1502,
				Message:  fmt.Sprintf("integer overflow: %d + %d", a, b),
			}
		}
		return result, nil
	case "-":
		result := a - b
		if (b > 0 && result > a) || (b < 0 && result < a) {
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     1502,
				Message:  fmt.Sprintf("integer overflow: %d - %d", a, b),
			}
		}
		return result, nil
	case "*":
		if a != 0 && b != 0 {
			result := a * b
			if result/a != b {
				return 0, &diagnostics.Diagnostic{
					Severity: diagnostics.SeverityError,
					Code:     1502,
					Message:  fmt.Sprintf("integer overflow: %d * %d", a, b),
				}
			}
			return result, nil
		}
		return 0, nil
	case "/":
		if b == 0 {
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     1503,
				Message:  "division by zero",
			}
		}
		if a == math.MinInt64 && b == -1 {
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     1502,
				Message:  fmt.Sprintf("integer overflow: %d / %d", a, b),
			}
		}
		return a / b, nil
	case "%":
		if b == 0 {
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     1503,
				Message:  "division by zero in modulo",
			}
		}
		return a % b, nil
	default:
		return 0, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  fmt.Sprintf("unsupported integer operation: %s", op),
		}
	}
}

// FloatArith performs floating-point arithmetic.
func FloatArith(op string, a, b float64) (float64, *diagnostics.Diagnostic) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, &diagnostics.Diagnostic{
				Severity: diagnostics.SeverityError,
				Code:     1503,
				Message:  "floating-point division by zero",
			}
		}
		return a / b, nil
	default:
		return 0, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  fmt.Sprintf("unsupported float operation: %s", op),
		}
	}
}

// BoolLogic performs boolean logic operations.
func BoolLogic(op string, a, b bool) (bool, *diagnostics.Diagnostic) {
	switch op {
	case "and":
		return a && b, nil
	case "or":
		return a || b, nil
	case "not":
		return !a, nil
	default:
		return false, &diagnostics.Diagnostic{
			Severity: diagnostics.SeverityError,
			Code:     1500,
			Message:  fmt.Sprintf("unsupported boolean operation: %s", op),
		}
	}
}

// StringConcat concatenates two compile-time strings.
func StringConcat(a, b string) string {
	return a + b
}

// tokenText returns the source text of the token associated with a node.
func (ce *ComptimeEvaluator) tokenText(nodeIdx uint32) string {
	node := &ce.ast.Nodes[nodeIdx]
	tokIdx := node.TokenIdx
	if int(tokIdx) >= len(ce.ast.Tokens) {
		return ""
	}
	tok := ce.ast.Tokens[tokIdx]
	end := tok.Offset + uint32(tok.Len)
	if end > uint32(len(ce.ast.Source)) {
		return ""
	}
	return string(ce.ast.Source[tok.Offset:end])
}

// parseInt64 parses an integer literal string to int64.
// Supports decimal, hex (0x), binary (0b), and octal (0o) prefixes.
func parseInt64(s string) int64 {
	if len(s) == 0 {
		return 0
	}

	// Remove underscores (numeric separators)
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '_' {
			clean = append(clean, s[i])
		}
	}
	s = string(clean)

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	var val int64
	if len(s) > 2 {
		switch s[:2] {
		case "0x", "0X":
			for _, c := range s[2:] {
				val *= 16
				switch {
				case c >= '0' && c <= '9':
					val += int64(c - '0')
				case c >= 'a' && c <= 'f':
					val += int64(c-'a') + 10
				case c >= 'A' && c <= 'F':
					val += int64(c-'A') + 10
				}
			}
			if negative {
				val = -val
			}
			return val
		case "0b", "0B":
			for _, c := range s[2:] {
				val = val*2 + int64(c-'0')
			}
			if negative {
				val = -val
			}
			return val
		case "0o", "0O":
			for _, c := range s[2:] {
				val = val*8 + int64(c-'0')
			}
			if negative {
				val = -val
			}
			return val
		}
	}

	for _, c := range s {
		val = val*10 + int64(c-'0')
	}
	if negative {
		val = -val
	}
	return val
}

// parseFloat64 parses a float literal string to float64.
func parseFloat64(s string) float64 {
	// Remove underscores
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '_' {
			clean = append(clean, s[i])
		}
	}

	// Simple manual parse: integer.fraction[e[+-]exponent]
	var val float64
	negative := false
	i := 0
	if i < len(clean) && clean[i] == '-' {
		negative = true
		i++
	}
	for i < len(clean) && clean[i] >= '0' && clean[i] <= '9' {
		val = val*10 + float64(clean[i]-'0')
		i++
	}
	if i < len(clean) && clean[i] == '.' {
		i++
		frac := 0.1
		for i < len(clean) && clean[i] >= '0' && clean[i] <= '9' {
			val += float64(clean[i]-'0') * frac
			frac *= 0.1
			i++
		}
	}
	if i < len(clean) && (clean[i] == 'e' || clean[i] == 'E') {
		i++
		expNeg := false
		if i < len(clean) && clean[i] == '-' {
			expNeg = true
			i++
		} else if i < len(clean) && clean[i] == '+' {
			i++
		}
		var exp float64
		for i < len(clean) && clean[i] >= '0' && clean[i] <= '9' {
			exp = exp*10 + float64(clean[i]-'0')
			i++
		}
		if expNeg {
			val /= math.Pow(10, exp)
		} else {
			val *= math.Pow(10, exp)
		}
	}
	if negative {
		val = -val
	}
	return val
}
