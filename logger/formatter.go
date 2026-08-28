package logger

import (
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"
)

// Formatter 负责编码日志各部分到缓冲区
// 默认提供 ConsoleFormatter (控制台易读) 和 JSONFormatter (JSON 行) 两种实现
// 自定义实现可嵌入 ConsoleFormatter 只覆盖需要的方法
type Formatter interface {
	// Begin 写日志头: 时间、级别、logger名、预设字段
	Begin(buf *[]byte, lv Level, name string, context string)
	// 类型安全字段方法, 性能版链式 API 委托
	Str(buf *[]byte, key, val string)
	Int(buf *[]byte, key string, val int)
	Int64(buf *[]byte, key string, val int64)
	Uint64(buf *[]byte, key string, val uint64)
	Float64(buf *[]byte, key string, val float64)
	Bool(buf *[]byte, key string, val bool)
	Time(buf *[]byte, key string, val time.Time)
	Dur(buf *[]byte, key string, val time.Duration)
	Err(buf *[]byte, err error)
	// 通用字段方法, 便捷版 API 委托
	Any(buf *[]byte, key string, val any)
	// Msg 写消息文本
	Msg(buf *[]byte, msg string)
	// End 写日志尾: caller、换行
	End(buf *[]byte, caller uintptr, callerFn bool)
}

// --- 全局 Formatter 配置 ---

var defaultFormatterPtr unsafe.Pointer // *formatterBox

type formatterBox struct {
	f Formatter
}

func init() {
	atomic.StorePointer(&defaultFormatterPtr, unsafe.Pointer(&formatterBox{f: ConsoleFormatter{}}))
}

func loadDefaultFormatter() Formatter {
	return (*formatterBox)(atomic.LoadPointer(&defaultFormatterPtr)).f
}

// SetFormatter 设置全局默认 Formatter
func SetFormatter(f Formatter) {
	if f == nil {
		f = ConsoleFormatter{}
	}
	atomic.StorePointer(&defaultFormatterPtr, unsafe.Pointer(&formatterBox{f: f}))
}

// DefaultFormatter 返回当前全局默认 Formatter
func DefaultFormatter() Formatter {
	return loadDefaultFormatter()
}

// --- ConsoleFormatter: 控制台易读模式 ---
// 格式: 时间[名]级别: key=value ... msg caller:line
// 例:   2026-08-05 14:30:00.123[  app]I: user=moke id=42 login handler.go:42

// ConsoleFormatter 控制台易读格式
type ConsoleFormatter struct {
	// NameWidth logger 名称显示宽度, 0=自适应, 默认18
	NameWidth int
}

// defaultNameWidth 默认名称显示宽度
const defaultNameWidth = 18

func (f ConsoleFormatter) nw() int {
	if f.NameWidth > 0 {
		return f.NameWidth
	}
	return defaultNameWidth
}

func (f ConsoleFormatter) Begin(buf *[]byte, lv Level, name string, context string) {
	*buf = appendTime(*buf)
	nw := f.nw()
	*buf = append(*buf, '[')
	if len(name) > nw {
		// 超长截断
		*buf = append(*buf, name[:nw]...)
	} else {
		for i := len(name); i < nw; i++ {
			*buf = append(*buf, ' ')
		}
		*buf = append(*buf, name...)
	}
	*buf = append(*buf, ']')
	if len(context) > 0 {
		*buf = append(*buf, ' ')
		*buf = append(*buf, context...)
	}
}

func (f ConsoleFormatter) Str(buf *[]byte, key, val string) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = append(*buf, val...)
}

func (f ConsoleFormatter) Int(buf *[]byte, key string, val int) {
	f.Int64(buf, key, int64(val))
}

func (f ConsoleFormatter) Int64(buf *[]byte, key string, val int64) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = appendInt64(*buf, val)
}

func (f ConsoleFormatter) Uint64(buf *[]byte, key string, val uint64) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = appendUint64(*buf, val)
}

func (f ConsoleFormatter) Float64(buf *[]byte, key string, val float64) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = strconv.AppendFloat(*buf, val, 'f', -1, 64)
}

func (f ConsoleFormatter) Bool(buf *[]byte, key string, val bool) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = appendBool(*buf, val)
}

func (f ConsoleFormatter) Time(buf *[]byte, key string, val time.Time) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = val.AppendFormat(*buf, "01-02 15:04:05")
}

func (f ConsoleFormatter) Dur(buf *[]byte, key string, val time.Duration) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = appendDuration(*buf, val)
}

func (f ConsoleFormatter) Err(buf *[]byte, err error) {
	if err == nil {
		return
	}
	*buf = append(*buf, ' ')
	*buf = append(*buf, "err="...)
	*buf = append(*buf, err.Error()...)
}

func (f ConsoleFormatter) Any(buf *[]byte, key string, val any) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, key...)
	*buf = append(*buf, '=')
	*buf = appendAny(*buf, val)
}

func (f ConsoleFormatter) Msg(buf *[]byte, msg string) {
	*buf = append(*buf, ' ')
	*buf = append(*buf, msg...)
}

