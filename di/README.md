# Go Dependency Injection Container

一个高性能、类型安全的 Go 依赖注入容器，支持完整的生命周期管理、配置注入和钩子系统。

## 特性

### 核心功能
- **类型安全**: 基于泛型的类型安全依赖解析
- **命名服务**: 支持同一类型的多个命名实例
- **懒加载**: 按需创建实例，减少启动时间
- **立即加载**: 注册时立即创建实例
- **瞬态模式**: 每次获取都创建新实例

### 高级功能
- **配置注入**: 自动将配置注入到结构体字段
- **生命周期管理**: 支持 ServiceLifecycle 和 DestroyCallback 接口
- **钩子系统**: 完整的创建/销毁前后钩子，容器级启动/关闭钩子
- **循环依赖检测**: 自动检测并报告循环依赖
- **作用域隔离**: 支持父容器/子容器隔离
- **并发安全**: 线程安全的容器操作
- **性能监控**: 内置性能统计和指标
- **全局容器**: 内置单例容器和泛型便捷 API
- **服务包**: 支持将多个服务组合批量注册

## 安装

```bash
go get github.com/mzzsfy/go-util/di
```

## 快速开始

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

    // 注册命名服务（泛型便捷 API）
    di.ProvideNamed(container, "main-db", func(c di.Container) (*Database, error) {
        return &Database{Host: "localhost", Port: 5432}, nil
    })

    // 注册默认服务（类型推导，无需手写名称）
    di.Provide(container, func(c di.Container) (*AppService, error) {
        return &AppService{}, nil
    })

    // 获取服务（类型安全，失败时 panic）
    app := di.MustGet[*AppService](container)
    fmt.Printf("App running with DB: %s:%d\n", app.DB.Host, app.DB.Port)
}
```

### 全局容器

di 包内置一个全局单例容器，适合简单场景。

```go
// 直接使用全局容器
container := di.GlobalContainer()

di.Provide(container, func(c di.Container) (*Service, error) {
    return &Service{Value: "global"}, nil
})

s := di.MustGet[*Service](container)
```

### 泛型便捷 API

所有泛型 API 通过类型参数 T 消除手动类型断言。

```go
// 注册默认服务（名称为空）
di.Provide[T](c, func(c di.Container) (T, error) { ... })

// 注册带名称的服务
di.ProvideNamed[T](c, "name", func(c di.Container) (T, error) { ... })

// 直接注册一个已存在的实例值
di.ProvideValue[T](c, instance)
di.ProvideValueNamed[T](c, "name", instance)

// 获取服务，返回 (T, error)
v, err := di.Get[T](c)
v, err := di.GetNamed[T](c, "name")

// 获取服务，失败时 panic
v := di.MustGet[T](c)
v := di.MustGetNamed[T](c, "name")

// 检查是否已注册
ok := di.Has[T](c)
ok := di.HasNamed[T](c, "name")

// 获取某个类型的所有命名实例
all, err := di.GetNamedAll[T](c)
```

### 服务包

将多个相关服务组合成一个包函数，便于批量注册和复用。

```go
// 定义服务包
dbPackage := di.Package(
    func(c di.Container) error {
        return di.Provide(c, func(c di.Container) (*Database, error) {
            return &Database{Host: "localhost"}, nil
        })
    },
    func(c di.Container) error {
        return di.ProvideNamed(c, "readonly", func(c di.Container) (*Database, error) {
            return &Database{Host: "readonly-host"}, nil
        })
    },
)

// 加载到容器
c := di.New()
if _, err := di.LoadPackages(c, dbPackage); err != nil {
    panic(err)
}
```

### 配置注入

```go
type Config struct {
    DatabaseURL string `di.config:"db.url:localhost:5432"`
    Debug       bool   `di.config:"debug:false"`
    Callback    string `di.config:"https://${callback.url:localhost:8080}"`
}

di.ProvideNamed(container, "config", func(c di.Container) (*Config, error) {
    return &Config{}, nil
})

// 自动从配置源读取并注入字段值
config := di.MustGetNamed[*Config](container, "config")
// config.DatabaseURL = "localhost:5432" (默认值)
// config.Debug = false (默认值)
// config.Callback = "https://localhost:8080" (默认值)
```

配置源通过 SetConfigSource 设置，默认使用内存 map 配置源。

```go
source := di.NewMapConfigSource()
source.Set("db.url", "prod-db:5432")
container.SetConfigSource(source)

