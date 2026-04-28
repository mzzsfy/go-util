package logger

import (
    "reflect"
    "runtime"
    "strings"
    "sync"
    "sync/atomic"
    "unsafe"

    "github.com/mzzsfy/go-util/helper"
    "github.com/mzzsfy/go-util/pool"
)

var (
    _ LevelControl = (*logger)(nil)

    // 全局配置
    printYearInfo     int32 = 1 // 0=4位年份, 1=2位, 2=无年份
    compressedLogName int32 = 1
    // caller 配置: 低16位=skip层数(默认2), bit16=启用caller, bit17=显示函数名
    callerConfig int32 = 0x00002 // skip=2, caller=off, func=off

    // Hook 存储
    globalHooksMu sync.Mutex
    globalHooks   atomic.Value // []Hook
    hasHooks      int32        // 是否有注册的钩子, 快速跳过无钩子时的 atomic.Load

    levelGeneration int32 // 级别变更时递增, atomic.LoadInt32/AddInt32

    // kvs 格式化器
    kvFmtMap atomic.Value // map[string]KvFormatter
    hasKvFmt int32        // 快速跳过: 无注册格式化器时跳过查找
    kvFmtMu  sync.Mutex   // 保护 kvFmtMap 写入

    // caller 缓存
    callerCache   map[uintptr]string
    callerCacheMu sync.RWMutex
)

const (
    callerCacheOrigMaxSize = 1024 // 缓存原始上限
    callerCacheMaxMul      = 8    // 最大可扩展到原始上限的8倍
)

var (
    callerCacheMaxSize = callerCacheOrigMaxSize // 动态缓存上限, 可自适应扩容
)

func init() {
    globalHooks.Store([]Hook(nil))
    kvFmtMap.Store(map[string]KvFormatter(nil))
    callerCache = make(map[uintptr]string, 1024)
}

// SetPrintYearInfo 设置年份打印格式: 0=4位年份, 1=2位, 2=无年份
func SetPrintYearInfo(v int32) {
    atomic.StoreInt32(&printYearInfo, v)
}

// SetCompressedLogName 设置是否压缩长日志名
func SetCompressedLogName(v int32) {
    atomic.StoreInt32(&compressedLogName, v)
}

// maxHooks 最大hook数量, 防止无限添加导致每次日志调用开销线性增长
const maxHooks = 64

// AddHook 注册日志钩子, 最多允许 maxHooks 个
// 返回是否成功, 超过数量限制时返回 false
func AddHook(hooks ...Hook) bool {
    if len(hooks) == 0 {
        return true
    }
    globalHooksMu.Lock()
    current := globalHooks.Load().([]Hook)
    if len(current)+len(hooks) > maxHooks {
        globalHooksMu.Unlock()
        return false
    }
    next := make([]Hook, 0, len(current)+len(hooks))
    next = append(next, current...)
    next = append(next, hooks...)
    globalHooks.Store(next)
    atomic.StoreInt32(&hasHooks, 1)
    globalHooksMu.Unlock()
    return true
}

// CleanHooks 清理所有全局钩子
func CleanHooks() {
    globalHooksMu.Lock()
    globalHooks.Store([]Hook(nil))
    atomic.StoreInt32(&hasHooks, 0)
    globalHooksMu.Unlock()
}

// SetKvPrefix 注册 kvs 中 key 的前缀格式化器
// 当 With(key, value) 时, fn(value) 的结果添加到 level: 后、kvs 前
func SetKvPrefix(key string, fn KvFormatter) {
    kvFmtMu.Lock()
    old := kvFmtMap.Load().(map[string]KvFormatter)
    newMap := make(map[string]KvFormatter, len(old)+1)
    for k, v := range old {
        newMap[k] = v
    }
    newMap[key] = fn
    kvFmtMap.Store(newMap)
    atomic.StoreInt32(&hasKvFmt, 1)
    kvFmtMu.Unlock()
}

