package script

import (
	"bytes"
	"testing"
)

// ========== Context API 测试 ==========

func Test_Context_BindAndGet(t *testing.T) {
	ctx := NewContext()

	ctx.BindValue("x", 42)
	ctx.BindValue("s", "hello")
	ctx.BindValue("b", true)
	ctx.BindValue("n", nil)

	tests := []struct {
		name     string
		key      string
		hasValue bool
	}{
		{"int绑定", "x", true},
		{"string绑定", "s", true},
		{"bool绑定", "b", true},
		{"nil绑定", "n", true},
		{"不存在", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := ctx.GetBindValue(tt.key)
			if ok != tt.hasValue {
				t.Errorf("GetBindValue(%q) ok = %v, want %v", tt.key, ok, tt.hasValue)
			}
			if tt.hasValue && val.Type == TypeNil && tt.key != "n" {
				t.Errorf("GetBindValue(%q) 返回了nil值", tt.key)
			}
		})
	}
}

func Test_Context_BindFuncAndGet(t *testing.T) {
	ctx := NewContext()

	add := func(a, b int) int { return a + b }
	ctx.BindFunc("add", add)

	fn, ok := ctx.GetBindFunc("add")
	if !ok {
		t.Fatal("GetBindFunc应该找到已绑定的函数")
	}
	if fn == nil {
		t.Fatal("GetBindFunc不应该返回nil")
	}

	_, ok = ctx.GetBindFunc("missing")
	if ok {
		t.Error("GetBindFunc对不存在的函数应返回false")
	}
}

func Test_Context_BindOverwrite(t *testing.T) {
	ctx := NewContext()
	ctx.BindValue("x", 1)
	ctx.BindValue("x", 2)

	val, _ := ctx.GetBindValue("x")
	if val.Int() != 2 {
		t.Errorf("覆盖绑定后应返回最新值, got %d", val.Int())
	}
}

func Test_Context_SetLogFunc(t *testing.T) {
	ctx := NewContext()

	var loggedMsg string
	ctx.SetLogFunc(func(format string, args ...any) {
		loggedMsg = format
	})

	ctx.Log("test message")

	if loggedMsg != "test message" {
		t.Errorf("LogFunc未正确接收消息, got %q", loggedMsg)
	}
}

func Test_Context_LogWithArgs(t *testing.T) {
	ctx := NewContext()

	var loggedMsg string
	ctx.SetLogFunc(func(format string, args ...any) {
		loggedMsg = format
	})

	ctx.Log("hello %s", "world")

	if loggedMsg != "hello %s" {
		t.Errorf("format应原样传递, got %q", loggedMsg)
	}
}

func Test_Context_SetOutputAndGetOutput(t *testing.T) {
	ctx := NewContext()

	var buf bytes.Buffer
	ctx.SetOutput(&buf)

	if ctx.GetOutput() == nil {
		t.Error("GetOutput不应返回nil")
	}

	// 默认输出不应为nil
	ctx2 := NewContext()
	if ctx2.GetOutput() == nil {
		t.Error("默认输出不应为nil")
	}
}

func Test_Context_GetStats(t *testing.T) {
	ctx := NewContext()

	// 初始统计应为零值
	stats := ctx.GetStats()
	if stats.TotalRuns != 0 {
		t.Errorf("初始TotalRuns应为0, got %d", stats.TotalRuns)
	}

	// 执行脚本后统计应更新
	runScriptWithContext(t, "1 + 2", ctx)

	stats = ctx.GetStats()
	if stats.TotalRuns != 1 {
		t.Errorf("执行一次后TotalRuns应为1, got %d", stats.TotalRuns)
	}
}

func Test_Context_GetStats_AfterMultipleRuns(t *testing.T) {
	ctx := NewContext()

	for i := 0; i < 5; i++ {
		runScriptWithContext(t, "1 + 1", ctx)
	}

	stats := ctx.GetStats()
	if stats.TotalRuns != 5 {
		t.Errorf("执行5次后TotalRuns应为5, got %d", stats.TotalRuns)
	}
}

func Test_Context_Clone(t *testing.T) {
	ctx := NewContext()
	ctx.BindValue("x", 42)
	ctx.BindValue("s", "hello")
	ctx.BindFunc("add", func(a, b int) int { return a + b })

	clone := ctx.Clone()

	// 克隆应保留绑定的值
	val, ok := clone.GetBindValue("x")
	if !ok || val.Int() != 42 {
		t.Errorf("Clone后x应存在且值为42, got %v, ok=%v", val, ok)
	}

	val2, ok := clone.GetBindValue("s")
	if !ok || val2.String() != "hello" {
		t.Errorf("Clone后s应存在且值为hello, got %v, ok=%v", val2, ok)
	}

	// 克隆应保留绑定的函数
	_, ok = clone.GetBindFunc("add")
	if !ok {
		t.Error("Clone后add函数应存在")
	}
}

func Test_Context_Clone_Independence(t *testing.T) {
	ctx := NewContext()
	ctx.BindValue("x", 1)

	clone := ctx.Clone()

	// 修改克隆不应影响原Context
	clone.BindValue("y", 99)
	_, ok := ctx.GetBindValue("y")
	if ok {
		t.Error("修改Clone不应影响原Context")
	}

	// 修改原不应影响克隆
	ctx.BindValue("z", 88)
	_, ok = clone.GetBindValue("z")
	if ok {
		t.Error("修改原Context不应影响Clone")
	}
}

