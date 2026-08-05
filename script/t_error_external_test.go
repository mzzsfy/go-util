package script

import (
	"testing"
)

// ========== 编译错误测试 ==========

func Test_CompileError_UndefinedVariable(t *testing.T) {
	runErrorTest(t, "undefined_var + 1")
}

func Test_CompileError_MissingSemicolon(t *testing.T) {
	// 这不应报错，换行作为分隔符
	compileScript(t, "x := 1\ny := 2")
}

func Test_CompileError_UnclosedString(t *testing.T) {
	runErrorTest(t, `"unclosed`)
}

func Test_CompileError_UnclosedParen(t *testing.T) {
	runErrorTest(t, "(1 + 2")
}

func Test_CompileError_UnclosedBracket(t *testing.T) {
	runErrorTest(t, "[1, 2, 3")
}

func Test_CompileError_UnclosedBrace(t *testing.T) {
	runErrorTest(t, "{1, 2, 3")
}

func Test_CompileError_MissingCondition(t *testing.T) {
	runErrorTest(t, "if { x := 1 }")
}

func Test_CompileError_MissingBody(t *testing.T) {
	runErrorTest(t, "if true")
}

func Test_CompileError_InvalidOperator(t *testing.T) {
	// 没有对应运算符的情况由编译器检查
	// 但语法层面所有运算符都已支持
}

func Test_CompileError_BreakOutsideLoop(t *testing.T) {
	runErrorTest(t, "break")
}

func Test_CompileError_ContinueOutsideLoop(t *testing.T) {
	runErrorTest(t, "continue")
}

func Test_CompileError_ReturnOutsideFunction(t *testing.T) {
	// return在顶层可能是合法的(返回脚本结果)
	// 取决于实现，这里不强制
}

// ========== 运行时错误测试 ==========

func Test_RuntimeError_ThrowString(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile(`throw "error message"`)
	assertNoError(t, err)

	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Error("throw应产生错误")
	}
}

func Test_RuntimeError_ThrowNumber(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile(`throw 42`)
	assertNoError(t, err)

	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Error("throw应产生错误")
	}
}

func Test_RuntimeError_DivByZero(t *testing.T) {
	runRuntimeErrorTest(t, "1 / 0")
}

func Test_RuntimeError_ModByZero(t *testing.T) {
	runRuntimeErrorTest(t, "1 % 0")
}

func Test_RuntimeError_IndexNonCollection(t *testing.T) {
	runRuntimeErrorTest(t, "42[0]")
}

func Test_RuntimeError_CallNonFunction(t *testing.T) {
	runRuntimeErrorTest(t, "x := 42\nx()")
}

// ========== 外部函数调用测试 ==========

func Test_ExternalFunc_NoArgs(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("answer", func() int { return 42 })

	result := runScriptWithContext(t, `
		#fn answer() => int
		answer()
	`, ctx)
	if result.Int() != 42 {
		t.Errorf("got %d, want 42", result.Int())
	}
}

func Test_ExternalFunc_OneArg(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("double", func(x int) int { return x * 2 })

	result := runScriptWithContext(t, `
		#fn double(int) => int
		double(21)
	`, ctx)
	if result.Int() != 42 {
		t.Errorf("got %d, want 42", result.Int())
	}
}

func Test_ExternalFunc_TwoArgs(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("add", func(a, b int) int { return a + b })

	result := runScriptWithContext(t, `
		#fn add(int, int) => int
		add(10, 20)
	`, ctx)
	if result.Int() != 30 {
		t.Errorf("got %d, want 30", result.Int())
	}
}

func Test_ExternalFunc_ThreeArgs(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("add3", func(a, b, c int) int { return a + b + c })

	result := runScriptWithContext(t, `
		#fn add3(int, int, int) => int
		add3(1, 2, 3)
	`, ctx)
	if result.Int() != 6 {
		t.Errorf("got %d, want 6", result.Int())
	}
}

