package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func setupResolverTest() (*ast.AstTree, *ast.InternPool, *sema.SymbolTable, *types.TypeTable, *sema.LazyResolver) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	tree := ast.NewTree(nil, nil)
	// Create root program node (index 0)
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeProgram})
	lr := sema.NewLazyResolver(st, tt, nil)
	return tree, pool, st, tt, lr
}

func addChild(tree *ast.AstTree, parent uint32, child uint32) {
	if tree.Nodes[parent].FirstChild == 0 {
		tree.Nodes[parent].FirstChild = child
	} else {
		curr := tree.Nodes[parent].FirstChild
		for tree.Nodes[curr].NextSibling != 0 {
			curr = tree.Nodes[curr].NextSibling
		}
		tree.Nodes[curr].NextSibling = child
	}
}

func TestNameResolver_Variable(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	xID := pool.Intern([]byte("x"))
	
	declIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: xID})
	addChild(tree, 0, declIdx)
	
	refIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: xID})
	addChild(tree, 0, refIdx)

	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	symIdx := tree.Nodes[refIdx].Payload
	if symIdx == xID {
		t.Errorf("expected payload to be rewritten to symbol index")
	}
	sym := st.SymbolAt(symIdx)
	if sym.Kind != sema.SymVar {
		t.Errorf("expected SymVar")
	}
}

func TestNameResolver_Undefined(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	xID := pool.Intern([]byte("x"))
	
	refIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: xID})
	addChild(tree, 0, refIdx)

	diags := nr.Resolve()
	if len(diags) != 1 || diags[0].Code != 2010 {
		t.Fatalf("expected undefined error (2010), got %v", diags)
	}
}

func TestNameResolver_Shadowing(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	xID := pool.Intern([]byte("x"))
	
	outerDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: xID})
	addChild(tree, 0, outerDecl)
	
	block := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIfStmt})
	addChild(tree, 0, block)

	innerDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: xID})
	addChild(tree, block, innerDecl)
	
	refIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: xID})
	addChild(tree, block, refIdx)

	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	symIdx := tree.Nodes[refIdx].Payload
	sym := st.SymbolAt(symIdx)
	if sym.DeclNode != innerDecl {
		t.Errorf("expected inner decl shadowing outer")
	}
}

func TestNameResolver_ScopeExit(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	xID := pool.Intern([]byte("x"))
	
	block := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIfStmt})
	addChild(tree, 0, block)

	innerDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: xID})
	addChild(tree, block, innerDecl)
	
	refIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: xID})
	addChild(tree, 0, refIdx) // Add to global, not block

	diags := nr.Resolve()
	if len(diags) != 1 || diags[0].Code != 2010 {
		t.Fatalf("expected undefined error since x is out of scope")
	}
}

func TestNameResolver_ImportLazy(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// Add math export using the loader logic
	mathID := pool.Intern([]byte("math"))
	sqrtID := pool.Intern([]byte("sqrt"))

	lr = sema.NewLazyResolver(st, tt, func(m *sema.ModuleInfo, st *sema.SymbolTable, tt *types.TypeTable) error {
		idx, _ := st.Define(sqrtID, sema.SymFunc, 0, 999)
		m.Exports[sqrtID] = idx
		return nil
	})

	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	importDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeImportDecl, Payload: mathID})
	addChild(tree, 0, importDecl)
	
	fieldAcc := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeFieldExpr, Payload: sqrtID})
	addChild(tree, 0, fieldAcc)

	lhsIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: mathID})
	addChild(tree, fieldAcc, lhsIdx)

	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	if tree.Nodes[fieldAcc].Payload == sqrtID {
		t.Errorf("expected FieldExpr Payload to be rewritten to symbol index, not nameID")
	}
}

func TestNameResolver_DuplicateError(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	xID := pool.Intern([]byte("x"))
	
	decl1 := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: xID})
	addChild(tree, 0, decl1)
	
	decl2 := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: xID})
	addChild(tree, 0, decl2)

	diags := nr.Resolve()
	if len(diags) != 1 || diags[0].Code != 2001 {
		t.Fatalf("expected duplicate definition error (2001)")
	}
}

func TestNameResolver_GenericFunc(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	sortID := pool.Intern([]byte("sort"))
	tID := pool.Intern([]byte("T"))

	fnDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeFuncDecl, Payload: sortID})
	tree.SetFlags(fnDecl, ast.FlagIsGeneric)
	addChild(tree, 0, fnDecl)

	gpParams := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParams})
	addChild(tree, fnDecl, gpParams)

	gpParam := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParam, Payload: tID})
	addChild(tree, gpParams, gpParam)

	// A reference to T inside the function body
	refIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, fnDecl, refIdx)

	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	symIdx := tree.Nodes[refIdx].Payload
	sym := st.SymbolAt(symIdx)
	if sym.Kind != sema.SymGenericParam {
		t.Errorf("expected SymGenericParam for generic parameter")
	}

	tmpl, found := tt.FindGenericTemplate(tree.Nodes[fnDecl].Payload)
	if !found {
		t.Fatalf("expected to find generic template for sort")
	}
	if len(tmpl.Params) != 1 || tmpl.Params[0].NameID != tID {
		t.Errorf("expected 1 generic param T, got %v", tmpl.Params)
	}
}

func TestNameResolver_GenericStruct(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)

	stackID := pool.Intern([]byte("Stack"))
	tID := pool.Intern([]byte("T"))

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

	// A reference to T inside a field type
	fieldDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeFieldDecl, Payload: pool.Intern([]byte("items"))})
	addChild(tree, structDecl, fieldDecl)

	refIdx := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, fieldDecl, refIdx)

	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors: %v", diags)
	}

	symIdx := tree.Nodes[refIdx].Payload
	sym := st.SymbolAt(symIdx)
	if sym.Kind != sema.SymGenericParam {
		t.Errorf("expected SymGenericParam for generic parameter")
	}

	tmpl, found := tt.FindGenericTemplate(tree.Nodes[structDecl].Payload)
	if !found {
		t.Fatalf("expected to find generic template for struct")
	}
	if len(tmpl.Params) != 1 || tmpl.Params[0].NameID != tID {
		t.Errorf("expected 1 generic param T, got %v", tmpl.Params)
	}
}

