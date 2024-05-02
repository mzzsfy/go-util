package concurrent

import (
    "runtime"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

const consumer = 10
const producer = 1000

func zipAny(vars ...any) []any {
    return vars
}
func Test_DequeueTimeout(t *testing.T) {
    t.Parallel()
    queue := BlockQueueWrapper[int](BlockQueueWrapper(newLinkedQueue[int]()))
    go func() {
        for {
            i := 1
            time.Sleep(time.Millisecond * 500)
            queue.Enqueue(i)
        }
    }()
    t.Log(time.Now(), zipAny(queue.DequeueBlock(time.Millisecond*300)))
    t.Log(time.Now(), zipAny(queue.DequeueBlock(time.Millisecond*300)))
    t.Log(time.Now(), zipAny(queue.DequeueBlock()))
    t.Log(time.Now(), zipAny(queue.DequeueBlock(time.Millisecond*300)))
    t.Log(time.Now(), zipAny(queue.DequeueBlock()))
    t.Log(time.Now(), zipAny(queue.DequeueBlock()))
}

func Test_LkQueue1(b *testing.T) {
    num := 100000
    queue := NewQueue[int]()
    for i := 0; i < num; i++ {
        queue.Enqueue(1)
    }
    if queue.Size() != num {
        b.Fatal("插入数据量数量不正确")
        return
    }
    var x = Int64Adder{}
    for i := 0; i < num; i++ {
        _, exist := queue.Dequeue()
        if exist {
            x.IncrementSimple()
        } else {
            return
        }
    }
    if x.SumInt() != num {
        b.Fatal("消费数据量数量不正确", x.SumInt(), num)
    }
}

func Test_LkQueue2(b *testing.T) {
    gn := 10
    num := 10000
    n := num * gn
    wg := NewWaitGroup(gn)
    queue := NewQueue[int]()
    for g := 0; g < gn; g++ {
        go func() {
            defer wg.Done()
            for i := 0; i < num; i++ {
                queue.Enqueue(1)
            }
        }()
    }
    wg.Wait()
    if queue.Size() != n {
        b.Fatal("插入数据量数量不正确")
        return
    }
    var x = Int64Adder{}
    wg.Add(gn)
    for g := 0; g < gn; g++ {
        go func() {
            defer wg.Done()
            for i := 0; i < num; i++ {
                _, exist := queue.Dequeue()
                if exist {
                    x.IncrementSimple()
                } else {
                    i--
                }
            }
        }()
    }
    wg.Wait()
    if x.SumInt() != n {
        b.Fatal("消费数据量数量不正确", x.SumInt(), n)
    }
}
func Benchmark_LkQueue(b *testing.B) {
    b.Run("Enqueue", func(b *testing.B) {
        queue := NewQueue(WithTypeLink[int]())
        over := int32(0)
        b.Cleanup(func() {
            atomic.StoreInt32(&over, 1)
        })
        for i := 0; i < 32; i++ {
            go func() {
                for {
                    _, ok := queue.Dequeue()
                    if !ok {
                        if atomic.LoadInt32(&over) == 1 {
                            return
                        }
                    }
                }
            }()
        }
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                queue.Enqueue(1)
            }
        })
        b.Log("hit", hit.Sum(), "noHit", noHit.Sum())
    })
    b.Run("Dequeue", func(b *testing.B) {
        queue := NewQueue(WithTypeLink[int]())
        over := int32(0)
        b.Cleanup(func() {
            atomic.StoreInt32(&over, 1)
        })
        for i := 0; i < 3; i++ {
            go func() {
                for {
                    if queue.Size() > 3000 {
                        time.Sleep(100 * time.Millisecond)
                    } else {
                        for j := 0; j < 1000; j++ {
                            queue.Enqueue(1)
                        }
                    }
                    if atomic.LoadInt32(&over) == 1 {
                        return
                    }
                }
            }()
        }
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                for {
                    _, ok := queue.Dequeue()
                    if ok {
                        break
                    }
                }
            }
        })
        b.Log("hit", hit.Sum(), "noHit", noHit.Sum())
    })
    b.Run("chan Enqueue", func(b *testing.B) {
        queue := make(chan int, 10)
        over := int32(0)
        b.Cleanup(func() {
            atomic.StoreInt32(&over, 1)
        })
        for i := 0; i < 10; i++ {
            go func() {
                for {
                    _, ok := <-queue
                    if !ok {
                        if atomic.LoadInt32(&over) == 1 {
                            return
                        }
                    }
                }
            }()
        }
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                queue <- 1
            }
        })
    })
    b.Run("chan Dequeue", func(b *testing.B) {
        queue := make(chan int)
        over := int32(0)
        b.Cleanup(func() {
            atomic.StoreInt32(&over, 1)
        })
        for i := 0; i < 3; i++ {
            go func() {
                for {
                    for j := 0; j < 100; j++ {
                        queue <- 1
                    }
                    if atomic.LoadInt32(&over) == 1 {
                        return
                    }
                }
            }()
        }
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                for {
                    _, ok := <-queue
                    if ok {
                        break
                    }
                }
            }
        })
    })
}

