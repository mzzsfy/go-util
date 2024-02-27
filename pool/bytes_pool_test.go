package pool

import (
    "math/rand"
    "strconv"
    "strings"
    "sync"
    "testing"
    "time"
)

func TestBufferPool(t *testing.T) {
    bp := NewBufferPool()
    bp.SetMaxCap(1024)

    b := bp.Get()
    b.WriteString("Hello, World!")
    bp.Put(b)

    b2 := bp.Get()
    if b2.String() != "" {
        t.Errorf("Expected empty buffer, got %s", b2.String())
    }
}

func TestBytePool(t *testing.T) {
    bp := NewSimpleBytesPool()
    bp.SetMaxCap(1024)
    bp.SetInitCap(512)

    b := bp.Get()
    b.Write([]byte("Hello, World!"))
    bp.Put(b)

    b2 := bp.Get()
    if len(b2.buf) != 0 {
        t.Errorf("Expected empty buffer, got %s", string(b2.buf))
    }
}

var (
    shortStr []string
    midStr   []string
    longStr  []string
)

func TestMain(m *testing.M) {
    rand.Seed(time.Now().UnixNano())
    for i := 0; i < 1000; i++ {
        shortStr = append(shortStr, strings.Repeat(strconv.Itoa(rand.Int()), 1))
        midStr = append(midStr, strings.Repeat(strconv.Itoa(rand.Int()), 10))
        longStr = append(longStr, strings.Repeat(strconv.Itoa(rand.Int()), 100))
    }
    m.Run()
}

func BenchmarkBufferPool(b *testing.B) {
    bp := NewBufferPool()
    bp.SetMaxCap(1024)

    wg := sync.WaitGroup{}
    wg.Add(10)
    for i := 0; i < 10; i++ {
        go func() {
            defer wg.Done()
            for i := 0; i < b.N; i++ {
                for z := range shortStr {
                    buf := bp.Get()
                    buf.WriteString(shortStr[z])
                    bp.Put(buf)
                    buf = bp.Get()
                    buf.WriteString(shortStr[z])
                    bp.Put(buf)
                    buf = bp.Get()
                    buf.WriteString(shortStr[z])
                    bp.Put(buf)
                }
            }
        }()
    }
    wg.Wait()
}

func BenchmarkBytePool(b *testing.B) {
    bp := NewSimpleBytesPool()

    wg := sync.WaitGroup{}
    wg.Add(10)
    for i := 0; i < 10; i++ {
        go func() {
            defer wg.Done()
            for i := 0; i < b.N; i++ {
                for z := range shortStr {
                    buf := bp.Get()
                    buf.WriteString(shortStr[z])
                    bp.Put(buf)
                    buf = bp.Get()
                    buf.WriteString(shortStr[z])
                    bp.Put(buf)
                    buf = bp.Get()
                    buf.WriteString(shortStr[z])
                    bp.Put(buf)
                }
            }
        }()
    }
    wg.Wait()
}