// 字段值也可直接通过容器读取
v := container.Value("db.url")
```

### 钩子系统

钩子分为提供者级别（仅作用于单个服务）和容器级别（作用于所有服务）。

#### 提供者级别钩子

```go
di.ProvideNamed(container, "service", func(c di.Container) (*Service, error) {
    return &Service{Value: "original"}, nil
},
    di.WithBeforeCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        // 创建前: info.Instance 为 nil
        return nil, nil // 返回 nil 使用默认创建
    }),
    di.WithAfterCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        // 创建后: info.Instance 为已创建的实例
        // 返回值会替换原实例
        return info.Instance, nil
    }),
    di.WithBeforeDestroy(func(c di.Container, info di.EntryInfo) {
        // 销毁前回调
    }),
    di.WithAfterDestroy(func(c di.Container, info di.EntryInfo) {
        // 销毁后回调
    }),
)
```

#### 容器级别钩子

容器级别钩子在 New 时通过 ContainerOption 配置，对所有服务生效。

```go
container := di.New(
    di.WithContainerBeforeCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        fmt.Printf("Creating: %s\n", info.Name)
        return nil, nil
    }),
    di.WithContainerAfterCreate(func(c di.Container, info di.EntryInfo) (any, error) {
        fmt.Printf("Created: %s\n", info.Name)
        return info.Instance, nil
    }),
    di.WithContainerBeforeDestroy(func(c di.Container, info di.EntryInfo) {
        fmt.Printf("Destroying: %s\n", info.Name)
    }),
    di.WithContainerAfterDestroy(func(c di.Container, info di.EntryInfo) {
        fmt.Printf("Destroyed: %s\n", info.Name)
    }),
)
```

#### 启动与关闭钩子

容器提供独立的启动和关闭钩子链。

```go
container := di.New(
    // Start() 时执行
    di.WithContainerOnStart(func(c di.Container) error {
        fmt.Println("container starting")
        return nil
    }),
    // Start() 后执行
    di.WithContainerAfterStart(func(c di.Container) error {
        fmt.Println("container started")
        return nil
    }),
    // Shutdown() 时最先执行（插队到其他关闭钩子之前）
    di.WithContainerBeforeShutdown(func(ctx context.Context) error {
        fmt.Println("before shutdown")
        return nil
    }),
    // Shutdown() 时按注册顺序执行
    di.WithContainerShutdown(func(ctx context.Context) error {
        fmt.Println("on shutdown")
        return nil
    }),
)

if err := container.Start(); err != nil {
    panic(err)
}
```

#### EntryInfo

钩子函数接收 EntryInfo 参数，携带服务元数据。

```go
type EntryInfo struct {
    Name     string            // 服务名称
    Instance any               // 服务实例（beforeCreate 时可能为 nil）
    Ctx      ContainerContext  // 容器上下文（销毁钩子中携带父 context）
}
```

### 生命周期管理

服务可通过实现以下接口自动接入容器关闭流程。

#### ServiceLifecycle 接口

```go
type LifecycleService struct {
    Value string
}

// Shutdown 在容器关闭时被调用
func (s *LifecycleService) Shutdown(ctx context.Context) error {
    fmt.Println("shutting down lifecycle service")
    return nil
}

di.Provide(container, func(c di.Container) (*LifecycleService, error) {
    return &LifecycleService{Value: "test"}, nil
})

// 关闭容器会自动调用所有 ServiceLifecycle 的 Shutdown
container.Shutdown(context.Background())
```

#### DestroyCallback 接口

```go
type ManagedResource struct{}

// OnDestroyCallback 在实例销毁时被调用，先于 ServiceLifecycle
func (m *ManagedResource) OnDestroyCallback() error {
    fmt.Println("destroy callback")
    return nil
}
```

销毁执行顺序: beforeDestroy 钩子 -> DestroyCallback.OnDestroyCallback -> ServiceLifecycle.Shutdown -> afterDestroy 钩子。

### 信号监听关闭

ShutdownOnSignals 监听系统信号并自动关闭容器，默认监听 SIGTERM 和 Interrupt。

```go
container := di.New()

di.Provide(container, func(c di.Container) (*App, error) {
    return &App{}, nil
})

container.Start()

// 收到 SIGTERM/SIGINT 时自动调用 Shutdown，然后 os.Exit
container.ShutdownOnSignals()

// 或自定义监听的信号
container.ShutdownOnSignals(syscall.SIGHUP)

// Done 返回关闭通知通道，通道关闭表示容器已关闭
<-container.Done()
```

### 加载模式

```go
// 默认模式 - 第一次获取时创建，之后缓存复用
di.ProvideNamed(c, "default", provider, di.WithLoadMode(di.LoadModeDefault))

// 立即加载 - 注册时立即创建实例
di.ProvideNamed(c, "eager", provider, di.WithLoadMode(di.LoadModeImmediate))

// 懒加载 - 延迟到第一次获取时创建，并检测循环依赖
di.ProvideNamed(c, "lazy", provider, di.WithLoadMode(di.LoadModeLazy))

