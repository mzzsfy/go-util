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

func Test_Seq_MapSliceN(t *testing.T) {
    n := 999
    seq := FromIntSeq().Take(n)
    s := CastAny[[]int](seq.MapSliceN(3))
    s.ForEach(func(is []int) {
        if len(is) != 3 {
            t.Fail()
        }
    })
}

func Test_Seq_ParallelOrdered(t *testing.T) {
    preTest(t)
    n := 1000
    start := time.Now()
    var maxConcurrent int32
    var nowConcurrent int32
    lock := sync.Mutex{}
    concurrent := int(float64(n/10+rand.Intn(n-n/10)) * 0.9)
    FromIntSeq().Take(n).MapParallel(func(i int) any {
        c := atomic.AddInt32(&nowConcurrent, 1)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        s := 60*time.Millisecond + time.Duration(rand.Intn(100000))*time.Microsecond
        //t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrent, -1)
        //t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 1, concurrent).Complete()
    if maxConcurrent != int32(concurrent) {
        t.Log("maxConcurrent:", maxConcurrent, "concurrent:", concurrent)
        t.Fail()
    }
    t.Log("ok,use", time.Now().Sub(start).String())
}

func Test_Seq_ParallelOrdered1(t *testing.T) {
    preTest(t)
    start := time.Now()
    it := IteratorInt()
    var count int32
    n := 1000
    var maxConcurrent int32
    var nowConcurrent int32
    lock := sync.Mutex{}
    concurrent := int(float64(n/10+rand.Intn(n-n/10)) * 0.9)
    FromIntSeq().Take(n).MapParallel(func(i int) any {
        c := atomic.AddInt32(&nowConcurrent, 1)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        s := 60*time.Millisecond + time.Duration(rand.Intn(100000))*time.Microsecond
        //t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrent, -1)
        //t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 2, concurrent).ForEach(func(ia any) {
        atomic.AddInt32(&count, 1)
        runtime.Gosched()
        i := ia.(int)
        i2, _ := it()
        //t.Log("test", i, "expect", i2)
        if i != i2 {
            t.Fail()
            runtime.Gosched()
            t.Log("test", i, "expect", i2)
            os.Exit(1)
        }
    })
    t.Logf("ok,use %s, n:%d,concurrent:%d,maxConcurrent:%d", time.Now().Sub(start).String(), n, concurrent, maxConcurrent)
    if count != int32(n) {
        t.Log("count:", count, "n:", n)
        t.Fail()
    }
    if maxConcurrent != int32(concurrent) {
        t.Log("maxConcurrent:", maxConcurrent, "concurrent:", concurrent)
        t.Fail()
    }
}

func Test_Seq_MergeBiInt(t *testing.T) {
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