func (f ConsoleFormatter) End(buf *[]byte, caller uintptr, callerFn bool) {
	if caller != 0 {
		*buf = append(*buf, ' ')
		*buf = appendCaller(*buf, caller, callerFn)
	}
	*buf = append(*buf, '\n')
}

// --- JSONFormatter: JSON 行模式 ---
// 格式: {"t":"2026-08-05T14:30:00.123","lv":"info","logger":"app","msg":"login","user":"moke"}
// 字段顺序: t, lv, logger, [context字段], [链式字段], msg, [caller]

// JSONFormatter JSON 行格式
type JSONFormatter struct{}

func (JSONFormatter) Begin(buf *[]byte, lv Level, name string, context string) {
	*buf = append(*buf, '{')
	// 时间
	*buf = append(*buf, `"t":"`...)
	*buf = appendJSONTime(*buf)
	*buf = append(*buf, `","lv":"`...)
	*buf = append(*buf, lv.String()...)
	*buf = append(*buf, `","logger":`...)
	*buf = appendJSONString(*buf, name)
	if len(context) > 0 {
		// context 已包含 JSON 字段或 console 格式字段
		// JSON 模式下 context 是 JSON 字段, 需以逗号开头
		// 由 Logger.With 构建时保证格式
	}
}

func (JSONFormatter) Str(buf *[]byte, key, val string) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = appendJSONString(*buf, val)
}

func (JSONFormatter) Int(buf *[]byte, key string, val int) {
	JSONFormatter{}.Int64(buf, key, int64(val))
}

func (JSONFormatter) Int64(buf *[]byte, key string, val int64) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = appendInt64(*buf, val)
}

func (JSONFormatter) Uint64(buf *[]byte, key string, val uint64) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = appendUint64(*buf, val)
}

func (JSONFormatter) Float64(buf *[]byte, key string, val float64) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = strconv.AppendFloat(*buf, val, 'f', -1, 64)
}

func (JSONFormatter) Bool(buf *[]byte, key string, val bool) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = appendBool(*buf, val)
}

func (JSONFormatter) Time(buf *[]byte, key string, val time.Time) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, `:"`...)
	*buf = val.AppendFormat(*buf, time.RFC3339Nano)
	*buf = append(*buf, '"')
}

func (JSONFormatter) Dur(buf *[]byte, key string, val time.Duration) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = appendInt64(*buf, int64(val))
}

func (JSONFormatter) Err(buf *[]byte, err error) {
	if err == nil {
		return
	}
	*buf = append(*buf, ',')
	*buf = append(*buf, `"err":`...)
	*buf = appendJSONString(*buf, err.Error())
}

func (JSONFormatter) Any(buf *[]byte, key string, val any) {
	*buf = append(*buf, ',')
	*buf = appendJSONString(*buf, key)
	*buf = append(*buf, ':')
	*buf = appendJSONAny(*buf, val)
}

func (JSONFormatter) Msg(buf *[]byte, msg string) {
	*buf = append(*buf, `,"msg":`...)
	*buf = appendJSONString(*buf, msg)
}

func (JSONFormatter) End(buf *[]byte, caller uintptr, callerFn bool) {
	if caller != 0 {
		// caller 在 JSON 中作为对象字段
		*buf = append(*buf, `,"caller":"`...)
		*buf = appendCallerRaw(*buf, caller, callerFn)
		*buf = append(*buf, '"')
	}
	*buf = append(*buf, '}', '\n')
}

// --- JSON 编码辅助 ---

// appendJSONString 追加 JSON 字符串 (含引号), 转义特殊字符
func appendJSONString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			if c < 0x20 {
				buf = append(buf, '\\', 'u', '0', '0',
					"0123456789abcdef"[c>>4],
					"0123456789abcdef"[c&0xF])
			} else {
				buf = append(buf, c)
			}
		}
	}
	return append(buf, '"')
}

// appendJSONAny 追加 JSON 值 (类型分发)
func appendJSONAny(buf []byte, v any) []byte {
	if v == nil {
		return append(buf, "null"...)
	}
	switch val := v.(type) {
	case string:
		return appendJSONString(buf, val)
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
	case float32:
		return strconv.AppendFloat(buf, float64(val), 'f', -1, 32)
	case float64:
		return strconv.AppendFloat(buf, val, 'f', -1, 64)
	case []byte:
		return appendJSONString(buf, string(val))
	case error:
		return appendJSONString(buf, val.Error())
	default:
		return appendJSONString(buf, toStr(v))
	}
}

// appendJSONTime 追加 RFC3339 时间 (不含引号)
func appendJSONTime(buf []byte) []byte {
	now := time.Now()
	return now.AppendFormat(buf, "2006-01-02T15:04:05.000")
}

// appendCallerRaw 追加 caller 信息 (不含引号), 供 JSONFormatter.End 使用
func appendCallerRaw(buf []byte, pc uintptr, showFunc bool) []byte {
	return appendCaller(buf, pc, showFunc)
}

// toStr 简易 any → string
func toStr(v any) string {
	if s, ok := v.(interface{ String() string }); ok {
		return s.String()
	}
	return ""
}
