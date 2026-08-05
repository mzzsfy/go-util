package logger

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 辅助: 创建写到 buf 的 Logger
func newTestLogger(buf *bytes.Buffer, lv Level) *Logger {
	return New("test", WithLevel(lv), WithWriter(buf))
}

// --- BDD: 基本日志输出 ---

// Given Info 级别 Logger
// When 调用 Info().Msg("hello")
// Then 输出包含 "hello" 和级别标识 "I"
func Test_Event_BasicOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l.Info().Msg("hello")

	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Fatalf("missing message: %q", out)
	}
	if !strings.Contains(out, "]I:") {
		t.Fatalf("missing level tag: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("missing newline: %q", out)
	}
}

// --- BDD: 级别过滤 ---

// Given Warn 级别 Logger
// When 调用 Debug().Str("k","v").Msg("filtered")
// Then 无输出
func Test_Event_LevelFiltered(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, WarnLevel)
	l.Debug().Str("k", "v").Msg("filtered")

	if buf.Len() != 0 {
		t.Fatalf("expected no output, got %q", buf.String())
	}
}

// --- BDD: 类型安全链式字段 ---

// Given Info 级别 Logger
// When 调用 Info().Str("user","moke").Int("id",42).Bool("ok",true).Msg("login")
// Then 输出包含所有字段
func Test_Event_ChainedFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l.Info().Str("user", "moke").Int("id", 42).Bool("ok", true).Msg("login")

	out := buf.String()
	checks := []string{"user=moke", "id=42", "ok=true", "login"}
	for _, s := range checks {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

// --- BDD: 所有类型方法 ---

func Test_Event_AllFieldTypes(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, TraceLevel)

	l.Trace().
		Str("s", "val").
		Int("i", -7).
		Int64("i64", -100).
		Uint64("u", 999).
		Float64("f", 3.14).
		Bool("b", false).
		Time("t", time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)).
		Dur("d", 5*time.Second).
		Err(io.EOF).
		Any("any", 42).
		Msg("done")

	out := buf.String()
	checks := []string{
		`s=val`, `i=-7`, `i64=-100`, `u=999`, `f=3.14`,
		`b=false`, `err=EOF`, `any=42`, `done`,
	}
	for _, s := range checks {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

// --- BDD: 级别控制 ---

func Test_Logger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, ErrorLevel)

	l.Info().Msg("filtered")
	if buf.Len() != 0 {
		t.Fatalf("info should be filtered: %q", buf.String())
	}

	l.SetLevel(DebugLevel)
	l.Info().Msg("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Fatalf("info should be visible after SetLevel: %q", buf.String())
	}
}

func Test_Logger_Enabled(t *testing.T) {
	l := New("t", WithLevel(WarnLevel))
	if l.Enabled(DebugLevel) {
		t.Fatal("debug should not be enabled at warn level")
	}
	if !l.Enabled(ErrorLevel) {
		t.Fatal("error should be enabled at warn level")
	}
}

// --- BDD: Msgf 格式化消息 ---

func Test_Event_Msgf(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l.Info().Msgf("count=%d name=%s", 5, "test")
	if !strings.Contains(buf.String(), "count=5 name=test") {
		t.Fatalf("msgf output wrong: %q", buf.String())
	}
}

// --- BDD: Send 无消息 ---

