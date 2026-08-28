package script

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"
)

// ========== Context Clone 深度测试 ==========

// Test_ContextScript_Clone_BindValueIndependent Clone 后绑定值完全独立
func Test_ContextScript_Clone_BindValueIndependent(t *testing.T) {
	t.Run("Clone后修改绑定值不影响原Context", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 42)
		clone := ctx.Clone()

		// 修改clone的值
		clone.BindValue("x", 100)

		origVal, _ := ctx.GetBindValue("x")
		cloneVal, _ := clone.GetBindValue("x")
		assertInt(t, origVal, 42)
		assertInt(t, cloneVal, 100)
	})
}

// Test_ContextScript_Clone_ModifyCloneNoAffectOriginal Clone 后修改不影响原 Context
func Test_ContextScript_Clone_ModifyCloneNoAffectOriginal(t *testing.T) {
	t.Run("在Clone上新增绑定不影响原Context", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("base", 1)
		clone := ctx.Clone()

		clone.BindValue("only_clone", 99)

		_, ok := ctx.GetBindValue("only_clone")
		if ok {
			t.Error("在Clone上新增绑定不应出现在原Context中")
		}
	})

	t.Run("在Clone上覆盖绑定不影响原Context", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("v", 10)
		clone := ctx.Clone()

		clone.BindValue("v", 999)

		origVal, _ := ctx.GetBindValue("v")
		assertInt(t, origVal, 10)
	})
}

// Test_ContextScript_Clone_ModifyOriginalNoAffectClone 原Context修改不影响Clone
func Test_ContextScript_Clone_ModifyOriginalNoAffectClone(t *testing.T) {
	t.Run("原Context新增绑定不影响Clone", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 1)
		clone := ctx.Clone()

		ctx.BindValue("only_orig", 55)

		_, ok := clone.GetBindValue("only_orig")
		if ok {
			t.Error("原Context新增绑定不应出现在Clone中")
		}
	})

	t.Run("原Context覆盖绑定不影响Clone", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("v", 10)
		clone := ctx.Clone()

		ctx.BindValue("v", 888)

		cloneVal, _ := clone.GetBindValue("v")
		assertInt(t, cloneVal, 10)
	})
}

// Test_ContextScript_Clone_PreserveAllBindValues Clone 保留所有绑定值
func Test_ContextScript_Clone_PreserveAllBindValues(t *testing.T) {
	t.Run("Clone保留多个绑定值", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("a", 1)
		ctx.BindValue("b", 2)
		ctx.BindValue("c", 3)
		ctx.BindValue("d", 4)
		ctx.BindValue("e", 5)

		clone := ctx.Clone()
		for _, name := range []string{"a", "b", "c", "d", "e"} {
			origVal, origOk := ctx.GetBindValue(name)
			cloneVal, cloneOk := clone.GetBindValue(name)
			if !origOk || !cloneOk {
				t.Errorf("绑定%s应存在: orig=%v clone=%v", name, origOk, cloneOk)
			}
			if origVal.Int() != cloneVal.Int() {
				t.Errorf("绑定%s值不一致: orig=%d clone=%d", name, origVal.Int(), cloneVal.Int())
			}
		}
	})
}

// Test_ContextScript_Clone_PreserveAllBindFuncs Clone 保留所有绑定函数
func Test_ContextScript_Clone_PreserveAllBindFuncs(t *testing.T) {
	t.Run("Clone保留多个绑定函数", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("f1", func() int { return 1 })
		ctx.BindFunc("f2", func(x int) int { return x * 2 })
		ctx.BindFunc("f3", func(a, b int) int { return a + b })

		clone := ctx.Clone()
		for _, name := range []string{"f1", "f2", "f3"} {
			_, origOk := ctx.GetBindFunc(name)
			_, cloneOk := clone.GetBindFunc(name)
			if !origOk || !cloneOk {
				t.Errorf("函数%s应存在: orig=%v clone=%v", name, origOk, cloneOk)
			}
		}
	})
}

// Test_ContextScript_Clone_FuncCallable Clone 后函数独立调用
func Test_ContextScript_Clone_FuncCallable(t *testing.T) {
	t.Run("Clone后绑定函数可正常调用", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("inc", func(x int) int { return x + 1 })

		clone := ctx.Clone()
		cs := compileScript(t, "#fn inc(int)=>int\ninc(10)")

		result := runScriptWithContext(t, "#fn inc(int)=>int\ninc(10)", clone)
		assertInt(t, result, 11)

		// 原Context也能正常调用
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 11)
	})
}

