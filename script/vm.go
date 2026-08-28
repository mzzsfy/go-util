package script

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ========== 操作码定义 ==========

// OpCode 操作码类型
type OpCode byte

const (
	// OpConst 加载常量到栈顶，参数为常量池索引
	OpConst OpCode = iota
	// OpNil 加载nil值到栈顶
	OpNil
	// OpTrue 加载true值到栈顶
	OpTrue
	// OpFalse 加载false值到栈顶
	OpFalse
	// OpLoadLocal 加载局部变量到栈顶，参数为变量索引
	OpLoadLocal
	// OpStoreLocal 将栈顶值存储到局部变量，参数为变量索引
	OpStoreLocal
	// OpLoadBind 加载绑定值到栈顶
	OpLoadBind
	// OpIndex 执行索引访问 arr[index]
	OpIndex
	// OpSlice 执行切片操作 arr[start:end]
	OpSlice
	// OpStoreIndex 执行索引赋值 arr[index] = value
	OpStoreIndex
	// OpAdd 执行加法运算
	OpAdd
	// OpSub 执行减法运算
	OpSub
	// OpMul 执行乘法运算
	OpMul
	// OpDiv 执行除法运算
	OpDiv
	// OpMod 执行取模运算
	OpMod
	// OpNeg 执行负号运算
	OpNeg
	// OpBitAnd 执行位与运算
	OpBitAnd
	// OpBitOr 执行位或运算
	OpBitOr
	// OpBitXor 执行位异或运算
	OpBitXor
	// OpBitNot 执行位取反运算
	OpBitNot
	// OpLShift 执行左移运算
	OpLShift
	// OpRShift 执行右移运算
	OpRShift
	// OpEqual 执行相等比较
	OpEqual
	// OpNotEqual 执行不等比较
	OpNotEqual
	// OpLess 执行小于比较
	OpLess
	// OpLessEq 执行小于等于比较
	OpLessEq
	// OpGreater 执行大于比较
	OpGreater
	// OpGreaterEq 执行大于等于比较
	OpGreaterEq
	// OpNot 执行逻辑非运算
	OpNot
	// OpToInt 将值转换为整数
	OpToInt
	// OpToFloat 将值转换为浮点数
	OpToFloat
	// OpToString 将值转换为字符串
	OpToString
	// OpTypeOf 获取值的类型
	OpTypeOf
	// OpJump 无条件跳转，参数为目标指令偏移
	OpJump
	// OpJumpIfFalse 条件为假时跳转
	OpJumpIfFalse
	// OpJumpIfTrue 条件为真时跳转
	OpJumpIfTrue
	// OpCall 调用函数
	OpCall
	// OpCallBind 调用绑定函数
	OpCallBind
	// OpReturn 从函数返回
	OpReturn
	// OpArrayNew 创建新数组
	OpArrayNew
	// OpLen 获取长度
	OpLen
	// OpMapNew 创建新Map
	OpMapNew
	// OpThrow 抛出异常
	OpThrow
	// OpPop 弹出栈顶值
	OpPop
	// OpDup 复制栈顶值
	OpDup
	// OpPush 向数组追加元素
	OpPush
	// OpDelete 从Map中删除key
	OpDelete
	// OpMapKeys 提取Map的key数组
	OpMapKeys
	// OpPrint 输出到output, 参数为参数数量
	OpPrint
	// OpPrintln 输出到output并换行, 参数为参数数量
	OpPrintln
)

// Instruction 字节码指令
// 表示虚拟机可执行的单条指令
type Instruction struct {
	// Op 操作码，定义指令类型
	Op OpCode
	// ArgCount 参数数量（0-2）
	ArgCount uint8
	// Args 指令参数数组，最多2个参数
	Args [2]int
}

// CompiledFunction 编译后的函数
// 包含函数的字节码指令、常量池和局部变量信息
type CompiledFunction struct {
	// Name 函数名称
	Name string
	// Instructions 字节码指令序列
	Instructions []Instruction
	// Constants 函数局部常量池
	Constants []Value
	// NumLocals 局部变量数量
	NumLocals int
	// NumParams 参数数量
	NumParams int
}

