package config

import (
	"strconv"
	"testing"
)

// TestValueNil_Int_Panic 测试 valueNil.Int() 应当 panic
func TestValueNil_Int_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("valueNil.Int() 应该 panic")
		}
	}()
	v := valueNil{}
	_ = v.Int()
}

// TestValueNil_Float_Panic 测试 valueNil.Float() 应当 panic
func TestValueNil_Float_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("valueNil.Float() 应该 panic")
		}
	}()
	v := valueNil{}
	_ = v.Float()
}

// TestValueNil_Bool_Panic 测试 valueNil.Bool() 应当 panic
func TestValueNil_Bool_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("valueNil.Bool() 应该 panic")
		}
	}()
	v := valueNil{}
	_ = v.Bool()
}

// TestValueNil_DefaultValues 测试 valueNil 默认值方法
func TestValueNil_DefaultValues(t *testing.T) {
	t.Parallel()
	v := valueNil{}

	if v.Any() != nil {
		t.Error("valueNil.Any() 应该返回 nil")
	}
	if v.AnyD("default") != "default" {
		t.Error("valueNil.AnyD() 应该返回默认值")
	}
	if v.String() != "<nil>" {
		t.Errorf("valueNil.String() = %q, 期望 %q", v.String(), "<nil>")
	}
	if v.StringD("default") != "default" {
		t.Error("valueNil.StringD() 应该返回默认值")
	}
	if v.IntD(42) != 42 {
		t.Error("valueNil.IntD() 应该返回默认值")
	}
	if v.FloatD(3.14) != 3.14 {
		t.Error("valueNil.FloatD() 应该返回默认值")
	}
	if v.BoolD(true) != true {
		t.Error("valueNil.BoolD() 应该返回默认值")
	}
	if _, ok := v.Child("test").(valueNil); !ok {
		t.Error("valueNil.Child() 应该返回 valueNil")
	}
}

// TestValueString_ParseFailed 测试 valueString 解析失败返回 0
func TestValueString_ParseFailed(t *testing.T) {
	t.Parallel()
	v := valueString("not-a-number")

	if v.Int() != 0 {
		t.Errorf("valueString.Int() 解析失败应该返回 0, 实际 %d", v.Int())
	}
	if v.Float() != 0 {
		t.Errorf("valueString.Float() 解析失败应该返回 0, 实际 %f", v.Float())
	}
}

// TestValueString_Parsing 测试 valueString 正常解析
func TestValueString_Parsing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		wantInt  int
		wantFloat float64
		wantBool bool
	}{
		{"整数", "42", 42, 42.0, false},
		{"负整数", "-10", -10, -10.0, false},
		{"浮点数字符串", "3.14", 0, 3.14, false}, // 解析为 Int 失败返回 0
		{"布尔 true", "true", 0, 0, true},
		{"布尔 yes", "yes", 0, 0, true},
		{"布尔 false", "false", 0, 0, false},
		{"空字符串", "", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := valueString(tt.input)
			if got := v.Int(); got != tt.wantInt {
				t.Errorf("Int() = %d, 期望 %d", got, tt.wantInt)
			}
			if got := v.Float(); got != tt.wantFloat {
				t.Errorf("Float() = %f, 期望 %f", got, tt.wantFloat)
			}
			if got := v.Bool(); got != tt.wantBool {
				t.Errorf("Bool() = %v, 期望 %v", got, tt.wantBool)
			}
		})
	}
}

// TestValueString_DefaultValues 测试 valueString 默认值方法
func TestValueString_DefaultValues(t *testing.T) {
	t.Parallel()
	v := valueString("invalid")

	if v.IntD(100) != 100 {
		t.Error("解析失败时 IntD 应该返回默认值")
	}
	if v.FloatD(2.5) != 2.5 {
		t.Error("解析失败时 FloatD 应该返回默认值")
	}

	// 有效值时应该返回解析值
	v2 := valueString("50")
	if v2.IntD(100) != 50 {
		t.Error("解析成功时 IntD 应该返回解析值")
	}
}

