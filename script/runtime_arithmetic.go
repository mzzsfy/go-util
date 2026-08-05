package script

// ========== 运行时算术运算 ==========
// 包含加法、减法、乘法、除法、取模、一元运算

// ========== 加法运算 ==========

// addHandler 加法运算处理器类型
type addHandler func(*VM, Value, Value) (Value, error)

// addHandlers 加法运算处理器映射表
// 包含所有支持的类型组合，使用map查找替代条件分支
var addHandlers = map[[2]ValueType]addHandler{
	// 精确类型匹配
	{TypeString, TypeString}: (*VM).addStringString,
	{TypeInt, TypeInt}:       (*VM).addIntInt,
	{TypeArray, TypeArray}:   (*VM).addArrayArray,
	// 浮点数运算
	{TypeFloat, TypeFloat}: (*VM).addFloatFloat,
	// 数值类型隐式提升
	{TypeInt, TypeFloat}: (*VM).addFloatFloat,
	{TypeFloat, TypeInt}: (*VM).addFloatFloat,
	// 字符串拼接（string + basic）
	{TypeString, TypeInt}:   (*VM).addStringOther,
	{TypeString, TypeFloat}: (*VM).addStringOther,
	{TypeString, TypeBool}:  (*VM).addStringOther,
	{TypeString, TypeNil}:   (*VM).addStringOther,
	// 字符串拼接（basic + string）
	{TypeInt, TypeString}:   (*VM).addOtherString,
	{TypeFloat, TypeString}: (*VM).addOtherString,
	{TypeBool, TypeString}:  (*VM).addOtherString,
	{TypeNil, TypeString}:   (*VM).addOtherString,
}

// addStringString 字符串+字符串
func (vm *VM) addStringString(a, b Value) (Value, error) {
	return NewValue(a.String() + b.String()), nil
}

// addStringOther 字符串+其他类型
func (vm *VM) addStringOther(a, b Value) (Value, error) {
	return NewValue(a.String() + b.GoString()), nil
}

// addOtherString 其他类型+字符串
func (vm *VM) addOtherString(a, b Value) (Value, error) {
	return NewValue(a.GoString() + b.String()), nil
}

// addIntInt 整数+整数
func (vm *VM) addIntInt(a, b Value) (Value, error) {
	return NewValue(a.Int() + b.Int()), nil
}

// addFloatFloat 浮点数运算（任一为浮点数）
func (vm *VM) addFloatFloat(a, b Value) (Value, error) {
	return NewValue(a.Float() + b.Float()), nil
}

// addArrayArray 数组连接
func (vm *VM) addArrayArray(a, b Value) (Value, error) {
	return vm.concatArrays(a, b), nil
}

// concatArrays 连接两个数组
func (vm *VM) concatArrays(a, b Value) Value {
	arr1 := a.Array()
	arr2 := b.Array()
	result := make([]Value, len(arr1.Elements)+len(arr2.Elements))
	copy(result, arr1.Elements)
	copy(result[len(arr1.Elements):], arr2.Elements)
	return NewValue(result)
}

// add 执行加法运算
// 使用类型映射表分派到具体处理器
func (vm *VM) add(a, b Value) (Value, error) {
	if handler, ok := addHandlers[[2]ValueType{a.Type, b.Type}]; ok {
		return handler(vm, a, b)
	}
	return Value{}, vm.runtimeErrorWithCode(ErrTypeMismatch,
		"类型错误：无法对 %s 和 %s 执行加法运算。\n"+
			"→ 问题：加法运算符（+）不支持这两个类型的组合。\n"+
			"→ 支持的类型组合：\n"+
			"  - int + int → int（整数相加）\n"+
			"  - float + float → float（浮点数相加）\n"+
			"  - int + float → float（整数与浮点数相加，自动提升）\n"+
			"  - string + string → string（字符串拼接）\n"+
			"  - string + 基本类型 → string（字符串与其他类型拼接）\n"+
			"  - array + array → array（数组连接）\n"+
			"→ 不支持的组合：%s + %s\n"+
			"→ 建议：检查操作数类型，必要时使用类型转换函数（int(), float(), string()）",
		typeName(a.Type), typeName(b.Type), typeName(a.Type), typeName(b.Type))
}

// ========== 算术运算辅助函数 ==========

// numericOpFunc 数值运算函数类型
type numericOpFunc func(a, b int) int

