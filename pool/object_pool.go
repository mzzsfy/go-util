package pool

import (
    "sync"
)

type ObjectPool[T any] struct {
    p     sync.Pool
    reset reset[T]
}

type defaultValue[T any]func() *T

type reset[T any]func(*T)

func NewObjectPool[T any](new defaultValue[T], reset reset[T]) *ObjectPool[T] {
    return &ObjectPool[T]{
        p: sync.Pool{
            New: func() any {
                return new()
            },
        },
        reset: reset,
    }
}

func (p *ObjectPool[T]) Get() *T {
    return p.p.Get().(*T)
}

func (p *ObjectPool[T]) Put(t *T) {
    if p.reset != nil {
        p.reset(t)
    }
    p.p.Put(t)
}
