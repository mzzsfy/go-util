// Package di 提供依赖注入容器功能
// 支持懒加载、生命周期管理、配置注入等高级特性
package di

import (
	"context"
	"os"
	"reflect"
	"time"

	"github.com/mzzsfy/go-util/config"
	"github.com/mzzsfy/go-util/helper"
)

// blackTypeMap 黑名单类型映射
// 这些类型不能在没有名称的情况下注册到容器中
// 防止意外覆盖常用基础类型
var blackTypeMap = map[string]bool{
	"context.Context": true,
	"string":          true,
	"int":             true,
}

// isBlacklistType 检查类型是否在黑名单中
// 参数 t: 要检查的反射类型
// 返回值: 如果类型在黑名单中返回 true
func isBlacklistType(t reflect.Type) bool {
	return blackTypeMap[t.String()]
}

// Container 是依赖注入容器的核心接口
// 提供服务注册、获取、生命周期管理等完整功能
type Container interface {
	// ProvideNamedWith 注册带名称和选项的服务构造函数
	// name: 服务名称，可以为空表示默认服务
	// provider: 服务构造函数，签名为 func(Container) (T, error)
	// opts: 可选的提供者配置选项
	ProvideNamedWith(name string, provider any, opts ...ProviderOption) error

	// GetNamed 获取带名称的服务实例
	// serviceType: 服务类型，可以是 reflect.Type 或具体类型
	// name: 服务名称，为空时使用默认名称
	// 返回服务实例和可能的错误
	GetNamed(serviceType any, name string) (any, error)

	// GetNamedAll 获取所有匹配类型的服务实例
	// 支持按接口匹配，性能较低，请勿滥用
	// 返回键值对映射，键为类型#名称格式
	GetNamedAll(serviceType any) (map[string]any, error)

	// HasNamed 检查带名称的服务是否已注册
	HasNamed(serviceType any, name string) bool

	// AppendOption 添加容器选项，启动后不可使用
	AppendOption(opt ...ContainerOption) error

	// Start 启动容器，调用启动钩子
	Start() error

	// Shutdown 关闭容器，调用所有服务的 Shutdown 方法
	Shutdown(context.Context) error

	// ShutdownOnSignals 监听系统信号并自动关闭
	ShutdownOnSignals(signals ...os.Signal)

	// Done 返回关闭通知通道
	Done() <-chan struct{}

	// CreateChildScope 创建子容器（作用域）
	// 子容器可以访问父容器的服务
	CreateChildScope() Container

	// 配置相关方法
	SetConfigSource(source ConfigSource)
	Value(key string) config.Value
	GetConfigSource() ConfigSource

	// 统计相关方法
	GetStats() ContainerStats
	ResetStats()
	GetInstanceCount() int
	GetProviderCount() int
	GetAverageCreateDuration() time.Duration
	GetAllInstances() map[string]any
	GetProviders() map[string]string
	ClearInstances()

	// 实例管理方法
	ReplaceInstance(serviceType any, name string, newInstance any) error
	RemoveInstance(serviceType any, name string) error
}

// ServiceLifecycle 服务生命周期接口
// 实现 Shutdown 方法的服务会在容器关闭时被调用
type ServiceLifecycle interface {
	// Shutdown 关闭服务，释放资源
	// ctx: 上下文，用于控制关闭超时
	Shutdown(context.Context) error
}

// DestroyCallback 销毁回调接口
// 实现 OnDestroyCallback 方法的服务会在实例销毁时被调用
type DestroyCallback interface {
	// OnDestroyCallback 执行销毁时的回调逻辑
	OnDestroyCallback() error
}

// ShutdownHook 服务关闭钩子函数类型
// 在容器关闭时按注册顺序执行
type ShutdownHook func(context.Context) error

// LoadMode 定义服务加载模式
type LoadMode int

const (
	// LoadModeDefault 默认加载模式
	// 第一次 Get 时创建实例，之后缓存复用
	LoadModeDefault LoadMode = iota

	// LoadModeImmediate 立即加载模式
	// 注册时立即创建实例
	LoadModeImmediate

	// LoadModeLazy 懒加载模式
	// 延迟到第一次获取时创建实例，并检测循环依赖
	LoadModeLazy

	// LoadModeTransient 瞬态模式
	// 每次获取都创建新实例，类似工厂模式
	LoadModeTransient
)

// ErrorConditionFail 条件失败错误
// 当 WithCondition 条件不满足时返回此错误
const ErrorConditionFail = helper.StringError("Condition failed")

// ConfigSource 配置源接口
// 提供配置值的获取功能
type ConfigSource interface {
	// Get 根据 key 获取配置值
	Get(key string) config.Value
}

// ConfigModifySource 可修改的配置源接口
// 扩展 ConfigSource，支持设置和删除配置
type ConfigModifySource interface {
	ConfigSource
	// Set 设置配置值
	Set(key string, value any)
	// Has 检查配置是否存在
	Has(key string) bool
	// Clear 清空所有配置
	Clear()
}

// ContainerStats 容器运行统计信息
// 用于监控容器性能和服务状态
type ContainerStats struct {
	// CreatedInstances 创建的实例总数
	CreatedInstances int
	// GetCalls Get 调用次数
	GetCalls int
	// ProvideCalls Provide 调用次数
	ProvideCalls int
	// ConfigHits 配置命中次数
	ConfigHits int
	// ConfigMisses 配置未命中次数
	ConfigMisses int
	// CreateDuration 总创建耗时
	CreateDuration time.Duration
}
