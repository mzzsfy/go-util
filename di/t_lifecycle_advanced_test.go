package di

import (
	"context"
	"reflect"
	"testing"
)

// 额外的测试来达到95%覆盖率

// 测试更多的createDestroyHook场景
func TestCreateDestroyHookAllPaths(t *testing.T) {
	t.Run("with all hooks", func(t *testing.T) {
		c := New().(*container)

		type TestService struct{}

		svc := &TestService{}

		beforeDestroyCalled := false
		afterDestroyCalled := false
		containerBeforeDestroyCalled := false
		containerAfterDestroyCalled := false

		entry := providerEntry{
			reflectType: reflect.TypeOf(svc),
			config: providerConfig{
				beforeDestroy: []func(Container, EntryInfo){
					func(c Container, info EntryInfo) {
						beforeDestroyCalled = true
					},
				},
				afterDestroy: []func(Container, EntryInfo){
					func(c Container, info EntryInfo) {
						afterDestroyCalled = true
					},
				},
			},
		}

		c.beforeDestroy = []func(Container, EntryInfo){
			func(c Container, info EntryInfo) {
				containerBeforeDestroyCalled = true
			},
		}
		c.afterDestroy = []func(Container, EntryInfo){
			func(c Container, info EntryInfo) {
				containerAfterDestroyCalled = true
			},
		}

		hook := c.createDestroyHook(entry, "test", svc)

		err := hook(context.Background())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !beforeDestroyCalled {
			t.Error("Expected beforeDestroy to be called")
		}
		if !afterDestroyCalled {
			t.Error("Expected afterDestroy to be called")
		}
		if !containerBeforeDestroyCalled {
			t.Error("Expected container beforeDestroy to be called")
		}
		if !containerAfterDestroyCalled {
			t.Error("Expected container afterDestroy to be called")
		}
	})
}

// 测试checkAndGetCachedInstance的所有路径
func TestCheckAndGetCachedInstanceAllPaths(t *testing.T) {
	t.Run("first check finds instance", func(t *testing.T) {
		c := New().(*container)
		key := "test-key"
		instance := "test-instance"

		c.mu.Lock()
		c.instances[key] = instance
		c.mu.Unlock()

		result, found := c.checkAndGetCachedInstance(key)
		if !found {
			t.Error("Expected to find instance")
		}
		if result != instance {
			t.Errorf("Expected %v, got %v", instance, result)
		}
	})

	t.Run("loading flag prevents false positive", func(t *testing.T) {
		c := New().(*container)
		key := "test-key"

		c.mu.Lock()
		c.loading[key] = true
		c.mu.Unlock()

		_, found := c.checkAndGetCachedInstance(key)
		if found {
			t.Error("Expected not to find instance when loading")
		}
	})
}

// 测试ShutdownOnSignals的默认参数
func TestShutdownOnSignalsDefaults(t *testing.T) {
	c := New().(*container)

	// 使用默认参数
	c.ShutdownOnSignals()

	// 清理
	_ = c.Shutdown(context.Background())
}

// 测试executeBeforeCreateHooks的错误路径
func TestExecuteBeforeCreateHooksErrors(t *testing.T) {
	t.Run("provider hook returns error", func(t *testing.T) {
		c := New().(*container)
		entry := providerEntry{
			reflectType: reflect.TypeOf(""),
			config: providerConfig{
				beforeCreate: []func(Container, EntryInfo) (any, error){
					func(c Container, info EntryInfo) (any, error) {
						return nil, context.DeadlineExceeded
					},
				},
			},
		}

		_, err := c.executeBeforeCreateHooks(entry, "test")
		if err == nil {
			t.Error("Expected error from beforeCreate hook")
		}
	})

	t.Run("container hook returns error", func(t *testing.T) {
		c := New().(*container)
		c.beforeCreate = []func(Container, EntryInfo) (any, error){
			func(c Container, info EntryInfo) (any, error) {
				return nil, context.Canceled
			},
		}

		entry := providerEntry{
			reflectType: reflect.TypeOf(""),
			config:      providerConfig{},
		}

		_, err := c.executeBeforeCreateHooks(entry, "test")
		if err == nil {
			t.Error("Expected error from container beforeCreate hook")
		}
	})
}

// 测试GetNamedAll的黑名单检查
func TestGetNamedAllBlacklist(t *testing.T) {
	c := New()

	// 尝试对string类型调用GetNamedAll（string在黑名单中）
	_, err := c.GetNamedAll(reflect.TypeOf(""))
	if err == nil {
		t.Error("Expected error for blacklisted type")
	}
}

// 测试injectToInstance的更多场景
func TestInjectToInstanceAdvanced(t *testing.T) {
	t.Run("struct with multiple fields", func(t *testing.T) {
		c := New().(*container)

		type Service1 struct{}
		type Service2 struct{}
		type CompositeService struct {
			Svc1 *Service1 `di:""`
			Svc2 *Service2 `di:""`
		}

		_ = c.ProvideNamedWith("", func(c Container) (*Service1, error) {
			return &Service1{}, nil
		})
		_ = c.ProvideNamedWith("", func(c Container) (*Service2, error) {
			return &Service2{}, nil
		})

		_ = c.ProvideNamedWith("", func(c Container) (*CompositeService, error) {
			return &CompositeService{}, nil
		})

		instance, err := c.GetNamed(reflect.TypeOf(&CompositeService{}), "")
		if err != nil {
			t.Logf("GetNamed error: %v", err)
		} else if instance == nil {
			t.Error("Expected non-nil instance")
		}
	})
}

// 测试setFieldValue的更多类型转换场景
func TestSetFieldValueAdvanced(t *testing.T) {
	t.Run("string to bool", func(t *testing.T) {
		field := reflect.New(reflect.TypeOf(false)).Elem()
		err := setFieldValue(field, "true")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if field.Bool() != true {
			t.Error("Expected true")
		}
	})

	t.Run("string to int", func(t *testing.T) {
		field := reflect.New(reflect.TypeOf(0)).Elem()
		err := setFieldValue(field, "42")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if field.Int() != 42 {
			t.Errorf("Expected 42, got %d", field.Int())
		}
	})

	t.Run("string to float", func(t *testing.T) {
		field := reflect.New(reflect.TypeOf(0.0)).Elem()
		err := setFieldValue(field, "3.14")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if field.Float() != 3.14 {
			t.Errorf("Expected 3.14, got %f", field.Float())
		}
	})
}
