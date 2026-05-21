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
	if text == "null" {
		return "NULL"
	}
	symIdx := node.Payload
	if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
		sym := g.Symbols.SymbolAt(symIdx)
		switch sym.Kind {
		case sema.SymConst:
			return MangleGlobalName("", text)
		case sema.SymVar:
			if sym.ScopeID == 0 {
				return MangleGlobalName("", text)
			}
		case sema.SymFunc:
			return GetFuncMangledName(symIdx, text, g.Table, g.Symbols, g.Intern)
		}
	}
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

	tLeft := g.NodeType(children[0])
	tRight := g.NodeType(children[1])
	if tLeft == types.TypeString || tRight == types.TypeString {
		if opText == "==" {
			return fmt.Sprintf("ax_str_eq(%s, %s)", left, right)
		} else if opText == "!=" {
			return fmt.Sprintf("(!ax_str_eq(%s, %s))", left, right)
		}
	}

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

	isExtern := false
	funcNode := g.Tree.Node(children[0])
	if funcNode.Kind == ast.NodeIdent {
		funcName := string(g.Tree.TokenText(funcNode.TokenIdx))
		symIdx := funcNode.Payload
		fmt.Printf("[DEBUG-CGEN-CALL] name=%s symIdx=%d\n", funcName, symIdx)
		if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
			sym := g.Symbols.SymbolAt(symIdx)
			fmt.Printf("[DEBUG-CGEN-CALL]   sym.Kind=%v sym.Flags=%d\n", sym.Kind, sym.Flags)
			if sym.Kind == sema.SymFunc && (sym.Flags&sema.SymFlagExtern != 0) {
				isExtern = true
				fmt.Printf("[DEBUG-CGEN-CALL]   SET isExtern=true!\n")
			}
		}
	}

	// Collect arguments first (needed for both builtin and normal paths)
	args := make([]string, 0, len(children)-1)
	for i := 1; i < len(children); i++ {
		argNodeIdx := children[i]
		argNode := g.Tree.Node(argNodeIdx)
		var val string
		var actualIdx uint32
		if argNode.Kind == ast.NodeNamedArg {
			actualIdx = argNode.FirstChild
			if actualIdx != ast.NullIdx {
				val = g.Emit(actualIdx)
			}
		} else {
			actualIdx = argNodeIdx
			val = g.Emit(actualIdx)
		}

		if isExtern && actualIdx != ast.NullIdx {
			argType := g.NodeType(actualIdx)
			fmt.Printf("[DEBUG-CGEN-CALL]   arg %d kind=%v type=%v\n", i, g.Tree.Node(actualIdx).Kind, argType)
			if argType == types.TypeString {
				val = fmt.Sprintf("(const char*)(%s).ptr", val)
				fmt.Printf("[DEBUG-CGEN-CALL]     wrapped to C string: %s\n", val)
			}
		}
		args = append(args, val)
	}

	// Check if the callee is a qualified/namespaced built-in call
	if qualifiedName, ok := g.getQualifiedFieldName(children[0]); ok {
		if call := EmitBuiltinCall(qualifiedName, args); call != "" {
			return call
		}
	}

	// First child is the function expression (usually an Ident)
	funcNode = g.Tree.Node(children[0])
	if funcNode.Kind == ast.NodeIdent {
		funcName := string(g.Tree.TokenText(funcNode.TokenIdx))

		// Check if it's actually a struct constructor call
		symIdx := funcNode.Payload
		if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
			sym := g.Symbols.SymbolAt(symIdx)
			if sym.Kind == sema.SymStruct {
				structType := types.TypeID(sym.TypeID)
				ctype := CTypeName(structType, g.Table, g.Intern, g.Queue)
				var fields []string
				for i := 1; i < len(children); i++ {
					childIdx := children[i]
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
		}

		// Check builtin table first
		if call := EmitBuiltinCall(funcName, args); call != "" {
			return call
		}

		// Not a builtin — apply standard mangling
		funcExpr := GetFuncMangledName(symIdx, funcName, g.Table, g.Symbols, g.Intern)
		// Adapt arguments!
		adaptedArgs := make([]string, len(args))
		for j, val := range args {
			actualIdx := children[j+1]
			isImplicitPtr := g.isImplicitPointerC(actualIdx)
			if g.expectsPointer(symIdx, j) {
				if !isImplicitPtr {
					adaptedArgs[j] = "&(" + val + ")"
				} else {
					adaptedArgs[j] = val
				}
			} else {
				if isImplicitPtr {
					adaptedArgs[j] = "*(" + val + ")"
				} else {
					adaptedArgs[j] = val
				}
			}
		}
		return fmt.Sprintf("%s(%s)", funcExpr, strings.Join(adaptedArgs, ", "))
	}

	// Non-identifier callee (e.g. field call, closure)
	if funcNode.Kind == ast.NodeFieldExpr {
		fieldChildren := g.Tree.Children(children[0])
		if len(fieldChildren) >= 1 {
			receiverIdx := fieldChildren[0]
			objType := g.NodeType(receiverIdx)
			if objType != types.TypeUnknown {
				entry := g.Table.Entry(objType)
				baseType := objType
				if entry.Kind == types.KindPointer {
					baseType = g.Table.PointerElem(objType)
					entry = g.Table.Entry(baseType)
				}
				if entry.Kind == types.KindStruct {
					// Method call!
					fieldNameText := string(g.Tree.TokenText(funcNode.TokenIdx))
					fieldNameID := g.Intern.Intern([]byte(fieldNameText))
					
					methodSymIdx, found := g.findMethodSymbol(baseType, fieldNameID)
					if found {
						receiverExpr := g.Emit(receiverIdx)
						
						var adaptedReceiver string
						// Check if parameter 0 (receiver self) expects a pointer!
						isRecImplicitPtr := g.isImplicitPointerC(receiverIdx)
						if g.expectsPointer(methodSymIdx, 0) {
							if !isRecImplicitPtr {
								adaptedReceiver = "&(" + receiverExpr + ")"
							} else {
								adaptedReceiver = receiverExpr
							}
						} else {
							if isRecImplicitPtr {
								adaptedReceiver = "*(" + receiverExpr + ")"
							} else {
								adaptedReceiver = receiverExpr
							}
						}
						
						// Add receiver as first argument
						callArgs := []string{adaptedReceiver}
						
						// Adapt other arguments!
						for j, val := range args {
							actualIdx := children[j+1]
							isImplicitPtr := g.isImplicitPointerC(actualIdx)
							if g.expectsPointer(methodSymIdx, j+1) {
								if !isImplicitPtr {
									callArgs = append(callArgs, "&("+val+")")
								} else {
									callArgs = append(callArgs, val)
								}
							} else {
								if isImplicitPtr {
									callArgs = append(callArgs, "*("+val+")")
								} else {
									callArgs = append(callArgs, val)
								}
							}
						}
						
						mangledName := GetFuncMangledName(methodSymIdx, fieldNameText, g.Table, g.Symbols, g.Intern)
						return fmt.Sprintf("%s(%s)", mangledName, strings.Join(callArgs, ", "))
					} else {
						// Fallback: if method symbol not found (e.g. mock test)
						structName := "anon_struct"
						if entry.NameID != 0 {
							structName = string(g.Intern.Get(entry.NameID))
						}
						receiverExpr := g.Emit(receiverIdx)
						callArgs := []string{receiverExpr}
						callArgs = append(callArgs, args...)
						
						mangledName := "ax_" + structName + "_" + fieldNameText
						return fmt.Sprintf("%s(%s)", mangledName, strings.Join(callArgs, ", "))
					}
				}
			}
		}
	}

	funcExpr := g.Emit(children[0])
	return fmt.Sprintf("%s(%s)", funcExpr, strings.Join(args, ", "))
}


