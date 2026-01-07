package di

import (
    "context"
    "os"
    "time"

    "github.com/mzzsfy/go-util/config"
    "github.com/mzzsfy/go-util/helper"
)

// Container 是依赖注入容器的接口
type Container interface {
    // ProvideNamedWith 注册带名称和选项的服务构造函数
    ProvideNamedWith(name string, provider any, opts ...ProviderOption) error

    // GetNamed 获取带名称的服务实例,name为空时时使用默认名称
    GetNamed(serviceType any, name string) (any, error)
    // GetNamedAll 找到所有匹配类型的服务实例，包括不同名称的实例,支持按接口匹配,性能较低，请勿滥用
    GetNamedAll(serviceType any) (map[string]any, error)

    // HasNamed 检查带名称的服务是否已注册
    HasNamed(serviceType any, name string) bool
    // AppendOption 添加容器选项,启动后不可使用
    AppendOption(opt ...ContainerOption) error
    // Start 启动容器，调用启动钩子
    Start() error

    // Shutdown 关闭容器，调用所有服务的 Shutdown 方法
    Shutdown(context.Context) error
    // ShutdownOnSignals 监听信号并自动关闭
    ShutdownOnSignals(signals ...os.Signal)
    Done() <-chan struct{}

    // CreateChildScope 创建子容器（作用域）
    CreateChildScope() Container

    SetConfigSource(source ConfigSource)
    Value(key string) config.Value
    GetConfigSource() ConfigSource

    GetStats() ContainerStats
    ResetStats()
    GetInstanceCount() int
    GetProviderCount() int
    GetAverageCreateDuration() time.Duration
    GetAllInstances() map[string]any
    GetProviders() map[string]string
    ClearInstances()

    ReplaceInstance(serviceType any, name string, newInstance any) error
    RemoveInstance(serviceType any, name string) error
}

// ServiceLifecycle 服务生命周期接口
type ServiceLifecycle interface {
    Shutdown(context.Context) error
}

type DestroyCallback interface {
    OnDestroyCallback() error
}

// ShutdownHook 服务关闭钩子
type ShutdownHook func(context.Context) error

// LoadMode 加载模式
type LoadMode int

const (
    LoadModeDefault   LoadMode = iota // 合适时机加载（第一次 Get 时创建，之后缓存）
    LoadModeImmediate                 // 立刻加载（注册时立即创建实例）
    LoadModeLazy                      // 懒加载：延迟到第一次获取时创建实例，并检测循环依赖
    LoadModeTransient                 // 每次创建,类似工厂模式。
)

const (
    ErrorConditionFail = helper.StringError("Condition failed")
)

// ConfigSource 配置源接口
type ConfigSource interface {
    // Get 获取配置值
    Get(key string) config.Value
}
type ConfigModifySource interface {
    ConfigSource
    // Set 设置配置值
    Set(key string, value any)
    // Has 检查配置是否存在
    Has(key string) bool
    // Clear 清空所有配置
    Clear()
}

// ContainerStats 容器统计信息
type ContainerStats struct {
    CreatedInstances int           // 创建的实例总数
    GetCalls         int           // Get调用次数
    ProvideCalls     int           // Provide调用次数
    ConfigHits       int           // 配置命中次数
    ConfigMisses     int           // 配置未命中次数
    CreateDuration   time.Duration // 总创建耗时
}
