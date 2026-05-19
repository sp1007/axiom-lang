package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestSumTypeColor(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// type Color = Red | Green | Blue
	colorID := pool.Intern([]byte("Color"))
	redID := pool.Intern([]byte("Red"))
	greenID := pool.Intern([]byte("Green"))
	blueID := pool.Intern([]byte("Blue"))

	aliasDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeAliasDecl, Payload: colorID})
	addChild(tree, 0, aliasDecl)

	sumType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeSumType})
	addChild(tree, aliasDecl, sumType)

	varRed := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: redID})
	addChild(tree, sumType, varRed)

	varGreen := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: greenID})
	addChild(tree, sumType, varGreen)

	varBlue := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: blueID})
	addChild(tree, sumType, varBlue)

	// 1. Resolve names
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors resolving: %v", diags)
	}

	// 2. Type check
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tc.Check()

	if len(tc.Errors()) > 0 {
		t.Fatalf("unexpected type errors: %v", tc.Errors())
	}

	// Verify symbol and type
	symIdx := tree.Nodes[aliasDecl].Payload
	sym := st.SymbolAt(symIdx)

	if sym.TypeID == 0 {
		t.Fatalf("TypeID not assigned to Color")
	}

	sumInfo := tt.SumInfo(types.TypeID(sym.TypeID))
	if len(sumInfo.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(sumInfo.Variants))
	}

	if sumInfo.Variants[0].NameID != redID || sumInfo.Variants[0].Tag != 0 || sumInfo.Variants[0].PayloadType != 0 {
		t.Errorf("incorrect Red variant")
	}

	if sumInfo.Variants[1].NameID != greenID || sumInfo.Variants[1].Tag != 1 || sumInfo.Variants[1].PayloadType != 0 {
		t.Errorf("incorrect Green variant")
	}

	if sumInfo.Variants[2].NameID != blueID || sumInfo.Variants[2].Tag != 2 || sumInfo.Variants[2].PayloadType != 0 {
		t.Errorf("incorrect Blue variant")
	}

	// Verify variant symbols have the correct TypeID
	varRedSym := st.SymbolAt(tree.Nodes[varRed].Payload)
	if varRedSym.TypeID != sym.TypeID {
		t.Errorf("Red variant symbol does not have Color TypeID")
	}
}

func TestSumTypeResult(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// Register primitives
	i32ID := pool.Intern([]byte("i32"))
	symI32, _ := st.Define(i32ID, sema.SymBuiltinType, 0, 0)
	st.SymbolAt(symI32).TypeID = uint32(types.TypeI32)

	strID := pool.Intern([]byte("string"))
	symStr, _ := st.Define(strID, sema.SymBuiltinType, 0, 0)
	st.SymbolAt(symStr).TypeID = uint32(types.TypeString)

	// type Result[T, E] = Ok(T) | Err(E)
	resID := pool.Intern([]byte("Result"))
	tID := pool.Intern([]byte("T"))
	eID := pool.Intern([]byte("E"))
	okID := pool.Intern([]byte("Ok"))
	errID := pool.Intern([]byte("Err"))

	aliasDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeAliasDecl, Payload: resID})
	tree.SetFlags(aliasDecl, ast.FlagIsGeneric)
	addChild(tree, 0, aliasDecl)

	gpParams := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParams})
	addChild(tree, aliasDecl, gpParams)

	gpT := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParam, Payload: tID})
	addChild(tree, gpParams, gpT)

	gpE := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeGenericParam, Payload: eID})
	addChild(tree, gpParams, gpE)

	sumType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeSumType})
	addChild(tree, aliasDecl, sumType)

	varOk := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: okID})
	addChild(tree, sumType, varOk)
	
	typeT := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, varOk, typeT)
	
	identT := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: tID})
	addChild(tree, typeT, identT)

	varErr := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: errID})
	addChild(tree, sumType, varErr)
	
	typeE := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, varErr, typeE)
	
	identE := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: eID})
	addChild(tree, typeE, identE)

	// 1. Resolve names
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	diags := nr.Resolve()
	if len(diags) > 0 {
		t.Fatalf("unexpected errors resolving: %v", diags)
	}

	// 2. Instantiate Result[i32, string]
	mono := sema.NewMonomorphizer(tree, pool, st, tt)
	symIdx := tree.Nodes[aliasDecl].Payload
	instSymID, mDiags := mono.InstantiateFunction(symIdx, []types.TypeID{types.TypeI32, types.TypeString})
	if len(mDiags) > 0 {
		t.Fatalf("unexpected errors monomorphizing: %v", mDiags)
	}

	sym := st.SymbolAt(instSymID)
	if sym.TypeID == 0 {
		t.Fatalf("TypeID not assigned to Result[i32, string]")
	}

	sumInfo := tt.SumInfo(types.TypeID(sym.TypeID))
	if len(sumInfo.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(sumInfo.Variants))
	}

	if sumInfo.Variants[0].NameID != okID || sumInfo.Variants[0].Tag != 0 || sumInfo.Variants[0].PayloadType != types.TypeI32 {
		t.Errorf("incorrect Ok variant, got payload %d", sumInfo.Variants[0].PayloadType)
	}

	if sumInfo.Variants[1].NameID != errID || sumInfo.Variants[1].Tag != 1 || sumInfo.Variants[1].PayloadType != types.TypeString {
		t.Errorf("incorrect Err variant, got payload %d", sumInfo.Variants[1].PayloadType)
	}
}

