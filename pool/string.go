package pool

import (
    "github.com/mzzsfy/go-util/storage"
    "sync"
    "sync/atomic"
)

// StringPool 字符串池,用数字代替字符串,用于Map的Key场景
type StringPool struct {
    idGen uint32
    lock  sync.RWMutex
    m     storage.Map[string, struct {
        id    uint32
        using uint32
    }]
}

func (p *StringPool) Use(s string) uint32 {
    p.lock.RLock()
    if v, ok := p.m.Get(s); ok {
        p.lock.RUnlock()
        atomic.AddUint32(&v.using, 1)
        return v.id
    }
    p.lock.RUnlock()
    p.lock.Lock()
    defer p.lock.Unlock()
    if v, ok := p.m.Get(s); ok {
        atomic.AddUint32(&v.using, 1)
        return v.id
    }
    id := atomic.AddUint32(&p.idGen, 1)
    p.m.Put(s, struct {
        id    uint32
        using uint32
    }{
        id:    id,
        using: 1,
    })
    return id
}

func (p *StringPool) UnUse(s string) {
    p.lock.RLock()
    if v, ok := p.m.Get(s); ok {
        p.lock.RUnlock()
        if atomic.AddUint32(&v.using, ^uint32(0)) == 0 {
            p.lock.Lock()
            defer p.lock.Unlock()
            if v, ok := p.m.Get(s); ok {
                if v.using == 0 {
                    p.m.Delete(s)
                }
            }
        }
        return
    }
    p.lock.RUnlock()
}
