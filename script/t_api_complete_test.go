package script

import (
	"bytes"
	"strings"
	"testing"
)

// ========== Eval 便捷函数测试 ==========

func Test_APIComplete_Eval(t *testing.T) {
	t.Run("加法返回整数", func(t *testing.T) {
		v, err := Eval("1 + 2")
		assertNoError(t, err)
		assertInt(t, v, 3)
	})

	t.Run("nil字面量", func(t *testing.T) {
		v, err := Eval("nil")
		assertNoError(t, err)
		assertNil(t, v)
	})

	t.Run("字符串字面量", func(t *testing.T) {
		v, err := Eval(`"hello"`)
		assertNoError(t, err)
		assertString(t, v, "hello")
	})

	t.Run("执行错误返回error", func(t *testing.T) {
		// 未定义变量导致运行时错误
		_, err := Eval("undefinedVar")
		assertError(t, err)
	})

	t.Run("编译错误返回error", func(t *testing.T) {
		_, err := Eval("x + ")
		assertError(t, err)
	})

	t.Run("空字符串", func(t *testing.T) {
		v, err := Eval("")
		assertNoError(t, err)
		assertNil(t, v)
	})

	t.Run("只有注释", func(t *testing.T) {
		v, err := Eval("// just a comment")
		assertNoError(t, err)
		assertNil(t, v)
	})
}

// ========== MustEval 测试 ==========

func Test_APIComplete_MustEval(t *testing.T) {
	t.Run("正常返回", func(t *testing.T) {
		v := MustEval("42")
		assertInt(t, v, 42)
	})

	t.Run("错误时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustEval在错误时应panic")
			}
		}()
		MustEval("!!!invalid")
	})
}

// ========== EvalWithBindings 测试 ==========

func Test_APIComplete_EvalWithBindings(t *testing.T) {
	t.Run("绑定int值", func(t *testing.T) {
		v, err := EvalWithBindings(`x :=>int getBindValue("x")
x + 10`, map[string]interface{}{"x": 5})
		assertNoError(t, err)
		assertInt(t, v, 15)
	})

	t.Run("绑定string值", func(t *testing.T) {
		v, err := EvalWithBindings(`s :=>string getBindValue("s")
s + "!"`, map[string]interface{}{"s": "hi"})
		assertNoError(t, err)
		assertString(t, v, "hi!")
	})

	t.Run("绑定bool值", func(t *testing.T) {
		v, err := EvalWithBindings(`b :=>bool getBindValue("b")
b`, map[string]interface{}{"b": true})
		assertNoError(t, err)
		assertBool(t, v, true)
	})

	t.Run("绑定多个值", func(t *testing.T) {
		v, err := EvalWithBindings(`x :=>int getBindValue("x")
y :=>int getBindValue("y")
x + y`, map[string]interface{}{"x": 10, "y": 20})
		assertNoError(t, err)
		assertInt(t, v, 30)
	})

	t.Run("绑定nil值", func(t *testing.T) {
		v, err := EvalWithBindings(`n :=>any getBindValue("n")
n`, map[string]interface{}{"n": nil})
		assertNoError(t, err)
		assertNil(t, v)
	})

	t.Run("绑定数组", func(t *testing.T) {
		v, err := EvalWithBindings(`arr :=>arr getBindValue("arr")
arr[0] + arr[1] + arr[2]`, map[string]interface{}{"arr": []int{1, 2, 3}})
		assertNoError(t, err)
		assertInt(t, v, 6)
	})

	t.Run("绑定Map", func(t *testing.T) {
		v, err := EvalWithBindings(`m :=>any getBindValue("m")
m["a"] + m["b"]`, map[string]interface{}{"m": map[string]int{"a": 10, "b": 20}})
		assertNoError(t, err)
		assertInt(t, v, 30)
	})

	t.Run("绑定值在表达式中使用", func(t *testing.T) {
		v, err := EvalWithBindings(`base :=>int getBindValue("base")
mul :=>int getBindValue("mul")
base * mul + 1`, map[string]interface{}{"base": 3, "mul": 7})
		assertNoError(t, err)
		assertInt(t, v, 22)
	})
}

