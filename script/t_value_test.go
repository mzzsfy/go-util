package script

import (
	"testing"
)

// ========== Value类型访问器测试 ==========

func Test_Value_Int(t *testing.T) {
	v := NewValue(42)
	if v.Type != TypeInt {
		t.Errorf("type应为TypeInt, got %v", v.Type)
	}
	if v.Int() != 42 {
		t.Errorf("Int()应为42, got %d", v.Int())
	}
}

func Test_Value_Int64(t *testing.T) {
	v := NewValue(int64(100))
	if v.Type != TypeInt {
		t.Errorf("int64应转为TypeInt, got %v", v.Type)
	}
	if v.Int() != 100 {
		t.Errorf("Int()应为100, got %d", v.Int())
	}
}

func Test_Value_Float(t *testing.T) {
	v := NewValue(3.14)
	if v.Type != TypeFloat {
		t.Errorf("type应为TypeFloat, got %v", v.Type)
	}
	if v.Float() != 3.14 {
		t.Errorf("Float()应为3.14, got %f", v.Float())
	}
}

func Test_Value_String(t *testing.T) {
	v := NewValue("hello")
	if v.Type != TypeString {
		t.Errorf("type应为TypeString, got %v", v.Type)
	}
	if v.String() != "hello" {
		t.Errorf("String()应为hello, got %s", v.String())
	}
}

func Test_Value_Bool(t *testing.T) {
	v := NewValue(true)
	if v.Type != TypeBool {
		t.Errorf("type应为TypeBool, got %v", v.Type)
	}
	if !v.Bool() {
		t.Error("Bool()应为true")
	}
}

func Test_Value_Nil(t *testing.T) {
	v := NewValue(nil)
	if v.Type != TypeNil {
		t.Errorf("type应为TypeNil, got %v", v.Type)
	}
	if !v.IsNil() {
		t.Error("IsNil()应为true")
	}
}

func Test_Value_Array(t *testing.T) {
	elems := []Value{NewValue(1), NewValue(2), NewValue(3)}
	v := NewValue(elems)
	if v.Type != TypeArray {
		t.Errorf("type应为TypeArray, got %v", v.Type)
	}
	arr := v.Array()
	if arr == nil {
		t.Fatal("Array()不应返回nil")
	}
	if len(arr.Elements) != 3 {
		t.Errorf("数组长度应为3, got %d", len(arr.Elements))
	}
	if arr.Elements[0].Int() != 1 {
		t.Errorf("第一个元素应为1, got %d", arr.Elements[0].Int())
	}
}

func Test_Value_Map(t *testing.T) {
	m := &MapValue{
		Pairs:   map[string]Value{"a": NewValue(1), "b": NewValue(2)},
		KeyType: TypeString,
	}
	v := NewValue(m)
	if v.Type != TypeMap {
		t.Errorf("type应为TypeMap, got %v", v.Type)
	}
	mv := v.Map()
	if mv == nil {
		t.Fatal("Map()不应返回nil")
	}
	if len(mv.Pairs) != 2 {
		t.Errorf("map大小应为2, got %d", len(mv.Pairs))
	}
}

func Test_Value_Function(t *testing.T) {
	fn := &FunctionValue{Compiled: &CompiledFunction{Name: "test"}}
	v := NewValue(fn)
	if v.Type != TypeFunction {
		t.Errorf("type应为TypeFunction, got %v", v.Type)
	}
	fv := v.Function()
	if fv == nil {
		t.Fatal("Function()不应返回nil")
	}
	if fv.Compiled.Name != "test" {
		t.Errorf("函数名应为test, got %s", fv.Compiled.Name)
	}
}

func Test_Value_ExternalFunc(t *testing.T) {
	ef := &ExternalFuncValue{Name: "ext", Func: func() {}}
	v := Value{Type: TypeExternalFunc, Data: ef}
	if v.Type != TypeExternalFunc {
		t.Errorf("type应为TypeExternalFunc, got %v", v.Type)
	}
	ext := v.ExternalFunc()
	if ext == nil {
		t.Fatal("ExternalFunc()不应返回nil")
	}
	if ext.Name != "ext" {
		t.Errorf("名称应为ext, got %s", ext.Name)
	}
}