// ========== Engine API 测试 ==========

func Test_Engine_NewEngine(t *testing.T) {
	e1 := NewEngine()
	if e1 == nil {
		t.Fatal("NewEngine不应返回nil")
	}

	e2 := NewEngine(WithMaxCallDepth(512))
	if e2 == nil {
		t.Fatal("NewEngine with option不应返回nil")
	}
}

func Test_Engine_WithMaxCallDepth(t *testing.T) {
	e := NewEngine(WithMaxCallDepth(100))

	ctx := NewContext()
	script := compileScript(t, "1 + 1")
	result, err := e.Run(ctx, script)
	assertNoError(t, err)
	if result.Int() != 2 {
		t.Errorf("结果应为2, got %d", result.Int())
	}
}

func Test_Engine_Run_Basic(t *testing.T) {
	e := NewEngine()
	ctx := NewContext()

	tests := []struct {
		input    string
		expected int
	}{
		{"1 + 2", 3},
		{"10 - 5", 5},
		{"3 * 4", 12},
		{"20 / 4", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			script := compileScript(t, tt.input)
			result, err := e.Run(ctx, script)
			assertNoError(t, err)
			if result.Int() != tt.expected {
				t.Errorf("got %d, want %d", result.Int(), tt.expected)
			}
		})
	}
}

func Test_Engine_Run_UpdatesStats(t *testing.T) {
	e := NewEngine()
	ctx := NewContext()
	script := compileScript(t, "1 + 1")

	for i := 0; i < 3; i++ {
		e.Run(ctx, script)
	}

	stats := ctx.GetStats()
	if stats.TotalRuns != 3 {
		t.Errorf("TotalRuns应为3, got %d", stats.TotalRuns)
	}
}

func Test_Engine_Run_ReturnsError(t *testing.T) {
	e := NewEngine()
	ctx := NewContext()
	script := compileScript(t, `throw "test error"`)

	_, err := e.Run(ctx, script)
	if err == nil {
		t.Error("throw应该返回错误")
	}
}

func Test_Engine_Stop(t *testing.T) {
	e := NewEngine()

	// Stop应不panic，即使没有运行中的脚本
	e.Stop()

	// 执行简单脚本后Stop也应正常
	ctx := NewContext()
	script := compileScript(t, "1 + 1")
	e.Run(ctx, script)
	e.Stop()
}

// ========== Script API 测试 ==========

func Test_Script_NewAndGetCompiled(t *testing.T) {
	cs := compileScript(t, "1 + 2")
	s := NewScript(cs)

	if s.GetCompiled() == nil {
		t.Error("GetCompiled不应返回nil")
	}
}

func Test_Script_Clone(t *testing.T) {
	cs := compileScript(t, "1 + 2")
	s1 := NewScript(cs)
	s2 := s1.Clone()

	if s2.GetCompiled() == nil {
		t.Error("Clone后的Script应有compiled数据")
	}
}

func Test_EvalWithBindings(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		binds    map[string]interface{}
		expected int
	}{
		{
			name: "int绑定",
			source: `x :=>int getBindValue("x")
x + 10`,
			binds:    map[string]interface{}{"x": 5},
			expected: 15,
		},
		{
			name: "多绑定",
			source: `x :=>int getBindValue("x")
y :=>int getBindValue("y")
x + y`,
			binds:    map[string]interface{}{"x": 10, "y": 20},
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvalWithBindings(tt.source, tt.binds)
			assertNoError(t, err)
			if result.Int() != tt.expected {
				t.Errorf("got %d, want %d", result.Int(), tt.expected)
			}
		})
	}
}

func Test_EvalWithBindings_Error(t *testing.T) {
	_, err := EvalWithBindings("invalid syntax !!", nil)
	if err == nil {
		t.Error("语法错误应返回error")
	}
}

func Test_MustEvalWithBindings(t *testing.T) {
	result := MustEvalWithBindings("1 + 2", nil)
	if result.Int() != 3 {
		t.Errorf("got %d, want 3", result.Int())
	}
}

func Test_MustEvalWithBindings_PanicOnSyntaxError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustEvalWithBindings在错误时应panic")
		}
	}()

	MustEvalWithBindings("!!!invalid", nil)
}

// ========== Context并发安全测试 ==========

func Test_Context_ConcurrentAccess(t *testing.T) {
	ctx := NewContext()
	done := make(chan bool, 4)

	// 并发写入
	go func() {
		for i := 0; i < 100; i++ {
			ctx.BindValue("x", i)
		}
		done <- true
	}()

	// 并发读取
	go func() {
		for i := 0; i < 100; i++ {
			ctx.GetBindValue("x")
		}
		done <- true
	}()

	// 并发统计
	go func() {
		for i := 0; i < 100; i++ {
			ctx.GetStats()
		}
		done <- true
	}()

	// 并发Clone
	go func() {
		for i := 0; i < 100; i++ {
			ctx.Clone()
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}
