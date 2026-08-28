package logger

import (
	"bytes"
	"context"
	"io"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- BDD: 双 hook 点 ---

// Given 已注册 PreHook 与 PostHook
// When 输出一条消息
// Then PreHook 内容在消息前(name后), PostHook 内容在消息后
func Test_Hook_PreAndPostOrder(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	AddHook(func(e *Event) { e.AppendString(" [pre]") })
	AddPostHook(func(e *Event) { e.AppendString(" [post]") })

	var buf bytes.Buffer
	l := New("hook.order", WithLevel(InfoLevel), WithWriter(&buf))
	l.Info().Msg("mid")

	out := buf.String()
	pre := strings.Index(out, "[pre]")
	msg := strings.Index(out, "mid")
	post := strings.Index(out, "[post]")
	if pre < 0 || msg < 0 || post < 0 {
		t.Fatalf("missing parts: %q", out)
	}
	if !(pre < msg && msg < post) {
		t.Fatalf("order wrong, want pre < msg < post: %q", out)
	}
}

// Given 仅注册 PostHook
// When 输出消息
// Then PostHook 内容在消息后、换行前
func Test_Hook_PostOnly(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	called := 0
	AddPostHook(func(e *Event) {
		called++
		e.AppendString(" [post]")
	})

	var buf bytes.Buffer
	l := New("hook.post", WithLevel(InfoLevel), WithWriter(&buf))
	l.Info().Msg("body")

	out := buf.String()
	if called != 1 {
		t.Fatalf("post hook called %d times, want 1", called)
	}
	msgIdx := strings.Index(out, "body")
	postIdx := strings.Index(out, "[post]")
	if msgIdx < 0 || postIdx < 0 || postIdx < msgIdx {
		t.Fatalf("post hook content should follow message: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("newline should be last: %q", out)
	}
}

// Given 被过滤的级别
// When 调用 Debug
// Then Pre/Post hook 均不执行
func Test_Hook_NotRunWhenFiltered(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	preCalled, postCalled := 0, 0
	AddHook(func(e *Event) { preCalled++ })
	AddPostHook(func(e *Event) { postCalled++ })

	l := New("hook.filtered", WithLevel(WarnLevel), WithWriter(io.Discard))
	l.Debug().Msg("filtered")

	if preCalled != 0 || postCalled != 0 {
		t.Fatalf("hooks should not run when filtered: pre=%d post=%d", preCalled, postCalled)
	}
}

// Given PreHook 读取 Context
// When WithContext 注入后输出
// Then PreHook 可取到 ctx 值
func Test_Hook_PreHookContext(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	type ctxKey struct{}
	var got any
	AddHook(func(e *Event) {
		if ctx := e.Context(); ctx != nil {
			got = ctx.Value(ctxKey{})
		}
	})

	l := New("hook.prectx", WithLevel(InfoLevel), WithWriter(io.Discard))
	ctx := context.WithValue(context.Background(), ctxKey{}, "u1")
	l.WithContext(ctx).Info().Msg("x")

	if got != "u1" {
		t.Fatalf("pre hook ctx value = %v, want u1", got)
	}
}

// --- BDD: hookList 并发注册精确性 ---

// Given 并发注册 N 个不同 hook, 总量不超上限
// When 全部 add
// Then 列表长度恰好为 N, 无重复无丢失
func Test_HookList_ConcurrentAddExact(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	const n = 32
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !AddHook(func(e *Event) {}) {
				t.Error("add should succeed within limit")
			}
		}()
	}
	wg.Wait()

	if got := len(preHooks.va.Load().([]Hook)); got != n {
		t.Fatalf("hook count = %d, want %d", got, n)
	}
	if atomic.LoadInt32(&preHooks.has) != 1 {
		t.Fatal("has flag should be 1")
	}
}

// Given hook 已注册
// When RemoveHook
// Then 从 Pre 与 Post 同时移除
func Test_Hook_RemoveFromBothLists(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	h := func(e *Event) {}
	AddHook(h)
	AddPostHook(h)

	RemoveHook(h)

	if got := len(preHooks.va.Load().([]Hook)); got != 0 {
		t.Fatalf("pre list should be empty, got %d", got)
	}
	if got := len(postHooks.va.Load().([]Hook)); got != 0 {
		t.Fatalf("post list should be empty, got %d", got)
	}
}

// Given hook 空参数注册
// When AddHook 无参数
// Then 返回 true 且列表不变
func Test_Hook_AddEmptyNoop(t *testing.T) {
	CleanHooks()
	defer CleanHooks()

	if !AddHook() {
		t.Fatal("add with no args should return true")
	}
	if got := len(preHooks.va.Load().([]Hook)); got != 0 {
		t.Fatalf("list should stay empty, got %d", got)
	}
}

