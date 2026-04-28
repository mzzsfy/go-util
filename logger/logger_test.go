package logger

import (
    "bytes"
    "fmt"
    "io"
    "regexp"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "testing"
    "time"
    "unicode/utf8"

    "github.com/mzzsfy/go-util/pool"
)

func Test_Logger_1(t *testing.T) {
    info := atomic.LoadInt32(&printYearInfo)
    defer func() {
        atomic.StoreInt32(&printYearInfo, info)
    }()
    Logger("test.test").I("test")
    atomic.StoreInt32(&printYearInfo, 1)
    Logger("test.test").I("test")
    atomic.StoreInt32(&printYearInfo, 2)
    Logger("test.test").I("test")
    atomic.StoreInt32(&printYearInfo, 3)
    Logger("test.test").I("test")
    atomic.StoreInt32(&printYearInfo, 1)
    log1 := Logger("test.test")
    log1.I("test")
    length := int(atomic.LoadInt32(&showNameMaxLength))
    defer SetLogNameMaxLength(length)
    SetLogNameMaxLength(6)
    log1.I("short test")
}

func Test_Logger_2(t *testing.T) {
    Logger("test.test.test.test.test.test.test.test.test.test.test").I("test")
    Logger("test").I("test")
    Logger("test.test").I("test")
    Logger("test.test.test").I("test")
    Logger("test.test.test.test").I("test")
    Logger("test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test.test.test").I("test")
    Logger("test.test.test.test.test.test.test.test.test.test.test").I("test")
    Logger("你好.你好").I("test")
    Logger("你好.你好.你好.你好").I("test")
    Logger("你好.你好.你好.你好.你好.你好.你好").I("test")
    Logger("你好.你好.你好.你好.你好.你好.你好.你好.你好").I("test")
}

func TestInitName_MultiByteRunes(t *testing.T) {
    oldMax := int(atomic.LoadInt32(&showNameMaxLength))
    defer SetLogNameMaxLength(oldMax)
    SetLogNameMaxLength(18)

    tests := []struct {
        name     string
        fullName string
    }{
        {"chinese short", "你好.世界"},
        {"chinese long", "你好.你好.你好.你好.你好.你好.你好.你好.你好"},
        {"mixed", "test.测试.module.模块.包名.名称"},
        {"ascii long", "a.b.c.d.e.f.g.h.i.j.k"},
        {"single chinese segment", "中文模块名测试"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            names := strings.Split(tt.fullName, ".")
            result := initName(len(names)-1, names, tt.fullName)
            if !utf8.ValidString(result) {
                t.Errorf("invalid UTF-8 in showName for %q: %q", tt.fullName, result)
            }
            runeCount := utf8.RuneCountInString(result)
            maxLen := int(atomic.LoadInt32(&showNameMaxLength))
            if runeCount > maxLen {
                t.Errorf("showName too long for %q: got %d runes, max %d: %q", tt.fullName, runeCount, maxLen, result)
            }
        })
    }
}

func TestInitName_SmallMaxLength_NoPanic(t *testing.T) {
    t.Parallel()
    oldMax := int(atomic.LoadInt32(&showNameMaxLength))
    oldCompressed := atomic.LoadInt32(&compressedLogName)
    defer func() {
        atomic.StoreInt32(&compressedLogName, oldCompressed)
        SetLogNameMaxLength(oldMax)
    }()

    atomic.StoreInt32(&compressedLogName, 1)
    names := strings.Split("alpha.beta.gamma", ".")
    for _, maxLen := range []int{0, 1, 2, 3, 4, 5} {
        t.Run(fmt.Sprintf("max_%d", maxLen), func(t *testing.T) {
            SetLogNameMaxLength(maxLen)
            result := initName(len(names)-1, names, "alpha.beta.gamma")
            if !utf8.ValidString(result) {
                t.Fatalf("invalid UTF-8 for maxLen=%d: %q", maxLen, result)
            }
        })
    }
}

