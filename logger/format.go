package logger

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

// --- 时间格式化 ---

const (
	YearFull  int32 = 0
	YearShort int32 = 1
	YearNone  int32 = 2
)

var yearMode int32 = YearShort

func SetYearMode(mode int32) {
	atomic.StoreInt32(&yearMode, mode)
	atomic.StorePointer(&timeCachePtr, unsafe.Pointer(&emptyTimeCache))
}

// SetPrintYearInfo 旧版兼容, 等价于 SetYearMode
func SetPrintYearInfo(v int32) { SetYearMode(v) }

// timeCache 秒级缓存已格式化的时间前缀
type timeCache struct {
	unix   int64
	prefix []byte
}

var emptyTimeCache = timeCache{}

// timeCachePtr 用 unsafe.Pointer 替代 atomic.Value, 省类型检查开销
var timeCachePtr unsafe.Pointer // *timeCache

// millisTable 预计算毫秒字节, 避免运行时除法
var millisTable [1000][3]byte

// digits2Table 预计算两位十进制字节
var digits2Table [100][2]byte

func init() {
	for i := 0; i < 1000; i++ {
		millisTable[i][0] = byte('0' + i/100)
		millisTable[i][1] = byte('0' + i/10%10)
		millisTable[i][2] = byte('0' + i%10)
	}
	for i := 0; i < 100; i++ {
		digits2Table[i][0] = byte('0' + i/10)
		digits2Table[i][1] = byte('0' + i%10)
	}
	atomic.StorePointer(&timeCachePtr, unsafe.Pointer(&emptyTimeCache))
}

// appendTime 追加当前时间到 buf, 格式由 yearMode 决定
func appendTime(buf []byte) []byte {
	now := time.Now()
	unix := now.Unix()

	c := (*timeCache)(atomic.LoadPointer(&timeCachePtr))
	if c.unix == unix {
		buf = append(buf, c.prefix...)
		return append(buf, millisTable[now.Nanosecond()/1e6][:]...)
	}

	newCache := formatTimeCache(now, unix)
	if !atomic.CompareAndSwapPointer(&timeCachePtr, unsafe.Pointer(c), unsafe.Pointer(newCache)) {
		newCache = (*timeCache)(atomic.LoadPointer(&timeCachePtr))
	}
	buf = append(buf, newCache.prefix...)
	return append(buf, millisTable[now.Nanosecond()/1e6][:]...)
}

// formatTimeCache 格式化时间前缀, 年份部分由 yearMode 决定
func formatTimeCache(now time.Time, unix int64) *timeCache {
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	mode := atomic.LoadInt32(&yearMode)

	var buf []byte
	switch mode {
	case YearFull:
		buf = make([]byte, 0, 20)
		buf = append4digits(buf, year)
		buf = append(buf, '-')
	case YearShort:
		buf = make([]byte, 0, 18)
		buf = append(buf, digits2Table[year%100][:]...)
		buf = append(buf, '-')
	default:
		buf = make([]byte, 0, 15)
	}
	buf = append(buf, digits2Table[month][:]...)
	buf = append(buf, '-')
	buf = append(buf, digits2Table[day][:]...)
	buf = append(buf, ' ')
	buf = append(buf, digits2Table[hour][:]...)
	buf = append(buf, ':')
	buf = append(buf, digits2Table[min][:]...)
	buf = append(buf, ':')
	buf = append(buf, digits2Table[sec][:]...)
	buf = append(buf, '.')
	return &timeCache{unix: unix, prefix: buf}
}

func append4digits(buf []byte, v int) []byte {
	return append(buf,
		byte('0'+v/1000),
		byte('0'+v/100%10),
		byte('0'+v/10%10),
		byte('0'+v%10),
	)
}

// --- 整数格式化 ---
// 手写实现 (非 strconv) 使 Int/Int64/Uint64 方法可被编译器内联

// appendInt64 负数取反后 uint64 位回绕正确处理 MinInt64
func appendInt64(buf []byte, n int64) []byte {
	if n < 0 {
		buf = append(buf, '-')
		n = -n
	}
	return appendUint64(buf, uint64(n))
}

// appendUint64 整数转十进制追加
// noinline 控制 appendInt64 内联成本, 避免 appendUint64 体被递归展开
//
//go:noinline
func appendUint64(buf []byte, n uint64) []byte {
	if n < 10 {
		return append(buf, byte('0'+n))
	}
	var tmp [20]byte
	pos := len(tmp)
	for n >= 10 {
		pos--
		tmp[pos] = byte('0' + n%10)
		n /= 10
	}
	pos--
	tmp[pos] = byte('0' + n)
	return append(buf, tmp[pos:]...)
}

// --- 其他类型格式化 ---

func appendBool(buf []byte, b bool) []byte {
	if b {
		return append(buf, "true"...)
	}
	return append(buf, "false"...)
}

func appendFloat64(buf []byte, f float64) []byte {
	return strconv.AppendFloat(buf, f, 'f', -1, 64)
}