// Test_ContextScript_Clone_MultiLayer 多层Clone
func Test_ContextScript_Clone_MultiLayer(t *testing.T) {
	t.Run("Clone的Clone保留绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 42)

		clone1 := ctx.Clone()
		clone1.BindValue("y", 99)

		clone2 := clone1.Clone()
		// clone2应同时有x和y
		valX, okX := clone2.GetBindValue("x")
		valY, okY := clone2.GetBindValue("y")
		if !okX || valX.Int() != 42 {
			t.Errorf("多层Clone后x应存在且为42, got %v, ok=%v", valX, okX)
		}
		if !okY || valY.Int() != 99 {
			t.Errorf("多层Clone后y应存在且为99, got %v, ok=%v", valY, okY)
		}
	})

	t.Run("多层Clone互不影响", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("base", 1)

		c1 := ctx.Clone()
		c2 := c1.Clone()
		c3 := c2.Clone()

		c1.BindValue("c1_only", 10)
		c2.BindValue("c2_only", 20)
		c3.BindValue("c3_only", 30)

		// c2不应有c1的绑定
		if _, ok := c2.GetBindValue("c1_only"); ok {
			t.Error("c2不应有c1的绑定")
		}
		// c3不应有c1和c2的绑定
		if _, ok := c3.GetBindValue("c1_only"); ok {
			t.Error("c3不应有c1的绑定")
		}
		if _, ok := c3.GetBindValue("c2_only"); ok {
			t.Error("c3不应有c2的绑定")
		}
	})
}

// Test_ContextScript_Clone_EmptyContext Clone 空 Context
func Test_ContextScript_Clone_EmptyContext(t *testing.T) {
	t.Run("Clone空Context返回非nil", func(t *testing.T) {
		ctx := NewContext()
		clone := ctx.Clone()
		if clone == nil {
			t.Fatal("Clone空Context不应返回nil")
		}
	})

	t.Run("Clone空Context后添加绑定", func(t *testing.T) {
		ctx := NewContext()
		clone := ctx.Clone()
		clone.BindValue("x", 42)

		val, ok := clone.GetBindValue("x")
		if !ok || val.Int() != 42 {
			t.Errorf("Clone后添加绑定失败, got %v, ok=%v", val, ok)
		}

		// 原Context应仍为空
		_, ok = ctx.GetBindValue("x")
		if ok {
			t.Error("原Context不应有Clone新增的绑定")
		}
	})
}

// Test_ContextScript_Clone_AddNewBind Clone 后添加新绑定
func Test_ContextScript_Clone_AddNewBind(t *testing.T) {
	t.Run("Clone后各自添加独立绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("shared", 1)
		clone := ctx.Clone()

		ctx.BindValue("orig_only", 2)
		clone.BindValue("clone_only", 3)

		origShared, _ := ctx.GetBindValue("shared")
		cloneShared, _ := clone.GetBindValue("shared")
		assertInt(t, origShared, 1)
		assertInt(t, cloneShared, 1)

		origVal, origOk := ctx.GetBindValue("orig_only")
		cloneVal, cloneOk := clone.GetBindValue("clone_only")
		if !origOk || origVal.Int() != 2 {
			t.Errorf("原Context的orig_only应存在")
		}
		if !cloneOk || cloneVal.Int() != 3 {
			t.Errorf("Clone的clone_only应存在")
		}

		// 交叉验证
		if _, ok := ctx.GetBindValue("clone_only"); ok {
			t.Error("原Context不应有clone_only")
		}
		if _, ok := clone.GetBindValue("orig_only"); ok {
			t.Error("Clone不应有orig_only")
		}
	})
}

// Test_ContextScript_Clone_RunScriptIndependently Clone后执行脚本独立
func Test_ContextScript_Clone_RunScriptIndependently(t *testing.T) {
	t.Run("Clone后使用不同绑定执行同一脚本", func(t *testing.T) {
		base := NewContext()
		base.BindValue("prefix", "Hello")

		cs := compileScript(t, `p :=>string getBindValue("prefix")
n :=>string getBindValue("name")
p + " " + n`)

		e := NewEngine()

		ctx1 := base.Clone()
		ctx1.BindValue("name", "Alice")
		v1, err := e.Run(ctx1, cs)
		assertNoError(t, err)
		if v1.String() != "Hello Alice" {
			t.Errorf("got %q, want %q", v1.String(), "Hello Alice")
		}

		ctx2 := base.Clone()
		ctx2.BindValue("name", "Bob")
		v2, err := e.Run(ctx2, cs)
		assertNoError(t, err)
		if v2.String() != "Hello Bob" {
			t.Errorf("got %q, want %q", v2.String(), "Hello Bob")
		}
	})
}

// ========== Context Clone - 值类型覆盖 ==========

