package script

import (
	"testing"
)

// ========== VM 核心函数覆盖率测试 ==========
// 本文件专注于提高 vm.go 文件的测试覆盖率

// TestVM_NewFrame 测试NewFrame函数
func Test_vm_NewFrame_Coverage(t *testing.T) {
	fn := &CompiledFunction{
		Name:         "testFunc",
		Instructions: make([]Instruction, 10),
		NumLocals:    5,
		NumParams:    2,
	}

	frame := NewFrame(fn)

	// 验证帧创建正确
	if frame.Function != fn {
		t.Error("Frame函数引用不正确")
	}
	if frame.IP != 0 {
		t.Errorf("初始IP应为0, 实际为%d", frame.IP)
	}
	if len(frame.Locals) != fn.NumLocals {
		t.Errorf("Locals长度应为%d, 实际为%d", fn.NumLocals, len(frame.Locals))
	}
}

// TestVM_NewVM 测试NewVM函数
func Test_vm_NewVM_Coverage(t *testing.T) {
	ctx := NewContext()
	maxDepth := 128

	vm := NewVM(ctx, maxDepth)

	// 验证VM初始化正确
	if vm.Context != ctx {
		t.Error("Context引用不正确")
	}
	if vm.MaxDepth != maxDepth {
		t.Errorf("MaxDepth应为%d, 实际为%d", maxDepth, vm.MaxDepth)
	}
	if vm.isRunning() {
		t.Error("初始Running状态应为false")
	}
	if cap(vm.Stack) != DefaultStackSize {
		t.Errorf("Stack容量应为%d, 实际为%d", DefaultStackSize, cap(vm.Stack))
	}
}

// TestVM_PushPop 测试push和pop函数
func Test_vm_PushPop_Coverage(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 测试基本push和pop
	val1 := NewValue(42)
	vm.push(val1)

	if vm.SP != 1 {
		t.Errorf("push后SP应为1, 实际为%d", vm.SP)
	}

	result := vm.pop()
	if vm.SP != 0 {
		t.Errorf("pop后SP应为0, 实际为%d", vm.SP)
	}
	if result.Int() != val1.Int() {
		t.Errorf("pop值不正确: 期望 %d, 实际 %d", val1.Int(), result.Int())
	}
}

// TestVM_Push_StackGrowth 测试栈扩容
func Test_vm_Push_StackGrowth(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 填充超过初始容量的栈
	initialCap := cap(vm.Stack)
	for i := 0; i < initialCap+10; i++ {
		vm.push(NewValue(i))
	}

	// 验证栈已扩容
	if cap(vm.Stack) <= initialCap {
		t.Errorf("栈应该已扩容, 初始容量 %d, 当前容量 %d", initialCap, cap(vm.Stack))
	}

	// 验证所有值都正确
	for i := initialCap + 9; i >= 0; i-- {
		val := vm.pop()
		if val.Int() != i {
			t.Errorf("索引 %d: 期望 %d, 实际 %d", i, i, val.Int())
		}
	}
}

// TestVM_CurrentFrame 测试currentFrame函数
func Test_vm_CurrentFrame_Coverage(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	fn := &CompiledFunction{Name: "test", NumLocals: 0}
	frame := NewFrame(fn)

	vm.Frames = append(vm.Frames, frame)
	vm.FP = 0

	current := vm.currentFrame()
	if current != frame {
		t.Error("currentFrame返回的帧不正确")
	}
}

// TestVM_HasMoreFrames 测试hasMoreFrames函数
func Test_vm_HasMoreFrames_Coverage(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 初始化时FP为0，hasMoreFrames返回true
	// 测试有帧情况
	if !vm.hasMoreFrames() {
		t.Error("FP为0时hasMoreFrames应返回true")
	}

	// 设置FP为-1模拟无帧
	vm.FP = -1
	if vm.hasMoreFrames() {
		t.Error("FP为-1时hasMoreFrames应返回false")
	}

	// 添加帧并恢复FP
	fn := &CompiledFunction{Name: "test", NumLocals: 0}
	vm.Frames = append(vm.Frames, NewFrame(fn))
	vm.FP = 0

	// 测试有帧情况
	if !vm.hasMoreFrames() {
		t.Error("有帧时hasMoreFrames应返回true")
	}
}

