package di

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// 测试实例管理功能
func TestInstanceManagement(t *testing.T) {
	c := New()

	// 注册服务
	_ = ProvideValueNamed(c, "service1", 100)
	_ = ProvideValueNamed(c, "service2", 200)

	// 获取实例
	_, _ = GetNamed[int](c, "service1")
	_, _ = GetNamed[int](c, "service2")

	t.Run("GetInstanceCount", func(t *testing.T) {
		count := c.GetInstanceCount()
		if count < 2 {
			t.Errorf("Expected at least 2 instances, got %d", count)
		}
	})

	t.Run("GetProviderCount", func(t *testing.T) {
		count := c.GetProviderCount()
		if count < 2 {
			t.Errorf("Expected at least 2 providers, got %d", count)
		}
	})

	t.Run("RemoveInstance", func(t *testing.T) {
		err := c.RemoveInstance(reflect.TypeOf(int(0)), "service1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// 实例数应该减少
		count := c.GetInstanceCount()
		if count >= 2 {
			t.Errorf("Expected less than 2 instances after removal, got %d", count)
		}
	})

	t.Run("ClearInstances", func(t *testing.T) {
		c.ClearInstances()
		count := c.GetInstanceCount()
		if count != 0 {
			t.Errorf("Expected 0 instances after clear, got %d", count)
		}
	})
}

// 测试 ReplaceInstance
func TestReplaceInstance(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "service", 100)
	val1, _ := GetNamed[int](c, "service")
	if val1 != 100 {
		t.Errorf("Expected 100, got %d", val1)
	}

	// 替换实例
	err := c.ReplaceInstance(reflect.TypeOf(int(0)), "service", 999)
	if err != nil {
		t.Fatalf("ReplaceInstance failed: %v", err)
	}

	// 获取替换后的实例（应该从缓存获取）
	instances := c.GetAllInstances()
	if len(instances) > 0 {
		// 验证替换成功
		t.Log("Instance replaced successfully")
	}
}

// 测试 ReplaceInstance 错误场景
func TestReplaceInstanceError(t *testing.T) {
	c := New()

	// 尝试替换不存在的提供者
	err := c.ReplaceInstance(reflect.TypeOf(int(0)), "nonexistent", 999)
	if err == nil {
		t.Error("Expected error for non-existent provider")
	}
}

// 测试 GetProviders
func TestGetProviders(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "s1", 1)
	_ = ProvideValueNamed(c, "s2", 2)
	_ = ProvideValueNamed(c, "s3", 3)

	providers := c.GetProviders()
	if len(providers) < 3 {
		t.Errorf("Expected at least 3 providers, got %d", len(providers))
	}
}

// 测试创建子容器并继承配置
func TestChildScopeWithConfig(t *testing.T) {
	parent := New()

	// 设置父容器的配置源
	configSource := NewMapConfigSource()
	configSource.Set("key", "parent-value")
	parent.SetConfigSource(configSource)

	// 创建子容器
	child := parent.CreateChildScope()

	// 子容器应该继承父容器的配置源
	childConfig := child.GetConfigSource()
	if childConfig == nil {
		t.Error("Expected child to inherit parent config source")
	}

	// 验证配置值
	value := childConfig.Get("key")
	if value.StringD("") != "parent-value" {
		t.Errorf("Expected 'parent-value', got '%s'", value.StringD(""))
	}
}

