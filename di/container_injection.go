// Package di 提供依赖注入功能
package di

import (
	"fmt"
	"reflect"
	"strings"
)

// injectService 注入服务到字段
// 根据 di 标签获取服务名称并注入对应的服务实例
// 参数:
//   - fieldValue: 字段反射值，用于设置注入的服务
//   - fieldType: 字段类型信息，包含 di 标签
// 返回:
//   - 注入失败时返回错误
func (c *container) injectService(fieldValue reflect.Value, fieldType reflect.StructField) error {
	serviceName := fieldType.Tag.Get("di")
	serviceInstance, err := c.GetNamed(fieldType.Type, serviceName)
	if err != nil {
		return fieldInjectionError(fieldType.Name, err)
	}

	if err := setFieldValue(fieldValue, serviceInstance); err != nil {
		return fieldInjectionError(fieldType.Name, err)
	}

	return nil
}

// resolveConfigTag 解析配置标签值
// 支持变量替换和传统格式两种方式
func (c *container) resolveConfigTag(configTag string) any {
	if strings.Contains(configTag, "${") {
		return c.resolveConfigValue(configTag)
	}
	return c.resolveTraditionalConfig(configTag)
}

// resolveTraditionalConfig 解析传统配置格式
// 格式为 "key:defaultValue"
func (c *container) resolveTraditionalConfig(configTag string) any {
	configKey, defaultValue := parseConfigInjection(configTag)
	return c.getConfigValue(configKey).AnyD(defaultValue)
}

// injectConfig 注入配置到字段
// 根据 di.config 标签获取配置值并注入
// 参数:
//   - fieldValue: 字段反射值，用于设置配置值
//   - fieldType: 字段类型信息，包含 di.config 标签
// 返回:
//   - 注入失败时返回错误
func (c *container) injectConfig(fieldValue reflect.Value, fieldType reflect.StructField) error {
	configTag := fieldType.Tag.Get("di.config")
	actualValue := c.resolveConfigTag(configTag)

	if err := setFieldValue(fieldValue, actualValue); err != nil {
		return fieldInjectionError(fieldType.Name, err)
	}

	return nil
}

// processFieldInjection 处理单个字段的注入
// 根据标签决定注入类型
func (c *container) processFieldInjection(fieldType reflect.StructField, fieldValue reflect.Value) error {
	_, hasDiTag := fieldType.Tag.Lookup("di")
	_, hasConfigTag := fieldType.Tag.Lookup("di.config")

	switch {
	case hasDiTag:
		return c.injectService(fieldValue, fieldType)
	case hasConfigTag:
		return c.injectConfig(fieldValue, fieldType)
	default:
		return nil
	}
}

// injectStruct 注入配置和服务到结构体字段
// 解引用指针后遍历所有字段进行注入
func (c *container) injectStruct(target reflect.Value) error {
	target = c.dereferencePointer(target)
	if !target.IsValid() {
		return nil
	}

	if target.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type for injection, got %v", target.Kind())
	}

	return c.injectAllFields(target)
}

// dereferencePointer 解引用指针
// 如果是空指针返回无效的 Value
// 参数:
//   - target: 目标反射值
// 返回:
//   - 解引用后的值，如果是空指针则返回无效 Value
func (c *container) dereferencePointer(target reflect.Value) reflect.Value {
	if target.Kind() == reflect.Ptr && target.IsNil() {
		return reflect.Value{}
	}
	if target.Kind() == reflect.Ptr {
		return target.Elem()
	}
	return target
}

// injectAllFields 注入所有字段
// 遍历结构体字段并执行注入
func (c *container) injectAllFields(target reflect.Value) error {
	targetType := target.Type()
	for i := 0; i < target.NumField(); i++ {
		if err := c.processFieldInjection(targetType.Field(i), target.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

// findDepend 找到所有待注入的字段
// 递归处理指针和结构体类型
// 参数:
//   - t: 要分析的类型
// 返回:
//   - 依赖键列表
//   - 如果类型不是结构体或指针返回错误
func (c *container) findDepend(t reflect.Type) ([]string, error) {
	switch t.Kind() {
	case reflect.Pointer:
		return c.findDepend(t.Elem())
	case reflect.Struct:
		return c.findStructDependencies(t)
	default:
		return nil, fmt.Errorf("provider must be a struct")
	}
}

// findStructDependencies 查找结构体的所有依赖
// 返回依赖的键列表
func (c *container) findStructDependencies(t reflect.Type) ([]string, error) {
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		depKey, err := c.getFieldDependency(field)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze dependency for field %s in struct %v: %w", field.Name, t, err)
		}
		if depKey != "" {
			fields = append(fields, depKey)
		}
	}
	return fields, nil
}

// getFieldDependency 获取单个字段的依赖键
// 根据标签解析依赖
func (c *container) getFieldDependency(field reflect.StructField) (string, error) {
	tag, hasTag := field.Tag.Lookup("di")
	if !hasTag {
		return "", nil
	}

	name := tag
	serviceType := field.Type
	typeName := typeKey(serviceType, name)

	if _, ok := c.providers[typeName]; ok {
		return typeName, nil
	}
	return "", fmt.Errorf("no provider found for type %s with name '%s'", serviceType, name)
}

// injectToInstance 根据实例类型进行依赖注入
// 处理结构体和指针类型
func (c *container) injectToInstance(instance any) (any, error) {
	instanceValue := reflect.ValueOf(instance)
	kind := instanceValue.Kind()

	// 不可寻址的结构体需要特殊处理
	if kind == reflect.Struct && !instanceValue.CanAddr() {
		return c.injectUnaddressableStruct(instanceValue)
	}

	// 非结构体和指针类型直接返回
	if kind != reflect.Struct && kind != reflect.Ptr {
		return instance, nil
	}

	if err := c.injectStruct(instanceValue); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies to instance of type %T: %w", instance, err)
	}
	return instance, nil
}

// injectUnaddressableStruct 注入到不可寻址的结构体
func (c *container) injectUnaddressableStruct(instanceValue reflect.Value) (any, error) {
	addr := reflect.New(instanceValue.Type())
	addr.Elem().Set(instanceValue)
	if err := c.injectStruct(addr); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies to unaddressable struct of type %v: %w", instanceValue.Type(), err)
	}
	return addr.Elem().Interface(), nil
}

// validateAndInject 验证实例并执行依赖注入
func (c *container) validateAndInject(instance any) (any, error) {
	if instance == nil {
		return nil, nil
	}
	instanceValue := reflect.ValueOf(instance)
	if !c.validateInstance(instance, instanceValue) {
		return nil, nil
	}
	return c.injectToInstance(instance)
}
