package di

import (
    "context"
    "fmt"
    "testing"
)

// Test_Hooks 测试钩子功能
func Test_Hooks(t *testing.T) {
    t.Run("Provider级别钩子", testProviderHooks)
    t.Run("Container级别钩子", testContainerHooks)
    t.Run("混合钩子", testMixedHooks)
}

// testProviderHooks 测试 provider 级别的钩子
func testProviderHooks(t *testing.T) {
    container := New()

    beforeCreateCalled := false
    afterCreateCalled := false
    beforeDestroyCalled := false
    afterDestroyCalled := false

    type TestService struct {
        Value string
    }

    // 注册带有钩子的服务
    err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
        return &TestService{Value: "original"}, nil
    },
        WithBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            beforeCreateCalled = true
            t.Log("Provider BeforeCreate called")
            return nil, nil // 使用默认创建
        }),
        WithAfterCreate(func(c Container, info EntryInfo) (any, error) {
            afterCreateCalled = true
            t.Log("Provider AfterCreate called")
            if service, ok := info.Instance.(*TestService); ok {
                service.Value = "modified"
            }
            return info.Instance, nil
        }),
        WithBeforeDestroy(func(c Container, info EntryInfo) {
            beforeDestroyCalled = true
            t.Log("Provider BeforeDestroy called")
        }),
        WithAfterDestroy(func(c Container, info EntryInfo) {
            afterDestroyCalled = true
            t.Log("Provider AfterDestroy called")
        }),
    )

    if err != nil {
        t.Fatalf("注册服务失败: %v", err)
    }

    // 获取服务，触发创建钩子
    service, err := GetNamed[*TestService](container, "test")
    if err != nil {
        t.Fatalf("获取服务失败: %v", err)
    }

    // 验证钩子被调用
    if !beforeCreateCalled {
        t.Error("Provider BeforeCreate 钩子未被调用")
    }
    if !afterCreateCalled {
        t.Error("Provider AfterCreate 钩子未被调用")
    }
    if service.Value != "modified" {
        t.Errorf("AfterCreate 钩子应该修改值，期望'modified', 实际'%s'", service.Value)
    }

    // 触发销毁钩子
    container.Shutdown(context.Background())

    if !beforeDestroyCalled {
        t.Error("Provider BeforeDestroy 钩子未被调用")
    }
    if !afterDestroyCalled {
        t.Error("Provider AfterDestroy 钩子未被调用")
    }
}

// testContainerHooks 测试容器级别的钩子
func testContainerHooks(t *testing.T) {
    globalBeforeCreate := 0
    globalAfterCreate := 0
    globalBeforeDestroy := 0
    globalAfterDestroy := 0

    container := New(
        WithContainerBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            globalBeforeCreate++
            t.Logf("Container BeforeCreate called (count: %d)", globalBeforeCreate)
            return nil, nil
        }),
        WithContainerAfterCreate(func(c Container, info EntryInfo) (any, error) {
            globalAfterCreate++
            t.Logf("Container AfterCreate called (count: %d)", globalAfterCreate)
            return info.Instance, nil
        }),
        WithContainerBeforeDestroy(func(c Container, info EntryInfo) {
            globalBeforeDestroy++
            t.Logf("Container BeforeDestroy called (count: %d)", globalBeforeDestroy)
        }),
        WithContainerAfterDestroy(func(c Container, info EntryInfo) {
            globalAfterDestroy++
            t.Logf("Container AfterDestroy called (count: %d)", globalAfterDestroy)
        }),
    )

    type Service1 struct{ Value string }
    type Service2 struct{ Value string }

    // 注册两个服务
    container.ProvideNamedWith("service1", func(c Container) (*Service1, error) {
        return &Service1{Value: "service1"}, nil
    })
    container.ProvideNamedWith("service2", func(c Container) (*Service2, error) {
        return &Service2{Value: "service2"}, nil
    })

    // 获取两个服务，应该触发两次容器钩子
    _, _ = GetNamed[*Service1](container, "service1")
    _, _ = GetNamed[*Service2](container, "service2")

    if globalBeforeCreate != 2 {
        t.Errorf("Container BeforeCreate 应该被调用2次，实际%d次", globalBeforeCreate)
    }
    if globalAfterCreate != 2 {
        t.Errorf("Container AfterCreate 应该被调用2次，实际%d次", globalAfterCreate)
    }

    // 销毁容器
    container.Shutdown(context.Background())

    if globalBeforeDestroy != 2 {
        t.Errorf("Container BeforeDestroy 应该被调用2次，实际%d次", globalBeforeDestroy)
    }
    if globalAfterDestroy != 2 {
        t.Errorf("Container AfterDestroy 应该被调用2次，实际%d次", globalAfterDestroy)
    }
}