// numericBinaryOp 执行数值二元运算的通用函数
// 优先执行整数运算，若有浮点数则提升为浮点数运算
func (vm *VM) numericBinaryOp(a, b Value, intOp numericOpFunc, floatOp func(a, b float64) float64, opName string) (Value, error) {
	// 整数运算
	if a.Type == TypeInt && b.Type == TypeInt {
		return NewValue(intOp(a.Int(), b.Int())), nil
	}

	// 浮点数运算（任一为浮点数则提升）
	if a.Type == TypeFloat || b.Type == TypeFloat {
		return NewValue(floatOp(a.Float(), b.Float())), nil
	}

	return Value{}, vm.runtimeError(
		"类型错误：无法对 %s 和 %s 执行%s运算。\n"+
			"→ 问题：%s运算符只支持数值类型。\n"+
			"→ 支持的类型组合：\n"+
			"  - int %s int → int\n"+
			"  - float %s float → float\n"+
			"  - int %s float → float（自动类型提升）\n"+
			"  - float %s int → float（自动类型提升）\n"+
			"→ 不支持的类型：%s, %s\n"+
			"→ 建议：使用 int() 或 float() 进行类型转换",
		typeName(a.Type), typeName(b.Type), opName, opName, opName, opName, opName, opName, typeName(a.Type), typeName(b.Type))
}

// sub 执行减法运算
// 支持类型：整数、浮点数
func (vm *VM) sub(a, b Value) (Value, error) {
	return vm.numericBinaryOp(a, b,
		func(x, y int) int { return x - y },
		func(x, y float64) float64 { return x - y },
		"减法")
}

// mul 执行乘法运算
// 支持类型：整数、浮点数
func (vm *VM) mul(a, b Value) (Value, error) {
	return vm.numericBinaryOp(a, b,
		func(x, y int) int { return x * y },
		func(x, y float64) float64 { return x * y },
		"乘法")
}

// div 执行除法运算
// 支持类型：整数、浮点数
// 注意：会检查除零错误
func (vm *VM) div(a, b Value) (Value, error) {
	if isZeroValue(b) {
		return Value{}, vm.runtimeErrorWithCode(ErrDivisionByZero,
			"除零错误：除数不能为零。\n"+
				"→ 问题：尝试除以零，这在数学上是未定义的。\n"+
				"→ 表达式：%s / 0\n"+
				"→ 原因：\n"+
				"  - 除数为整数 0\n"+
				"  - 除数为浮点数 0.0\n"+
				"→ 建议：\n"+
				"  - 在除法前检查除数是否为零\n"+
				"  - 示例：if divisor != 0 { result = x / divisor } else { result = 0 }",
			typeName(a.Type))
	}

	return vm.numericBinaryOp(a, b,
		func(x, y int) int { return x / y },
		func(x, y float64) float64 { return x / y },
		"除法")
}

// mod 执行取模运算
// 支持类型：整数
// 注意：会检查除零错误
func (vm *VM) mod(a, b Value) (Value, error) {
	if isZeroValue(b) {
		return Value{}, vm.runtimeErrorWithCode(ErrDivisionByZero,
			"取模运算错误：除数不能为零。\n"+
				"→ 问题：尝试对零取模，这在数学上是未定义的。\n"+
				"→ 表达式：%s %% 0\n"+
				"→ 原因：\n"+
				"  - 模数为整数 0\n"+
				"→ 建议：\n"+
				"  - 在取模前检查模数是否为零\n"+
				"  - 示例：if divisor != 0 { result = x %% divisor } else { result = 0 }",
			typeName(a.Type))
	}

	// 只支持整数取模
	if a.Type == TypeInt && b.Type == TypeInt {
		return NewValue(a.Int() % b.Int()), nil
	}

	return Value{}, vm.runtimeError(
		"类型错误：取模运算只支持整数类型。\n"+
			"→ 问题：取模运算符（%%）不支持浮点数。\n"+
			"→ 支持的类型：int %% int\n"+
			"→ 实际类型：%s %% %s\n"+
			"→ 建议：使用 int() 将浮点数转换为整数后再取模",
		typeName(a.Type), typeName(b.Type))
}

// ========== 一元运算 ==========

// negFunc 一元取负函数类型
type negFunc func(Value) Value

// negFuncs 取负运算映射表
var negFuncs = map[ValueType]negFunc{
	TypeInt:   func(v Value) Value { return NewValue(-v.Int()) },
	TypeFloat: func(v Value) Value { return NewValue(-v.Float()) },
}

// neg 执行一元负号运算
// 支持类型：整数、浮点数
func (vm *VM) neg(a Value) (Value, error) {
	if fn, ok := negFuncs[a.Type]; ok {
		return fn(a), nil
	}
	return Value{}, vm.runtimeError(
		"类型错误：无法对 %s 执行取负运算。\n"+
			"→ 问题：取负运算符（-）只支持数值类型。\n"+
			"→ 支持的类型：\n"+
			"  - int：整数（如 -123）\n"+
			"  - float：浮点数（如 -3.14）\n"+
			"→ 不支持的类型：%s（string、bool、array、map等）\n"+
			"→ 建议：使用 int() 或 float() 进行类型转换后再取负",
		typeName(a.Type), typeName(a.Type))
}
