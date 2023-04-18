package seq

import (
    "testing"
    "time"
)

func Test_Bi1(t *testing.T) {
    seq := BiFrom(func(k func(int, int)) {
        FromIntSeq(1, 10).ForEach(func(i int) {
            k(i, i+1)
        })
    })
    ok1 := 0
    ok2 := 0
    ok3 := 0
    seq.OnEach(func(i int, j int) {
        ok1++
    }).Filter(func(i int, j int) bool {
        return i%2 == 0
    }).OnEach(func(i int, j int) {
        ok2++
    }).MapFlat(func(i int, j int) BiSeq[any, any] {
        return func(f func(any, any)) {
            f(i, j)
            f(i+1, j+1)
        }
    }).ForEach(func(i any, j any) {
        t.Log(i.(int), j.(int))
        ok3++
    })
    if ok1 != 10 || ok2 != 5 || ok3 != 10 {
        t.Logf("ok1=%d, ok2=%d, ok3=%d", ok1, ok2, ok3)
        t.Fail()
    }
}
func Test_BiAsync(t *testing.T) {
    duration := time.Millisecond * 100
    go func() {
        seq := BiFrom(func(k func(int, int)) { FromIntSeq(1, 10).ForEach(func(i int) { k(i, i+1) }) })
        now := time.Now()
        seq.AsyncEach(func(i int, j int) {
            time.Sleep(duration)
        })
        sub := time.Now().Sub(now)
        if sub < duration || sub.Truncate(duration) != duration {
            t.Fail()
        }
    }()
    seq := BiFrom(func(k func(int, int)) { FromIntSeq(1, 10).ForEach(func(i int) { k(i, i+1) }) })
    now := time.Now()
    seq.Parallel().Map(func(i int, j int) (any, any) {
        time.Sleep(duration)
        return i, j
    }).Complete()
    sub := time.Now().Sub(now)
    if sub < duration || sub.Truncate(duration) != duration {
        t.Fail()
    }
}

func Test_BiTake(t *testing.T) {
    seq := BiFrom(func(k func(int, int)) { FromIntSeq().ForEach(func(i int) { k(i, i+1) }) })

    var r []int
    seq.Take(5).ForEach(func(i int, j int) {
        r = append(r, i)
    })
    if len(r) != 5 {
        t.Fail()
    }
    for i := 0; i < 5; i++ {
        if r[i] != i {
            t.Fail()
        }
    }
}

func Test_BiDrop(t *testing.T) {
    seq := BiFrom(func(k func(int, int)) { FromIntSeq(0, 9).ForEach(func(i int) { k(i, i+1) }) })

    var r []int
    seq.Drop(5).ForEach(func(i int, j int) { r = append(r, i) })
    if len(r) != 5 {
        t.Fail()
    }
    for i := 0; i < 5; i++ {
        if r[i] != i+5 {
            t.Fail()
        }
    }
}

func Test_BiDropTake(t *testing.T) {
    seq := BiFrom(func(k func(int, int)) { FromIntSeq().ForEach(func(i int) { k(i, i+1) }) })

    var r []int
    seq.Drop(5).Take(5).ForEach(func(i int, j int) {
        r = append(r, i)
    })
    if len(r) != 5 {
        t.Fail()
    }
    for i := 0; i < 5; i++ {
        if r[i] != i+5 {
            t.Fail()
        }
    }
}
