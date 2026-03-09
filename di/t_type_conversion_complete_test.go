package di

import (
	"reflect"
	"testing"
)

// 定义测试用的类型和方法（在包级别）
type testStringer interface {
	String() string
}

type testMyString string

func (ms testMyString) String() string {
	return string(ms)
}

type testMyStruct struct {
	value string
}

func (m *testMyStruct) String() string {
	return m.value
}

// 测试 tryDirectOrInterfaceMatch 更多场景
func TestTryDirectOrInterfaceMatch(t *testing.T) {
	// 测试接口类型
	field := reflect.ValueOf(new(testStringer)).Elem()
	value := testMyString("test")
	valueReflect := reflect.ValueOf(value)

	result := tryDirectOrInterfaceMatch(field, valueReflect, field.Type(), reflect.TypeOf(value))
	if !result {
		t.Error("Expected interface match to succeed")
	}
}

func TestTryDirectOrInterfaceMatchPointerType(t *testing.T) {
	// 测试指针类型实现接口
	field := reflect.ValueOf(new(testStringer)).Elem()
	value := &testMyStruct{value: "test"}
	valueReflect := reflect.ValueOf(value)

	result := tryDirectOrInterfaceMatch(field, valueReflect, field.Type(), reflect.TypeOf(value))
	if !result {
		t.Error("Expected pointer type interface match to succeed")
	}
}

func TestTryDirectOrInterfaceMatchFailure(t *testing.T) {
	field := reflect.ValueOf(new(int)).Elem()
	value := "string"
	valueReflect := reflect.ValueOf(value)

	result := tryDirectOrInterfaceMatch(field, valueReflect, field.Type(), reflect.TypeOf(value))
	if result {
		t.Error("Expected match to fail for incompatible types")
	}
}

// 测试 tryPointerConversion
func TestTryPointerConversion(t *testing.T) {
	t.Run("Value to Pointer", func(t *testing.T) {
		// 创建可寻址的字段
		original := 42
		field := reflect.ValueOf(&original)

		// 值类型
		value := 100
		valueReflect := reflect.ValueOf(value)

		result := tryPointerConversion(field.Elem(), valueReflect, field.Elem().Type(), reflect.TypeOf(value))
		if result {
			t.Error("Expected pointer conversion to not apply for value field")
		}
	})

	t.Run("Pointer to Value", func(t *testing.T) {
		// 字段是值类型
		field := reflect.ValueOf(new(int)).Elem()

		// 值是指针类型
		value := 42
		valuePtr := &value
		valueReflect := reflect.ValueOf(valuePtr)

		result := tryPointerConversion(field, valueReflect, field.Type(), reflect.TypeOf(valuePtr))
		if !result {
			t.Error("Expected pointer dereference to succeed")
		}
		if field.Int() != 42 {
			t.Errorf("Expected field to be 42, got %d", field.Int())
		}
	})
}

func TestTryPointerConversionNonAddressable(t *testing.T) {
	// 测试不可寻址的值创建新指针
	field := reflect.ValueOf(new(*int)).Elem()

	// 不可寻址的值
	value := 42
	valueReflect := reflect.ValueOf(value)

	result := tryPointerConversion(field, valueReflect, field.Type(), reflect.TypeOf(value))
	if !result {
		t.Error("Expected pointer conversion to succeed")
	}
}

// 测试 tryConvertibleOrSmartConversion
func TestTryConvertibleOrSmartConversion(t *testing.T) {
	t.Run("Convertible Type", func(t *testing.T) {
		field := reflect.ValueOf(new(float64)).Elem()
		value := int(42)
		valueReflect := reflect.ValueOf(value)

		result := tryConvertibleOrSmartConversion(field, valueReflect, field.Type(), reflect.TypeOf(value))
		if !result {
			t.Error("Expected convertible type conversion to succeed")
		}
		if field.Float() != 42.0 {
			t.Errorf("Expected field to be 42.0, got %f", field.Float())
		}
	})

	t.Run("Smart Conversion", func(t *testing.T) {
		field := reflect.ValueOf(new(int)).Elem()
		value := "42"
		valueReflect := reflect.ValueOf(value)

		result := tryConvertibleOrSmartConversion(field, valueReflect, field.Type(), reflect.TypeOf(value))
		if !result {
			t.Error("Expected smart conversion to succeed")
		}
		if field.Int() != 42 {
			t.Errorf("Expected field to be 42, got %d", field.Int())
		}
	})
}

