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
    WaiterCount() int32                               // 阻塞等待中的消费者数量
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
fmt.Println(adder.SumInt()) // 返回 int

adder.Reset() // 重置为零

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

依赖goid实现可重入语义,同一goroutine可多次加锁。返回`Locker`接口,支持`TryLock`非阻塞尝试。

```go
type Locker interface {
    sync.Locker
    TryLock() bool
}

lock := NewReentrantLock()
lock.Lock()
lock.Lock() // 可重入
lock.Unlock()
lock.Unlock()

if lock.TryLock() { // 非阻塞尝试加锁
    lock.Unlock()
}
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

## 滑动窗口限流器

时间滑动窗口实现的限流器。

### API

```go
// SlidingWindow 时间滑动窗口限流器
type SlidingWindow struct {
    // Now 时钟函数,返回毫秒时间戳;nil 时 CanDo 使用 time.Now().UnixMilli()
    // 注入假时钟时须以真实毫秒为基准(内部分片时间戳按 time.Now 初始化),并须在首次 CanDo 前设置;
    // 建议取构造时刻或之后,若 now 与构造时刻之差长期小于 everyWindowTime,首片填满后会持续拒绝,直到差值达到该阈值
    Now func() int64
}

// NewSlidingWindow 创建滑动窗口
// windowTime: 时间窗口长度(毫秒),须大于 windowNumber
// allowNumber: 窗口内允许的最大请求数,须不小于 windowNumber
// windowNumber: 子窗口数量,须大于2
func NewSlidingWindow(windowTime int64, allowNumber, windowNumber int32) (*SlidingWindow, error)

// CanDo 获取能否执行,允许则计数并返回true,超限返回false
// 并发安全,可在多goroutine同时调用
func (sw *SlidingWindow) CanDo() bool
```

### 使用示例

```go
// 创建: 1秒窗口,最多100次请求,分成10个子窗口
sw, err := NewSlidingWindow(1000, 100, 10)
if err != nil {
    panic(err) // 参数非法返回哨兵错误
}

// 测试时注入假时钟(毫秒),须在首次 CanDo 前设置,并以构造时刻为基准
base := time.Now().UnixMilli()
sw.Now = func() int64 { return base } // 随测试推进调整 base 的偏移

if sw.CanDo() {
    // 允许执行
}
```

参数说明:
- `windowTime`: 时间窗口长度(毫秒),须大于 `windowNumber`,否则返回 `ErrWindowTime`
- `allowNumber`: 窗口内允许的最大请求数,须不小于 `windowNumber`,否则返回 `ErrAllowNumber`
- `windowNumber`: 子窗口数量,须大于2,否则返回 `ErrWindowNumber`

参数非法时返回包级哨兵错误,可用 errors.Is 匹配。

实现说明: 配额按窗口均分到各子窗口,不能整除时余数分配到前几个分片;当前分片计数满后,须等到 `now` 与上一分片记录时间的差达到 `everyWindowTime`(即 `windowTime/windowNumber`)才清空过期分片并前移指针,未到时间则直接拒绝。差值达到两个分片间隔时一次前进多片(最多整窗);各分片独立记录时间戳,差值达到一个间隔后也可能在同一时刻连续放行整窗配额。滑动过程用互斥锁保护指针与时间戳,计数本身为原子操作。

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