// --- BDD: SetFormatter 即时生效 ---

// Given 已存在的 Log 无本地 Formatter
// When SetFormatter 替换全局 Formatter
// Then 该 Log 后续事件立即使用新 Formatter
func Test_SetFormatter_TakesEffectImmediately(t *testing.T) {
	old := DefaultFormatter()
	defer SetFormatter(old)

	var buf bytes.Buffer
	l := New("fmt.live", WithLevel(InfoLevel), WithWriter(&buf))

	SetFormatter(JSONFormatter{})
	l.Info().Msg("after")

	out := buf.String()
	if !strings.HasPrefix(out, "{") {
		t.Fatalf("new events should use new formatter: %q", out)
	}
	if !strings.Contains(out, `"msg":"after"`) {
		t.Fatalf("json msg missing: %q", out)
	}
}

// Given Log 创建时指定 WithFormatter
// When 全局 SetFormatter 变更
// Then 该 Log 仍使用本地 Formatter
func Test_SetFormatter_LocalOverrideWins(t *testing.T) {
	old := DefaultFormatter()
	defer SetFormatter(old)

	var buf bytes.Buffer
	l := New("fmt.local", WithLevel(InfoLevel), WithWriter(&buf), WithFormatter(JSONFormatter{}))

	SetFormatter(ConsoleFormatter{})
	l.Info().Msg("keep")

	out := buf.String()
	if !strings.HasPrefix(out, "{") {
		t.Fatalf("local formatter should win: %q", out)
	}
}

// --- BDD: Begin 超长 name 截断 ---

// Given ConsoleFormatter NameWidth 为 w, name 长度超过 w
// When 输出
// Then name 被截断为前 w 字节
func Test_ConsoleBegin_LongNameTruncated(t *testing.T) {
	var buf bytes.Buffer
	l := New("abcdefghijk", WithLevel(InfoLevel), WithWriter(&buf), WithFormatter(ConsoleFormatter{NameWidth: 4}))
	l.Info().Msg("x")

	out := buf.String()
	if !strings.Contains(out, "[abcd]") {
		t.Fatalf("name should be truncated to width: %q", out)
	}
	if strings.Contains(out, "abcdefghijk") {
		t.Fatalf("full name should not appear: %q", out)
	}
}

// Given name 短于 NameWidth
// When 输出
// Then name 右侧补空格对齐
func Test_ConsoleBegin_ShortNamePadded(t *testing.T) {
	var buf bytes.Buffer
	l := New("ab", WithLevel(InfoLevel), WithWriter(&buf), WithFormatter(ConsoleFormatter{NameWidth: 4}))
	l.Info().Msg("x")

	if !strings.Contains(buf.String(), "[  ab]") {
		t.Fatalf("name should be right-aligned padded: %q", buf.String())
	}
}

// --- BDD: caller 渲染 ---

// Given 合法 pc
// When appendCaller showFunc=false
// Then 输出仅含文件名(无目录分隔符)与行号
func Test_AppendCaller_FileNameOnly(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	out := string(appendCaller(nil, pc, false))
	if out == "" {
		t.Fatal("caller output should not be empty")
	}
	if strings.ContainsAny(out, "/\\") {
		t.Fatalf("should keep only base file name: %q", out)
	}
	if !strings.Contains(out, "caller_api_test.go:") && !strings.Contains(out, ".go:") {
		t.Fatalf("should contain file:line: %q", out)
	}
}

// Given 合法 pc
// When appendCaller showFunc=true
// Then 输出以函数短名(不含包路径)开头, 后跟 file:line
func Test_AppendCaller_FuncShortName(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	out := string(appendCaller(nil, pc, true))

	idx := strings.Index(out, " ")
	if idx <= 0 {
		t.Fatalf("func name should be followed by space: %q", out)
	}
	fn := out[:idx]
	if strings.Contains(fn, "/") || strings.Contains(fn, "go-util") {
		t.Fatalf("func name should be short: %q", fn)
	}
	if !strings.Contains(fn, "Test_AppendCaller_FuncShortName") {
		t.Fatalf("func short name wrong: %q", fn)
	}
	if !strings.Contains(out[idx:], ".go:") {
		t.Fatalf("file:line missing: %q", out)
	}
}

// Given pc 为 0
// When appendCaller
// Then 不追加任何内容
func Test_AppendCaller_ZeroPcNoop(t *testing.T) {
	if got := string(appendCaller(nil, 0, true)); got != "" {
		t.Fatalf("zero pc should append nothing, got %q", got)
	}
}