// ExternalFunc 外部函数定义
// 用于声明可以从脚本中调用的Go函数
type ExternalFunc struct {
	// Name 函数名称
	Name string
	// Params 参数列表
	Params []Param
	// Return 返回类型
	Return TypeExpr
}

// CompiledScript 编译后的脚本
// 包含所有函数定义、主函数和外部函数声明
type CompiledScript struct {
	// Functions 所有函数定义
	Functions []*CompiledFunction
	// Main 主函数入口
	Main *CompiledFunction
	// Externals 外部函数声明列表
	Externals []ExternalFunc
	// Constants 全局常量池
	Constants []Value
}

// ========== 虚拟机配置常量 ==========

const (
	// DefaultStackSize 默认栈大小
	DefaultStackSize    = 1024
	DefaultMaxCallDepth = 256
	// DefaultMaxSteps 默认最大指令数, 防止死循环无限阻塞
	DefaultMaxSteps = 1000 * 1000 * 10
)

// ========== 虚拟机实例池 ==========

// vmPool 用于复用VM实例，减少内存分配和GC压力
var vmPool = sync.Pool{
	New: func() interface{} {
		return &VM{
			Stack:  make([]Value, 0, DefaultStackSize),
			Frames: make([]*Frame, 0, DefaultMaxCallDepth),
		}
	},
}

// getVMFromPool 从池中获取VM实例
func getVMFromPool(ctx *Context, maxDepth int) *VM {
	vm := vmPool.Get().(*VM)
	vm.Context = ctx
	vm.MaxDepth = maxDepth
	vm.SP = 0
	vm.FP = 0
	return vm
}

// returnVMToPool 将VM实例归还到池中
func returnVMToPool(vm *VM) {
	if vm != nil {
		// 清零实际使用的栈区, 释放Value引用以便GC回收
		for i := 0; i < vm.SP; i++ {
			vm.Stack[i] = Value{}
		}
		vm.Stack = vm.Stack[:0]
		vm.Frames = vm.Frames[:0]
		vm.Context = nil
		vm.script = nil
		atomic.StoreInt32(&vm.timedOut, 0)
		vm.stop()
		vmPool.Put(vm)
	}
}

// Frame 调用栈帧
// 每次函数调用都会创建一个新的栈帧
type Frame struct {
	// Function 当前执行的函数
	Function *CompiledFunction
	// IP 指令指针，指向下一条要执行的指令
	IP int
	// Locals 局部变量存储
	Locals []Value
}

// NewFrame 创建新的调用栈帧
func NewFrame(fn *CompiledFunction) *Frame {
	return &Frame{
		Function: fn,
		IP:       0,
		Locals:   make([]Value, fn.NumLocals),
	}
}

// VM 虚拟机
type VM struct {
	// Stack 操作数栈
	Stack []Value
	// SP 栈指针
	SP int
	// Frames 调用栈
	Frames []*Frame
	// FP 帧指针
	FP int
	// Context 运行时上下文
	Context *Context
	// MaxDepth 最大调用深度
	MaxDepth int
	// running 运行状态(原子操作)
	running int32
	// timedOut 超时标志(原子操作), timer 设置后 execute 退出时返回超时错误
	timedOut int32
	// script 当前执行的脚本（用于外部函数调用）
	script *CompiledScript
	// mainFrame 预分配的主帧，避免每次Run分配
	mainFrame Frame
	// stepCount 已执行的指令数
	stepCount int64
	// maxSteps 最大指令数, 0 表示无限制
	maxSteps int64
}

// NewVM 创建虚拟机
// 注意：推荐使用getVMFromPool以获得更好的性能
func NewVM(ctx *Context, maxDepth int) *VM {
	return &VM{
		Stack:    make([]Value, 0, DefaultStackSize),
		Frames:   make([]*Frame, 0, maxDepth),
		Context:  ctx,
		MaxDepth: maxDepth,
	}
}

