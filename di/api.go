package di

import (
    "reflect"
    "time"

    "github.com/mzzsfy/go-util/config"
)

// 全局容器实例
var globalContainer = New()

// GlobalContainer 获取全局容器
func GlobalContainer() Container {
    return globalContainer
}

// GetNamedAll 获取指定类型的所有命名实例（泛型版本）
func GetNamedAll[T any](c Container) (map[string]T, error) {
    results, err := c.GetNamedAll(reflect.TypeOf((*T)(nil)).Elem())
    if err != nil {
        return nil, err
    }
    typedResults := make(map[string]T, len(results))
    for k, v := range results {
        typedResults[k] = v.(T)
    }
    return typedResults, nil
}

// ========== 泛型包装函数 ==========

// Provide 注册服务构造函数
func Provide[T any](c Container, provider func(Container) (T, error), opts ...ProviderOption) error {
    return ProvideNamed(c, "", provider, opts...)
}

// ProvideNamed 注册带名称的服务构造函数
func ProvideNamed[T any](c Container, name string, provider func(Container) (T, error), opts ...ProviderOption) error {
    return c.ProvideNamedWith(name, provider, opts...)
}

// ProvideValue 注册实例值
func ProvideValue[T any](c Container, instance T, opts ...ProviderOption) error {
    return ProvideNamed(c, "", func(Container) (T, error) { return instance, nil }, opts...)
}

// ProvideValueNamed 注册带名称的实例值
func ProvideValueNamed[T any](c Container, name string, instance T, opts ...ProviderOption) error {
    return ProvideNamed(c, name, func(Container) (T, error) { return instance, nil }, opts...)
}

// MustGet 获取服务实例，失败时 panic
func MustGet[T any](c Container) T {
    return MustGetNamed[T](c, "")
}

// MustGetNamed 获取带名称的服务实例，失败时 panic
func MustGetNamed[T any](c Container, name string) T {
    t, err := GetNamed[T](c, name)
    if err != nil {
        panic(err)
    }
    return t
}

// Get 获取服务实例
func Get[T any](c Container) (T, error) {
    return GetNamed[T](c, "")
}

// GetNamed 获取带名称的服务实例
func GetNamed[T any](c Container, name string) (T, error) {
    result, err := c.GetNamed(reflect.TypeOf((*T)(nil)).Elem(), name)
    if err != nil {
        var zero T
        return zero, err
    }
    return result.(T), nil
}

// Has 检查服务是否已注册
func Has[T any](c Container) bool {
    return c.HasNamed(reflect.TypeOf((*T)(nil)).Elem(), "")
}

// HasNamed 检查带名称的服务是否已注册
func HasNamed[T any](c Container, name string) bool {
    return c.HasNamed(reflect.TypeOf((*T)(nil)).Elem(), name)
}

// GetProviders 获取所有注册的提供者信息（使用全局容器）
func GetProviders() map[string]string {
    return globalContainer.GetProviders()
}

// GetStats 获取容器统计信息（使用全局容器）
func GetStats() ContainerStats {
    return globalContainer.GetStats()
}

// GetProviderCount 获取注册的提供者数量（使用全局容器）
func GetProviderCount() int {
    return globalContainer.GetProviderCount()
}

// GetInstanceCount 获取当前缓存的实例数量（使用全局容器）
func GetInstanceCount() int {
    return globalContainer.GetInstanceCount()
}

// GetAverageCreateDuration 获取平均创建耗时（使用全局容器）
func GetAverageCreateDuration() time.Duration {
    return globalContainer.GetAverageCreateDuration()
}

// ResetStats 重置统计信息（使用全局容器）
func ResetStats() {
    globalContainer.ResetStats()
}

// Value 获取配置值（使用全局容器）
func Value(key string) config.Value {
    return globalContainer.Value(key)
}