// 测试 setFieldValue 更多错误场景
func TestSetFieldValueErrors(t *testing.T) {
	t.Run("Unsettable Field", func(t *testing.T) {
		field := reflect.ValueOf(42) // 不可设置
		err := setFieldValue(field, "test")
		if err == nil {
			t.Error("Expected error for unsettable field")
		}
	})

	t.Run("Incompatible Types", func(t *testing.T) {
		field := reflect.ValueOf(new(chan int)).Elem()
		err := setFieldValue(field, "not-a-channel")
		if err == nil {
			t.Error("Expected error for incompatible types")
		}
	})
}

// 测试 ConvertStringToInt 错误场景
func TestConvertStringToIntErrors(t *testing.T) {
	tests := []string{
		"not-a-number",
		"",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			var val int64
			field := reflect.ValueOf(&val).Elem()
			err := convertStringToInt(field, tt)

			if err == nil {
				t.Errorf("Expected error for input '%s'", tt)
			}
		})
	}
}

// 测试 ConvertStringToBool
func TestConvertStringToBoolAdvanced(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", false}, // 大写不应该是 true
		{"false", false},
		{"yes", true},
		{"no", false},
		{"1", true},
		{"0", false},
		{"on", true},
		{"off", false},
		{"anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var val bool
			field := reflect.ValueOf(&val).Elem()
			err := convertStringToBool(field, tt.input)

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
			}
			if field.Bool() != tt.expected {
				t.Errorf("Input '%s': expected %v, got %v", tt.input, tt.expected, field.Bool())
			}
		})
	}
}

// 测试 smartTypeConversion 错误场景
func TestSmartTypeConversionErrors(t *testing.T) {
	t.Run("Non-String Source", func(t *testing.T) {
		field := reflect.ValueOf(new(int)).Elem()
		value := 42
		valueReflect := reflect.ValueOf(value)

		err := smartTypeConversion(field, valueReflect, field.Type(), reflect.TypeOf(value))
		if err == nil {
			t.Error("Expected error for non-string source")
		}
	})

	t.Run("Unsupported Target Type", func(t *testing.T) {
		field := reflect.ValueOf(new(chan int)).Elem()
		value := "not-a-channel"
		valueReflect := reflect.ValueOf(value)

		err := smartTypeConversion(field, valueReflect, field.Type(), reflect.TypeOf(value))
		if err == nil {
			t.Error("Expected error for unsupported target type")
		}
	})
}

// 测试各种数值类型的字符串转换
func TestStringToNumericConversions(t *testing.T) {
	t.Run("Int Types", func(t *testing.T) {
		var val int64
		field := reflect.ValueOf(&val).Elem()
		err := convertStringToInt(field, "42")
		if err != nil {
			t.Errorf("Failed to convert string to int: %v", err)
		}
	})

	t.Run("Uint Types", func(t *testing.T) {
		var val uint64
		field := reflect.ValueOf(&val).Elem()
		err := convertStringToUint(field, "42")
		if err != nil {
			t.Errorf("Failed to convert string to uint: %v", err)
		}
	})

	t.Run("Float Types", func(t *testing.T) {
		var val float64
		field := reflect.ValueOf(&val).Elem()
		err := convertStringToFloat(field, "42.5")
		if err != nil {
			t.Errorf("Failed to convert string to float: %v", err)
		}
	})
}