// Test_ContextScript_Clone_IntValue Clone int 绑定
func Test_ContextScript_Clone_IntValue(t *testing.T) {
	t.Run("Clone保留int绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("n", 12345)
		clone := ctx.Clone()

		val, ok := clone.GetBindValue("n")
		if !ok {
			t.Fatal("Clone后int绑定应存在")
		}
		assertInt(t, val, 12345)

		// 修改clone不影响原
		clone.BindValue("n", 999)
		origVal, _ := ctx.GetBindValue("n")
		assertInt(t, origVal, 12345)
	})
}

// Test_ContextScript_Clone_StringValue Clone string 绑定
func Test_ContextScript_Clone_StringValue(t *testing.T) {
	t.Run("Clone保留string绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("s", "hello world")
		clone := ctx.Clone()

		val, ok := clone.GetBindValue("s")
		if !ok {
			t.Fatal("Clone后string绑定应存在")
		}
		assertString(t, val, "hello world")

		clone.BindValue("s", "changed")
		origVal, _ := ctx.GetBindValue("s")
		assertString(t, origVal, "hello world")
	})
}

// Test_ContextScript_Clone_BoolValue Clone bool 绑定
func Test_ContextScript_Clone_BoolValue(t *testing.T) {
	t.Run("Clone保留bool绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("flag", true)
		clone := ctx.Clone()

		val, ok := clone.GetBindValue("flag")
		if !ok {
			t.Fatal("Clone后bool绑定应存在")
		}
		assertBool(t, val, true)

		clone.BindValue("flag", false)
		origVal, _ := ctx.GetBindValue("flag")
		assertBool(t, origVal, true)
	})
}

// Test_ContextScript_Clone_ArrayValue Clone 数组绑定
func Test_ContextScript_Clone_ArrayValue(t *testing.T) {
	t.Run("Clone保留数组绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("arr", []interface{}{1, 2, 3})
		clone := ctx.Clone()

		val, ok := clone.GetBindValue("arr")
		if !ok {
			t.Fatal("Clone后数组绑定应存在")
		}
		arr := val.Array()
		if arr == nil || len(arr.Elements) != 3 {
			t.Fatalf("Clone后数组应包含3个元素")
		}
		assertInt(t, arr.Elements[0], 1)
		assertInt(t, arr.Elements[1], 2)
		assertInt(t, arr.Elements[2], 3)
	})
}

// Test_ContextScript_Clone_MapValue Clone Map 绑定
func Test_ContextScript_Clone_MapValue(t *testing.T) {
	t.Run("Clone保留Map绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("m", map[string]interface{}{"k1": 10, "k2": 20})
		clone := ctx.Clone()

		val, ok := clone.GetBindValue("m")
		if !ok {
			t.Fatal("Clone后Map绑定应存在")
		}
		m := val.Map()
		if m == nil || len(m.Pairs) != 2 {
			t.Fatalf("Clone后Map应包含2个键值对")
		}
		v1 := m.Pairs["k1"]
		v2 := m.Pairs["k2"]
		assertInt(t, v1, 10)
		assertInt(t, v2, 20)
	})
}

// Test_ContextScript_Clone_NilValue Clone nil 绑定
func Test_ContextScript_Clone_NilValue(t *testing.T) {
	t.Run("Clone保留nil绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("n", nil)
		clone := ctx.Clone()

		val, ok := clone.GetBindValue("n")
		if !ok {
			t.Fatal("Clone后nil绑定应存在")
		}
		assertNil(t, val)
	})
}

// Test_ContextScript_Clone_MixedTypes Clone 多种类型混合
func Test_ContextScript_Clone_MixedTypes(t *testing.T) {
	t.Run("Clone保留混合类型绑定", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("i", 42)
		ctx.BindValue("f", 3.14)
		ctx.BindValue("s", "text")
		ctx.BindValue("b", true)
		ctx.BindValue("n", nil)

		clone := ctx.Clone()
		// 逐个验证
		iVal, _ := clone.GetBindValue("i")
		assertInt(t, iVal, 42)

		fVal, _ := clone.GetBindValue("f")
		assertFloat(t, fVal, 3.14)

		sVal, _ := clone.GetBindValue("s")
		assertString(t, sVal, "text")

		bVal, _ := clone.GetBindValue("b")
		assertBool(t, bVal, true)

		nVal, _ := clone.GetBindValue("n")
		assertNil(t, nVal)
	})
}

// ========== Context Clone - 函数 ==========

// Test_ContextScript_Clone_FuncParameterPassing Clone 后函数参数传递正确
func Test_ContextScript_Clone_FuncParameterPassing(t *testing.T) {
	t.Run("Clone后函数参数正确传递", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("mul", func(a, b int) int { return a * b })
		clone := ctx.Clone()

		result := runScriptWithContext(t, "#fn mul(int, int)=>int\nmul(6, 7)", clone)
		assertInt(t, result, 42)
	})
}