func TestAllLogger_VisitsAllNames(t *testing.T) {
    names := []string{"allLogger.test.a", "allLogger.test.b", "allLogger.test.c"}
    for _, n := range names {
        Logger(n)
    }
    var collected []string
    AllLogger()(func(name string) {
        if strings.HasPrefix(name, "allLogger.test.") {
            collected = append(collected, name)
        }
    })
    if len(collected) < len(names) {
        t.Fatalf("AllLogger visited only %d, expected %d", len(collected), len(names))
    }
    set := make(map[string]bool, len(collected))
    for _, n := range collected {
        set[n] = true
    }
    for _, n := range names {
        if !set[n] {
            t.Errorf("AllLogger did not visit %q", n)
        }
    }
}

func TestDoLogFormat1(t *testing.T) {
    t.Parallel()
    t.Run("一一对应", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "a={} b={}", []any{1, "x"})
        if got := buf.String(); got != "a=1 b=x" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("参数不足", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "a={} b={}", []any{1})
        if got := buf.String(); got != "a=1 b={}" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("参数多余", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "a={}", []any{1, "x", 2})
        if got := buf.String(); got != "a=1 x 2" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("error详细格式", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "a={}", []any{1, fmt.Errorf("boom")})
        if got := buf.String(); got != "a=1 boom" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("无占位符追加", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "plain", []any{1, "x"})
        if got := buf.String(); got != "plain 1 x" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("显式位置{0}{1}", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "b={1} a={0}", []any{1, "x"})
        if got := buf.String(); got != "b=x a=1 1 x" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("显式位置超出范围", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "a={0} b={2}", []any{1, "x"})
        if got := buf.String(); got != "a=1 b={2} 1 x" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("混合自动和显式", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "a={} b={0} c={}", []any{1, "x"})
        if got := buf.String(); got != "a=1 b=1 c=x" {
            t.Fatalf("got %q", got)
        }
    })
    t.Run("显式位置不消耗自动索引", func(t *testing.T) {
        buf := bfPool.Get()
        defer bfPool.Put(buf)
        doLogFormat1(buf, "{0} {} {1}", []any{"a", "b", "c"})
        if got := buf.String(); got != "a a b b c" {
            t.Fatalf("got %q", got)
        }
    })
}

func TestLogger_LazyMethods(t *testing.T) {
    t.Parallel()
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(io.Discard)
    defer SetDefaultWriterTarget(oldWriter)

    l := Logger("lazy.methods.test")
    l.SetLevel(WarnLevel)
    defer l.SetLevel(LevelUnset)

    l.DF("test", func() []any { t.Fatal("DF should not evaluate"); return nil })
    l.IF("test", func() []any { t.Fatal("IF should not evaluate"); return nil })

    var wfCalled, efCalled bool
    l.LF(WarnLevel, "test", func() []any { wfCalled = true; return []any{"w"} })
    l.LF(ErrorLevel, "test", func() []any { efCalled = true; return []any{"e"} })
    if !wfCalled || !efCalled {
        t.Fatal("LF(Warn/Error) should be evaluated")
    }
}

func TestLogger_LF_DirectLevel(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    l := Logger("lf.direct.level")
    l.L(InfoLevel, "test {}", "arg")
    l.LF(WarnLevel, "lazy {}", func() []any { return []any{"val"} })

    output := buf.String()
    if !strings.Contains(output, "test arg") || !strings.Contains(output, "lazy val") {
        t.Fatalf("output: %q", output)
    }
}

func TestLevel_MarshalUnmarshal(t *testing.T) {
    t.Parallel()
    t.Run("String", func(t *testing.T) {
        for _, tt := range []struct {
            level Level
            short string
            full  string
        }{
            {TraceLevel, "T", "Trace"}, {DebugLevel, "D", "Debug"},
            {InfoLevel, "I", "Info"}, {WarnLevel, "W", "Warn"},
            {ErrorLevel, "E", "Error"}, {FatalLevel, "F", "Fatal"},
        } {
            if got := tt.level.String(); got != tt.short {
                t.Errorf("got %q, want %q", got, tt.short)
            }
            if got := tt.level.FullName(); got != tt.full {
                t.Errorf("got %q, want %q", got, tt.full)
            }
        }
    })
    t.Run("FromString边界", func(t *testing.T) {
        if FromString("") != InfoLevel || FromString("xyz") != InfoLevel {
            t.Error("boundary cases failed")
        }
    })
}

