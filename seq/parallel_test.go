package seq

import (
	"sync/atomic"
	"testing"
)

func Test_Parallel(t *testing.T) {
	t.Parallel()
	//门闩同步:worker并发达到上限后统一放行,验证NewParallel并发度,任务阻塞保持并发无需sleep
	const concurrent = 4
	const taskCount = 50
	p, err := NewParallel(concurrent)
	if err != nil {
		t.Fatal(err)
	}
	var nowC, maxC int32
	gate := waitPeak(concurrent, &nowC, awaitTimeout)
	msg := awaitResult(awaitTimeout, func() {
		for i := 0; i < taskCount; i++ {
			p.Add(func() {
				c := atomic.AddInt32(&nowC, 1)
				recordPeak(&maxC, c)
				<-gate
				atomic.AddInt32(&nowC, -1)
			})
		}
		p.Wait()
	})
	if msg != "ok" {
		t.Fatalf("NewParallel执行异常:%s", msg)
	}
	if atomic.LoadInt32(&maxC) != concurrent {
		t.Fatalf("并发峰值 %d != %d", maxC, concurrent)
	}
}
