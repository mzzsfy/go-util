package seq

import (
    "math/rand"
    "testing"
    "time"
)

func Test_Parallel(t *testing.T) {
    preTest(t)
    n := 30 + rand.Intn(1000)
    duration := time.Millisecond * 800
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
        t.Fail()
    }
    println("ok,use ", sub.String())
}
