package di

import (
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"
)

// signalTestEnv 记录信号注册与退出调用的测试环境
// 替换 signalNotify/osExit,不注册真实信号,不触发进程退出
type signalTestEnv struct {
	t        *testing.T
	sigChan  chan<- os.Signal
	signals  []os.Signal
	exitCode int
	exited   chan struct{}
	once     sync.Once
}

// newSignalTestEnv 替换信号注册与退出函数为记录实现,测试结束自动恢复
// 依赖串行测试执行,不可用于并行测试
func newSignalTestEnv(t *testing.T) *signalTestEnv {
	env := &signalTestEnv{t: t, exited: make(chan struct{})}
	origNotify, origExit := signalNotify, osExit
	t.Cleanup(func() {
		signalNotify, osExit = origNotify, origExit
	})
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		env.sigChan = c
		env.signals = sig
	}
	osExit = func(code int) {
		env.exitCode = code
		env.once.Do(func() { close(env.exited) })
	}
	return env
}

// sendSignal 发送模拟信号
func (e *signalTestEnv) sendSignal(sig os.Signal) {
	e.t.Helper()
	if e.sigChan == nil {
		e.t.Fatal("signalNotify 尚未被调用")
	}
	e.sigChan <- sig
}

// waitExit 等待退出调用并返回退出码
func (e *signalTestEnv) waitExit() int {
	e.t.Helper()
	select {
	case <-e.exited:
		return e.exitCode
	case <-time.After(shutdownHookTimeout):
		e.t.Fatal("osExit 未在时限内被调用")
		return -1
	}
}

// assertSignals 断言注册的信号列表
func (e *signalTestEnv) assertSignals(want ...os.Signal) {
	e.t.Helper()
	if !reflect.DeepEqual(e.signals, want) {
		e.t.Errorf("期望注册信号 %v,实际 %v", want, e.signals)
	}
}

// assertShutdownComplete 断言退出码为 0 且容器已关闭
func (e *signalTestEnv) assertShutdownComplete(c Container) {
	e.t.Helper()
	if code := e.waitExit(); code != 0 {
		e.t.Errorf("期望退出码 0,实际 %d", code)
	}
	select {
	case <-c.Done():
	default:
		e.t.Error("收到信号后容器应已关闭")
	}
}

// 测试 ShutdownOnSignals 收到信号后关闭容器并退出
func TestShutdownOnSignals(t *testing.T) {
	c := New()

	// 注册一个服务
	_ = ProvideValue(c, "test-service")

	env := newSignalTestEnv(t)
	c.ShutdownOnSignals()

	env.sendSignal(os.Interrupt)
	env.assertShutdownComplete(c)
}

func TestShutdownOnSignalsWithCustomSignals(t *testing.T) {
	c := New()

	env := newSignalTestEnv(t)
	c.ShutdownOnSignals(os.Interrupt)

	env.assertSignals(os.Interrupt)

	env.sendSignal(os.Interrupt)
	env.assertShutdownComplete(c)
}

func TestShutdownOnSignalsEmpty(t *testing.T) {
	c := New()

	env := newSignalTestEnv(t)
	c.ShutdownOnSignals()

	env.assertSignals(syscall.SIGTERM, os.Interrupt)

	env.sendSignal(syscall.SIGTERM)
	env.assertShutdownComplete(c)
}

// 测试 GetAllInstances
func Test_GetAllInstances(t *testing.T) {
	c := New()

	// 初始应该为空
	instances := c.GetAllInstances()
	if len(instances) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(instances))
	}

	// 注册并获取一些服务（使用命名）
	_ = ProvideValueNamed(c, "service1", 1)
	_ = ProvideValueNamed(c, "service2", 2)

	_, _ = GetNamed[int](c, "service1")
	_, _ = GetNamed[int](c, "service2")

	// 现在应该有实例
	instances = c.GetAllInstances()
	if len(instances) < 2 {
		t.Errorf("Expected at least 2 instances, got %d", len(instances))
	}
}

// 测试 updateGetCallsStats
func Test_UpdateGetCallsStats(t *testing.T) {
	c := New().(*container) // 类型断言

	_ = ProvideValueNamed(c, "test", 42)

	// 初始统计
	c.updateGetCallsStats()
	stats := c.GetStats()
	initialCalls := stats.GetCalls

	// 再次调用
	c.updateGetCallsStats()

	stats = c.GetStats()
	if stats.GetCalls <= initialCalls {
		t.Errorf("Expected GetCalls to increase, got %d -> %d", initialCalls, stats.GetCalls)
	}
}

// 测试 prepareLazyDependencies 的错误路径
func Test_PrepareLazyDependenciesError(t *testing.T) {
	c := New()

	// 注册一个懒加载服务，依赖不存在的服务
	err := ProvideNamed(c, "lazy-service", func(c Container) (int, error) {
		// 尝试获取不存在的依赖
		dep, err := GetNamed[float64](c, "nonexistent")
		if err != nil {
			return 0, err
		}
		return int(dep), nil
	}, WithLoadMode(LoadModeLazy))

	if err != nil {
		t.Fatalf("ProvideNamed failed: %v", err)
	}

	// 尝试获取懒加载服务（应该失败，因为依赖不存在）
	_, err = GetNamed[int](c, "lazy-service")
	if err == nil {
		t.Error("Expected error when getting lazy service with missing dependency")
	}
}

// 测试 HasNamed 的更多场景
func Test_HasNamedAdvanced(t *testing.T) {
	c := New()

	// 测试不存在的服务
	if c.HasNamed(reflect.TypeOf(int(0)), "nonexistent") {
		t.Error("Expected HasNamed to return false for non-existent provider")
	}

	// 注册服务
	_ = ProvideValueNamed(c, "default", 42)
	_ = ProvideValueNamed(c, "named", 100)

	// 测试存在的服务
	if !c.HasNamed(reflect.TypeOf(int(0)), "default") {
		t.Error("Expected HasNamed to return true for default service")
	}

	if !c.HasNamed(reflect.TypeOf(int(0)), "named") {
		t.Error("Expected HasNamed to return true for named service")
	}
}

// 测试 findDepend 的更多场景
func Test_FindDependAdvanced(t *testing.T) {
	c := New().(*container) // 类型断言

	// 测试查找不存在的依赖
	_, err := c.findDepend(reflect.TypeOf(""))
	if err == nil {
		t.Error("Expected error when finding non-existent dependency")
	}
}

// 测试 ConvertStringToUint
func Test_ConvertStringToUint(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"999999", 999999, false},
		{"-1", 0, true}, // 负数应该失败
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// 创建一个 uint64 变量并获取其可设置的 Value
			var val uint64
			field := reflect.ValueOf(&val).Elem()

			err := convertStringToUint(field, tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s'", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				}
				if field.Uint() != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, field.Uint())
				}
			}
		})
	}
}

// 测试 ConvertStringToFloat 错误场景
func Test_ConvertStringToFloatErrors(t *testing.T) {
	tests := []string{
		"abc",
		"",
		"not-a-number",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			var val float64
			field := reflect.ValueOf(&val).Elem()
			err := convertStringToFloat(field, tt)

			if err == nil {
				t.Errorf("Expected error for input '%s'", tt)
			}
		})
	}
}