// emitField emits a field access expression: obj.field or obj->field
func (g *ExprGen) emitField(idx uint32, node *ast.AstNode) string {
	children := g.Tree.Children(idx)
	if len(children) < 1 {
		return "/* invalid field expr */"
	}

	// Check if this is a module-level field access
	lhsIdx := children[0]
	lhsNode := g.Tree.Node(lhsIdx)
	if lhsNode.Payload != 0 && g.Symbols != nil && int(lhsNode.Payload) < len(g.Symbols.Symbols) {
		lhsSym := g.Symbols.SymbolAt(lhsNode.Payload)
		if lhsSym.Kind == sema.SymModule {
			moduleName := string(g.Intern.Get(lhsSym.NameID))
			fieldName := string(g.Tree.TokenText(node.TokenIdx))
			
			// Replace dots with underscores in module name
			mangledModule := strings.ReplaceAll(moduleName, ".", "_")
			return "ax_" + mangledModule + "_" + fieldName
		}
	}

	obj := g.Emit(children[0])
	fieldName := string(g.Tree.TokenText(node.TokenIdx))

	if g.isPointerInC(children[0]) {
		return fmt.Sprintf("%s->%s", obj, fieldName)
	}
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

	colType := g.NodeType(children[0])
	if colType != types.TypeUnknown {
		entry := g.Table.Entry(colType)
		if entry.Kind == types.KindPointer {
			return fmt.Sprintf("((%s)[%s])", arr, index)
		}
	}

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
	fmt.Printf("[DEBUG-CGEN] emitCast nodeIdx=%d payload=%d targetType=%d\n", idx, node.Payload, targetType)

	if targetType == types.TypeString {
		srcType := g.NodeType(children[0])
		if srcType != types.TypeUnknown {
			srcEntry := g.Table.Entry(srcType)
			if srcEntry.Kind == types.KindPointer {
				return fmt.Sprintf("((ax_string){.ptr = (const ax_u8*)(%s), .len = strlen((const char*)(%s))})", inner, inner)
			}
		}
	}

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
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n', 't', 'r', '\\', '"', '\'', '0':
				b.WriteByte('\\')
				b.WriteByte(s[i+1])
				i += 2
			case 'x':
				if i+4 <= len(s) {
					b.WriteString(s[i : i+4])
					i += 4
				} else {
					b.WriteByte('\\')
					b.WriteByte('x')
					i += 2
				}
			case 'u':
				if i+6 <= len(s) {
					b.WriteString(s[i : i+6])
					i += 6
				} else {
					b.WriteByte('\\')
					b.WriteByte('u')
					i += 2
				}
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i+1])
				i += 2
			}
		} else if s[i] == '"' {
			b.WriteString(`\"`)
			i++
		} else if s[i] == '\n' {
			b.WriteString(`\n`)
			i++
		} else if s[i] == '\r' {
			b.WriteString(`\r`)
			i++
		} else if s[i] == '\t' {
			b.WriteString(`\t`)
			i++
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

// NodeType returns the TypeID of a given AST node, recursively resolving it if necessary.
func (g *ExprGen) NodeType(nodeIdx uint32) types.TypeID {
	if nodeIdx == ast.NullIdx {
		return types.TypeUnknown
	}
	node := g.Tree.Node(nodeIdx)
	switch node.Kind {
	case ast.NodeStringLit:
		return types.TypeString
	case ast.NodeIntLit:
		return types.TypeI64
	case ast.NodeFloatLit:
		return types.TypeF64
	case ast.NodeBoolLit:
		return types.TypeBool
	case ast.NodeCharLit:
		return types.TypeChar8
	case ast.NodeIdent:
		symIdx := node.Payload
		if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
			sym := g.Symbols.SymbolAt(symIdx)
			return types.TypeID(sym.TypeID)
		}
	case ast.NodeCastExpr, ast.NodeDerefExpr, ast.NodeStructLit, ast.NodeArrayLit:
		return types.TypeID(node.Payload)
	case ast.NodeFieldExpr:
		children := g.Tree.Children(nodeIdx)
		if len(children) >= 1 {
			objType := g.NodeType(children[0])
			if objType != types.TypeUnknown {
				entry := g.Table.Entry(objType)
				if entry.Kind == types.KindPointer {
					objType = g.Table.PointerElem(objType)
					entry = g.Table.Entry(objType)
				}
				if entry.Kind == types.KindStruct {
					structInfo := g.Table.StructInfo(objType)
					fieldNameID := node.Payload
					for _, f := range structInfo.Fields {
						if f.NameID == fieldNameID {
							return f.TypeID
						}
					}
					fieldName := string(g.Tree.TokenText(node.TokenIdx))
					for _, f := range structInfo.Fields {
						if resolveName(f.NameID, g.Intern) == fieldName {
							return f.TypeID
						}
					}
				}
			}
		}
	case ast.NodeIndexExpr:
		children := g.Tree.Children(nodeIdx)
		if len(children) >= 1 {
			colType := g.NodeType(children[0])
			if colType != types.TypeUnknown {
				entry := g.Table.Entry(colType)
				if entry.Kind == types.KindPointer {
					return g.Table.PointerElem(colType)
				}
				if entry.Kind == types.KindSlice {
					return g.Table.SliceElem(colType)
				}
			}
		}
	case ast.NodeCallExpr:
		children := g.Tree.Children(nodeIdx)
		if len(children) >= 1 {
			calleeType := g.NodeType(children[0])
			if calleeType != types.TypeUnknown {
				entry := g.Table.Entry(calleeType)
				if entry.Kind == types.KindFunction {
					funcInfo := g.Table.FuncInfo(calleeType)
					return funcInfo.Return
				} else if entry.Kind == types.KindStruct {
					return calleeType
				}
			}
		}
	}
	return types.TypeUnknown
}

