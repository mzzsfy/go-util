package pool

import (
	"bytes"
	"sync"
	"sync/atomic"
)

// minPoolCap 字节池最小容量阈值
const minPoolCap = 16

// defaultMaxBufferCap BufferPool默认最大容量,超过此容量的buffer不会被放回池中
const defaultMaxBufferCap = 2 * 1024

// defaultBytesMaxCap BytePool默认最大容量,超过此容量的字节切片不会被放回池中
const defaultBytesMaxCap = 256

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
		maxCap: defaultMaxBufferCap,
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

// Put 归还 buffer, Reset 后入池; 容量超过 maxCap 时丢弃不入池
// 归还后不得再使用该 buffer 及其 Bytes() 结果, 底层数组会被后续使用者覆盖
func (p *BufferPool) Put(b *bytes.Buffer) {
	if b.Cap() > int(atomic.LoadInt32(&p.maxCap)) {
		return
	}
	b.Reset()
	p.pool.Put(b)
}

// Bytes 可池化的字节缓冲
// 池化契约: Put 时会清空内容, Get 返回的对象长度为 0, 取出后无需再调用 Reset;
// Bytes()/String() 返回的结果不得在 Put 之后继续使用, 底层数组会被池的后续使用者覆盖
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
		maxCap:  defaultBytesMaxCap,
		initCap: minPoolCap,
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

// Get 从池中取出一个长度为 0 的缓冲(池化契约由 Put 保证), 无需再 Reset
func (p *BytePool) Get() *Bytes {
	return p.pool.Get().(*Bytes)
}

// Put 归还缓冲, 清空内容后入池; 容量超过 maxCap 时丢弃不入池
func (p *BytePool) Put(b *Bytes) {
	if cap(b.buf) > int(atomic.LoadInt32(&p.maxCap)) {
		return
	}
	b.buf = b.buf[:0]
	p.pool.Put(b)
}
