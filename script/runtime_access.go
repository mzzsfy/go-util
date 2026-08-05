package script

// ========== 运行时访问操作 ==========
// 包含索引访问、切片访问、索引赋值、长度计算

import (
	"math"
	"unicode/utf8"
)

// ========== 索引访问 ==========

// indexAccessor 索引访问函数类型
type indexAccessor func(*VM, Value, Value) (Value, error)

// indexAccessors 索引访问器映射表
var indexAccessors = map[ValueType]indexAccessor{
	TypeArray:  arrayIndexAccess,
	TypeString: stringIndexAccess,
	TypeMap:    mapIndexAccess,
}

// arrayIndexAccess 数组索引访问
func arrayIndexAccess(vm *VM, obj, index Value) (Value, error) {
	if index.Type != TypeInt {
		return Value{}, vm.runtimeErrorWithCode(ErrTypeMismatch,
			"类型错误：数组索引必须是整数。\n"+
				"→ 实际类型：%s\n"+
				"→ 用法：arr[0], arr[i] (i为整数)",
			typeName(index.Type))
	}
	arr := obj.Array()
	idx := index.Int()
	if !isValidIndex(idx, len(arr.Elements)) {
		return NewValue(nil), nil
	}
	return arr.Elements[idx], nil
}

// stringIndexAccess 字符串索引访问
// 按rune(Unicode字符)索引, 与len()和切片一致
func stringIndexAccess(vm *VM, obj, index Value) (Value, error) {
	if index.Type != TypeInt {
		return Value{}, vm.runtimeErrorWithCode(ErrTypeMismatch,
			"类型错误：字符串索引必须是整数。\n"+
				"→ 实际类型：%s\n"+
				"→ 用法：str[0], str[i] (i为整数)",
			typeName(index.Type))
	}
	s := obj.String()
	idx := index.Int()
	if idx < 0 {
		return NewValue(nil), nil
	}
	// 按rune遍历, 找到第idx个rune即返回, 无需[]rune分配
	// 注意: range string 的第一返回值是byte偏移而非rune索引, 故用独立计数器
	runeIdx := 0
	for _, r := range s {
		if runeIdx == idx {
			return NewValue(string(r)), nil
		}
		runeIdx++
	}
	return NewValue(nil), nil
}

// mapIndexAccess Map索引访问
func mapIndexAccess(vm *VM, obj, index Value) (Value, error) {
	m := obj.Map()
	key := index.String()
	if val, ok := m.Pairs[key]; ok {
		return val, nil
	}
	return NewValue(nil), nil
}

// indexAccess 执行索引访问
// 使用映射表分派到具体类型的访问函数
func (vm *VM) indexAccess(obj, index Value) (Value, error) {
	if accessor, ok := indexAccessors[obj.Type]; ok {
		return accessor(vm, obj, index)
	}
	typeStr := typeName(obj.Type)
	return Value{}, vm.runtimeErrorWithCode(ErrUnsupportedOp,
		"类型错误：无法对 '%s' 类型进行索引访问。\n"+
			"→ 问题：索引访问（obj[index]）只支持特定类型。\n"+
			"→ 支持的类型：\n"+
			"  - array：数组，使用整数索引，如 arr[0]\n"+
			"  - string：字符串，使用整数索引获取字符，如 str[0]\n"+
			"  - map：映射，使用字符串键，如 map[\"key\"]\n"+
			"→ 当前类型 '%s' 不支持索引操作\n"+
			"→ 建议：检查对象类型是否正确，或使用支持的类型",
		typeStr, typeStr)
}

// ========== 切片访问 ==========

// SliceEndDefault 切片结束索引的默认值
// 使用math.MinInt表示省略结束索引, 避免与用户输入的合法值冲突
const SliceEndDefault = math.MinInt

// sliceAccessor 切片访问函数类型
type sliceAccessor func(*VM, Value, int, int) (Value, error)

// sliceAccessors 切片访问器映射表
var sliceAccessors = map[ValueType]sliceAccessor{
	TypeArray:  arraySliceAccess,
	TypeString: stringSliceAccess,
}

// arraySliceAccess 数组切片访问
func arraySliceAccess(vm *VM, obj Value, s, e int) (Value, error) {
	arr := obj.Array()
	s, e, valid := normalizeSliceBounds(s, e, len(arr.Elements))
	if !valid {
		return NewValue([]Value{}), nil
	}
	return NewValue(arr.Elements[s:e]), nil
}

// stringSliceAccess 字符串切片访问
func stringSliceAccess(vm *VM, obj Value, start, end int) (Value, error) {
	s := obj.String()
	runeCount := utf8.RuneCountInString(s)
	start, end, valid := normalizeSliceBounds(start, end, runeCount)
	if !valid {
		return NewValue(""), nil
	}
	// 定位 start 的 byte 偏移
	byteStart := 0
	for i := 0; i < start; i++ {
		_, size := utf8.DecodeRuneInString(s[byteStart:])
		byteStart += size
	}
	// 定位 end 的 byte 偏移
	byteEnd := byteStart
	for i := start; i < end; i++ {
		_, size := utf8.DecodeRuneInString(s[byteEnd:])
		byteEnd += size
	}
	return NewValue(s[byteStart:byteEnd]), nil
}