func Test_Event_Send(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l.Info().Str("event", "startup").Send()
	out := buf.String()
	if !strings.Contains(out, "event=startup") {
		t.Fatalf("send output wrong: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("missing newline: %q", out)
	}
}

// --- BDD: 全局配置 ---

func Test_Global_DefaultLevel(t *testing.T) {
	old := DefaultLevel()
	defer SetDefaultLevel(old)

	SetDefaultLevel(DebugLevel)
	l := New("global-test")
	if l.Level() != DebugLevel {
		t.Fatalf("new logger should snapshot default level: got %v", l.Level())
	}
}

func Test_Global_DefaultWriter(t *testing.T) {
	old := DefaultWriter()
	defer SetDefaultWriter(old)

	var buf bytes.Buffer
	SetDefaultWriter(&buf)
	l := New("global-writer-test")
	l.Info().Msg("from global")
	if !strings.Contains(buf.String(), "from global") {
		t.Fatalf("global writer not used: %q", buf.String())
	}
}

func Test_Global_DefaultWriter_NilFallback(t *testing.T) {
	old := DefaultWriter()
	defer SetDefaultWriter(old)

	SetDefaultWriter(nil)
	if DefaultWriter() == nil {
		t.Fatal("nil should fallback to stdout")
	}
}

func Test_Global_DefaultLogger(t *testing.T) {
	l1 := Default()
	l2 := Default()
	if l1 != l2 {
		t.Fatal("Default should return same instance")
	}
}

// --- BDD: Level 工具函数 ---

func Test_Level_StringAndFromString(t *testing.T) {
	cases := []struct {
		lv   Level
		name string
	}{
		{TraceLevel, "Trace"}, {DebugLevel, "Debug"},
		{InfoLevel, "Info"}, {WarnLevel, "Warn"},
		{ErrorLevel, "Error"}, {FatalLevel, "Fatal"},
	}
	for _, c := range cases {
		if c.lv.String() != c.name {
			t.Errorf("%v.String() = %q, want %q", c.lv, c.lv.String(), c.name)
		}
		if FromString(c.name) != c.lv {
			t.Errorf("FromString(%q) = %v, want %v", c.name, FromString(c.name), c.lv)
		}
		// 大小写不敏感
		if FromString(strings.ToUpper(c.name)) != c.lv {
			t.Errorf("FromString(%q) failed case-insensitive", strings.ToUpper(c.name))
		}
	}
	if FromString("") != InfoLevel {
		t.Error("empty string should return InfoLevel")
	}
	if FromString("unknown") != InfoLevel {
		t.Error("unknown should return InfoLevel")
	}
}

func Test_Level_Tag(t *testing.T) {
	tags := map[Level]byte{
		TraceLevel: 'T', DebugLevel: 'D', InfoLevel: 'I',
		WarnLevel: 'W', ErrorLevel: 'E', FatalLevel: 'F',
	}
	for lv, want := range tags {
		if lv.tag() != want {
			t.Errorf("%v.tag() = %c, want %c", lv, lv.tag(), want)
		}
	}
}

// --- 并发安全 ---

func Test_Logger_ConcurrentSafety(t *testing.T) {
	l := New("concurrent", WithLevel(TraceLevel), WithWriter(io.Discard))
	done := make(chan struct{}, 4)
	for i := 0; i < 4; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				l.Info().Int("i", j).Msg("concurrent")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		<-done
	}
}

// --- 并行级别切换 ---

func Test_Logger_ConcurrentSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	done := make(chan struct{}, 2)
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			l.SetLevel(Level(i % 6))
		}
	}()
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			l.Info().Msg("msg")
		}
	}()
	<-done
	<-done
}

// --- 时间格式验证 ---

func Test_Format_TimeLayout(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l.Info().Msg("time-check")
	out := buf.String()
	// 格式: MM-DD HH:MM:SS.mmm I name msg
	if len(out) < 19 { // "MM-DD HH:MM:SS.mmm" = 18 字符 + 空格
		t.Fatalf("output too short, time prefix missing: %q", out)
	}
}

// --- benchmark ---

func Benchmark_Logger_NoArgs(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("simple message")
	}
}

func Benchmark_Logger_WithFields(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info().Int("id", i).Str("method", "GET").Int("status", 200).Msg("request")
	}
}

func Benchmark_Logger_FilteredOut(b *testing.B) {
	l := New("bench", WithLevel(WarnLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug().Int("id", i).Str("k", "v").Msg("filtered")
	}
}

func Benchmark_Logger_Parallel(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			l.Info().Int("id", i).Msg("parallel")
		}
	})
}

func Benchmark_Logger_ParallelFiltered(b *testing.B) {
	l := New("bench", WithLevel(WarnLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			l.Debug().Int("id", i).Msg("filtered")
		}
	})
}

