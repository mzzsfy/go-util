package di

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
)

// 测试 ShutdownOnSignals
func TestShutdownOnSignals(t *testing.T) {
	c := New()

	// 注册一个服务
	_ = ProvideValue(c, "test-service")

	// 调用 ShutdownOnSignals（不发送信号，只测试设置）
	c.ShutdownOnSignals()

	// 由于信号监听是异步的，我们只测试不会立即 panic
	time.Sleep(10 * time.Millisecond)

	// 手动关闭容器
	_ = c.Shutdown(context.Background())
}

func TestShutdownOnSignalsWithCustomSignals(t *testing.T) {
	c := New()

	// 使用自定义信号列表
	c.ShutdownOnSignals(os.Interrupt)

	// 给 goroutine 一点时间启动
	time.Sleep(10 * time.Millisecond)

	// 清理
	_ = c.Shutdown(context.Background())
}

func TestShutdownOnSignalsEmpty(t *testing.T) {
	c := New()

	// 不提供信号参数（使用默认）
	c.ShutdownOnSignals()

	time.Sleep(10 * time.Millisecond)

	_ = c.Shutdown(context.Background())
}

// 测试 GetAllInstances
func TestGetAllInstances(t *testing.T) {
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
func TestUpdateGetCallsStats(t *testing.T) {
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
func TestPrepareLazyDependenciesError(t *testing.T) {
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
func TestHasNamedAdvanced(t *testing.T) {
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
func TestFindDependAdvanced(t *testing.T) {
	c := New().(*container) // 类型断言

	// 测试查找不存在的依赖
	_, err := c.findDepend(reflect.TypeOf(""))
	if err == nil {
		t.Error("Expected error when finding non-existent dependency")
	}
}

// 测试 ConvertStringToUint
func TestConvertStringToUint(t *testing.T) {
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
func TestConvertStringToFloatErrors(t *testing.T) {
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
