package seq

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}
func preTest(t *testing.T) {
    t.Parallel()
}

func Test_1(t *testing.T) {
    seq := FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
    ok1 := 0
    ok2 := 0
    ok3 := 0
    ok4 := 1
    CastAnyT(
        seq.OnEach(func(i int) {
            ok1++
        }).Take(50).Filter(func(i int) bool {
            return i%2 == 0
        }).OnEach(func(i int) {
            ok2++
        }).MapFlat(func(i int) Seq[any] {
            return FromSlice([]any{i, i + 1})
        }), 0,
    ).ForEach(func(i int) {
        ok3++
        ok4++
        if ok4 != i {
            t.Fail()
        }
    })
    if ok1 != 10 || ok2 != 5 || ok3 != 10 {
        t.Fail()
    }
    ok4 = 0
    seq.ForEach(func(i int) {
        ok4++
        if ok4 != i {
            t.Fail()
        }
    })
}

func Test_FromIntSeq(t *testing.T) {
    seq := FromIntSeq(1, 10)
    ok := 0
    seq.ForEach(func(i int) {
        ok++
    })
    if ok != 10 {
        t.Fail()
    }
}

func Test_Take(t *testing.T) {
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

func Test_Drop(t *testing.T) {
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

func Test_DropTake(t *testing.T) {
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

func Test_Rand(t *testing.T) {
    slice := FromRandIntSeq().OnEach(func(i int) {
        fmt.Printf("%d,", i)
    }).Filter(func(i int) bool {
        return i%2 == 0
    }).Drop(10).Take(5).ToSlice()
    if len(slice) != 5 {
        t.Fail()
        println("len(slice) != 5")
    }
    fmt.Println(slice)
}

func Test_Seq_Complete(t *testing.T) {
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