// 编译期保证 atomic 使用符合规范
var _ = atomic.LoadInt32

// --- BDD: With 预设字段 ---

// Given 带 With 预设字段的 Logger
// When 调用 Info().Msg("request")
// Then 输出包含预设字段和消息
func Test_With_PresetFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l2 := l.With().Str("svc", "api").Int("ver", 2).Logger()

	l2.Info().Msg("request")
	out := buf.String()
	for _, s := range []string{"svc=api", "ver=2", "request"} {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

// Given 带 With 预设字段的 Logger
// When 追加事件级字段
// Then 预设字段和事件字段都出现
func Test_With_EventFieldsAppend(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, TraceLevel)
	l2 := l.With().Str("preset", "yes").Logger()

	l2.Info().Str("event", "login").Int("uid", 1).Msg("done")
	out := buf.String()
	// 预设字段在事件字段之前
	presetIdx := strings.Index(out, "preset=yes")
	eventIdx := strings.Index(out, "event=login")
	if presetIdx < 0 || eventIdx < 0 {
		t.Fatalf("missing fields: %q", out)
	}
	if presetIdx > eventIdx {
		t.Fatalf("preset should come before event: %q", out)
	}
}

// Given 多层 With
// When 链式派生
// Then 所有层级的字段都出现
func Test_With_MultipleLevels(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l2 := l.With().Str("a", "1").Logger()
	l3 := l2.With().Str("b", "2").Logger()

	l3.Info().Msg("multi")
	out := buf.String()
	for _, s := range []string{"a=1", "b=2", "multi"} {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

// With 不修改原 Logger
func Test_With_Immutable(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l2 := l.With().Str("k", "v").Logger()

	// 原 Logger 不带预设字段
	l.Info().Msg("original")
	out1 := buf.String()
	if strings.Contains(out1, "k=v") {
		t.Fatalf("original logger should not have preset: %q", out1)
	}

	// 派生 Logger 带预设字段
	buf.Reset()
	l2.Info().Msg("derived")
	out2 := buf.String()
	if !strings.Contains(out2, "k=v") {
		t.Fatalf("derived logger should have preset: %q", out2)
	}
}

// With 派生 Logger 级别独立
func Test_With_LevelIndependent(t *testing.T) {
	l := newTestLogger(&bytes.Buffer{}, InfoLevel)
	l2 := l.With().Str("k", "v").Logger()
	l2.SetLevel(DebugLevel)

	if l.Level() != InfoLevel {
		t.Fatalf("original level changed: %v", l.Level())
	}
	if l2.Level() != DebugLevel {
		t.Fatalf("derived level wrong: %v", l2.Level())
	}
}

// --- With benchmark ---

func Benchmark_Logger_With(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.Run("CreateDerived", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.With().Str("svc", "api").Int("ver", 2).Logger()
		}
	})
	b.Run("LogWithPreset", func(b *testing.B) {
		l2 := l.With().Str("svc", "api").Int("ver", 2).Logger()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l2.Info().Int("id", i).Msg("request")
		}
	})
}

// --- BDD: 命名 Logger 管理 ---

// Given Get("app.user")
// When 再次 Get("app.user")
// Then 返回同一实例
func Test_Named_GetSameInstance(t *testing.T) {
	RemoveLogger("ns.same")
	defer RemoveLogger("ns.same")

	l1 := Get("ns.same")
	l2 := Get("ns.same")
	if l1 != l2 {
		t.Fatal("same name should return same instance")
	}
}

// Given Get("ns.inherit.parent") 设为 ErrorLevel
// When Get("ns.inherit.parent.child")
// Then child 继承 parent 的 ErrorLevel
func Test_Named_LevelInheritance(t *testing.T) {
	RemoveLogger("ns.inherit")
	defer RemoveLogger("ns.inherit")

	parent := Get("ns.inherit.parent")
	parent.SetLevel(ErrorLevel)

	child := Get("ns.inherit.parent.child")
	if child.Level() != ErrorLevel {
		t.Fatalf("child should inherit parent level: got %v", child.Level())
	}
}

