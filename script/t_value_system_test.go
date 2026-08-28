package script

import (
	"strconv"
	"strings"
	"testing"
)

// ========== Value 类型系统完整测试 ==========
// 本文件覆盖 value.go 中所有公开方法与关键内部函数
// 包括: NewValue 类型推断, 类型访问器, Equal, GoString, 边界, 类型转换链, 复合值操作, formatValueForPrint

// ---------- NewValue 类型推断 ----------

// Test_ValueSystem_NewValue_Int int 推断为 TypeInt
func Test_ValueSystem_NewValue_Int(t *testing.T) {
	t.Run("正整数", func(t *testing.T) {
		v := NewValue(42)
		if v.Type != TypeInt {
			t.Errorf("NewValue(42).Type = %v, want TypeInt", v.Type)
		}
		if v.Int() != 42 {
			t.Errorf("Int() = %d, want 42", v.Int())
		}
	})
	t.Run("零", func(t *testing.T) {
		v := NewValue(0)
		if v.Type != TypeInt {
			t.Errorf("NewValue(0).Type = %v, want TypeInt", v.Type)
		}
	})
}

// Test_ValueSystem_NewValue_Int64 int64 推断为 TypeInt
func Test_ValueSystem_NewValue_Int64(t *testing.T) {
	v := NewValue(int64(100))
	if v.Type != TypeInt {
		t.Errorf("NewValue(int64).Type = %v, want TypeInt", v.Type)
	}
	if v.Int() != 100 {
		t.Errorf("Int() = %d, want 100", v.Int())
	}
}

// Test_ValueSystem_NewValue_Float64 float64 推断为 TypeFloat
func Test_ValueSystem_NewValue_Float64(t *testing.T) {
	v := NewValue(3.14)
	if v.Type != TypeFloat {
		t.Errorf("NewValue(3.14).Type = %v, want TypeFloat", v.Type)
	}
	if v.Float() != 3.14 {
		t.Errorf("Float() = %f, want 3.14", v.Float())
	}
}

// Test_ValueSystem_NewValue_String string 推断为 TypeString
func Test_ValueSystem_NewValue_String(t *testing.T) {
	v := NewValue("hello")
	if v.Type != TypeString {
		t.Errorf("NewValue(\"hello\").Type = %v, want TypeString", v.Type)
	}
	if v.String() != "hello" {
		t.Errorf("String() = %q, want \"hello\"", v.String())
	}
}

// Test_ValueSystem_NewValue_Bool bool 推断为 TypeBool
func Test_ValueSystem_NewValue_Bool(t *testing.T) {
	v := NewValue(true)
	if v.Type != TypeBool {
		t.Errorf("NewValue(true).Type = %v, want TypeBool", v.Type)
	}
	if !v.Bool() {
		t.Error("Bool() = false, want true")
	}
}

// Test_ValueSystem_NewValue_Nil nil 推断为 TypeNil
func Test_ValueSystem_NewValue_Nil(t *testing.T) {
	v := NewValue(nil)
	if v.Type != TypeNil {
		t.Errorf("NewValue(nil).Type = %v, want TypeNil", v.Type)
	}
	if !v.IsNil() {
		t.Error("IsNil() = false, want true")
	}
}

// Test_ValueSystem_NewValue_SliceValue []Value 推断为 TypeArray
func Test_ValueSystem_NewValue_SliceValue(t *testing.T) {
	elems := []Value{NewValue(1), NewValue(2), NewValue(3)}
	v := NewValue(elems)
	if v.Type != TypeArray {
		t.Errorf("NewValue([]Value).Type = %v, want TypeArray", v.Type)
	}
	arr := v.Array()
	if arr == nil {
		t.Fatal("Array() = nil, want non-nil")
	}
	if len(arr.Elements) != 3 {
		t.Errorf("len(Elements) = %d, want 3", len(arr.Elements))
	}
}

// Test_ValueSystem_NewValue_ArrayValuePtr *ArrayValue 推断为 TypeArray
func Test_ValueSystem_NewValue_ArrayValuePtr(t *testing.T) {
	av := &ArrayValue{Elements: []Value{NewValue(10)}}
	v := NewValue(av)
	if v.Type != TypeArray {
		t.Errorf("NewValue(*ArrayValue).Type = %v, want TypeArray", v.Type)
	}
	// 验证共享同一指针
	if v.Array() != av {
		t.Error("Array() 应返回同一个指针")
	}
}

// Test_ValueSystem_NewValue_MapValuePtr *MapValue 推断为 TypeMap
func Test_ValueSystem_NewValue_MapValuePtr(t *testing.T) {
	mv := &MapValue{Pairs: map[string]Value{"k": NewValue(1)}, KeyType: TypeString}
	v := NewValue(mv)
	if v.Type != TypeMap {
		t.Errorf("NewValue(*MapValue).Type = %v, want TypeMap", v.Type)
	}
	if v.Map() != mv {
		t.Error("Map() 应返回同一个指针")
	}
}

// Test_ValueSystem_NewValue_FunctionValuePtr *FunctionValue 推断为 TypeFunction
func Test_ValueSystem_NewValue_FunctionValuePtr(t *testing.T) {
	fv := &FunctionValue{Compiled: &CompiledFunction{Name: "fn"}}
	v := NewValue(fv)
	if v.Type != TypeFunction {
		t.Errorf("NewValue(*FunctionValue).Type = %v, want TypeFunction", v.Type)
	}
	if v.Function() != fv {
		t.Error("Function() 应返回同一个指针")
	}
}

