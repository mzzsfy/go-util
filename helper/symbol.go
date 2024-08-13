package helper

import (
    "sync/atomic"
)

var idGen = uint64(0)

type Symbol interface {
    Equal(Symbol) bool
    id() uint64
}

type NamedSymbol interface {
    Symbol
    String() string
}

type anonymousSymbol uint64

func (a anonymousSymbol) Equal(b Symbol) bool {
    return a.id() == b.id()
}

func (a anonymousSymbol) id() uint64 {
    return uint64(a)
}

type namedSymbol struct {
    name string
    anonymousSymbol
}

func (n namedSymbol) String() string {
    return n.name
}

func (n namedSymbol) MarshalText() (text []byte, err error) {
    return StringToBytes(n.name), nil
}

func NewSymbols(name string) NamedSymbol {
    return &namedSymbol{name: name, anonymousSymbol: NewAnonymousSymbols().(anonymousSymbol)}
}

func NewAnonymousSymbols() Symbol {
    return anonymousSymbol(atomic.AddUint64(&idGen, 1))
}
