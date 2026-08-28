//go:build !go1.27

package seq

import (
	"math/rand"
	"sync/atomic"
	"testing"
)

func Test_Seq_OnLast(t *testing.T) {
	exec := 0
	FromIntSeq().Take(10).OnLast(func(i *int) {
		if *i != 9 {
			t.Fail()
		}
		exec++
	}).ForEach(func(i int) {})
	if exec != 1 {
		t.Fail()
	}
}

func Test_Seq_Parallel(t *testing.T) {
	preTest(t)
	//门闩事件同步:并发数达到目标后统一放行,证明真并行,替代goroutine双跑与固定时序断言
	const concurrent = 4
	var nowC, maxC int32
	gate := waitPeak(concurrent, &nowC, awaitTimeout)
	msg := awaitResult(awaitTimeout, func() {
		FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).Parallel(concurrent).ForEach(func(i int) {
			c := atomic.AddInt32(&nowC, 1)
			recordPeak(&maxC, c)
			<-gate
			atomic.AddInt32(&nowC, -1)
		})
	})
	if msg != "ok" {
		t.Fatalf("Parallel执行异常:%s", msg)
	}
	if atomic.LoadInt32(&maxC) != concurrent {
		t.Fatalf("并发峰值 %d != %d", maxC, concurrent)
	}
}

func Test_Seq_ParallelN(t *testing.T) {
	preTest(t)
	//门闩同步:任务阻塞至并发达到上限,精确验证并发数控制,无需sleep维持并发
	n := 30 + rand.Intn(200)
	concurrent := 2 + rand.Intn(8)
	var nowC, maxC int32
	gate := waitPeak(int32(concurrent), &nowC, awaitTimeout)
	msg := awaitResult(awaitTimeout, func() {
		FromIntSeq().Take(n).Parallel(concurrent).ForEach(func(i int) {
			c := atomic.AddInt32(&nowC, 1)
			recordPeak(&maxC, c)
			<-gate
			atomic.AddInt32(&nowC, -1)
		})
	})
	if msg != "ok" {
		t.Fatalf("ParallelN执行异常:%s", msg)
	}
	if atomic.LoadInt32(&maxC) != int32(concurrent) {
		t.Fatalf("并发峰值 %d != %d", maxC, concurrent)
	}
}

func Test_Cache(t *testing.T) {
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

func Test_Sort(t *testing.T) {
	//结果 "10,9,8 ... 3,2,1"
	joinString := FromIntSeq(1).Take(10).Sort(func(i, j int) bool {
		return i > j
	}).JoinString(",")
	if "10,9,8,7,6,5,4,3,2,1" != joinString {
		t.Fatal(joinString)
	}
}

func Test_Seq_Repeat(t *testing.T) {
	testI := 0
	repeatI := 0
	FromIntSeq(0, 10).Repeat(3).ForEach(func(i int) {
		if i != testI {
			t.Fail()
		}
		testI++
		if testI > 10 {
			testI = 0
			repeatI++
		}
	})
	if repeatI != 3 {
		t.Fail()
	}
}
func Test_Seq_Recover(t *testing.T) {
	//awaitResult 包裹验证错误链在Stop后能终止,不挂起不死循环
	msg := awaitResult(awaitTimeout, func() {
		BiCastAnyK[int](BiMapK(MapBiInt(FromIntSeq().RecoverErr(func(a any) {
			t.Log("recover1", a)
			panic(a)
		}).Parallel(), func(i int) int {
			return i
		}), func(_ int, i int) any {
			return i
		}).Finally(func() {
			t.Log("finally1")
		}).RecoverErr(func(a any) {
			t.Log("recover2", a)
		}).Finally(func() {
			t.Log("finally2")
		})).ForEach(func(i, _ int) {
			if i > 10 {
				panic("stop")
			}
		})
		t.Log("ok~")
		biSeq1 := BiMapExchangeKV(
			BiMapK(BiMapExchangeKV(
				MapBiInt(
					FromIntSeq().MapParallel(func(i int) any {
						return i
					}).RecoverErr(func(a any) {
						t.Log("recover1", a)
						panic(a)
					}).Parallel().RecoverErr(func(a any) {
						t.Log("recover2", a)
						panic(a)
					}).ParallelCustomize(func(i any, f func()) {
						go f()
					}).RecoverErr(func(a any) {
						t.Log("recover3", a)
						panic(a)
					}).MapInt(func(a any) int {
						return a.(int)
					}), func(i int) int {
						return i
					},
				).MapVParallel(func(k, v int) any {
					return v
				}),
			).RecoverErr(func(a any) {
				t.Log("recover4", a)
				panic(a)
			}), func(a any, i int) int {
				return a.(int)
			}),
		)
		BiCastAnyV[int](BiMapV(biSeq1, func(_ int, i int) any {
			return i
		}).RecoverErrWithValue(func(a int, _, _ any) {
			t.Log("recover5", a)
			panic(a)
		}).Finally(func() {
			t.Log("finally1")
		}).RecoverErr(func(a any) {
			t.Log("recover6", a)
		}).Finally(func() {
			t.Log("finally2")
		})).ForEach(func(i, _ int) {
			if i > 10 {
				panic("stop")
			}
		})
	})
	if msg != "ok" {
		t.Fatalf("Recover错误链未正常终止:%s", msg)
	}
}