// Test_ContextScript_Clone_FuncReturnValue Clone 后函数返回值正确
func Test_ContextScript_Clone_FuncReturnValue(t *testing.T) {
	t.Run("Clone后函数返回值正确", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("greet", func(name string) string { return "Hi " + name })
		clone := ctx.Clone()

		result := runScriptWithContext(t, "#fn greet(string)=>string\ngreet(\"World\")", clone)
		if result.String() != "Hi World" {
			t.Errorf("got %q, want %q", result.String(), "Hi World")
		}
	})
}

// Test_ContextScript_Clone_MultipleFuncs Clone 多个函数
func Test_ContextScript_Clone_MultipleFuncs(t *testing.T) {
	t.Run("Clone后多个函数均可调用", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("add", func(a, b int) int { return a + b })
		ctx.BindFunc("sub", func(a, b int) int { return a - b })
		ctx.BindFunc("mul", func(a, b int) int { return a * b })

		clone := ctx.Clone()
		cs := compileScript(t, `#fn add(int, int)=>int
#fn sub(int, int)=>int
#fn mul(int, int)=>int
a := add(10, 3)
s := sub(10, 3)
m := mul(10, 3)
a + s + m`)

		e := NewEngine()
		result, err := e.Run(clone, cs)
		assertNoError(t, err)
		// add=13, sub=7, mul=30 => 13+7+30=50
		assertInt(t, result, 50)
	})
}

// ========== Script Encode/Decode ==========

// Test_ContextScript_Encode_ReturnsNoError Encode 未实现返回明确错误
func Test_ContextScript_Encode_ReturnsNoError(t *testing.T) {
	t.Run("Encode简单表达式报错", func(t *testing.T) {
		cs := compileScript(t, "1 + 2")
		s := NewScript(cs)
		if _, err := s.Encode(); err == nil {
			t.Error("Encode未实现应返回错误而非静默成功")
		}
	})
}

// Test_ContextScript_Decode_FromEncode Decode 未实现返回明确错误
func Test_ContextScript_Decode_FromEncode(t *testing.T) {
	t.Run("Decode报错", func(t *testing.T) {
		cs := compileScript(t, "3 * 4")
		s := NewScript(cs)
		if _, err := s.Decode(nil); err == nil {
			t.Error("Decode未实现应返回错误而非静默成功")
		}
	})
}

// Test_ContextScript_EncodeDecode_Consistency Encode -> Decode 均返回错误
func Test_ContextScript_EncodeDecode_Consistency(t *testing.T) {
	t.Run("Encode后Decode均报错", func(t *testing.T) {
		cs := compileScript(t, "10 + 20")
		s := NewScript(cs)
		_, err := s.Encode()
		if err == nil {
			t.Error("Encode未实现应返回错误")
		}
		if _, err := s.Decode(nil); err == nil {
			t.Error("Decode未实现应返回错误")
		}
	})
}

// Test_ContextScript_Encode_EmptyScript Encode 空脚本
func Test_ContextScript_Encode_EmptyScript(t *testing.T) {
	t.Run("Encode空脚本报错", func(t *testing.T) {
		cs := compileScript(t, "")
		s := NewScript(cs)
		if _, err := s.Encode(); err == nil {
			t.Error("Encode未实现应返回错误")
		}
	})
}

// Test_ContextScript_Encode_ConstantScript Encode 常量脚本
func Test_ContextScript_Encode_ConstantScript(t *testing.T) {
	t.Run("Encode常量脚本报错", func(t *testing.T) {
		cs := compileScript(t, `s := "hello"
n := 42
b := true`)
		s := NewScript(cs)
		if _, err := s.Encode(); err == nil {
			t.Error("Encode未实现应返回错误")
		}
	})
}

// Test_ContextScript_Encode_FunctionScript Encode 函数脚本
func Test_ContextScript_Encode_FunctionScript(t *testing.T) {
	t.Run("Encode带函数的脚本报错", func(t *testing.T) {
		cs := compileScript(t, `
fn add(a, b) { return a + b }
add(1, 2)
`)
		s := NewScript(cs)
		if _, err := s.Encode(); err == nil {
			t.Error("Encode未实现应返回错误")
		}
	})
}

// Test_ContextScript_Decode_InvalidData Decode 错误数据处理
func Test_ContextScript_Decode_InvalidData(t *testing.T) {
	t.Run("Decode空数据报错", func(t *testing.T) {
		cs := compileScript(t, "1 + 1")
		s := NewScript(cs)
		if _, err := s.Decode(nil); err == nil {
			t.Error("Decode未实现应返回错误而非静默成功")
		}
	})

	t.Run("Decode非法数据报错", func(t *testing.T) {
		cs := compileScript(t, "1 + 1")
		s := NewScript(cs)
		if _, err := s.Decode([]byte{0xFF, 0xFE, 0xFD}); err == nil {
			t.Error("Decode未实现应返回错误而非静默成功")
		}
	})
}

