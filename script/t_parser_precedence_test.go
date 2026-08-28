package script

import (
	"testing"
)

// ========== Parser运算符优先级测试 ==========

func Test_Parser_Precedence_AddSub(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1 + 2 + 3", 6},
		{"1 + 2 - 3", 0},
		{"10 - 3 - 2", 5},
		{"1 - 2 + 3", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Parser_Precedence_MulDiv(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"2 * 3 * 4", 24},
		{"12 / 3 / 2", 2},
		{"2 * 3 / 2", 3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Parser_Precedence_MixedAddMul(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1 + 2 * 3", 7},
		{"2 * 3 + 1", 7},
		{"2 + 3 * 4 + 5", 19},
		{"10 - 2 * 3", 4},
		{"2 * 3 + 4 * 5", 26},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Parser_Precedence_Parentheses(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"(1 + 2) * 3", 9},
		{"(2 + 3) * (4 + 5)", 45},
		{"((1 + 2) + 3) + 4", 10},
		{"2 * (3 + 4)", 14},
		{"(10 - 2) * (10 - 2)", 64},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Parser_Precedence_Comparison(t *testing.T) {
	// 比较运算符优先级低于算术
	runBoolTest(t, "1 + 1 > 1", true)
	runBoolTest(t, "1 + 1 < 3", true)
	runBoolTest(t, "2 * 3 == 6", true)
	runBoolTest(t, "2 * 3 != 7", true)
}

func Test_Parser_Precedence_Logical(t *testing.T) {
	// 逻辑运算符优先级低于比较
	runBoolTest(t, "1 > 0 && 2 > 1", true)
	runBoolTest(t, "1 > 2 || 2 > 1", true)
	runBoolTest(t, "1 > 0 && 2 > 1 && 3 > 2", true)
}

func Test_Parser_Precedence_Bitwise(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1 | 2 & 3", 3},  // & 优先于 |: 2&3=2, 1|2=3
		{"1 << 2 + 1", 8}, // + 优先于 <<: 2+1=3, 1<<3=8
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

func Test_Parser_Precedence_Unary(t *testing.T) {
	intTests := []struct {
		input    string
		expected int
	}{
		{"-2 * 3", -6},
		{"-(2 + 3)", -5},
		{"-(-5)", 5},
	}

	for _, tt := range intTests {
		t.Run(tt.input, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}

	runBoolTest(t, "!true", false)
	runBoolTest(t, "!(1 > 2)", true)
}

func Test_Parser_Precedence_ComplexChain(t *testing.T) {
	runIntTest(t, "1 + 2 * 3 - 4 / 2", 5)      // 1+6-2=5
	runIntTest(t, "2 * (3 + 4) - 5", 9)        // 2*7-5=9
	runIntTest(t, "(1 + 2) * (3 + 4) - 1", 20) // 3*7-1=20
	runIntTest(t, "100 / 10 / 2 + 3 * 4", 17)  // (100/10)/2=5, 3*4=12, 5+12=17
}

func Test_Parser_Precedence_Associativity(t *testing.T) {
	// 左结合: a - b - c = (a - b) - c
	runIntTest(t, "10 - 3 - 2", 5)   // (10-3)-2=5
	runIntTest(t, "100 / 5 / 2", 10) // (100/5)/2=10
	// 右结合赋值: a = b = c  (实际不支持链式赋值，测试单个)
}

// ========== Parser复合表达式解析测试 ==========

func Test_Parser_NestedArrayAccess(t *testing.T) {
	runIntTest(t, "[[1, 2], [3, 4]][0][1]", 2)
	runIntTest(t, "[[1, 2], [3, 4]][1][0]", 3)
}

func Test_Parser_ChainedCall(t *testing.T) {
	runIntTest(t, `
		fn id(x) { return x }
		id(id(42))
	`, 42)
}

func Test_Parser_MultipleStatements(t *testing.T) {
	result := runScript(t, `
		x := 1
		y := 2
		z := 3
		x + y + z
	`)
	if result.Int() != 6 {
		t.Errorf("got %d, want 6", result.Int())
	}
}

func Test_Parser_ExpressionInCondition(t *testing.T) {
	runBoolTest(t, "1 + 1 == 2", true)
	runBoolTest(t, "2 * 3 > 5", true)
	runBoolTest(t, "10 % 3 == 1", true)
}

func Test_Parser_ArrayInExpression(t *testing.T) {
	runIntTest(t, "[1, 2, 3][0] + [4, 5, 6][1]", 6) // 1+5=6
}

func Test_Parser_FunctionCallInExpression(t *testing.T) {
	runIntTest(t, `
		fn double(x) { return x * 2 }
		double(5) + double(10)
	`, 30) // 10+20=30
}

// ========== Parser错误恢复测试 ==========

func Test_Parser_MissingSemicolon(t *testing.T) {
	// 多个语句之间应该用换行分隔
	runIntTest(t, "x := 1\ny := 2\nx + y", 3)
}

func Test_Parser_EmptyBlock(t *testing.T) {
	// 空块应该能正常编译
	compileScript(t, "if true { }")
	compileScript(t, "for false { }")
}

func Test_Parser_TrailingNewline(t *testing.T) {
	runIntTest(t, "1 + 2\n", 3)
	runIntTest(t, "\n1 + 2\n", 3)
	runIntTest(t, "\n\n1 + 2\n\n", 3)
}
