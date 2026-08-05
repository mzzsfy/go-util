package script

import (
	"fmt"
	"strings"
)

// ========== 操作码处理器 ==========
// 本文件包含所有操作码处理器的类型定义、工厂函数和注册表

// ========== 预计算常量值 ==========
// 避免每次操作都创建新的Value对象，提升性能

var (
	// precomputedNil 预计算的nil值，避免重复创建
	precomputedNil = NewValue(nil)
	// precomputedTrue 预计算的true值，避免重复创建
	precomputedTrue = NewValue(true)
	// precomputedFalse 预计算的false值，避免重复创建
	precomputedFalse = NewValue(false)
)

// ========== 运行时错误模板 ==========
// 集中管理错误消息模板，便于维护和国际化

// runtimeErrorTemplates 运行时错误消息模板
// 使用format占位符，按顺序填充参数
var runtimeErrorTemplates = struct {
	// 绑定值相关
	BindValueNotFound string
	// 函数相关
	FuncNotSupported    string
	ExternalFuncInvalid string
	BindFuncNotFound    string
	ExternalFuncFailed  string
	CallStackOverflow   string
	// 指令相关
	UnknownOpcode string
}{
	BindValueNotFound: "运行时错误：未找到绑定值 '%s'。\n" +
		"→ 问题：尝试访问一个未绑定的值。\n" +
		"→ 可能的原因：\n" +
		"  - 值名拼写错误\n" +
		"  - 忘记绑定该值\n" +
		"  - 在错误的上下文中访问\n" +
		"→ 解决方法：在执行脚本前使用 context.BindValue() 绑定值\n" +
		"→ 示例：ctx.BindValue(\"myValue\", 123)\n" +
		"→ 提示：检查绑定名称是否与脚本中使用的名称一致",

	FuncNotSupported: "运行时错误：函数调用失败。\n" +
		"→ 问题：尝试调用的值不是有效的用户定义函数。\n" +
		"→ 可能的原因：\n" +
		"  - 调用了nil或未初始化的函数\n" +
		"  - 调用了非函数类型的值\n" +
		"  - 栈状态异常（参数不足）\n" +
		"→ 解决方法：检查函数是否正确定义和赋值",

	CallStackOverflow: "运行时错误：调用栈溢出（当前深度：%d，最大深度：%d）。\n" +
		"→ 问题：函数递归调用深度超过限制。\n" +
		"→ 可能的原因：\n" +
		"  - 无限递归（缺少终止条件）\n" +
		"  - 递归深度过大\n" +
		"→ 解决方法：\n" +
		"  - 检查递归终止条件是否正确\n" +
		"  - 使用 WithMaxCallDepth() 调整最大调用深度",

	ExternalFuncInvalid: "内部错误：外部函数引用无效（索引：%d）。\n" +
		"→ 这是一个编译器或虚拟机内部错误，通常不应出现。\n" +
		"→ 可能的原因：\n" +
		"  - #fn 指令解析错误\n" +
		"  - 外部函数列表损坏\n" +
		"  - 编译器生成的索引不正确\n" +
		"→ 建议：\n" +
		"  - 检查 #fn 指令的语法是否正确\n" +
		"  - 确保所有 #fn 声明的函数都已绑定\n" +
		"  - 重新编译脚本\n" +
		"  - 如果问题持续，请联系开发者",

	BindFuncNotFound: "运行时错误：未找到绑定的外部函数 '%s'。\n" +
		"→ 问题：尝试调用一个未绑定的外部函数。\n" +
		"→ 可能的原因：\n" +
		"  - 函数名拼写错误\n" +
		"  - 忘记绑定该函数\n" +
		"  - 在错误的上下文中调用\n" +
		"→ 解决方法：在执行脚本前使用 context.BindFunc() 绑定函数\n" +
		"→ 示例：\n" +
		"  Go代码：\n" +
		"    ctx.BindFunc(\"%s\", myGoFunc)\n" +
		"  脚本中已声明：\n" +
		"    #fn %s()=>...\n" +
		"→ 提示：检查绑定名称是否与 #fn 声明的名称一致",

	ExternalFuncFailed: "调用外部函数 '%s' 失败：%v。\n" +
		"→ 问题：外部函数调用过程中发生错误。\n" +
		"→ 可能的原因：\n" +
		"  - 函数签名与 #fn 声明不匹配\n" +
		"  - 参数类型不正确\n" +
		"  - 参数数量不匹配\n" +
		"  - 函数内部发生panic\n" +
		"→ 建议：\n" +
		"  - 检查 Go 函数签名是否与 #fn 声明一致\n" +
		"  - 确保传入的参数类型正确\n" +
		"  - 检查 Go 函数实现是否有错误",

	UnknownOpcode: "内部错误：未知的指令码 %d。\n" +
		"→ 这是一个严重的虚拟机内部错误，通常不应出现。\n" +
		"→ 可能的原因：\n" +
		"  - 编译器与虚拟机版本不兼容\n" +
		"  - 字节码损坏或被意外修改\n" +
		"  - 脚本引擎内部bug\n" +
		"→ 当前指令位置：%s 函数的第 %d 条指令\n" +
		"→ 建议：\n" +
		"  - 确保脚本引擎各组件版本一致\n" +
		"  - 重新编译脚本\n" +
		"  - 如果问题持续，请联系开发者并提供脚本代码",
}