// Run 执行脚本
func (vm *VM) Run(script *CompiledScript) (Value, error) {
	// 重置执行状态,确保多次调用安全
	vm.SP = 0
	vm.Frames = vm.Frames[:0]
	// 设置运行状态
	atomic.StoreInt32(&vm.running, 1)
	vm.script = script

	// 复用预分配的主帧，避免每次Run都分配Frame和Locals
	n := script.Main.NumLocals
	if cap(vm.mainFrame.Locals) < n {
		vm.mainFrame.Locals = make([]Value, n)
	} else {
		vm.mainFrame.Locals = vm.mainFrame.Locals[:n]
	}
	// 清零Locals
	for i := range vm.mainFrame.Locals {
		vm.mainFrame.Locals[i] = Value{}
	}
	vm.mainFrame.Function = script.Main
	vm.mainFrame.IP = 0

	vm.Frames = append(vm.Frames, &vm.mainFrame)
	vm.FP = 0

	// 执行指令循环
	err := vm.execute(script)

	// 显式清理，避免defer闭包分配
	vm.stop()
	vm.script = nil

	if err != nil {
		return Value{}, err
	}

	// 返回栈顶值
	if vm.SP > 0 {
		return vm.Stack[vm.SP-1], nil
	}

	return NewValue(nil), nil
}

// execute 执行指令循环（优化版）
// 内联最高频指令和常见算术运算以减少函数调用开销
func (vm *VM) execute(script *CompiledScript) error {
	for vm.FP >= 0 && vm.isRunning() {
		// 指令计数, 仅在配置了上限时检查
		if vm.maxSteps > 0 {
			vm.stepCount++
			if vm.stepCount > vm.maxSteps {
				return vm.runtimeErrorWithCode(ErrStepLimit, "运行时错误：脚本执行超过指令数上限（%d 条指令）。\n"+
					"→ 可能原因：死循环或计算量过大\n"+
					"→ 建议：使用 WithMaxSteps() 调整上限或优化脚本",
					vm.maxSteps)
			}
		}

		frame := vm.Frames[vm.FP]
		if frame.IP >= len(frame.Function.Instructions) {
			break
		}

		inst := frame.Function.Instructions[frame.IP]

		// 内联最高频指令以提升性能
		// 这些指令占据了大部分执行时间
		switch inst.Op {
		case OpConst:
			// 常量加载 - 最高频指令
			vm.push(frame.Function.Constants[inst.Args[0]])

		case OpLoadLocal:
			// 局部变量加载 - 高频指令
			vm.push(frame.Locals[inst.Args[0]])

		case OpStoreLocal:
			// 局部变量存储 - 高频指令
			frame.Locals[inst.Args[0]] = vm.pop()

		case OpAdd:
			// 加法 - 内联int+int，用intVal避免装箱
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt {
				vm.push(intVal(a.Data.(int) + b.Data.(int)))
			} else {
				result, err := vm.add(a, b)
				if err != nil {
					return err
				}
				vm.push(result)
			}

		case OpSub:
			// 减法 - 内联int-int
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt {
				vm.push(intVal(a.Data.(int) - b.Data.(int)))
			} else {
				result, err := vm.sub(a, b)
				if err != nil {
					return err
				}
				vm.push(result)
			}

		case OpMul:
			// 乘法 - 内联int*int
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt {
				vm.push(intVal(a.Data.(int) * b.Data.(int)))
			} else {
				result, err := vm.mul(a, b)
				if err != nil {
					return err
				}
				vm.push(result)
			}

		case OpDiv:
			// 除法 - 内联int/int（除零保护）
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt && b.Data.(int) != 0 {
				vm.push(intVal(a.Data.(int) / b.Data.(int)))
			} else {
				result, err := vm.div(a, b)
				if err != nil {
					return err
				}
				vm.push(result)
			}

		case OpMod:
			// 取模 - 内联int%int（除零保护）
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt && b.Data.(int) != 0 {
				vm.push(intVal(a.Data.(int) % b.Data.(int)))
			} else {
				result, err := vm.mod(a, b)
				if err != nil {
					return err
				}
				vm.push(result)
			}

		case OpLess:
			// 小于比较 - 内联int<int
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt {
				vm.push(NewValue(a.Data.(int) < b.Data.(int)))
			} else {
				result, err := vm.less(a, b)
				if err != nil {
					return err
				}
				vm.push(NewValue(result))
			}

		case OpGreater:
			// 大于比较 - 内联int>int
			b := vm.pop()
			a := vm.pop()
			if a.Type == TypeInt && b.Type == TypeInt {
				vm.push(NewValue(a.Data.(int) > b.Data.(int)))
			} else {
				result, err := vm.greater(a, b)
				if err != nil {
					return err
				}
				vm.push(NewValue(result))
			}

		case OpEqual:
			// 相等比较 - 高频指令，内联优化
			b := vm.pop()
			a := vm.pop()
			// 直接使用Value.Equal()方法
			vm.push(NewValue(a.Equal(b)))

		default:
			// 其他指令使用处理器数组
			if err := vm.executeInstruction(inst, script); err != nil {
				return err
			}
			// 控制流指令(OpCall, OpReturn)会改变帧，需要由处理器自己管理IP
			// 只有非控制流指令才在这里递增IP
			if inst.Op != OpCall && inst.Op != OpReturn {
				frame.IP++
			}
			continue
		}

		// 内联指令递增IP
		frame.IP++
	}

	// 超时检查: timer 设置 timedOut 后 stop VM, 循环退出
	if atomic.LoadInt32(&vm.timedOut) != 0 {
		return vm.runtimeErrorWithCode(ErrTimeout, "脚本执行超时")
	}

	return nil
}