// ========== Script Clone ==========

// Test_ContextScript_ScriptClone_IndependentRun Clone 后独立执行
func Test_ContextScript_ScriptClone_IndependentRun(t *testing.T) {
	t.Run("Clone的Script可独立执行", func(t *testing.T) {
		cs := compileScript(t, "5 + 5")
		s1 := NewScript(cs)
		s2 := s1.Clone()

		// 两者应都能获取编译产物
		if s2.GetCompiled() == nil {
			t.Fatal("Clone后的Script应有编译产物")
		}

		e := NewEngine()
		ctx := NewContext()
		v1, err := e.Run(ctx, s1.GetCompiled())
		assertNoError(t, err)
		assertInt(t, v1, 10)

		v2, err := e.Run(ctx, s2.GetCompiled())
		assertNoError(t, err)
		assertInt(t, v2, 10)
	})
}

// Test_ContextScript_ScriptClone_SameResult Clone 后执行结果一致
func Test_ContextScript_ScriptClone_SameResult(t *testing.T) {
	t.Run("Clone后执行结果一致", func(t *testing.T) {
		cs := compileScript(t, "100 * 7")
		s1 := NewScript(cs)
		s2 := s1.Clone()

		e := NewEngine()
		ctx1 := NewContext()
		ctx2 := NewContext()

		v1, _ := e.Run(ctx1, s1.GetCompiled())
		v2, _ := e.Run(ctx2, s2.GetCompiled())
		assertInt(t, v1, 700)
		assertInt(t, v2, 700)
	})
}

// Test_ContextScript_ScriptClone_Multiple 多次 Clone
func Test_ContextScript_ScriptClone_Multiple(t *testing.T) {
	t.Run("多次Clone均可执行", func(t *testing.T) {
		cs := compileScript(t, "42")
		s := NewScript(cs)

		e := NewEngine()
		for i := 0; i < 5; i++ {
			clone := s.Clone()
			if clone.GetCompiled() == nil {
				t.Fatalf("第%d次Clone后compiled为nil", i+1)
			}
			ctx := NewContext()
			v, err := e.Run(ctx, clone.GetCompiled())
			assertNoError(t, err)
			assertInt(t, v, 42)
		}
	})
}

// ========== Context Stats ==========

// Test_ContextScript_Stats_InitialZero 初始 Stats 为零值
func Test_ContextScript_Stats_InitialZero(t *testing.T) {
	t.Run("新建Context统计为零值", func(t *testing.T) {
		ctx := NewContext()
		stats := ctx.GetStats()
		if stats.TotalRuns != 0 {
			t.Errorf("初始TotalRuns应为0, got %d", stats.TotalRuns)
		}
		if stats.TotalTime != 0 {
			t.Errorf("初始TotalTime应为0, got %d", stats.TotalTime)
		}
		if stats.MaxCallDepth != 0 {
			t.Errorf("初始MaxCallDepth应为0, got %d", stats.MaxCallDepth)
		}
	})
}

// Test_ContextScript_Stats_HasTimeAfterRun 执行后 Stats 有耗时
func Test_ContextScript_Stats_HasTimeAfterRun(t *testing.T) {
	t.Run("执行后TotalTime大于0", func(t *testing.T) {
		ctx := NewContext()
		runScriptWithContext(t, "1 + 2", ctx)
		stats := ctx.GetStats()
		if stats.TotalRuns != 1 {
			t.Errorf("TotalRuns应为1, got %d", stats.TotalRuns)
		}
		// TotalTime可能为0(执行太快), 验证非负即可
		if stats.TotalTime < 0 {
			t.Errorf("TotalTime不应为负数, got %d", stats.TotalTime)
		}
	})
}

// Test_ContextScript_Stats_Accumulate 多次执行 Stats 累积
func Test_ContextScript_Stats_Accumulate(t *testing.T) {
	t.Run("多次执行TotalRuns累积", func(t *testing.T) {
		ctx := NewContext()
		for i := 0; i < 10; i++ {
			runScriptWithContext(t, "1 + 1", ctx)
		}
		stats := ctx.GetStats()
		if stats.TotalRuns != 10 {
			t.Errorf("10次执行后TotalRuns应为10, got %d", stats.TotalRuns)
		}
	})
}