// Given 子 Logger 已创建
// When 修改父 Logger 级别
// Then 子 Logger 自动感知父级变化
func Test_Named_LevelChangePropagates(t *testing.T) {
	RemoveLogger("ns.prop")
	defer RemoveLogger("ns.prop")

	parent := Get("ns.prop.parent")
	child := Get("ns.prop.parent.child")

	// 初始都是 Info (默认)
	if child.Level() != InfoLevel {
		t.Fatalf("initial child level: %v", child.Level())
	}

	// 父级改为 Debug
	parent.SetLevel(DebugLevel)

	// 子级应感知
	if child.Level() != DebugLevel {
		t.Fatalf("child should inherit new level: got %v", child.Level())
	}
}

// Given 全局默认级别变更
// When 继承默认的命名 Logger 解析级别
// Then 感知新默认级别
func Test_Named_DefaultLevelChange(t *testing.T) {
	oldDefault := DefaultLevel()
	defer SetDefaultLevel(oldDefault)

	RemoveLogger("ns.default")
	defer RemoveLogger("ns.default")

	l := Get("ns.default.test")
	if l.Level() != oldDefault {
		t.Fatalf("initial: %v", l.Level())
	}

	SetDefaultLevel(DebugLevel)
	if l.Level() != DebugLevel {
		t.Fatalf("should sense default change: got %v", l.Level())
	}
}

// SetLevelRecursive
func Test_Named_SetLevelRecursive(t *testing.T) {
	RemoveLogger("ns.rec")
	defer RemoveLogger("ns.rec")

	Get("ns.rec.parent")
	Get("ns.rec.parent.child")
	Get("ns.rec.parent.child.grandchild")
	Get("ns.rec.other")

	SetLevelRecursive("ns.rec.parent", ErrorLevel)

	if Get("ns.rec.parent").Level() != ErrorLevel {
		t.Error("parent level wrong")
	}
	if Get("ns.rec.parent.child").Level() != ErrorLevel {
		t.Error("child level wrong")
	}
	if Get("ns.rec.parent.child.grandchild").Level() != ErrorLevel {
		t.Error("grandchild level wrong")
	}
	// 不相关的不受影响
	if Get("ns.rec.other").Level() == ErrorLevel {
		t.Error("unrelated should not be affected")
	}
}

// RemoveLogger
func Test_Named_RemoveLogger(t *testing.T) {
	RemoveLogger("ns.remove")
	Get("ns.remove.a")
	Get("ns.remove.a.b")

	if RegistryCount() == 0 {
		t.Fatal("should have registered loggers")
	}

	RemoveLogger("ns.remove")

	// 重新 Get 应该是新实例
	l1 := Get("ns.remove.a")
	RemoveLogger("ns.remove")
	l2 := Get("ns.remove.a")
	if l1 == l2 {
		t.Fatal("after remove, Get should create new instance")
	}
	RemoveLogger("ns.remove")
}

// AllLogger 遍历
func Test_Named_AllLogger(t *testing.T) {
	RemoveLogger("ns.all")
	defer RemoveLogger("ns.all")

	Get("ns.all.x")
	Get("ns.all.y")

	seen := make(map[string]bool)
	AllLogger(func(name string) {
		if strings.HasPrefix(name, "ns.all.") {
			seen[name] = true
		}
	})
	if !seen["ns.all.x"] || !seen["ns.all.y"] {
		t.Fatalf("AllLogger missed entries: %v", seen)
	}
}

// 命名 Logger + With 派生
func Test_Named_WithDerived(t *testing.T) {
	var buf bytes.Buffer
	RemoveLogger("ns.withd")
	defer RemoveLogger("ns.withd")

	parent := Get("ns.withd", WithWriter(&buf))
	derived := parent.With().Str("trace", "abc").Logger()

	derived.Info().Msg("tracked")
	out := buf.String()
	if !strings.Contains(out, "trace=abc") || !strings.Contains(out, "tracked") {
		t.Fatalf("derived output: %q", out)
	}

	// 派生 Logger 级别独立
	derived.SetLevel(ErrorLevel)
	if parent.Level() == ErrorLevel {
		t.Fatal("parent level should not change")
	}
}