func BenchmarkLogger_Logging(b *testing.B) {
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(io.Discard)
    defer SetDefaultWriterTarget(oldWriter)
    log := Logger("bench.logging")

    b.Run("NoArgs", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            log.I("simple message")
        }
    })
    b.Run("BracePlaceholder_1Arg", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            log.I("value={}", i)
        }
    })
    b.Run("LevelCheck_FiltersOut", func(b *testing.B) {
        log.SetLevel(WarnLevel)
        defer log.SetLevel(LevelUnset)
        for i := 0; i < b.N; i++ {
            log.D("filtered {}", i)
        }
    })
    b.Run("Parallel_NoArgs", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for i := 0; pb.Next(); i++ {
                log.I("simple message")
            }
        })
    })
    b.Run("Parallel_BracePlaceholder", func(b *testing.B) {
        b.RunParallel(func(pb *testing.PB) {
            for i := 0; pb.Next(); i++ {
                log.I("value={}", i)
            }
        })
    })
}

func TestLogger_NoPlaceholderAppendArgs_NoExtraFormatNoise(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    Logger("regression.append.args").I("value:", 42)
    output := buf.String()
    if strings.Contains(output, "%!(EXTRA") {
        t.Fatalf("should not contain fmt extra, got %q", output)
    }
    if !strings.Contains(output, "value: 42") {
        t.Fatalf("should contain appended args, got %q", output)
    }
}

func TestPrintYearInfo_OutputAssertion(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    oldYearInfo := atomic.LoadInt32(&printYearInfo)
    SetDefaultWriterTarget(&buf)
    defer func() {
        atomic.StoreInt32(&printYearInfo, oldYearInfo)
        SetDefaultWriterTarget(oldWriter)
    }()

    for _, tt := range []struct {
        name    string
        value   int32
        pattern string
    }{
        {"four digit", 0, `^\d{4}-\d{2}-\d{2} `},
        {"two digit", 1, `^\d{2}-\d{2}-\d{2} `},
        {"no year", 2, `^\d{2}-\d{2} `},
    } {
        t.Run(tt.name, func(t *testing.T) {
            buf.Reset()
            atomic.StoreInt32(&printYearInfo, tt.value)
            cachedTime.Store(&timeCache{})
            Logger("year.info").I("test")
            matched, _ := regexp.MatchString(tt.pattern, buf.String())
            if !matched {
                t.Fatalf("output %q doesn't match %q", buf.String(), tt.pattern)
            }
        })
    }
}

func TestCompressedLogName_OutputAssertion(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    oldCompressed := atomic.LoadInt32(&compressedLogName)
    oldMax := int(atomic.LoadInt32(&showNameMaxLength))
    SetDefaultWriterTarget(&buf)
    defer func() {
        atomic.StoreInt32(&compressedLogName, oldCompressed)
        SetLogNameMaxLength(oldMax)
        SetDefaultWriterTarget(oldWriter)
    }()

    const name = "alpha.beta.gamma.delta.epsilon.zeta"
    SetLogNameMaxLength(12)

    atomic.StoreInt32(&compressedLogName, 0)
    Logger(name).I("plain")
    if !strings.Contains(buf.String(), "..") {
        t.Fatalf("should be truncated, got %q", buf.String())
    }

    buf.Reset()
    atomic.StoreInt32(&compressedLogName, 1)
    SetLogNameMaxLength(12)
    Logger(name).I("compressed")
    if !strings.Contains(buf.String(), "a...d.e.zeta") {
        t.Fatalf("should be compressed, got %q", buf.String())
    }
}

func TestInitName_EmptySegment_NoPanic(t *testing.T) {
    oldCompressed := atomic.LoadInt32(&compressedLogName)
    oldMax := int(atomic.LoadInt32(&showNameMaxLength))
    atomic.StoreInt32(&compressedLogName, 1)
    SetLogNameMaxLength(6)
    defer func() {
        atomic.StoreInt32(&compressedLogName, oldCompressed)
        SetLogNameMaxLength(oldMax)
    }()

    for _, tt := range []struct {
        step     int
        names    []string
        fullname string
    }{
        {2, strings.Split("alpha..beta", "."), "alpha..beta"},
        {3, strings.Split("alpha...beta", "."), "alpha...beta"},
    } {
        t.Run(tt.fullname, func(t *testing.T) {
            defer func() {
                if r := recover(); r != nil {
                    t.Fatalf("panicked: %v", r)
                }
            }()
            result := initName(tt.step, tt.names, tt.fullname)
            if !utf8.ValidString(result) {
                t.Fatalf("invalid UTF-8: %q", result)
            }
        })
    }
}

