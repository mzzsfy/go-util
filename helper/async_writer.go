package helper

import (
    "bytes"
    "github.com/mzzsfy/go-util/pool"
    "io"
    "os"
    "runtime"
    "strconv"
    "sync"
    "time"
)

var (
    defaultAsyncConsole = func() *AsyncWrite {
        a := NewAsyncWriter(os.Stdout)
        a.CacheSize = 2000
        a.FlushSize = 1024 * 8 //8k
        return a
    }()
)

type AsyncWriter interface {
    io.Writer
    WriterAsync(p []byte, callback func()) error
}

func AsyncConsole() *AsyncWrite {
    return defaultAsyncConsole
}

func NewAsyncWriter(writer io.Writer) *AsyncWrite {
    a := &AsyncWrite{
        target:    writer,
        FlushSize: 1024 * 2, //2k
        CacheSize: 128,
        pool: pool.NewObjectPool[cell](func() *cell { return &cell{} }, func(c *cell) {
            c.bs = nil
            c.callback = nil
        }),
    }
    a.cache = make(chan *cell)
    close(a.cache)
    return a
}

type cell struct {
    bs       []byte
    callback func()
}

type AsyncWrite struct {
    target    io.Writer
    CacheSize int //缓存大小
    FlushSize int //在内存中保留数据最大值
    BusySize  int //待写入数据超过这个数字,开始丢弃数据
    cache     chan *cell
    bf        bytes.Buffer
    pool      *pool.ObjectPool[cell]
    *sync.Mutex
}

func (c *AsyncWrite) Write(p []byte) (n int, err error) {
    return len(p), c.WriterAsync(p, nil)
}

func (c *AsyncWrite) WriterAsync(p []byte, callback func()) (err error) {
    b := c.pool.Get()
    b.bs = p
    b.callback = callback
    defer func() {
        if e := recover(); e != nil {
            if c.Mutex == nil {
                c.Mutex = new(sync.Mutex)
            }
            runtime.Gosched()
            c.Mutex.Lock()
            defer c.Mutex.Unlock()
            if cap(c.cache) != 0 {
                c.cache <- b
                return
            }
            if c.target == nil {
                panic("target is nil")
            }
            c.cache = make(chan *cell, c.CacheSize)
            if c.BusySize <= 0 {
                c.BusySize = int(float64(c.CacheSize) * .75)
            }
            if c.FlushSize <= 0 {
                c.BusySize = 1024 * 2 //2k
            }
            c.cache <- b
            go func() {
                tick := time.NewTicker(5 * time.Millisecond)
                defer tick.Stop()
                skip := false
                for {
                    select {
                    case ce, ok := <-c.cache:
                        if !ok {
                            return
                        }
                        skip = true
                        c.bf.Write(ce.bs)
                        if ce.callback != nil {
                            ce.callback()
                        }
                        c.pool.Put(ce)
                        if c.bf.Len() > c.FlushSize {
                            c.sync()
                        }
                        i := len(c.cache)
                        if i > c.BusySize {
                            l := i - 10
                            c.bf.WriteString("待写入数据过多,丢弃" + strconv.Itoa(l) + "条\n")
                            for j := 0; j < l; j++ {
                                ce = <-c.cache
                                if ce.callback != nil {
                                    ce.callback()
                                }
                                c.pool.Put(ce)
                            }
                        }
                    case <-tick.C:
                        if !skip && c.bf.Len() != 0 {
                            c.sync()
                        }
                        skip = false
                    }
                }
            }()
        }
    }()
    c.cache <- b
    return nil
}

func (c *AsyncWrite) sync() {
    bs := c.bf.Bytes()
    c.bf.Reset()
    c.target.Write(bs)
}

func (c *AsyncWrite) Reset() {
    func() {
        defer func() { recover() }()
        close(c.cache)
    }()
    c2 := make(chan *cell)
    close(c2)
    c.cache = c2
    c.bf.Reset()
}
