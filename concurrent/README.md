# 并发工具

包含高性能原子计数器、可重入锁、MPMC队列、滑动窗口限流器等并发安全工具。

## 队列

MPMC安全的泛型队列,算法参考Dmitry Vyukov。

### 实现类型

| 类型 | 构造选项 | 特点 | 适用场景 |
|------|----------|------|----------|
| 分段队列 | `WithTypeSegment[T]()` | 动态大小,自动扩容 | 队列大小未知,通用场景 |
| 环形队列 | `WithTypeRing[T](cap)` | 固定容量,零分配,预分配内存 | 已知上限,高频MPMC |
| channel队列 | `WithTypeChan[T](buffer)` | 固定容量,零分配,原生阻塞语义 | 需要阻塞语义的场景 |
| 延时队列 | `WithTypeDelay[T](delay)` | 入队后延时出队 | 延时任务调度 |

### 接口

```go
// Queue 队列接口
type Queue[T any] interface {
    Enqueue(v T)           // 入队
    Dequeue() (T, bool)    // 出队,返回元素和是否成功
    Size() int             // 队列大小
}

// BlockQueue 阻塞队列接口
type BlockQueue[T any] interface {
    Queue[T]
    DequeueBlock(timeout ...time.Duration) (T, bool) // 阻塞出队
}

// TryDequeuer 非阻塞尝试出队接口
type TryDequeuer[T any] interface {
    TryDequeue() (T, bool) // 空队列立即返回false
}
```

### 使用示例

```go
// 默认分段队列(动态大小)
q := NewQueue[int]()

// 环形队列(固定容量,零分配)
q = NewQueue(WithTypeRing[int](1024))

// channel队列(原生阻塞语义)
q = NewQueue(WithTypeChan[int](100))

// 延时队列(入队后延时出队)
q = NewQueue(WithTypeDelay[int](time.Second))

// 阻塞队列包装
bq := BlockQueueWrapper(q)
v, ok := bq.DequeueBlock(time.Second) // 阻塞等待1秒
```

### 性能对比

详见 [queue_perf_analysis.md](queue_perf_analysis.md)

| 指标 | segQueue | ringQueue | 说明 |
|------|----------|-----------|------|
| 1P1C ns/op | 20.7 | 19.0 | 单生产单消费差距小 |
| MPMC ns/op | 117~132 | 54~58 | ring在MPMC下优势明显 |
| 内存分配 | 动态 | 预分配 | ring零分配 |

选型建议:
- 队列大小未知: `WithTypeSegment`
- 已知上限的高频MPMC: `WithTypeRing`
- 需要延时: `WithTypeDelay`

## Int64Adder

类似Java LongAdder的高性能原子计数器,高并发场景下比`atomic.AddInt64`性能更好。

### 使用示例

```go
adder := &Int64Adder{}
adder.AddSimple(10)
adder.IncrementSimple()
adder.DecrementSimple()
fmt.Println(adder.Sum())

// 高性能模式:手动传入goid
goid := GoID()
adder.Add(goid, 100)
adder.Increment(goid)
adder.Decrement(goid)
```

### 内存布局优化

通过条件编译调整cache line填充,缓解伪共享:

| tag | 填充字节 | 说明 |
|-----|----------|------|
| 默认 | 56 | 平衡性能与内存 |
| `concurrent_fast` | 120 | 最大性能,内存占用更高 |
| `concurrent_memory` | 24 | 节省内存,性能略有下降 |

```shell
go test -tags=concurrent_fast -bench=. ./concurrent
```

### 性能数据

```shell
$ go test -bench=Benchmark1.+ ./concurrent
Benchmark1Int64Adder/Int64Adder_32-6         137929      9034 ns/op
Benchmark1Int64Adder/Atomic_32-6              51117     34734 ns/op
```

## 锁

### 可重入锁

依赖goid实现可重入语义,同一goroutine可多次加锁。

```go
lock := NewReentrantLock()
lock.Lock()
lock.Lock() // 可重入
lock.Unlock()
lock.Unlock()
```

### RwLocker接口

```go
type RwLocker interface {
    Locker
    RLock()
    RUnlock()
    TryRLock() bool
}
```

### CasRwLocker (已弃用)

基于CAS实现的读写锁,性能不如`sync.RWMutex`,仅保留用于基准测试。

```go
// Deprecated: 使用sync.RWMutex替代
var lock CasRwLocker
```

### NoLock

空锁实现,用于占位或测试。

```go
var l NoLock
l.Lock()   // 无操作
l.Unlock() // 无操作
```

## IdGenerator

ID生成器接口,支持雪花算法和原子计数器两种实现。

```go
type IdGenerator interface {
    NextId() uint64
}
```

内置实现(包内使用):
- `snowFlake`: 雪花算法,CAS优化,高并发性能好
- `atomIdGenerator`: 原子递增

## 滑动窗口限流器

时间滑动窗口实现的限流器。

```go
// 创建: 1秒窗口,最多100次请求,分成10个子窗口
sw := NewSlidingWindow(1000, 100, 10)

if sw.CanDo() {
    // 允许执行
}
```

参数说明:
- `time`: 时间窗口长度(毫秒)
- `allowNumber`: 窗口内允许的最大请求数
- `windowNumber`: 子窗口数量,必须大于2

## 工具函数

### GoID

获取当前goroutine id,可替换实现。

```go
goid := GoID()
```

### Helper

锁辅助工具。

```go
h := Helper{Locker: &sync.Mutex{}}
h.RunWithLock(func() { /* 临界区 */ })
defer h.Lock1()() // 延迟解锁模式
```
