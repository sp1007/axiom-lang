package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestMonoBasic(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// Register primitive types for test
	i32ID := pool.Intern([]byte("i32"))
	symI32, _ := st.Define(i32ID, sema.SymBuiltinType, 0, 0)
	symI32Obj := st.SymbolAt(symI32)
	symI32Obj.TypeID = uint32(types.TypeI32)

	// Create generic identity function: fn identity[T](x: T) -> T { x }
	identityID := pool.Intern([]byte("identity"))
	tID := pool.Intern([]byte("T"))
	xID := pool.Intern([]byte("x"))

	fnDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeFuncDecl, Payload: identityID})
	tree.SetFlags(fnDecl, ast.FlagIsGeneric)
	addChild(tree, 0, fnDecl)

	gpParams := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParams})
	addChild(tree, fnDecl, gpParams)

	gpParam := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParam, Payload: tID})
	addChild(tree, gpParams, gpParam)

	// param: x: T
	paramDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeParamDecl, Payload: xID})
	addChild(tree, fnDecl, paramDecl)
	paramType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, paramDecl, paramType)
	paramIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, paramType, paramIdent)

	// Return type: T
	retType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, fnDecl, retType)
	retIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, retType, retIdent)

	// body: x
	body := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeBlock})
	addChild(tree, fnDecl, body)
	
	refX := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: xID})
	addChild(tree, body, refX)

	// First pass: resolve names and register template
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors resolving template: %v", diags)
	}

	symIdx := tree.Nodes[fnDecl].Payload
	_, found := tt.FindGenericTemplate(symIdx)
	if !found {
		t.Fatalf("expected to find generic template for identity")
	}

	// Now instantiate identity[i32]
	mono := sema.NewMonomorphizer(tree, pool, st, tt)
	instSymID, instDiags := mono.InstantiateFunction(symIdx, []types.TypeID{types.TypeI32})
	if len(instDiags) > 0 {
		t.Fatalf("unexpected errors during instantiation: %v", instDiags)
	}

	instSym := st.SymbolAt(instSymID)
	instName := string(pool.Get(instSym.NameID))
	
	if instName != "_AX_std_identity__i32" {
		t.Errorf("expected mangled name _AX_std_identity__i32, got %s", instName)
	}

	// Second instantiation should return cached ID
	instSymID2, _ := mono.InstantiateFunction(symIdx, []types.TypeID{types.TypeI32})
	if instSymID2 != instSymID {
		t.Errorf("expected cached symbol ID %d, got %d", instSymID, instSymID2)
	}
}

func TestMonoStruct(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// Register primitive types for test
	i32ID := pool.Intern([]byte("i32"))
	symI32, _ := st.Define(i32ID, sema.SymBuiltinType, 0, 0)
	st.SymbolAt(symI32).TypeID = uint32(types.TypeI32)

	strID := pool.Intern([]byte("string"))
	symStr, _ := st.Define(strID, sema.SymBuiltinType, 0, 0)
	st.SymbolAt(symStr).TypeID = uint32(types.TypeString)

	// Create generic struct: struct Stack[T] { value: T }
	stackID := pool.Intern([]byte("Stack"))
	tID := pool.Intern([]byte("T"))
	valueID := pool.Intern([]byte("value"))

	structDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeStructDecl, Payload: stackID})
	tree.SetFlags(structDecl, ast.FlagIsGeneric)
	addChild(tree, 0, structDecl)

	gpParams := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParams})
	addChild(tree, structDecl, gpParams)

	gpParam := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParam, Payload: tID})
	addChild(tree, gpParams, gpParam)

	// Field: value: T
	fieldDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeFieldDecl, Payload: valueID})
	addChild(tree, structDecl, fieldDecl)

	fieldType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, fieldDecl, fieldType)

	fieldIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, fieldType, fieldIdent)

	// 1. Resolve names
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors resolving struct: %v", diags)
	}

	symIdx := tree.Nodes[structDecl].Payload

	// 2. Instantiate Stack[i32]
	mono := sema.NewMonomorphizer(tree, pool, st, tt)
	instSymID1, diags := mono.InstantiateFunction(symIdx, []types.TypeID{types.TypeI32})
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	// 3. Instantiate Stack[string]
	instSymID2, diags := mono.InstantiateFunction(symIdx, []types.TypeID{types.TypeString})
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	sym1 := st.SymbolAt(instSymID1)
	sym2 := st.SymbolAt(instSymID2)

	if sym1.TypeID == 0 || sym2.TypeID == 0 {
		t.Errorf("expected instantiated structs to be assigned TypeIDs")
	}

	if sym1.TypeID == sym2.TypeID {
		t.Errorf("expected different TypeIDs for Stack[i32] and Stack[string]")
	}
}

func TestMonoRecursive(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// Register primitive types for test
	i32ID := pool.Intern([]byte("i32"))
	symI32, _ := st.Define(i32ID, sema.SymBuiltinType, 0, 0)
	st.SymbolAt(symI32).TypeID = uint32(types.TypeI32)

	// Create generic recursive function: fn fib[T](n: T) -> T { fib[T](n) }
	fibID := pool.Intern([]byte("fib"))
	tID := pool.Intern([]byte("T"))
	nID := pool.Intern([]byte("n"))

	fnDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeFuncDecl, Payload: fibID})
	tree.SetFlags(fnDecl, ast.FlagIsGeneric)
	addChild(tree, 0, fnDecl)

	gpParams := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParams})
	addChild(tree, fnDecl, gpParams)

	gpParam := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParam, Payload: tID})
	addChild(tree, gpParams, gpParam)

	// param: n: T
	paramDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeParamDecl, Payload: nID})
	addChild(tree, fnDecl, paramDecl)
	paramType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, paramDecl, paramType)
	paramIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, paramType, paramIdent)

	// Return type: T
	retType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, fnDecl, retType)
	retIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, retType, retIdent)

	// body: fib[T](n)
	body := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeBlock})
	addChild(tree, fnDecl, body)

	callExpr := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeCallExpr})
	addChild(tree, body, callExpr)

	// callee: fib[T] (IndexExpr)
	idxExpr := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIndexExpr})
	addChild(tree, callExpr, idxExpr)

	calleeIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: fibID})
	addChild(tree, idxExpr, calleeIdent)

	typeArg := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, idxExpr, typeArg)

	// arg: n
	argN := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: nID})
	addChild(tree, callExpr, argN)

	// 1. Resolve names
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors resolving func: %v", diags)
	}

	symIdx := tree.Nodes[fnDecl].Payload

	// 2. Instantiate fib[i32]
	mono := sema.NewMonomorphizer(tree, pool, st, tt)
	instSymID, diags := mono.InstantiateFunction(symIdx, []types.TypeID{types.TypeI32})
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	sym := st.SymbolAt(instSymID)
	instName := string(pool.Get(sym.NameID))
	if instName != "_AX_std_fib__i32" {
		t.Errorf("expected mangled name _AX_std_fib__i32, got %s", instName)
	}
}
