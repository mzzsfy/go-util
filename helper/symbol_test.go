package helper

import (
	"testing"
)

// TestNewSymbols tests the NewSymbols function.
func Test_NewSymbols(t *testing.T) {
	sym := NewSymbols("test")
	sym1 := NewAnonymousSymbols()
	if sym.String() != "test" {
		t.Errorf("Expected name 'test' but got '%s'", sym)
	}
	if !sym.Equal(sym) {
		t.Errorf("Expected symbol to be equal to itself but it was not")
	}
	if !sym1.Equal(sym1) {
		t.Errorf("Expected symbol to be equal to itself but it was not")
	}
}

// TestSymbolEqual tests the anonymousSymbol Equal method.
func Test_SymbolEqual(t *testing.T) {
	sym1 := NewAnonymousSymbols()
	sym2 := NewAnonymousSymbols()
	sym3 := NewSymbols("test")

	if sym1.Equal(sym2) {
		t.Errorf("Expected two different symbols but they were equal")
	}

	if sym1.Equal(sym3) {
		t.Errorf("Expected two different symbols but they were equal")
	}
}

// Test_NamedSymbol_MarshalText 命名符号序列化为名称
func Test_NamedSymbol_MarshalText(t *testing.T) {
	t.Parallel()
	sym := NewSymbols("config-name")
	mt, ok := sym.(interface {
		MarshalText() ([]byte, error)
	})
	if !ok {
		t.Fatal("命名符号应实现 MarshalText")
	}
	text, err := mt.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText 不应报错: %v", err)
	}
	if string(text) != "config-name" {
		t.Errorf("MarshalText = %q, 期望名称原样输出", text)
	}
}

// Test_AnonymousSymbolNotNamed 匿名符号不含名称语义
func Test_AnonymousSymbolNotNamed(t *testing.T) {
	t.Parallel()
	// 匿名符号未实现 NamedSymbol 接口
	if _, ok := NewAnonymousSymbols().(NamedSymbol); ok {
		t.Error("匿名符号不应实现 NamedSymbol 接口")
	}
	if _, ok := NewSymbols("x").(NamedSymbol); !ok {
		t.Error("命名符号应实现 NamedSymbol 接口")
	}
}