func TestMatchExhaustive(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// type Color = Red | Green
	colorID := pool.Intern([]byte("Color"))
	redID := pool.Intern([]byte("Red"))
	greenID := pool.Intern([]byte("Green"))

	aliasDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeAliasDecl, Payload: colorID})
	addChild(tree, 0, aliasDecl)

	sumType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeSumType})
	addChild(tree, aliasDecl, sumType)

	varRed := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: redID})
	addChild(tree, sumType, varRed)

	varGreen := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: greenID})
	addChild(tree, sumType, varGreen)

	// var c: Color
	cID := pool.Intern([]byte("c"))
	varDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: cID})
	addChild(tree, 0, varDecl)
	varType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, varDecl, varType)
	varIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: colorID})
	addChild(tree, varType, varIdent)

	// match c { Red: 1, Green: 2 }
	matchStmt := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeMatchStmt})
	addChild(tree, 0, matchStmt)

	matchScrutinee := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: cID})
	addChild(tree, matchStmt, matchScrutinee)

	// Arm 1: Red
	arm1 := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeMatchArm})
	addChild(tree, matchStmt, arm1)
	arm1Pat := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: redID})
	addChild(tree, arm1, arm1Pat)
	arm1Body := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeBlock})
	addChild(tree, arm1, arm1Body)

	// Arm 2: Green
	arm2 := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeMatchArm})
	addChild(tree, matchStmt, arm2)
	arm2Pat := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: greenID})
	addChild(tree, arm2, arm2Pat)
	arm2Body := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeBlock})
	addChild(tree, arm2, arm2Body)

	// Resolve & Check
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	nr.Resolve()
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tc.Check()

	if len(tc.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", tc.Errors())
	}
}

func TestMatchNonExhaustive(t *testing.T) {
	tree, pool, st, tt, lr := setupResolverTest()

	// type Color = Red | Green
	colorID := pool.Intern([]byte("Color"))
	redID := pool.Intern([]byte("Red"))
	greenID := pool.Intern([]byte("Green"))

	aliasDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeAliasDecl, Payload: colorID})
	addChild(tree, 0, aliasDecl)

	sumType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeSumType})
	addChild(tree, aliasDecl, sumType)

	varRed := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: redID})
	addChild(tree, sumType, varRed)

	varGreen := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVariantDecl, Payload: greenID})
	addChild(tree, sumType, varGreen)

	// var c: Color
	cID := pool.Intern([]byte("c"))
	varDecl := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeVarDecl, Payload: cID})
	addChild(tree, 0, varDecl)
	varType := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeTypeExpr})
	addChild(tree, varDecl, varType)
	varIdent := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: colorID})
	addChild(tree, varType, varIdent)

	// match c { Red: 1 } // missing Green
	matchStmt := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeMatchStmt})
	addChild(tree, 0, matchStmt)

	matchScrutinee := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: cID})
	addChild(tree, matchStmt, matchScrutinee)

	// Arm 1: Red
	arm1 := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeMatchArm})
	addChild(tree, matchStmt, arm1)
	arm1Pat := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeIdent, Payload: redID})
	addChild(tree, arm1, arm1Pat)
	arm1Body := uint32(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, ast.AstNode{Kind: ast.NodeBlock})
	addChild(tree, arm1, arm1Body)

	// Resolve & Check
	nr := sema.NewNameResolver(tree, pool, st, tt, lr)
	nr.Resolve()
	ie := sema.NewInferenceEngine(tree, st, tt, nil)
	tc := sema.NewTypeChecker(tree, pool, st, tt, ie)
	tc.Check()

	if len(tc.Errors()) == 0 {
		t.Fatalf("expected non-exhaustive match error")
	}

	found := false
	for _, e := range tc.Errors() {
		if e.Code == 3030 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error 3030, got %v", tc.Errors())
	}
}
