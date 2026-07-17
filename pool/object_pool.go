package pool

import (
	"sync"
)

// ObjectPool 泛型对象池,用于复用对象以减少内存分配
type ObjectPool[T any] struct {
	p     sync.Pool
	reset reset[T]
}

// defaultValue 对象创建函数类型
type defaultValue[T any]func() *T

// reset 对象重置函数类型
type reset[T any]func(*T)

// NewObjectPool 创建对象池
// new: 对象创建函数,用于池为空时创建新对象
// reset: 对象重置函数,对象归还池时调用以重置状态,可为nil
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

// Get 从池中获取一个对象
func (p *ObjectPool[T]) Get() *T {
	return p.p.Get().(*T)
}

// Put 将对象归还到池中
// 如果设置了reset函数,会先调用reset重置对象状态
func (p *ObjectPool[T]) Put(t *T) {
	if p.reset != nil {
		p.reset(t)
	}
	p.p.Put(t)
}
