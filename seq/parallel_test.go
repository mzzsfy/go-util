package seq

import (
    "math/rand"
    "testing"
    "time"
)

func Test_Parallel(t *testing.T) {
    n := 30 + rand.Intn(1000)
    duration := time.Millisecond * 2000
    concurrent := 5 + rand.Intn(n-1)/(rand.Intn(10)+1)
    p := NewParallel(concurrent)
    now := time.Now()
    t.Logf("开始,concurrent=%d,n=%d", concurrent, n)
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