// Test_ValueSystem_NewValue_UnsupportedSlices 不支持的切片类型降级为 TypeNil
// 引擎仅处理 []Value, 其他切片类型不经过自动转换
func Test_ValueSystem_NewValue_UnsupportedSlices(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"[]int", []int{1, 2, 3}},
		{"[]string", []string{"a", "b"}},
		{"[]interface{}", []interface{}{1, "x", true}},
		{"[]int64", []int64{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.input)
			if v.Type != TypeNil {
				t.Errorf("NewValue(%s).Type = %v, want TypeNil(不支持类型降级)", tt.name, v.Type)
			}
		})
	}
}

// Test_ValueSystem_NewValue_UnsupportedMaps 不支持的 map 类型降级为 TypeNil
func Test_ValueSystem_NewValue_UnsupportedMaps(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"map[string]int", map[string]int{"a": 1}},
		{"map[string]interface{}", map[string]interface{}{"a": 1}},
		{"map[int]string", map[int]string{1: "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.input)
			if v.Type != TypeNil {
				t.Errorf("NewValue(%s).Type = %v, want TypeNil", tt.name, v.Type)
			}
		})
	}
}

// Test_ValueSystem_NewValue_UnknownType 未知结构体类型降级为 TypeNil
func Test_ValueSystem_NewValue_UnknownType(t *testing.T) {
	type custom struct{ X int }
	v := NewValue(custom{X: 1})
	if !v.IsNil() {
		t.Errorf("未知类型应降级为 TypeNil, got Type=%v", v.Type)
	}
}

// Test_ValueSystem_NewValue_ZeroValues 零值类型正确
func Test_ValueSystem_NewValue_ZeroValues(t *testing.T) {
	t.Run("int零值", func(t *testing.T) {
		if NewValue(0).Type != TypeInt {
			t.Error("NewValue(0) 应为 TypeInt")
		}
	})
	t.Run("string零值", func(t *testing.T) {
		if NewValue("").Type != TypeString {
			t.Error("NewValue(\"\") 应为 TypeString")
		}
	})
	t.Run("bool零值", func(t *testing.T) {
		if NewValue(false).Type != TypeBool {
			t.Error("NewValue(false) 应为 TypeBool")
		}
	})
	t.Run("float零值", func(t *testing.T) {
		if NewValue(0.0).Type != TypeFloat {
			t.Error("NewValue(0.0) 应为 TypeFloat")
		}
	})
}

// ---------- 类型访问器方法 ----------

// Test_ValueSystem_Int_Accessor Int() 正常路径
func Test_ValueSystem_Int_Accessor(t *testing.T) {
	v := NewValue(99)
	if v.Int() != 99 {
		t.Errorf("Int() = %d, want 99", v.Int())
	}
}

// Test_ValueSystem_Int_OnFloat Int() 对 float 值返回零值(无自动转换)
func Test_ValueSystem_Int_OnFloat(t *testing.T) {
	v := NewValue(3.14)
	if v.Int() != 0 {
		t.Errorf("float 的 Int() = %d, want 0 (getTyped 类型不匹配返回零值)", v.Int())
	}
}

// Test_ValueSystem_Int_OnString Int() 对 string 值返回零值
func Test_ValueSystem_Int_OnString(t *testing.T) {
	v := NewValue("123")
	if v.Int() != 0 {
		t.Errorf("string 的 Int() = %d, want 0", v.Int())
	}
}

// Test_ValueSystem_Int_OnBool Int() 对 bool 值返回零值
func Test_ValueSystem_Int_OnBool(t *testing.T) {
	v := NewValue(true)
	if v.Int() != 0 {
		t.Errorf("bool 的 Int() = %d, want 0", v.Int())
	}
}

// Test_ValueSystem_Int_OnNil Int() 对 nil 值返回零值
func Test_ValueSystem_Int_OnNil(t *testing.T) {
	v := NewValue(nil)
	if v.Int() != 0 {
		t.Errorf("nil 的 Int() = %d, want 0", v.Int())
	}
}

// Test_ValueSystem_Float_Accessor Float() 正常路径
func Test_ValueSystem_Float_Accessor(t *testing.T) {
	v := NewValue(2.5)
	if v.Float() != 2.5 {
		t.Errorf("Float() = %f, want 2.5", v.Float())
	}
}

// Test_ValueSystem_Float_OnInt Float() 对 int 值自动转换为float64
func Test_ValueSystem_Float_OnInt(t *testing.T) {
	v := NewValue(42)
	if v.Float() != 42.0 {
		t.Errorf("int 的 Float() = %f, want 42.0 (自动转换)", v.Float())
	}
}

// Test_ValueSystem_Float_OnNil Float() 对 nil 值返回零值
func Test_ValueSystem_Float_OnNil(t *testing.T) {
	v := NewValue(nil)
	if v.Float() != 0 {
		t.Errorf("nil 的 Float() = %f, want 0", v.Float())
	}
}

// Test_ValueSystem_String_Accessor String() 正常路径
func Test_ValueSystem_String_Accessor(t *testing.T) {
	v := NewValue("world")
	if v.String() != "world" {
		t.Errorf("String() = %q, want \"world\"", v.String())
	}
}