func TestLogger_SetLogNameMaxLength_ConcurrentSafety(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    l := Logger("concurrent.test.name")
    done := make(chan struct{})
    go func() {
        defer close(done)
        for i := 0; i < 100; i++ {
            SetLogNameMaxLength(4 + i%20)
        }
    }()
    for i := 0; i < 100; i++ {
        l.I("msg" + strconv.Itoa(i))
    }
    <-done
    if buf.Len() == 0 {
        t.Fatal("expected output")
    }
}

func TestDefaultLogLevel_BasicBehavior(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    oldLevel := DefaultLogLevel()
    SetDefaultWriterTarget(&buf)
    defer func() {
        SetDefaultLogLevel(oldLevel)
        SetDefaultWriterTarget(oldWriter)
    }()

    SetDefaultLogLevel(WarnLevel)
    log := Logger("default.level.test")
    log.I("filtered")
    if buf.String() != "" {
        t.Fatalf("info should be filtered, got %q", buf.String())
    }
    log.L(WarnLevel, "visible")
    if !strings.Contains(buf.String(), "visible") {
        t.Fatalf("warn should be emitted, got %q", buf.String())
    }
}

func TestDefaultWriterTarget_NilFallback(t *testing.T) {
    oldWriter := DefaultWriterTarget()
    defer SetDefaultWriterTarget(oldWriter)
    SetDefaultWriterTarget(nil)
    if DefaultWriterTarget() == nil {
        t.Fatal("should fallback to stdout")
    }
    Logger("writer.nil").I("test")
}

func TestFormatTimeCache_YearLessThan100(t *testing.T) {
    oldYearInfo := atomic.LoadInt32(&printYearInfo)
    atomic.StoreInt32(&printYearInfo, 1)
    defer atomic.StoreInt32(&printYearInfo, oldYearInfo)

    cache := formatTimeCache(time.Date(5, 1, 15, 10, 30, 45, 0, time.UTC), 0)
    if !bytes.HasPrefix(cache.prefix, []byte("05-01-15 10:30:45.")) {
        t.Fatalf("year=5 prefix wrong: %q", cache.prefix)
    }
}

func TestLogger_DerivedLevelInheritance(t *testing.T) {
    log := Logger("level.derive.parent")
    log.SetLevel(ErrorLevel)

    if log.Level() != ErrorLevel {
        t.Fatalf("base level wrong: %v", log.Level())
    }
    if d := log.With("k", "v"); d.Level() != ErrorLevel {
        t.Fatalf("With derived wrong: %v", d.Level())
    }
    if d := log.With("k", "v").With("k2", "v2"); d.Level() != ErrorLevel {
        t.Fatalf("multi derived wrong: %v", d.Level())
    }
}

func TestLogger_WithKvsOutput(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    Logger("with.kvs").With("userId", 42, "name", "test").I("action")
    output := buf.String()
    if !strings.Contains(output, "userId=42") || !strings.Contains(output, "name=test") {
        t.Fatalf("output: %q", output)
    }
}

func TestHooks_ConcurrentAddAndLog(t *testing.T) {
    CleanHooks()
    defer CleanHooks()

    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(io.Discard)
    defer SetDefaultWriterTarget(oldWriter)

    var count int64
    const writers, loggers, perLogger = 5, 10, 50
    var wg sync.WaitGroup
    wg.Add(writers + loggers)

    for i := 0; i < writers; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < 10; j++ {
                AddHook(func(lv Level, buf *pool.Bytes, log *logger) {
                    atomic.AddInt64(&count, 1)
                })
            }
        }()
    }
    for i := 0; i < loggers; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < perLogger; j++ {
                Logger("hook.concurrent").I("msg", j)
            }
        }()
    }
    wg.Wait()
    t.Logf("hook count=%d", atomic.LoadInt64(&count))
}

func TestHook_Output(t *testing.T) {
    CleanHooks()
    defer CleanHooks()

    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    AddHook(func(lv Level, buf *pool.Bytes, log *logger) {
        buf.WriteString(" [svc]")
    })

    Logger("hook.test").I("hello")
    if !strings.Contains(buf.String(), "[svc]") {
        t.Fatalf("output: %q", buf.String())
    }
}

