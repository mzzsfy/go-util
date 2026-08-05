// Package di 实现依赖注入容器
package di

import (
	"reflect"
	"sync"
)

// cacheKey 实例缓存键,值类型避免字符串拼接
type cacheKey struct {
	t    reflect.Type
	name string
}

// String 返回缓存键的字符串表示,供管理接口使用
func (k cacheKey) String() string {
	if k.name != "" {
		return k.t.String() + "#" + k.name
	}
	return k.t.String()
}

// container 实现 Container 接口
type container struct {
	// beforeCreate 容器级别的创建前钩子
	beforeCreate []func(Container, EntryInfo) (any, error)
	// afterCreate 容器级别的创建后钩子
	afterCreate []func(Container, EntryInfo) (any, error)
	// beforeDestroy 容器级别的销毁前钩子
	beforeDestroy []func(Container, EntryInfo)
	// afterDestroy 容器级别的销毁后钩子
	afterDestroy []func(Container, EntryInfo)
	// providers 服务提供者映射
	providers map[cacheKey]providerEntry
	// instances 已创建的实例缓存
	instances map[cacheKey]any
	// mu 读写锁，保护 providers 和 instances
	mu sync.RWMutex
	// parent 父容器引用，用于作用域继承
	parent *container
	// shutdown 关闭钩子列表
	shutdown []ShutdownHook
	// loading 正在创建的实例标记，用于检测循环依赖
	loading map[cacheKey]bool
	// configSource 配置源
	configSource ConfigSource
	// configMu 配置源读写锁
	configMu sync.RWMutex
	// stats 容器统计信息（使用 atomic 操作）
	stats containerStats
	// done 关闭通知通道
	done chan struct{}
	// onStartup 启动前钩子列表
	onStartup []func(Container) error
	// afterStartup 启动后钩子列表
	afterStartup []func(Container) error
	// started 是否已启动
	started bool
}

// containerStats 容器运行统计信息
// 使用 int64 配合 atomic 操作实现无锁统计
type containerStats struct {
	// createdInstances 创建的实例总数
	createdInstances int64
	// getCalls Get 调用次数
	getCalls int64
	// provideCalls Provide 调用次数
	provideCalls int64
	// configHits 配置命中次数
	configHits int64
	// configMisses 配置未命中次数
	configMisses int64
	// createDuration 总创建耗时（纳秒）
	createDuration int64
}

// providerEntry 服务提供者条目
// 存储服务的类型、构造函数和配置信息
type providerEntry struct {
	// reflectType 服务的反射类型
	reflectType reflect.Type
	// provider 服务构造函数
	provider func(Container) (any, error)
	// config 提供者配置
	config providerConfig
}

// New 创建新的 DI 容器
// opts: 可选的容器配置选项
// 返回配置好的容器实例
func New(opts ...ContainerOption) Container {
    c := &container{
        providers:    make(map[cacheKey]providerEntry),
        instances:    make(map[cacheKey]any),
        loading:      make(map[cacheKey]bool),
        configSource: NewMapConfigSource(),
        done:         make(chan struct{}),
        onStartup:    make([]func(Container) error, 0),
        afterStartup: make([]func(Container) error, 0),
    }

    // 应用所有配置选项
    for _, opt := range opts {
        opt(c)
    }

    return c
}
