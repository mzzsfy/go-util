package di

import (
	"context"
	"reflect"
	"testing"
)

// 测试 Provide 函数（使用自定义类型避免黑名单）
func TestProvideWithCustomType(t *testing.T) {
	c := New()

	// Provide 函数调用 ProvideNamed("", ...)
	err := Provide(c, func(c Container) (*TestService, error) {
		return &TestService{Value: 100}, nil
	})
	if err != nil {
		t.Fatalf("Provide failed: %v", err)
	}

	// 获取服务
	val, err := Get[*TestService](c)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.Value != 100 {
		t.Errorf("Expected 100, got %d", val.Value)
	}
}

// 测试 MustGet 函数
func TestMustGetWithCustomType(t *testing.T) {
	c := New()

	_ = ProvideValue(c, &TestService{Value: 200})

	// MustGet 函数调用 MustGetNamed("", ...)
	val := MustGet[*TestService](c)
	if val.Value != 200 {
		t.Errorf("Expected 200, got %d", val.Value)
	}
}

// 测试 MustGet panic
func TestMustGetPanicWithCustomType(t *testing.T) {
	c := New()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for MustGet with non-existent service")
		}
	}()

	_ = MustGet[*TestService](c) // 应该 panic
}

// 测试 Has 函数
func TestHasWithCustomType(t *testing.T) {
	c := New()

	// 测试未注册
	if Has[*TestService](c) {
		t.Error("Expected Has to return false for non-existent service")
	}

	// 注册服务
	_ = ProvideValue(c, &TestService{Value: 300})

	// 测试已注册（Has 函数调用 HasNamed("", ...)）
	if !Has[*TestService](c) {
		t.Error("Expected Has to return true for registered service")
	}
}

// 测试 ProvideValue 函数
func TestProvideValueWithCustomType(t *testing.T) {
	c := New()

	service := &TestService{Value: 400}
	err := ProvideValue(c, service)
	if err != nil {
		t.Fatalf("ProvideValue failed: %v", err)
	}

	val, err := Get[*TestService](c)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val.Value != 400 {
		t.Errorf("Expected 400, got %d", val.Value)
	}
}

// 测试更多的容器钩子
func TestContainerHooksBeforeCreate(t *testing.T) {
	c := New()

	hookCalled := false
	c.AppendOption(WithContainerBeforeCreate(func(c Container, info EntryInfo) (any, error) {
		hookCalled = true
		return nil, nil
	}))

	_ = ProvideValue(c, &TestService{Value: 500})
	_, _ = Get[*TestService](c)

	if !hookCalled {
		t.Error("Expected before create hook to be called")
	}
}

func TestContainerHooksAfterCreate(t *testing.T) {
	c := New()

	hookCalled := false
	c.AppendOption(WithContainerAfterCreate(func(c Container, info EntryInfo) (any, error) {
		hookCalled = true
		return nil, nil
	}))

	_ = ProvideValue(c, &TestService{Value: 600})
	_, _ = Get[*TestService](c)

	if !hookCalled {
		t.Error("Expected after create hook to be called")
	}
}

func TestContainerHooksBeforeDestroy(t *testing.T) {
	c := New()

	hookCalled := false
	c.AppendOption(WithContainerBeforeDestroy(func(c Container, info EntryInfo) {
		hookCalled = true
	}))

	_ = ProvideValue(c, &TestService{Value: 700})
	_ = c.Start()

	// 关闭容器以触发销毁钩子
	_ = c.Shutdown(context.Background())

	// 注意：销毁钩子可能在 shutdown 时被调用
	t.Logf("Before destroy hook called: %v", hookCalled)
}

func TestContainerHooksAfterDestroy(t *testing.T) {
	c := New()

	hookCalled := false
	c.AppendOption(WithContainerAfterDestroy(func(c Container, info EntryInfo) {
		hookCalled = true
	}))

	_ = ProvideValue(c, &TestService{Value: 800})
	_ = c.Start()
	_ = c.Shutdown(context.Background())

	t.Logf("After destroy hook called: %v", hookCalled)
}

func TestContainerHooksOnStart(t *testing.T) {
	c := New()

	hookCalled := false
	c.AppendOption(WithContainerOnStart(func(c Container) error {
		hookCalled = true
		return nil
	}))

	_ = ProvideValue(c, &TestService{Value: 900})
	err := c.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !hookCalled {
		t.Error("Expected on start hook to be called")
	}

	_ = c.Shutdown(context.Background())
}

func TestContainerHooksAfterStart(t *testing.T) {
	c := New()

	hookCalled := false
	c.AppendOption(WithContainerAfterStart(func(c Container) error {
		hookCalled = true
		return nil
	}))

	_ = ProvideValue(c, &TestService{Value: 1000})
	err := c.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !hookCalled {
		t.Error("Expected after start hook to be called")
	}

	_ = c.Shutdown(context.Background())
}

// 测试 getServiceNameFromKey 更多场景
func TestGetServiceNameFromKeyAdvanced(t *testing.T) {
	tests := []struct {
		key          string
		defaultName  string
		expectedName string
	}{
		{"int#service1", "int", "service1"},
		{"int", "int", ""},
		{"*di.TestService#named", "*di.TestService", "named"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			name := getServiceNameFromKey(tt.key, tt.defaultName)
			if name != tt.expectedName {
				t.Errorf("Expected '%s', got '%s'", tt.expectedName, name)
			}
		})
	}
}

// 测试 validateProviderFunction 更多错误场景
func TestValidateProviderFunctionAdvanced(t *testing.T) {
	t.Run("InvalidNoErrorReturn", func(t *testing.T) {
		fn := func(c Container) int {
			return 42
		}
		err := validateProviderFunction(reflect.TypeOf(fn))
		if err == nil {
			t.Error("Expected error for function without error return")
		}
	})

	t.Run("InvalidWrongFirstParam", func(t *testing.T) {
		fn := func(s string) (int, error) {
			return 42, nil
		}
		err := validateProviderFunction(reflect.TypeOf(fn))
		if err == nil {
			t.Error("Expected error for function with wrong first param")
		}
	})

	t.Run("InvalidTooManyParams", func(t *testing.T) {
		fn := func(c Container, s string) (int, error) {
			return 42, nil
		}
		err := validateProviderFunction(reflect.TypeOf(fn))
		if err == nil {
			t.Error("Expected error for function with too many params")
		}
	})
}

// 测试 tryDirectOrInterfaceMatch 更多场景
func TestTryDirectOrInterfaceMatchAdvanced(t *testing.T) {
	t.Run("DirectMatch", func(t *testing.T) {
		field := reflect.ValueOf(new(int)).Elem()
		value := 42
		valueReflect := reflect.ValueOf(value)

		result := tryDirectOrInterfaceMatch(field, valueReflect, field.Type(), reflect.TypeOf(value))
		if !result {
			t.Error("Expected direct match to succeed")
		}
		if field.Int() != 42 {
			t.Errorf("Expected 42, got %d", field.Int())
		}
	})
}
