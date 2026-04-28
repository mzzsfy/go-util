package helper

import (
    "runtime"
    "strings"
    "sync/atomic"
)

type Err struct {
    Error any
    Stack Stacks
}

// tryBoundary 缓存 TryWithStack 帧的 PC, 后续 panic 走 O(1) 整数比较
var tryBoundary atomic.Uintptr

// TryWithStack 执行函数 f, 在 panic 时捕获错误和调用栈
// 首次 panic 通过函数名查找 boundary PC 并缓存, 后续走 PC 整数比较
func TryWithStack(f func(), callback func(recoverValue any, stack Stacks)) {
    defer func() {
        if err := recover(); err != nil {
            stack := CallerStack(2)
            cleanTryFrame(&stack)
            callback(err, stack)
        }
    }()
    f()
}

// cleanTryFrame 从调用栈中移除 TryWithStack 帧
// 快速路径: 缓存命中时仅做一次 uintptr 比较
// 慢路径: 首次调用时从当前 panic 上下文获取 PC 并缓存
func cleanTryFrame(stack *Stacks) {
    boundary := tryBoundary.Load()
    if boundary != 0 {
        // 快速路径: O(n) 遍历, 单次整数比较
        for i, s := range *stack {
            if s.PC == boundary {
                *stack = append((*stack)[:i], (*stack)[i+1:]...)
                return
            }
        }
    }
    // 慢路径: 从当前 panic 上下文获取 TryWithStack 帧的 PC
    var pcs [8]uintptr
    n := runtime.Callers(0, pcs[:])
    frames := runtime.CallersFrames(pcs[:n])
    for {
        f, more := frames.Next()
        if strings.HasSuffix(f.Function, ".TryWithStack") {
            tryBoundary.Store(f.PC)
            // 用获取到的 PC 匹配
            for i, s := range *stack {
                if s.PC == f.PC {
                    *stack = append((*stack)[:i], (*stack)[i+1:]...)
                    return
                }
            }
            return
        }
        if !more {
            return
        }
    }
}
