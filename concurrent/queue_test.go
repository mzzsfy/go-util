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

func Benchmark_LkQueue_q1(b *testing.B) {
    i := f1(b.N)
    if i != 0 {
        b.Fatal("count is not 0", i)
    }

}
func Benchmark_LkQueue_q2(b *testing.B) {
    i := f2(b.N)
    if i != 0 {
        b.Fatal("count is not 0", i)
    }
}

const num = 1000

//func Test_LkQueue_q1(t *testing.T) {
//    file, _ := os.OpenFile("cpu1.prof", os.O_RDWR|os.O_CREATE, os.ModePerm)
//    pprof.StartCPUProfile(file)
//    defer pprof.StopCPUProfile()
//    f1(num)
//}
//func Test_LkQueue_q2(t *testing.T) {
//    file, _ := os.OpenFile("cpu2.prof", os.O_RDWR|os.O_CREATE, os.ModePerm)
//    pprof.StartCPUProfile(file)
//    defer pprof.StopCPUProfile()
//    f2(num)
//}

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
func Benchmark_LkQueue_c(b *testing.B) {
    queue := make(chan int, 1024)
    wg := sync.WaitGroup{}
    count := int64(0)
    for i := 0; i < consumer; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-queue:
                    atomic.AddInt64(&count, -1)
                    continue
                default:
                    runtime.Gosched()
                }
                select {
                case <-queue:
                    atomic.AddInt64(&count, -1)
                case <-time.After(time.Millisecond * 10):
                    return
                }
            }
        }()
    }
    for i := 0; i < producer; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < b.N; i++ {
                queue <- i
                atomic.AddInt64(&count, 1)
            }
        }()
    }
    wg.Wait()
    if atomic.LoadInt64(&count) != 0 {
        b.Fatal("count is not 0")
    }
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