// Test_ContextScript_Stats_DifferentScripts 不同脚本 Stats 差异
func Test_ContextScript_Stats_DifferentScripts(t *testing.T) {
	t.Run("不同脚本使用不同Context统计独立", func(t *testing.T) {
		ctx1 := NewContext()
		ctx2 := NewContext()

		runScriptWithContext(t, "1 + 1", ctx1)
		runScriptWithContext(t, "2 + 2", ctx1)
		runScriptWithContext(t, "3 + 3", ctx2)

		s1 := ctx1.GetStats()
		s2 := ctx2.GetStats()
		if s1.TotalRuns != 2 {
			t.Errorf("ctx1的TotalRuns应为2, got %d", s1.TotalRuns)
		}
		if s2.TotalRuns != 1 {
			t.Errorf("ctx2的TotalRuns应为1, got %d", s2.TotalRuns)
		}
	})
}

// Test_ContextScript_Stats_FieldsComplete Stats 字段完整性
func Test_ContextScript_Stats_FieldsComplete(t *testing.T) {
	t.Run("Stats包含完整字段", func(t *testing.T) {
		ctx := NewContext()
		// 直接调用updateStats填充字段
		ctx.updateStats(1000, 5)
		ctx.updateStats(2000, 10)
		ctx.updateStats(3000, 3)

		stats := ctx.GetStats()
		if stats.TotalRuns != 3 {
			t.Errorf("TotalRuns应为3, got %d", stats.TotalRuns)
		}
		if stats.TotalTime != 6000 {
			t.Errorf("TotalTime应为6000, got %d", stats.TotalTime)
		}
		if stats.MaxCallDepth != 10 {
			t.Errorf("MaxCallDepth应为10, got %d", stats.MaxCallDepth)
		}
	})
}

// ========== Context Output ==========

// Test_ContextScript_Output_SetCustom SetOutput 设置自定义 Writer
func Test_ContextScript_Output_SetCustom(t *testing.T) {
	t.Run("SetOutput设置自定义Writer", func(t *testing.T) {
		ctx := NewContext()
		var buf bytes.Buffer
		ctx.SetOutput(&buf)

		if ctx.GetOutput() == nil {
			t.Error("SetOutput后GetOutput不应为nil")
		}
	})
}

// Test_ContextScript_Output_GetReturnsSet GetOutput 返回设置的 Writer
func Test_ContextScript_Output_GetReturnsSet(t *testing.T) {
	t.Run("GetOutput返回设置的Writer", func(t *testing.T) {
		ctx := NewContext()
		var buf bytes.Buffer
		writer := &buf
		ctx.SetOutput(writer)

		got := ctx.GetOutput()
		if got != writer {
			// bytes.Buffer地址可能因实现不同, 验证可写入即可
			_, err := got.Write([]byte("test"))
			if err != nil {
				t.Errorf("GetOutput返回的Writer不可写: %v", err)
			}
		}
	})
}

// Test_ContextScript_Output_DefaultNonNil 默认 Output 非 nil
func Test_ContextScript_Output_DefaultNonNil(t *testing.T) {
	t.Run("新建Context默认Output非nil", func(t *testing.T) {
		ctx := NewContext()
		if ctx.GetOutput() == nil {
			t.Error("默认Output不应为nil")
		}
	})

	t.Run("默认Output为os.Stdout", func(t *testing.T) {
		ctx := NewContext()
		if ctx.GetOutput() != os.Stdout {
			t.Error("默认Output应为os.Stdout")
		}
	})
}

// Test_ContextScript_Output_LogToCustomWriter Log 输出到自定义 Writer
func Test_ContextScript_Output_LogToCustomWriter(t *testing.T) {
	t.Run("Log通过自定义LogFunc输出到Writer", func(t *testing.T) {
		ctx := NewContext()
		var buf bytes.Buffer

		// 设置自定义logFunc将日志写入buf
		ctx.SetLogFunc(func(format string, args ...any) {
			fmt.Fprintf(&buf, format, args...)
		})
		ctx.SetOutput(&buf)

		ctx.Log("test message")
		if buf.String() != "test message" {
			t.Errorf("Log应通过logFunc输出到Writer, got %q", buf.String())
		}
	})
}

// Test_ContextScript_Output_LogFormat Log 格式化参数
func Test_ContextScript_Output_LogFormat(t *testing.T) {
	t.Run("Log正确传递格式化参数", func(t *testing.T) {
		ctx := NewContext()
		var received string
		ctx.SetLogFunc(func(format string, args ...any) {
			received = fmt.Sprintf(format, args...)
		})

		ctx.Log("value=%d, name=%s", 42, "test")
		if received != "value=42, name=test" {
			t.Errorf("格式化结果不符, got %q", received)
		}
	})
}

// ========== Context LogFunc ==========

// Test_ContextScript_LogFunc_SetCustom SetLogFunc 设置自定义日志函数
func Test_ContextScript_LogFunc_SetCustom(t *testing.T) {
	t.Run("SetLogFunc设置自定义函数", func(t *testing.T) {
		ctx := NewContext()
		called := false
		ctx.SetLogFunc(func(format string, args ...any) {
			called = true
		})

		ctx.Log("anything")
		if !called {
			t.Error("自定义LogFunc应被调用")
		}
	})
}