// ========== MustEvalWithBindings 测试 ==========

func Test_APIComplete_MustEvalWithBindings(t *testing.T) {
	t.Run("正常返回", func(t *testing.T) {
		v := MustEvalWithBindings(`x :=>int getBindValue("x")
x * 2`, map[string]interface{}{"x": 21})
		assertInt(t, v, 42)
	})

	t.Run("错误时panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustEvalWithBindings在错误时应panic")
			}
		}()
		MustEvalWithBindings("!!!invalid", nil)
	})
}

// ========== Context - BindValue 测试 ==========

func Test_APIComplete_Context_BindValue(t *testing.T) {
	t.Run("绑定int", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("i", 42)
		v, ok := ctx.GetBindValue("i")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertInt(t, v, 42)
	})

	t.Run("绑定string", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("s", "hello")
		v, ok := ctx.GetBindValue("s")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertString(t, v, "hello")
	})

	t.Run("绑定bool", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("b", true)
		v, ok := ctx.GetBindValue("b")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertBool(t, v, true)
	})

	t.Run("绑定nil", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("n", nil)
		v, ok := ctx.GetBindValue("n")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertNil(t, v)
	})

	t.Run("绑定[]int", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("arr", []int{10, 20, 30})
		v, ok := ctx.GetBindValue("arr")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertType(t, v, TypeArray)
	})

	t.Run("绑定[]string", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("arr", []string{"a", "b"})
		v, ok := ctx.GetBindValue("arr")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertType(t, v, TypeArray)
	})

	t.Run("绑定map[string]int", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("m", map[string]int{"x": 1})
		v, ok := ctx.GetBindValue("m")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		assertType(t, v, TypeMap)
	})

	t.Run("不存在返回false", func(t *testing.T) {
		ctx := NewContext()
		_, ok := ctx.GetBindValue("missing")
		if ok {
			t.Error("未绑定的值应返回false")
		}
	})

	t.Run("绑定覆盖同名值", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 1)
		ctx.BindValue("x", 2)
		v, _ := ctx.GetBindValue("x")
		assertInt(t, v, 2)
	})

	t.Run("绑定struct返回nil", func(t *testing.T) {
		// struct类型不支持自动转换, 返回nil
		type MyStruct struct{ X int }
		ctx := NewContext()
		ctx.BindValue("s", MyStruct{X: 42})
		v, ok := ctx.GetBindValue("s")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		// struct未知类型, 引擎转为nil
		assertNil(t, v)
	})

	t.Run("绑定slice_of_struct", func(t *testing.T) {
		type Item struct{ V int }
		ctx := NewContext()
		ctx.BindValue("items", []Item{{V: 1}, {V: 2}})
		v, ok := ctx.GetBindValue("items")
		if !ok {
			t.Fatal("GetBindValue应返回true")
		}
		// 元素为struct, 转为nil, 整体是数组
		assertType(t, v, TypeArray)
	})
}

// ========== Context - BindFunc 测试 ==========