// Test_ValueSystem_String_OnInt String() 对 int 值返回空字符串
func Test_ValueSystem_String_OnInt(t *testing.T) {
	v := NewValue(42)
	if v.String() != "" {
		t.Errorf("int 的 String() = %q, want \"\"", v.String())
	}
}

// Test_ValueSystem_Bool_Accessor Bool() 正常路径
func Test_ValueSystem_Bool_Accessor(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		if !NewValue(true).Bool() {
			t.Error("Bool() = false, want true")
		}
	})
	t.Run("false", func(t *testing.T) {
		if NewValue(false).Bool() {
			t.Error("Bool() = true, want false")
		}
	})
}

// Test_ValueSystem_Bool_OnInt Bool() 对 int 值返回 false
func Test_ValueSystem_Bool_OnInt(t *testing.T) {
	v := NewValue(1)
	if v.Bool() {
		t.Error("int 的 Bool() = true, want false")
	}
}

// Test_ValueSystem_Array_Accessor Array() 正常路径
func Test_ValueSystem_Array_Accessor(t *testing.T) {
	elems := []Value{NewValue(1), NewValue(2)}
	v := NewValue(elems)
	arr := v.Array()
	if arr == nil {
		t.Fatal("Array() = nil")
	}
	if len(arr.Elements) != 2 {
		t.Errorf("len = %d, want 2", len(arr.Elements))
	}
	if arr.Elements[1].Int() != 2 {
		t.Errorf("Elements[1] = %d, want 2", arr.Elements[1].Int())
	}
}

// Test_ValueSystem_Array_OnNonArray Array() 对非数组类型返回 nil
func Test_ValueSystem_Array_OnNonArray(t *testing.T) {
	tests := []struct {
		name string
		v    Value
	}{
		{"int", NewValue(1)},
		{"string", NewValue("x")},
		{"nil", NewValue(nil)},
		{"bool", NewValue(true)},
		{"float", NewValue(1.0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.v.Array() != nil {
				t.Errorf("%s 的 Array() 应返回 nil", tt.name)
			}
		})
	}
}

// Test_ValueSystem_Map_Accessor Map() 正常路径
func Test_ValueSystem_Map_Accessor(t *testing.T) {
	mv := &MapValue{Pairs: map[string]Value{"key": NewValue(42)}, KeyType: TypeString}
	v := NewValue(mv)
	result := v.Map()
	if result == nil {
		t.Fatal("Map() = nil")
	}
	if result.Pairs["key"].Int() != 42 {
		t.Errorf("Pairs[\"key\"] = %d, want 42", result.Pairs["key"].Int())
	}
}

// Test_ValueSystem_Map_OnNonMap Map() 对非 Map 类型返回 nil
func Test_ValueSystem_Map_OnNonMap(t *testing.T) {
	tests := []struct {
		name string
		v    Value
	}{
		{"int", NewValue(1)},
		{"string", NewValue("x")},
		{"array", NewValue([]Value{NewValue(1)})},
		{"nil", NewValue(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.v.Map() != nil {
				t.Errorf("%s 的 Map() 应返回 nil", tt.name)
			}
		})
	}
}

// Test_ValueSystem_Function_Accessor Function() 正常路径
func Test_ValueSystem_Function_Accessor(t *testing.T) {
	fv := &FunctionValue{Compiled: &CompiledFunction{Name: "myFn"}}
	v := NewValue(fv)
	if v.Function() == nil {
		t.Fatal("Function() = nil")
	}
	if v.Function().Compiled.Name != "myFn" {
		t.Errorf("Compiled.Name = %q, want \"myFn\"", v.Function().Compiled.Name)
	}
}

// Test_ValueSystem_Function_OnNonFunction Function() 对非函数类型返回 nil
func Test_ValueSystem_Function_OnNonFunction(t *testing.T) {
	v := NewValue(42)
	if v.Function() != nil {
		t.Error("int 的 Function() 应返回 nil")
	}
}

// Test_ValueSystem_ExternalFunc_Accessor ExternalFunc() 正常路径
func Test_ValueSystem_ExternalFunc_Accessor(t *testing.T) {
	ef := &ExternalFuncValue{Name: "ext", Func: func() {}}
	v := Value{Type: TypeExternalFunc, Data: ef}
	result := v.ExternalFunc()
	if result == nil {
		t.Fatal("ExternalFunc() = nil")
	}
	if result.Name != "ext" {
		t.Errorf("Name = %q, want \"ext\"", result.Name)
	}
}

// Test_ValueSystem_ExternalFunc_OnNonExternalFunc ExternalFunc() 对非外部函数类型返回 nil
func Test_ValueSystem_ExternalFunc_OnNonExternalFunc(t *testing.T) {
	v := NewValue(42)
	if v.ExternalFunc() != nil {
		t.Error("int 的 ExternalFunc() 应返回 nil")
	}
}

// ---------- Equal 相等性比较 ----------

// Test_ValueSystem_Equal_Int int 相等比较
func Test_ValueSystem_Equal_Int(t *testing.T) {
	t.Run("相同值", func(t *testing.T) {
		if !NewValue(1).Equal(NewValue(1)) {
			t.Error("1 == 1 应为 true")
		}
	})
	t.Run("不同值", func(t *testing.T) {
		if NewValue(1).Equal(NewValue(2)) {
			t.Error("1 == 2 应为 false")
		}
	})
}

