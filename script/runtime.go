package script

// ========== 运行时核心辅助函数 ==========
// 提供通用的辅助函数，供其他runtime_*文件使用

import "strconv"

// ========== 通用辅助函数 ==========

// isZeroValue 检查值是否为零值
func isZeroValue(v Value) bool {
	return (v.Type == TypeInt && v.Int() == 0) ||
		(v.Type == TypeFloat && v.Float() == 0.0)
}

// isValidIndex 检查索引是否在有效范围内
func isValidIndex(idx, length int) bool {
	return idx >= 0 && idx < length
}

// normalizeSliceBounds 规范化切片边界
// 负数索引从末尾倒数(Python风格), 返回规范化后的起始和结束位置及有效性
func normalizeSliceBounds(s, e, length int) (int, int, bool) {
	if s < 0 {
		s += length
		if s < 0 {
			s = 0
		}
	}
	if e < 0 {
		e += length
		if e < 0 {
			e = 0
		}
	}
	if e > length {
		e = length
	}
	return s, e, s < e
}

// valueToString 将值转换为字符串
// 字符串类型直接返回，其他类型使用GoString
func valueToString(v Value) string {
	if v.Type == TypeString {
		return v.String()
	}
	return v.GoString()
}

// ========== 类型转换 ==========

// convertStringToInt 将字符串转换为整数
func convertStringToInt(vm *VM, a Value) (Value, error) {
	i, err := strconv.Atoi(a.String())
	if err != nil {
		return Value{}, vm.runtimeError(
			"类型转换错误：无法将字符串转换为整数。\n"+
				"→ 问题：字符串 '%s' 不是有效的整数格式。\n"+
				"→ 支持的格式：\n"+
				"  - 十进制：\"123\", \"-456\"\n"+
				"  - 不支持十六进制、八进制、二进制\n"+
				"→ 失败原因：%v\n"+
				"→ 建议：检查字符串格式，确保只包含数字和可选的负号",
			a.String(), err)
	}
	return NewValue(i), nil
}

// toInt 执行整数类型转换
func (vm *VM) toInt(a Value) (Value, error) {
	switch a.Type {
	case TypeInt:
		return a, nil
	case TypeFloat:
		return NewValue(int(a.Float())), nil
	case TypeString:
		return convertStringToInt(vm, a)
	case TypeBool:
		if a.Bool() {
			return NewValue(1), nil
		}
		return NewValue(0), nil
	}
	return Value{}, vm.runtimeError(
		"类型转换错误：无法将 %s 转换为整数。\n"+
			"→ 问题：int() 函数只支持特定类型。\n"+
			"→ 支持的类型：\n"+
			"  - int：直接返回\n"+
			"  - float：截断小数部分（如 3.14 → 3）\n"+
			"  - string：解析字符串（如 \"123\" → 123）\n"+
			"  - bool：true → 1, false → 0\n"+
			"→ 不支持的类型：%s（array、map、function等）\n"+
			"→ 建议：先转换为支持的类型",
		typeName(a.Type), typeName(a.Type))
}

// ========== 浮点数转换 ==========

// convertStringToFloat 将字符串转换为浮点数
func convertStringToFloat(vm *VM, a Value) (Value, error) {
	f, err := strconv.ParseFloat(a.String(), 64)
	if err != nil {
		return Value{}, vm.runtimeError(
			"类型转换错误：无法将字符串转换为浮点数。\n"+
				"→ 问题：字符串 '%s' 不是有效的浮点数格式。\n"+
				"→ 支持的格式：\n"+
				"  - 小数：\"3.14\", \"-0.5\"\n"+
				"  - 科学计数法：\"1e10\", \"2.5e-3\"\n"+
				"→ 失败原因：%v\n"+
				"→ 建议：检查字符串格式，确保符合浮点数语法",
			a.String(), err)
	}
	return NewValue(f), nil
}

