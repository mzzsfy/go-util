package script

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

// runScriptWithMaxSteps 使用指定步数上限执行脚本
func runScriptWithMaxSteps(t *testing.T, input string, maxSteps int) (Value, error) {
	t.Helper()
	parser := NewParser()
	compiled, err := parser.Compile(input)
	if err != nil {
		return Value{}, err
	}
	engine := NewEngine(WithMaxSteps(maxSteps))
	return engine.Run(NewContext(), compiled)
}

// Test_ShiftInvalidAmount 负数或超宽移位量返回明确错误而非 runtime panic
func Test_ShiftInvalidAmount(t *testing.T) {
	t.Parallel()
	bits := strconv.IntSize
	cases := []struct {
		name string
		code string
	}{
		{"左移负数", `1 << -1`},
		{"右移负数", `16 >> -1`},
		{"左移超宽", "1 << " + strconv.Itoa(bits)},
		{"右移超宽", "16 >> " + strconv.Itoa(bits)},
		{"左移远超宽", "1 << " + strconv.Itoa(bits*1000)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Eval(c.code)
			if err == nil {
				t.Fatalf("%s 应返回错误", c.code)
			}
			if !errors.Is(err, &RuntimeError{Code: ErrInvalidShift}) {
				t.Errorf("%s 错误码应为 ErrInvalidShift, 实际: %v", c.code, err)
			}
		})
	}
}

// Test_ShiftValid 合法移位行为不变
func Test_ShiftValid(t *testing.T) {
	t.Parallel()
	maxShift := uint(strconv.IntSize - 1)
	cases := []struct {
		code string
		want int
	}{
		{`1 << 4`, 16},
		{`16 >> 2`, 4},
		{`0 >> 0`, 0},
		{`1 << 0`, 1},
		{"1 << " + strconv.Itoa(strconv.IntSize-1), 1 << maxShift},
	}
	for _, c := range cases {
		v := runScript(t, c.code)
		if v.Int() != c.want {
			t.Errorf("%s = %d, 期望 %d", c.code, v.Int(), c.want)
		}
	}
}

// Test_ShiftTypeMismatch 非整数移位仍返回类型错误
func Test_ShiftTypeMismatch(t *testing.T) {
	t.Parallel()
	_, err := Eval(`1.5 << 2`)
	if err == nil {
		t.Fatal("浮点移位应返回类型错误")
	}
}

// Test_NestedSameNameCountFor 同名 for 计数变量嵌套不越界不死循环, 各层槽位独立
func Test_NestedSameNameCountFor(t *testing.T) {
	t.Parallel()
	t.Run("嵌套计数不越界", func(t *testing.T) {
		code := `fn f() {
	a := 1
	i := 0
	for i := 2 {
		a = i
	}
	return a
}
f()`
		v, err := runScriptWithMaxSteps(t, code, 10000)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if v.Int() != 2 {
			t.Errorf("结果 %d, 期望 2", v.Int())
		}
	})

	t.Run("嵌套计数不死循环", func(t *testing.T) {
		code := `c := 0
for i := 3 {
	for i := 2 {
		c = c + 1
	}
}
c`
		v, err := runScriptWithMaxSteps(t, code, 100000)
		if err != nil {
			t.Fatalf("执行失败(疑似死循环): %v", err)
		}
		if v.Int() != 6 {
			t.Errorf("结果 %d, 期望 6", v.Int())
		}
	})

	t.Run("内层结束后外层变量恢复", func(t *testing.T) {
		code := `total := 0
for i := 2 {
	for i := 1 {
		total = total + 10
	}
	total = total + i
}
total`
		v, err := runScriptWithMaxSteps(t, code, 100000)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if v.Int() != 23 {
			t.Errorf("结果 %d, 期望 23", v.Int())
		}
	})

	t.Run("顺序同名循环独立", func(t *testing.T) {
		code := `s := 0
for i := 3 {
	s = s + i
}
for i := 5 {
	s = s + i
}
s`
		v := runScript(t, code)
		if v.Int() != 21 {
			t.Errorf("结果 %d, 期望 21", v.Int())
		}
	})
}

