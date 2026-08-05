package logger

import (
    "context"
    "fmt"
    "io"
    "sync/atomic"
)

// levelInherit 哨兵值: 本地级别未设置, 继承父级或全局默认
const levelInherit int32 = -1

// Logger 日志记录器
// 并发安全: resolved 用 atomic 读, SetLevel 递增 globalGen 传播变更
// 独立 Logger (New 创建) parent=nil, 级别独立无继承开销
// 命名 Logger (Get 创建) 有 parent 链, localLv=levelInherit 时继承父级
type Logger struct {
    name     string
    writer   io.Writer
    context  string    // 预编码字段前缀 "k=v k2=v2 ", 空=无
    parent   *Logger   // 命名继承链, nil=独立
    localLv  int32     // 本地级别, levelInherit=继承
    resolved int32     // atomic, 缓存的有效级别
    gen      int32     // atomic, 上次 resolved 更新时的 generation
    isAsync  bool             // writer 是否实现 asyncWriter, flush 中避免 type assertion
    fmt      Formatter        // 编码器, 从全局快照
    ctx      context.Context  // 请求级上下文, WithContext 注入, 默认 nil 无开销
}

// Option Logger 配置选项
type Option func(*Logger)

// WithLevel 设置 Logger 级别
func WithLevel(lv Level) Option {
    return func(l *Logger) {
        l.localLv = int32(lv)
        atomic.StoreInt32(&l.resolved, int32(lv))
    }
}

// WithWriter 设置 Logger 输出目标, 传 nil 忽略
func WithWriter(w io.Writer) Option {
    return func(l *Logger) {
        if w != nil {
            l.writer = w
            _, l.isAsync = w.(asyncWriter)
        }
    }
}

// WithFormatter 设置 Logger 编码器, 传 nil 忽略
func WithFormatter(f Formatter) Option {
    return func(l *Logger) {
        if f != nil {
            l.fmt = f
        }
    }
}

// New 创建独立 Logger, 从全局默认快照级别、输出目标和编码器
// parent=nil, 不参与命名继承
func New(name string, opts ...Option) *Logger {
    lv := atomic.LoadInt32(&defaultLevel)
    g := atomic.LoadInt32(&globalGen)
    l := &Logger{
        name:     name,
        writer:   loadDefaultWriter(),
        localLv:  lv,
        resolved: lv,
        gen:      g,
        fmt:      loadDefaultFormatter(),
    }
    for _, opt := range opts {
        opt(l)
    }
    _, l.isAsync = l.writer.(asyncWriter)
    return l
}

// --- 级别方法 ---

func (l *Logger) Trace() *Event { return l.event(TraceLevel) }
func (l *Logger) Debug() *Event { return l.event(DebugLevel) }
func (l *Logger) Info() *Event  { return l.event(InfoLevel) }
func (l *Logger) Warn() *Event  { return l.event(WarnLevel) }
func (l *Logger) Error() *Event { return l.event(ErrorLevel) }

// Fatal 创建 Fatal 级别事件, Msg/Msgf/Send 后触发 os.Exit(1)
func (l *Logger) Fatal() *Event {
    e := l.event(FatalLevel)
    if e.enabled {
        e.exitAfter = true
    }
    return e
}

// event 级别检查 + 事件创建, 过滤时返回 disabled 单例
// 快速路径内联: 独立 Logger (New/With/SetLevel) 不走 level() 函数调用
func (l *Logger) event(lv Level) *Event {
    local := atomic.LoadInt32(&l.localLv)
    if local != levelInherit {
        if Level(local) > lv {
            return disabledEvent
        }
    } else if l.level() > lv {
        return disabledEvent
    }
    return newEvent(l, lv)
}

// level O(1) 级别解析
// 快路径: localLv 非 inherit 时直接返回 (New/With/SetLevel 创建的 Logger)
// 慢路径: inherit 时检查 gen, 决定是否遍历父链 (Get 创建的命名 Logger)
func (l *Logger) level() Level {
    lv := atomic.LoadInt32(&l.localLv)
    if lv != levelInherit {
        return Level(lv)
    }
    if atomic.LoadInt32(&l.gen) == atomic.LoadInt32(&globalGen) {
        return Level(atomic.LoadInt32(&l.resolved))
    }
    return l.resolveLevel()
}

// resolveLevel 慢路径: 遍历父链解析有效级别, 更新缓存
func (l *Logger) resolveLevel() Level {
    lv := atomic.LoadInt32(&l.localLv)
    if lv == levelInherit {
        if l.parent != nil {
            lv = int32(l.parent.level())
        } else {
            lv = atomic.LoadInt32(&defaultLevel)
        }
    }
    g := atomic.LoadInt32(&globalGen)
    atomic.StoreInt32(&l.resolved, lv)
    atomic.StoreInt32(&l.gen, g)
    return Level(lv)
}