// opcodeHandler 定义操作码处理函数类型
type opcodeHandler func(vm *VM, inst Instruction, frame *Frame) error

// opcodeHandlers 存储所有操作码的处理器
// 使用数组而不是map以获得更好的性能（OpCode是byte类型）
var opcodeHandlers [256]opcodeHandler

// ========== 常量操作处理器 ==========

// handleConst 处理常量加载
// 从常量池中加载指定索引的常量值到栈顶
func handleConst(vm *VM, inst Instruction, frame *Frame) error {
	constant := frame.Function.Constants[inst.Args[0]]
	vm.push(constant)
	return nil
}

// makeConstHandler 创建常量加载处理器
// 用于生成返回固定预计算值的处理器
func makeConstHandler(val Value) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		vm.push(val)
		return nil
	}
}

// 使用工厂函数创建常量加载处理器
var (
	handleNil   = makeConstHandler(precomputedNil)
	handleTrue  = makeConstHandler(precomputedTrue)
	handleFalse = makeConstHandler(precomputedFalse)
)

// ========== 变量操作处理器 ==========

// makeLoadHandler 创建变量加载处理器
// getVar: 从指定存储区获取变量的函数
func makeLoadHandler(getVar func(vm *VM, frame *Frame, idx int) Value) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		vm.push(getVar(vm, frame, inst.Args[0]))
		return nil
	}
}

// makeStoreHandler 创建变量存储处理器
// setVar: 将值存储到指定存储区的函数
func makeStoreHandler(setVar func(vm *VM, frame *Frame, idx int, val Value)) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		setVar(vm, frame, inst.Args[0], vm.pop())
		return nil
	}
}

// 使用工厂函数创建变量加载/存储处理器
var (
	handleLoadLocal = makeLoadHandler(func(vm *VM, frame *Frame, idx int) Value {
		return frame.Locals[idx]
	})
	handleStoreLocal = makeStoreHandler(func(vm *VM, frame *Frame, idx int, val Value) {
		frame.Locals[idx] = val
	})
)

// handleLoadBind 处理绑定值加载
// 从上下文中获取绑定值并压入栈顶
func handleLoadBind(vm *VM, inst Instruction, frame *Frame) error {
	// 从栈顶获取绑定值名称
	nameVal := vm.pop()
	name := nameVal.String()
	val, ok := vm.Context.GetBindValue(name)
	if !ok {
		return vm.runtimeError(runtimeErrorTemplates.BindValueNotFound, name)
	}
	vm.push(val)
	return nil
}

