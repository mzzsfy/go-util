package pool

import (
	"sync"
	"sync/atomic"

	"github.com/mzzsfy/go-util/storage"
)

type stringPoolEntry struct {
	id    uint64
	using uint32
}

// NewStringPool 创建字符串池
func NewStringPool() *StringPool {
	return &StringPool{
		m: storage.NewMap[string, *stringPoolEntry](),
	}
}

// StringPool 字符串池, 用数字代替字符串, 用于 Map 的 Key 场景
type StringPool struct {
	idGen uint64
	lock  sync.RWMutex
	m     storage.Map[string, *stringPoolEntry]
}

// Peek 查看字符串对应的ID,不存在则返回0
func (p *StringPool) Peek(s string) uint64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if v, ok := p.m.Get(s); ok {
		return v.id
	}
	return 0
}

// Use 获取字符串对应的ID并增加引用计数,不存在则创建
func (p *StringPool) Use(s string) uint64 {
	p.lock.RLock()
	if v, ok := p.m.Get(s); ok {
		// 在读锁内完成原子递增，避免 RUnlock 后 UnUse 将引用计数减到 0 并删除条目
		atomic.AddUint32(&v.using, 1)
		id := v.id
		p.lock.RUnlock()
		return id
	}
	p.lock.RUnlock()
	p.lock.Lock()
	defer p.lock.Unlock()
	if v, ok := p.m.Get(s); ok {
		atomic.AddUint32(&v.using, 1)
		return v.id
	}
	id := atomic.AddUint64(&p.idGen, 1)
	p.m.Put(s, &stringPoolEntry{
		id:    id,
		using: 1,
	})
	return id
}

// UnUse 释放一次引用, 引用归零时删除条目
// 线程安全: 使用写锁保证与 Use 的原子递增之间无 TOCTOU 窗口
func (p *StringPool) UnUse(s string) {
	p.lock.Lock()
	if v, ok := p.m.Get(s); ok {
		if atomic.AddUint32(&v.using, ^uint32(0)) == 0 {
			p.m.Delete(s)
		}
	}
	p.lock.Unlock()
}
