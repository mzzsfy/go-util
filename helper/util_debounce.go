package helper

import (
    "sync"
    "time"
)

// Debounce 创建一个防抖函数,在duration时间内只执行一次
func Debounce(call func(), duration time.Duration) func() {
    var lastCall *time.Time
    return func() {
        if lastCall == nil {
            call()
            t := time.Now()
            lastCall = &t
        } else {
            now := time.Now()
            if now.Sub(*lastCall) > duration {
                lastCall = &now
                call()
            }
        }
    }
}

// DebounceConcurrent 创建一个并发安全的防抖函数
func DebounceConcurrent(call func(), duration time.Duration) func() {
    f := Debounce(call, duration)
    var lock sync.Mutex
    return func() {
        lock.Lock()
        defer lock.Unlock()
        f()
    }
}
