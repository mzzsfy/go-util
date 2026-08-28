package script

import (
	"testing"
)

// ========== Context 完整测试 ==========

func Test_context_SetOutput(t *testing.T) {
	ctx := NewContext()
	var buf []byte
	writer := &testWriter{buf: &buf}
	ctx.SetOutput(writer)

	if ctx.GetOutput() != writer {
		t.Errorf("SetOutput失败")
	}
}

type testWriter struct {
	buf *[]byte
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func Test_context_Log(t *testing.T) {
	ctx := NewContext()
	logCalled := false
	ctx.SetLogFunc(func(format string, args ...interface{}) {
		logCalled = true
	})

	ctx.Log("test")
	if !logCalled {
		t.Errorf("Log函数未被调用")
	}
}

// ========== Engine 完整测试 ==========

func Test_engine_Stop(t *testing.T) {
	engine := NewEngine()
	engine.Stop()
	// Stop方法目前是空实现，只测试不会panic
}

// TestEngine_OptionsComplete 测试引擎选项完整场景
func Test_engine_OptionsComplete(t *testing.T) {
	engine := NewEngine(
		WithMaxCallDepth(500),
	)

	if engine == nil {
		t.Error("NewEngine不应返回nil")
	}
}

// TestEngine_Run_MultipleTimes 测试多次执行
func Test_engine_Run_MultipleTimes(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile("1 + 1")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine()

	// 执行多次
	for i := 0; i < 3; i++ {
		_, err = engine.Run(ctx, script)
		if err != nil {
			t.Errorf("第%d次执行失败: %v", i+1, err)
		}
	}

	stats := ctx.GetStats()
	if stats.TotalRuns != 3 {
		t.Errorf("TotalRuns应为3, 实际为%d", stats.TotalRuns)
	}
}

// ========== Value 完整测试 ==========

func Test_value_NewValueTypes(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected ValueType
	}{
		{10, TypeInt},
		{10.5, TypeFloat},
		{"hello", TypeString},
		{true, TypeBool},
		{nil, TypeNil},
	}

	for _, tt := range tests {
		result := NewValue(tt.input)
		if result.Type != tt.expected {
			t.Errorf("NewValue(%v) 类型错误：期望 %v，得到 %v", tt.input, tt.expected, result.Type)
		}
	}
}

func Test_value_ArrayMethods(t *testing.T) {
	arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})

	if arr.Type != TypeArray {
		t.Errorf("类型应为TypeArray")
	}

	array := arr.Array()
	if array == nil {
		t.Errorf("Array()不应返回nil")
		return
	}

	if len(array.Elements) != 3 {
		t.Errorf("数组长度应为3")
	}
}

// TestValue_PrintFormatting 测试值打印格式
func Test_value_PrintFormatting(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		nonEmpty bool
	}{
		{"nil值", NewValue(nil), true},
		{"整数值", NewValue(42), true},
		{"浮点值", NewValue(3.14), true},
		{"字符串值", NewValue("hello"), true},
		{"布尔值true", NewValue(true), true},
		{"布尔值false", NewValue(false), true},
		{"数组值", NewValue([]Value{NewValue(1), NewValue(2)}), true},
		{"Map值", NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用GoString方法测试打印格式
			str := tt.value.GoString()
			if tt.nonEmpty && str == "" {
				t.Errorf("%s: GoString表示不应为空", tt.name)
			}
			if tt.nonEmpty && str == "<unknown>" {
				t.Errorf("%s: GoString表示不应为<unknown>", tt.name)
			}
		})
	}
}

func Test_value_MapMethods(t *testing.T) {
	m := &MapValue{
		Pairs: map[string]Value{
			"a": NewValue(1),
		},
	}
	val := NewValue(m)

	if val.Type != TypeMap {
		t.Errorf("类型应为TypeMap")
	}

	mp := val.Map()
	if mp == nil {
		t.Errorf("Map()不应返回nil")
		return
	}

	if len(mp.Pairs) != 1 {
		t.Errorf("Map大小应为1")
	}
}

// ========== Frame测试 ==========

// TestNewFrame 测试创建新帧
func Test_newframe(t *testing.T) {
	fn := &CompiledFunction{
		Name:         "testFunc",
		Instructions: []Instruction{},
		NumLocals:    5,
		NumParams:    2,
	}

	frame := NewFrame(fn)

	if frame.Function.Name != "testFunc" {
		t.Errorf("Frame函数名不正确: %s", frame.Function.Name)
	}

	if len(frame.Locals) != 5 {
		t.Errorf("Locals长度应为5, 实际为%d", len(frame.Locals))
	}

	if frame.IP != 0 {
		t.Errorf("初始IP应为0, 实际为%d", frame.IP)
	}
}

// ========== 编译错误测试 ==========

// TestCompileError_NoLine 测试无行号错误
func Test_compileerror_NoLine(t *testing.T) {
	err := &CompileError{Line: 0, Column: 0, Message: "test error"}
	result := err.Error()
	if result == "" {
		t.Error("Error()不应返回空字符串")
	}
}

// TestRuntimeError_EmptyStack 测试空调用栈
func Test_runtimeerror_EmptyStack(t *testing.T) {
	err := &RuntimeError{
		Message:    "runtime error",
		StackTrace: []string{},
	}
	result := err.Error()
	if result == "" {
		t.Error("Error()不应返回空字符串")
	}
}

// ========== Value边界测试 ==========

// TestValue_ExternalFunc 测试外部函数值
func Test_value_ExternalFunc(t *testing.T) {
	ef := &ExternalFuncValue{
		Name: "extFunc",
		Func: func(x int) int { return x },
	}
	val := Value{Type: TypeExternalFunc, Data: ef}

	result := val.ExternalFunc()
	if result == nil {
		t.Error("ExternalFunc()不应返回nil")
	}

	if result.Name != "extFunc" {
		t.Errorf("外部函数名不正确: %s", result.Name)
	}
}

// TestValue_Function 测试函数值
func Test_value_Function(t *testing.T) {
	fn := &FunctionValue{
		Compiled: &CompiledFunction{Name: "myFunc"},
	}
	val := Value{Type: TypeFunction, Data: fn}

	result := val.Function()
	if result == nil {
		t.Error("Function()不应返回nil")
	}
}

// ========== 统计信息测试 ==========

// TestExecStats 测试执行统计
func Test_execstats(t *testing.T) {
	ctx := NewContext()

	// 执行一些操作以更新统计
	ctx.updateStats(100, 5)
	ctx.updateStats(200, 3)
	ctx.updateStats(300, 10)

	stats := ctx.GetStats()

	if stats.TotalRuns != 3 {
		t.Errorf("TotalRuns应为3, 实际为%d", stats.TotalRuns)
	}

	if stats.MaxCallDepth != 10 {
		t.Errorf("MaxCallDepth应为10, 实际为%d", stats.MaxCallDepth)
	}
}

// ========== 特殊值处理器测试 ==========

func Test_vmspecialvalues(t *testing.T) {
	t.Run("nil处理器", func(t *testing.T) {
		result := runScript(t, "nil")
		if !result.IsNil() {
			t.Errorf("nil处理器测试失败")
		}
	})

	t.Run("true处理器", func(t *testing.T) {
		result := runScript(t, "true")
		if !result.Bool() {
			t.Errorf("true处理器测试失败")
		}
	})

	t.Run("false处理器", func(t *testing.T) {
		result := runScript(t, "false")
		if result.Bool() {
			t.Errorf("false处理器测试失败")
		}
	})
}