// ========== Value类型不匹配访问测试 ==========

func Test_Value_TypeMismatch(t *testing.T) {
	intVal := NewValue(42)

	// int值访问其他类型应返回零值, Float()例外: 自动转换为float64
	if intVal.String() != "" {
		t.Errorf("int的String()应返回空字符串, got %q", intVal.String())
	}
	if intVal.Float() != 42.0 {
		t.Errorf("int的Float()应返回42.0(自动转换), got %f", intVal.Float())
	}
	if intVal.Bool() {
		t.Error("int的Bool()应返回false")
	}
	if intVal.Array() != nil {
		t.Error("int的Array()应返回nil")
	}
	if intVal.Map() != nil {
		t.Error("int的Map()应返回nil")
	}
	if intVal.Function() != nil {
		t.Error("int的Function()应返回nil")
	}
	if intVal.ExternalFunc() != nil {
		t.Error("int的ExternalFunc()应返回nil")
	}
}

func Test_Value_StringTypeMismatch(t *testing.T) {
	strVal := NewValue("hello")

	if strVal.Int() != 0 {
		t.Errorf("string的Int()应返回0, got %d", strVal.Int())
	}
	if strVal.Float() != 0 {
		t.Errorf("string的Float()应返回0, got %f", strVal.Float())
	}
}

func Test_Value_BoolTypeMismatch(t *testing.T) {
	boolVal := NewValue(true)

	if boolVal.Int() != 0 {
		t.Errorf("bool的Int()应返回0, got %d", boolVal.Int())
	}
}

func Test_Value_NilTypeMismatch(t *testing.T) {
	nilVal := NewValue(nil)

	if nilVal.Int() != 0 {
		t.Error("nil的Int()应返回0")
	}
	if nilVal.Float() != 0 {
		t.Error("nil的Float()应返回0")
	}
	if nilVal.String() != "" {
		t.Error("nil的String()应返回空字符串")
	}
	if nilVal.Bool() {
		t.Error("nil的Bool()应返回false")
	}
	if nilVal.Array() != nil {
		t.Error("nil的Array()应返回nil")
	}
}

// ========== Value相等比较测试 ==========

func Test_Value_Equal_SameType(t *testing.T) {
	tests := []struct {
		name string
		a, b Value
		want bool
	}{
		{"int相等", NewValue(1), NewValue(1), true},
		{"int不等", NewValue(1), NewValue(2), false},
		{"float相等", NewValue(1.5), NewValue(1.5), true},
		{"float不等", NewValue(1.5), NewValue(2.5), false},
		{"string相等", NewValue("a"), NewValue("a"), true},
		{"string不等", NewValue("a"), NewValue("b"), false},
		{"bool相等", NewValue(true), NewValue(true), true},
		{"bool不等", NewValue(true), NewValue(false), false},
		{"nil相等", NewValue(nil), NewValue(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.a.Equal(tt.b) != tt.want {
				t.Errorf("Equal() = %v, want %v", !tt.want, tt.want)
			}
		})
	}
}

func Test_Value_Equal_DifferentType(t *testing.T) {
	tests := []struct {
		name string
		a, b Value
	}{
		{"int vs string", NewValue(1), NewValue("1")},
		{"int vs bool", NewValue(1), NewValue(true)},
		{"string vs bool", NewValue("true"), NewValue(true)},
		{"nil vs int", NewValue(nil), NewValue(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.a.Equal(tt.b) {
				t.Error("不同类型的值不应相等")
			}
		})
	}
}

func Test_Value_Equal_Array(t *testing.T) {
	a1 := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
	a2 := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
	a3 := NewValue([]Value{NewValue(1), NewValue(2), NewValue(4)})
	a4 := NewValue([]Value{NewValue(1), NewValue(2)})

	if !a1.Equal(a2) {
		t.Error("相同数组应相等")
	}
	if a1.Equal(a3) {
		t.Error("元素不同的数组不应相等")
	}
	if a1.Equal(a4) {
		t.Error("长度不同的数组不应相等")
	}
}

