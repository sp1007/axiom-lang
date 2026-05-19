// Package ast defines the AXIOM abstract syntax tree node types.
// AST nodes reference types only by TypeID (uint32), never by
// pointer to TypeInfo, to avoid circular imports with the types package.
package ast

import (
	"fmt"
	"unsafe"
)

// String returns a human-readable name for the node kind.
func (k NodeKind) String() string {
	if int(k) < len(nodeKindNames) {
		if s := nodeKindNames[k]; s != "" {
			return s
		}
	}
	return fmt.Sprintf("NodeKind(%d)", k)
}

// nodeKindNames maps each NodeKind to its display name.
var nodeKindNames = [NodeKindCount]string{
	NodeInvalid:       "Invalid",
	NodeProgram:       "Program",
	NodeFuncDecl:      "FuncDecl",
	NodeStructDecl:    "StructDecl",
	NodeInterfaceDecl: "InterfaceDecl",
	NodeImportDecl:    "ImportDecl",
	NodeConstDecl:     "ConstDecl",
	NodeTypeAliasDecl: "TypeAliasDecl",
	NodeParamDecl:     "ParamDecl",
	NodeFieldDecl:     "FieldDecl",
	NodeMethodSig:     "MethodSig",
	NodeVariantDecl:   "VariantDecl",
	NodeBlock:         "Block",
	NodeVarDecl:       "VarDecl",
	NodeAssignStmt:    "AssignStmt",
	NodeReturnStmt:    "ReturnStmt",
	NodeIfStmt:        "IfStmt",
	NodeElifClause:    "ElifClause",
	NodeElseClause:    "ElseClause",
	NodeForStmt:       "ForStmt",
	NodeWhileStmt:     "WhileStmt",
	NodeMatchStmt:     "MatchStmt",
	NodeMatchArm:      "MatchArm",
	NodeDeferStmt:     "DeferStmt",
	NodeUnsafeBlock:   "UnsafeBlock",
	NodeArenaBlock:    "ArenaBlock",
	NodeBinaryExpr:    "BinaryExpr",
	NodeUnaryExpr:     "UnaryExpr",
	NodeCallExpr:      "CallExpr",
	NodeIndexExpr:     "IndexExpr",
	NodeFieldExpr:     "FieldExpr",
	NodeCastExpr:      "CastExpr",
	NodeDerefExpr:     "DerefExpr",
	NodeSpawnExpr:     "SpawnExpr",
	NodeAwaitExpr:     "AwaitExpr",
	NodeClosureExpr:   "ClosureExpr",
	NodeIntLit:        "IntLit",
	NodeFloatLit:      "FloatLit",
	NodeStringLit:     "StringLit",
	NodeCharLit:       "CharLit",
	NodeBoolLit:       "BoolLit",
	NodeNilLit:        "NilLit",
	NodeIdent:         "Ident",
	NodeArrayLit:      "ArrayLit",
	NodeStructLit:     "StructLit",
	NodeNamedArg:      "NamedArg",
	NodeTypeExpr:      "TypeExpr",
	NodePtrType:       "PtrType",
	NodeSliceType:     "SliceType",
	NodeArrayType:     "ArrayType",
	NodeFuncType:      "FuncType",
	NodeGenericType:   "GenericType",
	NodeIsolatedType:  "IsolatedType",
	NodeFutureType:    "FutureType",
	NodeSumType:       "SumType",
	NodeWildcardPat:   "WildcardPat",
	NodeLiteralPat:    "LiteralPat",
	NodeBindingPat:    "BindingPat",
	NodeVariantPat:    "VariantPat",
	NodeTuplePat:      "TuplePat",
	NodeGenericParams: "GenericParams",
	NodeGenericParam:  "GenericParam",
	NodeEffectAnnotation: "EffectAnnotation",
	NodeError:         "Error",
	NodeDestroyStmt:   "DestroyStmt",
	NodeAliasStmt:     "AliasStmt",
}

// AstNode is a node in the flat-array AST.
// Layout is FROZEN at 24 bytes. Do not add fields without an RFC.
//
// Tree structure is encoded via index fields, not pointers.
// All nodes of a compilation unit live in a single []AstNode slice.
// Index 0 is always the root Program node.
//
// Field layout (24 bytes total):
//
//	Kind        NodeKind  1B   @ offset 0
//	_pad        uint8     1B   @ offset 1  (reserved)
//	Flags       uint16    2B   @ offset 2
//	TokenIdx    uint32    4B   @ offset 4
//	FirstChild  uint32    4B   @ offset 8
//	NextSibling uint32    4B   @ offset 12
//	Payload     uint32    4B   @ offset 16
//	ExtraIdx    uint32    4B   @ offset 20
//
// FROZEN: do not modify without RFC
type AstNode struct {
	Kind        NodeKind // discriminant
	_           uint8    // padding, reserved
	Flags       uint16   // bit flags, see Flag* constants below
	TokenIdx    uint32   // index into the token slice for this node's primary token
	FirstChild  uint32   // index of first child node (0 = no children)
	NextSibling uint32   // index of next sibling node (0 = last sibling)
	Payload     uint32   // multipurpose: SymbolIdx, TypeID, or literal value depending on Kind
	ExtraIdx    uint32   // index into AstTree.Extras for overflow data
}