// --- 命名 Logger benchmark ---

func Benchmark_Named_Get(b *testing.B) {
	RemoveLogger("bench.named")
	Get("bench.named")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get("bench.named")
	}
	RemoveLogger("bench.named")
}

func Benchmark_Named_GetCreate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RemoveLogger("bench.create")
		Get("bench.create")
	}
}

// --- BDD: Hook 系统 ---

func Test_Hook_Output(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	var buf bytes.Buffer
	oldW := DefaultWriter()
	SetDefaultWriter(&buf)
	defer SetDefaultWriter(oldW)

	AddHook(func(e *Event) {
		e.AppendString(" [hooked]")
	})

	New("hook.test").Info().Msg("hello")
	out := buf.String()
	if !strings.Contains(out, "[hooked]") {
		t.Fatalf("hook output missing: %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("message missing: %q", out)
	}
}

func Test_Hook_WithContext(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	type ctxKey struct{}
	var captured context.Context

	AddHook(func(e *Event) {
		captured = e.Context()
		if ctx := e.Context(); ctx != nil {
			if v, ok := ctx.Value(ctxKey{}).(string); ok {
				e.Str("user", v)
			}
		}
	})

	l := New("ctx.test", WithLevel(InfoLevel), WithWriter(io.Discard))
	ctx := context.WithValue(context.Background(), ctxKey{}, "user123")

	// WithContext 注入 ctx, hook 能读取
	l.WithContext(ctx).Info().Msg("login")
	if captured == nil {
		t.Fatal("hook should receive non-nil ctx")
	}
	if v := captured.Value(ctxKey{}); v != "user123" {
		t.Fatalf("ctx value mismatch: got %v", v)
	}

	// 不传 ctx, hook 收到 nil
	captured = nil
	l.Info().Msg("plain")
	if captured != nil {
		t.Fatal("hook should receive nil ctx when not set")
	}

	// WithKvs 派生 Logger 继承 ctx
	captured = nil
	l.WithContext(ctx).WithKvs("svc", "api").Info().Msg("derived")
	if captured == nil {
		t.Fatal("derived Logger should inherit ctx")
	}
}

func Test_Hook_RemoveAndClean(t *testing.T) {
	CleanHooks()

	marker := func(e *Event) { e.AppendString(" [m]") }
	AddHook(marker)
	if atomic.LoadInt32(&hasHooks) != 1 {
		t.Fatal("hasHooks should be 1")
	}

	RemoveHook(marker)
	if atomic.LoadInt32(&hasHooks) != 0 {
		t.Fatal("hasHooks should be 0 after remove")
	}

	AddHook(func(e *Event) {})
	AddHook(func(e *Event) {})
	CleanHooks()
	if atomic.LoadInt32(&hasHooks) != 0 {
		t.Fatal("hasHooks should be 0 after clean")
	}
}

func Test_Hook_MaxLimit(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	for i := 0; i < maxHooks; i++ {
		if !AddHook(func(e *Event) {}) {
			t.Fatalf("add %d should succeed", i)
		}
	}
	if AddHook(func(e *Event) {}) {
		t.Fatal("exceeding maxHooks should fail")
	}
}

func Test_Hook_ConcurrentAddAndLog(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	SetDefaultWriter(io.Discard)
	defer SetDefaultWriter(os.Stdout)

	var count int64
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			AddHook(func(e *Event) {
				atomic.AddInt64(&count, 1)
			})
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			New("hook.concurrent").Info().Msg("msg")
		}
	}()
	wg.Wait()
}

// --- BDD: Caller ---

