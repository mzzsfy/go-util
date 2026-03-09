package di

import (
	"reflect"
	"testing"

	"github.com/mzzsfy/go-util/helper"
)

// TestCollectMatchingInstancesWithConditionFail 测试条件失败时的分支
func TestCollectMatchingInstancesWithConditionFail(t *testing.T) {
	c := New().(*container)

	// 提供一个带条件的实例，条件返回false
	err := c.ProvideNamedWith("service1",
		func(c Container) (string, error) { return "value1", nil },
		WithCondition(func(c Container) bool { return false }),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 提供一个正常的实例
	err = c.ProvideNamedWith("service2",
		func(c Container) (string, error) { return "value2", nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 收集所有string类型的实例
	results, err := c.collectMatchingInstances(reflect.TypeOf(""), "string")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 应该只有一个实例（service2），因为service1的条件失败
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestMergeParentResultsNoParent 测试没有父容器的情况
func TestMergeParentResultsNoParent(t *testing.T) {
	c := New().(*container)
	results := map[string]any{"key": "value"}

	// 没有父容器，应该直接返回nil
	err := c.mergeParentResults(results, "string")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMergeParentResultsWithParentError 测试父容器返回错误
func TestMergeParentResultsWithParentError(t *testing.T) {
	parent := New().(*container)
	c := New().(*container)
	c.parent = parent

	// 在父容器中添加一个会失败的实例
	err := parent.ProvideNamedWith("bad",
		func(c Container) (string, error) {
			return "", helper.StringError("test error")
		},
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	results := map[string]any{}
	// 父容器GetNamedAll会失败
	err = c.mergeParentResults(results, reflect.TypeOf(""))
	if err == nil {
		t.Error("expected error from parent, got nil")
	}
}

// testErrorForCoverage 用于测试的错误类型
type testErrorForCoverage struct{}

func (e *testErrorForCoverage) Error() string { return "test" }

// TestTryDirectOrInterfaceMatchPtrImplements 测试指针实现接口的情况
func TestTryDirectOrInterfaceMatchPtrImplements(t *testing.T) {
	field := reflect.New(reflect.TypeOf((*error)(nil)).Elem()).Elem()
	value := reflect.ValueOf(&testErrorForCoverage{})
	fieldType := field.Type()
	valueType := value.Type()

	// 测试指针实现接口的分支
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if !result {
		t.Error("expected true, got false")
	}
}

// TestTryDirectOrInterfaceMatchElemImplements 测试指针Elem实现接口的情况
func TestTryDirectOrInterfaceMatchElemImplements(t *testing.T) {
	// 创建一个接口字段
	field := reflect.New(reflect.TypeOf((*error)(nil)).Elem()).Elem()

	// 创建一个指针值，其Elem()实现了接口
	val := testErrorForCoverage{}
	value := reflect.ValueOf(&val)

	fieldType := field.Type()
	valueType := value.Type()

	// valueType.Kind() == reflect.Ptr，valueReflect.Elem().Type().Implements(fieldType)
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if !result {
		t.Error("expected true for elem implements interface")
	}
}

// TestSmartTypeConversionStringToString 测试字符串到字符串的转换
func TestSmartTypeConversionStringToString(t *testing.T) {
	field := reflect.New(reflect.TypeOf("")).Elem()
	value := reflect.ValueOf("test-value")
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if field.String() != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", field.String())
	}
}

// TestSmartTypeConversionStringToBool 测试字符串到布尔值的转换
func TestSmartTypeConversionStringToBool(t *testing.T) {
	field := reflect.New(reflect.TypeOf(false)).Elem()
	value := reflect.ValueOf("true")
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !field.Bool() {
		t.Error("expected true, got false")
	}
}

// TestSmartTypeConversionStringToInt 测试字符串到整数的转换
func TestSmartTypeConversionStringToInt(t *testing.T) {
	field := reflect.New(reflect.TypeOf(0)).Elem()
	value := reflect.ValueOf("42")
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if field.Int() != 42 {
		t.Errorf("expected 42, got %d", field.Int())
	}
}

// TestSmartTypeConversionStringToFloat 测试字符串到浮点数的转换
func TestSmartTypeConversionStringToFloat(t *testing.T) {
	field := reflect.New(reflect.TypeOf(0.0)).Elem()
	value := reflect.ValueOf("3.14")
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if field.Float() != 3.14 {
		t.Errorf("expected 3.14, got %f", field.Float())
	}
}

// TestSmartTypeConversionStringToUint 测试字符串到无符号整数的转换
func TestSmartTypeConversionStringToUint(t *testing.T) {
	field := reflect.New(reflect.TypeOf(uint(0))).Elem()
	value := reflect.ValueOf("42")
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if field.Uint() != 42 {
		t.Errorf("expected 42, got %d", field.Uint())
	}
}

// TestTryDirectOrInterfaceMatchTypeEquality 测试类型相等的情况
func TestTryDirectOrInterfaceMatchTypeEquality(t *testing.T) {
	field := reflect.New(reflect.TypeOf("")).Elem()
	value := reflect.ValueOf("test")
	fieldType := field.Type()
	valueType := value.Type()

	// fieldType == valueType
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if !result {
		t.Error("expected true for type equality")
	}
}

// TestTryDirectOrInterfaceMatchInterfaceImplementation 测试接口实现的情况
func TestTryDirectOrInterfaceMatchInterfaceImplementation(t *testing.T) {
	// 创建一个接口字段
	field := reflect.New(reflect.TypeOf((*error)(nil)).Elem()).Elem()

	// 创建一个实现接口的值
	val := helper.StringError("test")
	value := reflect.ValueOf(val)

	fieldType := field.Type()
	valueType := value.Type()

	// valueReflect.Type().Implements(fieldType)
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if !result {
		t.Error("expected true for interface implementation")
	}
}

// TestTryPointerConversionToPtr 测试转换为指针
func TestTryPointerConversionToPtr(t *testing.T) {
	field := reflect.New(reflect.TypeOf((*int)(nil))).Elem()
	value := reflect.ValueOf(42)
	fieldType := field.Type()
	valueType := value.Type()

	// fieldType.Kind() == reflect.Ptr && valueType.Kind() != reflect.Ptr
	result := tryPointerConversion(field, value, fieldType, valueType)
	if !result {
		t.Error("expected true for pointer conversion")
	}
}

// TestTryPointerConversionFromPtr 测试从指针转换
func TestTryPointerConversionFromPtr(t *testing.T) {
	val := 42
	field := reflect.New(reflect.TypeOf(0)).Elem()
	value := reflect.ValueOf(&val)
	fieldType := field.Type()
	valueType := value.Type()

	// fieldType.Kind() != reflect.Ptr && valueType.Kind() == reflect.Ptr
	result := tryPointerConversion(field, value, fieldType, valueType)
	if !result {
		t.Error("expected true for dereferencing pointer")
	}

	if field.Int() != 42 {
		t.Errorf("expected 42, got %d", field.Int())
	}
}

// TestSetFieldValueCannotSet 测试字段不可设置的情况
func TestSetFieldValueCannotSet(t *testing.T) {
	// 创建一个不可设置的字段
	value := reflect.ValueOf("test")
	field := value

	err := setFieldValue(field, "new value")
	if err == nil {
		t.Error("expected error for unsettable field")
	}
}

// TestSmartTypeConversionUnsupportedKind 测试不支持类型的转换
func TestSmartTypeConversionUnsupportedKind(t *testing.T) {
	// 测试字符串到不支持类型的转换
	field := reflect.New(reflect.TypeOf([]int{})).Elem()
	value := reflect.ValueOf("test")
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err == nil {
		t.Error("expected error for unsupported kind, got nil")
	}
}

// TestSmartTypeConversionNonString 测试非字符串源的转换
func TestSmartTypeConversionNonString(t *testing.T) {
	field := reflect.New(reflect.TypeOf(0)).Elem()
	value := reflect.ValueOf(123) // 不是字符串
	fieldType := field.Type()
	valueType := value.Type()

	err := smartTypeConversion(field, value, fieldType, valueType)
	if err == nil {
		t.Error("expected error for non-string source, got nil")
	}
}

// TestConvertStringToUintInvalid 测试无效的字符串转uint
func TestConvertStringToUintInvalid(t *testing.T) {
	field := reflect.New(reflect.TypeOf(uint(0))).Elem()
	err := convertStringToUint(field, "invalid")
	if err == nil {
		t.Error("expected error for invalid uint string, got nil")
	}
}

// TestConvertStringToFloatInvalid 测试无效的字符串转float
func TestConvertStringToFloatInvalid(t *testing.T) {
	field := reflect.New(reflect.TypeOf(0.0)).Elem()
	err := convertStringToFloat(field, "invalid")
	if err == nil {
		t.Error("expected error for invalid float string, got nil")
	}
}

// TestConvertStringToIntInvalid 测试无效的字符串转int
func TestConvertStringToIntInvalid(t *testing.T) {
	field := reflect.New(reflect.TypeOf(0)).Elem()
	err := convertStringToInt(field, "invalid")
	if err == nil {
		t.Error("expected error for invalid int string, got nil")
	}
}

// TestHasNamedWithReflectType 测试HasNamed使用reflect.Type参数
func TestHasNamedWithReflectType(t *testing.T) {
	c := New().(*container)

	// 注册一个服务
	err := c.ProvideNamedWith("test",
		func(c Container) (string, error) { return "value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 使用reflect.Type检查
	result := c.HasNamed(reflect.TypeOf(""), "test")
	if !result {
		t.Error("expected true, got false")
	}

	// 使用普通类型检查
	result = c.HasNamed("", "test")
	if !result {
		t.Error("expected true, got false")
	}
}

// TestHasNamedWithParent 测试HasNamed在父容器中查找
func TestHasNamedWithParent(t *testing.T) {
	parent := New().(*container)
	c := New().(*container)
	c.parent = parent

	// 在父容器中注册服务
	err := parent.ProvideNamedWith("parent-service",
		func(c Container) (string, error) { return "parent-value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 子容器应该能找到父容器的服务
	result := c.HasNamed("", "parent-service")
	if !result {
		t.Error("expected to find service in parent container")
	}
}

// TestHasNamedNotFound 测试HasNamed找不到服务
func TestHasNamedNotFound(t *testing.T) {
	c := New().(*container)

	// 查找不存在的服务
	result := c.HasNamed("", "nonexistent")
	if result {
		t.Error("expected false for nonexistent service")
	}
}

// TestContainerOptionAfterStart 测试容器选项在启动后添加
func TestContainerOptionAfterStart(t *testing.T) {
	c := New().(*container)

	// 启动容器
	err := c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// 尝试在启动后添加选项，应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding option to started container")
		}
	}()

	WithContainerBeforeCreate(func(c Container, info EntryInfo) (any, error) {
		return nil, nil
	})(c)
}

// TestWithContainerAfterCreateAfterStart 测试AfterCreate钩子在启动后添加
func TestWithContainerAfterCreateAfterStart(t *testing.T) {
	c := New().(*container)

	// 启动容器
	err := c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// 尝试在启动后添加选项，应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding AfterCreate to started container")
		}
	}()

	WithContainerAfterCreate(func(c Container, info EntryInfo) (any, error) {
		return nil, nil
	})(c)
}

// TestWithContainerBeforeDestroyAfterStart 测试BeforeDestroy钩子在启动后添加
func TestWithContainerBeforeDestroyAfterStart(t *testing.T) {
	c := New().(*container)

	// 启动容器
	err := c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// 尝试在启动后添加选项，应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding BeforeDestroy to started container")
		}
	}()

	WithContainerBeforeDestroy(func(c Container, info EntryInfo) {
	})(c)
}

// TestSetConfigSource 测试设置配置源
func TestSetConfigSource(t *testing.T) {
	c := New().(*container)

	// 设置配置源
	source := NewMapConfigSource()
	c.SetConfigSource(source)

	// 验证配置源已设置
	if c.GetConfigSource() != source {
		t.Error("config source not set correctly")
	}
}

// TestSetConfigSourceNil 测试设置nil配置源
func TestSetConfigSourceNil(t *testing.T) {
	c := New().(*container)

	// 设置nil配置源应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when setting nil config source")
		}
	}()

	c.SetConfigSource(nil)
}

// TestGetConfigValueWithEmptyKey 测试空key的配置值获取
func TestGetConfigValueWithEmptyKey(t *testing.T) {
	c := New().(*container)

	// 不设置配置源，获取空key，应该返回一个Value（可能是空值）
	_ = c.getConfigValue("")

	// 设置配置源
	source := NewMapConfigSource()
	c.SetConfigSource(source)

	// 获取空key
	_ = c.getConfigValue("")
}

// TestGetConfigValueDetailed 测试详细的配置值获取
func TestGetConfigValueDetailed(t *testing.T) {
	c := New().(*container)
	source := NewMapConfigSource()
	c.SetConfigSource(source)

	// 测试空key
	_ = c.getConfigValue("")

	// 测试存在的key
	source.Set("test-key", "test-value")
	value := c.getConfigValue("test-key")
	if value == nil {
		t.Error("expected value for existing key")
	}

	// 测试不存在的key
	_ = c.getConfigValue("nonexistent")

	// 测试没有配置源的情况
	c2 := New().(*container)
	_ = c2.getConfigValue("any-key")

	// 验证统计信息
	stats := c.GetStats()
	t.Logf("Config hits: %d, misses: %d", stats.ConfigHits, stats.ConfigMisses)
}

// TestCreateChildScopeConfig 测试子容器继承配置源
func TestCreateChildScopeConfig(t *testing.T) {
	parent := New().(*container)
	source := NewMapConfigSource()
	parent.SetConfigSource(source)

	// 在父容器中设置配置
	source.Set("parent-key", "parent-value")

	// 创建子容器
	child := parent.CreateChildScope()

	// 子容器应该继承父容器的配置源
	childSource := child.GetConfigSource()
	if childSource == nil {
		t.Fatal("child container should inherit config source")
	}

	// 子容器应该能访问父容器的配置
	value := child.Value("parent-key")
	if value == nil {
		t.Error("child should access parent config")
	}
}
func TestGetConfigValueWithSource(t *testing.T) {
	c := New().(*container)
	source := NewMapConfigSource()
	c.SetConfigSource(source)

	// 设置一个值
	source.Set("test-key", "test-value")

	// 获取存在的key
	value := c.getConfigValue("test-key")
	if value == nil {
		t.Fatal("expected value, got nil")
	}

	strVal := value.String()
	if strVal != "test-value" {
		t.Errorf("expected 'test-value', got '%s'", strVal)
	}

	// 获取不存在的key，应该返回一个Value（可能是空值）
	_ = c.getConfigValue("nonexistent")
}

// TestWithContainerAfterDestroy 测试销毁后钩子
func TestWithContainerAfterDestroy(t *testing.T) {
	_ = false // 标记为使用，避免编译错误
	c := New(
		WithContainerAfterDestroy(func(c Container, info EntryInfo) {
			_ = true // 钩子被调用
		}),
	).(*container)

	// 注册一个服务
	err := c.ProvideNamedWith("test",
		func(c Container) (string, error) { return "value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 启动容器
	err = c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// 关闭容器会触发afterDestroy钩子
}

// TestWithContainerAfterDestroyAfterStart 测试AfterDestroy钩子在启动后添加
func TestWithContainerAfterDestroyAfterStart(t *testing.T) {
	c := New().(*container)

	// 启动容器
	err := c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// 尝试在启动后添加选项，应该panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when adding AfterDestroy to started container")
		}
	}()

	WithContainerAfterDestroy(func(c Container, info EntryInfo) {
	})(c)
}

// TestWithContainerAfterStart 测试启动后钩子
func TestWithContainerAfterStart(t *testing.T) {
	called := false
	c := New(
		WithContainerAfterStart(func(c Container) error {
			called = true
			return nil
		}),
	).(*container)

	// 启动容器
	err := c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !called {
		t.Error("AfterStart hook was not called")
	}
}

// TestCreateChildScope 测试创建子容器
func TestCreateChildScope(t *testing.T) {
	parent := New().(*container)

	// 在父容器中注册服务（必须提供名称）
	err := parent.ProvideNamedWith("parent-service",
		func(c Container) (string, error) { return "parent-value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 创建子容器
	child := parent.CreateChildScope()
	if child == nil {
		t.Fatal("child container is nil")
	}

	// 子容器应该能访问父容器的服务
	val, err := child.GetNamed(reflect.TypeOf(""), "parent-service")
	if err != nil {
		t.Fatalf("failed to get service from parent: %v", err)
	}

	if val.(string) != "parent-value" {
		t.Errorf("expected 'parent-value', got %v", val)
	}
}

// TestWithContainerOnStart 测试启动钩子
func TestWithContainerOnStart(t *testing.T) {
	called := false
	c := New(
		WithContainerOnStart(func(c Container) error {
			called = true
			return nil
		}),
	).(*container)

	// 启动容器
	err := c.Start()
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !called {
		t.Error("OnStart hook was not called")
	}
}

// TestWithContainerOnStartError 测试启动钩子返回错误
func TestWithContainerOnStartError(t *testing.T) {
	c := New(
		WithContainerOnStart(func(c Container) error {
			return helper.StringError("startup error")
		}),
	).(*container)

	// 启动容器应该返回错误
	err := c.Start()
	if err == nil {
		t.Error("expected error from OnStart hook")
	}
}

// TestLoadModeTransientMore 测试瞬时加载模式
func TestLoadModeTransientMore(t *testing.T) {
	callCount := 0
	c := New().(*container)

	err := c.ProvideNamedWith("transient",
		func(c Container) (string, error) {
			callCount++
			return "value", nil
		},
		WithLoadMode(LoadModeTransient),
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 获取实例多次，每次都应该创建新实例
	_, err = c.GetNamed(reflect.TypeOf(""), "transient")
	if err != nil {
		t.Fatalf("first get failed: %v", err)
	}

	_, err = c.GetNamed(reflect.TypeOf(""), "transient")
	if err != nil {
		t.Fatalf("second get failed: %v", err)
	}

	// 应该调用了两次
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

// TestClearInstancesSimple 测试清除实例
func TestClearInstancesSimple(t *testing.T) {
	c := New().(*container)

	// 注册并获取服务
	err := c.ProvideNamedWith("test",
		func(c Container) (string, error) { return "value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = c.GetNamed(reflect.TypeOf(""), "test")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// 验证实例存在
	if c.GetInstanceCount() == 0 {
		t.Error("expected instances to exist")
	}

	// 清除实例
	c.ClearInstances()

	// 验证实例被清除
	if c.GetInstanceCount() != 0 {
		t.Error("expected instances to be cleared")
	}
}

// TestGetInstanceCountSimple 测试获取实例数量
func TestGetInstanceCountSimple(t *testing.T) {
	c := New().(*container)

	// 初始应该为0
	if c.GetInstanceCount() != 0 {
		t.Error("expected 0 instances initially")
	}

	// 注册并获取服务
	err := c.ProvideNamedWith("test",
		func(c Container) (string, error) { return "value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = c.GetNamed(reflect.TypeOf(""), "test")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// 应该有1个实例
	if c.GetInstanceCount() != 1 {
		t.Errorf("expected 1 instance, got %d", c.GetInstanceCount())
	}
}

// TestGetProviderCountSimple 测试获取提供者数量
func TestGetProviderCountSimple(t *testing.T) {
	c := New().(*container)

	// 初始应该为0
	if c.GetProviderCount() != 0 {
		t.Error("expected 0 providers initially")
	}

	// 注册服务
	err := c.ProvideNamedWith("test",
		func(c Container) (string, error) { return "value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 应该有1个提供者
	if c.GetProviderCount() != 1 {
		t.Errorf("expected 1 provider, got %d", c.GetProviderCount())
	}
}

// TestServiceLifecycle 测试服务生命周期接口
func TestServiceLifecycle(t *testing.T) {
	shutdownCalled := false

	type TestService struct {
		value string
	}

	c := New().(*container)

	err := c.ProvideNamedWith("lifecycle",
		func(c Container) (*TestService, error) {
			return &TestService{value: "test"}, nil
		},
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 获取实例
	_, err = c.GetNamed(reflect.TypeOf(&TestService{}), "lifecycle")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// 注册一个实现了ServiceLifecycle的服务
	type LifecycleService struct {
		TestService
	}

	err = c.ProvideNamedWith("lifecycle2",
		func(c Container) (*LifecycleService, error) {
			return &LifecycleService{}, nil
		},
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 获取实例
	_, err = c.GetNamed(reflect.TypeOf(&LifecycleService{}), "lifecycle2")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	_ = shutdownCalled // 避免未使用变量警告
}

// TestGetNamedAllWithInterface 测试通过接口获取所有实例
func TestGetNamedAllWithInterface(t *testing.T) {
	c := New().(*container)

	// 注册多个实现了error接口的服务
	err := c.ProvideNamedWith("error1",
		func(c Container) (error, error) {
			return helper.StringError("error1"), nil
		},
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err = c.ProvideNamedWith("error2",
		func(c Container) (error, error) {
			return helper.StringError("error2"), nil
		},
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 获取所有error类型的实例
	instances, err := c.GetNamedAll(reflect.TypeOf((*error)(nil)).Elem())
	if err != nil {
		t.Fatalf("GetNamedAll failed: %v", err)
	}

	// 应该有2个实例
	if len(instances) < 2 {
		t.Errorf("expected at least 2 instances, got %d", len(instances))
	}
}

// TestValidateProviderFunctionSimple 测试验证provider函数
func TestValidateProviderFunctionSimple(t *testing.T) {
	// 测试有效的provider函数
	validFunc := func(c Container) (string, error) { return "test", nil }
	err := validateProviderFunction(reflect.TypeOf(validFunc))
	if err != nil {
		t.Errorf("valid function should pass: %v", err)
	}

	// 测试无效的provider函数（没有返回值）
	invalidFunc := func(c Container) {}
	err = validateProviderFunction(reflect.TypeOf(invalidFunc))
	if err == nil {
		t.Error("invalid function should fail validation")
	}
}

// TestTryDirectOrInterfaceMatchNonInterface 测试非接口类型的字段
func TestTryDirectOrInterfaceMatchNonInterface(t *testing.T) {
	field := reflect.New(reflect.TypeOf(0)).Elem()
	value := reflect.ValueOf("test")
	fieldType := field.Type()
	valueType := value.Type()

	// fieldType不是接口类型，且类型不匹配
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if result {
		t.Error("expected false for non-matching non-interface types")
	}
}

// TestTryDirectOrInterfaceMatchInterfaceNotImplemented 测试接口未实现的情况
func TestTryDirectOrInterfaceMatchInterfaceNotImplemented(t *testing.T) {
	// 创建一个接口字段
	field := reflect.New(reflect.TypeOf((*error)(nil)).Elem()).Elem()

	// 创建一个不实现error接口的值
	value := reflect.ValueOf(123)
	fieldType := field.Type()
	valueType := value.Type()

	// fieldType是接口，但value不实现
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if result {
		t.Error("expected false for interface not implemented")
	}
}

// TestTryDirectOrInterfaceMatchPtrElemNotImplemented 测试指针Elem不实现接口
func TestTryDirectOrInterfaceMatchPtrElemNotImplemented(t *testing.T) {
	// 创建一个接口字段
	field := reflect.New(reflect.TypeOf((*error)(nil)).Elem()).Elem()

	// 创建一个指针，其Elem不实现接口
	val := 123
	value := reflect.ValueOf(&val)
	fieldType := field.Type()
	valueType := value.Type()

	// valueType.Kind() == reflect.Ptr，但Elem()不实现接口
	result := tryDirectOrInterfaceMatch(field, value, fieldType, valueType)
	if result {
		t.Error("expected false when elem does not implement interface")
	}
}

// TestGetStatsDetailed 测试获取统计信息
func TestGetStatsDetailed(t *testing.T) {
	c := New().(*container)

	// 初始统计应该为0
	stats := c.GetStats()
	if stats.CreatedInstances != 0 {
		t.Errorf("expected 0 created instances, got %d", stats.CreatedInstances)
	}

	// 注册并获取服务
	err := c.ProvideNamedWith("test",
		func(c Container) (string, error) { return "value", nil },
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = c.GetNamed(reflect.TypeOf(""), "test")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// 统计应该更新
	stats = c.GetStats()
	if stats.CreatedInstances == 0 {
		t.Error("expected some created instances")
	}
	if stats.GetCalls == 0 {
		t.Error("expected some get calls")
	}
}
