//go:build !go1.27

package seq

import (
	"sync/atomic"
	"testing"
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
		//t.Log(i.(int), j.(int))
		ok3++
	})
	if ok1 != 10 || ok2 != 5 || ok3 != 10 {
		t.Logf("ok1=%d, ok2=%d, ok3=%d", ok1, ok2, ok3)
		t.Fail()
	}
}
func Test_BiParallel(t *testing.T) {
	preTest(t)
	//门闩事件同步:并发数达到目标后统一放行,证明真并行,替代固定时序断言
	const concurrent = 4
	var nowC, maxC int32
	gate := waitPeak(concurrent, &nowC, awaitTimeout)
	msg := awaitResult(awaitTimeout, func() {
		seq := BiFrom(func(k func(int, int)) { FromIntSeq(1, 8).ForEach(func(i int) { k(i, i+1) }) })
		seq.Parallel(concurrent).ForEach(func(i int, j int) {
			c := atomic.AddInt32(&nowC, 1)
			recordPeak(&maxC, c)
			<-gate
			atomic.AddInt32(&nowC, -1)
		})
	})
	if msg != "ok" {
		t.Fatalf("BiParallel执行异常:%s", msg)
	}
	if atomic.LoadInt32(&maxC) != concurrent {
		t.Fatalf("并发峰值 %d != %d", maxC, concurrent)
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