// ========== 数组和Map操作处理器 ==========

// handleIndex 处理索引访问
// 支持数组、字符串、Map的索引访问
func handleIndex(vm *VM, inst Instruction, frame *Frame) error {
	index := vm.pop()
	obj := vm.pop()
	result, err := vm.indexAccess(obj, index)
	if err != nil {
		return err
	}
	vm.push(result)
	return nil
}

// handleSlice 处理切片操作
// 支持数组和字符串的切片访问
func handleSlice(vm *VM, inst Instruction, frame *Frame) error {
	end := vm.pop()
	start := vm.pop()
	obj := vm.pop()
	result, err := vm.sliceAccess(obj, start, end)
	if err != nil {
		return err
	}
	vm.push(result)
	return nil
}

// handleStoreIndex 处理索引赋值
// 支持数组和Map的索引赋值
// 栈布局：obj, index, value -> value (保留值在栈上作为表达式结果)
func handleStoreIndex(vm *VM, inst Instruction, frame *Frame) error {
	val := vm.pop()
	index := vm.pop()
	obj := vm.pop()
	if err := vm.storeIndex(obj, index, val); err != nil {
		return err
	}
	// 将赋值的值保留在栈上，作为表达式的结果
	vm.push(val)
	return nil
}

// ========== 算术运算处理器类型 ==========

// binaryValueFunc 二元值运算函数类型（VM方法）
type binaryValueFunc func(vm *VM, a, b Value) (Value, error)

// unaryValueFunc 一元值运算函数类型（VM方法）
type unaryValueFunc func(vm *VM, a Value) (Value, error)

// handleBinaryOp 执行二元运算的通用处理器
func handleBinaryOp(vm *VM, op binaryValueFunc) error {
	b := vm.pop()
	a := vm.pop()
	result, err := op(vm, a, b)
	if err != nil {
		return err
	}
	vm.push(result)
	return nil
}

// handleUnaryOp 执行一元运算的通用处理器
func handleUnaryOp(vm *VM, op unaryValueFunc) error {
	a := vm.pop()
	result, err := op(vm, a)
	if err != nil {
		return err
	}
	vm.push(result)
	return nil
}

// makeBinaryHandler 创建二元运算处理器
func makeBinaryHandler(op binaryValueFunc) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		return handleBinaryOp(vm, op)
	}
}

// makeUnaryHandler 创建一元运算处理器
func makeUnaryHandler(op unaryValueFunc) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		return handleUnaryOp(vm, op)
	}
}

// ========== 比较运算处理器类型 ==========

// compareFunc 比较函数类型（VM方法）
type compareFunc func(vm *VM, a, b Value) (bool, error)

// handleCompareOp 执行比较运算的通用处理器
func handleCompareOp(vm *VM, cmp compareFunc) error {
	b := vm.pop()
	a := vm.pop()
	result, err := cmp(vm, a, b)
	if err != nil {
		return err
	}
	vm.push(NewValue(result))
	return nil
}

// makeCompareHandler 创建比较运算处理器
func makeCompareHandler(cmp compareFunc) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		return handleCompareOp(vm, cmp)
	}
}

// makeEqualHandler 创建相等比较处理器
func makeEqualHandler(negate bool) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		b := vm.pop()
		a := vm.pop()
		result := a.Equal(b)
		if negate {
			result = !result
		}
		vm.push(NewValue(result))
		return nil
	}
}

// handleNot 处理逻辑非运算
func handleNot(vm *VM, inst Instruction, frame *Frame) error {
	a := vm.pop()
	vm.push(NewValue(!a.IsTruthy()))
	return nil
}

// ========== 运算处理器实例 ==========