// Test_NestedSameNameDynamicFor 动态for同名变量嵌套不共用计数器槽
func Test_NestedSameNameDynamicFor(t *testing.T) {
	t.Parallel()
	code := `n1 := 2
n2 := 3
c := 0
for i := n1 {
	for i := n2 {
		c = c + 1
	}
}
c`
	v, err := runScriptWithMaxSteps(t, code, 100000)
	if err != nil {
		t.Fatalf("执行失败(疑似死循环): %v", err)
	}
	if v.Int() != 6 {
		t.Errorf("结果 %d, 期望 6", v.Int())
	}
}

// Test_NestedSameNameRangeFor range循环同名变量嵌套行为正确
func Test_NestedSameNameRangeFor(t *testing.T) {
	t.Parallel()
	code := `c := 0
for i := [1, 2, 3] {
	for i := [1] {
		c = c + 1
	}
}
c`
	v, err := runScriptWithMaxSteps(t, code, 100000)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if v.Int() != 3 {
		t.Errorf("结果 %d, 期望 3", v.Int())
	}
}

// Test_NestedFuncState 嵌套函数定义后编译器状态完整恢复
func Test_NestedFuncState(t *testing.T) {
	t.Parallel()
	t.Run("嵌套函数引用外部函数", func(t *testing.T) {
		code := `fn helper(x) {
	return x + 1
}
fn outer(y) {
	fn inner(z) {
		return helper(z) * 2
	}
	return inner(y)
}
outer(5)`
		v := runScript(t, code)
		if v.Int() != 12 {
			t.Errorf("结果 %d, 期望 12", v.Int())
		}
	})

	t.Run("嵌套函数多次调用", func(t *testing.T) {
		code := `fn outer(y) {
	fn inner(z) {
		return z * 2
	}
	return inner(y) + inner(1)
}
outer(5)`
		v := runScript(t, code)
		if v.Int() != 12 {
			t.Errorf("结果 %d, 期望 12", v.Int())
		}
	})

	t.Run("嵌套函数与切片混用", func(t *testing.T) {
		code := `fn outer(arr) {
	fn inner(x) {
		return x[1:3]
	}
	s := arr[0:2]
	return inner(arr)[0] + s[1]
}
outer([1, 2, 3, 4])`
		v := runScript(t, code)
		if v.Int() != 4 {
			t.Errorf("结果 %d, 期望 4", v.Int())
		}
	})
}

// Test_DefaultStepLimit Eval 默认具备步数保护, 死循环不会无限阻塞
func Test_DefaultStepLimit(t *testing.T) {
	t.Run("默认上限触发", func(t *testing.T) {
		done := make(chan error, 1)
		go func() {
			_, err := Eval(`x := 0
for true {
	x = x + 1
}
x`)
			done <- err
		}()
		select {
		case err := <-done:
			if err == nil {
				t.Fatal("死循环应被默认步数上限终止")
			}
			if !errors.Is(err, &RuntimeError{Code: ErrStepLimit}) {
				t.Errorf("错误码应为 ErrStepLimit, 实际: %v", err)
			}
		case <-time.After(60 * time.Second):
			t.Fatal("默认配置下死循环未被终止")
		}
	})

	t.Run("上限0表示无限制", func(t *testing.T) {
		v, err := runScriptWithMaxSteps(t, `1 + 2`, 0)
		if err != nil {
			t.Fatalf("正常脚本不应报错: %v", err)
		}
		if v.Int() != 3 {
			t.Errorf("结果 %d, 期望 3", v.Int())
		}
	})
}

