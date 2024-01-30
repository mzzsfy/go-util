package pool

import (
    "sync"
)

type ObjectPool[T any] struct {
    p     sync.Pool
    new   DefaultValue[T]
    reset Reset[T]
}

type DefaultValue[T any]func() *T

type Reset[T any]func(*T)

func NewObjectPool[T any](new DefaultValue[T], reset Reset[T]) *ObjectPool[T] {
    return &ObjectPool[T]{
        p: sync.Pool{
            New: func() any {
                return new()
            },
        },
        new:   new,
        reset: reset,
    }
}

func (p *ObjectPool[T]) Get() *T {
    return p.p.Get().(*T)
}

func (p *ObjectPool[T]) Put(t *T) {
    p.reset(t)
    p.p.Put(t)
}