// 算术运算处理器
var (
	handleAdd    = makeBinaryHandler((*VM).add)
	handleSub    = makeBinaryHandler((*VM).sub)
	handleMul    = makeBinaryHandler((*VM).mul)
	handleDiv    = makeBinaryHandler((*VM).div)
	handleMod    = makeBinaryHandler((*VM).mod)
	handleNeg    = makeUnaryHandler((*VM).neg)
	handleBitAnd = makeBinaryHandler((*VM).bitAnd)
	handleBitOr  = makeBinaryHandler((*VM).bitOr)
	handleBitXor = makeBinaryHandler((*VM).bitXor)
	handleBitNot = makeUnaryHandler((*VM).bitNot)
	handleLShift = makeBinaryHandler((*VM).lshift)
	handleRShift = makeBinaryHandler((*VM).rshift)
)

// 比较运算处理器
var (
	handleLess      = makeCompareHandler((*VM).less)
	handleLessEq    = makeCompareHandler((*VM).lessEq)
	handleGreater   = makeCompareHandler((*VM).greater)
	handleGreaterEq = makeCompareHandler((*VM).greaterEq)
	handleEqual     = makeEqualHandler(false)
	handleNotEqual  = makeEqualHandler(true)
)

// ========== 类型转换处理器 ==========

// makeTypeConvertHandler 创建类型转换处理器
func makeTypeConvertHandler(convert func(vm *VM, a Value) (Value, error)) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		result, err := convert(vm, vm.pop())
		if err != nil {
			return err
		}
		vm.push(result)
		return nil
	}
}

// makeSimpleUnaryHandler 创建简单的一元操作处理器
func makeSimpleUnaryHandler(op func(vm *VM, a Value) Value) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		vm.push(op(vm, vm.pop()))
		return nil
	}
}

var (
	handleToInt    = makeTypeConvertHandler((*VM).toInt)
	handleToFloat  = makeTypeConvertHandler((*VM).toFloat)
	handleToString = makeSimpleUnaryHandler(func(vm *VM, a Value) Value {
		// 如果已经是字符串类型，直接返回，避免重复加引号
		if a.Type == TypeString {
			return a
		}
		// 其他类型调用GoString()转换为字符串表示
		return NewValue(a.GoString())
	})
	handleTypeOf = makeSimpleUnaryHandler(func(vm *VM, a Value) Value {
		return NewValue(typeName(a.Type))
	})
	handleLen = makeTypeConvertHandler(func(vm *VM, a Value) (Value, error) {
		n, err := vm.length(a)
		if err != nil {
			return Value{}, err
		}
		return NewValue(n), nil
	})
)

// ========== 控制流处理器 ==========

// handleJump 处理无条件跳转指令
func handleJump(vm *VM, inst Instruction, frame *Frame) error {
	frame.IP = inst.Args[0] - 1
	return nil
}

// makeJumpHandler 创建条件跳转处理器
func makeJumpHandler(jumpOnTrue bool) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		if vm.pop().IsTruthy() == jumpOnTrue {
			frame.IP = inst.Args[0] - 1
		}
		return nil
	}
}

var (
	handleJumpIfFalse = makeJumpHandler(false)
	handleJumpIfTrue  = makeJumpHandler(true)
)

// ========== 函数操作处理器 ==========

