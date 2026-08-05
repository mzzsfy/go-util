package script

import "testing"

// ========== 类型转换测试 ==========
// 参照 goja 和 expr-lang 的类型转换测试
// 本文件测试 int()、float()、string()、typeof()、类型比较语义和隐式类型行为

// ---------- int() 转换 ----------

// Test_Type_IntFromString int()从字符串转换
func Test_Type_IntFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"正数字符串", `int("42")`, 42},
		{"零字符串", `int("0")`, 0},
		{"负数字符串", `int("-5")`, -5},
		{"大正数", `int("123456")`, 123456},
		{"大负数", `int("-99999")`, -99999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_IntFromFloat int()从浮点数转换(截断)
func Test_Type_IntFromFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"正浮点截断", `int(3.14)`, 3},
		{"接近整数截断", `int(3.99)`, 3},
		{"负浮点截断", `int(-3.14)`, -3},
		{"负接近整数截断", `int(-3.99)`, -3},
		{"零浮点", `int(0.0)`, 0},
		{"整数浮点", `int(10.0)`, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_IntFromBool int()从布尔值转换
func Test_Type_IntFromBool(t *testing.T) {
	runIntTest(t, `int(true)`, 1)
	runIntTest(t, `int(false)`, 0)
}

// Test_Type_IntInvalidString int()对无效字符串的行为
func Test_Type_IntInvalidString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"字母字符串", `int("abc")`},
		{"浮点数字符串", `int("3.14")`},
		{"空字符串", `int("")`},
		{"带空格", `int(" 123 ")`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRuntimeErrorTest(t, tt.input)
		})
	}
}

// Test_Type_IntUnsupportedType int()对不支持的类型报错
func Test_Type_IntUnsupportedType(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"数组转int", `int([1, 2])`},
		{"Map转int", `int({"a": 1})`},
		{"nil转int", `int(nil)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRuntimeErrorTest(t, tt.input)
		})
	}
}

// ---------- float() 转换 ----------

// Test_Type_FloatFromString float()从字符串转换
func Test_Type_FloatFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"小数", `float("3.14")`, 3.14},
		{"零", `float("0")`, 0.0},
		{"负小数", `float("-5.5")`, -5.5},
		{"整数串", `float("42")`, 42.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runFloatTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_FloatFromInt float()从整数转换
func Test_Type_FloatFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"正整数", `float(42)`, 42.0},
		{"零", `float(0)`, 0.0},
		{"负整数", `float(-7)`, -7.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runFloatTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_FloatFromBool float()从布尔值转换
func Test_Type_FloatFromBool(t *testing.T) {
	runFloatTest(t, `float(true)`, 1.0)
	runFloatTest(t, `float(false)`, 0.0)
}

// Test_Type_FloatInvalidString float()对无效字符串的行为
func Test_Type_FloatInvalidString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"字母", `float("abc")`},
		{"空串", `float("")`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRuntimeErrorTest(t, tt.input)
		})
	}
}

// Test_Type_FloatUnsupportedType float()对不支持的类型报错
func Test_Type_FloatUnsupportedType(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"数组转float", `float([1.0, 2.0])`},
		{"Map转float", `float({"a": 1})`},
		{"nil转float", `float(nil)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRuntimeErrorTest(t, tt.input)
		})
	}
}

// ---------- string() 转换 ----------

