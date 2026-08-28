package seq

import (
	"runtime"
	"sync"
	"sync/atomic"
)

//======增强,不改变内容========

// RecoverErr defer recover 的简单封装,在发生panic时,会调用f函数,任何位置可用
func (t Seq[T]) RecoverErr(f func(any)) Seq[T] {
	return func(c func(T)) {
		defer func() {
			if a := recover(); a != nil {
				if a == &stop {
					panic(a)
				}
				f(a)
			}
		}()
		t(c)
	}
}

////Deprecated: 不要使用这个方法,方法名称有歧义,请使用 RecoverErr
//func (t Seq[T]) Catch(f func(any)) Seq[T] {
//    return t.RecoverErr(f)
//}

// RecoverErrWithValue defer recover 的简单封装,保留最后一次调用的值
func (t Seq[T]) RecoverErrWithValue(f func(T, any)) Seq[T] {
	return func(c func(T)) {
		var last T
		defer func() {
			if a := recover(); a != nil {
				if a == &stop {
					panic(a)
				}
				f(last, a)
			}
		}()
		t(func(t T) {
			last = t
			c(t)
		})
	}
}

//// Deprecated: 不要使用这个方法,方法名称有歧义,请使用 RecoverErrWithValue
//func (t Seq[T]) CatchWithValue(f func(T, any)) Seq[T] {
//    return t.RecoverErrWithValue(f)
//}

// Finally defer 的简单封装
func (t Seq[T]) Finally(f func()) Seq[T] {
	return func(c func(T)) {
		defer f()
		t(func(t T) { c(t) })
	}
}

// OnEach 每个元素额外执行
func (t Seq[T]) OnEach(f func(T)) Seq[T] {
	return func(c func(T)) {
		t(func(t T) {
			f(t)
			c(t)
		})
	}
}

// OnEachF 每个元素根据f函数判断是否需要额外执行
func (t Seq[T]) OnEachF(step func(T) bool, f func(T), skip ...int) Seq[T] {
	return func(c func(T)) {
		x := 0
		if len(skip) > 0 {
			x = -skip[0]
		}
		t(func(t T) {
			x++
			if x > 0 && step(t) {
				f(t)
			}
			c(t)
		})
	}
}

// OnEachN 每n个元素额外执行
func (t Seq[T]) OnEachN(step int, f func(T), skip ...int) Seq[T] {
	if step <= 0 {
		panic("step must > 0")
	}
	return func(c func(T)) {
		x := 0
		if len(skip) > 0 {
			x = -skip[0]
		}
		t(func(t T) {
			x++
			if x > 0 && x%step == 0 {
				f(t)
			}
			c(t)
		})
	}
}

// OnEachNX 每n个元素额外执行一次,当结束时,如果剩余元素不足n个,额外执行一次
// 注意:并行消费(Parallel)之后使用本方法,last 记录位于并行分支内,存在数据竞争,禁止该组合
func (t Seq[T]) OnEachNX(step int, f func(T), skip ...int) Seq[T] {
	if step <= 0 {
		panic("step must > 0")
	}
	return func(c func(T)) {
		x := 0
		if len(skip) > 0 {
			x = -skip[0]
		}
		var last T
		exist := false
		t(func(t T) {
			x++
			last = t
			exist = true
			if x > 0 && x%step == 0 {
				f(t)
			}
			c(t)
		})
		//空流不触发收尾,避免旧实现在 skip 导致 x%step!=0 时的nil解引用
		if exist && x%step != 0 {
			f(last)
		}
	}
}

// OnBefore 指定位置前(包含),每个元素额外执行
func (t Seq[T]) OnBefore(i int, f func(T)) Seq[T] {
	return func(c func(T)) {
		x := 0
		t(func(t T) {
			if x < i {
				x++
				f(t)
			}
			c(t)
		})
	}
}

// OnAfter 指定位置后(包含),每个元素额外执行
func (t Seq[T]) OnAfter(i int, f func(T)) Seq[T] {
	return func(c func(T)) {
		x := 0
		t(func(t T) {
			if x >= i {
				f(t)
			} else {
				x++
			}
			c(t)
		})
	}
}

// OnFirst 执行前额外执行
func (t Seq[T]) OnFirst(f func(T)) Seq[T] {
	return func(c func(T)) {
		x := 0
		t(func(t T) {
			if x == 0 {
				x++
				f(t)
			}
			c(t)
		})
	}
}

// OnLast 执行完成后额外执行,f 的参数为最后一个元素,空流时为nil
// 注意:并行消费(Parallel)之后使用本方法,last 记录位于并行分支内,存在数据竞争,禁止该组合
// 并行流的收尾场景应使用 Finally,其在所有并行任务完成后于调用方goroutine执行
func (t Seq[T]) OnLast(f func(*T)) Seq[T] {
	return func(c func(T)) {
		var last T
		exist := false
		t(func(t T) {
			last = t
			exist = true
			c(t)
		})
		if exist {
			f(&last)
		} else {
			f(nil)
		}
	}
}

// Sync 串行执行
func (t Seq[T]) Sync() Seq[T] {
	return func(c func(T)) {
		lock := sync.Mutex{}
		t(func(t T) {
			lock.Lock()
			defer lock.Unlock()
			c(t)
		})
	}
}