// Test_ValueSystem_Equal_Float float 相等比较
func Test_ValueSystem_Equal_Float(t *testing.T) {
	t.Run("相同值", func(t *testing.T) {
		if !NewValue(1.5).Equal(NewValue(1.5)) {
			t.Error("1.5 == 1.5 应为 true")
		}
	})
	t.Run("不同值", func(t *testing.T) {
		if NewValue(1.5).Equal(NewValue(2.5)) {
			t.Error("1.5 == 2.5 应为 false")
		}
	})
}

// Test_ValueSystem_Equal_String string 相等比较
func Test_ValueSystem_Equal_String(t *testing.T) {
	t.Run("相同值", func(t *testing.T) {
		if !NewValue("abc").Equal(NewValue("abc")) {
			t.Error("\"abc\" == \"abc\" 应为 true")
		}
	})
	t.Run("不同值", func(t *testing.T) {
		if NewValue("abc").Equal(NewValue("abd")) {
			t.Error("\"abc\" == \"abd\" 应为 false")
		}
	})
}

// Test_ValueSystem_Equal_Bool bool 相等比较
func Test_ValueSystem_Equal_Bool(t *testing.T) {
	t.Run("相同值true", func(t *testing.T) {
		if !NewValue(true).Equal(NewValue(true)) {
			t.Error("true == true 应为 true")
		}
	})
	t.Run("不同值", func(t *testing.T) {
		if NewValue(true).Equal(NewValue(false)) {
			t.Error("true == false 应为 false")
		}
	})
}

// Test_ValueSystem_Equal_Nil nil 相等比较
func Test_ValueSystem_Equal_Nil(t *testing.T) {
	if !NewValue(nil).Equal(NewValue(nil)) {
		t.Error("nil == nil 应为 true")
	}
}

// Test_ValueSystem_Equal_CrossType 跨类型比较均为 false
// 注意: int/float数值比较自动提升, 不在此测试中
func Test_ValueSystem_Equal_CrossType(t *testing.T) {
	tests := []struct {
		name string
		a, b Value
	}{
		{"int vs string", NewValue(1), NewValue("1")},
		{"int vs bool", NewValue(1), NewValue(true)},
		{"int vs nil", NewValue(0), NewValue(nil)},
		{"string vs bool", NewValue("true"), NewValue(true)},
		{"float vs string", NewValue(1.0), NewValue("1.0")},
		{"nil vs int", NewValue(nil), NewValue(0)},
		{"array vs map", NewValue([]Value{}), NewValue(&MapValue{Pairs: map[string]Value{}})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.a.Equal(tt.b) {
				t.Error("跨类型比较应为 false")
			}
		})
	}
}

// Test_ValueSystem_Equal_Array 数组相等比较
func Test_ValueSystem_Equal_Array(t *testing.T) {
	t.Run("元素相同", func(t *testing.T) {
		a := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
		b := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
		if !a.Equal(b) {
			t.Error("元素相同的数组应相等")
		}
	})
	t.Run("元素不同", func(t *testing.T) {
		a := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
		b := NewValue([]Value{NewValue(1), NewValue(2), NewValue(4)})
		if a.Equal(b) {
			t.Error("元素不同的数组不应相等")
		}
	})
	t.Run("长度不同", func(t *testing.T) {
		a := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
		b := NewValue([]Value{NewValue(1), NewValue(2)})
		if a.Equal(b) {
			t.Error("长度不同的数组不应相等")
		}
	})
}

// Test_ValueSystem_Equal_Map Map 相等比较
func Test_ValueSystem_Equal_Map(t *testing.T) {
	t.Run("相同键值", func(t *testing.T) {
		a := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1), "b": NewValue(2)}, KeyType: TypeString})
		b := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1), "b": NewValue(2)}, KeyType: TypeString})
		if !a.Equal(b) {
			t.Error("相同键值的 Map 应相等")
		}
	})
	t.Run("不同键", func(t *testing.T) {
		a := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}, KeyType: TypeString})
		b := NewValue(&MapValue{Pairs: map[string]Value{"b": NewValue(1)}, KeyType: TypeString})
		if a.Equal(b) {
			t.Error("键不同的 Map 不应相等")
		}
	})
	t.Run("不同值", func(t *testing.T) {
		a := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}, KeyType: TypeString})
		b := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(2)}, KeyType: TypeString})
		if a.Equal(b) {
			t.Error("值不同的 Map 不应相等")
		}
	})
}

// Test_ValueSystem_Equal_NestedArray 嵌套数组递归比较
func Test_ValueSystem_Equal_NestedArray(t *testing.T) {
	a := NewValue([]Value{
		NewValue([]Value{NewValue(1), NewValue(2)}),
		NewValue([]Value{NewValue(3), NewValue(4)}),
	})
	b := NewValue([]Value{
		NewValue([]Value{NewValue(1), NewValue(2)}),
		NewValue([]Value{NewValue(3), NewValue(4)}),
	})
	if !a.Equal(b) {
		t.Error("嵌套数组内容相同应递归相等")
	}
	c := NewValue([]Value{
		NewValue([]Value{NewValue(1), NewValue(99)}),
		NewValue([]Value{NewValue(3), NewValue(4)}),
	})
	if a.Equal(c) {
		t.Error("嵌套数组内层元素不同不应相等")
	}
}

