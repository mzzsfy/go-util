//go:build !go1.24

package storage

import (
    "sync"
    "testing"
)

func BenchmarkConcurrentSwissMap_ReadAndWrite(b *testing.B) {
    m := makeSwissConcurrentMap[int, int]()
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

func BenchmarkSyncMap_ReadAndWrite(b *testing.B) {
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

func BenchmarkConcurrentSwissMap_Read(b *testing.B) {
    m := makeSwissConcurrentMap[int, int]()
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

func BenchmarkConcurrentSwissMap_Write(b *testing.B) {
    m := makeSwissConcurrentMap[int, int]()
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            m.Put(i%1000, 0)
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