// Parallel 对后续操作启用并行执行,使用 Sync() 保证消费不竞争
// concurrent 为空时默认并发数为 2*GOMAXPROCS,防止 goroutine 洪泛
// 注意:Parallel 之后的链上不得使用 OnLast/OnEachNX,其内部状态写入位于并行分支,存在数据竞争
// 需要"最后一个元素"语义时将 OnLast 置于 Parallel 之前,需要收尾语义时使用 Finally
func (t Seq[T]) Parallel(concurrent ...int) Seq[T] {
	sl := runtime.GOMAXPROCS(0) * 2
	if len(concurrent) > 0 && concurrent[0] > 0 {
		sl = concurrent[0]
	}
	return func(c func(T)) {
		p := newParallel(sl)
		//生产方panic中止时兜底回收worker,正常路径Wait已回收,幂等
		defer p.abort()
		t(func(t T) {
			p.Add(func() {
				c(t)
			})
		})
		p.Wait()
	}
}

// ParallelCustomize 自定义并行执行策略
func (t Seq[T]) ParallelCustomize(fn func(T, func())) Seq[T] {
	return func(c func(T)) {
		wg := sync.WaitGroup{}
		//err保护与hasErr读取分离,hasErr供生产方快速失败无锁检查,与parallel实现一致
		var errMu sync.Mutex
		var err any
		var hasErr int32
		//hasStop标记消费者提前停止,哨兵不入错误,停止语义为正常返回
		var hasStop int32
		t(func(t T) {
			wg.Add(1)
			fn(t, func() {
				defer func() {
					if a := recover(); a != nil {
						//消费者提前停止的哨兵不入错误,标记停止供生产方快速失败
						if a == &stop {
							atomic.StoreInt32(&hasStop, 1)
						} else if atomic.CompareAndSwapInt32(&hasErr, 0, 1) {
							errMu.Lock()
							err = a
							errMu.Unlock()
						}
					}
					wg.Done()
				}()
				c(t)
			})
			//停止或失败均中止生产,停止哨兵由源头stopRecover吸收
			if atomic.LoadInt32(&hasStop) == 1 {
				panic(&stop)
			}
			if atomic.LoadInt32(&hasErr) == 1 {
				errMu.Lock()
				e := err
				errMu.Unlock()
				panic(e)
			}
		})
		wg.Wait()
		//生产已结束,失败在Wait后重抛,避免静默吞掉;停止哨兵不入错误,无需重抛
		if atomic.LoadInt32(&hasErr) == 1 {
			errMu.Lock()
			e := err
			errMu.Unlock()
			panic(e)
		}
	}
}

// Sort 排序,禁止无限流,会导致内存溢出
func (t Seq[T]) Sort(less func(T, T) bool) Seq[T] {
	var r []T
	once := sync.Once{}
	fn := func() {
		defer stopRecover()
		t(func(t T) { r = append(r, t) })
		sortSlice(r, less)
	}
	return func(t func(T)) {
		once.Do(fn)
		for _, v := range r {
			t(v)
		}
	}
}

// SortCustomize 自定义排序,禁止无限流,会导致内存溢出
func (t Seq[T]) SortCustomize(sort func([]T)) Seq[T] {
	var r []T
	once := sync.Once{}
	fn := func() {
		defer stopRecover()
		t(func(t T) { r = append(r, t) })
		sort(r)
	}
	return func(t func(T)) {
		once.Do(fn)
		for _, v := range r {
			t(v)
		}
	}
}

// Reverse 逆序,禁止无限流,会导致内存溢出
func (t Seq[T]) Reverse() Seq[T] {
	var r []T
	once := sync.Once{}
	fn := func() {
		defer stopRecover()
		t(func(t T) { r = append(r, t) })
	}
	return func(t func(T)) {
		once.Do(fn)
		for i := len(r) - 1; i >= 0; i-- {
			t(r[i])
		}
	}
}

// Cache 缓存Seq,使该Seq可以多次重复消费,init为true时,会立刻触发消费行为
func (t Seq[T]) Cache(init ...bool) Seq[T] {
	var r []T
	once := sync.Once{}
	fn := func() {
		defer stopRecover()
		t(func(t T) { r = append(r, t) })
	}
	if len(init) > 0 && init[0] {
		once.Do(fn)
	}
	return func(t func(T)) {
		defer stopRecover()
		once.Do(fn)
		for _, v := range r {
			t(v)
		}
	}
}

// Repeat 重复该Seq n次,如果不传递n,则无限重复,当前seq如果比较重,建议先使用 Cache 缓存
// 标准构造源会吸收消费者的Stop哨兵后正常返回,Repeat须识别哨兵并停止重启源,否则无限重复+提前终止会死循环
// 禁止置于 Parallel 之后使用,Repeat 内部状态与重复推进仅支持串行消费(与 OnLast 同类约束)
func (t Seq[T]) Repeat(n ...int) Seq[T] {
	return func(f func(T)) {
		stopped := false
		wf := func(v T) {
			defer func() {
				if a := recover(); a != nil {
					if a == &stop {
						stopped = true
					}
					//哨兵与错误继续向上传播,由源的stopRecover或本函数兜底吸收
					panic(a)
				}
			}()
			f(v)
		}
		defer stopRecover()
		if len(n) == 0 {
			for {
				t(wf)
				if stopped {
					return
				}
			}
		} else {
			l := n[0]
			for i := 0; i < l; i++ {
				t(wf)
				if stopped {
					return
				}
			}
		}
	}
}
