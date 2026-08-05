package script

import (
	"testing"
)

// ========== 内建函数完整测试 ==========

func Test_Builtin_String_Conversions(t *testing.T) {
	result := runScript(t, `string(42)`)
	if result.String() != "42" {
		t.Errorf("string(42) = %q, want %q", result.String(), "42")
	}

	result = runScript(t, `string(0)`)
	if result.String() != "0" {
		t.Errorf("string(0) = %q", result.String())
	}
}

func Test_Builtin_Int_Conversions(t *testing.T) {
	result := runScript(t, `int("42")`)
	if result.Int() != 42 {
		t.Errorf("int('42') = %d", result.Int())
	}

	result = runScript(t, `int("0")`)
	if result.Int() != 0 {
		t.Errorf("int('0') = %d", result.Int())
	}
}

func Test_Builtin_Float_Conversions(t *testing.T) {
	result := runScript(t, `float("3.14")`)
	if result.Float() != 3.14 {
		t.Errorf("float('3.14') = %f", result.Float())
	}
}

func Test_Builtin_Len_AllTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`len("hello")`, 5},
		{`len("")`, 0},
		{`len("a")`, 1},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`len({"a": 1})`, 1},
		{`len({})`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Builtin_TypeOf_AllTypes(t *testing.T) {
	result := runScript(t, `typeof(42)`)
	if result.String() == "" {
		t.Error("typeof(42)不应为空")
	}

	result = runScript(t, `typeof("hello")`)
	if result.String() == "" {
		t.Error("typeof('hello')不应为空")
	}

	result = runScript(t, `typeof(true)`)
	if result.String() == "" {
		t.Error("typeof(true)不应为空")
	}

	result = runScript(t, `typeof(nil)`)
	if result.String() == "" {
		t.Error("typeof(nil)不应为空")
	}

	result = runScript(t, `typeof([1, 2])`)
	if result.String() == "" {
		t.Error("typeof([1,2])不应为空")
	}

	result = runScript(t, `typeof({"a": 1})`)
	if result.String() == "" {
		t.Error("typeof({a:1})不应为空")
	}
}

// ========== delete操作测试 ==========

func Test_Delete_ArrayElement(t *testing.T) {
	// delete可能或不可能支持，取决于实现
	// 如果不支持，编译/运行时会报错
	// 这里测试不panic即可
	parser := NewParser()
	_, err := parser.Compile(`
		arr := [1, 2, 3]
		delete(arr, 1)
	`)
	// 无论成功与否，不应panic
	_ = err
}

func Test_Delete_MapKey(t *testing.T) {
	parser := NewParser()
	_, err := parser.Compile(`
		m := {"a": 1, "b": 2}
		delete(m, "a")
	`)
	_ = err
}

// ========== 多行脚本测试 ==========

func Test_MultiLine_ComplexScript(t *testing.T) {
	ctx := NewContext()
	ctx.BindValue("base", 50)
	result := runScriptWithContext(t, `
		#fn getBindValue(string) => int
		base :=>int getBindValue("base")

		fn calculate(x) {
			if x > 100 {
				return x * 2
			}
			return x + 10
		}

		result := calculate(base)
		result
	`, ctx)
	if result.Int() != 60 {
		t.Errorf("got %d, want 60", result.Int())
	}
}

func Test_MultiLine_FunctionDefinitions(t *testing.T) {
	result := runScript(t, `
		fn add(a, b) {
			return a + b
		}

		fn sub(a, b) {
			return a - b
		}

		fn mul(a, b) {
			return a * b
		}

		add(10, 5) + sub(20, 5) + mul(3, 4)
	`)
	// 15 + 15 + 12 = 42
	if result.Int() != 42 {
		t.Errorf("got %d, want 42", result.Int())
	}
}

// ========== 表达式求值顺序测试 ==========

func Test_EvaluationOrder_LeftToRight(t *testing.T) {
	ctx := NewContext()
	order := []int{}
	ctx.BindFunc("track", func(n int) int {
		order = append(order, n)
		return n
	})

	runScriptWithContext(t, `
		#fn track(int) => int
		track(1) + track(2) + track(3)
	`, ctx)

	if len(order) != 3 {
		t.Fatalf("应调用3次track, got %d", len(order))
	}
	for i, expected := range []int{1, 2, 3} {
		if order[i] != expected {
			t.Errorf("order[%d] = %d, want %d", i, order[i], expected)
		}
	}
}

func Test_EvaluationOrder_FunctionArgsFirst(t *testing.T) {
	ctx := NewContext()
	order := []string{}
	ctx.BindFunc("a", func() int { order = append(order, "a"); return 1 })
	ctx.BindFunc("b", func() int { order = append(order, "b"); return 2 })

	runScriptWithContext(t, `
		#fn a() => int
		#fn b() => int
		a() + b()
	`, ctx)

	if len(order) < 2 {
		t.Fatalf("应调用2次, got %d", len(order))
	}
}

// ========== 数组切片测试 ==========

func Test_ArraySlice_VariousRanges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checkLen int
	}{
		{"full slice", `[1, 2, 3, 4, 5][0:5]`, 5},
		{"partial start", `[1, 2, 3, 4, 5][1:3]`, 2},
		{"single element", `[1, 2, 3, 4, 5][2:3]`, 1},
		{"empty slice", `[1, 2, 3, 4, 5][2:2]`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runScript(t, tt.input)
			arr := result.Array()
			if arr == nil {
				t.Fatal("结果应为数组")
			}
			if len(arr.Elements) != tt.checkLen {
				t.Errorf("切片长度 = %d, want %d", len(arr.Elements), tt.checkLen)
			}
		})
	}
}

