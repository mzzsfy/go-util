package helper

import (
    "bytes"
    "io"
    "os"
    "runtime"
    "strconv"
    "sync"
    "time"
)

var (
    defaultAsyncConsole = func() *AsyncWriter {
        a := NewAsyncWriter(os.Stdout)
        a.CacheSize = 2000
        a.FlushSize = 1024 * 8 //8k
        return a
    }()
)

func AsyncConsole() *AsyncWriter {
    return defaultAsyncConsole
}

func NewAsyncWriter(writer io.Writer) *AsyncWriter {
    a := &AsyncWriter{
        target:    writer,
        FlushSize: 1024 * 2, //2k
        CacheSize: 128,
    }
    a.cache = make(chan []byte)
    close(a.cache)
    return a
}

type AsyncWriter struct {
    target    io.Writer
    CacheSize int //缓存大小
    FlushSize int //在内存中保留数据最大值
    BusySize  int //待写入数据超过这个数字,开始丢弃数据
    cache     chan []byte
    bf        bytes.Buffer
    *sync.Mutex
}

func (c *AsyncWriter) Write(p []byte) (n int, err error) {
    defer func() {
        if e := recover(); e != nil {
            if c.Mutex == nil {
                c.Mutex = new(sync.Mutex)
            }
            runtime.Gosched()
            c.Mutex.Lock()
            defer c.Mutex.Unlock()
            if cap(c.cache) != 0 {
                c.cache <- p
                return
            }
            if c.target == nil {
                panic("target is nil")
            }
            c.cache = make(chan []byte, c.CacheSize)
            if c.BusySize <= 0 {
                c.BusySize = int(float64(c.CacheSize) * .75)
            }
            if c.FlushSize <= 0 {
                c.BusySize = 1024 * 2 //2k
            }
            c.cache <- p
            go func() {
                tick := time.NewTicker(5 * time.Millisecond)
                defer tick.Stop()
                skip := false
                for {
                    select {
                    case b, ok := <-c.cache:
                        if !ok {
                            return
                        }
                        skip = true
                        c.bf.Write(b)
                        if c.bf.Len() > c.FlushSize {
                            c.sync()
                        }
                        i := len(c.cache)
                        if i > c.BusySize {
                            l := i - 10
                            c.bf.WriteString("待写入数据过多,丢弃" + strconv.Itoa(l) + "条\n")
                            for j := 0; j < l; j++ {
                                <-c.cache
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
    c.cache <- p
    return len(p), nil
}

func (c *AsyncWriter) sync() {
    bs := c.bf.Bytes()
    c.bf.Reset()
    c.target.Write(bs)
}

func (c *AsyncWriter) Reset() {
    func() {
        defer func() { recover() }()
        close(c.cache)
    }()
    c2 := make(chan []byte)
    close(c2)
    c.cache = c2
    c.bf.Reset()
}
