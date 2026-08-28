package script

import (
	"testing"
)

// ========== 索引访问测试 ==========

func Test_runtime_IndexAccess(t *testing.T) {
	vm := newTestVM(t)

	// 数组索引访问
	arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
	result, err := vm.indexAccess(arr, NewValue(1))
	if err != nil || result.Int() != 2 {
		t.Errorf("数组索引访问失败: %v", err)
	}

	// 字符串索引访问
	str := NewValue("hello")
	result, err = vm.indexAccess(str, NewValue(0))
	if err != nil || result.String() != "h" {
		t.Errorf("字符串索引访问失败: %v", err)
	}

	// Map索引访问
	m := &MapValue{Pairs: map[string]Value{"key": NewValue("value")}}
	result, err = vm.indexAccess(NewValue(m), NewValue("key"))
	if err != nil || result.String() != "value" {
		t.Errorf("Map索引访问失败: %v", err)
	}
}

func Test_runtime_SliceAccess(t *testing.T) {
	vm := newTestVM(t)

	// 数组切片
	arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3), NewValue(4)})
	result, err := vm.sliceAccess(arr, NewValue(1), NewValue(3))
	if err != nil {
		t.Errorf("数组切片失败: %v", err)
		return
	}
	slice := result.Array()
	if len(slice.Elements) != 2 {
		t.Errorf("切片长度错误: 期望2, 得到%d", len(slice.Elements))
	}

	// 字符串切片
	str := NewValue("hello world")
	result, err = vm.sliceAccess(str, NewValue(0), NewValue(5))
	if err != nil || result.String() != "hello" {
		t.Errorf("字符串切片失败: got=%s, want=hello", result.String())
	}
}

func Test_runtime_Length(t *testing.T) {
	vm := newTestVM(t)

	// 数组长度
	arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
	if n, _ := vm.length(arr); n != 3 {
		t.Errorf("数组长度错误")
	}

	// 字符串长度
	str := NewValue("hello")
	if n, _ := vm.length(str); n != 5 {
		t.Errorf("字符串长度错误")
	}

	// Map长度
	m := &MapValue{Pairs: map[string]Value{"a": NewValue(1), "b": NewValue(2)}}
	if n, _ := vm.length(NewValue(m)); n != 2 {
		t.Errorf("Map长度错误")
	}
}

func Test_runtime_MapAccess(t *testing.T) {
	t.Run("获取已存在的键", func(t *testing.T) {
		runIntTest(t, `m := {"a": 100, "b": 200}
m["a"]`, 100)
	})

	t.Run("获取不存在的键返回nil", func(t *testing.T) {
		result := runScript(t, `m := {"a": 100}
m["b"]`)
		if !result.IsNil() {
			t.Errorf("期望 nil，得到 %v", result)
		}
	})
}

func Test_runtime_ArrayIndexAccess2(t *testing.T) {
	t.Run("正常访问", func(t *testing.T) {
		runIntTest(t, `arr := [10, 20, 30]
arr[1]`, 20)
	})

	t.Run("越界访问返回nil", func(t *testing.T) {
		result := runScript(t, `arr := [10, 20, 30]
arr[10]`)
		if !result.IsNil() {
			t.Errorf("期望 nil，得到 %v", result)
		}
	})
}

func Test_runtime_StringIndexAccess_EdgeCases(t *testing.T) {
	vm := newTestVM(t)

	t.Run("越界索引返回nil", func(t *testing.T) {
		str := NewValue("hello")
		result, err := vm.indexAccess(str, NewValue(100))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if !result.IsNil() {
			t.Errorf("越界索引应返回nil，得到: %v", result)
		}
	})

	t.Run("空字符串越界", func(t *testing.T) {
		str := NewValue("")
		result, err := vm.indexAccess(str, NewValue(0))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if !result.IsNil() {
			t.Errorf("空字符串索引应返回nil")
		}
	})
}

func Test_runtime_SliceAccess_EdgeCases(t *testing.T) {
	vm := newTestVM(t)

	t.Run("省略结束索引使用对象长度", func(t *testing.T) {
		arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
		result, err := vm.sliceAccess(arr, NewValue(1), NewValue(SliceEndDefault))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		elements := result.Array().Elements
		if len(elements) != 2 {
			t.Errorf("期望2个元素，得到%d", len(elements))
		}
	})

	t.Run("字符串切片正常", func(t *testing.T) {
		str := NewValue("hello world")
		result, err := vm.sliceAccess(str, NewValue(0), NewValue(5))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		if result.String() != "hello" {
			t.Errorf("期望hello，得到%s", result.String())
		}
	})

	t.Run("空数组切片", func(t *testing.T) {
		arr := NewValue([]Value{})
		result, err := vm.sliceAccess(arr, NewValue(0), NewValue(SliceEndDefault))
		if err != nil {
			t.Errorf("不应返回错误: %v", err)
		}
		elements := result.Array().Elements
		if len(elements) != 0 {
			t.Errorf("期望0个元素，得到%d", len(elements))
		}
	})
}

func Test_runtime_Length_EdgeCases(t *testing.T) {
	vm := newTestVM(t)

	tests := []struct {
		name        string
		value       Value
		expectError bool
	}{
		{"整数报错", NewValue(42), true},
		{"布尔报错", NewValue(true), true},
		{"nil报错", NewValue(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.length(tt.value)
			if tt.expectError && err == nil {
				t.Errorf("期望错误，但未返回错误")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望错误，但返回: %v", err)
			}
		})
	}
}
