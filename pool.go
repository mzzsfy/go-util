package util

import (
    "sync"
)

type Pool[T any] struct {
    p     sync.Pool
    new   DefaultValue[T]
    reset Reset[T]
}

type DefaultValue[T any]func() *T

type Reset[T any]func(*T)

func NewPool[T any](new DefaultValue[T], reset Reset[T]) *Pool[T] {
    return &Pool[T]{
        p: sync.Pool{
            New: func() any {
                return new()
            },
        },
        new:   new,
        reset: reset,
    }
}

func (p *Pool[T]) Get() *T {
    return p.p.Get().(*T)
}

func (p *Pool[T]) Put(t *T) {
    p.reset(t)
    p.p.Put(t)
}
