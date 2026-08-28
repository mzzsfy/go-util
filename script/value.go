package script

// ========== 运行时值类型 ==========
// 本文件包含运行时值的核心类型定义和基本操作

import (
	"fmt"
	"strconv"
	"strings"
)

// ValueType 值类型枚举
type ValueType int

const (
	// TypeNil nil类型
	TypeNil ValueType = iota

	// TypeInt 整数类型
	TypeInt

	// TypeFloat 浮点数类型
	TypeFloat

	// TypeString 字符串类型
	TypeString

	// TypeBool 布尔类型
	TypeBool

	// TypeArray 数组类型
	TypeArray

	// TypeMap Map类型
	TypeMap

	// TypeFunction 函数类型
	TypeFunction

	// TypeExternalFunc 外部函数类型
	TypeExternalFunc
)

// Value 运行时值
type Value struct {
	// Type 值类型
	Type ValueType

	// Data 实际数据
	Data any
}

// ArrayValue 数组值
type ArrayValue struct {
	// Elements 元素列表
	Elements []Value
}

// MapValue Map值
type MapValue struct {
	// Pairs 键值对映射
	Pairs map[string]Value

	// KeyType 键类型，固定为String
	KeyType ValueType
}

// FunctionValue 函数值
type FunctionValue struct {
	// Compiled 编译后的函数
	Compiled *CompiledFunction
}

// ExternalFuncValue 外部函数值
type ExternalFuncValue struct {
	// Name 函数名称
	Name string

	// Func 函数对象
	Func interface{}
}

// NewValue 创建新值
// 支持所有数值类型(int/int64/.../uint/.../float64/float32), string, bool, nil, []Value, *ArrayValue, *MapValue, *FunctionValue
func NewValue(data any) Value {
	if data == nil {
		return nilValue()
	}

	// 尝试基本类型转换
	if v, ok := tryPrimitiveValue(data); ok {
		return v
	}

	// 尝试复合类型转换
	return newCompositeValue(data)
}

// nilValue 返回nil值
func nilValue() Value {
	return Value{Type: TypeNil, Data: nil}
}

// smallInts 预缓存小整数Value(0-255)，避免重复接口装箱
var smallInts [256]Value

func init() {
	for i := 0; i < 256; i++ {
		smallInts[i] = Value{Type: TypeInt, Data: i}
	}
}

// intVal 快速构造int类型的Value，命中缓存时零分配
func intVal(v int) Value {
	if uint(v) < 256 {
		return smallInts[v]
	}
	return intValSlow(v)
}

//go:noinline
func intValSlow(v int) Value {
	return Value{Type: TypeInt, Data: v}
}

// tryPrimitiveValue 尝试将基本类型转换为Value
func tryPrimitiveValue(data any) (Value, bool) {
	switch v := data.(type) {
	case int:
		return intVal(v), true
	case int64:
		return intVal(int(v)), true
	case int32:
		return intVal(int(v)), true
	case int16:
		return intVal(int(v)), true
	case int8:
		return intVal(int(v)), true
	case uint:
		return intVal(int(v)), true
	case uint64:
		return intVal(int(v)), true
	case uint32:
		return intVal(int(v)), true
	case uint16:
		return intVal(int(v)), true
	case uint8:
		return intVal(int(v)), true
	case float64:
		return Value{Type: TypeFloat, Data: v}, true
	case float32:
		return Value{Type: TypeFloat, Data: float64(v)}, true
	case string:
		return Value{Type: TypeString, Data: v}, true
	case bool:
		return Value{Type: TypeBool, Data: v}, true
	default:
		return Value{}, false
	}
}

// newCompositeValue 将复合类型转换为Value
func newCompositeValue(data any) Value {
	switch v := data.(type) {
	case []Value:
		return Value{Type: TypeArray, Data: &ArrayValue{Elements: v}}
	case *ArrayValue:
		return Value{Type: TypeArray, Data: v}
	case *MapValue:
		return Value{Type: TypeMap, Data: v}
	case *FunctionValue:
		return Value{Type: TypeFunction, Data: v}
	default:
		return nilValue()
	}
}

// IsNil 判断值是否为nil
// 检查类型为nil或数据为nil
func (v Value) IsNil() bool {
	return v.Type == TypeNil
} // getTyped 获取指定类型的值（泛型辅助函数）
func getTyped[T any](v Value, typ ValueType) T {
	if v.Type == typ {
		return v.Data.(T)
	}
	var zero T
	return zero
}

// Int 获取整数值
func (v Value) Int() int {
	return getTyped[int](v, TypeInt)
}

// Float 获取浮点值
// TypeInt 自动提升为 float64, 支持混合运算
func (v Value) Float() float64 {
	if v.Type == TypeFloat {
		return v.Data.(float64)
	}
	if v.Type == TypeInt {
		return float64(v.Data.(int))
	}
	return 0
}

// String 获取字符串值
func (v Value) String() string {
	return getTyped[string](v, TypeString)
}

// Bool 获取布尔值
func (v Value) Bool() bool {
	return getTyped[bool](v, TypeBool)
}