// 瞬态模式 - 每次获取都创建新实例，类似工厂
di.ProvideNamed(c, "transient", provider, di.WithLoadMode(di.LoadModeTransient))
```

### 条件注册

```go
di.ProvideNamed(container, "conditional", func(c di.Container) (*Service, error) {
    return &Service{}, nil
}, di.WithCondition(func(c di.Container) bool {
    return os.Getenv("ENV") == "production"
}))
```

### 作用域隔离

```go
parent := di.New()
di.ProvideNamed(parent, "parent-service", func(c di.Container) (*Service, error) {
    return &Service{Value: "parent"}, nil
})

// 创建子容器，继承父容器的配置源但服务隔离
child := parent.CreateChildScope()
di.ProvideNamed(child, "child-service", func(c di.Container) (*Service, error) {
    return &Service{Value: "child"}, nil
})

// 子容器可以访问父容器服务
parentService, _ := di.GetNamed[*Service](child, "parent-service")

// 父容器不能访问子容器服务（返回错误）
_, err := di.GetNamed[*Service](parent, "child-service")

// 父容器关闭时会级联关闭子容器
parent.Shutdown(context.Background())
```

### 实例管理

容器支持运行时替换、移除和清空实例，用于测试或热更新。

```go
// 运行时替换实例（必须已注册提供者）
container.ReplaceInstance(reflect.TypeOf((*Service)(nil)).Elem(), "", newService)

// 移除已缓存实例（不影响提供者注册）
container.RemoveInstance(reflect.TypeOf((*Service)(nil)).Elem(), "")

// 清空所有缓存实例
container.ClearInstances()

// 查询实例和提供者数量
container.GetInstanceCount()
container.GetProviderCount()

// 获取所有已缓存实例和已注册提供者
container.GetAllInstances()
container.GetProviders()
```

### 性能监控

```go
stats := container.GetStats()
fmt.Printf("创建实例数: %d\n", stats.CreatedInstances)
fmt.Printf("Get 调用次数: %d\n", stats.GetCalls)
fmt.Printf("Provide 调用次数: %d\n", stats.ProvideCalls)
fmt.Printf("配置命中/未命中: %d / %d\n", stats.ConfigHits, stats.ConfigMisses)
fmt.Printf("总创建耗时: %v\n", stats.CreateDuration)

// 平均创建耗时
avg := container.GetAverageCreateDuration()

// 重置统计
container.ResetStats()
```

## API 参考

### 构造函数

```go
// 创建容器，可携带容器级选项
func New(opts ...ContainerOption) Container

// 获取全局单例容器
func GlobalContainer() Container
```

### Container 接口

```go
type Container interface {
    // 服务注册
    ProvideNamedWith(name string, provider any, opts ...ProviderOption) error

    // 服务获取
    GetNamed(serviceType any, name string) (any, error)
    GetNamedAll(serviceType any) (map[string]any, error)
    HasNamed(serviceType any, name string) bool

    // 容器选项
    AppendOption(opt ...ContainerOption) error

    // 生命周期
    Start() error
    Shutdown(context.Context) error
    ShutdownOnSignals(signals ...os.Signal)
    Done() <-chan struct{}
    CreateChildScope() Container

    // 配置
    SetConfigSource(source ConfigSource)
    GetConfigSource() ConfigSource
    Value(key string) config.Value

    // 统计
    GetStats() ContainerStats
    ResetStats()
    GetInstanceCount() int
    GetProviderCount() int
    GetAverageCreateDuration() time.Duration
    GetAllInstances() map[string]any
    GetProviders() map[string]string
    ClearInstances()

    // 实例管理
    ReplaceInstance(serviceType any, name string, newInstance any) error
    RemoveInstance(serviceType any, name string) error
}
```

### 泛型便捷函数

| 函数 | 说明 |
| --- | --- |
| `Provide[T](c, provider, opts...)` | 注册默认服务 |
| `ProvideNamed[T](c, name, provider, opts...)` | 注册命名服务 |
| `ProvideValue[T](c, instance, opts...)` | 注册实例值 |
| `ProvideValueNamed[T](c, name, instance, opts...)` | 注册命名实例值 |
| `Get[T](c)` | 获取默认服务 |
| `GetNamed[T](c, name)` | 获取命名服务 |
| `MustGet[T](c)` | 获取默认服务，失败 panic |
| `MustGetNamed[T](c, name)` | 获取命名服务，失败 panic |
| `Has[T](c)` | 检查默认服务是否已注册 |
| `HasNamed[T](c, name)` | 检查命名服务是否已注册 |
| `GetNamedAll[T](c)` | 获取某类型的所有命名实例 |

### 服务包函数

| 函数 | 说明 |
| --- | --- |
| `Package(providers...)` | 将多个 provider 组合成一个包函数 |
| `LoadPackages(c, packages...)` | 将多个包加载到容器 |

### 生命周期接口

| 接口 | 方法 | 说明 |
| --- | --- | --- |
| `ServiceLifecycle` | `Shutdown(ctx) error` | 容器关闭时调用 |
| `DestroyCallback` | `OnDestroyCallback() error` | 实例销毁时调用，先于 ServiceLifecycle |

### 钩子函数签名

```go
// 创建前/后钩子（提供者级和容器级）
func(Container, EntryInfo) (any, error)

