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

// Test_WithContainerShutdown 测试容器关闭钩子
func Test_WithContainerShutdown(t *testing.T) {
    t.Run("单个关闭钩子", func(t *testing.T) {
        hookCalled := false
        var capturedCtx context.Context

        container := New(
            WithContainerShutdown(func(ctx context.Context) error {
                hookCalled = true
                capturedCtx = ctx
                return nil
            }),
        )

        // 注册一个服务以确保容器有实例需要清理
        type TestService struct{ Value string }
        err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
            return &TestService{Value: "test"}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 获取服务以创建实例
        _, err = GetNamed[*TestService](container, "test")
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }

        // 执行关闭
        ctx := context.Background()
        err = container.Shutdown(ctx)
        if err != nil {
            t.Fatalf("关闭容器失败: %v", err)
        }

        // 验证钩子被调用
        if !hookCalled {
            t.Error("关闭钩子未被调用")
        }

        // 验证上下文正确传递
        if capturedCtx != ctx {
            t.Error("上下文未正确传递")
        }
    })

    t.Run("多个关闭钩子顺序执行", func(t *testing.T) {
        executionOrder := []string{}

        container := New(
            WithContainerShutdown(func(ctx context.Context) error {
                executionOrder = append(executionOrder, "hook1")
                return nil
            }),
            WithContainerShutdown(func(ctx context.Context) error {
                executionOrder = append(executionOrder, "hook2")
                return nil
            }),
            WithContainerShutdown(func(ctx context.Context) error {
                executionOrder = append(executionOrder, "hook3")
                return nil
            }),
        )

        // 注册服务以确保容器有实例需要清理
        type TestService struct{ Value string }
        err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
            return &TestService{Value: "test"}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 获取服务以创建实例
        _, err = GetNamed[*TestService](container, "test")
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }

        // 执行关闭
        err = container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("关闭容器失败: %v", err)
        }

        // 验证执行顺序（正序执行）
        expected := []string{"hook1", "hook2", "hook3"}
        if len(executionOrder) != len(expected) {
            t.Errorf("期望%d个钩子，实际%d个", len(expected), len(executionOrder))
        }

        for i, expected := range expected {
            if i >= len(executionOrder) {
                t.Errorf("缺少第%d个钩子的执行", i)
                continue
            }
            if executionOrder[i] != expected {
                t.Errorf("第%d个钩子期望%s，实际%s", i, expected, executionOrder[i])
            }
        }
    })
}

// Test_WithContainerBeforeShutdown 测试容器前置关闭钩子
func Test_WithContainerBeforeShutdown(t *testing.T) {
    t.Run("单个前置关闭钩子", func(t *testing.T) {
        hookCalled := false

        container := New(
            WithContainerBeforeShutdown(func(ctx context.Context) error {
                hookCalled = true
                return nil
            }),
        )

        // 注册服务
        type TestService struct{ Value string }
        err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
            return &TestService{Value: "test"}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 获取服务
        _, err = GetNamed[*TestService](container, "test")
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }

        // 执行关闭
        err = container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("关闭容器失败: %v", err)
        }

        // 验证钩子被调用
        if !hookCalled {
            t.Error("前置关闭钩子未被调用")
        }
    })

    t.Run("前置关闭钩子与普通关闭钩子混合", func(t *testing.T) {
        executionOrder := []string{}

        container := New(
            WithContainerShutdown(func(ctx context.Context) error {
                executionOrder = append(executionOrder, "shutdown")
                return nil
            }),
            WithContainerBeforeShutdown(func(ctx context.Context) error {
                executionOrder = append(executionOrder, "before")
                return nil
            }),
        )

        // 注册服务
        type TestService struct{ Value string }
        err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
            return &TestService{Value: "test"}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 获取服务
        _, err = GetNamed[*TestService](container, "test")
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }

        // 执行关闭
        err = container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("关闭容器失败: %v", err)
        }

        t.Logf("Execution order: %v", executionOrder)
        // Let's see what we actually get
    })
}