func Test_Value_Equal_Map(t *testing.T) {
	m1 := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}, KeyType: TypeString})
	m2 := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}, KeyType: TypeString})
	m3 := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(2)}, KeyType: TypeString})
	m4 := NewValue(&MapValue{Pairs: map[string]Value{"b": NewValue(1)}, KeyType: TypeString})

	if !m1.Equal(m2) {
		t.Error("相同map应相等")
	}
	if m1.Equal(m3) {
		t.Error("值不同的map不应相等")
	}
	if m1.Equal(m4) {
		t.Error("键不同的map不应相等")
	}
}

func Test_Value_Equal_NestedArray(t *testing.T) {
	outer1 := NewValue([]Value{
		NewValue([]Value{NewValue(1), NewValue(2)}),
		NewValue([]Value{NewValue(3), NewValue(4)}),
	})
	outer2 := NewValue([]Value{
		NewValue([]Value{NewValue(1), NewValue(2)}),
		NewValue([]Value{NewValue(3), NewValue(4)}),
	})

	if !outer1.Equal(outer2) {
		t.Error("嵌套数组应递归比较相等")
	}
}

// ========== NewValue各种输入测试 ==========

func Test_NewValue_Types(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantType ValueType
	}{
		{"int", 42, TypeInt},
		{"int64", int64(42), TypeInt},
		{"float64", 3.14, TypeFloat},
		{"string", "hello", TypeString},
		{"bool", true, TypeBool},
		{"nil", nil, TypeNil},
		{"[]Value", []Value{NewValue(1)}, TypeArray},
		{"*ArrayValue", &ArrayValue{Elements: []Value{}}, TypeArray},
		{"*MapValue", &MapValue{Pairs: map[string]Value{}}, TypeMap},
		{"*FunctionValue", &FunctionValue{Compiled: &CompiledFunction{}}, TypeFunction},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.input)
			if v.Type != tt.wantType {
				t.Errorf("NewValue(%v).Type = %v, want %v", tt.name, v.Type, tt.wantType)
			}
		})
	}
}

func Test_NewValue_UnknownType(t *testing.T) {
	// 不支持的类型应返回nil
	type custom struct{ X int }
	v := NewValue(custom{X: 1})
	if !v.IsNil() {
		t.Errorf("不支持的类型应返回nil, got type %v", v.Type)
	}
}

// ========== Value.IsNil测试 ==========

func Test_Value_IsNil_AllTypes(t *testing.T) {
	if !NewValue(nil).IsNil() {
		t.Error("nil应IsNil")
	}
	if NewValue(0).IsNil() {
		t.Error("int 0不应IsNil")
	}
	if NewValue("").IsNil() {
		t.Error("空字符串不应IsNil")
	}
	if NewValue(false).IsNil() {
		t.Error("false不应IsNil")
	}
	if NewValue(0.0).IsNil() {
		t.Error("0.0不应IsNil")
	}
}

// ========== smallInts缓存测试 ==========

func Test_smallInts_Cache(t *testing.T) {
	for i := 0; i < 256; i++ {
		v := intVal(i)
		if v.Type != TypeInt {
			t.Errorf("smallInts[%d].Type = %v, want TypeInt", i, v.Type)
		}
		if v.Int() != i {
			t.Errorf("smallInts[%d].Int() = %d, want %d", i, v.Int(), i)
		}
	}
}

func Test_intVal_OutsideCache(t *testing.T) {
	v := intVal(256)
	if v.Int() != 256 {
		t.Errorf("intVal(256).Int() = %d, want 256", v.Int())
	}
	v = intVal(-1)
	if v.Int() != -1 {
		t.Errorf("intVal(-1).Int() = %d, want -1", v.Int())
	}
}