func Test_Caller_BasicOutput(t *testing.T) {
	oldCfg := atomic.LoadInt32(&callerConfig)
	defer atomic.StoreInt32(&callerConfig, oldCfg)

	var buf bytes.Buffer
	l := New("caller.test", WithLevel(InfoLevel), WithWriter(&buf))

	SetCaller(true)
	buf.Reset()
	l.Info().Msg("with caller")
	if !strings.Contains(buf.String(), ".go:") {
		t.Fatalf("caller info missing: %q", buf.String())
	}

	SetCaller(false)
	buf.Reset()
	l.Info().Msg("no caller")
	if strings.Contains(buf.String(), ".go:") {
		t.Fatalf("caller should be absent: %q", buf.String())
	}
}

func Test_Caller_FuncName(t *testing.T) {
	oldCfg := atomic.LoadInt32(&callerConfig)
	defer atomic.StoreInt32(&callerConfig, oldCfg)

	var buf bytes.Buffer
	l := New("caller.func", WithLevel(InfoLevel), WithWriter(&buf))

	SetCaller(true)
	SetCallerFunc(true)
	l.Info().Msg("with func")
	out := buf.String()
	if !strings.Contains(out, ".go:") {
		t.Fatalf("caller missing: %q", out)
	}
	// 函数名格式: "Test_Caller_FuncName caller_test.go:line"
	if !strings.Contains(out, "Test_Caller_FuncName ") {
		t.Fatalf("func name missing: %q", out)
	}
}

// --- Hook + Caller benchmark ---

func Benchmark_Logger_WithHook(b *testing.B) {
	CleanHooks()
	defer CleanHooks()
	AddHook(func(e *Event) {})

	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("hooked")
	}
}

func Benchmark_Logger_WithCaller(b *testing.B) {
	oldCfg := atomic.LoadInt32(&callerConfig)
	defer atomic.StoreInt32(&callerConfig, oldCfg)
	SetCaller(true)

	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("caller")
	}
}

// --- BDD: 异步写入 ---

// mockAsyncWriter 模拟异步 writer, 用于测试异步路径
type mockAsyncWriter struct {
	mu    sync.Mutex
	bufs  [][]byte
	calls int
}

func (m *mockAsyncWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (m *mockAsyncWriter) WriterAsync(p []byte, callback func()) error {
	cp := make([]byte, len(p))
	copy(cp, p)
	m.mu.Lock()
	m.bufs = append(m.bufs, cp)
	m.calls++
	m.mu.Unlock()
	if callback != nil {
		callback()
	}
	return nil
}

func Test_AsyncWriter_Path(t *testing.T) {
	aw := &mockAsyncWriter{}
	l := New("async", WithLevel(InfoLevel), WithWriter(aw))

	l.Info().Str("k", "v").Msg("async msg")

	aw.mu.Lock()
	defer aw.mu.Unlock()
	if aw.calls != 1 {
		t.Fatalf("expected 1 async call, got %d", aw.calls)
	}
	out := string(aw.bufs[0])
	if !strings.Contains(out, "async msg") || !strings.Contains(out, "k=v") {
		t.Fatalf("async output wrong: %q", out)
	}
}

func Test_AsyncWriter_ConcurrentSafe(t *testing.T) {
	aw := &mockAsyncWriter{}
	l := New("async.concurrent", WithLevel(TraceLevel), WithWriter(aw))

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				l.Info().Int("i", j).Msg("msg")
			}
		}()
	}
	wg.Wait()

	aw.mu.Lock()
	defer aw.mu.Unlock()
	if aw.calls != 400 {
		t.Fatalf("expected 400 calls, got %d", aw.calls)
	}
}

// --- BDD: 时间格式 ---

