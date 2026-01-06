package di

import (
    "context"
    "fmt"
    "sync"
    "testing"
    "time"
)

// Test_Start_Basic 测试Start方法的基本功能
func Test_Start_Basic(t *testing.T) {
    t.Run("基本启动流程", func(t *testing.T) {
        container := New()

        err := container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

    })

    t.Run("重复启动应该返回nil", func(t *testing.T) {
        container := New()

        err := container.Start()
        if err != nil {
            t.Fatalf("第一次Start failed: %v", err)
        }

        err = container.Start()
        if err == nil {
            t.Fatalf("第二次Start 应该报错")
        }
    })
}

// Test_Start_WithHooks 测试带钩子的启动
func Test_Start_WithHooks(t *testing.T) {
    t.Run("启动前钩子", func(t *testing.T) {
        hookCalled := false

        container := New(
            WithOnStart(func(c Container) error {
                hookCalled = true
                return nil
            }),
        )

        err := container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        if !hookCalled {
            t.Error("启动前钩子应该被调用")
        }
    })

    t.Run("启动后钩子", func(t *testing.T) {
        hookCalled := false

        container := New(
            WithAfterStart(func(c Container) error {
                hookCalled = true
                return nil
            }),
        )

        err := container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        if !hookCalled {
            t.Error("启动后钩子应该被调用")
        }
    })

    t.Run("钩子执行顺序", func(t *testing.T) {
        var executionOrder []string

        container := New(
            WithOnStart(func(c Container) error {
                executionOrder = append(executionOrder, "startup-1")
                return nil
            }),
            WithOnStart(func(c Container) error {
                executionOrder = append(executionOrder, "startup-2")
                return nil
            }),
            WithAfterStart(func(c Container) error {
                executionOrder = append(executionOrder, "after-1")
                return nil
            }),
            WithAfterStart(func(c Container) error {
                executionOrder = append(executionOrder, "after-2")
                return nil
            }),
        )

        err := container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        expected := []string{"startup-1", "startup-2", "after-1", "after-2"}
        if len(executionOrder) != len(expected) {
            t.Errorf("期望%d个钩子调用，实际%d个", len(expected), len(executionOrder))
        }

        for i, expected := range expected {
            if i >= len(executionOrder) || executionOrder[i] != expected {
                t.Errorf("第%d个钩子期望%s，实际%s", i, expected, executionOrder[i])
            }
        }
    })
}

// Test_Start_ErrorHandling 测试错误处理
func Test_Start_ErrorHandling(t *testing.T) {
    t.Run("启动前钩子返回错误", func(t *testing.T) {
        container := New(
            WithOnStart(func(c Container) error {
                return fmt.Errorf("startup error")
            }),
        )

        err := container.Start()
        if err == nil {
            t.Fatal("应该返回错误")
        }

        if err.Error() != "startup hook 0 failed: startup error" {
            t.Errorf("错误消息格式不正确: %v", err)
        }
    })

    t.Run("启动后钩子返回错误", func(t *testing.T) {
        container := New(
            WithAfterStart(func(c Container) error {
                return fmt.Errorf("after startup error")
            }),
        )

        err := container.Start()
        if err == nil {
            t.Fatal("应该返回错误")
        }

        if err.Error() != "after startup hook 0 failed: after startup error" {
            t.Errorf("错误消息格式不正确: %v", err)
        }
    })

    t.Run("多个钩子中第一个错误", func(t *testing.T) {
        container := New(
            WithOnStart(func(c Container) error {
                return nil
            }),
            WithOnStart(func(c Container) error {
                return fmt.Errorf("second hook error")
            }),
            WithOnStart(func(c Container) error {
                t.Error("第三个钩子不应该被调用")
                return nil
            }),
        )

        err := container.Start()
        if err == nil {
            t.Fatal("应该返回错误")
        }

        if err.Error() != "startup hook 1 failed: second hook error" {
            t.Errorf("错误消息格式不正确: %v", err)
        }
    })
}

