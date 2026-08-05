package script

import (
	"fmt"
	"testing"
)

// ========== 表达式求值与集成测试 ==========
// 参照 expr-lang 和 tengo 的测试模式

// ========== A. 表达式求值模式 ==========

// --- A1. 算术表达式树 ---

func Test_Expr_DeepNestedArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"((((1+2)*3)-4)/2)", 2},
		{"((1+2)*(3+4))/(5-2)", 7},
		{"((((10-3)*2)+6)/4)", 5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_MixedPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1+2*3-4/2+5*3", 20},
		{"2*3+4*5-6/2", 23},
		{"100/10+3*4-2*5", 12},
		{"1+1+1+1+1*0", 4},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_ParenthesesOverride(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"(1+2)*(3+4)", 21},
		{"((1+2)*(3+4))/(5-2)", 7},
		{"(10-2)*(10-2)", 64},
		{"(2+3)*(4+5)*(1+0)", 45},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_UnaryNegInArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"-5+3", -2},
		{"3+(-5)", -2},
		{"(-5)*(-3)", 15},
		{"-(2+3)", -5},
		{"-(-5)", 5},
		{"-(2*3)+1", -5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_NestedModulo(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"17%5%2", 0},
		{"17%5", 2},
		{"(17%5)%2", 0},
		{"100%7%3", 2},
		{"20%3*4", 8},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// --- A2. 逻辑表达式 ---

func Test_Expr_ComplexBoolean(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"(true && false) || (true && true)", true},
		{"(true && false) || (true && false)", false},
		{"(1 > 0 && 2 > 0) || (3 > 0 && 4 > 0)", true},
		{"(1 < 0 && 2 > 0) || (3 < 0 && 4 > 0)", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_DeMorgansLaw(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!(false||false)", true},
		{"!false && !false", true},
		{"!(true||false)", false},
		{"!true && !false", false},
		{"!(1>0 || 2>0)", false},
		{"!(1>0) && !(2>0)", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_ShortCircuitEvaluation(t *testing.T) {
	runBoolTest(t, "false && (1 > 0)", false)
	runBoolTest(t, "true || (1 < 0)", true)
	runBoolTest(t, "(1 > 2) && (3 > 0)", false)
	runBoolTest(t, "(1 < 2) || (3 > 4)", true)
}

func Test_Expr_NestedLogical(t *testing.T) {
	runBoolTest(t, "(1>0) && ((2>1) || (3>2))", true)
	runBoolTest(t, "(1<0) || ((2>1) && (3<2))", false)
	runBoolTest(t, "(1==1) && (2==2) && (3==3)", true)
	runBoolTest(t, "(1==1) && (2!=2) || (3==3)", true)
}

// --- A3. 比较链 ---

func Test_Expr_ComparisonWithArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1+1 == 2 && 3*3 == 9", true},
		{"2*3 != 7 && 10/2 == 5", true},
		{"1+1 == 3 || 2*2 == 4", true},
		{"5-1 == 3 && 2+2 == 5", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_RangeCheckPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"50 >= 0 && 50 < 100", true},
		{"100 >= 0 && 100 < 100", false},
		{"0 >= 0 && 0 < 100", true},
		{"99 >= 0 && 99 < 100", true},
		{"100 >= 0 && 101 < 100", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_InequalityCheck(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 != 0 && 1 != 0", true},
		{"0 != 0 && 1 != 0", false},
		{"5 != 3 && 7 != 0", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// --- A4. 位运算组合 ---

func Test_Expr_BitwiseWithArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1 + 2 << 1", 6},
		{"1 << 2 + 1", 8},
		{"2 * 3 | 1", 7},
		{"4 & 3 + 1", 4},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_BitwiseChain(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0xFF & 0x0F", 15},
		{"0xF0 | 0x0F", 255},
		{"0xFF & 0x0F | 0xF0", 255},
		{"0xAA ^ 0xFF", 85},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_BitwiseShiftAndMask(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"(1 << 4) & 0x10", 16},
		{"(0xFF >> 4) & 0x0F", 15},
		{"1 << 8", 256},
		{"256 >> 4", 16},
		{"(0xAA & 0x0F) << 4", 160},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Expr_BitwiseNotExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"^0", -1},
		{"^5", -6},
		{"3 + ^5", -3},
		{"^0 & 0xFF", 255},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// ========== B. 真实场景集成测试 ==========

// --- B5. 数值算法 ---

func Test_Integration_BubbleSort(t *testing.T) {
	result := runScript(t, `
		arr := [5, 3, 8, 1, 9, 2, 7]
		n := len(arr)
		for i := 0; i < n; i = i + 1 {
			for j := 0; j < n - 1 - i; j = j + 1 {
				if arr[j] > arr[j + 1] {
					tmp := arr[j]
					arr[j] = arr[j + 1]
					arr[j + 1] = tmp
				}
			}
		}
		arr[0] * 100000 + arr[1] * 10000 + arr[2] * 1000 + arr[3] * 100 + arr[4] * 10 + arr[5]
	`)
	expected := 1*100000 + 2*10000 + 3*1000 + 5*100 + 7*10 + 8
	if result.Int() != expected {
		t.Errorf("bubble sort: got %d, want %d", result.Int(), expected)
	}
}

func Test_Integration_FindMaxInArray(t *testing.T) {
	runIntTest(t, `
		arr := [3, 7, 2, 9, 4, 8, 1]
		maxVal := arr[0]
		for i := 1; i < len(arr); i = i + 1 {
			if arr[i] > maxVal {
				maxVal = arr[i]
			}
		}
		maxVal
	`, 9)
}

func Test_Integration_ArraySum(t *testing.T) {
	runIntTest(t, `
		arr := [10, 20, 30, 40, 50]
		sum := 0
		for v := range arr {
			sum = sum + v
		}
		sum
	`, 150)
}

func Test_Integration_FactorialLoop(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 1},
		{1, 1},
		{5, 120},
		{10, 3628800},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("factorial(%d)", tt.n)
		t.Run(name, func(t *testing.T) {
			input := fmt.Sprintf(`
				n := %d
				result := 1
				for i := 1; i <= n; i = i + 1 {
					result = result * i
				}
				result
			`, tt.n)
			runIntTest(t, input, tt.expected)
		})
	}
}

func Test_Integration_FibonacciLoop(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{5, 5},
		{10, 55},
		{15, 610},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("fib(%d)", tt.n)
		t.Run(name, func(t *testing.T) {
			input := fmt.Sprintf(`
				n := %d
				if n <= 1 {
					n
				} else {
					a := 0
					b := 1
					for i := 2; i <= n; i = i + 1 {
						c := a + b
						a = b
						b = c
					}
					b
				}
			`, tt.n)
			runIntTest(t, input, tt.expected)
		})
	}
}

// --- B6. 字符串处理 ---

func Test_Integration_StringConcat(t *testing.T) {
	runStringTest(t, `"hello" + " " + "world"`, "hello world")
	runStringTest(t, `"a" + "b" + "c" + "d" + "e"`, "abcde")
}

func Test_Integration_StringBuildInLoop(t *testing.T) {
	result := runScript(t, `
		s := ""
		arr := ["a", "b", "c", "d"]
		for v := range arr {
			s = s + v
		}
		s
	`)
	if result.String() != "abcd" {
		t.Errorf("got %s, want abcd", result.String())
	}
}

func Test_Integration_StringLength(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{`len("hello")`, 5},
		{`len("")`, 0},
		{`len("hello world")`, 11},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Integration_StringSlice(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"[0:3]`, "hel"},
		{`"hello"[1:4]`, "ell"},
		{`"hello"[2:5]`, "llo"},
		{`"hello world"[0:5]`, "hello"},
		{`"hello world"[6:11]`, "world"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}
}

// --- B7. 条件逻辑 ---

func Test_Integration_Grade(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{95, "A"},
		{85, "B"},
		{75, "C"},
		{65, "D"},
		{55, "F"},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("score_%d", tt.score)
		t.Run(name, func(t *testing.T) {
			input := fmt.Sprintf(`
				score := %d
				grade := "F"
				if score >= 90 {
					grade = "A"
				} else if score >= 80 {
					grade = "B"
				} else if score >= 70 {
					grade = "C"
				} else if score >= 60 {
					grade = "D"
				} else {
					grade = "F"
				}
				grade
			`, tt.score)
			runStringTest(t, input, tt.expected)
		})
	}
}

func Test_Integration_LeapYear(t *testing.T) {
	tests := []struct {
		year     int
		expected bool
	}{
		{2000, true},
		{1900, false},
		{2004, true},
		{2001, false},
		{2020, true},
		{2023, false},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("year_%d", tt.year)
		t.Run(name, func(t *testing.T) {
			input := fmt.Sprintf("%d %% 4 == 0 && (%d %% 100 != 0 || %d %% 400 == 0)", tt.year, tt.year, tt.year)
			runBoolTest(t, input, tt.expected)
		})
	}
}

func Test_Integration_OddEven(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{2, true},
		{3, false},
		{0, true},
		{100, true},
		{101, false},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("n_%d_even", tt.n)
		t.Run(name, func(t *testing.T) {
			input := fmt.Sprintf("%d %% 2 == 0", tt.n)
			runBoolTest(t, input, tt.expected)
		})
	}
}

func Test_Integration_ThreeMax(t *testing.T) {
	runIntTest(t, `
		a := 10
		b := 25
		c := 15
		max := a
		if b > max {
			max = b
		}
		if c > max {
			max = c
		}
		max
	`, 25)

	runIntTest(t, `
		a := 30
		b := 25
		c := 15
		max := a
		if b > max {
			max = b
		}
		if c > max {
			max = c
		}
		max
	`, 30)

	runIntTest(t, `
		a := 10
		b := 25
		c := 50
		max := a
		if b > max {
			max = b
		}
		if c > max {
			max = c
		}
		max
	`, 50)
}

// --- B8. 数据转换 ---

func Test_Integration_IntStringRoundTrip(t *testing.T) {
	runIntTest(t, `int(string(42))`, 42)
	runIntTest(t, `int(string(-100))`, -100)
	runIntTest(t, `int(string(0))`, 0)
}

func Test_Integration_FloatTruncation(t *testing.T) {
	runIntTest(t, `int(3.7)`, 3)
	runIntTest(t, `int(3.2)`, 3)
	runIntTest(t, `int(-3.7)`, -3)
}

func Test_Integration_TypeCheckThenAct(t *testing.T) {
	result := runScript(t, `
		v := 42
		t := typeof(v)
		if t == "int" {
			v * 2
		} else {
			0
		}
	`)
	if result.Int() != 84 {
		t.Errorf("got %d, want 84", result.Int())
	}

	result = runScript(t, `
		v := "hello"
		t := typeof(v)
		if t == "string" {
			len(v)
		} else {
			0
		}
	`)
	if result.Int() != 5 {
		t.Errorf("got %d, want 5", result.Int())
	}
}

// --- B9. 外部函数集成 ---

func Test_Integration_ExternalFuncCompute(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("square", func(x int) int { return x * x })

	parser := NewParser()
	engine := NewEngine()
	compiled, err := parser.Compile(`
		#fn square(int)=>int
		square(7) + square(3)
	`)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Int() != 58 {
		t.Errorf("got %d, want 58", result.Int())
	}
}

func Test_Integration_ExternalFuncInLoop(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("double", func(x int) int { return x * 2 })

	parser := NewParser()
	engine := NewEngine()
	compiled, err := parser.Compile(`
		#fn double(int)=>int
		sum := 0
		for i := 1; i <= 5; i = i + 1 {
			sum = sum + double(i)
		}
		sum
	`)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Int() != 30 {
		t.Errorf("got %d, want 30", result.Int())
	}
}

func Test_Integration_ExternalFuncWithScriptFunc(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("pow", func(base, exp int) int {
		result := 1
		for i := 0; i < exp; i++ {
			result *= base
		}
		return result
	})

	parser := NewParser()
	engine := NewEngine()
	compiled, err := parser.Compile(`
		#fn pow(int, int)=>int
		fn computeArea(side) {
			return pow(side, 2)
		}
		computeArea(5) + computeArea(3)
	`)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Int() != 34 {
		t.Errorf("got %d, want 34", result.Int())
	}
}

func Test_Integration_ExternalFuncReturnTypes(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("greet", func(name string) string { return "Hello, " + name })

		parser := NewParser()
		engine := NewEngine()
		compiled, err := parser.Compile(`
			#fn greet(string)=>string
			greet("World")
		`)
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		result, err := engine.Run(ctx, compiled)
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if result.String() != "Hello, World" {
			t.Errorf("got %s, want Hello, World", result.String())
		}
	})

	t.Run("bool", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("isPositive", func(n int) bool { return n > 0 })

		parser := NewParser()
		engine := NewEngine()
		compiled, err := parser.Compile(`
			#fn isPositive(int)=>bool
			isPositive(5) && !isPositive(-3)
		`)
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		result, err := engine.Run(ctx, compiled)
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if !result.Bool() {
			t.Errorf("got %v, want true", result.Bool())
		}
	})

	t.Run("float", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("half", func(n int) float64 { return float64(n) / 2.0 })

		parser := NewParser()
		engine := NewEngine()
		compiled, err := parser.Compile(`
			#fn half(int)=>float
			half(10)
		`)
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		result, err := engine.Run(ctx, compiled)
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if result.Float() != 5.0 {
			t.Errorf("got %f, want 5.0", result.Float())
		}
	})
}

func Test_Integration_ExternalFuncArraySum(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("sumArray", func(arr []int) int {
		sum := 0
		for _, v := range arr {
			sum += v
		}
		return sum
	})

	parser := NewParser()
	engine := NewEngine()
	compiled, err := parser.Compile(`
		#fn sumArray(array)=>int
		arr := [1, 2, 3, 4, 5]
		sumArray(arr)
	`)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Int() != 15 {
		t.Errorf("got %d, want 15", result.Int())
	}
}
