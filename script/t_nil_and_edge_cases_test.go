package script

import (
	"math"
	"strconv"
	"testing"
)

// ========== Nil值和边界测试 ==========
// 参照 goja/otto 和 expr-lang 的 nil 处理与边界值测试模式
// 覆盖 nil 语义、数值边界、字符串边界、空集合操作

// ========== A. Nil基本行为 ==========

// Test_Nil_Literal nil作为字面量返回
func Test_Nil_Literal(t *testing.T) {
	result := runScript(t, "nil")
	assertNil(t, result)
}

// Test_Nil_AssignToVar nil赋值给变量
func Test_Nil_AssignToVar(t *testing.T) {
	result := runScript(t, "x := nil\nx")
	assertNil(t, result)
}

// Test_Nil_Typeof typeof(nil)返回"nil"
func Test_Nil_Typeof(t *testing.T) {
	runStringTest(t, "typeof(nil)", "nil")
}

// Test_Nil_InArray nil在数组中作为元素
func Test_Nil_InArray(t *testing.T) {
	result := runScript(t, "[nil, 2][0]")
	assertNil(t, result)
}

// Test_Nil_AsMapValue nil作为Map值
func Test_Nil_AsMapValue(t *testing.T) {
	result := runScript(t, `{"a": nil}["a"]`)
	assertNil(t, result)
}

// Test_Nil_Print print(nil)不报错且返回nil
func Test_Nil_Print(t *testing.T) {
	result := runScript(t, "print(nil)")
	assertNil(t, result)
}

// Test_Nil_FunctionImplicitReturn 函数尾随表达式隐式返回其值
func Test_Nil_FunctionImplicitReturn(t *testing.T) {
	// 引擎将函数末尾的表达式作为隐式返回值
	runIntTest(t, `fn f() { 1 + 2 }
f()`, 3)
}

// Test_Nil_FunctionEmptyBody 空函数体返回nil
func Test_Nil_FunctionEmptyBody(t *testing.T) {
	result := runScript(t, `fn f() {}
f()`)
	assertNil(t, result)
}

// Test_Nil_FunctionReturnNil 函数显式return nil
func Test_Nil_FunctionReturnNil(t *testing.T) {
	result := runScript(t, `fn f() { return nil }
f()`)
	assertNil(t, result)
}

// ========== B. Nil比较语义 ==========

// Test_Nil_EqualSelf nil == nil为true
func Test_Nil_EqualSelf(t *testing.T) {
	runBoolTest(t, "nil == nil", true)
}

// Test_Nil_NotEqualSelf nil != nil为false
func Test_Nil_NotEqualSelf(t *testing.T) {
	runBoolTest(t, "nil != nil", false)
}

// Test_Nil_EqualInt nil == 0为false（不同类型）
func Test_Nil_EqualInt(t *testing.T) {
	runBoolTest(t, "nil == 0", false)
}

// Test_Nil_NotEqualInt nil != 0为true（不同类型）
func Test_Nil_NotEqualInt(t *testing.T) {
	runBoolTest(t, "nil != 0", true)
}

// Test_Nil_EqualString nil == ""为false
func Test_Nil_EqualString(t *testing.T) {
	runBoolTest(t, `nil == ""`, false)
}

// Test_Nil_EqualBool nil == false为false
func Test_Nil_EqualBool(t *testing.T) {
	runBoolTest(t, "nil == false", false)
}

// Test_Nil_VarEquality 赋值后的nil变量互相比较
func Test_Nil_VarEquality(t *testing.T) {
	result := runScript(t, "x := nil\ny := nil\nx == y")
	assertBool(t, result, true)
}

// Test_Nil_GreaterThan nil > 0报运行时错误
func Test_Nil_GreaterThan(t *testing.T) {
	runRuntimeErrorTest(t, "nil > 0")
}

// Test_Nil_Negative -nil报运行时错误
func Test_Nil_Negative(t *testing.T) {
	runRuntimeErrorTest(t, "-nil")
}

// Test_Nil_Arithmetic nil参与算术运算报运行时错误
func Test_Nil_Arithmetic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"nil加int", "nil + 1"},
		{"nil减int", "nil - 1"},
		{"nil乘int", "nil * 2"},
		{"nil除int", "nil / 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRuntimeErrorTest(t, tt.input)
		})
	}
}

// Test_Nil_AssignThenArith 不能将nil赋值给变量再用于算术
func Test_Nil_AssignThenArith(t *testing.T) {
	runRuntimeErrorTest(t, "x := nil\nx + 1")
}

// Test_Nil_IndexAccess nil不支持索引访问
func Test_Nil_IndexAccess(t *testing.T) {
	runRuntimeErrorTest(t, "nil[0]")
}

// ========== C. Nil类型转换 ==========

