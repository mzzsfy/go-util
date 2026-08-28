package script

import (
	"testing"
)

// ========== 比较运算测试 ==========

func Test_runtime_Comparison(t *testing.T) {
	vm := newTestVM(t)

	tests := []struct {
		name     string
		a, b     Value
		op       string
		expected bool
	}{
		{"小于-true", NewValue(10), NewValue(20), "<", true},
		{"小于-false", NewValue(20), NewValue(10), "<", false},
		{"小于等于-true", NewValue(10), NewValue(10), "<=", true},
		{"大于-true", NewValue(20), NewValue(10), ">", true},
		{"大于等于-true", NewValue(10), NewValue(10), ">=", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool
			var err error

			switch tt.op {
			case "<":
				result, err = vm.less(tt.a, tt.b)
			case "<=":
				result, err = vm.lessEq(tt.a, tt.b)
			case ">":
				result, err = vm.greater(tt.a, tt.b)
			case ">=":
				result, err = vm.greaterEq(tt.a, tt.b)
			}

			if err != nil {
				t.Errorf("比较失败: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("期望 %v, 得到 %v", tt.expected, result)
			}
		})
	}
}

func Test_runtime_Comparison_String(t *testing.T) {
	vm := newTestVM(t)

	t.Run("字符串小于", func(t *testing.T) {
		result, err := vm.less(NewValue("abc"), NewValue("abd"))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if !result {
			t.Error("期望true")
		}
	})

	t.Run("字符串小于等于", func(t *testing.T) {
		result, err := vm.lessEq(NewValue("abc"), NewValue("abc"))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if !result {
			t.Error("期望true")
		}
	})

	t.Run("字符串大于", func(t *testing.T) {
		result, err := vm.greater(NewValue("abd"), NewValue("abc"))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if !result {
			t.Error("期望true")
		}
	})

	t.Run("字符串大于等于", func(t *testing.T) {
		result, err := vm.greaterEq(NewValue("abc"), NewValue("abc"))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if !result {
			t.Error("期望true")
		}
	})
}

