package script

import (
	"testing"
)

// ========== 其他运行时测试（Context, Builtin, External, Utils等）==========

func Test_isvalidindex_Complete(t *testing.T) {
	tests := []struct {
		idx      int
		length   int
		expected bool
	}{
		{0, 5, true},
		{4, 5, true},
		{5, 5, false},
		{-1, 5, false},
		{0, 0, false},
		{10, 5, false},
	}

	for _, tt := range tests {
		result := isValidIndex(tt.idx, tt.length)
		if result != tt.expected {
			t.Errorf("isValidIndex(%d, %d) = %v, expected %v",
				tt.idx, tt.length, result, tt.expected)
		}
	}
}

func Test_normalizeslicebounds_Complete(t *testing.T) {
	tests := []struct {
		s, e, length int
		expS, expE   int
		expValid     bool
	}{
		{0, 5, 5, 0, 5, true},
		{1, 3, 5, 1, 3, true},
		{0, 10, 5, 0, 5, true},
		{2, 2, 5, 2, 2, false},
		{-1, 3, 5, 4, 3, false}, // 负数start从末尾倒数, -1+5=4
	}

	for _, tt := range tests {
		s, e, valid := normalizeSliceBounds(tt.s, tt.e, tt.length)
		if s != tt.expS || e != tt.expE || valid != tt.expValid {
			t.Errorf("normalizeSliceBounds(%d, %d, %d) = (%d, %d, %v), expected (%d, %d, %v)",
				tt.s, tt.e, tt.length, s, e, valid, tt.expS, tt.expE, tt.expValid)
		}
	}
}

func Test_valuetostring_Complete(t *testing.T) {
	tests := []struct {
		val      Value
		expected string
	}{
		{NewValue("hello"), "hello"},
		{NewValue(42), "42"},
		{NewValue(3.14), "3.14"},
		{NewValue(true), "true"},
		{NewValue(false), "false"},
		{NewValue(nil), "nil"},
	}

	for _, tt := range tests {
		result := valueToString(tt.val)
		if result != tt.expected {
			t.Errorf("valueToString(%v) = %s, expected %s", tt.val, result, tt.expected)
		}
	}
}

func Test_typename_Complete(t *testing.T) {
	tests := []struct {
		valType  ValueType
		expected string
	}{
		{TypeInt, "int"},
		{TypeFloat, "float"},
		{TypeString, "string"},
		{TypeBool, "bool"},
		{TypeNil, "nil"},
		{TypeArray, "array"},
		{TypeMap, "map"},
		{TypeFunction, "function"},
		{ValueType(999), "unknown"},
	}

	for _, tt := range tests {
		result := typeName(tt.valType)
		if result != tt.expected {
			t.Errorf("typeName(%v) = %s, expected %s", tt.valType, result, tt.expected)
		}
	}
}

func Test_iszerovalue_Complete(t *testing.T) {
	tests := []struct {
		val      Value
		expected bool
	}{
		{NewValue(0), true},
		{NewValue(1), false},
		{NewValue(0.0), true},
		{NewValue(0.1), false},
		{NewValue(""), false},
		{NewValue(nil), false},
		{NewValue(false), false},
	}

	for _, tt := range tests {
		result := isZeroValue(tt.val)
		if result != tt.expected {
			t.Errorf("isZeroValue(%v) = %v, expected %v", tt.val, result, tt.expected)
		}
	}
}

func Test_context_BindFunc_Complete(t *testing.T) {
	ctx := NewContext()
	fn := func(x, y int) int { return x + y }
	ctx.BindFunc("add", fn)

	fn2, ok := ctx.GetBindFunc("add")
	if !ok {
		t.Error("绑定函数应存在")
		return
	}
	result := fn2.(func(int, int) int)(1, 2)
	if result != 3 {
		t.Errorf("函数调用失败: %d", result)
	}
}

func Test_context_BindValue_Complete(t *testing.T) {
	ctx := NewContext()
	ctx.BindValue("num", 42)
	ctx.BindValue("str", "hello")
	ctx.BindValue("arr", []int{1, 2, 3})

	val1, ok1 := ctx.GetBindValue("num")
	if !ok1 || val1.Int() != 42 {
		t.Error("整数绑定失败")
	}

	val2, ok2 := ctx.GetBindValue("str")
	if !ok2 || val2.String() != "hello" {
		t.Error("字符串绑定失败")
	}
}

func Test_context_GetBindValue_NotExist(t *testing.T) {
	ctx := NewContext()
	_, ok := ctx.GetBindValue("nonexistent")
	if ok {
		t.Error("不存在的绑定值应该返回false")
	}
}

func Test_context_GetBindFunc_NotExist(t *testing.T) {
	ctx := NewContext()
	_, ok := ctx.GetBindFunc("nonexistent")
	if ok {
		t.Error("不存在的绑定函数应该返回false")
	}
}

func Test_formatvalueforprint(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"nil类型", NewValue(nil), "nil"},
		{"整数类型", NewValue(42), "42"},
		{"浮点类型", NewValue(3.14), "3.14"},
		{"字符串类型", NewValue("hello"), "\"hello\""},
		{"布尔类型true", NewValue(true), "true"},
		{"布尔类型false", NewValue(false), "false"},
		{"数组类型", NewValue([]Value{NewValue(1), NewValue(2)}), "[1, 2]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValueForPrint(tt.value)
			if result != tt.expected {
				t.Errorf("期望 %s, 得到 %s", tt.expected, result)
			}
		})
	}
}

func Test_formatvalueforprint_Function(t *testing.T) {
	// 创建一个简单的编译函数
	parser := NewParser()
	script, err := parser.Compile("fn testFunc() { 1 }")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	if len(script.Functions) == 0 {
		t.Fatal("没有编译任何函数")
	}

	fn := script.Functions[0]
	fnValue := Value{Type: TypeFunction, Data: &FunctionValue{Compiled: fn}}

	result := formatValueForPrint(fnValue)
	if result != "<function testFunc>" {
		t.Errorf("期望 <function testFunc>, 得到 %s", result)
	}
}

func Test_formatvalueforprint_ExternalFunc(t *testing.T) {
	extValue := Value{Type: TypeExternalFunc, Data: &ExternalFuncValue{Name: "myExternalFunc"}}

	result := formatValueForPrint(extValue)
	if result != "<external myExternalFunc>" {
		t.Errorf("期望 <external myExternalFunc>, 得到 %s", result)
	}
}

func Test_formatvalueforprint_Unknown(t *testing.T) {
	// 创建一个未知类型的值
	unknownValue := Value{Type: ValueType(255), Data: nil}

	result := formatValueForPrint(unknownValue)
	if result != "<unknown>" {
		t.Errorf("期望 <unknown>, 得到 %s", result)
	}
}
