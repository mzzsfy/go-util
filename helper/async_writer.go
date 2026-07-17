package helper

import (
    "io"
    "os"
    "strconv"
    "sync"
    "sync/atomic"
    "time"
)

// 默认异步写入器配置常量
const (
    // defaultCacheSize 默认缓存通道大小
    defaultCacheSize = 128
    // defaultFlushSize 默认刷写阈值(字节)
    defaultFlushSize = 2 * 1024
    // defaultConsoleFlushSize 默认控制台刷写阈值(字节)
    defaultConsoleFlushSize = 8 * 1024
)

// 默认异步控制台输出
var defaultAsyncConsole = func() *AsyncWrite {
    a := NewAsyncWriter(os.Stdout)
    a.SetFlushSize(defaultConsoleFlushSize)
    return a
}()

// AsyncWriter 异步写入接口
type AsyncWriter interface {
    io.Writer
    WriterAsync(p []byte, callback func()) error
}

// AsyncConsole 返回默认的异步控制台写入器
func AsyncConsole() *AsyncWrite {
    return defaultAsyncConsole
}

// NewAsyncWriter 创建异步写入器,立即启动后台消费协程
func NewAsyncWriter(writer io.Writer) *AsyncWrite {
    a := &AsyncWrite{
        target:    writer,
        cacheSize: defaultCacheSize,
        flushSize: defaultFlushSize,
        pool: sync.Pool{
            New: func() any { return &cell{} },
        },
    }
    cache := make(chan *cell, a.CacheSize())
    done := make(chan struct{})
    stopped := make(chan struct{})
    a.cache = cache
    a.done = done
    a.stopped = stopped
    atomic.StoreInt32(&a.busySize, int32(float64(a.CacheSize())*0.75))
    go a.consumeLoop(cache, done, stopped)
    return a
}

type cell struct {
    bs       []byte
    callback func()
}

// AsyncWrite 异步写入器
type AsyncWrite struct {
    target    io.Writer
    cacheSize int32
    flushSize int32
    busySize  int32
    cache     chan *cell
    pool      sync.Pool
    done      chan struct{}
    stopped   chan struct{}
    mu        sync.Mutex
}

func (c *AsyncWrite) CacheSize() int     { return int(atomic.LoadInt32(&c.cacheSize)) }
func (c *AsyncWrite) FlushSize() int     { return int(atomic.LoadInt32(&c.flushSize)) }
func (c *AsyncWrite) BusySize() int      { return int(atomic.LoadInt32(&c.busySize)) }
func (c *AsyncWrite) SetCacheSize(v int) { atomic.StoreInt32(&c.cacheSize, int32(v)) }
func (c *AsyncWrite) SetFlushSize(v int) { atomic.StoreInt32(&c.flushSize, int32(v)) }
func (c *AsyncWrite) SetBusySize(v int) {
    if v < 0 {
        v = 0
    }
    atomic.StoreInt32(&c.busySize, int32(v))
}

// Write 同步写入,始终返回 len(p), nil
func (c *AsyncWrite) Write(p []byte) (n int, err error) {
    return len(p), c.WriterAsync(p, nil)
}

// WriterAsync 异步写入数据,回调在数据真正写入底层目标后触发
func (c *AsyncWrite) WriterAsync(p []byte, callback func()) error {
    b := c.pool.Get().(*cell)
    b.bs = p
    b.callback = callback

    // 锁内非阻塞发送: 消除获取 cache 引用后到发送之间的竞态窗口
    // Reset 需要获取同一把锁才能替换 cache, 保证发送期间 cache 不会变化
    c.mu.Lock()
    select {
    case c.cache <- b:
        c.mu.Unlock()
    default:
        c.mu.Unlock()
        // channel 满时丢弃数据,直接触发回调,与背压丢弃语义一致
        _, cb := c.recycleCell(b)
        if cb != nil {
            cb()
        }
    }

    return nil
}

// recycleCell 回收单个 cell: 清零字段并归还 pool, 返回 cell 的数据和回调
func (c *AsyncWrite) recycleCell(ce *cell) ([]byte, func()) {
    bs, cb := ce.bs, ce.callback
    ce.bs = nil
    ce.callback = nil
    c.pool.Put(ce)
    return bs, cb
}

