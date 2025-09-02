//go:build !arm64 && !arm && !amd64 && !amd64p32 && !386

package unsafe

import (
    "runtime"
    "strconv"
    "strings"
)

func GoID() int64 {
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
