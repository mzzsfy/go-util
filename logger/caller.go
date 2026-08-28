package logger

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// caller 配置位布局: bit0=enabled, bit1=showFunc, bit8-15=userSkip
const (
	callerEnabledBit  int32 = 1 << 0
	callerFuncBit     int32 = 1 << 1
	callerSkipShift         = 8
	callerSkipMask    int32 = 0xFF << callerSkipShift
	defaultCallerSkip       = 0
)

var callerConfig int32

func init() {
	atomic.StoreInt32(&callerConfig, int32(defaultCallerSkip<<callerSkipShift))
}

// SetCaller 启用/禁用调用者信息 (file:line)
func SetCaller(enabled bool) {
	for {
		old := atomic.LoadInt32(&callerConfig)
		nb := old &^ callerEnabledBit
		if enabled {
			nb |= callerEnabledBit
		}
		if atomic.CompareAndSwapInt32(&callerConfig, old, nb) {
			return
		}
	}
}

// SetCallerFunc 启用/禁用函数名显示 (需先 SetCaller(true))
func SetCallerFunc(enabled bool) {
	for {
		old := atomic.LoadInt32(&callerConfig)
		nb := old &^ callerFuncBit
		if enabled {
			nb |= callerFuncBit
		}
		if atomic.CompareAndSwapInt32(&callerConfig, old, nb) {
			return
		}
	}
}

// SetCallerInfo 旧版兼容, 等价于 SetCaller
func SetCallerInfo(enabled bool) { SetCaller(enabled) }

// SetCallerSkip 设置用户侧额外跳过层数 (0-255), 0 为直接调用日志的用户帧
func SetCallerSkip(skip int) {
	if skip < 0 {
		skip = 0
	}
	if skip > 0xFF {
		skip = 0xFF
	}
	for {
		old := atomic.LoadInt32(&callerConfig)
		nb := (old &^ callerSkipMask) | int32(skip<<callerSkipShift)
		if atomic.CompareAndSwapInt32(&callerConfig, old, nb) {
			return
		}
	}
}

// isCallerImplFrame 判定 logger 实现帧, 名单须与内部调用链保持一致
// 不能按包名前缀过滤, 否则误杀包内测试与示例代码的用户帧
func isCallerImplFrame(name string) bool {
	switch name {
	case "github.com/mzzsfy/go-util/logger.newEvent",
		"github.com/mzzsfy/go-util/logger.emitFormat",
		"github.com/mzzsfy/go-util/logger.(*Log).event",
		"github.com/mzzsfy/go-util/logger.(*Log).Trace",
		"github.com/mzzsfy/go-util/logger.(*Log).Debug",
		"github.com/mzzsfy/go-util/logger.(*Log).Info",
		"github.com/mzzsfy/go-util/logger.(*Log).Warn",
		"github.com/mzzsfy/go-util/logger.(*Log).Error",
		"github.com/mzzsfy/go-util/logger.(*Log).Fatal",
		"github.com/mzzsfy/go-util/logger.(*Log).D",
		"github.com/mzzsfy/go-util/logger.(*Log).I",
		"github.com/mzzsfy/go-util/logger.(*Log).W",
		"github.com/mzzsfy/go-util/logger.(*Log).E",
		"github.com/mzzsfy/go-util/logger.(*Log).L",
		"github.com/mzzsfy/go-util/logger.(*Log).DF",
		"github.com/mzzsfy/go-util/logger.(*Log).IF",
		"github.com/mzzsfy/go-util/logger.(*Log).WF",
		"github.com/mzzsfy/go-util/logger.(*Log).EF",
		"github.com/mzzsfy/go-util/logger.(*Log).LF":
		return true
	}
	return false
}

// captureCaller 捕获用户侧调用者 PC
// 固定 skip 层数对编译器内联决策敏感, 内联状态随 GOARCH/版本变化, 故按名单过滤实现帧
// Callers 返回的 pc 为物理 pc, 必须经 CallersFrames 展开才是逻辑帧
func captureCaller(skip int) uintptr {
	var pcs [16]uintptr
	n := runtime.Callers(0, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	user := 0
	for i := 0; ; i++ {
		f, more := frames.Next()
		// 0=runtime.Callers, 1=captureCaller
		if i >= 2 && !isCallerImplFrame(f.Function) {
			if user == skip {
				return f.PC
			}
			user++
		}
		if !more {
			return 0
		}
	}
}

// callerCache 缓存 pc -> 格式化字节, 避免重复 runtime.FuncForPC
var callerCache sync.Map

// callerKey 为可比较结构体, 避免位压缩在 32 位平台溢出
type callerKey struct {
	pc       uintptr
	showFunc bool
}

// appendCaller 格式化调用者信息并追加到 buf
func appendCaller(buf []byte, pc uintptr, showFunc bool) []byte {
	if pc == 0 {
		return buf
	}
	key := callerKey{pc: pc, showFunc: showFunc}
	if v, ok := callerCache.Load(key); ok {
		return append(buf, v.([]byte)...)
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return buf
	}
	file, line := fn.FileLine(pc)
	// 只保留文件名
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			file = file[i+1:]
			break
		}
	}

	var result []byte
	if showFunc {
		name := fn.Name()
		for i := len(name) - 1; i >= 0; i-- {
			if name[i] == '.' {
				name = name[i+1:]
				break
			}
		}
		result = make([]byte, 0, len(name)+1+len(file)+8)
		result = append(result, name...)
		result = append(result, ' ')
	} else {
		result = make([]byte, 0, len(file)+8)
	}
	result = append(result, file...)
	result = append(result, ':')
	result = appendUint64(result, uint64(line))

	callerCache.Store(key, result)
	return append(buf, result...)
}
