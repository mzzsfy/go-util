package script

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// Context 运行时上下文
// 提供脚本执行时的运行时环境，包含变量绑定、函数绑定、日志输出和执行统计等功能
// 线程安全，支持并发访问
type Context struct {
	// mu 读写锁，保护并发访问
	mu sync.RWMutex
	// bindValues 绑定的值映射
	bindValues map[string]Value
	// bindFuncs 绑定的函数映射
	bindFuncs map[string]interface{}
	// logFunc 日志输出函数
	logFunc func(format string, args ...any)
	// output 标准输出目标
	output io.Writer
	// stats 执行统计信息
	stats *ExecStats
}

// ExecStats 执行统计
type ExecStats struct {
	// TotalRuns 总执行次数
	TotalRuns int64
	// MaxCallDepth 最大调用深度
	MaxCallDepth int
	// TotalTime 总执行时间，单位为纳秒
	TotalTime int64
}

// NewContext 创建运行时上下文
func NewContext() *Context {
	return &Context{
		bindValues: make(map[string]Value),
		bindFuncs:  make(map[string]interface{}),
		logFunc:    func(format string, args ...any) {},
		output:     os.Stdout,
		stats:      &ExecStats{},
	}
}

// withLock 提供通用的写锁操作封装
func (c *Context) withLock(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn()
}

// withRLock 提供通用的读锁操作封装
func (c *Context) withRLock(fn func()) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fn()
}