func Benchmark_LkQueue111(b *testing.B) {
    b.Run("lk", func(b *testing.B) {
        queue := NewQueue(WithTypeLink[int]())
        wg := NewWaitGroup(0)
        en := Int64Adder{}
        de := Int64Adder{}
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func() {
                id := GoID()
                for i := 0; i < b.N; i++ {
                    en.Increment(id)
                    queue.Enqueue(1)
                }
            }()
            go func() {
                defer wg.Done()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    x, ok := queue.Dequeue()
                    if !ok {
                        i--
                        runtime.Gosched()
                    } else {
                        if x != 1 {
                            b.Error("数据错误")
                            b.FailNow()
                            return
                        }
                        de.Increment(id)
                    }
                }
            }()
        }
        time.Sleep(10 * time.Millisecond)
        wg.Wait()
        if en.Sum() != de.Sum() {
            b.Error("入队出队数量不匹配")
        }
    })
    b.Run("chan", func(b *testing.B) {
        queue := make(chan int, 10)
        wg := NewWaitGroup(0)
        en := Int64Adder{}
        de := Int64Adder{}
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func() {
                id := GoID()
                for i := 0; i < b.N; i++ {
                    en.Increment(id)
                    queue <- 1
                }
            }()
            go func() {
                defer wg.Done()
                id := GoID()
                for i := 0; i < b.N; i++ {
                    x, ok := <-queue
                    if !ok {
                        i--
                    } else {
                        if x != 1 {
                            b.Error("数据错误")
                            b.FailNow()
                        }
                        de.Increment(id)
                    }
                }
            }()
        }
        time.Sleep(10 * time.Millisecond)
        wg.Wait()
        if en.Sum() != de.Sum() {
            b.Error("入队出队数量不匹配")
        }
    })
}

func Benchmark_LkQueue_q1(b *testing.B) {
    b.Run("f1", func(b *testing.B) {
        i := f1(b.N)
        if i != 0 {
            b.Fatal("count is not 0", i)
        }
    })
    b.Run("f2", func(b *testing.B) {
        i := f2(b.N)
        if i != 0 {
            b.Fatal("count is not 0", i)
        }
    })
}

func f1(num int) int64 {
    queue := BlockQueueWrapper(newLinkedQueue[int]())
    wg := sync.WaitGroup{}
    count := int64(0)
    for i := 0; i < consumer; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                _, b := queue.DequeueBlock(time.Millisecond * 10)
                if !b {
                    return
                }
                atomic.AddInt64(&count, -1)
            }
        }()
    }
    for i := 0; i < producer; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < num; i++ {
                queue.Enqueue(i)
                atomic.AddInt64(&count, 1)
            }
        }()
    }
    wg.Wait()
    return atomic.LoadInt64(&count)
}

func f2(num int) int64 {
    queue := newLinkedQueue[int]()
    wg := sync.WaitGroup{}
    count := int64(0)
    for i := 0; i < consumer; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            x := int32(0)
            for {
                _, b := queue.Dequeue()
                if b {
                    atomic.AddInt64(&count, -1)
                    x = 0
                } else {
                    if atomic.AddInt32(&x, 1) > 1000 {
                        return
                    }
                    runtime.Gosched()
                }
            }
        }()
    }
    for i := 0; i < producer; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < num; i++ {
                queue.Enqueue(i)
                atomic.AddInt64(&count, 1)
            }
        }()
    }
    wg.Wait()
    return atomic.LoadInt64(&count)
}

func TestDelayQueue_Dequeue(t *testing.T) {
    t.Parallel()
    queue := newDelayQueue[int](time.Millisecond * 100)
    queue.Enqueue(1)
    queue.Enqueue(2)
    queue.Enqueue(3)
    queue.Enqueue(4)
    queue.Enqueue(5)
    dequeue, b := queue.Dequeue()
    if b {
        t.Fatal("dequeue1", dequeue)
    }
    time.Sleep(time.Millisecond * 10)
    dequeue, b = queue.Dequeue()
    if b {
        t.Fatal("dequeue2", dequeue)
    }
    time.Sleep(time.Millisecond * 100)
    dequeue, b = queue.Dequeue()
    if !b {
        t.Fatal("dequeue3", dequeue)
    }
    t.Log(queue.Dequeue())
    t.Log(queue.Dequeue())
    t.Log(queue.Dequeue())
    t.Log(queue.Dequeue())
    t.Log(queue.Dequeue())
}
