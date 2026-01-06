# Go Dependency Injection Container

一个高性能、类型安全的 Go 依赖注入容器，支持完整的生命周期管理、配置注入和钩子系统。

## ✨ 特性

### 核心功能
- **类型安全**: 基于泛型的类型安全依赖解析
- **命名服务**: 支持同一类型的多个命名实例
- **懒加载**: 按需创建实例，减少启动时间
- **立即加载**: 注册时立即创建实例
- **瞬态模式**: 每次获取都创建新实例

### 高级功能
- **配置注入**: 自动将配置注入到结构体字段
- **生命周期管理**: 支持服务销毁钩子
- **钩子系统**: 完整的创建/销毁前后钩子
- **循环依赖检测**: 自动检测并报告循环依赖
- **作用域隔离**: 支持父容器/子容器隔离
- **并发安全**: 线程安全的容器操作
- **性能监控**: 内置性能统计和指标

## 📦 安装

```bash
go get github.com/mzzsfy/go-util/di
```

## 🚀 快速开始

### 基础使用

```go
package main

import (
    "fmt"
    "github.com/mzzsfy/go-util/di"
)

type Database struct {
    Host string
    Port int
}

type AppService struct {
    DB *Database `di:"main-db"` // 依赖注入标签
}

func main() {
    container := di.New()

    // 注册服务
    container.ProvideNamedWith("main-db", func(c di.Container) (*Database, error) {
        return &Database{Host: "localhost", Port: 5432}, nil
    })

    // 注册应用服务
    container.ProvideNamedWith("app", func(c di.Container) (*AppService, error) {
        return &AppService{}, nil
    })

    // 获取服务（自动注入依赖）
    app, err := di.GetNamed[*AppService](container, "app")
    if err != nil {
        panic(err)
    }

    fmt.Printf("App running with DB: %s:%d\n", app.DB.Host, app.DB.Port)
}
```

### 配置注入

```go
type Config struct {
    DatabaseURL string `di.config:"db.url:localhost:5432"`
    Debug       bool   `di.config:"debug:false"`
    Callback    string `di.config:"https://${callback.url:localhost:8080}"`
}

container.ProvideNamedWith("config", func(c di.Container) (*Config, error) {
    return &Config{}, nil
})

// 自动从配置源读取并注入字段值
config, _ := di.GetNamed[*Config](container, "config")
// config.DatabaseURL = "localhost:5432" (默认值)
// config.Debug = false (默认值)
// config.Callback = https://localhost:8080 (默认值)
```

### 钩子系统

```go
// Provider 级别钩子
container.ProvideNamedWith("service", func(c di.Container) (*Service, error) {
    return &Service{Value: "original"}, nil
},
    di.WithBeforeCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        // 准备创建: info.Name = "service", info.Instance = nil
        return nil, nil // 使用默认创建
    }),
    di.WithAfterCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        // 创建完成: info.Name = "service", info.Instance = *Service
        if service, ok := info.Instance.(*Service); ok {
            service.Value = "modified"
        }
        return info.Instance, nil
    }),
    di.WithAfterDestroy(func(c di.Container, info di.EntryInfo) {
        // 清理资源
        cleanup(info.Instance)
    }),
)

// 容器级别钩子（应用于所有服务）
container := di.NewWithOptions(
    di.WithContainerBeforeCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        fmt.Printf("Creating: %s\n", info.Name)
        return nil, nil
    }),
    di.WithContainerAfterCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        fmt.Printf("Created: %s (%T)\n", info.Name, info.Instance)
        return info.Instance, nil
    }),
)
```

### 生命周期管理

```go
// 实现 ServiceLifecycle 接口
type LifecycleService struct {
    Value string
}

func (s *LifecycleService) Shutdown(ctx context.Context) error {
    fmt.Println("Shutting down...")
    return nil
}

container.ProvideNamedWith("lifecycle", func(c di.Container) (*LifecycleService, error) {
    return &LifecycleService{Value: "test"}, nil
})

// 使用钩子管理生命周期
container.ProvideNamedWith("managed", func(c di.Container) (*Service, error) {
    return &Service{}, nil
}, di.WithAfterDestroy(func(c di.Container, info di.EntryInfo) {
    // 在容器关闭时执行清理
    fmt.Println("Cleaning up managed service")
}))

// 关闭容器（自动调用所有销毁钩子）
container.Shutdown(context.Background())
```

### 加载模式

```go
// 懒加载（默认）- 第一次获取时创建
container.ProvideNamedWith("lazy", func(c di.Container) (*Service, error) {
    return &Service{}, nil
}, di.WithLoadMode(di.LoadModeLazy))

// 立即加载 - 注册时立即创建
container.ProvideNamedWith("eager", func(c di.Container) (*Service, error) {
    return &Service{}, nil
}, di.WithLoadMode(di.LoadModeImmediate))

// 瞬态模式 - 每次获取都创建新实例
container.ProvideNamedWith("transient", func(c di.Container) (*Service, error) {
    return &Service{}, nil
}, di.WithLoadMode(di.LoadModeTransient))
```

### 条件注册

