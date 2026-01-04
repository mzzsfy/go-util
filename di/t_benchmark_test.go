package di

import (
    "fmt"
    "testing"
)

// BenchmarkServiceRegistration 测试服务注册性能
func BenchmarkServiceRegistration(b *testing.B) {
    container := New()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idx := i
        name := fmt.Sprintf("bench%d", i)
        _ = container.ProvideNamedWith(name, func(c Container) (*BenchService, error) {
            return &BenchService{Value: idx}, nil
        })
    }
}

// BenchmarkServiceGet 测试服务获取性能（包含首次创建）
func BenchmarkServiceGet(b *testing.B) {
    container := New()

    // 预先注册服务
    for i := 0; i < 100; i++ {
        idx := i
        name := fmt.Sprintf("bench%d", i)
        container.ProvideNamedWith(name, func(c Container) (*BenchService, error) {
            return &BenchService{Value: idx}, nil
        })
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        name := fmt.Sprintf("bench%d", i%100)
        _, _ = GetNamed[*BenchService](container, name)
    }
}

// BenchmarkServiceGetCached 测试缓存服务获取性能
func BenchmarkServiceGetCached(b *testing.B) {
    container := New()

    // 预先注册并获取服务（确保缓存）
    for i := 0; i < 100; i++ {
        idx := i
        name := fmt.Sprintf("bench%d", i)
        container.ProvideNamedWith(name, func(c Container) (*BenchService, error) {
            return &BenchService{Value: idx}, nil
        })
        _, _ = GetNamed[*BenchService](container, name) // 预热缓存
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        name := fmt.Sprintf("bench%d", i%100)
        _, _ = GetNamed[*BenchService](container, name)
    }
}

// BenchmarkConfigInjection 测试配置注入性能
func BenchmarkConfigInjection(b *testing.B) {
    container := New()

    // 设置配置
    source := NewMapConfigSource()
    for i := 0; i < 100; i++ {
        source.Set(fmt.Sprintf("bench.key%d", i), fmt.Sprintf("value%d", i))
        source.Set(fmt.Sprintf("bench.port%d", i), 8000+i)
    }
    container.SetConfigSource(source)

    // 注册带配置注入的服务
    type BenchConfigService struct {
        Name string `di.config:"bench.key0"`
        Port int    `di.config:"bench.port0"`
    }

    container.ProvideNamedWith("bench", func(c Container) (*BenchConfigService, error) {
        return &BenchConfigService{}, nil
    })

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = GetNamed[*BenchConfigService](container, "bench")
    }
}

// BenchmarkConcurrentGet 测试并发获取性能
func BenchmarkConcurrentGet(b *testing.B) {
    container := New()

    // 预先注册服务
    for i := 0; i < 10; i++ {
        idx := i
        name := fmt.Sprintf("concurrent%d", i)
        container.ProvideNamedWith(name, func(c Container) (*BenchService, error) {
            return &BenchService{Value: idx}, nil
        })
        _, _ = GetNamed[*BenchService](container, name) // 预热
    }

    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            name := fmt.Sprintf("concurrent%d", i%10)
            _, _ = GetNamed[*BenchService](container, name)
            i++
        }
    })
}

// BenchmarkMixedOperations 测试混合操作性能
func BenchmarkMixedOperations(b *testing.B) {
    container := New()

    // 设置配置
    source := NewMapConfigSource()
    source.Set("bench.app.name", "BenchmarkApp")
    source.Set("bench.app.version", "1.0")
    container.SetConfigSource(source)

    // 注册基础服务
    type BenchDBService struct {
        Name string `di.config:"bench.app.name"`
    }

    container.ProvideNamedWith("", func(c Container) (*BenchDBService, error) {
        return &BenchDBService{}, nil
    })

    // 注册依赖服务
    type BenchAppService struct {
        DB *BenchDBService `di:""`
    }

    container.ProvideNamedWith("", func(c Container) (*BenchAppService, error) {
        return &BenchAppService{}, nil
    })

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = Get[*BenchAppService](container)
    }
}

// BenchmarkStats 测试统计功能的性能
func BenchmarkStats(b *testing.B) {
    container := New()

    // 预先注册和获取一些服务
    for i := 0; i < 10; i++ {
        idx := i
        name := fmt.Sprintf("stats%d", i)
        container.ProvideNamedWith(name, func(c Container) (*BenchService, error) {
            return &BenchService{Value: idx}, nil
        })
        _, _ = GetNamed[*BenchService](container, name)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = container.GetStats()
        _ = container.GetInstanceCount()
        _ = container.GetProviderCount()
        _ = container.GetAverageCreateDuration()
    }
}

// BenchmarkConfigSource 测试配置源操作性能
func BenchmarkConfigSource(b *testing.B) {
    source := NewMapConfigSource()

    // 预先设置一些配置
    for i := 0; i < 50; i++ {
        source.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key := fmt.Sprintf("key%d", i%50)
        _ = source.Get(key)
        _ = source.Has(key)
    }
}

// BenchmarkProvideAndGet 测试注册和获取的组合性能
func BenchmarkProvideAndGet(b *testing.B) {
    container := New()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idx := i
        name := fmt.Sprintf("combined%d", i)
        _ = container.ProvideNamedWith(name, func(c Container) (*BenchService, error) {
            return &BenchService{Value: idx}, nil
        })
        _, _ = GetNamed[*BenchService](container, name)
    }
}

// BenchmarkConfigInjectionWithDefaults 测试带默认值的配置注入性能
func BenchmarkConfigInjectionWithDefaults(b *testing.B) {
    container := New()

    // 设置部分配置（有些配置不存在，会使用默认值）
    source := NewMapConfigSource()
    source.Set("bench.key0", "actualValue")
    container.SetConfigSource(source)

    // 注册带配置注入的服务（包含默认值）
    type BenchConfigDefaultService struct {
        Name    string `di.config:"bench.key0:defaultName"`
        Port    int    `di.config:"bench.port0:8080"`
        Timeout int    `di.config:"bench.timeout:30"`
    }

    container.ProvideNamedWith("bench", func(c Container) (*BenchConfigDefaultService, error) {
        return &BenchConfigDefaultService{}, nil
    })

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = GetNamed[*BenchConfigDefaultService](container, "bench")
    }
}

// BenchService 用于基准测试的服务类型
type BenchService struct {
    Value int
}
