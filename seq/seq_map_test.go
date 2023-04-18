package seq

import (
    "math/rand"
    "os"
    "runtime"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

func TestSeq_MapSliceN(t *testing.T) {
    n := 999
    seq := FromIntSeq().Take(n)
    s := seq.MapSliceN(3)
    s.ForEach(func(is []any) {
        if len(is) != 3 {
            t.Fail()
        }
    })
}

func TestSeq_ParallelOrdered(t *testing.T) {
    preTest(t)
    n := 1000
    var maxConcurrency int32
    var nowConcurrency int32
    lock := sync.Mutex{}
    concurrency := 5 + rand.Intn(n/3)
    FromIntSeq().Take(n).MapParallel(func(i int) any {
        c := atomic.AddInt32(&nowConcurrency, 1)
        if c > atomic.LoadInt32(&maxConcurrency) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrency)
            if x <= c {
                maxConcurrency = c
            }
            lock.Unlock()
        }
        s := 20*time.Millisecond + time.Duration(rand.Intn(60000))*time.Microsecond
        //println("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrency, -1)
        //println("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 1, concurrency).Sync().Complete()
    if maxConcurrency != int32(concurrency) {
        println("maxConcurrency:", maxConcurrency, "concurrency:", concurrency)
        t.Fail()
    }
}

func TestSeq_ParallelOrdered1(t *testing.T) {
    preTest(t)
    it := IteratorInt()
    n := 1000
    var maxConcurrency int32
    var nowConcurrency int32
    lock := sync.Mutex{}
    concurrency := 5 + rand.Intn(n/2)
    FromIntSeq().Take(n).MapParallel(func(i int) any {
        c := atomic.AddInt32(&nowConcurrency, 1)
        if c > atomic.LoadInt32(&maxConcurrency) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrency)
            if x <= c {
                maxConcurrency = c
            }
            lock.Unlock()
        }
        s := 20*time.Millisecond + time.Duration(rand.Intn(60000))*time.Microsecond
        //println("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrency, -1)
        //println("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 2, concurrency).ForEach(func(ia any) {
        runtime.Gosched()
        i := ia.(int)
        i2, _ := it()
        //println("test", i, "expect", i2)
        if i != i2 {
            t.Fail()
            runtime.Gosched()
            println("test", i, "expect", i2)
            os.Exit(1)
        }
    })
    if maxConcurrency != int32(concurrency) {
        println("maxConcurrency:", maxConcurrency, "concurrency:", concurrency)
        t.Fail()
    }
}

func TestSeq_MergeBiInt(t *testing.T) {
    s := FromIntSeq().Take(1000).MergeBiInt(IteratorInt(111)).Cache()
    {
        it := IteratorInt()
        s.SeqV().ForEach(func(i int) {
            i2, _ := it()
            if i != i2 {
                t.Fail()
            }
        })
    }
    {
        it := IteratorInt(111)
        s.SeqK().ForEach(func(i int) {
            i2, _ := it()
            if i != i2 {
                t.Fail()
            }
        })
    }
}