func TestSetLevelRecursive(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    // 创建父子层级
    Logger("rec.parent")
    Logger("rec.parent.child")
    Logger("rec.parent.child.grandchild")
    Logger("rec.other")

    // 递归设置为 ErrorLevel
    SetLevelRecursive("rec.parent", ErrorLevel)

    // rec.parent 及其子 logger 的 I 应该被过滤
    Logger("rec.parent").I("filtered")
    Logger("rec.parent.child").I("filtered")
    Logger("rec.parent.child.grandchild").I("filtered")

    // rec.other 不受影响
    Logger("rec.other").I("visible")

    if strings.Contains(buf.String(), "filtered") {
        t.Fatalf("recursive set should filter children: %q", buf.String())
    }
    if !strings.Contains(buf.String(), "visible") {
        t.Fatalf("unrelated logger should not be affected: %q", buf.String())
    }
}

func TestKvsCopy_DefenseMutation(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    kvs := []any{"key", "val1"}
    log := Logger("kvs.copy").With(kvs...)
    kvs[1] = "val2" // 修改原始切片

    log.I("test")
    if !strings.Contains(buf.String(), "key=val1") {
        t.Fatalf("kvs should not be affected by mutation: %q", buf.String())
    }
}

func TestLevelGeneration_ParentChangePropagates(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    // 创建父子层级
    Logger("gen.parent")
    child := Logger("gen.parent.child")

    // 子 logger 应该继承默认 Info 级别
    child.D("should be filtered")
    if buf.String() != "" {
        t.Fatalf("debug should be filtered: %q", buf.String())
    }

    // 修改父 logger 为 Debug 级别
    Logger("gen.parent").SetLevel(DebugLevel)

    // 子 logger 应该自动感知父级变化
    child.D("should be visible now")
    if !strings.Contains(buf.String(), "should be visible now") {
        t.Fatalf("child should inherit parent's new level: %q", buf.String())
    }

    // 恢复
    Logger("gen.parent").SetLevel(LevelUnset)
}

func TestLevelGeneration_DefaultLevelChange(t *testing.T) {
    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    oldLevel := DefaultLogLevel()
    SetDefaultWriterTarget(&buf)
    defer func() {
        SetDefaultLogLevel(oldLevel)
        SetDefaultWriterTarget(oldWriter)
    }()

    SetDefaultLogLevel(DebugLevel)
    log := Logger("gen.default")
    log.D("visible after default change")
    if !strings.Contains(buf.String(), "visible") {
        t.Fatalf("should see debug after default level change: %q", buf.String())
    }
    SetDefaultLogLevel(oldLevel)
}

func TestWithWriter_PerLoggerOutput(t *testing.T) {
    var globalBuf, localBuf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&globalBuf)
    defer SetDefaultWriterTarget(oldWriter)

    // 全局 logger 写到 globalBuf
    globalLog := Logger("writer.global")
    // 独立 writer 的 logger 写到 localBuf
    localLog := Logger("writer.local", WithWriter(&localBuf))

    globalLog.I("global msg")
    localLog.I("local msg")

    if !strings.Contains(globalBuf.String(), "global msg") {
        t.Fatalf("global writer: %q", globalBuf.String())
    }
    if strings.Contains(globalBuf.String(), "local msg") {
        t.Fatalf("local should NOT write to global: %q", globalBuf.String())
    }
    if !strings.Contains(localBuf.String(), "local msg") {
        t.Fatalf("local writer: %q", localBuf.String())
    }
}

func TestWithWriter_DerivedInheritsWriter(t *testing.T) {
    var localBuf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(io.Discard)
    defer SetDefaultWriterTarget(oldWriter)

    log := Logger("writer.derive", WithWriter(&localBuf))
    log.With("k", "v").I("derived msg")

    if !strings.Contains(localBuf.String(), "derived msg") {
        t.Fatalf("derived should inherit writer: %q", localBuf.String())
    }
}

