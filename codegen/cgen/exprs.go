package cgen

import (
	"fmt"
	"strings"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

// ExprGen generates C expression strings from typed AST expression nodes.
// It carries context about safe/unsafe mode to control safety check emission.
type ExprGen struct {
	Table   *types.TypeTable
	Intern  *ast.InternPool
	Symbols *sema.SymbolTable
	Tree    *ast.AstTree
	Queue   *TypeDeclQueue
	Unsafe  bool // when true, omits bounds checks and generational checks
}

// NewExprGen creates a new expression generator.
func NewExprGen(
	table *types.TypeTable,
	intern *ast.InternPool,
	symbols *sema.SymbolTable,
	tree *ast.AstTree,
	queue *TypeDeclQueue,
) *ExprGen {
	return &ExprGen{
		Table:   table,
		Intern:  intern,
		Symbols: symbols,
		Tree:    tree,
		Queue:   queue,
	}
}

// WithUnsafe returns a new ExprGen with unsafe mode enabled.
// The original ExprGen is not modified.
func (g *ExprGen) WithUnsafe() *ExprGen {
	clone := *g
	clone.Unsafe = true
	return &clone
}

// Emit returns the C expression string for the given AST expression node.
func (g *ExprGen) Emit(nodeIdx uint32) string {
	if nodeIdx == ast.NullIdx {
		return "/* null expr */"
	}
	node := g.Tree.Node(nodeIdx)

	switch node.Kind {
	case ast.NodeIntLit:
		return g.emitIntLit(nodeIdx, node)
	case ast.NodeFloatLit:
		return g.emitFloatLit(nodeIdx, node)
	case ast.NodeBoolLit:
		return g.emitBoolLit(nodeIdx, node)
	case ast.NodeStringLit:
		return g.emitStringLit(nodeIdx, node)
	case ast.NodeCharLit:
		return g.emitCharLit(nodeIdx, node)
	case ast.NodeNilLit:
		return "((void*)0)"
	case ast.NodeIdent:
		return g.emitIdent(nodeIdx, node)
	case ast.NodeBinaryExpr:
		return g.emitBinary(nodeIdx, node)
	case ast.NodeUnaryExpr:
		return g.emitUnary(nodeIdx, node)
	case ast.NodeCallExpr:
		return g.emitCall(nodeIdx, node)
	case ast.NodeFieldExpr:
		return g.emitField(nodeIdx, node)
	case ast.NodeIndexExpr:
		return g.emitIndex(nodeIdx, node)
	case ast.NodeCastExpr:
		return g.emitCast(nodeIdx, node)
	case ast.NodeDerefExpr:
		return g.emitDeref(nodeIdx, node)
	case ast.NodeStructLit:
		return g.emitStructLit(nodeIdx, node)
	case ast.NodeArrayLit:
		return g.emitArrayLit(nodeIdx, node)
	case ast.NodeSpawnExpr:
		return g.emitSpawn(nodeIdx, node)
	case ast.NodeAwaitExpr:
		return g.emitAwait(nodeIdx, node)
	case ast.NodeClosureExpr:
		return "/* closure: not yet supported */"
	default:
		return fmt.Sprintf("/* unknown expr kind %d */", node.Kind)
	}
}

// emitIntLit emits an integer literal.
func (g *ExprGen) emitIntLit(idx uint32, node *ast.AstNode) string {
	text := string(g.Tree.TokenText(node.TokenIdx))
	typeID := types.TypeID(node.Payload)

	// Annotate with type suffix for unsigned 64-bit
	switch typeID {
	case types.TypeU64:
		return fmt.Sprintf("((ax_u64)%sULL)", text)
	case types.TypeI64:
		return fmt.Sprintf("((ax_i64)%sLL)", text)
	case types.TypeU32:
		return fmt.Sprintf("((ax_u32)%sU)", text)
	default:
		return text
	}
}

// emitFloatLit emits a float literal.
func (g *ExprGen) emitFloatLit(idx uint32, node *ast.AstNode) string {
	text := string(g.Tree.TokenText(node.TokenIdx))
	typeID := types.TypeID(node.Payload)

	if typeID == types.TypeF32 {
		return text + "f"
	}
	return text
}

// emitBoolLit emits a boolean literal using the AX_TRUE/AX_FALSE macros.
func (g *ExprGen) emitBoolLit(idx uint32, node *ast.AstNode) string {
	text := string(g.Tree.TokenText(node.TokenIdx))
	if text == "true" {
		return "AX_TRUE"
	}
	return "AX_FALSE"
}

// emitStringLit emits a string literal as an ax_string compound literal.
func (g *ExprGen) emitStringLit(idx uint32, node *ast.AstNode) string {
	raw := string(g.Tree.TokenText(node.TokenIdx))
	// Strip surrounding quotes if present
	content := raw
	if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
		content = content[1 : len(content)-1]
	}
	// Compute byte length (content is already escaped in source form)
	byteLen := computeByteLen(content)
	escaped := escapeForC(content)
	return fmt.Sprintf(`(ax_string){.ptr=(const ax_u8*)"%s", .len=%d}`, escaped, byteLen)
}

