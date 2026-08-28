package script

// ========== 外部函数调用桥接 ==========
// 本文件实现了脚本引擎与Go函数之间的互操作
// 包括：外部函数调用、返回值处理、类型转换

import (
	"fmt"
	"reflect"
)

// ========== 外部函数调用 ==========

// callExternalFunc 调用外部Go函数
// 参数fn为Go函数（必须是function类型），args为脚本值参数列表
// 返回转换后的脚本值和可能的错误
// 支持的Go函数签名：无返回值、单返回值、(value, error)双返回值
// 不缓存函数信息：闭包与方法值共享代码指针，按指针缓存会静默调用错误目标
func callExternalFunc(fn interface{}, args []Value) (Value, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// 验证函数类型
	if fnType.Kind() != reflect.Func {
		return Value{}, fmt.Errorf("参数必须是函数类型，但传入了%v类型", fnType.Kind())
	}

	// 验证参数数量
	if fnType.NumIn() != len(args) {
		return Value{}, fmt.Errorf("参数数量不匹配: 期望 %d, 实际 %d", fnType.NumIn(), len(args))
	}

	// 转换参数
	in, err := convertArgs(fnType, args)
	if err != nil {
		return Value{}, err
	}

	// 调用函数
	out := fnValue.Call(in)
	return handleReturnValues(out)
}

// wrapConvertError 包装转换错误，添加详细的参数信息
// 用于提供更详细的错误定位，方便调试
func wrapConvertError(index int, targetType reflect.Type, actualValue Value, err error) error {
	return fmt.Errorf(
		"参数 %d 转换失败: %v\n"+
			"→ 期望类型: %v\n"+
			"→ 实际类型: %s\n"+
			"→ 建议：检查函数调用时传入的参数类型是否正确",
		index, err, targetType, typeName(actualValue.Type))
}

// convertArgs 将脚本值参数列表转换为Go函数参数
// 遍历每个参数并根据目标类型进行转换
func convertArgs(fnType reflect.Type, args []Value) ([]reflect.Value, error) {
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		paramType := fnType.In(i)
		converted, err := convertValueToGo(arg, paramType)
		if err != nil {
			return nil, wrapConvertError(i, paramType, arg, err)
		}
		in[i] = converted
	}
	return in, nil
}

// ========== 返回值处理器 ==========

// returnValueHandler 返回值处理器函数类型
// 定义处理不同数量返回值的函数签名
type returnValueHandler func(out []reflect.Value) (Value, error)

// returnValueHandlers 返回值处理器映射表
// 根据返回值数量分发到对应处理器
// 支持0个返回值、1个返回值、2个返回值(value, error模式)
var returnValueHandlers = map[int]returnValueHandler{
	0: handleNoReturnValue,
	1: handleSingleReturnValue,
	2: handleDualReturnValue,
}

// handleNoReturnValue 处理无返回值情况
// Go函数没有返回值时，脚本侧返回nil
func handleNoReturnValue(out []reflect.Value) (Value, error) {
	return NewValue(nil), nil
}

// handleSingleReturnValue 处理单返回值情况
// 如果返回值是error类型，则作为错误返回
// 否则将返回值转换为脚本值
func handleSingleReturnValue(out []reflect.Value) (Value, error) {
	val := out[0].Interface()
	if err, ok := val.(error); ok {
		return Value{}, err
	}
	return convertGoToValue(val), nil
}

// handleDualReturnValue 处理双返回值情况（value, error模式）
func handleDualReturnValue(out []reflect.Value) (Value, error) {
	val := out[0].Interface()
	if err, ok := out[1].Interface().(error); ok && err != nil {
		return Value{}, err
	}
	return convertGoToValue(val), nil
}

// handleReturnValues 根据返回值数量分发到对应处理器
func handleReturnValues(out []reflect.Value) (Value, error) {
	handler, ok := returnValueHandlers[len(out)]
	if !ok {
		return Value{}, fmt.Errorf("外部函数返回值数量不支持: %d个（支持0、1、2个）", len(out))
	}
	return handler(out)
}

// ========== 脚本值 -> Go值转换 ==========

// errConvertFormat 类型转换错误格式字符串
const errConvertFormat = "无法将脚本%s值转换为Go类型: %v"

// intCompatibleKinds 整数兼容的Go类型集合
// 包含所有有符号整数类型
var intCompatibleKinds = map[reflect.Kind]bool{
	reflect.Int: true, reflect.Int8: true, reflect.Int16: true,
	reflect.Int32: true, reflect.Int64: true,
}

// floatCompatibleKinds 浮点数兼容的Go类型集合
var floatCompatibleKinds = map[reflect.Kind]bool{
	reflect.Float32: true, reflect.Float64: true,
}

// valueConverter 脚本值到Go值的转换函数类型
// 定义将脚本值转换为Go值的函数签名
type valueConverter func(val Value, targetType reflect.Type) (reflect.Value, error)

// valueConverters 各类型的转换器映射
// 根据脚本值类型分发到对应的转换函数
var valueConverters map[ValueType]valueConverter

// convertValueToGo 将脚本值转换为Go类型
// 根据脚本值类型查找对应的转换器并执行转换
func convertValueToGo(val Value, targetType reflect.Type) (reflect.Value, error) {
	converter, ok := valueConverters[val.Type]
	if !ok {
		return reflect.Value{}, fmt.Errorf("无法将脚本类型(%s)转换为Go类型(%v)", typeName(val.Type), targetType)
	}
	return converter(val, targetType)
}

