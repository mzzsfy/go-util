package seq

import (
    "math/rand"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

func Test_Seq_OnLast(t *testing.T) {
    exec := 0
    FromIntSeq().Take(10).OnLast(func(i *int) {
        if *i != 9 {
            t.Fail()
        }
        exec++
    }).ForEach(func(i int) {})
    if exec != 1 {
        t.Fail()
    }
}

func Test_Parallel(t *testing.T) {
    preTest(t)
    n := 30 + rand.Intn(1000)
    duration := time.Millisecond * 3000
    concurrent := 1 + rand.Intn(n-1)/(rand.Intn(10)+1)
    p := NewParallel(concurrent)
    now := time.Now()
    for i := 0; i < n; i++ {
        p.Add(func() {
            time.Sleep(duration / time.Duration(n/concurrent))
        })
    }
    p.Wait()
    sub := time.Now().Sub(now)
    if sub < duration || sub.Truncate(duration) != duration {
        t.Log("运行时间不正确", duration.String(), sub.String())
        t.Fail()
    } else {
        t.Log("ok,use ", sub.String())
    }
}

func Test_Seq_Parallel(t *testing.T) {
    preTest(t)
    duration := time.Millisecond * 2000
    seq := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    go func() {
        now := time.Now()
        seq.AsyncEach(func(i int) {
            time.Sleep(duration)
        })
        sub := time.Now().Sub(now)
        if sub < duration || sub.Truncate(duration) != duration {
            t.Fail()
        }
    }()
    now := time.Now()
    seq.Parallel().Map(func(i int) any {
        time.Sleep(duration)
        return i
    }).Complete()
    sub := time.Now().Sub(now)
    if sub < duration || sub.Truncate(duration) != duration {
        t.Fail()
    }
}

func Test_Seq_ParallelN(t *testing.T) {
    preTest(t)
    n := 30 + rand.Intn(1000)
    seq := FromIntSeq().Take(n)
    now := time.Now()
    duration := time.Millisecond * 2000
    concurrency := 1 + rand.Intn(n-1)/2
    var maxConcurrency int32
    var nowConcurrency int32
    lock := sync.Mutex{}
    seq.Parallel(concurrency).ForEach(func(i int) {
        c := atomic.AddInt32(&nowConcurrency, 1)
        if c > atomic.LoadInt32(&maxConcurrency) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrency)
            if x <= c {
                maxConcurrency = c
            }
            lock.Unlock()
        }
        time.Sleep(duration / time.Duration(n/concurrency))
        atomic.AddInt32(&nowConcurrency, -1)
    })
    sub := time.Now().Sub(now)
    if sub < duration || sub.Truncate(duration) != duration {
        t.Fail()
    }
    if maxConcurrency != int32(concurrency) {
        t.Log("maxConcurrency:", maxConcurrency, "concurrency:", concurrency)
        t.Fail()
    }
    t.Log("ok,use ", sub.String())
}

func Test_Cache(t *testing.T) {
    d := 0
    seq := From(func(f func(i int)) {
        d++
        for i := 0; i < 1000; i++ {
            //懒加载,可中断,所以不会执行到100以上
            if i > 100 {
                t.Fail()
            }
            f(i)
        }
    })
    cacheSeq := seq.Take(100)
    {
        var r []int
        cacheSeq.Drop(5).Take(5).ForEach(func(i int) { r = append(r, i) })
        if len(r) != 5 {
            t.Fail()
        }
        for i := 0; i < 5; i++ {
            if r[i] != i+5 {
                t.Fail()
            }
        }
    }
    {
        var r []int
        cacheSeq.Take(10).ForEach(func(i int) { r = append(r, i) })
        if len(r) != 10 {
            t.Fail()
        }
        for i := 0; i < 10; i++ {
            if r[i] != i {
                t.Fail()
            }
        }
    }
    if d != 2 {
        t.Fail()
    }
    cacheSeq = cacheSeq.Cache()
    {
        var r []int
        cacheSeq.Drop(5).Take(5).ForEach(func(i int) { r = append(r, i) })
        if len(r) != 5 {
            t.Fail()
        }
        for i := 0; i < 5; i++ {
            if r[i] != i+5 {
                t.Fail()
            }
        }
    }
    {
        var r []int
        cacheSeq.Take(10).ForEach(func(i int) { r = append(r, i) })
        if len(r) != 10 {
            t.Fail()
        }
        for i := 0; i < 10; i++ {
            if r[i] != i {
                t.Fail()
            }
        }
    }
    if d != 3 {
        t.Fail()
    }
}

func Test_Sort(t *testing.T) {
    //结果 "10,9,8 ... 3,2,1"
    if "10,9,8,7,6,5,4,3,2,1" != FromIntSeq(1).Take(10).Sort(func(i, j int) bool {
        return i > j
    }).JoinString(",") {
        t.Fail()
    }
}

func Test_Seq_Repeat(t *testing.T) {
    testI := 0
    repeatI := 0
    FromIntSeq(0, 10).Repeat(3).ForEach(func(i int) {
        if i != testI {
            t.Fail()
        }
        testI++
        if testI > 10 {
            testI = 0
            repeatI++
        }
    })
    if repeatI != 3 {
        t.Fail()
    }
}
