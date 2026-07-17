package concurrent

import (
    "time"
)

// Queue 队列接口,MPMC安全(算法参考Dmitry Vyukov)
//
// 三种实现:
//   - 默认分段队列(WithTypeSegment): 动态大小,segment链表自动扩容
//   - 环形队列(WithTypeRing): 固定容量,零分配,适合已知上限的场景
//   - channel包装(WithTypeChan): 固定容量,零分配,原生阻塞语义
//
// 阻塞队列使用 BlockQueueWrapper 包装,支持 DequeueBlock 阻塞等待
//
// 示例:
//
//	q := NewQueue[int]()
//	q = NewQueue(WithTypeRing[int](1024))
//	bq := BlockQueueWrapper(q)
//	v, ok := bq.DequeueBlock(time.Second)
type Queue[T any] interface {
    // Enqueue 入队
    Enqueue(v T)
    // Dequeue 出队,返回元素和是否成功
    Dequeue() (T, bool)
    // Size 返回队列大小
    Size() int
}

// BlockQueue 阻塞队列接口
type BlockQueue[T any] interface {
    Queue[T]
    // DequeueBlock 阻塞出队,支持超时参数
    DequeueBlock(timeout ...time.Duration) (T, bool)
}

// TryDequeuer 可选接口:非阻塞尝试出队,空队列立即返回false
type TryDequeuer[T any] interface {
    TryDequeue() (T, bool)
}

type opt[T any] struct {
    Type func() Queue[T]
    opt  []func(Queue[T]) Queue[T]
}

// Opt[T] 队列配置选项
type Opt[T any] func(*opt[T])

// NewQueue 创建队列,默认使用分段队列实现
func NewQueue[T any](opts ...Opt[T]) Queue[T] {
    if len(opts) == 0 {
        return newSegQueue[T]()
    }
    opt := &opt[T]{
        Type: newSegQueue[T],
    }
    for _, o := range opts {
        o(opt)
    }
    r := opt.Type()
    for _, f := range opt.opt {
        r = f(r)
    }
    return r
}
