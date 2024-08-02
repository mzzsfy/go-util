package helper

import (
    "testing"
)

// TestNewSymbols tests the NewSymbols function.
func TestNewSymbols(t *testing.T) {
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
func TestSymbolEqual(t *testing.T) {
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
