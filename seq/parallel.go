package seq

import (
	"errors"
	"sync"
	"sync/atomic"
)

// ErrConcurrent 并发数非法
var ErrConcurrent = errors.New("concurrent must > 0")

type Parallel interface {
	// Add 添加一个函数到并行执行队列中,当并发数达到上限时, 会阻塞等待
	// 已有任务panic或消费者提前停止后,Add会以相同panic值中止生产
	Add(fn func())
	// Wait 关闭任务通道并等待worker全部退出,已有任务panic时以该panic值重抛
	Wait()
}

// NewParallel 创建并行执行器,concurrent为并发worker数量
func NewParallel(concurrent int) (Parallel, error) {
	if concurrent <= 0 {
		return nil, ErrConcurrent
	}
	return newParallel(concurrent), nil
}

// newParallel 内部构造,调用方保证并发数合法,返回具体类型以暴露abort兜底回收
func newParallel(concurrent int) *parallel {
	p := &parallel{tasks: make(chan func())}
	for i := 0; i < concurrent; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p
}

// parallel 固定worker池,任务经无缓冲channel分发,worker全忙时Add阻塞,语义与旧版并发上限一致
type parallel struct {
	tasks chan func()
	wg    sync.WaitGroup
	//closeOnce保证close(tasks)幂等,生产方panic中止时由abort兜底回收,避免重复close
	closeOnce sync.Once
	// err保护与hasErr读取分离,hasErr供Add快速失败无锁检查
	errMu  sync.Mutex
	err    any
	hasErr int32
	//hasStop标记消费者提前停止,哨兵不入错误通道,停止语义为正常返回
	hasStop int32
}

func (p *parallel) worker() {
	defer p.wg.Done()
	for fn := range p.tasks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					//消费者提前停止的哨兵不入错误,标记停止供Add快速失败,由源头stopRecover吸收
					if r == &stop {
						atomic.StoreInt32(&p.hasStop, 1)
						return
					}
					p.errMu.Lock()
					if p.err == nil {
						p.err = r
						atomic.StoreInt32(&p.hasErr, 1)
					}
					p.errMu.Unlock()
				}
			}()
			fn()
		}()
	}
}

func (p *parallel) Add(fn func()) {
	//消费者已提前停止时中止生产,哨兵由源头stopRecover吸收
	if atomic.LoadInt32(&p.hasStop) == 1 {
		panic(&stop)
	}
	//已有任务失败时快速失败,终止生产,与旧版Add阻塞后panic行为一致
	if atomic.LoadInt32(&p.hasErr) == 1 {
		p.errMu.Lock()
		e := p.err
		p.errMu.Unlock()
		panic(e)
	}
	p.tasks <- fn
}

func (p *parallel) Wait() {
	p.closeOnce.Do(func() { close(p.tasks) })
	p.wg.Wait()
	if atomic.LoadInt32(&p.hasErr) == 1 {
		p.errMu.Lock()
		e := p.err
		p.errMu.Unlock()
		panic(e)
	}
}

// abort 生产方panic中止时的兜底回收,幂等且不重抛错误,确保常驻worker全部退出
func (p *parallel) abort() {
	p.closeOnce.Do(func() { close(p.tasks) })
	p.wg.Wait()
}
