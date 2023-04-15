package seq

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

func Test1(t *testing.T) {
    seq := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    ok1 := 0
    ok2 := 0
    ok3 := 0
    CastAnyT(
        seq.OnEach(func(i int) {
            ok1++
        }).Take(50).Filter(func(i int) bool {
            return i%2 == 0
        }).OnEach(func(i int) {
            ok2++
        }).FlatMap(func(i int) Seq[any] {
            return FromSlice([]any{i, i + 1})
        }), 0,
    ).ForEach(func(i int) {
        t.Log(i)
        ok3++
    })
    if ok1 != 10 || ok2 != 5 || ok3 != 10 {
        t.Fail()
    }
    seq.ForEach(func(i int) {
        t.Log("test", i)
    })
}

func TestAsync(t *testing.T) {
    duration := time.Millisecond * 100
    seq := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    go func() {
        seq := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
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

func TestConcurrencyControl(t *testing.T) {
    n := 100 + rand.Intn(1000)
    seq := FromIntSeq().Take(n)
    now := time.Now()
    duration := time.Millisecond * 100
    concurrency := 1 + rand.Intn(n-1)
    seq.Parallel(concurrency).Map(func(i int) any {
        //println(i, "start")
        time.Sleep(duration / time.Duration(n/concurrency))
        //println(i, "end")
        return i
    }).Complete()
    sub := time.Now().Sub(now)
    if sub < duration || sub.Truncate(duration) != duration {
        t.Fail()
    }
}

func TestFromIntSeq(t *testing.T) {
    seq := FromIntSeq(1, 10)
    ok := 0
    seq.ForEach(func(i int) {
        ok++
    })
    if ok != 10 {
        t.Fail()
    }
}

func TestTake(t *testing.T) {
    seq := FromIntSeq(0, 9)
    var r []int
    seq.Take(5).ForEach(func(i int) { r = append(r, i) })
    if len(r) != 5 {
        t.Fail()
    }
    for i := 0; i < 5; i++ {
        if r[i] != i {
            t.Fail()
        }
    }
}

func TestDrop(t *testing.T) {
    seq := FromIntSeq(0, 9)
    var r []int
    seq.Drop(5).ForEach(func(i int) { r = append(r, i) })
    if len(r) != 5 {
        t.Fail()
    }
    for i := 0; i < 5; i++ {
        if r[i] != i+5 {
            t.Fail()
        }
    }
}

func TestDropTake(t *testing.T) {
    seq := FromIntSeq()
    var r []int
    seq.Drop(5).Take(5).ForEach(func(i int) { r = append(r, i) })
    if len(r) != 5 {
        t.Fail()
    }
    for i := 0; i < 5; i++ {
        if r[i] != i+5 {
            t.Fail()
        }
    }
}

func TestCache(t *testing.T) {
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

func TestRand(t *testing.T) {
    slice := FromRandIntSeq().OnEach(func(i int) {
        fmt.Println("", i)
    }).Filter(func(i int) bool {
        return i%2 == 0
    }).Drop(10).Take(5).ToSlice()
    if len(slice) != 5 {
        t.Fail()
    }
    fmt.Println(slice)
}

func TestSort(t *testing.T) {
    //结果 "10,9,8 ... 3,2,1"
    if "10,9,8,7,6,5,4,3,2,1" != FromIntSeq(1).Take(10).Sort(func(i, j int) bool {
        return i > j
    }).JoinString(",") {
        t.Fail()
    }
}

func TestSeq_Complete(t *testing.T) {
    s := FromIntSeq().Take(1000).MapBiSerialNumber(100).Cache()
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
        it := IteratorInt(100)
        s.SeqK().ForEach(func(i int) {
            i2, _ := it()
            if i != i2 {
                t.Fail()
            }
        })
    }
}