// appendAny 追加任意类型值, 热路径应优先用类型安全方法
func appendAny(buf []byte, v any) []byte {
	if v == nil {
		return append(buf, "<nil>"...)
	}
	switch val := v.(type) {
	case []byte:
		return append(buf, val...)
	case string:
		return append(buf, val...)
	case bool:
		return appendBool(buf, val)
	case int:
		return appendInt64(buf, int64(val))
	case int8:
		return appendInt64(buf, int64(val))
	case int16:
		return appendInt64(buf, int64(val))
	case int32:
		return appendInt64(buf, int64(val))
	case int64:
		return appendInt64(buf, val)
	case uint:
		return appendUint64(buf, uint64(val))
	case uint8:
		return appendUint64(buf, uint64(val))
	case uint16:
		return appendUint64(buf, uint64(val))
	case uint32:
		return appendUint64(buf, uint64(val))
	case uint64:
		return appendUint64(buf, val)
	case uintptr:
		return appendUint64(buf, uint64(val))
	case float32:
		return strconv.AppendFloat(buf, float64(val), 'f', -1, 32)
	case float64:
		return appendFloat64(buf, val)
	case error:
		return append(buf, val.Error()...)
	case fmt.Stringer:
		return append(buf, val.String()...)
	default:
		return append(buf, fmt.Sprint(v)...)
	}
}

// b2s 零分配 []byte → string, 共享底层数组
// 调用方需确保返回的 string 在原始 []byte 被修改前使用完毕
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// --- 占位符消息格式化 ---

// doFormatPlaceholders {}占位符风格
// 支持: {} 自动递增, {0} 显式位置, {%s} 格式化, {0%d} 显式+格式化
func doFormatPlaceholders(buf []byte, format string, args []any) []byte {
	argIdx := 0
	argLen := len(args)
	lastWrite := 0
	for i := 0; i < len(format); {
		if format[i] != '{' {
			i++
			continue
		}
		// 快速路径: {}
		if i+1 < len(format) && format[i+1] == '}' {
			if i > lastWrite {
				buf = append(buf, format[lastWrite:i]...)
			}
			if argIdx < argLen {
				buf = appendAny(buf, args[argIdx])
			} else {
				buf = append(buf, '{', '}')
			}
			argIdx++
			lastWrite = i + 2
			i += 2
			continue
		}
		// 慢路径: {0} {%s} {0%d} 等
		j := i + 1
		for j < len(format) && format[j] != '}' {
			j++
		}
		if j >= len(format) {
			break
		}
		idx, verb, ok := parsePlaceholder(format[i+1 : j])
		if !ok {
			i++
			continue
		}
		if i > lastWrite {
			buf = append(buf, format[lastWrite:i]...)
		}
		if idx >= 0 {
			// 显式位置
			if idx < argLen {
				buf = appendArg(buf, verb, args[idx])
			} else {
				buf = append(buf, format[i:j+1]...)
			}
		} else {
			// 自动递增
			if argIdx < argLen {
				buf = appendArg(buf, verb, args[argIdx])
			} else {
				buf = append(buf, format[i:j+1]...)
			}
			argIdx++
		}
		lastWrite = j + 1
		i = j + 1
	}
	if lastWrite < len(format) {
		buf = append(buf, format[lastWrite:]...)
	}
	// 追加剩余未消费参数 (仅自动递增的)
	if argIdx < argLen {
		for k := argIdx; k < argLen; k++ {
			buf = append(buf, ' ')
			buf = appendAny(buf, args[k])
		}
	}
	return buf
}

// parsePlaceholder 解析占位符内容
// 返回 idx(>=0=显式位置, -1=自动递增), verb(格式化动词, 空=appendAny), ok
func parsePlaceholder(s string) (idx int, verb string, ok bool) {
	if len(s) == 0 {
		return 0, "", false
	}
	pct := strings.IndexByte(s, '%')
	if pct < 0 {
		n := parseDigitIndex(s)
		if n >= 0 {
			return n, "", true
		}
		return 0, "", false
	}
	verb = s[pct:]
	if pct == 0 {
		return -1, verb, true
	}
	n := parseDigitIndex(s[:pct])
	if n >= 0 {
		return n, verb, true
	}
	return 0, "", false
}

// appendArg 追加单个参数, verb非空时用格式化
// 常见verb优化, 避免fmt.Sprintf分配
func appendArg(buf []byte, verb string, val any) []byte {
	if verb == "" {
		return appendAny(buf, val)
	}
	switch verb {
	case "%d":
		return appendVerbD(buf, val)
	case "%s":
		return appendVerbS(buf, val)
	case "%v":
		return appendAny(buf, val)
	case "%t":
		return appendBool(buf, val.(bool))
	case "%x":
		return strconv.AppendUint(buf, toUint64(val), 16)
	}
	return append(buf, fmt.Sprintf(verb, val)...)
}

func toUint64(val any) uint64 {
	switch v := val.(type) {
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint:
		return uint64(v)
	case uint64:
		return v
	}
	return 0
}

func appendVerbD(buf []byte, val any) []byte {
	switch v := val.(type) {
	case int:
		return appendInt64(buf, int64(v))
	case int8:
		return appendInt64(buf, int64(v))
	case int16:
		return appendInt64(buf, int64(v))
	case int32:
		return appendInt64(buf, int64(v))
	case int64:
		return appendInt64(buf, v)
	case uint:
		return appendUint64(buf, uint64(v))
	case uint8:
		return appendUint64(buf, uint64(v))
	case uint16:
		return appendUint64(buf, uint64(v))
	case uint32:
		return appendUint64(buf, uint64(v))
	case uint64:
		return appendUint64(buf, v)
	}
	return append(buf, fmt.Sprintf("%d", val)...)
}

func appendVerbS(buf []byte, val any) []byte {
	switch v := val.(type) {
	case string:
		return append(buf, v...)
	case []byte:
		return append(buf, v...)
	case error:
		return append(buf, v.Error()...)
	}
	return append(buf, fmt.Sprintf("%s", val)...)
}

func parseDigitIndex(s string) int {
	if len(s) == 0 {
		return -1
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
		if n > 100 {
			return -1
		}
	}
	return n
}