// isSameHook 比较 Hook 是否为同一函数
func isSameHook(a, b Hook) bool {
    return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// RemoveHook 移除指定的钩子函数
func RemoveHook(hook Hook) {
    globalHooksMu.Lock()
    current := globalHooks.Load().([]Hook)
    next := make([]Hook, 0, len(current))
    for _, h := range current {
        if !isSameHook(h, hook) {
            next = append(next, h)
        }
    }
    globalHooks.Store(next)
    if len(next) > 0 {
        atomic.StoreInt32(&hasHooks, 1)
    } else {
        atomic.StoreInt32(&hasHooks, 0)
    }
    globalHooksMu.Unlock()
}

// logger 统一日志类型
type logger struct {
    showName    unsafe.Pointer     // *string, 原子读写避免 interface 装箱
    kvs         []any              // 结构化字段
    kvPrefix    unsafe.Pointer     // *string, 缓存的 kvs 前缀, With() 时计算
    fullName    string             // 完整名称
    parent      *logger            // 父 logger
    localWriter *writerTargetValue // nil = 使用全局 writerTarget
    local       int32              // 本地级别设置
    resolved    int32              // 缓存的有效级别
    levelGen    int32              // 上次计算 resolved 时的 generation
    derived     bool               // 是否为派生 logger (With/WithWriter), 可被 Unuse 回收
}

// --- 基础方法 ---

func (l *logger) FullName() string {
    return l.fullName
}

// getShowName 热路径: 直接 atomic.LoadPointer + 指针解引用, 无 interface 装箱
func (l *logger) getShowName() string {
    p := (*string)(atomic.LoadPointer(&l.showName))
    if p == nil {
        return ""
    }
    return *p
}

func (l *logger) setShowName(name string) {
    atomic.StorePointer(&l.showName, unsafe.Pointer(&name))
}

func (l *logger) setFullName(name string) {
    l.fullName = name
}

// Level O(1) 原子读取缓存的有效级别
func (l *logger) Level() Level {
    return Level(atomic.LoadInt32(&l.resolved))
}

// SetLevel 设置本地级别, 并更新缓存
// 传递 LevelUnset 清除本地设置, 继承父级或默认级别
func (l *logger) SetLevel(lv Level) {
    if lv != LevelUnset {
        atomic.StoreInt32(&l.local, int32(lv))
        atomic.StoreInt32(&l.resolved, int32(lv))
        atomic.StoreInt32(&l.levelGen, atomic.AddInt32(&levelGeneration, 1))
    } else {
        atomic.StoreInt32(&l.local, levelUnset)
        atomic.AddInt32(&levelGeneration, 1)
    }
}

// --- 日志方法 ---

// resolveLevel O(1) 快速路径, 拆分慢路径使快速路径可被编译器内联
func (l *logger) resolveLevel() Level {
    if atomic.LoadInt32(&l.levelGen) == atomic.LoadInt32(&levelGeneration) {
        return Level(atomic.LoadInt32(&l.resolved))
    }
    return l.resolveLevelSlow()
}

// resolveLevelSlow 慢路径: 遍历父链重新解析有效级别
func (l *logger) resolveLevelSlow() Level {
    if v := atomic.LoadInt32(&l.local); v >= 0 {
        gen := atomic.LoadInt32(&levelGeneration)
        atomic.StoreInt32(&l.resolved, v)
        atomic.StoreInt32(&l.levelGen, gen)
        return Level(v)
    }
    resolved := Level(atomic.LoadInt32(&defaultLevel))
    for p := l.parent; p != nil; p = p.parent {
        if v := atomic.LoadInt32(&p.local); v >= 0 {
            resolved = Level(v)
            break
        }
    }
    gen := atomic.LoadInt32(&levelGeneration)
    atomic.StoreInt32(&l.resolved, int32(resolved))
    atomic.StoreInt32(&l.levelGen, gen)
    return resolved
}

func (l *logger) D(msg string, args ...any) *Log {
    if l.resolveLevel() <= DebugLevel {
        l.write(DebugLevel, msg, args)
    }
    return l
}

func (l *logger) DF(msg string, f func() []any) *Log {
    if l.resolveLevel() <= DebugLevel {
        l.write(DebugLevel, msg, f())
    }
    return l
}

func (l *logger) I(msg string, args ...any) *Log {
    if l.resolveLevel() <= InfoLevel {
        l.write(InfoLevel, msg, args)
    }
    return l
}

func (l *logger) IF(msg string, f func() []any) *Log {
    if l.resolveLevel() <= InfoLevel {
        l.write(InfoLevel, msg, f())
    }
    return l
}

func (l *logger) L(lv Level, msg string, args ...any) *Log {
    if l.resolveLevel() <= lv {
        l.write(lv, msg, args)
    }
    return l
}

func (l *logger) LF(lv Level, msg string, f func() []any) *Log {
    if l.resolveLevel() <= lv {
        l.write(lv, msg, f())
    }
    return l
}

// --- 派生方法 ---

func (l *logger) With(kvs ...any) *Log {
    if len(kvs) == 0 {
        return l
    }
    return l.deriveWith(kvs)
}

// Unuse 回收派生 logger 到对象池
// 仅回收 With()/WithWriter() 派生的 logger, 命名 logger 调用无效
func (l *logger) Unuse() {
    if l.derived {
        logPool.Put(l)
    }
}

// UpdatePrefix 重新计算 kvPrefix 缓存
// 当 kvs 中的值状态变化时(如用户改名), 调用此方法刷新前缀
func (l *logger) UpdatePrefix() *Log {
    if atomic.LoadInt32(&hasKvFmt) == 0 || len(l.kvs) == 0 {
        return l
    }
    fm := kvFmtMap.Load().(map[string]KvFormatter)
    var sb strings.Builder
    for i := 0; i+1 < len(l.kvs); i += 2 {
        if key, ok := l.kvs[i].(string); ok {
            if fn := fm[key]; fn != nil {
                sb.WriteString(fn(l.kvs[i+1]))
            }
        }
    }
    s := sb.String()
    atomic.StorePointer(&l.kvPrefix, unsafe.Pointer(&s))
    return l
}

// deriveWith 从当前 logger 派生新实例, 合并父级和新传入的结构化字段
// 被 SetKvPrefix 注册的 key 会从 kvs 中移除, 只在前缀中显示
func (l *logger) deriveWith(kvs []any) *Log {
    nl := l.derive()
    // 合并父级 kvs + 新 kvs, 同时提取匹配的前缀
    hasFmt := atomic.LoadInt32(&hasKvFmt) != 0
    var fm map[string]KvFormatter
    if hasFmt {
        fm = kvFmtMap.Load().(map[string]KvFormatter)
    }
    // 先计算需要保留的 kvs 数量和前缀大小
    totalKvs := len(l.kvs) + len(kvs)
    keepCount := totalKvs
    prefixSize := 0
    if hasFmt {
        // 检查新 kvs
        for i := 0; i+1 < len(kvs); i += 2 {
            if key, ok := kvs[i].(string); ok && fm[key] != nil {
                keepCount -= 2
                prefixSize += 20
            }
        }
        // 检查父级 kvs (继承的也可能被提取)
        for i := 0; i+1 < len(l.kvs); i += 2 {
            if key, ok := l.kvs[i].(string); ok && fm[key] != nil {
                keepCount -= 2
                prefixSize += 20
            }
        }
    }
    // 分配 kvs 数组
    if keepCount > 0 {
        nl.kvs = make([]any, 0, keepCount)
        // 复制父级 kvs (跳过被提取的)
        for i := 0; i+1 < len(l.kvs); i += 2 {
            if key, ok := l.kvs[i].(string); ok && fm[key] != nil {
                continue // 跳过,已提取到前缀
            }
            nl.kvs = append(nl.kvs, l.kvs[i], l.kvs[i+1])
        }
        // 复制新 kvs (跳过被提取的)
        for i := 0; i+1 < len(kvs); i += 2 {
            if key, ok := kvs[i].(string); ok && fm[key] != nil {
                continue // 跳过,已提取到前缀
            }
            nl.kvs = append(nl.kvs, kvs[i], kvs[i+1])
        }
    }
    // 构建前缀
    if prefixSize > 0 {
        var sb strings.Builder
        sb.Grow(prefixSize)
        // 先提取父级匹配的前缀 (保持顺序)
        for i := 0; i+1 < len(l.kvs); i += 2 {
            if key, ok := l.kvs[i].(string); ok {
                if fn := fm[key]; fn != nil {
                    sb.WriteString(fn(l.kvs[i+1]))
                }
            }
        }
        // 再提取新 kvs 匹配的前缀
        for i := 0; i+1 < len(kvs); i += 2 {
            if key, ok := kvs[i].(string); ok {
                if fn := fm[key]; fn != nil {
                    sb.WriteString(fn(kvs[i+1]))
                }
            }
        }
        s := sb.String()
        nl.kvPrefix = unsafe.Pointer(&s)
    }
    return nl
}

// --- 核心写入 ---

// write 单 buffer 构建完整日志行: time[name]level: kvs msg caller\n
func (l *logger) write(lv Level, msg string, args []any) {
    var caller uintptr
    cfg := atomic.LoadInt32(&callerConfig)
    if cfg&(1<<16) != 0 {
        var pcs [1]uintptr
        runtime.Callers(int(cfg&0xFFFF), pcs[:])
        caller = pcs[0]
    }

    // 选择 writer: 独立 writer 或全局, 使用指针避免 struct 拷贝
    wtv := l.localWriter
    if wtv == nil {
        wtv = writerTarget.Load().(*writerTargetValue)
    }
    buf := bfPool.Get()
    if !wtv.isAsync {
        defer bfPool.Put(buf)
    }

    // 构建前缀
    AppendNowTime(buf)
    buf.WriteByte('[')
    buf.WriteString(l.getShowName())
    buf.WriteByte(']')
    // 内联 lv.String(), Level 短名固定单字符
    if lv >= 0 && int(lv) < len(levelByte) {
        buf.WriteByte(levelByte[lv])
    } else if lv > FatalLevel {
        buf.WriteByte('F')
    } else {
        buf.WriteByte('T')
    }
    buf.WriteByte(':')
    buf.WriteByte(' ')

    // 缓存的 kvs 前缀, With() 时已计算
    if p := (*string)(atomic.LoadPointer(&l.kvPrefix)); p != nil {
        buf.WriteString(*p)
    }

    // 结构化字段
    for i := 0; i+1 < len(l.kvs); i += 2 {
        buf.WriteByte(' ')
        appendAny(buf, l.kvs[i])
        buf.WriteByte('=')
        appendAny(buf, l.kvs[i+1])
    }
    if len(l.kvs)%2 == 1 {
        buf.WriteByte(' ')
        appendAny(buf, l.kvs[len(l.kvs)-1])
    }
    // kv 与消息间分隔空格
    if len(l.kvs) > 0 {
        buf.WriteByte(' ')
    }

    // 消息格式化
    if len(args) != 0 {
        doLogFormat1(buf, msg, args)
    } else {
        buf.WriteString(msg)
    }

    // 调用者信息
    if caller != 0 {
        buf.WriteByte(' ')
        formatCaller(buf, caller, cfg&(1<<17) != 0)
    }

    buf.WriteByte('\n')

    // 钩子: hasHooks 快速跳过无钩子时的 atomic.Value 开销
    if atomic.LoadInt32(&hasHooks) != 0 {
        hooks := globalHooks.Load().([]Hook)
        for i := 0; i < len(hooks); i++ {
            hooks[i](lv, buf, l)
        }
    }

    // 输出
    if wtv.isAsync {
        wtv.asyncW.WriterAsync(buf.Bytes(), func() {
            bfPool.Put(buf)
        })
    } else {
        wtv.writer.Write(buf.Bytes())
    }
}

// --- 辅助 ---

var bfPool = func() *pool.BytePool {
    p := pool.NewSimpleBytesPool()
    p.SetInitCap(128)
    p.SetMaxCap(512)
    return p
}()

// formatCaller 将 pc 格式化为 "file:line" 或 "func file:line" 写入 buffer
func formatCaller(s *pool.Bytes, pc uintptr, showFunc bool) {
    if pc == 0 {
        return
    }
    // 检查缓存 (读锁保护)
    cacheKey := pc
    if showFunc {
        // 带 func 的用不同 key, 避免与无 func 缓存冲突
        cacheKey = pc | (1 << 63)
    }
    callerCacheMu.RLock()
    v, ok := callerCache[cacheKey]
    callerCacheMu.RUnlock()
    if ok {
        s.WriteString(v)
        return
    }
    // 计算
    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return
    }
    file, line := fn.FileLine(pc)
    for i := len(file) - 1; i >= 0; i-- {
        if file[i] == '/' || file[i] == '\\' {
            file = file[i+1:]
            break
        }
    }
    result := file + ":" + formatIntStr(line)
    if showFunc {
        // 只取函数名部分 (去掉包路径)
        name := fn.Name()
        for i := len(name) - 1; i >= 0; i-- {
            if name[i] == '.' {
                name = name[i+1:]
                break
            }
        }
        result = name + " " + result
    }
    // 缓存或淘汰
    callerCacheMu.Lock()
    if len(callerCache) >= callerCacheMaxSize {
        // 清空重建: 热点 caller 会因后续访问迅速重新填充
        // 同时自适应扩容上限 1.5 倍, 最大不超过原始上限的 callerCacheMaxMul 倍
        callerCache = make(map[uintptr]string, callerCacheMaxSize/2)
        newSize := callerCacheMaxSize + callerCacheMaxSize/2
        if max := callerCacheOrigMaxSize * callerCacheMaxMul; newSize > max {
            newSize = max
        }
        callerCacheMaxSize = newSize
    }
    callerCache[cacheKey] = result
    callerCacheMu.Unlock()
    s.WriteString(result)
}

