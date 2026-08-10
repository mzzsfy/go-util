package logger

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Hook 日志行钩子
// AddHook 注册 PreHook: Begin之后、链式字段和消息之前执行
// AddPostHook 注册 PostHook: 消息之后、End之前执行
type Hook func(e *Event)

// maxHooks 每个 hook 列表的最大数量
const maxHooks = 64

// hookList 并发安全的 hook 列表
type hookList struct {
	mu  sync.Mutex
	va  atomic.Value // []Hook
	has int32        // atomic, 快速跳过无 hook 场景
}

func (hl *hookList) init() { hl.va.Store([]Hook(nil)) }

func (hl *hookList) add(h ...Hook) bool {
	if len(h) == 0 {
		return true
	}
	hl.mu.Lock()
	current := hl.va.Load().([]Hook)
	if len(current)+len(h) > maxHooks {
		hl.mu.Unlock()
		return false
	}
	next := make([]Hook, 0, len(current)+len(h))
	next = append(next, current...)
	next = append(next, h...)
	hl.va.Store(next)
	atomic.StoreInt32(&hl.has, 1)
	hl.mu.Unlock()
	return true
}

func (hl *hookList) clean() {
	hl.mu.Lock()
	hl.va.Store([]Hook(nil))
	atomic.StoreInt32(&hl.has, 0)
	hl.mu.Unlock()
}

func (hl *hookList) remove(hook Hook) {
	hl.mu.Lock()
	current := hl.va.Load().([]Hook)
	target := reflect.ValueOf(hook).Pointer()
	next := make([]Hook, 0, len(current))
	for _, h := range current {
		if reflect.ValueOf(h).Pointer() != target {
			next = append(next, h)
		}
	}
	hl.va.Store(next)
	if len(next) > 0 {
		atomic.StoreInt32(&hl.has, 1)
	} else {
		atomic.StoreInt32(&hl.has, 0)
	}
	hl.mu.Unlock()
}

func (hl *hookList) run(e *Event) {
	if atomic.LoadInt32(&hl.has) == 0 {
		return
	}
	for _, h := range hl.va.Load().([]Hook) {
		h(e)
	}
}

var (
	preHooks  hookList // Begin后执行, 追加内容在name后、消息前
	postHooks hookList // 消息后执行, 追加内容在消息后、End前
)

func init() {
	preHooks.init()
	postHooks.init()
}

// AddHook 注册 PreHook, Begin之后执行, 追加内容出现在name后、消息前
func AddHook(h ...Hook) bool { return preHooks.add(h...) }

// AddPostHook 注册 PostHook, 消息之后执行, 追加内容出现在消息后、End前
func AddPostHook(h ...Hook) bool { return postHooks.add(h...) }

// CleanHooks 清理所有钩子 (Pre 和 Post)
func CleanHooks() {
	preHooks.clean()
	postHooks.clean()
}

// RemoveHook 移除指定的钩子函数 (从 Pre 和 Post 中同时移除)
func RemoveHook(hook Hook) {
	preHooks.remove(hook)
	postHooks.remove(hook)
}