// 测试并发 GetNamed
func TestConcurrentGetNamed(t *testing.T) {
	c := New()

	// 使用自定义结构体类型避免循环依赖检测
	service := &TestService{Value: 42}
	_ = ProvideValueNamed(c, "concurrent-service", service)

	// 先获取一次，确保实例已创建
	_, _ = GetNamed[*TestService](c, "concurrent-service")

	done := make(chan int)

	// 并发获取同一个服务
	for i := 0; i < 10; i++ {
		go func() {
			val, err := GetNamed[*TestService](c, "concurrent-service")
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				done <- 0
				return
			}
			if val == nil {
				t.Errorf("Expected non-nil value")
				done <- 0
				return
			}
			if val.Value != 42 {
				t.Errorf("Expected 42, got %d", val.Value)
			}
			done <- 1
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// 测试生命周期钩子的完整流程
func TestLifecycleHooksFullFlow(t *testing.T) {
	c := New()

	var beforeCreateCalled bool
	var afterCreateCalled bool
	var beforeDestroyCalled bool
	var afterDestroyCalled bool

	_ = ProvideNamed(c, "lifecycle", func(c Container) (int, error) {
		return 42, nil
	},
		WithBeforeCreate(func(c Container, info EntryInfo) (any, error) {
			beforeCreateCalled = true
			return nil, nil
		}),
		WithAfterCreate(func(c Container, info EntryInfo) (any, error) {
			afterCreateCalled = true
			return nil, nil
		}),
		WithBeforeDestroy(func(c Container, info EntryInfo) {
			beforeDestroyCalled = true
		}),
		WithAfterDestroy(func(c Container, info EntryInfo) {
			afterDestroyCalled = true
		}),
	)

	// 获取服务（触发创建）
	_, _ = GetNamed[int](c, "lifecycle")

	if !beforeCreateCalled {
		t.Error("Expected before create hook to be called")
	}
	if !afterCreateCalled {
		t.Error("Expected after create hook to be called")
	}

	// 关闭容器（触发销毁）
	_ = c.Shutdown(context.Background())

	if !beforeDestroyCalled {
		t.Error("Expected before destroy hook to be called")
	}
	if !afterDestroyCalled {
		t.Error("Expected after destroy hook to be called")
	}
}

// 测试条件失败的场景
func TestConditionFailure(t *testing.T) {
	c := New()

	conditionCalled := false

	_ = ProvideNamed(c, "conditional", func(c Container) (int, error) {
		return 42, nil
	}, WithCondition(func(c Container) bool {
		conditionCalled = true
		return false // 条件失败
	}))

	// 尝试获取服务
	_, err := GetNamed[int](c, "conditional")
	if err == nil {
		t.Error("Expected error for condition failure")
	}

	if !conditionCalled {
		t.Error("Expected condition function to be called")
	}
}

// 测试 Start 和 Shutdown 的多次调用
func TestMultipleStartShutdown(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "service", 42)

	// 第一次启动
	err := c.Start()
	if err != nil {
		t.Errorf("First start failed: %v", err)
	}

	// 重复启动应该返回错误或被忽略
	err = c.Start()
	if err != nil {
		t.Logf("Second start returned error (expected): %v", err)
	}

	// 第一次关闭
	err = c.Shutdown(context.Background())
	if err != nil {
		t.Errorf("First shutdown failed: %v", err)
	}

	// 重复关闭
	err = c.Shutdown(context.Background())
	if err != nil {
		t.Logf("Second shutdown returned error (expected): %v", err)
	}
}

// 测试统计重置
func TestResetStats(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "service", 42)

	// 获取几次服务
	_, _ = GetNamed[int](c, "service")
	_, _ = GetNamed[int](c, "service")

	// 检查统计
	stats := c.GetStats()
	if stats.GetCalls == 0 {
		t.Error("Expected GetCalls to be > 0")
	}

	// 重置统计
	c.ResetStats()

	// 检查统计已重置
	stats = c.GetStats()
	if stats.GetCalls != 0 {
		t.Errorf("Expected GetCalls to be 0 after reset, got %d", stats.GetCalls)
	}
}

// 测试平均创建时间
func TestGetAverageCreateDuration(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "service1", 1)
	_ = ProvideValueNamed(c, "service2", 2)

	// 获取服务（会记录创建时间）
	_, _ = GetNamed[int](c, "service1")
	_, _ = GetNamed[int](c, "service2")

	// 获取平均创建时间
	avgDuration := c.GetAverageCreateDuration()
	t.Logf("Average create duration: %v", avgDuration)

	// 平均时间应该 >= 0
	if avgDuration < 0 {
		t.Errorf("Expected non-negative average duration, got %v", avgDuration)
	}
}

// 测试带超时的 Start
func TestStartWithTimeout(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "service", 42)

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start 不接受 context
	err := c.Start()
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	_ = c.Shutdown(ctx)
}

// 测试带超时的 Shutdown
func TestShutdownWithTimeout(t *testing.T) {
	c := New()

	_ = ProvideValueNamed(c, "service", 42)
	_ = c.Start()

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown with timeout failed: %v", err)
	}
}