// Test_TypeAnnotGetBindValue :=> 声明对 getBindValue 返回值做类型校验
func Test_TypeAnnotGetBindValue(t *testing.T) {
	t.Run("类型匹配成功", func(t *testing.T) {
		v := runScriptWithBinds(t, "x :=>int getBindValue(\"n\")\nx", map[string]interface{}{"n": 5})
		if v.Int() != 5 {
			t.Errorf("结果 %d, 期望 5", v.Int())
		}
	})

	t.Run("类型不匹配报错", func(t *testing.T) {
		parser := NewParser()
		compiled, err := parser.Compile("x :=>int getBindValue(\"n\")\nx")
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		ctx := NewContext()
		ctx.BindValue("n", "not an int")
		_, err = NewEngine().Run(ctx, compiled)
		if err == nil {
			t.Fatal("int 注解绑定 string 值应报类型错误")
		}
		if !strings.Contains(err.Error(), "int") {
			t.Errorf("错误信息应包含注解类型, 实际: %v", err)
		}
	})

	t.Run("any注解不校验", func(t *testing.T) {
		v := runScriptWithBinds(t, "x :=>any getBindValue(\"n\")\nx", map[string]interface{}{"n": "str"})
		if v.String() != "str" {
			t.Errorf("结果 %q, 期望 str", v.String())
		}
	})

	t.Run("非法注解类型编译报错", func(t *testing.T) {
		_, err := Eval(`x :=>foo getBindValue("n")`)
		if err == nil {
			t.Fatal("未知注解类型应编译报错")
		}
	})

	t.Run("非声明用法仍编译报错", func(t *testing.T) {
		_, err := Eval(`x := getBindValue("n")`)
		if err == nil {
			t.Fatal("getBindValue 未在 :=> 中使用应编译报错")
		}
	})
}

// Test_P2Misc P2 六项缺陷修复验证
func Test_P2Misc(t *testing.T) {
	t.Parallel()

	t.Run("切片共享底层数组", func(t *testing.T) {
		code := `a := [1, 2, 3]
s := a[0:2]
s[0] = 99
a[0]`
		v := runScript(t, code)
		if v.Int() != 99 {
			t.Errorf("结果 %d, 期望 99(切片共享底层数组)", v.Int())
		}
	})

	t.Run("EncodeDecode返回明确错误", func(t *testing.T) {
		s := NewScript(nil)
		_, err := s.Encode()
		if err == nil {
			t.Error("Encode 未实现应返回错误而非静默成功")
		}
		_, err = s.Decode(nil)
		if err == nil {
			t.Error("Decode 未实现应返回错误而非静默成功")
		}
	})

	t.Run("Clone浅拷贝共享编译产物", func(t *testing.T) {
		compiled, err := NewParser().Compile(`1 + 2`)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		s := NewScript(compiled)
		clone := s.Clone()
		if clone.GetCompiled() != s.GetCompiled() {
			t.Error("Clone 应与原脚本共享同一编译产物")
		}
		v, err := NewEngine().Run(NewContext(), clone.GetCompiled())
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if v.Int() != 3 {
			t.Errorf("结果 %d, 期望 3", v.Int())
		}
	})

	t.Run("d与f开头注释不被当指令", func(t *testing.T) {
		code := `# debug 目的注释
# fn 字样注释
1 + 2`
		parser := NewParser()
		compiled, err := parser.Compile(code)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		if len(compiled.Externals) != 0 {
			t.Errorf("注释不应产生外部函数声明, 实际 %d 个", len(compiled.Externals))
		}
	})

	t.Run("BindFunc非函数返回错误", func(t *testing.T) {
		ctx := NewContext()
		if err := ctx.BindFunc("f", nil); err == nil {
			t.Error("BindFunc(nil) 应返回错误")
		}
		if err := ctx.BindFunc("f", "not func"); err == nil {
			t.Error("BindFunc(字符串) 应返回错误")
		}
		if err := ctx.BindFunc("f", func() {}); err != nil {
			t.Errorf("合法函数不应报错: %v", err)
		}
	})

	t.Run("前导零整数编译报错", func(t *testing.T) {
		_, err := Eval(`x := 0123`)
		if err == nil {
			t.Fatal("前导零整数应编译报错而非按八进制解析")
		}
		v := runScript(t, `0xFF + 10`)
		if v.Int() != 265 {
			t.Errorf("十六进制 0xFF+10 = %d, 期望 265", v.Int())
		}
	})

	t.Run("跨行字符串更新行号", func(t *testing.T) {
		code := "a := \"line1\nline2\"\n1 +"
		_, err := NewParser().Compile(code)
		if err == nil {
			t.Fatal("残缺脚本应产生解析错误")
		}
		ce, ok := err.(*CompileError)
		if !ok {
			t.Fatalf("错误类型应为 CompileError, 实际: %T", err)
		}
		if ce.Line != 3 {
			t.Errorf("错误行号应为 3(跨行字符串后), 实际: %d", ce.Line)
		}
	})
}