// TestValueAny_TypeConversions 测试 valueAny 类型转换
func TestValueAny_TypeConversions(t *testing.T) {
	t.Parallel()
	// 整数类型转换
	intTests := []struct {
		name  string
		value any
		want  int
	}{
		{"int", int(42), 42},
		{"int8", int8(8), 8},
		{"int16", int16(16), 16},
		{"int32", int32(32), 32},
		{"int64", int64(64), 64},
		{"uint", uint(10), 10},
		{"uint8", uint8(8), 8},
		{"uint16", uint16(16), 16},
		{"uint32", uint32(32), 32},
		{"uint64", uint64(64), 64},
	}

	for _, tt := range intTests {
		t.Run("Int_"+tt.name, func(t *testing.T) {
			v := valueAny{value: tt.value}
			if got := v.Int(); got != tt.want {
				t.Errorf("Int() = %d, 期望 %d", got, tt.want)
			}
		})
	}

	// 浮点类型转换
	floatTests := []struct {
		name  string
		value any
		want  float64
	}{
		{"int", int(42), 42.0},
		{"float32", float32(3.14), float64(float32(3.14))},
		{"float64", float64(2.718), 2.718},
	}

	for _, tt := range floatTests {
		t.Run("Float_"+tt.name, func(t *testing.T) {
			v := valueAny{value: tt.value}
			if got := v.Float(); got != tt.want {
				t.Errorf("Float() = %f, 期望 %f", got, tt.want)
			}
		})
	}
}

// TestValueAny_Bool 测试 valueAny 布尔转换
func TestValueAny_Bool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"布尔 true", true, true},
		{"布尔 false", false, false},
		{"字符串 true", "true", true},
		{"字符串 yes", "yes", true},
		{"字符串 false", "false", false},
		{"其他类型", 123, false}, // 默认值
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := valueAny{value: tt.value}
			if got := v.Bool(); got != tt.want {
				t.Errorf("Bool() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

// TestValueAny_DefaultValues 测试 valueAny 默认值
func TestValueAny_DefaultValues(t *testing.T) {
	t.Parallel()
	// 无效类型返回默认值
	v := valueAny{value: "not-a-number"}
	if v.IntD(999) != 999 {
		t.Error("无效类型 IntD 应该返回默认值")
	}

	v2 := valueAny{value: "invalid"}
	if v2.FloatD(1.5) != 1.5 {
		t.Error("无效类型 FloatD 应该返回默认值")
	}

	v3 := valueAny{value: 123}
	if v3.BoolD(true) != true {
		t.Error("无效类型 BoolD 应该返回默认值")
	}
}

// TestValue_Child_NestedAccess 测试 Child 嵌套访问
func TestValue_Child_NestedAccess(t *testing.T) {
	t.Parallel()
	// 创建嵌套结构
	nested := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"value": "deep",
			},
		},
	}

	v := valueAny{value: nested}

	// 第一层访问
	child1 := v.Child("level1")
	if child1.Any() == nil {
		t.Error("Child(level1) 不应该返回 nil")
	}

	// 第二层访问
	child2 := child1.Child("level2")
	if child2.Any() == nil {
		t.Error("Child(level2) 不应该返回 nil")
	}

	// 第三层访问获取值
	child3 := child2.Child("value")
	if child3.String() != "deep" {
		t.Errorf("Child(level2).Child(value) = %q, 期望 %q", child3.String(), "deep")
	}

	// 不存在的路径返回 valueNil
	missing := v.Child("nonexistent")
	if _, ok := missing.(valueNil); !ok {
		t.Error("不存在的路径应该返回 valueNil")
	}
}

// TestValueString_Child 测试 valueString.Child 返回 valueNil
func TestValueString_Child(t *testing.T) {
	t.Parallel()
	v := valueString("test")
	child := v.Child("any")
	if _, ok := child.(valueNil); !ok {
		t.Error("valueString.Child 应该返回 valueNil")
	}
}

// TestValueFromPath 测试路径取值
func TestValueFromPath(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"key":      "value",
		"nested":   map[string]any{"inner": 42},
		"nilKey":   nil,
		"number":   100,
		"floating": 1.5,
	}

	tests := []struct {
		name     string
		path     string
		wantNil  bool
		wantStr  string
		wantInt  int
		wantFloat float64
	}{
		{"简单字符串", "key", false, "value", 0, 0},
		{"嵌套值", "nested.inner", false, "42", 42, 42.0},
		{"不存在的路径", "nonexistent", true, "", 0, 0},
		{"nil 值", "nilKey", true, "", 0, 0},
		{"整数", "number", false, "100", 100, 100.0},
		{"浮点数", "floating", false, "1.5", 1, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ValueFromPath(data, tt.path)
			if tt.wantNil {
				if _, ok := v.(valueNil); !ok {
					t.Error("应该返回 valueNil")
				}
				return
			}
			if v.String() != tt.wantStr {
				t.Errorf("String() = %q, 期望 %q", v.String(), tt.wantStr)
			}
			if v.Int() != tt.wantInt {
				t.Errorf("Int() = %d, 期望 %d", v.Int(), tt.wantInt)
			}
			if v.Float() != tt.wantFloat {
				t.Errorf("Float() = %f, 期望 %f", v.Float(), tt.wantFloat)
			}
		})
	}
}

