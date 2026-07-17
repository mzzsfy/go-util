package concurrent

import (
    "runtime"
    "sync"
    "sync/atomic"
    "testing"
    "unsafe"
)

func TestSize(t *testing.T) {
    t.Log(0, unsafe.Sizeof(struct {
        int64
    }{}))
    t.Log(0, unsafe.Sizeof(struct {
        int64
        _ [0]byte
    }{}))
    t.Log(8, unsafe.Sizeof(struct {
        int64
        _ [8]byte
    }{}))
    t.Log(16, unsafe.Sizeof(struct {
        int64
        _ [16]byte
    }{}))
    t.Log(24, unsafe.Sizeof(struct {
        int64
        _ [24]byte
    }{}))
    t.Log(56, unsafe.Sizeof(struct {
        int64
        _ [56]byte
    }{}))
    t.Log(120, unsafe.Sizeof(struct {
        int64
        _ [120]byte
    }{}))
}

func NewWaitGroup(init int) *sync.WaitGroup {
    waitGroup := sync.WaitGroup{}
    waitGroup.Add(init)
    return &waitGroup
}

func Benchmark_bitInt64Adder_Bit(b *testing.B) {
    goroutineNum := 1000
    b.Run("Benchmark_bitInt64Adder_0Bit", func(b *testing.B) {
        adder := &struct {
            int32
            int64
            values []struct {
                int64
            }
        }{
            values: make([]struct{ int64 }, slotNumber),
        }
        wg := NewWaitGroup(goroutineNum)
        wg1 := NewWaitGroup(goroutineNum)
        wg2 := NewWaitGroup(1)
        for i := 0; i < goroutineNum; i++ {
            go func() {
                wg1.Done()
                defer wg.Done()
                wg2.Wait()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    atomic.AddInt64(&adder.values[int(id)&modNumber].int64, 1)
                    if i%128 == 0 {
                        runtime.Gosched()
                    }
                }
            }()
        }
        wg1.Wait()
        b.ResetTimer()
        wg2.Done()
        wg.Wait()
    })
    b.Run("Benchmark_bitInt64Adder_8Bit", func(b *testing.B) {

        adder := &struct {
            int32
            int64
            values []struct {
                int64
                _ [8]byte
            }
        }{
            values: make([]struct {
                int64
                _ [8]byte
            }, slotNumber),
        }
        wg := NewWaitGroup(goroutineNum)
        wg1 := NewWaitGroup(goroutineNum)
        wg2 := NewWaitGroup(1)
        for i := 0; i < goroutineNum; i++ {
            go func() {
                wg1.Done()
                defer wg.Done()
                wg2.Wait()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    atomic.AddInt64(&adder.values[int(id)&modNumber].int64, 1)
                    if i%128 == 0 {
                        runtime.Gosched()
                    }
                }
            }()
        }
        wg1.Wait()
        b.ResetTimer()
        wg2.Done()
        wg.Wait()
    })
    b.Run("Benchmark_bitInt64Adder_16Bit", func(b *testing.B) {

        adder := &struct {
            int32
            int64
            values []struct {
                int64
                _ [16]byte
            }
        }{
            values: make([]struct {
                int64
                _ [16]byte
            }, slotNumber),
        }
        wg := NewWaitGroup(goroutineNum)
        wg1 := NewWaitGroup(goroutineNum)
        wg2 := NewWaitGroup(1)
        for i := 0; i < goroutineNum; i++ {
            go func() {
                wg1.Done()
                defer wg.Done()
                wg2.Wait()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    atomic.AddInt64(&adder.values[int(id)&modNumber].int64, 1)
                    if i%128 == 0 {
                        runtime.Gosched()
                    }
                }
            }()
        }
        wg1.Wait()
        b.ResetTimer()
        wg2.Done()
        wg.Wait()
    })
    b.Run("Benchmark_bitInt64Adder_24Bit", func(b *testing.B) {

        adder := &struct {
            int32
            int64
            values []struct {
                int64
                _ [24]byte
            }
        }{
            values: make([]struct {
                int64
                _ [24]byte
            }, slotNumber),
        }
        wg := NewWaitGroup(goroutineNum)
        wg1 := NewWaitGroup(goroutineNum)
        wg2 := NewWaitGroup(1)
        for i := 0; i < goroutineNum; i++ {
            go func() {
                wg1.Done()
                defer wg.Done()
                wg2.Wait()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    atomic.AddInt64(&adder.values[int(id)&modNumber].int64, 1)
                    if i%128 == 0 {
                        runtime.Gosched()
                    }
                }
            }()
        }
        wg1.Wait()
        b.ResetTimer()
        wg2.Done()
        wg.Wait()
    })
    b.Run("Benchmark_bitInt64Adder_56Bit", func(b *testing.B) {

        adder := &struct {
            int32
            int64
            values []struct {
                int64
                _ [56]byte
            }
        }{
            values: make([]struct {
                int64
                _ [56]byte
            }, slotNumber),
        }
        wg := NewWaitGroup(goroutineNum)
        wg1 := NewWaitGroup(goroutineNum)
        wg2 := NewWaitGroup(1)
        for i := 0; i < goroutineNum; i++ {
            go func() {
                wg1.Done()
                defer wg.Done()
                wg2.Wait()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    atomic.AddInt64(&adder.values[int(id)&modNumber].int64, 1)
                    if i%128 == 0 {
                        runtime.Gosched()
                    }
                }
            }()
        }
        wg1.Wait()
        b.ResetTimer()
        wg2.Done()
        wg.Wait()
    })
    b.Run("Benchmark_bitInt64Adder_120Bit", func(b *testing.B) {

        adder := &struct {
            int32
            int64
            values []struct {
                int64
                _ [120]byte
            }
        }{
            values: make([]struct {
                int64
                _ [120]byte
            }, slotNumber),
        }
        wg := NewWaitGroup(goroutineNum)
        wg1 := NewWaitGroup(goroutineNum)
        wg2 := NewWaitGroup(1)
        for i := 0; i < goroutineNum; i++ {
            go func() {
                wg1.Done()
                defer wg.Done()
                wg2.Wait()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    atomic.AddInt64(&adder.values[int(id)&modNumber].int64, 1)
                    if i%128 == 0 {
                        runtime.Gosched()
                    }
                }
            }()
        }
        wg1.Wait()
        b.ResetTimer()
        wg2.Done()
        wg.Wait()
    })
}