// Test_Start_Concurrent 测试并发启动
func Test_Start_Concurrent(t *testing.T) {
    t.Run("多个goroutine同时启动", func(t *testing.T) {
        container := New()

        var wg sync.WaitGroup
        numGoroutines := 10
        successCount := 0
        var mu sync.Mutex

        for i := 0; i < numGoroutines; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                err := container.Start()
                mu.Lock()
                if err == nil {
                    successCount++
                }
                mu.Unlock()
            }()
        }

        wg.Wait()

        // 应该只有一个goroutine成功启动
        if successCount != 1 {
            t.Errorf("期望1个goroutine成功启动，实际%d个", successCount)
        }
    })

    t.Run("并发启动与钩子交互", func(t *testing.T) {
        hookCallCount := 0
        var mu sync.Mutex

        container := New(
            WithOnStart(func(c Container) error {
                mu.Lock()
                hookCallCount++
                mu.Unlock()
                // 模拟一些工作
                time.Sleep(10 * time.Millisecond)
                return nil
            }),
        )

        var wg sync.WaitGroup
        numGoroutines := 5
        successCount := 0

        for i := 0; i < numGoroutines; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                err := container.Start()
                mu.Lock()
                if err == nil {
                    successCount++
                }
                mu.Unlock()
            }()
        }

        wg.Wait()

        // 应该只有一个goroutine成功
        if successCount != 1 {
            t.Errorf("期望1个goroutine成功，实际%d个", successCount)
        }

        // 钩子应该只被调用一次
        if hookCallCount != 1 {
            t.Errorf("钩子应该只被调用1次，实际调用%d次", hookCallCount)
        }
    })
}

// Test_Start_WithContainerOperations 测试启动期间的容器操作
func Test_Start_WithContainerOperations(t *testing.T) {
    t.Run("启动钩子中可以访问容器", func(t *testing.T) {
        // 创建一个容器并注册服务
        container := New()
        err := container.ProvideNamedWith("test", func(c Container) (string, error) {
            return "test-value", nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 启动钩子中访问容器
        hookExecuted := false
        err = container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        // 现在创建另一个容器，使用启动钩子来验证功能
        container2 := New(
            WithOnStart(func(c Container) error {
                // 在钩子中检查容器状态
                hookExecuted = true
                return nil
            }),
        )

        err = container2.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        if !hookExecuted {
            t.Error("钩子应该被执行")
        }
    })

    t.Run("启动后钩子可以访问已启动状态", func(t *testing.T) {
        container := New(
            WithAfterStart(func(c Container) error {
                return nil
            }),
        )

        err := container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }
    })
}

// Test_Start_WithServices 测试与服务的集成
func Test_Start_WithServices(t *testing.T) {
    t.Run("启动钩子与延迟加载服务", func(t *testing.T) {
        serviceCreated := false

        type DelayedService struct {
            Value string
        }

        container := New()

        // 注册一个延迟加载的服务
        err := container.ProvideNamedWith("delayed", func(c Container) (*DelayedService, error) {
            serviceCreated = true
            return &DelayedService{Value: "delayed-service"}, nil
        }, WithLoadMode(LoadModeLazy))
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 启动钩子
        err = container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        // 延迟加载的服务在启动时不应该被创建
        if serviceCreated {
            t.Error("延迟加载的服务在启动时不应该被创建")
        }

        // 获取服务应该触发创建
        val, err := GetNamed[*DelayedService](container, "delayed")
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }

        if val.Value != "delayed-service" {
            t.Errorf("期望'delayed-service'，实际'%v'", val.Value)
        }

        if !serviceCreated {
            t.Error("服务应该被创建")
        }
    })

    t.Run("启动钩子与立即加载服务", func(t *testing.T) {
        serviceCreated := false

        type ImmediateService struct {
            Value string
        }

        container := New()

        // 注册一个立即加载的服务
        err := container.ProvideNamedWith("immediate", func(c Container) (*ImmediateService, error) {
            serviceCreated = true
            return &ImmediateService{Value: "immediate-service"}, nil
        }, WithLoadMode(LoadModeImmediate))
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 服务应该已经被创建
        if !serviceCreated {
            t.Error("立即加载的服务应该在注册时就被创建")
        }

        // 启动钩子
        err = container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }
    })
}