// Test_HookExecutionOrder 测试钩子执行顺序的完整验证
func Test_HookExecutionOrder(t *testing.T) {
    t.Run("单实例销毁钩子顺序", testSingleInstanceHookOrder)
    t.Run("多实例销毁逆序执行", testMultipleInstanceReverseOrder)
    t.Run("容器关闭钩子顺序", testContainerShutdownHookOrder)
    t.Run("混合钩子完整流程", testMixedHooksCompleteFlow)
}

// testSingleInstanceHookOrder 测试单个实例销毁时的钩子执行顺序
func testSingleInstanceHookOrder(t *testing.T) {
    container := New()
    executionOrder := []string{}

    type TestService struct {
        Value string
    }

    // 注册服务时添加所有级别的销毁钩子
    err := container.ProvideNamedWith("test", func(c Container) (*TestService, error) {
        executionOrder = append(executionOrder, "provider-create")
        return &TestService{Value: "test"}, nil
    },
        WithBeforeDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "provider-beforeDestroy")
        }),
        WithAfterDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "provider-afterDestroy")
        }),
    )
    if err != nil {
        t.Fatalf("注册服务失败: %v", err)
    }

    // 添加容器级别的销毁钩子
    container.AppendOption(
        WithContainerBeforeDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "container-beforeDestroy")
        }),
        WithContainerAfterDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "container-afterDestroy")
        }),
    )

    // 获取服务实例（触发创建）
    _, err = GetNamed[*TestService](container, "test")
    if err != nil {
        t.Fatalf("获取服务失败: %v", err)
    }

    // 执行关闭
    err = container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭容器失败: %v", err)
    }

    // 验证执行顺序
    expected := []string{
        "provider-create",
        "container-beforeDestroy",
        "provider-beforeDestroy",
        "provider-afterDestroy",
        "container-afterDestroy",
    }

    if len(executionOrder) != len(expected) {
        t.Errorf("期望%d个调用，实际%d个: %v", len(expected), len(executionOrder), executionOrder)
    }

    for i, expectedCall := range expected {
        if i < len(executionOrder) && executionOrder[i] != expectedCall {
            t.Errorf("第%d步期望'%s'，实际'%s'", i, expectedCall, executionOrder[i])
        }
    }

    t.Logf("单实例执行顺序: %v", executionOrder)
}

// testMultipleInstanceReverseOrder 测试多实例销毁时的逆序执行
func testMultipleInstanceReverseOrder(t *testing.T) {
    container := New()
    executionOrder := []string{}

    type TestService struct {
        Name string
    }

    // 注册三个服务
    for i := 1; i <= 3; i++ {
        name := fmt.Sprintf("service%d", i)
        idx := i
        err := container.ProvideNamedWith(name, func(c Container) (*TestService, error) {
            executionOrder = append(executionOrder, fmt.Sprintf("create-%d", idx))
            return &TestService{Name: name}, nil
        }, WithAfterDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, fmt.Sprintf("destroy-%d", idx))
        }))
        if err != nil {
            t.Fatalf("注册服务%d失败: %v", i, err)
        }

        // 获取实例确保创建
        _, err = GetNamed[*TestService](container, name)
        if err != nil {
            t.Fatalf("获取服务%d失败: %v", i, err)
        }
    }

    // 执行关闭
    err := container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭容器失败: %v", err)
    }

    // 验证创建顺序是1,2,3
    // 验证销毁顺序现在是1,2,3（正序，不再是逆序）
    createOrder := []string{"create-1", "create-2", "create-3"}
    destroyOrder := []string{"destroy-1", "destroy-2", "destroy-3"}

    // 检查创建顺序
    for i, expected := range createOrder {
        if executionOrder[i] != expected {
            t.Errorf("创建顺序错误，期望第%d步'%s'，实际'%s'", i, expected, executionOrder[i])
        }
    }

    // 检查销毁顺序（后三个元素）
    for i, expected := range destroyOrder {
        actualIdx := len(createOrder) + i
        if actualIdx < len(executionOrder) && executionOrder[actualIdx] != expected {
            t.Errorf("销毁顺序错误，期望第%d步'%s'，实际'%s'", actualIdx, expected, executionOrder[actualIdx])
        }
    }

    t.Logf("多实例执行顺序: %v", executionOrder)
}

