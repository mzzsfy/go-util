package seq

import (
    "math/rand"
    "testing"
    "time"
)

const allSleepDuration = time.Millisecond * 1000

func Test_Parallel(t *testing.T) {
    preTest(t)
    FromIntSeq().Take(10).Parallel().ForEach(func(i int) {
        n := 30 + rand.Intn(10000)
        concurrent := 5 + rand.Intn(int(float64(n)*0.9))
        p := NewParallel(concurrent)
        now := time.Now()
        t.Logf("%d,开始,concurrent=%d,n=%d", i, concurrent, n)
        for i := 0; i < n; i++ {
            //i := i
            p.Add(func() {
                d := time.Duration(float64(allSleepDuration) / (float64(n) / float64(concurrent)))
                //t.Logf("%d sleep %s", i, d.String())
                time.Sleep(d)
            })
        }
        p.Wait()
        sub := time.Now().Sub(now)
        if sub < allSleepDuration || sub.Truncate(allSleepDuration) != allSleepDuration {
            t.Log("运行时间不正确", i, allSleepDuration.String(), sub.String())
            t.FailNow()
        } else {
            t.Log("ok,use ", i, sub.String())
        }
    })
}