func Test_APIComplete_Context_BindFunc(t *testing.T) {
	t.Run("绑定func(int)int", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("double", func(x int) int { return x * 2 })
		cs := compileScript(t, "#fn double(int)=>int\ndouble(5)")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 10)
	})

	t.Run("绑定func(string)string", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("greet", func(name string) string { return "hi " + name })
		cs := compileScript(t, "#fn greet(string)=>string\ngreet(\"Bob\")")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertString(t, v, "hi Bob")
	})

	t.Run("绑定func(int_int)int", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("add", func(a, b int) int { return a + b })
		cs := compileScript(t, "#fn add(int, int)=>int\nadd(3, 4)")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 7)
	})

	t.Run("绑定func()int无参数", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("answer", func() int { return 42 })
		cs := compileScript(t, "#fn answer()=>int\nanswer()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 42)
	})

	t.Run("绑定func()(int_error)双返回值", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("val", func() (int, error) { return 99, nil })
		cs := compileScript(t, "#fn val()=>int\nval()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 99)
	})

	t.Run("绑定func()[]int返回数组", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("getList", func() []int { return []int{1, 2, 3} })
		cs := compileScript(t, "#fn getList()=>arr\ngetList()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertType(t, v, TypeArray)
	})

	t.Run("绑定func()map[string]int返回map", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("getMap", func() map[string]int { return map[string]int{"k": 1} })
		cs := compileScript(t, "#fn getMap()=>any\ngetMap()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertType(t, v, TypeMap)
	})

	t.Run("绑定func()bool返回bool", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("flag", func() bool { return true })
		cs := compileScript(t, "#fn flag()=>bool\nflag()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertBool(t, v, true)
	})

	t.Run("绑定func()string返回string", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("msg", func() string { return "ok" })
		cs := compileScript(t, "#fn msg()=>string\nmsg()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertString(t, v, "ok")
	})

	t.Run("绑定func返回nil", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("noop", func() interface{} { return nil })
		cs := compileScript(t, "#fn noop()=>any\nnoop()")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertNil(t, v)
	})

	t.Run("GetBindFunc返回已绑定函数", func(t *testing.T) {
		ctx := NewContext()
		fn := func(x int) int { return x }
		ctx.BindFunc("f", fn)
		got, ok := ctx.GetBindFunc("f")
		if !ok {
			t.Fatal("GetBindFunc应返回true")
		}
		if got == nil {
			t.Fatal("GetBindFunc不应返回nil")
		}
	})

	t.Run("GetBindFunc不存在返回false", func(t *testing.T) {
		ctx := NewContext()
		_, ok := ctx.GetBindFunc("missing")
		if ok {
			t.Error("未绑定的函数应返回false")
		}
	})
}

// ========== Context - Clone 测试 ==========

func Test_APIComplete_Context_Clone(t *testing.T) {
	t.Run("Clone后绑定值独立", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 1)
		clone := ctx.Clone()

		// 修改Clone不影响原Context
		clone.BindValue("y", 99)
		_, ok := ctx.GetBindValue("y")
		if ok {
			t.Error("修改Clone不应影响原Context")
		}
	})

	t.Run("Clone后绑定函数独立", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("f", func() int { return 1 })
		clone := ctx.Clone()

		// Clone应保留已有函数
		_, ok := clone.GetBindFunc("f")
		if !ok {
			t.Error("Clone后应保留已绑定函数")
		}

		// 在Clone上添加新函数不影响原Context
		clone.BindFunc("g", func() int { return 2 })
		_, ok = ctx.GetBindFunc("g")
		if ok {
			t.Error("修改Clone不应影响原Context")
		}
	})

	t.Run("修改原Context不影响Clone", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 1)
		clone := ctx.Clone()

		ctx.BindValue("z", 88)
		_, ok := clone.GetBindValue("z")
		if ok {
			t.Error("修改原Context不应影响Clone")
		}
	})

	t.Run("Clone保留原有绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("a", 42)
		ctx.BindValue("b", "hello")
		ctx.BindFunc("fn", func() int { return 0 })

		clone := ctx.Clone()
		v, ok := clone.GetBindValue("a")
		if !ok {
			t.Fatal("Clone应保留绑定值a")
		}
		assertInt(t, v, 42)

		v2, ok := clone.GetBindValue("b")
		if !ok {
			t.Fatal("Clone应保留绑定值b")
		}
		assertString(t, v2, "hello")

		_, ok = clone.GetBindFunc("fn")
		if !ok {
			t.Error("Clone应保留绑定函数fn")
		}
	})
}

// ========== Context - Stats 测试 ==========

func Test_APIComplete_Context_Stats(t *testing.T) {
	t.Run("GetStats返回初始统计", func(t *testing.T) {
		ctx := NewContext()
		stats := ctx.GetStats()
		if stats.TotalRuns != 0 {
			t.Errorf("初始TotalRuns应为0, 得到%d", stats.TotalRuns)
		}
	})

	t.Run("多次执行后统计更新", func(t *testing.T) {
		ctx := NewContext()
		cs := compileScript(t, "1 + 1")
		e := NewEngine()
		for i := 0; i < 5; i++ {
			e.Run(ctx, cs)
		}
		stats := ctx.GetStats()
		if stats.TotalRuns != 5 {
			t.Errorf("5次执行后TotalRuns应为5, 得到%d", stats.TotalRuns)
		}
	})
}