// convertNumericValue 转换数值类型（整数和浮点数）
// 支持Go的所有整数和浮点数类型
// 采用延迟计算，仅在需要时获取对应的值
func convertNumericValue(val Value, targetType reflect.Type) (reflect.Value, error) {
	kind := targetType.Kind()

	if intCompatibleKinds[kind] {
		return reflect.ValueOf(val.Int()).Convert(targetType), nil
	}
	if floatCompatibleKinds[kind] {
		return reflect.ValueOf(val.Float()).Convert(targetType), nil
	}
	return reflect.Value{}, fmt.Errorf(errConvertFormat, "数值", targetType)
}

// convertStringValue 转换字符串值
// 脚本字符串直接转换为Go字符串
func convertStringValue(val Value, targetType reflect.Type) (reflect.Value, error) {
	return reflect.ValueOf(val.String()), nil
}

// convertBoolValue 转换布尔值
// 脚本布尔值直接转换为Go布尔值
func convertBoolValue(val Value, targetType reflect.Type) (reflect.Value, error) {
	return reflect.ValueOf(val.Bool()), nil
}

// convertNilValue 转换nil值
// 返回目标类型的零值
func convertNilValue(val Value, targetType reflect.Type) (reflect.Value, error) {
	return reflect.Zero(targetType), nil
}

// convertArrayValue 转换数组值到Go切片
// 递归转换每个元素到目标元素类型
func convertArrayValue(val Value, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf(errConvertFormat, "数组", targetType)
	}

	arr := val.Array()
	slice := reflect.MakeSlice(targetType, len(arr.Elements), len(arr.Elements))
	for i, elem := range arr.Elements {
		converted, err := convertValueToGo(elem, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		slice.Index(i).Set(converted)
	}
	return slice, nil
}

// convertMapValue 转换Map值到Go map
// 递归转换每个值到目标值类型
// 注意：脚本Map的键固定为字符串类型
func convertMapValue(val Value, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf(errConvertFormat, "Map", targetType)
	}

	m := val.Map()
	newMap := reflect.MakeMap(targetType)
	for k, v := range m.Pairs {
		keyVal := reflect.ValueOf(k)
		valVal, err := convertValueToGo(v, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		newMap.SetMapIndex(keyVal, valVal)
	}
	return newMap, nil
}

// ========== Go值 -> 脚本值转换 ==========

// goConverter Go值到脚本值的转换函数类型
// 定义将Go值转换为脚本值的函数签名
type goConverter func(v reflect.Value) Value

// goConverters 各类型的转换器映射
// 根据Go值的Kind分发到对应的转换函数
var goConverters map[reflect.Kind]goConverter

// convertGoToValue 将Go值转换为脚本值
// 支持基本类型、切片、数组、Map和指针
func convertGoToValue(val interface{}) Value {
	if val == nil {
		return NewValue(nil)
	}

	v := reflect.ValueOf(val)
	converter, ok := goConverters[v.Kind()]
	if !ok {
		return NewValue(nil)
	}
	return converter(v)
}

// convertGoString 转换Go字符串到脚本字符串
func convertGoString(v reflect.Value) Value { return NewValue(v.String()) }

// convertGoBool 转换Go布尔值到脚本布尔值
func convertGoBool(v reflect.Value) Value { return NewValue(v.Bool()) }

// convertGoSlice 转换Go切片/数组到脚本数组
// 递归转换每个元素
func convertGoSlice(v reflect.Value) Value {
	length := v.Len()
	arr := make([]Value, length)
	for i := 0; i < length; i++ {
		arr[i] = convertGoToValue(v.Index(i).Interface())
	}
	return NewValue(arr)
}

// convertGoMap 转换Go Map到脚本Map
// 键转换为字符串，值递归转换
// 优化：预分配Map容量，减少扩容开销
func convertGoMap(v reflect.Value) Value {
	length := v.Len()
	m := &MapValue{
		Pairs:   make(map[string]Value, length), // 预分配容量
		KeyType: TypeString,
	}
	for _, key := range v.MapKeys() {
		k := fmt.Sprintf("%v", key.Interface())
		m.Pairs[k] = convertGoToValue(v.MapIndex(key).Interface())
	}
	return NewValue(m)
}

// convertGoPtr 转换Go指针
// 解引用后递归转换，nil指针返回nil
func convertGoPtr(v reflect.Value) Value {
	if v.IsNil() {
		return NewValue(nil)
	}
	return convertGoToValue(v.Elem().Interface())
}

// ========== 转换器注册 ==========

// init 初始化类型转换器映射表
func init() {
	// 注册脚本值转换器
	valueConverters = map[ValueType]valueConverter{
		TypeInt:    convertNumericValue,
		TypeFloat:  convertNumericValue,
		TypeString: convertStringValue,
		TypeBool:   convertBoolValue,
		TypeNil:    convertNilValue,
		TypeArray:  convertArrayValue,
		TypeMap:    convertMapValue,
	}

	// 注册Go值转换器（基础类型）
	goConverters = map[reflect.Kind]goConverter{
		reflect.String: convertGoString,
		reflect.Bool:   convertGoBool,
		reflect.Slice:  convertGoSlice,
		reflect.Array:  convertGoSlice,
		reflect.Map:    convertGoMap,
		reflect.Ptr:    convertGoPtr,
	}

	// 批量注册数值类型转换器
	// 整数类型：使用int(v.Int())转换
	for _, k := range []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64} {
		goConverters[k] = func(v reflect.Value) Value { return NewValue(int(v.Int())) }
	}
	// 浮点类型：使用v.Float()转换
	for _, k := range []reflect.Kind{reflect.Float32, reflect.Float64} {
		goConverters[k] = func(v reflect.Value) Value { return NewValue(v.Float()) }
	}
	// 无符号整数类型：使用int(v.Uint())转换
	for _, k := range []reflect.Kind{reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64} {
		goConverters[k] = func(v reflect.Value) Value { return NewValue(int(v.Uint())) }
	}
}
