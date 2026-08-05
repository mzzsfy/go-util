package script

import (
	"testing"
)

// ========== VM类型强制转换测试 ==========

func Test_VM_IntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1 + 2", 3},
		{"10 - 3", 7},
		{"3 * 4", 12},
		{"20 / 4", 5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_FloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1.5 + 2.5", 4.0},
		{"5.5 - 2.5", 3.0},
		{"2.5 * 4.0", 10.0},
		{"10.0 / 4.0", 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runFloatTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_FloatDivision(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"10 / 3", 3}, // 整数除法
		{"7 / 2", 3},  // 整数除法截断
		{"100 / 7", 14},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_Modulo(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"10 % 3", 1},
		{"7 % 2", 1},
		{"8 % 4", 0},
		{"15 % 7", 1},
		{"100 % 9", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_BitwiseOperations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"AND", "12 & 10", 8},
		{"OR", "12 | 10", 14},
		{"XOR", "12 ^ 10", 6},
		{"左移", "1 << 4", 16},
		{"右移", "256 >> 2", 64},
		{"位反", "^5", -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_BitwiseComplex(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"(0xFF & 0x0F)", 15},
		{"(0xF0 | 0x0F)", 255},
		{"(0xFF ^ 0x0F)", 240},
		{"(1 << 8)", 256},
		{"(1024 >> 3)", 128},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_NegativeNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"-5", -5},
		{"-5 + 3", -2},
		{"3 + (-5)", -2},
		{"-10 * -3", 30},
		{"-(-5)", 5},
		{"2 * -3", -6},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_StringConcatenation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"a" + "b"`, "ab"},
		{`"hello" + " " + "world"`, "hello world"},
		{`"" + "x"`, "x"},
		{`"x" + ""`, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_ComparisonAllTypes(t *testing.T) {
	intTests := []struct {
		input    string
		expected bool
	}{
		{"1 < 2", true},
		{"2 < 1", false},
		{"2 <= 2", true},
		{"3 <= 2", false},
		{"3 > 2", true},
		{"2 > 3", false},
		{"2 >= 2", true},
		{"1 >= 3", false},
	}

	for _, tt := range intTests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_EqualityAllTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 == 1", true},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 != 1", false},
		{`"a" == "a"`, true},
		{`"a" == "b"`, false},
		{"true == true", true},
		{"true == false", false},
		{"nil == nil", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_LogicalAnd(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"1 > 0 && 2 > 1", true},
		{"1 > 0 && 2 < 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_LogicalOr(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{"1 > 2 || 2 > 1", true},
		{"1 > 2 || 3 > 2", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_NotOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!(1 > 2)", true},
		{"!(1 < 2)", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_ShortCircuit(t *testing.T) {
	ctx := NewContext()
	called := false
	ctx.BindFunc("check", func() bool {
		called = true
		return true
	})

	runScriptWithContext(t, `
		#fn check() => bool
		false && check()
	`, ctx)
	if called {
		t.Error("&& 短路应阻止第二个操作数求值")
	}

	called = false
	runScriptWithContext(t, `
		#fn check() => bool
		true || check()
	`, ctx)
	if called {
		t.Error("|| 短路应阻止第二个操作数求值")
	}
}

// ========== VM数组操作测试 ==========

func Test_VM_ArrayCreate(t *testing.T) {
	result := runScript(t, `[1, 2, 3]`)
	arr := result.Array()
	if arr == nil {
		t.Fatal("结果应为数组")
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("数组长度应为3, got %d", len(arr.Elements))
	}
}

func Test_VM_ArrayAccess(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"[10, 20, 30][0]", 10},
		{"[10, 20, 30][1]", 20},
		{"[10, 20, 30][2]", 30},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_ArrayLength(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"len([1, 2, 3])", 3},
		{"len([])", 0},
		{"len([1])", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_VM_ArraySlice(t *testing.T) {
	result := runScript(t, `[1, 2, 3, 4, 5][1:3]`)
	arr := result.Array()
	if arr == nil {
		t.Fatal("切片结果应为数组")
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("切片长度应为2, got %d", len(arr.Elements))
	}
	if arr.Elements[0].Int() != 2 {
		t.Errorf("切片第一个元素应为2, got %d", arr.Elements[0].Int())
	}
}

func Test_VM_ArrayStore(t *testing.T) {
	result := runScript(t, `
		arr := [1, 2, 3]
		arr[0] = 99
		x := arr[0]
		x
	`)
	if result.Int() != 99 {
		t.Errorf("数组赋值后应返回99, got %d", result.Int())
	}
}

// ========== VM Map操作测试 ==========

func Test_VM_MapCreate(t *testing.T) {
	result := runScript(t, `{"a": 1, "b": 2}`)
	m := result.Map()
	if m == nil {
		t.Fatal("结果应为map")
	}
	if len(m.Pairs) != 2 {
		t.Fatalf("map大小应为2, got %d", len(m.Pairs))
	}
}

func Test_VM_MapAccess(t *testing.T) {
	runIntTest(t, `{"a": 1, "b": 2}["a"]`, 1)
	runIntTest(t, `{"a": 1, "b": 2}["b"]`, 2)
}

func Test_VM_MapLength(t *testing.T) {
	runIntTest(t, `len({"a": 1, "b": 2, "c": 3})`, 3)
}

func Test_VM_NestedDataAccess(t *testing.T) {
	result := runScript(t, `
		data := {"users": [{"name": "alice"}, {"name": "bob"}]}
		data["users"][0]["name"]
	`)
	if result.String() != "alice" {
		t.Errorf("嵌套访问应返回alice, got %q", result.String())
	}
}

// ========== VM条件表达式测试 ==========

func Test_VM_NestedIf(t *testing.T) {
	result := runScript(t, `
		x := 15
		result := 0
		if x > 10 {
			if x > 20 {
				result = 1
			} else {
				result = 2
			}
		} else {
			result = 3
		}
		result
	`)
	if result.Int() != 2 {
		t.Errorf("嵌套if应返回2, got %d", result.Int())
	}
}

func Test_VM_IfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"if true { 1 } else { 2 }", 1},
		{"if false { 1 } else { 2 }", 2},
		{"if 1 > 0 { 10 } else { 20 }", 10},
		{"if 1 < 0 { 10 } else { 20 }", 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// ========== VM函数测试 ==========

func Test_VM_FunctionReturn(t *testing.T) {
	runIntTest(t, `
		fn double(x) {
			return x * 2
		}
		double(21)
	`, 42)
}

func Test_VM_FunctionMultipleArgs(t *testing.T) {
	runIntTest(t, `
		fn add(a, b, c) {
			return a + b + c
		}
		add(1, 2, 3)
	`, 6)
}

func Test_VM_FunctionNoArgs(t *testing.T) {
	runIntTest(t, `
		fn answer() {
			return 42
		}
		answer()
	`, 42)
}

func Test_VM_FunctionNestedCall(t *testing.T) {
	runIntTest(t, `
		fn inc(x) {
			return x + 1
		}
		fn double(x) {
			return x * 2
		}
		inc(double(5))
	`, 11)
}

func Test_VM_FunctionParamPassing(t *testing.T) {
	runIntTest(t, `
		fn add(a, b) {
			return a + b
		}
		fn apply(x, y) {
			return add(x, y)
		}
		apply(10, 20)
	`, 30)
}

// ========== VM循环测试 ==========

func Test_VM_ForLoop(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for i := 0; i < 5; i = i + 1 {
			sum = sum + i
		}
		sum
	`)
	// 0+1+2+3+4 = 10
	if result.Int() != 10 {
		t.Errorf("for循环求和应为10, got %d", result.Int())
	}
}

