package logger

import (
    "bytes"
    "context"
    "fmt"
    "github.com/mzzsfy/go-util/helper"
    "github.com/mzzsfy/go-util/pool"
    "io"
    "strings"
    "sync"
)

var (
    _ Log = (*logger)(nil)

    // PrintBlankTag 是否打印空白tag
    PrintBlankTag bool
    // PrintYearInfo 打印年份信息,0打印4位,1,打印2位,2不打印
    PrintYearInfo = 1
    // CompressedLogName 是否压缩日志名称
    CompressedLogName = true
    globalPlugin      []Plugin
)

func AddGlobalPlugin(plugin ...Plugin) {
    globalPlugin = append(globalPlugin, plugin...)
}
func GlobalPlugins() []Plugin {
    r := make([]Plugin, len(globalPlugin))
    copy(r, globalPlugin)
    return r
}
func CleanGlobalPlugin() {
    globalPlugin = nil
}

type logger struct {
    level    *Level
    tag      rune
    showName string
    fullName string
    parent   *logger
    plugin   []Plugin
    context  context.Context
    //writer   io.Writer
}

func (l *logger) FullName() string {
    return l.fullName
}

func (l *logger) target() io.Writer {
    //if l.writer != nil {
    //    return l.writer
    //}
    //if l.parent != nil {
    //    return l.parent.target()
    //}
    return writerTarget
}
func (l *logger) Level() Level {
    if l.level != nil && *l.level >= 0 {
        return *l.level
    }
    if l.parent != nil {
        return l.parent.Level()
    }
    return *defaultLevel
}
func (l *logger) SetLevel(lv *Level) Log {
    if lv != nil {
        if l.level == nil {
            l.level = new(Level)
        }
        *l.level = *lv
    } else {
        *l.level = -1
    }
    return l
}

// T Trace
func (l *logger) T(message string, args ...any) Log { return l.L(TraceLevel, message, args...) }

// TF Trace 懒加载
func (l *logger) TF(message string, f func() []any) Log { return l.LF(TraceLevel, message, f) }

// D debug
func (l *logger) D(message string, args ...any) Log { return l.L(DebugLevel, message, args...) }

// DF debug 懒加载
func (l *logger) DF(message string, f func() []any) Log { return l.LF(DebugLevel, message, f) }

// I info
func (l *logger) I(message string, args ...any) Log { return l.L(InfoLevel, message, args...) }

// IF info 懒加载
func (l *logger) IF(message string, f func() []any) Log { return l.LF(InfoLevel, message, f) }

// W warn
func (l *logger) W(message string, args ...any) Log { return l.L(WarnLevel, message, args...) }

// WF warn 懒加载
func (l *logger) WF(message string, f func() []any) Log { return l.LF(WarnLevel, message, f) }

// E error
func (l *logger) E(message string, args ...any) Log { return l.L(ErrorLevel, message, args...) }

// EF error 懒加载
func (l *logger) EF(message string, f func() []any) Log { return l.LF(ErrorLevel, message, f) }

// L 打印日志
func (l *logger) L(lv Level, message string, args ...any) Log {
    if l.Level() <= lv {
        l.doLog(lv, message, args)
    }
    return l
}

// LF 打印日志
func (l *logger) LF(lv Level, message string, f func() []any) Log {
    if l.Level() <= lv {
        l.doLog(lv, message, f())
    }
    return l
}

func (l *logger) Plugin() []Plugin {
    return l.plugin
}

// WithPlugin 产生一个新的Log,当输出日志前,会尝试调用该方法,多次调用只保留最后一个
func (l *logger) WithPlugin(p Plugin) Log {
    if p == nil {
        return l
    }
    nl := logPool.Get()
    *nl = *l
    nl.plugin = append(nl.plugin, p)
    return nl
}

func (l *logger) WithContext(ctx context.Context) Log {
    if ctx == nil && l.context == nil {
        return l
    }
    nl := logPool.Get()
    *nl = *l
    nl.context = ctx
    return nl
}

func (l *logger) UnUse() {
    if l.context == nil && l.plugin == nil {
        return
    }
    if Logger(l.FullName()) == l {
        return
    }
    logPool.Put(l)
}

func (l *logger) Context() context.Context {
    if l.context != nil {
        return l.context
    } else if l.parent != nil {
        return l.parent.Context()
    }
    return nil
}

func (l *logger) doLog(lv Level, message string, args []any) {
    target := l.target()
    w, aw := target.(helper.AsyncWriter)
    format := bfPool.Get()
    if !aw {
        defer bfPool.Put(format)
    }
    if len(args) != 0 {
        if strings.Contains(message, bigParenthesisString) || !strings.Contains(message, percentSignString) {
            doLogFormat1(format, message, args)
        } else {
            doLogFormat2(format, message, args)
        }
    } else {
        format.Write(helper.StringToBytes(message))
    }
    l.beforeWrite(lv, format)
    for i := 0; i < len(globalPlugin); i++ {
        if p, ok := globalPlugin[i].(PluginWrite); ok {
            p.BeforeWrite(lv, format, l, p)
        }
    }
    if format.Len() > 0 {
        start := FormatStart(l, lv)
        if !aw {
            defer bfPool.Put(start)
        }
        start.Write(format.Bytes())
        start.WriteByte('\n')
        if aw {
            w.WriterAsync(start.Bytes(), func() {
                bfPool.Put(start)
                bfPool.Put(format)
            })
        } else {
            target.Write(helper.StringToBytes(start.String()))
        }
    }
}