// toFloat 执行浮点数类型转换
func (vm *VM) toFloat(a Value) (Value, error) {
	switch a.Type {
	case TypeInt:
		return NewValue(float64(a.Int())), nil
	case TypeFloat:
		return a, nil
	case TypeString:
		return convertStringToFloat(vm, a)
	case TypeBool:
		if a.Bool() {
			return NewValue(1.0), nil
		}
		return NewValue(0.0), nil
	}
	return Value{}, vm.runtimeError(
		"类型转换错误：无法将 %s 转换为浮点数。\n"+
			"→ 问题：float() 函数只支持特定类型。\n"+
			"→ 支持的类型：\n"+
			"  - int：转换为浮点数（如 123 → 123.0）\n"+
			"  - float：直接返回\n"+
			"  - string：解析字符串（如 \"3.14\" → 3.14）\n"+
			"  - bool：true → 1.0, false → 0.0\n"+
			"→ 不支持的类型：%s（array、map、function等）\n"+
			"→ 建议：先转换为支持的类型",
		typeName(a.Type), typeName(a.Type))
}

// ========== 位运算 ==========

// bitOpFunc 位运算函数类型
type bitOpFunc func(a, b int) int

// bitUnaryFunc 位一元运算函数类型
type bitUnaryFunc func(a int) int

// bitBinaryOp 执行位二元运算的通用函数
// 只支持整数类型
func (vm *VM) bitBinaryOp(a, b Value, op bitOpFunc, opName string) (Value, error) {
	if a.Type != TypeInt || b.Type != TypeInt {
		return Value{}, vm.runtimeError(
			"类型错误：位运算只支持整数类型。\n"+
				"→ 问题：%s 运算符只支持整数。\n"+
				"→ 支持的类型：int %s int\n"+
				"→ 实际类型：%s %s %s\n"+
				"→ 建议：使用 int() 将浮点数转换为整数",
			opName, opName, typeName(a.Type), opName, typeName(b.Type))
	}
	return NewValue(op(a.Int(), b.Int())), nil
}

// bitUnaryOp 执行位一元运算的通用函数
// 只支持整数类型
func (vm *VM) bitUnaryOp(a Value, op bitUnaryFunc, opName string) (Value, error) {
	if a.Type != TypeInt {
		return Value{}, vm.runtimeError(
			"类型错误：位运算只支持整数类型。\n"+
				"→ 问题：%s 运算符只支持整数。\n"+
				"→ 支持的类型：int\n"+
				"→ 实际类型：%s\n"+
				"→ 建议：使用 int() 将浮点数转换为整数",
			opName, typeName(a.Type))
	}
	return NewValue(op(a.Int())), nil
}

// bitAnd 执行位与运算
// 支持类型：整数
func (vm *VM) bitAnd(a, b Value) (Value, error) {
	return vm.bitBinaryOp(a, b, func(x, y int) int { return x & y }, "位与(&)")
}

// bitOr 执行位或运算
// 支持类型：整数
func (vm *VM) bitOr(a, b Value) (Value, error) {
	return vm.bitBinaryOp(a, b, func(x, y int) int { return x | y }, "位或(|)")
}

// bitXor 执行位异或运算
// 支持类型：整数
func (vm *VM) bitXor(a, b Value) (Value, error) {
	return vm.bitBinaryOp(a, b, func(x, y int) int { return x ^ y }, "位异或(^)")
}

// bitNot 执行位取反运算
// 支持类型：整数
func (vm *VM) bitNot(a Value) (Value, error) {
	return vm.bitUnaryOp(a, func(x int) int { return ^x }, "位取反(^)")
}

// lshift 执行左移运算
// 支持类型：整数
func (vm *VM) lshift(a, b Value) (Value, error) {
	return vm.bitBinaryOp(a, b, func(x, y int) int { return x << y }, "左移(<<)")
}

// rshift 执行右移运算
// 支持类型：整数
func (vm *VM) rshift(a, b Value) (Value, error) {
	return vm.bitBinaryOp(a, b, func(x, y int) int { return x >> y }, "右移(>>)")
}

// ========== 比较运算 ==========

// compareOp 比较操作类型
type compareOp int

