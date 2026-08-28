package logger

import (
	"io"
	"os"
	"sync/atomic"
)

// writerTarget 全局默认输出目标包装, 用指针存入 atomic.Value 避免类型不一致 panic
type writerTarget struct {
	w io.Writer
}

var (
	defaultLevel    int32        // atomic, Level 值
	defaultWriterVa atomic.Value // *writerTarget
	globalGen       int32        // atomic, 级别变更代, 每次级别变更递增
)

func init() {
	atomic.StoreInt32(&defaultLevel, int32(InfoLevel))
	defaultWriterVa.Store(&writerTarget{w: os.Stdout})
}

// loadDefaultWriter 读取全局默认 writer
func loadDefaultWriter() io.Writer {
	return defaultWriterVa.Load().(*writerTarget).w
}

// SetDefaultLevel 设置全局默认级别, 影响 New 创建的 Log 和继承父级的命名 Log
func SetDefaultLevel(lv Level) {
	atomic.StoreInt32(&defaultLevel, int32(lv))
	atomic.AddInt32(&globalGen, 1)
}

// DefaultLevel 返回全局默认级别
func DefaultLevel() Level {
	return Level(atomic.LoadInt32(&defaultLevel))
}

// SetDefaultWriter 设置全局默认输出目标, 传 nil 回退到 os.Stdout
func SetDefaultWriter(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	defaultWriterVa.Store(&writerTarget{w: w})
}

// DefaultWriter 返回全局默认输出目标
func DefaultWriter() io.Writer {
	return loadDefaultWriter()
}

// --- 旧版兼容函数 ---

// SetDefaultLogLevel 旧版兼容, 等价于 SetDefaultLevel
func SetDefaultLogLevel(l Level) bool { SetDefaultLevel(l); return true }

// DefaultLogLevel 旧版兼容, 等价于 DefaultLevel
func DefaultLogLevel() Level { return DefaultLevel() }

// SetDefaultWriterTarget 旧版兼容, 等价于 SetDefaultWriter
func SetDefaultWriterTarget(w io.Writer) { SetDefaultWriter(w) }

// DefaultWriterTarget 旧版兼容, 等价于 DefaultWriter
func DefaultWriterTarget() io.Writer { return DefaultWriter() }

// defaultLogger 全局默认 Log, 懒初始化
var defaultLogger atomic.Value // *Log

// Default 返回全局默认 Log
func Default() *Log {
	l := defaultLogger.Load()
	if l == nil {
		// 竞争无碍, 多创建一次只是多一次 GC
		nl := New("")
		defaultLogger.Store(nl)
		return nl
	}
	return l.(*Log)
}
