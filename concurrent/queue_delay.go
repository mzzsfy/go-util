package concurrent

import (
	"sync/atomic"
	"time"
)

func WithTypeDelay[T any](delay time.Duration) Opt[T] {
	return func(opt *opt[T]) {
		opt.Type = func() Queue[T] {
			return newDelayQueue[T](delay)
		}
	}
}

func newDelayQueue[T any](delayTime time.Duration) Queue[T] {
	return &delayQueue[T]{
		in:    newSegQueue[delay[T]](),
		out:   newSegQueue[T](),
		delay: delayTime,
	}
}

// 延时队列
//
// 已知局限: 到期序不严格——单消费者串行sleep, 队头长延时元素会推迟后续短延时元素的到期执行;
// 严格到期序需最小堆+条件唤醒, 代价与收益不成比例, 按入队序近似到期
// 已知局限: 无公开 Stop/Close, Queue 接口亦无 Close 语义, 队列随进程生命周期存续
type delay[T any] struct {
	t     time.Time
	value T
}

type delayQueue[T any] struct {
	in      Queue[delay[T]]
	out     Queue[T]
	delay   time.Duration
	running int32
}

func (q *delayQueue[T]) Size() int {
	return q.in.Size() + q.out.Size()
}

func (q *delayQueue[T]) start() {
	if atomic.CompareAndSwapInt32(&q.running, 0, 1) {
		go q.consume()
	}
}

func (q *delayQueue[T]) consume() {
	defer q.stopConsumer()
	for {
		v, b := q.in.Dequeue()
		if !b {
			return
		}
		remaining := q.delay - time.Since(v.t)
		if remaining > 0 {
			time.Sleep(remaining)
		}
		q.out.Enqueue(v.value)
	}
}

// stopConsumer 重置消费者状态; 退出窗口期入队的元素(生产者CAS失败未启动新消费者)由双检重启接管
func (q *delayQueue[T]) stopConsumer() {
	atomic.StoreInt32(&q.running, 0)
	if q.in.Size() > 0 && atomic.CompareAndSwapInt32(&q.running, 0, 1) {
		go q.consume()
	}
}

func (q *delayQueue[T]) Enqueue(v T) {
	q.in.Enqueue(delay[T]{t: time.Now(), value: v})
	q.start()
}

func (q *delayQueue[T]) Dequeue() (T, bool) {
	return q.out.Dequeue()
}
