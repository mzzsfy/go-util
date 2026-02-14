package helper

import (
    "github.com/mzzsfy/go-util/concurrent"
    "github.com/mzzsfy/go-util/pool"
    "github.com/mzzsfy/go-util/unsafe"
    "math"
    "math/rand"
    "runtime"
    "testing"
    "time"
)

func TestAddDelayTask(t *testing.T) {
    t.Parallel()
    interval := time.Millisecond * 10
    s := NewScheduler(interval)
    start := time.Now()
    maxTime := time.Millisecond
    for i := 0; i < 8; i++ {
        i := i
        duration := time.Microsecond * time.Duration(rand.Int63n(int64(math.Pow10(i))))
        if duration < 0 {
            duration = -duration
        }
        maxTime = duration
        s.AddDelayTask(duration, func() {
            if duration > time.Millisecond {
                // Allow small tolerance for scheduler interval precision
                if time.Since(start) < duration-interval*2 {
                    t.Error("Task ran too early", duration.String(), time.Since(start))
                }
                if time.Since(start) > duration+interval*10 {
                    t.Error("Task ran too late", duration.String(), time.Since(start))
                }
            }
            t.Log("task", i, time.Since(start).String())
        })
    }
    defer s.Stop()
    t.Log("maxTime", maxTime.String())
    time.Sleep(maxTime + time.Millisecond*10)
}

func TestAddIntervalTask(t *testing.T) {
    t.Parallel()
    s := NewScheduler()
    counter := 0
    s.AddIntervalTask(time.Millisecond*200, func() {
        counter++
        if counter >= 5 {
            s.Stop()
        }
    })

    time.Sleep(time.Second * 2)

    if counter != 5 {
        t.Errorf("Expected counter to be 5, got %d", counter)
    }
}

func BenchmarkScheduler_AddCronTask(b *testing.B) {
    s := NewScheduler(time.Millisecond * 10)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            s.AddDelayTask(time.Millisecond, func() {})
            s.AddDelayTask(time.Millisecond*10, func() {})
            s.AddDelayTask(time.Millisecond*100, func() {})
            s.AddDelayTask(time.Second, func() {})
            s.AddDelayTask(time.Second*10, func() {})
            s.AddDelayTask(time.Minute, func() {})
            s.AddDelayTask(time.Minute*10, func() {})
            s.AddDelayTask(time.Hour, func() {})
        }
    })
    b.Cleanup(func() {
        if b.N > 1000000 {
            b.Log(s.log())
        }
        s.Stop()
    })
}

func TestScheduler_CronTask(t *testing.T) {
    s := NewScheduler(time.Millisecond * 10)
    gp := pool.NewGopool()
    s.DoCallFn = func(f func()) { gp.Go(f) }
    t.Cleanup(func() { s.Stop() })
    var goid int64
    go func() {
        goid = unsafe.GoID()
    }()
    wg := NewWaitGroup(0)
    add := concurrent.Int64Adder{}
    run := concurrent.Int64Adder{}
    f := func() {
        run.IncrementSimple()
        wg.Done()
    }
    num := 3000000
    st := 5
    n := 7
    go func() {
        for {
            sum := run.SumInt()
            if sum == num*n {
                return
            }
            if sum == 0 {
                t.Log("adding", add.Sum(), "/", num*n)
                time.Sleep(time.Millisecond)
            } else {
                t.Log("running", run.Sum(), "/", num*n, gp.WorkerCount(), runtime.NumGoroutine())
                time.Sleep(time.Millisecond * 300)
            }
        }
    }()
    for i := 0; i < num; i++ {
        wg.Add(n)
        for j := st; j < st+n; j++ {
            add.IncrementSimple()
            s.AddDelayTask(time.Millisecond*time.Duration(j)*time.Duration(j), f)
        }
    }
    for i := 0; i < 1000; i++ {
        s.AddDelayTask(time.Hour*1, func() {})
        s.AddDelayTask(time.Hour*10, func() {})
        s.AddDelayTask(time.Hour*100, func() {})
        s.AddDelayTask(time.Hour*1000, func() {})
    }
    wg.Wait()
    time.Sleep(time.Millisecond * 50)
    go func() {
        t.Log("goid start", goid, "end", unsafe.GoID(), "new", unsafe.GoID()-goid)
    }()
}