func Test_Format_YearMode(t *testing.T) {
	oldMode := atomic.LoadInt32(&yearMode)
	defer SetYearMode(oldMode)

	var buf bytes.Buffer
	l := New("yearmode", WithLevel(InfoLevel), WithWriter(&buf))

	// YearFull: YYYY-
	SetYearMode(YearFull)
	buf.Reset()
	l.Info().Msg("full")
	if len(buf.String()) < 4 || buf.String()[4] != '-' {
		t.Fatalf("YearFull should start with YYYY-: %q", buf.String())
	}

	// YearShort: YY-
	SetYearMode(YearShort)
	buf.Reset()
	l.Info().Msg("short")
	if len(buf.String()) < 2 || buf.String()[2] != '-' {
		t.Fatalf("YearShort should start with YY-: %q", buf.String())
	}

	// YearNone: MM-DD
	SetYearMode(YearNone)
	buf.Reset()
	l.Info().Msg("none")
	if len(buf.String()) < 5 || buf.String()[2] != '-' {
		t.Fatalf("YearNone should start with MM-DD: %q", buf.String())
	}
}

// --- 异步写入 benchmark ---

// benchAsyncWriter 不 copy 数据, 模拟真实 helper.AsyncWrite 行为
// (真实 AsyncWrite 在后台 goroutine 中 copy, WriterAsync 本身不 copy)
type benchAsyncWriter struct{}

func (benchAsyncWriter) Write(p []byte) (int, error) { return len(p), nil }
func (benchAsyncWriter) WriterAsync(p []byte, callback func()) error {
	if callback != nil {
		callback()
	}
	return nil
}

func Benchmark_Logger_AsyncWriter(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(benchAsyncWriter{}))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info().Int("id", i).Msg("async")
	}
}

func Benchmark_Logger_WithKvs_Create(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.WithKvs("svc", "api", "ver", 2)
	}
}

func Benchmark_Logger_WithKvs_OddArgs(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.WithKvs("svc", "api", "ver")
	}
}

func Benchmark_Logger_WithKvs_Log(b *testing.B) {
	l := New("bench", WithLevel(InfoLevel), WithWriter(io.Discard)).WithKvs("svc", "api", "ver", 2)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info().Int("id", i).Msg("request")
	}
}

// --- 旧版命令式 API 测试 ---

// D/I/L 便捷方法, {}占位符格式化和级别过滤
func Test_Legacy_ConvenienceMethods(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)

	// D 在 InfoLevel 下不输出
	l.D("debug-msg", "arg")
	if buf.Len() != 0 {
		t.Fatalf("D should be filtered at InfoLevel: %q", buf.String())
	}

	// I 占位符替换
	l.I("hello {}", "world")
	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Fatalf("placeholder {} not replaced: %q", out)
	}

	// 显式位置占位符
	buf.Reset()
	l.L(InfoLevel, "{0} -> {1}", "a", "b")
	out = buf.String()
	if !strings.Contains(out, "a -> b") {
		t.Fatalf("explicit placeholder not replaced: %q", out)
	}

	// 无占位符时追加未消费参数
	buf.Reset()
	l.I("no slots", "extra1", "extra2")
	out = buf.String()
	if !strings.Contains(out, "no slots extra1 extra2") {
		t.Fatalf("unconsumed args not appended: %q", out)
	}

	// %x 格式化动词
	buf.Reset()
	l.I("count={%d} name={%s}", 42, "moke")
	out = buf.String()
	if !strings.Contains(out, "count=42 name=moke") {
		t.Fatalf("%%x format not applied: %q", out)
	}

	// 显式位置 + 格式化
	buf.Reset()
	l.I("{0%s} and {1%d}", "hello", 99)
	out = buf.String()
	if !strings.Contains(out, "hello and 99") {
		t.Fatalf("idx+%%x format not applied: %q", out)
	}
}

// DF/IF/LF 延迟参数方法, 级别不够时 f 不执行
func Test_Legacy_DeferredMethods(t *testing.T) {
	t.Parallel()

	// DebugLevel: DF 占位符输出
	var buf bytes.Buffer
	l := newTestLogger(&buf, DebugLevel)
	l.DF("user {} login", func() []any { return []any{"moke"} })
	out := buf.String()
	if !strings.Contains(out, "user moke login") {
		t.Fatalf("DF placeholder output wrong: %q", out)
	}

	// InfoLevel: DF 延迟函数不调用
	l2 := newTestLogger(&bytes.Buffer{}, InfoLevel)
	l2.DF("msg", func() []any {
		t.Fatal("DF deferred function should not be called when filtered")
		return nil
	})
}