// Test_ValueSystem_Equal_NestedMap 嵌套 Map 递归比较
func Test_ValueSystem_Equal_NestedMap(t *testing.T) {
	a := NewValue(&MapValue{
		Pairs: map[string]Value{
			"inner": NewValue(&MapValue{Pairs: map[string]Value{"x": NewValue(1)}, KeyType: TypeString}),
		},
		KeyType: TypeString,
	})
	b := NewValue(&MapValue{
		Pairs: map[string]Value{
			"inner": NewValue(&MapValue{Pairs: map[string]Value{"x": NewValue(1)}, KeyType: TypeString}),
		},
		KeyType: TypeString,
	})
	if !a.Equal(b) {
		t.Error("嵌套 Map 内容相同应递归相等")
	}
}

// Test_ValueSystem_Equal_NilVsOther nil 与其他类型比较
func Test_ValueSystem_Equal_NilVsOther(t *testing.T) {
	others := []Value{
		NewValue(0),
		NewValue(""),
		NewValue(false),
		NewValue(0.0),
	}
	for i, other := range others {
		if NewValue(nil).Equal(other) {
			t.Errorf("nil 与类型 %v 不应相等 (case %d)", other.Type, i)
		}
	}
}

// Test_ValueSystem_Equal_Function 函数类型无相等处理器, 同类型比较返回 false
func Test_ValueSystem_Equal_Function(t *testing.T) {
	fv := &FunctionValue{Compiled: &CompiledFunction{Name: "f"}}
	a := NewValue(fv)
	b := NewValue(fv)
	// TypeFunction 未注册 equalHandler, 即使指向同一指针也返回 false
	if a.Equal(b) {
		t.Error("TypeFunction 无 Equal 处理器, 同类型比较应返回 false")
	}
}

// ---------- GoString 格式化 ----------

// Test_ValueSystem_GoString_Int int 的 GoString
func Test_ValueSystem_GoString_Int(t *testing.T) {
	if NewValue(42).GoString() != "42" {
		t.Errorf("GoString() = %q, want \"42\"", NewValue(42).GoString())
	}
}

// Test_ValueSystem_GoString_Float float 的 GoString
func Test_ValueSystem_GoString_Float(t *testing.T) {
	v := NewValue(3.14)
	want := strconv.FormatFloat(3.14, 'f', -1, 64)
	if v.GoString() != want {
		t.Errorf("GoString() = %q, want %q", v.GoString(), want)
	}
}

// Test_ValueSystem_GoString_String string 的 GoString (带引号)
func Test_ValueSystem_GoString_String(t *testing.T) {
	if NewValue("hello").GoString() != "\"hello\"" {
		t.Errorf("GoString() = %q, want \"\\\"hello\\\"\"", NewValue("hello").GoString())
	}
}

// Test_ValueSystem_GoString_Bool bool 的 GoString
func Test_ValueSystem_GoString_Bool(t *testing.T) {
	if NewValue(true).GoString() != "true" {
		t.Errorf("GoString() = %q, want \"true\"", NewValue(true).GoString())
	}
	if NewValue(false).GoString() != "false" {
		t.Errorf("GoString() = %q, want \"false\"", NewValue(false).GoString())
	}
}

// Test_ValueSystem_GoString_Nil nil 的 GoString
func Test_ValueSystem_GoString_Nil(t *testing.T) {
	if NewValue(nil).GoString() != "nil" {
		t.Errorf("GoString() = %q, want \"nil\"", NewValue(nil).GoString())
	}
}

// Test_ValueSystem_GoString_Array 数组的 GoString
func Test_ValueSystem_GoString_Array(t *testing.T) {
	v := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
	if v.GoString() != "[1, 2, 3]" {
		t.Errorf("GoString() = %q, want \"[1, 2, 3]\"", v.GoString())
	}
}

// Test_ValueSystem_GoString_EmptyArray 空数组的 GoString
func Test_ValueSystem_GoString_EmptyArray(t *testing.T) {
	v := NewValue([]Value{})
	if v.GoString() != "[]" {
		t.Errorf("GoString() = %q, want \"[]\"", v.GoString())
	}
}

// Test_ValueSystem_GoString_Map 单键 Map 的 GoString (多键顺序不确定)
func Test_ValueSystem_GoString_Map(t *testing.T) {
	v := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}, KeyType: TypeString})
	// 单键确定: {"a": 1}
	want := "{\"a\": 1}"
	if v.GoString() != want {
		t.Errorf("GoString() = %q, want %q", v.GoString(), want)
	}
}

// Test_ValueSystem_GoString_EmptyMap 空 Map 的 GoString
func Test_ValueSystem_GoString_EmptyMap(t *testing.T) {
	v := NewValue(&MapValue{Pairs: map[string]Value{}, KeyType: TypeString})
	if v.GoString() != "{}" {
		t.Errorf("GoString() = %q, want \"{}\"", v.GoString())
	}
}

// Test_ValueSystem_GoString_NestedArray 嵌套数组的 GoString
func Test_ValueSystem_GoString_NestedArray(t *testing.T) {
	inner := &ArrayValue{Elements: []Value{NewValue(1), NewValue(2)}}
	outer := &ArrayValue{Elements: []Value{
		{Type: TypeArray, Data: inner},
		NewValue(3),
	}}
	v := Value{Type: TypeArray, Data: outer}
	if v.GoString() != "[[1, 2], 3]" {
		t.Errorf("GoString() = %q, want \"[[1, 2], 3]\"", v.GoString())
	}
}

