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
    s := CastAny[[]int](MapSliceN(seq, 3))
    s.ForEach(func(is []int) {
        if len(is) != 3 {
            t.Fail()
        }
    })
}

func TestSeq_ParallelOrdered1_100(t *testing.T) {
    for i := 0; i < 100; i++ {
        t.Run("Test_Seq_ParallelOrdered1", Test_Seq_ParallelOrdered1)
    }
}

func Test_Seq_ParallelOrdered1(t *testing.T) {
    preTest(t)
    var count int32
    n := rand.Intn(3000) + 100
    start := time.Now()
    var maxConcurrent int32
    var nowConcurrent int32
    lock := sync.Mutex{}
    concurrent := 1 + int(float64(n/10+rand.Intn(n-n/7))*0.9)
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
        s := 150*time.Millisecond + time.Duration(rand.Intn(150000))*time.Microsecond
        //t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrent, -1)
        c = atomic.LoadInt32(&nowConcurrent)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        //t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 1, concurrent).ForEach(func(ia any) {
        atomic.AddInt32(&count, 1)
    })
    t.Logf("ok,use %s, n:%d,concurrent:%d,maxConcurrent:%d", time.Now().Sub(start).String(), n, concurrent, maxConcurrent)
    if maxConcurrent != int32(concurrent) {
        t.Fail()
    }
    if count != int32(n) {
        t.Fail()
    }
}

func TestSeq_ParallelOrdered2_100(t *testing.T) {
    for i := 0; i < 100; i++ {
        t.Run("Test_Seq_ParallelOrdered2", Test_Seq_ParallelOrdered2)
    }
}

func Test_Seq_ParallelOrdered2(t *testing.T) {
    start := time.Now()
    it := IteratorInt()
    var count int32
    n := rand.Intn(50) + 20
    var maxConcurrent int32
    var nowConcurrent int32
    var nowIndex int32
    var maxDifference int
    lock := sync.Mutex{}
    concurrent := 3 + int(float64(n/10+rand.Intn(n-n/3))*0.9)
    //t.Logf("n:%d,concurrent:%d,n:%d", n, concurrent, n)
    FromIntSeq().Take(n).MapParallel(func(i int) any {
        atomic.AddInt32(&nowIndex, 1)
        c := atomic.AddInt32(&nowConcurrent, 1)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        s := 6*time.Millisecond + time.Duration(rand.Intn(20000))*time.Microsecond
        //t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrent, -1)
        c = atomic.LoadInt32(&nowConcurrent)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        //t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 2, concurrent).ForEach(func(ia any) {
        count++
        atomic.AddInt32(&nowConcurrent, -1)
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
        time.Sleep(time.Millisecond * 2)
        c := int(nowIndex) - i
        if c > maxDifference {
            maxDifference = c
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
    if maxDifference <= concurrent {
        t.Log("maxDifference:", maxDifference, "concurrent:", concurrent)
        t.Fail()
    }
}

func Test_Seq_ParallelOrdered3(t *testing.T) {
    preTest(t)
    start := time.Now()
    it := IteratorInt()
    var count int32
    n := rand.Intn(100) + 20
    var maxConcurrent int32
    var nowConcurrent int32
    lock := sync.Mutex{}
    concurrent := 2 + int(float64(n/10+rand.Intn(n-n/10))*0.9)
    var nowIndex int32
    FromIntSeq().Take(n).MapParallel(func(i int) any {
        atomic.AddInt32(&nowIndex, 1)
        c := atomic.AddInt32(&nowConcurrent, 1)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        s := 5*time.Millisecond + time.Duration(rand.Intn(10000))*time.Microsecond
        //t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
        time.Sleep(s)
        atomic.AddInt32(&nowConcurrent, -1)
        //t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
        return i
    }, 3, concurrent).ForEach(func(ia any) {
        count++
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
        s := time.Duration(rand.Intn(100)) * time.Microsecond
        time.Sleep(s)
        if int(nowIndex) > concurrent+ia.(int) {
            t.Fail()
        }
    })
    t.Logf("ok,use %s, n:%d,concurrent:%d,maxConcurrent:%d", time.Now().Sub(start).String(), n, concurrent, maxConcurrent)
    if count != int32(n) {
        t.Log("count:", count, "n:", n)
        t.Fail()
    }
}

func Test_Seq_MergeBiInt(t *testing.T) {
    s := MergeBiInt(FromIntSeq().Take(1000), IteratorInt(111)).Cache()
    {
        it := IteratorInt()
        FromBiV(s).ForEach(func(i int) {
            i2, _ := it()
            if i != i2 {
                t.Fail()
            }
        })
    }
    {
        it := IteratorInt(111)
        FromBiK(s).ForEach(func(i int) {
            i2, _ := it()
            if i != i2 {
                t.Fail()
            }
        })
    }
}

func Test_Seq_MapFlat(t *testing.T) {
    testI := 0
    testRounds := 0
    MapFlatInt(FromIntSeq().Take(100), func(i int) Seq[int] {
        return FromIntSeq(i).Take(10)
    }).ForEach(func(i int) {
        if testRounds+testI != i {
            t.FailNow()
        }
        testI++
        if testI == 10 {
            testI = 0
            testRounds++
        }
    })
}