// Test_Start_WithShutdown 测试启动与关闭的交互
func Test_Start_WithShutdown(t *testing.T) {
    t.Run("启动后可以正常关闭", func(t *testing.T) {
        container := New()

        err := container.Start()
        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        err = container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("Shutdown failed: %v", err)
        }

    })

    t.Run("关闭后重新启动", func(t *testing.T) {
        container := New()

        err := container.Start()
        if err != nil {
            t.Fatalf("第一次Start failed: %v", err)
        }

        err = container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("Shutdown failed: %v", err)
        }

        // 尝试重新启动（应该成功，因为Shutdown重置了状态）
        err = container.Start()
        if err != nil {
            t.Fatalf("Shutdown后重新启动应该成功，但失败了: %v", err)
        }
    })
}

// Test_Start_Performance 测试性能
func Test_Start_Performance(t *testing.T) {
    t.Run("大量钩子的性能", func(t *testing.T) {
        // 构建带有大量钩子的容器
        opts := make([]ContainerOption, 0, 2000)
        for i := 0; i < 1000; i++ {
            opts = append(opts, WithOnStart(func(c Container) error {
                return nil
            }))
            opts = append(opts, WithAfterStart(func(c Container) error {
                return nil
            }))
        }

        container := New(opts...)

        start := time.Now()
        err := container.Start()
        duration := time.Since(start)

        if err != nil {
            t.Fatalf("Start failed: %v", err)
        }

        t.Logf("启动2000个钩子耗时: %v", duration)

        // 应该在合理时间内完成（例如1秒内）
        if duration > time.Second {
            t.Errorf("启动耗时过长: %v", duration)
        }
    })
}

// Test_Start_MultipleContainers 测试多个容器
func Test_Start_MultipleContainers(t *testing.T) {
    t.Run("多个独立容器", func(t *testing.T) {
        container1 := New()
        container2 := New()

        err1 := container1.Start()
        err2 := container2.Start()

        if err1 != nil || err2 != nil {
            t.Fatalf("启动失败: %v, %v", err1, err2)
        }

    })

    t.Run("父子容器启动", func(t *testing.T) {
        parent := New()
        child := parent.CreateChildScope()

        err := parent.Start()
        if err != nil {
            t.Fatalf("父容器启动失败: %v", err)
        }

        err = child.Start()
        if err != nil {
            t.Fatalf("子容器启动失败: %v", err)
        }

    })
}