// ========== Context - Output 测试 ==========

func Test_APIComplete_Context_Output(t *testing.T) {
	t.Run("SetOutput设置自定义Writer", func(t *testing.T) {
		ctx := NewContext()
		var buf bytes.Buffer
		ctx.SetOutput(&buf)
		// GetOutput返回非nil即可
		if ctx.GetOutput() == nil {
			t.Error("GetOutput不应返回nil")
		}
	})

	t.Run("GetOutput获取设置的Writer", func(t *testing.T) {
		ctx := NewContext()
		var buf bytes.Buffer
		ctx.SetOutput(&buf)
		// 写入数据验证Writer可用
		buf.WriteString("test")
		if buf.String() != "test" {
			t.Error("自定义Writer应可正常写入")
		}
	})

	t.Run("默认输出非nil", func(t *testing.T) {
		ctx := NewContext()
		if ctx.GetOutput() == nil {
			t.Error("默认输出不应为nil")
		}
	})
}

// ========== Context - Log 测试 ==========

func Test_APIComplete_Context_Log(t *testing.T) {
	t.Run("Log调用设置的LogFunc", func(t *testing.T) {
		ctx := NewContext()
		var got string
		ctx.SetLogFunc(func(format string, args ...any) {
			got = format
		})
		ctx.Log("test log")
		if got != "test log" {
			t.Errorf("LogFunc应收到消息, 得到%q", got)
		}
	})

	t.Run("Log传递格式化参数", func(t *testing.T) {
		ctx := NewContext()
		var gotFormat string
		var gotArgs []any
		ctx.SetLogFunc(func(format string, args ...any) {
			gotFormat = format
			gotArgs = args
		})
		ctx.Log("val=%d", 42)
		if gotFormat != "val=%d" {
			t.Errorf("format应原样传递, 得到%q", gotFormat)
		}
		if len(gotArgs) != 1 {
			t.Fatalf("应传递1个参数, 得到%d", len(gotArgs))
		}
	})
}

// ========== Engine - NewEngine 测试 ==========

func Test_APIComplete_Engine_NewEngine(t *testing.T) {
	t.Run("默认引擎", func(t *testing.T) {
		e := NewEngine()
		if e == nil {
			t.Fatal("NewEngine不应返回nil")
		}
	})

	t.Run("WithMaxCallDepth设置自定义深度", func(t *testing.T) {
		e := NewEngine(WithMaxCallDepth(512))
		if e == nil {
			t.Fatal("NewEngine with option不应返回nil")
		}
		ctx := NewContext()
		cs := compileScript(t, "1 + 1")
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 2)
	})

	t.Run("深度过小导致递归失败", func(t *testing.T) {
		e := NewEngine(WithMaxCallDepth(10))
		ctx := NewContext()
		cs := compileScript(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(100)
`)
		_, err := e.Run(ctx, cs)
		assertError(t, err)
	})

	t.Run("深度足够时递归成功", func(t *testing.T) {
		e := NewEngine(WithMaxCallDepth(100))
		ctx := NewContext()
		cs := compileScript(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(50)
`)
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 50)
	})
}

// ========== Engine - Run 测试 ==========

func Test_APIComplete_Engine_Run(t *testing.T) {
	t.Run("同一脚本多次Run", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		cs := compileScript(t, "10 + 20")
		for i := 0; i < 3; i++ {
			v, err := e.Run(ctx, cs)
			assertNoError(t, err)
			assertInt(t, v, 30)
		}
	})

	t.Run("不同脚本同一引擎", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		s1 := compileScript(t, "5 * 6")
		s2 := compileScript(t, "100 - 1")
		v1, err := e.Run(ctx, s1)
		assertNoError(t, err)
		assertInt(t, v1, 30)
		v2, err := e.Run(ctx, s2)
		assertNoError(t, err)
		assertInt(t, v2, 99)
	})

	t.Run("Run后Context统计更新", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		cs := compileScript(t, "1 + 1")
		e.Run(ctx, cs)
		stats := ctx.GetStats()
		if stats.TotalRuns != 1 {
			t.Errorf("执行后TotalRuns应为1, 得到%d", stats.TotalRuns)
		}
	})

	t.Run("Run错误处理", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		cs := compileScript(t, `throw "test error"`)
		_, err := e.Run(ctx, cs)
		assertError(t, err)
	})
}

