// Package di 提供服务注册功能
package di

import (
	"fmt"
	"reflect"
)

// containerType 缓存 Container 类型
// 用于验证 provider 函数签名
var containerType = reflect.TypeOf((*Container)(nil)).Elem()

// ProvideNamedWith 注册带名称和选项的服务构造函数
// provider 必须是签名为 func(Container) (T, error) 的函数
func (c *container) ProvideNamedWith(name string, provider any, opts ...ProviderOption) error {
	// 验证 provider 是正确的函数类型
	providerType := reflect.TypeOf(provider)
	if err := c.validateProviderType(providerType); err != nil {
		return err
	}

	return c.registerProvider(name, providerType, provider, opts)
}

// validateProviderType 验证 provider 函数类型
// 检查函数签名是否符合要求
func (c *container) validateProviderType(providerType reflect.Type) error {
	return validateProviderFunction(providerType)
}

// validateProviderFunction 验证 provider 函数签名
// 独立函数，供测试和外部使用
func validateProviderFunction(providerType reflect.Type) error {
	if providerType.Kind() != reflect.Func {
		return fmt.Errorf("provider must be a function")
	}
	if providerType.NumIn() != 1 || providerType.In(0) != containerType {
		return fmt.Errorf("provider must have signature func(Container) (T, error)")
	}
	if providerType.NumOut() != 2 || providerType.Out(1).String() != "error" {
		return fmt.Errorf("provider must return (T, error)")
	}
	return nil
}

// checkBlacklistForProvider 检查黑名单类型
// 黑名单类型必须有名称才能注册
func checkBlacklistForProvider(returnType reflect.Type, name string) error {
	if name != "" {
		return nil
	}
	if blackTypeMap[returnType.String()] {
		return fmt.Errorf("cannot register type %s without name", returnType)
	}
	return nil
}

// registerProvider 注册提供者
// 验证黑名单后执行实际注册
func (c *container) registerProvider(name string, providerType reflect.Type, provider any, opts []ProviderOption) error {
	returnType := providerType.Out(0)
	key := typeKey(returnType, name)

	if err := checkBlacklistForProvider(returnType, name); err != nil {
		return err
	}

	return c.doRegisterProvider(key, returnType, provider, opts)
}

// doRegisterProvider 执行提供者注册
// 检查重复注册，应用选项，存储提供者
func (c *container) doRegisterProvider(key string, returnType reflect.Type, provider any, opts []ProviderOption) error {
	c.mu.Lock()

	if _, exists := c.providers[key]; exists {
		c.mu.Unlock()
		return providerExistsError(returnType, extractNameFromKey(key))
	}

	// 应用选项，默认为 LoadModeDefault
	p := providerConfig{}
	for _, opt := range opts {
		opt(&p)
	}

	c.providers[key] = providerEntry{
		reflectType: returnType,
		provider:    createProviderWrapper(provider),
		config:      p,
	}

	c.statsMu.Lock()
	c.stats.provideCalls++
	c.statsMu.Unlock()
	c.mu.Unlock()

	// 如果是立即加载模式，额外调用一次,保证创建实例
	return c.handleImmediateLoad(p.loadMode, returnType, key)
}

// createProviderWrapper 创建提供者包装函数
// 使用反射调用 provider 函数，保证泛型兼容
func createProviderWrapper(provider any) func(Container) (any, error) {
	return func(cont Container) (any, error) {
		results := reflect.ValueOf(provider).Call([]reflect.Value{reflect.ValueOf(cont)})
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	}
}

// handleImmediateLoad 处理立即加载模式
// 立即加载模式下立即创建实例
func (c *container) handleImmediateLoad(loadMode LoadMode, returnType reflect.Type, key string) error {
	if loadMode == LoadModeImmediate {
		_, err := c.GetNamed(returnType, extractNameFromKey(key))
		if err != nil {
			return err
		}
	}
	return nil
}