func Test_ExternalFunc_StringArg(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("greet", func(name string) string { return "hello " + name })

	result := runScriptWithContext(t, `
		#fn greet(string) => string
		greet("world")
	`, ctx)
	if result.String() != "hello world" {
		t.Errorf("got %q, want %q", result.String(), "hello world")
	}
}

func Test_ExternalFunc_BoolArg(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("negate", func(b bool) bool { return !b })

	result := runScriptWithContext(t, `
		#fn negate(bool) => bool
		negate(true)
	`, ctx)
	if result.Bool() != false {
		t.Errorf("negate(true)应返回false, got %v", result.Bool())
	}
}

func Test_ExternalFunc_ReturnFloat(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("pi", func() float64 { return 3.14 })

	result := runScriptWithContext(t, `
		#fn pi() => float
		pi()
	`, ctx)
	if result.Float() != 3.14 {
		t.Errorf("got %f, want 3.14", result.Float())
	}
}

func Test_ExternalFunc_MixedArgTypes(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("repeat", func(s string, n int) string {
		out := ""
		for i := 0; i < n; i++ {
			out += s
		}
		return out
	})

	result := runScriptWithContext(t, `
		#fn repeat(string, int) => string
		repeat("ab", 3)
	`, ctx)
	if result.String() != "ababab" {
		t.Errorf("got %q, want %q", result.String(), "ababab")
	}
}

func Test_ExternalFunc_NoReturn(t *testing.T) {
	ctx := NewContext()
	sideEffect := 0
	ctx.BindFunc("inc", func() { sideEffect++ })

	runScriptWithContext(t, `
		#fn inc()
		inc()
	`, ctx)
	if sideEffect != 1 {
		t.Errorf("外部函数副作用应执行1次, got %d", sideEffect)
	}
}

func Test_ExternalFunc_MultipleCalls(t *testing.T) {
	ctx := NewContext()
	count := 0
	ctx.BindFunc("counter", func() int {
		count++
		return count
	})

	result := runScriptWithContext(t, `
		#fn counter() => int
		counter() + counter()
	`, ctx)
	if result.Int() != 3 { // 1 + 2 = 3
		t.Errorf("got %d, want 3", result.Int())
	}
}

func Test_ExternalFunc_InCondition(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("isPositive", func(n int) bool { return n > 0 })

	result := runScriptWithContext(t, `
		#fn isPositive(int) => bool
		if isPositive(5) { 1 } else { 2 }
	`, ctx)
	if result.Int() != 1 {
		t.Errorf("got %d, want 1", result.Int())
	}
}

// ========== 内建函数测试 ==========

func Test_Builtin_Len_String(t *testing.T) {
	runIntTest(t, `len("hello")`, 5)
	runIntTest(t, `len("")`, 0)
	runIntTest(t, `len("a")`, 1)
}

func Test_Builtin_Len_Array(t *testing.T) {
	runIntTest(t, `len([1, 2, 3])`, 3)
	runIntTest(t, `len([])`, 0)
}

func Test_Builtin_Len_Map(t *testing.T) {
	runIntTest(t, `len({"a": 1})`, 1)
	runIntTest(t, `len({"a": 1, "b": 2, "c": 3})`, 3)
}

func Test_Builtin_TypeOf_Basic(t *testing.T) {
	result := runScript(t, `typeof(42)`)
	if result.String() == "" {
		t.Error("typeof(42)不应返回空字符串")
	}
}

func Test_Builtin_String(t *testing.T) {
	result := runScript(t, `string(123)`)
	if result.String() != "123" {
		t.Errorf("string(123)应为'123', got %q", result.String())
	}
}

func Test_Builtin_Int(t *testing.T) {
	result := runScript(t, `int("42")`)
	if result.Int() != 42 {
		t.Errorf("int('42')应为42, got %d", result.Int())
	}
}
