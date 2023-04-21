package seq

import (
    "math/rand"
    "testing"
    "time"
)

const allSleepDuration = time.Millisecond * 2000

func Test_Parallel(t *testing.T) {
    preTest(t)
    FromIntSeq(1).Take(1).ForEach(func(x int) {
        now := time.Now()
        FromIntSeq().Take(20).Parallel().OnLast(func(i *int) {
            t.Logf("%d ok,use %s", x, time.Now().Sub(now).String())
        }).ForEach(func(i int) {
            n := 30 + rand.Intn(10000)
            concurrent := int(float64(n/10+rand.Intn(n-10)) * 0.9)
            p := NewParallel(concurrent)
            sleepDuration := time.Duration(float64(allSleepDuration) / (float64(n) / float64(concurrent)))
            //t.Logf("%d,开始,concurrent=%d,n=%d", i, concurrent, n)
            for i := 0; i < n; i++ {
                //i := i
                p.Add(func() {
                    d := sleepDuration
                    //t.Logf("%d sleep %s", i, d.String())
                    time.Sleep(d)
                })
            }
            p.Wait()
            sub := time.Now().Sub(now)
            if sub < allSleepDuration || sub > 3*allSleepDuration {
                t.Logf("%d,运行时间不正确%s,%s,concurrent=%d,n=%d,sleepDuration=%s", i, allSleepDuration.String(), sub.String(), concurrent, n, sleepDuration.String())
                t.FailNow()
            } else {
                //t.Log("ok,use ", i, sub.String())
            }
        })
    })
}