// copyStringMap 复制字符串映射的通用函数
func copyStringMap[T any](src map[string]T) map[string]T {
	dst := make(map[string]T, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// BindValue 绑定值
// 自动将Go值转换为脚本Value类型
func (c *Context) BindValue(name string, value any) {
	c.withLock(func() {
		// 如果已经是Value类型，直接使用
		if v, ok := value.(Value); ok {
			c.bindValues[name] = v
		} else {
			// 否则转换后存储
			c.bindValues[name] = convertGoToValue(value)
		}
	})
}

// BindFunc 绑定函数
// 参数name为函数名称，fn为Go函数对象
// 注意：如果fn不是函数类型会触发panic，这是编程错误而非运行时错误
func (c *Context) BindFunc(name string, fn interface{}) {
	// 验证函数类型（编程错误应该尽早暴露）
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panic(fmt.Sprintf("BindFunc参数错误：'%s'绑定的第二个参数必须是函数类型，但传入了%v类型。\n"+
			"正确用法示例：ctx.BindFunc(\"myFunc\", myGoFunc)\n"+
			"错误用法示例：ctx.BindFunc(\"myFunc\", \"not a function\")",
			name, reflect.TypeOf(fn).Kind()))
	}
	c.withLock(func() { c.bindFuncs[name] = fn })
}

// GetBindValue 获取绑定值
func (c *Context) GetBindValue(name string) (Value, bool) {
	var val Value
	var ok bool
	c.withRLock(func() { val, ok = c.bindValues[name] })
	return val, ok
}

// GetBindFunc 获取绑定函数
func (c *Context) GetBindFunc(name string) (interface{}, bool) {
	var fn interface{}
	var ok bool
	c.withRLock(func() { fn, ok = c.bindFuncs[name] })
	return fn, ok
}

// SetLogFunc 设置日志函数
func (c *Context) SetLogFunc(fn func(format string, args ...any)) {
	c.withLock(func() { c.logFunc = fn })
}

// Log 记录日志
func (c *Context) Log(format string, args ...any) {
	var logFunc func(format string, args ...any)
	c.withRLock(func() { logFunc = c.logFunc })
	logFunc(format, args...)
}

// SetOutput 设置输出
func (c *Context) SetOutput(w io.Writer) {
	c.withLock(func() { c.output = w })
}

// GetOutput 获取输出
func (c *Context) GetOutput() io.Writer {
	var output io.Writer
	c.withRLock(func() { output = c.output })
	return output
}

// GetStats 获取执行统计
// 返回统计信息的副本，避免并发访问时的数据竞争
func (c *Context) GetStats() ExecStats {
	var stats ExecStats
	c.withRLock(func() {
		stats = *c.stats
	})
	return stats
}

// Clone 克隆上下文
func (c *Context) Clone() *Context {
	var newCtx *Context
	c.withRLock(func() {
		newCtx = NewContext()
		newCtx.bindValues = copyStringMap(c.bindValues)
		newCtx.bindFuncs = copyStringMap(c.bindFuncs)
		newCtx.logFunc = c.logFunc
		newCtx.output = c.output
	})
	return newCtx
}

// updateStats 更新统计信息
// duration: 本次执行耗时（纳秒）
// depth: 当前调用栈深度
func (c *Context) updateStats(duration int64, depth int) {
	c.withLock(func() {
		c.stats.TotalRuns++
		c.stats.TotalTime += duration
		if depth > c.stats.MaxCallDepth {
			c.stats.MaxCallDepth = depth
		}
	})
}

// ========== 脚本与执行引擎 ==========

// Script 编译后的脚本，包装CompiledScript提供便捷的执行接口
type Script struct {
	// compiled 编译后的脚本数据
	compiled *CompiledScript
}

// NewScript 从编译产物创建脚本实例
// 参数compiled为Parser.Compile()的返回值
func NewScript(compiled *CompiledScript) *Script {
	return &Script{compiled: compiled}
}

// Clone 克隆脚本，创建独立的脚本副本
// 注意：当前实现为浅拷贝，深拷贝功能待实现
func (s *Script) Clone() *Script {
	// TODO: 实现深拷贝
	return &Script{compiled: s.compiled}
}

// Encode 序列化脚本为字节数组
// 用于脚本持久化存储和传输
func (s *Script) Encode() ([]byte, error) {
	// TODO: 实现序列化
	return nil, nil
}

// Decode 从字节数组反序列化脚本
// 与Encode配对使用
func (s *Script) Decode(data []byte) (*Script, error) {
	// TODO: 实现反序列化
	return nil, nil
}

// GetCompiled 获取底层编译产物
// 用于访问编译后的字节码和常量池
func (s *Script) GetCompiled() *CompiledScript {
	return s.compiled
}

// Engine 执行引擎，负责管理脚本执行的上下文和配置
// 提供脚本执行、停止等核心功能
type Engine struct {
	// maxDepth 最大调用栈深度，防止无限递归导致栈溢出
	maxDepth int
	// maxSteps 最大指令数, 0 表示无限制
	maxSteps int64
	// timeout 执行超时, 0 表示无限制
	timeout time.Duration
	// runningVMs 当前正在运行的VM实例映射（用于Stop功能）
	// key: VM指针的字符串表示，value: *VM
	runningVMs sync.Map
}

// EngineOption 引擎配置选项函数类型
// 用于函数式配置模式
type EngineOption func(*Engine)

// WithMaxCallDepth 创建设置最大调用栈深度的配置选项
// 参数depth为允许的最大调用栈帧数
func WithMaxCallDepth(depth int) EngineOption {
	return func(e *Engine) {
		e.maxDepth = depth
	}
}

// WithMaxSteps 设置脚本执行的最大指令数
// n 为 0 或负数时表示无限制
func WithMaxSteps(n int) EngineOption {
	return func(e *Engine) {
		e.maxSteps = int64(n)
	}
}

// WithTimeout 设置脚本执行的超时时间
// d 为 0 时表示无限制
func WithTimeout(d time.Duration) EngineOption {
	return func(e *Engine) {
		e.timeout = d
	}
}

// NewEngine 创建执行引擎实例
// opts: 可选配置项列表，用于自定义引擎行为
// 默认最大调用深度为DefaultMaxCallDepth
func NewEngine(opts ...EngineOption) *Engine {
	engine := &Engine{
		maxDepth: DefaultMaxCallDepth,
	}

	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// Run 执行已编译的脚本
// ctx: 运行时上下文，提供变量绑定和函数绑定
// script: 已编译的脚本对象
// 返回: 脚本执行的返回值和可能的错误
func (e *Engine) Run(ctx *Context, script *CompiledScript) (result Value, err error) {
	startTime := time.Now()

	// 使用VM实例池，减少内存分配
	vm := getVMFromPool(ctx, e.maxDepth)
	defer returnVMToPool(vm)

	// 配置指令计数
	vm.maxSteps = e.maxSteps
	vm.stepCount = 0

	// 注册VM到运行列表（用于Stop功能）
	vmKey := fmt.Sprintf("%p", vm)
	e.runningVMs.Store(vmKey, vm)
	defer e.runningVMs.Delete(vmKey)

	// 超时: 后台 timer 到期后 stop VM
	if e.timeout > 0 {
		timer := time.AfterFunc(e.timeout, func() {
			atomic.StoreInt32(&vm.timedOut, 1)
			vm.stop()
		})
		defer timer.Stop()
	}

	// panic 防护: 外部函数 panic 转为 RuntimeError
	defer func() {
		if r := recover(); r != nil {
			err = &RuntimeError{
				Code:       ErrPanic,
				Message:    fmt.Sprintf("脚本执行期间发生 panic: %v", r),
				StackTrace: vm.buildStackTrace(),
			}
		}
	}()

	result, err = vm.Run(script)
	if err != nil {
		return Value{}, err
	}

	// 更新统计信息
	duration := time.Since(startTime).Nanoseconds()
	ctx.updateStats(duration, vm.FP)

	return result, nil
}

// Stop 停止所有正在运行的脚本执行
// 用于中断长时间运行的脚本
// 线程安全，可以在任何goroutine中调用
func (e *Engine) Stop() {
	// 遍历所有运行中的VM并设置停止标志
	e.runningVMs.Range(func(key, value interface{}) bool {
		if vm, ok := value.(*VM); ok {
			vm.stop()
		}
		return true
	})
}