func (g *ExprGen) findMethodSymbol(structType types.TypeID, methodNameID uint32) (uint32, bool) {
	for idx, sym := range g.Symbols.Symbols {
		if sym.Kind == sema.SymFunc && sym.NameID == methodNameID {
			tID := types.TypeID(sym.TypeID)
			if tID != types.TypeUnknown {
				entry := g.Table.Entry(tID)
				if entry.Kind == types.KindFunction {
					fi := g.Table.FuncInfo(tID)
					if len(fi.Params) > 0 {
						firstParamType := fi.Params[0]
						if g.baseTypeEquals(firstParamType, structType) {
							return uint32(idx), true
						}
					}
				}
			}
		}
	}
	return 0, false
}

func (g *ExprGen) baseTypeEquals(t1, target types.TypeID) bool {
	entry := g.Table.Entry(t1)
	if entry.Kind == types.KindPointer {
		return g.Table.PointerElem(t1) == target
	}
	if entry.Kind == types.KindRef {
		return types.TypeID(entry.Extra) == target
	}
	return t1 == target
}

func (g *ExprGen) getQualifiedFieldName(nodeIdx uint32) (string, bool) {
	if nodeIdx == ast.NullIdx {
		return "", false
	}
	node := g.Tree.Node(nodeIdx)
	if node.Kind == ast.NodeIdent {
		return string(g.Tree.TokenText(node.TokenIdx)), true
	}
	if node.Kind == ast.NodeFieldExpr {
		children := g.Tree.Children(nodeIdx)
		if len(children) >= 1 {
			lhs, ok := g.getQualifiedFieldName(children[0])
			if ok {
				rhs := string(g.Tree.TokenText(node.TokenIdx))
				return lhs + "." + rhs, true
			}
		}
	}
	return "", false
}