```go
// 只有满足条件时才注册服务
container.ProvideNamedWith("conditional", func(c di.Container) (*Service, error) {
    return &Service{}, nil
}, di.WithCondition(func(c di.Container) bool {
    // 检查环境或其他条件
    return os.Getenv("ENV") == "production"
}))
```

### 作用域隔离

```go
parent := di.New()
parent.ProvideNamedWith("parent-service", func(c di.Container) (*Service, error) {
    return &Service{Value: "parent"}, nil
})

// 创建子容器，继承父容器的配置但服务隔离
child := parent.CreateChildScope()
child.ProvideNamedWith("child-service", func(c di.Container) (*Service, error) {
    return &Service{Value: "child"}, nil
})

// 子容器可以访问父容器服务
parentService, _ := di.GetNamed[*Service](child, "parent-service")

// 父容器不能访问子容器服务
_, err := di.GetNamed[*Service](parent, "child-service") // 错误
```

### 性能监控

```go
stats := container.GetStats()
fmt.Printf("创建实例数: %d\n", stats.CreatedInstances)
fmt.Printf("Get调用次数: %d\n", stats.GetCalls)
fmt.Printf("平均创建耗时: %v\n", stats.CreateDuration/time.Duration(stats.CreatedInstances))
```

## 🔧 API 参考

### 核心接口

#### `Container`
```go
type Container interface {
    ProvideNamedWith(name string, provider any, opts ...ProviderOption) error
    GetNamed(serviceType any, name string) (any, error)
    GetNamedAll(serviceType any) (map[string]any, error)
    HasNamed(serviceType any, name string) bool
    Shutdown(ctx context.Context) error
    CreateChildScope() Container
    // ... 更多方法
}
```

#### `EntryInfo`
```go
type EntryInfo struct {
    Name     string // 服务名称
    Instance any    // 服务实例
}
```

### 钩子函数签名

```go
// 创建前/后钩子
func(Container, EntryInfo) (any, error)

// 销毁前/后钩子
func(Container, EntryInfo)
```

### 选项函数

- `WithBeforeCreate` - 创建前钩子
- `WithAfterCreate` - 创建后钩子
- `WithBeforeDestroy` - 销毁前钩子
- `WithAfterDestroy` - 销毁后钩子
- `WithContainerBeforeCreate` - 容器级别创建前钩子
- `WithContainerAfterCreate` - 容器级别创建后钩子
- `WithContainerBeforeDestroy` - 容器级别销毁前钩子
- `WithContainerAfterDestroy` - 容器级别销毁后钩子
- `WithLoadMode` - 设置加载模式
- `WithCondition` - 设置条件函数

### 加载模式

```go
const (
    LoadModeDefault   LoadMode = iota // 懒加载（第一次Get时创建，之后缓存）
    LoadModeImmediate                  // 立即加载（注册时创建）
    LoadModeLazy                       // 懒加载（延迟创建，检测循环依赖）
    LoadModeTransient                  // 瞬态（每次创建，类似工厂）
)
```

## 🎯 使用场景

### 1. Web 应用
```go
container.ProvideNamedWith("http-server", func(c di.Container) (*http.Server, error) {
    return &http.Server{Addr: ":8080"}, nil
})

container.ProvideNamedWith("router", func(c di.Container) (*mux.Router, error) {
    return mux.NewRouter(), nil
})
```

### 2. 数据库连接
```go
container.ProvideNamedWith("db", func(c di.Container) (*sql.DB, error) {
    return sql.Open("postgres", "connection-string")
}, di.WithAfterDestroy(func(c di.Container, info di.EntryInfo) {
    if db, ok := info.Instance.(*sql.DB); ok {
        db.Close()
    }
}))
```

### 3. 配置管理
```go
type AppConfig struct {
    Port     int    `di.config:"server.port:8080"`
    LogLevel string `di.config:"log.level:info"`
}

container.ProvideNamedWith("config", func(c di.Container) (*AppConfig, error) {
    return &AppConfig{}, nil
})
```

### 4. 缓存服务
```go
container.ProvideNamedWith("redis", func(c di.Container) (*redis.Client, error) {
    return redis.NewClient(&redis.Options{Addr: "localhost:6379"}), nil
}, di.WithLoadMode(di.LoadModeImmediate)) // 提前连接
```

## ⚡ 性能

- **注册 1000 个服务**: ~1ms
- **首次获取 1000 个服务**: ~3ms
- **缓存获取 1000 个服务**: <1ms
- **并发安全**: 5000 请求/100 goroutines ~1ms
- **QPS**: 400万+

## 🔒 线程安全

所有容器操作都是线程安全的，可以在多个 goroutine 中安全使用。

## 🚨 限制

1. **基本类型限制**: `string`, `int`, `context.Context` 等基本类型不能注册为空名称服务
2. **结构体要求**: 依赖注入要求目标是结构体或结构体指针
3. **循环依赖**: 会检测并返回错误，不支持循环依赖

## 📝 最佳实践

1. **使用命名服务**: 避免类型冲突，提高可读性
2. **合理使用钩子**: 避免在钩子中执行耗时操作
3. **及时清理**: 在销毁钩子中释放资源
4. **错误处理**: 始终检查返回的错误
5. **配置注入**: 使用标签简化配置管理

## 🧪 测试

```bash
cd di
go test -v
```

## 📄 许可证

MIT License