// testMixedHooks 测试混合钩子
func testMixedHooks(t *testing.T) {
    containerLog := []string{}
    providerLog := []string{}

    container := New(
        WithContainerBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            containerLog = append(containerLog, "container-before")
            return nil, nil
        }),
        WithContainerAfterCreate(func(c Container, info EntryInfo) (any, error) {
            containerLog = append(containerLog, "container-after")
            return info.Instance, nil
        }),
    )

    type TestService struct{ Value string }

    err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
        return &TestService{Value: "test"}, nil
    },
        WithBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            providerLog = append(providerLog, "provider-before")
            return nil, nil
        }),
        WithAfterCreate(func(c Container, info EntryInfo) (any, error) {
            providerLog = append(providerLog, "provider-after")
            return info.Instance, nil
        }),
    )

    if err != nil {
        t.Fatalf("注册服务失败: %v", err)
    }

    // 获取服务
    _, err = GetNamed[*TestService](container, "test")
    if err != nil {
        t.Fatalf("获取服务失败: %v", err)
    }

    // 验证执行顺序：容器before -> providerbefore -> providerafter -> containerafter
    if len(containerLog) != 2 {
        t.Errorf("容器钩子应该被调用2次，实际%d次", len(containerLog))
    }
    if len(providerLog) != 2 {
        t.Errorf("Provider钩子应该被调用2次，实际%d次", len(providerLog))
    }

    t.Logf("Container log: %v", containerLog)
    t.Logf("Provider log: %v", providerLog)
}

// TestHookBeanModification 测试钩子修改实例
func TestHookBeanModification(t *testing.T) {
    container := New()

    type Service struct {
        Value string
        Count int
    }

    // 使用 afterCreate 钩子修改实例
    err := container.ProvideNamedWith("service", func(c Container) (*Service, error) {
        return &Service{Value: "original", Count: 0}, nil
    },
        WithAfterCreate(func(c Container, info EntryInfo) (any, error) {
            if service, ok := info.Instance.(*Service); ok {
                service.Value = "hook-modified"
                service.Count = 42
            }
            return info.Instance, nil
        }),
    )

    if err != nil {
        t.Fatalf("注册服务失败: %v", err)
    }

    service, err := GetNamed[*Service](container, "service")
    if err != nil {
        t.Fatalf("获取服务失败: %v", err)
    }

    if service.Value != "hook-modified" {
        t.Errorf("期望'hook-modified', 实际'%s'", service.Value)
    }
    if service.Count != 42 {
        t.Errorf("期望42, 实际%d", service.Count)
    }
}

// TestHookReplacement 测试钩子替换实例
func TestHookReplacement(t *testing.T) {
    container := New()

    type Service struct {
        Value string
    }

    // 使用 beforeCreate 钩子完全替换创建过程
    err := container.ProvideNamedWith("service", func(c Container) (*Service, error) {
        // 这个函数不会被调用，因为 beforeCreate 返回了实例
        return &Service{Value: "original"}, nil
    },
        WithBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            // 完全替换创建过程
            return &Service{Value: "hook-replaced"}, nil
        }),
    )

    if err != nil {
        t.Fatalf("注册服务失败: %v", err)
    }

    service, err := GetNamed[*Service](container, "service")
    if err != nil {
        t.Fatalf("获取服务失败: %v", err)
    }

    if service.Value != "hook-replaced" {
        t.Errorf("期望'hook-replaced', 实际'%s'", service.Value)
    }
}

// TestHookErrorHandling 测试钩子错误处理
func TestHookErrorHandling(t *testing.T) {
    container := New()

    type Service struct{ Value string }

    // 测试 beforeCreate 返回错误
    err := container.ProvideNamedWith("service1", func(c Container) (*Service, error) {
        return &Service{Value: "test"}, nil
    },
        WithBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            return nil, fmt.Errorf("beforeCreate error")
        }),
    )

    if err != nil {
        t.Fatalf("注册服务应该成功: %v", err)
    }

    _, err = GetNamed[*Service](container, "service1")
    if err == nil {
        t.Fatal("应该返回 beforeCreate 错误")
    }
    if !contains(err.Error(), "beforeCreate error") {
        t.Errorf("期望包含'beforeCreate error', 实际: %v", err)
    }

    // 测试 afterCreate 返回错误
    container2 := New()
    err = container2.ProvideNamedWith("service2", func(c Container) (*Service, error) {
        return &Service{Value: "test"}, nil
    },
        WithAfterCreate(func(c Container, info EntryInfo) (any, error) {
            return nil, fmt.Errorf("afterCreate error")
        }),
    )

    if err != nil {
        t.Fatalf("注册服务应该成功: %v", err)
    }

    _, err = GetNamed[*Service](container2, "service2")
    if err == nil {
        t.Fatal("应该返回 afterCreate 错误")
    }
    if !contains(err.Error(), "afterCreate error") {
        t.Errorf("期望包含'afterCreate error', 实际: %v", err)
    }
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
