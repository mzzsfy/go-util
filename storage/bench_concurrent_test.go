//go:build go1.24

package storage

import (
	"sync"
	"testing"
)

func BenchmarkConcurrentWrapper_Read(b *testing.B) {
	m := MapTypeConcurrentWrapper(MapTypeGo[int, int](8)).createMap()
	for i := 0; i < 1000; i++ {
		m.Put(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 1000)
			i++
		}
	})
}

func BenchmarkConcurrentGoMap_Read(b *testing.B) {
	m := MapTypeConcurrentGo[int, int]().createMap()
	for i := 0; i < 1000; i++ {
		m.Put(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(i % 1000)
			i++
		}
	})
}

func BenchmarkConcurrentWrapper_Write(b *testing.B) {
	m := MapTypeConcurrentWrapper(MapTypeGo[int, int](8)).createMap()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Put(i%1000, 0)
			i++
		}
	})
}

func BenchmarkConcurrentGoMap_Write(b *testing.B) {
	m := MapTypeConcurrentGo[int, int]().createMap()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Put(i%1000, 0)
			i++
		}
	})
}

func BenchmarkConcurrentWrapper_ReadWrite(b *testing.B) {
	m := MapTypeConcurrentWrapper(MapTypeGo[int, int](8)).createMap()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				m.Get(i % 1000)
			} else {
				m.Put(i, i)
			}
			i++
		}
	})
}

func BenchmarkConcurrentGoMap_ReadWrite(b *testing.B) {
	m := MapTypeConcurrentGo[int, int]().createMap()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				m.Get(i % 1000)
			} else {
				m.Put(i, i)
			}
			i++
		}
	})
}

func BenchmarkSyncMap_Read(b *testing.B) {
	m := &sync.Map{}
	for i := 0; i < 1000; i++ {
		m.Store(i, i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Load(i % 1000)
			i++
		}
	})
}

func BenchmarkSyncMap_Write(b *testing.B) {
	m := &sync.Map{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Store(i%1000, 0)
			i++
		}
	})
}

func BenchmarkSyncMap_ReadWrite(b *testing.B) {
	m := &sync.Map{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				m.Load(i % 1000)
			} else {
				m.Store(i, i)
			}
			i++
		}
	})
}