// 销毁前/后钩子（提供者级和容器级）
func(Container, EntryInfo)

// 启动钩子
func(Container) error

// 关闭钩子
func(context.Context) error
```

### 选项函数

**提供者选项**（ProviderOption）:
- `WithBeforeCreate` - 创建前钩子
- `WithAfterCreate` - 创建后钩子
- `WithBeforeDestroy` - 销毁前钩子
- `WithAfterDestroy` - 销毁后钩子
- `WithLoadMode` - 加载模式
- `WithCondition` - 条件函数

**容器选项**（ContainerOption）:
- `WithContainerBeforeCreate` - 容器级创建前钩子
- `WithContainerAfterCreate` - 容器级创建后钩子
- `WithContainerBeforeDestroy` - 容器级销毁前钩子
- `WithContainerAfterDestroy` - 容器级销毁后钩子
- `WithContainerOnStart` - 启动时钩子
- `WithContainerAfterStart` - 启动后钩子
- `WithContainerBeforeShutdown` - 关闭前钩子（插队到链首）
- `WithContainerShutdown` - 关闭钩子

> 注: 除 `WithContainerShutdown` 和 `WithContainerBeforeShutdown` 外，上述容器选项在 `Start` 之后调用会 panic。`AppendOption` 会将 panic 转为错误返回。

### 加载模式

```go
const (
    LoadModeDefault   LoadMode = iota // 默认: 第一次 Get 时创建并缓存
    LoadModeImmediate                 // 立即加载: 注册时创建
    LoadModeLazy                      // 懒加载: 延迟创建并检测循环依赖
    LoadModeTransient                 // 瞬态: 每次获取都创建新实例
)
```

### 配置源

```go
// 内置基于 map 的可修改配置源
source := di.NewMapConfigSource()
source.Set("key", "value")
source.Has("key")  // true
source.Clear()

// 自定义配置源只需实现 ConfigSource 接口
type ConfigSource interface {
    Get(key string) config.Value
}
```

## 使用场景

### 1. Web 应用

```go
c := di.New()
di.ProvideNamed(c, "http-server", func(c di.Container) (*http.Server, error) {
    return &http.Server{Addr: ":8080"}, nil
})
di.ProvideNamed(c, "router", func(c di.Container) (*mux.Router, error) {
    return mux.NewRouter(), nil
})
```

### 2. 数据库连接

```go
di.ProvideNamed(c, "db", func(c di.Container) (*sql.DB, error) {
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

di.Provide(c, func(c di.Container) (*AppConfig, error) {
    return &AppConfig{}, nil
})
```

### 4. 缓存服务

```go
di.ProvideNamed(c, "redis", func(c di.Container) (*redis.Client, error) {
    return redis.NewClient(&redis.Options{Addr: "localhost:6379"}), nil
}, di.WithLoadMode(di.LoadModeImmediate))
```

## 性能

- **注册 1000 个服务**: ~1ms
- **首次获取 1000 个服务**: ~3ms
- **缓存获取 1000 个服务**: <1ms
- **并发安全**: 5000 请求/100 goroutines ~1ms
- **QPS**: 400万+

## 线程安全

所有容器操作都是线程安全的，可以在多个 goroutine 中安全使用。

## 限制

1. **基本类型限制**: `string`、`int`、`context.Context` 等基本类型不能注册为空名称服务
2. **结构体要求**: 依赖注入要求目标是结构体或结构体指针
3. **循环依赖**: 会检测并返回错误，不支持循环依赖
4. **启动后追加选项**: 容器 Start 后追加 ContainerOption 会 panic（仅 WithContainerShutdown/WithContainerBeforeShutdown 例外）

## 最佳实践

1. **优先使用泛型 API**: 用 `di.Provide`/`di.MustGet` 替代裸 `ProvideNamedWith`/`GetNamed`，避免手动类型断言
2. **合理使用钩子**: 避免在钩子中执行耗时操作
3. **及时清理**: 通过 ServiceLifecycle 或 WithAfterDestroy 释放资源
4. **错误处理**: 使用 Get 系列函数时检查返回错误，或使用 MustGet 系列在初始化阶段快速失败
5. **配置注入**: 使用 `di.config` 标签简化配置管理
6. **服务包**: 将相关服务组合成 Package，便于模块化和复用