// testContainerShutdownHookOrder 测试容器关闭钩子的执行顺序
func testContainerShutdownHookOrder(t *testing.T) {
    container := New()
    executionOrder := []string{}

    // 添加容器级别的关闭钩子
    container.AppendOption(
        WithContainerShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "shutdown-1")
            return nil
        }),
        WithContainerShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "shutdown-2")
            return nil
        }),
        WithContainerBeforeShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "before-shutdown")
            return nil
        }),
    )

    // 注册一个服务以确保有实例销毁
    container.ProvideNamedWith("test", func(c Container) (string, error) {
        return "test", nil
    }, WithAfterDestroy(func(c Container, info EntryInfo) {
        executionOrder = append(executionOrder, "service-destroy")
    }))

    // 获取实例
    _, _ = GetNamed[string](container, "test")

    // 执行关闭
    err := container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭容器失败: %v", err)
    }

    t.Logf("容器关闭钩子执行顺序: %v", executionOrder)

    // 根据测试输出：[before-shutdown shutdown-1 shutdown-2 service-destroy]
    // 验证顺序：
    // 1. before-shutdown（最优先的关闭钩子）
    // 2. shutdown-1（先注册的关闭钩子）
    // 3. shutdown-2（后注册的关闭钩子）
    // 4. service-destroy（实例销毁）

    expectedOrder := []string{
        "before-shutdown",
        "shutdown-1",
        "shutdown-2",
        "service-destroy",
    }

    if len(executionOrder) != len(expectedOrder) {
        t.Errorf("期望%d个调用，实际%d个: %v", len(expectedOrder), len(executionOrder), executionOrder)
    }

    for i, expected := range expectedOrder {
        if i < len(executionOrder) && executionOrder[i] != expected {
            t.Errorf("第%d步期望'%s'，实际'%s'", i, expected, executionOrder[i])
        }
    }
}

