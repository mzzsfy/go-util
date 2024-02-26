package pool

import (
    "bytes"
    "sync"
)

// BufferPool is a pool of bytes.Buffer.
type BufferPool struct {
    pool   sync.Pool
    maxCap int
}

func NewBufferPool() *BufferPool {
    return &BufferPool{
        pool: sync.Pool{
            New: func() interface{} {
                return new(bytes.Buffer)
            },
        },
        maxCap: 2 * 1024,
    }
}
func (p *BufferPool) SetMaxCap(maxCap int) {
    if maxCap <= 16 {
        maxCap = 16
    }
    p.maxCap = maxCap
}

func (p *BufferPool) Get() *bytes.Buffer {
    return p.pool.Get().(*bytes.Buffer)
}
func (p *BufferPool) Put(b *bytes.Buffer) {
    b.Reset()
    if b.Cap() > p.maxCap {
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

type BytesPool struct {
    sync.Pool
    maxCap  int
    initCap int
}

func NewBytesPool() *BytesPool {
    b := &BytesPool{
        Pool:    sync.Pool{},
        maxCap:  2 * 1024,
        initCap: 16,
    }
    b.Pool.New = func() interface{} {
        return &Bytes{
            buf: make([]byte, b.initCap),
        }
    }
    return b
}

func (p *BytesPool) SetMaxCap(maxCap int) {
    if maxCap <= 16 {
        maxCap = 16
    }
    p.maxCap = maxCap
}
func (p *BytesPool) SetInitCap(initCap int) {
    if initCap <= 16 {
        initCap = 16
    }
    p.initCap = initCap
}
func (p *BytesPool) Get() *Bytes {
    return p.Pool.Get().(*Bytes)
}
func (p *BytesPool) Put(b *Bytes) {
    if cap(b.buf) > p.maxCap {
        return
    }
    b.buf = b.buf[:0]
    if cap(b.buf) < p.initCap {
        b.buf = make([]byte, 0, p.initCap)
    }
    p.Pool.Put(b)
}