// WithKvs 派生 Logger, 预设字段出现在输出中
func Test_Legacy_WithKvs(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := newTestLogger(&buf, InfoLevel)
	l2 := l.WithKvs("svc", "api", "ver", 2)

	l2.I("request")
	out := buf.String()
	for _, s := range []string{"svc=api", "ver=2", "request"} {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

// LevelUnset 常量, FullName 和 String 方法
func Test_LevelUnset_And_FullName(t *testing.T) {
	t.Parallel()

	if LevelUnset != Level(-1) {
		t.Fatalf("LevelUnset should be Level(-1), got %v", LevelUnset)
	}
	if LevelUnset.FullName() != "Unset" {
		t.Fatalf("LevelUnset.FullName() = %q, want %q", LevelUnset.FullName(), "Unset")
	}
	if InfoLevel.FullName() != "Info" {
		t.Fatalf("InfoLevel.FullName() = %q, want %q", InfoLevel.FullName(), "Info")
	}
	if InfoLevel.String() != "Info" {
		t.Fatalf("InfoLevel.String() = %q, want %q", InfoLevel.String(), "Info")
	}
}

// JSONFormatter 链式 API, 输出 JSON 行
func Test_JSONFormatter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := New("test", WithLevel(InfoLevel), WithWriter(&buf), WithFormatter(JSONFormatter{}))

	l.Info().Str("user", "moke").Int("id", 42).Msg("login")
	out := buf.String()

	if !strings.HasPrefix(out, "{") {
		t.Fatalf("should start with '{': %q", out)
	}
	if !strings.HasSuffix(out, "}\n") {
		t.Fatalf("should end with '}\\n': %q", out)
	}
	checks := []string{
		`"lv":"Info"`,
		`"user":"moke"`,
		`"id":42`,
		`"msg":"login"`,
		`"logger"`,
	}
	for _, s := range checks {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

// JSONFormatter + 命令式便捷方法
func Test_JSONFormatter_Convenience(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	l := New("test", WithLevel(InfoLevel), WithWriter(&buf), WithFormatter(JSONFormatter{}))

	l.I("hello {%s}", "world")
	out := buf.String()

	if !strings.HasPrefix(out, "{") || !strings.HasSuffix(out, "}\n") {
		t.Fatalf("expected JSON line: %q", out)
	}
	if !strings.Contains(out, `"msg":"hello world"`) {
		t.Fatalf("missing msg field: %q", out)
	}
}

// 旧版兼容函数: SetDefaultLogLevel/DefaultLogLevel/SetDefaultWriterTarget 等
func Test_Legacy_CompatFunctions(t *testing.T) {
	// 修改全局状态, 不并行

	// SetDefaultLogLevel / DefaultLogLevel
	oldLevel := DefaultLogLevel()
	defer SetDefaultLogLevel(oldLevel)
	if !SetDefaultLogLevel(InfoLevel) {
		t.Fatal("SetDefaultLogLevel should return true")
	}
	if DefaultLogLevel() != InfoLevel {
		t.Fatalf("DefaultLogLevel = %v, want InfoLevel", DefaultLogLevel())
	}

	// SetDefaultWriterTarget / DefaultWriterTarget
	oldWriter := DefaultWriterTarget()
	defer SetDefaultWriterTarget(oldWriter)
	var buf bytes.Buffer
	SetDefaultWriterTarget(&buf)
	if DefaultWriterTarget() == nil {
		t.Fatal("DefaultWriterTarget should not be nil")
	}

	// SetCallerInfo 不 panic
	oldCaller := atomic.LoadInt32(&callerConfig)
	defer atomic.StoreInt32(&callerConfig, oldCaller)
	SetCallerInfo(true)

	// SetPrintYearInfo 不 panic
	oldYear := atomic.LoadInt32(&yearMode)
	defer SetYearMode(oldYear)
	SetPrintYearInfo(0)
}