// handleCall 处理函数调用指令
// 栈布局: [函数, arg1, arg2, ...] -> 调用后栈顶为返回值
func handleCall(vm *VM, inst Instruction, frame *Frame) error {
	numArgs := inst.Args[0]

	// 检查栈是否有足够的元素
	if vm.SP < numArgs+1 {
		return vm.runtimeErrorWithCode(ErrUnsupportedOp, runtimeErrorTemplates.FuncNotSupported)
	}

	// 先弹出参数(栈顶是最后一个参数)
	args := vm.popArgs(numArgs)

	// 再弹出函数对象
	fnValue := vm.pop()

	// 检查是否是函数类型
	if fnValue.Type != TypeFunction {
		return vm.runtimeErrorWithCode(ErrUnsupportedOp, runtimeErrorTemplates.FuncNotSupported)
	}

	// 获取函数对象
	fn := fnValue.Function()
	if fn == nil || fn.Compiled == nil {
		return vm.runtimeErrorWithCode(ErrUnsupportedOp, runtimeErrorTemplates.FuncNotSupported)
	}

	compiledFn := fn.Compiled

	// 检查参数数量
	if len(args) != compiledFn.NumParams {
		return vm.runtimeError("参数数量不匹配：函数 %s 期望 %d 个参数，但得到 %d 个",
			compiledFn.Name, compiledFn.NumParams, len(args))
	}

	// 检查调用栈深度，防止无限递归导致内存溢出
	if vm.FP+1 >= vm.MaxDepth {
		return vm.runtimeErrorWithCode(ErrCallStackOverflow, runtimeErrorTemplates.CallStackOverflow, vm.FP+1, vm.MaxDepth)
	}

	// 在压入新帧之前，递增调用者帧的IP
	// 这样返回后会执行调用指令的下一条指令
	frame.IP++

	// 创建新的帧
	newFrame := NewFrame(compiledFn)

	// 将参数复制到新帧的局部变量
	for i, arg := range args {
		newFrame.Locals[i] = arg
	}

	// 压入调用栈
	vm.Frames = append(vm.Frames, newFrame)
	vm.FP++

	return nil
}

// getExternalFunc 从脚本中获取外部函数定义
func (vm *VM) getExternalFunc(extIdx int) (ExternalFunc, error) {
	if vm.script == nil || extIdx >= len(vm.script.Externals) {
		return ExternalFunc{}, vm.runtimeError(runtimeErrorTemplates.ExternalFuncInvalid, extIdx)
	}
	return vm.script.Externals[extIdx], nil
}

// popArgs 从栈上弹出指定数量的参数
func (vm *VM) popArgs(count int) []Value {
	args := make([]Value, count)
	for i := count - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	return args
}

// lookupBindFunc 查找绑定的外部函数
// 如果找不到，返回带有详细错误信息的运行时错误
func (vm *VM) lookupBindFunc(funcName string) (interface{}, error) {
	fn, ok := vm.Context.GetBindFunc(funcName)
	if !ok {
		return nil, vm.runtimeError(runtimeErrorTemplates.BindFuncNotFound,
			funcName, funcName, funcName)
	}
	return fn, nil
}

// callBindFunc 执行外部函数调用并处理结果
func (vm *VM) callBindFunc(funcName string, fn interface{}, args []Value) error {
	result, err := callExternalFunc(fn, args)
	if err != nil {
		return vm.runtimeError(runtimeErrorTemplates.ExternalFuncFailed, funcName, err)
	}
	vm.push(result)
	return nil
}

// handleCallBind 处理外部绑定函数调用指令
// 拆分为多个子函数以提高可读性和可维护性
func handleCallBind(vm *VM, inst Instruction, frame *Frame) error {
	// 获取外部函数定义
	extFunc, err := vm.getExternalFunc(inst.Args[0])
	if err != nil {
		return err
	}

	// 查找绑定的函数
	fn, err := vm.lookupBindFunc(extFunc.Name)
	if err != nil {
		return err
	}

	// 弹出参数并调用函数
	args := vm.popArgs(inst.Args[1])
	return vm.callBindFunc(extFunc.Name, fn, args)
}

// handleReturn 处理函数返回指令
// 栈顶的值作为返回值,弹出当前帧并将返回值压入调用者栈
func handleReturn(vm *VM, inst Instruction, frame *Frame) error {
	// 保存返回值(栈顶元素)
	var returnValue Value
	if vm.SP > 0 {
		returnValue = vm.pop()
	} else {
		returnValue = NewValue(nil)
	}

	// 弹出当前帧
	if vm.FP > 0 {
		vm.Frames[vm.FP] = nil
		vm.Frames = vm.Frames[:vm.FP]
		vm.FP--
	} else {
		// 主帧返回，停止VM执行
		vm.stop()
	}

	// 将返回值压入调用者栈
	vm.push(returnValue)

	return nil
}