// consumeLoop 后台消费协程,从缓存通道读取数据并批量写入
// cache/done/stopped 通过参数传入,确保 Reset 关闭的是本协程持有的同一组通道
func (c *AsyncWrite) consumeLoop(cache chan *cell, done, stopped chan struct{}) {

    tick := time.NewTicker(5 * time.Millisecond)
    defer tick.Stop()

    // 预分配 flushSize 容量, 减少 append 扩容次数
    var buf = make([]byte, 0, c.FlushSize())
    var callbacks []func()
    hasData := false

    for {
        // 单协程消费, 每轮循环顶部缓存 atomic 值, 避免每条消息都 load
        flushSize := c.FlushSize()
        busySize := c.BusySize()
        select {
        case ce, ok := <-cache:
            if !ok {
                // 通道关闭,刷写剩余数据后退出
                c.flushLocal(&buf, &callbacks)
                close(stopped)
                return
            }
            bs, cb := c.recycleCell(ce)
            buf = append(buf, bs...)
            if cb != nil {
                callbacks = append(callbacks, cb)
            }
            hasData = true

            // 缓冲区超过刷写阈值,立即刷写
            if len(buf) > flushSize {
                c.flushLocal(&buf, &callbacks)
                hasData = false
            }

            // 背压处理: 积压过多时丢弃一半数据
            backlog := len(cache)
            if backlog > busySize {
                drop := backlog >> 1
                if drop > 0 {
                    buf = append(buf, "待写入数据过多,丢弃"+strconv.Itoa(drop)+"条\n"...)
                    for j := 0; j < drop; j++ {
                        select {
                        case ce = <-cache:
                            _, cb := c.recycleCell(ce)
                            if cb != nil {
                                cb()
                            }
                        case <-done:
                            // 收到停止信号,刷写已缓冲数据并退出
                            c.flushLocal(&buf, &callbacks)
                            close(stopped)
                            return
                        }
                    }
                }
            }

        case <-tick.C:
            if hasData {
                c.flushLocal(&buf, &callbacks)
                hasData = false
            }

        case <-done:
            // 收到停止信号,排空缓存通道后退出
            c.drainChannel(cache, &buf, &callbacks)
            close(stopped)
            return
        }
    }
}

// drainChannel 排空缓存通道中所有剩余数据并刷写
func (c *AsyncWrite) drainChannel(cache chan *cell, buf *[]byte, callbacks *[]func()) {
    // 用 len 判断循环,避免 select+default 在通道瞬时为空时提前退出
    for len(cache) > 0 {
        ce := <-cache
        bs, cb := c.recycleCell(ce)
        *buf = append(*buf, bs...)
        if cb != nil {
            *callbacks = append(*callbacks, cb)
        }
    }
    c.flushLocal(buf, callbacks)
}

// flushLocal 将缓冲数据写入底层目标,并在写入完成后触发回调
func (c *AsyncWrite) flushLocal(buf *[]byte, callbacks *[]func()) {
    if len(*buf) == 0 {
        return
    }
    data := *buf
    cb := *callbacks
    // 保留底层数组容量供下个周期复用, 减少 GC 压力
    *buf = (*buf)[:0]
    *callbacks = (*callbacks)[:0]

    // 实际写入底层目标
    c.target.Write(data)

    // 写入完成后触发回调
    for _, fn := range cb {
        fn()
    }
}

// Reset 重置写入器,停止旧协程并启动新协程
func (c *AsyncWrite) Reset() {
    c.mu.Lock()
    oldDone := c.done
    oldStopped := c.stopped

    // 创建新的通道和信号
    cache := make(chan *cell, c.CacheSize())
    done := make(chan struct{})
    stopped := make(chan struct{})
    c.cache = cache
    c.done = done
    c.stopped = stopped
    c.mu.Unlock()

    // 通知旧协程停止并等待完成
    close(oldDone)
    <-oldStopped

    // 启动新的消费协程
    go c.consumeLoop(cache, done, stopped)
}