// sliceAccess 执行切片访问
// 使用映射表分派到具体类型的切片函数
// 当 end 为 SliceEndDefault 时，使用对象的长度作为结束索引
func (vm *VM) sliceAccess(obj, start, end Value) (Value, error) {
	s := start.Int()
	e := end.Int()

	// 处理省略结束索引的情况
	if e == SliceEndDefault {
		e, _ = vm.length(obj)
	}

	if accessor, ok := sliceAccessors[obj.Type]; ok {
		return accessor(vm, obj, s, e)
	}
	typeStr := typeName(obj.Type)
	return Value{}, vm.runtimeErrorWithCode(ErrUnsupportedOp,
		"类型错误：无法对 '%s' 类型进行切片操作。\n"+
			"→ 问题：切片（obj[start:end]）只支持特定类型。\n"+
			"→ 支持的类型：\n"+
			"  - array：数组切片，如 arr[1:3] 返回子数组\n"+
			"  - string：字符串切片，如 str[0:5] 返回子字符串\n"+
			"→ 当前类型 '%s' 不支持切片操作\n"+
			"→ 建议：检查对象类型是否正确，或使用支持的类型",
		typeStr, typeStr)
}

// ========== 索引赋值 ==========

// storeIndexer 索引赋值函数类型
type storeIndexer func(*VM, Value, Value, Value) error

// storeIndexers 索引赋值器映射表
var storeIndexers = map[ValueType]storeIndexer{
	TypeArray: arrayStoreIndex,
	TypeMap:   mapStoreIndex,
}

// arrayStoreIndex 数组索引赋值
func arrayStoreIndex(vm *VM, obj, index, val Value) error {
	arr := obj.Array()
	idx := index.Int()
	if !isValidIndex(idx, len(arr.Elements)) {
		return vm.runtimeErrorWithCode(ErrIndexOutOfBounds,
			"数组索引越界：索引 %d 超出数组范围。\n"+
				"→ 问题：尝试访问的索引超出了数组的有效范围。\n"+
				"→ 数组长度：%d（有效索引范围：0 到 %d）\n"+
				"→ 访问索引：%d\n"+
				"→ 原因：\n"+
				"  - 索引为负数\n"+
				"  - 索引大于等于数组长度\n"+
				"→ 建议：\n"+
				"  - 使用 len(arr) 获取数组长度\n"+
				"  - 在访问前检查索引是否有效：if i >= 0 && i < len(arr) { arr[i] = value }",
			idx, len(arr.Elements), len(arr.Elements)-1, idx)
	}
	arr.Elements[idx] = val
	return nil
}

// mapStoreIndex Map索引赋值
func mapStoreIndex(vm *VM, obj, index, val Value) error {
	m := obj.Map()
	key := index.String()
	m.Pairs[key] = val
	return nil
}

// storeIndex 执行索引赋值
// 使用映射表分派到具体类型的赋值函数
func (vm *VM) storeIndex(obj, index, val Value) error {
	if indexer, ok := storeIndexers[obj.Type]; ok {
		return indexer(vm, obj, index, val)
	}
	typeStr := typeName(obj.Type)
	return vm.runtimeErrorWithCode(ErrUnsupportedOp,
		"类型错误：无法对 '%s' 类型进行索引赋值。\n"+
			"→ 问题：索引赋值（obj[index] = value）只支持特定类型。\n"+
			"→ 支持的类型：\n"+
			"  - array：数组，使用整数索引，如 arr[0] = 123\n"+
			"  - map：映射，使用字符串键，如 map[\"key\"] = \"value\"\n"+
			"→ 当前类型 '%s' 不支持索引赋值\n"+
			"→ 注意：字符串不可变，不支持索引赋值\n"+
			"→ 建议：检查对象类型是否正确，或使用支持的类型",
		typeStr, typeStr)
}

// ========== 长度计算 ==========

// lengthFunc 长度计算函数类型
type lengthFunc func(Value) int

// lengthFuncs 类型到长度获取函数的映射
var lengthFuncs = map[ValueType]lengthFunc{
	TypeArray:  func(v Value) int { return len(v.Array().Elements) },
	TypeString: func(v Value) int { return utf8.RuneCountInString(v.String()) },
	TypeMap:    func(v Value) int { return len(v.Map().Pairs) },
}

// length 获取值的长度
// 支持类型：数组、字符串、Map
func (vm *VM) length(val Value) (int, error) {
	if fn, ok := lengthFuncs[val.Type]; ok {
		return fn(val), nil
	}
	return 0, vm.runtimeError(
		"类型错误：无法获取 %s 类型的长度。\n"+
			"→ 问题：len() 只支持特定类型。\n"+
			"→ 支持的类型：array、string、map\n"+
			"→ 不支持的类型：%s\n"+
			"→ 建议：检查操作数类型",
		typeName(val.Type), typeName(val.Type))
}