// IsTruthy 判断值在布尔上下文中是否为真
// nil: false; bool: 直接值; 数值: 非0为true; string: 非空为true; 其他: true
func (v Value) IsTruthy() bool {
	switch v.Type {
	case TypeNil:
		return false
	case TypeBool:
		return v.Data.(bool)
	case TypeInt:
		return v.Data.(int) != 0
	case TypeFloat:
		return v.Data.(float64) != 0
	case TypeString:
		return v.Data.(string) != ""
	default:
		return true
	}
}

// Array 获取数组值
func (v Value) Array() *ArrayValue {
	return getTyped[*ArrayValue](v, TypeArray)
}

// Map 获取Map值
func (v Value) Map() *MapValue {
	return getTyped[*MapValue](v, TypeMap)
}

// Function 获取函数值
func (v Value) Function() *FunctionValue {
	return getTyped[*FunctionValue](v, TypeFunction)
}

// ExternalFunc 获取外部函数值
func (v Value) ExternalFunc() *ExternalFuncValue {
	return getTyped[*ExternalFuncValue](v, TypeExternalFunc)
}

// ========== 相等比较处理器 ==========

// equalFunc 相等比较函数类型
type equalFunc func(a, b Value) bool

// equalGeneric 泛型相等比较函数，支持任意可比较类型
func equalGeneric[T comparable](a, b Value) bool {
	return a.Data.(T) == b.Data.(T)
}

// equalHandlers 相等比较处理器映射表
// 在init()中初始化以避免初始化循环
var equalHandlers map[ValueType]equalFunc

func init() {
	equalHandlers = map[ValueType]equalFunc{
		TypeNil:    func(a, b Value) bool { return true },
		TypeInt:    equalGeneric[int],
		TypeFloat:  equalGeneric[float64],
		TypeString: equalGeneric[string],
		TypeBool:   equalGeneric[bool],
		TypeArray:  equalArray,
		TypeMap:    equalMap,
	}
}

// equalArray 递归比较两个数组的元素
func equalArray(a, b Value) bool {
	la, lb := a.Array(), b.Array()
	if len(la.Elements) != len(lb.Elements) {
		return false
	}
	for i := range la.Elements {
		if !la.Elements[i].Equal(lb.Elements[i]) {
			return false
		}
	}
	return true
}

// equalMap 递归比较两个Map的键值对
func equalMap(a, b Value) bool {
	ma, mb := a.Map(), b.Map()
	if len(ma.Pairs) != len(mb.Pairs) {
		return false
	}
	for k, va := range ma.Pairs {
		vb, ok := mb.Pairs[k]
		if !ok || !va.Equal(vb) {
			return false
		}
	}
	return true
}

// Equal 判断两个值是否相等
// 数值类型(int/float)混合比较时自动提升为float
func (v Value) Equal(other Value) bool {
	// 数值类型混合比较: int == float 自动提升
	if (v.Type == TypeInt || v.Type == TypeFloat) &&
		(other.Type == TypeInt || other.Type == TypeFloat) {
		return v.Float() == other.Float()
	}
	if v.Type != other.Type {
		return false
	}
	if handler, ok := equalHandlers[v.Type]; ok {
		return handler(v, other)
	}
	return false
}

// ========== 值格式化处理 ==========

// stringFormatter 字符串格式化函数类型
type stringFormatter func(v Value) string

// stringFormatters 各类型的字符串格式化器
// 在init()中初始化以避免循环依赖
var stringFormatters map[ValueType]stringFormatter

func init() {
	stringFormatters = map[ValueType]stringFormatter{
		TypeNil:          func(v Value) string { return "nil" },
		TypeInt:          func(v Value) string { return strconv.Itoa(v.Int()) },
		TypeFloat:        func(v Value) string { return strconv.FormatFloat(v.Float(), 'f', -1, 64) },
		TypeString:       func(v Value) string { return strconv.Quote(v.String()) },
		TypeBool:         func(v Value) string { return strconv.FormatBool(v.Bool()) },
		TypeArray:        formatArray,
		TypeMap:          formatMap,
		TypeFunction:     func(v Value) string { return fmt.Sprintf("<function %s>", v.Function().Compiled.Name) },
		TypeExternalFunc: func(v Value) string { return fmt.Sprintf("<external %s>", v.ExternalFunc().Name) },
	}
}

// formatArray 格式化数组值
func formatArray(v Value) string {
	arr := v.Array()
	elements := make([]string, len(arr.Elements))
	for i, e := range arr.Elements {
		elements[i] = formatValueForPrint(e)
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

// formatMap 格式化Map值
func formatMap(v Value) string {
	m := v.Map()
	pairs := make([]string, 0, len(m.Pairs))
	for k, val := range m.Pairs {
		pairs = append(pairs, fmt.Sprintf("%q: %s", k, formatValueForPrint(val)))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// formatValueForPrint 格式化值用于打印
func formatValueForPrint(v Value) string {
	if fn, ok := stringFormatters[v.Type]; ok {
		return fn(v)
	}
	return "<unknown>"
}

// GoString 返回值的Go语法格式字符串表示
func (v Value) GoString() string {
	return formatValueForPrint(v)
}

// formatValue 格式化值为字符串
func formatValue(v any) string {
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v", v)
}