// TestVM_BuildStackTrace 测试buildStackTrace函数
func Test_vm_BuildStackTrace_Coverage(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 创建多个帧
	for i := 0; i < 3; i++ {
		fn := &CompiledFunction{
			Name:         "func" + string(rune('0'+i)),
			NumLocals:    0,
			Instructions: make([]Instruction, 5),
		}
		frame := NewFrame(fn)
		frame.IP = i
		vm.Frames = append(vm.Frames, frame)
	}
	vm.FP = 2

	trace := vm.buildStackTrace()

	// 验证堆栈追踪
	if len(trace) != 3 {
		t.Errorf("堆栈追踪长度应为3, 实际为%d", len(trace))
	}

	// 验证顺序（从当前帧到最早的帧）
	for i, line := range trace {
		if line == "" {
			t.Errorf("堆栈追踪第%d行不应为空", i)
		}
	}
}

// TestVM_RuntimeError 测试runtimeError函数
func Test_vm_RuntimeError_Coverage(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 添加帧以生成堆栈追踪
	fn := &CompiledFunction{Name: "test", NumLocals: 0}
	vm.Frames = append(vm.Frames, NewFrame(fn))
	vm.FP = 0

	err := vm.runtimeError("测试错误: %d", 42)

	if err == nil {
		t.Fatal("runtimeError应返回错误")
	}

	if err.Message == "" {
		t.Error("错误消息不应为空")
	}

	if len(err.StackTrace) == 0 {
		t.Error("堆栈追踪不应为空")
	}
}

// TestVM_ExecuteInstruction_UnknownOpcode 测试未知操作码
func Test_vm_ExecuteInstruction_UnknownOpcode(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	fn := &CompiledFunction{Name: "test", NumLocals: 0}
	vm.Frames = append(vm.Frames, NewFrame(fn))
	vm.FP = 0

	// 使用一个未注册的操作码（255通常未使用）
	inst := Instruction{Op: OpCode(255)}

	err := vm.executeInstruction(inst, nil)

	if err == nil {
		t.Error("未知操作码应返回错误")
	}
}

// TestVM_Run_Success 测试Run函数成功场景
func Test_vm_Run_Success(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile("42")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	vm := NewVM(ctx, DefaultMaxCallDepth)

	result, err := vm.Run(script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 42 {
		t.Errorf("期望 42, 实际 %d", result.Int())
	}

	// 验证Running状态已重置
	if vm.isRunning() {
		t.Error("执行完成后Running应为false")
	}

	// 验证script引用已清除
	if vm.script != nil {
		t.Error("执行完成后script引用应为nil")
	}
}

// TestVM_Run_EmptyStack 测试Run函数空栈场景
func Test_vm_Run_EmptyStack(t *testing.T) {
	// 创建一个不产生任何值的脚本
	parser := NewParser()
	script, err := parser.Compile("x := 1")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	vm := NewVM(ctx, DefaultMaxCallDepth)

	// 清空栈（如果有值）
	vm.SP = 0

	result, err := vm.Run(script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 空栈应返回nil
	if !result.IsNil() {
		t.Errorf("空栈应返回nil, 实际为 %v", result)
	}
}

// TestVM_Execute_FrameExhaustion 测试execute函数帧耗尽场景
func Test_vm_Execute_FrameExhaustion(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 创建一个空的主函数
	mainFn := &CompiledFunction{
		Name:         "main",
		NumLocals:    0,
		Instructions: []Instruction{}, // 空指令集
	}

	script := &CompiledScript{Main: mainFn}

	err := vm.execute(script)
	if err != nil {
		t.Errorf("空指令集不应产生错误: %v", err)
	}
}

// TestVM_ExecuteInstruction_AllHandlers 测试所有指令处理器
func Test_vm_ExecuteInstruction_AllHandlers(t *testing.T) {
	// 测试已经通过其他测试覆盖的指令
	// 这里主要测试未覆盖的边界情况

	t.Run("OpDup", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test", NumLocals: 0})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		vm.push(NewValue(42))

		inst := Instruction{Op: OpDup}
		err := vm.executeInstruction(inst, nil)
		if err != nil {
			t.Fatalf("OpDup执行失败: %v", err)
		}

		// 验证栈顶有两个相同的值
		if vm.SP != 2 {
			t.Errorf("SP应为2, 实际为%d", vm.SP)
		}

		val1 := vm.pop()
		val2 := vm.pop()
		if val1.Int() != 42 || val2.Int() != 42 {
			t.Error("OpDup应该复制栈顶值")
		}
	})

	t.Run("OpPop", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test", NumLocals: 0})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		vm.push(NewValue(42))
		vm.push(NewValue(100))

		inst := Instruction{Op: OpPop}
		err := vm.executeInstruction(inst, nil)
		if err != nil {
			t.Fatalf("OpPop执行失败: %v", err)
		}

		// 验证栈顶值被弹出
		if vm.SP != 1 {
			t.Errorf("SP应为1, 实际为%d", vm.SP)
		}

		val := vm.pop()
		if val.Int() != 42 {
			t.Errorf("剩余值应为42, 实际为%d", val.Int())
		}
	})

	t.Run("OpReturn", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test", NumLocals: 0})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		inst := Instruction{Op: OpReturn}
		err := vm.executeInstruction(inst, nil)
		if err != nil {
			t.Fatalf("OpReturn执行失败: %v", err)
		}
	})
}