func Test_comparison_IntComparison(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		// == 运算
		{"5 == 5", true},
		{"5 == 10", false},
		{"0 == 0", true},
		{"-5 == -5", true},
		// != 运算
		{"5 != 10", true},
		{"5 != 5", false},
		{"0 != 1", true},
		// < 运算
		{"5 < 10", true},
		{"10 < 5", false},
		{"5 < 5", false},
		{"-10 < -5", true},
		{"0 < 1", true},
		// > 运算
		{"10 > 5", true},
		{"5 > 10", false},
		{"5 > 5", false},
		{"-5 > -10", true},
		{"1 > 0", true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_FloatComparison(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		// == 运算
		{"5.0 == 5.0", true},
		{"5.5 == 5.5", true},
		{"5.0 == 10.0", false},
		// != 运算
		{"5.0 != 10.0", true},
		{"5.5 != 5.5", false},
		// < 运算
		{"5.0 < 10.0", true},
		{"10.5 < 5.0", false},
		{"5.0 < 5.0", false},
		{"-10.5 < -5.0", true},
		// > 运算
		{"10.0 > 5.0", true},
		{"5.0 > 10.0", false},
		{"5.0 > 5.0", false},
		{"-5.0 > -10.0", true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_StringComparison(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		// == 运算
		{`"abc" == "abc"`, true},
		{`"abc" == "def"`, false},
		{`"" == ""`, true},
		// != 运算
		{`"abc" != "def"`, true},
		{`"abc" != "abc"`, false},
		// < 运算（字典序）
		{`"abc" < "def"`, true},
		{`"def" < "abc"`, false},
		{`"abc" < "abc"`, false},
		{`"a" < "ab"`, true},
		// > 运算（字典序）
		{`"def" > "abc"`, true},
		{`"abc" > "def"`, false},
		{`"abc" > "abc"`, false},
		{`"ab" > "a"`, true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_BoolComparison(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		// == 运算
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		// != 运算
		{"true != false", true},
		{"true != true", false},
		{"false != false", false},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_NilComparison(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"nil == nil", true},
		{"nil != nil", false},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_unary_Negative(t *testing.T) {
	tests := []struct {
		Input    string
		Expected int
	}{
		{"-0", 0},
		{"-1", -1},
		{"-10", -10},
		{"-(-5)", 5},
		{"-100", -100},
		{"--10", 10},
	}
	RunIntTestsSimple(t, tests)
}

func Test_unary_NegativeFloat(t *testing.T) {
	tests := []struct {
		Name     string
		Input    string
		Expected float64
	}{
		{"零", "-0.0", 0.0},
		{"负数", "-1.5", -1.5},
		{"大负数", "-10.25", -10.25},
		{"双重负号", "-(-5.5)", 5.5},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			runFloatTest(t, tt.Input, tt.Expected)
		})
	}
}

func Test_comparison_Complex(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"(5 > 3) && (10 < 20)", true},
		{"(5 > 3) || (10 > 20)", true},
		{"(5 < 3) && (10 < 20)", false},
		{"!(5 > 3)", false},
		{"!(5 < 3)", true},
		{"5 == 5 && 10 == 10", true},
		{"5 == 5 || 10 == 20", true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_EdgeCases(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		// 大数比较
		{"1000000 == 1000000", true},
		{"1000000 > 999999", true},
		// 负数比较
		{"-1 < 0", true},
		{"-10 < -5", true},
		{"-100 > -200", true},
		// 零比较
		{"0 == 0", true},
		{"0 < 1", true},
		{"0 > -1", true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_logical_OrOperator(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{"1 > 0 || 2 < 1", true},
		{"1 < 0 || 2 > 1", true},
		{"1 < 0 || 2 < 1", false},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_logical_AndOperator(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"1 > 0 && 2 > 1", true},
		{"1 > 0 && 2 < 1", false},
		{"1 < 0 && 2 > 1", false},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_logical_ShortCircuitEvaluation(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"true || true", true},
		{"true || false", true},
		{"false && true", false},
		{"false && false", false},
		{"(true || false) && true", true},
		{"(false && true) || true", true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_LessEqOperator(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"5 <= 10", true},
		{"10 <= 10", true},
		{"15 <= 10", false},
		{"0 <= 0", true},
		{"-5 <= 0", true},
		{"-10 <= -5", true},
		{"5.0 <= 10.0", true},
		{"10.0 <= 10.0", true},
		{"15.0 <= 10.0", false},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_GreaterEqOperator(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"10 >= 5", true},
		{"10 >= 10", true},
		{"5 >= 10", false},
		{"0 >= 0", true},
		{"0 >= -5", true},
		{"-5 >= -10", true},
		{"10.0 >= 5.0", true},
		{"10.0 >= 10.0", true},
		{"5.0 >= 10.0", false},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_comparison_AllComparisonOperators(t *testing.T) {
	tests := []struct {
		Input    string
		Expected bool
	}{
		{"5 == 5", true},
		{"5 != 10", true},
		{"5 < 10", true},
		{"10 > 5", true},
		{"5 <= 5", true},
		{"5 >= 5", true},
		{"(5 <= 10) && (10 >= 5)", true},
		{"(5 >= 10) || (10 <= 5)", false},
		{"5 < 10 && 10 <= 10", true},
		{"10 > 5 && 5 >= 5", true},
	}
	RunBoolTestsSimple(t, tests)
}

func Test_shortcircuit_And(t *testing.T) {
	script := `
		x := 0
		y := false && (1 / x > 0)
		y
	`
	runBoolTest(t, script, false)
}

func Test_shortcircuit_AndTrue(t *testing.T) {
	script := `
		x := 5
		y := x > 0 && x < 10
		y
	`
	runBoolTest(t, script, true)
}

func Test_shortcircuit_Or(t *testing.T) {
	script := `
		x := 0
		y := true || (1 / x > 0)
		y
	`
	runBoolTest(t, script, true)
}

func Test_shortcircuit_OrFalse(t *testing.T) {
	script := `
		x := 5
		y := x < 0 || x > 10
		y
	`
	runBoolTest(t, script, false)
}

func Test_shortcircuit_Nested(t *testing.T) {
	script := `
		a := true
		b := false
		c := true
		y := (a && b) || c
		y
	`
	runBoolTest(t, script, true)
}

func Test_shortcircuit_Complex(t *testing.T) {
	script := `
		a := 1
		b := 2
		c := (a > 0 && b > 0) || (a < 0 && b < 0)
		c
	`
	runBoolTest(t, script, true)
}

func Test_shortcircuit_Chained(t *testing.T) {
	script := `
		x := 5
		y := x > 0 && x < 10 && x != 7
		y
	`
	runBoolTest(t, script, true)
}