// Test_Type_StringFromInt string()从整数转换
func Test_Type_StringFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"正整数", `string(42)`, "42"},
		{"零", `string(0)`, "0"},
		{"负数", `string(-5)`, "-5"},
		{"大数", `string(1000000)`, "1000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_StringFromFloat string()从浮点数转换
func Test_Type_StringFromFloat(t *testing.T) {
	runStringTest(t, `string(3.14)`, "3.14")
}

// Test_Type_StringFromBool string()从布尔值转换
func Test_Type_StringFromBool(t *testing.T) {
	runStringTest(t, `string(true)`, "true")
	runStringTest(t, `string(false)`, "false")
}

// Test_Type_StringFromString string()对字符串直接返回
func Test_Type_StringFromString(t *testing.T) {
	runStringTest(t, `string("already")`, "already")
}

// Test_Type_StringFromNil string()对nil转换
func Test_Type_StringFromNil(t *testing.T) {
	runStringTest(t, `string(nil)`, "nil")
}

// ---------- typeof() 全类型覆盖 ----------

// Test_Type_TypeofAllTypes typeof()覆盖所有内置类型
func Test_Type_TypeofAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"整数", `typeof(42)`, "int"},
		{"浮点", `typeof(3.14)`, "float"},
		{"字符串", `typeof("str")`, "string"},
		{"布尔true", `typeof(true)`, "bool"},
		{"布尔false", `typeof(false)`, "bool"},
		{"nil", `typeof(nil)`, "nil"},
		{"空数组", `typeof([])`, "array"},
		{"非空数组", `typeof([1, 2, 3])`, "array"},
		{"空Map", `typeof({})`, "map"},
		{"非空Map", `typeof({"a": 1})`, "map"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runStringTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_TypeofFunction typeof()检测函数类型
func Test_Type_TypeofFunction(t *testing.T) {
	runStringTest(t, `fn f() { 1 }
typeof(f)`, "function")
}

// ---------- 类型比较语义 ----------

// Test_Type_IntEqual int == int
func Test_Type_IntEqual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"相等", `42 == 42`, true},
		{"不等", `42 == 43`, false},
		{"零相等", `0 == 0`, true},
		{"负数相等", `-5 == -5`, true},
		{"正负不等", `5 == -5`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_FloatEqual float == float
func Test_Type_FloatEqual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"相等", `3.14 == 3.14`, true},
		{"不等", `3.14 == 2.71`, false},
		{"零相等", `0.0 == 0.0`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_StringEqual string == string
func Test_Type_StringEqual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"相等", `"hello" == "hello"`, true},
		{"不等", `"hello" == "world"`, false},
		{"空串相等", `"" == ""`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_BoolEqual bool == bool
func Test_Type_BoolEqual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true==true", `true == true`, true},
		{"false==false", `false == false`, true},
		{"true==false", `true == false`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_NilEqual nil == nil
func Test_Type_NilEqual(t *testing.T) {
	runBoolTest(t, `nil == nil`, true)
	runBoolTest(t, `nil != nil`, false)
}

// Test_Type_DifferentTypeCompare 不同类型比较结果
// 数值类型(int/float)自动提升比较
func Test_Type_DifferentTypeCompare(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"int与float相等", `5 == 5.0`, true},
		{"int与string不等", `42 == "42"`, false},
		{"bool与int不等", `true == 1`, false},
		{"nil与int不等", `nil == 0`, false},
		{"int与float等号", `5 != 5.0`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// ---------- 隐式类型行为 ----------

// Test_Type_IfAcceptsBool if条件接受布尔值
func Test_Type_IfAcceptsBool(t *testing.T) {
	runIntTest(t, `
r := 0
if true { r = 10 } else { r = 20 }
r`, 10)
}

// Test_Type_IfRejectsNonBool if条件对非bool值使用truthiness判断
func Test_Type_IfRejectsNonBool(t *testing.T) {
	// 非0整数和非空字符串为真走then, 0为假走else
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"整数条件走then", `
r := 0
if 1 { r = 10 } else { r = 20 }
r`, 10},
		{"零走else", `
r := 0
if 0 { r = 10 } else { r = 20 }
r`, 20},
		{"字符串走then", `
r := 0
if "hello" { r = 10 } else { r = 20 }
r`, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Type_LogicalAndReturnType && 返回bool类型
func Test_Type_LogicalAndReturnType(t *testing.T) {
	result := runScript(t, `typeof(true && false)`)
	if result.String() != "bool" {
		t.Errorf("期望 bool, 得到 %s", result.String())
	}
}

// Test_Type_LogicalOrReturnType || 返回bool类型
func Test_Type_LogicalOrReturnType(t *testing.T) {
	result := runScript(t, `typeof(true || false)`)
	if result.String() != "bool" {
		t.Errorf("期望 bool, 得到 %s", result.String())
	}
}

// Test_Type_NotReturnType ! 返回bool类型
func Test_Type_NotReturnType(t *testing.T) {
	result := runScript(t, `typeof(!true)`)
	if result.String() != "bool" {
		t.Errorf("期望 bool, 得到 %s", result.String())
	}
}

// Test_Type_ChainedConversion 链式类型转换
func Test_Type_ChainedConversion(t *testing.T) {
	// int -> float -> string -> int
	runIntTest(t, `int(string(float(42)))`, 42)
}

// Test_Type_ConversionInExpr 转换在表达式中使用
func Test_Type_ConversionInExpr(t *testing.T) {
	// int("10") + int("20") = 30
	runIntTest(t, `int("10") + int("20")`, 30)
}

// Test_Type_StringConcatWithString 字符串拼接使用string()转换
func Test_Type_StringConcatWithString(t *testing.T) {
	runStringTest(t, `"count: " + string(42)`, "count: 42")
}

// Test_Type_LenWithConversion len()配合类型使用
func Test_Type_LenWithConversion(t *testing.T) {
	// len(string(12345)) = 5
	runIntTest(t, `len(string(12345))`, 5)
}

// Test_Type_TypeofInCondition typeof()结果在条件中使用
func Test_Type_TypeofInCondition(t *testing.T) {
	runIntTest(t, `
x := 42
r := 0
if typeof(x) == "int" {
    r = 1
} else {
    r = 2
}
r`, 1)
}
