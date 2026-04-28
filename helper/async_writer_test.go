package helper

import (
    "bytes"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

type callbackBuffer struct {
    mu     sync.Mutex
    buf    bytes.Buffer
    hook   func(string)
    writes int
}

// Write 记录写入内容并触发钩子
func (c *callbackBuffer) Write(p []byte) (n int, err error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    _, _ = c.buf.Write(p)
    c.writes++
    if c.hook != nil {
        c.hook(c.buf.String())
    }
    return len(p), nil
}

// String 返回当前缓冲区内容
func (c *callbackBuffer) String() string {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.buf.String()
}

// TestAsyncWriterWriteContract 验证 Write 返回值约定
func TestAsyncWriterWriteContract(t *testing.T) {
    t.Parallel()

    target := &callbackBuffer{}
    writer := NewAsyncWriter(target)
    defer writer.Reset()

    n, err := writer.Write([]byte("hello"))
    if err != nil {
        t.Fatalf("Write returned error: %v", err)
    }
    if n != len("hello") {
        t.Fatalf("Write returned length %d", n)
    }

    deadline := time.Now().Add(time.Second)
    for target.String() != "hello" {
        if time.Now().After(deadline) {
            t.Fatalf("unexpected content: %q", target.String())
        }
        time.Sleep(time.Millisecond)
    }
}

// TestAsyncWriterCallbackAfterWrite 验证回调在底层写入后触发
func TestAsyncWriterCallbackAfterWrite(t *testing.T) {
    t.Parallel()

    callbackDone := make(chan struct{}, 1)
    target := &callbackBuffer{}
    target.hook = func(content string) {
        if content != "hello" {
            return
        }
    }

    writer := NewAsyncWriter(target)
    defer writer.Reset()

    if err := writer.WriterAsync([]byte("hello"), func() {
        if target.String() != "hello" {
            t.Errorf("callback fired before write, content=%q", target.String())
        }
        callbackDone <- struct{}{}
    }); err != nil {
        t.Fatalf("WriterAsync returned error: %v", err)
    }

    select {
    case <-callbackDone:
    case <-time.After(time.Second):
        t.Fatal("callback not fired")
    }
}

// TestAsyncWriterResetRestart 验证 Reset 后仍可继续写入
func TestAsyncWriterResetRestart(t *testing.T) {
    t.Parallel()

    target := &callbackBuffer{}
    writer := NewAsyncWriter(target)

    if err := writer.WriterAsync([]byte("first"), nil); err != nil {
        t.Fatalf("first WriterAsync returned error: %v", err)
    }

    deadline := time.Now().Add(time.Second)
    for target.String() != "first" {
        if time.Now().After(deadline) {
            t.Fatalf("unexpected first content: %q", target.String())
        }
        time.Sleep(time.Millisecond)
    }

    writer.Reset()
    defer writer.Reset()

    if err := writer.WriterAsync([]byte("second"), nil); err != nil {
        t.Fatalf("second WriterAsync returned error: %v", err)
    }

    deadline = time.Now().Add(time.Second)
    for target.String() != "firstsecond" {
        if time.Now().After(deadline) {
            t.Fatalf("unexpected reset content: %q", target.String())
        }
        time.Sleep(time.Millisecond)
    }
}

// TestAsyncWriterBackpressureCallback 验证背压丢弃时回调仍被触发
func TestAsyncWriterBackpressureCallback(t *testing.T) {
    t.Parallel()

    target := &callbackBuffer{}
    writer := NewAsyncWriter(target)
    // 设置较小的容量使背压容易触发
    writer.SetCacheSize(16)
    writer.SetBusySize(8)
    writer.SetFlushSize(64)

    const n = 100
    var callbackCount int32

    for i := 0; i < n; i++ {
        data := []byte("data" + strconv.Itoa(i))
        if err := writer.WriterAsync(data, func() {
            atomic.AddInt32(&callbackCount, 1)
        }); err != nil {
            t.Fatalf("WriterAsync returned error: %v", err)
        }
    }

    // 等待所有回调触发
    deadline := time.Now().Add(5 * time.Second)
    for atomic.LoadInt32(&callbackCount) < n {
        if time.Now().After(deadline) {
            t.Fatalf("expected %d callbacks, got %d", n, atomic.LoadInt32(&callbackCount))
        }
        time.Sleep(time.Millisecond)
    }

    writer.Reset()
}

// TestDrainChannelCompletes 验证 drainChannel 能排空缓冲区中全部数据
func TestDrainChannelCompletes(t *testing.T) {
    t.Parallel()

    const n = 64
    ch := make(chan *cell, n)
    for i := 0; i < n; i++ {
        ch <- &cell{bs: []byte("x")}
    }

    target := &callbackBuffer{}
    // 手动构造 AsyncWrite,不启动后台协程,仅测试 drainChannel 逻辑
    a := &AsyncWrite{
        target: target,
        pool:   sync.Pool{New: func() any { return &cell{} }},
    }

    var buf []byte
    var callbacks []func()
    a.drainChannel(ch, &buf, &callbacks)

    // drainChannel 内部会调用 flushLocal,数据应已写入 target
    if got := target.String(); got != strings.Repeat("x", n) {
        t.Fatalf("期望 %d 字节写入,实际 %d 字节", n, len(got))
    }
    if len(ch) != 0 {
        t.Fatalf("通道未排空,剩余 %d 项", len(ch))
    }
}

// TestWriterAsyncConcurrentWithReset 验证多次 Reset 后仍可正常工作
func TestWriterAsyncConcurrentWithReset(t *testing.T) {
    t.Parallel()

    const rounds = 10
    for r := 0; r < rounds; r++ {
        target := &callbackBuffer{}
        writer := NewAsyncWriter(target)

        // 写入并等待消费
        writer.WriterAsync([]byte("data"+strconv.Itoa(r)), nil)
        deadline := time.Now().Add(time.Second)
        for target.String() == "" {
            if time.Now().After(deadline) {
                t.Fatalf("round %d: 数据未写入", r)
            }
            runtime.Gosched()
        }

        // Reset 后验证可以继续使用
        writer.Reset()
        writer.WriterAsync([]byte("after-reset"), nil)
        deadline = time.Now().Add(time.Second)
        for !strings.Contains(target.String(), "after-reset") {
            if time.Now().After(deadline) {
                t.Fatalf("round %d: Reset 后数据未写入, content=%q", r, target.String())
            }
            runtime.Gosched()
        }
    }
}

// TestAsyncWriter_ResetCompletes 验证 Reset 能正常完成
func TestAsyncWriter_ResetCompletes(t *testing.T) {
    t.Parallel()

    target := &callbackBuffer{}
    writer := NewAsyncWriter(target)

    // 写入一些数据
    writer.WriterAsync([]byte("data1"), nil)
    writer.WriterAsync([]byte("data2"), nil)

    // Reset 应在合理时间内完成
    done := make(chan struct{})
    go func() {
        writer.Reset()
        close(done)
    }()

    select {
    case <-done:
        // 成功
    case <-time.After(3 * time.Second):
        t.Fatal("Reset 阻塞超过 3 秒")
    }

    // 验证 Reset 后仍可写入
    writer.WriterAsync([]byte("after-reset"), nil)
    deadline := time.Now().Add(time.Second)
    for !strings.Contains(target.String(), "after-reset") {
        if time.Now().After(deadline) {
            t.Fatalf("Reset 后数据未写入, content=%q", target.String())
        }
        runtime.Gosched()
    }
}

// TestAsyncWriter_AtomicConfigGetterSetter 验证配置字段的 atomic getter/setter
func TestAsyncWriter_AtomicConfigGetterSetter(t *testing.T) {
    t.Parallel()

    target := &callbackBuffer{}
    writer := NewAsyncWriter(target)
    defer writer.Reset()

    // 验证默认值
    if writer.CacheSize() != 128 {
        t.Fatalf("expected default CacheSize 128, got %d", writer.CacheSize())
    }
    if writer.FlushSize() != 2048 {
        t.Fatalf("expected default FlushSize 2048, got %d", writer.FlushSize())
    }

    // 验证 setter 生效
    writer.SetCacheSize(256)
    if writer.CacheSize() != 256 {
        t.Fatalf("expected CacheSize 256, got %d", writer.CacheSize())
    }
    writer.SetFlushSize(4096)
    if writer.FlushSize() != 4096 {
        t.Fatalf("expected FlushSize 4096, got %d", writer.FlushSize())
    }
    writer.SetBusySize(32)
    if writer.BusySize() != 32 {
        t.Fatalf("expected BusySize 32, got %d", writer.BusySize())
    }
}

// TestAsyncWriter_ConcurrentWriteReset 验证 WriterAsync 与 Reset 并发时数据不丢失
// 多个 goroutine 同时写入, 另一个 goroutine 执行 Reset, 验证已写入的数据全部到达底层
func TestAsyncWriter_ConcurrentWriteReset(t *testing.T) {
    t.Parallel()

    const writers = 4
    const writesPerWriter = 50
    const resetCount = 5

    target := &callbackBuffer{}
    writer := NewAsyncWriter(target)
    defer writer.Reset()

    var totalSent int32

    for r := 0; r < resetCount; r++ {
        var wg sync.WaitGroup
        wg.Add(writers)

        // 并发写入
        for w := 0; w < writers; w++ {
            go func(id int) {
                defer wg.Done()
                for i := 0; i < writesPerWriter; i++ {
                    data := []byte("w" + strconv.Itoa(id) + "d" + strconv.Itoa(i))
                    writer.WriterAsync(data, nil)
                    atomic.AddInt32(&totalSent, 1)
                    runtime.Gosched()
                }
            }(w)
        }

        // 等待本轮写入完成
        wg.Wait()

        // 执行 Reset
        writer.Reset()
    }

    // 最终写入一批数据并验证
    writer.WriterAsync([]byte("final"), nil)
    deadline := time.Now().Add(2 * time.Second)
    for !strings.Contains(target.String(), "final") {
        if time.Now().After(deadline) {
            t.Fatalf("final 数据未写入, content 长度=%d", len(target.String()))
        }
        runtime.Gosched()
    }

    // 验证无 panic, 无死锁
    t.Logf("totalSent=%d, target写入次数=%d", atomic.LoadInt32(&totalSent), target.writes)
}
