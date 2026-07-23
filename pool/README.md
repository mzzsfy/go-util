# 池化工具

提供对象池、字节池、协程池和字符串池,用于减少 GC 压力,提升高频分配场景的性能。

## 对象池

sync.Pool 的泛型封装,支持自定义创建和重置逻辑。

```go
pool := NewObjectPool[User](func() *User { return &User{} }, func(u *User) { u.id = 0; u.name = "" })

u := pool.Get()
// 使用 u ...
pool.Put(u)
```

## 协程池

弹性协程池,空闲 worker 阻塞等待任务而非自旋,超时后自动退出。支持优雅关闭和重启。

```go
// 使用默认协程池运行任务(返回 error,池关闭时返回 ErrPoolClosed)
err := Go(func() {
	// code
})

// 携带 context 运行任务
err := CtxGo(ctx, func() {
	// code
})

// 自定义协程池
pool := NewGopool(WithName("myPool"), WithMaxWorks(1000), WithIdleTimeout(60 * time.Second))
err := pool.Go(func() {
	// code
})
err := pool.CtxGo(ctx, func() {
	// code
})

// 关闭(停止接受新任务,等待已有 worker 执行完队列残留任务)
ok := pool.Shutdown()

// 重启(等待所有旧 worker 退出后恢复)
ok := pool.Restart()
```

### 协程池选项

| 选项 | 说明 |
|---|---|
| `WithName(name string)` | 设置协程池名称 |
| `WithMaxWorks(n int)` | 设置最大 worker 数量 |
| `WithIdleTimeout(d time.Duration)` | 设置 worker 空闲超时退出时间 |
| `WithPanicHandler(handler func(any, context.Context))` | 设置 panic 处理函数 |

### 协程池方法

| 方法 | 说明 |
|---|---|
| `Name() string` | 获取协程池名称 |
| `WorkerCount() uint64` | 获取当前工作中协程数量 |
| `TaskCount() uint64` | 获取队列任务数量 |
| `Go(f func()) error` | 提交任务 |
| `CtxGo(ctx context.Context, f func()) error` | 携带 context 提交任务 |
| `Shutdown() bool` | 优雅关闭 |
| `Restart() bool` | 重启 |

## 字节池

提供 `BytePool` 和 `BufferPool` 两种字节复用池,归还时超过最大容量的对象会被丢弃,避免大对象常驻内存。

### BytePool

池化自定义 `Bytes` 类型,支持设置初始容量和最大容量。

```go
pool := NewSimpleBytesPool()
pool.SetInitCap(64)
pool.SetMaxCap(1024)

b := pool.Get()
b.WriteString("hello")
// 使用 b.Bytes() 获取数据 ...
pool.Put(b)
```

`Bytes` 提供以下方法: `Write`、`WriteString`、`WriteByte`、`Len`、`Cap`、`Reset`、`String`、`Bytes`。

### BufferPool

池化标准库 `bytes.Buffer`,归还时自动调用 `Reset`。

```go
pool := NewBufferPool()
pool.SetMaxCap(2048)

buf := pool.Get()
buf.WriteString("hello")
// 使用 buf ...
pool.Put(buf)
```

## 字符串池

将字符串映射为数字 ID,适用于高频字符串作为 Map Key 的场景,减少字符串比较和内存开销。内部维护引用计数,引用归零时自动删除条目。

```go
sp := NewStringPool()

id := sp.Use("myKey") // 获取或创建 ID,引用计数 +1
// 使用 id 作为 key ...
peekID := sp.Peek("myKey") // 仅查看 ID,不存在返回 0
sp.UnUse("myKey")          // 引用计数 -1,归零时删除条目
```
