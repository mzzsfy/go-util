package logger

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Hook 日志行钩子, 在 Formatter.Begin 之后、链式字段和消息之前调用
// 可通过 Event 的 AppendString / AppendBytes 追加内容到缓冲区
// 追加内容出现在 name 之后、消息之前
type Hook func(e *Event)

// maxHooks 最大 hook 数量, 防止每次日志调用开销线性增长
const maxHooks = 64

var (
	hooksMu  sync.Mutex
	hooksVa  atomic.Value // []Hook
	hasHooks int32        // atomic, 快速跳过无 hook 场景
)

func init() {
	hooksVa.Store([]Hook(nil))
}

// AddHook 注册日志钩子, 最多 maxHooks 个, 返回是否成功
func AddHook(h ...Hook) bool {
	if len(h) == 0 {
		return true
	}
	hooksMu.Lock()
	current := hooksVa.Load().([]Hook)
	if len(current)+len(h) > maxHooks {
		hooksMu.Unlock()
		return false
	}
	next := make([]Hook, 0, len(current)+len(h))
	next = append(next, current...)
	next = append(next, h...)
	hooksVa.Store(next)
	atomic.StoreInt32(&hasHooks, 1)
	hooksMu.Unlock()
	return true
}

// CleanHooks 清理所有钩子
func CleanHooks() {
	hooksMu.Lock()
	hooksVa.Store([]Hook(nil))
	atomic.StoreInt32(&hasHooks, 0)
	hooksMu.Unlock()
}

// RemoveHook 移除指定钩子 (基于函数指针比较)
func RemoveHook(hook Hook) {
	hooksMu.Lock()
	current := hooksVa.Load().([]Hook)
	target := reflect.ValueOf(hook).Pointer()
	next := make([]Hook, 0, len(current))
	for _, h := range current {
		if reflect.ValueOf(h).Pointer() != target {
			next = append(next, h)
		}
	}
	hooksVa.Store(next)
	if len(next) > 0 {
		atomic.StoreInt32(&hasHooks, 1)
	} else {
		atomic.StoreInt32(&hasHooks, 0)
	}
	hooksMu.Unlock()
}