const (
	compareLess compareOp = iota
	compareLessEq
	compareGreater
	compareGreaterEq
)

// compare 执行比较运算
// 数值类型(int/float)混合比较时自动提升为float
func (vm *VM) compare(a, b Value, op compareOp) (bool, error) {
	// 数值类型混合比较: int vs float 提升为 float 比较
	if (a.Type == TypeInt || a.Type == TypeFloat) &&
		(b.Type == TypeInt || b.Type == TypeFloat) {
		return vm.compareFloat(a.Float(), b.Float(), op), nil
	}

	// 非数值类型需要类型一致
	if a.Type != b.Type {
		return false, vm.runtimeErrorWithCode(ErrTypeMismatch,
			"类型错误：无法比较不同类型的值。\n"+
				"→ 问题：比较运算符要求两边类型相同。\n"+
				"→ 左操作数类型：%s\n"+
				"→ 右操作数类型：%s\n"+
				"→ 建议：使用类型转换使两边类型一致",
			typeName(a.Type), typeName(b.Type))
	}

	// 类型相同，执行比较
	switch a.Type {
	case TypeInt:
		return vm.compareInt(a.Int(), b.Int(), op), nil
	case TypeFloat:
		return vm.compareFloat(a.Float(), b.Float(), op), nil
	case TypeString:
		return vm.compareString(a.String(), b.String(), op), nil
	default:
		return false, vm.runtimeError(
			"类型错误：类型 %s 不支持比较运算。\n"+
				"→ 问题：只有可排序的类型才支持比较运算。\n"+
				"→ 支持的类型：int、float、string\n"+
				"→ 不支持的类型：%s\n"+
				"→ 建议：检查操作数类型",
			typeName(a.Type), typeName(a.Type))
	}
}

// compareInt 比较两个整数
func (vm *VM) compareInt(a, b int, op compareOp) bool {
	switch op {
	case compareLess:
		return a < b
	case compareLessEq:
		return a <= b
	case compareGreater:
		return a > b
	case compareGreaterEq:
		return a >= b
	}
	return false
}

// compareFloat 比较两个浮点数
func (vm *VM) compareFloat(a, b float64, op compareOp) bool {
	switch op {
	case compareLess:
		return a < b
	case compareLessEq:
		return a <= b
	case compareGreater:
		return a > b
	case compareGreaterEq:
		return a >= b
	}
	return false
}

// compareString 比较两个字符串
func (vm *VM) compareString(a, b string, op compareOp) bool {
	switch op {
	case compareLess:
		return a < b
	case compareLessEq:
		return a <= b
	case compareGreater:
		return a > b
	case compareGreaterEq:
		return a >= b
	}
	return false
}

// less 执行小于比较
func (vm *VM) less(a, b Value) (bool, error) {
	return vm.compare(a, b, compareLess)
}

// lessEq 执行小于等于比较
func (vm *VM) lessEq(a, b Value) (bool, error) {
	return vm.compare(a, b, compareLessEq)
}

// greater 执行大于比较
func (vm *VM) greater(a, b Value) (bool, error) {
	return vm.compare(a, b, compareGreater)
}

// greaterEq 执行大于等于比较
func (vm *VM) greaterEq(a, b Value) (bool, error) {
	return vm.compare(a, b, compareGreaterEq)
}

// ========== 数组操作 ==========

// builtinPush 实现 push(arr, elem) 内置函数
// 向数组末尾追加元素并返回新数组
func (vm *VM) builtinPush(arr, elem Value) (Value, error) {
	if arr.Type != TypeArray {
		return Value{}, vm.runtimeError(
			"类型错误：push() 第一个参数必须是数组。\n"+
				"→ 实际类型：%s\n"+
				"→ 用法：push(array, element)",
			typeName(arr.Type))
	}

	// 创建新数组（避免修改原数组）
	oldElems := arr.Array().Elements
	newElems := make([]Value, len(oldElems)+1)
	copy(newElems, oldElems)
	newElems[len(oldElems)] = elem

	return NewValue(&ArrayValue{Elements: newElems}), nil
}