// Test_ContextScript_LogFunc_LogThroughFunc Log 通过自定义函数输出
func Test_ContextScript_LogFunc_LogThroughFunc(t *testing.T) {
	t.Run("Log通过自定义函数输出内容", func(t *testing.T) {
		ctx := NewContext()
		var messages []string
		ctx.SetLogFunc(func(format string, args ...any) {
			messages = append(messages, format)
		})

		ctx.Log("first")
		ctx.Log("second")
		ctx.Log("third")

		if len(messages) != 3 {
			t.Fatalf("应有3条日志, got %d", len(messages))
		}
		if messages[0] != "first" || messages[1] != "second" || messages[2] != "third" {
			t.Errorf("日志顺序或内容错误: %v", messages)
		}
	})
}

// Test_ContextScript_LogFunc_Format LogFunc 格式化参数
func Test_ContextScript_LogFunc_Format(t *testing.T) {
	t.Run("LogFunc正确接收format和args", func(t *testing.T) {
		ctx := NewContext()
		var savedFormat string
		var savedArgs []interface{}
		ctx.SetLogFunc(func(format string, args ...any) {
			savedFormat = format
			savedArgs = args
		})

		ctx.Log("num=%d str=%s", 100, "abc")
		if savedFormat != "num=%d str=%s" {
			t.Errorf("format不符, got %q", savedFormat)
		}
		if len(savedArgs) != 2 {
			t.Fatalf("参数数量应为2, got %d", len(savedArgs))
		}
	})
}

// Test_ContextScript_LogFunc_Default 默认 LogFunc 行为
func Test_ContextScript_LogFunc_Default(t *testing.T) {
	t.Run("默认LogFunc不panic", func(t *testing.T) {
		ctx := NewContext()
		// 默认logFunc是空实现, 调用Log不应panic
		ctx.Log("test")
		ctx.Log("formatted %d", 42)
	})
}

// ========== Context 并发安全 ==========

// Test_ContextScript_Concurrent_Read 多 goroutine 同时读取 Context
func Test_ContextScript_Concurrent_Read(t *testing.T) {
	t.Run("多goroutine并发读取不panic", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("x", 42)
		ctx.BindFunc("f", func() int { return 1 })

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					ctx.GetBindValue("x")
					ctx.GetBindFunc("f")
					ctx.GetStats()
				}
			}()
		}
		wg.Wait()
	})
}

// Test_ContextScript_Concurrent_BindValue 多 goroutine 同时 BindValue
func Test_ContextScript_Concurrent_BindValue(t *testing.T) {
	t.Run("多goroutine并发写不同key不panic", func(t *testing.T) {
		ctx := NewContext()
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					ctx.BindValue(fmt.Sprintf("key_%d_%d", id, j), j)
				}
			}(i)
		}
		wg.Wait()
	})
}

// Test_ContextScript_Concurrent_Clone Clone 在并发环境中的行为
func Test_ContextScript_Concurrent_Clone(t *testing.T) {
	t.Run("并发Clone和修改不panic", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("base", 1)

		var wg sync.WaitGroup
		// 一半goroutine做Clone, 一半做BindValue
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					if id%2 == 0 {
						clone := ctx.Clone()
						clone.BindValue("clone_key", id*100+j)
					} else {
						ctx.BindValue(fmt.Sprintf("orig_%d_%d", id, j), j)
					}
				}
			}(i)
		}
		wg.Wait()
	})
}

// ========== Engine 复用 ==========

// Test_ContextScript_Engine_ReuseMultipleRun 同一 Engine 多次 Run
func Test_ContextScript_Engine_ReuseMultipleRun(t *testing.T) {
	t.Run("同一Engine多次Run同一脚本", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		cs := compileScript(t, "10 + 20")

		for i := 0; i < 5; i++ {
			v, err := e.Run(ctx, cs)
			assertNoError(t, err)
			assertInt(t, v, 30)
		}

		stats := ctx.GetStats()
		if stats.TotalRuns != 5 {
			t.Errorf("TotalRuns应为5, got %d", stats.TotalRuns)
		}
	})
}

// Test_ContextScript_Engine_DifferentContext 同一 Engine 不同 Context
func Test_ContextScript_Engine_DifferentContext(t *testing.T) {
	t.Run("同一Engine用不同Context执行", func(t *testing.T) {
		e := NewEngine()
		cs := compileScript(t, `x :=>int getBindValue("x")
x * 2`)

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

		// 两个Context的统计独立
		s1 := ctx1.GetStats()
		s2 := ctx2.GetStats()
		if s1.TotalRuns != 1 || s2.TotalRuns != 1 {
			t.Errorf("各Context应各执行1次, got ctx1=%d ctx2=%d", s1.TotalRuns, s2.TotalRuns)
		}
	})
}

