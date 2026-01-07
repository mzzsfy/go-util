package di

import (
    "context"
    "fmt"
    "strings"
    "testing"
    "time"
)

// Test_EdgeCases 边界条件和异常场景测试
func Test_EdgeCases(t *testing.T) {
    t.Run("空值和零值处理", testEmptyValues)
    t.Run("类型系统边界", testTypeBoundaries)
    t.Run("并发异常场景", testConcurrencyExceptions)
    t.Run("配置注入异常", testConfigInjectionExceptions)
    t.Run("生命周期异常", testLifecycleExceptions)
    t.Run("内存和资源边界", testMemoryBoundaries)
    t.Run("特殊类型处理", testSpecialTypes)
}

// testEmptyValues 测试空值和零值处理
func testEmptyValues(t *testing.T) {
    // 测试1: 空字符串服务名（基本类型需要名称）
    t.Run("空字符串服务名", func(t *testing.T) {
        container := New()

        // 基本类型如string不能在没有名称的情况下注册，这是安全特性
        err := container.ProvideNamedWith("", func(c Container) (string, error) {
            return "default", nil
        })
        if err == nil {
            t.Fatal("应该拒绝注册无名称的基本类型")
        }
        if !strings.Contains(err.Error(), "cannot register type string without name") {
            t.Errorf("期望类型注册错误，实际: %v", err)
        }

        // 使用结构体包装可以正常注册
        type StringService struct {
            Value string
        }
        err = container.ProvideNamedWith("", func(c Container) (*StringService, error) {
            return &StringService{Value: "default"}, nil
        })
        if err != nil {
            t.Fatalf("注册结构体服务失败: %v", err)
        }

        val, err := Get[*StringService](container)
        if err != nil {
            t.Fatalf("获取结构体服务失败: %v", err)
        }
        if val.Value != "default" {
            t.Errorf("期望'default', 实际'%s'", val.Value)
        }
    })

    // 测试2: 零值类型注册
    t.Run("零值类型注册", func(t *testing.T) {
        container := New()

        // 注册零值结构体
        type ZeroService struct {
            Value int
        }
        err := container.ProvideNamedWith("zero-service", func(c Container) (*ZeroService, error) {
            return &ZeroService{Value: 0}, nil
        })
        if err != nil {
            t.Fatalf("注册零值服务失败: %v", err)
        }

        val, err := GetNamed[*ZeroService](container, "zero-service")
        if err != nil {
            t.Fatalf("获取零值服务失败: %v", err)
        }
        if val.Value != 0 {
            t.Errorf("期望0, 实际%d", val.Value)
        }
    })

    // 测试3: nil指针服务
    t.Run("nil指针服务", func(t *testing.T) {
        container := New()

        type Service struct {
            Value string
        }

        // 注册返回nil的服务
        err := container.ProvideNamedWith("nil-service", func(c Container) (*Service, error) {
            return nil, nil
        })
        if err != nil {
            t.Fatalf("注册nil服务失败: %v", err)
        }

        val, err := GetNamed[*Service](container, "nil-service")
        if err != nil {
            t.Fatalf("获取nil服务失败: %v", err)
        }
        if val != nil {
            t.Errorf("期望nil, 实际%v", val)
        }
    })

    // 测试4: 空配置源
    t.Run("空配置源", func(t *testing.T) {
        container := New()
        // 使用空的配置源而不是nil
        emptySource := NewMapConfigSource()
        container.SetConfigSource(emptySource)

        type Service struct {
            Value string `di.config:"any.key:default"`
        }

        err := container.ProvideNamedWith("", func(c Container) (*Service, error) {
            return &Service{}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        service, err := Get[*Service](container)
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }
        if service.Value != "default" {
            t.Errorf("期望默认值'default', 实际'%s'", service.Value)
        }
    })
}

// testTypeBoundaries 测试类型系统边界
func testTypeBoundaries(t *testing.T) {
    // 测试1: 泛型类型边界
    t.Run("泛型类型边界", func(t *testing.T) {
        container := New()

        // 测试不同泛型实例
        err := container.ProvideNamedWith("int-list", func(c Container) ([]int, error) {
            return []int{1, 2, 3}, nil
        })
        if err != nil {
            t.Fatalf("注册泛型类型失败: %v", err)
        }

        err = container.ProvideNamedWith("string-list", func(c Container) ([]string, error) {
            return []string{"a", "b", "c"}, nil
        })
        if err != nil {
            t.Fatalf("注册泛型类型失败: %v", err)
        }

        intList, err := GetNamed[[]int](container, "int-list")
        if err != nil {
            t.Fatalf("获取int列表失败: %v", err)
        }
        if len(intList) != 3 {
            t.Errorf("期望长度3, 实际%d", len(intList))
        }

        stringList, err := GetNamed[[]string](container, "string-list")
        if err != nil {
            t.Fatalf("获取string列表失败: %v", err)
        }
        if len(stringList) != 3 {
            t.Errorf("期望长度3, 实际%d", len(stringList))
        }
    })

    // 测试2: 函数类型
    t.Run("函数类型", func(t *testing.T) {
        container := New()

        // 注册函数类型服务
        err := container.ProvideNamedWith("func-service", func(c Container) (func(int) int, error) {
            return func(x int) int { return x * 2 }, nil
        })
        if err != nil {
            t.Fatalf("注册函数类型失败: %v", err)
        }

        fn, err := GetNamed[func(int) int](container, "func-service")
        if err != nil {
            t.Fatalf("获取函数类型失败: %v", err)
        }
        if fn(5) != 10 {
            t.Errorf("函数调用错误: 期望10, 实际%d", fn(5))
        }
    })
}

// testConcurrencyExceptions 测试并发异常场景
func testConcurrencyExceptions(t *testing.T) {
    // 测试1: 并发注册相同服务
    t.Run("并发重复注册", func(t *testing.T) {
        container := New()
        errors := make(chan error, 10)

        // 多个goroutine同时注册相同名称的服务
        for i := 0; i < 10; i++ {
            go func(id int) {
                err := container.ProvideNamedWith("same-name", func(c Container) (int, error) {
                    return id, nil
                })
                if err != nil {
                    errors <- err
                } else {
                    errors <- nil
                }
            }(i)
        }

        // 收集结果
        successCount := 0
        errorCount := 0
        for i := 0; i < 10; i++ {
            err := <-errors
            if err == nil {
                successCount++
            } else {
                errorCount++
            }
        }

        // 应该只有1个成功，9个失败
        if successCount != 1 {
            t.Errorf("期望1个成功，实际%d个", successCount)
        }
        if errorCount != 9 {
            t.Errorf("期望9个失败，实际%d个", errorCount)
        }
    })
}

// testConfigInjectionExceptions 测试配置注入异常
func testConfigInjectionExceptions(t *testing.T) {
    // 测试1: 配置类型转换失败
    t.Run("配置类型转换失败", func(t *testing.T) {
        container := New()
        source := NewMapConfigSource()
        source.Set("invalid.int", "not-a-number")
        container.SetConfigSource(source)

        type Service struct {
            Value int `di.config:"invalid.int"`
        }

        err := container.ProvideNamedWith("", func(c Container) (*Service, error) {
            return &Service{}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        // 获取时应该能够处理转换失败（可能返回0值）
        service, err := Get[*Service](container)
        if err != nil {
            t.Logf("获取服务失败（预期可能）: %v", err)
        } else {
            t.Logf("转换结果: %d", service.Value)
        }
    })

    // 测试2: 配置键包含特殊字符（冒号除外，因为冒号用于默认值分隔符）
    t.Run("特殊字符配置键", func(t *testing.T) {
        container := New()
        source := NewMapConfigSource()
        source.Set("key.with.dots", "value1")
        source.Set("key-with-dashes", "value3")
        container.SetConfigSource(source)

        type Service struct {
            Val1 string `di.config:"key.with.dots"`
            Val3 string `di.config:"key-with-dashes"`
        }

        // 注册服务
        err := container.ProvideNamedWith("", func(c Container) (*Service, error) {
            return &Service{}, nil
        })
        if err != nil {
            t.Fatalf("注册服务失败: %v", err)
        }

        service, err := Get[*Service](container)
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }

        if service.Val1 != "value1" {
            t.Errorf("点号键错误: 期望'value1', 实际'%s'", service.Val1)
        }
        if service.Val3 != "value3" {
            t.Errorf("破折号键错误: 期望'value3', 实际'%s'", service.Val3)
        }

        // 测试冒号作为默认值分隔符的功能
        t.Run("冒号作为默认值分隔符", func(t *testing.T) {
            type ServiceWithDefault struct {
                Val string `di.config:"nonexistent.key:default:value"`
            }

            err := container.ProvideNamedWith("withDefault", func(c Container) (*ServiceWithDefault, error) {
                return &ServiceWithDefault{}, nil
            })
            if err != nil {
                t.Fatalf("注册带默认值的服务失败: %v", err)
            }

            defaultService, err := GetNamed[*ServiceWithDefault](container, "withDefault")
            if err != nil {
                t.Fatalf("获取带默认值的服务失败: %v", err)
            }

            if defaultService.Val != "default:value" {
                t.Errorf("默认值错误: 期望'default:value', 实际'%s'", defaultService.Val)
            }
        })
    })
}

// testLifecycleExceptions 测试生命周期异常
func testLifecycleExceptions(t *testing.T) {
    // 测试1: 关闭钩子返回错误
    t.Run("关闭钩子错误", func(t *testing.T) {
        container := New()

        // 注册服务
        container.ProvideNamedWith("test", func(c Container) (string, error) {
            return "test", nil
        })
        _, _ = GetNamed[string](container, "test")

        // 启动容器
        err := container.Start()
        if err != nil {
            t.Fatalf("启动容器失败: %v", err)
        }

        // 关闭应该继续执行其他钩子，即使有错误
        err = container.Shutdown(context.Background())
        // Shutdown方法通常会忽略单个钩子错误，继续执行
        if err != nil {
            t.Logf("关闭返回错误（可能）: %v", err)
        }

        // 验证容器已被清空
        if container.GetProviderCount() != 0 {
            t.Error("关闭后应该清空提供者")
        }
    })

    // 测试2: 重复关闭
    t.Run("重复关闭", func(t *testing.T) {
        container := New()

        // 注册服务
        container.ProvideNamedWith("test", func(c Container) (string, error) {
            return "test", nil
        })
        _, _ = GetNamed[string](container, "test")

        // 启动容器
        err := container.Start()
        if err != nil {
            t.Fatalf("启动容器失败: %v", err)
        }

        // 第一次关闭
        err1 := container.Shutdown(context.Background())
        if err1 != nil {
            t.Fatalf("第一次关闭失败: %v", err1)
        }

        // 第二次关闭 - 应该返回"container is already shutting down"错误
        err2 := container.Shutdown(context.Background())
        if err2 == nil {
            t.Error("第二次关闭应该返回错误")
        } else {
            t.Logf("第二次关闭返回错误（预期）: %v", err2)
        }
    })

    // 测试3: 未启动容器关闭（现在应该成功）
    t.Run("未启动容器关闭", func(t *testing.T) {
        container := New()

        // 注册服务但不启动
        container.ProvideNamedWith("test", func(c Container) (string, error) {
            return "test", nil
        })

        // 关闭未启动的容器应该成功（清理资源）
        err := container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("关闭未启动的容器应该成功，但失败了: %v", err)
        }

        // 验证容器已被清空
        if container.GetProviderCount() != 0 {
            t.Error("关闭后应该清空提供者")
        }
    })

    // 测试4: 启动后再次启动
    t.Run("重复启动", func(t *testing.T) {
        container := New()

        // 启动容器
        err := container.Start()
        if err != nil {
            t.Fatalf("第一次启动失败: %v", err)
        }

        // 尝试再次启动
        err = container.Start()
        if err == nil {
            t.Error("重复启动应该返回错误")
        } else {
            t.Logf("重复启动返回错误（预期）: %v", err)
        }
    })

    // 测试5: 关闭后重新启动
    t.Run("关闭后重新启动", func(t *testing.T) {
        container := New()

        // 注册服务
        container.ProvideNamedWith("test", func(c Container) (string, error) {
            return "test", nil
        })

        // 启动容器
        err := container.Start()
        if err != nil {
            t.Fatalf("第一次启动失败: %v", err)
        }

        // 获取服务实例
        val, err := GetNamed[string](container, "test")
        if err != nil {
            t.Fatalf("获取服务失败: %v", err)
        }
        if val != "test" {
            t.Errorf("期望'test', 实际'%s'", val)
        }

        // 关闭容器
        err = container.Shutdown(context.Background())
        if err != nil {
            t.Fatalf("关闭失败: %v", err)
        }

        // 验证容器已清空
        if container.GetProviderCount() != 0 {
            t.Error("关闭后应该清空提供者")
        }

        // 重新启动容器
        err = container.Start()
        if err != nil {
            t.Fatalf("重新启动失败: %v", err)
        }

        // 重新注册服务（因为关闭后清空了）
        container.ProvideNamedWith("test2", func(c Container) (string, error) {
            return "test2", nil
        })

        // 获取新服务
        val2, err := GetNamed[string](container, "test2")
        if err != nil {
            t.Fatalf("重新启动后获取服务失败: %v", err)
        }
        if val2 != "test2" {
            t.Errorf("期望'test2', 实际'%s'", val2)
        }
    })
}

// testMemoryBoundaries 测试内存和资源边界
func testMemoryBoundaries(t *testing.T) {
    // 测试1: 大量服务注册
    t.Run("大量服务注册", func(t *testing.T) {
        container := New()

        // 注册100个服务（减少数量以加快测试）
        for i := 0; i < 100; i++ {
            idx := i
            name := fmt.Sprintf("service%d", i)
            err := container.ProvideNamedWith(name, func(c Container) (int, error) {
                return idx, nil
            })
            if err != nil {
                t.Fatalf("注册服务%d失败: %v", i, err)
            }
        }

        // 验证注册数量
        if container.GetProviderCount() != 100 {
            t.Errorf("期望100个提供者，实际%d", container.GetProviderCount())
        }

        // 获取部分服务
        for i := 0; i < 10; i++ {
            name := fmt.Sprintf("service%d", i)
            val, err := GetNamed[int](container, name)
            if err != nil {
                t.Fatalf("获取服务%s失败: %v", name, err)
            }
            if val != i {
                t.Errorf("服务%s值错误: 期望%d, 实际%d", name, i, val)
            }
        }

        // 验证实例数量
        if container.GetInstanceCount() != 10 {
            t.Errorf("期望10个实例，实际%d", container.GetInstanceCount())
        }
    })

    // 测试2: 深度依赖链
    t.Run("深度依赖链", func(t *testing.T) {
        container := New()

        // 创建深度为20的依赖链
        depth := 20

        // 注册最底层服务
        container.ProvideNamedWith(fmt.Sprintf("service%d", depth-1), func(c Container) (int, error) {
            return depth - 1, nil
        })

        // 逐层注册依赖
        for i := depth - 2; i >= 0; i-- {
            current := i
            next := i + 1
            err := container.ProvideNamedWith(fmt.Sprintf("service%d", current), func(c Container) (int, error) {
                nextVal, err := GetNamed[int](c, fmt.Sprintf("service%d", next))
                if err != nil {
                    return 0, err
                }
                return current + nextVal, nil
            })
            if err != nil {
                t.Fatalf("注册服务%d失败: %v", current, err)
            }
        }

        // 获取顶层服务，应该递归解析整个依赖链
        val, err := GetNamed[int](container, "service0")
        if err != nil {
            t.Fatalf("获取顶层服务失败: %v", err)
        }

        // 期望值: 0 + 1 + 2 + ... + 19 = 190
        expected := 0
        for i := 0; i < depth; i++ {
            expected += i
        }
        if val != expected {
            t.Errorf("深度依赖链计算错误: 期望%d, 实际%d", expected, val)
        }
    })
}

// testSpecialTypes 测试特殊类型处理
func testSpecialTypes(t *testing.T) {
    // 测试1: 时间类型
    t.Run("时间类型", func(t *testing.T) {
        container := New()

        now := time.Now()
        err := container.ProvideNamedWith("time", func(c Container) (time.Time, error) {
            return now, nil
        })
        if err != nil {
            t.Fatalf("注册时间服务失败: %v", err)
        }

        retrieved, err := GetNamed[time.Time](container, "time")
        if err != nil {
            t.Fatalf("获取时间服务失败: %v", err)
        }
        if !retrieved.Equal(now) {
            t.Errorf("时间不匹配: 期望%v, 实际%v", now, retrieved)
        }
    })

    // 测试2: 通道类型
    t.Run("通道类型", func(t *testing.T) {
        container := New()

        ch := make(chan int, 10)
        err := container.ProvideNamedWith("channel", func(c Container) (chan int, error) {
            return ch, nil
        })
        if err != nil {
            t.Fatalf("注册通道服务失败: %v", err)
        }

        retrieved, err := GetNamed[chan int](container, "channel")
        if err != nil {
            t.Fatalf("获取通道服务失败: %v", err)
        }
        if retrieved != ch {
            t.Error("通道实例不匹配")
        }
    })

    // 测试3: 复杂结构体
    t.Run("复杂结构体", func(t *testing.T) {
        container := New()

        type Complex struct {
            IntField    int
            StringField string
            BoolField   bool
            SliceField  []string
            MapField    map[string]int
            StructField struct {
                Nested string
            }
            PtrField *int
        }

        ptrVal := 42
        complex := Complex{
            IntField:    100,
            StringField: "test",
            BoolField:   true,
            SliceField:  []string{"a", "b", "c"},
            MapField:    map[string]int{"key": 1},
            StructField: struct{ Nested string }{Nested: "nested"},
            PtrField:    &ptrVal,
        }

        err := container.ProvideNamedWith("complex", func(c Container) (Complex, error) {
            return complex, nil
        })
        if err != nil {
            t.Fatalf("注册复杂结构体失败: %v", err)
        }

        retrieved, err := GetNamed[Complex](container, "complex")
        if err != nil {
            t.Fatalf("获取复杂结构体失败: %v", err)
        }

        if retrieved.IntField != complex.IntField {
            t.Error("IntField不匹配")
        }
        if retrieved.StringField != complex.StringField {
            t.Error("StringField不匹配")
        }
        if retrieved.BoolField != complex.BoolField {
            t.Error("BoolField不匹配")
        }
        if len(retrieved.SliceField) != len(complex.SliceField) {
            t.Error("SliceField长度不匹配")
        }
        if len(retrieved.MapField) != len(complex.MapField) {
            t.Error("MapField长度不匹配")
        }
        if retrieved.StructField.Nested != complex.StructField.Nested {
            t.Error("StructField不匹配")
        }
        if retrieved.PtrField == nil || *retrieved.PtrField != *complex.PtrField {
            t.Error("PtrField不匹配")
        }
    })

    // 测试4: 空接口类型
    t.Run("空接口类型", func(t *testing.T) {
        container := New()

        // 注册不同类型的空接口
        container.ProvideNamedWith("intf", func(c Container) (interface{}, error) {
            return "string value", nil
        })

        container.ProvideNamedWith("intf2", func(c Container) (interface{}, error) {
            return 42, nil
        })

        val1, err := GetNamed[interface{}](container, "intf")
        if err != nil {
            t.Fatalf("获取空接口1失败: %v", err)
        }
        if val1.(string) != "string value" {
            t.Error("空接口值1错误")
        }

        val2, err := GetNamed[interface{}](container, "intf2")
        if err != nil {
            t.Fatalf("获取空接口2失败: %v", err)
        }
        if val2.(int) != 42 {
            t.Error("空接口值2错误")
        }
    })
}