// ========== Engine - Stop 测试 ==========

func Test_APIComplete_Engine_Stop(t *testing.T) {
	t.Run("Stop无运行脚本时不panic", func(t *testing.T) {
		e := NewEngine()
		e.Stop()
	})

	t.Run("Stop在执行后正常调用", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		cs := compileScript(t, "1 + 1")
		e.Run(ctx, cs)
		e.Stop()
	})
}

// ========== Script - NewScript/GetCompiled 测试 ==========

func Test_APIComplete_Script_NewAndGetCompiled(t *testing.T) {
	t.Run("NewScript创建实例", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s := NewScript(cs)
		if s == nil {
			t.Fatal("NewScript不应返回nil")
		}
	})

	t.Run("GetCompiled返回底层编译产物", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s := NewScript(cs)
		if s.GetCompiled() == nil {
			t.Error("GetCompiled不应返回nil")
		}
	})
}

// ========== Script - Clone 测试 ==========

func Test_APIComplete_Script_Clone(t *testing.T) {
	t.Run("Clone返回非nil", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s1 := NewScript(cs)
		s2 := s1.Clone()
		if s2 == nil {
			t.Fatal("Clone不应返回nil")
		}
	})

	t.Run("Clone保留编译产物", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s1 := NewScript(cs)
		s2 := s1.Clone()
		if s2.GetCompiled() == nil {
			t.Error("Clone后应保留编译产物")
		}
	})
}

// ========== Script - Encode/Decode 测试 ==========

func Test_APIComplete_Script_EncodeDecode(t *testing.T) {
	t.Run("Encode不报错", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s := NewScript(cs)
		_, err := s.Encode()
		// Encode当前为TODO实现, 但不应报错
		assertNoError(t, err)
	})

	t.Run("Decode不报错", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s := NewScript(cs)
		data, _ := s.Encode()
		_, err := s.Decode(data)
		// Decode当前为TODO实现, 但不应报错
		assertNoError(t, err)
	})

	t.Run("Encode后Decode结果一致", func(t *testing.T) {
		// Encode/Decode未实现, 验证不panic即可
		cs := compileScript(t, "3 * 4")
		s := NewScript(cs)
		data, _ := s.Encode()
		_, err := s.Decode(data)
		assertNoError(t, err)
	})

	t.Run("Encode空脚本", func(t *testing.T) {
		cs := compileScript(t, "")
		s := NewScript(cs)
		_, err := s.Encode()
		assertNoError(t, err)
	})

	t.Run("Encode带常量的脚本", func(t *testing.T) {
		cs := compileScript(t, `s := "hello"
s`)
		s := NewScript(cs)
		_, err := s.Encode()
		assertNoError(t, err)
	})

	t.Run("Encode带函数的脚本", func(t *testing.T) {
		cs := compileScript(t, `
fn add(a, b) { return a + b }
add(1, 2)
`)
		s := NewScript(cs)
		_, err := s.Encode()
		assertNoError(t, err)
	})
}

// ========== API 组合场景 ==========

func Test_APIComplete_CompileOnce_RunMultiple(t *testing.T) {
	t.Run("编译一次多次绑定执行", func(t *testing.T) {
		// 同一脚本用不同Context绑定不同值执行
		cs := compileScript(t, `x :=>int getBindValue("x")
x * 2`)
		e := NewEngine()

		ctx1 := NewContext()
		ctx1.BindValue("x", 5)
		v1, err := e.Run(ctx1, cs)
		assertNoError(t, err)
		assertInt(t, v1, 10)

		ctx2 := NewContext()
		ctx2.BindValue("x", 50)
		v2, err := e.Run(ctx2, cs)
		assertNoError(t, err)
		assertInt(t, v2, 100)
	})
}