func (g *ExprGen) expectsPointer(symIdx uint32, paramIdx int) bool {
	if symIdx == 0 || g.Symbols == nil || int(symIdx) >= len(g.Symbols.Symbols) {
		return false
	}
	sym := g.Symbols.SymbolAt(symIdx)
	if sym.DeclNode == 0 {
		return false
	}
	// Traverse the children of sym.DeclNode to find the parameter nodes.
	paramCount := 0
	child := g.Tree.Node(sym.DeclNode).FirstChild
	for child != ast.NullIdx {
		childNode := g.Tree.Node(child)
		if childNode.Kind == ast.NodeParamDecl {
			if paramCount == paramIdx {
				isLent := (childNode.Flags & ast.FlagIsLent) != 0
				isMut := (childNode.Flags & ast.FlagIsMut) != 0
				
				if isLent {
					return true
				}
				if isMut {
					// Check if type is a struct or generic struct
					if sym.TypeID != 0 {
						entry := g.Table.Entry(types.TypeID(sym.TypeID))
						if entry.Kind == types.KindFunction {
							fi := g.Table.FuncInfo(types.TypeID(sym.TypeID))
							if paramIdx < len(fi.Params) {
								pt := fi.Params[paramIdx]
								ptEntry := g.Table.Entry(pt)
								if ptEntry.Kind == types.KindStruct || ptEntry.Kind == types.KindGenericInst {
									return true
								}
							}
						}
					}
				}
				return false
			}
			paramCount++
		}
		child = childNode.NextSibling
	}
	return false
}

func (g *ExprGen) isPointerInC(nodeIdx uint32) bool {
	if nodeIdx == ast.NullIdx {
		return false
	}
	node := g.Tree.Node(nodeIdx)
	objType := g.NodeType(nodeIdx)
	if objType != types.TypeUnknown {
		entry := g.Table.Entry(objType)
		if entry.Kind == types.KindPointer || entry.Kind == types.KindRef {
			return true
		}
	}
	
	if node.Kind == ast.NodeIdent {
		symIdx := node.Payload
		if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
			sym := g.Symbols.SymbolAt(symIdx)
			if sym.Kind == sema.SymParam && sym.DeclNode != 0 {
				paramFlags := g.Tree.Node(sym.DeclNode).Flags
				isLent := (paramFlags & ast.FlagIsLent) != 0
				isMut := (paramFlags & ast.FlagIsMut) != 0
				if isLent {
					return true
				} else if isMut && objType != types.TypeUnknown {
					entry := g.Table.Entry(objType)
					if entry.Kind == types.KindStruct || entry.Kind == types.KindGenericInst {
						return true
					}
				}
			}
		}
	}
	return false
}

func (g *ExprGen) isImplicitPointerC(nodeIdx uint32) bool {
	if nodeIdx == ast.NullIdx {
		return false
	}
	node := g.Tree.Node(nodeIdx)
	objType := g.NodeType(nodeIdx)
	
	if node.Kind == ast.NodeIdent {
		symIdx := node.Payload
		if symIdx != 0 && g.Symbols != nil && int(symIdx) < len(g.Symbols.Symbols) {
			sym := g.Symbols.SymbolAt(symIdx)
			if sym.Kind == sema.SymParam && sym.DeclNode != 0 {
				paramFlags := g.Tree.Node(sym.DeclNode).Flags
				isLent := (paramFlags & ast.FlagIsLent) != 0
				isMut := (paramFlags & ast.FlagIsMut) != 0
				if isLent {
					return true
				} else if isMut && objType != types.TypeUnknown {
					entry := g.Table.Entry(objType)
					if entry.Kind == types.KindStruct || entry.Kind == types.KindGenericInst {
						return true
					}
				}
			}
		}
	}
	return false
}