// Test_ValueSystem_GoString_NestedMap 嵌套 Map 的 GoString
func Test_ValueSystem_GoString_NestedMap(t *testing.T) {
	inner := &MapValue{Pairs: map[string]Value{"y": NewValue(2)}, KeyType: TypeString}
	outer := &MapValue{
		Pairs:   map[string]Value{"inner": {Type: TypeMap, Data: inner}},
		KeyType: TypeString,
	}
	v := Value{Type: TypeMap, Data: outer}
	got := v.GoString()
	// 单键嵌套, 输出确定
	want := "{\"inner\": {\"y\": 2}}"
	if got != want {
		t.Errorf("GoString() = %q, want %q", got, want)
	}
}

// Test_ValueSystem_GoString_Function 函数的 GoString
func Test_ValueSystem_GoString_Function(t *testing.T) {
	v := NewValue(&FunctionValue{Compiled: &CompiledFunction{Name: "testFunc"}})
	if v.GoString() != "<function testFunc>" {
		t.Errorf("GoString() = %q, want \"<function testFunc>\"", v.GoString())
	}
}

// Test_ValueSystem_GoString_ExternalFunc 外部函数的 GoString
func Test_ValueSystem_GoString_ExternalFunc(t *testing.T) {
	ef := &ExternalFuncValue{Name: "extFunc", Func: func() {}}
	v := Value{Type: TypeExternalFunc, Data: ef}
	if v.GoString() != "<external extFunc>" {
		t.Errorf("GoString() = %q, want \"<external extFunc>\"", v.GoString())
	}
}

// Test_ValueSystem_GoString_UnknownType 未知类型的 GoString
func Test_ValueSystem_GoString_UnknownType(t *testing.T) {
	v := Value{Type: ValueType(999), Data: nil}
	if v.GoString() != "<unknown>" {
		t.Errorf("GoString() = %q, want \"<unknown>\"", v.GoString())
	}
}

// ---------- intVal / intValSlow 与类型边界 ----------

// Test_ValueSystem_IntVal_CacheHit 0-255 命中缓存
func Test_ValueSystem_IntVal_CacheHit(t *testing.T) {
	for i := 0; i < 256; i++ {
		v := intVal(i)
		if v.Type != TypeInt {
			t.Errorf("intVal(%d).Type = %v, want TypeInt", i, v.Type)
		}
		if v.Int() != i {
			t.Errorf("intVal(%d).Int() = %d, want %d", i, v.Int(), i)
		}
	}
}

// Test_ValueSystem_IntVal_CacheMiss 256 走 intValSlow
func Test_ValueSystem_IntVal_CacheMiss(t *testing.T) {
	v := intVal(256)
	if v.Int() != 256 {
		t.Errorf("intVal(256).Int() = %d, want 256", v.Int())
	}
}

// Test_ValueSystem_IntVal_Negative 负数走 intValSlow
func Test_ValueSystem_IntVal_Negative(t *testing.T) {
	v := intVal(-1)
	if v.Int() != -1 {
		t.Errorf("intVal(-1).Int() = %d, want -1", v.Int())
	}
}

// Test_ValueSystem_IntValSlow 直接构造 Value
func Test_ValueSystem_IntValSlow(t *testing.T) {
	v := intValSlow(1000)
	if v.Type != TypeInt || v.Int() != 1000 {
		t.Errorf("intValSlow(1000) = {Type:%v, Int:%d}, want TypeInt/1000", v.Type, v.Int())
	}
}

// Test_ValueSystem_Boundary_LargeInt 大整数
func Test_ValueSystem_Boundary_LargeInt(t *testing.T) {
	// 平台宽度下的大值, 均超过小整数缓存
	big := 1<<(strconv.IntSize-2) + 1
	v := NewValue(big)
	if v.Int() != big {
		t.Errorf("Int() = %d, want %d", v.Int(), big)
	}
	if v.Type != TypeInt {
		t.Errorf("Type = %v, want TypeInt", v.Type)
	}
}

// Test_ValueSystem_Boundary_NegativeInt 负数
func Test_ValueSystem_Boundary_NegativeInt(t *testing.T) {
	v := NewValue(-500)
	if v.Int() != -500 {
		t.Errorf("Int() = %d, want -500", v.Int())
	}
}

// Test_ValueSystem_Boundary_EmptyString 空字符串
func Test_ValueSystem_Boundary_EmptyString(t *testing.T) {
	v := NewValue("")
	if v.Type != TypeString {
		t.Errorf("Type = %v, want TypeString", v.Type)
	}
	if v.String() != "" {
		t.Errorf("String() = %q, want \"\"", v.String())
	}
	if v.IsNil() {
		t.Error("空字符串不应 IsNil")
	}
}

// Test_ValueSystem_Boundary_EmptyArray 空数组
func Test_ValueSystem_Boundary_EmptyArray(t *testing.T) {
	v := NewValue([]Value{})
	if v.Type != TypeArray {
		t.Errorf("Type = %v, want TypeArray", v.Type)
	}
	arr := v.Array()
	if arr == nil || len(arr.Elements) != 0 {
		t.Errorf("空数组 Elements 长度应 0, got %v", arr)
	}
}