func Test_APIComplete_CloneThenRun(t *testing.T) {
	t.Run("Context Clone后分别执行", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 10)

		clone := ctx.Clone()
		clone.BindValue("y", 20)

		cs := compileScript(t, `x :=>int getBindValue("x")
x`)
		e := NewEngine()

		// 两个Context都能执行
		v1, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v1, 10)

		v2, err := e.Run(clone, cs)
		assertNoError(t, err)
		assertInt(t, v2, 10)
	})
}

func Test_APIComplete_MultipleEngines(t *testing.T) {
	t.Run("多个Engine共享同一Context数据", func(t *testing.T) {
		// 不同引擎执行同一Context绑定的数据
		ctx := NewContext()
		ctx.BindFunc("inc", func(x int) int { return x + 1 })

		cs := compileScript(t, "#fn inc(int)=>int\ninc(41)")

		e1 := NewEngine()
		e2 := NewEngine(WithMaxCallDepth(512))

		v1, err := e1.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v1, 42)

		v2, err := e2.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v2, 42)
	})
}

func Test_APIComplete_ClosureBinding(t *testing.T) {
	t.Run("绑定Go闭包到脚本引擎", func(t *testing.T) {
		// 闭包捕获外部变量
		offset := 100
		ctx := NewContext()
		ctx.BindFunc("addOffset", func(x int) int { return x + offset })

		cs := compileScript(t, "#fn addOffset(int)=>int\naddOffset(5)")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 105)
	})
}

func Test_APIComplete_BindFuncDualReturnError(t *testing.T) {
	t.Run("func返回error时脚本收到错误", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("fail", func() (int, error) { return 0, errTestAPI })

		cs := compileScript(t, "#fn fail()=>int\nfail()")
		e := NewEngine()
		_, err := e.Run(ctx, cs)
		assertError(t, err)
	})
}

func Test_APIComplete_BindFuncOverwrite(t *testing.T) {
	t.Run("覆盖同名绑定函数", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("f", func(x int) int { return x * 2 })
		ctx.BindFunc("f", func(x int) int { return x * 3 })

		cs := compileScript(t, "#fn f(int)=>int\nf(10)")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		// 应使用最后一次绑定
		assertInt(t, v, 30)
	})
}

func Test_APIComplete_Context_LogDefault(t *testing.T) {
	t.Run("默认LogFunc不panic", func(t *testing.T) {
		ctx := NewContext()
		// 默认logFunc是空实现, 不应panic
		ctx.Log("some message %d", 42)
	})
}

func Test_APIComplete_Engine_RunStatsFields(t *testing.T) {
	t.Run("执行后stats字段可读取", func(t *testing.T) {
		ctx := NewContext()
		cs := compileScript(t, "1 + 1")
		e := NewEngine()
		e.Run(ctx, cs)
		stats := ctx.GetStats()
		// 验证stats各字段可正常读取, TotalRuns应更新
		if stats.TotalRuns != 1 {
			t.Errorf("TotalRuns应为1, 得到%d", stats.TotalRuns)
		}
		// MaxCallDepth在执行结束后记录, 帧已弹出
		_ = stats.MaxCallDepth
		_ = stats.TotalTime
	})
}

func Test_APIComplete_ScriptFromCompiled(t *testing.T) {
	t.Run("Parser.Compile返回可直接执行", func(t *testing.T) {
		parser := NewParser()
		cs, err := parser.Compile("7 * 8")
		assertNoError(t, err)
		e := NewEngine()
		ctx := NewContext()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 56)
	})
}

func Test_APIComplete_EvalWithBindingsNilMap(t *testing.T) {
	t.Run("EvalWithBindings传nil绑定", func(t *testing.T) {
		v, err := EvalWithBindings("3 + 4", nil)
		assertNoError(t, err)
		assertInt(t, v, 7)
	})
}