// Compile-time size assertions — ensures AstNode is exactly 24 bytes.
var _ = [1]struct{}{}[24-unsafe.Sizeof(AstNode{})]
var _ = [1]struct{}{}[unsafe.Sizeof(AstNode{})-24]

// NodeKind is the type discriminant for AstNode.
type NodeKind uint8

const (
	NodeInvalid NodeKind = iota // 0: sentinel / error node

	// Top-level declarations
	NodeProgram       // root node
	NodeFuncDecl      // fn foo(...)
	NodeStructDecl    // struct Foo:
	NodeInterfaceDecl // interface Bar:
	NodeImportDecl    // import std.fs
	NodeConstDecl     // const X: T = expr
	NodeTypeAliasDecl // type Result = Ok(i32) | Err(string)

	// Sub-declarations
	NodeParamDecl   // function parameter
	NodeFieldDecl   // struct field
	NodeMethodSig   // interface method signature
	NodeVariantDecl // sum type variant

	// Statements
	NodeBlock       // indented block
	NodeVarDecl     // let x: T = expr
	NodeAssignStmt  // x = expr, x += expr
	NodeReturnStmt  // return expr
	NodeIfStmt      // if/elif/else chain
	NodeElifClause  // elif branch
	NodeElseClause  // else branch
	NodeForStmt     // for x in expr:
	NodeWhileStmt   // while cond:
	NodeMatchStmt   // match expr:
	NodeMatchArm    // pattern: body
	NodeDeferStmt   // defer expr
	NodeUnsafeBlock // unsafe:
	NodeArenaBlock  // in [arena]:

	// Expressions
	NodeBinaryExpr  // lhs op rhs
	NodeUnaryExpr   // op expr
	NodeCallExpr    // fn(args)
	NodeIndexExpr   // expr[idx]
	NodeFieldExpr   // expr.field
	NodeCastExpr    // expr as Type
	NodeDerefExpr   // expr.*
	NodeSpawnExpr   // spawn expr
	NodeAwaitExpr   // await expr
	NodeClosureExpr // |params| body

	// Literals and atoms
	NodeIntLit    // integer literal
	NodeFloatLit  // float literal
	NodeStringLit // string literal
	NodeCharLit   // character literal
	NodeBoolLit   // true / false
	NodeNilLit    // nil
	NodeIdent     // identifier reference
	NodeArrayLit  // [expr, ...]
	NodeStructLit // TypeName{field: expr, ...}
	NodeNamedArg  // field: expr in call/struct

	// Type expressions
	NodeTypeExpr     // generic type node wrapping a type expression
	NodePtrType      // *T or *mut T
	NodeSliceType    // [T]
	NodeArrayType    // [T; N]
	NodeFuncType     // fn(A, B) -> C
	NodeGenericType  // Foo[T]
	NodeIsolatedType // Isolated[T]
	NodeFutureType   // Future[T]
	NodeSumType      // A | B

	// Patterns
	NodeWildcardPat // _
	NodeLiteralPat  // literal in match arm
	NodeBindingPat  // name binding in match arm
	NodeVariantPat  // Variant(inner)
	NodeTuplePat    // (a, b)

	// Generics
	NodeGenericParams // [T: Interface]
	NodeGenericParam  // single T: Interface

	// Effects
	NodeEffectAnnotation // {.raises: [T].}

	// Error recovery
	NodeError // parse error node

	// Compiler-injected nodes
	NodeDestroyStmt // CTGC: compiler-injected destroy at scope exit
	NodeAliasStmt   // CTGC: alias reuse (destroy + alloc same type → reuse memory)

	NodeKindCount // sentinel — total count
)

// Flag constants for AstNode.Flags.
// Each flag is a distinct power of 2, suitable for bitwise OR.
const (
	FlagIsPub         uint16 = 1 << 0  // declaration is pub
	FlagIsMut         uint16 = 1 << 1  // variable is mut / pointer is *mut
	FlagIsAsync       uint16 = 1 << 2  // function is async
	FlagIsExtern      uint16 = 1 << 3  // function is extern
	FlagIsSink        uint16 = 1 << 4  // parameter is !T (sink/consumed)
	FlagIsLent        uint16 = 1 << 5  // parameter is lent (borrowed)
	FlagIsPacked      uint16 = 1 << 6  // struct is packed
	FlagEscapesToHeap uint16 = 1 << 7  // escape analysis: value escapes to heap
	FlagUsesArena     uint16 = 1 << 8  // allocation uses arena allocator
	FlagIsGeneric     uint16 = 1 << 9  // declaration has generic parameters
	FlagIsMoved       uint16 = 1 << 10 // value has been moved (ownership tracking)
)
