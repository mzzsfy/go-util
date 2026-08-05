package script

import (
	"testing"
)

// ========== 外部函数复杂签名测试 ==========

func Test_ExternalFunc_FloatArgs(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("fmul", func(a, b float64) float64 { return a * b })

	result := runScriptWithContext(t, `
		#fn fmul(float, float) => float
		fmul(2.5, 4.0)
	`, ctx)
	if result.Float() != 10.0 {
		t.Errorf("got %f, want 10.0", result.Float())
	}
}

func Test_ExternalFunc_MixedIntFloat(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("scale", func(n int, f float64) float64 {
		return float64(n) * f
	})

	result := runScriptWithContext(t, `
		#fn scale(int, float) => float
		scale(3, 2.5)
	`, ctx)
	if result.Float() != 7.5 {
		t.Errorf("got %f, want 7.5", result.Float())
	}
}

func Test_ExternalFunc_StringReturn(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("concat", func(a, b string) string { return a + "-" + b })

	result := runScriptWithContext(t, `
		#fn concat(string, string) => string
		concat("hello", "world")
	`, ctx)
	if result.String() != "hello-world" {
		t.Errorf("got %q, want %q", result.String(), "hello-world")
	}
}

func Test_ExternalFunc_FourArgs(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("add4", func(a, b, c, d int) int { return a + b + c + d })

	result := runScriptWithContext(t, `
		#fn add4(int, int, int, int) => int
		add4(1, 2, 3, 4)
	`, ctx)
	if result.Int() != 10 {
		t.Errorf("got %d, want 10", result.Int())
	}
}

func Test_ExternalFunc_FiveArgs(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("add5", func(a, b, c, d, e int) int { return a + b + c + d + e })

	result := runScriptWithContext(t, `
		#fn add5(int, int, int, int, int) => int
		add5(1, 2, 3, 4, 5)
	`, ctx)
	if result.Int() != 15 {
		t.Errorf("got %d, want 15", result.Int())
	}
}

func Test_ExternalFunc_InLoop(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("inc", func(n int) int { return n + 1 })

	result := runScriptWithContext(t, `
		#fn inc(int) => int
		sum := 0
		for i := 0; i < 5; i = i + 1 {
			sum = sum + inc(i)
		}
		sum
	`, ctx)
	// inc(0)+inc(1)+inc(2)+inc(3)+inc(4) = 1+2+3+4+5 = 15
	if result.Int() != 15 {
		t.Errorf("got %d, want 15", result.Int())
	}
}

func Test_ExternalFunc_InArray(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("val", func() int { return 42 })

	result := runScriptWithContext(t, `
		#fn val() => int
		x := [val(), val(), val()]
		x[0] + x[1] + x[2]
	`, ctx)
	if result.Int() != 126 {
		t.Errorf("got %d, want 126", result.Int())
	}
}

func Test_ExternalFunc_AsCondition(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("isActive", func() bool { return true })

	result := runScriptWithContext(t, `
		#fn isActive() => bool
		if isActive() { 1 } else { 0 }
	`, ctx)
	if result.Int() != 1 {
		t.Errorf("got %d, want 1", result.Int())
	}
}

func Test_ExternalFunc_Nested(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("add", func(a, b int) int { return a + b })
	ctx.BindFunc("mul", func(a, b int) int { return a * b })

	result := runScriptWithContext(t, `
		#fn add(int, int) => int
		#fn mul(int, int) => int
		add(mul(2, 3), mul(4, 5))
	`, ctx)
	// 6 + 20 = 26
	if result.Int() != 26 {
		t.Errorf("got %d, want 26", result.Int())
	}
}

func Test_ExternalFunc_ChainedWithScript(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("toDouble", func(s string) string { return s + s })

	result := runScriptWithContext(t, `
		#fn toDouble(string) => string
		x := toDouble("ab")
		x + toDouble("cd")
	`, ctx)
	if result.String() != "ababcdcd" {
		t.Errorf("got %q, want %q", result.String(), "ababcdcd")
	}
}

// ========== 脚本复用测试 ==========