// TestValueFromPath_EmptyConfig 测试空配置访问
func TestValueFromPath_EmptyConfig(t *testing.T) {
	t.Parallel()
	// 测试空map
	emptyMap := map[string]any{}
	v := ValueFromPath(emptyMap, "anyKey")
	if _, ok := v.(valueNil); !ok {
		t.Error("空map访问任意路径应返回 valueNil")
	}

	// 测试nil值
	v2 := ValueFromPath(nil, "anyKey")
	if _, ok := v2.(valueNil); !ok {
		t.Error("nil访问任意路径应返回 valueNil")
	}
}

// TestValueFromPath_DeepNested 测试超深嵌套路径
func TestValueFromPath_DeepNested(t *testing.T) {
	t.Parallel()
	// 构建深度嵌套结构
	data := map[string]any{}
	current := data
	for i := 0; i < 50; i++ {
		key := "level" + strconv.Itoa(i)
		if i == 49 {
			current[key] = "deepValue"
		} else {
			current[key] = map[string]any{}
			current = current[key].(map[string]any)
		}
	}

	// 构建路径
	path := ""
	for i := 0; i < 50; i++ {
		if i > 0 {
			path += "."
		}
		path += "level" + strconv.Itoa(i)
	}

	v := ValueFromPath(data, path)
	if v.String() != "deepValue" {
		t.Errorf("深度嵌套路径应返回 deepValue, 实际 %q", v.String())
	}

	// 测试超深嵌套后不存在路径
	pathNotExist := path + ".nonexistent"
	v2 := ValueFromPath(data, pathNotExist)
	if _, ok := v2.(valueNil); !ok {
		t.Error("深度嵌套后不存在的路径应返回 valueNil")
	}
}

// TestValueFromPath_InvalidPathFormat 测试无效路径格式
func TestValueFromPath_InvalidPathFormat(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"key": "value",
	}

	tests := []struct {
		name      string
		path      string
		wantPanic bool
	}{
		{"空路径", "", false}, // 空路径返回整个对象
		{"仅点号", ".", true},
		{"连续点号", "key..value", true},
		{"以点号结尾", "key.", true},
		{"以点号开始", ".key", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("路径 %q 应该触发 panic", tt.path)
					}
				}()
			}
			v := ValueFromPath(data, tt.path)
			if !tt.wantPanic {
				// 不应panic的情况检查返回值
				if tt.path == "" {
					// 空路径返回整个对象
					if v.Any() == nil {
						t.Error("空路径应返回整个对象")
					}
				}
			}
		})
	}
}

// TestValueFromPath_ArrayIndex 测试数组索引路径
func TestValueFromPath_ArrayIndex(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"arr": []any{"a", "b", "c"},
	}

	tests := []struct {
		name     string
		path     string
		wantNil  bool
		wantStr  string
		wantPanic bool
	}{
		{"有效索引", "arr[0]", false, "a", false},
		{"第二个索引", "arr[1]", false, "b", false},
		{"超范围索引", "arr[10]", true, "", false}, // 超范围返回nil而非panic
		{"负数索引", "arr[-1]", false, "", true},   // 负数索引会panic
		{"非数字索引", "arr[abc]", false, "", true}, // 非数字会panic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("路径 %q 应该触发 panic", tt.path)
					}
				}()
			}
			v := ValueFromPath(data, tt.path)
			if !tt.wantPanic {
				if tt.wantNil {
					if _, ok := v.(valueNil); !ok {
						t.Errorf("路径 %q 应返回 valueNil", tt.path)
					}
				} else if v.String() != tt.wantStr {
					t.Errorf("路径 %q String() = %q, 期望 %q", tt.path, v.String(), tt.wantStr)
				}
			}
		})
	}
}
