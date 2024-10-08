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

func Test_Seq_Parallel(t *testing.T) {
    preTest(t)
    seq := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    go func() {
        now := time.Now()
        seq.Parallel().ForEach(func(i int) {
            time.Sleep(allSleepDuration)
        })
        sub := time.Now().Sub(now)
        if sub < allSleepDuration || sub.Truncate(allSleepDuration) != allSleepDuration {
            t.Fail()
        }
    }()
    now := time.Now()
    seq.Parallel().Map(func(i int) any {
        time.Sleep(allSleepDuration)
        return i
    }).ForEach(func(a any) {})
    sub := time.Now().Sub(now)
    if sub < allSleepDuration || sub.Truncate(allSleepDuration) != allSleepDuration {
        t.Fail()
    }
}

func Test_Seq_ParallelN(t *testing.T) {
    preTest(t)
    n := 30 + rand.Intn(1000)
    seq := FromIntSeq().Take(n)
    now := time.Now()
    concurrent := int(float64(n/10+rand.Intn(n-10)) * 0.9)
    var maxConcurrent int32
    var nowConcurrent int32
    lock := sync.Mutex{}
    seq.Parallel(concurrent).ForEach(func(i int) {
        c := atomic.AddInt32(&nowConcurrent, 1)
        if c > atomic.LoadInt32(&maxConcurrent) {
            lock.Lock()
            x := atomic.LoadInt32(&maxConcurrent)
            if x <= c {
                maxConcurrent = c
            }
            lock.Unlock()
        }
        time.Sleep(time.Duration(float64(allSleepDuration) / (float64(n) / float64(concurrent))))
        atomic.AddInt32(&nowConcurrent, -1)
    })
    sub := time.Now().Sub(now)
    if sub < allSleepDuration || sub > 3*allSleepDuration {
        t.Fail()
    }
    if maxConcurrent != int32(concurrent) {
        t.Log("maxConcurrent:", maxConcurrent, "concurrent:", concurrent)
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
    joinString := FromIntSeq(1).Take(10).Sort(func(i, j int) bool {
        return i > j
    }).JoinString(",")
    if "10,9,8,7,6,5,4,3,2,1" != joinString {
        t.Fatal(joinString)
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
func Test_Seq_Recover(t *testing.T) {
    BiCastAnyK[int](BiMapK(MapBiInt(FromIntSeq().RecoverErr(func(a any) {
        t.Log("recover1", a)
        panic(a)
    }).Parallel(), func(i int) int {
        return i
    }), func(_ int, i int) any {
        return i
    }).Finally(func() {
        t.Log("finally1")
    }).RecoverErr(func(a any) {
        t.Log("recover2", a)
    }).Finally(func() {
        t.Log("finally2")
    })).ForEach(func(i, _ int) {
        if i > 10 {
            panic("stop")
        }
    })
    t.Log("ok~")
    biSeq1 := BiMapExchangeKV(
        BiMapK(BiMapExchangeKV(
            MapBiInt(
                FromIntSeq().MapParallel(func(i int) any {
                    return i
                }).RecoverErr(func(a any) {
                    t.Log("recover1", a)
                    panic(a)
                }).Parallel().RecoverErr(func(a any) {
                    t.Log("recover2", a)
                    panic(a)
                }).ParallelCustomize(func(i any, f func()) {
                    go f()
                }).RecoverErr(func(a any) {
                    t.Log("recover3", a)
                    panic(a)
                }).MapInt(func(a any) int {
                    return a.(int)
                }), func(i int) int {
                    return i
                },
            ).MapVParallel(func(k, v int) any {
                return v
            }),
        ).RecoverErr(func(a any) {
            t.Log("recover4", a)
            panic(a)
        }), func(a any, i int) int {
            return a.(int)
        }),
    )
    BiCastAnyV[int](BiMapV(biSeq1, func(_ int, i int) any {
        return i
    }).RecoverErrWithValue(func(a int, _, _ any) {
        t.Log("recover5", a)
        panic(a)
    }).Finally(func() {
        t.Log("finally1")
    }).RecoverErr(func(a any) {
        t.Log("recover6", a)
    }).Finally(func() {
        t.Log("finally2")
    })).ForEach(func(i, _ int) {
        if i > 10 {
            panic("stop")
        }
    })
}