// emitCharLit emits a character literal.
func (g *ExprGen) emitCharLit(idx uint32, node *ast.AstNode) string {
	text := string(g.Tree.TokenText(node.TokenIdx))
	return text
}

// emitIdent emits an identifier, mangled if it's a module-level symbol.
func (g *ExprGen) emitIdent(idx uint32, node *ast.AstNode) string {
	text := string(g.Tree.TokenText(node.TokenIdx))
	return text
}

// emitBinary emits a binary expression with proper operator mapping.
func (g *ExprGen) emitBinary(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		return "/* invalid binary expr */"
	}

	left := g.Emit(children[0])
	right := g.Emit(children[1])

	// The operator token index is stored in TokenIdx
	opText := string(g.Tree.TokenText(node.TokenIdx))
	cOp := mapBinaryOp(opText)

	// Power operator uses runtime helper
	if opText == "**" {
		return fmt.Sprintf("ax_pow(%s, %s)", left, right)
	}

	return fmt.Sprintf("(%s %s %s)", left, cOp, right)
}

// emitUnary emits a unary expression.
func (g *ExprGen) emitUnary(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 1 {
		return "/* invalid unary expr */"
	}

	operand := g.Emit(children[0])
	opText := string(g.Tree.TokenText(node.TokenIdx))

	switch opText {
	case "not":
		return fmt.Sprintf("(!%s)", operand)
	case "-":
		return fmt.Sprintf("(-%s)", operand)
	case "~":
		return fmt.Sprintf("(~%s)", operand)
	case "!":
		// Sink (ownership transfer) — in expression context, just pass through
		return operand
	default:
		return fmt.Sprintf("(%s%s)", opText, operand)
	}
}

// emitCall emits a function call expression.
// Checks the builtin table first; if the function is a recognized built-in,
// emits a direct call to the C runtime function.
func (g *ExprGen) emitCall(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 1 {
		return "/* invalid call expr */"
	}

	// Collect arguments first (needed for both builtin and normal paths)
	args := make([]string, 0, len(children)-1)
	for i := 1; i < len(children); i++ {
		argNode := g.Tree.Node(children[i])
		if argNode.Kind == ast.NodeNamedArg {
			if argNode.FirstChild != ast.NullIdx {
				args = append(args, g.Emit(argNode.FirstChild))
			}
		} else {
			args = append(args, g.Emit(children[i]))
		}
	}

	// First child is the function expression (usually an Ident)
	funcNode := g.Tree.Node(children[0])
	if funcNode.Kind == ast.NodeIdent {
		funcName := string(g.Tree.TokenText(funcNode.TokenIdx))

		// Check builtin table first
		if call := EmitBuiltinCall(funcName, args); call != "" {
			return call
		}

		// Not a builtin — apply standard mangling
		funcExpr := MangleFuncName("", funcName)
		return fmt.Sprintf("%s(%s)", funcExpr, strings.Join(args, ", "))
	}

	// Non-identifier callee (e.g. field call, closure)
	funcExpr := g.Emit(children[0])
	return fmt.Sprintf("%s(%s)", funcExpr, strings.Join(args, ", "))
}


// emitField emits a field access expression: obj.field
func (g *ExprGen) emitField(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 1 {
		return "/* invalid field expr */"
	}

	obj := g.Emit(children[0])
	fieldName := string(g.Tree.TokenText(node.TokenIdx))
	return fmt.Sprintf("%s.%s", obj, fieldName)
}

// emitIndex emits an array/slice index with bounds checking.
func (g *ExprGen) emitIndex(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 2 {
		return "/* invalid index expr */"
	}

	arr := g.Emit(children[0])
	index := g.Emit(children[1])

	if g.Unsafe {
		return fmt.Sprintf("(%s).ptr[%s]", arr, index)
	}

	return fmt.Sprintf("(ax_bounds_check((ax_u64)(%s), (%s).len), (%s).ptr[%s])",
		index, arr, arr, index)
}

// emitCast emits a type cast expression.
func (g *ExprGen) emitCast(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 1 {
		return "/* invalid cast expr */"
	}

	inner := g.Emit(children[0])
	targetType := types.TypeID(node.Payload)
	ctype := CTypeName(targetType, g.Table, g.Intern, g.Queue)
	return fmt.Sprintf("((%s)(%s))", ctype, inner)
}