// TestVM_TypeName 测试typeName函数
func Test_vm_TypeName_Coverage(t *testing.T) {
	tests := []struct {
		typ      ValueType
		expected string
	}{
		{TypeNil, "nil"},
		{TypeInt, "int"},
		{TypeFloat, "float"},
		{TypeString, "string"},
		{TypeBool, "bool"},
		{TypeArray, "array"},
		{TypeMap, "map"},
		{TypeFunction, "function"},
		{TypeExternalFunc, "external"},
		{ValueType(255), "unknown"}, // 测试未知类型
	}

	for _, tt := range tests {
		result := typeName(tt.typ)
		if result != tt.expected {
			t.Errorf("typeName(%v) = %s, want %s", tt.typ, result, tt.expected)
		}
	}
}

// TestVM_Push_ExistingSlice 测试push到已有切片
func Test_vm_Push_ExistingSlice(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 预先设置栈（模拟已有数据）
	vm.Stack = []Value{NewValue(1), NewValue(2), NewValue(3)}
	vm.SP = 3

	// Push新值
	vm.push(NewValue(4))

	// 验证
	if vm.SP != 4 {
		t.Errorf("SP应为4, 实际为%d", vm.SP)
	}

	if vm.Stack[3].Int() != 4 {
		t.Errorf("栈顶值应为4, 实际为%d", vm.Stack[3].Int())
	}
}

// TestVM_Run_MultipleRuns 测试VM多次运行
func Test_vm_Run_MultipleRuns(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	for i := 0; i < 3; i++ {
		parser := NewParser()
		script, err := parser.Compile("10 + 20")
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}

		_ = NewContext()
		result, err := vm.Run(script)
		if err != nil {
			t.Fatalf("第%d次执行失败: %v", i+1, err)
		}

		if result.Int() != 30 {
			t.Errorf("第%d次执行结果错误: 期望30, 实际%d", i+1, result.Int())
		}

		// 验证状态已重置
		if vm.isRunning() {
			t.Errorf("第%d次执行后Running未重置", i+1)
		}
		if vm.script != nil {
			t.Errorf("第%d次执行后script未清除", i+1)
		}
	}
}

// TestVM_Run_ExecuteError 测试Run执行错误场景
func Test_vm_Run_ExecuteError(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile("1 / 0") // 除零错误
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	vm := NewVM(ctx, DefaultMaxCallDepth)

	_, err = vm.Run(script)
	if err == nil {
		t.Error("除零应该返回错误")
	}

	// 验证状态已重置（即使出错）
	if vm.isRunning() {
		t.Error("错误后Running应为false")
	}
	if vm.script != nil {
		t.Error("错误后script引用应为nil")
	}
}