// Given 同一 pc 与 showFunc
// When 多次 appendCaller
// Then 输出一致(缓存命中路径)
func Test_AppendCaller_CacheConsistent(t *testing.T) {
	pc, _, _, _ := runtime.Caller(0)
	first := string(appendCaller(nil, pc, false))
	second := string(appendCaller(nil, pc, false))
	if first != second {
		t.Fatalf("cache path changed output: %q vs %q", first, second)
	}
}

// --- BDD: WithWriter 运行时替换 ---

// Given 命名 Log 已绑定 writer1
// When 再次 Get 并 WithWriter(writer2)
// Then 后续事件写入 writer2 且不再写 writer1
func Test_WithWriter_RuntimeReplace(t *testing.T) {
	RemoveLogger("wr.live")
	defer RemoveLogger("wr.live")

	var buf1, buf2 bytes.Buffer
	l := Get("wr.live", WithLevel(InfoLevel), WithWriter(&buf1))
	l.Info().Msg("first")
	if !strings.Contains(buf1.String(), "first") {
		t.Fatalf("first write should hit buf1: %q", buf1.String())
	}

	Get("wr.live", WithWriter(&buf2))
	l.Info().Msg("second")

	if strings.Contains(buf1.String(), "second") {
		t.Fatalf("old writer should not receive new events: %q", buf1.String())
	}
	if !strings.Contains(buf2.String(), "second") {
		t.Fatalf("new writer should receive events: %q", buf2.String())
	}
}

// Given writer 从同步换到异步再换回同步
// When 各输出一条
// Then 三个 writer 各自收到对应消息(同步兜底路径不丢内容)
func Test_WithWriter_AsyncToSyncFallback(t *testing.T) {
	var sync1, sync2 bytes.Buffer
	aw := &mockAsyncWriter{}

	RemoveLogger("wr.fallback")
	defer RemoveLogger("wr.fallback")

	l := Get("wr.fallback", WithLevel(InfoLevel), WithWriter(&sync1))
	l.Info().Msg("a")

	Get("wr.fallback", WithWriter(aw))
	l.Info().Msg("b")

	aw.mu.Lock()
	asyncCalls := aw.calls
	aw.mu.Unlock()
	if asyncCalls != 1 {
		t.Fatalf("async writer should be used once, got %d", asyncCalls)
	}

	// 换回同步 writer, aw 已被置 nil, flush 走同步兜底
	Get("wr.fallback", WithWriter(&sync2))
	l.Info().Msg("c")

	if !strings.Contains(sync2.String(), "c") {
		t.Fatalf("fallback sync write missing: %q", sync2.String())
	}
}

// --- BDD: appendDuration 极值边界 ---

// Given MinInt64 与 MaxInt64 时长
// When appendDuration
// Then 与标准库 String 一致
func Test_AppendDuration_ExtremeValues(t *testing.T) {
	for _, d := range []time.Duration{
		-1 << 63,  // MinInt64
		1<<63 - 1, // MaxInt64
		-1<<63 + 1,
	} {
		got := string(appendDuration(nil, d))
		want := d.String()
		if got != want {
			t.Errorf("appendDuration(%d) = %q, want %q", int64(d), got, want)
		}
	}
}

// --- BDD: 旧版兼容别名冒烟 ---

// Given 命名 Log
// When Logger(name) 与 Get(name)
// Then 返回同一实例
func Test_Legacy_LoggerAlias(t *testing.T) {
	RemoveLogger("legacy.alias")
	defer RemoveLogger("legacy.alias")

	if Logger("legacy.alias") != Get("legacy.alias") {
		t.Fatal("Logger should alias Get")
	}
}

// Given 旧版 SetLevel
// When 按名称设置级别
// Then 等价于 SetLevelByName
func Test_Legacy_SetLevelAlias(t *testing.T) {
	RemoveLogger("legacy.setlv")
	defer RemoveLogger("legacy.setlv")

	Get("legacy.setlv")
	SetLevel("legacy.setlv", ErrorLevel)
	if Get("legacy.setlv").Level() != ErrorLevel {
		t.Fatal("SetLevel should apply by name")
	}
	if SetLevelByName("legacy.setlv.missing", DebugLevel) {
		t.Fatal("SetLevelByName on missing name should return false")
	}
}

// WithFormatter/WithWriter 传 nil 忽略
func Test_Option_NilIgnored(t *testing.T) {
	var buf bytes.Buffer
	l := New("opt.nil", WithLevel(InfoLevel), WithWriter(nil), WithFormatter(nil), WithWriter(&buf))
	l.Info().Msg("ok")
	if !strings.Contains(buf.String(), "ok") {
		t.Fatalf("nil options should be ignored: %q", buf.String())
	}
}