// Test_Nil_StringConversion string(nil)返回"nil"
func Test_Nil_StringConversion(t *testing.T) {
	runStringTest(t, "string(nil)", "nil")
}

// Test_Nil_IntConversion int(nil)报运行时错误
func Test_Nil_IntConversion(t *testing.T) {
	runRuntimeErrorTest(t, "int(nil)")
}

// Test_Nil_FloatConversion float(nil)报运行时错误
func Test_Nil_FloatConversion(t *testing.T) {
	runRuntimeErrorTest(t, "float(nil)")
}

// Test_Nil_StringConcatWithNil 字符串与nil拼接
func Test_Nil_StringConcatWithNil(t *testing.T) {
	// string + nil 拼接为 "nil"
	runStringTest(t, `"" + nil`, "nil")
	runStringTest(t, `"x" + nil`, "xnil")
}

// ========== D. Nil在数据结构中的传播 ==========

// Test_Nil_ArrayOutOfBounds 数组越界访问返回nil
func Test_Nil_ArrayOutOfBounds(t *testing.T) {
	result := runScript(t, `[1, 2, 3][10]`)
	assertNil(t, result)
}

// Test_Nil_MapMissingKey Map中不存在的key返回nil
func Test_Nil_MapMissingKey(t *testing.T) {
	result := runScript(t, `m := {"a": 1}
m["b"]`)
	assertNil(t, result)
}

// Test_Nil_NestedAccess 嵌套访问nil值
func Test_Nil_NestedAccess(t *testing.T) {
	result := runScript(t, `m := {"a": {"b": nil}}
m["a"]["b"]`)
	assertNil(t, result)
}

// Test_Nil_InIfCondition nil在if条件中走else分支
func Test_Nil_InIfCondition(t *testing.T) {
	runIntTest(t, `
r := 0
if nil { r = 1 } else { r = 2 }
r`, 2)
}

// ========== E. 数值边界 ==========

// Test_Num_MaxInt 平台int最大值
func Test_Num_MaxInt(t *testing.T) {
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	runIntTest(t, maxInt, math.MaxInt)
}

// Test_Num_MinInt 平台int最小值（通过减法避免词法问题）
func Test_Num_MinInt(t *testing.T) {
	maxInt := strconv.FormatInt(math.MaxInt, 10)
	runIntTest(t, "0 - "+maxInt, math.MinInt+1)
}

