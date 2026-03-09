// Package di 提供泛型辅助函数和全局容器
package di

import (
	"reflect"
)

// globalContainer 全局容器实例
// 作为默认的容器使用
var globalContainer = New()

// GlobalContainer 获取全局容器
// 返回全局单例容器实例
func GlobalContainer() Container {
	return globalContainer
}

// GetNamedAll 获取指定类型的所有命名实例（泛型版本）
// 类型参数 T: 目标服务类型
// 参数:
//   - c: 容器实例
// 返回:
//   - 类型安全的实例映射
//   - 可能的错误
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
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - provider: 服务构造函数
//   - opts: 可选的提供者配置
// 返回:
//   - 可能的错误
func Provide[T any](c Container, provider func(Container) (T, error), opts ...ProviderOption) error {
	return ProvideNamed(c, "", provider, opts...)
}

// ProvideNamed 注册带名称的服务构造函数
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - name: 服务名称
//   - provider: 服务构造函数
//   - opts: 可选的提供者配置
// 返回:
//   - 可能的错误
func ProvideNamed[T any](c Container, name string, provider func(Container) (T, error), opts ...ProviderOption) error {
	return c.ProvideNamedWith(name, provider, opts...)
}

// ProvideValue 注册实例值
// 直接注册一个已创建的实例
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - instance: 服务实例
//   - opts: 可选的提供者配置
// 返回:
//   - 可能的错误
func ProvideValue[T any](c Container, instance T, opts ...ProviderOption) error {
	return ProvideNamed(c, "", func(Container) (T, error) { return instance, nil }, opts...)
}

// ProvideValueNamed 注册带名称的实例值
// 直接注册一个已创建的带名称实例
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - name: 服务名称
//   - instance: 服务实例
//   - opts: 可选的提供者配置
// 返回:
//   - 可能的错误
func ProvideValueNamed[T any](c Container, name string, instance T, opts ...ProviderOption) error {
	return ProvideNamed(c, name, func(Container) (T, error) { return instance, nil }, opts...)
}

// MustGet 获取服务实例，失败时 panic
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
// 返回:
//   - 服务实例
func MustGet[T any](c Container) T {
	return MustGetNamed[T](c, "")
}

// MustGetNamed 获取带名称的服务实例，失败时 panic
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - name: 服务名称
// 返回:
//   - 服务实例
func MustGetNamed[T any](c Container, name string) T {
	t, err := GetNamed[T](c, name)
	if err != nil {
		panic(err)
	}
	return t
}

// Get 获取服务实例
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
// 返回:
//   - 服务实例
//   - 可能的错误
func Get[T any](c Container) (T, error) {
	return GetNamed[T](c, "")
}

// GetNamed 获取带名称的服务实例
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - name: 服务名称
// 返回:
//   - 服务实例
//   - 可能的错误
func GetNamed[T any](c Container, name string) (T, error) {
	result, err := c.GetNamed(reflect.TypeOf((*T)(nil)).Elem(), name)
	if err != nil {
		var zero T
		return zero, err
	}
	return result.(T), nil
}

// Has 检查服务是否已注册
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
// 返回:
//   - 如果服务已注册返回 true
func Has[T any](c Container) bool {
	return c.HasNamed(reflect.TypeOf((*T)(nil)).Elem(), "")
}

// HasNamed 检查带名称的服务是否已注册
// 类型参数 T: 服务类型
// 参数:
//   - c: 容器实例
//   - name: 服务名称
// 返回:
//   - 如果服务已注册返回 true
func HasNamed[T any](c Container, name string) bool {
	return c.HasNamed(reflect.TypeOf((*T)(nil)).Elem(), name)
}