// executeInstruction 执行单条指令
// 使用处理器数组模式降低复杂度
func (vm *VM) executeInstruction(inst Instruction, script *CompiledScript) error {
	frame := vm.currentFrame()

	handler := opcodeHandlers[inst.Op]
	if handler == nil {
		return vm.runtimeError(runtimeErrorTemplates.UnknownOpcode,
			inst.Op, frame.Function.Name, frame.IP)
	}

	return handler(vm, inst, frame)
}

// currentFrame 获取当前帧
func (vm *VM) currentFrame() *Frame {
	return vm.Frames[vm.FP]
}

// hasMoreFrames 检查是否还有未执行的调用帧
func (vm *VM) hasMoreFrames() bool {
	return vm.FP >= 0
}

// push 压栈
// 快速路径直接索引，慢速路径由append处理容量增长
func (vm *VM) push(val Value) {
	if vm.SP < len(vm.Stack) {
		vm.Stack[vm.SP] = val
	} else {
		vm.Stack = append(vm.Stack, val)
	}
	vm.SP++
}

// pop 出栈
func (vm *VM) pop() Value {
	if vm.SP <= 0 {
		return precomputedNil
	}
	vm.SP--
	return vm.Stack[vm.SP]
}

// isRunning 原子检查运行状态
func (vm *VM) isRunning() bool {
	return atomic.LoadInt32(&vm.running) != 0
}

// stop 原子停止VM
func (vm *VM) stop() {
	atomic.StoreInt32(&vm.running, 0)
}

// buildStackTrace 构建调用堆栈追踪
func (vm *VM) buildStackTrace() []string {
	if len(vm.Frames) == 0 {
		return nil
	}
	trace := make([]string, 0, vm.FP+1)
	for i := vm.FP; i >= 0; i-- {
		frame := vm.Frames[i]
		trace = append(trace, fmt.Sprintf("  at %s (ip:%d)", frame.Function.Name, frame.IP))
	}
	return trace
}

// runtimeError 创建运行时错误
func (vm *VM) runtimeError(format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		Message:    fmt.Sprintf(format, args...),
		StackTrace: vm.buildStackTrace(),
	}
}

// runtimeErrorWithCode 创建带错误码的运行时错误
func (vm *VM) runtimeErrorWithCode(code ErrCode, format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		StackTrace: vm.buildStackTrace(),
	}
}

// typeName 获取类型名称
func typeName(t ValueType) string {
	switch t {
	case TypeNil:
		return "nil"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeBool:
		return "bool"
	case TypeArray:
		return "array"
	case TypeMap:
		return "map"
	case TypeFunction:
		return "function"
	case TypeExternalFunc:
		return "external"
	}
	return "unknown"
}
