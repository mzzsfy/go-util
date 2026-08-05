package pool

import (
	"bytes"
	"sync"
	"testing"
)

// 原生 sync.Pool 对照
func BenchmarkSyncPool_RawBuffer(b *testing.B) {
	p := sync.Pool{New: func() any { return new(bytes.Buffer) }}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get().(*bytes.Buffer)
			buf.Reset()
			buf.WriteString("hello")
			p.Put(buf)
		}
	})
}

func BenchmarkBufferPool_GetPut(b *testing.B) {
	p := NewBufferPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			buf.WriteString("hello")
			p.Put(buf)
		}
	})
}

func BenchmarkSyncPool_RawBytes(b *testing.B) {
	p := sync.Pool{New: func() any { b := make([]byte, 0, 16); return &b }}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bs := p.Get().(*[]byte)
			*bs = (*bs)[:0]
			*bs = append(*bs, "hello"...)
			p.Put(bs)
		}
	})
}

func BenchmarkBytePool_GetPut(b *testing.B) {
	p := NewSimpleBytesPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bs := p.Get()
			bs.WriteString("hello")
			p.Put(bs)
		}
	})
}

type benchObj struct {
	a int
	b string
}

func BenchmarkSyncPool_RawObject(b *testing.B) {
	p := sync.Pool{New: func() any { return &benchObj{} }}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			o := p.Get().(*benchObj)
			o.a = 0
			o.b = ""
			o.a = 42
			o.b = "x"
			p.Put(o)
		}
	})
}

func BenchmarkObjectPool_GetPut(b *testing.B) {
	p := NewObjectPool(func() *benchObj { return &benchObj{} }, func(o *benchObj) { o.a = 0; o.b = "" })
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			o := p.Get()
			o.a = 42
			o.b = "x"
			p.Put(o)
		}
	})
}