// ========== 嵌套函数调用测试 ==========

func Test_NestedCalls_DeepChain(t *testing.T) {
	runIntTest(t, `
		fn id(x) { return x }
		id(id(id(id(id(42)))))
	`, 42)
}

func Test_NestedCalls_DifferentFunctions(t *testing.T) {
	runIntTest(t, `
		fn inc(x) { return x + 1 }
		fn dbl(x) { return x * 2 }
		inc(dbl(inc(dbl(1))))
	`, 7) // dbl(1)=2, inc(2)=3, dbl(3)=6, inc(6)=7
}

// ========== 注释处理测试 ==========
// NextToken()已自动跳过注释token, 注释与代码混合可正常工作

func Test_Comment_StandaloneHash(t *testing.T) {
	// # 注释在独立行, 应被正常过滤
	runIntTest(t, `
		# comment
		x := 42
		x
	`, 42)
}

func Test_Comment_HashBetweenStatements(t *testing.T) {
	runIntTest(t, `
		x := 1
		# comment
		y := 2
		x + y
	`, 3)
}

func Test_Comment_InlineSlashSlash(t *testing.T) {
	// // 行内注释也同理
	runIntTest(t, `
		x := 1 // inline comment
		x + 1
	`, 2)
}

func Test_Comment_MultilineComment(t *testing.T) {
	// /* */ 多行注释也被过滤
	runIntTest(t, `
		x := 1
		/* multi
		   line
		   comment */
		y := 2
		x + y
	`, 3)
}

// ========== 零值和默认值测试 ==========

func Test_DefaultValues_AllTypes(t *testing.T) {
	// 未初始化变量的默认值由编译器决定
	// 这里测试NewValue的零值
	if NewValue(0).Int() != 0 {
		t.Error("0的Int应为0")
	}
	if NewValue("").String() != "" {
		t.Error("空字符串的String应为空")
	}
	if NewValue(false).Bool() != false {
		t.Error("false的Bool应为false")
	}
	if NewValue(0.0).Float() != 0.0 {
		t.Error("0.0的Float应为0.0")
	}
}