func (l *logger) beforeWrite(lv Level, format Buffer) {
    for _, plugin := range l.plugin {
        if p, ok := plugin.(PluginWrite); ok {
            p.BeforeWrite(lv, format, l, plugin)
        }
    }
    if l.parent != nil {
        l.parent.beforeWrite(lv, format)
    }
}

var bfPool = func() *pool.BytesPool {
    p := pool.NewBytesPool()
    p.SetMaxCap(128)
    return p
}()

func FormatStart(l *logger, lv Level) *pool.Bytes {
    s := bfPool.Get()
    AppendNowTime(s)
    s.WriteByte('[')
    AppendLoggerName(l, s)
    s.WriteByte(']')
    AppendLevel(lv, s)
    for _, plugin := range globalPlugin {
        if p, ok := plugin.(PluginAddSuffix); ok {
            p.AddSuffix(lv, s, l, plugin)
        }
    }
    s.WriteByte(':')
    s.WriteByte(' ')
    return s
}

func AppendLevel(lv Level, s Buffer) {
    s.Write(helper.StringToBytes(lv.String()))
}

func AppendLoggerName(l *logger, s Buffer) {
    if l.tag > 0 {
        if l.tag < 128 {
            s.WriteByte(byte(l.tag))
        } else {
            s.Write(helper.StringToBytes(string(l.tag)))
        }
    } else if PrintBlankTag {
        s.WriteByte(' ')
    }
    s.Write(helper.StringToBytes(l.showName))
}

var (
    percentSignString    = "%"
    bigParenthesisString = "{}"
    bigParenthesis       = helper.StringToBytes(bigParenthesisString)
)

//{}占位符风格
func doLogFormat1(s Buffer, format string, args []any) {
    //todo: 修改实现,性能有较大提升空间
    l := len(args)
    split := bytes.SplitN(helper.StringToBytes(format), bigParenthesis, -1)
    sl := len(split) - 1
    for i := 0; i < sl; i++ {
        s.Write(split[i])
        s.Write(FormatAny(args[i]))
    }
    s.Write(split[sl])
    if l > sl {
        if e, ok := args[l-1].(error); ok {
            args = args[:l-1]
            l = len(args)
            for i := sl; i < l; i++ {
                s.WriteByte(' ')
                s.Write(FormatAny(args[i]))
            }
            s.Write(helper.StringToBytes(fmt.Sprintf(" %+v", e)))
        } else {
            for i := sl; i < l; i++ {
                s.WriteByte(' ')
                s.Write(FormatAny(args[i]))
            }
        }
    }
}

//Format风格
func doLogFormat2(s Buffer, format string, args []any) {
    var err error
    l := len(args)
    if l > 0 {
        if e, ok := args[l-1].(error); ok {
            err = e
            args = args[:l-1]
        }
    }
    s.Write(helper.StringToBytes(fmt.Sprintf(format, args...)))
    if err != nil {
        s.WriteByte(' ')
        s.Write(helper.StringToBytes(fmt.Sprintf("%+v", err)))
    }
}

func FormatAny(arg any) []byte {
    if arg == nil {
        return helper.StringToBytes("<nil>")
    }

    // Some types can be done without reflection.
    switch v := arg.(type) {
    case []byte:
        return v
    case string:
        return helper.StringToBytes(v)
    case bool:
        if v {
            return helper.StringToBytes("true")
        }
        return helper.StringToBytes("false")
    case int:
        return formatInteger(v)
    case int8:
        return formatInteger(v)
    case int16:
        return formatInteger(v)
    case int32:
        return formatInteger(v)
    case int64:
        return formatInteger(v)
    case uint:
        return formatInteger(v)
    case uint8:
        return formatInteger(v)
    case uint16:
        return formatInteger(v)
    case uint32:
        return formatInteger(v)
    case uint64:
        return formatInteger(v)
    case uintptr:
        return formatInteger(v)
    default:
        return helper.StringToBytes(fmt.Sprint(v))
    }
}

//18446744073709551615 uint64
//-9223372036854775808 int64
var buf = sync.Pool{
    New: func() any {
        return &[20]byte{}
    },
}

func formatInteger[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) []byte {
    bs := buf.Get().(*[20]byte)
    defer buf.Put(bs)
    r := bs[:]
    negative := false
    i := len(r) - 1
    if n < 0 {
        n = -n
        //溢出,特殊处理
        // math.MinInt8
        // math.MinInt16
        // math.MinInt32
        // math.MinInt64
        if n < 0 {
            next := n / 10
            r[i] = byte('0' + -(n - next*10))
            n = -next
            i--
        }
        negative = true
    }
    for n > 9 {
        next := n / 10
        r[i] = byte('0' + n - next*10)
        n = next
        i--
    }
    r[i] = byte('0' + n)
    if negative {
        i--
        r[i] = '-'
    }
    return r[i:]
}
