package seq

import (
    "math/rand"
    "testing"
    "time"
)

func Test_Parallel(t *testing.T) {
    FromIntSeq().Take(10).Parallel().ForEach(func(i int) {
        n := 30 + rand.Intn(1000)
        duration := time.Millisecond * 2000
        concurrent := 1 + rand.Intn(n-1)/(rand.Intn(10)+2)
        p := NewParallel(concurrent)
        now := time.Now()
        t.Logf("%d,开始,concurrent=%d,n=%d", i, concurrent, n)
        for i := 0; i < n; i++ {
            //i := i
            p.Add(func() {
                d := duration / time.Duration(n/concurrent)
                //t.Logf("%d sleep %s", i, d.String())
                time.Sleep(d)
            })
        }
        p.Wait()
        sub := time.Now().Sub(now)
        if sub < duration || sub.Truncate(duration) != duration {
            t.Log("运行时间不正确", i, duration.String(), sub.String())
            t.FailNow()
        } else {
            t.Log("ok,use ", sub.String())
        }
    })
}