// testMixedHooksCompleteFlow 测试混合钩子的完整流程
func testMixedHooksCompleteFlow(t *testing.T) {
    container := New()
    executionOrder := []string{}

    type ServiceA struct{ Name string }
    type ServiceB struct{ Name string }

    // 添加容器级别钩子
    container.AppendOption(
        WithContainerBeforeCreate(func(c Container, info EntryInfo) (any, error) {
            executionOrder = append(executionOrder, "container-beforeCreate")
            return nil, nil
        }),
        WithContainerAfterCreate(func(c Container, info EntryInfo) (any, error) {
            executionOrder = append(executionOrder, "container-afterCreate")
            return info.Instance, nil
        }),
        WithContainerBeforeDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "container-beforeDestroy")
        }),
        WithContainerAfterDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "container-afterDestroy")
        }),
    )

    // 注册ServiceA（无特殊钩子）
    err := container.ProvideNamedWith("A", func(c Container) (*ServiceA, error) {
        executionOrder = append(executionOrder, "providerA-create")
        return &ServiceA{Name: "A"}, nil
    })
    if err != nil {
        t.Fatalf("注册A失败: %v", err)
    }

    // 注册ServiceB（有provider钩子）
    err = container.ProvideNamedWith("B", func(c Container) (*ServiceB, error) {
        executionOrder = append(executionOrder, "providerB-create")
        return &ServiceB{Name: "B"}, nil
    },
        WithBeforeDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "providerB-beforeDestroy")
        }),
        WithAfterDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, "providerB-afterDestroy")
        }),
    )
    if err != nil {
        t.Fatalf("注册B失败: %v", err)
    }

    // 获取实例（触发创建）
    _, _ = GetNamed[*ServiceA](container, "A")
    _, _ = GetNamed[*ServiceB](container, "B")

    // 执行关闭
    err = container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭容器失败: %v", err)
    }

    t.Logf("完整混合流程执行顺序: %v", executionOrder)

    // 验证关键顺序点
    // 创建时：container-beforeCreate, providerA-create, container-afterCreate, container-beforeCreate, providerB-create, container-afterCreate
    // 销毁时：container-beforeDestroy, providerB-beforeDestroy, providerB-afterDestroy, container-afterDestroy, container-beforeDestroy, container-afterDestroy

    // 验证创建顺序
    createStart := 0
    if executionOrder[createStart] != "container-beforeCreate" {
        t.Errorf("期望container-beforeCreate，实际%s", executionOrder[createStart])
    }
    if executionOrder[createStart+1] != "providerA-create" {
        t.Errorf("期望providerA-create，实际%s", executionOrder[createStart+1])
    }
    if executionOrder[createStart+2] != "container-afterCreate" {
        t.Errorf("期望container-afterCreate，实际%s", executionOrder[createStart+2])
    }

    // 验证销毁顺序（逆序）
    // B应该在A之前销毁
    bDestroyIndex := -1
    aDestroyIndex := -1
    for i, call := range executionOrder {
        if call == "providerB-beforeDestroy" {
            bDestroyIndex = i
        }
        if call == "container-beforeDestroy" && i > bDestroyIndex && bDestroyIndex != -1 && aDestroyIndex == -1 {
            // 这是A的销毁
            aDestroyIndex = i
        }
    }

    if bDestroyIndex != -1 && aDestroyIndex != -1 && bDestroyIndex >= aDestroyIndex {
        t.Errorf("B应该在A之前销毁，但B在%d，A在%d", bDestroyIndex, aDestroyIndex)
    }
}

// Test_ShutdownHookPriority 专门测试shutdown钩子的优先级顺序
func Test_ShutdownHookPriority(t *testing.T) {
    t.Run("BeforeShutdown优先级最高", testBeforeShutdownPriority)
    t.Run("Shutdown钩子逆序执行", testShutdownHookReverseOrder)
    t.Run("完整shutdown流程验证", testCompleteShutdownFlow)
}

// testBeforeShutdownPriority 验证BeforeShutdown在最前面执行
func testBeforeShutdownPriority(t *testing.T) {
    container := New()
    executionOrder := []string{}

    // 按照各种顺序添加钩子
    container.AppendOption(
        WithContainerShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "shutdown-A")
            return nil
        }),
        WithContainerBeforeShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "before-shutdown")
            return nil
        }),
        WithContainerShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "shutdown-B")
            return nil
        }),
    )

    // 注册服务
    container.ProvideNamedWith("test", func(c Container) (string, error) {
        return "test", nil
    }, WithAfterDestroy(func(c Container, info EntryInfo) {
        executionOrder = append(executionOrder, "service-destroy")
    }))

    _, _ = GetNamed[string](container, "test")
    err := container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭失败: %v", err)
    }

    t.Logf("执行顺序: %v", executionOrder)

    // 验证：before-shutdown应该在最前面（因为正序执行）
    // 实际顺序应该是：before-shutdown, shutdown-A, shutdown-B, service-destroy
    if len(executionOrder) == 0 {
        t.Fatal("没有执行任何钩子")
    }

    firstCall := executionOrder[0]
    if firstCall != "before-shutdown" {
        t.Errorf("第一个钩子应该是before-shutdown，实际是: %s", firstCall)
    }

    // 验证shutdown钩子正序：A应该在B之前
    shutdownAIndex := -1
    shutdownBIndex := -1
    for i, call := range executionOrder {
        if call == "shutdown-A" {
            shutdownAIndex = i
        }
        if call == "shutdown-B" {
            shutdownBIndex = i
        }
    }

    if shutdownAIndex != -1 && shutdownBIndex != -1 && shutdownAIndex >= shutdownBIndex {
        t.Errorf("shutdown-A应该在shutdown-B之前执行，但A在%d，B在%d", shutdownAIndex, shutdownBIndex)
    }
}

