package stdlib_test

import (
	"testing"

	stdlib "github.com/axiom-lang/axiom/std"
)

func TestLookupTypeByName_Vec(t *testing.T) {
	info := stdlib.LookupTypeByName("Vec")
	if info == nil {
		t.Fatal("Vec not found")
	}
	if info.Module != "std::collections" {
		t.Errorf("module = %q", info.Module)
	}
	if !info.IsGeneric {
		t.Error("Vec should be generic")
	}
	if info.TypeParams != 1 {
		t.Errorf("type params = %d, expected 1", info.TypeParams)
	}
}

func TestLookupTypeByName_Option(t *testing.T) {
	info := stdlib.LookupTypeByName("Option")
	if info == nil {
		t.Fatal("Option not found")
	}
	if info.Module != "std::option" {
		t.Errorf("module = %q", info.Module)
	}
	if info.TypeParams != 1 {
		t.Errorf("type params = %d, expected 1", info.TypeParams)
	}
}

func TestLookupTypeByName_Result(t *testing.T) {
	info := stdlib.LookupTypeByName("Result")
	if info == nil {
		t.Fatal("Result not found")
	}
	if info.TypeParams != 2 {
		t.Errorf("type params = %d, expected 2", info.TypeParams)
	}
}

func TestLookupTypeByName_HashMap(t *testing.T) {
	info := stdlib.LookupTypeByName("HashMap")
	if info == nil {
		t.Fatal("HashMap not found")
	}
	if info.TypeParams != 2 {
		t.Errorf("type params = %d, expected 2", info.TypeParams)
	}
}

func TestLookupTypeByName_String(t *testing.T) {
	info := stdlib.LookupTypeByName("String")
	if info == nil {
		t.Fatal("String not found")
	}
	if info.IsGeneric {
		t.Error("String should not be generic")
	}
}

func TestLookupType_WithModule(t *testing.T) {
	info := stdlib.LookupType("Mutex", "std::sync")
	if info == nil {
		t.Fatal("Mutex not found in std::sync")
	}
	if info.Type != stdlib.TypeMutex {
		t.Error("wrong type")
	}
}

func TestLookupTypeByName_NotFound(t *testing.T) {
	info := stdlib.LookupTypeByName("FooBar")
	if info != nil {
		t.Error("expected nil for non-existent type")
	}
}

func TestAllStdTypes_HaveModule(t *testing.T) {
	for _, st := range stdlib.StdTypes {
		if st.Module == "" {
			t.Errorf("type %q has empty module", st.Name)
		}
		if st.Name == "" {
			t.Errorf("type ID %d has empty name", st.Type)
		}
	}
}

func TestGenericTypes_HaveParams(t *testing.T) {
	for _, st := range stdlib.StdTypes {
		if st.IsGeneric && st.TypeParams == 0 {
			t.Errorf("generic type %q has 0 type params", st.Name)
		}
		if !st.IsGeneric && st.TypeParams > 0 {
			t.Errorf("non-generic type %q has %d type params", st.Name, st.TypeParams)
		}
	}
}