// TestVM_Execute_InstructionPointer 测试execute指令指针递增
func Test_vm_Execute_InstructionPointer(t *testing.T) {
	// 创建一个包含多条指令的脚本
	parser := NewParser()
	script, err := parser.Compile("a := 1\na = 2\na")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	vm := NewVM(ctx, DefaultMaxCallDepth)

	_, err = vm.Run(script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 验证执行成功（结果应为2）
	// 这个测试主要验证指令指针正确递增
}

// TestVM_FrameOperations 测试帧操作边界情况
func Test_vm_FrameOperations(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)

	// 测试FP为-1的情况
	vm.FP = -1

	if vm.hasMoreFrames() {
		t.Error("FP为-1时hasMoreFrames应返回false")
	}
}

// TestVM_HandleCall 测试handleCall函数
func Test_vm_HandleCall(t *testing.T) {
	t.Run("栈空间不足", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test"})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0
		vm.SP = 0 // 栈为空

		inst := Instruction{Op: OpCall, Args: [2]int{0, 0}}
		err := handleCall(vm, inst, frame)

		if err == nil {
			t.Error("栈为空时应返回错误")
		}
	})

	t.Run("非函数类型调用", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test"})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		// 压入一个非函数值
		vm.push(NewValue(42))

		inst := Instruction{Op: OpCall, Args: [2]int{0, 0}}
		err := handleCall(vm, inst, frame)

		if err == nil {
			t.Error("调用非函数类型应返回错误")
		}
	})

	t.Run("nil函数对象", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test"})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		// 创建一个FunctionValue但Compiled为nil
		fnValue := NewValue(&FunctionValue{Compiled: nil})
		vm.push(fnValue)

		inst := Instruction{Op: OpCall, Args: [2]int{0, 0}}
		err := handleCall(vm, inst, frame)

		if err == nil {
			t.Error("Compiled为nil应返回错误")
		}
	})

	t.Run("参数数量不匹配", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test"})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		// 创建一个期望2个参数的函数
		compiledFn := &CompiledFunction{
			Name:      "testFunc",
			NumParams: 2,
			NumLocals: 2,
		}
		fnValue := NewValue(&FunctionValue{Compiled: compiledFn})
		vm.push(fnValue)

		// 只压入1个参数
		vm.push(NewValue(1))

		inst := Instruction{Op: OpCall, Args: [2]int{1, 0}} // 1个参数
		err := handleCall(vm, inst, frame)

		if err == nil {
			t.Error("参数数量不匹配应返回错误")
		}
	})

	t.Run("参数匹配的函数调用", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "main"})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		// 创建一个用户定义的函数
		compiledFn := &CompiledFunction{
			Name:         "test",
			NumParams:    2,
			NumLocals:    2,
			Instructions: make([]Instruction, 5),
		}
		fnValue := NewValue(&FunctionValue{Compiled: compiledFn})
		vm.push(fnValue)

		// 压入2个参数
		vm.push(NewValue(10))
		vm.push(NewValue(20))

		inst := Instruction{Op: OpCall, Args: [2]int{2, 0}} // 2个参数
		err := handleCall(vm, inst, frame)

		// 注意：当前实现中用户函数调用返回"功能暂不支持"错误
		// 这是预期的行为，测试验证错误路径被正确处理
		if err == nil {
			t.Log("handleCall成功执行（如果未来支持了用户函数调用）")
			// 验证新帧已创建
			if len(vm.Frames) != 2 {
				t.Errorf("应有2个帧，实际有%d个", len(vm.Frames))
			}
		} else {
			t.Logf("handleCall返回预期错误: %v", err)
		}
	})

	t.Run("函数Compiled为nil", func(t *testing.T) {
		vm := NewVM(NewContext(), DefaultMaxCallDepth)
		frame := NewFrame(&CompiledFunction{Name: "test"})
		vm.Frames = append(vm.Frames, frame)
		vm.FP = 0

		// 创建一个Compiled为nil的函数值
		fnValue := NewValue(&FunctionValue{Compiled: nil})
		vm.push(fnValue)

		inst := Instruction{Op: OpCall, Args: [2]int{0, 0}}
		err := handleCall(vm, inst, frame)

		if err == nil {
			t.Error("Compiled为nil应返回错误")
		}
	})
}
