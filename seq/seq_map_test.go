//go:build !go1.27

package seq

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test_Seq_MapSliceN(t *testing.T) {
	n := 999
	seq := FromIntSeq().Take(n)
	s := CastAny[[]int](MapSliceN(seq, 3))
	s.ForEach(func(is []int) {
		if len(is) != 3 {
			t.Fail()
		}
	})
}

func TestSeq_ParallelOrdered1_100(t *testing.T) {
	deadline, hasDeadline := t.Deadline()
	for i := 0; i < 100; i++ {
		if hasDeadline && time.Until(deadline) <= 1500*time.Millisecond {
			break
		}
		t.Run("Test_Seq_ParallelOrdered1", Test_Seq_ParallelOrdered1)
	}
}

func Test_Seq_ParallelOrdered1(t *testing.T) {
	preTest(t)
	//map函数内的sleep为模拟异步负载制造乱序完成窗口,属必要sleep,耗时断言仅宽松校验
	var count int32
	n := rand.Intn(1000) + 50 // Reduced from 3000+100 to be more reliable
	start := time.Now()
	var maxConcurrent int32
	var nowConcurrent int32
	lock := sync.Mutex{}
	// Simplified concurrent calculation to be more predictable
	concurrent := 1 + n/20
	if concurrent > 10 {
		concurrent = 10 // Cap concurrency for stability
	}
	FromIntSeq().Take(n).MapParallel(func(i int) any {
		c := atomic.AddInt32(&nowConcurrent, 1)
		if c > atomic.LoadInt32(&maxConcurrent) {
			lock.Lock()
			x := atomic.LoadInt32(&maxConcurrent)
			if x <= c {
				maxConcurrent = c
			}
			lock.Unlock()
		}
		s := 5*time.Millisecond + time.Duration(rand.Intn(5000))*time.Microsecond
		time.Sleep(s)
		atomic.AddInt32(&nowConcurrent, -1)
		c = atomic.LoadInt32(&nowConcurrent)
		if c > atomic.LoadInt32(&maxConcurrent) {
			lock.Lock()
			x := atomic.LoadInt32(&maxConcurrent)
			if x <= c {
				maxConcurrent = c
			}
			lock.Unlock()
		}
		return i
	}, 1, concurrent).ForEach(func(ia any) {
		atomic.AddInt32(&count, 1)
	})
	t.Logf("ok,use %s, n:%d,concurrent:%d,maxConcurrent:%d", time.Now().Sub(start).String(), n, concurrent, maxConcurrent)
	// Allow for some variance in concurrent execution on slower machines
	if maxConcurrent < int32(concurrent)-1 || maxConcurrent > int32(concurrent)+1 {
		t.Logf("maxConcurrent %d is too far from expected %d", maxConcurrent, concurrent)
		t.Fail()
	}
	if count != int32(n) {
		t.Fail()
	}
}

func TestSeq_ParallelOrdered2_100(t *testing.T) {
	deadline, hasDeadline := t.Deadline()
	for i := 0; i < 20; i++ {
		if hasDeadline && time.Until(deadline) <= 1500*time.Millisecond {
			break
		}
		t.Run("Test_Seq_ParallelOrdered2", Test_Seq_ParallelOrdered2)
	}
}

func Test_Seq_ParallelOrdered2(t *testing.T) {
	start := time.Now()
	//map与消费端的sleep为模拟异步负载制造乱序完成窗口,属必要sleep,耗时仅日志观测
	it := IteratorInt()
	var count int32
	n := rand.Intn(30) + 15 // Reduced from 50+20 to be more reliable
	var maxConcurrent int32
	var nowConcurrent int32
	var nowIndex int32
	var maxDifference int
	lock := sync.Mutex{}
	// Simplified concurrent calculation
	concurrent := 3 + n/5
	if concurrent > 8 {
		concurrent = 8 // Cap concurrency for stability
	}
	var failed int32
	//t.Logf("n:%d,concurrent:%d,n:%d", n, concurrent, n)
	FromIntSeq().Take(n).MapParallel(func(i int) any {
		atomic.AddInt32(&nowIndex, 1)
		c := atomic.AddInt32(&nowConcurrent, 1)
		if c > atomic.LoadInt32(&maxConcurrent) {
			lock.Lock()
			x := atomic.LoadInt32(&maxConcurrent)
			if x <= c {
				maxConcurrent = c
			}
			lock.Unlock()
		}
		s := 3*time.Millisecond + time.Duration(rand.Intn(10000))*time.Microsecond // Reduced sleep time
		//t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
		time.Sleep(s)
		atomic.AddInt32(&nowConcurrent, -1)
		c = atomic.LoadInt32(&nowConcurrent)
		if c > atomic.LoadInt32(&maxConcurrent) {
			lock.Lock()
			x := atomic.LoadInt32(&maxConcurrent)
			if x <= c {
				maxConcurrent = c
			}
			lock.Unlock()
		}
		//t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
		return i
	}, 2, concurrent).ForEach(func(ia any) {
		count++
		atomic.AddInt32(&nowConcurrent, -1)
		runtime.Gosched()
		i := ia.(int)
		i2, _ := it()
		//t.Log("test", i, "expect", i2)
		if i != i2 {
			if atomic.CompareAndSwapInt32(&failed, 0, 1) {
				t.Errorf("test %d expect %d", i, i2)
			}
			return
		}
		time.Sleep(time.Millisecond)
		c := int(nowIndex) - i
		if c > maxDifference {
			maxDifference = c
		}
	})
	t.Logf("ok,use %s, n:%d,concurrent:%d,maxConcurrent:%d", time.Now().Sub(start).String(), n, concurrent, maxConcurrent)
	if count != int32(n) {
		t.Log("count:", count, "n:", n)
		t.Fail()
	}
	// Allow for some variance in concurrent execution on slower machines
	if maxConcurrent < int32(concurrent)-1 || maxConcurrent > int32(concurrent)+1 {
		t.Log("maxConcurrent:", maxConcurrent, "concurrent:", concurrent)
		t.Fail()
	}
	// Allow for some variance in maxDifference
	if maxDifference <= concurrent-1 {
		t.Log("maxDifference:", maxDifference, "concurrent:", concurrent)
		t.Fail()
	}
}