// Test_ValueSystem_Boundary_EmptyMap 空 Map
func Test_ValueSystem_Boundary_EmptyMap(t *testing.T) {
	v := NewValue(&MapValue{Pairs: map[string]Value{}, KeyType: TypeString})
	if v.Type != TypeMap {
		t.Errorf("Type = %v, want TypeMap", v.Type)
	}
	mv := v.Map()
	if mv == nil || len(mv.Pairs) != 0 {
		t.Errorf("空 Map Pairs 长度应 0, got %v", mv)
	}
}

// ---------- 类型转换链 (验证无自动转换) ----------

// Test_ValueSystem_Conversion_IntToFloat int.Float() 自动转换为float64
func Test_ValueSystem_Conversion_IntToFloat(t *testing.T) {
	v := NewValue(42)
	if v.Float() != 42.0 {
		t.Errorf("int 的 Float() = %f, want 42.0 (自动转换)", v.Float())
	}
}

// Test_ValueSystem_Conversion_FloatToInt float.Int() 不自动转换, 返回零值
func Test_ValueSystem_Conversion_FloatToInt(t *testing.T) {
	v := NewValue(3.99)
	if v.Int() != 0 {
		t.Errorf("float 的 Int() = %d, want 0 (无自动转换, 不截断也不四舍五入)", v.Int())
	}
}

// Test_ValueSystem_Conversion_BoolToInt bool.Int() 不自动转换, 返回零值
func Test_ValueSystem_Conversion_BoolToInt(t *testing.T) {
	if NewValue(true).Int() != 0 {
		t.Error("bool 的 Int() 应返回 0 (无自动转换)")
	}
}

// Test_ValueSystem_Conversion_StringToBool string.Bool() 不自动转换, 返回 false
func Test_ValueSystem_Conversion_StringToBool(t *testing.T) {
	if NewValue("true").Bool() {
		t.Error("string 的 Bool() 应返回 false (无自动转换)")
	}
}

// ---------- 复合值操作 ----------

// Test_ValueSystem_Composite_ArrayElementAccess 数组元素访问
func Test_ValueSystem_Composite_ArrayElementAccess(t *testing.T) {
	v := NewValue([]Value{NewValue(10), NewValue(20), NewValue(30)})
	arr := v.Array()
	t.Run("首元素", func(t *testing.T) {
		if arr.Elements[0].Int() != 10 {
			t.Errorf("Elements[0] = %d, want 10", arr.Elements[0].Int())
		}
	})
	t.Run("尾元素", func(t *testing.T) {
		if arr.Elements[2].Int() != 30 {
			t.Errorf("Elements[2] = %d, want 30", arr.Elements[2].Int())
		}
	})
}

// Test_ValueSystem_Composite_MapKVAccess Map 键值访问
func Test_ValueSystem_Composite_MapKVAccess(t *testing.T) {
	mv := &MapValue{
		Pairs:   map[string]Value{"name": NewValue("alice"), "age": NewValue(30)},
		KeyType: TypeString,
	}
	v := NewValue(mv)
	m := v.Map()
	t.Run("字符串值", func(t *testing.T) {
		if m.Pairs["name"].String() != "alice" {
			t.Errorf("Pairs[name] = %q, want \"alice\"", m.Pairs["name"].String())
		}
	})
	t.Run("整数值", func(t *testing.T) {
		if m.Pairs["age"].Int() != 30 {
			t.Errorf("Pairs[age] = %d, want 30", m.Pairs["age"].Int())
		}
	})
}

// Test_ValueSystem_Composite_NestedArray 嵌套数组
func Test_ValueSystem_Composite_NestedArray(t *testing.T) {
	inner := []Value{NewValue(1), NewValue(2)}
	outer := []Value{NewValue(inner), NewValue(3)}
	v := NewValue(outer)
	arr := v.Array()
	innerArr := arr.Elements[0].Array()
	if innerArr == nil {
		t.Fatal("嵌套内层 Array() = nil")
	}
	if innerArr.Elements[1].Int() != 2 {
		t.Errorf("内层 Elements[1] = %d, want 2", innerArr.Elements[1].Int())
	}
}

// Test_ValueSystem_Composite_NestedMap 嵌套 Map
func Test_ValueSystem_Composite_NestedMap(t *testing.T) {
	inner := &MapValue{Pairs: map[string]Value{"deep": NewValue(99)}, KeyType: TypeString}
	outer := &MapValue{
		Pairs:   map[string]Value{"nested": NewValue(inner)},
		KeyType: TypeString,
	}
	v := NewValue(outer)
	innerMap := v.Map().Pairs["nested"].Map()
	if innerMap == nil {
		t.Fatal("嵌套内层 Map() = nil")
	}
	if innerMap.Pairs["deep"].Int() != 99 {
		t.Errorf("内层 Pairs[deep] = %d, want 99", innerMap.Pairs["deep"].Int())
	}
}

// Test_ValueSystem_Composite_ArrayInMap Map 值为数组
func Test_ValueSystem_Composite_ArrayInMap(t *testing.T) {
	mv := &MapValue{
		Pairs: map[string]Value{
			"list": NewValue([]Value{NewValue(1), NewValue(2)}),
		},
		KeyType: TypeString,
	}
	v := NewValue(mv)
	arr := v.Map().Pairs["list"].Array()
	if arr == nil || len(arr.Elements) != 2 {
		t.Errorf("Map 中的数组元素访问失败, arr=%v", arr)
	}
}

