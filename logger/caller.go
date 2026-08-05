package logger

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// caller 配置位布局: bit0=enabled, bit1=showFunc, bit8-15=skip
const (
	callerEnabledBit  int32 = 1 << 0
	callerFuncBit     int32 = 1 << 1
	callerSkipShift         = 8
	callerSkipMask    int32 = 0xFF << callerSkipShift
	defaultCallerSkip       = 4
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

// SetCallerSkip 设置调用者跳过层数 (0-255)
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

// captureCaller 获取调用者 PC, skip 从 runtime.Callers 自身算起
func captureCaller(skip int) uintptr {
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])
	return pcs[0]
}

// callerCache 缓存 pc -> 格式化字节, 避免重复 runtime.FuncForPC
var callerCache sync.Map

// appendCaller 格式化调用者信息并追加到 buf
func appendCaller(buf []byte, pc uintptr, showFunc bool) []byte {
	if pc == 0 {
		return buf
	}
	key := pc
	if showFunc {
		key |= 1 << 63
	}
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
