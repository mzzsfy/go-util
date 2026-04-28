package pool

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// minPoolCap 字节池最小容量阈值
const minPoolCap = 16

// BufferPool is a pool of bytes.Buffer.
type BufferPool struct {
	pool   sync.Pool
	maxCap int32
}

func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
		maxCap: 2 * 1024,
	}
}

func (p *BufferPool) SetMaxCap(maxCap int) {
	if maxCap <= minPoolCap {
		maxCap = minPoolCap
	}
	atomic.StoreInt32(&p.maxCap, int32(maxCap))
}

func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}
func (p *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	if b.Cap() > int(atomic.LoadInt32(&p.maxCap)) {
		return
	}
	p.pool.Put(b)
}

type Bytes struct {
	buf []byte
}

func (b *Bytes) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}
func (b *Bytes) WriteString(s string) (n int, err error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}
func (b *Bytes) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}
func (b *Bytes) Len() int {
	return len(b.buf)

}
func (b *Bytes) Cap() int {
	return cap(b.buf)
}
func (b *Bytes) Reset() {
	b.buf = b.buf[:0]
}
func (b *Bytes) String() string {
	return string(b.buf)
}
func (b *Bytes) Bytes() []byte {
	return b.buf
}

type BytePool struct {
	pool    sync.Pool
	maxCap  int32
	initCap int32
}

// NewSimpleBytesPool 创建一个简单的字节池,池内的字节初始容量与最大容量相对稳定
func NewSimpleBytesPool() *BytePool {
	b := &BytePool{
		pool:    sync.Pool{},
		maxCap:  256,
		initCap: 16,
	}
	b.pool.New = func() any {
		return &Bytes{
			buf: make([]byte, 0, int(atomic.LoadInt32(&b.initCap))),
		}
	}
	return b
}

func (p *BytePool) SetMaxCap(maxCap int) {
	if maxCap <= minPoolCap {
		maxCap = minPoolCap
	}
	atomic.StoreInt32(&p.maxCap, int32(maxCap))
	// 保证 initCap 不超过 maxCap, 否则池会不断丢弃和重分配
	if ic := atomic.LoadInt32(&p.initCap); ic > int32(maxCap) {
		atomic.StoreInt32(&p.initCap, int32(maxCap))
	}
}

func (p *BytePool) SetInitCap(initCap int) {
	if initCap <= minPoolCap {
		initCap = minPoolCap
	}
	// 保证 initCap 不超过 maxCap
	if mc := atomic.LoadInt32(&p.maxCap); int32(initCap) > mc {
		initCap = int(mc)
	}
	atomic.StoreInt32(&p.initCap, int32(initCap))
}

func (p *BytePool) Get() *Bytes {
	b := p.pool.Get().(*Bytes)
	// 保证返回的 buffer 至少有 initCap 容量
	initCap := atomic.LoadInt32(&p.initCap)
	if cap(b.buf) < int(initCap) {
		b.buf = make([]byte, 0, int(initCap))
	}
	return b
}

func (p *BytePool) Put(b *Bytes) {
	if cap(b.buf) > int(atomic.LoadInt32(&p.maxCap)) {
		return
	}
	b.buf = b.buf[:0]
	p.pool.Put(b)
}