var bigParenthesis = helper.StringToBytes("{}")

// doLogFormat1 {}占位符风格, 支持 {} 自动递增和 {0} {1} 显式位置
func doLogFormat1(s *pool.Bytes, format string, args []any) {
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
                s.WriteString(format[lastWrite:i])
            }
            if argIdx < argLen {
                appendAny(s, args[argIdx])
            } else {
                s.Write(bigParenthesis)
            }
            argIdx++
            lastWrite = i + 2
            i += 2
            continue
        }
        // 慢路径: {0} {1} 等显式位置
        j := i + 1
        for j < len(format) && format[j] != '}' {
            j++
        }
        if j >= len(format) {
            break
        }
        idx := parseDigitIndex(format[i+1 : j])
        if idx >= 0 {
            if i > lastWrite {
                s.WriteString(format[lastWrite:i])
            }
            if idx < argLen {
                appendAny(s, args[idx])
            } else {
                s.Write(helper.StringToBytes(format[i : j+1]))
            }
            lastWrite = j + 1
            i = j + 1
        } else {
            i++
        }
    }
    if lastWrite < len(format) {
        s.WriteString(format[lastWrite:])
    }
    // 追加剩余未消费的参数 (仅自动递增的)
    if argIdx < argLen {
        for i := argIdx; i < argLen; i++ {
            s.WriteByte(' ')
            appendAny(s, args[i])
        }
    }
}

// parseDigitIndex 解析字符串为非负整数, 失败返回 -1
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

// formatIntStr 将正整数转换为字符串, 用于 caller 缓存构建
// 避免 strconv.FormatInt 的分配开销
func formatIntStr(n int) string {
    if n < 10 {
        // 单数字: 使用常量字符表
        return digits[n : n+1]
    }
    // 手动构建, 栈上数组不逃逸
    var buf [20]byte
    pos := len(buf)
    for n >= 10 {
        pos--
        buf[pos] = '0' + byte(n%10)
        n /= 10
    }
    pos--
    buf[pos] = '0' + byte(n)
    return string(buf[pos:])
}

// digits 0-9 的字符串表示, 用于快速单数字转换
const digits = "0123456789"
