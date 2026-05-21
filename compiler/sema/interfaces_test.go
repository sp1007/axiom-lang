package sema_test

import (
	"testing"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/sema"
	"github.com/axiom-lang/axiom/compiler/types"
)

func TestInterfaces_ImplicitImplementation(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	// 1. Register Struct (MyStruct)
	structNameID := pool.Intern([]byte("MyStruct"))
	structTypeID := tt.RegisterStruct(structNameID, nil, nil)

	// 2. Register Interface (Printable)
	ifaceNameID := pool.Intern([]byte("Printable"))
	printMethodNameID := pool.Intern([]byte("my_print"))
	ifaceTypeID := tt.RegisterInterface(ifaceNameID, []types.MethodSig{
		{NameID: printMethodNameID, Params: []types.TypeID{}, Return: types.TypeVoid},
	})

	// 3. Add `print` method for `MyStruct` to SymbolTable
	methodSymIdx, _ := st.Define(printMethodNameID, sema.SymFunc, 0, 0)
	
	// Create FuncType for method (first param is MyStruct)
	methodFuncTypeID := tt.RegisterFunction([]types.TypeID{structTypeID}, types.TypeVoid, nil)
	methodSym := st.SymbolAt(methodSymIdx)
	methodSym.TypeID = uint32(methodFuncTypeID)

	// 4. Check Implementation
	interfaces := sema.NewInterfaces(st, tt)
	ok, missing := interfaces.ImplementsInterface(structTypeID, ifaceTypeID)

	if !ok {
		t.Fatalf("expected MyStruct to implement Printable, but it failed. Missing: %v", missing)
	}
}

func TestInterfaces_MissingMethod(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	structNameID := pool.Intern([]byte("EmptyStruct"))
	structTypeID := tt.RegisterStruct(structNameID, nil, nil)

	ifaceNameID := pool.Intern([]byte("Printable"))
	printMethodNameID := pool.Intern([]byte("my_print"))
	ifaceTypeID := tt.RegisterInterface(ifaceNameID, []types.MethodSig{
		{NameID: printMethodNameID, Params: []types.TypeID{}, Return: types.TypeVoid},
	})

	interfaces := sema.NewInterfaces(st, tt)
	ok, missing := interfaces.ImplementsInterface(structTypeID, ifaceTypeID)

	if ok {
		t.Fatalf("expected EmptyStruct to NOT implement Printable")
	}

	if len(missing) != 1 || missing[0].NameID != printMethodNameID {
		t.Fatalf("expected missing method 'print', got %v", missing)
	}
}

func TestInterfaces_WrongSignature(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	structNameID := pool.Intern([]byte("WrongStruct"))
	structTypeID := tt.RegisterStruct(structNameID, nil, nil)

	ifaceNameID := pool.Intern([]byte("Printable"))
	printMethodNameID := pool.Intern([]byte("my_print"))
	ifaceTypeID := tt.RegisterInterface(ifaceNameID, []types.MethodSig{
		{NameID: printMethodNameID, Params: []types.TypeID{}, Return: types.TypeVoid},
	})

	// Add `print` method, but returning i32 instead of void
	methodSymIdx, _ := st.Define(printMethodNameID, sema.SymFunc, 0, 0)
	methodFuncTypeID := tt.RegisterFunction([]types.TypeID{structTypeID}, types.TypeI32, nil)
	methodSym := st.SymbolAt(methodSymIdx)
	methodSym.TypeID = uint32(methodFuncTypeID)

	interfaces := sema.NewInterfaces(st, tt)
	ok, missing := interfaces.ImplementsInterface(structTypeID, ifaceTypeID)

	if ok {
		t.Fatalf("expected WrongStruct to NOT implement Printable due to wrong signature")
	}

	if len(missing) != 1 || missing[0].NameID != printMethodNameID {
		t.Fatalf("expected missing method 'print' (due to signature mismatch), got %v", missing)
	}
}
func TestInterfaces_BuiltinOrd(t *testing.T) {
	pool := ast.NewInternPool(16)
	st := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()

	interfaces := sema.NewInterfaces(st, tt)
	ok, _ := interfaces.ImplementsInterface(types.TypeI32, types.TypeOrd)

	if !ok {
		t.Fatalf("expected i32 to implement Builtin Ord")
	}

	okStr, _ := interfaces.ImplementsInterface(types.TypeString, types.TypeOrd)
	if !okStr {
		t.Fatalf("expected string to implement Builtin Ord")
	}

}