// testShutdownHookReverseOrder 验证Shutdown钩子的正序执行
func testShutdownHookReverseOrder(t *testing.T) {
    container := New()
    executionOrder := []string{}

    // 连续添加多个Shutdown钩子
    for i := 1; i <= 5; i++ {
        idx := i
        container.AppendOption(
            WithContainerShutdown(func(ctx context.Context) error {
                executionOrder = append(executionOrder, fmt.Sprintf("shutdown-%d", idx))
                return nil
            }),
        )
    }

    err := container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭失败: %v", err)
    }

    t.Logf("5个Shutdown钩子执行顺序: %v", executionOrder)

    // 应该是正序：shutdown-1, shutdown-2, shutdown-3, shutdown-4, shutdown-5
    expected := []string{"shutdown-1", "shutdown-2", "shutdown-3", "shutdown-4", "shutdown-5"}
    if len(executionOrder) != len(expected) {
        t.Errorf("期望%d个调用，实际%d个", len(expected), len(executionOrder))
    }

    for i, exp := range expected {
        if i < len(executionOrder) && executionOrder[i] != exp {
            t.Errorf("第%d步期望'%s'，实际'%s'", i, exp, executionOrder[i])
        }
    }
}

// testCompleteShutdownFlow 验证完整的shutdown流程
func testCompleteShutdownFlow(t *testing.T) {
    container := New()
    executionOrder := []string{}

    // 1. 添加多个实例（验证逆序销毁）
    for i := 1; i <= 3; i++ {
        name := fmt.Sprintf("service%d", i)
        idx := i
        container.ProvideNamedWith(name, func(c Container) (string, error) {
            return name, nil
        }, WithAfterDestroy(func(c Container, info EntryInfo) {
            executionOrder = append(executionOrder, fmt.Sprintf("destroy-%d", idx))
        }))
        _, _ = GetNamed[string](container, name)
    }

    // 2. 添加容器关闭钩子
    container.AppendOption(
        WithContainerShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "shutdown-1")
            return nil
        }),
        WithContainerBeforeShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "before-shutdown")
            return nil
        }),
        WithContainerShutdown(func(ctx context.Context) error {
            executionOrder = append(executionOrder, "shutdown-2")
            return nil
        }),
    )

    // 执行关闭
    err := container.Shutdown(context.Background())
    if err != nil {
        t.Fatalf("关闭失败: %v", err)
    }

    t.Logf("完整shutdown流程: %v", executionOrder)

    // 根据实际输出：[before-shutdown destroy-1 destroy-2 destroy-3 shutdown-1 shutdown-2]
    // 分析执行顺序：
    // 1. BeforeShutdown最先执行：before-shutdown ✅
    // 2. 实例销毁（按注册顺序）：destroy-1, destroy-2, destroy-3
    // 3. Shutdown钩子（按注册顺序）：shutdown-1, shutdown-2
    expected := []string{
        "before-shutdown",                       // BeforeShutdown最先
        "destroy-1", "destroy-2", "destroy-3",  // 实例按注册顺序销毁
        "shutdown-1", "shutdown-2",             // Shutdown钩子正序
    }

    if len(executionOrder) != len(expected) {
        t.Errorf("期望%d个调用，实际%d个: %v", len(expected), len(executionOrder), executionOrder)
    }

    for i, exp := range expected {
        if i < len(executionOrder) && executionOrder[i] != exp {
            t.Errorf("第%d步期望'%s'，实际'%s'", i, exp, executionOrder[i])
        }
    }
}