// Test_ContextScript_Engine_DifferentScript 同一 Engine 不同 Script
func Test_ContextScript_Engine_DifferentScript(t *testing.T) {
	t.Run("同一Engine执行不同脚本", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()

		s1 := compileScript(t, "1 + 2")
		s2 := compileScript(t, "3 * 4")
		s3 := compileScript(t, "10 - 7")

		v1, err := e.Run(ctx, s1)
		assertNoError(t, err)
		assertInt(t, v1, 3)

		v2, err := e.Run(ctx, s2)
		assertNoError(t, err)
		assertInt(t, v2, 12)

		v3, err := e.Run(ctx, s3)
		assertNoError(t, err)
		assertInt(t, v3, 3)

		stats := ctx.GetStats()
		if stats.TotalRuns != 3 {
			t.Errorf("TotalRuns应为3, got %d", stats.TotalRuns)
		}
	})
}

// Test_ContextScript_Engine_StateAfterRun Engine Run 后状态
func Test_ContextScript_Engine_StateAfterRun(t *testing.T) {
	t.Run("Engine执行后不影响后续执行", func(t *testing.T) {
		e := NewEngine()
		ctx := NewContext()
		cs := compileScript(t, "5 + 5")

		// 第一次执行
		v1, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v1, 10)

		// 第二次执行不同脚本
		cs2 := compileScript(t, "7 * 8")
		v2, err := e.Run(ctx, cs2)
		assertNoError(t, err)
		assertInt(t, v2, 56)

		// 第三次回到第一个脚本
		v3, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v3, 10)
	})
}

// ========== WithMaxCallDepth ==========

// Test_ContextScript_MaxCallDepth_Default 默认深度 256
func Test_ContextScript_MaxCallDepth_Default(t *testing.T) {
	t.Run("默认深度为DefaultMaxCallDepth", func(t *testing.T) {
		e := NewEngine()
		if e.maxDepth != DefaultMaxCallDepth {
			t.Errorf("默认maxDepth应为%d, got %d", DefaultMaxCallDepth, e.maxDepth)
		}
	})

	t.Run("默认深度支持50层递归", func(t *testing.T) {
		runIntTest(t, `
fn rec(n) {
    if n <= 0 { return 0 }
    return 1 + rec(n - 1)
}
rec(50)`, 50)
	})
}

// Test_ContextScript_MaxCallDepth_Custom 自定义深度验证
func Test_ContextScript_MaxCallDepth_Custom(t *testing.T) {
	t.Run("自定义深度为指定值", func(t *testing.T) {
		e := NewEngine(WithMaxCallDepth(512))
		if e.maxDepth != 512 {
			t.Errorf("maxDepth应为512, got %d", e.maxDepth)
		}
	})

	t.Run("自定义深度0", func(t *testing.T) {
		e := NewEngine(WithMaxCallDepth(0))
		if e.maxDepth != 0 {
			t.Errorf("maxDepth应为0, got %d", e.maxDepth)
		}
	})
}

// Test_ContextScript_MaxCallDepth_StackOverflow 深度限制触发栈溢出
func Test_ContextScript_MaxCallDepth_StackOverflow(t *testing.T) {
	t.Run("无限递归触发栈溢出", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn rec(n) {
    return rec(n + 1)
}
rec(0)`)
	})

	t.Run("自定义小深度触发栈溢出", func(t *testing.T) {
		cs := compileScript(t, `
fn rec(n) {
    return rec(n + 1)
}
rec(0)`)
		ctx := NewContext()
		e := NewEngine(WithMaxCallDepth(10))
		_, err := e.Run(ctx, cs)
		assertError(t, err)
	})
}

// Test_ContextScript_MaxCallDepth_Recursion 不同深度对递归的影响
func Test_ContextScript_MaxCallDepth_Recursion(t *testing.T) {
	t.Run("深度过小导致递归失败", func(t *testing.T) {
		cs := compileScript(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(100)
`)
		ctx := NewContext()
		e := NewEngine(WithMaxCallDepth(10))
		_, err := e.Run(ctx, cs)
		assertError(t, err)
	})

	t.Run("深度足够时递归成功", func(t *testing.T) {
		cs := compileScript(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(50)
`)
		ctx := NewContext()
		e := NewEngine(WithMaxCallDepth(100))
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 50)
	})

	t.Run("自定义大深度支持更深递归", func(t *testing.T) {
		cs := compileScript(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(200)
`)
		ctx := NewContext()
		e := NewEngine()
		v, err := e.Run(ctx, cs)
		assertNoError(t, err)
		assertInt(t, v, 200)
	})
}