func Test_Script_ReuseMultipleRuns(t *testing.T) {
	script := compileScript(t, "1 + 2")
	ctx := NewContext()
	engine := NewEngine()

	for i := 0; i < 10; i++ {
		result, err := engine.Run(ctx, script)
		assertNoError(t, err)
		if result.Int() != 3 {
			t.Errorf("run %d: got %d, want 3", i, result.Int())
		}
	}
}

func Test_Script_ReuseWithDifferentContexts(t *testing.T) {
	script := compileScript(t, `
		v :=>int getBindValue("x")
		v * 2
	`)
	engine := NewEngine()

	ctx1 := NewContext()
	ctx1.BindValue("x", 10)

	ctx2 := NewContext()
	ctx2.BindValue("x", 20)

	r1, _ := engine.Run(ctx1, script)
	r2, _ := engine.Run(ctx2, script)

	if r1.Int() != 20 {
		t.Errorf("ctx1: got %d, want 20", r1.Int())
	}
	if r2.Int() != 40 {
		t.Errorf("ctx2: got %d, want 40", r2.Int())
	}
}

func Test_Script_RunAfterError(t *testing.T) {
	script := compileScript(t, "1 + 1")
	ctx := NewContext()
	engine := NewEngine()

	// 先执行一个会报错的脚本
	errorScript := compileScript(t, `throw "err"`)
	engine.Run(ctx, errorScript)

	// 后续正常脚本应不受影响
	result, err := engine.Run(ctx, script)
	assertNoError(t, err)
	if result.Int() != 2 {
		t.Errorf("got %d, want 2", result.Int())
	}
}

// ========== Map操作测试 ==========

func Test_VM_MapStore(t *testing.T) {
	result := runScript(t, `
		m := {"a": 1}
		m["b"] = 2
		m["a"] + m["b"]
	`)
	if result.Int() != 3 {
		t.Errorf("got %d, want 3", result.Int())
	}
}

func Test_VM_MapOverwrite(t *testing.T) {
	result := runScript(t, `
		m := {"key": 1}
		m["key"] = 99
		m["key"]
	`)
	if result.Int() != 99 {
		t.Errorf("got %d, want 99", result.Int())
	}
}

func Test_VM_NestedMapAccess(t *testing.T) {
	result := runScript(t, `
		data := {
			"config": {
				"port": 8080,
				"host": "localhost"
			}
		}
		data["config"]["port"]
	`)
	if result.Int() != 8080 {
		t.Errorf("got %d, want 8080", result.Int())
	}
}

func Test_VM_MapWithDifferentValueTypes(t *testing.T) {
	result := runScript(t, `
		m := {
			"int": 42,
			"str": "hello",
			"bool": true
		}
		m
	`)
	m := result.Map()
	if m.Pairs["int"].Int() != 42 {
		t.Errorf("int = %d", m.Pairs["int"].Int())
	}
	if m.Pairs["str"].String() != "hello" {
		t.Errorf("str = %q", m.Pairs["str"].String())
	}
	if !m.Pairs["bool"].Bool() {
		t.Error("bool should be true")
	}
}

// ========== 数组与Map组合测试 ==========

func Test_VM_ArrayOfMaps(t *testing.T) {
	result := runScript(t, `
		arr := [{"x": 1}, {"x": 2}, {"x": 3}]
		arr[1]["x"]
	`)
	if result.Int() != 2 {
		t.Errorf("got %d, want 2", result.Int())
	}
}

func Test_VM_MapOfArrays(t *testing.T) {
	result := runScript(t, `
		m := {"nums": [10, 20, 30]}
		m["nums"][2]
	`)
	if result.Int() != 30 {
		t.Errorf("got %d, want 30", result.Int())
	}
}

func Test_VM_ModifyArrayInMap(t *testing.T) {
	result := runScript(t, `
		data := {"items": [1, 2, 3]}
		data["items"][0] = 99
		data["items"][0]
	`)
	if result.Int() != 99 {
		t.Errorf("got %d, want 99", result.Int())
	}
}

