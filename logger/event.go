package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// maxEventBufCap Event buf 超过此容量时丢弃底层数组, 防止池中持有大块内存
const maxEventBufCap = 16 * 1024

// Event 链式日志事件, 由 Log 级别方法创建
// 级别被过滤时返回 disabled Event, 所有方法 noop, 零开销
// 必须调用 Msg/Msgf/Send 触发输出, 否则事件泄露
type Event struct {
	buf       []byte
	msgBuf    []byte // 占位符格式化缓冲区, 池化
	lg        *Log
	level     Level
	enabled   bool
	exitAfter bool // Fatal 级别, 输出后 os.Exit(1)
	caller    uintptr
	callerFn  bool // caller 是否显示函数名
	fmt       Formatter
	ctx       context.Context // 请求级上下文, 从 Logger 继承
}

var eventPool = sync.Pool{
	New: func() any { return &Event{} },
}

// disabledEvent 全局 disabled 单例, 级别被过滤时返回
var disabledEvent = &Event{enabled: false}

func newEvent(l *Log, lv Level) *Event {
	e := eventPool.Get().(*Event)
	e.lg = l
	e.level = lv
	e.enabled = true
	e.buf = e.buf[:0]
	e.ctx = l.ctx
	if l.fmt != nil {
		e.fmt = l.fmt
	} else {
		e.fmt = loadDefaultFormatter()
	}
	e.fmt.Begin(&e.buf, lv, l.name, l.context)
	// PreHook: Begin之后执行, 追加内容在name后、消息前
	preHooks.run(e)
	// caller: 单次读 callerConfig, 提取 enabled + skip + funcFlag
	cc := atomic.LoadInt32(&callerConfig)
	if cc&callerEnabledBit != 0 {
		e.caller = captureCaller(int((cc >> callerSkipShift) & 0xFF))
		e.callerFn = cc&callerFuncBit != 0
	}
	return e
}

func (e *Event) release() {
	if cap(e.buf) > maxEventBufCap {
		e.buf = nil
	}
	if cap(e.msgBuf) > maxEventBufCap {
		e.msgBuf = nil
	}
	e.lg = nil
	e.level = 0
	e.exitAfter = false
	e.caller = 0
	e.callerFn = false
	e.fmt = nil
	e.ctx = nil
	eventPool.Put(e)
}

// --- 公开方法 (Hook 用) ---

func (e *Event) AppendString(s string) {
	e.buf = append(e.buf, s...)
}

func (e *Event) AppendBytes(p []byte) {
	e.buf = append(e.buf, p...)
}

func (e *Event) Level() Level {
	return e.level
}

// Context 返回请求级上下文, 用于hook中提取请求信息(如当前用户)
// 未通过 WithContext 注入时返回 nil
func (e *Event) Context() context.Context {
	return e.ctx
}

// --- 类型安全字段方法, 无 interface 装箱 ---

func (e *Event) Str(key, val string) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Str(&e.buf, key, val)
	return e
}

func (e *Event) Int(key string, val int) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Int(&e.buf, key, val)
	return e
}

func (e *Event) Int64(key string, val int64) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Int64(&e.buf, key, val)
	return e
}

func (e *Event) Uint64(key string, val uint64) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Uint64(&e.buf, key, val)
	return e
}

func (e *Event) Float64(key string, val float64) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Float64(&e.buf, key, val)
	return e
}

func (e *Event) Bool(key string, val bool) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Bool(&e.buf, key, val)
	return e
}

func (e *Event) Time(key string, val time.Time) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Time(&e.buf, key, val)
	return e
}

func (e *Event) Dur(key string, val time.Duration) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Dur(&e.buf, key, val)
	return e
}

func (e *Event) Err(err error) *Event {
	if !e.enabled || err == nil {
		return e
	}
	e.fmt.Err(&e.buf, err)
	return e
}

// Any 通用字段方法, 有 interface 装箱开销, 热路径慎用
func (e *Event) Any(key string, val any) *Event {
	if !e.enabled {
		return e
	}
	e.fmt.Any(&e.buf, key, val)
	return e
}

// --- Logger 派生 ---

// Log 从 With 事件构建新 Log, 复用已写入的字段作为预设字段
func (e *Event) Logger() *Log {
	resolved := atomic.LoadInt32(&e.lg.resolved)
	nl := &Log{
		name:     e.lg.name,
		writer:   e.lg.writer,
		isAsync:  e.lg.isAsync,
		aw:       e.lg.aw,
		localLv:  resolved,
		resolved: resolved,
		fmt:      e.lg.fmt,
		ctx:      e.lg.ctx,
		// gen 默认 0, localLv != levelInherit 时 level() 不检查 gen
	}
	if len(e.buf) > 0 {
		nl.context = string(e.buf)
	}
	e.release()
	return nl
}

// --- 触发写入 ---

func (e *Event) Msg(msg string) {
	if !e.enabled {
		return
	}
	e.fmt.Msg(&e.buf, msg)
	e.flush()
}

// MsgFormat 占位符格式化消息后输出, 零分配
// 支持 {} 自动递增, {0} 显式位置, {%s} 格式化, {0%d} 位置+格式化
func (e *Event) MsgFormat(msg string, args []any) {
	if !e.enabled {
		return
	}
	e.msgBuf = doFormatPlaceholders(e.msgBuf[:0], msg, args)
	e.fmt.Msg(&e.buf, b2s(e.msgBuf))
	e.flush()
}

func (e *Event) Msgf(format string, args ...any) {
	if !e.enabled {
		return
	}
	e.fmt.Msg(&e.buf, fmt.Sprintf(format, args...))
	e.flush()
}

func (e *Event) Send() {
	if !e.enabled {
		return
	}
	e.flush()
}

// asyncWriter 异步写入接口, 与 helper.AsyncWriter 兼容
type asyncWriter interface {
	WriterAsync(p []byte, callback func()) error
}

// bufPool 异步写入的中转 buffer 池
var bufPool = sync.Pool{
	New: func() any { b := make([]byte, 0, 256); return &b },
}

// flush 写入输出并归还 Event
// 顺序: PostHook → fmt.End(caller+换行) → write → release
func (e *Event) flush() {
	postHooks.run(e)
	e.fmt.End(&e.buf, e.caller, e.callerFn)

	// 热替换 writer 存在瞬态窗口, aw 为 nil 时兜底回退同步写
	aw := e.lg.aw
	if e.lg.isAsync && aw != nil {
		// 拷贝到池化中转 buffer, Event buf 常驻复用
		bp := bufPool.Get().(*[]byte)
		*bp = append((*bp)[:0], e.buf...)
		aw(*bp, func() {
			bufPool.Put(bp)
		})
	} else {
		e.lg.writer.Write(e.buf)
	}

	fatal := e.exitAfter
	e.release()
	if fatal {
		os.Exit(1)
	}
}