func TestRemoveHook(t *testing.T) {
    CleanHooks()
    defer CleanHooks()

    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    marker := func(lv Level, b *pool.Bytes, log *logger) {
        b.WriteString(" [mark]")
    }
    AddHook(marker)
    Logger("remove.hook").I("with hook")
    if !strings.Contains(buf.String(), "[mark]") {
        t.Fatalf("hook should be active: %q", buf.String())
    }

    RemoveHook(marker)
    buf.Reset()
    Logger("remove.hook").I("without hook")
    if strings.Contains(buf.String(), "[mark]") {
        t.Fatalf("hook should be removed: %q", buf.String())
    }
}

func TestCallerConfig_SetCallerInfo(t *testing.T) {
    // 保存并恢复 callerConfig
    oldCfg := atomic.LoadInt32(&callerConfig)
    defer atomic.StoreInt32(&callerConfig, oldCfg)

    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    // 启用 caller
    SetCallerInfo(true)
    buf.Reset()
    Logger("caller.on").I("with caller")
    // 应包含 .go: 文件行号信息
    if !strings.Contains(buf.String(), ".go:") {
        t.Fatalf("caller should be present: %q", buf.String())
    }

    // 禁用 caller
    SetCallerInfo(false)
    buf.Reset()
    Logger("caller.off").I("no caller")
    if strings.Contains(buf.String(), ".go:") {
        t.Fatalf("caller should be absent: %q", buf.String())
    }
}

func TestCallerConfig_SetCallerFunc(t *testing.T) {
    oldCfg := atomic.LoadInt32(&callerConfig)
    defer atomic.StoreInt32(&callerConfig, oldCfg)

    var buf bytes.Buffer
    oldWriter := DefaultWriterTarget()
    SetDefaultWriterTarget(&buf)
    defer SetDefaultWriterTarget(oldWriter)

    // 启用 caller + 函数名
    SetCallerInfo(true)
    SetCallerFunc(true)
    buf.Reset()
    Logger("caller.func").I("with func")
    // 函数名格式: "FuncName file.go:line"
    output := buf.String()
    if !strings.Contains(output, ".go:") {
        t.Fatalf("caller should be present: %q", output)
    }

    // 关闭函数名
    SetCallerFunc(false)
    SetCallerInfo(false)
}

func TestRemoveLogger(t *testing.T) {
    // 创建一个唯一的logger层级
    prefix := "remove.test"
    l := Logger(prefix + ".a")
    if l == nil {
        t.Fatal("Logger should not be nil")
    }
    // 验证存在
    globalLock.RLock()
    _, ok1 := globalLog.Get(prefix + ".a")
    _, ok2 := globalLog.Get(prefix)
    globalLock.RUnlock()
    if !ok1 || !ok2 {
        t.Fatal("logger should exist in globalLog")
    }

    // 移除
    RemoveLogger(prefix)

    // 验证已删除
    globalLock.RLock()
    _, ok1 = globalLog.Get(prefix + ".a")
    _, ok2 = globalLog.Get(prefix)
    globalLock.RUnlock()
    if ok1 || ok2 {
        t.Fatal("logger should be removed from globalLog")
    }
}

func TestMaxLoggerCount(t *testing.T) {
    old := atomic.LoadInt32(&maxLoggerCount)
    defer func() {
        atomic.StoreInt32(&maxLoggerCount, old)
        RemoveLogger("maxlimit")
    }()

    RemoveLogger("maxlimit")

    // 记录当前已有logger数量
    globalLock.RLock()
    baseCount := globalLog.Count()
    globalLock.RUnlock()

    // 设置限制为 baseCount + 1: 只允许再创建1个
    // 创建 a.b 需要2个条目 (a, a.b), 第二个应超限
    SetMaxLoggerCount(baseCount + 1)

    // Logger("maxlimit") 占用唯一剩余槽位
    Logger("maxlimit")
    globalLock.RLock()
    _, okRoot := globalLog.Get("maxlimit")
    globalLock.RUnlock()
    if !okRoot {
        t.Fatal("maxlimit should be cached")
    }

    // maxlimit.child 超出限制, 应返回临时logger
    l := Logger("maxlimit.child")
    if l == nil {
        t.Fatal("should return temporary logger")
    }
    if !l.derived {
        t.Fatal("overflow logger should be derived (recyclable)")
    }
    globalLock.RLock()
    _, okChild := globalLog.Get("maxlimit.child")
    globalLock.RUnlock()
    if okChild {
        t.Fatal("overflow logger should not be cached")
    }
    l.Unuse()
}