// ========== 数组操作处理器 ==========

// handleArrayNew 处理数组创建指令
func handleArrayNew(vm *VM, inst Instruction, frame *Frame) error {
	count := inst.Args[0]
	elements := make([]Value, count)
	for i := count - 1; i >= 0; i-- {
		elements[i] = vm.pop()
	}
	vm.push(NewValue(elements))
	return nil
}

// ========== Map操作处理器 ==========

// handleMapNew 处理Map创建指令
// 优化：预分配Map容量，减少扩容开销
func handleMapNew(vm *VM, inst Instruction, frame *Frame) error {
	count := inst.Args[0]
	m := &MapValue{
		Pairs:   make(map[string]Value, count), // 预分配容量
		KeyType: TypeString,
	}
	// 栈顶是最后一对, 先出现的key如果已存在则跳过
	// 这样后出现的重复key会优先保留(与主流语言一致)
	for i := 0; i < count; i++ {
		val := vm.pop()
		key := vm.pop()
		keyStr := key.String()
		if _, exists := m.Pairs[keyStr]; !exists {
			m.Pairs[keyStr] = val
		}
	}
	vm.push(NewValue(m))
	return nil
}

// ========== 栈操作处理器 ==========

// handlePop 处理栈弹出指令
func handlePop(vm *VM, inst Instruction, frame *Frame) error {
	vm.pop()
	return nil
}

// handleDup 处理栈复制指令
func handleDup(vm *VM, inst Instruction, frame *Frame) error {
	val := vm.pop()
	vm.push(val)
	vm.push(val)
	return nil
}

// ========== 异常处理处理器 ==========

// handleThrow 处理异常抛出指令
func handleThrow(vm *VM, inst Instruction, frame *Frame) error {
	msg := vm.pop()
	return vm.runtimeError("%s", msg.GoString())
}

// ========== 内置函数处理器 ==========

// handleBuiltinPush 处理 push(arr, elem) 操作
func handleBuiltinPush(vm *VM, inst Instruction, frame *Frame) error {
	elem := vm.pop()
	arr := vm.pop()
	result, err := vm.builtinPush(arr, elem)
	if err != nil {
		return err
	}
	vm.push(result)
	return nil
}

// handleBuiltinDelete 处理 delete(map, key) 操作
// 从map中删除指定的key，返回nil
func handleBuiltinDelete(vm *VM, inst Instruction, frame *Frame) error {
	// 弹出栈上的key和map
	key := vm.pop()
	mapVal := vm.pop()

	// 检查第一个参数是否为map类型
	if mapVal.Type != TypeMap {
		return vm.runtimeError("运行时错误：delete() 第一个参数必须是map类型，但传入了 %s 类型。\n"+
			"→ 正确用法：delete(map, key)\n"+
			"→ 示例：delete(m, \"key\")",
			typeName(mapVal.Type))
	}

	// 从map中删除指定的key
	m := mapVal.Map()
	if m != nil && m.Pairs != nil {
		delete(m.Pairs, key.String())
	}

	// 返回nil
	vm.push(precomputedNil)
	return nil
}

// handleMapKeys 提取Map的key数组
// 栈顶弹出map, 压入key字符串数组
func handleMapKeys(vm *VM, inst Instruction, frame *Frame) error {
	mapVal := vm.pop()
	if mapVal.Type != TypeMap {
		return vm.runtimeError("运行时错误：mapKeys() 参数必须是map类型，但传入了 %s 类型。",
			typeName(mapVal.Type))
	}
	m := mapVal.Map()
	if m == nil || m.Pairs == nil {
		vm.push(NewValue([]Value{}))
		return nil
	}
	keys := make([]Value, 0, len(m.Pairs))
	for k := range m.Pairs {
		keys = append(keys, NewValue(k))
	}
	vm.push(NewValue(keys))
	return nil
}