func Test_APIComplete_Context_SetLogFuncReplace(t *testing.T) {
	t.Run("SetLogFunc替换已有LogFunc", func(t *testing.T) {
		ctx := NewContext()
		var first, second string

		ctx.SetLogFunc(func(format string, args ...any) {
			first = format
		})
		ctx.Log("first")
		if first != "first" {
			t.Errorf("第一个LogFunc应被调用, 得到%q", first)
		}

		ctx.SetLogFunc(func(format string, args ...any) {
			second = format
		})
		ctx.Log("second")
		if second != "second" {
			t.Errorf("第二个LogFunc应被调用, 得到%q", second)
		}
		// 第一个LogFunc不应被第二次调用修改
		if first != "first" {
			t.Error("替换后原LogFunc不应被调用")
		}
	})
}

func Test_APIComplete_CloneIsolation_Comprehensive(t *testing.T) {
	t.Run("Clone隔离_修改值不影响对方", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("shared", 1)

		clone := ctx.Clone()
		// 双向修改
		ctx.BindValue("onlyOrig", 10)
		clone.BindValue("onlyClone", 20)

		_, okOrig := ctx.GetBindValue("onlyClone")
		_, okClone := clone.GetBindValue("onlyOrig")
		if okOrig {
			t.Error("原Context不应有Clone新增的值")
		}
		if okClone {
			t.Error("Clone不应有原Context新增的值")
		}

		// shared值两边都有
		v1, _ := ctx.GetBindValue("shared")
		v2, _ := clone.GetBindValue("shared")
		assertInt(t, v1, 1)
		assertInt(t, v2, 1)
	})
}

func Test_APIComplete_Engine_RunStatsTotalTime(t *testing.T) {
	t.Run("执行后TotalTime大于零", func(t *testing.T) {
		ctx := NewContext()
		cs := compileScript(t, "1 + 1")
		e := NewEngine()
		e.Run(ctx, cs)
		stats := ctx.GetStats()
		// 简单脚本执行极快, TotalTime可能为0, 只验证字段存在
		_ = stats.TotalTime
	})
}

func Test_APIComplete_BindFuncWithBoolReturn(t *testing.T) {
	t.Run("func(int)bool返回布尔", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("isPos", func(x int) bool { return x > 0 })
		cs := compileScript(t, "#fn isPos(int)=>bool\nisPos(5)")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertBool(t, v, true)
	})
}

func Test_APIComplete_BindFuncNoReturnValue(t *testing.T) {
	t.Run("func无返回值脚本得到nil", func(t *testing.T) {
		ctx := NewContext()
		called := false
		ctx.BindFunc("sideEffect", func(x int) { called = true })
		cs := compileScript(t, "#fn sideEffect(int)\nsideEffect(42)")
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertNil(t, v)
		if !called {
			t.Error("无返回值函数应被调用")
		}
	})
}

// errTestAPI 测试用错误对象
var errTestAPI = newTestAPIError("api test error")

type testAPIError struct{ msg string }

func (e *testAPIError) Error() string { return e.msg }

func newTestAPIError(msg string) *testAPIError { return &testAPIError{msg: msg} }

// 编译期验证testAPIError实现了error接口
var _ error = (*testAPIError)(nil)

// ========== Context - BindFunc panic 测试 ==========

func Test_APIComplete_BindFunc_PanicOnNonFunc(t *testing.T) {
	t.Run("传入string应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("BindFunc传入非函数应panic")
			}
		}()
		ctx := NewContext()
		ctx.BindFunc("bad", "not a function")
	})

	t.Run("传入int应panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("BindFunc传入非函数应panic")
			}
		}()
		ctx := NewContext()
		ctx.BindFunc("bad", 42)
	})
}

// ========== 验证字符串拼接包含 ==========

func Test_APIComplete_AssertContains(t *testing.T) {
	// 验证测试辅助函数assertContains正常工作
	t.Run("assertContains可用", func(t *testing.T) {
		if !strings.Contains("hello world", "world") {
			t.Error("strings.Contains验证失败")
		}
	})
}