func Test_VM_ForBreak(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for i := 0; i < 100; i = i + 1 {
			if i >= 3 {
				break
			}
			sum = sum + i
		}
		sum
	`)
	// 0+1+2 = 3
	if result.Int() != 3 {
		t.Errorf("break后求和应为3, got %d", result.Int())
	}
}

func Test_VM_ForContinue(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for i := 0; i < 5; i = i + 1 {
			if i == 2 {
				continue
			}
			sum = sum + i
		}
		sum
	`)
	// 0+1+3+4 = 8
	if result.Int() != 8 {
		t.Errorf("continue后求和应为8, got %d", result.Int())
	}
}

func Test_VM_NestedLoop(t *testing.T) {
	result := runScript(t, `
		total := 0
		for i := 0; i < 3; i = i + 1 {
			for j := 0; j < 3; j = j + 1 {
				total = total + 1
			}
		}
		total
	`)
	if result.Int() != 9 {
		t.Errorf("嵌套循环应为9, got %d", result.Int())
	}
}

func Test_VM_RangeLoop(t *testing.T) {
	result := runScript(t, `
		sum := 0
		for v := range [1, 2, 3, 4, 5] {
			sum = sum + v
		}
		sum
	`)
	// 1+2+3+4+5 = 15
	if result.Int() != 15 {
		t.Errorf("range求和应为15, got %d", result.Int())
	}
}

// ========== VM变量赋值测试 ==========

func Test_VM_VariableReassign(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"x := 10\nx = 20\nx", 20},
		{"x := 10\nx = x + 5\nx", 15},
		{"x := 10\nx = x * 2\nx", 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}