// formatValueForOutput 格式化值用于print/println输出
// 字符串类型直接输出不加引号, 其他类型用GoString
func formatValueForOutput(v Value) string {
	if v.Type == TypeString {
		return v.String()
	}
	return v.GoString()
}

// makePrintHandler 创建print/println处理器
func makePrintHandler(newline bool) opcodeHandler {
	return func(vm *VM, inst Instruction, frame *Frame) error {
		count := inst.Args[0]
		args := vm.popArgs(count)
		parts := make([]string, count)
		for i, arg := range args {
			parts[i] = formatValueForOutput(arg)
		}
		output := vm.Context.GetOutput()
		if output != nil {
			if newline {
				fmt.Fprintln(output, strings.Join(parts, " "))
			} else {
				fmt.Fprint(output, strings.Join(parts, " "))
			}
		}
		vm.push(precomputedNil)
		return nil
	}
}

// ========== 操作码注册表 ==========

// opcodeRegistry 操作码注册表
var opcodeRegistry = []struct {
	op OpCode
	h  opcodeHandler
}{
	// 常量与变量
	{OpConst, handleConst},
	{OpNil, handleNil},
	{OpTrue, handleTrue},
	{OpFalse, handleFalse},
	{OpLoadLocal, handleLoadLocal},
	{OpStoreLocal, handleStoreLocal},
	{OpLoadBind, handleLoadBind},
	// 数组和Map
	{OpIndex, handleIndex},
	{OpSlice, handleSlice},
	{OpStoreIndex, handleStoreIndex},
	{OpArrayNew, handleArrayNew},
	{OpMapNew, handleMapNew},
	// 算术运算
	{OpAdd, handleAdd},
	{OpSub, handleSub},
	{OpMul, handleMul},
	{OpDiv, handleDiv},
	{OpMod, handleMod},
	{OpNeg, handleNeg},
	// 位运算
	{OpBitAnd, handleBitAnd},
	{OpBitOr, handleBitOr},
	{OpBitXor, handleBitXor},
	{OpBitNot, handleBitNot},
	{OpLShift, handleLShift},
	{OpRShift, handleRShift},
	// 比较运算
	{OpEqual, handleEqual},
	{OpNotEqual, handleNotEqual},
	{OpLess, handleLess},
	{OpLessEq, handleLessEq},
	{OpGreater, handleGreater},
	{OpGreaterEq, handleGreaterEq},
	// 逻辑运算
	{OpNot, handleNot},
	// 类型转换
	{OpToInt, handleToInt},
	{OpToFloat, handleToFloat},
	{OpToString, handleToString},
	{OpTypeOf, handleTypeOf},
	{OpLen, handleLen},
	// 控制流
	{OpJump, handleJump},
	{OpJumpIfFalse, handleJumpIfFalse},
	{OpJumpIfTrue, handleJumpIfTrue},
	// 函数操作
	{OpCall, handleCall},
	{OpCallBind, handleCallBind},
	{OpReturn, handleReturn},
	// 栈操作
	{OpPop, handlePop},
	{OpDup, handleDup},
	// 内置函数
	{OpPush, handleBuiltinPush},
	{OpDelete, handleBuiltinDelete},
	{OpMapKeys, handleMapKeys},
	{OpPrint, makePrintHandler(false)},
	{OpPrintln, makePrintHandler(true)},
	// 异常处理
	{OpThrow, handleThrow},
}

// init 初始化所有操作码处理器
func init() {
	for _, reg := range opcodeRegistry {
		opcodeHandlers[reg.op] = reg.h
	}
}
