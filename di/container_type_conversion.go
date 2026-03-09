// Package di 提供类型转换功能
package di

import (
	"fmt"
	"reflect"
)

// tryDirectOrInterfaceMatch 尝试直接类型匹配或接口匹配
// 如果类型相同或值实现了接口，直接设置字段
func tryDirectOrInterfaceMatch(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) bool {
	if fieldType == valueType {
		field.Set(valueReflect)
		return true
	}

	if fieldType.Kind() != reflect.Interface {
		return false
	}

	return tryInterfaceMatch(field, valueReflect, fieldType, valueType)
}

// tryInterfaceMatch 尝试接口匹配
// 检查值是否实现了目标接口类型
func tryInterfaceMatch(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) bool {
	if valueReflect.Type().Implements(fieldType) {
		field.Set(valueReflect)
		return true
	}

	if valueType.Kind() == reflect.Ptr && valueReflect.Elem().Type().Implements(fieldType) {
		field.Set(valueReflect)
		return true
	}

	return false
}

// tryPointerConversion 尝试指针和值的转换
// 自动处理 *T 和 T 之间的转换
func tryPointerConversion(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) bool {
	fieldIsPtr := fieldType.Kind() == reflect.Ptr
	valueIsPtr := valueType.Kind() == reflect.Ptr

	if fieldIsPtr == valueIsPtr {
		return false
	}

	if fieldIsPtr {
		convertValueToPointer(field, valueReflect, valueType)
		return true
	}

	field.Set(valueReflect.Elem())
	return true
}

// convertValueToPointer 将值类型转换为指针类型
// 处理可寻址和不可寻址的情况
func convertValueToPointer(field reflect.Value, valueReflect reflect.Value, valueType reflect.Type) {
	if valueReflect.CanAddr() {
		field.Set(valueReflect.Addr())
		return
	}

	ptr := reflect.New(valueType)
	ptr.Elem().Set(valueReflect)
	field.Set(ptr)
}

// tryConvertibleOrSmartConversion 尝试类型转换或智能转换
// 先尝试标准类型转换，再尝试智能转换
func tryConvertibleOrSmartConversion(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) bool {
	if valueReflect.Type().ConvertibleTo(fieldType) {
		field.Set(valueReflect.Convert(fieldType))
		return true
	}

	if err := smartTypeConversion(field, valueReflect, fieldType, valueType); err == nil {
		return true
	}

	return false
}

// conversionStrategy 类型转换策略函数
// 定义类型转换的通用接口
type conversionStrategy func(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) bool

// conversionStrategies 类型转换策略链
// 按顺序尝试不同的转换策略
var conversionStrategies = []conversionStrategy{
	tryDirectOrInterfaceMatch,
	tryPointerConversion,
	tryConvertibleOrSmartConversion,
}

// setFieldValue 设置字段值
// 处理类型兼容性，自动进行必要的类型转换
func setFieldValue(field reflect.Value, value any) error {
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: field is not settable (private or constant)")
	}

	valueReflect := reflect.ValueOf(value)
	fieldType := field.Type()
	valueType := valueReflect.Type()

	for _, strategy := range conversionStrategies {
		if strategy(field, valueReflect, fieldType, valueType) {
			return nil
		}
	}

	return conversionError(valueType, fieldType)
}

// trueStrings 真值字符串映射
// 用于字符串到布尔值的转换
var trueStrings = map[string]bool{
	"true": true,
	"yes":  true,
	"1":    true,
	"on":   true,
}

// conversionFunc 类型转换函数签名
// 定义字符串到其他类型的转换函数接口
type conversionFunc func(reflect.Value, string) error

// convertStringToBool 字符串转布尔值
// 支持 true/yes/1/on 作为真值
func convertStringToBool(field reflect.Value, strValue string) error {
	field.SetBool(trueStrings[strValue])
	return nil
}

// convertStringToNumber 字符串转数值
// 统一处理整数、无符号整数和浮点数的转换
func convertStringToNumber(field reflect.Value, strValue string, kind reflect.Kind) error {
	switch {
	case kind == reflect.Float32 || kind == reflect.Float64:
		var floatValue float64
		if _, err := fmt.Sscanf(strValue, "%f", &floatValue); err != nil {
			return fmt.Errorf("cannot convert string '%s' to float: %w", strValue, err)
		}
		field.SetFloat(floatValue)
	case kind >= reflect.Int && kind <= reflect.Int64:
		var intValue int64
		if _, err := fmt.Sscanf(strValue, "%d", &intValue); err != nil {
			return fmt.Errorf("cannot convert string '%s' to integer: %w", strValue, err)
		}
		field.SetInt(intValue)
	default:
		var uintValue uint64
		if _, err := fmt.Sscanf(strValue, "%d", &uintValue); err != nil {
			return fmt.Errorf("cannot convert string '%s' to uint: %w", strValue, err)
		}
		field.SetUint(uintValue)
	}
	return nil
}

// convertStringToInt 字符串转整数（向后兼容）
func convertStringToInt(field reflect.Value, strValue string) error {
	return convertStringToNumber(field, strValue, reflect.Int)
}

// convertStringToUint 字符串转无符号整数（向后兼容）
func convertStringToUint(field reflect.Value, strValue string) error {
	return convertStringToNumber(field, strValue, reflect.Uint)
}

// convertStringToFloat 字符串转浮点数（向后兼容）
func convertStringToFloat(field reflect.Value, strValue string) error {
	return convertStringToNumber(field, strValue, reflect.Float64)
}

// convertStringToString 字符串到字符串的转换
// 直接设置字符串值
func convertStringToString(field reflect.Value, strValue string) error {
	field.SetString(strValue)
	return nil
}

// stringConverters 字符串转换器映射
// 根据 Kind 选择对应的转换函数
var stringConverters = map[reflect.Kind]conversionFunc{
	reflect.Bool:   convertStringToBool,
	reflect.String: convertStringToString,
}

// smartTypeConversion 智能类型转换
// 支持从字符串自动转换为各种基本类型
func smartTypeConversion(field reflect.Value, valueReflect reflect.Value, fieldType reflect.Type, valueType reflect.Type) error {
	if valueType.Kind() != reflect.String {
		return fmt.Errorf("smart conversion only supports string source type, got %s", valueType.Kind())
	}

	strValue := valueReflect.String()
	kind := fieldType.Kind()

	// 检查是否有专用转换器
	if converter, exists := stringConverters[kind]; exists {
		if err := converter(field, strValue); err != nil {
			return fmt.Errorf("smart conversion failed for field type %s: %w", kind, err)
		}
		return nil
	}

	// 处理数值类型
	if !isNumericKind(kind) {
		return fmt.Errorf("unsupported smart conversion: cannot convert string to %s", kind)
	}

	if err := convertStringToNumber(field, strValue, kind); err != nil {
		return fmt.Errorf("smart conversion failed for field type %s: %w", kind, err)
	}
	return nil
}

// isNumericKind 检查是否为数值类型
func isNumericKind(kind reflect.Kind) bool {
	return (kind >= reflect.Int && kind <= reflect.Int64) ||
		(kind >= reflect.Uint && kind <= reflect.Uint64) ||
		kind == reflect.Float32 || kind == reflect.Float64
}
