package concurrent

import (
    "github.com/mzzsfy/go-util/helper"
    "sync/atomic"
)

var idGen = uint64(0)

type Symbol interface {
    foo()
}

type NamedSymbol interface {
    Symbol
    String() string
}

type anonymousSymbol uint64

func (a anonymousSymbol) foo() {}

type namedSymbol struct {
    name string
    anonymousSymbol
}

func (n *namedSymbol) String() string {
    return n.name
}

func (n *namedSymbol) MarshalText() (text []byte, err error) {
    return helper.StringToBytes(n.name), nil
}

func NewSymbols(name string) NamedSymbol {
    return &namedSymbol{name: name, anonymousSymbol: NewAnonymousSymbols().(anonymousSymbol)}
}

func NewAnonymousSymbols() Symbol {
    return anonymousSymbol(atomic.AddUint64(&idGen, 1))
}