// emitDeref emits a heap dereference with generational check.
func (g *ExprGen) emitDeref(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 1 {
		return "/* invalid deref expr */"
	}

	ref := g.Emit(children[0])
	targetType := types.TypeID(node.Payload)
	ctype := CTypeName(targetType, g.Table, g.Intern, g.Queue)

	if g.Unsafe {
		return fmt.Sprintf("(*((%s*)(%s).ptr))", ctype, ref)
	}
	return fmt.Sprintf("(*((%s*)ax_deref(%s)))", ctype, ref)
}

// emitStructLit emits a struct literal as a C compound literal.
func (g *ExprGen) emitStructLit(idx uint32, node *ast.AstNode) string {
	typeID := types.TypeID(node.Payload)
	ctype := CTypeName(typeID, g.Table, g.Intern, g.Queue)

	children := g.Tree.Children(idx)
	fields := make([]string, 0, len(children))
	for _, childIdx := range children {
		childNode := g.Tree.Node(childIdx)
		if childNode.Kind == ast.NodeNamedArg {
			fieldName := string(g.Tree.TokenText(childNode.TokenIdx))
			if childNode.FirstChild != ast.NullIdx {
				value := g.Emit(childNode.FirstChild)
				fields = append(fields, fmt.Sprintf(".%s=%s", fieldName, value))
			}
		}
	}

	return fmt.Sprintf("((%s){%s})", ctype, strings.Join(fields, ", "))
}

// emitArrayLit emits an array literal as a C compound literal slice.
func (g *ExprGen) emitArrayLit(idx uint32, node *ast.AstNode) string {
	typeID := types.TypeID(node.Payload)
	elemType := CTypeName(typeID, g.Table, g.Intern, g.Queue)

	children := g.Tree.Children(idx)
	elems := make([]string, 0, len(children))
	for _, childIdx := range children {
		elems = append(elems, g.Emit(childIdx))
	}

	count := len(elems)
	if count == 0 {
		sliceName := "ax_slice_" + sanitizeName(elemType)
		return fmt.Sprintf("((%s){.ptr=NULL, .len=0, .cap=0})", sliceName)
	}

	sliceName := "ax_slice_" + sanitizeName(elemType)
	return fmt.Sprintf("((%s){.ptr=(%s[]){%s}, .len=%d, .cap=%d})",
		sliceName, elemType, strings.Join(elems, ", "), count, count)
}

// emitSpawn: MVP — just emit a synchronous call.
func (g *ExprGen) emitSpawn(idx uint32, node *ast.AstNode) string {
	if node.FirstChild != ast.NullIdx {
		return g.Emit(node.FirstChild) + " /* spawn: MVP sync call */"
	}
	return "/* spawn: no expr */"
}

// emitAwait: MVP — await is identity (no-op).
func (g *ExprGen) emitAwait(idx uint32, node *ast.AstNode) string {
	if node.FirstChild != ast.NullIdx {
		return g.Emit(node.FirstChild) + " /* await: MVP no-op */"
	}
	return "/* await: no expr */"
}

// mapBinaryOp maps AXIOM binary operators to C operators.
func mapBinaryOp(axOp string) string {
	switch axOp {
	case "+":
		return "+"
	case "-":
		return "-"
	case "*":
		return "*"
	case "/":
		return "/"
	case "%":
		return "%"
	case "==":
		return "=="
	case "!=":
		return "!="
	case "<":
		return "<"
	case "<=":
		return "<="
	case ">":
		return ">"
	case ">=":
		return ">="
	case "and":
		return "&&"
	case "or":
		return "||"
	case "&":
		return "&"
	case "|":
		return "|"
	case "^":
		return "^"
	case "<<":
		return "<<"
	case ">>":
		return ">>"
	default:
		return axOp
	}
}

// computeByteLen counts the byte length of a string content,
// accounting for escape sequences.
func computeByteLen(content string) int {
	n := 0
	i := 0
	for i < len(content) {
		if content[i] == '\\' && i+1 < len(content) {
			switch content[i+1] {
			case 'n', 't', 'r', '\\', '"', '\'', '0':
				n++
				i += 2
			case 'x':
				// \xHH — 1 byte
				n++
				i += 4
			case 'u':
				// \uHHHH — up to 3 UTF-8 bytes
				n += 3
				i += 6
			default:
				n++
				i++
			}
		} else {
			n++
			i++
		}
	}
	return n
}

// escapeForC escapes a string for use inside a C string literal.
func escapeForC(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}
