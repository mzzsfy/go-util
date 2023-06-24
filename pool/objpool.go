package pool

import (
    "sync"
)

type Objpool[T any] struct {
    p     sync.Pool
    new   DefaultValue[T]
    reset Reset[T]
}

type DefaultValue[T any]func() *T

type Reset[T any]func(*T)

func NewObjpool[T any](new DefaultValue[T], reset Reset[T]) *Objpool[T] {
    return &Objpool[T]{
        p: sync.Pool{
            New: func() any {
                return new()
            },
        },
        new:   new,
        reset: reset,
    }
}

func (p *Objpool[T]) Get() *T {
    return p.p.Get().(*T)
}

func (p *Objpool[T]) Put(t *T) {
    p.reset(t)
    p.p.Put(t)
}