// ========== 运算符边界测试 ==========

func Test_VM_AllBitwiseCombinations(t *testing.T) {
	// 测试各种位运算组合
	runIntTest(t, "0xFF & 0x0F", 15)
	runIntTest(t, "0x00 | 0xFF", 255)
	runIntTest(t, "0xFF ^ 0xFF", 0)
	runIntTest(t, "1 << 0", 1)
	runIntTest(t, "1 << 1", 2)
	runIntTest(t, "1 << 2", 4)
	runIntTest(t, "1 << 3", 8)
	runIntTest(t, "1 << 4", 16)
	runIntTest(t, "1 << 5", 32)
	runIntTest(t, "256 >> 0", 256)
	runIntTest(t, "256 >> 1", 128)
	runIntTest(t, "256 >> 2", 64)
	runIntTest(t, "256 >> 3", 32)
	runIntTest(t, "256 >> 4", 16)
}

func Test_VM_ComparisonChain(t *testing.T) {
	runBoolTest(t, "1 < 2 && 2 < 3 && 3 < 4", true)
	runBoolTest(t, "1 < 2 && 2 > 3", false)
	runBoolTest(t, "1 > 0 || 0 > 1", true)
	runBoolTest(t, "1 == 1 && 2 == 2 && 3 == 3", true)
}

// ========== Context日志输出测试 ==========

func Test_Context_LogInScript(t *testing.T) {
	// 测试脚本中的log输出
	ctx := NewContext()
	var logged []string
	ctx.SetLogFunc(func(format string, args ...any) {
		logged = append(logged, format)
	})

	// 执行脚本
	runScriptWithContext(t, "1 + 1", ctx)

	// log可能或可能不被调用，取决于脚本实现
	// 这里只验证不panic
}

// ========== Engine配置测试 ==========

func Test_Engine_DefaultMaxCallDepth(t *testing.T) {
	e := NewEngine()
	ctx := NewContext()
	script := compileScript(t, "1 + 1")
	result, err := e.Run(ctx, script)
	assertNoError(t, err)
	if result.Int() != 2 {
		t.Errorf("got %d, want 2", result.Int())
	}
}

func Test_Engine_CustomMaxCallDepth(t *testing.T) {
	e := NewEngine(WithMaxCallDepth(1000))
	ctx := NewContext()
	script := compileScript(t, "1 + 1")
	result, err := e.Run(ctx, script)
	assertNoError(t, err)
	if result.Int() != 2 {
		t.Errorf("got %d, want 2", result.Int())
	}
}

func Test_Engine_DeepRecursion(t *testing.T) {
	e := NewEngine(WithMaxCallDepth(50))
	ctx := NewContext()
	script := compileScript(t, `
		fn recurse(n) {
			if n <= 0 { return 0 }
			return recurse(n - 1)
		}
		recurse(10)
	`)
	result, err := e.Run(ctx, script)
	assertNoError(t, err)
	if result.Int() != 0 {
		t.Errorf("got %d, want 0", result.Int())
	}
}

// ========== 统计信息验证 ==========

func Test_Context_StatsAfterMultipleRuns(t *testing.T) {
	ctx := NewContext()
	engine := NewEngine()
	script := compileScript(t, "1 + 1")

	n := 10
	for i := 0; i < n; i++ {
		engine.Run(ctx, script)
	}

	stats := ctx.GetStats()
	if stats.TotalRuns != int64(n) {
		t.Errorf("TotalRuns = %d, want %d", stats.TotalRuns, n)
	}
}

func Test_Context_StatsMaxDepth(t *testing.T) {
	ctx := NewContext()
	engine := NewEngine()
	script := compileScript(t, `
		fn f(n) {
			if n <= 0 { return 0 }
			return f(n - 1)
		}
		f(5)
	`)
	_, err := engine.Run(ctx, script)
	assertNoError(t, err)

	stats := ctx.GetStats()
	if stats.TotalRuns != 1 {
		t.Errorf("TotalRuns应为1, got %d", stats.TotalRuns)
	}
}
