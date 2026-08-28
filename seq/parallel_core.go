package seq

import (
	"sync"
	"sync/atomic"
)

// resultR 有序归并的结果单元
type resultR[R any] struct {
	id int32
	v  R
}

// resultKV 有序归并的结果单元,双值场景
type resultKV[K, R any] struct {
	id int32
	k  K
	v  R
}

// mapParallelCore 并行转换的有序调度核心,order>0 时使用,order 语义见 MapParallel
func mapParallelCore[T, R any](t Seq[T], syncFn func(T) R, o int, sl int) Seq[R] {
	return func(c func(R)) {
		p := newParallel(sl)
		//生产方panic中止时兜底回收worker,正常路径Wait已回收,幂等
		defer p.abort()
		var id int32
		switch o {
		case 1:
			//完成即消费,消费并发竞争,顺序由调用方Sync保证
			t(func(t T) {
				id++
				p.Add(func() { c(syncFn(t)) })
			})
			p.Wait()
		case 2:
			//缓冲与并发数一致,worker投递结果时减少阻塞
			results := make(chan resultR[R], sl)
			done := make(chan struct{})
			//归并消费独立于生产,按序号输出,乱序结果暂存,单goroutine访问无锁
			go func() {
				defer close(done)
				//下游提前停止时,吸收Stop并丢弃剩余结果,避免worker投递阻塞
				defer func() {
					for range results {
					}
				}()
				defer stopRecover()
				next := int32(1)
				pending := make(map[int32]R)
				for r := range results {
					if r.id != next {
						pending[r.id] = r.v
						continue
					}
					c(r.v)
					next++
					for {
						v, ok := pending[next]
						if !ok {
							break
						}
						delete(pending, next)
						c(v)
						next++
					}
				}
			}()
			func() {
				//兜底关闭保证panic中止时归并goroutine可退出,abort先行确保worker投递完成后才关闭
				defer close(results)
				defer p.abort()
				t(func(t T) {
					id++
					i := id
					p.Add(func() { results <- resultR[R]{i, syncFn(t)} })
				})
				p.Wait()
			}()
			<-done
		default:
			//以消费为准限速,worker持有至前序消费完成
			var currentIndex int32 = 1
			lock := sync.Mutex{}
			cond := sync.NewCond(&lock)
			t(func(t T) {
				id++
				i := id
				p.Add(func() {
					defer func() {
						if r := recover(); r != nil {
							//任务panic仍需轮到自身序号后跳过,否则后续任务永久等待,失败经Wait重抛
							lock.Lock()
							for currentIndex < i {
								cond.Wait()
							}
							currentIndex = i + 1
							lock.Unlock()
							cond.Broadcast()
							panic(r)
						}
					}()
					a := syncFn(t)
					//消费持锁保证顺序,锁经defer归还,消费者panic时锁先行释放,避免recover内重入死锁
					func() {
						lock.Lock()
						defer lock.Unlock()
						for currentIndex != i {
							cond.Wait()
						}
						c(a)
						currentIndex = i + 1
					}()
					cond.Broadcast()
				})
			})
			p.Wait()
		}
	}
}

// biMapVParallelCore 并行转换的有序调度核心,order>0 时使用
func biMapVParallelCore[K, V, R any](t BiSeq[K, V], f func(K, V) R, o int, sl int) BiSeq[K, R] {
	return func(c func(K, R)) {
		p := newParallel(sl)
		//生产方panic中止时兜底回收worker,正常路径Wait已回收,幂等
		defer p.abort()
		var id int32
		switch o {
		case 1:
			t(func(k K, v V) {
				id++
				p.Add(func() { c(k, f(k, v)) })
			})
			p.Wait()
		case 2:
			results := make(chan resultKV[K, R], sl)
			done := make(chan struct{})
			go func() {
				defer close(done)
				//下游提前停止时,吸收Stop并丢弃剩余结果,避免worker投递阻塞
				defer func() {
					for range results {
					}
				}()
				defer stopRecover()
				next := int32(1)
				pending := make(map[int32]resultKV[K, R])
				for r := range results {
					if r.id != next {
						pending[r.id] = r
						continue
					}
					c(r.k, r.v)
					next++
					for {
						g, ok := pending[next]
						if !ok {
							break
						}
						delete(pending, next)
						c(g.k, g.v)
						next++
					}
				}
			}()
			func() {
				//兜底关闭保证panic中止时归并goroutine可退出,abort先行确保worker投递完成后才关闭
				defer close(results)
				defer p.abort()
				t(func(k K, v V) {
					id++
					i := id
					p.Add(func() { results <- resultKV[K, R]{i, k, f(k, v)} })
				})
				p.Wait()
			}()
			<-done
		default:
			var currentIndex int32 = 1
			lock := sync.Mutex{}
			cond := sync.NewCond(&lock)
			t(func(k K, v V) {
				id++
				i := id
				p.Add(func() {
					defer func() {
						if r := recover(); r != nil {
							//任务panic仍需轮到自身序号后跳过,否则后续任务永久等待,失败经Wait重抛
							lock.Lock()
							for currentIndex < i {
								cond.Wait()
							}
							currentIndex = i + 1
							lock.Unlock()
							cond.Broadcast()
							panic(r)
						}
					}()
					a := f(k, v)
					//消费持锁保证顺序,锁经defer归还,消费者panic时锁先行释放,避免recover内重入死锁
					func() {
						lock.Lock()
						defer lock.Unlock()
						for currentIndex != i {
							cond.Wait()
						}
						c(k, a)
						currentIndex = i + 1
					}()
					cond.Broadcast()
				})
			})
			p.Wait()
		}
	}
}

// biMapVCore 常规顺序映射核心
func biMapVCore[K, V, R any](t BiSeq[K, V], f func(K, V) R) BiSeq[K, R] {
	return func(c func(K, R)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// mapParallelCustomizeCore 并行转换自定义调度核心
func mapParallelCustomizeCore[T, R any](t Seq[T], asyncFn func(T, func(R))) Seq[R] {
	return func(c func(R)) {
		wg := sync.WaitGroup{}
		//err保护与hasErr读取分离,hasErr供生产方快速失败无锁检查,与parallel实现一致
		var errMu sync.Mutex
		var err any
		var hasErr int32
		//hasStop标记消费者提前停止,哨兵不入错误,停止语义为正常返回
		var hasStop int32
		t(func(t T) {
			wg.Add(1)
			asyncFn(t, func(r R) {
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
				c(r)
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