// Test_Num_ZeroArithmetic 零的特殊运算
func Test_Num_ZeroArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"零加零", "0 + 0", 0},
		{"零减零", "0 - 0", 0},
		{"零乘百", "0 * 100", 0},
		{"百乘零", "100 * 0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Num_LargeMultiply 大数乘法
func Test_Num_LargeMultiply(t *testing.T) {
	// 期望值与引擎同为int回绕语义
	a := 1000000
	runIntTest(t, "1000000 * 1000000", a*a)
}

// Test_Num_NegativeArithmetic 负数运算
func Test_Num_NegativeArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"负数加正数", "-5 + 3", -2},
		{"负数加负数", "-5 + -3", -8},
		{"负数减正数", "-5 - 3", -8},
		{"双重负号", "--10", 10},
		{"负负得正", "-(-5)", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// ========== F. 浮点数边界 ==========

// Test_Num_FloatPrecision 浮点数精度问题
func Test_Num_FloatPrecision(t *testing.T) {
	// 1.1 + 2.2 在IEEE754浮点运算中产生精度误差
	// Go编译期常量折叠得到3.3, 但运行时计算得到3.3000000000000003
	// 脚本引擎行为与运行时浮点运算一致
	result := runScript(t, "1.1 + 2.2")
	// 验证结果在合理精度范围内
	delta := result.Float() - 3.3
	if delta < 0 {
		delta = -delta
	}
	if delta > 1e-14 {
		t.Errorf("1.1 + 2.2 精度误差过大: got %v, want ~3.3", result.Float())
	}
}

// Test_Num_FloatTruncation 浮点数截断
func Test_Num_FloatTruncation(t *testing.T) {
	runIntTest(t, "int(3.99)", 3)
	runIntTest(t, "int(-3.99)", -3)
}

// Test_Num_NegativeFloatZero 负零与正零相等
func Test_Num_NegativeFloatZero(t *testing.T) {
	runBoolTest(t, "-0.0 == 0.0", true)
}

// Test_Num_SmallFloat 非常小的浮点数
func Test_Num_SmallFloat(t *testing.T) {
	runFloatTest(t, "0.000001", 0.000001)
}

// ========== G. 除法和取模边界 ==========

// Test_Num_DivByZero 除以零报运行时错误
func Test_Num_DivByZero(t *testing.T) {
	runRuntimeErrorTest(t, "1 / 0")
}

// Test_Num_FloatDivByZero 浮点数除以零报运行时错误
func Test_Num_FloatDivByZero(t *testing.T) {
	runRuntimeErrorTest(t, "1.0 / 0.0")
}

// Test_Num_ModByZero 对零取模报运行时错误
func Test_Num_ModByZero(t *testing.T) {
	runRuntimeErrorTest(t, "5 % 0")
}

// Test_Num_NegativeDivision 负数除法
func Test_Num_NegativeDivision(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"负数除正数", "-7 / 2", -3},
		{"正数除负数", "7 / -2", -3},
		{"负数除负数", "-7 / -2", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Num_NegativeModulo 负数取模
func Test_Num_NegativeModulo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"负数取模正数", "-7 % 3", -1},
		{"正数取模负数", "7 % -3", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// ========== H. 字符串索引和切片边界 ==========

// Test_Str_Edge_SliceEmpty 字符串切片返回空字符串
func Test_Str_Edge_SliceEmpty(t *testing.T) {
	runStringTest(t, `"hello"[0:0]`, "")
}

// Test_Str_Edge_SliceFull 字符串切片返回完整字符串
func Test_Str_Edge_SliceFull(t *testing.T) {
	runStringTest(t, `"hello"[0:5]`, "hello")
}

// Test_Str_Edge_SliceWithLen 使用len()的字符串切片
func Test_Str_Edge_SliceWithLen(t *testing.T) {
	runStringTest(t, `s := "hello"
s[0:len(s)]`, "hello")
}

// Test_Str_Edge_EmptyStringSlice 空字符串切片
func Test_Str_Edge_EmptyStringSlice(t *testing.T) {
	runStringTest(t, `""[0:0]`, "")
}

// Test_Str_Edge_EmptyStringIndex 空字符串索引返回nil
func Test_Str_Edge_EmptyStringIndex(t *testing.T) {
	result := runScript(t, `""[0]`)
	assertNil(t, result)
}

// Test_Str_Edge_OutOfBoundsIndex 字符串越界索引返回nil
func Test_Str_Edge_OutOfBoundsIndex(t *testing.T) {
	result := runScript(t, `"abc"[100]`)
	assertNil(t, result)
}

// Test_Str_Edge_OversizeSlice 越界切片限制到字符串长度
func Test_Str_Edge_OversizeSlice(t *testing.T) {
	// 越界切片不报错, 截断到字符串长度
	runStringTest(t, `"abc"[0:10]`, "abc")
}

// ========== I. 字符串拼接边界 ==========

// Test_Str_Edge_EmptyConcat 空字符串拼接
func Test_Str_Edge_EmptyConcat(t *testing.T) {
	runStringTest(t, `"" + ""`, "")
	runStringTest(t, `"" + "x"`, "x")
	runStringTest(t, `"x" + ""`, "x")
}

// Test_Str_Edge_MultiConcat 多个字符串拼接
func Test_Str_Edge_MultiConcat(t *testing.T) {
	runStringTest(t, `"a" + "b" + "c" + "d"`, "abcd")
}

// ========== J. 字符串比较 ==========

// Test_Str_Edge_Comparison 字符串比较语义
func Test_Str_Edge_Comparison(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"空串相等", `"" == ""`, true},
		{"字典序小于", `"abc" < "abd"`, true},
		{"前缀小于长串", `"ab" < "abc"`, true},
		{"大小写敏感", `"A" < "a"`, true},
		{"相等不小于", `"abc" < "abc"`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBoolTest(t, tt.input, tt.expected)
		})
	}
}

// ========== K. 空集合边界 ==========

// Test_Empty_ArrayLen 空数组长度为0
func Test_Empty_ArrayLen(t *testing.T) {
	runIntTest(t, "len([])", 0)
}

// Test_Empty_ArrayOutOfBounds 空数组越界访问返回nil
func Test_Empty_ArrayOutOfBounds(t *testing.T) {
	result := runScript(t, "[][0]")
	assertNil(t, result)
}

// Test_Empty_ArrayPush 向空数组push元素
func Test_Empty_ArrayPush(t *testing.T) {
	runIntTest(t, "len(push([], 42))", 1)
}

// Test_Empty_MapLen 空Map长度为0
func Test_Empty_MapLen(t *testing.T) {
	runIntTest(t, "len({})", 0)
}

// Test_Empty_MapMissingKey 空Map访问key返回nil
func Test_Empty_MapMissingKey(t *testing.T) {
	result := runScript(t, `{}["key"]`)
	assertNil(t, result)
}

// Test_Empty_MapDelete 删除空Map中不存在的key不报错
func Test_Empty_MapDelete(t *testing.T) {
	runIntTest(t, `m := {}
delete(m, "key")
len(m)`, 0)
}