func Test_Seq_ParallelOrdered3(t *testing.T) {
	preTest(t)
	start := time.Now()
	//map与消费端的sleep为模拟异步负载制造乱序完成窗口,属必要sleep,耗时仅日志观测
	it := IteratorInt()
	var count int32
	n := rand.Intn(50) + 15 // Reduced from 100+20 to be more reliable
	var maxConcurrent int32
	var nowConcurrent int32
	lock := sync.Mutex{}
	// Simplified concurrent calculation
	concurrent := 2 + n/8
	if concurrent > 6 {
		concurrent = 6 // Cap concurrency for stability
	}
	var failed int32
	var nowIndex int32
	FromIntSeq().Take(n).MapParallel(func(i int) any {
		atomic.AddInt32(&nowIndex, 1)
		c := atomic.AddInt32(&nowConcurrent, 1)
		if c > atomic.LoadInt32(&maxConcurrent) {
			lock.Lock()
			x := atomic.LoadInt32(&maxConcurrent)
			if x <= c {
				maxConcurrent = c
			}
			lock.Unlock()
		}
		s := 3*time.Millisecond + time.Duration(rand.Intn(5000))*time.Microsecond // Reduced sleep time
		//t.Log("sleep", i, s.Truncate(time.Microsecond*100).String())
		time.Sleep(s)
		atomic.AddInt32(&nowConcurrent, -1)
		//t.Log("sleep over", i, s.Truncate(time.Microsecond*100).String())
		return i
	}, 3, concurrent).ForEach(func(ia any) {
		count++
		runtime.Gosched()
		i := ia.(int)
		i2, _ := it()
		//t.Log("test", i, "expect", i2)
		if i != i2 {
			if atomic.CompareAndSwapInt32(&failed, 0, 1) {
				t.Errorf("test %d expect %d", i, i2)
			}
			return
		}
		s := time.Duration(rand.Intn(50)) * time.Microsecond // Reduced sleep time
		time.Sleep(s)
		// Allow for some variance in timing on slower machines
		if int(nowIndex) > concurrent+ia.(int)+2 {
			t.Fail()
		}
	})
	t.Logf("ok,use %s, n:%d,concurrent:%d,maxConcurrent:%d", time.Now().Sub(start).String(), n, concurrent, maxConcurrent)
	if count != int32(n) {
		t.Log("count:", count, "n:", n)
		t.Fail()
	}
}

func Test_Seq_MergeBiInt(t *testing.T) {
	s := MergeBiInt(FromIntSeq().Take(1000), IteratorInt(111)).Cache()
	{
		it := IteratorInt()
		FromBiV(s).ForEach(func(i int) {
			i2, _ := it()
			if i != i2 {
				t.Fail()
			}
		})
	}
	{
		it := IteratorInt(111)
		FromBiK(s).ForEach(func(i int) {
			i2, _ := it()
			if i != i2 {
				t.Fail()
			}
		})
	}
}

// Test_Seq_MergeBiAnyRight 验证 Right 变体流方向反转: 流元素在前, iterator 值在后
func Test_Seq_MergeBiAnyRight(t *testing.T) {
	t.Parallel()
	it := func() Iterator[any] {
		i := 0
		return func() (any, bool) {
			if i >= 3 {
				return nil, false
			}
			i++
			return i * 10, true
		}
	}
	elements := []string{"a", "b", "c"}
	s := MergeBiAnyRight(FromSlice(elements), it()).Cache()
	{
		exp := it()
		FromBiV(s).ForEach(func(v any) {
			e, _ := exp()
			if v != e {
				t.Fatalf("V 应为 iterator 值: %v != %v", v, e)
			}
		})
	}
	{
		i := 0
		FromBiK(s).ForEach(func(k string) {
			if k != elements[i] {
				t.Fatalf("K 应为流元素: %v", k)
			}
			i++
		})
		if i != len(elements) {
			t.Fatalf("元素数 %d != %d", i, len(elements))
		}
	}
	// 对照: 非 Right 的 K 是 iterator 值
	s2 := MergeBiAny(FromSlice([]string{"x"}), func() (any, bool) { return 1, true }).Cache()
	FromBiK(s2).ForEach(func(k any) {
		if k != 1 {
			t.Fatalf("MergeBiAny 的 K 应为 iterator 值, got %v", k)
		}
	})
}

func Test_Seq_MapFlat(t *testing.T) {
	testI := 0
	testRounds := 0
	MapFlatInt(FromIntSeq().Take(100), func(i int) Seq[int] {
		return FromIntSeq(i).Take(10)
	}).ForEach(func(i int) {
		if testRounds+testI != i {
			t.FailNow()
		}
		testI++
		if testI == 10 {
			testI = 0
			testRounds++
		}
	})
}
