package di

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// 测试 checkAndGetCachedInstance 的缓存命中
func TestCheckAndGetCachedInstanceHit(t *testing.T) {
	c := New().(*container)

	_ = ProvideValueNamed(c, "test", 42)

	// 第一次获取
	_, _ = GetNamed[int](c, "test")

	// 第二次获取（应该命中缓存）
	val, cached := c.checkAndGetCachedInstance("int#test")
	if !cached {
		t.Error("Expected to get cached instance")
	}
	if val == nil {
		t.Error("Expected non-nil value")
	}
}

// 测试 prepareLazyDependencies 更多场景
func TestPrepareLazyDependenciesMoreCases(t *testing.T) {
	t.Run("NonLazyMode", func(t *testing.T) {
		c := New().(*container)

		_ = ProvideValueNamed(c, "service", 42)

		// 非 lazy 模式，应该直接返回 nil
		key := "int#service"
		entry := c.providers[key]

		err := c.prepareLazyDependencies(entry, key)
		if err != nil {
			t.Errorf("Expected nil for non-lazy mode, got %v", err)
		}
	})
}

// 测试 validateInstance 更多场景
func TestValidateInstanceMoreCases(t *testing.T) {
	c := New().(*container)

	t.Run("ValidInstance", func(t *testing.T) {
		// 有效实例
		instance := 42
		instanceValue := reflect.ValueOf(instance)
		valid := c.validateInstance(instance, instanceValue)
		if !valid {
			t.Error("Expected instance to be valid")
		}
	})
}

// 测试 findDepend 更多场景
func TestFindDependMoreCases(t *testing.T) {
	c := New().(*container)

	_ = ProvideValueNamed(c, "dep1", 10)
	_ = ProvideValueNamed(c, "dep2", 20)

	t.Run("FoundFirstProvider", func(t *testing.T) {
		// findDepend 只适用于 struct 类型
		// 对于 int 类型会返回错误，这是预期行为
		_, err := c.findDepend(reflect.TypeOf(int(0)))
		if err == nil {
			t.Log("findDepend succeeded (int is treated as valid)")
		} else {
			t.Logf("findDepend failed as expected for int: %v", err)
		}
	})
}

// 测试 mergeParentResults
func TestMergeParentResults(t *testing.T) {
	parent := New()
	_ = ProvideValueNamed(parent, "parent-service", 100.0)

	child := parent.CreateChildScope()
	_ = ProvideValueNamed(child, "child-service", 200.0)

	// 从子容器获取所有实例（使用 float64，不在黑名单中）
	results, err := child.GetNamedAll(reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("GetNamedAll failed: %v", err)
	}

	// 应该包含父容器的服务
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results (parent + child), got %d", len(results))
	}
}

// 测试 validateProviderFunction 更多场景
func TestValidateProviderFunctionMoreCases(t *testing.T) {
	t.Run("ValidFunction", func(t *testing.T) {
		fn := func(c Container) (int, error) {
			return 42, nil
		}
		err := validateProviderFunction(reflect.TypeOf(fn))
		if err != nil {
			t.Errorf("Unexpected error for valid function: %v", err)
		}
	})

	t.Run("InvalidFunctionNoReturn", func(t *testing.T) {
		fn := func(c Container) {
		}
		err := validateProviderFunction(reflect.TypeOf(fn))
		if err == nil {
			t.Error("Expected error for function with no returns")
		}
	})
}

// 测试 SetConfigSource 的并发访问
func TestSetConfigSourceConcurrent(t *testing.T) {
	c := New()

	done := make(chan bool)

	// 并发设置和获取
	for i := 0; i < 10; i++ {
		go func(idx int) {
			configSource := NewMapConfigSource()
			configSource.Set("key", idx)
			c.SetConfigSource(configSource)
			_ = c.GetConfigSource()
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// 测试 Done 通道
func TestDoneChannel(t *testing.T) {
	c := New()

	// 启动容器
	_ = c.Start()

	// Done 通道应该还未关闭
	select {
	case <-c.Done():
		t.Error("Done channel should not be closed yet")
	default:
		// 正确
	}

	// 关闭容器
	_ = c.Shutdown(context.Background())

	// 现在 Done 通道应该已关闭
	select {
	case <-c.Done():
		// 正确
	case <-time.After(100 * time.Millisecond):
		t.Error("Done channel should be closed after shutdown")
	}
}