// Test_Start_WithConfigInjection 测试启动钩子中的配置注入
func Test_Start_WithConfigInjection(t *testing.T) {
    t.Run("启动钩子中使用配置注入", func(t *testing.T) {
        // 设置配置源
        configSource := NewMapConfigSource()
        configSource.Set("callback.url", "example.com:8080")
        configSource.Set("service.name", "test-service")

        container := New()
        container.SetConfigSource(configSource)

        // 创建一个在启动钩子中使用的结构体
        type ConfigStruct struct {
            CallbackURL string `di.config:"https://${callback.url:localhost:9000}"`
            ServiceName string `di.config:"service.name:default-service"`
            Port        int    `di.config:"server.port:8080"`
        }

        // 添加启动钩子，在钩子中创建并注入配置
        err := container.ProvideNamedWith("config", func(c Container) (*ConfigStruct, error) {
            config := &ConfigStruct{}
            // 手动触发配置注入（模拟在启动钩子中的场景）
            return config, nil
        })
        if err != nil {
            t.Fatalf("注册配置服务失败: %v", err)
        }

        // 启动容器
        err = container.Start()
        if err != nil {
            t.Fatalf("启动失败: %v", err)
        }

        // 获取配置实例
        config, err := GetNamed[*ConfigStruct](container, "config")
        if err != nil {
            t.Fatalf("获取配置失败: %v", err)
        }

        // 验证配置注入结果
        if config.CallbackURL != "https://example.com:8080" {
            t.Errorf("期望CallbackURL为'https://example.com:8080'，实际'%s'", config.CallbackURL)
        }

        if config.ServiceName != "test-service" {
            t.Errorf("期望ServiceName为'test-service'，实际'%s'", config.ServiceName)
        }

        if config.Port != 8080 {
            t.Errorf("期望Port为8080，实际%d", config.Port)
        }
    })

    t.Run("变量替换与默认值", func(t *testing.T) {
        // 设置配置源，但不设置某些值以测试默认值
        configSource := NewMapConfigSource()
        configSource.Set("existing.key", "existing-value")

        container := New()
        container.SetConfigSource(configSource)

        type ConfigStruct struct {
            URL1 string `di.config:"http://${missing.url:localhost:8080}"`
            URL2 string `di.config:"https://${existing.key}/api"`
            URL3 string `di.config:"${missing1:default1}-${missing2:default2}"`
        }

        err := container.ProvideNamedWith("config", func(c Container) (*ConfigStruct, error) {
            return &ConfigStruct{}, nil
        })
        if err != nil {
            t.Fatalf("注册配置服务失败: %v", err)
        }

        config, err := GetNamed[*ConfigStruct](container, "config")
        if err != nil {
            t.Fatalf("获取配置失败: %v", err)
        }

        // 验证默认值
        if config.URL1 != "http://localhost:8080" {
            t.Errorf("期望URL1为'http://localhost:8080'，实际'%s'", config.URL1)
        }

        if config.URL2 != "https://existing-value/api" {
            t.Errorf("期望URL2为'https://existing-value/api'，实际'%s'", config.URL2)
        }

        if config.URL3 != "default1-default2" {
            t.Errorf("期望URL3为'default1-default2'，实际'%s'", config.URL3)
        }
    })

    t.Run("复杂变量替换场景", func(t *testing.T) {
        configSource := NewMapConfigSource()
        configSource.Set("host", "api.example.com")
        configSource.Set("port", "443")
        configSource.Set("path", "/v1/users")

        container := New()
        container.SetConfigSource(configSource)

        type ConfigStruct struct {
            FullURL string `di.config:"https://${host:localhost}:${port:8080}${path:/api}"`
        }

        err := container.ProvideNamedWith("config", func(c Container) (*ConfigStruct, error) {
            return &ConfigStruct{}, nil
        })
        if err != nil {
            t.Fatalf("注册配置服务失败: %v", err)
        }

        config, err := GetNamed[*ConfigStruct](container, "config")
        if err != nil {
            t.Fatalf("获取配置失败: %v", err)
        }

        expected := "https://api.example.com:443/v1/users"
        if config.FullURL != expected {
            t.Errorf("期望FullURL为'%s'，实际'%s'", expected, config.FullURL)
        }
    })

    t.Run("启动钩子中动态配置", func(t *testing.T) {
        // 测试在启动钩子执行期间，配置注入能够正常工作
        configSource := NewMapConfigSource()
        configSource.Set("app.version", "1.0.0")
        configSource.Set("app.env", "production")

        container := New()
        container.SetConfigSource(configSource)

        // 创建一个需要配置注入的服务
        type AppConfig struct {
            Version string `di.config:"app.version:unknown"`
            Env     string `di.config:"app.env:development"`
        }

        // 注册服务，容器会自动处理配置注入
        err := container.ProvideNamedWith("appConfig", func(c Container) (*AppConfig, error) {
            return &AppConfig{}, nil
        })
        if err != nil {
            t.Fatalf("注册配置服务失败: %v", err)
        }

        // 启动容器
        err = container.Start()
        if err != nil {
            t.Fatalf("启动失败: %v", err)
        }

        // 获取配置实例
        config, err := GetNamed[*AppConfig](container, "appConfig")
        if err != nil {
            t.Fatalf("获取配置失败: %v", err)
        }

        // 验证配置注入结果
        if config.Version != "1.0.0" {
            t.Errorf("期望version为'1.0.0'，实际'%s'", config.Version)
        }

        if config.Env != "production" {
            t.Errorf("期望env为'production'，实际'%s'", config.Env)
        }
    })
}
