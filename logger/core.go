package logger

import (
    "io"
    "os"
    "strconv"
    "sync/atomic"

    "github.com/mzzsfy/go-util/helper"
    "github.com/mzzsfy/go-util/pool"
)

const (
    // TraceLevel 记录细节,参数等信息
    TraceLevel Level = iota
    // DebugLevel 记录重要的参数等信息
    DebugLevel
    // InfoLevel 记录信息用户易读信息
    InfoLevel
    // WarnLevel 记录用户需要关注信息
    WarnLevel
    // ErrorLevel 记录用户需要处理的信息
    ErrorLevel
    // FatalLevel 致命错误,一般影响到系统正常运行
    FatalLevel

    // LevelUnset 哨兵值: 清除本地设置, 继承父级或默认级别
    LevelUnset Level = -1
)

type Level int8

var (
    // levelByte 短名称查找表, 索引对应 Level 值
    levelByte = [6]byte{'T', 'D', 'I', 'W', 'E', 'F'}
    // levelFull 全名查找表, 索引对应 Level 值
    levelFull = [6]string{"Trace", "Debug", "Info", "Warn", "Error", "Fatal"}
)

func (l Level) String() string {
    if l == LevelUnset {
        return "-"
    }
    if l >= 0 && int(l) < len(levelByte) {
        return string(levelByte[l])
    }
    if l > FatalLevel {
        return "F"
    }
    return "T"
}

func (l Level) FullName() string {
    if l == LevelUnset {
        return "Unset"
    }
    if l >= 0 && int(l) < len(levelFull) {
        return levelFull[l]
    }
    if l > FatalLevel {
        return "Fatal"
    }
    return "Trace"
}

func (l *Level) UnmarshalJSON(b []byte) error {
    s, err := strconv.Unquote(helper.BytesToString(b))
    if err != nil {
        return err
    }
    *l = FromString(s)
    return nil
}

func (l *Level) UnmarshalBinary(b []byte) error {
    *l = FromString(helper.BytesToString(b))
    return nil
}

func (l *Level) UnmarshalText(b []byte) error {
    return l.UnmarshalBinary(b)
}

func (l Level) MarshalJSON() ([]byte, error) {
    return helper.StringToBytes(`"` + l.String() + `"`), nil
}
func (l Level) MarshalYAML() (any, error) {
    return l.String(), nil
}
func (l Level) MarshalBinary() ([]byte, error) {
    return helper.StringToBytes(l.String()), nil
}
func (l Level) MarshalText() ([]byte, error) {
    return l.MarshalBinary()
}

func FromString(s string) Level {
    if s == "" {
        return InfoLevel
    }
    // 特殊处理: "-" 表示 LevelUnset
    if s == "-" {
        return LevelUnset
    }
    // 通过 | 0x20 将首字符转为小写,统一处理大小写
    b := s[0] | 0x20
    switch b {
    case 'i':
        return InfoLevel
    case 'd':
        return DebugLevel
    case 't':
        return TraceLevel
    case 'w':
        return WarnLevel
    case 'e':
        return ErrorLevel
    case 'f':
        return FatalLevel
    case 'u':
        return LevelUnset
    default:
        return InfoLevel
    }
}

var (
    // defaultLevel 使用 atomic int32 存储 Level, 避免 interface{} 装箱开销
    defaultLevel int32
    // writerTarget 使用 atomic.Value 存储 *writerTargetValue 指针, 避免每次 Load 时拷贝 40 字节 struct
    writerTarget atomic.Value
)

// LevelControl 级别控制接口, 用于运行时调整日志级别
// 传递 LevelUnset 表示清除本地设置, 继承父级或默认级别
type LevelControl interface {
    Level() Level
    SetLevel(lv Level)
}

// Hook 日志行钩子, 在写入输出前调用, 可直接操作缓冲区
type Hook func(lv Level, buf *pool.Bytes, log *Log)

// KvFormatter kvs 值的自定义格式化器
// 返回的字符串作为日志行的前缀或后缀, 返回空字符串则不添加
type KvFormatter func(value any) string

// Log 日志记录器类型
type Log = logger

// writerTargetValue 存储 writer 及其异步能力缓存
// 通过指针存储在 atomic.Value 中, 热路径 Load 时只拷贝指针不拷贝 struct
type writerTargetValue struct {
    writer  io.Writer
    isAsync bool
    asyncW  helper.AsyncWriter
}

func init() {
    atomic.StoreInt32(&defaultLevel, int32(InfoLevel))
    wt := newWriterTargetValue(os.Stdout)
    writerTarget.Store(&wt)
}

func DefaultWriterTarget() io.Writer {
    return writerTarget.Load().(*writerTargetValue).writer
}

// newWriterTargetValue 构建 writerTargetValue, 在设置时缓存异步写能力
func newWriterTargetValue(w io.Writer) writerTargetValue {
    v := writerTargetValue{writer: w}
    if aw, ok := w.(helper.AsyncWriter); ok {
        v.isAsync = true
        v.asyncW = aw
    }
    return v
}

func SetDefaultWriterTarget(w io.Writer) {
    if w == nil {
        w = os.Stdout
    }
    wt := newWriterTargetValue(w)
    writerTarget.Store(&wt)
}

// SetDefaultLogLevel 设置默认日志级别, 返回是否成功
// 不接受 LevelUnset 等负值, 默认级别必须是有效级别
func SetDefaultLogLevel(l Level) bool {
    if l < 0 {
        return false
    }
    atomic.StoreInt32(&defaultLevel, int32(l))
    atomic.AddInt32(&levelGeneration, 1)
    return true
}
func DefaultLogLevel() Level {
    return Level(atomic.LoadInt32(&defaultLevel))
}

// SetCallerInfo 设置是否输出调用者信息
func SetCallerInfo(enabled bool) {
    for {
        old := atomic.LoadInt32(&callerConfig)
        // 清除 bit16, 根据 enabled 设置
        newCfg := old &^ (1 << 16)
        if enabled {
            newCfg |= 1 << 16
        }
        if atomic.CompareAndSwapInt32(&callerConfig, old, newCfg) {
            return
        }
    }
}

// SetCallerSkip 设置调用者跳过层数 (0-65535)
func SetCallerSkip(skip int) {
    if skip < 0 {
        skip = 0
    }
    if skip > 0xFFFF {
        skip = 0xFFFF
    }
    for {
        old := atomic.LoadInt32(&callerConfig)
        // 保留高16位, 替换低16位
        newCfg := (old &^ 0xFFFF) | int32(skip)
        if atomic.CompareAndSwapInt32(&callerConfig, old, newCfg) {
            return
        }
    }
}

// SetCallerFunc 设置是否显示函数名
func SetCallerFunc(enabled bool) {
    for {
        old := atomic.LoadInt32(&callerConfig)
        // 清除 bit17, 根据 enabled 设置
        newCfg := old &^ (1 << 17)
        if enabled {
            newCfg |= 1 << 17
        }
        if atomic.CompareAndSwapInt32(&callerConfig, old, newCfg) {
            return
        }
    }
}
