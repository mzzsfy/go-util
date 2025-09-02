package unsafe_test

import (
    "github.com/mzzsfy/go-util/unsafe"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "testing"
    _ "unsafe"
)

func Test_GoID(t *testing.T) {
    var wg sync.WaitGroup
    for i := 0; i < 5000; i++ {
        wg.Add(1)
        go func() {
            testGoID(t)
            wg.Done()
        }()
    }
    wg.Wait()
}

func testGoID(t *testing.T) {
    id := unsafe.GoID()
    id1 := goID1()
    if id != id1 {
        t.Errorf("goroutine id error: %d != %d", id, id1)
    }
}

func goID1() int64 {
    var buf = make([]byte, 29) // "goroutine 9223372036854775807"
    runtime.Stack(buf, false)
    s := string(buf[10:])
    i := strings.IndexByte(s, ' ')
    if i == -1 {
        i = len(s)
    }
    r, _ := strconv.ParseInt(s[:i], 10, 64)
    return r
}
