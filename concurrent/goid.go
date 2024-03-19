package concurrent

import (
    "github.com/mzzsfy/go-util/unsafe"
)

// GoID 返回当前goroutine的id,公开以允许替换实现
var GoID = unsafe.GoID