// Test_ValueSystem_Composite_MapInArray 数组元素为 Map
func Test_ValueSystem_Composite_MapInArray(t *testing.T) {
	mv := &MapValue{Pairs: map[string]Value{"k": NewValue(1)}, KeyType: TypeString}
	v := NewValue([]Value{NewValue(mv), NewValue(42)})
	arr := v.Array()
	mapInArray := arr.Elements[0].Map()
	if mapInArray == nil {
		t.Fatal("数组中的 Map 元素访问失败")
	}
	if mapInArray.Pairs["k"].Int() != 1 {
		t.Errorf("Pairs[k] = %d, want 1", mapInArray.Pairs["k"].Int())
	}
}

// ---------- formatValueForPrint / formatValue ----------

// Test_ValueSystem_FormatValueForPrint_AllTypes 各类型打印格式
func Test_ValueSystem_FormatValueForPrint_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want string
	}{
		{"nil", NewValue(nil), "nil"},
		{"int", NewValue(7), "7"},
		{"float", NewValue(2.5), "2.5"},
		{"string", NewValue("hi"), "\"hi\""},
		{"bool", NewValue(true), "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValueForPrint(tt.v)
			if got != tt.want {
				t.Errorf("formatValueForPrint() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Test_ValueSystem_FormatValueForPrint_Nested 嵌套结构打印
func Test_ValueSystem_FormatValueForPrint_Nested(t *testing.T) {
	v := NewValue([]Value{
		NewValue(1),
		NewValue([]Value{NewValue(2), NewValue(3)}),
	})
	got := formatValueForPrint(v)
	if got != "[1, [2, 3]]" {
		t.Errorf("formatValueForPrint() = %q, want \"[1, [2, 3]]\"", got)
	}
}

// Test_ValueSystem_FormatValueForPrint_EmptyCollections 空集合打印
func Test_ValueSystem_FormatValueForPrint_EmptyCollections(t *testing.T) {
	t.Run("空数组", func(t *testing.T) {
		v := NewValue([]Value{})
		if formatValueForPrint(v) != "[]" {
			t.Errorf("空数组打印应 \"[]\"")
		}
	})
	t.Run("空Map", func(t *testing.T) {
		v := NewValue(&MapValue{Pairs: map[string]Value{}, KeyType: TypeString})
		if formatValueForPrint(v) != "{}" {
			t.Errorf("空 Map 打印应 \"{}\"")
		}
	})
}

// Test_ValueSystem_FormatValue_SecondaryFunc formatValue(any) 函数
func Test_ValueSystem_FormatValue_SecondaryFunc(t *testing.T) {
	t.Run("string带引号", func(t *testing.T) {
		if formatValue("hello") != "\"hello\"" {
			t.Errorf("formatValue(string) 应加引号, got %q", formatValue("hello"))
		}
	})
	t.Run("int", func(t *testing.T) {
		if formatValue(42) != "42" {
			t.Errorf("formatValue(int) = %q, want \"42\"", formatValue(42))
		}
	})
	t.Run("float", func(t *testing.T) {
		if formatValue(3.14) != "3.14" {
			t.Errorf("formatValue(float) = %q, want \"3.14\"", formatValue(3.14))
		}
	})
}

// ---------- IsNil 补充 ----------

// Test_ValueSystem_IsNil_AllTypes 各类型 IsNil 行为
func Test_ValueSystem_IsNil_AllTypes(t *testing.T) {
	t.Run("nil为true", func(t *testing.T) {
		if !NewValue(nil).IsNil() {
			t.Error("nil IsNil 应为 true")
		}
	})
	t.Run("int零值不为nil", func(t *testing.T) {
		if NewValue(0).IsNil() {
			t.Error("0 不应 IsNil")
		}
	})
	t.Run("空字符串不为nil", func(t *testing.T) {
		if NewValue("").IsNil() {
			t.Error("\"\" 不应 IsNil")
		}
	})
	t.Run("false不为nil", func(t *testing.T) {
		if NewValue(false).IsNil() {
			t.Error("false 不应 IsNil")
		}
	})
	t.Run("空数组不为nil", func(t *testing.T) {
		if NewValue([]Value{}).IsNil() {
			t.Error("空数组不应 IsNil")
		}
	})
	t.Run("空Map不为nil", func(t *testing.T) {
		if NewValue(&MapValue{Pairs: map[string]Value{}}).IsNil() {
			t.Error("空 Map 不应 IsNil")
		}
	})
}

// ---------- MultiKey Map GoString 格式验证 ----------

// Test_ValueSystem_GoString_MultiKeyMap 多键 Map GoString 包含所有键值对
func Test_ValueSystem_GoString_MultiKeyMap(t *testing.T) {
	v := NewValue(&MapValue{
		Pairs: map[string]Value{
			"a": NewValue(1),
			"b": NewValue(2),
		},
		KeyType: TypeString,
	})
	got := v.GoString()
	// Map 迭代顺序不确定, 验证包含所有键值对
	if !strings.Contains(got, `"a": 1`) {
		t.Errorf("GoString 应包含 \"a\": 1, got %q", got)
	}
	if !strings.Contains(got, `"b": 2`) {
		t.Errorf("GoString 应包含 \"b\": 2, got %q", got)
	}
	if !strings.HasPrefix(got, "{") || !strings.HasSuffix(got, "}") {
		t.Errorf("GoString 应以 {} 包裹, got %q", got)
	}
}