// Enabled 检查指定级别是否会产生输出
func (l *Logger) Enabled(lv Level) bool {
    return l.level() <= lv
}

// SetLevel 设置 Logger 级别, 递增 globalGen 通知所有命名 Logger 重新解析
func (l *Logger) SetLevel(lv Level) {
    atomic.StoreInt32(&l.localLv, int32(lv))
    atomic.StoreInt32(&l.resolved, int32(lv))
    atomic.AddInt32(&globalGen, 1)
}

// Level 返回当前有效级别
func (l *Logger) Level() Level {
    return l.level()
}

// Name 返回 Logger 名称
func (l *Logger) Name() string {
    return l.name
}

// --- Logger 派生 ---

// With 返回字段构建器, 用于创建带预设字段的 Logger
//
//	l2 := l.With().Str("svc", "api").Int("ver", 2).Logger()
func (l *Logger) With() *Event {
    return newWithEvent(l)
}

// newWithEvent 创建用于构建预设字段的 Event (不写头部, 只写 context)
func newWithEvent(l *Logger) *Event {
    e := eventPool.Get().(*Event)
    e.lg = l
    e.enabled = true
    e.buf = e.buf[:0]
    e.fmt = l.fmt
    if len(l.context) > 0 {
        e.buf = append(e.buf, l.context...)
    }
    return e
}

// --- 旧版命令式便捷方法 ---

// emitFormat 便捷日志内部: 先检查级别, 再格式化占位符消息
// 级别不够时跳过格式化, 避免无谓分配
func emitFormat(l *Logger, lv Level, msg string, args []any) {
    if len(args) == 0 {
        l.event(lv).Msg(msg)
        return
    }
    if !l.Enabled(lv) {
        return
    }
    e := l.event(lv)
    if !e.enabled {
        return
    }
    e.MsgFormat(msg, args)
}

// D 便捷Debug日志, msg支持{}占位符, 性能版用 Debug().Str().Msg()
func (l *Logger) D(msg string, args ...any) *Logger {
    emitFormat(l, DebugLevel, msg, args)
    return l
}

// I 便捷Info日志, msg支持{}占位符, 性能版用 Info().Str().Msg()
func (l *Logger) I(msg string, args ...any) *Logger {
    emitFormat(l, InfoLevel, msg, args)
    return l
}

// E 便捷Error日志, msg支持{}占位符, 性能版用 Error().Str().Msg()
func (l *Logger) E(msg string, args ...any) *Logger {
    emitFormat(l, ErrorLevel, msg, args)
    return l
}

// L 便捷指定级别日志, msg支持{}占位符, 性能版用 Trace()/Debug()/...().Str().Msg()
func (l *Logger) L(lv Level, msg string, args ...any) *Logger {
    emitFormat(l, lv, msg, args)
    return l
}

// DF 延迟参数Debug日志, f返回占位符参数, 级别不够时f不会被调用
func (l *Logger) DF(msg string, f func() []any) *Logger {
    if !l.Enabled(DebugLevel) {
        return l
    }
    emitFormat(l, DebugLevel, msg, f())
    return l
}

// IF 延迟参数Info日志, f返回占位符参数, 级别不够时f不会被调用
func (l *Logger) IF(msg string, f func() []any) *Logger {
    if !l.Enabled(InfoLevel) {
        return l
    }
    emitFormat(l, InfoLevel, msg, f())
    return l
}

// EF 延迟参数Error日志, f返回占位符参数, 级别不够时f不会被调用
func (l *Logger) EF(msg string, f func() []any) *Logger {
    if !l.Enabled(ErrorLevel) {
        return l
    }
    emitFormat(l, ErrorLevel, msg, f())
    return l
}

// LF 延迟参数指定级别日志, f返回占位符参数, 级别不够时f不会被调用
func (l *Logger) LF(lv Level, msg string, f func() []any) *Logger {
    if !l.Enabled(lv) {
        return l
    }
    emitFormat(l, lv, msg, f())
    return l
}

// --- 派生方法 ---

// WithContext 派生带请求上下文的Logger, 用于hook中提取请求级信息(如当前用户)
// ctx为nil时返回原Logger, 避免无谓拷贝
func (l *Logger) WithContext(ctx context.Context) *Logger {
    if ctx == nil {
        return l
    }
    nl := *l
    nl.ctx = ctx
    return &nl
}

// WithKvs 派生带预设字段的Logger, args为key-value对
// 奇数个参数时, 最后一个落单key配nil值
func (l *Logger) WithKvs(args ...any) *Logger {
    e := l.With()
    n := len(args) &^ 1
    for i := 0; i < n; i += 2 {
        e.Any(fmt.Sprint(args[i]), args[i+1])
    }
    if n < len(args) {
        e.Any(fmt.Sprint(args[n]), nil)
    }
    return e.Logger()
}

// Log 旧版类型别名, 兼容旧代码
type Log = Logger
